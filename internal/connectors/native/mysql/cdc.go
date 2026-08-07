package mysql

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"

	gomysql "github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/replication"

	"polymetrics.ai/internal/connectors"
)

const (
	mysqlBinlogExecutorID          = "mysql-binlog-row-v1"
	mysqlCDCSchemaFingerprintState = "schema_fingerprint"
)

// MySQL 8.4 removed SHOW MASTER STATUS with the legacy replication terms.
// SHOW BINARY LOG STATUS retains the File/Position output required for a
// position checkpoint.
const currentBinlogStatusQuery = "SHOW BINARY LOG STATUS"

const currentBinlogRequirementsQuery = "SELECT @@GLOBAL.binlog_format AS binlog_format, @@GLOBAL.binlog_row_image AS binlog_row_image"

// ChangefeedExecutorDescriptor is the runtime half of defs/mysql/changefeed.
// It must remain exactly in sync with the checked-in declaration; this is what
// keeps public CDC capability fail-closed.
func (c Connector) ChangefeedExecutorDescriptor() connectors.ChangefeedExecutorDescriptor {
	return connectors.ChangefeedExecutorDescriptor{
		Status:    connectors.ChangefeedStatusImplemented,
		Mechanism: connectors.ChangefeedMechanismBinlogReplication,
		Executor: connectors.ChangefeedExecutorRef{
			Kind: "native",
			ID:   mysqlBinlogExecutorID,
		},
		Checkpoint: connectors.ChangefeedCheckpoint{
			Kind:        "binlog_position",
			Keys:        []string{"binlog_file", "binlog_pos", mysqlCDCSchemaFingerprintState},
			CommitAfter: "downstream_ack",
			OnInvalid:   "resnapshot_required",
		},
	}
}

