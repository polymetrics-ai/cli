# Overview

Front is a wave2 fan-out declarative-HTTP migration. It reads Front contacts, conversations,
inboxes, tags, teammates, and channels through the Front Core REST API
(`GET https://api2.frontapp.com/<resource>`). This bundle is capability-parity migrated from
`internal/connectors/front` (the hand-written connector it migrates); the legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Front API token via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`) and is never logged, matching legacy's
`connsdk.Bearer(secret)` (`front.go:250`). `base_url` defaults to `https://api2.frontapp.com`
(legacy's `frontDefaultBaseURL`) and may be overridden for tests/proxies.

## Streams notes

All 6 streams (`contacts`, `conversations`, `inboxes`, `tags`, `teammates`, `channels`) share the
same shape: `GET` against a Front list endpoint, records at `_results`, primary key `["id"]`.
`contacts` and `conversations` additionally declare `x-cursor-field` (`updated_at` /
`last_message_at` respectively), matching legacy's own `CursorFields` — the other 4 streams declare
none, matching legacy exactly (legacy's own `frontStreams()` leaves `PrimaryKey`-only Stream
structs for `inboxes`/`tags`/`teammates`/`channels`).

Pagination follows Front's own body-cursor convention (`pagination.type: next_url`,
`next_url_path: "_pagination.next"`): the first page is requested at `<resource>?limit=<page_limit>`
and every subsequent page follows the absolute `_pagination.next` URL Front returns, matching
legacy's `harvest` function (`front.go:148-187`) exactly, including its same-host SSRF-guard
default (THREAT-MODEL §3; Front's own `next` URL is always same-origin in production, so
`allow_cross_host` is left at its default `false`). `page_limit` (default `50`, legacy's
`frontDefaultPageSize`) is sent as `limit` on the first request via a `{{ config.page_limit }}`
template; like bitly's identical divergence (documented in `docs/migration/conventions.md`'s
bitly worked example), the engine's `readDeclarative` re-merges `stream.Query` onto every
subsequent page request, including ones that follow an absolute next-page URL — unlike legacy,
which resets to an empty `url.Values{}` once it starts following `_pagination.next`
(`front.go:183-184`, `query = nil`). This is a wire-request-shape divergence verified benign in
DATA terms only: Front's own `next` URL already carries the same `limit` value the engine
re-applies (the replace is idempotent), so the effective query on every page is value-identical to
what Front's own `next` URL already carries.

## Write actions & risks

None. Front is read-only (`capabilities.write: false`, no `writes.json`), matching legacy's
`Write` returning `connectors.ErrUnsupportedOperation` (legacy's own package doc: "the API key only
grants list access... so no reverse-ETL write surface is exposed").

## Known limits

- **Single-page fixtures for every stream (sanctioned `next_url` exception, conventions.md §4).**
  Front's `_pagination.next` URL is the conformance replay server's own runtime address —
  unknown ahead of time — so a static fixture file cannot embed a correct second-page URL. Every
  stream therefore ships exactly one fixture page (`_pagination.next: null`, terminating
  immediately); `conformance`'s `pagination_terminates` dynamic check runs against whichever stream
  it selects first, which correctly proves single-page termination. Real 2-page `next_url`
  correctness for this connector is not proven by a live `paritytest/front` suite in this wave (no
  Go files were authored per this wave's Tier-1-only mandate) — this is a scope narrowing relative
  to the sanctioned bitly/calendly pattern (which does ship a live parity test), not a claim that
  multi-page correctness has been verified beyond the engine's own `next_url` paginator unit tests.
- **`max_pages` is not modeled.** Legacy's hard request-count cap override (`frontMaxPages`,
  `front.go:302-315`) has no equivalent knob on the `next_url` paginator; pagination is bounded only
  by an empty/null `_pagination.next` value, matching Front's own real termination signal.
- **`page_size` alias is not modeled.** Legacy accepts `page_size` as an alias for `page_limit`
  (`frontPageSize`, `front.go:283-300`). Only `page_limit` (the catalog's canonical config field
  name) is wired into this bundle's `spec.json`/`streams.json`; the alias is a documented,
  narrow scope reduction (a caller using the alias name would need to switch to `page_limit`).
