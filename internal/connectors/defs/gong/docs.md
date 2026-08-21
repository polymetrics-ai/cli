# Overview

Reads Gong users, calls, scorecards, targets, settings, flows, and related API resources through the Gong REST API.

Executable ETL streams: `users`, `calls`, `scorecards`, `crm_integrations`, `workspaces`, `trackers`, `briefs`, `library_folders`, `flows`, `flow_folders`, `call_outcomes`, `permission_profiles`. Bounded target definition lookup is exposed as `pm gong targets list --workspaceId <id>` because the official endpoint requires a workspace query parameter rather than a connection-global stream.

Bounded direct-read commands cover the official GET detail/query endpoints and all 13 POST read-query endpoints with the `json_redacted` output policy. That policy preserves all ordinary provider response fields; only concrete configured credential values are masked with an explicit marker. POST reads execute through typed operation metadata with connector-authored flags and schema-gated JSON bodies; no raw body flag exists. Call transcripts are available through `pm gong calls transcript` with call ID, time range, workspace, and cursor filters and a 16 MiB response cap. JSON mutations, typed multipart uploads, target assignment CSV uploads, and top-level array CRM schema uploads are modeled as reverse-ETL write actions where the engine supports the request shape.

Service API documentation: https://gong.app.gong.io/ajax/settings/api/documentation/specs?version=.

## Auth setup

Connection fields:

- `access_key` (required, secret, string); Gong generated access key. Used for HTTP Basic auth; never logged.
- `access_key_secret` (required, secret, string); Gong generated access key secret. Used for HTTP Basic auth; never logged.
- `base_url` (optional, string); default `https://api.gong.io/v2`; format `uri`; Gong API base URL override for tests or proxies.
- `max_pages` (optional, string); default `0`; maximum pages; use 0, all, or unlimited to exhaust a paginated stream.
- `mode` (optional, string); fixture mode is used by credential-free conformance.
- `page_size` (optional, string); default `100`; compatibility page-size request value for legacy Gong list streams; current Gong OpenAPI does not document a provider-side page-size parameter. CLI `--limit` remains the PM output cap.
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound for supported incremental streams.

Secret fields are redacted in logs and write previews: `access_key`, `access_key_secret`.

Authentication behavior: HTTP Basic authentication using `secrets.access_key`, `secrets.access_key_secret`.

Connection checks call GET `/users` with query `limit`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `cursor`; next token from `records.cursor`.

The original incremental streams remain `users`, `calls`, and `scorecards`; additional list streams are full-refresh stream runners over public Gong list endpoints. `pm gong calls list` accepts `--from` and `--to` ISO-8601/RFC3339 bounds mapped to Gong `fromDateTime` and `toDateTime`; the upper bound is exclusive. Commands such as `pm gong workspaces list --json` use the same credential and bounded PM output `--limit` behavior as other connector stream commands. Global `--limit` caps emitted PM records only; it does not claim to control Gong page size or total provider-side results.

## Write actions & risks

Write actions are declared in `writes.json` for JSON Gong mutations, including calls, meetings, CRM integration registration/deletion, permission profiles, calls user access, flows assignments, targets assignment CSV uploads, engagement events, digital interactions, tasks, integration settings, data privacy erasure, bounded multipart media/CRM uploads, and the top-level JSON array CRM entity-schema upload.

Safety gates:

- Use reverse ETL plan -> preview -> approval -> execute.
- Destructive/admin actions declare `confirm: destructive`.
- No generic raw HTTP write, raw JSON body, arbitrary GraphQL mutation, shell write, or SQL write is exposed.
- Multipart upload commands accept only declared project-local file path fields, bind approvals to a SHA-256 content digest, snapshot and verify the approved bytes before any HTTP request, and enforce byte limits during preflight, snapshotting, and streaming; file/path/content-like fields are redacted in command plans.
- Top-level JSON array writes use a declared `body_field` and `body_schema`; no raw JSON CLI flag is exposed.
- Gong DELETE operations (`meetings delete`, `crm integrations delete`, and `calls users-access delete`) are canonical reverse-ETL write actions with `confirm: destructive`, typed record schemas, and plan -> preview -> explicit approval -> execute safeguards.

Read risk: external Gong API read of call, user, CRM, settings, flow, and activity data; direct reads are bounded and preserve ordinary provider response fields. Only concrete configured credential values are masked.

Write risk: typed Gong reverse ETL mutations for calls, meetings, CRM, permissions, flows, targets, engagement, and data privacy erasure.

Approval: reverse ETL writes require plan, preview, approval, execute; destructive/admin actions require --confirm destructive.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage was re-audited against the public Gong OpenAPI 3.0.1 spec on 2026-08-21 UTC: 59 paths and 69 operations (GET 29, POST 28, PUT 8, PATCH 1, DELETE 3). The source lock records the exact current artifact digest and semantic inventory in `sources/gong-operation-source-lock.json`; the source remains `https://gong.app.gong.io/ajax/settings/api/documentation/specs?version=`.
- Executable coverage after the 2026-08-21 reconciliation: 12 stream endpoints, 30 bounded direct reads (17 GET plus all 13 typed POST read-query commands), 27 typed reverse-ETL write actions, and three bounded multipart upload actions; `api_surface.json` has 69/69 covered rows and 0 excluded/planned/blocked rows. Gong's official source has no binary-download response operation.
- The checkpoint found 0 missing/stale official operation rows, 0 required write parameter/schema gaps, and 0 required direct-read flag gaps after fixing connector-local metadata. Certification remains 0 because this work used fixture/local validation only and no live Gong credentials or provider calls.
- `targets upload-assignments` exposes Gong's optional `validateOnly` query parameter as `--validate-only`. When absent, it is omitted; when supplied, its exact value is transmitted through the declaration-owned optional query mapping. The operation retains typed target/workspace/file inputs and destructive approval.
- POST read-query filters are allow-listed as typed command flags. Arbitrary/raw request bodies remain intentionally unavailable.
