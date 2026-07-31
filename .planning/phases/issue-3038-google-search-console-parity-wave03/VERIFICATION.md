# Verification — Google Search Console parity wave03

## Required local gates

- [x] `go run ./cmd/connectorgen validate internal/connectors/defs/google-search-console` — passed, 1 connector checked, 0 findings.
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/google-search-console' -count=1` — passed.
- [x] `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1` — passed.
- [x] `go build ./cmd/pm` — passed.
- [x] `make connector-boundary` — passed with clean boundary report and existing approved exceptions only.
- [x] `make verify` — passed.
- [x] `git diff --check` — passed.

## Additional credential-free checks

- [x] Connector-owned contract tests: `TestGoogleSearchConsoleOfficialParityContracts` and `TestGoogleSearchConsoleDirectReadBodySchemasAreClosed` pass after red-first failure.
- [x] `go run ./cmd/pm help docs` read before docs generation.
- [x] `go run ./cmd/pm help google-search-console` shows the typed direct-read command surface.
- [x] Generated docs/catalog/website data regenerated and diff-reviewed; unrelated connector manual rewrites were reverted, leaving Google Search Console docs plus catalog/website generated surfaces.
- [x] CLI direct-read fixture test covers all three typed POST direct operations without live provider calls.
- [x] `gh-axi` issue addendum marker appended once on #3038–#3045 with truthful post-change counts and no certification claim.

## Safety assertions

- [x] No live provider calls or credential requests.
- [x] No secret-like fixture literals; only synthetic fixture tokens/URLs are used in tests.
- [x] No certification claim; `certification.json` is fixture/live-sweep metadata only.
- [x] No new dependencies, pushes, PR edits, no-mistakes run, VPS, Thaalam, or main merge.
