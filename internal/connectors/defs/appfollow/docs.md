# Overview

AppFollow is a declarative-HTTP migration. It reads AppFollow account users, app collections,
per-collection app lists, and per-app rating breakdowns through the read-only AppFollow REST API
v2 (`GET https://api.appfollow.io/api/v2/...`). This bundle targets full capability parity with
`internal/connectors/appfollow` (the hand-written connector it migrates) across all 4 legacy
streams; the legacy package stays registered and unchanged until wave6's registry flip.

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

None of the 4 streams are incremental — AppFollow's v2 API has no server-side cursor filter
legacy uses (no `CursorFields` declared on any legacy stream).

## Write actions & risks

None. Legacy's `Write` unconditionally returns `connectors.ErrUnsupportedOperation`;
`capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- **`ratings`'s `store` field is not populated per-row (ACCEPTABLE, documented deviation).**
  AppFollow's `/meta/ratings` response nests each day's rating breakdown under `ratings.list`,
  with `ext_id`/`store` as SIBLINGS of `list` (not fields on each row) —
  `{"ratings":{"ext_id":"...","store":"...","list":[{...no ext_id/store...}]}}`. Legacy's
  `ratingRows` (`appfollow.go:230`) reads both sibling fields once per request and injects them
  onto every extracted row. The engine's `computed_fields`/`fan_out.stamp_field` mechanisms only
  ever see the INDIVIDUAL extracted row (the object at `records.path`, here `ratings.list`'s
  element) — there is no reference to a sibling path elsewhere in the same page body. `ext_id` is
  still populated exactly (via `fan_out.stamp_field`, whose value is the same config-supplied id
  legacy's fallback-to-config-id path also uses), but `store` is never set by this bundle — it
  will read as `null` on every `ratings` record, a real, narrow field-level parity gap, not a
  silent one. Closing this fully would need a "stamp a value read from a named sibling path in the
  page body" primitive the dialect does not have today; adding one is out of scope for a
  single-connector fan-out migration.
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
