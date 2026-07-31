---
name: pm-notion
description: Notion connector knowledge and safe action guide.
---

# pm-notion

## Purpose

Reads and writes Notion users, pages, blocks, data sources, databases, comments, views, meeting notes, search results, and file-upload metadata through the official Notion API.

## Icon

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
- block_id
- comment_id
- data_source_id
- database_id
- file_upload_id
- max_pages
- page_id
- page_size
- property_id
- query_id
- task_id
- user_id
- view_id
- token (secret)

## ETL Streams

- databases:
  - primary key: id
  - cursor: last_edited_time
  - fields: archived(), created_time(), id(), in_trash(), last_edited_time(), object(), parent(), properties(), title(), type(), url()
- pages:
  - primary key: id
  - cursor: last_edited_time
  - fields: archived(), created_time(), id(), in_trash(), last_edited_time(), object(), parent(), properties(), type(), url()
- search_results:
  - primary key: id
  - cursor: last_edited_time
  - fields: archived(), created_time(), id(), in_trash(), last_edited_time(), object(), type(), url()
- users:
  - primary key: id
  - fields: archived(), avatar_url(), bot(), created_time(), id(), in_trash(), last_edited_time(), name(), object(), person(), type(), url()
- bot_user:
  - primary key: id
  - fields: archived(), avatar_url(), bot(), created_time(), id(), in_trash(), last_edited_time(), name(), object(), person(), type(), url()
- user:
  - primary key: id
  - fields: archived(), avatar_url(), bot(), created_time(), id(), in_trash(), last_edited_time(), name(), object(), person(), type(), url()
- page:
  - primary key: id
  - cursor: last_edited_time
  - fields: archived(), created_time(), id(), in_trash(), last_edited_time(), object(), parent(), properties(), type(), url()
- page_property_items:
  - primary key: id
  - fields: archived(), created_time(), has_more(), id(), in_trash(), last_edited_time(), next_cursor(), object(), property_item(), results(), type(), url()
- page_markdown:
  - primary key: id
  - fields: archived(), created_time(), id(), in_trash(), last_edited_time(), markdown(), object(), truncated(), type(), unknown_block_ids(), url()
- async_task:
  - primary key: id
  - fields: archived(), created_time(), id(), in_trash(), last_edited_time(), object(), operation(), poll_after_seconds(), status(), status_url(), type(), url()
- block:
  - primary key: id
  - cursor: last_edited_time
  - fields: archived(), created_time(), has_children(), id(), in_trash(), last_edited_time(), object(), parent(), type(), url()
- block_children:
  - primary key: id
  - cursor: last_edited_time
  - fields: archived(), created_time(), has_children(), id(), in_trash(), last_edited_time(), object(), parent(), type(), url()
- data_source:
  - primary key: id
  - cursor: last_edited_time
  - fields: archived(), created_time(), id(), in_trash(), last_edited_time(), object(), parent(), properties(), title(), type(), url()
- data_source_query:
  - primary key: id
  - cursor: last_edited_time
  - fields: archived(), created_time(), id(), in_trash(), last_edited_time(), object(), parent(), properties(), type(), url()
- data_source_templates:
  - primary key: id
  - fields: archived(), created_time(), description(), id(), in_trash(), last_edited_time(), name(), object(), type(), url()
- database:
  - primary key: id
  - cursor: last_edited_time
  - fields: archived(), created_time(), id(), in_trash(), last_edited_time(), object(), parent(), properties(), title(), type(), url()
- comments:
  - primary key: id
  - fields: archived(), created_time(), discussion_id(), id(), in_trash(), last_edited_time(), object(), parent(), rich_text(), type(), url()
- comment:
  - primary key: id
  - fields: archived(), created_time(), discussion_id(), id(), in_trash(), last_edited_time(), object(), parent(), rich_text(), type(), url()
- file_uploads:
  - primary key: id
  - fields: archived(), content_type(), created_time(), file_import_result(), filename(), id(), in_trash(), last_edited_time(), object(), status(), type(), url()
- file_upload:
  - primary key: id
  - fields: archived(), content_type(), created_time(), file_import_result(), filename(), id(), in_trash(), last_edited_time(), object(), status(), type(), url()
- custom_emojis:
  - primary key: id
  - fields: archived(), created_time(), emoji(), external_url(), id(), in_trash(), last_edited_time(), name(), object(), type(), url()
- views:
  - primary key: id
  - fields: archived(), created_time(), data_source_id(), database_id(), id(), in_trash(), last_edited_time(), name(), object(), type(), url()
- view:
  - primary key: id
  - fields: archived(), created_time(), data_source_id(), database_id(), id(), in_trash(), last_edited_time(), name(), object(), type(), url()
