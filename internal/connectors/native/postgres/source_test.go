package postgres_test

import (
	"context"
	"testing"

	"polymetrics.ai/internal/connectors"
	native "polymetrics.ai/internal/connectors/native/postgres"
)

// fixtureConfig is a minimal valid fixture-mode config. Fixture mode
// short-circuits all network access, so this suite needs no live database.
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
