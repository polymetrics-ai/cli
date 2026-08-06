package mysql

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func testConfig() connectors.RuntimeConfig {
	return connectors.RuntimeConfig{Config: map[string]string{
		"host":                  "db.internal",
		"port":                  "3317",
		"database":              "analytics",
		"username":              "reader",
		"cursor_field":          "sequence",
		"page_size":             "2",
		"read_limit":            "10",
		"replication_server_id": "731001",
	}}
}

func TestNameMetadataAndExecutableChangefeed(t *testing.T) {
	c := New()
	if c.Name() != "mysql" {
		t.Fatalf("Name() = %q, want mysql", c.Name())
	}
	caps := connectors.MetadataOf(c).Capabilities
	if !caps.Check || !caps.Catalog || !caps.Read || !caps.CDC || caps.Write {
		t.Fatalf("capabilities = %+v, want check/catalog/read/cdc only", caps)
	}
	definition, ok := any(c).(connectors.DefinitionProvider)
	if !ok {
		t.Fatal("mysql connector has no bundle definition")
	}
	if !connectors.HasImplementedChangefeed(c, definition.Definition().Changefeed) {
		t.Fatal("mysql connector must expose CDC only through its matching binlog executor")
	}
}

func TestResolveConfigRejectsUnsafeInputsWithoutSecretLeakage(t *testing.T) {
	base := testConfig()
	cases := []struct {
		name   string
		mutate func(connectors.RuntimeConfig) connectors.RuntimeConfig
	}{
		{name: "missing host", mutate: func(cfg connectors.RuntimeConfig) connectors.RuntimeConfig { cfg.Config["host"] = ""; return cfg }},
		{name: "url host", mutate: func(cfg connectors.RuntimeConfig) connectors.RuntimeConfig {
			cfg.Config["host"] = "https://db.internal"
			return cfg
		}},
		{name: "unsafe database", mutate: func(cfg connectors.RuntimeConfig) connectors.RuntimeConfig {
			cfg.Config["database"] = "analytics;drop"
			return cfg
		}},
		{name: "invalid port", mutate: func(cfg connectors.RuntimeConfig) connectors.RuntimeConfig { cfg.Config["port"] = "0"; return cfg }},
		{name: "invalid user", mutate: func(cfg connectors.RuntimeConfig) connectors.RuntimeConfig {
			cfg.Config["username"] = "reader\nother"
			return cfg
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := connectors.RuntimeConfig{Config: map[string]string{}}
			for key, value := range base.Config {
				cfg.Config[key] = value
			}
			_, err := resolveConfig(tc.mutate(cfg))
			if err == nil {
				t.Fatal("resolveConfig() succeeded, want error")
			}
		})
	}
}

func TestSnapshotQueryUsesOnlyBoundCursorValue(t *testing.T) {
	query, args := snapshotQuery("analytics", "events", "sequence", "4", 2)
	if query != "SELECT * FROM `analytics`.`events` WHERE `sequence` > ? ORDER BY `sequence` ASC LIMIT 2" {
		t.Fatalf("snapshotQuery() = %q", query)
	}
	if !reflect.DeepEqual(args, []any{"4"}) {
		t.Fatalf("snapshot args = %#v", args)
	}
}

func TestIdentifierErrorDoesNotEchoCallerValue(t *testing.T) {
	value := "not-an-identifier;"
	err := validateIdentifier(value)
	if err == nil {
		t.Fatal("validateIdentifier() succeeded for unsafe input")
	}
	if strings.Contains(err.Error(), value) {
		t.Fatalf("identifier error exposed its caller value: %q", err)
	}
}

func TestCatalogStreamsIncludesConfiguredCursorOnlyWhenPresent(t *testing.T) {
	streams, err := catalogStreams("analytics", []connectors.Record{
		{"table_name": []byte("events"), "column_name": []byte("id"), "data_type": []byte("bigint")},
		{"table_name": []byte("events"), "column_name": []byte("sequence"), "data_type": []byte("bigint")},
		{"table_name": "other", "column_name": "id", "data_type": "bigint"},
	}, map[string][]string{"events": {"id"}, "other": {"id"}}, "sequence")
	if err != nil {
		t.Fatalf("catalogStreams(): %v", err)
	}
	if len(streams) != 2 || !reflect.DeepEqual(streams[0].CursorFields, []string{"sequence"}) || len(streams[1].CursorFields) != 0 {
		t.Fatalf("catalog streams = %#v", streams)
	}
}

func TestBinlogPositionStateRequiresCompleteValidPosition(t *testing.T) {
	if _, err := binlogPositionFromState(map[string]string{"binlog_file": "mysql-bin.000001"}); err == nil {
		t.Fatal("partial binlog state accepted")
	}
	position, err := binlogPositionFromState(map[string]string{"binlog_file": "mysql-bin.000001", "binlog_pos": "4"})
	if err != nil || position.Name != "mysql-bin.000001" || position.Pos != 4 {
		t.Fatalf("binlogPositionFromState() = %#v, %v", position, err)
	}
}

func TestCurrentBinlogStatusQueryUsesMySQL84Syntax(t *testing.T) {
	if currentBinlogStatusQuery != "SHOW BINARY LOG STATUS" {
		t.Fatalf("current binlog status query = %q", currentBinlogStatusQuery)
	}
}

func TestWriteUnsupported(t *testing.T) {
	_, err := New().Write(context.Background(), connectors.WriteRequest{}, nil)
	if !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write() = %v, want ErrUnsupportedOperation", err)
	}
}
