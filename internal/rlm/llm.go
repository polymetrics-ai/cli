package rlm

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"polymetrics.ai/internal/connectors/connsdk"
)

// LLMClient is a minimal provider-agnostic chat completion interface. It is
// satisfied by the OpenAI-compatible client below and mocked in tests. The RLM
// router (Tier-2 classification) consumes it; the in-container PI mono agent uses
// pi-ai natively and does not use this client.
type LLMClient interface {
	Complete(ctx context.Context, prompt string) (string, error)
}

// LLMConfig configures an OpenAI-compatible chat client. Provider sets sensible
// defaults (base URL); explicit BaseURL/Model/APIKey override them.
type LLMConfig struct {
	Provider string // "openrouter" (default), "openai", "ollama", or "" for custom
	BaseURL  string // OpenAI-compatible base, e.g. https://openrouter.ai/api/v1
	APIKey   string // bearer token; may be empty for keyless local servers
	Model    string // model id, e.g. "anthropic/claude-fable-5"
}

const openrouterBaseURL = "https://openrouter.ai/api/v1"

// LLMConfigFromEnv builds an LLMConfig from environment variables, defaulting to
// OpenRouter. Precedence for the key: OPENROUTER_API_KEY (when provider is
// openrouter/unset) then PM_LLM_API_KEY. Base/model fall back to PM_LLM_* and the
// provider default.
func LLMConfigFromEnv(getenv func(string) string) LLMConfig {
	if getenv == nil {
		getenv = func(string) string { return "" }
	}
	return LLMConfigFromSettings(
		getenv("PM_LLM_PROVIDER"),
		getenv("PM_LLM_BASE_URL"),
		getenv("PM_LLM_MODEL"),
		getenv,
	)
}

// LLMConfigFromSettings combines non-secret typed settings with env-only API-key
// intake. API keys remain environment-only and are never loaded from config files.
func LLMConfigFromSettings(provider, baseURL, model string, getenv func(string) string) LLMConfig {
	if getenv == nil {
		getenv = func(string) string { return "" }
	}
	provider = strings.TrimSpace(provider)
	if provider == "" {
		provider = "openrouter"
	}
	cfg := LLMConfig{
		Provider: provider,
		BaseURL:  strings.TrimSpace(baseURL),
		APIKey:   strings.TrimSpace(getenv("PM_LLM_API_KEY")),
		Model:    strings.TrimSpace(model),
	}
	if provider == "openrouter" {
		if cfg.BaseURL == "" {
			cfg.BaseURL = openrouterBaseURL
		}
		if cfg.APIKey == "" {
			cfg.APIKey = strings.TrimSpace(getenv("OPENROUTER_API_KEY"))
		}
	}
	return cfg
}

// Resolvable reports whether the config has enough to make a call (a base URL,
// and an API key unless the endpoint is an obviously-local keyless server).
func (c LLMConfig) Resolvable() bool {
	if c.BaseURL == "" {
		return false
	}
	if c.APIKey != "" {
		return true
	}
	return isLocalEndpoint(c.BaseURL)
}

func isLocalEndpoint(base string) bool {
	b := strings.ToLower(base)
	return strings.Contains(b, "localhost") || strings.Contains(b, "127.0.0.1") || strings.Contains(b, "://0.0.0.0")
}

// httpLLM is an OpenAI-compatible chat client built on connsdk.Requester so it
// reuses the repo's retry/backoff/rate-limit transport.
type httpLLM struct {
	req   *connsdk.Requester
	model string
}

// NewLLM constructs an LLMClient from cfg. It errors if the config is not
// resolvable so callers get a clear message instead of an opaque HTTP failure.
func NewLLM(cfg LLMConfig) (LLMClient, error) {
	return newLLMWithClient(cfg, nil)
}

func newLLMWithClient(cfg LLMConfig, httpClient *http.Client) (LLMClient, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("rlm: LLM base URL is not set (set PM_LLM_BASE_URL or use --provider openrouter with OPENROUTER_API_KEY)")
	}
	if cfg.Model == "" {
		return nil, fmt.Errorf("rlm: LLM model is not set (set PM_LLM_MODEL or --model)")
	}
	if cfg.APIKey == "" && !isLocalEndpoint(cfg.BaseURL) {
		return nil, fmt.Errorf("rlm: LLM API key is not set (set OPENROUTER_API_KEY / PM_LLM_API_KEY)")
	}
	var auth connsdk.Authenticator
	if cfg.APIKey != "" {
		auth = connsdk.Bearer(cfg.APIKey)
	}
	return &httpLLM{
		req: &connsdk.Requester{
			Client:    httpClient,
			BaseURL:   strings.TrimRight(cfg.BaseURL, "/"),
			Auth:      auth,
			UserAgent: "polymetrics-go-cli",
		},
		model: cfg.Model,
	}, nil
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

// Complete sends a single user prompt to the chat/completions endpoint and
// returns the assistant message content.
func (l *httpLLM) Complete(ctx context.Context, prompt string) (string, error) {
	body := chatRequest{
		Model:    l.model,
		Messages: []chatMessage{{Role: "user", Content: prompt}},
	}
	var out chatResponse
	if err := l.req.DoJSON(ctx, http.MethodPost, "chat/completions", nil, body, &out); err != nil {
		return "", fmt.Errorf("rlm: LLM request: %w", err)
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("rlm: LLM returned no choices")
	}
	return out.Choices[0].Message.Content, nil
}
