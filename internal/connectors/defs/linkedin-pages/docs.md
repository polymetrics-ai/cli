# Overview

LinkedIn Pages is a read-only company-page source connector. It reads the LinkedIn organization
(company page) profile, lifetime follower statistics, lifetime share (content) statistics, and
total first-degree follower count for one configured `org_id`, through the LinkedIn Community
Management REST API (`https://api.linkedin.com/rest`). This bundle migrates
`internal/connectors/linkedin-pages` (the hand-written legacy connector); the legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

Provide a LinkedIn member OAuth2 `access_token` (secret, sent as `Authorization: Bearer`) and the
`org_id` of the company page to read. `org_id` is a plain LinkedIn organization identifier (a bare
numeric id, e.g. `123`), not a credential — it is declared as ordinary `config`, not `x-secret`,
specifically so it can be stamped onto every emitted record via `computed_fields` (the engine's
`computed_fields` Vars environment deliberately never resolves `secrets.*`, conventions.md §3 — an
`x-secret` org_id could not be stamped onto records at all). Legacy's `access_token` resolution
tolerated two secret-key spellings (`credentials.access_token` and a bare `access_token`
fallback); this bundle declares only `access_token`, since the engine's `secrets.<key>` reference
resolves a single flat map key — a dotted `credentials.access_token` reference would only ever
look up the literal key `credentials` (`interpolate.go`'s `resolveRefValue` uses `segs[1]` alone as
the secret key), not a nested `credentials` object. The refresh_token OAuth2 exchange itself (when
the long-lived access token needs refreshing) is performed by the operator/agent layer before this
connector runs, exactly as in legacy — this bundle only ever consumes the resolved bearer token.

Every request also carries the mandatory `LinkedIn-Version` header (`config.linkedin_version`,
default `202601`, LinkedIn's monthly YYYYMM version scheme) and
`X-Restli-Protocol-Version: 2.0.0`, matching legacy's `requester()` exactly.

## Streams notes

The 4 streams are heterogeneous, mirroring legacy's per-endpoint routing table exactly:

- **`organizations`**: a single-object GET at `/organizations/{{ config.org_id }}` (`records.path:
  "."`, `pagination: none`). The raw Organization Lookup response is entirely camelCase
  (`vanityName`, `localizedName`, `$URN`, ...); `computed_fields` renames each to legacy's exact
  snake_case output field (`vanity_name`, `localized_name`, `urn`, ...) — plain schema projection
  only copies fields whose name matches exactly, so every renamed field needs an explicit
  `computed_fields` entry (conventions.md §2's schema-as-projection rule). `id` needs no rename
  (already lowercase in both raw and legacy output).
- **`follower_statistics`** / **`share_statistics`**: finder endpoints
  (`/organizationalEntityFollowerStatistics`, `/organizationalEntityShareStatistics`) scoped by
  `q=organizationalEntity&organizationalEntity=urn:li:organization:{{ config.org_id }}`, returning
  `{"elements":[...]}` paged with LinkedIn's `start`/`count` offset convention
  (`pagination.type: offset_limit`, `offset_param: start`, `limit_param: count`, `page_size: 100`,
  legacy's own default). Legacy passes each element through largely verbatim (only a fixed known-key
  allowlist, no renaming), so plain schema projection (camelCase properties declared as-is) covers
  every field; only `org_id` needs a `computed_fields` stamp, since the raw element does not
  self-describe which organization it belongs to.
- **`total_follower_count`**: a single-object GET at `/networkSizes/urn:li:organization:{{
  config.org_id }}?edgeType=COMPANY_FOLLOWED_BY_MEMBER` (`records.path: "."`, `pagination: none`).
  `first_degree_size` is a bare single-reference `computed_fields` entry
  (`{{ record.firstDegreeSize }}`), so the engine's typed-extraction rule (conventions.md §3)
  copies the raw JSON integer verbatim rather than stringifying it — the schema declares
  `"integer"`, matching legacy's own untouched-int passthrough.

Every stream stamps `org_id` (`config.org_id`) onto every emitted record via `computed_fields`,
matching legacy's `stamp` wrapper in `Read` (`rec["org_id"] = orgID`) exactly — statistics elements
do not always echo the entity urn themselves, so this is the only way a downstream destination can
tell which organization a given follower/share-statistics record belongs to.

`check` issues a single bounded `GET /organizations/{{ config.org_id }}`, mirroring legacy's own
`Check` (a bounded read of the organization profile confirms auth, org id, and connectivity without
mutating anything).

None of the 4 streams declare an `incremental` block: these are lifetime/point-in-time statistics
with no cursor field, matching legacy's `InitialState` (always an empty cursor) and its own
`linkedinPagesStreams()` catalog (no `CursorFields` declared anywhere) exactly.

## Write actions & risks

None. LinkedIn Pages is read-only (`capabilities.write: false`); legacy's own `Write` always
returns `connectors.ErrUnsupportedOperation` — there is no approved reverse-ETL write surface for
this data set.

## Known limits

- Only the 4 legacy-parity read streams are implemented; see `api_surface.json`. LinkedIn's much
  larger Community Management/Marketing surface (posts, comments, reactions, ad campaigns, lead gen
  forms) is out of scope until Pass B; LinkedIn Ads has its own sibling connector.
- **`page_size`/`max_pages` config overrides are not modeled.** Legacy accepts optional `page_size`
  (1-1000, default 100) and `max_pages` (default unlimited) config keys read at request time
  (`linkedinPageSize`/`linkedinMaxPages`, `linkedin_pages.go:433-461`). The engine's
  `PaginationSpec.PageSize`/`MaxPages` fields are plain fixed JSON integers baked into
  `streams.json`'s `base.pagination` block — there is no templating/config-driven override
  mechanism for either. This bundle declares a fixed `page_size: 100` (legacy's own default) and no
  `max_pages` cap (unbounded, matching legacy's own default). Neither `page_size` nor `max_pages` is
  declared in `spec.json` (a declared-but-unwireable key is worse than an absent one, conventions.md
  F6). The required 2-page conformance fixture
  (`fixtures/streams/follower_statistics/{page_1,page_2}.json`) is sized to match live behavior:
  page 1 returns a full 100-element page (so the paginator continues to page 2) and page 2 returns
  the remainder — a fixture-convenience page size is never leaked into the live pagination config.
- `linkedin_version` is a plain `config` override (default `202601`) rather than a hardcoded
  constant, so an operator can pin a specific LinkedIn API version if needed; legacy's own default
  behaves identically when unset.
- `base_url` supports the same test/proxy override legacy's `linkedinBaseURL` does (any absolute
  `http`/`https` URL with a host); this is enforced structurally by the engine's SSRF-guarded path
  interpolation and requester construction, not by a bundle-specific check.
