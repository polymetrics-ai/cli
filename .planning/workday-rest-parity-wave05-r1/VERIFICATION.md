# Verification checklist — Workday REST parity wave05-r1

Completed verification:

- [x] Official inventory generation summary from fetched Workday 2026.30 docs: 52 production services, 920 HTTP operations (`GET` 655, `POST` 154, `PATCH` 58, `PUT` 20, `DELETE` 33).
- [x] Post-change Workday counts from generated files: `api_surface` rows 920; streams 463; direct read operations/commands 174; writes 251; blocked binary/file or current-contract gaps 32; stream fixtures 463; write fixtures 251; certified 0.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs --json` filtered for `workday-rest`: 0 findings.
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/workday-rest' -count=1`: pass.
- [x] `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1 -timeout=10m`: pass.
- [x] `go run ./cmd/pm docs validate --dir docs/cli --connectors-dir docs/connectors --website-dir website/content/docs`: pass.
- [x] `go vet ./internal/connectors/... ./internal/cli/...`: pass.
- [x] `go build ./cmd/pm`: pass.
- [x] `make connector-boundary`: clean.
- [x] `git diff --check`: pass.
- [x] `GOFLAGS='-p=1' make verify`: pass (serial package execution used after default parallel `make verify` hit local package-test timeout/SIGTERM instability; no code changes were made for that environment issue).
- [x] Update parent #3231 and subissues #3232-#3238 once through `gh-axi` with truthful counts and verification evidence.
- [x] Clean commit on `fm/cli-workday-rest-parity-wave05-r1`.
