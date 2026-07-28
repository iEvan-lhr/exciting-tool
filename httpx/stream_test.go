package httpx

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func TestDoStreamValidationAndLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/pdf; charset=binary")
		_, _ = w.Write([]byte("12345"))
	}))
	defer server.Close()

	response, err := New().DoStream(context.Background(), http.MethodGet, server.URL, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Close()
	if !response.OK() {
		t.Fatalf("status = %d", response.StatusCode)
	}
	if err := response.RequireContentType("application/pdf"); err != nil {
		t.Fatal(err)
	}
	if err := response.RequireContentType("image/*"); !errors.Is(err, ErrUnexpectedContentType) {
		t.Fatalf("content type error = %v", err)
	}

	response.LimitBody(4)
	data, err := io.ReadAll(response.Body)
	if !errors.Is(err, ErrBodyTooLarge) {
		t.Fatalf("limited read error = %v", err)
	}
	if string(data) != "1234" {
		t.Fatalf("limited body = %q", data)
	}
}

func TestStreamCheckStatusBoundsErrorBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "failure", http.StatusBadGateway)
	}))
	defer server.Close()

	response, err := New().DoStream(context.Background(), http.MethodGet, server.URL, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = response.CheckStatus(4)
	var statusErr *StatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("status error = %v", err)
	}
	if statusErr.StatusCode != http.StatusBadGateway || string(statusErr.Body) != "fail" || !statusErr.Truncated {
		t.Fatalf("status error = %+v", statusErr)
	}
}

func TestRetryStatusAndHook(t *testing.T) {
	var attempts atomic.Int32
	var hooks atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if attempts.Add(1) < 3 {
			w.Header().Set("Retry-After", "0")
			http.Error(w, "retry", http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	client := New(WithRetry(RetryPolicy{
		MaxAttempts:       3,
		BaseDelay:         time.Microsecond,
		MaxDelay:          time.Millisecond,
		RespectRetryAfter: true,
		OnRetry: func(event RetryEvent) {
			if event.StatusCode != http.StatusServiceUnavailable {
				t.Errorf("retry status = %d", event.StatusCode)
			}
			hooks.Add(1)
		},
	}))
	response, err := client.DoStream(context.Background(), http.MethodGet, server.URL, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "ok" || attempts.Load() != 3 || hooks.Load() != 2 {
		t.Fatalf("body = %q, attempts = %d, hooks = %d", data, attempts.Load(), hooks.Load())
	}
}

func TestRetryReplayableRequestBody(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request: %v", err)
		}
		if string(data) != "payload" {
			t.Errorf("request body = %q", data)
		}
		if attempts.Add(1) == 1 {
			http.Error(w, "retry", http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	client := New(WithRetry(RetryPolicy{
		MaxAttempts: 2,
		BaseDelay:   time.Microsecond,
		MaxDelay:    time.Millisecond,
		Methods:     []string{http.MethodPost},
	}))
	response, err := client.DoRequest(context.Background(), Request{
		Method: http.MethodPost,
		URL:    server.URL,
		Body: func() (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("payload")), nil
		},
		ContentLength: 7,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer response.Close()
	if attempts.Load() != 2 {
		t.Fatalf("attempts = %d", attempts.Load())
	}
}

func TestRetryTransportError(t *testing.T) {
	var attempts atomic.Int32
	httpClient := &http.Client{
		Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			if attempts.Add(1) == 1 {
				return nil, errors.New("temporary transport failure")
			}
			return &http.Response{
				StatusCode:    http.StatusOK,
				Header:        make(http.Header),
				Body:          io.NopCloser(strings.NewReader("ok")),
				ContentLength: 2,
			}, nil
		}),
	}
	client := New(
		WithHTTPClient(httpClient),
		WithRetry(RetryPolicy{
			MaxAttempts:          2,
			BaseDelay:            time.Microsecond,
			MaxDelay:             time.Millisecond,
			RetryTransportErrors: true,
		}),
	)
	response, err := client.DoStream(context.Background(), http.MethodGet, "http://example.test", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Close()
	if attempts.Load() != 2 {
		t.Fatalf("attempts = %d", attempts.Load())
	}
}

func TestRetryRejectsOneShotBody(t *testing.T) {
	client := New(WithRetry(RetryPolicy{
		MaxAttempts: 2,
		Methods:     []string{http.MethodPost},
	}))
	body := io.LimitReader(strings.NewReader("payload"), 7)
	_, err := client.DoStream(context.Background(), http.MethodPost, "http://example.test", body, nil)
	if !errors.Is(err, ErrBodyNotReplayable) {
		t.Fatalf("error = %v", err)
	}
}

func TestRetryHonorsContextDuringBackoff(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "retry", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client := New(WithRetry(RetryPolicy{
		MaxAttempts: 3,
		BaseDelay:   time.Second,
		MaxDelay:    time.Second,
	}))
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	_, err := client.DoStream(ctx, http.MethodGet, server.URL, nil, nil)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("error = %v", err)
	}
}

func TestAllowHostsValidator(t *testing.T) {
	client := New(WithRequestValidator(AllowHosts("allowed.example")))
	_, err := client.DoStream(context.Background(), http.MethodGet, "http://blocked.example/path", nil, nil)
	if err == nil || !strings.Contains(err.Error(), "not allowed") {
		t.Fatalf("error = %v", err)
	}
}
