# Overview

Opinion Stage is a declarative-HTTP migration of the legacy hand-written connector at
`internal/connectors/opinion-stage/` (read-only ground truth, unchanged until wave6's registry
flip). It reads items (polls, quizzes, and forms) through the Opinion Stage Public Result API
(`GET {base_url}/api/v2/items`), a JSON:API-shaped endpoint (`{id, type, attributes, links,
relationships}` records under a top-level `data` array). It also reads legacy's per-item
`responses` and `questions` substreams by listing item ids and using `stream.fan_out` to request each
child endpoint.

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

`items` (`GET /api/v2/items`) extracts records from the
top-level `data` array (`records.path: "data"`); each JSON:API record's `id`, `type`, `links`, and
`relationships` fields survive schema projection unchanged (they are already top-level keys on the
raw record — no rename needed). `attributes.title`, `attributes.status`, and `attributes.embed` are
lifted to top-level output fields via `computed_fields` (`{{ record.attributes.title }}`, etc.),
and `attributes.timestamps.created`/`attributes.timestamps.modified` are renamed to top-level
`created`/`modified`, matching legacy's `jsonAPIRecord` flattening helper
(`streams.go:139-158`) field-for-field.

`responses` (`GET /api/v2/items/{id}/responses`) and `questions`
(`GET /api/v2/items/{id}/questions`) use `fan_out.ids_from.request` against `/api/v2/items` to
collect every item id, then substitute the id into the child path with `{{ fanout.id }}` and stamp
that value to `item_id`. `responses` lifts `attributes.answers`, `attributes.result`,
`attributes.result.title`, `attributes.result.text`, `attributes.timestamps.created`,
`attributes.timings.duration`, and `attributes.utm`, preserving top-level `id`, `type`, and `links`.
`questions` lifts `attributes.title`, `attributes.kind`, `attributes.lead`,
`attributes.timestamps.created`, and `attributes.timestamps.modified`, preserving top-level `id` and
`type`. Those field sets match legacy's `opinionStageResponseRecord` and
`opinionStageQuestionRecord` mappers.

Pagination is JSON:API-style page-number pagination: `page[number]` (1-based) and `page[size]`,
matching legacy's `harvest` loop (`opinion_stage.go:131-161`) exactly — a page returning fewer
records than `page[size]` stops the read (`len(records) < pageSize`); the base pagination block's
`page_size: 50` matches legacy's own runtime default (`opinionStageDefaultPageSize`,
config-overridable up to 1000), fixed as a bundle-baked value rather than config-templated (see
Known limits). There is no cursor/incremental field: legacy is full-refresh only
(`docs/connectors/source-opinion-stage/MANUAL.md`: `supports incremental: false`), so
`schemas/items.json` declares no `x-cursor-field`.

## Write actions & risks

None. Opinion Stage's Public Result API is read-only; `capabilities.write` is `false` and this
bundle ships no `writes.json`, matching legacy's `Write` stub
(`connectors.ErrUnsupportedOperation`, `opinion_stage.go:312-314`).

## Known limits

- **`page_size`/`max_pages` config knobs are not carried over.** Legacy accepts `page_size`
  (1-1000, default 50) and `max_pages` (0/`all`/`unlimited` = unbounded) as runtime config
  overrides. The declarative `PaginationSpec.PageSize`/`MaxPages` fields are fixed values baked
  into `streams.json`'s `base.pagination` block, not config-templated — there is no mechanism in
  the dialect to route a `spec.json` property into either field (unlike `stream.Query`'s opt-in
  templating). Declaring `page_size`/`max_pages` in `spec.json` anyway would be genuinely dead
  config no template in this bundle ever consumes (F6, `docs/migration/conventions.md`), so
  neither is declared. `page_size` in this bundle is fixed at 50, matching legacy's own default
  (the client-side short-page stop threshold); `max_pages` is unbounded (absent, matching legacy's
  own `all`/`unlimited` default).
