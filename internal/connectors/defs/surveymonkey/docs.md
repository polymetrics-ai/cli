# Overview

SurveyMonkey is a Tier-1 declarative-HTTP migration. It reads SurveyMonkey surveys, collectors,
and bulk survey responses through the SurveyMonkey v3 REST API
(`GET https://api.surveymonkey.com/v3/...`). This bundle targets capability parity with
`internal/connectors/surveymonkey` (the hand-written connector it migrates); the legacy package
stays registered and unchanged until wave6's registry flip. SurveyMonkey is read-only both in
legacy and here (legacy's `Write` always returns `connectors.ErrUnsupportedOperation`).

**Tier justification**: legacy is a pure `connsdk`-based HTTP connector — a `connsdk.Requester`
with `connsdk.Bearer` auth, a hand-rolled `harvest` loop that follows a `links.next` URL, and plain
per-record field copies (`surveyMonkeySurveyRecord`/`surveyMonkeyCollectorRecord`/
`surveyMonkeyResponseRecord`). No signature auth, no async job polling, no multipart/XML bodies,
no sub-resource fan-out beyond a single required path substitution (`survey_id`) — nothing that
needs a Go hook. This is a clean Tier-1 declarative bundle.

## Auth setup

Provide a SurveyMonkey OAuth access token via the `access_token` secret; it is sent as a Bearer
token (`Authorization: Bearer <access_token>`) and is never logged, matching legacy's
`connsdk.Bearer(token)` (`surveymonkey.go:158`). `base_url` defaults to
`https://api.surveymonkey.com/v3` (legacy's `surveyMonkeyDefaultBaseURL`) and may be overridden
for tests/proxies.

## Streams notes

`surveys` is a simple, non-scoped list endpoint (`GET /surveys`); records live under the `data`
key, matching `connsdk.RecordsAt(resp.Body, "data")` (`surveymonkey.go:114`).

`collectors` and `survey_responses` are both scoped to one survey: their path templates
(`/surveys/{{ config.survey_id }}/collectors`, `/surveys/{{ config.survey_id }}/responses/bulk`)
substitute the required `survey_id` config value (urlencoded by `InterpolatePath`'s per-segment
default, matching legacy's own `"surveys/" + url.PathEscape(surveyID) + "/" + endpoint.resource`);
an absent `survey_id` hard-errors on both sides (legacy: `"surveymonkey connector requires config
survey_id for this stream"`; engine: an unresolved `config.survey_id` path-template key — same
failure classification, different literal text, per conventions.md §5's precedent for
config-validation-error parity). `survey_responses`' resource is `responses/bulk` (SurveyMonkey's
own bulk-response endpoint, `surveyMonkeyStreamEndpoints["survey_responses"].resource`), preserved
verbatim.

Pagination follows SurveyMonkey's own `links.next` **relative-or-absolute** URL convention
(`pagination.type: next_url`, `next_url_path: "links.next"`), matching legacy's own `harvest` loop
exactly: `per_page=100` (legacy's `surveyMonkeyDefaultPageSize`) is sent as a static per-stream
query value on the FIRST request only; legacy explicitly resets `query = nil` once it follows a
`links.next` URL (`surveymonkey.go:130-131`), stopping when `links.next` is empty
(`strings.TrimSpace(next) == ""`). The engine's `next_url` paginator re-merges `stream.Query` onto
every subsequent request too (`engine/read.go`'s `mergeQuery`) rather than legacy's explicit
one-time-only reset — this is the identical wire-request-shape divergence documented for bitly's
`bitlinks` stream (conventions.md's bitly worked example), verified benign in DATA terms only
because SurveyMonkey's own `links.next` URL already carries the equivalent `per_page` value the
engine re-applies (the replace is idempotent); if SurveyMonkey's `next` URL ever omitted or
diverged from `per_page`, this bundle's request would differ from legacy's — it does not today.
The engine's `next_url` paginator's same-host SSRF guard (THREAT-MODEL §3) passes cleanly since
SurveyMonkey's `links.next` URL is always same-origin as `base_url` in production.

None of the 3 streams declare an `incremental` block: none of SurveyMonkey's list endpoints expose
a server-side incremental cursor in legacy's own catalog (no `CursorFields` declared anywhere in
`surveyMonkeyStreams()`) — full refresh only, matching legacy exactly.

## Write actions & risks

None. SurveyMonkey is read-only both in legacy and here: legacy's own `Write` method returns
`connectors.ErrUnsupportedOperation` unconditionally. `capabilities.write` is `false` and this
bundle ships no `writes.json`.

## Known limits

- **`next_url` pagination fixtures are single-page** (conventions.md §4's sanctioned exception):
  the second-page URL is the replay server's own runtime address, unknown until the harness picks
  a port, so a static fixture file cannot embed it. Each stream's fixture terminates cleanly after
  page 1 (`links.next: ""`), satisfying `fixtures_present`/`read_fixture_nonempty`/
  `pagination_terminates`; live 2-page `next_url` correctness for this exact shape is already
  proven by bitly's own `TestParityBitly_BitlinksStreamPaginates`-style live parity test pattern —
  a dedicated `paritytest/surveymonkey` live-server test asserting a second `links.next` page is
  actually followed is recommended future work, matching bitly's precedent, but is not required to
  reach `migrated` status (SurveyMonkey ships no dedicated parity suite in legacy either).
- **`page_size` is not runtime-configurable.** Legacy exposes `page_size` as a config-driven
  override (`surveyMonkeyPageSize`, `surveymonkey.go:210-212`, bounded 1-1000). The engine's
  `next_url` paginator has no analogous config-driven page-size knob (it never reads
  `PaginationSpec.PageSize`, unlike `page_number`/`offset_limit`), and `stream.Query` templating
  has no absent-key-falsy tolerance for a required override (conventions.md §3), so a
  `{{ config.page_size }}` template would hard-error whenever `page_size` is unset — the common
  case. This bundle therefore sends SurveyMonkey's own default (`per_page=100`) as a static
  per-stream query literal, matching bitly's `size=50` precedent, and does not declare `page_size`
  in `spec.json` at all (F6, REVIEW.md: a declared-but-unwireable config key is worse than an
  absent one).
- **`max_pages` is not modeled.** Legacy exposes `max_pages` as a hard request-count cap override
  (`surveyMonkeyMaxPages`, `surveymonkey.go:214-216`). The engine's `next_url` paginator has no
  `MaxPages`-equivalent knob wired to a config value; pagination is bounded only by the
  short/empty-page stop signal (an empty `links.next` value), matching SurveyMonkey's own real
  termination behavior for any config that never sets `max_pages` (the common case) and differing
  only in the rare case an operator relied on the hard cap to stop a sync early.
