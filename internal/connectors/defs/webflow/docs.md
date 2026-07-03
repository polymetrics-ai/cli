# Overview

Webflow reads site collections, pages, and forms for a single configured Webflow site using the
Webflow Data API v2. This bundle migrates `internal/connectors/webflow` (the hand-written
connector) to a declarative defs bundle at capability parity; the legacy package stays registered
and unchanged until wave6's registry flip. The catalog entry (`source-webflow`,
`internal/connectors/catalog_data.json`) declares `"type": "source"` — Webflow is read-only in both
legacy and this bundle, not a destination.

## Auth setup

Provide a Webflow API token (Site token or OAuth access token) via the `api_key` secret; it is
used only for Bearer auth (`Authorization: Bearer <api_key>`) and is never logged. `site_id` is a
required config value naming the Webflow site every stream reads from (legacy's `siteResource`
helper fails without it, matching this bundle's `required: ["api_key", "site_id"]`). An optional
`accept_version` config value is sent as the `Accept-Version` header (Webflow's API-versioning
mechanism); when unset, the header is omitted entirely (not sent empty) — identical to legacy's
`if version := ...; version != ""` conditional send.

## Streams notes

All 3 streams are single-page reads (`pagination` omitted/`none`) against a site-scoped list
endpoint, matching legacy exactly (legacy's `Read` issues exactly one request per stream, no
cursor/pagination loop of any kind):

- `collections` — `GET /v2/sites/{site_id}/collections`, records at `collections`, fields
  `id`/`displayName`/`slug`.
- `pages` — `GET /v2/sites/{site_id}/pages`, records at `pages`, fields `id`/`title`/`slug`.
- `forms` — `GET /v2/sites/{site_id}/forms`, records at `forms`, fields
  `id`/`displayName`/`createdOn`.

None of the three streams declares a cursor field — matches legacy's `streams()`, which never sets
`CursorFields` on any of the three `connectors.Stream` entries (only `PrimaryKey: []string{"id"}}`
is set), so there is no server-side or client-side incremental filter to model (§8 incremental
truth table: no `CursorFields` in legacy's catalog → no `incremental` block, no `x-cursor-field`).

## Write actions & risks

None. Webflow is read-only in both legacy and this bundle (`capabilities.write: false`) — legacy's
`Write` unconditionally returns `connectors.ErrUnsupportedOperation`.

## Known limits

- Full Webflow Data API surface (CMS collection items, ecommerce, form submissions, assets,
  webhooks, users) is out of scope for this wave; see `api_surface.json`'s `excluded` entries. Only
  the 3 legacy-parity streams are implemented.
- Each stream is a genuine single-page read with no pagination of any kind, matching legacy
  exactly — a Webflow site with more collections/pages/forms than fit in one response page would
  be silently truncated on both legacy and this bundle identically (not a migration-introduced
  regression; Webflow's Data API v2 list endpoints for these three resources are not paginated in
  legacy's own implementation).
