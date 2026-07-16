# Issue 400 Prompt Trace — Cobra Router Shell

## Kickoff snapshot

**Task:** Execute polymetrics-ai/cli#400 as one bounded mutating worker for parent #397.

**Branch:** `refactor/400-cobra-router-shell`
**Worker directory:** `/Users/karthiksivadas/Development/polymetrics-cli-agents/wt-400-cobra-router-shell`
**Sub-PR base:** `feat/cli-architecture-v2`
**Parent PR:** #438 draft (`feat/cli-architecture-v2` -> `main`)

**Allowed write scope:** root Cobra tree/router wrappers, Cobra/pflag error mapping, focused CLI tests under `internal/cli/**`; `go.mod` / `go.sum` only for exact ADR-0002-approved Cobra v1.10.x and expected pflag/mousetrap transitives; issue-local GSD artifacts under `.planning/phases/400-*`; minimal CLI docs/golden fixture changes only if acceptance requires.

**Required command path:**

```bash
scripts/gsd doctor
scripts/gsd prompt plan-phase 400 --skip-research >/tmp/gsd-plan-phase-400.prompt
scripts/gsd prompt programming-loop init --phase 400 --dry-run >/tmp/gsd-programming-loop-400.prompt
```

**Actual command path:**

- `scripts/gsd doctor` passed.
- `scripts/gsd prompt plan-phase 400 --skip-research >/tmp/gsd-plan-phase-400.prompt` passed; prompt length 142 lines.
- `scripts/gsd prompt programming-loop init --phase 400 --dry-run >/tmp/gsd-programming-loop-400.prompt` failed with `scripts/gsd: unknown GSD command: programming-loop`; using `.pi/prompts/pm-gsd-loop.md` fallback.

## Downstream artifacts

- `PLAN.md`: created.
- `TDD-LEDGER.md`: created.
- `VERIFICATION.md`: created.
- `SUMMARY.md`: created.
- `RUN-STATE.json`: created.

## Verification result

Focused/full local gates and post-commit diff checks passed for initial implementation. PR #440 open. Review-fix cycle completed locally after pm-reviewer findings; push pending at artifact update time.

Key initial results:

- `go test ./internal/cli/ -run Golden -count=1` passed byte-identical.
- `go test ./internal/cli/ -run Certify -count=1` passed.
- `go test ./internal/cli/ -count=1` passed.
- `go vet ./...` passed.
- `go test ./...` passed.
- `go build ./cmd/pm` passed.
- `make verify` passed.
- `git diff --check origin/feat/cli-architecture-v2...HEAD` passed.
- `git diff origin/feat/cli-architecture-v2...HEAD -- go.mod go.sum` recorded expected Cobra/pflag/mousetrap dependency delta.

Review-fix results:

- Red: `go test ./internal/cli/ -run TestCobraRouterShell -count=1` failed on missing persistent root flags and legacy error reclassification.
- Green: `go test ./internal/cli/ -run TestCobraRouterShell -count=1` passed: `ok  	polymetrics.ai/internal/cli	2.026s`.
- Exact gates passed: `gofmt -w cmd internal`; `go test ./internal/cli/ -run 'TestCobraRouterShell|Golden' -count=1`; `go test ./internal/cli/ -run Certify -count=1`; `go test ./internal/cli/ -count=1`; `go vet ./...`; `go test ./...`; `go build ./cmd/pm`; `make verify`; `git diff --check origin/feat/cli-architecture-v2...HEAD`; `git diff origin/feat/cli-architecture-v2...HEAD -- go.mod go.sum`.
