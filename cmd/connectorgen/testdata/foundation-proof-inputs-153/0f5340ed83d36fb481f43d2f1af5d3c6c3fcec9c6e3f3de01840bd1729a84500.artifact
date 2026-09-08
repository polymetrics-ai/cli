package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	duckdb "github.com/marcboeker/go-duckdb"
)

// ArrowTransformer owns one DuckDB connection for an entire transport run.
// It uses a closed TransformPlanV1 and typed Arrow views only; it has no
// connector or destination dependency. Transform serializes calls because an
// Arrow view is connection-local, while each returned reader already owns its
// result batches and can be consumed after the call returns.
type ArrowTransformer struct {
	plan TransformPlanV1
	db   *sql.DB
	conn *sql.Conn

	mu     sync.Mutex
	closed bool
}

// NewArrowTransformer opens the local vector engine once. It must be closed
// after the run, not after each bounded input range.
func NewArrowTransformer(ctx context.Context, plan TransformPlanV1) (*ArrowTransformer, error) {
	if ctx == nil || !plan.valid() {
		return nil, ErrTransformPlanInvalid
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// go-duckdb's Arrow stream callback is not safe when DuckDB fans one input
	// stream into its internal parallel scan workers (it can free the callback
	// buffer while another worker still reads it). One vectorized worker keeps
	// the Arrow ownership protocol sound; source and COPY units remain pipelined
	// by the transport controller rather than relying on unsafe intra-record
	// parallelism.
	db, err := sql.Open("duckdb", "?threads=1")
	if err != nil {
		return nil, fmt.Errorf("open reusable DuckDB transform engine: %w", err)
	}
	connection, err := db.Conn(ctx)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("open DuckDB transform connection: %w", err)
	}
	return &ArrowTransformer{plan: plan, db: db, conn: connection}, nil
}

// Close releases the one local DuckDB connection. It is idempotent so failed
// source, segment, COPY, publish, and checkpoint paths can all defer it.
func (t *ArrowTransformer) Close() error {
	if t == nil {
		return nil
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return nil
	}
	t.closed = true
	var first error
	if t.conn != nil {
		if err := t.conn.Close(); err != nil {
			first = err
		}
	}
	if t.db != nil {
		if err := t.db.Close(); err != nil && first == nil {
			first = err
		}
	}
	return first
}