// ReadCDC consumes row-based binary-log events for one discovered table. It
// starts from a committed binlog file/position when supplied; otherwise it
// captures the current master position before opening the replication stream.
// A binlog event's state is committed only after all its records were accepted
// by emit, preventing a partial page from advancing a durable checkpoint.
func (c Connector) ReadCDC(ctx context.Context, req connectors.CDCReadRequest, emit func(connectors.CDCEvent) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if strings.TrimSpace(req.Stream) == "" {
		return errors.New("mysql CDC requires a stream (table or database.table)")
	}
	conn, err := resolveConfig(req.Config)
	if err != nil {
		return err
	}
	database, table, err := qualifyStream(conn.database, req.Stream)
	if err != nil {
		return err
	}
	serverID, err := replicationServerID(req.Config)
	if err != nil {
		return err
	}
	if err := validateBinlogRequirements(ctx, conn); err != nil {
		return err
	}
	position, checkpointSchema, err := binlogCheckpointFromState(req.State)
	if err != nil {
		return err
	}
	if position.Name == "" {
		position, err = currentBinlogPosition(ctx, conn)
		if err != nil {
			return err
		}
	}
	columns, err := loadColumnNames(ctx, conn, database, table)
	if err != nil {
		return err
	}
	schemaFingerprint := cdcSchemaFingerprint(columns)
	if checkpointSchema != "" && checkpointSchema != schemaFingerprint {
		return errors.New("mysql CDC schema changed; resnapshot required")
	}

	// Replication must run under the same transport-security choice as every
	// other statement; a syncer left at its zero value would connect in
	// plaintext regardless of the configured mode.
	replicationTLS, err := conn.replicationTLS(ctx)
	if err != nil {
		return err
	}
	syncer := replication.NewBinlogSyncer(replication.BinlogSyncerConfig{
		ServerID:  serverID,
		Flavor:    gomysql.MySQLFlavor,
		Host:      conn.host,
		Port:      uint16(conn.port),
		User:      conn.username,
		Password:  conn.password,
		TLSConfig: replicationTLS,
		// The upstream library emits configuration/position diagnostics through
		// this logger. Discard it: test and connector output must never become a
		// credential or endpoint logging path.
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	defer syncer.Close()
	streamer, err := syncer.StartSync(position)
	if err != nil {
		return fmt.Errorf("start mysql binlog replication: %w", err)
	}
	currentFile := position.Name
	for {
		event, err := streamer.GetEvent(ctx)
		if err != nil {
			return err
		}
		if event == nil || event.Header == nil {
			return errors.New("mysql binlog returned an event without a header")
		}
		if rotate, ok := event.Event.(*replication.RotateEvent); ok {
			currentFile = string(rotate.NextLogName)
			continue
		}
		if query, ok := event.Event.(*replication.QueryEvent); ok {
			if err := validateCDCQueryEvent(query); err != nil {
				return err
			}
			continue
		}
		rows, ok := event.Event.(*replication.RowsEvent)
		if !ok || rows.Table == nil {
			continue
		}
		if string(rows.Table.Schema) != database || string(rows.Table.Table) != table {
			continue
		}
		state := connectors.Record{
			"binlog_file":                  currentFile,
			"binlog_pos":                   strconv.FormatUint(uint64(event.Header.LogPos), 10),
			mysqlCDCSchemaFingerprintState: schemaFingerprint,
		}
		events, err := rowsToCDCEvents(rows, columns, state)
		if err != nil {
			return err
		}
		for _, cdcEvent := range events {
			if err := emit(cdcEvent); err != nil {
				return err
			}
		}
		if len(events) > 0 && req.CheckpointCommitter != nil {
			nextState := map[string]string{
				"binlog_file":                  currentFile,
				"binlog_pos":                   strconv.FormatUint(uint64(event.Header.LogPos), 10),
				mysqlCDCSchemaFingerprintState: schemaFingerprint,
			}
			if err := req.CheckpointCommitter.CommitChangefeedCheckpoint(ctx, nextState); err != nil {
				return fmt.Errorf("commit mysql binlog checkpoint: %w", err)
			}
		}
	}
}

func validateCDCQueryEvent(event *replication.QueryEvent) error {
	if event == nil {
		return errors.New("mysql CDC received an invalid statement event")
	}
	fields := strings.Fields(strings.TrimSpace(strings.TrimSuffix(string(event.Query), ";")))
	if len(fields) == 0 {
		return errors.New("mysql CDC encountered a statement event; resnapshot required")
	}
	switch strings.ToUpper(fields[0]) {
	case "BEGIN", "COMMIT", "ROLLBACK":
		return nil
	case "SET":
		statement := strings.ToUpper(strings.Join(fields, " "))
		if !strings.Contains(statement, "BINLOG_FORMAT") && !strings.Contains(statement, "BINLOG_ROW_IMAGE") {
			return nil
		}
	}
	return errors.New("mysql CDC encountered a statement event; resnapshot required")
}

func binlogPositionFromState(state map[string]string) (gomysql.Position, error) {
	if len(state) == 0 {
		return gomysql.Position{}, nil
	}
	file := strings.TrimSpace(state["binlog_file"])
	pos := strings.TrimSpace(state["binlog_pos"])
	if file == "" && pos == "" {
		return gomysql.Position{}, nil
	}
	if file == "" || pos == "" {
		return gomysql.Position{}, errors.New("mysql CDC state requires both binlog_file and binlog_pos")
	}
	parsed, err := strconv.ParseUint(pos, 10, 32)
	if err != nil || parsed < 4 {
		return gomysql.Position{}, errors.New("mysql CDC state binlog_pos must be a valid binlog position")
	}
	return gomysql.Position{Name: file, Pos: uint32(parsed)}, nil
}

func binlogCheckpointFromState(state map[string]string) (gomysql.Position, string, error) {
	position, err := binlogPositionFromState(state)
	if err != nil {
		return gomysql.Position{}, "", err
	}
	schema, schemaPresent := state[mysqlCDCSchemaFingerprintState]
	if position.Name == "" {
		if schemaPresent {
			return gomysql.Position{}, "", errors.New("mysql CDC schema fingerprint requires a binlog position")
		}
		return gomysql.Position{}, "", nil
	}
	if !schemaPresent || len(schema) != sha256.Size*2 {
		return gomysql.Position{}, "", errors.New("mysql CDC state requires schema fingerprint")
	}
	if _, err := hex.DecodeString(schema); err != nil {
		return gomysql.Position{}, "", errors.New("mysql CDC state schema fingerprint is invalid")
	}
	return position, schema, nil
}

func cdcSchemaFingerprint(columns []string) string {
	sum := sha256.Sum256([]byte(strings.Join(columns, "\x00")))
	return hex.EncodeToString(sum[:])
}

func currentBinlogPosition(ctx context.Context, conn connConfig) (gomysql.Position, error) {
	db, err := conn.open(ctx)
	if err != nil {
		return gomysql.Position{}, err
	}
	defer func() { _ = db.Close() }()
	result, err := db.Execute(currentBinlogStatusQuery)
	if err != nil {
		return gomysql.Position{}, fmt.Errorf("read mysql master status: %w", err)
	}
	records, err := resultRecords(result)
	result.Close()
	if err != nil {
		return gomysql.Position{}, err
	}
	if len(records) != 1 {
		return gomysql.Position{}, errors.New("mysql binary logging is unavailable")
	}
	file := recordCursor(records[0]["File"])
	pos, err := strconv.ParseUint(recordCursor(records[0]["Position"]), 10, 32)
	if err != nil || file == "" || pos < 4 {
		return gomysql.Position{}, errors.New("mysql binary log status returned an invalid position")
	}
	return gomysql.Position{Name: file, Pos: uint32(pos)}, nil
}

func validateBinlogRequirements(ctx context.Context, conn connConfig) error {
	db, err := conn.open(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	result, err := db.Execute(currentBinlogRequirementsQuery)
	if err != nil {
		return errors.New("read mysql CDC binlog requirements failed")
	}
	records, err := resultRecords(result)
	result.Close()
	if err != nil {
		return err
	}
	return validateBinlogRequirementRecords(records)
}

func validateBinlogRequirementRecords(records []connectors.Record) error {
	if len(records) != 1 {
		return errors.New("mysql CDC could not determine binlog requirements")
	}
	format, formatOK := recordString(records[0]["binlog_format"])
	rowImage, rowImageOK := recordString(records[0]["binlog_row_image"])
	if !formatOK || !rowImageOK || !strings.EqualFold(strings.TrimSpace(format), "ROW") || !strings.EqualFold(strings.TrimSpace(rowImage), "FULL") {
		return errors.New("mysql CDC requires binlog_format=ROW and binlog_row_image=FULL")
	}
	return nil
}

func loadColumnNames(ctx context.Context, conn connConfig, database, table string) ([]string, error) {
	db, err := conn.open(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = db.Close() }()
	result, err := db.Execute(`
SELECT column_name AS column_name
FROM information_schema.columns
WHERE table_schema = ? AND table_name = ?
ORDER BY ordinal_position`, database, table)
	if err != nil {
		return nil, fmt.Errorf("load mysql CDC column metadata: %w", err)
	}
	records, err := resultRecords(result)
	result.Close()
	if err != nil {
		return nil, err
	}
	columns := make([]string, 0, len(records))
	for _, record := range records {
		name, ok := recordString(record["column_name"])
		if !ok || validateIdentifier(name) != nil {
			return nil, errors.New("mysql CDC column metadata is invalid")
		}
		columns = append(columns, name)
	}
	if len(columns) == 0 {
		return nil, errors.New("mysql CDC stream has no columns")
	}
	return columns, nil
}

func rowsToCDCEvents(rows *replication.RowsEvent, columns []string, state connectors.Record) ([]connectors.CDCEvent, error) {
	if rows == nil {
		return nil, errors.New("mysql CDC rows event is nil")
	}
	if len(columns) == 0 {
		return nil, errors.New("mysql CDC has no column metadata")
	}
	var operation string
	switch rows.Type() {
	case replication.EnumRowsEventTypeInsert:
		operation = "insert"
	case replication.EnumRowsEventTypeUpdate:
		operation = "update"
	case replication.EnumRowsEventTypeDelete:
		operation = "delete"
	default:
		return nil, errors.New("mysql CDC received an unsupported rows event")
	}
	if rows.ColumnCount != uint64(len(columns)) {
		return nil, errors.New("mysql CDC row event does not match its column metadata")
	}
	if err := validateCDCRowImages(rows.Rows, rows.SkippedColumns, columns); err != nil {
		return nil, err
	}
	values := rows.Rows
	if operation == "update" {
		if len(values)%2 != 0 {
			return nil, errors.New("mysql CDC update event has an incomplete before/after row pair")
		}
		afterImages := make([][]any, 0, len(values)/2)
		for index := 1; index < len(values); index += 2 {
			afterImages = append(afterImages, values[index])
		}
		values = afterImages
	}
	return cdcEventsFromRows(operation, values, columns, state), nil
}

func validateCDCRowImages(rows [][]any, skippedColumns [][]int, columns []string) error {
	if len(rows) != len(skippedColumns) {
		return errors.New("mysql CDC row image metadata is incomplete")
	}
	for index, row := range rows {
		if len(row) != len(columns) {
			return errors.New("mysql CDC row does not match its column metadata")
		}
		if len(skippedColumns[index]) != 0 {
			return errors.New("mysql CDC requires complete row images")
		}
	}
	return nil
}

func cdcEventsFromRows(operation string, values [][]any, columns []string, state connectors.Record) []connectors.CDCEvent {
	events := make([]connectors.CDCEvent, 0, len(values))
	for ordinal, row := range values {
		record := make(connectors.Record, len(columns))
		for index, column := range columns {
			record[column] = copyCDCValue(row[index])
		}
		eventState := copyCDCState(state)
		eventState["binlog_row"] = strconv.Itoa(ordinal)
		events = append(events, connectors.CDCEvent{
			Operation: operation,
			Record:    record,
			State:     eventState,
		})
	}
	return events
}

func copyCDCState(state connectors.Record) connectors.Record {
	copy := make(connectors.Record, len(state))
	for key, value := range state {
		copy[key] = value
	}
	return copy
}

func copyCDCValue(value any) any {
	if bytes, ok := value.([]byte); ok {
		return append([]byte(nil), bytes...)
	}
	return value
}
