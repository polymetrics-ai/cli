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
	native "polymetrics.ai/internal/connectors/native/postgres"
)

// fixtureConfig is a minimal valid fixture-mode config (mirrors legacy
// postgres_test.go's fixtureConfig): fixture mode short-circuits all network
// access so this suite needs no live database.
func fixtureConfig() connectors.RuntimeConfig {
	return connectors.RuntimeConfig{
		Config: map[string]string{
			"mode":     "fixture",
			"host":     "db.internal",
			"database": "analytics",
			"username": "reader",
			"sslmode":  "require",
		},
	}
}

func TestNameAndMetadata(t *testing.T) {
	c := native.New()
	if c.Name() != "postgres" {
		t.Fatalf("Name() = %q, want postgres", c.Name())
	}
	caps := c.Metadata().Capabilities
	if !caps.Check || !caps.Catalog || !caps.Read {
		t.Fatalf("capabilities = %+v, want Check && Catalog && Read", caps)
	}
	if caps.Write {
		t.Fatalf("postgres source connector must be read-only, got Write=true")
	}
	if connectors.MetadataOf(c).Capabilities.CDC {
		t.Fatal("PostgreSQL CDC must remain undiscoverable until streamed staging is available")
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
	if manifest.Metadata.Capabilities.CDC {
		t.Fatal("PostgreSQL manifest must keep CDC fail-closed")
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

func TestCheckFixtureModeOK(t *testing.T) {
	c := native.New()
	if err := c.Check(context.Background(), fixtureConfig()); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

func TestCheckRejectsCtxCancelled(t *testing.T) {
	c := native.New()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := c.Check(ctx, fixtureConfig()); err == nil {
		t.Fatal("Check(cancelled ctx) = nil, want error")
	}
}

func TestCatalogFixtureMode(t *testing.T) {
	c := native.New()
	cat, err := c.Catalog(context.Background(), fixtureConfig())
	if err != nil {
		t.Fatalf("Catalog(fixture) = %v", err)
	}
	if cat.Connector != "postgres" {
		t.Fatalf("catalog connector = %q, want postgres", cat.Connector)
	}
	if len(cat.Streams) == 0 {
		t.Fatal("catalog returned no streams")
	}
	s := cat.Streams[0]
	if len(s.PrimaryKey) == 0 {
		t.Fatalf("fixture stream %q missing primary key", s.Name)
	}
	if len(s.Fields) == 0 {
		t.Fatalf("fixture stream %q missing fields", s.Name)
	}
	if len(s.CursorFields) == 0 {
		t.Fatalf("fixture stream %q missing cursor fields", s.Name)
	}
}

func TestReadFixtureEmitsRows(t *testing.T) {
	c := native.New()
	cat, err := c.Catalog(context.Background(), fixtureConfig())
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	stream := cat.Streams[0].Name

	var got []connectors.Record
	err = c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: fixtureConfig()}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read(fixture): %v", err)
	}
	if len(got) < 2 {
		t.Fatalf("read emitted %d rows, want >= 2", len(got))
	}
	pk := cat.Streams[0].PrimaryKey[0]
	cursor := cat.Streams[0].CursorFields[0]
	for _, rec := range got {
		if rec[pk] == nil {
			t.Fatalf("record missing primary key %q: %+v", pk, rec)
		}
		if rec[cursor] == nil {
			t.Fatalf("record missing cursor field %q: %+v", cursor, rec)
		}
	}
}

func TestReadFixtureIncrementalCursor(t *testing.T) {
	c := native.New()
	cat, _ := c.Catalog(context.Background(), fixtureConfig())
	stream := cat.Streams[0].Name

	countWithState := func(state map[string]string) int {
		var n int
		_ = c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: fixtureConfig(), State: state}, func(connectors.Record) error {
			n++
			return nil
		})
		return n
	}

	full := countWithState(nil)
	high := countWithState(map[string]string{"cursor": "99999999"})
	if high >= full {
		t.Fatalf("incremental read returned %d rows with high cursor, want fewer than full %d", high, full)
	}
}

