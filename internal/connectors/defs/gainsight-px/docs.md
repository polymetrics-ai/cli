# Overview

Gainsight PX (aptrinsic) is a product-experience analytics platform. This bundle reads Gainsight
PX accounts, users, features, and segments through the aptrinsic REST API
(`https://api.aptrinsic.com/v1`), migrating `internal/connectors/gainsight-px` (the hand-written
legacy connector, which stays registered and unchanged until wave6's registry flip) to a
declarative bundle at capability parity. Gainsight PX is read-only here — no write actions.

## Auth setup

Provide a Gainsight PX (aptrinsic) API key via the `api_key` secret. It is sent as the
`X-APTRINSIC-API-KEY` header (`auth: [{"mode": "api_key_header", "header":
"X-APTRINSIC-API-KEY", "value": "{{ secrets.api_key }}"}]`) and is never logged.

## Streams notes

Four streams, each following the aptrinsic list-endpoint shape but with per-stream path/records-key
quirks preserved exactly from legacy:

| stream | request path | records key |
|---|---|---|
| `accounts` | `/accounts` | `accounts` |
| `users` | `/users` | `users` |
| `feature` | `/feature` (singular) | `features` (plural) |
| `segments` | `/segment` (singular) | `segments` (plural) |

Every stream's primary key is `["id"]`; no `x-cursor-field`/incremental cursor is declared,
matching legacy's `gainsightStreams()` (no `CursorFields` published — the API supports only
full-refresh sync).

Pagination is aptrinsic's `scrollId` cursor convention (`pagination.type: cursor` with
`cursor_param: scrollId`, `token_path: scrollId`): the next page's `scrollId` query param is read
from the current response body's own `scrollId` field, and pagination stops when that field is
empty — identical to legacy's `harvest` loop, whose own fixture-backed test
(`gainsightpx_test.go`) uses exactly this empty-`scrollId` stop signal as its terminal case. Every
request sends `pageSize=100` (matches legacy's default `page_size`).

## Write actions & risks

None. Gainsight PX is exposed read-only (`capabilities.write: false`), matching legacy's
`Capabilities{Write: false}` and its `Write` method returning `connectors.ErrUnsupportedOperation`
unconditionally.

## Known limits

- **`scrollId`-present-with-zero-records edge case (data-safe, unexercised by legacy's own tests)**:
  legacy's `harvest` loop stops on the FIRST of three conditions: an empty `scrollId`, an empty
  page (`len(records) == 0`), or a repeated `scrollId` value. The engine's `cursor` pagination
  type (`token_path` variant) checks only the token's own emptiness and a repeated-token loop
  guard — it does not additionally stop on a zero-record page that still carries a non-empty
  `scrollId`. Every real aptrinsic response observed in legacy's own test fixtures returns an
  empty `scrollId` exactly on the terminal (possibly-empty) page, so this corner case does not
  arise in practice; if the live API ever did return a non-empty token alongside zero records,
  the engine would issue one additional (likely also-empty) request rather than stopping
  immediately — this never omits, duplicates, or reorders emitted records, so it is data-safe
  under `docs/migration/conventions.md` §5's meta-rule, not a blocker. Fully closing this would
  need the `cursor`/`token_path` paginator to also accept a record-count-based stop signal (there
  is no boolean "has more" field in this API to use as `stop_path`), which is not needed for any
  observed real response shape.
- Full Gainsight PX API surface (custom events ingest, engagements, journeys, NPS surveys,
  webhooks) is out of scope for this wave; see `api_surface.json`'s `excluded` entries.
- No incremental sync: none of the four streams declares an `incremental` block, matching legacy
  (`InitialState` is not overridden by a stateful-cursor mechanism; the upstream Gainsight PX API
  has no documented "updated since" filter for these list endpoints). `client_filtered`
  incremental was considered and rejected: adding client-side cursor filtering here would be new
  behavior legacy never had, not a migration.
