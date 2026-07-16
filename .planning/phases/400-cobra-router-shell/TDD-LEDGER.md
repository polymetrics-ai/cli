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

Actual red evidence captured before router implementation:

```bash
set -o pipefail; go list -deps ./internal/cli | grep '^github.com/spf13/cobra$'
```

Result: no output, exit 1 (Cobra absent from `internal/cli` dependency graph).

```bash
go test ./internal/cli/ -run TestCobraRouterShell -count=1
```

Result: setup failure, exit 1:

```text
# polymetrics.ai/internal/cli
internal/cli/cobra_router_test.go:8:2: no required module provides package github.com/spf13/cobra; to add it:
	go get github.com/spf13/cobra
FAIL	polymetrics.ai/internal/cli [setup failed]
FAIL
```

## Cycle 2 — green implementation evidence

Implemented Cobra router shell with fresh `newRootCmd` per `Run`, `DisableFlagParsing` top-level wrappers, hidden `extract`/`worker`, legacy help/manual rendering, dynamic connector fallback, and `mapCobraErr` feeding `writeError`.

```bash
go test ./internal/cli/ -run TestCobraRouterShell -count=1
```

Result:

```text
ok  	polymetrics.ai/internal/cli	0.532s
```

```bash
go test ./internal/cli/ -run Golden -count=1
```

First result after initial implementation: failed because Cobra used Go test process args when `Run(nil, ...)` passed a nil arg slice to `SetArgs`; fixed by forcing a non-nil empty arg slice for root execution.

```text
--- FAIL: TestGoldenTranscripts (5.33s)
    --- FAIL: TestGoldenTranscripts/root_bare_manual (0.50s)
        golden_transcript_test.go:154: exit code = 2, want 0
            stdout=
            stderr=error: unknown command "-test.paniconexit0"
FAIL
FAIL	polymetrics.ai/internal/cli	6.826s
FAIL
```

Green rerun:

```text
ok  	polymetrics.ai/internal/cli	6.446s
```

Focused gates after fix/refactor:

```bash
gofmt -w internal/cli/cobra_router.go
go test ./internal/cli/ -run Golden -count=1
go test ./internal/cli/ -run Certify -count=1
go test ./internal/cli/ -count=1
```

Result:

```text
ok  	polymetrics.ai/internal/cli	6.369s
ok  	polymetrics.ai/internal/cli	95.783s
ok  	polymetrics.ai/internal/cli	154.218s
```

Dependency delta:

- Direct: `github.com/spf13/cobra v1.10.2`.
- Indirect go.mod: `github.com/spf13/pflag v1.0.9`, `github.com/inconshreveable/mousetrap v1.1.0`.
- go.sum also includes Cobra module-metadata checksums for `github.com/cpuguy83/go-md2man/v2 v2.0.6/go.mod`, `github.com/russross/blackfriday/v2 v2.1.0/go.mod`, and `go.yaml.in/yaml/v3 v3.0.4/go.mod`; `go mod why -m` reports the main module does not import those modules. No additional direct dependency introduced.

## Cycle 3 — refactor evidence

Refactor kept behavior byte-identical:

- `Run` now delegates to `newRootCmd` + `executeRootCmd`; top-level switch removed from `Run`.
- Legacy handlers and `parseFlags` remain unchanged under wrappers.
- `runManualAlias` fixed to preserve `--json help` / `--json man` root manual envelopes.
- Root fallback preserves legacy `nosuch --help` / dynamic connector `github help` manual interception before connector dispatch.
- Golden fixture file was not regenerated or edited.

Refactor gate:

```bash
git diff --check && git diff --cached --check
```

Result: no output, exit 0.

## Final verification evidence

See `VERIFICATION.md`. Full `make verify` initially failed at tidy-check while dependency files were intentionally unstaged; after staging `go.mod`/`go.sum`, the same gate passed. No quality gate was weakened.

Post-commit checks:

```bash
git diff --check origin/feat/cli-architecture-v2...HEAD
git diff origin/feat/cli-architecture-v2...HEAD -- go.mod go.sum
```

Result: whitespace check passed with no output; dependency diff shows only the recorded Cobra v1.10.2 / pflag / mousetrap delta plus go.sum module-metadata checksums.

Post-self-review fallback-help test and final gates after commit `e027a12e`:

```bash
go test ./internal/cli/ -run 'TestCobraRouterShell|Golden' -count=1
# ok  	polymetrics.ai/internal/cli	6.374s

go test ./internal/cli/ -run Certify -count=1
# ok  	polymetrics.ai/internal/cli	91.679s

go vet ./...
go test ./...
go build ./cmd/pm
make verify
git diff --check origin/feat/cli-architecture-v2...HEAD
```

Result: pass; `make verify` ended with `connectorgen validate: 547 connector(s) checked, 0 findings`.
