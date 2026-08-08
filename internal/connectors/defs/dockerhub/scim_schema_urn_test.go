package dockerhub_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	connectorDefs "polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

const canonicalSCIMSchemaURN = "urn:ietf:params:scim:schemas:core:2.0:User"

func TestDockerhubSCIMSchemaGetAcceptsCanonicalURNPathParameter(t *testing.T) {
	bundle, err := engine.Load(connectorDefs.FS, "dockerhub")
	if err != nil {
		t.Fatalf("load Docker Hub bundle: %v", err)
	}

	var requests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		const wantPath = "/v2/scim/2.0/Schemas/urn:ietf:params:scim:schemas:core:2.0:User"
		if got := r.URL.Path; got != wantPath {
			t.Errorf("request path = %q, want %q", got, wantPath)
		}
		if got := r.URL.RawQuery; got != "" {
			t.Errorf("request query = %q, want empty", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"urn:ietf:params:scim:schemas:core:2.0:User"}`))
	}))
	t.Cleanup(server.Close)

	// The real bundle uses a separate SCIM bearer hook. This isolated path-resolution
	// test removes only auth so it can prove the operation reaches the local server
	// without a credential or Docker Hub network request.
	bundle.HTTP.URL = server.URL
	bundle.HTTP.Auth = nil

	_, err = engine.OperationDirectRead(context.Background(), bundle, connectors.OperationDirectReadRequest{
		Operation: "dockerhub.get_scim_schema",
		Config:    connectors.RuntimeConfig{Config: map[string]string{"base_url": server.URL}},
		PathParams: map[string]string{
			"id": canonicalSCIMSchemaURN,
		},
	}, nil)
	if err != nil {
		t.Fatalf("read canonical Docker Hub SCIM schema URN: %v", err)
	}
	if requests != 1 {
		t.Fatalf("requests = %d, want 1", requests)
	}
}
