# Verification checklist — Recurly parity-resume r1

## Required gates

- [x] `go run ./cmd/connectorgen surface-sync`
- [x] `go run ./cmd/connectorgen surface-sync --check`
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs/recurly --json`
- [x] `go test ./internal/connectors/conformance/...`
- [x] `go test ./internal/connectors/commandrunner -run TestEveryImplementedCommandPassesRuntimePreflight`
- [x] `go test ./internal/cli/...` (green after regenerating the required root CLI transcript)
- [x] `go vet` for changed packages
- [x] `go build ./cmd/pm`
- [x] `pm recurly --help`, bare `pm recurly`, and all three binary command help paths; fixture-backed commandrunner execution covers the binary requests without credentials or live calls
- [x] `cd website && pnpm run gen:website-data`
- [ ] `git diff --check`

## Scope and safety

- [ ] Changed files remain Recurly-owned bundle/generated surfaces plus this phase's artifacts.
- [ ] No credentials, live Recurly calls, reverse-ETL execution, or shared runtime/schema edits.
- [x] Provider OpenAPI source, operation total, and reachable count are recorded honestly.
- [ ] Final citation-convention metadata is applied and field-citation coverage is recorded.

## GSD verification status

`verificationPassed` remains false in `RUN-STATE.json` until a declared full verification equivalent
passes. The shared resume contract intentionally replaces `make verify` with focused gates because
the complete suite exceeds bounded execution windows.
