# Overview

Partnerize reads Partnerize conversions, campaigns, and publishers through the REST API. This
bundle migrates `internal/connectors/partnerize` (the hand-written legacy connector) to a
declarative Tier-1 defs bundle; the legacy package stays registered and unchanged until the wave6
registry flip.

## Auth setup

Provide both a Partnerize `application_key` and `user_api_key` secret; they are sent as HTTP Basic
auth (application_key as username, user_api_key as password) and are never logged, matching
legacy's `connsdk.Basic(applicationKey, userAPIKey)` exactly.

## Streams notes

Three streams: `conversions` (`GET /conversions`), `campaigns` (`GET /campaigns`), `publishers`
(`GET /publishers`). All share the same shape: records at `data`, primary key `["id"]`, incremental
cursor field `created_at`. Pagination is `offset_limit` (`limit`/`offset` query params, `page_size:
100` matching legacy's default `limit`) — the engine's offset paginator advances by exactly
`page_size` per page and stops once a page returns fewer than `page_size` records, closely tracking
legacy's own `harvestOffset` stop rule (`len(records) < limit || len(records) == 0 ||
(total > 0 && offset+len(records) >= total)`): the engine's paginator's short-page-count stop
subsumes legacy's first two conditions exactly; legacy's additional `meta.total_count`-aware early
stop (stopping before a short page if the running total already reaches the reported total) is a
defensive early-exit legacy adds on top — a page that is not yet short but whose count has already
reached `total_count` would, in the rare case that the API always returns full pages until
mid-page, no longer be reachable by the engine's own paginator, since the engine's stop condition
depends only on the current page's record count vs `page_size`, not on a cumulative comparison
against a separate `meta.total_count` field (which the pagination dialect's `offset_limit` type has
no path-reference for at all). In every real-world case observed (a `total_count`-truncated final
page is also necessarily a short page), the two stop rules agree; the divergent corner (an API
returning a full-`page_size` final page that also happens to reach `total_count` exactly on a page
boundary) is undetectable from `meta.total_count` alone, only from the following empty page, which
the engine's paginator observes on the next iteration in that same edge case, converging identically
one round later.

## Write actions & risks

None. This connector is read-only, matching legacy (`Capabilities.Write: false`,
`Write` returns `ErrUnsupportedOperation`).

## Known limits

- Only `conversions`, `campaigns`, and `publishers` are implemented, matching legacy's exact stream
  set. Clicks, reporting exports, and webhooks are out of scope for this wave; see
  `api_surface.json`'s `excluded` entries.
- Legacy's runtime `limit`/`max_pages` config overrides are not exposed as `spec.json` properties:
  `page_size` on an `offset_limit` pagination block is a static integer, not `{{ }}`-templatable in
  this engine version, so a declared-but-unwireable `spec.json` property would be dead config (F6).
  This bundle uses a fixed `page_size: 100` (matching legacy's default `limit`) and leaves the page
  count unbounded (matching legacy's `max_pages=0`/unlimited default).
- See "Streams notes" above for the one edge-case divergence between legacy's `meta.total_count`-
  aware early stop and the engine's page-size-only stop signal — both terminate identically for
  every response shape actually observed from the Partnerize API (a truncated final page is always
  a short page), so this is not treated as a documented parity deviation, only a noted theoretical
  corner.
