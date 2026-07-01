package qwen_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/leketech/OpsPilot-AI/backend/internal/config"
	"github.com/leketech/OpsPilot-AI/backend/internal/llm/qwen"
)

// newTestClient points a Client at srv and uses a dummy API key and model.
func newTestClient(srv *httptest.Server) *qwen.Client {
	return qwen.New(&config.Config{
		QwenAPIKey:  "test-key",
		QwenBaseURL: srv.URL,
		QwenModel:   "test-model",
	})
}

// okResponse builds a minimal valid completionResponse JSON string.
func okResponse(content string) []byte {
	resp := map[string]any{
		"choices": []map[string]any{
			{
				"message":       map[string]any{"role": "assistant", "content": content},
				"finish_reason": "stop",
			},
		},
		"usage": map[string]any{
			"prompt_tokens":     10,
			"completion_tokens": 5,
			"total_tokens":      15,
		},
	}
	b, _ := json.Marshal(resp)
	return b
}

// ── Tests ─────────────────────────────────────────────────────────────────────

// TestChat_RequestSerialization verifies that the client sends the correct
// Authorization header, Content-Type, and body to the API.
func TestChat_RequestSerialization(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/chat/completions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("unexpected auth header: %s", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("unexpected content-type: %s", got)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("could not decode request body: %v", err)
		}
		if body["model"] != "test-model" {
			t.Errorf("unexpected model: %v", body["model"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(okResponse("hello"))
	}))
	defer srv.Close()

	client := newTestClient(srv)
	result, err := client.Chat(context.Background(), []qwen.Message{
		{Role: "user", Content: "ping"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Content != "hello" {
		t.Errorf("unexpected content: %q", result.Content)
	}
}

// TestChat_ResponseParsing verifies that token usage is captured correctly.
func TestChat_ResponseParsing(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(okResponse("world"))
	}))
	defer srv.Close()

	client := newTestClient(srv)
	result, err := client.Chat(context.Background(), []qwen.Message{{Role: "user", Content: "hi"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Content != "world" {
		t.Errorf("expected 'world', got %q", result.Content)
	}
	if result.Usage.PromptTokens != 10 {
		t.Errorf("expected 10 prompt tokens, got %d", result.Usage.PromptTokens)
	}
	if result.Usage.CompletionTokens != 5 {
		t.Errorf("expected 5 completion tokens, got %d", result.Usage.CompletionTokens)
	}
	if result.Usage.TotalTokens != 15 {
		t.Errorf("expected 15 total tokens, got %d", result.Usage.TotalTokens)
	}
}

// TestChat_RetryOnRateLimit verifies that the client retries on HTTP 429 and
// eventually succeeds when the server recovers.
func TestChat_RetryOnRateLimit(t *testing.T) {
	var callCount int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&callCount, 1)
		if n < 3 {
			// First two calls fail with rate-limit.
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":{"message":"rate limit exceeded","code":"429"}}`))
			return
		}
		// Third call succeeds.
		w.Header().Set("Content-Type", "application/json")
		w.Write(okResponse("recovered"))
	}))
	defer srv.Close()

	client := newTestClient(srv)
	result, err := client.Chat(context.Background(), []qwen.Message{{Role: "user", Content: "retry me"}})
	if err != nil {
		t.Fatalf("expected success after retries, got error: %v", err)
	}
	if result.Content != "recovered" {
		t.Errorf("unexpected content: %q", result.Content)
	}
	if atomic.LoadInt32(&callCount) != 3 {
		t.Errorf("expected 3 HTTP calls (1 + 2 retries), got %d", callCount)
	}
}

// TestChat_NoRetryOnClientError verifies that 4xx errors (other than 429)
// are NOT retried — they indicate a programming error, not a transient issue.
func TestChat_NoRetryOnClientError(t *testing.T) {
	var callCount int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":{"message":"invalid request","code":"400"}}`))
	}))
	defer srv.Close()

	client := newTestClient(srv)
	_, err := client.Chat(context.Background(), []qwen.Message{{Role: "user", Content: "bad request"}})
	if err == nil {
		t.Fatal("expected error for 400, got nil")
	}
	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("expected exactly 1 HTTP call for non-retryable error, got %d", callCount)
	}
}

// TestChat_ContextCancellation verifies that a cancelled context stops the
// request immediately and returns a context error.
func TestChat_ContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate a slow server that takes longer than the cancelled context.
		time.Sleep(200 * time.Millisecond)
		w.Write(okResponse("too late"))
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	client := newTestClient(srv)
	_, err := client.Chat(ctx, []qwen.Message{{Role: "user", Content: "hello"}})
	if err == nil {
		t.Fatal("expected context cancellation error, got nil")
	}
}

// TestCompleteJSON_ParsesStructuredResponse verifies that CompleteJSON
// correctly unmarshals a JSON body into the target struct.
func TestCompleteJSON_ParsesStructuredResponse(t *testing.T) {
	type result struct {
		Answer string `json:"answer"`
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// The model responds with a JSON object as its content.
		w.Write(okResponse(`{"answer":"42"}`))
	}))
	defer srv.Close()

	client := newTestClient(srv)
	var out result
	_, err := client.CompleteJSON(context.Background(), []qwen.Message{{Role: "user", Content: "what is the answer?"}}, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Answer != "42" {
		t.Errorf("expected answer '42', got %q", out.Answer)
	}
}
