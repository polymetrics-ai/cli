# Overview

SurveySparrow is a wave2 fan-out declarative-HTTP migration. It reads SurveySparrow surveys,
contacts, responses, and questions through the SurveySparrow API
(`GET https://api.surveysparrow.com/v3/...`). This bundle targets capability parity with
`internal/connectors/survey-sparrow` (the hand-written `surveysparrow` package it migrates); the
legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a SurveySparrow access token via the `access_token` secret; it is sent as a Bearer token
(`Authorization: Bearer <access_token>`) and is never logged, matching legacy's
`connsdk.Bearer(token)` (`survey_sparrow.go:119`). `base_url` defaults to
`https://api.surveysparrow.com/v3` and may be overridden for tests/proxies.

## Streams notes

`surveys`, `contacts`, and `responses` are top-level `GET` list endpoints (`/surveys`, `/contacts`,
`/responses`); records live at the `data` key on all 4 streams. `questions` is scoped to one
SurveySparrow survey via the required `survey_id` config value, substituted into the
`/surveys/{{ config.survey_id }}/questions` path template (urlencoded by `InterpolatePath`'s
per-segment default, matching legacy's own `url.PathEscape(surveyID)` in `resolveResource`); an
absent `survey_id` hard-errors on both sides (legacy: `"survey-sparrow stream requires config
survey_id for path %q"`; engine: an unresolved `config.survey_id` path-template key).

`responses` declares `incremental.cursor_field: completed_time`, matching legacy's own
`CursorFields: []string{"completed_time"}` declaration; neither this bundle nor legacy sends a
server-side lower-bound filter or performs client-side filtering for this stream (legacy's `Read`
performs no incremental filtering at all) — this bundle matches that exactly (no `request_param`/
`client_filtered` declared). `surveys`, `contacts`, and `questions` have no cursor field, matching
legacy (full refresh only). `id` is a bare integer on every stream (SurveySparrow's real wire
shape), matching legacy's `{Name: "id", Type: "integer"}` field declarations exactly.

Pagination is page-number (`pagination.type: page_number`, `page_param: page`, `size_param: limit`,
`start_page: 1`, `page_size: 100`), identical to legacy's
`connsdk.PageNumberPaginator{PageParam: "page", SizeParam: "limit", StartPage: 1, PageSize:
pageSize}` with legacy's default `pageSize` of 100.

## Write actions & risks

None. Legacy's `Write` always returns `connectors.ErrUnsupportedOperation`; `capabilities.write` is
`false` and this bundle ships no `writes.json`.

## Known limits

- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes both as config-driven
  overrides (`survey_sparrow.go`'s `pageSize`/`maxPages`, bounded 1-500 / a non-negative integer or
  `all`/`unlimited`). The engine's `page_number` paginator has no config-driven page-size or
  max-pages knob (`PaginationSpec.PageSize`/`MaxPages` are static bundle JSON, never templated), so
  this bundle uses legacy's own default (`page_size: 100`) as a fixed bundle value and does not
  declare `page_size`/`max_pages` in `spec.json` at all (a declared-but-unwireable config key is
  worse than an absent one, per `docs/migration/conventions.md` F6). Pagination is unbounded by
  default (reads every page until a short page), matching legacy's own default of `maxPages=0`
  (unbounded) when `max_pages` is unset.
- Full SurveySparrow API surface (survey/contact mutation, webhooks, NPS/reports) is out of scope
  for this wave; see `api_surface.json`'s `excluded: {category: out_of_scope}` entries.
