package httpx

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	ErrInvalidMultipartName = errors.New("multipart field name or filename is invalid")
	ErrMultipartStarted     = errors.New("multipart body has already been opened")
)

type multipartField struct {
	name  string
	value string
}

type multipartFile struct {
	fieldName string
	fileName  string
	open      BodyFactory
}

// Multipart builds replayable multipart/form-data bodies without buffering
// complete files in memory. Build the form before sending it.
type Multipart struct {
	mutex     sync.Mutex
	boundary  string
	fields    []multipartField
	files     []multipartFile
	sealed    bool
	openCount int
	oneShot   bool
}

func NewMultipart() *Multipart {
	return &Multipart{
		boundary: newMultipartBoundary(),
	}
}

func (m *Multipart) AddField(name, value string) error {
	if err := validateMultipartName(name); err != nil {
		return err
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.sealed {
		return ErrMultipartStarted
	}
	m.fields = append(m.fields, multipartField{name: name, value: value})
	return nil
}

// AddBytes adds an in-memory file and copies data so later caller mutations do
// not change the request.
func (m *Multipart) AddBytes(fieldName, fileName string, data []byte) error {
	copied := append([]byte(nil), data...)
	return m.addFile(fieldName, fileName, func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(copied)), nil
	}, true)
}

// AddFile adds a local file. The file is reopened for every request attempt.
func (m *Multipart) AddFile(fieldName, path string) error {
	fileName := filepath.Base(path)
	return m.addFile(fieldName, fileName, func() (io.ReadCloser, error) {
		return os.Open(path)
	}, true)
}

// AddReader adds a one-shot reader. Forms containing a one-shot reader cannot
// be retried; use AddReaderFunc when the source can be reopened.
func (m *Multipart) AddReader(fieldName, fileName string, reader io.Reader) error {
	if reader == nil {
		return errors.New("multipart reader cannot be nil")
	}
	return m.addFile(fieldName, fileName, func() (io.ReadCloser, error) {
		if closer, ok := reader.(io.ReadCloser); ok {
			return closer, nil
		}
		return io.NopCloser(reader), nil
	}, false)
}

// AddReaderFunc adds a replayable file source. Open must return a fresh reader
// each time it is called.
func (m *Multipart) AddReaderFunc(fieldName, fileName string, open BodyFactory) error {
	if open == nil {
		return errors.New("multipart body factory cannot be nil")
	}
	return m.addFile(fieldName, fileName, open, true)
}

func (m *Multipart) addFile(
	fieldName string,
	fileName string,
	open BodyFactory,
	replayable bool,
) error {
	if err := validateMultipartName(fieldName); err != nil {
		return err
	}
	if err := validateMultipartName(fileName); err != nil {
		return err
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.sealed {
		return ErrMultipartStarted
	}
	m.files = append(m.files, multipartFile{
		fieldName: fieldName,
		fileName:  fileName,
		open:      open,
	})
	m.oneShot = m.oneShot || !replayable
	return nil
}

func (m *Multipart) ContentType() string {
	if m == nil {
		return ""
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.ensureBoundary()
	return "multipart/form-data; boundary=" + m.boundary
}

func (m *Multipart) Replayable() bool {
	if m == nil {
		return false
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return !m.oneShot
}

// Open creates a streaming body. It satisfies BodyFactory.
func (m *Multipart) Open() (io.ReadCloser, error) {
	if m == nil {
		return nil, errors.New("multipart form cannot be nil")
	}
	m.mutex.Lock()
	if m.oneShot && m.openCount > 0 {
		m.mutex.Unlock()
		return nil, ErrBodyNotReplayable
	}
	m.sealed = true
	m.openCount++
	m.ensureBoundary()
	boundary := m.boundary
	fields := append([]multipartField(nil), m.fields...)
	files := append([]multipartFile(nil), m.files...)
	m.mutex.Unlock()

	reader, writer := io.Pipe()
	go writeMultipartBody(writer, boundary, fields, files)
	return reader, nil
}

func (m *Multipart) ensureBoundary() {
	if m.boundary == "" {
		m.boundary = newMultipartBoundary()
	}
}

func newMultipartBoundary() string {
	writer := multipart.NewWriter(io.Discard)
	boundary := writer.Boundary()
	_ = writer.Close()
	return boundary
}

func validateMultipartName(value string) error {
	if value == "" || strings.ContainsAny(value, "\r\n") {
		return ErrInvalidMultipartName
	}
	return nil
}

func writeMultipartBody(
	pipe *io.PipeWriter,
	boundary string,
	fields []multipartField,
	files []multipartFile,
) {
	writer := multipart.NewWriter(pipe)
	if err := writer.SetBoundary(boundary); err != nil {
		_ = pipe.CloseWithError(err)
		return
	}
	for _, field := range fields {
		if err := writer.WriteField(field.name, field.value); err != nil {
			_ = pipe.CloseWithError(err)
			return
		}
	}
	for _, file := range files {
		part, err := writer.CreateFormFile(file.fieldName, file.fileName)
		if err != nil {
			_ = pipe.CloseWithError(err)
			return
		}
		source, err := file.open()
		if err != nil {
			_ = pipe.CloseWithError(err)
			return
		}
		_, copyErr := io.Copy(part, source)
		closeErr := source.Close()
		if copyErr != nil {
			_ = pipe.CloseWithError(copyErr)
			return
		}
		if closeErr != nil {
			_ = pipe.CloseWithError(closeErr)
			return
		}
	}
	if err := writer.Close(); err != nil {
		_ = pipe.CloseWithError(err)
		return
	}
	_ = pipe.Close()
}

// PostMultipartStream sends a streaming multipart request and leaves the
// response body open for the caller.
func (c *Client) PostMultipartStream(
	ctx context.Context,
	url string,
	form *Multipart,
	headers http.Header,
) (*StreamResponse, error) {
	if form == nil {
		return nil, errors.New("multipart form cannot be nil")
	}
	policy := c.retry()
	if policy.MaxAttempts > 1 && policy.allowsMethod(http.MethodPost) && !form.Replayable() {
		return nil, ErrBodyNotReplayable
	}
	headers = headers.Clone()
	if headers == nil {
		headers = make(http.Header)
	}
	headers.Set("Content-Type", form.ContentType())
	return c.DoRequest(ctx, Request{
		Method:        http.MethodPost,
		URL:           url,
		Header:        headers,
		Body:          form.Open,
		ContentLength: -1,
	})
}

// PostMultipart sends a multipart request and buffers the bounded response.
func (c *Client) PostMultipart(
	ctx context.Context,
	url string,
	form *Multipart,
	headers http.Header,
) (*Response, error) {
	stream, err := c.PostMultipartStream(ctx, url, form, headers)
	if err != nil {
		return nil, err
	}
	defer stream.Close()
	limit := c.bodyLimit()
	data, err := io.ReadAll(io.LimitReader(stream.Body, limit+1))
	result := &Response{
		StatusCode: stream.StatusCode,
		Header:     stream.Header.Clone(),
		Body:       data,
	}
	if err != nil {
		return result, err
	}
	if int64(len(data)) > limit {
		result.Body = result.Body[:limit]
		return result, ErrBodyTooLarge
	}
	return result, nil
}
