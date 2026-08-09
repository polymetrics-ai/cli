package mysql

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"reflect"
	"strings"
	"testing"
	"time"

	gomysql "github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/replication"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/native/sqltls"
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

func TestNameMetadataKeepsTheInternalCDCReaderNonPublic(t *testing.T) {
	c := New()
	if c.Name() != "mysql" {
		t.Fatalf("Name() = %q, want mysql", c.Name())
	}
	caps := connectors.MetadataOf(c).Capabilities
	if !caps.Check || !caps.Catalog || !caps.Read || caps.CDC || caps.Write {
		t.Fatalf("capabilities = %+v, want check/catalog/read only", caps)
	}
	definition, ok := any(c).(connectors.DefinitionProvider)
	if !ok {
		t.Fatal("mysql connector has no bundle definition")
	}
	if definition.Definition().Changefeed != nil {
		t.Fatal("mysql connector must not declare a public changefeed before an operator entrypoint exists")
	}
	if _, ok := any(c).(connectors.ChangefeedExecutor); ok {
		t.Fatal("mysql connector must not expose its internal CDC reader as a public changefeed executor")
	}
	if _, ok := any(c).(connectors.CDCReader); !ok {
		t.Fatal("mysql connector lost its internally proven CDC reader")
	}
	manifest := connectors.ManifestOf(c)
	if manifest.Risk.Read == "" || len(manifest.ConfigFields) == 0 || len(manifest.SecretFields) != 1 || manifest.SecretFields[0].Name != "password" {
		t.Fatalf("native mysql manifest = %+v, want bundle configuration and risk", manifest)
	}
}

// #3902's DirectReadPage contract applies to one-page HTTP/API exploration.
// MySQL exposes no direct-read command or REST operation: Read is the ETL
// interface and deliberately drains its deterministic SQL pages into the
// caller's sync pipeline, bounded by page_size and read_limit. Adding a
// DirectReader later must add page context rather than silently repurposing
// this bulk reader.
func TestReadIsETLNotPagewiseDirectRead(t *testing.T) {
	connector := New()
	if _, ok := any(connector).(connectors.DirectReader); ok {
		t.Fatal("native MySQL Read must not be exposed as a pagewise DirectReader")
	}
	if _, ok := any(connector).(connectors.OperationDirectReader); ok {
		t.Fatal("native MySQL has no declared operation direct-read surface")
	}
}

func TestInitialStateOmitsCursorUntilOneExists(t *testing.T) {
	state, err := New().InitialState(context.Background(), "analytics.events", testConfig())
	if err != nil {
		t.Fatalf("InitialState(): %v", err)
	}
	if state["stream"] != "analytics.events" {
		t.Fatalf("initial state = %#v, want stream", state)
	}
	if _, present := state[connsdk.CursorStateKey]; present {
		t.Fatalf("initial state = %#v, want no cursor", state)
	}
}

func TestRawCursorStatePreservesExplicitValues(t *testing.T) {
	for _, tc := range []struct {
		name    string
		state   map[string]string
		want    string
		present bool
	}{
		{name: "missing state"},
		{name: "missing cursor", state: map[string]string{"stream": "analytics.events"}},
		{name: "empty cursor", state: map[string]string{connsdk.CursorStateKey: ""}, present: true},
		{name: "whitespace cursor", state: map[string]string{connsdk.CursorStateKey: "  "}, want: "  ", present: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, present := rawCursorState(tc.state)
			if got != tc.want || present != tc.present {
				t.Fatalf("rawCursorState(%#v) = (%q, %t), want (%q, %t)", tc.state, got, present, tc.want, tc.present)
			}
		})
	}
}

