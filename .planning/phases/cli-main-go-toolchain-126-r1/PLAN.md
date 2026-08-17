# Phase cli-main-go-toolchain-126-r1 — align main's Go security toolchain

## GSD and skills

- GSD setup passed: `scripts/gsd doctor`; `go run ./cmd/agentcontractgen check`.
- Resolved commands: `discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and
  `code-review` via `scripts/gsd sources <command>`.
- Manual-GSD fallback: execute generated `scripts/gsd prompt` instructions inline because Pi is
  unavailable and the canonical single-worker contract forbids role spawning.
- Skills loaded: `golang-how-to`, `golang-continuous-integration`,
  `golang-dependency-management`, `golang-security`, and `golang-testing`.

## Scope

Update active Go toolchain pins only: `go.mod`'s `toolchain` directive and setup-go (including
the security job's explicit `GOTOOLCHAIN`) in the five named workflows. Retain `go 1.25.4` and do
not alter completed historical planning records.

## TDD execution plan

| Order | Slice | RED evidence | GREEN implementation | Guard |
| --- | --- | --- | --- | --- |
| 1 | Active toolchain inventory | The scoped source search reports `1.25.12` in every active pin. | Replace only active toolchain pins with `1.26.6`. | The language directive remains `go 1.25.4`; no historical `.planning` record changes. |
| 2 | Vulnerability gate | `govulncheck` on the old pin is known to fail for the listed Go standard-library advisories. | Run `govulncheck ./...` using Go 1.26.6. | Verify the command exits successfully and no `1.25.12` remains in active target files. |

## Commit checkpoints

1. Planning/TDD evidence and scoped toolchain pin.
2. Green local vulnerability and restraint verification.
3. Review-fix checkpoint only for bounded in-scope findings.
