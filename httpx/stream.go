package httpx

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
	"sync"
	"time"
)

const defaultMaxErrorBodyBytes int64 = 64 << 10

var ErrUnexpectedContentType = errors.New("unexpected response content type")

// BodyFactory returns a new request body for each attempt. Implementations
// must be safe to call again when retries are enabled.
type BodyFactory func() (io.ReadCloser, error)

// Request describes a request with an optional replayable body.
type Request struct {
	Method        string
	URL           string
	Header        http.Header
	Body          BodyFactory
	ContentLength int64
}

// StreamResponse exposes the response body without buffering it. The caller
// owns Body and must close it.
type StreamResponse struct {
	StatusCode    int
	Header        http.Header
	Body          io.ReadCloser
	ContentLength int64
}

type ContentTypeError struct {
	Actual  string
	Allowed []string
}

func (e *ContentTypeError) Error() string {
	return fmt.Sprintf("%v: got %q, allowed %s", ErrUnexpectedContentType, e.Actual, strings.Join(e.Allowed, ", "))
}

func (e *ContentTypeError) Unwrap() error {
	return ErrUnexpectedContentType
}

func (r *StreamResponse) OK() bool {
	return r != nil && r.StatusCode >= http.StatusOK && r.StatusCode < http.StatusMultipleChoices
}

func (r *StreamResponse) Close() error {
	if r == nil || r.Body == nil {
		return nil
	}
	return r.Body.Close()
}

// CheckStatus consumes and closes the body only for non-2xx responses. Error
// bodies are bounded so a server cannot force an unbounded allocation.
func (r *StreamResponse) CheckStatus(maxErrorBodyBytes int64) error {
	if r == nil {
		return errors.New("cannot check a nil response")
	}
	if r.OK() {
		return nil
	}
	if maxErrorBodyBytes <= 0 {
		maxErrorBodyBytes = defaultMaxErrorBodyBytes
	}
	if r.Body == nil {
		return &StatusError{
			StatusCode: r.StatusCode,
			Header:     r.Header.Clone(),
		}
	}
	defer r.Close()
	data, err := io.ReadAll(io.LimitReader(r.Body, maxErrorBodyBytes+1))
	if err != nil {
		return err
	}
	truncated := int64(len(data)) > maxErrorBodyBytes
	if truncated {
		data = data[:maxErrorBodyBytes]
	}
	return &StatusError{
		StatusCode: r.StatusCode,
		Header:     r.Header.Clone(),
		Body:       data,
		Truncated:  truncated,
	}
}

// RequireContentType validates the response media type. Patterns such as
// "image/*" are supported.
func (r *StreamResponse) RequireContentType(allowed ...string) error {
	if r == nil {
		return errors.New("cannot validate a nil response")
	}
	actual := r.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(actual)
	if err != nil {
		return &ContentTypeError{Actual: actual, Allowed: append([]string(nil), allowed...)}
	}
	for _, candidate := range allowed {
		candidateType, _, parseErr := mime.ParseMediaType(candidate)
		if parseErr != nil {
			candidateType = strings.TrimSpace(candidate)
		}
		if contentTypeMatches(mediaType, candidateType) {
			return nil
		}
	}
	return &ContentTypeError{Actual: mediaType, Allowed: append([]string(nil), allowed...)}
}

// LimitBody replaces Body with a reader that returns ErrBodyTooLarge if more
// than maxBytes are read.
func (r *StreamResponse) LimitBody(maxBytes int64) {
	if r == nil || r.Body == nil || maxBytes <= 0 {
		return
	}
	r.Body = &limitedReadCloser{
		body:      r.Body,
		remaining: maxBytes,
	}
}

func contentTypeMatches(actual, allowed string) bool {
	actual = strings.ToLower(strings.TrimSpace(actual))
	allowed = strings.ToLower(strings.TrimSpace(allowed))
	if actual == allowed {
		return true
	}
	if strings.HasSuffix(allowed, "/*") {
		return strings.HasPrefix(actual, strings.TrimSuffix(allowed, "*"))
	}
	return false
}

type limitedReadCloser struct {
	body      io.ReadCloser
	remaining int64
	exceeded  bool
}

func (r *limitedReadCloser) Read(target []byte) (int, error) {
	if r.exceeded {
		return 0, ErrBodyTooLarge
	}
	if r.remaining == 0 {
		var extra [1]byte
		count, err := r.body.Read(extra[:])
		if count > 0 {
			r.exceeded = true
			return 0, ErrBodyTooLarge
		}
		return 0, err
	}
	if int64(len(target)) > r.remaining {
		target = target[:r.remaining]
	}
	count, err := r.body.Read(target)
	r.remaining -= int64(count)
	return count, err
}

func (r *limitedReadCloser) Close() error {
	return r.body.Close()
}

