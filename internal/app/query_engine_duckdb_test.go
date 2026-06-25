//go:build duckdb

package app_test

import (
	"context"
	"encoding/json"
	"testing"

	"polymetrics.ai/internal/app"
)

// TestDuckDBJoinAndAggregate is the red-first test for the DuckDB-backed query
// engine (only built with -tags duckdb). It proves real analytical SQL — a JOIN
// plus GROUP BY aggregation over two warehouse tables.
func TestDuckDBJoinAndAggregate(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject: %v", err)
	}
	a, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if got := a.QueryEngineName(); got != "duckdb" {
		t.Fatalf("QueryEngineName() = %q, want duckdb (-tags duckdb build)", got)
	}

	seedWarehouseTable(t, root, "customers", []map[string]any{
		{"customer_id": "c1", "name": "Ada"},
		{"customer_id": "c2", "name": "Grace"},
	})
	seedWarehouseTable(t, root, "orders", []map[string]any{
		{"order_id": "o1", "customer_id": "c1", "amount": 100},
		{"order_id": "o2", "customer_id": "c1", "amount": 50},
		{"order_id": "o3", "customer_id": "c2", "amount": 70},
	})

	rows, err := a.QuerySQL(ctx, `
		SELECT c.name AS name, SUM(o.amount) AS total
		FROM orders o JOIN customers c USING (customer_id)
		GROUP BY c.name ORDER BY total DESC`, 10)
	if err != nil {
		t.Fatalf("QuerySQL(join+aggregate): %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("rows = %d, want 2", len(rows))
	}
	// Ada: 150, Grace: 70 (ordered desc).
	if name := toString(rows[0]["name"]); name != "Ada" {
		t.Fatalf("row0 name = %q, want Ada", name)
	}
	if total := toFloat(rows[0]["total"]); total != 150 {
		t.Fatalf("row0 total = %v, want 150", rows[0]["total"])
	}
	if total := toFloat(rows[1]["total"]); total != 70 {
		t.Fatalf("row1 total = %v, want 70", rows[1]["total"])
	}
}

// TestDuckDBSelectOnlyRejectsMutation asserts the engine refuses non-SELECT SQL.
func TestDuckDBSelectOnlyRejectsMutation(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject: %v", err)
	}
	a, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	seedWarehouseTable(t, root, "orders", []map[string]any{{"order_id": "o1", "amount": 1}})

	for _, bad := range []string{
		"INSERT INTO orders VALUES ('x', 2)",
		"DROP VIEW orders",
		"select * from orders; drop view orders",
	} {
		if _, err := a.QuerySQL(ctx, bad, 10); err == nil {
			t.Errorf("QuerySQL(%q) = nil error, want rejection", bad)
		}
	}
}

func toString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func toFloat(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int64:
		return float64(n)
	case int:
		return float64(n)
	case json.Number:
		f, _ := n.Float64()
		return f
	default:
		return -1
	}
}
