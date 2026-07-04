# Overview

AppFollow is a declarative-HTTP migration, Pass-B full-surface expanded. It reads AppFollow
account users, app collections, app lists, reviews, review summaries, rating breakdowns/history,
ASO keywords, rankings, and version/what's-new metadata through the AppFollow REST API v2
(`https://api.appfollow.io/api/v2/...`), and writes review replies/tags/notes, ASO keyword edits,
and account user/app-collection/tracked-app management actions. This bundle originally targeted
capability parity with `internal/connectors/appfollow` (the hand-written connector it migrates,
covering 4 streams, read-only); Pass B (`docs/migration/conventions.md` §8, `api_surface.json`)
researched AppFollow's full published API v2 reference
(https://docs.api.appfollow.io/reference/overview, 44 documented operations) and expanded to 11
streams + 11 writes covering every practical list/detail resource and every dialect-expressible
mutation. The legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide an AppFollow API token via the `api_secret` secret; it is sent as the
`X-AppFollow-API-Token` header and is never logged, matching legacy's
`connsdk.APIKeyHeader(appfollowTokenHeader, secret, "")` (`appfollow.go:322`). `base_url` defaults
to `https://api.appfollow.io/api/v2` and may be overridden for tests/proxies.

## Streams notes

`users` and `app_collections` are single-request, unpaginated `GET` endpoints (`pagination`
omitted; legacy's reporting endpoints return their full result in one body, matching
`readSimple`):

- `users` — `GET /account/users`, records at the response root (`records.path: ""`), primary key
  `["id"]`.
- `app_collections` — `GET /account/apps`, records at the `apps` array, primary key `["id"]`.

`app_lists` and `ratings` are sub-resource fan-out reads, now expressible via the S4 engine
mini-wave's `stream.fan_out` primitive (`docs/migration/conventions.md` §3):

- `app_lists` — `GET /account/apps/app`, one request per `app_collection_ids` config entry
  (`fan_out.ids_from.config_key`), forwarded as the `apps_id` query parameter
  (`fan_out.into.query_param`), records at `apps_app`, primary key `["app_id"]`. The fan-out id is
  stamped onto every emitted record's `app_collection_id` field (`fan_out.stamp_field`), matching
  legacy's `readAppLists` (`appfollow.go:157`) `if rec["app_collection_id"] == nil { rec[...] = id
  }` injection.
- `ratings` — `GET /meta/ratings`, one request per `ext_ids` config entry, forwarded as the
  `ext_id` query parameter, records at `ratings.list` (the dotted-path convention selects straight
  into the nested rows array, no special-casing needed), primary key
  `["ext_id", "date", "country"]`. The fan-out id is stamped onto every row's `ext_id` field,
  matching legacy's `readRatings` (`appfollow.go:205`) ext_id injection.

Both fan-out streams require their respective config list (`app_collection_ids` for `app_lists`,
`ext_ids` for `ratings`) — comma-separated, split/trimmed/empty-entries-dropped, matching legacy's
`splitList`.

**Pass B additions** — 7 new streams, all fanning out over `ext_ids` (the same config list
`ratings` already requires; every one of these is scoped per tracked app, exactly like `ratings`):

- `reviews` — `GET /reviews`, requires `report_from`/`report_to` config (AppFollow's own `from`/
  `to` query params are hard-required per its OpenAPI spec), records at the response root, primary
  key `["id"]`.
- `reviews_summary` — `GET /reviews/summary`, aggregate per-date/country/version review stats,
  primary key `["ext_id", "date", "country"]`.
- `keywords` — `GET /aso/keywords`, tracked ASO keyword positions, primary key
  `["ext_id", "country", "device", "date"]`.
- `rankings` — `GET /meta/rankings`, category ranking positions, primary key
  `["ext_id", "country", "device", "genre_id", "date"]`.
- `versions` — `GET /meta/versions`, requires `report_country` config (hard-required by the API;
  defaults to `us`), version/release history, primary key `["ext_id", "version", "country"]`.
- `versions_whatsnew` — `GET /meta/versions/whatsnew`, release-notes-only variant of `versions`,
  primary key `["ext_id", "version", "country"]`.
- `ratings_history` — `GET /meta/ratings/history`, requires `report_store` config (hard-required;
  defaults to `itunes`), a fuller per-date rating history than the pre-existing `ratings` stream's
  snapshot, primary key `["ext_id", "date", "country", "version"]`.

None of the 11 streams are incremental — AppFollow's v2 API has no server-side cursor filter
legacy uses (no `CursorFields` declared on any legacy stream), and the newly-researched endpoints'
own filters are `from`/`to` explicit date-range parameters, not a repeat-sync cursor contract.

## Write actions & risks

Pass B added 11 write actions (`capabilities.write` is now `true`); every action's `risk` field in
`writes.json` states whether it is externally visible (a public review reply the app-store shows
end users) or irreversible (account/collection/app/user removal) — approval is required for all of
them:

- `reply_to_review` (`POST /reviews/reply`) — posts a **public** developer reply to a live
  app-store review; cannot be unsent programmatically once posted.
- `update_review_tags` (`POST /reviews/tags`) — overwrites a review's tag set.
- `update_review_notes` (`POST /reviews/notes`) — overwrites a review's internal note.
- `edit_keywords` (`POST /aso/keywords`) — replaces the tracked ASO keyword list for a
  country/device pair.
- `add_user` / `update_user` / `remove_user` (`POST`/`PATCH`/`DELETE /account/users`) — account
  user management; `remove_user` is irreversible.
- `add_collection` / `remove_collection` (`POST`/`DELETE /account/apps`) — app-collection
  management; `remove_collection` irreversibly drops every app tracked under it.
- `add_app` / `remove_app` (`POST`/`DELETE /account/apps/app`) — tracked-app management within a
  collection; `remove_app` is irreversible.

All 4 `DELETE` actions (`remove_user`, `remove_collection`, `remove_app`) and `remove_app` send a
JSON request body carrying the identifying fields (AppFollow's real DELETE endpoints take a body,
not path parameters) — modeled via `body_fields` (an explicit allow-list) since there is no
`{{ record.id }}`-shaped path segment to exclude via `path_fields`; `kind: "delete"` is declared for
audit/risk classification even though none of them declare `missing_ok_status` (AppFollow's DELETE
endpoints are not documented as returning a distinguishable "already gone" 404 the way idempotent
REST deletes typically do, so no missing-ok status code is asserted — an unexpected non-2xx is
always treated as a genuine write failure, the safe default).

## Known limits

- **Pass B streams' schemas are field-name-accurate but response-envelope-unconfirmed (documented
  research gap, not a guess).** AppFollow's public OpenAPI reference declares every 200 response as
  a bare `"schema": {}` for all 7 new streams — the vendor genuinely does not publish response
  bodies in its machine-readable spec. Field names for `reviews` (`title`/`store`/`is_answer`/
  `was_changed`/`user_id`/`ext_id`/`review_id`/`dt`/`created`/`content`/`app_version`/`updated`/
  `app_id`/`time`/`rating`/`date`/`locale`/`rating_prev`/`author`/`id`) and `keywords`
  (`date`/`store`/`device`/`total`/`ext_id`/`country`/`page`/`no_pos`/`pos`/`popularity`) and
  `ratings_history` (`from`/`to`/`store`/`ext_id`/`version`/`country`/`period`/`offset`/`limit`/
  `total`/`date`/`stars`/`stars1..5`/`avg_rating`) come from AppFollow's own published "Response
  Body Parameters" reference pages (a field-glossary the docs site ships alongside, but not inside,
  the OpenAPI operation) — these are real, vendor-documented field names, not invented ones.
  `rankings`/`versions`/`versions_whatsnew` have no equivalent published field-glossary page at
  all; their schemas here use the same field vocabulary AppFollow uses elsewhere in its docs
  (`position`/`category`/`genre_id`/`release_date`/`whats_new`/`last_modified`) as the
  best-available, still-undocumented-exact-shape approximation. `records.path: ""` (response root)
  is assumed uniformly for all 7 new streams by analogy with `users`' and `app_lists`' sibling
  `GET`-list conventions; none of this could be confirmed against a live response since the
  connector was authored credential-free. A capability-expansion agent with live AppFollow
  credentials should verify the actual response envelope/field set against a real account and
  correct `schemas/*.json`/`records.path` if it differs — this is flagged here specifically so that
  correction is a schema/fixture edit, not a rediscovery.
- **`ratings` sibling metadata is stamped before projection.** AppFollow's `/meta/ratings`
  response nests each day's rating breakdown under `ratings.list`, with `ext_id`/`store` as
  siblings of `list`. `fan_out.stamp_field` restores legacy's `ext_id` fallback from the request id,
  and `response_fields.store` restores legacy's `ratings.store` copy onto every emitted row.
- **`app_lists`'s `app_collection_id` is typed as `string`, not legacy's `integer` (ACCEPTABLE,
  documented deviation).** `fan_out.stamp_field` always writes the fan-out id as the STRING split
  out of `app_collection_ids` (matching every other `stamp_field`-using bundle in this repo, e.g.
  cisco-meraki's `organizationId`, metricool's `blogId` — see `docs/migration/conventions.md` §3),
  whereas legacy's raw API response and `appListRecord` mapping carry `app_collection_id` as a
  JSON integer. The emitted VALUE is identical (the same collection id), only its JSON type
  differs (`"11"` vs `11`); every downstream consumer that treats ids as opaque strings is
  unaffected, but a warehouse column-type comparison would see `string` here vs `integer` in
  legacy's schema. Widening/retyping avoids silently emitting a value that would fail the
  declared schema, per the meta-rule (`docs/migration/conventions.md` §5) — this is the identical
  class of deviation already accepted for every other `fan_out`-migrated bundle's stamped id field.
- **Config-driven collection-id auto-discovery is not modeled.** Legacy's `app_lists` stream falls
  back to auto-discovering collection ids from `/account/apps` when `app_collection_ids` is unset
  (`discoverCollectionIDs`, `appfollow.go:185`). `fan_out.ids_from` supports this shape too (the
  `request` variant — a preliminary GET, fully paginated, extracting an `id_field` off each
  record), and `/account/apps`'s `apps` array does carry an `id` field per collection, so this
  fallback IS expressible with the dialect as it stands today; it was intentionally left out of
  this migration to keep `app_lists`'s required config surface explicit and mirror this bundle's
  existing `ext_ids`-required (no-fallback) precedent for `ratings`. A future capability-expansion
  pass MAY add the `request`-based `ids_from` variant as an alternative when `app_collection_ids`
  is unset, without needing any further engine changes.
