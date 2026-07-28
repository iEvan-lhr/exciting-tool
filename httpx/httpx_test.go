package httpx

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestJSONHelpers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Client") != "exciting-tool" {
			t.Errorf("missing default header")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"method":"` + r.Method + `"}`))
	}))
	defer server.Close()

	client := New(WithHeader("X-Client", "exciting-tool"))
	var result struct {
		Method string `json:"method"`
	}
	response, err := client.PostJSON(context.Background(), server.URL, map[string]string{"value": "x"}, &result)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK || result.Method != http.MethodPost {
		t.Fatalf("response = %+v, result = %+v", response, result)
	}
}

func TestStatusAndBodyLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "large") {
			_, _ = w.Write([]byte("12345"))
			return
		}
		http.Error(w, "no", http.StatusBadRequest)
	}))
	defer server.Close()

	client := New(WithMaxBodyBytes(4))
	if _, err := client.Do(context.Background(), http.MethodGet, server.URL+"/large", nil, nil); !errors.Is(err, ErrBodyTooLarge) {
		t.Fatalf("large response error = %v", err)
	}
	_, err := client.GetJSON(context.Background(), server.URL+"/status", nil)
	var statusErr *StatusError
	if !errors.As(err, &statusErr) || statusErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("status error = %v", err)
	}
}