- view_query_results:
  - primary key: id
  - cursor: last_edited_time
  - fields: archived(), created_time(), id(), in_trash(), last_edited_time(), object(), parent(), properties(), type(), url()
- meeting_notes_query:
  - primary key: id
  - cursor: last_edited_time
  - fields: archived(), created_time(), has_children(), id(), in_trash(), last_edited_time(), object(), parent(), type(), url()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_page:
  - endpoint: POST /pages
  - required fields: parent
  - risk: creates a Notion page under the supplied parent; may create visible workspace content
- update_page:
  - endpoint: PATCH /pages/{{ record.page_id }}
  - required fields: page_id
  - risk: updates mutable fields on an existing Notion page or archives/restores it
- move_page:
  - endpoint: POST /pages/{{ record.page_id }}/move
  - required fields: page_id, parent
  - risk: moves an existing Notion page to another parent
- update_page_markdown:
  - endpoint: PATCH /pages/{{ record.page_id }}/markdown
  - required fields: page_id, markdown
  - risk: replaces or updates a page content body from typed markdown
- update_block:
  - endpoint: PATCH /blocks/{{ record.block_id }}
  - required fields: block_id
  - risk: updates a Notion block or archives/restores it
- delete_block:
  - endpoint: DELETE /blocks/{{ record.block_id }}
  - required fields: block_id
  - risk: archives/deletes a block by ID; destructive confirmation required
- append_block_children:
  - endpoint: PATCH /blocks/{{ record.block_id }}/children
  - required fields: block_id, children
  - risk: appends children blocks under an existing block
- create_data_source:
  - endpoint: POST /data_sources
  - required fields: parent, properties
  - risk: creates a Notion data source under the supplied parent
- update_data_source:
  - endpoint: PATCH /data_sources/{{ record.data_source_id }}
  - required fields: data_source_id
  - risk: updates a Notion data source schema or metadata; may archive or trash existing Notion content when archived/in_trash is supplied
- create_database:
  - endpoint: POST /databases
  - required fields: parent
  - risk: creates a legacy Notion database object as documented by the official API
- update_database:
  - endpoint: PATCH /databases/{{ record.database_id }}
  - required fields: database_id
  - risk: updates a legacy Notion database object as documented by the official API
- create_comment:
  - endpoint: POST /comments
  - required fields: rich_text
  - risk: creates a Notion comment on a page/block/discussion
- update_comment:
  - endpoint: PATCH /comments/{{ record.comment_id }}
  - required fields: comment_id, rich_text
  - risk: updates a Notion comment rich_text body
- delete_comment:
  - endpoint: DELETE /comments/{{ record.comment_id }}
  - required fields: comment_id
  - risk: deletes a Notion comment by ID; destructive confirmation required
- create_view:
  - endpoint: POST /views
  - required fields: data_source_id, name
  - risk: creates a Notion view for a data source or database
- update_view:
  - endpoint: PATCH /views/{{ record.view_id }}
  - required fields: view_id
  - risk: updates a Notion view
- delete_view:
  - endpoint: DELETE /views/{{ record.view_id }}
  - required fields: view_id
  - risk: deletes a Notion view; destructive confirmation required
- create_view_query:
  - endpoint: POST /views/{{ record.view_id }}/queries
  - required fields: view_id
  - risk: creates a bounded Notion view query resource for later result retrieval
- delete_view_query:
  - endpoint: DELETE /views/{{ record.view_id }}/queries/{{ record.query_id }}
  - required fields: view_id, query_id
  - risk: deletes a Notion view query resource; destructive confirmation required
- create_meeting_note:
  - endpoint: POST /blocks/meeting_notes
  - required fields: title
  - risk: creates a Notion meeting note block
- create_file_upload:
  - endpoint: POST /file_uploads
  - required fields: mode, filename
  - risk: creates a bounded Notion file-upload resource before upload or external import
- complete_file_upload:
  - endpoint: POST /file_uploads/{{ record.file_upload_id }}/complete
  - required fields: file_upload_id
  - risk: completes a Notion multi-part file upload after all approved parts are sent

## Security

- read risk: external Notion API reads for workspace users, pages, blocks, data sources, databases, comments, views, meeting notes, search results, and file-upload metadata
- write risk: typed reverse-ETL actions create or update Notion pages, blocks, data sources, databases, comments, views, meeting notes, view queries, and file-upload lifecycle resources; archive/delete-like actions require destructive confirmation
- approval: reverse ETL writes require plan preview and explicit approval before execution; fixture-only conformance is not live certification
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

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
