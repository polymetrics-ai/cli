# Overview

Reads GitLab projects, groups, users, and issues through the GitLab REST API v4 and carries a complete connector-local operation ledger for the pinned official GitLab OpenAPI v2 source.

Implemented fixture-backed streams: `projects`, `groups`, `users`, `issues`.

The complete official ledger has 1,146 operations: 397 ETL/read, 637 reverse-ETL write/mutation, 6 direct/provider query/search/metadata, 89 binary/file read/transfer, 15 CDC/changefeed/audit/webhook, and 2 excluded/not-applicable callback endpoints.

Only the four streams are executable in this wave. `api_surface.json`, `operations.json`, and `cli_surface.json` keep every other operation represented as typed planned/blocked metadata until a future connector-local stream/action/command adds fixtures and execution evidence.

`api_surface.json` also includes one connector-local supplemental GET `/users` row for the fixture-backed users stream because that documented top-level endpoint is absent from the pinned OpenAPI source; official GET `/projects/{id}/users` remains planned metadata.

Official source: https://gitlab.com/gitlab-org/gitlab/-/raw/9cd04099eb59d87335798e4f57a2bc5a2622e4cc/doc/api/openapi/openapi_v2.yaml
Branch provenance source: https://gitlab.com/api/v4/projects/gitlab-org%2Fgitlab/repository/branches/master

## Auth setup

Connection fields:

- `access_token` (required, secret, string); GitLab personal access token or OAuth access token. Used only for Bearer auth and redacted by the connector runtime.
- `base_url` (optional, string); default `https://gitlab.com/api/v4`; format `uri`; use `https://gitlab.example.com/api/v4` for self-managed instances or fixture replay.
- `start_date` (optional, string); format `date-time`; used by the current read stream filters where upstream supports a date bound.
- `page_size` (optional, string); default `50`; current streams send `per_page=50`.
- `mode` (optional, string); fixture/live mode marker used by local harnesses.

Requests use Bearer authentication from `secrets.access_token`. No fixture or metadata file contains credential values.

Connection checks call GET `/user` against the configured API base URL.

## Streams notes

Default pagination follows RFC 5988 `Link` headers with `rel=next`; fixture pages are bounded and sanitized.

- `projects`: GET `/projects`; records at the response root; sends `per_page=50`; optionally sends `last_activity_after` from `start_date`.
- `groups`: GET `/groups`; records at the response root; sends `per_page=50`.
- `users`: GET `/users`; records at the response root; sends `per_page=50`; optionally sends `created_after` from `start_date`. The pinned OpenAPI source does not enumerate this top-level users collection, so `api_surface.json` uses a connector-local supplemental coverage row and leaves GET `/projects/{id}/users` as planned metadata.
- `issues`: GET `/issues`; records at the response root; sends `per_page=50`; optionally sends `updated_after` from `start_date`; derives `author_id` from `author.id`.

Planned ETL/read rows in `operations.json` are metadata only. They are not advertised as executable streams until each has a schema, fixture replay, and conformance evidence.

## Write actions & risks

No `writes.json` actions are executable in this wave, and connector metadata keeps `capabilities.write=false` until named actions and fixtures are added.

The official ledger still includes all 637 mutation operations, including DELETE, destructive, admin, token/key/variable, hook, runner, package delete, and other high-risk operations. These are not blanket-excluded as unsafe. They are represented as planned/blocked typed metadata with risk, source URL, bounded request schemas where available, and approval notes.

Before any GitLab mutation can execute it must become a named reverse-ETL action with:

1. a bounded record schema and redaction contract;
2. dry-run plan and preview evidence;
3. explicit approval token;
4. `confirm: "destructive"` and typed confirmation for destructive/admin actions;
5. idempotency and cleanup notes where upstream behavior allows them;
6. fixture/conformance evidence proving request shape without live credentials.

No generic HTTP method/path/body, arbitrary GraphQL, shell, file, SQL write/read, extension, binary, or raw passthrough command is exposed.

## Known limits

- Fixture-backed implementation remains limited to 4 streams; all other official operations are planned/blocked metadata.
- This connector is not live-certified; fixture success must not be reported as provider certification.
- Direct/provider search/query/metadata rows depend on shared foundation #2985 before execution can be claimed.
- Binary/file transfer rows depend on shared foundation #2987 before bounded download/upload execution can be claimed.
- CDC/changefeed/audit/webhook rows depend on shared foundations #2986 and #2988 before CDC/changefeed claims can be made.
- Destructive/admin write rows depend on per-action schemas, redaction, fixtures, and typed confirmation evidence before execution can be claimed.
- The current top-level `/users` stream is fixture-backed legacy behavior; the pinned OpenAPI source omits that exact row, so api surface coverage uses a connector-local supplemental row and does not mark project-scoped users as implemented.
- The two excluded rows are GitLab Slack integration callback endpoints (`/integrations/slack/interactions` and `/integrations/slack/options`), not user-invoked connector operations.
- Generated lane counts: etl_read=397, reverse_etl_write=637, direct_read_query_search=6, binary_file=89, cdc_changefeed=15, excluded_not_applicable=2.
