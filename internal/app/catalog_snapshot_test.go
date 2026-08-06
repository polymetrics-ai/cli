package app

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
)

func TestShowCatalogUsesOneAccountFileForTwoConnections(t *testing.T) {
	instance, runtime := catalogTestApp(t)
	ref, err := instance.catalogReference("hubspot", runtime)
	if err != nil {
		t.Fatalf("catalogReference: %v", err)
	}
	stored := connectors.Catalog{Connector: "hubspot", Streams: []connectors.Stream{{Name: "2-1"}}, Discovery: &connectors.DiscoveryStatus{Complete: true, ExpiresAt: time.Now().UTC().Add(-time.Minute)}}
	if err := instance.catalogs.write(ref, stored, time.Now().UTC()); err != nil {
		t.Fatalf("write: %v", err)
	}
	instance.state.Catalogs = []catalogReference{ref}

	first, err := instance.ShowCatalog(context.Background(), "account-a")
	if err != nil {
		t.Fatalf("ShowCatalog account-a: %v", err)
	}
	second, err := instance.ShowCatalog(context.Background(), "account-b")
	if err != nil {
		t.Fatalf("ShowCatalog account-b: %v", err)
	}
	if len(instance.state.Catalogs) != 1 || first.Connection != "account-a" || second.Connection != "account-b" {
		t.Fatalf("account catalog references = %#v, snapshots = %#v / %#v", instance.state.Catalogs, first, second)
	}
	if first.Catalog.Discovery == nil || !first.Catalog.Discovery.Stale || !first.Catalog.Discovery.Cached {
		t.Fatalf("shown discovery status = %#v, want explicit stale cached status", first.Catalog.Discovery)
	}
}

func TestCatalogFileContainsSchemaOnlyNotAccountKey(t *testing.T) {
	instance, runtime := catalogTestApp(t)
	ref, err := instance.catalogReference("hubspot", runtime)
	if err != nil {
		t.Fatalf("catalogReference: %v", err)
	}
	catalog := connectors.Catalog{Connector: "hubspot", Streams: []connectors.Stream{{Name: "2-1", Schema: []byte(`{"type":"object","properties":{"field_a":{"type":"string"}}}`)}}}
	if err := instance.catalogs.write(ref, catalog, time.Now().UTC()); err != nil {
		t.Fatalf("write: %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(instance.projectDir, ref.File))
	if err != nil {
		t.Fatalf("read catalog file: %v", err)
	}
	for _, forbidden := range [][]byte{[]byte(ref.AccountKey), []byte("account-a"), []byte("credential"), []byte("config"), []byte("secrets")} {
		if bytes.Contains(raw, forbidden) {
			t.Fatalf("catalog file contains forbidden non-schema material")
		}
	}
}

func TestPersistAccountCatalogDoesNotWritePointerBeforeFileSync(t *testing.T) {
	instance, runtime := catalogTestApp(t)
	ref, err := instance.catalogReference("hubspot", runtime)
	if err != nil {
		t.Fatalf("catalogReference: %v", err)
	}
	instance.catalogs.syncDirectory = func(string) error { return errors.New("sync failed") }
	err = instance.persistAccountCatalog(ref, connectors.Catalog{Connector: "hubspot"}, time.Now().UTC())
	if err == nil {
		t.Fatal("persistAccountCatalog succeeded despite failed file durability")
	}
	if len(instance.state.Catalogs) != 0 {
		t.Fatalf("state referenced a catalog after file sync failed: %#v", instance.state.Catalogs)
	}
}

func TestCatalogStorageSyncsCreatedDirectoryChain(t *testing.T) {
	projectDir := filepath.Join(t.TempDir(), "project", ".polymetrics")
	storage := newCatalogStorage(projectDir)
	var synced []string
	storage.syncDirectory = func(path string) error {
		synced = append(synced, path)
		return nil
	}
	reference := catalogReference{Connector: "hubspot", AccountKey: connectors.AuthCohortKey("opaque-key-1")}
	file, err := catalogRelativePath(reference.Connector, reference.AccountKey)
	if err != nil {
		t.Fatalf("catalogRelativePath: %v", err)
	}
	reference.File = file
	if err := storage.write(reference, connectors.Catalog{Connector: "hubspot"}, time.Now().UTC()); err != nil {
		t.Fatalf("write: %v", err)
	}
	for _, required := range []string{filepath.Join(projectDir, "catalogs", "hubspot"), filepath.Join(projectDir, "catalogs"), projectDir} {
		if !containsPath(synced, required) {
			t.Fatalf("sync calls = %v, missing created directory %q", synced, required)
		}
	}
}

func catalogTestApp(t *testing.T) (*App, connectors.RuntimeConfig) {
	t.Helper()
	projectDir := filepath.Join(t.TempDir(), ".polymetrics")
	identity, err := connectors.NewCoordinationIdentity([]byte("catalog-test-project-salt"), connectors.CredentialBinding{
		BindingID: "binding-catalog", ProviderFamily: "hubspot", AuthProfile: "private_app",
	})
	if err != nil {
		t.Fatalf("NewCoordinationIdentity: %v", err)
	}
	credential := CredentialMeta{ID: "credential-id", Name: "credential-alias", Connector: "hubspot", ProviderFamily: "hubspot", AuthProfile: "private_app"}
	instance := &App{
		projectDir: projectDir,
		catalogs:   newCatalogStorage(projectDir),
		state: state{
			CoordinationSalt: "catalog-test-project-salt",
			Credentials:      []CredentialMeta{credential},
			CredentialBindings: map[string]credentialBindingState{
				credential.ID: {BindingID: "binding-catalog", ProviderFamilyDeclared: true, AuthProfileDeclared: true, DeclarationProvenanceRecorded: true},
			},
			Connections: []Connection{
				{Name: "account-a", Source: EndpointConfig{Connector: "hubspot", Credential: credential.Name}},
				{Name: "account-b", Source: EndpointConfig{Connector: "hubspot", Credential: credential.Name}},
			},
		},
	}
	return instance, connectors.RuntimeConfig{CoordinationIdentity: identity}
}

func containsPath(paths []string, want string) bool {
	for _, path := range paths {
		if path == want {
			return true
		}
	}
	return false
}
