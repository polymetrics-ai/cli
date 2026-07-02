# Overview

Formbricks is an open-source survey/experience-management platform. This bundle reads five
read-only streams — `surveys`, `responses`, `action_classes`, `attribute_classes`, `webhooks` —
through the Formbricks management API (`https://app.formbricks.com/api/v1/<resource>` by default,
or a self-hosted instance's equivalent). It migrates `internal/connectors/formbricks` (the
hand-written legacy connector), which stays registered and unchanged until wave6's registry flip.
Formbricks has no write/mutation surface exposed for reverse ETL in this connector, matching legacy:
`capabilities.write` is `false` and this bundle ships no `writes.json`.

## Auth setup

Formbricks authenticates every request with an `X-API-Key` request header (`streams.json`
`base.auth`'s `api_key_header` mode), sourced from the required `api_key` secret, matching legacy's
`connsdk.APIKeyHeader` wiring (`formbricks.go:236-240`) exactly. `base_url` defaults to
`https://app.formbricks.com/api/v1` (legacy's `formbricksDefaultBaseURL`) via `spec.json`'s
`default`; override it for a self-hosted Formbricks deployment.

## Streams notes

All five streams share the same envelope: records are extracted from the top-level `data` array
(matching legacy's `connsdk.RecordsAt(resp.Body, "data")`). Every stream needs a `computed_fields`
rename from the raw API's camelCase field names (`environmentId`, `createdAt`, `updatedAt`,
`surveyId`, `contactId`) to this bundle's snake_case schema properties — plain schema projection
copies by exact key match only, so without the rename these fields would silently drop (mirrors
searxng's documented `publishedDate`-vs-`published_date` pattern, `docs/migration/conventions.md`
§5 item 5).

Only `responses` paginates: Formbricks' management API accepts `limit`/`skip` offset pagination for
that endpoint (`offset_limit` type, `limit_param: "limit"`, `offset_param: "skip"`), matching
legacy's `harvest` loop (`formbricks.go:145-181`) exactly — the loop stops on a short page (fewer
than the requested size). The remaining four streams (`surveys`, `action_classes`,
`attribute_classes`, `webhooks`) return their entire collection in a single, unpaginated page,
matching legacy's `endpoint.paginated == false` branch (a single request, no query params sent).

`webhooks`' raw API fields `surveyIds`/`triggers` are emitted verbatim (unrenamed, camelCase) in
both legacy (`formbricksWebhookRecord`, `streams.go:187-198`) and this bundle's schema — an
intentional exception to the otherwise snake_case schema, preserved for exact parity rather than
"fixed" to snake_case, since a rename here would be a data-shape change legacy itself never made.

No stream has a working incremental capability: legacy declares `CursorFields: []string{"updated_at"}`
on `surveys`/`responses`/`action_classes`/`attribute_classes` (manifest metadata only — `webhooks`
declares none at all) but never actually filters or advances a read by it; this bundle mirrors that
exactly via `x-cursor-field` on those same four schemas (and no cursor field on `webhooks`), with no
`incremental` block on any stream.

## Write actions & risks

None. Formbricks' write/mutation surface is not exposed by this connector (legacy:
`Capabilities.Write: false`, `Write` returns `ErrUnsupportedOperation`); `capabilities.write` is
`false` and this bundle ships no `writes.json`.

## Known limits

- **`page_size`/`max_pages` config overrides are not modeled.** Legacy accepts optional `page_size`
  (1-100, default 50) and `max_pages` (default unlimited) config keys read at request time
  (`formbricksPageSize`/`formbricksMaxPages`, `formbricks.go:271-298`), applied uniformly across
  every stream (even non-paginated ones, where `page_size` is simply unused). The engine's
  `PaginationSpec.PageSize`/`MaxPages` fields are plain fixed JSON integers baked into
  `streams.json` — there is no templating/config-driven override mechanism for them. This bundle
  declares a fixed `page_size: 2` for the `responses` stream (chosen small so the required 2-page
  conformance fixture is realistic and exercises the short-page stop rule; legacy's own default is
  50) and no `max_pages` cap (unbounded, matching legacy's own default). Neither `page_size` nor
  `max_pages` is declared in `spec.json` (F6, `docs/migration/conventions.md`: dead, unwireable
  config is worse than absent config).
