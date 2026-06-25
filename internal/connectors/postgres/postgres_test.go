package postgres_test

import (
	"context"
	"errors"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/postgres"
)

// fixtureConfig is a minimal valid fixture-mode config. Fixture mode short
// circuits all network access so the unit suite needs no live database.
func fixtureConfig() connectors.RuntimeConfig {
	return connectors.RuntimeConfig{
		Config: map[string]string{
			"mode":     "fixture",
			"host":     "db.internal",
			"database": "analytics",
			"username": "reader",
			"sslmode":  "require",
		},
		Secrets: map[string]string{"password": "s3cret"},
	}
}

// liveConfigMissingPassword is a non-fixture config with required fields set but
// no password secret; Check must reject it without dialing.
func liveConfigMissingPassword() connectors.RuntimeConfig {
	return connectors.RuntimeConfig{
		Config: map[string]string{
			"host":     "db.internal",
			"database": "analytics",
			"username": "reader",
			"sslmode":  "require",
		},
	}
}

func TestRegisteredAndMetadata(t *testing.T) {
	c := postgres.New()
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
	r := connectors.NewRegistry()
	if _, ok := r.Get("postgres"); !ok {
		t.Fatal("registry did not resolve postgres (self-registration)")
	}
}

func TestCheckFixtureModeOK(t *testing.T) {
	c := postgres.New()
	if err := c.Check(context.Background(), fixtureConfig()); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

func TestCheckConfigValidation(t *testing.T) {
	c := postgres.New()
	cases := []struct {
		name string
		cfg  connectors.RuntimeConfig
	}{
		{
			name: "missing host",
			cfg: connectors.RuntimeConfig{
				Config:  map[string]string{"database": "d", "username": "u"},
				Secrets: map[string]string{"password": "p"},
			},
		},
		{
			name: "missing database",
			cfg: connectors.RuntimeConfig{
				Config:  map[string]string{"host": "h", "username": "u"},
				Secrets: map[string]string{"password": "p"},
			},
		},
		{
			name: "missing username",
			cfg: connectors.RuntimeConfig{
				Config:  map[string]string{"host": "h", "database": "d"},
				Secrets: map[string]string{"password": "p"},
			},
		},
		{
			name: "missing password secret",
			cfg:  liveConfigMissingPassword(),
		},
		{
			name: "invalid sslmode",
			cfg: connectors.RuntimeConfig{
				Config:  map[string]string{"host": "h", "database": "d", "username": "u", "sslmode": "bananas"},
				Secrets: map[string]string{"password": "p"},
			},
		},
		{
			name: "invalid port",
			cfg: connectors.RuntimeConfig{
				Config:  map[string]string{"host": "h", "database": "d", "username": "u", "port": "not-a-number"},
				Secrets: map[string]string{"password": "p"},
			},
		},
		{
			name: "ssrf host with scheme",
			cfg: connectors.RuntimeConfig{
				Config:  map[string]string{"host": "http://evil.example.com", "database": "d", "username": "u"},
				Secrets: map[string]string{"password": "p"},
			},
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

func TestCatalogFixtureMode(t *testing.T) {
	c := postgres.New()
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
	c := postgres.New()
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

// TestReadFixtureIncrementalCursor verifies the cursor lower-bound from state is
// applied: with the cursor set past the first fixture row, fewer rows come back.
func TestReadFixtureIncrementalCursor(t *testing.T) {
	c := postgres.New()
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
	// A cursor at the highest fixture id should filter out earlier rows.
	high := countWithState(map[string]string{"cursor": "99999999"})
	if high >= full {
		t.Fatalf("incremental read returned %d rows with high cursor, want fewer than full %d", high, full)
	}
}

func TestInitialStateStatefulReader(t *testing.T) {
	c := postgres.New()
	sr, ok := c.(connectors.StatefulReader)
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

func TestReadUnknownFixtureStream(t *testing.T) {
	c := postgres.New()
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "public.does_not_exist", Config: fixtureConfig()}, func(connectors.Record) error {
		return nil
	})
	if err == nil {
		t.Fatal("Read(unknown stream) = nil, want error")
	}
}

func TestCDCUnsupportedStub(t *testing.T) {
	c := postgres.New()
	cdc, ok := c.(connectors.CDCReader)
	if !ok {
		t.Fatal("postgres connector must implement CDCReader (documented stub)")
	}
	err := cdc.ReadCDC(context.Background(), connectors.CDCReadRequest{Stream: "public.users", Config: fixtureConfig()}, func(connectors.CDCEvent) error {
		return nil
	})
	if !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("ReadCDC = %v, want ErrUnsupportedOperation", err)
	}
}

func TestWriteUnsupported(t *testing.T) {
	c := postgres.New()
	_, err := c.Write(context.Background(), connectors.WriteRequest{Stream: "public.users", Config: fixtureConfig()}, []connectors.Record{{"id": 1}})
	if !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write = %v, want ErrUnsupportedOperation", err)
	}
}
