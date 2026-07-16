# Phase 406 Prompts

## Kickoff snapshot

- Command: `scripts/gsd prompt plan-phase 406 --skip-research >/tmp/gsd-plan-phase-406.prompt`
- Programming-loop attempt: `scripts/gsd prompt programming-loop init --phase 406 --dry-run >/tmp/gsd-programming-loop-406.prompt`
- Downstream artifact: `PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, `SUMMARY.md`, `RUN-STATE.json`
- Verification result: local gates passed; PR creation pending.

## Adapter gap

`programming-loop` is not present in `.gsd/commands.json`; manual GSD fallback follows `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`.

## Terminal snapshot

- Red tests: `go test ./internal/cli/ -run 'Catalog|CobraRouterShell' -count=1`; `go test ./internal/cli/ -run 'TestCatalogConnectionFlagFormsPreserveLegacySemantics' -count=1`.
- Green focused: `go test ./internal/cli/ -run 'Catalog|CobraRouterShell|Golden' -count=1`; `go test ./internal/cli/ -run Certify -count=1`.
- Full gates: `gofmt -w cmd internal`, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify`.
- Downstream artifact: `TDD-LEDGER.md`, `VERIFICATION.md`, `SUMMARY.md`, `RUN-STATE.json`.
- Verification result: passed; review route remains human/parent fallback pending, Claude/Copilot not requested.
