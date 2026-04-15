package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientAddsHeadersAndIdempotency(t *testing.T) {
	var gotAPIKey string
	var gotIdempotency string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAPIKey = r.Header.Get("X-API-Key")
		gotIdempotency = r.Header.Get("Idempotency-Key")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	client := newAPIClient(apiClientConfig{
		APIKey:  "k_test",
		BaseURL: srv.URL,
		Timeout: 2 * time.Second,
	})

	type outModel struct {
		OK bool `json:"ok"`
	}
	var out outModel
	err := client.post(context.Background(), "/instances", nil, map[string]string{"name": "x"}, "idem-1", &out)
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}

	if gotAPIKey != "k_test" {
		t.Fatalf("expected api key header, got %q", gotAPIKey)
	}
	if gotIdempotency != "idem-1" {
		t.Fatalf("expected idempotency header, got %q", gotIdempotency)
	}
	if !out.OK {
		t.Fatalf("expected response decode")
	}
}

func TestDescribeAPIError(t *testing.T) {
	err := &apiError{statusCode: http.StatusUnauthorized}
	got := describeAPIError(err)
	if got == "" {
		t.Fatalf("expected non-empty diagnostic description")
	}
}
