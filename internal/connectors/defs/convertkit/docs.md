# Overview

ConvertKit (Kit) reads subscribers, forms, sequences, tags, and broadcasts through the ConvertKit
v3 REST API. This bundle migrates `internal/connectors/convertkit` (the legacy hand-written
connector, kept registered and unchanged until wave6's registry flip) to a declarative defs bundle
at capability parity. Read-only: the upstream source is full-refresh only and there are no safe
reverse-ETL writes to expose, matching legacy's `Capabilities.Write: false`.

## Auth setup

Provide the ConvertKit v3 API secret via the `api_key` secret; it is sent as the `api_secret` query
parameter on every request and never logged. Legacy resolved the same credential from any of three
secret names in precedence order (`api_key`, then `access_token`, then `api_secret`, first non-empty
wins) to tolerate differently-named credential storage across catalog versions. This bundle
reproduces that exact precedence with a 3-candidate `base.auth` list, each candidate gated by a
`when` clause on its own secret key (`docs/migration/conventions.md`'s dual-auth first-match-wins
ordering pattern) — `api_key` is declared first, `access_token` second, `api_secret` last, matching
legacy's fallback order exactly.

## Streams notes

`subscribers` and `broadcasts` use ConvertKit's page-based pagination (`page`/`per_page` query
params, `page_number` pagination type, `page_size: 50` matching legacy's default). The engine's
`page_number` paginator stops on a short (or empty) page — legacy's own stop condition combines a
short/empty page with the response's `total_pages` field, but in every real ConvertKit response
shape a final page IS the short/empty one once `total_pages` is reached, so this is a faithful,
non-data-changing representation of the same stop signal (the response's own `total_pages` field is
not separately consulted, matching how e.g. searxng's/every other `page_number` bundle in this repo
represents the same class of API).

`forms`, `sequences`, and `tags` return their full collection in a single unpaginated array under
their own resource key (`base.pagination` is overridden per-stream to `type: none`) — matches
legacy's `paginated: false` endpoint table exactly.

Every ConvertKit object schema declares `x-cursor-field: "created_at"`, matching legacy's published
catalog (`CursorFields: ["created_at"]` on every stream via `convertkitStreams()` in
`internal/connectors/convertkit/streams.go`), and every stream now also declares a bare
`incremental: { cursor_field: "created_at" }` block with no `request_param` — legacy's
`Read`/`harvest` never sends a server-side incremental filter (the upstream source is full-refresh
only in practice; ConvertKit's v3 list endpoints accept no since/updated-since query param), so no
`request_param` is added, but the engine derives `Manifest.SyncModes`'
`incremental_append`/`incremental_append_deduped` strictly from `stream.Incremental != nil`
(`DerivedSyncModes`, `internal/connectors/engine/connector.go`), not from schema `x-cursor-field`
alone. Omitting the block (as an earlier draft of this bundle did) would silently narrow the
published catalog versus legacy's `Catalog()`/`CursorFields`, exactly the same class of gap chameleon
and circleci correct for their own cursor-only-no-server-filter streams (see
`internal/connectors/defs/chameleon/streams.json` and
`internal/connectors/defs/circleci/streams.json`) — the bare block reproduces legacy's selectable
sync-mode capability with zero change to any emitted record's data.

## Write actions & risks

None. ConvertKit is read-only in this port (`writes.json` is intentionally absent).

## Known limits

- Legacy's configurable `page_size` (1-50, default 50) and `max_pages` (0/all/unlimited or a
  positive integer cap) config knobs are not modeled: `streams.json`'s `pagination.page_size` is a
  fixed JSON literal with no config-driven override mechanism (same class of limitation as
  searxng's `page_size`/`max_pages`, `docs/migration/conventions.md`'s read-only/no-auth golden). A
  declared-but-unwireable config key is worse than an absent one (F6, REVIEW.md), so neither is
  declared in `spec.json`; every paginated stream is fixed at `page_size: 50`, matching legacy's
  own default exactly.
- Full ConvertKit v3/v4 API surface (custom fields, purchases, webhooks, segments, courses, account
  metadata, subscribe/unsubscribe writes) is out of scope; see `api_surface.json`'s `excluded:
  {category: out_of_scope, reason: "Pass B capability expansion"}` entries. Only the 5
  legacy-parity read streams are implemented.
- `fixtures/streams/subscribers/**` and `fixtures/streams/broadcasts/**` ship a full 50-record
  first page (matching the real `page_size: 50` stop threshold exactly, per
  `docs/migration/conventions.md` §4's "no string-ification workaround" / real-wire-shape rule) so
  `pagination_terminates` proves the engine's actual production page-size threshold, not an
  artificially shrunk one.
