# Overview

SendGrid is a wave2 fan-out declarative-HTTP migration. It reads SendGrid Marketing Campaigns
lists, segments, and contacts, plus suppression bounces, through the SendGrid v3 REST API
(`GET https://api.sendgrid.com/v3/...`). This bundle targets capability parity with
`internal/connectors/sendgrid` (the hand-written connector it migrates); the legacy package stays
registered and unchanged until wave6's registry flip. SendGrid is read-only (`capabilities.write`
is `false`, matching legacy's `Write` returning `connectors.ErrUnsupportedOperation`).

## Auth setup

Provide a SendGrid API key via the `api_key` secret; it is sent as a Bearer token (`Authorization:
Bearer <api_key>`) and never logged, matching legacy's `connsdk.Bearer(secret)`
(`sendgrid.go:262-267`). `base_url` defaults to `https://api.sendgrid.com/v3` (legacy's
`sendgridDefaultBaseURL`) and may be overridden for tests/proxies.

## Streams notes

`lists` (`GET /marketing/lists`), `segments` (`GET /marketing/segments/2.0`), and `contacts` (`GET
/marketing/contacts`) all share SendGrid's marketing-API envelope,
`{<result_key>:[...],"_metadata":{"next":"<full url|absent>"}}` — records live at `result` (lists,
contacts) or `results` (segments), matching legacy's per-endpoint `recordsPath`
(`streams.go:34-37`). Pagination is `next_url` (`next_url_path: "_metadata.next"`) — SendGrid's own
`_metadata.next` value is a fully-qualified absolute URL that already carries the next page's
`page_token`/`page_size` query, matching legacy's `harvestMetadataNext`
(`sendgrid.go:141-180`), which follows it verbatim with no extra params. `page_size=100`
(SendGrid's own default, `sendgridDefaultPageSize`) is declared as a static per-stream `query`
value re-sent on every page request the engine issues (`streams.json`'s `"query": {"page_size":
"100"}`); see Known limits for the same re-sent-vs-legacy's-reset-to-nil divergence already
documented on bitly's/aircall's bundles (this wave's `next_url` siblings).

`suppression_bounces` (`GET /suppression/bounces`) returns a top-level JSON array rather than an
enveloped object (`records.path: "."`, matching legacy's `recordsPath: ""` /
`connsdk.RecordsAt(resp.Body, "")` root-array extraction), and uses `offset_limit` pagination
(`limit_param: limit`, `offset_param: offset`, `page_size: 100` — legacy's own `harvestOffset`,
`sendgrid.go:185-217`); a page shorter than 100 records signals the end, matching legacy's
`len(records) < pageSize` stop condition exactly — the engine's `OffsetPaginator.Next` uses the
identical `recordCount < PageSize` rule.

None of the four streams expose a request-time incremental filter parameter in legacy — legacy's
own `Stream.CursorFields` catalog metadata (`segments`/`contacts`: `updated_at`;
`suppression_bounces`: `created`) is purely descriptive cataloging, never wired into an actual
request-narrowing filter anywhere in `harvestMetadataNext`/`harvestOffset`. This bundle mirrors
that exactly: `x-cursor-field` is declared on each schema (matching legacy's catalog metadata) but
no stream declares an `incremental` block, so every read is full refresh, identical to legacy.

## Write actions & risks

None. SendGrid's marketing/suppression read endpoints have no obviously-safe reverse-ETL writes in
legacy (legacy's own package doc: "Read-only."); `capabilities.write` is `false` and this bundle
ships no `writes.json`, matching legacy's `Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **`next_url` fixtures for `lists`/`segments`/`contacts` are single-page, per the sanctioned
  exception (conventions.md §4).** A `next_url` stream's next-page URL is the replay server's own
  runtime address, unknown until the harness picks a port — a static fixture file cannot embed the
  correct absolute URL for a second page. `pagination_terminates` passes on the first stream
  (`lists`) with its single page (`hits == len(pages) == 1`), which is not a 2-page pagination proof
  but is not a false failure either. `suppression_bounces` (the one `offset_limit` stream in this
  bundle) ships the required 2-page fixture (`fixtures/streams/suppression_bounces/{page_1,
  page_2}.json`, 100 then 1 record) proving genuine multi-page consumption. Real 2-page
  `next_url`-following correctness for `lists`/`segments`/`contacts` is proven by legacy's own
  existing test (`internal/connectors/sendgrid/sendgrid_test.go`'s
  `TestReadListsPaginatesAndAuthenticates`, which drives a real 2-page `httptest.Server` and asserts
  the second page is requested via the served `_metadata.next` URL), plus the engine's own generic
  `next_url` paginator unit tests (`internal/connectors/engine/paginate_test.go`) and read-path
  integration test (`internal/connectors/engine/read_test.go`). This wave does not add a new
  `paritytest/sendgrid` package (out of scope per this wave's JSON-only mandate); a future wave
  adding hand-written parity suites should follow bitly's/calendly's
  `TestParity<Name>_..._TwoPagePagination` pattern for sendgrid specifically.
- **`page_size` is re-sent on every page request for the `next_url` streams, unlike legacy's
  reset-to-nil-then-follow.** The engine's `readDeclarative` loop merges `stream.Query` into every
  page request regardless of pagination type, and `connsdk.Requester.resolveURL` re-applies that
  merged query onto the absolute next-page URL (replacing any same-named param already present).
  Legacy instead passes `query = nil` once it follows an absolute next-page URL (`sendgrid.go:177`),
  sending `page_size` only on the first request. This mirrors bitly's/aircall's identical,
  already-ledgered divergence (conventions.md §5) and is benign in DATA terms only because
  SendGrid's own `_metadata.next` URL already carries the identical `page_size=100` value the engine
  re-applies (the replace is idempotent) — if SendGrid's `next` URL ever diverged from `page_size`,
  this bundle's request would differ from legacy's; today it does not.
- **`page_size`/`max_pages` config overrides are not modeled for the `next_url` streams.** Legacy
  exposes `page_size` (1-1000, default 100) and `max_pages` (0/all/unlimited or a positive integer
  cap) as config-driven overrides (`sendgridPageSize`/`sendgridMaxPages`, applied to ALL streams
  including `suppression_bounces`). The engine's `next_url` paginator has no config-driven
  page-size or request-count-cap knob at all (mirrors bitly's/aircall's identical limitation); these
  are therefore not declared in `spec.json`, and `lists`/`segments`/`contacts` send SendGrid's own
  default (`page_size=100`) as a static per-stream query literal. `suppression_bounces`'
  `offset_limit` paginator does support a config-driven page size in principle, but since the other
  three streams cannot honor one, `page_size` is left undeclared bundle-wide for consistency and to
  avoid a config surface that only some streams would actually respect (F6, REVIEW.md: a
  declared-but-partially-unwireable config key is worse than an absent one). `max_pages` is
  likewise not modeled for any stream; pagination is bounded only by each paginator's own natural
  stop signal (an empty `_metadata.next` value, or a short page).