func TestOpaqueCursorStatePreservesNativeMySQLBoundaryValues(t *testing.T) {
	connector := New()
	for _, tc := range []struct {
		name  string
		value any
	}{
		{name: "string", value: "delta"},
		{name: "binary", value: []byte{0x00, 0xff, 0x01}},
		{name: "signed", value: int64(-42)},
		{name: "unsigned", value: uint64(1<<63 + 7)},
		{name: "floating", value: math.Float64frombits(0x400921fb54442d18)},
		{name: "boolean", value: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state, err := connector.CursorStateFromRecord(connectors.Record{"sequence": tc.value}, "sequence")
			if err != nil {
				t.Fatalf("CursorStateFromRecord(): %v", err)
			}
			got, present, err := readCursorState(connectors.ReadRequest{CursorState: state})
			if err != nil || !present || !reflect.DeepEqual(got, tc.value) {
				t.Fatalf("readCursorState() = (%#v, %t, %v), want (%#v, true, nil)", got, present, err, tc.value)
			}
			_, args := snapshotQuery("analytics", "events", "sequence", "id", got, nil, true, false, 2)
			if !reflect.DeepEqual(args, []any{tc.value}) {
				t.Fatalf("resumed query args = %#v, want %#v", args, []any{tc.value})
			}
		})
	}

	original := []byte{0x00, 0xff}
	state, err := connector.CursorStateFromRecord(connectors.Record{"sequence": original}, "sequence")
	if err != nil {
		t.Fatal(err)
	}
	original[0] = 0x7f
	got, present, err := readCursorState(connectors.ReadRequest{CursorState: state})
	if err != nil || !present || !reflect.DeepEqual(got, []byte{0x00, 0xff}) {
		t.Fatalf("copied binary cursor = %#v, %t, %v", got, present, err)
	}

	if _, _, err := readCursorState(connectors.ReadRequest{CursorState: connectors.OpaqueCursorState{Token: []byte("unknown"), Present: true}}); err == nil {
		t.Fatal("readCursorState() accepted an unrecognized opaque cursor")
	}
}

func TestSourceOrderedCursorFieldMustMatchMySQLConfiguration(t *testing.T) {
	connector := New()
	config := testConfig()
	if err := connector.ValidateCursorField(config, "sequence"); err != nil {
		t.Fatalf("ValidateCursorField(matching): %v", err)
	}
	for _, field := range []string{"", "updated_at"} {
		if err := connector.ValidateCursorField(config, field); err == nil {
			t.Fatalf("ValidateCursorField(%q) error = nil", field)
		}
	}
	config.Config["cursor_field"] = ""
	if err := connector.ValidateCursorField(config, "sequence"); err == nil {
		t.Fatal("ValidateCursorField(missing configuration) error = nil")
	}
}

