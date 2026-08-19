# Verification checklist — Google Ads connector parity wave03-r1 (#3021-#3028)

Credential-free only. No live Google Ads API calls or provider writes were made.

## Required local gates

- [x] Source inventory script: official v22 methods counted and local surface partition summarized.
  - `python3 .planning/phases/google-ads-parity-wave03-r1/generate_google_ads_parity.py`
  - Generated counts: `api_surface_rows=164`, `streams=3`, `direct_reads=21`, `write_actions=7`, `write_fixtures=7`, `blocked_rows=133`.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs/google-ads`
  - `connectorgen validate: 1 connector(s) checked, 0 findings`
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/google-ads' -count=1`
  - `ok polymetrics.ai/internal/connectors/conformance 3.285s`
- [x] `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1`
  - final rerun: `ok polymetrics.ai/internal/cli 162.881s`
- [x] `go build ./cmd/pm`
- [x] `make connector-boundary`
  - outcome `clean`, 0 findings
- [x] `make verify`
  - final run passed. Earlier attempts exposed the known long-running `internal/connectors/certify` package nearing the Makefile's 20m package timeout; the final run completed with connector packages revalidated.
- [x] `git diff --check`

## Additional targeted gates

- [x] Google Ads hook tests: `go test ./internal/connectors/hooks/google-ads -count=1`
  - `ok polymetrics.ai/internal/connectors/hooks/google-ads 0.360s`
- [x] Documentation generation/validation:
  - [x] `go run ./cmd/pm docs generate --dir docs/cli`
  - [x] `go run ./cmd/pm skills generate --dir docs/skills` (unrelated generated skill/doc churn reverted)
  - [x] `go run ./cmd/pm docs validate --connectors-dir docs/connectors`
  - [x] `node website/scripts/gen-connector-bundles.mjs && node website/scripts/gen-connectors.mjs && node website/scripts/gen-connector-catalog.mjs`
- [x] Help/inspect checks (no credentials):
  - [x] `./pm help connectors`
  - [x] `./pm connectors inspect google-ads`
  - [x] `go run ./cmd/pm connectors inspect google-ads --json` summarized as 3 streams and 7 write actions.
  - [x] `go run ./cmd/pm google-ads --help`
  - [x] `go run ./cmd/pm google-ads customers generate-ad-group-themes --help`

## Safety evidence

- [x] No secrets requested, printed, stored, summarized, or added to fixtures.
- [x] No live provider calls, writes, certification, VPS, Herdr, or Thaalam changes.
- [x] No new dependencies.
- [x] Reverse ETL write actions remain approval-gated and destructive/admin actions carry `confirm: destructive` metadata.
- [x] Blocked operation rows record connector-local evidence and are not advertised as executable.
