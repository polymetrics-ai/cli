# TDD LEDGER — issue 3585

## Mandatory fallback note

`pm-gsd-worker` could not be spawned in this Pi session because parent scout smoke reported missing `grep`/`find`/`ls` parent tools. This worker is the project-agent fallback and must not spawn subagents. `scripts/gsd prompt programming-loop ...` is unavailable (`unknown GSD command: programming-loop`), so this ledger records manual GSD universal-loop evidence.

## Planned red/green evidence

| Slice | Red evidence | Green evidence | Status |
| --- | --- | --- | --- |
| Ledger completeness | `go test ./.planning/phases/connector-guardrail-remediation-r1/workers/issue-3585` failed before `DISPOSITION-LEDGER.md` existed (`open DISPOSITION-LEDGER.md: no such file or directory`) | Same command passes after ledger rows exist and classify each audited path | Green |
| Existing shared foundation inspection | Targeted package tests identify whether Stripe DELETE/no-body and commandrunner/connectorgen foundations are already covered | `go test ./internal/connectors/engine ./internal/connectors/commandrunner ./cmd/connectorgen` passed; no production correction needed | Green |
| Production correction, if required | Not triggered: no focused test exposed a current production bug | Not applicable; ledger/proof-only implementation | Not applicable |
| Verification | `git diff --check`, worker Go test, focused package tests, vet/build/no-mistakes doctor | Passing/blocked results recorded in `VERIFICATION.md` | Green |

## Required audited rows

### Stripe #3530 (`86d510927a05aa56b184bf5a8778b5444c69b9b1`)

- `internal/connectors/engine/write.go` — shared runtime/tooling, production
- `internal/connectors/engine/write_test.go` — shared runtime/tooling, test proof

### Google Ads #3535 (`5d61794f76c42cc7a97a4b29f6ffcc09dd39dbae`)

- `cmd/connectorgen/main.go` — shared runtime/tooling, production
- `cmd/connectorgen/main_test.go` — shared runtime/tooling, test proof
- `cmd/connectorgen/validate.go` — shared runtime/tooling, production
- `internal/cli/cli.go` — shared runtime/tooling, production command-surface wiring
- `internal/connectors/command_surface.go` — shared runtime/tooling, production metadata surface
- `internal/connectors/commandrunner/runner.go` — shared runtime/tooling, production runner
- `internal/connectors/commandrunner/runner_test.go` — shared runtime/tooling, test proof
- `internal/connectors/engine/bundle.go` — shared runtime/tooling, production metadata loader
- `internal/connectors/engine/connector.go` — shared runtime/tooling, production definition summary
- `internal/connectors/engine/schema/cli_surface.schema.json` — shared runtime/tooling, schema contract
- `internal/connectors/hooks/google-ads/hooks.go` — connector-owned hook/native surface, allowed Google Ads scope
- `internal/connectors/hooks/google-ads/hooks_test.go` — connector-owned hook/native surface test, allowed Google Ads scope
- `internal/connectors/defs/gong/cli_surface.json` — unrelated connector definition, not remediated here; record as delegated to #3586

### Freshchat #3536 (`b053dc4a3ad7f9895637a09560ea8a9a76bec507`)

- `internal/connectors/commandrunner/runner.go` — shared runtime/tooling, production runner
- `internal/connectors/commandrunner/runner_test.go` — shared runtime/tooling, test proof

## Actual evidence log

- RED: `go test ./.planning/phases/connector-guardrail-remediation-r1/workers/issue-3585` failed because `DISPOSITION-LEDGER.md` did not exist.
- GREEN: `go test ./.planning/phases/connector-guardrail-remediation-r1/workers/issue-3585` passed after adding `DISPOSITION-LEDGER.md`.
- GREEN: `go test ./internal/connectors/engine ./internal/connectors/commandrunner ./cmd/connectorgen` passed, supporting a ledger-only disposition with no production correction.