func TestOpaqueCursorComparisonRetainsNativeBinaryOrder(t *testing.T) {
	connector := New()
	first, err := connector.CursorStateFromRecord(connectors.Record{"sequence": []byte{0x00}}, "sequence")
	if err != nil {
		t.Fatal(err)
	}
	last, err := connector.CursorStateFromRecord(connectors.Record{"sequence": []byte{0xff}}, "sequence")
	if err != nil {
		t.Fatal(err)
	}
	cmp, err := connector.CompareCursorStates(last, first)
	if err != nil || cmp <= 0 {
		t.Fatalf("CompareCursorStates(0xff, 0x00) = (%d, %v), want positive", cmp, err)
	}
	text, err := connector.CursorStateFromRecord(connectors.Record{"sequence": "later"}, "sequence")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := connector.CompareCursorStates(text, text); !errors.Is(err, connectors.ErrOpaqueCursorOrderUnavailable) {
		t.Fatalf("CompareCursorStates(text, text) error = %v, want unavailable order", err)
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
			name:        "empty resumed cursor boundary",
			cursorField: "sequence",
			resume:      true,
			wantQuery:   "SELECT * FROM `analytics`.`events` WHERE `sequence` > ? ORDER BY `sequence` ASC, `id` ASC LIMIT 2",
			wantArgs:    []any{""},
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

func TestResultRecordsNormalizesTextAndCopiesBinaryValues(t *testing.T) {
	binaryValue := []byte{0x01, 0x02, 0x03}
	result := &gomysql.Result{Resultset: &gomysql.Resultset{
		Fields: []*gomysql.Field{
			{Name: []byte("label"), Type: gomysql.MYSQL_TYPE_VAR_STRING, Charset: 45},
			{Name: []byte("occurred_at"), Type: gomysql.MYSQL_TYPE_DATETIME, Charset: mysqlBinaryCharsetID, Flag: gomysql.BINARY_FLAG},
			{Name: []byte("metadata"), Type: gomysql.MYSQL_TYPE_JSON, Charset: mysqlBinaryCharsetID},
			{Name: []byte("payload"), Type: gomysql.MYSQL_TYPE_BLOB, Charset: mysqlBinaryCharsetID, Flag: gomysql.BINARY_FLAG},
		},
		Values: [][]gomysql.FieldValue{{
			gomysql.NewFieldValue(gomysql.FieldValueTypeString, 0, []byte("alpha")),
			gomysql.NewFieldValue(gomysql.FieldValueTypeString, 0, []byte("2026-08-07 12:34:56")),
			gomysql.NewFieldValue(gomysql.FieldValueTypeString, 0, []byte(`{"enabled":true}`)),
			gomysql.NewFieldValue(gomysql.FieldValueTypeString, 0, binaryValue),
		}},
	}}
	records, err := resultRecords(result)
	if err != nil {
		t.Fatalf("resultRecords(): %v", err)
	}
	if len(records) != 1 || records[0]["label"] != "alpha" || records[0]["occurred_at"] != "2026-08-07 12:34:56" || records[0]["metadata"] != `{"enabled":true}` {
		t.Fatalf("result records = %#v, want normalized text values", records)
	}
	payload, ok := records[0]["payload"].([]byte)
	if !ok || !reflect.DeepEqual(payload, binaryValue) {
		t.Fatalf("payload = %#v, want copied binary value %#v", records[0]["payload"], binaryValue)
	}
	source, ok := result.Values[0][3].Value().([]byte)
	if !ok {
		t.Fatal("test result did not retain binary source bytes")
	}
	source[0] = 0xff
	if payload[0] != 0x01 {
		t.Fatal("resultRecords() retained an alias to binary source bytes")
	}
	if _, ok := copyReadBoundaryValue(records[0]["label"]).(string); !ok {
		t.Fatal("text cursor boundary is not a string")
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

func TestBinlogCheckpointStateRequiresSchemaFingerprint(t *testing.T) {
	columns := []string{"id", "label"}
	fingerprint := cdcSchemaFingerprint(columns)
	if _, _, err := binlogCheckpointFromState(nil); err != nil {
		t.Fatalf("initial checkpoint state = %v, want accepted", err)
	}
	if _, _, err := binlogCheckpointFromState(map[string]string{"binlog_file": "mysql-bin.000001", "binlog_pos": "4"}); err == nil {
		t.Fatal("checkpoint without schema fingerprint accepted")
	}
	if _, _, err := binlogCheckpointFromState(map[string]string{mysqlCDCSchemaFingerprintState: fingerprint}); err == nil {
		t.Fatal("schema fingerprint without checkpoint position accepted")
	}
	if _, _, err := binlogCheckpointFromState(map[string]string{"binlog_file": "mysql-bin.000001", "binlog_pos": "4", mysqlCDCSchemaFingerprintState: "not-a-fingerprint"}); err == nil {
		t.Fatal("invalid schema fingerprint accepted")
	}
	position, gotFingerprint, err := binlogCheckpointFromState(map[string]string{
		"binlog_file":                  "mysql-bin.000001",
		"binlog_pos":                   "4",
		mysqlCDCSchemaFingerprintState: fingerprint,
	})
	if err != nil || position.Name != "mysql-bin.000001" || position.Pos != 4 || gotFingerprint != fingerprint {
		t.Fatalf("binlogCheckpointFromState() = %#v, %q, %v", position, gotFingerprint, err)
	}
	if cdcSchemaFingerprint([]string{"label", "id"}) == fingerprint {
		t.Fatal("schema fingerprint did not bind ordered column metadata")
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

// TestCDCQueryEventsFailClosed pins the blast radius of a statement event. The
// binary log is server-wide, so a changefeed on one database must survive
// unrelated schema activity while still failing closed on anything that can
// reach its own database or that cannot be attributed at all.
func TestCDCQueryEventsFailClosed(t *testing.T) {
	const database = "pm_harness"
	for _, tc := range []struct {
		name    string
		schema  string
		query   string
		wantErr bool
	}{
		{name: "begin", schema: database, query: "BEGIN"},
		{name: "commit", schema: database, query: "COMMIT;"},
		{name: "rollback", schema: database, query: "ROLLBACK"},
		{name: "session metadata", schema: database, query: "SET TIMESTAMP=1723034096"},
		{name: "session binlog format", schema: database, query: "SET SESSION binlog_format = 'STATEMENT'", wantErr: true},
		{name: "session row image", schema: database, query: "SET SESSION binlog_row_image = 'MINIMAL'", wantErr: true},
		// A replication format change silences row events for every schema, so
		// it is rejected even when it arrives under an unrelated one.
		{name: "row image change from another schema", schema: "other_app", query: "SET GLOBAL binlog_row_image = 'MINIMAL'", wantErr: true},
		{name: "schema change", schema: database, query: "ALTER TABLE events ADD COLUMN ignored INT", wantErr: true},
		{name: "statement mutation", schema: database, query: "INSERT INTO events VALUES (1)", wantErr: true},
		{name: "empty statement", schema: database, wantErr: true},
		// Unrelated schema activity must not end this changefeed.
		{name: "unrelated schema change", schema: "other_app", query: "ALTER TABLE events ADD COLUMN ignored INT"},
		{name: "unrelated qualified schema change", schema: "other_app", query: "ALTER TABLE other_app.events ADD COLUMN ignored INT"},
		{name: "server wide grant", schema: "other_app", query: "GRANT SELECT ON other_app.* TO 'reader'@'%'"},
		{name: "unrelated analyze", schema: "other_app", query: "ANALYZE TABLE inventory"},
		// A qualified name reaches across the default schema, so the default is
		// never proof of what a statement touched.
		{name: "cross schema qualified alter", schema: "other_app", query: "ALTER TABLE pm_harness.events ADD COLUMN ignored INT", wantErr: true},
		{name: "cross schema backquoted alter", schema: "other_app", query: "ALTER TABLE `pm_harness`.`events` ADD COLUMN ignored INT", wantErr: true},
		// Under sql_mode=ANSI_QUOTES a double-quoted span is an identifier, and
		// a rename or type change keeps the column count, so nothing downstream
		// would notice it. What the quote means is undecidable here, so it is
		// read as a name rather than skipped as a constant.
		{name: "cross schema ansi quoted alter", schema: "other_app", query: `ALTER TABLE "pm_harness"."events" CHANGE COLUMN label title VARCHAR(64)`, wantErr: true},
		{name: "cross schema ansi quoted drop", schema: "other_app", query: `DROP TABLE "pm_harness"."events"`, wantErr: true},
		{name: "unrelated ansi quoted alter", schema: "other_app", query: `ALTER TABLE "other_app"."events" CHANGE COLUMN label title VARCHAR(64)`},
		// An escaped quote inside a double-quoted span used to end it one quote
		// early, shifting every later boundary until the target reference was
		// swallowed into one unread blob. A RENAME keeps the column count, so
		// nothing downstream would have caught the mis-projection.
		{name: "escaped double quote before a target reference", schema: "other_app", query: `ALTER TABLE t COMMENT "a\"b", RENAME TO pm_harness.t2`, wantErr: true},
		{name: "escaped single quote before a target reference", schema: "other_app", query: `ALTER TABLE t COMMENT 'a\'b', RENAME TO pm_harness.t2`, wantErr: true},
		{name: "doubled quote inside an ansi quoted target", schema: "other_app", query: `ALTER TABLE "pm_harness"."od""d" ADD COLUMN ignored INT`, wantErr: true},
		{name: "doubled quote inside an unrelated backquoted name", schema: "other_app", query: "ALTER TABLE `od``d` ADD COLUMN ignored INT"},
		// A span the scan cannot close means it no longer knows where any later
		// quote sits, so the statement is unattributable rather than unrelated.
		{name: "unterminated double quoted span", schema: "other_app", query: `ALTER TABLE t COMMENT "never closed`, wantErr: true},
		{name: "unterminated single quoted span", schema: "other_app", query: `INSERT INTO audit (note) VALUES ('never closed`, wantErr: true},
		{name: "unterminated backquoted identifier", schema: "other_app", query: "ALTER TABLE `never closed", wantErr: true},
		{name: "cross schema drop database", schema: "other_app", query: "DROP DATABASE pm_harness", wantErr: true},
		{name: "cross schema rename into target", schema: "other_app", query: "RENAME TABLE other_app.events TO pm_harness.events", wantErr: true},
		// A literal that merely spells the database name is not an object
		// reference, but an unattributable statement still fails closed.
		{name: "database name only inside a literal", schema: "other_app", query: "INSERT INTO audit (note) VALUES ('pm_harness')"},
		{name: "absent default schema", query: "ANALYZE TABLE inventory", wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			event := &replication.QueryEvent{Query: []byte(tc.query), Schema: []byte(tc.schema)}
			err := validateCDCQueryEvent(event, database)
			if (err != nil) != tc.wantErr {
				t.Fatalf("validateCDCQueryEvent(%q) under schema %q error = %v, want error=%t", tc.query, tc.schema, err, tc.wantErr)
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
	columns := []string{"id", "label"}
	fingerprint := cdcSchemaFingerprint(columns)
	state := connectors.Record{"binlog_file": "mysql-bin.000001", "binlog_pos": "412", mysqlCDCSchemaFingerprintState: fingerprint}
	events := cdcEventsFromRows("insert", [][]any{{[]byte("6"), "foxtrot"}, {[]byte("7"), "golf"}}, columns, state)
	if len(events) != 2 || events[0].State["binlog_row"] != "0" || events[1].State["binlog_row"] != "1" {
		t.Fatalf("CDC events = %#v, want ordered per-row dedupe state", events)
	}
	if events[0].State["binlog_file"] != "mysql-bin.000001" || events[0].State["binlog_pos"] != "412" || events[0].State[mysqlCDCSchemaFingerprintState] != fingerprint {
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

func TestTransportSecurityModeIsResolvedFromSharedSQLShape(t *testing.T) {
	for _, tc := range []struct {
		name         string
		mode         string
		wantMode     sqltls.Mode
		wantEncrypt  bool
		wantFallback bool
	}{
		{name: "default when unset", mode: "", wantMode: sqltls.ModePreferred, wantEncrypt: true, wantFallback: true},
		{name: "explicit disabled", mode: "disabled", wantMode: sqltls.ModeDisabled},
		{name: "explicit required", mode: "required", wantMode: sqltls.ModeRequired, wantEncrypt: true},
		{name: "libpq spelling", mode: "verify-full", wantMode: sqltls.ModeVerifyIdentity, wantEncrypt: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := testConfig()
			if tc.mode != "" {
				cfg.Config["sslmode"] = tc.mode
			}
			conn, err := resolveConfig(cfg)
			if err != nil {
				t.Fatalf("resolveConfig(): %v", err)
			}
			if conn.tls.Mode != tc.wantMode {
				t.Fatalf("mode = %q, want %q", conn.tls.Mode, tc.wantMode)
			}
			if conn.tls.Encrypted() != tc.wantEncrypt {
				t.Fatalf("Encrypted() = %t, want %t", conn.tls.Encrypted(), tc.wantEncrypt)
			}
			if conn.tls.MayFallBackToPlaintext() != tc.wantFallback {
				t.Fatalf("MayFallBackToPlaintext() = %t, want %t", conn.tls.MayFallBackToPlaintext(), tc.wantFallback)
			}
		})
	}
}

func TestResolveConfigRejectsUnknownTransportSecurityMode(t *testing.T) {
	cfg := testConfig()
	cfg.Config["sslmode"] = "definitely-not-a-mode"
	if _, err := resolveConfig(cfg); err == nil {
		t.Fatal("resolveConfig() accepted an unknown sslmode instead of refusing it")
	}
}

// A server that offers no TLS is the exact situation in which a connector may
// be tempted to downgrade. Prove that only "preferred" does.
func TestStrictTransportSecurityIsNotDowngradedAgainstATLSLessServer(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer func() { _ = listener.Close() }()
	go serveTLSLessMySQLGreeting(listener)

	host, portText, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatalf("split addr: %v", err)
	}

	for _, tc := range []struct {
		mode        string
		wantRefusal bool
	}{
		{mode: "required", wantRefusal: true},
		{mode: "verify-ca", wantRefusal: true},
		{mode: "verify-identity", wantRefusal: true},
		{mode: "preferred"},
		{mode: "disabled"},
	} {
		t.Run(tc.mode, func(t *testing.T) {
			cfg := testConfig()
			cfg.Config["host"] = host
			cfg.Config["port"] = portText
			cfg.Config["sslmode"] = tc.mode
			conn, err := resolveConfig(cfg)
			if err != nil {
				t.Fatalf("resolveConfig(): %v", err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			tlsConfig, err := conn.tls.TLSConfig(conn.host)
			if err != nil {
				t.Fatalf("TLSConfig(): %v", err)
			}
			_, dialErr := conn.dial(ctx, tlsConfig)
			if tc.wantRefusal {
				if dialErr == nil || !strings.Contains(dialErr.Error(), serverRefusedTLS) {
					t.Fatalf("dial under %q error = %v, want the driver's TLS refusal", tc.mode, dialErr)
				}
				// The refusal must not be convertible into a plaintext retry.
				if conn.tls.MayFallBackToPlaintext() {
					t.Fatalf("mode %q permitted a plaintext fallback", tc.mode)
				}
				return
			}
			if tlsConfig != nil && !conn.tls.MayFallBackToPlaintext() {
				t.Fatalf("mode %q would encrypt without a documented fallback", tc.mode)
			}
		})
	}
}

// serveTLSLessMySQLGreeting writes a MySQL 8.4 initial handshake whose
// capability flags omit CLIENT_SSL, then closes. That is all the client needs
// to see to decide whether TLS is possible.
func serveTLSLessMySQLGreeting(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		go func(conn net.Conn) {
			defer func() { _ = conn.Close() }()
			// Capabilities: CLIENT_PROTOCOL_41 (0x00000200) only. CLIENT_SSL
			// (0x00000800) is deliberately absent.
			const capabilities = 0x00000200
			body := []byte{10}                                                         // protocol version
			body = append(body, []byte("8.4.11-test")...)                              // server version
			body = append(body, 0)                                                     // version terminator
			body = append(body, 1, 0, 0, 0)                                            // connection id
			body = append(body, make([]byte, 8)...)                                    // auth-plugin-data-part-1
			body = append(body, 0)                                                     // filler
			body = append(body, byte(capabilities&0xff), byte((capabilities>>8)&0xff)) // capability flags lower
			body = append(body, 0xff)                                                  // charset
			body = append(body, 2, 0)                                                  // status flags
			body = append(body, byte((capabilities>>16)&0xff), byte((capabilities>>24)&0xff))
			body = append(body, 21)                  // auth plugin data length
			body = append(body, make([]byte, 10)...) // reserved
			body = append(body, make([]byte, 13)...) // auth-plugin-data-part-2
			body = append(body, []byte("mysql_native_password")...)
			body = append(body, 0)

			header := []byte{byte(len(body)), byte(len(body) >> 8), byte(len(body) >> 16), 0}
			_, _ = conn.Write(append(header, body...))
		}(conn)
	}
}

// recordingExecutor captures the statement a discovery helper issues. It
// returns a server error so the caller stops at the query it was asked about.
type recordingExecutor struct {
	queries []string
	args    [][]any
}

func (e *recordingExecutor) Execute(query string, args ...any) (*gomysql.Result, error) {
	e.queries = append(e.queries, query)
	e.args = append(e.args, append([]any(nil), args...))
	return nil, errors.New("scripted server error")
}

// Catalog needs every table's primary key; a Read needs one. The
// information_schema join is the expensive part of starting a read, so it must
// be bounded to the stream actually being paged.
func TestPrimaryKeyDiscoveryIsBoundedToTheStreamBeingRead(t *testing.T) {
	read := &recordingExecutor{}
	if _, err := discoverPrimaryKeys(context.Background(), read, "analytics", "events"); err == nil {
		t.Fatal("discoverPrimaryKeys() swallowed the server error")
	}
	if !strings.Contains(read.queries[0], "t.table_name = ?") {
		t.Fatalf("read discovery query = %q, want a single-table predicate", read.queries[0])
	}
	if want := []any{"analytics", "events"}; !reflect.DeepEqual(read.args[0], want) {
		t.Fatalf("read discovery args = %v, want %v", read.args[0], want)
	}

	catalog := &recordingExecutor{}
	if _, err := discoverPrimaryKeys(context.Background(), catalog, "analytics", ""); err == nil {
		t.Fatal("discoverPrimaryKeys() swallowed the server error")
	}
	if strings.Contains(catalog.queries[0], "t.table_name = ?") {
		t.Fatalf("catalog discovery query = %q, want the whole schema", catalog.queries[0])
	}
	if want := []any{"analytics"}; !reflect.DeepEqual(catalog.args[0], want) {
		t.Fatalf("catalog discovery args = %v, want %v", catalog.args[0], want)
	}
}

// A cancelled sync must surface as a cancellation, not as a connection
// failure, while still returning none of the driver's endpoint text.
func TestDialErrorsKeepCancellationDistinguishableWithoutLeakingConfiguration(t *testing.T) {
	const leaky = "dial tcp 10.0.0.5:3306: user=admin: connection refused"
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()

	for _, tc := range []struct {
		name string
		ctx  context.Context
		err  error
		want error
	}{
		{name: "cancelled context", ctx: cancelled, err: errors.New(leaky), want: context.Canceled},
		{name: "cancelled dial", ctx: context.Background(), err: fmt.Errorf("dial: %w", context.Canceled), want: context.Canceled},
		{name: "expired dial", ctx: context.Background(), err: fmt.Errorf("dial: %w", context.DeadlineExceeded), want: context.DeadlineExceeded},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := dialError(tc.ctx, tc.err)
			if !errors.Is(err, tc.want) {
				t.Fatalf("dialError() = %v, want %v", err, tc.want)
			}
			if strings.Contains(err.Error(), "10.0.0.5") || strings.Contains(err.Error(), "admin") {
				t.Fatalf("dialError() = %q, want no configuration or endpoint material", err)
			}
		})
	}

	generic := dialError(context.Background(), errors.New(leaky))
	if generic.Error() != "connect mysql failed" {
		t.Fatalf("dialError() = %q, want the opaque connection failure", generic)
	}
	if errors.Is(generic, context.Canceled) || errors.Is(generic, context.DeadlineExceeded) {
		t.Fatalf("dialError() = %v, want an ordinary failure to stay distinct from cancellation", generic)
	}
}

// A database almost always contains a table this connector cannot quote.
// Discovery must stay usable for the rest, and must never advertise a stream
// whose every read would fail.
func TestCatalogSkipsUnreadableIdentifiersWithoutFailingDiscovery(t *testing.T) {
	columns := []connectors.Record{
		{"table_name": "events", "column_name": "id", "data_type": "bigint", "ordinal_position": "1"},
		{"table_name": "events", "column_name": "label", "data_type": "varchar", "ordinal_position": "2"},
		// Unquotable table name.
		{"table_name": "legacy-report", "column_name": "id", "data_type": "bigint", "ordinal_position": "1"},
		// Readable table carrying one unquotable column.
		{"table_name": "orders", "column_name": "id", "data_type": "bigint", "ordinal_position": "1"},
		{"table_name": "orders", "column_name": "total amount", "data_type": "decimal", "ordinal_position": "2"},
	}
	streams, err := catalogStreams("analytics", columns, map[string][]string{"events": {"id"}}, "", nil)
	if err != nil {
		t.Fatalf("catalogStreams() = %v, want the readable tables rather than a whole-database failure", err)
	}
	got := make([]string, 0, len(streams))
	for _, stream := range streams {
		got = append(got, stream.Name)
	}
	if want := []string{"analytics.events"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("streams = %v, want %v", got, want)
	}
	// Every advertised stream must survive the reader's own parsing.
	for _, stream := range streams {
		if _, _, err := qualifyStream("analytics", stream.Name); err != nil {
			t.Fatalf("catalog advertised unreadable stream %q: %v", stream.Name, err)
		}
	}
}
