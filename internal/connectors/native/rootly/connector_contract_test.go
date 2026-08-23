package rootly

import (
	"context"
	"crypto/sha256"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/credential"
)

func TestConnectorContract(t *testing.T) {
	assertConnectorContract(t, New(), "rootly")
}

func TestCheckRejectsRetainedHeaderControlCredentialBeforeProvider(t *testing.T) {
	const source = "rootly-retained-newline-canary\n\n"
	token := credential.NormalizeStdin(source)
	want := source[:len(source)-1]
	if gotLength, wantLength := len(token), len(want); gotLength != wantLength {
		t.Fatalf("normalized token length = %d, want %d", gotLength, wantLength)
	}
	if gotHash, wantHash := sha256.Sum256([]byte(token)), sha256.Sum256([]byte(want)); gotHash != wantHash {
		t.Fatalf("normalized token SHA-256 = %x, want %x", gotHash, wantHash)
	}

	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		atomic.AddInt32(&calls, 1)
	}))
	defer srv.Close()

	err := (Connector{Client: srv.Client()}).Check(context.Background(), connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": token},
	})
	var invalid *credential.InvalidSecretValueError
	if !errors.As(err, &invalid) {
		t.Fatalf("Check() error type = %T, want InvalidSecretValueError", err)
	}
	if got := atomic.LoadInt32(&calls); got != 0 {
		t.Fatalf("provider calls = %d, want 0", got)
	}
	if strings.Contains(err.Error(), "rootly-retained-newline-canary") {
		t.Fatal("header validation error exposed credential bytes")
	}
}

func assertConnectorContract(t *testing.T, c connectors.Connector, wantName string) {
	t.Helper()
	if c == nil {
		t.Fatal("New() = nil")
	}
	if got := c.Name(); got != wantName {
		t.Fatalf("Name() = %q, want %q", got, wantName)
	}
	meta := c.Metadata()
	if meta.Name != wantName {
		t.Fatalf("Metadata().Name = %q, want %q", meta.Name, wantName)
	}
	caps := meta.Capabilities
	if !caps.Check || !caps.Catalog || !caps.Read {
		t.Fatalf("capabilities = %+v, want Check, Catalog, and Read", caps)
	}
	if caps.Write {
		t.Fatalf("%s is read-only; Write capability must be false", wantName)
	}
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture): %v", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != wantName {
		t.Fatalf("Catalog().Connector = %q, want %q", cat.Connector, wantName)
	}
	if len(cat.Streams) == 0 {
		t.Fatal("Catalog returned zero streams")
	}
}
