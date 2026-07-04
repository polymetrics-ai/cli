# Overview

Survicate is a Tier-1 declarative-HTTP connector for the Survicate **Data Export API v2**
(`https://data-api.survicate.com/v2`), Survicate's fixed, small, read-oriented API for exporting
survey, response, and respondent data (`https://developers.survicate.com/data-export/`). This is a
Pass B full-surface expansion: every list-shaped endpoint with a discovery path is now a stream
(`surveys`, `survey_questions`, `responses`); every other documented endpoint is either a
duplicate-shaped detail read already covered by a list stream, or genuinely un-syncable without an
externally-supplied id (respondent/personal-data endpoints — see `api_surface.json`).

**Tier justification**: a plain declarative-HTTP bundle. Auth is a single static header (no
signature/token-exchange scheme), pagination is the engine's `next_url` type, and the two
sub-resource streams are ordinary `fan_out` reads over the `surveys` list — nothing needs a Go
hook.

## Auth setup

Provide a Survicate Data Export API key (Survicate panel > Surveys Settings > Access Keys) via the
`api_key` secret. Survicate's own docs (`developers.survicate.com/data-export/setup`) state the
scheme precisely: *"The API key should be included in the `Authorization` header... The format for
this should be `Basic {{apiKey}}`"* — this is **not** standard HTTP Basic auth (base64 of a
`user:pass` pair); it is the literal API key sent verbatim after the `Basic ` prefix
(`Authorization: Basic <api_key>`). This bundle expresses that with
`auth.mode: api_key_header, header: Authorization, prefix: "Basic "` rather than
`mode: basic` (which would incorrectly base64-encode the key as though it were a username). The
previous version of this bundle used `mode: bearer` (`Authorization: Bearer <api_key>`), which does
not match Survicate's documented scheme at all — this is a Pass B correctness fix, not a behavior
change any real caller depended on (a `bearer`-shaped request would have been rejected by the real
API with 401 for every caller). `base_url` defaults to `https://data-api.survicate.com/v2` (the
current documented API version; the bundle's prior default pointed at a now-superseded `v1` path)
and may be overridden for tests/proxies.

## Streams notes

All three streams share the base's `next_url` pagination (`pagination_data.next_url` in the
response body — Survicate's documented `has_more`/`next_url` envelope, identical shape across
`surveys`, `survey_questions`, and `responses`). Survicate's `next_url` value is a **relative path**
(e.g. `/surveys?start=2023-01-01T00:00:00.000000Z`), not an absolute URL — the engine's `next_url`
paginator's same-origin SSRF guard (`checkOrigin`) fails closed on any URL with no parseable host,
which a relative path always has. `allow_cross_host: true` is set on the base pagination block to
bypass that guard; despite the name, this is a narrow, same-host-relative-path use of the flag (not
a genuine cross-host follow) — the only way to accommodate a documented relative-URL pagination
cursor with this paginator type. The request layer (`connsdk.Requester.resolveURL`) still resolves
the relative path against `config.base_url` correctly either way.

`surveys` (`GET /surveys`) lists every survey in the workspace; `items_per_page` is sent statically
at 100 (the documented max). The legacy connector's catalog publishes `updated_at` as the survey
cursor field and its `mapSurvey` mapper emits only `id`, `name`, `created_at`, and `updated_at`, so
the schema keeps exactly those fields plus `x-cursor-field` for legacy catalog/data parity. The v2
Data Export API docs and fixtures include additional fields, but widening the schema would emit
fields legacy always dropped. The API docs do not show a server-side incremental filter for
`updated_at`, so no `incremental` block is declared.

`survey_questions` (`GET /surveys/{survey_id}/questions`) and `responses`
(`GET /surveys/{survey_id}/responses`) both `fan_out` over every survey id (`ids_from.request`:
`GET /surveys`, `records_path: data`, `id_field: id`), stamping `survey_id` onto every emitted
record — there is no "list all questions/responses across every survey" endpoint; each is
inherently survey-scoped. `responses` does not declare a cursor field: the API's documented
`start`/`end` filters order responses **latest-to-oldest** (descending), the opposite of the
engine's ascending lower-bound-filter incremental model, so wiring a naive `request_param` against
`start` would silently invert which slice of history a "resume from cursor" sync actually returns —
narrower (full-refresh-only) is the correct, honest choice here.

`respondents/{respondent_uuid}/attributes` and `respondents/{respondent_uuid}/responses` are not
modeled as streams: neither has a "list all respondent ids" discovery endpoint, so there is no
config-free way to fan out over them (`api_surface.json`: `requires_elevated_scope`). The
`personal-data` GDPR endpoints (`GET`/`DELETE /personal-data`, both keyed by a single caller-supplied
email) are likewise not list-shaped resources and are excluded (`requires_elevated_scope`
/`destructive_admin`).

## Write actions & risks

None. The Data Export API is read-only apart from the GDPR `DELETE /personal-data` erasure
endpoint, which is excluded as `destructive_admin` (irreversibly deletes all data for an email
address across Survicate and connected services) rather than modeled as a reverse-ETL write.
`capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- **`next_url` pagination fixtures are single-page** (the sanctioned exception,
  `docs/migration/conventions.md` §4): Survicate's `next_url` is a relative path whose correct
  absolute resolution depends on the replay server's own runtime address, which a static fixture
  file cannot embed. `pagination_terminates` exercises the `surveys` stream (1 fixture page, 1
  request — proves fixture-consumption termination, not multi-page traversal). No
  `paritytest/survicate` package exists in this repo to prove live 2-page `next_url` correctness
  end-to-end; this is a documented gap, not a hidden one.
- **No respondent- or personal-data-scoped streams** — see Streams notes above; both require an
  externally-supplied id/email with no in-API discovery path.
- **`responses` is full-refresh-only** — no `x-cursor-field` or `incremental` block is declared
  because the API's `start`/`end` filters order latest-to-oldest, the opposite of the engine's
  ascending lower-bound model; every sync is full-refresh.
- Rate limits per Survicate's docs: 1000 requests/minute per workspace, 5 concurrent requests max;
  `metadata.json.rate_limit.requests_per_minute` documents this (informational-only per
  `docs/migration/conventions.md` §3 — the engine never enforces it; no `streams.json`
  `base.rate_limit` is declared since there is no evidence Survicate's own client behavior
  throttles proactively rather than reacting to a `429`).
