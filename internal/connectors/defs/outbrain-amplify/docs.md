# Overview

Outbrain Amplify is a wave2 declarative-HTTP migration. It reads Outbrain Amplify marketers,
campaigns, and performance reports through the Outbrain Amplify REST API
(`GET https://api.outbrain.com/amplify/v0.1/...`). This bundle is a Tier-1 pure declarative port of
`internal/connectors/outbrain-amplify` (the hand-written connector it migrates, itself a
connsdk-HTTP-based read-only connector — no signature auth, no async polling, no protocol other
than plain REST — so no Go escape hatch is needed); the legacy package stays registered and
unchanged until wave6's registry flip.

## Auth setup

Two credential shapes are supported, matching legacy's own first-match-wins precedence
(`outbrain_amplify.go`'s `authenticator`, which checks `access_token` before `username`/`password`):

1. **Bearer token** (preferred): set the `access_token` secret. Sent as `Authorization: Bearer
   <access_token>`.
2. **HTTP Basic** (fallback for local/proxy deployments that expose that mode): set the `username`
   config value and the `password` secret. Used only when `access_token` is not set.

`streams.json`'s `base.auth` declares bearer FIRST with `when: "{{ secrets.access_token }}"`, then
basic with `when: "{{ secrets.password }}"` — reproducing legacy's exact precedence when both
credential shapes happen to be configured (`docs/migration/conventions.md`'s dual-auth ordering
rule). Neither credential is ever logged.

## Streams notes

All 3 streams share the same emitted-record shape (`id`, `name`, `enabled`, `status`, `created_at`,
`impressions`, `clicks`, `spend`) — legacy's own `record()` mapper is identical across all three
streams (`outbrain_amplify.go:270`).

- `marketers` — `GET /marketers`, records at `marketers`.
- `campaigns` — `GET /campaigns`, records at `campaigns`.
- `performance_reports` — `GET /reports`, records at `results`. Optional report filters
  (`start_date`, `end_date`, `report_granularity`, `conversion_count`, `geo_location_breakdown`) are
  sent only when their config value is set (`stream.Query`'s `omit_when_absent` dialect), matching
  legacy's own conditional-set loop (`outbrain_amplify.go:99-103`).

Pagination is `offset_limit` (`limit`/`offset` query params, `page_size: 100` — legacy's own
`defaultPageSize`), stopping on a short/empty page. Legacy ALSO stops early when the running
offset reaches the response body's `totalResults`/`total` field; the engine's `offset_limit`
paginator implements only the short-page stop signal. This can cause, at most, one harmless extra
request on the rare page where a full-size page happens to exactly exhaust the total (the
following request then returns an empty/short page and stops normally) — it never omits,
duplicates, or reorders any record for any input legacy itself would accept (same class of
deviation as `docs/migration/conventions.md`'s jamf-pro ledger entry).

## Write actions & risks

None. Legacy's `Write` returns `connectors.ErrUnsupportedOperation`; `capabilities.write` is
`false` and this bundle ships no `writes.json`.

## Known limits

- **`marketer_id`-scoped paths are not modeled.** Legacy optionally scopes `campaigns` and
  `performance_reports` to `/marketers/{marketer_id}/campaigns`/`/reports` when a `marketer_id`
  config value is set (`campaignsPath`/`reportsPath`, `outbrain_amplify.go:251-263`); the
  unscoped `/campaigns`/`/reports` paths are legacy's own base-case fallback when `marketer_id` is
  absent (the common case). Path templating has no conditional-branch mechanism for an OPTIONAL
  config value that changes the path structure itself (as opposed to a query param, which the
  `omit_when_absent` dialect handles) — modeling both shapes would need a second stream declaration
  or a Tier-2 hook, neither justified for this narrow a scoping feature. Only the unscoped base case
  is implemented; out of scope, not silently wrong (same pattern as searxng's unmodeled subreddit
  narrowing, `docs/migration/conventions.md` ledger item 7). `marketer_id` remains declared in
  `spec.json` for documentation/forward-compat purposes even though no template currently consumes
  it — unlike a dead config key with no future use, this one names a real, well-understood legacy
  capability intentionally deferred to Pass B.
- **`page_size`/`max_pages` runtime overrides are declared but `max_pages` has no engine-side
  enforcement mechanism for `offset_limit` pagination** — the engine's `offset_limit` paginator
  (`connsdk.OffsetPaginator`) has no request-count cap knob analogous to `page_number`'s implicit
  `MaxPages` read-path enforcement; termination is bounded solely by the short/empty-page stop
  signal, matching Outbrain's own real termination behavior in every case that does not hit the
  `totalResults` edge case documented above.
- Full Outbrain Amplify surface (budget/creative/promoted-link management, marketer-scoped
  campaign/report paths, campaign mutation) is out of scope for this wave; see
  `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B capability expansion"}`
  entries.
