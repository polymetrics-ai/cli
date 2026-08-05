# Verification checklist — issue-3752-rate-limit-admission-r1

Status: **planned; no implementation verification claimed yet**.

## Targeted behavior gates

- [ ] Run the red B1 defect test before modifying requester retry code; capture the actual `30s` vs `90s` failure in `TDD-LEDGER.md`.
- [ ] `go test ./internal/connectors/connsdk -run 'TestRequester.*(RateLimit|RetryAfter|Admission|Observation|Jitter)' -count=1`
- [ ] `go test ./internal/connectors/engine -run 'TestBundleLoad.*RateLimit|Test.*RateLimit.*Declaration' -count=1`
- [ ] `go test -race ./internal/connectors/connsdk -count=1`
- [ ] Run the changed-package suites: `go test ./internal/connectors/connsdk ./internal/connectors/engine ./internal/connectors/bundleregistry ./cmd/connectorgen -count=1`.
- [ ] Confirm a terminal 429 is both `*connsdk.RateLimitError` and its wrapped safe `*connsdk.HTTPError`; no fixture credential text appears in `err.Error()` or observer values.
- [ ] Confirm admission cancellation occurs before any `httptest` handler hit for JSON/form/multipart and `DoStream`.

## Loader and fleet compatibility gates

- [ ] `go test ./internal/connectors/engine -run 'TestBundleLoad|TestNewLoadsDeclarativeBundlesWithHooksAndNativeOverrides' -count=1`
- [ ] `go run ./cmd/connectorgen validate`
- [ ] `go run ./cmd/connectorgen surface-sync --check`
- [ ] Confirm `rate_limits.json` is optional and no existing bundle declaration is edited or required.
- [ ] Confirm no command `availability: implemented`, `api_surface`, `operations.json`, `cli_surface.json`, or commandrunner preflight contract changed. The 213-command defect class is therefore not expanded by this foundation.

## Hygiene and project gates

- [ ] `gofmt -w internal/connectors/connsdk internal/connectors/engine`
- [ ] `go vet ./internal/connectors/connsdk ./internal/connectors/engine ./internal/connectors/bundleregistry ./cmd/connectorgen`
- [ ] `go build ./cmd/pm`
- [ ] Run `make verify`'s non-full-suite gates separately: `tidy-check`, `lint`, `docs-check`, `smoke-no-build`, `agent-contract-check`, `connectorgen-validate`, `connectorgen-surface-sync`, `connector-boundary`, and `release-workflow-check`.
- [ ] `scripts/verify-gsd-workflow origin/main`
- [ ] Check changed paths and enforce scope: no `internal/connectors/defs/<connector>/` migration, no `commandrunner`, and no deferred #3753/#3754/#3755 engine/CLI surface edits.

## Deliberately not applicable in this slice

- CLI help/manual/website parity: #3755 owns the first operator-visible rate-limit surface. This
  foundation adds no command, flag, help topic, output format, or website documentation.
- Live/integration provider tests: prohibited. Only unit fixtures and `httptest` are permitted.
- Full `go test ./...` and monolithic `make verify`: CI owns these due the documented 550-connector
  timeout limitation; individual non-test gates remain local requirements.

## Final review sequence

After local gates, generate and execute the GSD `execute-phase`, `verify-work`, and `code-review`
prompts inline; record real outputs, plan gaps if any, and do not start no-mistakes until firstmate
directs the validation/ship stage. Review findings are dispositioned, never silently ignored.
