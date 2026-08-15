package postgres_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	native "polymetrics.ai/internal/connectors/native/postgres"
)

func TestNameAndMetadata(t *testing.T) {
	c := native.New()
	if c.Name() != "postgres" {
		t.Fatalf("Name() = %q, want postgres", c.Name())
	}
	caps := c.Metadata().Capabilities
	if !caps.Check || !caps.Catalog || !caps.Read {
		t.Fatalf("capabilities = %+v, want Check && Catalog && Read", caps)
	}
	if !caps.Write || !caps.CDC || caps.Query {
		t.Fatalf("PostgreSQL capabilities = %+v, want write=true cdc=true query=false", caps)
	}
	if !connectors.MetadataOf(c).Capabilities.CDC {
		t.Fatal("PostgreSQL CDC must be discoverable with the matching pgoutput v2 executor")
	}
	definition, ok := connectors.DefinitionOf(c)
	if !ok || definition.Changefeed == nil || !definition.Capabilities.CDC {
		t.Fatalf("PostgreSQL definition = %#v, want an executable matching changefeed", definition)
	}
}

func TestPostgresDeclaresProviderHTTPRateLimitsNotApplicable(t *testing.T) {
	bundle, err := engine.Load(defs.FS, "postgres")
	if err != nil {
		t.Fatal(err)
	}
	if bundle.RateLimits == nil || bundle.RateLimits.State != connsdk.RateLimitStateNotApplicable || bundle.RateLimits.Reason != "PostgreSQL uses its native wire protocol and makes no provider HTTP API requests." {
		t.Fatalf("PostgreSQL rate limits = %#v, want explicit no-provider-HTTP not_applicable declaration", bundle.RateLimits)
	}
}

func TestManifestProjectsBundleCredentials(t *testing.T) {
	manifest := connectors.ManifestOf(native.New())
	configFields := make(map[string]connectors.ConfigField, len(manifest.ConfigFields))
	for _, field := range manifest.ConfigFields {
		configFields[field.Name] = field
	}
	wantConfigFields := []string{
		"cdc_publication",
		"cursor_field",
		"database",
		"host",
		"mode",
		"port",
		"read_limit",
		"schema",
		"sslmode",
		"sslrootcert",
		"sslservername",
		"username",
	}
	if len(configFields) != len(wantConfigFields) {
		t.Fatalf("manifest config fields = %#v, want %#v", configFields, wantConfigFields)
	}
	for _, name := range wantConfigFields {
		if _, ok := configFields[name]; !ok {
			t.Fatalf("manifest config fields = %#v, missing %q", configFields, name)
		}
	}
	for _, name := range []string{"database", "host", "username"} {
		if !configFields[name].Required {
			t.Fatalf("manifest config field %q = %#v, want required", name, configFields[name])
		}
	}
	if len(manifest.SecretFields) != 1 || manifest.SecretFields[0].Name != "password" {
		t.Fatalf("manifest secret fields = %#v, want password", manifest.SecretFields)
	}
	password := manifest.SecretFields[0]
	if password.Required || password.RequiredWhen != "mode is not fixture" {
		t.Fatalf("manifest password requirement = %#v, want conditional fixture exemption", password)
	}
	if len(manifest.AuthModes) != 1 || manifest.AuthModes[0].Name != "password" {
		t.Fatalf("manifest auth modes = %#v, want password authentication", manifest.AuthModes)
	}
	if !manifest.Metadata.Capabilities.CDC {
		t.Fatal("PostgreSQL manifest must advertise the proven pgoutput v2 capability")
	}
	if !manifest.Metadata.Capabilities.Write || manifest.Metadata.Capabilities.Query {
		t.Fatalf("PostgreSQL manifest capabilities = %+v, want write=true query=false", manifest.Metadata.Capabilities)
	}
}

func TestGeneratedDocsDescribeAuthenticationRequirements(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	docsDir := filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..", "docs", "connectors", "postgres")
	for _, name := range []string{"MANUAL.md", "SKILL.md"} {
		raw, err := os.ReadFile(filepath.Join(docsDir, name))
		if err != nil {
			t.Fatalf("Read %s: %v", name, err)
		}
		for _, want := range []string{
			"database (required)",
			"host (required)",
			"username (required)",
			"password (secret) (required when mode is not fixture)",
			"password: Live connections require password authentication; peer/socket and client-certificate modes, including ambient certificates, are unsupported.",
		} {
			if !strings.Contains(string(raw), want) {
				t.Fatalf("%s missing %q:\n%s", name, want, raw)
			}
		}
	}
}

