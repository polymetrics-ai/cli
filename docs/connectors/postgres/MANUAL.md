# pm connectors inspect postgres

```text
NAME
  pm connectors inspect postgres - PostgreSQL connector manual

SYNOPSIS
  pm connectors inspect postgres
  pm connectors inspect postgres --json
  pm credentials add <name> --connector postgres [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads PostgreSQL tables by discovering schemas/columns from information_schema, snapshots tables with bounded cursor-incremental reads, and supports fixture-safe bounded row DML/truncate reverse ETL actions. CDC is a documented stub pending the gated pglogrepl dependency.

ICON
  asset: icons/postgresql.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://www.postgresql.org/docs/

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: database

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  host (required): Bare hostname or IP of the PostgreSQL server; no scheme, path, or credentials.
  port: TCP port, 1-65535. Defaults to 5432.
  database (required): Database name to connect to.
  username (required): Database role used to authenticate.
  sslmode: libpq sslmode: disable, allow, prefer, require, verify-ca, or verify-full. Defaults to disable.
  schema: Default schema for discovery/read/write actions. Defaults to public.
  cursor_field: Optional column name used for incremental reads.
  read_limit: Maximum rows returned per snapshot SELECT; defaults to 10000. A smaller request limit wins.
  mode: Set to fixture for test/conformance replay without network access.
  password (secret) (required): Database role password. Never logged.

SYNC MODES
  Source modes: full_refresh, incremental

REVERSE ETL ACTIONS
  insert_row: Insert one row with typed column values.
    endpoint: SQL INSERT
    required fields: table, values
    optional fields: schema
    risk: high: inserts a row into the selected table
  update_row: Update rows selected by explicit key fields.
    endpoint: SQL UPDATE
    required fields: table, values, keys
    optional fields: schema
    risk: high: updates rows matching supplied keys
  upsert_row: Insert or update one row through a declared conflict key.
    endpoint: SQL MERGE
    required fields: table, values, keys
    optional fields: schema
    risk: high: bounded upsert; no raw MERGE text
  delete_row: Delete rows selected by explicit key fields.
    endpoint: SQL DELETE
    required fields: table, keys
    optional fields: schema
    risk: critical: deletes rows matching supplied keys
  truncate_table: Truncate one table only when confirm_phrase=truncate is supplied.
    endpoint: SQL TRUNCATE
    required fields: table, confirm_phrase
    optional fields: schema
    risk: critical: truncates the selected table; cascade/restart identity are not exposed

SECURITY
  read risk: low: schema/table/cursor identifiers are validated and values are bound parameters
  write risk: high/critical: bounded row DML/truncate only; no arbitrary SQL; destructive actions require approval
  approval: reverse ETL writes require plan -> preview -> approval -> execute; truncate_table requires confirm_phrase=truncate
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Inspect PostgreSQL schemas and run bounded reverse-ETL row actions without raw SQL.
  Usage: pm postgres <command> [flags]
  Source CLI: psql (https://www.postgresql.org/docs/current/sql-commands.html)
  Global flags:
    --json (boolean): Write machine-readable JSON output.
    --connection (string): Use a saved PostgreSQL connector credential.: maps_to=connection
  Bounded row writes
    row insert - Insert one row with typed string or integer column flags. [intent=reverse_etl availability=implemented write=insert_row]; approval: reverse ETL writes require plan -> preview -> approval -> execute.; risk: Inserts a row into the configured table using typed fields only; no SQL text is accepted.; flags: --schema, --table, --column, --value, --value-int
    row update - Update row values selected by an explicit key column. [intent=reverse_etl availability=implemented write=update_row]; approval: reverse ETL writes require plan -> preview -> approval -> execute.; risk: Updates rows matching explicit key fields. Missing keys are rejected.; flags: --schema, --table, --column, --value, --value-int, --key-column, --key-value, --key-value-int
    row upsert - Insert or update one row through a declared conflict key. [intent=reverse_etl availability=implemented write=upsert_row]; approval: reverse ETL writes require plan -> preview -> approval -> execute.; risk: Performs a bounded row upsert; it does not expose arbitrary MERGE or INSERT text.; flags: --schema, --table, --key-column, --key-value, --key-value-int, --column, --value, --value-int
    row delete - Delete rows selected by explicit key columns. [intent=reverse_etl availability=implemented write=delete_row]; approval: Requires destructive approval through plan -> preview -> approval -> execute.; risk: Destructive: deletes rows matching the supplied key values. Missing keys are rejected.; flags: --schema, --table, --key-column, --key-value, --key-value-int
    table truncate - Truncate one table with typed confirmation. [intent=reverse_etl availability=implemented write=truncate_table]; approval: Requires destructive approval and confirm-phrase truncate through plan -> preview -> approval -> execute.; risk: Destructive: truncates only the named table. CASCADE and RESTART IDENTITY are not exposed.; flags: --schema, --table, --confirm-phrase
  Help topics:
    safety - PostgreSQL commands intentionally do not expose raw SQL, COPY streams, shell escapes, extension APIs, or arbitrary protocol messages. Use catalog/read for ETL and bounded row/table write actions for reverse ETL.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect postgres

  # Inspect as structured JSON
  pm connectors inspect postgres --json

AGENT WORKFLOW
  - Run pm connectors inspect postgres before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
