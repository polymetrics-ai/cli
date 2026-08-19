# Overview

Reads the existing Zoom users, meetings, and webinars ETL streams, and plans two approval-gated
warehouse destination actions through the Zoom REST API.

The committed public source lock records 35 Zoom Developer Docs OpenAPI documents (12,127,228
bytes and 1,937 REST operations), captured on 2026-08-19. The provider documents identify
themselves as OpenAPI 3.0.0; `api_surface.json` deliberately retains its earlier OpenAPI 3.1.1
ledger snapshot for inventory continuity. The committed crosswalk finds 1,911 exact ledger/source
identities, 26 current-source-only operations, and two ledger-only Zoom Phone paths. Neither side
is silently renamed or dropped.

The executable surface is deliberately small: the three preserved stream-backed reads plus two
source-backed reverse-ETL actions. The actions have loopback fixture proof only and are explicitly
pending live certification. `operations.json` is a source-contract inventory (776 reads, 971 REST
writes, including 311 DELETE contracts, and one bounded binary candidate); those entries are not
terminal commands and do not claim `api_surface` coverage.

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

Read behavior is an external Zoom API read of user, meeting, and webinar data. The two declared
warehouse destination actions are high risk and always use plan → preview → explicit approval →
execute:

- `pm zoom healthcare clinical-notes update` updates a clinical note completion status. The note
  identifier is redacted in write errors.
- `pm zoom quality-management interactions create` imports a Quality Management interaction from
  a third-party download URL.

These actions are fixture-proven but not live-certified. They are not a general REST-write surface.
Every other provider write remains disabled in `sources/zoom-declaration-disposition.json` with a
fixed-vocabulary reason, source evidence, and recoverability status. DELETE contracts retain
`mutation_class=delete`, destructive confirmation, and approval requirements in `operations.json`;
they are not fabricated as warehouse destination actions.

## Known limits

- Batch default: `read_page_size=100`.
- The current branch declares five executable ledger rows (three ETL streams and two reverse-ETL
  actions), while the already-delivered Wave 2 sub-PR owns 70 additional direct reads until parent
  integration. This is not provider-wide Zoom parity.
- `sources/zoom-declaration-disposition.json` accounts for all 1,913 ledger rows and the 26
  source-only rows. It records 1,838 disabled ledger rows on this branch, 54 provider-deprecated
  rows, 704 rows requiring paid or elevated Zoom capability review, and the two unresolved Phone
  path mismatches.
- The operation-backed REST write coverage gap tracked by #4281 applies to 971 source contracts,
  including all 311 source-backed DELETE contracts. The 34 multipart upload contracts are deferred
  on the bounded `file_upload` executor gap (G12); unsupported legacy extension fields are not
  coerced into a substitute.
- The bounded Clip download remains disabled: Zoom documents a 302 redirect but does not provide a
  provider-declared safe redirect host for the current binary contract. No `sync_transport.json` is
  declared because `declarative_stream_source` lacks a registered, proven source-to-warehouse
  transport adapter.
- Foundation-gap detail, source evidence, exact code locations, minimal remedies, and recoverability
  are recorded in `sources/zoom-foundation-gaps.json`. No auth, engine, or certification-scope code
  is changed by this connector lane.