// TestCertificationArchitectureUsesExecutablePostgresResumeBindings guards the
// #4015 architecture's keyset predicate from drifting back to symbolic values.
// It documents the target contract; #3855 separately identifies the current
// scalar reader predicate as a legacy implementation gap.
func TestCertificationArchitectureUsesExecutablePostgresResumeBindings(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	docPath := filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..", "docs", "architecture", "github-postgres-warehouse-certification.md")
	raw, err := os.ReadFile(docPath)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", docPath, err)
	}
	doc := string(raw)
	const bindingOrder = "bind the prior cursor value as pgx argument 1 (`$1`) and the stable primary-key tie breaker as argument 2 (`$2`)"
	if !strings.Contains(doc, bindingOrder) {
		t.Fatalf("certification architecture missing PostgreSQL pgx binding order: %q", bindingOrder)
	}
	const want = `WHERE cursor > $1
   OR (cursor = $1 AND primary_key > $2)
ORDER BY cursor, primary_key`
	if !strings.Contains(doc, want) {
		t.Fatalf("certification architecture missing executable PostgreSQL resume predicate:\n%s", want)
	}
	for _, forbidden := range []string{"WHERE cursor > $cursor", "primary_key > $primary_key"} {
		if strings.Contains(doc, forbidden) {
			t.Fatalf("certification architecture contains symbolic PostgreSQL placeholder %q", forbidden)
		}
	}
}

// TestNoInitRegistration is the required grep-guard (T-17): the native
// package must NOT call RegisterFactory from anywhere in
// its own source. The registration flip (wiring native/postgres into the
// production registry) is a wave6 change; wave0 only builds and tests the
// package. This is a structural guard, not a behavioral one, so it inspects
// the actual .go source files rather than runtime registry state (a
// same-process runtime check could pass even if some other test in the
// binary happened to import a package that registers "postgres" under a
// different mechanism).
func TestNoInitRegistration(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed; cannot locate package directory")
	}
	dir := filepath.Dir(thisFile)

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir(%s): %v", dir, err)
	}

	found := false
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		if strings.HasSuffix(e.Name(), "_test.go") {
			// The grep-guard covers the package's own production source, not
			// its tests (this very test file legitimately mentions the
			// forbidden identifiers in prose/identifiers above).
			continue
		}
		found = true
		raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatalf("ReadFile(%s): %v", e.Name(), err)
		}
		src := string(raw)
		if strings.Contains(src, "RegisterFactory(") {
			t.Fatalf("%s calls RegisterFactory — native/postgres must NOT self-register in wave0 (registration flip is wave6)", e.Name())
		}
		if strings.Contains(src, "func init()") {
			t.Fatalf("%s declares an init() function — native/postgres must perform no registration side effects in wave0", e.Name())
		}
	}
	if !found {
		t.Fatal("no non-test .go source files found in native/postgres; grep-guard did not actually scan anything")
	}
}

// TestConnectorSatisfiesCoreInterfaces compile/runtime-asserts the shape
// required by API-CONTRACT.md / design §B.7 Tier-3. PostgreSQL CDC is admitted
// only when the native reader also supplies the exact executor descriptor.
func TestConnectorSatisfiesCoreInterfaces(t *testing.T) {
	c := native.New()
	var _ connectors.Connector = c
	if _, ok := any(c).(connectors.CDCReader); !ok {
		t.Fatal("native postgres connector must implement connectors.CDCReader")
	}
	if _, ok := any(c).(connectors.ChangefeedExecutor); !ok {
		t.Fatal("native postgres connector must implement ChangefeedExecutor for logical replication")
	}
	if _, ok := any(c).(connectors.StatefulReader); !ok {
		t.Fatal("native postgres connector must implement connectors.StatefulReader")
	}
	if _, ok := any(c).(connectors.DefinitionProvider); !ok {
		t.Fatal("native postgres connector must implement connectors.DefinitionProvider (engine.Base)")
	}
}

func TestWriteUnsupported(t *testing.T) {
	c := native.New()
	_, err := c.Write(context.Background(), connectors.WriteRequest{Stream: "public.users", Config: fixtureConfig()}, []connectors.Record{{"id": 1}})
	if !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write = %v, want ErrUnsupportedOperation", err)
	}
}
