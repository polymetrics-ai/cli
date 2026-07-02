# Overview

Opinion Stage is a Tier-1 declarative-HTTP migration of the legacy hand-written connector at
`internal/connectors/opinion-stage/` (read-only ground truth, unchanged until wave6's registry
flip). It reads items (polls, quizzes, and forms) through the Opinion Stage Public Result API
(`GET {base_url}/api/v2/items`), a JSON:API-shaped endpoint (`{id, type, attributes, links,
relationships}` records under a top-level `data` array). This bundle covers the `items` stream
only — see **Known limits** below for why legacy's `responses`/`questions` per-item substreams are
out of scope for this Tier-1 pass.

## Auth setup

The Opinion Stage Public Result API authenticates with HTTP Basic auth using your personal API key
as the username and an empty password — this is the API's own documented convention, not a
polymetrics invention (see `docs/connectors/source-opinion-stage/MANUAL.md`'s `authMethods`).
`spec.json` declares a single required secret, `api_key`; `streams.json`'s `base.auth` wires it as
`{"mode": "basic", "username": "{{ secrets.api_key }}", "password": ""}`, matching legacy's
`connsdk.Basic(strings.TrimSpace(secret), "")` exactly (`opinion_stage.go:251`). Retrieve your key
from your account page at `https://app.opinionstage.com/dashboard/settings/account/edit`.

`base_url` defaults to `https://api.opinionstage.com` (materialized via `spec.json`'s `"default"`,
matching legacy's `opinionStageDefaultBaseURL` in-code fallback) and may be overridden for tests or
proxies.

## Streams notes

`items` (`GET /api/v2/items`) is the only stream in this bundle. Records are extracted from the
top-level `data` array (`records.path: "data"`); each JSON:API record's `id`, `type`, `links`, and
`relationships` fields survive schema projection unchanged (they are already top-level keys on the
raw record — no rename needed). `attributes.title`, `attributes.status`, and `attributes.embed` are
lifted to top-level output fields via `computed_fields` (`{{ record.attributes.title }}`, etc.),
and `attributes.timestamps.created`/`attributes.timestamps.modified` are renamed to top-level
`created`/`modified`, matching legacy's `jsonAPIRecord` flattening helper
(`streams.go:139-158`) field-for-field.

Pagination is JSON:API-style page-number pagination: `page[number]` (1-based) and `page[size]`,
matching legacy's `harvest` loop (`opinion_stage.go:131-161`) exactly — a page returning fewer
records than `page[size]` stops the read (`len(records) < pageSize`); the base pagination block's
`page_size: 2` is the bundle's own fixed per-page-request size (legacy's own runtime default is 50,
config-overridable up to 1000, but the declarative dialect's `PaginationSpec.PageSize` is a fixed
value baked into the bundle — see Known limits). There is no cursor/incremental field: legacy is
full-refresh only (`docs/connectors/source-opinion-stage/MANUAL.md`: `supports incremental: false`),
so `schemas/items.json` declares no `x-cursor-field`.

## Write actions & risks

None. Opinion Stage's Public Result API is read-only; `capabilities.write` is `false` and this
bundle ships no `writes.json`, matching legacy's `Write` stub
(`connectors.ErrUnsupportedOperation`, `opinion_stage.go:312-314`).

## Known limits

- **`responses` and `questions` per-item substreams are not modeled.** Legacy additionally exposes
  `GET /api/v2/items/{id}/responses` and `GET /api/v2/items/{id}/questions`, each requiring a
  list-then-fan-out read: first page through `items` to collect every item id, then issue one
  request per id per substream (`readSubstream`/`itemIDs`, `opinion_stage.go:163-197`). The
  declarative dialect's `StreamSpec` has exactly one fixed (templated) `path` per stream and no
  parent-id-driven fan-out primitive — this is precisely the `StreamHook` (sub-resource fan-out)
  use case named in `docs/migration/conventions.md` §1's Tier-2 hook table ("issue → comments per
  issue"). A `StreamHook` would close this gap but is Go and out of scope for this Tier-1-only
  migration pass; `api_surface.json` documents both endpoints as `excluded: {category:
  "out_of_scope"}` rather than silently dropping them or faking a declarative approximation. A
  future capability-expansion pass may promote this bundle to Tier 2 (bundle + hook) specifically to
  add these two substreams.
- **`page_size`/`max_pages` config knobs are not carried over.** Legacy accepts `page_size`
  (1-1000, default 50) and `max_pages` (0/`all`/`unlimited` = unbounded) as runtime config
  overrides. The declarative `PaginationSpec.PageSize`/`MaxPages` fields are fixed values baked
  into `streams.json`'s `base.pagination` block, not config-templated — there is no mechanism in
  the dialect to route a `spec.json` property into either field (unlike `stream.Query`'s opt-in
  templating). Declaring `page_size`/`max_pages` in `spec.json` anyway would be genuinely dead
  config no template in this bundle ever consumes (F6, `docs/migration/conventions.md`), so
  neither is declared. `page_size` in this bundle is fixed at 2 (the client-side short-page stop
  threshold, analogous to searxng's identical fixed-`page_size` shape); `max_pages` is unbounded
  (absent, matching legacy's own `all`/`unlimited` default).
