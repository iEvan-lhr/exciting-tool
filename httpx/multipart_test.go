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

func TestPostMultipartStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if value := r.FormValue("mode"); value != "convert" {
			t.Errorf("mode = %q", value)
		}
		file, header, err := r.FormFile("files")
		if err != nil {
			t.Errorf("form file: %v", err)
			return
		}
		defer file.Close()
		data, err := io.ReadAll(file)
		if err != nil {
			t.Errorf("read file: %v", err)
		}
		if header.Filename != "deck.pptx" || string(data) != "pptx-data" {
			t.Errorf("file = %q, data = %q", header.Filename, data)
		}
		w.Header().Set("Content-Type", "application/pdf")
		_, _ = w.Write([]byte("pdf-data"))
	}))
	defer server.Close()

	form := NewMultipart()
	if err := form.AddField("mode", "convert"); err != nil {
		t.Fatal(err)
	}
	if err := form.AddBytes("files", "deck.pptx", []byte("pptx-data")); err != nil {
		t.Fatal(err)
	}
	response, err := New().PostMultipartStream(context.Background(), server.URL, form, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Close()
	if err := response.CheckStatus(0); err != nil {
		t.Fatal(err)
	}
	if err := response.RequireContentType("application/pdf"); err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(response.Body)
	if err != nil || string(data) != "pdf-data" {
		t.Fatalf("body = %q, error = %v", data, err)
	}
}

func TestMultipartRetryReopensBody(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		data, err := io.ReadAll(file)
		_ = file.Close()
		if err != nil || string(data) != "data" {
			t.Errorf("body = %q, error = %v", data, err)
		}
		if attempts.Add(1) == 1 {
			http.Error(w, "retry", http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	form := NewMultipart()
	if err := form.AddBytes("file", "file.bin", []byte("data")); err != nil {
		t.Fatal(err)
	}
	client := New(WithRetry(RetryPolicy{
		MaxAttempts: 2,
		BaseDelay:   time.Microsecond,
		MaxDelay:    time.Millisecond,
		Methods:     []string{http.MethodPost},
	}))
	response, err := client.PostMultipartStream(context.Background(), server.URL, form, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Close()
	if attempts.Load() != 2 {
		t.Fatalf("attempts = %d", attempts.Load())
	}
}

func TestMultipartOneShotReaderCannotRetry(t *testing.T) {
	form := NewMultipart()
	if err := form.AddReader("file", "file.bin", strings.NewReader("data")); err != nil {
		t.Fatal(err)
	}
	client := New(WithRetry(RetryPolicy{
		MaxAttempts: 2,
		Methods:     []string{http.MethodPost},
	}))
	_, err := client.PostMultipartStream(context.Background(), "http://example.test", form, nil)
	if !errors.Is(err, ErrBodyNotReplayable) {
		t.Fatalf("error = %v", err)
	}
}

func TestMultipartRejectsHeaderInjectionAndMutationAfterOpen(t *testing.T) {
	form := NewMultipart()
	if err := form.AddField("bad\r\nname", "value"); !errors.Is(err, ErrInvalidMultipartName) {
		t.Fatalf("invalid name error = %v", err)
	}
	body, err := form.Open()
	if err != nil {
		t.Fatal(err)
	}
	_ = body.Close()
	if err := form.AddField("late", "value"); !errors.Is(err, ErrMultipartStarted) {
		t.Fatalf("late mutation error = %v", err)
	}
}
