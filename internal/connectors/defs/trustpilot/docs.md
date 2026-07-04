# Overview

Trustpilot is a wave2 fan-out declarative-HTTP migration, expanded in Pass B against Trustpilot's
current published Business API docs (`https://developers.trustpilot.com/`, fetched 2026-07-03 —
see `api_surface.json` for the full endpoint-by-endpoint accounting). It reads Trustpilot
business-unit reviews, review invitations, business-unit profile metadata, and (new in Pass B)
business-unit categories through the Trustpilot REST API
(`GET https://api.trustpilot.com/v1/business-units/<business_unit_id>/...`). This bundle is
capability-parity migrated from `internal/connectors/trustpilot` (the hand-written connector it
migrates) for its original 3 streams; the legacy package stays registered and unchanged until
wave6's registry flip. No write actions were added in this pass — see Write actions & risks.

## Auth setup

Provide a Trustpilot API key via the `api_key` secret; it is sent as the `apikey` query parameter
on every request (an `api_key_query` auth candidate), matching legacy's
`connsdk.APIKeyQuery("apikey", key)` (`trustpilot.go:154`) and is never logged. `business_unit_id`
is required and is substituted into every stream's business-unit-scoped path, matching legacy's
`resourcePath` requirement.

## Streams notes

- `reviews` — `GET /v1/business-units/<business_unit_id>/reviews`, records at `reviews`, primary
  key `["id"]`. Legacy's `reviewRecord` renames the raw API's camelCase `createdAt` to
  `created_at` (falling back to an already-snake_case `created_at` if present); this bundle
  reproduces the rename via a typed coalesce computed field.
- `invitations` — `GET /v1/private/business-units/<business_unit_id>/invitations`, records at
  `invitations`, primary key `["id"]`. Same `createdAt`/`created_at` -> `created_at` fallback as
  `reviews`.
- `business_units` — `GET /v1/business-units/<business_unit_id>` (a single-object response, not a
  list); `records.path` is `"."`, which the engine's record extractor treats as one record when
  the body root is a JSON object (matching legacy's `recordsPath: "."` /
  `connsdk.RecordsAt(body, ".")` behavior exactly). Primary key `["id"]`; `displayName` is renamed
  to `display_name` via `computed_fields`, with fallback to an already-snake_case `display_name`,
  matching legacy's `businessUnitRecord`.
- `categories` (new in Pass B) — `GET /v1/business-units/<business_unit_id>/categories`, records
  at `categories`, primary key `["category_id"]`. Not paginated (`pagination: {"type": "none"}`
  overrides the base's `page_number` spec) — Trustpilot's own docs declare no page/perPage
  parameters on this endpoint, only `country`/`locale` filters this bundle does not (yet) expose as
  config. `computed_fields` renames the raw API's camelCase `categoryId`/`displayName`/`isPrimary`
  to `category_id`/`display_name`/`is_primary`; `is_primary` is a bare single-reference template so
  it keeps its native boolean type (typed `computed_fields` extraction, conventions.md §3).

Pagination (`reviews`/`invitations` only; `business_units` and `categories` are unpaginated) is
`page_number` (`page`/`perPage` query params, `page_size: 100` matching legacy's
`defaultPageSize`/`maxPageSize`) capped at `max_pages: 1` (legacy's own `defaultMaxPages`), matching
legacy's default: only the first page is fetched unless legacy's config override
(`max_pages=all`/`unlimited`/N) raises the cap — see Known limits for why that per-request override
is not modeled. A short page (fewer than `page_size` records) still stops pagination early exactly
like legacy's `len(records) < pageSize` check.

## Write actions & risks

None added in Pass B, and legacy has none either (`Write` always returns
`connectors.ErrUnsupportedOperation`; `metadata.json` declares `capabilities.write: false` and no
`writes.json` file exists). **Every mutation-capable and every private/OAuth-gated endpoint in
Trustpilot's current API surface (review replies, review tags, invitation send/delete, product-review
conversations, private reviews, private products) requires a "Business user OAuth Token" credential**
(`docs/migration/conventions.md` §6's `AUTH_COMPLEX`-adjacent reasoning, though this is a scope
decision rather than a dialect gap: `oauth2_client_credentials`/`bearer` auth modes both exist in the
engine dialect and could express this credential shape mechanically) — legacy's spec only ever
declared a plain `apikey` query-parameter credential and never implemented any write at all, so
adding OAuth would be a new credential-shape/spec.json change, a distinct follow-on decision outside
this Pass B pass's mandate to preserve accepted-input behavior. See `api_surface.json` for every
OAuth-gated endpoint this excludes, each tagged `requires_elevated_scope` with this exact reasoning.

## Known limits

- **The legacy `business_units` (`GET /v1/business-units/{business_unit_id}`) and `invitations`
  (`GET /v1/private/business-units/{business_unit_id}/invitations`) endpoint shapes no longer
  appear in Trustpilot's current public Business Units / Invitation API docs** (fetched
  2026-07-03). The modern documented equivalent of `business_units` is
  `GET /v1/business-units/{businessUnitId}/profileinfo` (a materially different, much larger field
  set — company contact/address/social-media info rather than legacy's narrow `id`/`display_name`
  pair); the modern invitations surface has moved entirely to a separate
  `invitations-api.trustpilot.com` host and requires a Business user OAuth Token, not the
  `apikey` query-parameter credential legacy's `invitations` stream authenticates with. Both
  streams are left exactly as legacy implemented them (unchanged parity behavior, meta-rule
  compliant) rather than silently repointed at a different endpoint shape or credential — that
  would be an accepted-input behavior change, not a Pass B capability expansion. `api_surface.json`
  documents both the legacy path (`covered_by`) and every real modern equivalent this pass
  discovered (`profileinfo`, the OAuth-gated invitations-api.trustpilot.com endpoints), each
  correctly categorized rather than silently omitted.
