# Overview

Teamwork is a wave2 fan-out declarative-HTTP migration of `internal/connectors/teamwork` (the
hand-written legacy connector this bundle migrates; the legacy package stays registered and
unchanged until wave6's registry flip). It reads Teamwork projects through the Teamwork REST API
(`GET https://api.teamwork.com/projects.json`). Read-only.

## Auth setup

Provide a Teamwork username (email) via the `username` config value and an API token via the
`password` secret; both are required. They are sent as HTTP Basic auth
(`Authorization: Basic base64(username:password)`), matching legacy's
`connsdk.Basic(username, password)` (`teamwork.go:110`). `password` is never logged. `base_url`
defaults to `https://api.teamwork.com` and may be overridden for tests/proxies.

Legacy's own `Check` is a pure config-presence validation with no network call
(`teamwork.go:34-48`); this bundle's declarative `check` (`GET /projects.json`) performs a real,
richer network round-trip instead — an unavoidable, structural consequence of the engine's
declarative `check` dialect (every Tier-1 golden bundle in this repo does the same; there is no
"validate-only, no request" primitive), not a per-connector authoring choice.

## Streams notes

`projects` is the only stream: `GET /projects.json`, records at `projects`, primary key `["id"]`.
Teamwork's real wire shape names the creation-timestamp field `created-on` (hyphenated); a
`computed_fields` rename (`"created_at": "{{ record.created-on }}"`) reproduces legacy's effective
output field name exactly (`teamwork.go:86`). `id` and `name` need no rename — they already match
the raw API's field names and survive plain schema projection unaided.

Pagination follows a 1-based page-number convention (`pagination.type: page_number`,
`page_param: page`, `size_param: pageSize`, `page_size: 100`) — matches legacy's hand-rolled loop
(`teamwork.go:72-93`: `page`/`pageSize` query params, stopping when a page returns fewer records
than the configured size).

## Write actions & risks

None. Legacy `teamwork` is read-only (`Write` returns `connectors.ErrUnsupportedOperation`);
`metadata.json` declares `capabilities.write: false` and this bundle ships no `writes.json`.

## Known limits

- Full Teamwork API surface (tasks, task lists, time entries, milestones, people) is out of scope
  for wave2; see `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B
  capability expansion"}` entries. Only the single legacy-parity `projects` stream is implemented.
- **`page_size` is not runtime-configurable.** Legacy exposes a config-driven `page_size` override
  (`teamwork.go:124-130`, `pageSize(cfg, 100)`, any positive integer, defaulting to 100 when unset
  or invalid). The engine's `page_number` paginator constructor reads `PaginationSpec.PageSize` as
  a static bundle-level integer from `streams.json`, not a config-templated field, so there is no
  mechanism to make it runtime-configurable from `config.page_size` without inventing Go. This
  bundle hardcodes `page_size: 100`, legacy's own default, matching every input that does not
  explicitly override the page size (the common case); an operator who previously set a
  smaller/larger `page_size` config value loses that override here. `page_size` is not declared in
  `spec.json` at all (F6, REVIEW.md: a declared-but-unwireable key is worse than an absent one).
- All fixtures (`fixtures/streams/projects/**`, `fixtures/check.json`) represent Teamwork's real
  wire shape, including the hyphenated `created-on` field name before the `computed_fields` rename.
  `fixtures/streams/projects/page_1.json` carries exactly 100 synthetic records (matching the
  static `page_size: 100`) so the real page-number paginator's short-page-stop rule genuinely
  exercises a second request, proving `pagination_terminates` rather than short-circuiting after
  page 1.
