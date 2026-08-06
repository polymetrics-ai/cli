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
	manifest := connectors.ManifestOf(c)
	if manifest.Risk.Read == "" || len(manifest.ConfigFields) == 0 || len(manifest.SecretFields) != 1 || manifest.SecretFields[0].Name != "password" {
		t.Fatalf("native mysql manifest = %+v, want bundle configuration and risk", manifest)
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

func TestSnapshotQueryUsesCompleteDeterministicPageBounds(t *testing.T) {
	for _, tc := range []struct {
		name                                      string
		cursorField, lowerCursor, lowerPrimaryKey string
		resume, hasPageBoundary                   bool
		wantQuery                                 string
		wantArgs                                  []any
	}{
		{
			name:        "unfiltered cursor snapshot",
			cursorField: "sequence",
			wantQuery:   "SELECT * FROM `analytics`.`events` ORDER BY `sequence` ASC, `id` ASC LIMIT 2",
		},
		{
			name:        "unique resumed cursor boundary",
			cursorField: "sequence",
			lowerCursor: "4",
			resume:      true,
			wantQuery:   "SELECT * FROM `analytics`.`events` WHERE `sequence` > ? ORDER BY `sequence` ASC, `id` ASC LIMIT 2",
			wantArgs:    []any{"4"},
		},
		{
			name:            "cursor primary key continuation",
			cursorField:     "sequence",
			lowerCursor:     "4",
			lowerPrimaryKey: "12",
			hasPageBoundary: true,
			wantQuery:       "SELECT * FROM `analytics`.`events` WHERE (`sequence` > ? OR (`sequence` = ? AND `id` > ?)) ORDER BY `sequence` ASC, `id` ASC LIMIT 2",
			wantArgs:        []any{"4", "4", "12"},
		},
		{
			name:      "unfiltered primary key snapshot",
			wantQuery: "SELECT * FROM `analytics`.`events` ORDER BY `id` ASC LIMIT 2",
		},
		{
			name:            "primary key continuation",
			lowerPrimaryKey: "12",
			hasPageBoundary: true,
			wantQuery:       "SELECT * FROM `analytics`.`events` WHERE `id` > ? ORDER BY `id` ASC LIMIT 2",
			wantArgs:        []any{"12"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			query, args := snapshotQuery("analytics", "events", tc.cursorField, "id", tc.lowerCursor, tc.lowerPrimaryKey, tc.resume, tc.hasPageBoundary, 2)
			if query != tc.wantQuery {
				t.Fatalf("snapshotQuery() = %q, want %q", query, tc.wantQuery)
			}
			if !reflect.DeepEqual(args, tc.wantArgs) {
				t.Fatalf("snapshot args = %#v, want %#v", args, tc.wantArgs)
			}
		})
	}
}

func TestSingleColumnPrimaryKeyIsRequiredForCompletePaging(t *testing.T) {
	for _, primaryKey := range [][]string{nil, []string{"id", "tenant_id"}, []string{"unsafe-key"}} {
		if _, err := singleColumnPrimaryKey(primaryKey); err == nil {
			t.Fatalf("singleColumnPrimaryKey(%v) succeeded", primaryKey)
		}
	}
	if primaryKey, err := singleColumnPrimaryKey([]string{"id"}); err != nil || primaryKey != "id" {
		t.Fatalf("singleColumnPrimaryKey() = %q, %v", primaryKey, err)
	}
}

func TestUniqueCursorFieldMetadataIsRequired(t *testing.T) {
	if err := validateUniqueCursorFieldRecords([]connectors.Record{{"column_name": "sequence"}}, "sequence"); err != nil {
		t.Fatalf("validateUniqueCursorFieldRecords() error = %v", err)
	}
	for _, records := range [][]connectors.Record{nil, []connectors.Record{{"column_name": "other"}}} {
		if err := validateUniqueCursorFieldRecords(records, "sequence"); err == nil {
			t.Fatalf("validateUniqueCursorFieldRecords(%#v) succeeded", records)
		}
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
	columns := []connectors.Record{
		{"table_name": []byte("events"), "column_name": []byte("id"), "data_type": []byte("bigint")},
		{"table_name": []byte("events"), "column_name": []byte("sequence"), "data_type": []byte("bigint")},
		{"table_name": "other", "column_name": "id", "data_type": "bigint"},
	}
	primaryKeys := map[string][]string{"events": {"id"}, "other": {"id"}}
	streams, err := catalogStreams("analytics", columns, primaryKeys, "sequence", map[string]bool{"events": true})
	if err != nil {
		t.Fatalf("catalogStreams(): %v", err)
	}
	if len(streams) != 2 || !reflect.DeepEqual(streams[0].CursorFields, []string{"sequence"}) || len(streams[1].CursorFields) != 0 {
		t.Fatalf("catalog streams = %#v", streams)
	}
	withoutUniqueCursor, err := catalogStreams("analytics", columns, primaryKeys, "sequence", nil)
	if err != nil {
		t.Fatalf("catalogStreams() without unique cursor: %v", err)
	}
	if len(withoutUniqueCursor[0].CursorFields) != 0 {
		t.Fatalf("catalog streams = %#v, want no unsafe cursor", withoutUniqueCursor)
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

func TestValidateBinlogRequirementsFailsClosed(t *testing.T) {
	for _, tc := range []struct {
		name    string
		records []connectors.Record
		wantErr bool
	}{
		{name: "row full", records: []connectors.Record{{"binlog_format": "ROW", "binlog_row_image": "FULL"}}},
		{name: "case insensitive row full", records: []connectors.Record{{"binlog_format": []byte("row"), "binlog_row_image": []byte("full")}}},
		{name: "statement", records: []connectors.Record{{"binlog_format": "STATEMENT", "binlog_row_image": "FULL"}}, wantErr: true},
		{name: "minimal row image", records: []connectors.Record{{"binlog_format": "ROW", "binlog_row_image": "MINIMAL"}}, wantErr: true},
		{name: "missing settings", records: []connectors.Record{{}}, wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := validateBinlogRequirementRecords(tc.records)
			if (err != nil) != tc.wantErr {
				t.Fatalf("validateBinlogRequirementRecords() error = %v, want error=%t", err, tc.wantErr)
			}
		})
	}
}

func TestCDCRowImagesAndDedupeOrdinal(t *testing.T) {
	if err := validateCDCRowImages([][]any{{"6", "foxtrot"}}, [][]int{{1}}, []string{"id", "label"}); err == nil {
		t.Fatal("validateCDCRowImages() accepted an omitted column")
	}
	if err := validateCDCRowImages([][]any{{"6", "foxtrot"}}, [][]int{}, []string{"id", "label"}); err == nil {
		t.Fatal("validateCDCRowImages() accepted incomplete row image metadata")
	}
	state := connectors.Record{"binlog_file": "mysql-bin.000001", "binlog_pos": "412"}
	events := cdcEventsFromRows("insert", [][]any{{[]byte("6"), "foxtrot"}, {[]byte("7"), "golf"}}, []string{"id", "label"}, state)
	if len(events) != 2 || events[0].State["binlog_row"] != "0" || events[1].State["binlog_row"] != "1" {
		t.Fatalf("CDC events = %#v, want ordered per-row dedupe state", events)
	}
	if events[0].State["binlog_file"] != "mysql-bin.000001" || events[0].State["binlog_pos"] != "412" {
		t.Fatalf("CDC event state = %#v, want preserved binlog position", events[0].State)
	}
	if _, ok := state["binlog_row"]; ok {
		t.Fatal("cdcEventsFromRows() mutated the checkpoint state")
	}
}

func TestWriteUnsupported(t *testing.T) {
	_, err := New().Write(context.Background(), connectors.WriteRequest{}, nil)
	if !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write() = %v, want ErrUnsupportedOperation", err)
	}
}
