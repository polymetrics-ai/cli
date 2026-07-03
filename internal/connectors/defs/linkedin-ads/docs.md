# LinkedIn Ads

## Overview

Reads LinkedIn Ads accounts, campaign groups, campaigns, and creatives through the LinkedIn
Marketing REST API (`https://api.linkedin.com/rest`). Read-only: there is no approved reverse-ETL
write surface for LinkedIn Ads in pm, matching the legacy `internal/connectors/linkedin-ads`
package.

## Auth setup

A LinkedIn member OAuth2 **access token** (`secrets.access_token`) is sent as a Bearer token on
every request. The refresh_token exchange (client_id/client_secret/refresh_token) is performed
outside this connector, exactly as legacy documents — only the resolved long-lived access token is
consumed here. Every request also carries two static headers the LinkedIn Marketing API requires:
`LinkedIn-Version` (monthly `YYYYMM` version string, default `202601`, overridable via
`config.linkedin_version`) and `X-Restli-Protocol-Version: 2.0.0`.

## Streams notes

All four streams (`accounts`, `campaign_groups`, `campaigns`, `creatives`) read from LinkedIn's
`{resource}?q=search` finder endpoints and share identical `start`/`count` offset pagination (a
short page — fewer records than the declared page size — stops the read). Every entity exposes a
numeric `id` except `creatives`, whose id is a string URN
(`urn:li:sponsoredCreative:...`), matching legacy's per-stream primary-key typing.

`accounts`, `campaign_groups`, and `campaigns` derive their `created_at`/`last_modified` cursor
fields from LinkedIn's nested `changeAuditStamps.created.time` /
`changeAuditStamps.lastModified.time` (epoch millis) via `computed_fields`, exactly mirroring
legacy's `changeAuditStamps` helper. `creatives` instead reads the flat `createdAt`/`lastModifiedAt`
epoch-millis fields the creatives endpoint exposes directly (legacy's own comment documents this
newer creatives endpoint shape as the primary source, falling back to `changeAuditStamps` only when
both flat fields are absent — a fallback that never triggers against the real creatives API and is
not modeled here; see Known limits).

## Write actions & risks

None. LinkedIn Ads is read-only in pm (`capabilities.write: false`), matching legacy.

## Known limits

- **`page_size`/`max_pages` config overrides are not modeled.** `streams.json`'s
  `base.pagination.page_size` is set to legacy's real production default, `100` (legacy:
  `linkedinDefaultPageSize = 100`) — this is the actual value a live deployment's paginator sends;
  it is not a fixture convenience. `config.page_size` from legacy's spec is not wired into the
  engine's pagination block (the same "config value looks live but is dead" shape documented for
  stripe's `page_size`/`limit_param`, conventions.md §5 item 3) since the dialect's pagination spec
  is a fixed, bundle-declared value, not a per-request override; the value is still declared in
  `spec.json` for config-surface documentation but is otherwise inert. The mandatory 2-page
  conformance fixture (`fixtures/streams/accounts/{page_1,page_2}.json`) is sized to match: page 1
  returns a full 100-record page (so the paginator continues to page 2), page 2 returns the
  1-record remainder — matching aviationstack's and awin-advertiser's identical repaired precedent
  (`docs/migration/conventions.md`, wave2 sweep class C3). Every other stream's single-page fixture
  requests `count=100` to match.
- **`creatives`' `changeAuditStamps` fallback is not modeled**: legacy falls back to
  `changeAuditStamps.created.time`/`changeAuditStamps.lastModified.time` only when a creatives
  record has neither `createdAt` nor `lastModifiedAt` set. The real LinkedIn creatives API always
  populates the flat fields, so this fallback path is legacy dead code in practice; only the flat-field
  mapping is implemented here. If a future creatives API response omits both flat fields, this
  bundle emits `null` for `created_at`/`last_modified` rather than falling back — a documented,
  narrow scope narrowing, not a silent data change for any input the real API is known to send.
- **`secrets.access_token` config-surface renaming**: legacy accepts the secret under the dotted key
  `credentials.access_token` (with a bare `access_token` fallback). The engine's `secrets.<key>`
  template reference only resolves a single flat segment (`interpolate.go`'s `resolveRefValue`
  splits on `.` and uses only the first segment as the key), so a dotted spec property name cannot
  be referenced by `{{ secrets.credentials.access_token }}`. This bundle declares the secret as
  `spec.json`'s flat `access_token` property instead — the resolved token value and its Bearer-auth
  usage are byte-identical; only the config-surface field name differs from legacy's dotted key.
- Pass B (full LinkedIn Marketing API surface — analytics/reporting finders, audiences, conversions,
  lead gen forms, video ads, write actions) is out of scope; see `api_surface.json`.
