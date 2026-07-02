# Overview

Teamtailor is a wave2 fan-out declarative-HTTP migration of `internal/connectors/teamtailor` (the
hand-written legacy connector this bundle migrates; the legacy package stays registered and
unchanged until wave6's registry flip). It reads Teamtailor jobs through the Teamtailor JSON:API
REST API (`GET https://api.teamtailor.com/v1/jobs`). Read-only.

## Auth setup

Legacy accepts either of two credential shapes with a defined precedence: a plain config value
`api` (read first, `teamtailor.go:117-122`'s `apiKey()` helper) OR a secret `api_key` (fallback
when `api` is unset). Both are marked `x-secret: true` here (the marker describes a field's
credential nature, not which RuntimeConfig namespace legacy happened to read it from) and are
wired as TWO auth candidates in `streams.json` `base.auth`, in the SAME precedence order legacy
uses — `api` first, `api_key` second — since `selectAuth` is first-match-wins and reordering would
silently change which credential wins when both are configured. Either credential is sent as
`Authorization: Token token=<key>` (`api_key_header` mode, `prefix: "Token token="`), matching
legacy's `connsdk.APIKeyHeader("Authorization", key, "Token token=")`. An optional `x_api_version`
config value is sent as the `X-Api-Version` header; when unset, the header is omitted entirely
(not sent empty) per the engine's conditional-header omission rule for an optional, non-required
`config.*` key. Never logged.

`base_url` defaults to `https://api.teamtailor.com/v1` and may be overridden for tests/proxies.

## Streams notes

`jobs` is the only stream: `GET /jobs`, records at `data` (Teamtailor's JSON:API envelope), primary
key `["id"]`. `id` is a top-level JSON:API field and survives schema projection directly; `title`
and `created_at` live nested under `attributes` (`attributes.title`, `attributes.created-at` —
note the real wire hyphen, not underscore) and are NOT present at the raw record's top level, so
plain schema projection alone would silently drop them. `computed_fields` rehydrates both
(`"title": "{{ record.attributes.title }}"`, `"created_at": "{{ record.attributes.created-at }}"`)
from the raw pre-projection record, matching legacy's `object(item["attributes"])`-then-field-pick
shape exactly (`teamtailor.go:86-87`).

Pagination follows a 1-based page-number convention (`pagination.type: page_number`,
`page_param: page[number]`, `size_param: page[size]`, `page_size: 100`) — matches legacy's
hand-rolled loop (`teamtailor.go:72-94`: `page[number]`/`page[size]` query params, stopping when a
page returns fewer records than the configured size).

## Write actions & risks

None. Legacy `teamtailor` is read-only (`Write` returns `connectors.ErrUnsupportedOperation`);
`metadata.json` declares `capabilities.write: false` and this bundle ships no `writes.json`.

## Known limits

- Full Teamtailor JSON:API surface (candidates, applications, departments, roles, stages) is out
  of scope for wave2; see `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B
  capability expansion"}` entries. Only the single legacy-parity `jobs` stream is implemented.
- **`page_size` is not runtime-configurable.** Legacy exposes a config-driven `page_size` override
  (`teamtailor.go:143-149`, `pageSize(cfg, 100)`, any positive integer, defaulting to 100 when
  unset or invalid). The engine's `page_number` paginator constructor reads `PaginationSpec.PageSize`
  as a static bundle-level integer from `streams.json`, not a config-templated field, so there is
  no mechanism to make it runtime-configurable from `config.page_size` without inventing Go. This
  bundle hardcodes `page_size: 100`, legacy's own default, matching every input that does not
  explicitly override the page size (the common case); an operator who previously set a
  smaller/larger `page_size` config value loses that override here. `page_size` is not declared in
  `spec.json` at all (F6, REVIEW.md: a declared-but-unwireable key is worse than an absent one).
- All fixtures (`fixtures/streams/jobs/**`, `fixtures/check.json`) represent Teamtailor's real
  JSON:API wire shape, including the `attributes.created-at` hyphenated field name before the
  `computed_fields` rename. `fixtures/streams/jobs/page_1.json` carries exactly 100 synthetic
  records (matching the static `page_size: 100`) so the real page-number paginator's
  short-page-stop rule genuinely exercises a second request, proving `pagination_terminates`
  rather than short-circuiting after page 1.
- Dual-auth precedence (`api` over `api_key`) is exercised at authoring time per the ordering rule
  above; a both-present parity assertion is left to a future parity-suite extension (no
  `paritytest/teamtailor` package exists yet in this wave — Tier 1 fan-out ships bundle files only,
  per the migration's hard rules).
