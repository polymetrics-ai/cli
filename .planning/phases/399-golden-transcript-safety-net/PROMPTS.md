# Prompts — Issue 399 Golden Transcript Safety Net

## Kickoff snapshot

Task: Execute polymetrics-ai/cli#399 as one bounded mutating worker for parent #397.

Worker branch: `test/399-golden-transcript-safety-net`  
Worker directory: `/Users/karthiksivadas/Development/polymetrics-cli-agents/wt-399-golden-transcript-safety-net`  
Parent branch: `feat/cli-architecture-v2`  
Parent PR: https://github.com/polymetrics-ai/cli/pull/438

Required command path:

```bash
scripts/gsd doctor
scripts/gsd prompt plan-phase 399 --skip-research >/tmp/gsd-plan-phase-399.prompt
scripts/gsd prompt programming-loop init --phase 399 --dry-run >/tmp/gsd-programming-loop-399.prompt
```

Programming-loop result:

```text
scripts/gsd: unknown GSD command: programming-loop
```

Fallback: use `.pi/prompts/pm-gsd-loop.md` as manual GSD/TDD contract.

## Downstream artifact

- `PLAN.md`: created before test harness edits.
- `TDD-LEDGER.md`: created; red evidence pending.
- `VERIFICATION.md`: created; results pending.
- `RUN-STATE.json`: created; `verificationPassed=false` until full `make verify` passes.

## Verification result

Passed local gates:

- `go test ./internal/cli/ -run Golden -count=1`
- `go test ./internal/cli/ -count=1`
- `go vet ./...`
- `go test ./...`
- `go build ./cmd/pm`
- `make verify`
- `git diff -- go.mod` (empty)
- CLI parity spot checks for `pm help docs`, bare `pm connectors`, `pm docs --help`, and docs/website grep.
