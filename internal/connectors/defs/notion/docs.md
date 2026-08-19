# Notion

## Overview

Reads and writes Notion pages, databases, data sources, blocks, comments, views, and file uploads
through the Notion REST API.

The documented surface is **51 operations**, re-derived on 2026-08-07 from the provider's official
OpenAPI 3.1.0 document at <https://developers.notion.com/openapi.json>: 49 HTTP operations across 34
paths (20 GET, 17 POST, 8 PATCH, 4 DELETE), plus the 2 legacy endpoints documented only under the
reference nav's explicit "Databases (deprecated)" group and absent from the current OpenAPI document
(`GET /v1/databases`, `POST /v1/databases/{database_id}/query`).

The document's 31 top-level `webhooks` entries are webhook **events**, not request operations, and
are excluded per the sweep counting policy. Notion publishes **no webhook management endpoints**
(there is no `/v1/webhooks` path), so there is nothing in that area to implement.

Every documented operation is partitioned exactly once in `api_surface.json` and carries exactly one
disposition — executable, blocked with a named dependency, or not executable with a source citation.
None is blank.

## Auth setup

A single `x-secret` field, `token`, carried as a bearer credential. Add it from an environment
variable or stdin, never as prompt text:

```
pm credentials add notion --from-env NOTION_TOKEN
```

`Notion-Version` is pinned to `2022-06-28` in `streams.json`'s base headers.

## Streams notes

Six ETL streams. Three are Tier-2 `StreamHook`-driven because `POST /v1/search` carries its
pagination cursor and object filter **in the request body**, which the declarative read path cannot
express (`engine/read.go` issues a nil body on every declarative read):

| Stream | Endpoint | Mode |
| --- | --- | --- |
| `databases` | `POST /v1/search` (object=database) | hook |
| `pages` | `POST /v1/search` (object=page) | hook |
| `users` | `GET /v1/users` | hook |
| `custom_emojis` | `GET /v1/custom_emojis` | declarative |
| `file_uploads` | `GET /v1/file_uploads` | declarative |
| `views` | `GET /v1/views` | declarative |

`POST /v1/search` is **one documented operation** exercised with two documented object filters. The
`api_surface.json` `covered_by` block accepts a single stream name and no bundle leaves a declared
stream without a row, so it is carried on two qualified rows — the convention this bundle already
shipped. The same applies to `PATCH /v1/comments/{comment_id}`, which carries its two modelled union
arms.

The three declarative streams use Notion's standard cursor pagination
(`start_cursor` → `next_cursor`, stopping on `has_more`), declared once in `streams.json`'s base.
The hook-driven streams run their own harvest loop and ignore that block.

## Write actions & risks

24 write actions. Every mutation goes through plan → preview → approval → execute; Notion publishes
no idempotency header, so the runtime submits each approved record exactly once rather than relying
on provider-side deduplication.

Four actions are destructive and additionally require destructive confirmation: `delete_block`,
`delete_comment`, `delete_view`, `delete_view_query`.

### Union-rooted request bodies

`AGENTS.md` is explicit that a `record_schema` rooted at `oneOf`/`anyOf` is not one executable
command contract. Three Notion bodies are union-rooted and are modelled by **naming the arm each
action executes**, selected by the field that arm requires rather than by position:

| Endpoint | Arms | Actions |
| --- | --- | --- |
| `PATCH /v1/comments/{comment_id}` | `{rich_text}` / `{markdown}` | `update_comment`, `update_comment_markdown` |
| `POST /v1/pages/{page_id}/move` | `parent.{page_id}` / `parent.{data_source_id}` | `move_page`, `move_page_to_data_source` |
| `PATCH /v1/blocks/{block_id}` | `{in_trash}` / 31-arm block-type union | `update_block` (the `in_trash` arm only) |

## Known limits

- **Block content updates are not promoted.** `PATCH /v1/blocks/{block_id}`'s content arm is a
  31-way discriminated union over block types with no single closed contract. Only the universally
  applicable `in_trash` archive/restore arm is promoted. The endpoint is reachable; changing a
  block's *content* is not exposed as a typed command.
- **Three writes are `partial`, not `implemented`.** `append_block_children`, `update_comment`, and
  `create_data_source` each have a required structured field with no typed scalar leaf, so no honest
  flag contract exists for them:
  - `append_block_children.children` — an array over the 31-arm block-type union;
  - `update_comment.rich_text` — an array of rich-text objects;
  - `create_data_source.properties` — a free-form map of property definitions.

  Each keeps its full typed `record_schema` in `writes.json` and stays usable through reverse ETL
  from a source record. Only the direct flag-driven command form is partial.
- **File upload is blocked with a named dependency.** `POST /v1/file_uploads/{file_upload_id}/send`
  is bounded in `operations.json` as a `file_upload` operation, but promotion waits on the **shared
  file upload runner** that binds provider paths to approved payload digests. This follows the same
  disposition xero's 22 attachment uploads already carry. There is no Notion binary *download*
  endpoint — file bytes are served from provider-signed storage URLs, not the API.
- **OAuth endpoints are permanently disallowed, not deferred.** `POST /v1/oauth/token`,
  `/v1/oauth/introspect`, and `/v1/oauth/revoke` mint, describe, or revoke a caller credential. Their
  responses are entirely token-derived, and this repository does not emit credential values, so they
  are never exposed as connector commands. Credential lifecycle belongs to `pm credentials`.
- **Direct reads use `json_redacted`.** That is the only general-purpose output policy the runtime
  supports for direct reads (`commandrunner.supportedDirectReadOutputPolicies`).
- **Flags exist for scalar fields only.** An object or array field is supplied by the reverse-ETL
  source record, not by a shell flag — the same rule the merged `recurly` bundle follows.