// Transform evaluates one bounded input Arrow record. The returned reader
// owns its records. Call Release after consuming it. The input remains owned
// by the caller and is never decoded into maps or Go structs one row at a
// time.
func (t *ArrowTransformer) Transform(ctx context.Context, input arrow.Record) (array.RecordReader, error) {
	if t == nil || ctx == nil || input == nil || !t.plan.valid() {
		return nil, ErrTransformPlanInvalid
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	query, err := t.plan.duckDBArrowQuery(input.Schema())
	if err != nil {
		return nil, ErrTransformPlanInvalid
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed || t.conn == nil {
		return nil, ErrTransformPlanInvalid
	}

	var output array.RecordReader
	err = t.conn.Raw(func(raw any) error {
		driverConnection, ok := raw.(driver.Conn)
		if !ok {
			return ErrTransformPlanInvalid
		}
		arrowConnection, err := duckdb.NewArrowFromConn(driverConnection)
		if err != nil {
			return fmt.Errorf("open DuckDB Arrow bridge: %w", err)
		}
		inputReader, err := array.NewRecordReader(input.Schema(), []arrow.Record{input})
		if err != nil {
			return ErrTransformPlanInvalid
		}
		defer inputReader.Release()
		releaseView, err := arrowConnection.RegisterView(inputReader, "pm_input")
		if err != nil {
			return fmt.Errorf("register Arrow transform input: %w", err)
		}
		output, err = arrowConnection.QueryContext(ctx, query)
		// go-duckdb's Arrow registration owns a stream-backed view. Release and
		// remove it before another range is registered on this reusable
		// connection; retaining a prior stream makes a second Arrow scan race
		// freed callback buffers in DuckDB.
		releaseView()
		execer, ok := driverConnection.(driver.ExecerContext)
		if !ok {
			if output != nil {
				output.Release()
				output = nil
			}
			return ErrTransformPlanInvalid
		}
		if _, dropErr := execer.ExecContext(ctx, "DROP VIEW IF EXISTS pm_input", nil); dropErr != nil {
			if output != nil {
				output.Release()
				output = nil
			}
			return fmt.Errorf("release Arrow transform input: %w", dropErr)
		}
		if err != nil {
			return fmt.Errorf("evaluate closed DuckDB transform: %w", err)
		}
		return nil
	})
	if err != nil {
		if output != nil {
			output.Release()
		}
		if err == ErrTransformPlanInvalid {
			return nil, err
		}
		return nil, err
	}
	return output, nil
}

// TransformArrowRecord preserves the small single-record API for callers and
// tests. The high-throughput controller instead constructs ArrowTransformer
// once and calls Transform for each source range.
func TransformArrowRecord(ctx context.Context, plan TransformPlanV1, input arrow.Record) (array.RecordReader, error) {
	transformer, err := NewArrowTransformer(ctx, plan)
	if err != nil {
		return nil, err
	}
	defer func() { _ = transformer.Close() }()
	return transformer.Transform(ctx, input)
}

func (p TransformPlanV1) duckDBArrowQuery(schema *arrow.Schema) (string, error) {
	if schema == nil {
		return "", ErrTransformPlanInvalid
	}
	available := make(map[string]struct{}, schema.NumFields())
	for _, field := range schema.Fields() {
		available[field.Name] = struct{}{}
	}
	var document transformPlanDocument
	if err := json.Unmarshal(p.normalized, &document); err != nil {
		return "", ErrTransformPlanInvalid
	}
	columns := make([]string, len(document.Select))
	for index, projection := range document.Select {
		output, _, err := normalizeTransformProjection(projection)
		if err != nil || (output.Source != "" && !hasArrowTransformField(available, output.Source)) {
			return "", ErrTransformPlanInvalid
		}
		if output.Source != "" {
			columns[index] = quoteDuckDBTransformIdentifier(output.Source) + " AS " + quoteDuckDBTransformIdentifier(output.Target)
			continue
		}
		expression, err := renderDuckDBTransformExpression(output.Expression, available)
		if err != nil {
			return "", err
		}
		columns[index] = expression + " AS " + quoteDuckDBTransformIdentifier(output.Target)
	}
	query := "SELECT " + strings.Join(columns, ", ") + " FROM pm_input"
	if len(document.Where) != 0 {
		expression, err := renderDuckDBTransformExpression(document.Where, available)
		if err != nil {
			return "", err
		}
		query += " WHERE " + expression
	}
	return query, nil
}

func hasArrowTransformField(fields map[string]struct{}, field string) bool {
	_, ok := fields[field]
	return ok
}

func renderDuckDBTransformExpression(raw []byte, fields map[string]struct{}) (string, error) {
	expression, err := normalizeTransformExpression(raw)
	if err != nil {
		return "", ErrTransformPlanInvalid
	}
	return renderDuckDBNormalizedTransformExpression(expression.Operation, expression.JSON, fields)
}

func renderDuckDBNormalizedTransformExpression(operation string, raw []byte, fields map[string]struct{}) (string, error) {
	switch operation {
	case "upper", "date":
		field, err := arrowTransformField(raw, operation, fields)
		if err != nil {
			return "", err
		}
		if operation == "upper" {
			return "upper(" + quoteDuckDBTransformIdentifier(field) + ")", nil
		}
		// Arrow timestamps obtained from PostgreSQL timestamptz columns reach
		// DuckDB as TIMESTAMP WITH TIME ZONE. DuckDB's Arrow bridge cannot cast
		// that type directly to DATE, so the closed compiler fixes the instant
		// to a timestamp first; it remains impossible for callers to supply SQL.
		return "CAST(CAST(" + quoteDuckDBTransformIdentifier(field) + " AS TIMESTAMP) AS DATE)", nil
	case "cast":
		var node struct {
			Cast json.RawMessage `json:"cast"`
		}
		if err := json.Unmarshal(raw, &node); err != nil || len(node.Cast) == 0 {
			return "", ErrTransformPlanInvalid
		}
		nested, err := normalizeTransformExpression(node.Cast)
		if err != nil || nested.Operation != "multiply" {
			return "", ErrTransformPlanInvalid
		}
		value, err := renderDuckDBNormalizedTransformExpression(nested.Operation, nested.JSON, fields)
		if err != nil {
			return "", err
		}
		return "CAST(" + value + " AS BIGINT)", nil
	case "multiply", "mod", "not_equal":
		var node map[string][]json.RawMessage
		if err := json.Unmarshal(raw, &node); err != nil || len(node[operation]) != 2 {
			return "", ErrTransformPlanInvalid
		}
		left, err := renderDuckDBTransformOperand(node[operation][0], fields)
		if err != nil {
			return "", err
		}
		right, err := renderDuckDBTransformOperand(node[operation][1], fields)
		if err != nil {
			return "", err
		}
		switch operation {
		case "multiply":
			return "(" + left + " * " + right + ")", nil
		case "mod":
			return "(" + left + " % " + right + ")", nil
		default:
			return "(" + left + " <> " + right + ")", nil
		}
	default:
		return "", ErrTransformPlanInvalid
	}
}

func renderDuckDBTransformOperand(raw []byte, fields map[string]struct{}) (string, error) {
	var field string
	if err := json.Unmarshal(raw, &field); err == nil {
		if !hasArrowTransformField(fields, field) {
			return "", ErrTransformPlanInvalid
		}
		return quoteDuckDBTransformIdentifier(field), nil
	}
	var number int64
	if err := json.Unmarshal(raw, &number); err == nil {
		return fmt.Sprintf("%d", number), nil
	}
	if expression, err := normalizeTransformExpression(raw); err == nil {
		return renderDuckDBNormalizedTransformExpression(expression.Operation, expression.JSON, fields)
	}
	return "", ErrTransformPlanInvalid
}

func arrowTransformField(raw []byte, operation string, fields map[string]struct{}) (string, error) {
	field, err := transformExpressionField(raw, operation)
	if err != nil || !hasArrowTransformField(fields, field) {
		return "", ErrTransformPlanInvalid
	}
	return field, nil
}

func quoteDuckDBTransformIdentifier(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}
