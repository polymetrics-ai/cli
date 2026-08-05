# TDD Ledger — Google Search Console documented-operation parity resume

| Cycle | Red evidence | Green evidence | Status |
| --- | --- | --- | --- |
| Recovery preflight | PR #3555 head `e7dc502c5` rehydrated and rebased onto current `origin/main`; stale shared changes were deliberately not restored. | Pending focused validator run. | In progress |
| Validator repair | `go run ./cmd/connectorgen validate google-search-console` exposed the CLI's directory argument contract rather than connector-name shorthand. The correct focused gate, `go run ./cmd/connectorgen validate internal/connectors/defs/google-search-console`, failed as expected with nine findings: four Search Analytics stream-target mismatches and five unrequired required-body flags. | The focused gate now reports `0 findings`; four duplicate convenience-command endpoint references were retired and the remaining required direct-read flags are explicit. | Green |
| Citation coverage | Provider-field matrix must show source/evidence/requiredness for every request field before canonical citation metadata is authored. | `REQUEST-FIELD-MATRIX.md` records 32/32 operation-specific request-field paths from Google-owned sources. Canonical citation metadata remains pending the shared convention landing. | Research complete |
| Command reachability | A fixture-backed `direct search-analytics query` with the documented `https://.../` site property failed before dispatch: `path variable siteUrl contains invalid character ':'`. | The false redundant direct command and its operation/certification candidate were retired. `search-analytics by-date` sent a properly encoded request to a local fixture server; URL Inspection, Mobile-Friendly Test, and Sites list also returned success from that server. | Green |
| Final verification | The pre-regeneration CLI suite failed only because the root manual golden transcript did not yet contain the newly exposed namespace. | The repository's `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1` generator updated the golden artifact; `go test ./internal/cli/...` then passed in 387.038s. Focused connector validation, conformance, runtime preflight, vet/build, docs validation, and website data generation all pass on `origin/main` `36b431cf1`. | Green |

The nine red findings are:

1. `search-analytics by-date` maps the shared POST endpoint to `search_analytics_by_query`, not `search_analytics_by_date`.
2. `search-analytics by-country` maps it to `search_analytics_by_query`, not `search_analytics_by_country`.
3. `search-analytics by-device` maps it to `search_analytics_by_query`, not `search_analytics_by_device`.
4. `search-analytics by-page` maps it to `search_analytics_by_query`, not `search_analytics_by_page`.
5. Direct Search Analytics `body.startDate` lacks a required `--start-date` flag.
6. Direct Search Analytics `body.endDate` lacks a required `--end-date` flag.
7. Direct URL Inspection `body.inspectionUrl` lacks a required `--inspection-url` flag.
8. Direct URL Inspection `body.siteUrl` lacks a required `--site-url` flag.
9. Direct Mobile-Friendly Test `body.url` lacks a required `--url` flag.

No new production edit for this resume phase occurred before this current-main validator red evidence.

The sixth omitted `required` declaration (`direct search-analytics query --site-url`) was
also identified during implementation as semantically necessary for a provider-required path.
It is no longer applicable because the direct command was removed after the runtime test
proved it could not accept Google’s documented URL-valued path input under current-main
safety rules. The Search Analytics operation remains implemented through its ETL streams.
