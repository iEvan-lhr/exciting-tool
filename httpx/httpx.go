// Package httpx provides a small, context-aware HTTP client with bounded
// response reads and JSON helpers.
package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultMaxBodyBytes int64 = 8 << 20

var ErrBodyTooLarge = errors.New("http response body exceeds configured limit")

type Option func(*Client)

type Client struct {
	httpClient   *http.Client
	headers      http.Header
	maxBodyBytes int64
}

type Response struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

type StatusError struct {
	StatusCode int
	Body       []byte
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("unexpected HTTP status %d", e.StatusCode)
}

// New creates a client with a 30-second timeout and an 8 MiB body limit.
func New(options ...Option) *Client {
	client := &Client{
		httpClient:   &http.Client{Timeout: 30 * time.Second},
		headers:      make(http.Header),
		maxBodyBytes: defaultMaxBodyBytes,
	}
	for _, option := range options {
		if option != nil {
			option(client)
		}
	}
	return client
}

func WithHTTPClient(httpClient *http.Client) Option {
	return func(client *Client) {
		if httpClient != nil {
			client.httpClient = httpClient
		}
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(client *Client) {
		copy := *client.httpClient
		copy.Timeout = timeout
		client.httpClient = &copy
	}
}

func WithMaxBodyBytes(limit int64) Option {
	return func(client *Client) {
		if limit > 0 {
			client.maxBodyBytes = limit
		}
	}
}

func WithHeader(key, value string) Option {
	return func(client *Client) {
		client.headers.Set(key, value)
	}
}

func (r *Response) OK() bool {
	return r != nil && r.StatusCode >= http.StatusOK && r.StatusCode < http.StatusMultipleChoices
}

func (r *Response) DecodeJSON(target any) error {
	if r == nil {
		return errors.New("cannot decode a nil response")
	}
	if target == nil {
		return nil
	}
	return json.Unmarshal(r.Body, target)
}

func (c *Client) Do(
	ctx context.Context,
	method string,
	url string,
	body io.Reader,
	headers http.Header,
) (*Response, error) {
	if c == nil {
		c = New()
	}
	if ctx == nil {
		ctx = context.Background()
	}
	httpClient := c.httpClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	limit := c.maxBodyBytes
	if limit <= 0 {
		limit = defaultMaxBodyBytes
	}
	request, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	request.Header = c.headers.Clone()
	for key, values := range headers {
		request.Header.Del(key)
		for _, value := range values {
			request.Header.Add(key, value)
		}
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	data, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	result := &Response{
		StatusCode: response.StatusCode,
		Header:     response.Header.Clone(),
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

func (c *Client) DoJSON(
	ctx context.Context,
	method string,
	url string,
	payload any,
	target any,
	headers http.Header,
) (*Response, error) {
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(data)
	}
	headers = headers.Clone()
	if headers == nil {
		headers = make(http.Header)
	}
	headers.Set("Accept", "application/json")
	if payload != nil {
		headers.Set("Content-Type", "application/json")
	}

	response, err := c.Do(ctx, method, url, body, headers)
	if err != nil {
		return response, err
	}
	if !response.OK() {
		return response, &StatusError{
			StatusCode: response.StatusCode,
			Body:       append([]byte(nil), response.Body...),
		}
	}
	if target != nil {
		if err := response.DecodeJSON(target); err != nil {
			return response, err
		}
	}
	return response, nil
}

func (c *Client) GetJSON(ctx context.Context, url string, target any) (*Response, error) {
	return c.DoJSON(ctx, http.MethodGet, url, nil, target, nil)
}

func (c *Client) PostJSON(ctx context.Context, url string, payload, target any) (*Response, error) {
	return c.DoJSON(ctx, http.MethodPost, url, payload, target, nil)
}
