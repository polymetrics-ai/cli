package trello

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

func init() {
	engine.RegisterHooks("trello", func() engine.Hooks { return New() })
}

type Hooks struct{}

func New() *Hooks { return &Hooks{} }

func (h *Hooks) ConnectorName() string { return "trello" }

var (
	_ engine.Hooks    = (*Hooks)(nil)
	_ engine.AuthHook = (*Hooks)(nil)
)

func (h *Hooks) Authenticator(ctx context.Context, cfg connectors.RuntimeConfig, _ engine.AuthSpec) (connsdk.Authenticator, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	key := strings.TrimSpace(cfg.Secrets["key"])
	if key == "" {
		return nil, errors.New("trello auth requires secret key")
	}
	token := strings.TrimSpace(cfg.Secrets["token"])
	if token == "" {
		return nil, errors.New("trello auth requires secret token")
	}
	return connsdk.AuthFunc(func(ctx context.Context, req *http.Request) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := req.URL.Query()
		query.Set("key", key)
		query.Set("token", token)
		req.URL.RawQuery = query.Encode()
		return nil
	}), nil
}