// DoStream sends one streaming request. When retries are configured, body
// must be one of the replayable reader types recognized by net/http.
func (c *Client) DoStream(
	ctx context.Context,
	method string,
	url string,
	body io.Reader,
	headers http.Header,
) (*StreamResponse, error) {
	probe, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	var mutex sync.Mutex
	firstBody := probe.Body
	firstUsed := false
	var factory BodyFactory
	if body != nil {
		factory = func() (io.ReadCloser, error) {
			mutex.Lock()
			defer mutex.Unlock()
			if !firstUsed {
				firstUsed = true
				return firstBody, nil
			}
			if probe.GetBody == nil {
				return nil, ErrBodyNotReplayable
			}
			return probe.GetBody()
		}
	}

	policy := c.retry()
	if body != nil && policy.MaxAttempts > 1 && policy.allowsMethod(probe.Method) && probe.GetBody == nil {
		_ = probe.Body.Close()
		return nil, ErrBodyNotReplayable
	}
	return c.DoRequest(ctx, Request{
		Method:        probe.Method,
		URL:           url,
		Header:        headers,
		Body:          factory,
		ContentLength: probe.ContentLength,
	})
}

// DoRequest sends a streaming request. Request.Body must create a fresh body
// for every attempt.
func (c *Client) DoRequest(ctx context.Context, request Request) (*StreamResponse, error) {
	if c == nil {
		c = New()
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if request.Method == "" {
		request.Method = http.MethodGet
	}

	policy := c.retry()
	attempts := 1
	if policy.allowsMethod(request.Method) {
		attempts = policy.MaxAttempts
	}

	for attempt := 1; attempt <= attempts; attempt++ {
		response, err := c.doAttempt(ctx, request)
		shouldRetry := attempt < attempts &&
			((err != nil && policy.RetryTransportErrors) ||
				(err == nil && response != nil && policy.allowsStatus(response.StatusCode)))
		if !shouldRetry || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return response, err
		}

		var rawResponse *http.Response
		if response != nil {
			rawResponse = &http.Response{
				StatusCode: response.StatusCode,
				Header:     response.Header,
			}
			drainAndClose(response.Body)
		}
		delay := policy.delay(attempt, rawResponse)
		if policy.OnRetry != nil {
			event := RetryEvent{
				Attempt:     attempt,
				NextAttempt: attempt + 1,
				Method:      request.Method,
				URL:         request.URL,
				Err:         err,
				Delay:       delay,
			}
			if response != nil {
				event.StatusCode = response.StatusCode
			}
			policy.OnRetry(event)
		}
		if err := waitForRetry(ctx, delay); err != nil {
			return nil, err
		}
	}
	return nil, errors.New("http retry loop ended unexpectedly")
}

func (c *Client) doAttempt(ctx context.Context, spec Request) (*StreamResponse, error) {
	var body io.ReadCloser
	var err error
	if spec.Body != nil {
		body, err = spec.Body()
		if err != nil {
			return nil, err
		}
	}

	request, err := http.NewRequestWithContext(ctx, spec.Method, spec.URL, body)
	if err != nil {
		if body != nil {
			_ = body.Close()
		}
		return nil, err
	}
	if body != nil && spec.ContentLength >= 0 {
		request.ContentLength = spec.ContentLength
	}
	request.Header = c.headers.Clone()
	mergeHeaders(request.Header, spec.Header)

	for _, validator := range c.requestValidators {
		if err := validator(request); err != nil {
			if request.Body != nil {
				_ = request.Body.Close()
			}
			return nil, err
		}
	}

	httpClient := c.httpClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	response, err := httpClient.Do(request)
	if err != nil {
		if response != nil && response.Body != nil {
			_ = response.Body.Close()
		}
		return nil, err
	}
	return &StreamResponse{
		StatusCode:    response.StatusCode,
		Header:        response.Header.Clone(),
		Body:          response.Body,
		ContentLength: response.ContentLength,
	}, nil
}

func (c *Client) retry() RetryPolicy {
	if c == nil {
		return RetryPolicy{MaxAttempts: 1}.normalized()
	}
	return c.retryPolicy.normalized()
}

func mergeHeaders(target, override http.Header) {
	for key, values := range override {
		target.Del(key)
		for _, value := range values {
			target.Add(key, value)
		}
	}
}

func drainAndClose(body io.ReadCloser) {
	if body == nil {
		return
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(body, 32<<10))
	_ = body.Close()
}

// AllowHosts returns a request validator that accepts only the provided host
// names. An entry with a port matches URL.Host; otherwise it matches Hostname.
func AllowHosts(hosts ...string) RequestValidator {
	allowed := make(map[string]struct{}, len(hosts))
	for _, host := range hosts {
		host = strings.ToLower(strings.TrimSpace(host))
		if host != "" {
			allowed[host] = struct{}{}
		}
	}
	return func(request *http.Request) error {
		host := strings.ToLower(request.URL.Host)
		hostname := strings.ToLower(request.URL.Hostname())
		if _, ok := allowed[host]; ok {
			return nil
		}
		if _, ok := allowed[hostname]; ok {
			return nil
		}
		return fmt.Errorf("request host %q is not allowed", request.URL.Host)
	}
}
