package postgres

import (
	"context"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func testFixtureConfig() connectors.RuntimeConfig {
	return connectors.RuntimeConfig{
		Config: map[string]string{
			"mode":     "fixture",
			"host":     "db.internal",
			"database": "analytics",
			"username": "writer",
			"sslmode":  "require",
			"schema":   "public",
		},
		Secrets: map[string]string{"password": "fixture-password"},
	}
}

func TestPostgresWriteValidationRejectsArbitrarySQLAndUnsafeIdentifiers(t *testing.T) {
	c := New()
	ctx := context.Background()

	valid := []connectors.Record{{
		"schema": "public",
		"table":  "customers",
		"values": []any{
			map[string]any{"name": "email", "value_string": "customer@example.invalid"},
		},
	}}
	if err := c.ValidateWrite(ctx, connectors.WriteRequest{Action: "insert_row", Config: testFixtureConfig()}, valid); err != nil {
		t.Fatalf("ValidateWrite(valid insert) = %v", err)
	}

	cases := []struct {
		name    string
		action  string
		records []connectors.Record
		want    string
	}{
		{
			name:   "unknown action blocks arbitrary sql",
			action: "execute_sql",
			records: []connectors.Record{{
				"sql": "DROP TABLE public.customers",
			}},
			want: "unsupported postgres write action",
		},
		{
			name:   "unsafe table identifier",
			action: "insert_row",
			records: []connectors.Record{{
				"table":  "customers;drop",
				"values": []any{map[string]any{"name": "email", "value_string": "customer@example.invalid"}},
			}},
			want: "table",
		},
		{
			name:   "missing delete keys",
			action: "delete_row",
			records: []connectors.Record{{
				"table": "customers",
			}},
			want: "keys",
		},
		{
			name:   "truncate requires typed confirm phrase",
			action: "truncate_table",
			records: []connectors.Record{{
				"table": "customers",
			}},
			want: "confirm_phrase",
		},
		{
			name:   "value item is closed typed schema",
			action: "insert_row",
			records: []connectors.Record{{
				"table":  "customers",
				"values": []any{map[string]any{"name": "email", "value_string": "customer@example.invalid", "sql": "raw"}},
			}},
			want: "unsupported field",
		},
		{
			name:   "integer float must fit int64",
			action: "insert_row",
			records: []connectors.Record{{
				"table":  "customers",
				"values": []any{map[string]any{"name": "id", "value_int": 1e40}},
			}},
			want: "fit int64",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := c.ValidateWrite(ctx, connectors.WriteRequest{Action: tc.action, Config: testFixtureConfig()}, tc.records)
			if err == nil {
				t.Fatalf("ValidateWrite succeeded, want error containing %q", tc.want)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("ValidateWrite error = %v, want substring %q", err, tc.want)
			}
		})
	}
}