- **Per-request `max_pages` override is not modeled.** Legacy accepts a `max_pages` config value
  (`"all"`/`"unlimited"` or a non-negative integer) to raise or remove the `defaultMaxPages: 1`
  cap at read time (`trustpilot.go:186-199`, `configuredMaxPages`). The engine's `base.pagination`
  block is a static, bundle-load-time spec with no per-request config indirection for `max_pages`,
  so this bundle fixes `max_pages: 1` — matching legacy's own default behavior exactly, but not
  legacy's opt-in override to page through the full result set. Declaring a `max_pages` spec
  property would be dead config (no template consumes it), so it is not declared at all (F6,
  `docs/migration/conventions.md`).
- **Per-request `page_size` override is not modeled**, for the identical structural reason
  (`base.pagination.page_size` is a static field, not config-templated); `page_size: 100` matches
  legacy's `defaultPageSize`/`maxPageSize` (both were 100 in legacy, so there was no meaningful
  override range beyond the default in practice).
- **`base_url` scheme/host validation is not re-implemented as a distinct pre-flight check.**
  Legacy validates `base_url` with a dedicated parse-and-check step before every request
  (`trustpilot.go:168-184`); the engine performs the equivalent validation implicitly whenever a
  templated URL is actually dispatched (an invalid or schemeless URL fails the request itself),
  so no observable behavior differs for any valid or invalid `base_url` legacy would have accepted
  or rejected.
- Fixture pagination is single-page (`max_pages: 1` matches legacy's default, so no bundle
  configuration ever requests a second page — there is nothing for a second fixture page to
  prove); `pagination_terminates` exercises the real short-circuit (exactly 1 request made per
  paginated stream).

- **Per-request `max_pages` override is not modeled.** Legacy accepts a `max_pages` config value
  (`"all"`/`"unlimited"` or a non-negative integer) to raise or remove the `defaultMaxPages: 1`
  cap at read time (`trustpilot.go:186-199`, `configuredMaxPages`). The engine's `base.pagination`
  block is a static, bundle-load-time spec with no per-request config indirection for `max_pages`,
  so this bundle fixes `max_pages: 1` — matching legacy's own default behavior exactly, but not
  legacy's opt-in override to page through the full result set. Declaring a `max_pages` spec
  property would be dead config (no template consumes it), so it is not declared at all (F6,
  `docs/migration/conventions.md`).
- **Per-request `page_size` override is not modeled**, for the identical structural reason
  (`base.pagination.page_size` is a static field, not config-templated); `page_size: 100` matches
  legacy's `defaultPageSize`/`maxPageSize` (both were 100 in legacy, so there was no meaningful
  override range beyond the default in practice).
- **`base_url` scheme/host validation is not re-implemented as a distinct pre-flight check.**
  Legacy validates `base_url` with a dedicated parse-and-check step before every request
  (`trustpilot.go:168-184`); the engine performs the equivalent validation implicitly whenever a
  templated URL is actually dispatched (an invalid or schemeless URL fails the request itself),
  so no observable behavior differs for any valid or invalid `base_url` legacy would have accepted
  or rejected.
- Fixture pagination is single-page (`max_pages: 1` matches legacy's default, so no bundle
  configuration ever requests a second page — there is nothing for a second fixture page to
  prove); `pagination_terminates` exercises the real short-circuit (exactly 1 request made per
  paginated stream).