func TestReadUnknownFixtureStream(t *testing.T) {
	c := native.New()
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "public.does_not_exist", Config: fixtureConfig()}, func(connectors.Record) error {
		return nil
	})
	if err == nil {
		t.Fatal("Read(unknown stream) = nil, want error")
	}
}

func TestReadRequiresStream(t *testing.T) {
	c := native.New()
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "", Config: fixtureConfig()}, func(connectors.Record) error {
		return nil
	})
	if err == nil {
		t.Fatal("Read(no stream) = nil, want error")
	}
}

func TestInitialStateStatefulReader(t *testing.T) {
	c := native.New()
	sr, ok := any(c).(connectors.StatefulReader)
	if !ok {
		t.Fatal("postgres connector must implement StatefulReader")
	}
	state, err := sr.InitialState(context.Background(), "public.users", fixtureConfig())
	if err != nil {
		t.Fatalf("InitialState: %v", err)
	}
	if state == nil {
		t.Fatal("InitialState returned nil state map")
	}
}

func TestCDCIsFailClosedUntilStreamedStagingExists(t *testing.T) {
	c := native.New()
	cdc, ok := any(c).(connectors.ChangefeedExecutor)
	if !ok {
		t.Fatal("postgres connector must implement a logical-replication ChangefeedExecutor")
	}
	err := cdc.ReadCDC(context.Background(), connectors.CDCReadRequest{Stream: "public.users", Config: fixtureConfig()}, func(connectors.CDCEvent) error {
		return nil
	})
	if !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("ReadCDC = %v, want fail-closed unsupported result", err)
	}
}

func TestWriteUnsupported(t *testing.T) {
	c := native.New()
	_, err := c.Write(context.Background(), connectors.WriteRequest{Stream: "public.users", Config: fixtureConfig()}, []connectors.Record{{"id": 1}})
	if !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write = %v, want ErrUnsupportedOperation", err)
	}
}

// --- config validation table (component-level, mirrors legacy behavior) ---

func TestCheckConfigValidationTable(t *testing.T) {
	c := native.New()
	cases := []struct {
		name string
		cfg  connectors.RuntimeConfig
	}{
		{
			name: "missing host",
			cfg:  connectors.RuntimeConfig{Config: map[string]string{"database": "d", "username": "u"}},
		},
		{
			name: "missing database",
			cfg:  connectors.RuntimeConfig{Config: map[string]string{"host": "h", "username": "u"}},
		},
		{
			name: "missing username",
			cfg:  connectors.RuntimeConfig{Config: map[string]string{"host": "h", "database": "d"}},
		},
		{
			name: "missing password secret",
			cfg: connectors.RuntimeConfig{
				Config: map[string]string{"host": "h", "database": "d", "username": "u", "sslmode": "require"},
			},
		},
		{
			name: "invalid sslmode",
			cfg:  connectors.RuntimeConfig{Config: map[string]string{"host": "h", "database": "d", "username": "u", "sslmode": "bananas"}},
		},
		{
			name: "invalid port (non-numeric)",
			cfg:  connectors.RuntimeConfig{Config: map[string]string{"host": "h", "database": "d", "username": "u", "port": "not-a-number"}},
		},
		{
			name: "invalid port (out of range)",
			cfg:  connectors.RuntimeConfig{Config: map[string]string{"host": "h", "database": "d", "username": "u", "port": "70000"}},
		},
		{
			name: "host with scheme (SSRF guard)",
			cfg:  connectors.RuntimeConfig{Config: map[string]string{"host": "http://evil.example.com", "database": "d", "username": "u"}},
		},
		{
			name: "host with bracketed non-IPv6",
			cfg:  connectors.RuntimeConfig{Config: map[string]string{"host": "[not-an-ip]", "database": "d", "username": "u"}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := c.Check(context.Background(), tc.cfg); err == nil {
				t.Fatalf("Check(%s) = nil, want validation error", tc.name)
			}
		})
	}
}
