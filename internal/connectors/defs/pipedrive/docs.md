# Overview

Pipedrive is a read-only declarative migration of `internal/connectors/pipedrive` (legacy Go
connector). It reads Pipedrive deals, persons, organizations, activities, products, and users
through Pipedrive's REST API v1. This bundle is capability-parity with legacy; legacy stays
registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Pipedrive API token via the `api_token` secret. It is sent as the documented `api_token`
query-string parameter on every request (`auth.mode: api_key_query`), matching legacy's
`connsdk.APIKeyQuery("api_token", key)` exactly — legacy accepts either an `api_token` or (as a
fallback) an `api_key` secret name (`firstNonEmpty(secret(cfg,"api_token"), secret(cfg,"api_key"))`);
this bundle declares only `api_token` since the `api_key` fallback name has no separate documented
meaning and every legacy call site that actually authenticates supplies `api_token`. Never logged.

## Streams notes

All 6 streams (`deals`, `persons`, `organizations`, `activities`, `products`, `users`) pass through
the RAW record with no field-filtering (`"projection": "passthrough"`) — legacy's `Read` emits
`connectors.Record(rec)` directly from the decoded `data` array with no `mapRecord`-style shaping
function. Pagination follows Pipedrive's `additional_data.pagination.next_start` cursor convention
(`pagination.type: cursor` with `token_path: additional_data.pagination.next_start`, `cursor_param:
start`): the next page's `start` query value is read straight from the previous response body, and
pagination stops when `next_start` is absent/null/empty — identical to legacy's
`strings.TrimSpace(next) == ""` stop check. No `stop_path` is declared: legacy's own stop condition
looks only at `next_start`, never at `more_items_in_collection`, so adding a `stop_path` gate on that
field would be new behavior, not a port.

`deals`, `persons`, and `organizations` send `since_timestamp` (an optional-query entry,
`omit_when_absent: true`) when `replication_start_date` is configured; `activities` sends the same
value as `since` (a different param name — matches legacy's per-endpoint `sinceParam` map:
`deals`/`persons`/`organizations` → `since_timestamp`, `activities` → `since`, `products`/`users` →
`""`, i.e. no since param wired at all for those two, matching this bundle's `products`/`users`
streams declaring no such query key). Legacy resolves the since value from `req.State["cursor"]`
first, falling back to `req.Config.Config["replication_start_date"]`; this bundle does not attempt to
express the state-cursor half of that fallback as a stream-level `incremental` block because none of
Pipedrive's since params here is wired as this dialect's `incremental.request_param` (no per-stream
`x-cursor-field`/state-cursor round-trip is declared) — see Known limits.

## Write actions & risks

None. Legacy `Write` always returns `connectors.ErrUnsupportedOperation`; `metadata.json` declares
`capabilities.write: false` and no `writes.json` file exists, matching legacy exactly.

## Known limits

- Legacy's `since`/`since_timestamp` value can come from EITHER the sync's persisted
  `state["cursor"]` OR the static `replication_start_date` config value (state takes precedence).
  This bundle only wires the static config half (`config.replication_start_date`) via the
  optional-query dialect; the state-cursor half is not modeled as a stream `incremental` block
  because no stream here declares `x-cursor-field`/`incremental.cursor_field` — the raw
  `update_time`/`add_time` fields legacy tracks are ordinary passthrough fields, not a declared
  incremental cursor tied to a `request_param`. This means every read is effectively a
  config-anchored (not truly resumable, state-cursor-advancing) read from wave2's bundle's
  perspective. Out of scope for wave2 fan-out; a future capability-expansion pass can promote these
  streams to full `incremental` blocks once the since-param-vs-state-cursor precedence is worth
  modeling declaratively.
- `page_size`/`max_pages` config knobs (legacy's client-side page-size and page-count bounds,
  1-500 and 0/all/unlimited respectively) have no bundle-level equivalent; `limit=100` is a fixed
  per-stream `query` value (matching legacy's `defaultPageSize`) and pagination relies solely on the
  cursor paginator's own token-based stop signal (no `MaxPages` cap declared), matching legacy's own
  unbounded default. Out of scope for wave2 fan-out (Pass B).
- The 2-page conformance fixture lives on the `deals` stream (`fixtures/streams/deals/page_1.json`
  /`page_2.json`); `persons`/`organizations`/`activities`/`products`/`users` ship single-page
  fixtures only, since they share the identical `cursor`+`token_path` pagination shape already proven
  by `deals`'s fixture pair.
