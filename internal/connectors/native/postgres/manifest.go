package postgres

import "polymetrics.ai/internal/connectors"

// Manifest exposes the native connector's runtime config and bounded write
// actions to docs, CLI inspection, and command-runner planning. engine.Base
// provides Definition()/CommandSurface() from defs/postgres, but Manifest is
// not part of Base today, so the Tier-3 connector publishes the equivalent
// user-facing contract explicitly.
func (c Connector) Manifest() connectors.Manifest {
	return connectors.Manifest{
		Metadata: c.Metadata(),
		ConfigFields: []connectors.ConfigField{
			{Name: "host", Description: "Bare hostname or IP of the PostgreSQL server; no scheme, path, or credentials.", Required: true},
			{Name: "port", Description: "TCP port, 1-65535. Defaults to 5432."},
			{Name: "database", Description: "Database name to connect to.", Required: true},
			{Name: "username", Description: "Database role used to authenticate.", Required: true},
			{Name: "sslmode", Description: "libpq sslmode: disable, allow, prefer, require, verify-ca, or verify-full. Defaults to disable."},
			{Name: "schema", Description: "Default schema for discovery/read/write actions. Defaults to public."},
			{Name: "cursor_field", Description: "Optional column name used for incremental reads."},
			{Name: "read_limit", Description: "Maximum rows returned per snapshot SELECT; defaults to 10000. A smaller request limit wins."},
			{Name: "mode", Description: "Set to fixture for test/conformance replay without network access."},
		},
		SecretFields: []connectors.SecretField{
			{Name: "password", Description: "Database role password. Never logged.", Required: true},
		},
		WriteActions: []connectors.WriteActionSpec{
			{Name: "insert_row", Description: "Insert one row with typed column values.", RequiredFields: []string{"table", "values"}, OptionalFields: []string{"schema"}, Method: "SQL", Path: "INSERT", Risk: "high: inserts a row into the selected table"},
			{Name: "update_row", Description: "Update rows selected by explicit key fields.", RequiredFields: []string{"table", "values", "keys"}, OptionalFields: []string{"schema"}, Method: "SQL", Path: "UPDATE", Risk: "high: updates rows matching supplied keys"},
			{Name: "upsert_row", Description: "Insert or update one row through a declared conflict key.", RequiredFields: []string{"table", "values", "keys"}, OptionalFields: []string{"schema"}, Method: "SQL", Path: "MERGE", Risk: "high: bounded upsert; no raw MERGE text"},
			{Name: "delete_row", Description: "Delete rows selected by explicit key fields.", RequiredFields: []string{"table", "keys"}, OptionalFields: []string{"schema"}, Method: "SQL", Path: "DELETE", Risk: "critical: deletes rows matching supplied keys", Confirm: "destructive"},
			{Name: "truncate_table", Description: "Truncate one table only when confirm_phrase=truncate is supplied.", RequiredFields: []string{"table", "confirm_phrase"}, OptionalFields: []string{"schema"}, Method: "SQL", Path: "TRUNCATE", Risk: "critical: truncates the selected table; cascade/restart identity are not exposed", Confirm: "destructive"},
		},
		SourceSyncModes: []string{"full_refresh", "incremental"},
		Risk: connectors.RiskSpec{
			Read:     "low: schema/table/cursor identifiers are validated and values are bound parameters",
			Write:    "high/critical: bounded row DML/truncate only; no arbitrary SQL; destructive actions require approval",
			Approval: "reverse ETL writes require plan -> preview -> approval -> execute; truncate_table requires confirm_phrase=truncate",
		},
	}
}
