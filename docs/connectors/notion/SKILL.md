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
- authorization_code (secret)
- oauth_client_id (secret)
- oauth_client_secret (secret)
- refresh_token (secret)
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

- Work with Notion pages, databases, data sources, blocks, comments, views and file uploads from the command line.
- Usage: pm notion <command> [flags]
- Source CLI: Notion API (https://developers.notion.com/openapi.json)
- Global flags:
  - --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
- Databases
- Pages
- Users
- Custom Emojis
- File Uploads
- Views
- Async Task
- Block
- Comment
- Data Source
- Database
- File Upload
- Page
- User
- View
- Meeting Note
- Other Commands
  - databases list - List Notion databases as ETL records. [intent=etl availability=implemented stream=databases]
  - pages list - List Notion pages as ETL records. [intent=etl availability=implemented stream=pages]
  - users list - List Notion users as ETL records. [intent=etl availability=implemented stream=users]
  - custom-emojis list - List Notion custom emojis as ETL records. [intent=etl availability=implemented stream=custom_emojis]
  - file-uploads list - List Notion file uploads as ETL records. [intent=etl availability=implemented stream=file_uploads]
  - views list - List Notion views as ETL records. [intent=etl availability=implemented stream=views]
  - async-task get - Retrieve an async task [intent=direct_read availability=implemented operation=notion.retrieve_async_task]; notes: Bounded Notion read; fixed method and path with typed request fields.; flags: --task-id (required), --page, --page-cursor
  - block get - Retrieve a block [intent=direct_read availability=implemented operation=notion.retrieve_block]; notes: Bounded Notion read; fixed method and path with typed request fields.; flags: --block-id (required), --page, --page-cursor
  - block children list - Retrieve block children [intent=direct_read availability=implemented operation=notion.retrieve_block_children]; notes: Bounded Notion read; fixed method and path with typed request fields.; flags: --block-id (required), --page-size, --page, --page-cursor
  - comment list - List comments [intent=direct_read availability=implemented operation=notion.list_comments]; notes: Bounded Notion read; fixed method and path with typed request fields.; flags: --block-id (required), --page-size, --page, --page-cursor
  - comment get - Retrieve a comment [intent=direct_read availability=implemented operation=notion.retrieve_comment]; notes: Bounded Notion read; fixed method and path with typed request fields.; flags: --comment-id (required), --page, --page-cursor
  - data-source get - Retrieve a data source [intent=direct_read availability=implemented operation=notion.retrieve_data_source]; notes: Bounded Notion read; fixed method and path with typed request fields.; flags: --data-source-id (required), --page, --page-cursor
  - data-source templates list - List templates in a data source [intent=direct_read availability=implemented operation=notion.list_data_source_templates]; notes: Bounded Notion read; fixed method and path with typed request fields.; flags: --data-source-id (required), --name, --page-size, --page, --page-cursor
  - database get - Retrieve a database [intent=direct_read availability=implemented operation=notion.retrieve_database]; notes: Bounded Notion read; fixed method and path with typed request fields.; flags: --database-id (required), --page, --page-cursor
  - file-upload get - Retrieve a file upload [intent=direct_read availability=implemented operation=notion.retrieve_file_upload]; notes: Bounded Notion read; fixed method and path with typed request fields.; flags: --file-upload-id (required), --page, --page-cursor
  - page get - Retrieve a page [intent=direct_read availability=implemented operation=notion.retrieve_page]; notes: Bounded Notion read; fixed method and path with typed request fields.; flags: --page-id (required), --filter-properties, --page, --page-cursor
  - page markdown get - Retrieve a page as markdown [intent=direct_read availability=implemented operation=notion.retrieve_page_markdown]; notes: Bounded Notion read; fixed method and path with typed request fields.; flags: --page-id (required), --include-transcript, --page, --page-cursor
  - page property get - Retrieve a page property item [intent=direct_read availability=implemented operation=notion.retrieve_page_property]; notes: Bounded Notion read; fixed method and path with typed request fields.; flags: --page-id (required), --property-id (required), --page-size, --page, --page-cursor
  - user me get - Retrieve your token's bot user [intent=direct_read availability=implemented operation=notion.retrieve_bot_user]; notes: Bounded Notion read; fixed method and path with typed request fields.; flags: --page, --page-cursor
  - user get - Retrieve a user [intent=direct_read availability=implemented operation=notion.retrieve_user]; notes: Bounded Notion read; fixed method and path with typed request fields.; flags: --user-id (required), --page, --page-cursor
  - view get - Retrieve a view [intent=direct_read availability=implemented operation=notion.retrieve_view]; notes: Bounded Notion read; fixed method and path with typed request fields.; flags: --view-id (required), --page, --page-cursor
  - view query results get - Get view query results [intent=direct_read availability=implemented operation=notion.retrieve_view_query_results]; notes: Bounded Notion read; fixed method and path with typed request fields.; flags: --view-id (required), --query-id (required), --page-size, --page, --page-cursor
  - data-source query - Query a data source [intent=direct_read availability=implemented operation=notion.query_data_source]; notes: Bounded Notion read; fixed method and path with typed request fields.; flags: --data-source-id (required), --filter-properties, --page-size, --is-archived, --result-type, --page, --page-cursor
  - meeting-note query - Query meeting notes [intent=direct_read availability=implemented operation=notion.query_meeting_notes]; notes: Bounded Notion read; fixed method and path with typed request fields.; flags: --limit, --page, --page-cursor
  - meeting-note create - Create a meeting note [intent=reverse_etl availability=implemented write=create_meeting_note]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.; flags: --language, --title
  - block update - Update a block [intent=reverse_etl availability=implemented write=update_block]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.; flags: --block-id (required), --in-trash
  - block delete - Delete a block [intent=reverse_etl availability=implemented write=delete_block]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.; flags: --block-id (required)
  - block children append - Append block children [intent=reverse_etl availability=partial write=append_block_children]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.; notes: Reverse ETL from a source record is the supported path: children is an array over Notion's 31-arm block-type union; no typed scalar leaf exists for a flag contract; flags: --block-id (required)
  - comment create - Create a comment [intent=reverse_etl availability=implemented write=create_comment]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.
  - comment update - Update a comment [intent=reverse_etl availability=partial write=update_comment]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.; notes: Reverse ETL from a source record is the supported path: rich_text is an array of rich-text objects with no scalar leaf; supply it from a reverse-ETL source record; flags: --comment-id (required)
  - comment update-markdown - Update a comment [intent=reverse_etl availability=implemented write=update_comment_markdown]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.; flags: --comment-id (required), --markdown (required)
  - comment delete - Delete a comment [intent=reverse_etl availability=implemented write=delete_comment]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.; flags: --comment-id (required)
  - data-source create - Create a data source [intent=reverse_etl availability=partial write=create_data_source]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.; notes: Reverse ETL from a source record is the supported path: properties is a free-form map of property definitions with no declared required scalar leaf; flags: --parent-database-id (required)
  - data-source update - Update a data source [intent=reverse_etl availability=implemented write=update_data_source]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.; flags: --data-source-id (required), --in-trash
  - database create - Create a database [intent=reverse_etl availability=implemented write=create_database]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.; flags: --is-inline, --parent (required)
  - database update - Update a database [intent=reverse_etl availability=implemented write=update_database]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.; flags: --database-id (required), --in-trash, --is-inline, --is-locked, --parent
  - page create - Create a page [intent=reverse_etl availability=implemented write=create_page]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.; flags: --allow-async, --markdown
  - page update - Update page [intent=reverse_etl availability=implemented write=update_page]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.; flags: --erase-content, --in-trash, --is-archived, --is-locked, --page-id (required)
  - page markdown update - Update a page's content as markdown [intent=reverse_etl availability=implemented write=update_page_markdown]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.; flags: --allow-async, --page-id (required)
  - page move - Move a page [intent=reverse_etl availability=implemented write=move_page]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.; flags: --page-id (required), --parent-page-id (required)
  - page move-to-data-source - Move a page [intent=reverse_etl availability=implemented write=move_page_to_data_source]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.; flags: --page-id (required), --parent-data-source-id (required)
  - view create - Create a view [intent=reverse_etl availability=implemented write=create_view]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.; flags: --data-source-id (required), --database-id, --name (required), --type (required), --view-id
  - view update - Update a view [intent=reverse_etl availability=implemented write=update_view]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.; flags: --name, --view-id (required)
  - view delete - Delete a view [intent=reverse_etl availability=implemented write=delete_view]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.; flags: --view-id (required)
  - view query create - Create a view query [intent=reverse_etl availability=implemented write=create_view_query]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.; flags: --page-size, --view-id (required)
  - view query delete - Delete a view query [intent=reverse_etl availability=implemented write=delete_view_query]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.; flags: --query-id (required), --view-id (required)
  - file-upload create - Create a file upload [intent=reverse_etl availability=implemented write=create_file_upload]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.; flags: --content-type, --external-url, --filename, --mode, --number-of-parts
  - file-upload complete - Complete a multi-part file upload [intent=reverse_etl availability=implemented write=complete_file_upload]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium — Notion has no provider idempotency header; keep reverse ETL in plan/preview/approve/execute so each approved record is submitted exactly once.; flags: --file-upload-id (required)
  - file-upload send - Upload a file [intent=reverse_etl availability=planned operation=notion.send_file_upload]; approval: binary/file uploads require plan, preview, explicit approval, and payload digest approval before execution; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; promotion is blocked on the shared file upload runner that binds provider paths to approved payload digests. Named dependency, not a blank disposition.
- Help topics:
  - safety - Reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.
  - pagination - Notion paginates with start_cursor/next_cursor and has_more; ETL streams follow the cursor to exhaustion within the configured page bounds.

## Sync Transport

- Source transport: declared
- Destination transport: declared
- A declared transport still requires runtime preflight and externally verified conformance; it is not a certification claim.
- Source executor: declarative_api/declarative_stream_source
- Destination executor: declarative_api/declarative_typed_destination

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
