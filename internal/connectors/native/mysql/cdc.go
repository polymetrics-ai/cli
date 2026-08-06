package mysql

import (
	"context"
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

const mysqlBinlogExecutorID = "mysql-binlog-row-v1"

// MySQL 8.4 removed SHOW MASTER STATUS with the legacy replication terms.
// SHOW BINARY LOG STATUS retains the File/Position output required for a
// position checkpoint.
const currentBinlogStatusQuery = "SHOW BINARY LOG STATUS"

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
			Keys:        []string{"binlog_file", "binlog_pos"},
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
	columns, err := loadColumnNames(ctx, conn, database, table)
	if err != nil {
		return err
	}
	position, err := binlogPositionFromState(req.State)
	if err != nil {
		return err
	}
	if position.Name == "" {
		position, err = currentBinlogPosition(ctx, conn)
		if err != nil {
			return err
		}
	}

	syncer := replication.NewBinlogSyncer(replication.BinlogSyncerConfig{
		ServerID: serverID,
		Flavor:   gomysql.MySQLFlavor,
		Host:     conn.host,
		Port:     uint16(conn.port),
		User:     conn.username,
		Password: conn.password,
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
		rows, ok := event.Event.(*replication.RowsEvent)
		if !ok || rows.Table == nil {
			continue
		}
		if string(rows.Table.Schema) != database || string(rows.Table.Table) != table {
			continue
		}
		state := connectors.Record{
			"binlog_file": currentFile,
			"binlog_pos":  strconv.FormatUint(uint64(event.Header.LogPos), 10),
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
				"binlog_file": currentFile,
				"binlog_pos":  strconv.FormatUint(uint64(event.Header.LogPos), 10),
			}
			if err := req.CheckpointCommitter.CommitChangefeedCheckpoint(ctx, nextState); err != nil {
				return fmt.Errorf("commit mysql binlog checkpoint: %w", err)
			}
		}
	}
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
	events := make([]connectors.CDCEvent, 0, len(values))
	for _, row := range values {
		if len(row) != len(columns) {
			return nil, errors.New("mysql CDC row does not match its column metadata")
		}
		record := make(connectors.Record, len(columns))
		for index, column := range columns {
			record[column] = copyCDCValue(row[index])
		}
		events = append(events, connectors.CDCEvent{
			Operation: operation,
			Record:    record,
			State:     copyCDCState(state),
		})
	}
	return events, nil
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
