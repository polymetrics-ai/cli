# Summary — Twenty S7 fixtures/certify

Status: CORRECT-TURN63 review-fix slice complete locally; pending commit/push.

Delivered:
- Credential-free Twenty fixture coverage for 28 stream directories (29 page fixtures including existing two-page attachments) and 112 write fixtures.
- Focused conformance regression `TestTwentyFixturesCoverAllStreamsAndWrites`.
- `docs.md` S7 fixture/certification notes, parity-deviation ledger, and certify-harness limitation notes.
- Website generated connector data refreshed/idempotent.

Passed:
- JSON parse; `connectorgen validate` 0 findings/0 warnings; `make connectorgen-validate`; focused Twenty conformance; `go vet ./...`; `go test ./...`; `go build ./cmd/pm`; docs check; website generation idempotency; `git diff --check`.

Correction delivered:
- Certify pre-bootstrap connection now uses stream metadata from `connectors inspect --json` before catalog refresh, so Twenty attachments use `updatedAt` instead of fallback `updated_at`.
- Full-sweep certify connection/table/capture names now use bounded deterministic safe names for long stream names, preventing `_pm_raw` filename length failures.
- TURN63 F1: no-`--stream` certify defaults to first known cursor stream, else first known stream; hardcoded fallback only when specs are unavailable.
- TURN63 F2: conformance write fixtures now support exact decoded `body_exact` JSON (arrays included) and `no_body`; Twenty batch/delete fixtures assert those shapes.
- TURN63 F3: `fixture_conformance` runs for available bundle fixtures (Twenty) and real conformance failures fail certification; only missing bundles/fixtures skip.
- TURN63 F4: Twenty docs removed stale current limitations for fixed `updatedAt`/long-name certify issues and kept true live-credential/reverse-ETL caveats.

Correction gates passed:
- Focused certify red/green tests; broader focused certify tests; Twenty conformance focused tests; connectorgen validate JSON (`0 findings`, `0 warnings`); `go vet ./...`; `go build ./cmd/pm`; `gofmt -l cmd internal` clean.
- Credential-free localhost `pm connectors certify twenty` non-full and `--full --skip write` both exited 0 with `.report.passed=true`.
- TURN63 focused certify tests PASS (`ok polymetrics.ai/internal/connectors/certify 44.503s`); full certify package PASS (`334.029s`); conformance focused/package tests PASS (`1.183s` / `9.392s`); `go test ./...` PASS; localhost no-`--stream` certify non-full and full `--skip write` PASS with `fixture_conformance.passed=true`; website generated data idempotent after Twenty docs update; `git diff --check` PASS.

Blocked/not run:
- `make verify` not run because it executes `./pm reverse run` through `smoke-no-build`; correction forbids reverse ETL execution.

Safety: no secrets, no live Twenty API, no reverse ETL execution, no new dependencies.
