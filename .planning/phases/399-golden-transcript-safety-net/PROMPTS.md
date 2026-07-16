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

- `PLAN.md`: created before test harness edits; updated with review-fix dispositions and `local_critical_path` decision.
- `TDD-LEDGER.md`: records initial red/absent evidence, docs-diff red, green evidence, and review-fix validation.
- `VERIFICATION.md`: records original gates and review-fix requested gates.
- `RUN-STATE.json`: updated with allowed `docs/cli/**` scope evidence and `verificationPassed=true` after review-fix `make verify` passed.

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

Sub-PR: https://github.com/polymetrics-ai/cli/pull/439  
Pre-review-fix head SHA: `d7ffbb1ee01b709a3470f62976cba65c2c586921`; review-fix commit recorded in worker handoff.
Automated review: Claude workflow observed `disabled_manually`; coverage pending / parent-PR fallback pending. Review-fix cycle did not post `@claude review`.
