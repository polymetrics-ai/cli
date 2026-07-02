# Overview

Countercyclical is a financial-intelligence platform for investment teams; this bundle reads its
investments, valuations, and research memos streams through the Countercyclical REST API. It
migrates `internal/connectors/countercyclical` (the legacy hand-written connector, kept registered
and unchanged until wave6's registry flip) to a declarative defs bundle at capability parity.
Read-only.

## Auth setup

Provide the Countercyclical API key via the `api_key` secret; it is sent as the `apiKey` query
parameter on every request (matching the upstream `ApiKeyAuthenticator`, `inject_into
request_parameter` convention) and never logged.

## Streams notes

All 3 streams (`investments`, `valuations`, `memos`) are `GET` requests returning a root-level JSON
array (`records.path: "."` selects the body root), primary key `["id"]`. The upstream API is
documented as having no pagination at all; this bundle nonetheless declares
`pagination.type: offset_limit` (`limit`/`offset` query params, `page_size: 100`) as a defensive
measure exactly mirroring legacy's own `harvest` loop, which sends `limit`/`offset` and continues
only while a full page (`== pageSize`) comes back — a short/empty page (the expected, honest case
for a genuinely unpaginated API) stops after exactly one request. Legacy only sends `offset` when
it is greater than 0 (page 1 omits it entirely); the engine's `offset_limit` paginator always sends
`offset=0` explicitly on the first request. This is a documented, non-data-changing parity
deviation: `offset=0` and an absent `offset` param are semantically identical for every standard
offset-pagination API (including this one, per its own docs), and unknown/redundant query params are
ignored by APIs of this shape — see the parity-deviation ledger below.

No stream declares an `incremental` block: the upstream manifest-only source connector this
migrates advertises full-refresh only (no incremental cursor), matching legacy's own
`streamCatalog()`, which leaves `CursorFields` empty for every stream.

## Write actions & risks

None. Countercyclical is read-only in this port (`writes.json` is intentionally absent), matching
legacy's `Capabilities.Write: false`.

## Known limits

- Legacy's configurable `page_size` (1-1000, default 100) and `max_pages` (0/all/unlimited or a
  positive integer cap) config knobs are not modeled: `streams.json`'s `pagination.page_size` is a
  fixed JSON literal with no config-driven override mechanism (same class of limitation as
  searxng's `page_size`/`max_pages`). A declared-but-unwireable config key is worse than an absent
  one (F6, REVIEW.md), so neither is declared in `spec.json`; the stop threshold is fixed at 100,
  matching legacy's own default exactly.
- Parity deviation (meta-rule, `docs/migration/conventions.md` §5): the engine's `offset_limit`
  paginator sends `offset=0` explicitly on the very first request; legacy omits the `offset` query
  param entirely when its value is 0. This never changes emitted record data for any input legacy
  itself would accept — `offset=0` and an absent `offset` are the same request to any
  standards-conforming offset-pagination API, and this API's own docs confirm unknown/redundant
  params are ignored. ACCEPTABLE.
