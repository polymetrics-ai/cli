package rlm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLLMConfigFromEnv_DefaultsOpenRouter(t *testing.T) {
	env := map[string]string{"OPENROUTER_API_KEY": "sk-or-xyz", "PM_LLM_MODEL": "anthropic/claude-fable-5"}
	cfg := LLMConfigFromEnv(func(k string) string { return env[k] })
	if cfg.Provider != "openrouter" {
		t.Fatalf("provider = %q", cfg.Provider)
	}
	if cfg.BaseURL != openrouterBaseURL {
		t.Fatalf("base = %q, want %q", cfg.BaseURL, openrouterBaseURL)
	}
	if cfg.APIKey != "sk-or-xyz" {
		t.Fatalf("key = %q", cfg.APIKey)
	}
	if !cfg.Resolvable() {
		t.Fatal("config should be resolvable")
	}
}

func TestLLMConfigFromEnv_CustomBaseAndKey(t *testing.T) {
	env := map[string]string{
		"PM_LLM_PROVIDER": "openai",
		"PM_LLM_BASE_URL": "https://api.openai.com/v1",
		"PM_LLM_API_KEY":  "sk-abc",
		"PM_LLM_MODEL":    "gpt-5.4",
	}
	cfg := LLMConfigFromEnv(func(k string) string { return env[k] })
	if cfg.BaseURL != "https://api.openai.com/v1" || cfg.APIKey != "sk-abc" || cfg.Model != "gpt-5.4" {
		t.Fatalf("unexpected cfg: %+v", cfg)
	}
}

func TestLLMConfig_Resolvable_LocalKeyless(t *testing.T) {
	cfg := LLMConfig{BaseURL: "http://localhost:11434/v1", Model: "llama3"}
	if !cfg.Resolvable() {
		t.Fatal("local keyless endpoint should be resolvable")
	}
	cfg2 := LLMConfig{BaseURL: "https://api.openrouter.ai/v1", Model: "x"}
	if cfg2.Resolvable() {
		t.Fatal("remote endpoint without key should not be resolvable")
	}
}

func TestNewLLM_ErrorsWhenUnresolvable(t *testing.T) {
	if _, err := NewLLM(LLMConfig{}); err == nil {
		t.Fatal("want error for empty config")
	}
	if _, err := NewLLM(LLMConfig{BaseURL: "https://x/v1", Model: "m"}); err == nil {
		t.Fatal("want error for missing key on remote endpoint")
	}
}

func TestLLMComplete_OpenAICompat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/chat/completions") {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer sk-test" {
			t.Errorf("auth header = %q", got)
		}
		var req chatRequest
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &req); err != nil {
			t.Errorf("bad request body: %v", err)
		}
		if req.Model != "test-model" || len(req.Messages) != 1 || req.Messages[0].Content != "hello" {
			t.Errorf("unexpected request: %+v", req)
		}
		_ = json.NewEncoder(w).Encode(chatResponse{Choices: []struct {
			Message chatMessage `json:"message"`
		}{{Message: chatMessage{Role: "assistant", Content: "world"}}}})
	}))
	defer srv.Close()

	c, err := newLLMWithClient(LLMConfig{BaseURL: srv.URL, APIKey: "sk-test", Model: "test-model"}, srv.Client())
	if err != nil {
		t.Fatalf("newLLM: %v", err)
	}
	got, err := c.Complete(context.Background(), "hello")
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if got != "world" {
		t.Fatalf("got %q, want %q", got, "world")
	}
}
