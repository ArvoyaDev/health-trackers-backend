package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimitMiddleware(t *testing.T) {
	// Create a test server with the rate-limited handler
	handler := rateLimitMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	server := httptest.NewServer(handler)
	defer server.Close()

	client := &http.Client{}
	url := server.URL

	// Send multiple requests in quick succession
	requestCount := 20
	successCount := 0
	rateLimitCount := 0

	for i := 0; i < requestCount; i++ {
		resp, err := client.Get(url)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			successCount++
		} else if resp.StatusCode == http.StatusTooManyRequests {
			rateLimitCount++
		} else {
			t.Fatalf("Unexpected status code: %d", resp.StatusCode)
		}

		// Sleep for a short duration to simulate quick succession
		time.Sleep(50 * time.Millisecond)
	}

	// Check that some requests were rate-limited
	if rateLimitCount == 0 {
		t.Errorf("Expected some requests to be rate-limited, but none were")
	}

	t.Logf("Success count: %d, Rate limit count: %d", successCount, rateLimitCount)
}
