# Issue 400 TDD Ledger — Cobra Router Shell

**Issue:** #400
**Parent:** #397
**Branch:** `refactor/400-cobra-router-shell`
**Sub-PR base:** `feat/cli-architecture-v2`

## Loaded skills

- `gsd-core` — `.pi/skills/gsd-core/SKILL.md`.
- `caveman` — `.agents/skills/caveman/SKILL.md`.
- `golang-how-to` — routing table: CLI command tree -> `golang-spf13-cobra` + `golang-cli`; writing tests -> `golang-testing`; security audit -> `golang-security` + `golang-safety`.
- `golang-cli` — `SilenceUsage`/`SilenceErrors`, stdout/stderr discipline, injected writer tests.
- `golang-testing` — Best Practices #1 named subtests, #3 independent tests, #5 observable behavior.
- `golang-error-handling` — Best Practices #5 error chain inspection, #7 single handling rule.
- `golang-design-patterns` — Best Practices #3 explicit constructors/no hidden init, #20 dependency restraint, #21 testability.
- `golang-safety` — Best Practices #10 useful zero values, #11 avoid unnecessary shared state; Cobra state must not leak across invocations.
- `golang-spf13-cobra` — Best Practices #1 `RunE`, #4 injected writers, #5 fresh command tree per test/run; `DisableFlagParsing` available for strangler wrappers.
- `golang-security` — Security Thinking Model #1-#3 for untrusted argv; avoid generic shell/write surfaces and command injection patterns.
- `golang-documentation` — loaded for CLI help/manual parity review; docs changes currently expected not applicable.

Missing repo-local stack skill: `.pi/skills/go-implementation/SKILL.md` returned `ENOENT`; recorded in plan and handoff. User-required Go skills were loaded.

## GSD evidence

```bash
scripts/gsd doctor
```

Result: pass (`ok` for node, repo root, official docs, command registry, upstream lock, Pi settings/extension/skill/prompt, commands=69).

```bash
scripts/gsd prompt plan-phase 400 --skip-research >/tmp/gsd-plan-phase-400.prompt && test -s /tmp/gsd-plan-phase-400.prompt && wc -l /tmp/gsd-plan-phase-400.prompt
```

Result: pass (`142 /tmp/gsd-plan-phase-400.prompt`).

```bash
scripts/gsd prompt programming-loop init --phase 400 --dry-run >/tmp/gsd-programming-loop-400.prompt
```

Result: expected adapter gap, exit 1:

```text
scripts/gsd: unknown GSD command: programming-loop
```

Manual fallback: `.pi/prompts/pm-gsd-loop.md` loaded and followed.

## Dependency red / decision evidence

```bash
go list -m -versions github.com/spf13/cobra
```

Result: `v1.10.0 v1.10.1 v1.10.2` available in approved ADR 0002 v1.10.x range. Selected `github.com/spf13/cobra v1.10.2`.

## Cycle 0 — planning setup

Status: in progress.

Artifacts created before production edits:

- `.planning/phases/400-cobra-router-shell/PLAN.md`
- `.planning/phases/400-cobra-router-shell/TDD-LEDGER.md`
- `.planning/phases/400-cobra-router-shell/VERIFICATION.md`
- `.planning/phases/400-cobra-router-shell/SUMMARY.md`
- `.planning/phases/400-cobra-router-shell/PROMPTS.md`
- `.planning/phases/400-cobra-router-shell/RUN-STATE.json`

## Cycle 1 — red validation/test evidence

Planned commands before router implementation:

```bash
go list -deps ./internal/cli | grep '^github.com/spf13/cobra$'
go test ./internal/cli/ -run TestCobraRouterShell -count=1
```

Expected red:

- Cobra dependency absent before Phase 2 implementation.
- Focused test fails to compile/run until `newRootCmd` / Cobra shell exists.

Actual red evidence: pending.

## Cycle 2 — green implementation evidence

Pending. Target commands:

```bash
go test ./internal/cli/ -run TestCobraRouterShell -count=1
go test ./internal/cli/ -run Golden -count=1
go test ./internal/cli/ -run Certify -count=1
go test ./internal/cli/ -count=1
```

## Cycle 3 — refactor evidence

Pending. Refactor only while golden suite remains byte-identical.

## Final verification evidence

Pending; see `VERIFICATION.md`.
