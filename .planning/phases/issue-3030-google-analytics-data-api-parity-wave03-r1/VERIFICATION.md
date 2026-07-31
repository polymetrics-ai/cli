# Verification checklist: Google Analytics Data API parity wave03 r1

## Required gates

- [x] `go run ./cmd/connectorgen validate internal/connectors/defs/google-analytics-data-api` — pass, 1 connector, 0 findings.
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/google-analytics-data-api' -count=1` — pass.
- [x] `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1` — pass.
- [x] `go build ./cmd/pm` — pass.
- [x] `make connector-boundary` — pass, clean boundary report.
- [x] `make verify` — pass. Note: first two attempts hit the repo's hardcoded 20m package timeout while warming `internal/cli`/`internal/connectors/certify`; reran the slow certify package once, then `make verify` passed with cached package results plus fresh build/docs/smoke/lint/connectorgen/boundary/release checks.
- [x] `git diff --check` — pass.

## Focused gates

- [x] `go test ./cmd/connectorgen -run 'TestValidate_(AcceptsSingleBundleDir|APISurfaceAllowsPOSTBackedReadStreamWhenWriteFalse|CLISurface)' -count=1` — pass.
- [x] `go test ./cmd/connectorgen ./internal/connectors/conformance ./internal/connectors/native/google-analytics-data-api ./internal/connectors/hooks/google-analytics-data-api -count=1` — pass.
- [x] Generated docs/skills/catalog/website diff reviewed: JSON generated catalogs changed exactly one connector entry (`google-analytics-data-api` / `Google Analytics 4 (GA4)`); no `*_gen.go` diffs from accidental generator run.

## CLI/docs parity checklist

- [x] `pm help connectors` / root golden behavior updated only for the new GA provider command line.
- [x] `pm connectors inspect google-analytics-data-api` generated manual reflects auth, streams, operation counts, blocked/planned rows, and no live certification.
- [x] `pm connectors inspect google-analytics-data-api --json` remains metadata-only/secret-free by construction; validate, docs validate, and golden tests passed.
- [x] Provider command help is definition-owned (`cli_surface.json`) and exposes fixed commands only; no generic raw API escape hatch exists.
- [x] `docs/connectors/google-analytics-data-api/MANUAL.md` and `SKILL.md` updated.
- [x] `docs/connectors/README.md`, `docs/connectors/catalog/all-connectors.{json,md}`, and website connector generated data updated with GA-only diffs.

## Issue update checklist

- [x] Append idempotent captain-policy addendum to #3030 — appended with marker `ga4-parity-wave03-r1-captain-addendum`.
- [x] Append idempotent captain-policy addendum to #3031-#3037 — appended with same marker.
- [x] Addendum includes actual counts (11 official, 4 executable, 7 blocked/planned; 5 stream fixtures, 3 direct fixtures) and no certification claim.
