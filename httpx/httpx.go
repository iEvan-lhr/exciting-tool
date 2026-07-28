// Package httpx provides context-aware buffered and streaming HTTP helpers.
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
	httpClient        *http.Client
	headers           http.Header
	maxBodyBytes      int64
	retryPolicy       RetryPolicy
	requestValidators []RequestValidator
}

type Response struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

type StatusError struct {
	StatusCode int
	Header     http.Header
	Body       []byte
	Truncated  bool
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
		retryPolicy:  RetryPolicy{MaxAttempts: 1},
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

// WithRetry enables retry behavior for requests allowed by policy.
// MaxAttempts includes the initial request. A value smaller than two disables
// retries.
func WithRetry(policy RetryPolicy) Option {
	return func(client *Client) {
		client.retryPolicy = policy.clone()
	}
}

// WithRequestValidator adds a validator that runs before every request
// attempt. It can enforce policies such as allowed URL schemes or hosts.
func WithRequestValidator(validator RequestValidator) Option {
	return func(client *Client) {
		if validator != nil {
			client.requestValidators = append(client.requestValidators, validator)
		}
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
	stream, err := c.DoStream(ctx, method, url, body, headers)
	if err != nil {
		return nil, err
	}
	defer stream.Body.Close()

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
			Header:     response.Header.Clone(),
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

func (c *Client) bodyLimit() int64 {
	if c == nil || c.maxBodyBytes <= 0 {
		return defaultMaxBodyBytes
	}
	return c.maxBodyBytes
}
