# Issue 400 Plan — Cobra Router Shell

**Issue:** [#400](https://github.com/polymetrics-ai/cli/issues/400)
**Parent:** [#397](https://github.com/polymetrics-ai/cli/issues/397)
**Parent PR:** [#438](https://github.com/polymetrics-ai/cli/pull/438) (`feat/cli-architecture-v2` → `main`, draft)
**Worker branch:** `refactor/400-cobra-router-shell`
**Sub-PR base:** `feat/cli-architecture-v2`
**Parent dependency integrated:** #399 via PR #439, parent commit `379cb5015335ff7c9b20e5bb780952ead22c53b2`
**Mode:** spawned bounded mutating worker / stacked sub-PR
**GSD command path:** `scripts/gsd doctor`; `scripts/gsd prompt plan-phase 400 --skip-research`; `scripts/gsd prompt programming-loop init --phase 400 --dry-run` failed with `scripts/gsd: unknown GSD command: programming-loop`, so `.pi/prompts/pm-gsd-loop.md` is the recorded manual GSD programming-loop fallback.

## Objective

Replace the hand-written top-level `Run` switch with a fresh-per-invocation Cobra router shell while keeping the CLI transcript byte-identical. Cobra owns only top-level routing/help/error plumbing for this phase; all legacy handlers and `parseGlobal`/`parseFlags` semantics stay underneath `DisableFlagParsing` wrappers.

## Scope

Allowed writes:

- Root Cobra tree/router wrappers and Cobra/pflag error mapping under `internal/cli/**`.
- Focused CLI tests under `internal/cli/**`.
- `go.mod` / `go.sum` only for `github.com/spf13/cobra v1.10.2` and expected transitives `github.com/spf13/pflag` / `github.com/inconshreveable/mousetrap`.
- Issue-local GSD artifacts under `.planning/phases/400-cobra-router-shell/`.
- Minimal CLI docs/golden fixture changes only if required; default target is byte-identical golden suite.

Forbidden / out of scope:

- No shared parent orchestration artifacts, parent PR body, `.planning/PROJECT.md`, `.planning/ROADMAP.md`, `.planning/STATE.md`, or `.planning/traces/cli-architecture-v2-orchestration-state.yaml` edits.
- No later namespace nativization, Viper/config, TUI, event bus, telemetry, completion, help-tree churn, or connector bundle work.
- No credentialed connector checks, secrets, generic shell/HTTP/SQL write surfaces, reverse ETL execution, quality-gate reductions, or merges to `main`.

## Dependency decision

Selected exact approved dependency: `github.com/spf13/cobra v1.10.2`.

Rationale: ADR 0002 explicitly approves Cobra v1.10.x plus expected pflag/mousetrap transitives for Phase 2; `go list -m -versions github.com/spf13/cobra` shows `v1.10.2` as the latest v1.10.x patch. Stop if `go get` or `go mod tidy` introduces any additional direct dependency or dependency outside ADR-expected Cobra transitives.

## Required skills / references loaded

Skills loaded and applied:

- `gsd-core` — repo-local GSD/Pi adapter workflow.
- `caveman` — compact handoff prose only.
- `golang-how-to` — always-on Go skill routing; CLI task maps to Cobra + CLI + testing skills.
- `golang-cli` — command routing, flags, exit codes, stdout/stderr discipline.
- `golang-testing` — table-driven/red-green tests and golden fixtures.
- `golang-error-handling` — stable error taxonomy and single `writeError` authority.
- `golang-design-patterns` — minimal strangler wrapper, explicit constructors, dependency restraint.
- `golang-safety` — fresh command tree per run, nil/zero-value and state-leak avoidance.
- `golang-spf13-cobra` — `RunE`, `SilenceErrors/SilenceUsage`, `DisableFlagParsing`, hidden commands, fresh trees.
- `golang-security` — untrusted argv / dynamic connector passthrough / no generic write surfaces.
- `golang-documentation` — loaded proactively for CLI help/manual parity review; docs changes expected not applicable unless goldens/docs drift.

Rule references to cite in PR/handoff:

- `.agents/agentic-delivery/references/required-skills-routing.md`: **Always-on Go skill routing** and **CLI and command behavior** require `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, and `golang-security`; Cobra code additionally requires `golang-spf13-cobra`; docs changes require `golang-documentation`.
- `golang-cli`: root command `SilenceUsage`/`SilenceErrors`; stdout/stderr discipline; command tests through injected writers.
- `golang-spf13-cobra`: Best Practices #1 (`RunE`), #4 (`cmd.OutOrStdout`/`ErrOrStderr`), #5 (fresh command tree per test/run).
- `golang-testing`: Best Practices #1 named subtests, #3 independent tests, #5 observable behavior.
- `golang-error-handling`: Best Practices #5 `errors.Is`/`errors.As`, #7 errors logged or returned not both.
- `golang-design-patterns`: Best Practices #3 avoid hidden `init`, #20 dependency restraint, #21 design for testability.
- `golang-safety`: Best Practices #10 useful zero values, #11 `sync.Once` only when needed; avoid retained mutable Cobra state.
- `golang-security`: Security Thinking Model questions #1-#3 for untrusted argv and passthrough flags; Common Mistakes avoid command-injection/generic shell surfaces.

Note: `.pi/skills/go-implementation/SKILL.md` was required by the worker policy but does not exist in this checkout (`ENOENT`); only `.pi/skills/gsd-core/SKILL.md` is present. Recorded as a missing optional repo-local skill artifact, not a blocker because user-required Go skills above were loaded.

## Slice plan

### Slice 0 — Plan/TDD setup

1. Confirm branch `refactor/400-cobra-router-shell` is based on `origin/feat/cli-architecture-v2` at `379cb5015335ff7c9b20e5bb780952ead22c53b2`.
2. Create issue-local `PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, `SUMMARY.md`, `PROMPTS.md`, and `RUN-STATE.json` before production edits.
3. Record GSD command output and `programming-loop` fallback.
4. Commit/push the planning checkpoint if gates are clean and useful.

### Slice 1 — Red validation/test evidence

1. Run baseline red validation: `go list -deps ./internal/cli | grep '^github.com/spf13/cobra$'` should fail before dependency/router implementation.
2. Add a focused internal CLI test for `newRootCmd` / Cobra shell invariants: fresh command tree, root `DisableFlagParsing`, `SilenceErrors`, `SilenceUsage`, `help`/`man` commands, and hidden `extract`/`worker` commands.
3. Run `go test ./internal/cli/ -run TestCobraRouterShell -count=1`; expected red compile failure before production router exists.
4. Record exact output in `TDD-LEDGER.md`.

### Slice 2 — Minimal Cobra router shell

1. Add `github.com/spf13/cobra v1.10.2` with expected pflag/mousetrap transitives only.
2. Introduce a fresh-per-`Run` root command builder.
3. Rework `Run(args, stdout, stderr) int` to:
   - call existing `parseGlobal` first to preserve global `--root` / `--json` behavior anywhere in argv;
   - construct a fresh Cobra tree per invocation;
   - set `SetArgs(cleanArgs)`, `SetOut(stdout)`, `SetErr(stderr)`;
   - execute Cobra with `SilenceErrors/SilenceUsage`;
   - map errors through `mapCobraErr` then `writeError`.
4. Register `DisableFlagParsing` wrappers for all existing top-level commands; wrappers call the same legacy handlers and pass untouched `args` beneath them.
5. Keep `extract` and `worker` hidden from Cobra help while preserving existing hidden-command golden behavior.
6. Root fallback calls `runMaybeConnectorCommand` to preserve dynamic `pm <connector> <path…>` arbitrary flag passthrough.
7. Custom help/usage functions render existing `docs` map / root manual; no Cobra boilerplate or ANSI.

### Slice 3 — Green and refactor checks

1. Run focused red-green gates:
   - `go test ./internal/cli/ -run TestCobraRouterShell -count=1`
   - `go test ./internal/cli/ -run Golden -count=1`
   - `go test ./internal/cli/ -run Certify -count=1`
   - `go test ./internal/cli/ -count=1`
2. Refactor only while goldens stay byte-identical.
3. Ensure `writeError` remains sole exit-code authority; no direct `os.Exit`, no direct stderr diagnostics outside existing handlers.
4. Record go.mod/go.sum delta exactly.

### Slice 4 — Verification, parity, PR

1. Run required full gates:
   - `gofmt -w cmd internal`
   - `go test ./internal/cli/ -run Golden -count=1`
   - `go test ./internal/cli/ -run Certify -count=1`
   - `go test ./internal/cli/ -count=1`
   - `go vet ./...`
   - `go test ./...`
   - `go build ./cmd/pm`
   - `make verify`
   - `git diff --check origin/feat/cli-architecture-v2...HEAD`
   - `git diff origin/feat/cli-architecture-v2...HEAD -- go.mod go.sum`
2. CLI parity checks without credentials:
   - Build local binary (`go build -o /tmp/pm-400 ./cmd/pm`) before running unfamiliar `pm` commands.
   - Use `pm help <topic>` before unfamiliar commands: `/tmp/pm-400 help connectors`.
   - Bare namespace: `/tmp/pm-400 connectors` exits 0 and prints existing manual.
   - Command help: `/tmp/pm-400 docs --help` and `/tmp/pm-400 worker --help` preserve current golden behavior.
   - Docs/website/generated artifacts: mark no source update required if golden suite and docs-generate-diff remain byte-identical and no CLI help text/command/flag surface changes.
3. Update artifacts with real results.
4. Commit/push coherent green slice(s) to `origin/refactor/400-cobra-router-shell`.
5. Open non-draft stacked sub-PR to `feat/cli-architecture-v2` with Conventional Commit title, `Refs #400`, `Refs #397`, GSD/TDD/skills/dependency/parity/verification evidence.
6. Review route: Claude workflow already `disabled_manually`; Copilot quota exhausted. Do not post `@claude review`; do not request Copilot again. Record human/parent-PR fallback pending.

## Review-fix cycle — PR #440 pm-reviewer dispositions

Accepted findings:

1. `mapCobraErr` must not reclassify legacy/root-fallback plain errors that merely contain Cobra-like text (`unknown flag`, `unknown command`). Keep `writeError` taxonomy unchanged by marking legacy handler/root-fallback errors before they reach `mapCobraErr`, then bypassing them there. Add regression tests for legacy bypass and genuine Cobra parse-error usage mapping.
2. Fresh root commands must define persistent `--root` and `--json` flags per ADR 0002 while `parseGlobal` remains the semantic owner and root/wrappers keep `DisableFlagParsing`. Flag defaults must reflect already-parsed invocation state and not share state across fresh command trees.

Residual scope to close:

- Assert every registered top-level wrapper, including `init`, has `DisableFlagParsing` and expected visibility (`extract`/`worker` hidden only).
- Add deterministic dynamic connector passthrough coverage for arbitrary connector flags with late global flags, using local `httptest` and no credentials beyond test-local fixture credentials.
- Exercise actual Cobra error mapping and legacy bypass without changing transcript behavior.

Review-fix TDD plan:

1. Add focused failing tests in `internal/cli/cobra_router_test.go` for persistent root flags, wrapper visibility/flag parsing, legacy handler error bypass, genuine Cobra parse error mapping, and dynamic connector passthrough with late globals.
2. Capture red output with `go test ./internal/cli/ -run TestCobraRouterShell -count=1` before production edits.
3. Implement the smallest router changes: root persistent flag definitions plus legacy error marking/bypass.
4. Run the exact gates requested by the review-fix task, update `TDD-LEDGER.md`, `VERIFICATION.md`, `SUMMARY.md`, and `RUN-STATE.json`, then commit/push.
5. Update PR #440 body with accepted disposition summary and corrected verification. Do not post `@claude review`; do not request Copilot again.

## Spawn decision for this cycle

`local_critical_path`: review-fix work is handled inline by the assigned issue worker in the existing isolated worktree; no subagents are available to this worker.

## Human gates

- No secrets, no credential prompts, no credentialed connector checks.
- No dependency deviation beyond Cobra v1.10.2 and expected pflag/mousetrap transitives.
- No generic shell/HTTP/SQL write tools.
- No reverse ETL execution outside existing test/smoke gates.
- No quality gate reductions.
- No merge to `main` or parent PR merge.
