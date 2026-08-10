package dockerhub_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"

	"polymetrics.ai/internal/connectors"
	connectorDefs "polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	dockerhubhooks "polymetrics.ai/internal/connectors/hooks/dockerhub"
)

var (
	dockerHubRawPathParameter    = regexp.MustCompile(`\{[A-Za-z_][A-Za-z0-9_]*\}`)
	dockerHubRecordPathParameter = regexp.MustCompile(`\{\{\s*record\.([A-Za-z_][A-Za-z0-9_]*)\s*\}\}`)
)

func TestDockerhubReverseETLWritePathsAreEngineRelativeAndInterpolated(t *testing.T) {
	bundle, err := engine.Load(connectorDefs.FS, "dockerhub")
	if err != nil {
		t.Fatalf("load Docker Hub bundle: %v", err)
	}

	for _, action := range bundle.Writes {
		t.Run(action.Name, func(t *testing.T) {
			if strings.HasPrefix(action.Path, "/v2/") {
				t.Errorf("path = %q, want engine-relative path without the base URL's /v2 prefix", action.Path)
			}
			if raw := dockerHubRawPathParameter.FindString(action.Path); raw != "" {
				t.Errorf("path = %q retains raw OpenAPI parameter %q, want a {{ record.* }} template", action.Path, raw)
			}

			pathFields := make(map[string]bool, len(action.PathFields))
			for _, field := range action.PathFields {
				pathFields[field] = true
			}
			for _, match := range dockerHubRecordPathParameter.FindAllStringSubmatch(action.Path, -1) {
				if !pathFields[match[1]] {
					t.Errorf("path template %q is missing %q from path_fields %v", match[0], match[1], action.PathFields)
				}
			}
		})
	}

	preview, err := engine.DryRunWrite(context.Background(), bundle, connectors.WriteRequest{
		Action: "create_repository",
		Config: connectors.RuntimeConfig{Config: map[string]string{
			"base_url": "https://hub.docker.com/v2",
		}},
	}, []connectors.Record{{
		"name":       "fixture-repository",
		"namespace":  "polymetrics",
		"is_private": true,
	}}, nil)
	if err != nil {
		t.Fatalf("dry-run create_repository: %v", err)
	}

	const wantRequest = "POST https://hub.docker.com/v2/namespaces/polymetrics/repositories"
	if warnings := strings.Join(preview.Warnings, "\n"); !strings.Contains(warnings, wantRequest) {
		t.Errorf("create_repository preview warnings = %q, want resolved request %q", warnings, wantRequest)
	}
}

func TestDockerhubSCIMWritesNormalizeProxyBaseAndUseSCIMBearer(t *testing.T) {
	bundle, err := engine.Load(connectorDefs.FS, "dockerhub")
	if err != nil {
		t.Fatalf("load Docker Hub bundle: %v", err)
	}

	var dataHits int32
	dataServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&dataHits, 1)
		if r.Method != http.MethodPost || r.URL.Path != "/proxy/v2/scim/2.0/Users" {
			t.Errorf("SCIM request = %s %s, want POST /proxy/v2/scim/2.0/Users", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer fixture-scim-bearer" {
			t.Errorf("SCIM Authorization = %q, want the separately configured SCIM bearer", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/scim+json" {
			t.Errorf("SCIM Content-Type = %q, want application/scim+json", got)
		}
		w.Header().Set("Content-Type", "application/scim+json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"fixture-user"}`))
	}))
	t.Cleanup(dataServer.Close)

	var authHits int32
	authServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&authHits, 1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"token":"fixture-session"}`))
	}))
	t.Cleanup(authServer.Close)

	records := []connectors.Record{{
		"schemas":  []any{"urn:ietf:params:scim:schemas:core:2.0:User"},
		"userName": "fixture-user@example.test",
	}}
	for _, tt := range []struct {
		name    string
		secrets map[string]string
	}{
		{name: "SCIM-only", secrets: map[string]string{"scim_bearer_token": "fixture-scim-bearer"}},
		{name: "dual-credential", secrets: map[string]string{"docker_pat": "fixture-pat", "scim_bearer_token": "fixture-scim-bearer"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cfg := connectors.RuntimeConfig{
				Config: map[string]string{
					"base_url":        dataServer.URL + "/proxy",
					"auth_url":        authServer.URL + "/trusted-auth",
					"docker_username": "fixture-user",
				},
				Secrets: tt.secrets,
			}
			hooks := dockerhubhooks.New().(*dockerhubhooks.Hooks)
			hooks.Client = authServer.Client()
			req := connectors.WriteRequest{Action: "create_scim_user", Config: cfg}

			preview, err := engine.DryRunWrite(context.Background(), bundle, req, records, hooks)
			if err != nil {
				t.Fatalf("preview SCIM write: %v", err)
			}
			wantPreview := "POST " + dataServer.URL + "/proxy/v2/scim/2.0/Users"
			if warnings := strings.Join(preview.Warnings, "\n"); !strings.Contains(warnings, wantPreview) {
				t.Fatalf("preview warnings = %q, want %q", warnings, wantPreview)
			}

			result, err := engine.Write(context.Background(), bundle, req, records, hooks)
			if err != nil {
				t.Fatalf("execute SCIM write: %v", err)
			}
			if result.RecordsWritten != 1 || result.RecordsFailed != 0 {
				t.Fatalf("SCIM write result = %+v, want one success", result)
			}
		})
	}
	if got := atomic.LoadInt32(&dataHits); got != 2 {
		t.Fatalf("SCIM data requests = %d, want 2", got)
	}
	if got := atomic.LoadInt32(&authHits); got != 0 {
		t.Fatalf("Docker Hub session-login requests = %d, want 0 for SCIM writes", got)
	}
}
