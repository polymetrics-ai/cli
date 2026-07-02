# Overview

Twitter (X) reads tweets and their authors matching a configured search query from the Twitter API
v2 recent-search endpoint, using an App-only Bearer token. This bundle migrates
`internal/connectors/twitter` (the hand-written connector) to a declarative defs bundle at
capability parity; the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Twitter API v2 App-only Bearer token via the `api_key` secret; it is used only for
Bearer auth (`Authorization: Bearer <api_key>`) and is never logged.

## Streams notes

Both streams read the SAME endpoint (`GET /2/tweets/search/recent`) with the SAME query
parameters; they differ only in which JSON path the records live at:

- `tweets` reads the top-level `data[]` array — primary key `id`, cursor field `created_at`.
- `authors` reads the `includes.users[]` expansion array (populated by the shared
  `expansions=author_id` query parameter) — primary key `id`, no cursor field (matches legacy,
  which declares `CursorFields: nil` for this stream: Twitter's recent-search endpoint has no
  server-side incremental filter on authors, and legacy surfaces no client-side one either).

Pagination follows Twitter v2's `meta.next_token` cursor convention (`pagination.type: cursor`
with `token_path: meta.next_token`): the next page's `next_token` query param is read from the
current page's `meta.next_token` body field, and pagination stops when that field is absent or
empty — identical to legacy's `harvest` loop.

The required `query` config value (Twitter v2 search syntax, e.g. `from:example`) is sent
unconditionally, matching legacy's own hard requirement (`Check`/`Read` both fail without it).
`start_date`/`end_date` are optional RFC3339 bounds sent as `start_time`/`end_time` only when
configured (`omit_when_absent: true` on both), matching legacy's own `if start := ...; start != ""`
conditional sends. `page_size` (default `100`) is sent as `max_results`; legacy additionally bounds
it to 10-100 and `max_pages` to a non-negative integer, `all`, or `unlimited` — this bundle does not
enforce those input-validation range checks (the engine dialect has no config-value range
validation primitive), documented as a scope narrowing in Known limits, not a data-emission
deviation: any value a user configures within the range legacy would have accepted behaves
identically here.

## Write actions & risks

None. Twitter is read-only in both legacy and this bundle (`capabilities.write: false`) — posting
tweets or other mutating actions are side-effecting actions inappropriate for a generic reverse-ETL
source.

## Known limits

- `page_size`/`max_pages` range/shape validation (legacy's 10-100 bound on `page_size`, its
  non-negative/`all`/`unlimited` parsing for `max_pages`) is not enforced by this bundle — the
  engine has no declarative config-range-validation primitive. Any value a user would have
  configured successfully under legacy behaves identically here; only the "friendly error before
  the first request" behavior for a wildly out-of-range value is dropped.
- Full Twitter v2 API surface (full-archive search, spaces, DMs, lists, likes, follows, tweet
  writes) is out of scope for this wave; see `api_surface.json`'s `excluded` entries. Only the
  2 legacy-parity streams are implemented.
- `authors` has no incremental cursor field, matching legacy exactly (Twitter's recent-search
  `includes.users` expansion carries no per-author timestamp legacy ever surfaced as a cursor).
