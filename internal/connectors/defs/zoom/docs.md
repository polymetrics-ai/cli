# Overview

Exposes source-backed Zoom API commands through the Zoom REST API. The current branch contains 712
runnable commands: three preserved ETL streams, 505 direct reads, and 204 approval-gated typed
write commands, including 185 guarded deletes. The 707 newly mapped endpoint rows are implemented
pending certification; none is a certified claim. One bounded REST read has accepted live proof,
but it remains uncertified until the matrix can project an operation-specific fixture.

The committed public source lock records 35 Zoom Developer Docs OpenAPI documents (12,127,228
bytes and 1,937 REST operations), captured on 2026-08-19. The provider documents identify
themselves as OpenAPI 3.0.0; `api_surface.json` deliberately retains its earlier OpenAPI 3.1.1
ledger snapshot for inventory continuity. The committed crosswalk finds 1,911 exact ledger/source
identities, 26 current-source-only operations, and two ledger-only Zoom Phone paths. Neither side
is silently renamed or dropped.

`operations.json` retains the typed source-contract inventory (776 reads, 971 REST writes,
including 311 DELETE contracts, and one bounded binary candidate). The command and action surface
is derived from the same pinned contracts and bound to exact `api_surface.json` endpoints. Provider
OAuth scope requirements are command metadata: an insufficient token receives Zoom's provider 403,
not a locally disabled command.

Service API documentation: https://developers.zoom.us/docs/api/.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Zoom OAuth access token, sent as a Bearer token
  (Authorization: Bearer `<access_token>`). Never logged.
- `base_url` (optional, string); default `https://api.zoom.us/v2`; format `uri`; Zoom API base URL
  override for tests or proxies.
- `max_pages` (optional, string); default `0`. The field remains in the connection specification,
  but the current Zoom cursor paginator does not consume it.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; records per page (1-300); sent as the `page_size`
  query parameter.
- `user_id` (optional, string); Zoom user ID or email that scopes the `meetings` and `webinars`
  streams. Set it in credential configuration or override it with `--user-id` for either command.

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://api.zoom.us/v2`, `max_pages=0`, `page_size=100`.
`max_pages` is materialized as configuration but does not affect Zoom pagination.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/users` with query `page_size`=`1`.

## Streams notes

Default pagination is cursor pagination: send `next_page_token` and take the next token from
`next_page_token`. The stream command `--limit` bounds emitted records (default 100).

- `users`: `pm zoom users list` reads GET `/v2/users` through stream `users`. Records are at
  `users`; query `page_size`=`{{ config.page_size }}`; computed output fields are `name` and
  `updated_at`; emits passthrough records. Provider reference:
  https://developers.zoom.us/docs/api/users.md.
- `meetings`: `pm zoom meetings list` reads GET `/v2/users/{userId}/meetings` through stream
  `meetings`. Supply `--user-id` or configure `user_id` in the credential. Records are at
  `meetings`; query `page_size`=`{{ config.page_size }}`; computed output fields are `id`, `name`,
  and `updated_at`; emits passthrough records. Provider reference:
  https://developers.zoom.us/docs/api/meetings.md.
- `webinars`: `pm zoom webinars list` reads GET `/v2/users/{userId}/webinars` through stream
  `webinars`. Supply `--user-id` or configure `user_id` in the credential. Records are at
  `webinars`; query `page_size`=`{{ config.page_size }}`; computed output fields are `id`, `name`,
  and `updated_at`; emits passthrough records. Provider reference:
  https://developers.zoom.us/docs/api/meetings.md.

## Write actions & risks

Read behavior is an external Zoom API read. All 204 typed write actions use plan → preview →
explicit approval → execute. The 185 DELETE actions also require destructive confirmation. The two
fixture-backed warehouse destination actions remain useful examples:

- `pm zoom healthcare clinical-notes update` updates a clinical note completion status. The note
  identifier is redacted in write errors.
- `pm zoom quality-management interactions create` imports a Quality Management interaction from
  a third-party download URL.

No newly mapped command is certified. A bounded authenticated `operation:rest_read` proof records
two HTTP 200 exchanges with fingerprint-only evidence and `observed_operations` scope; it does not
certify the operation because the current matrix has no operation-specific fixture projection.
All 204 generated mutation candidates are explicitly unassessed/deferred on the typed destination
foundation gap. `sources/zoom-declaration-disposition.json` supplies the per-operation status,
evidence, fixed-vocabulary rejection reason, and recoverability record. It does not silently treat
ordinary deletes as unsafe.

## Known limits

- Batch default: `read_page_size=100`.
- The current branch exposes 712 runnable commands. A separately delivered Wave 2 sub-PR owns 70
  additional direct reads until its parent integration; this branch neither duplicates nor claims
  those commands. Zoom provider-wide completeness is not claimed.
- `sources/zoom-declaration-disposition.json` accounts for all 1,913 ledger rows and the 26
  source-only rows. It records 1,131 disabled ledger rows: 535 require paid provider entitlement,
  538 have a documented foundation gap, 54 are provider-deprecated, and two are schema-incompatible.
- The remaining 471 JSON-body writes need a typed root-body input contract. Twenty-five operation
  contracts need array-query encoding. The 34 multipart uploads remain deferred on the bounded `file_upload`
  executor gap (G12); unsupported legacy extension fields are not coerced into a substitute.
- The bounded Clip download remains disabled: Zoom documents a 302 redirect but does not provide a
  provider-declared safe redirect host for the current binary contract. `sync_transport.json`
  declares the three preserved streams through the connector-neutral declarative source adapter.
  Zoom does not declare a reverse-ETL transport destination: no connector-neutral typed destination
  executor exists yet, and no action transport binding is invented as a substitute.
- `operation-specific-fixture-evidence-projection` prevents the one accepted `operation:rest_read`
  live proof from becoming a certified operation cell. The minimal foundation change is an exact
  direct-operation fixture/replay projection; a stream fixture is not substituted as evidence.
- A bounded full live run passed the catalog, append ETL, and query read-back stages for Users and
  Meetings. It cannot publish those capability proofs yet because the shared harness unconditionally
  reports its obsolete definition-fixture stage as skipped/failing, aggregates unrelated stream
  refusals into one report, and invokes flow/schedule checks when Zoom declares no executable flow
  pair. These are `definition-fixture-conformance-certification-stage`,
  `capability-scoped-live-evidence-publication`, and `schedule-roundtrip-source-only-skip` foundation
  gaps; their remedies preserve all-or-nothing full-parity semantics and do not promote a connector
  from partial stage proof.
- The pinned Users and Meetings OpenAPI response schemas declare durable creation timestamps for all
  three preserved streams: `user_created_at`/`created_at` for users and `created_at` for meetings
  and webinars. The connector projects those exact provider fields as `created_at` and declares
  that field as its cursor; it does not infer a watermark. The bounded Webinar probe returned
  HTTP 400 with Zoom error code 200: the Webinar plan is missing and must be subscribed/enabled for
  the user. It remains a runnable command for entitled accounts but is explicitly uncertified for
  this account with `requires-paid-tier` evidence; the account identifier is redacted.
- Foundation-gap detail, source evidence, exact code locations, minimal remedies, and recoverability
  are recorded in `sources/zoom-foundation-gaps.json`. No auth, engine, or certification-scope code
  is changed by this connector lane.
