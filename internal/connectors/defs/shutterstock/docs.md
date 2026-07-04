# Overview

Shutterstock is a declarative-HTTP connector for the Shutterstock REST API. The legacy streams
(`images`, `videos`, `audio`) keep the hand-written connector's four-field record mapping for
search results. Pass B reviewed the static API reference generated on 2026-07-01 and enumerated
109 documented endpoints: 74 GET endpoints and 35 POST/PATCH/DELETE endpoints.

This bundle now covers the legacy media searches plus additional media, license-list, collection,
editorial, catalog, contributor, computer-vision similarity, and subscription reads. It also exposes
collection/lightbox create, rename, delete, and item-add write actions. Licensing, download,
OAuth/test, reference/autocomplete, binary upload, and DELETE item-removal endpoints remain
excluded in `api_surface.json` with concrete reasons.

## Auth setup

Provide a Shutterstock OAuth access token via the `access_token` secret; it is sent as a Bearer
token (`Authorization: Bearer <access_token>`, matching legacy's `connsdk.Bearer(secret)` at
`shutterstock.go:127`) and is never logged. `base_url` defaults to `https://api.shutterstock.com`
and may be overridden for tests/proxies.

Optional search filters `query`, `sort`, `orientation`, and `category` are passed only to
search-style streams. Detail/list-by-ID streams use comma-separated ID config keys such as
`image_ids`, `video_ids`, `audio_ids`, `sfx_ids`, `editorial_image_ids`, `image_collection_ids`,
`visual_asset_ids`, and `contributor_ids`.

## Streams notes

The three legacy streams (`images`, `videos`, `audio`) use `GET /v2/{images,videos,audio}/search`,
records at `data`, and the same `page`/`per_page` page-number pagination as legacy. Their schemas
and `computed_fields` intentionally preserve legacy's emitted record data: `id`, `description`,
`media_type`, and `updated_at`.

New Pass B streams use `projection: "passthrough"` so the connector does not truncate live
Shutterstock response fields that are outside the documented core schema. Non-paginated reference
or detail endpoints explicitly set `pagination.type: none`; paginated list/search endpoints inherit
the base `page`/`per_page` pagination.

ID-scoped endpoints are modeled with `fan_out.ids_from.config_key`. If a caller leaves the relevant
comma-separated ID config empty, the stream emits no records. When IDs are provided, the engine
runs the endpoint once per ID and stamps the parent ID onto emitted records where useful. Nested
contributor collection detail streams require `contributor_id` plus `contributor_collection_ids`
because the endpoint has two path variables.

## Write actions & risks

`capabilities.write` is true for collection/lightbox metadata writes only. The write actions are:
create, rename, delete, and add-items for image, video, and audio collections; plus create, update,
delete, and add-items for catalog collections. Delete actions are marked destructive and tolerate
404 as an idempotent missing result.

Licensing and download endpoints are not exposed as writes: they create rights-bearing
transactions or binary asset delivery rather than ordinary metadata mutations. DELETE item-removal
endpoints are also excluded because Shutterstock documents `item_id` as a DELETE query parameter,
and the current `writes.json` dialect has no query-param field for write actions.

## Known limits

- **`page_size`/`per_page` is not runtime-configurable.** Legacy accepts `page_size`/`per_page`
  overrides, but the engine's paginator page size is a bundle-authored literal. The spec no longer
  declares a dead `page_size` property; the bundle uses Shutterstock's legacy default of 100.
- **Legacy fallback field names are not modeled.** Legacy maps `description` from
  `description || title`, `media_type` from `media_type || asset_type`, and `updated_at` from
  `updated_time || updated_at`. The declarative `computed_fields` use the primary names exercised
  by the legacy tests; there is no first-non-null template filter.
- **`max_pages` is not runtime-configurable.** Legacy accepts `max_pages`; this bundle leaves
  pagination unbounded and relies on short-page termination.
- **Fixture-mode-only stamped fields are not modeled.** Legacy's synthetic fixture path stamps
  `fixture: true`; the declarative fixture replay harness replaces that test affordance.
- **Reference and helper endpoints are excluded.** Categories, genres, instruments, moods,
  autocomplete, CV keyword suggestions, user/token introspection, OAuth, and test endpoints are
  not syncable account data streams.
