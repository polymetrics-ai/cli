package mysql

import (
	"context"
	"errors"
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
	if checkpoint := c.ChangefeedExecutorDescriptor().Checkpoint.Keys; !reflect.DeepEqual(checkpoint, []string{"binlog_file", "binlog_pos", mysqlCDCSchemaFingerprintState}) {
		t.Fatalf("CDC checkpoint keys = %v, want schema-bound position", checkpoint)
	}
	manifest := connectors.ManifestOf(c)
	if manifest.Risk.Read == "" || len(manifest.ConfigFields) == 0 || len(manifest.SecretFields) != 1 || manifest.SecretFields[0].Name != "password" {
		t.Fatalf("native mysql manifest = %+v, want bundle configuration and risk", manifest)
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

func TestCDCQueryEventsFailClosed(t *testing.T) {
	for _, tc := range []struct {
		name    string
		query   string
		wantErr bool
	}{
		{name: "begin", query: "BEGIN"},
		{name: "commit", query: "COMMIT;"},
		{name: "rollback", query: "ROLLBACK"},
		{name: "session metadata", query: "SET TIMESTAMP=1723034096"},
		{name: "session binlog format", query: "SET SESSION binlog_format = 'STATEMENT'", wantErr: true},
		{name: "session row image", query: "SET SESSION binlog_row_image = 'MINIMAL'", wantErr: true},
		{name: "schema change", query: "ALTER TABLE events ADD COLUMN ignored INT", wantErr: true},
		{name: "statement mutation", query: "INSERT INTO events VALUES (1)", wantErr: true},
		{name: "empty statement", wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := validateCDCQueryEvent(&replication.QueryEvent{Query: []byte(tc.query)})
			if (err != nil) != tc.wantErr {
				t.Fatalf("validateCDCQueryEvent(%q) error = %v, want error=%t", tc.query, err, tc.wantErr)
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
