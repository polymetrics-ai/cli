---
name: pm-notion
description: Notion connector knowledge and safe action guide.
---

# pm-notion

## Purpose

Reads and writes Notion pages, databases, data sources, blocks, comments, views, and file uploads through the Notion REST API.

## Icon

- id: notion
- asset: icons/notion.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.notion.com/reference/changes-by-version

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- max_pages
- page_size
- token (secret)

## ETL Streams

- databases:
  - primary key: id
  - cursor: last_edited_time
  - fields: archived(boolean), created_time(string), id(string), in_trash(boolean), last_edited_time(string), object(string), parent(object), title(array), url(string)
- pages:
  - primary key: id
  - cursor: last_edited_time
  - fields: archived(boolean), created_time(string), id(string), in_trash(boolean), last_edited_time(string), object(string), parent(object), properties(object), url(string)
- users:
  - primary key: id
  - fields: avatar_url(string), bot(object), id(string), name(string), object(string), person(object), type(string)
- custom_emojis:
  - primary key: id
  - fields: id(string), name(string), object(string), url(string)
- file_uploads:
  - primary key: id
  - cursor: last_edited_time
  - fields: content_length(integer), content_type(string), created_time(string), expiry_time(string), filename(string), id(string), last_edited_time(string), number_of_parts(object), object(string), status(string), upload_url(string)
