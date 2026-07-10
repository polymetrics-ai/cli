# Overview

Reads Gong users, calls, scorecards, settings, flows, and related API resources through the Gong REST API.

Executable ETL streams: `users`, `calls`, `scorecards`, `crm_integrations`, `workspaces`, `trackers`, `briefs`, `library_folders`, `flows`, `flow_folders`, `call_outcomes`, `permission_profiles`.

Bounded direct-read commands cover the remaining official GET detail/query endpoints with a 1 MiB response cap and the `json_redacted` output policy. Selected POST read-query commands now execute through a typed operation direct-read path with schema-gated JSON bodies and no raw body flag. JSON mutations, typed multipart uploads, and top-level array CRM schema uploads are modeled as reverse-ETL write actions where the engine supports the request shape.

Service API documentation: https://gong.app.gong.io/ajax/settings/api/documentation/specs?version=.

## Auth setup

Connection fields:

- `access_key` (required, secret, string); Gong generated access key. Used for HTTP Basic auth; never logged.
- `access_key_secret` (required, secret, string); Gong generated access key secret. Used for HTTP Basic auth; never logged.
- `base_url` (optional, string); default `https://api.gong.io/v2`; format `uri`; Gong API base URL override for tests or proxies.
- `max_pages` (optional, string); default `0`; maximum pages; use 0, all, or unlimited to exhaust a paginated stream.
- `mode` (optional, string); fixture mode is used by credential-free conformance.
- `page_size` (optional, string); default `100`; records per page (1-100).
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound for supported incremental streams.

Secret fields are redacted in logs and write previews: `access_key`, `access_key_secret`.

Authentication behavior: HTTP Basic authentication using `secrets.access_key`, `secrets.access_key_secret`.

Connection checks call GET `/users` with query `limit`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `cursor`; next token from `records.cursor`.

The original incremental streams remain `users`, `calls`, and `scorecards`; additional list streams are full-refresh stream runners over public Gong list endpoints. Commands such as `pm gong workspaces list --json` use the same credential and bounded `--limit` behavior as other connector stream commands.

## Write actions & risks

Write actions are declared in `writes.json` for JSON Gong mutations, including calls, meetings, CRM integration registration/deletion, permission profiles, calls user access, flows assignments, engagement events, digital interactions, tasks, integration settings, data privacy erasure, bounded multipart media/CRM uploads, and the top-level JSON array CRM entity-schema upload.

Safety gates:

- Use reverse ETL plan -> preview -> approval -> execute.
- Destructive/admin actions declare `confirm: destructive`.
- No generic raw HTTP write, raw JSON body, arbitrary GraphQL mutation, shell write, or SQL write is exposed.
- Multipart upload commands accept only declared project-local file path fields and enforce byte limits before sending; file/path/content-like fields are redacted in command plans.
- Top-level JSON array writes use a declared `body_field` and `body_schema`; no raw JSON CLI flag is exposed.

Read risk: external Gong API read of call, user, CRM, settings, flow, and activity data; direct reads are bounded and redacted.

Write risk: typed Gong reverse ETL mutations for calls, meetings, CRM, permissions, flows, engagement, and data privacy erasure.

Approval: reverse ETL writes require plan, preview, approval, execute; destructive/admin actions require --confirm destructive.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage is inventoried from the public Gong OpenAPI 3.0.1 spec fetched on 2026-07-10: 57 paths and 67 operations.
- Executable coverage: 12 stream endpoints, 19 bounded direct reads (16 GET plus 3 typed POST read-query commands), and 26 typed reverse-ETL write actions.
- Operation metadata coverage: 10 broad POST read-query operations remain blocked until safe typed filter bodies are modeled.
- `/v2/calls/extensive`, `/v2/calls/transcript`, and `/v2/stats/interaction` are POST read-query operations in the official spec; they remain typed operations rather than unsafe raw API commands because their broad filter bodies are not yet exposed through narrow typed flags.
