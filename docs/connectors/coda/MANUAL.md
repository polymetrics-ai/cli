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
  id: coda
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
  auth_token (secret) (required)

ETL STREAMS
  docs:
    primary key: id
    fields: browserLink(string), createdAt(string), folderId(string), href(string), id(string), name(string), owner(string), ownerName(string), type(string), updatedAt(string), workspaceId(string)
  tables:
    primary key: id
    fields: browserLink(string), doc_id(string), href(string), id(string), name(string), rowCount(integer), tableType(string), type(string)
  pages:
    primary key: id
    fields: browserLink(string), contentType(string), doc_id(string), href(string), id(string), name(string), subtitle(string), type(string)
  formulas:
    primary key: id
    fields: doc_id(string), href(string), id(string), name(string), type(string)
  controls:
    primary key: id
    fields: controlType(string), doc_id(string), href(string), id(string), name(string), type(string)
  columns:
    primary key: id
    fields: calculated(boolean), defaultValue(string), display(boolean), doc_id(string), format(object), formula(string), href(string), id(string), name(string), table_id(string), type(string)
  rows:
    primary key: id
    fields: browserLink(string), createdAt(string), doc_id(string), href(string), id(string), index(integer), name(string), table_id(string), type(string), updatedAt(string), values(object)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

REVERSE ETL ACTIONS
  upsert_rows:
    endpoint: POST /docs/{{ config.doc_id }}/tables/{{ record.table_id }}/rows
    required fields: table_id, rows
    optional fields: keyColumns
    risk: inserts new rows, or upserts existing ones when keyColumns is set, into a Coda table; queued for async processing (202) and generally applied within seconds
  update_row:
    endpoint: PUT /docs/{{ config.doc_id }}/tables/{{ record.table_id }}/rows/{{ record.row_id }}
    required fields: table_id, row_id, row
    risk: overwrites cell values on an existing row; queued for async processing (202) and generally applied within seconds
  delete_row:
    endpoint: DELETE /docs/{{ config.doc_id }}/tables/{{ record.table_id }}/rows/{{ record.row_id }}
    required fields: table_id, row_id
    risk: permanently removes a row from a Coda table; irreversible, queued for async processing (202)
  delete_rows:
    endpoint: DELETE /docs/{{ config.doc_id }}/tables/{{ record.table_id }}/rows
    required fields: table_id, rowIds
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

COMMAND SURFACE
  Run Coda's declared typed write actions.
  Usage: pm coda <command> [flags]
  Global flags:
    --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
  Reverse ETL writes
  Other Commands
    create page apply - Typed action create_page [intent=reverse_etl availability=partial write=create_page]; approval: Blocked pending a faithful CLI record binding: declaration-pending: typed action path uses unrecognized placeholder {{ config.doc_id }}.; risk: creates a new page in the configured doc; requires Doc Maker access in the workspace, queued for async processing (202); notes: Generated from the connector-owned typed action; declaration-pending: typed action path uses unrecognized placeholder {{ config.doc_id }}.
    delete page apply - Typed action delete_page [intent=reverse_etl availability=partial write=delete_page]; approval: Blocked pending a faithful CLI record binding: declaration-pending: typed action path uses unrecognized placeholder {{ config.doc_id }}.; risk: permanently removes a page (and its subpages/content) from the doc; irreversible, queued for async processing (202); notes: Generated from the connector-owned typed action; declaration-pending: typed action path uses unrecognized placeholder {{ config.doc_id }}.; flags: --page-id (required)
    delete row apply - Typed action delete_row [intent=reverse_etl availability=partial write=delete_row]; approval: Blocked pending a faithful CLI record binding: declaration-pending: typed action path uses unrecognized placeholder {{ config.doc_id }}.; risk: permanently removes a row from a Coda table; irreversible, queued for async processing (202); notes: Generated from the connector-owned typed action; declaration-pending: typed action path uses unrecognized placeholder {{ config.doc_id }}.; flags: --row-id (required), --table-id (required)
    delete rows apply - Typed action delete_rows [intent=reverse_etl availability=partial write=delete_rows]; approval: Blocked pending a faithful CLI record binding: declaration-pending: typed action path uses unrecognized placeholder {{ config.doc_id }}.; risk: permanently removes multiple rows from a Coda table in one request; irreversible, queued for async processing (202); notes: Generated from the connector-owned typed action; declaration-pending: typed action path uses unrecognized placeholder {{ config.doc_id }}.; flags: --row-ids (required), --table-id (required)
    push button apply - Typed action push_button [intent=reverse_etl availability=partial write=push_button]; approval: Blocked pending a faithful CLI record binding: declaration-pending: typed action path uses unrecognized placeholder {{ config.doc_id }}.; risk: pushes a button on a row; the underlying button can perform ANY action the doc's formulas define, including writes to other tables and Pack actions outside this connector's declared surface — high blast-radius, approval required; notes: Generated from the connector-owned typed action; declaration-pending: typed action path uses unrecognized placeholder {{ config.doc_id }}.; flags: --column-id (required), --row-id (required), --table-id (required)
    update page apply - Typed action update_page [intent=reverse_etl availability=partial write=update_page]; approval: Blocked pending a faithful CLI record binding: declaration-pending: typed action path uses unrecognized placeholder {{ config.doc_id }}.; risk: renames, hides, or restyles an existing page; renaming/re-iconing requires Doc Maker access in the workspace, queued for async processing (202); notes: Generated from the connector-owned typed action; declaration-pending: typed action path uses unrecognized placeholder {{ config.doc_id }}.; flags: --page-id (required)
    update row apply - Typed action update_row [intent=reverse_etl availability=partial write=update_row]; approval: Blocked pending a faithful CLI record binding: declaration-pending: the closed CLI flag set cannot faithfully represent required record field row.cells.0.value.; risk: overwrites cell values on an existing row; queued for async processing (202) and generally applied within seconds; notes: Generated from the connector-owned typed action; declaration-pending: the closed CLI flag set cannot faithfully represent required record field row.cells.0.value.
    upsert rows apply - Typed action upsert_rows [intent=reverse_etl availability=partial write=upsert_rows]; approval: Blocked pending a faithful CLI record binding: declaration-pending: the closed CLI flag set cannot faithfully represent required record field rows.0.cells.0.value.; risk: inserts new rows, or upserts existing ones when keyColumns is set, into a Coda table; queued for async processing (202) and generally applied within seconds; notes: Generated from the connector-owned typed action; declaration-pending: the closed CLI flag set cannot faithfully represent required record field rows.0.cells.0.value.

SYNC TRANSPORT
  Source transport: declared
  Destination transport: unsupported
  A declared transport still requires runtime preflight and externally verified conformance; it is not a certification claim.
  Source executor: declarative_api/declarative_stream_source

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