- views:
  - primary key: id
  - cursor: last_edited_time
  - fields: created_time(string), data_source_id(string), id(string), last_edited_time(string), name(string), object(string), parent(object), type(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_meeting_note:
  - endpoint: POST /v1/blocks/meeting_notes
  - optional fields: language, options, title
  - risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.
- update_block:
  - endpoint: PATCH /v1/blocks/{{ record.block_id }}
  - required fields: block_id
  - optional fields: in_trash
  - risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.
- delete_block:
  - endpoint: DELETE /v1/blocks/{{ record.block_id }}
  - required fields: block_id
  - risk: high — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.
- append_block_children:
  - endpoint: PATCH /v1/blocks/{{ record.block_id }}/children
  - required fields: block_id, children
  - optional fields: position
  - risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.
- create_comment:
  - endpoint: POST /v1/comments
  - optional fields: attachments, display_name
  - risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.
- update_comment:
  - endpoint: PATCH /v1/comments/{{ record.comment_id }}
  - required fields: comment_id, rich_text
  - risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.
- update_comment_markdown:
  - endpoint: PATCH /v1/comments/{{ record.comment_id }}
  - required fields: comment_id, markdown
  - risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.
- delete_comment:
  - endpoint: DELETE /v1/comments/{{ record.comment_id }}
  - required fields: comment_id
  - risk: high — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.
- create_data_source:
  - endpoint: POST /v1/data_sources
  - required fields: parent, properties
  - optional fields: icon, title
  - risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.
- update_data_source:
  - endpoint: PATCH /v1/data_sources/{{ record.data_source_id }}
  - required fields: data_source_id
  - optional fields: icon, in_trash, parent, properties, title
  - risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.
- create_database:
  - endpoint: POST /v1/databases
  - required fields: parent
  - optional fields: cover, description, icon, initial_data_source, is_inline, title
  - risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.
- update_database:
  - endpoint: PATCH /v1/databases/{{ record.database_id }}
  - required fields: database_id
  - optional fields: cover, description, icon, in_trash, is_inline, is_locked, parent, title
  - risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.
- create_page:
  - endpoint: POST /v1/pages
  - optional fields: allow_async, children, content, cover, icon, markdown, parent, position, properties, template
  - risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.
- update_page:
  - endpoint: PATCH /v1/pages/{{ record.page_id }}
  - required fields: page_id
  - optional fields: cover, erase_content, icon, in_trash, is_archived, is_locked, properties, template
  - risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.
- update_page_markdown:
  - endpoint: PATCH /v1/pages/{{ record.page_id }}/markdown
  - required fields: page_id
  - optional fields: allow_async
  - risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.
- move_page:
  - endpoint: POST /v1/pages/{{ record.page_id }}/move
  - required fields: page_id, parent
  - risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.
- move_page_to_data_source:
  - endpoint: POST /v1/pages/{{ record.page_id }}/move
  - required fields: page_id, parent
  - risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.
- create_view:
  - endpoint: POST /v1/views
  - required fields: data_source_id, name, type
  - optional fields: configuration, create_database, database_id, filter, placement, position, quick_filters, sorts, view_id
  - risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.
- update_view:
  - endpoint: PATCH /v1/views/{{ record.view_id }}
  - required fields: view_id
  - optional fields: configuration, filter, name, quick_filters, sorts
  - risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.
- delete_view:
  - endpoint: DELETE /v1/views/{{ record.view_id }}
  - required fields: view_id
  - risk: high — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.
- create_view_query:
  - endpoint: POST /v1/views/{{ record.view_id }}/queries
  - required fields: view_id
  - optional fields: page_size
  - risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.
- delete_view_query:
  - endpoint: DELETE /v1/views/{{ record.view_id }}/queries/{{ record.query_id }}
  - required fields: view_id, query_id
  - risk: high — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.
- create_file_upload:
  - endpoint: POST /v1/file_uploads
  - optional fields: content_type, external_url, filename, mode, number_of_parts
  - risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.
- complete_file_upload:
  - endpoint: POST /v1/file_uploads/{{ record.file_upload_id }}/complete
  - required fields: file_upload_id
  - risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.

## Security

- read risk: external Notion API read of workspace databases/pages/users
- approval: none; read-only source connector
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run Notion's declared streams and reverse-ETL actions.
- Usage: pm notion <command> [flags]
- Read streams
- Reverse ETL writes
- Other Commands
  - api get v1 databases - Documented GET /v1/databases (not implemented) [intent=direct_read availability=not_implemented operation=notion.get.v1-databases]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api post v1 databases database-id query - Documented POST /v1/databases/{database_id}/query (not implemented) [intent=direct_write availability=not_implemented operation=notion.post.v1-databases-database-id-query]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: low; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 file-uploads file-upload-id send - Documented POST /v1/file_uploads/{file_upload_id}/send (not implemented) [intent=direct_write availability=not_implemented operation=notion.post.v1-file-uploads-file-upload-id-send]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 oauth introspect - Documented POST /v1/oauth/introspect (not implemented) [intent=direct_write availability=not_implemented operation=notion.post.v1-oauth-introspect]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 oauth revoke - Documented POST /v1/oauth/revoke (not implemented) [intent=direct_write availability=not_implemented operation=notion.post.v1-oauth-revoke]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 oauth token - Documented POST /v1/oauth/token (not implemented) [intent=direct_write availability=not_implemented operation=notion.post.v1-oauth-token]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 search - Documented POST /v1/search (not implemented) [intent=direct_write availability=not_implemented operation=notion.post.v1-search]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - async-task get - Retrieve an async task [intent=direct_read availability=implemented operation=notion.retrieve_async_task]; notes: Bounded Notion read; fixed method and path with typed request fields.; flags: --task-id (required), --page, --page-cursor
  - block children append - Append block children [intent=reverse_etl availability=partial write=append_block_children]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.; notes: Reverse ETL from a source record is the supported path: children is an array over Notion's 31-arm block-type union; no typed scalar leaf exists for a flag contract; flags: --block-id (required)
  - block children list - Retrieve block children [intent=direct_read availability=implemented operation=notion.retrieve_block_children]; notes: Bounded Notion read; fixed method and path with typed request fields.; flags: --block-id (required), --page-size, --page, --page-cursor
  - block delete - Delete a block [intent=reverse_etl availability=implemented write=delete_block]; approval: requires plan, preview, approval, and execute; risk: high — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.; flags: --block_id (required)
  - block get - Retrieve a block [intent=direct_read availability=implemented operation=notion.retrieve_block]; notes: Bounded Notion read; fixed method and path with typed request fields.; flags: --block-id (required), --page, --page-cursor
  - block update - Update a block [intent=reverse_etl availability=implemented write=update_block]; approval: requires plan, preview, approval, and execute; risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.; flags: --block_id (required)
  - comment create - Create a comment [intent=reverse_etl availability=implemented write=create_comment]; approval: requires plan, preview, approval, and execute; risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.
  - comment delete - Delete a comment [intent=reverse_etl availability=implemented write=delete_comment]; approval: requires plan, preview, approval, and execute; risk: high — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.; flags: --comment_id (required)
  - comment get - Retrieve a comment [intent=direct_read availability=implemented operation=notion.retrieve_comment]; notes: Bounded Notion read; fixed method and path with typed request fields.; flags: --comment-id (required), --page, --page-cursor
  - comment list - List comments [intent=direct_read availability=implemented operation=notion.list_comments]; notes: Bounded Notion read; fixed method and path with typed request fields.; flags: --block-id (required), --page-size, --page, --page-cursor
  - comment update - Update a comment [intent=reverse_etl availability=partial write=update_comment]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.; notes: Reverse ETL from a source record is the supported path: rich_text is an array of rich-text objects with no scalar leaf; supply it from a reverse-ETL source record; flags: --comment-id (required)
  - comment update-markdown - Update a comment [intent=reverse_etl availability=implemented write=update_comment_markdown]; approval: requires plan, preview, approval, and execute; risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.; flags: --comment_id (required), --markdown (required)
  - custom-emojis list - List Notion custom emojis as ETL records. [intent=etl availability=implemented stream=custom_emojis]
  - data-source create - Create a data source [intent=reverse_etl availability=partial write=create_data_source]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.; notes: Reverse ETL from a source record is the supported path: properties is a free-form map of property definitions with no declared required scalar leaf; flags: --parent-database-id (required)
  - data-source get - Retrieve a data source [intent=direct_read availability=implemented operation=notion.retrieve_data_source]; notes: Bounded Notion read; fixed method and path with typed request fields.; flags: --data-source-id (required), --page, --page-cursor
  - data-source query - Query a data source [intent=direct_read availability=implemented operation=notion.query_data_source]; notes: Bounded Notion read; fixed method and path with typed request fields.; flags: --data-source-id (required), --filter-properties, --page-size, --is-archived, --result-type, --page, --page-cursor
  - data-source templates list - List templates in a data source [intent=direct_read availability=implemented operation=notion.list_data_source_templates]; notes: Bounded Notion read; fixed method and path with typed request fields.; flags: --data-source-id (required), --name, --page-size, --page, --page-cursor
  - data-source update - Update a data source [intent=reverse_etl availability=implemented write=update_data_source]; approval: requires plan, preview, approval, and execute; risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.; flags: --data_source_id (required)
  - database create - Create a database [intent=reverse_etl availability=implemented write=create_database]; approval: requires plan, preview, approval, and execute; risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.; flags: --parent (required)
  - database get - Retrieve a database [intent=direct_read availability=implemented operation=notion.retrieve_database]; notes: Bounded Notion read; fixed method and path with typed request fields.; flags: --database-id (required), --page, --page-cursor
  - database update - Update a database [intent=reverse_etl availability=implemented write=update_database]; approval: requires plan, preview, approval, and execute; risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.; flags: --database_id (required)
  - databases list - List Notion databases as ETL records. [intent=etl availability=implemented stream=databases]
  - file-upload complete - Complete a multi-part file upload [intent=reverse_etl availability=implemented write=complete_file_upload]; approval: requires plan, preview, approval, and execute; risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.; flags: --file_upload_id (required)
  - file-upload create - Create a file upload [intent=reverse_etl availability=implemented write=create_file_upload]; approval: requires plan, preview, approval, and execute; risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.
  - file-upload get - Retrieve a file upload [intent=direct_read availability=implemented operation=notion.retrieve_file_upload]; notes: Bounded Notion read; fixed method and path with typed request fields.; flags: --file-upload-id (required), --page, --page-cursor
  - file-uploads list - List Notion file uploads as ETL records. [intent=etl availability=implemented stream=file_uploads]
  - meeting-note create - Create a meeting note [intent=reverse_etl availability=implemented write=create_meeting_note]; approval: requires plan, preview, approval, and execute; risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.
  - meeting-note query - Query meeting notes [intent=direct_read availability=implemented operation=notion.query_meeting_notes]; notes: Bounded Notion read; fixed method and path with typed request fields.; flags: --limit, --page, --page-cursor
  - page create - Create a page [intent=reverse_etl availability=implemented write=create_page]; approval: requires plan, preview, approval, and execute; risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.
  - page get - Retrieve a page [intent=direct_read availability=implemented operation=notion.retrieve_page]; notes: Bounded Notion read; fixed method and path with typed request fields.; flags: --page-id (required), --filter-properties, --page, --page-cursor
  - page markdown get - Retrieve a page as markdown [intent=direct_read availability=implemented operation=notion.retrieve_page_markdown]; notes: Bounded Notion read; fixed method and path with typed request fields.; flags: --page-id (required), --include-transcript, --page, --page-cursor
  - page markdown update - Update a page's content as markdown [intent=reverse_etl availability=implemented write=update_page_markdown]; approval: requires plan, preview, approval, and execute; risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.; flags: --page_id (required)
  - page move - Move a page [intent=reverse_etl availability=not_implemented write=move_page]; approval: requires plan, preview, approval, and execute; risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - page move-to-data-source - Move a page [intent=reverse_etl availability=not_implemented write=move_page_to_data_source]; approval: requires plan, preview, approval, and execute; risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - page property get - Retrieve a page property item [intent=direct_read availability=implemented operation=notion.retrieve_page_property]; notes: Bounded Notion read; fixed method and path with typed request fields.; flags: --page-id (required), --property-id (required), --page-size, --page, --page-cursor
  - page update - Update page [intent=reverse_etl availability=implemented write=update_page]; approval: requires plan, preview, approval, and execute; risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.; flags: --page_id (required)
  - pages list - List Notion pages as ETL records. [intent=etl availability=implemented stream=pages]
  - user get - Retrieve a user [intent=direct_read availability=implemented operation=notion.retrieve_user]; notes: Bounded Notion read; fixed method and path with typed request fields.; flags: --user-id (required), --page, --page-cursor
  - user me get - Retrieve your token's bot user [intent=direct_read availability=implemented operation=notion.retrieve_bot_user]; notes: Bounded Notion read; fixed method and path with typed request fields.; flags: --page, --page-cursor
  - users list - List Notion users as ETL records. [intent=etl availability=implemented stream=users]
  - view create - Create a view [intent=reverse_etl availability=implemented write=create_view]; approval: requires plan, preview, approval, and execute; risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.; flags: --data_source_id (required), --name (required), --type (required)
  - view delete - Delete a view [intent=reverse_etl availability=implemented write=delete_view]; approval: requires plan, preview, approval, and execute; risk: high — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.; flags: --view_id (required)
  - view get - Retrieve a view [intent=direct_read availability=implemented operation=notion.retrieve_view]; notes: Bounded Notion read; fixed method and path with typed request fields.; flags: --view-id (required), --page, --page-cursor
  - view query create - Create a view query [intent=reverse_etl availability=implemented write=create_view_query]; approval: requires plan, preview, approval, and execute; risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.; flags: --view_id (required)
  - view query delete - Delete a view query [intent=reverse_etl availability=implemented write=delete_view_query]; approval: requires plan, preview, approval, and execute; risk: high — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.; flags: --query_id (required), --view_id (required)
  - view query results get - Get view query results [intent=direct_read availability=implemented operation=notion.retrieve_view_query_results]; notes: Bounded Notion read; fixed method and path with typed request fields.; flags: --view-id (required), --query-id (required), --page-size, --page, --page-cursor
  - view update - Update a view [intent=reverse_etl availability=implemented write=update_view]; approval: requires plan, preview, approval, and execute; risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.; flags: --view_id (required)
  - views list - List Notion views as ETL records. [intent=etl availability=implemented stream=views]

## Commands

### Inspect as a manual

```bash
pm connectors inspect notion
```

### Inspect as structured JSON

```bash
pm connectors inspect notion --json
```

## Agent Rules

- Run pm connectors inspect notion before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
