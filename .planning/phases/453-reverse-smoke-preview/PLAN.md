# Phase 453 Plan — Reverse smoke preview safety

Issue: polymetrics-ai/cli#453
Parent: #397 / PR #438
Branch: `fix/453-reverse-smoke-preview`
Base branch: `feat/cli-architecture-v2`
Base parent at dispatch: `5680debb`
Execution decision: `local_critical_path` — isolated worker cwd/branch; no subagent tool; one issue/scope.

## Required reading complete

- Issue #453 body and acceptance criteria; parent #397 body and human gates.
- `AGENTS.md`; issue-agent, stacked parent/subissue, parent-orchestrator, automated-review, Claude-review, and worker-handoff contracts/workflows.
- GSD universal runtime loop; `.planning/config.json`, `PROJECT.md`, `ROADMAP.md`, `STATE.md`, universal programming loop PRD/prompts.
- GSD Pi adapter, required skill routing, CLI help/docs/website parity reference.
- Reverse safety docs/help/tests: `docs/skills/pm-reverse-etl/SKILL.md`, `docs/skills/recipe-preview-approve-reverse-etl/SKILL.md`, `docs/cli/reverse.md`, `docs/GUIDE.md`, `internal/app/reverse_confirmation_test.go`, `internal/app/reverse_approval_test.go`, `internal/cli/reverse_cli_test.go`.
- `Makefile` `smoke-no-build` recipe.

## GSD adapter

- `scripts/gsd doctor` — pass.
- `scripts/gsd list` — pass; 69 commands listed.
- `scripts/gsd prompt plan-phase 453 --skip-research >/tmp/gsd-plan-phase-453.prompt` — pass; prompt written.
- `scripts/gsd prompt programming-loop init --phase 453 --dry-run >/tmp/gsd-programming-loop-453.prompt` — blocked: `scripts/gsd: unknown GSD command: programming-loop`; manual GSD fallback active using `.pi/prompts/pm-gsd-loop.md`, `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`, and issue contract.
- Repo skill gap: `.pi/skills/go-implementation/SKILL.md` required by worker instructions but missing (`ENOENT`); global Go skills loaded and recorded.

## Required skills loaded

- GSD/status: `gsd-core`, `caveman`.
- Go/test/CLI/safety: `golang-how-to`, `golang-testing`, `golang-cli`, `golang-security`, `golang-safety`, `golang-error-handling`.
- Skill rule anchors for handoff: required-skill routing Go always-on and CLI command behavior; `golang-how-to` testing/CLI/security routing table; `golang-testing` best practices #1, #3, #5; `golang-cli` stdout/stderr, machine-readable output, CLI testing; `golang-security` trust-boundary questions #1-#3 and no secrets; `golang-safety` resource and nil/panic safety; `golang-error-handling` #1/#2/#7/#9.

## Scope / exclusions

Allowed:

- `Makefile` `smoke-no-build` reverse smoke ordering only.
- A durable minimal regression check that statically enforces plan ID extraction → preview → approval extraction → run ordering.
- Issue-local `.planning/phases/453-*` artifacts.

Excluded:

- CLI product behavior, reverse app logic, connectors, docs/website, generated manuals, dependencies, parent ledger/PR body, external services/writes, credentialed checks, runtime-backed checks, and `main` merge.

## Safety interlock

Do not run `make smoke` or `make verify` until both are true:

1. `smoke-no-build` contains `./pm reverse preview "$$PLAN_ID" --root "$$SMOKE_DIR" --json >/dev/null;` between plan ID extraction and `reverse run`.
2. Focused ordering regression is green.

After green, only run existing temporary local smoke; no external connector/runtime/credentialed checks.

## Delivered implementation matrix

| Scope | Delivery |
|---|---|
| Red regression | Added `internal/safety/smoke_makefile_test.go`, a static Makefile ordering test for `smoke-no-build`. |
| Smoke ordering | Inserted `./pm reverse preview "$$PLAN_ID" --root "$$SMOKE_DIR" --json >/dev/null;` after plan ID extraction and before approval extraction / run. |
| Token safety | Kept approval token parsing from existing `PLAN_OUTPUT`; no new token output or storage. |
| Local write scope | Smoke still writes only to the existing temp local warehouse/outbox fixture. |
| CLI/docs parity | N/A; no CLI behavior/help/docs/website/generated artifacts changed. |

## Slice plan

1. Planning checkpoint ✅
   - Created issue-local PLAN/TDD-LEDGER/VERIFICATION/SUMMARY/RUN-STATE/PROMPTS before production edits.
   - Recorded GSD adapter fallback and loaded skills.

2. Red regression ✅
   - Added a focused Go test under `internal/safety` that reads `Makefile` and asserts `smoke-no-build` orders `reverse plan` output extraction, `reverse preview`, approval extraction, and `reverse run`.
   - Ran `go test ./internal/safety -run TestSmokeNoBuildReversePlanPreviewRunOrdering -count=1` and captured exact failure before Makefile edit.

3. Minimal green implementation ✅
   - Inserted `./pm reverse preview "$$PLAN_ID" --root "$$SMOKE_DIR" --json >/dev/null;` after `PLAN_ID=` and before `APPROVAL=` / `reverse run`.
   - Kept temp root/outbox flow unchanged.

4. Verification / PR ✅
   - Ran focused ordering test before smoke.
   - Ran required gates after fix: `gofmt -w cmd internal`, `go test ./...`, `go vet ./...`, `go build ./cmd/pm`, `make smoke`, `make verify`, `git diff --check origin/feat/cli-architecture-v2...HEAD`, `git diff -- go.mod go.sum`.
   - `make verify` redundantly covered gofmt/tidy/vet/test/build/docs/smoke/lint/connectorgen; exact coverage recorded.
   - Opened non-draft stacked PR #454 to `feat/cli-architecture-v2` with `Refs #453` and `Refs #397`.

## Parity stance

CLI help/docs/website parity is N/A: no command, flag, help text, runtime output contract, connector surface, docs, website, or generated manual behavior changes. This issue only changes repository smoke ordering and a regression test.
