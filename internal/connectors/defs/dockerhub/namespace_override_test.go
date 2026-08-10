package dockerhub_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"polymetrics.ai/internal/connectors"
	connectorDefs "polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

func TestDockerhubNamespaceOverrideDrivesAllStreamAndCheckPaths(t *testing.T) {
	bundle, err := engine.Load(connectorDefs.FS, "dockerhub")
	if err != nil {
		t.Fatalf("load Docker Hub bundle: %v", err)
	}

	var (
		mu       sync.Mutex
		gotPaths []string
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		gotPaths = append(gotPaths, r.URL.EscapedPath())
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[]}`))
	}))
	t.Cleanup(server.Close)

	config := connectors.RuntimeConfig{Config: map[string]string{
		"base_url":        server.URL,
		"docker_username": "auth-identity",
		"namespace":       "target-namespace",
		"repository":      "fixture-repository",
		"tag":             "fixture-tag",
	}}

	t.Run("missing namespace fails before HTTP", func(t *testing.T) {
		mu.Lock()
		gotPaths = nil
		mu.Unlock()

		missingNamespace := connectors.RuntimeConfig{Config: map[string]string{
			"base_url":        server.URL,
			"docker_username": "auth-identity",
		}}
		err := engine.Read(context.Background(), bundle, connectors.ReadRequest{
			Stream: "repositories",
			Config: missingNamespace,
		}, nil, func(connectors.Record) error { return nil })
		if err == nil {
			t.Fatal("read without namespace: want local unresolved-namespace error")
		}
		if !strings.Contains(err.Error(), "namespace") {
			t.Fatalf("read without namespace error = %q, want namespace failure", err)
		}
		assertDockerhubNoRequest(t, &mu, gotPaths)
	})

	readCases := []struct {
		stream string
		path   string
	}{
		{stream: "repositories", path: "/namespaces/target-namespace/repositories"},
		{stream: "tags", path: "/namespaces/target-namespace/repositories/fixture-repository/tags"},
		{stream: "repository_detail", path: "/namespaces/target-namespace/repositories/fixture-repository"},
		{stream: "tag_detail", path: "/namespaces/target-namespace/repositories/fixture-repository/tags/fixture-tag"},
	}
	for _, tc := range readCases {
		t.Run(tc.stream, func(t *testing.T) {
			mu.Lock()
			gotPaths = nil
			mu.Unlock()

			err := engine.Read(context.Background(), bundle, connectors.ReadRequest{
				Stream: tc.stream,
				Config: config,
			}, nil, func(connectors.Record) error { return nil })
			if err != nil {
				t.Fatalf("read %s: %v", tc.stream, err)
			}

			assertDockerhubRequestPath(t, &mu, gotPaths, tc.path)
		})
	}

	t.Run("check", func(t *testing.T) {
		mu.Lock()
		gotPaths = nil
		mu.Unlock()

		if err := engine.Check(context.Background(), bundle, config, nil); err != nil {
			t.Fatalf("check Docker Hub bundle: %v", err)
		}

		assertDockerhubRequestPath(t, &mu, gotPaths, "/namespaces/target-namespace/repositories")
	})
}

func assertDockerhubRequestPath(t *testing.T, mu *sync.Mutex, gotPaths []string, want string) {
	t.Helper()
	mu.Lock()
	defer mu.Unlock()
	if len(gotPaths) != 1 {
		t.Fatalf("request paths = %v, want exactly one request to %q", gotPaths, want)
	}
	if gotPaths[0] != want {
		t.Errorf("request path = %q, want namespace override path %q", gotPaths[0], want)
	}
}

func assertDockerhubNoRequest(t *testing.T, mu *sync.Mutex, gotPaths []string) {
	t.Helper()
	mu.Lock()
	defer mu.Unlock()
	if len(gotPaths) != 0 {
		t.Fatalf("request paths = %v, want no request when namespace is absent", gotPaths)
	}
}