func TestPostgresWriteBuildsParameterizedSQL(t *testing.T) {
	cases := []struct {
		name       string
		action     string
		record     connectors.Record
		wantSQL    string
		wantArgs   int
		forbidText string
	}{
		{
			name:   "insert",
			action: "insert_row",
			record: connectors.Record{"schema": "public", "table": "customers", "values": []any{
				map[string]any{"name": "email", "value_string": "customer@example.invalid"},
				map[string]any{"name": "active", "value_bool": true},
			}},
			wantSQL:    `INSERT INTO "public"."customers" ("email", "active") VALUES ($1, $2)`,
			wantArgs:   2,
			forbidText: "customer@example.invalid",
		},
		{
			name:   "update",
			action: "update_row",
			record: connectors.Record{"schema": "public", "table": "customers", "values": []any{
				map[string]any{"name": "active", "value_bool": false},
			}, "keys": []any{
				map[string]any{"name": "id", "value_int": int64(42)},
			}},
			wantSQL:  `UPDATE "public"."customers" SET "active" = $1 WHERE "id" = $2`,
			wantArgs: 2,
		},
		{
			name:   "upsert",
			action: "upsert_row",
			record: connectors.Record{"schema": "public", "table": "customers", "values": []any{
				map[string]any{"name": "id", "value_int": int64(42)},
				map[string]any{"name": "email", "value_string": "customer@example.invalid"},
			}, "keys": []any{
				map[string]any{"name": "id", "value_int": int64(42)},
			}},
			wantSQL:    `MERGE INTO "public"."customers" AS target USING (VALUES ($1, $2)) AS source ("id", "email") ON target."id" = source."id" WHEN MATCHED THEN UPDATE SET "email" = source."email" WHEN NOT MATCHED THEN INSERT ("id", "email") VALUES (source."id", source."email")`,
			wantArgs:   2,
			forbidText: "customer@example.invalid",
		},
		{
			name:   "delete",
			action: "delete_row",
			record: connectors.Record{"schema": "public", "table": "customers", "keys": []any{
				map[string]any{"name": "id", "value_int": int64(42)},
			}},
			wantSQL:  `DELETE FROM "public"."customers" WHERE "id" = $1`,
			wantArgs: 1,
		},
		{
			name:     "truncate",
			action:   "truncate_table",
			record:   connectors.Record{"schema": "public", "table": "customers", "confirm_phrase": "truncate"},
			wantSQL:  `TRUNCATE TABLE ONLY "public"."customers"`,
			wantArgs: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stmt, err := buildWriteStatement("public", tc.action, tc.record)
			if err != nil {
				t.Fatalf("buildWriteStatement(%s): %v", tc.action, err)
			}
			if stmt.SQL != tc.wantSQL {
				t.Fatalf("SQL = %q, want %q", stmt.SQL, tc.wantSQL)
			}
			if len(stmt.Args) != tc.wantArgs {
				t.Fatalf("len(args) = %d, want %d (%+v)", len(stmt.Args), tc.wantArgs, stmt.Args)
			}
			if tc.forbidText != "" && strings.Contains(stmt.SQL, tc.forbidText) {
				t.Fatalf("SQL inlined sensitive value %q: %s", tc.forbidText, stmt.SQL)
			}
		})
	}
}

func TestPostgresFixtureWriteDryRunAndExecute(t *testing.T) {
	c := New()
	ctx := context.Background()
	req := connectors.WriteRequest{Action: "delete_row", Config: testFixtureConfig()}
	records := []connectors.Record{{
		"schema": "public",
		"table":  "customers",
		"keys": []any{
			map[string]any{"name": "id", "value_int": int64(42)},
		},
	}}

	preview, err := c.DryRunWrite(ctx, req, records)
	if err != nil {
		t.Fatalf("DryRunWrite: %v", err)
	}
	if preview.RecordsStaged != 1 || preview.Action != "delete_row" {
		t.Fatalf("preview = %+v, want one staged delete_row", preview)
	}
	joined := strings.Join(preview.Warnings, "\n")
	if !strings.Contains(joined, `DELETE FROM "public"."customers" WHERE "id" = $1`) {
		t.Fatalf("preview warnings did not include redacted SQL template: %+v", preview.Warnings)
	}
	if strings.Contains(joined, "42") {
		t.Fatalf("preview warning leaked bound value: %+v", preview.Warnings)
	}

	result, err := c.Write(ctx, req, records)
	if err != nil {
		t.Fatalf("Write(fixture): %v", err)
	}
	if result.RecordsWritten != 1 || result.RecordsFailed != 0 {
		t.Fatalf("result = %+v, want one written and no failures", result)
	}
}

func TestPostgresReadUsesRequestLimitBeforeConfiguredLimit(t *testing.T) {
	c := New()
	cfg := testFixtureConfig()
	cfg.Config["read_limit"] = "100"
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "public.users", Config: cfg, Limit: 1}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read fixture with request limit: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("Read emitted %d fixture rows, want request limit 1", len(got))
	}
}
