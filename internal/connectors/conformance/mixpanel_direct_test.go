package conformance

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

func TestMixpanelOperationDirectReadsReplay(t *testing.T) {
	b := loadTestBundle(t, "../defs", "mixpanel")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "ok",
			"path":   r.URL.Path,
		})
	}))
	defer server.Close()

	cfg := runtimeConfigForEngine(b)
	cfg.Config["base_url"] = server.URL

	implementedDirect := map[string]bool{}
	if b.CLISurface != nil {
		for _, cmd := range b.CLISurface.Commands {
			if cmd.Intent == "direct_read" && cmd.Availability == "implemented" && cmd.Operation != "" {
				implementedDirect[cmd.Operation] = true
			}
		}
	}
	if len(implementedDirect) != 15 {
		t.Fatalf("implemented direct read commands = %d, want 15", len(implementedDirect))
	}

	for _, op := range b.Operations {
		op := op
		if !implementedDirect[op.ID] {
			continue
		}
		t.Run(op.ID, func(t *testing.T) {
			result, err := engine.OperationDirectRead(context.Background(), b, connectors.OperationDirectReadRequest{
				Operation:    op.ID,
				Config:       cfg,
				MaxBytes:     1 << 20,
				OutputPolicy: "json_redacted",
				RedactFields: []string{"token", "secret", "password", "api_secret", "project_token", "authorization"},
			}, engine.HooksFor(b.Name))
			if err != nil {
				t.Fatalf("OperationDirectRead(%s): %v", op.ID, err)
			}
			if result.Status != http.StatusOK {
				t.Fatalf("status = %d, want 200", result.Status)
			}
		})
	}
}
