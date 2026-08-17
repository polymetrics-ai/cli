# pm connectors inspect coda

```text
NAME
  pm connectors inspect coda - Coda connector manual

SYNOPSIS
  pm connectors inspect coda
  pm connectors inspect coda --json
  pm credentials add <name> --connector coda [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Coda docs and doc-scoped tables, rows, columns, pages, formulas, and controls, and writes rows/pages, through the Coda REST API v1.

ICON
  asset: icons/coda.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://coda.io/developers/apis/v1

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  doc_id
  mode
  page_size
  auth_token (secret)

ETL STREAMS
  docs:
    primary key: id
    fields: browserLink(), createdAt(), folderId(), href(), id(), name(), owner(), ownerName(), type(), updatedAt(), workspaceId()
  tables:
    primary key: id
    fields: browserLink(), doc_id(), href(), id(), name(), rowCount(), tableType(), type()
  pages:
    primary key: id
    fields: browserLink(), contentType(), doc_id(), href(), id(), name(), subtitle(), type()
  formulas:
    primary key: id
    fields: doc_id(), href(), id(), name(), type()
  controls:
    primary key: id
    fields: controlType(), doc_id(), href(), id(), name(), type()
  columns:
    primary key: id
    fields: calculated(), defaultValue(), display(), doc_id(), format(), formula(), href(), id(), name(), table_id(), type()
  rows:
    primary key: id
    fields: browserLink(), createdAt(), doc_id(), href(), id(), index(), name(), table_id(), type(), updatedAt(), values()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  upsert_rows:
    endpoint: POST /docs/{{ config.doc_id }}/tables/{{ record.table_id }}/rows
    required fields: table_id
    optional fields: rows, keyColumns
    risk: inserts new rows, or upserts existing ones when keyColumns is set, into a Coda table; queued for async processing (202) and generally applied within seconds
  update_row:
    endpoint: PUT /docs/{{ config.doc_id }}/tables/{{ record.table_id }}/rows/{{ record.row_id }}
    required fields: table_id, row_id
    optional fields: row
    risk: overwrites cell values on an existing row; queued for async processing (202) and generally applied within seconds
  delete_row:
    endpoint: DELETE /docs/{{ config.doc_id }}/tables/{{ record.table_id }}/rows/{{ record.row_id }}
    required fields: table_id, row_id
    risk: permanently removes a row from a Coda table; irreversible, queued for async processing (202)
  delete_rows:
    endpoint: DELETE /docs/{{ config.doc_id }}/tables/{{ record.table_id }}/rows
    required fields: table_id
    optional fields: rowIds
    risk: permanently removes multiple rows from a Coda table in one request; irreversible, queued for async processing (202)
  push_button:
    endpoint: POST /docs/{{ config.doc_id }}/tables/{{ record.table_id }}/rows/{{ record.row_id }}/buttons/{{ record.column_id }}
    required fields: table_id, row_id, column_id
    risk: pushes a button on a row; the underlying button can perform ANY action the doc's formulas define, including writes to other tables and Pack actions outside this connector's declared surface — high blast-radius, approval required
  create_page:
    endpoint: POST /docs/{{ config.doc_id }}/pages
    risk: creates a new page in the configured doc; requires Doc Maker access in the workspace, queued for async processing (202)
  update_page:
    endpoint: PUT /docs/{{ config.doc_id }}/pages/{{ record.page_id }}
    required fields: page_id
    risk: renames, hides, or restyles an existing page; renaming/re-iconing requires Doc Maker access in the workspace, queued for async processing (202)
  delete_page:
    endpoint: DELETE /docs/{{ config.doc_id }}/pages/{{ record.page_id }}
    required fields: page_id
    risk: permanently removes a page (and its subpages/content) from the doc; irreversible, queued for async processing (202)

SECURITY
  read risk: external Coda API read of docs and doc-scoped tables, rows, columns, pages, formulas, and controls
  write risk: external mutation of Coda table rows and doc pages (insert/upsert/update/delete rows, push a row button, create/update/delete a page); push_button and delete actions are approval-gated per writes.json risk text
  approval: row/page create+update: none; delete_row/delete_rows/delete_page: approval required (irreversible); push_button: approval required (arbitrary doc-defined side effects)
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect coda

  # Inspect as structured JSON
  pm connectors inspect coda --json

AGENT WORKFLOW
  - Run pm connectors inspect coda before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
