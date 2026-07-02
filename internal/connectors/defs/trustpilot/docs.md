# Overview

Trustpilot is a wave2 fan-out declarative-HTTP migration. It reads Trustpilot business-unit
reviews, review invitations, and business-unit profile metadata through the Trustpilot REST API
(`GET https://api.trustpilot.com/v1/business-units/<business_unit_id>/...`). This bundle is
capability-parity migrated from `internal/connectors/trustpilot` (the hand-written connector it
migrates); the legacy package stays registered and unchanged until wave6's registry flip.

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
  reproduces the rename via a `computed_fields` entry (`"created_at": "{{ record.createdAt }}"`).
  The legacy fallback-to-snake_case branch is dead in practice (Trustpilot's real API always sends
  camelCase `createdAt`); see Known limits.
- `invitations` — `GET /v1/private/business-units/<business_unit_id>/invitations`, records at
  `invitations`, primary key `["id"]`. Same `createdAt` -> `created_at` rename as `reviews`.
- `business_units` — `GET /v1/business-units/<business_unit_id>` (a single-object response, not a
  list); `records.path` is `"."`, which the engine's record extractor treats as one record when
  the body root is a JSON object (matching legacy's `recordsPath: "."` /
  `connsdk.RecordsAt(body, ".")` behavior exactly). Primary key `["id"]`; `displayName` is renamed
  to `display_name` via `computed_fields`, matching legacy's `businessUnitRecord`.

Pagination (`reviews`/`invitations` only; `business_units` is a single-object endpoint with no
pagination block) is `page_number` (`page`/`perPage` query params, `page_size: 100` matching
legacy's `defaultPageSize`/`maxPageSize`) capped at `max_pages: 1` (legacy's own
`defaultMaxPages`), matching legacy's default: only the first page is fetched unless legacy's
config override (`max_pages=all`/`unlimited`/N) raises the cap — see Known limits for why that
per-request override is not modeled. A short page (fewer than `page_size` records) still stops
pagination early exactly like legacy's `len(records) < pageSize` check.

## Write actions & risks

None. Legacy `Write` always returns `connectors.ErrUnsupportedOperation`; `metadata.json` declares
`capabilities.write: false` and no `writes.json` file exists, matching legacy exactly.

## Known limits

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
