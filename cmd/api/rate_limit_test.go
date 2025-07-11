package main

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func testHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"ok":true}`))
}

func TestRateLimiter_UnderLimit(t *testing.T) {
	limiter := NewIPRateLimitMiddleware(3, 100*time.Millisecond)
	defer limiter.Stop()
	ts := httptest.NewServer(limiter.Middleware(http.HandlerFunc(testHandler)))
	defer ts.Close()

	client := &http.Client{}
	for i := 0; i < 3; i++ {
		req, _ := http.NewRequest("GET", ts.URL, nil)
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}
	}
}

func TestRateLimiter_OverLimit(t *testing.T) {
	limiter := NewIPRateLimitMiddleware(2, 200*time.Millisecond)
	defer limiter.Stop()
	ts := httptest.NewServer(limiter.Middleware(http.HandlerFunc(testHandler)))
	defer ts.Close()

	client := &http.Client{}
	for i := 0; i < 2; i++ {
		resp, _ := client.Get(ts.URL)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}
	}
	resp, _ := client.Get(ts.URL)
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("Expected 429, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Too many requests") {
		t.Errorf("Expected error message, got %s", string(body))
	}
}

func TestRateLimiter_Expiration(t *testing.T) {
	limiter := NewIPRateLimitMiddleware(1, 100*time.Millisecond)
	defer limiter.Stop()
	ts := httptest.NewServer(limiter.Middleware(http.HandlerFunc(testHandler)))
	defer ts.Close()

	client := &http.Client{}
	resp, _ := client.Get(ts.URL)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}
	resp, _ = client.Get(ts.URL)
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("Expected 429, got %d", resp.StatusCode)
	}
	time.Sleep(120 * time.Millisecond)
	resp, _ = client.Get(ts.URL)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 after window reset, got %d", resp.StatusCode)
	}
}

func TestRateLimiter_MultipleIPs(t *testing.T) {
	limiter := NewIPRateLimitMiddleware(1, 100*time.Millisecond)
	defer limiter.Stop()
	ts := httptest.NewServer(limiter.Middleware(http.HandlerFunc(testHandler)))
	defer ts.Close()

	client := &http.Client{}
	// First IP
	req1, _ := http.NewRequest("GET", ts.URL, nil)
	req1.RemoteAddr = "1.2.3.4:1234"
	resp1, _ := client.Do(req1)
	if resp1.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", resp1.StatusCode)
	}
	// Second IP
	req2, _ := http.NewRequest("GET", ts.URL, nil)
	req2.RemoteAddr = "5.6.7.8:5678"
	resp2, _ := client.Do(req2)
	if resp2.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", resp2.StatusCode)
	}
}

func TestRateLimiter_Cleanup(t *testing.T) {
	limiter := NewIPRateLimitMiddleware(1, 50*time.Millisecond)
	limiter.cleanupInt = 50 * time.Millisecond // speed up cleanup
	defer limiter.Stop()

	key := "9.9.9.9|/test"
	entry := &rateLimitEntry{Count: 1, WindowStart: time.Now().Add(-11 * time.Minute), LastSeen: time.Now().Add(-11 * time.Minute)}
	limiter.store.Store(key, entry)

	// Wait for cleanup
	var found bool
	for i := 0; i < 10; i++ {
		time.Sleep(60 * time.Millisecond)
		_, found = limiter.store.Load(key)
		if !found {
			break
		}
	}
	if found {
		t.Errorf("Expected entry to be cleaned up, but it was still present")
	}
}

func TestRateLimiter_LogsOnLimit(t *testing.T) {
	limiter := NewIPRateLimitMiddleware(1, 100*time.Millisecond)
	defer limiter.Stop()
	ts := httptest.NewServer(limiter.Middleware(http.HandlerFunc(testHandler)))
	defer ts.Close()

	// Capture log output
	old := log.Writer()
	r, w, _ := os.Pipe()
	log.SetOutput(w)

	client := &http.Client{}
	client.Get(ts.URL) // first ok
	client.Get(ts.URL) // should trigger limit

	w.Close()
	log.SetOutput(old)

	out, _ := io.ReadAll(r)
	if !strings.Contains(string(out), "RATE LIMIT") {
		t.Errorf("Expected log output to contain 'RATE LIMIT', got %s", string(out))
	}
}
