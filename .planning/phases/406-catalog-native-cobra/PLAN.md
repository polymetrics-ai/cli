# Phase 406 Plan — Catalog native Cobra namespace

Issue: polymetrics-ai/cli#406
Parent: #397 / PR #438
Branch: `refactor/406-catalog-native-cobra`
Base branch: `feat/cli-architecture-v2`
Base parent head at dispatch: `e5ee4075`
Execution decision: `local_critical_path` — bounded mutating worker already isolated in cwd/branch; worker has no subagent tool and must not delegate.

## Required reading complete

- Issue #406 and parent #397 bodies, parent PR #438 context.
- `AGENTS.md`; issue-agent, stacked PR, automated-review, Claude-review, worker-handoff contracts/workflows.
- GSD universal runtime loop; `.planning/config.json`, `PROJECT.md`, `ROADMAP.md`, `STATE.md`, universal programming loop PRD/prompts.
- Required-skill routing, GSD Pi adapter, CLI help/docs/website parity.
- CLI Architecture v2 plan Stage/Phase 8, execution prompt Stage 8, ADR 0002.
- Current router/catalog/docs/goldens: `internal/cli/cobra_router.go`, `internal/cli/cli.go`, `internal/cli/parse.go`, `internal/cli/docs.go`, `internal/cli/*catalog*test.go`, `internal/cli/golden_transcript_test.go`, `internal/cli/testdata/golden_transcripts.json`, `docs/cli/catalog.md`, website CLI-reference grep.

## GSD adapter

- `scripts/gsd doctor` — pass.
- `scripts/gsd prompt plan-phase 406 --skip-research >/tmp/gsd-plan-phase-406.prompt` — pass.
- `scripts/gsd prompt programming-loop init --phase 406 --dry-run >/tmp/gsd-programming-loop-406.prompt` — blocked: `scripts/gsd: unknown GSD command: programming-loop`; manual GSD fallback active using `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md` and issue contract.
- Adapter gap recorded because `.gsd/commands.json` exposes `gsd-plan-phase`, `gsd-verify-work`, and `gsd-code-review`, but not `programming-loop`.

## Required skills loaded

- GSD/status: `gsd-core`, `caveman`.
- Go/CLI: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-documentation`, `golang-spf13-cobra`, `golang-security`, `golang-safety`.
- Skill rule anchors for handoff: go-how-to CLI routing table; CLI exit-code/stdout-stderr/testing rules; testing best-practices #1, #3, #5; error-handling #1, #2, #7, #9; docs principles + application CLI help; cobra best practices #1, #3, #4, #5 and flags StringArray guidance; security trust-boundary questions #1-#3/no secrets; safety #2 and #10.

## Scope / exclusions

Allowed:

- Top-level `pm catalog` Cobra node/flags/help/usage under `internal/cli/**`.
- Catalog handler adaptation and catalog-focused tests/golden checks.
- Catalog-only docs/website/generated data only if output intentionally changes.
- Issue-local `.planning/phases/406-*` artifacts.

Excluded:

- app/flow/certify/worker/events/logging/telemetry/connsdk, config semantics, other namespaces, connector catalog (`pm connectors catalog`), go.mod/go.sum, parent orchestration state, parent PR body, parent roadmap/state, and other workers' branches.
- Credentials, credentialed connector checks, runtime services, external writes, reverse ETL execution outside existing local smoke gates, dependency changes, and `main` merge.

## Current behavior notes before edits

- `catalog` is still a legacy Cobra wrapper with `DisableFlagParsing: true` and no native subcommands/flags.
- Legacy `runCatalog` parses `refresh|show` with `parseFlags`, repeated `--connection` last-wins, bare `--connection` becomes `true`, unknown flags inside actions are ignored, and action-specific extra positional args are ignored.
- Bare `pm catalog` is already contextual help exit 0 through wrapper help interception.
- Invalid action without `--connection` currently reports missing connection before usage; acceptance requires invalid action usage error.
- Help and docs currently come from canonical `docs` map and generated `docs/cli/catalog.md`; golden catalog help fixtures exist and should not change.

## Delivered implementation matrix

| Scope | Delivery |
|---|---|
| Native catalog Cobra node | `catalog` removed from legacy wrapper list and registered as a native Cobra command with custom docs-map help/usage. |
| Catalog actions | Added native `refresh` and `show` subcommands with StringArray `--connection`, `NoOptDefVal="true"`, unknown-flag whitelist, arbitrary extra-arg tolerance, and last-wins adaptation. |
| Legacy flag parity | Added catalog-specific pre-normalization so `--connection value` keeps legacy space-form behavior while `NoOptDefVal` preserves bare `--connection` -> `true`. |
| Handler adaptation | Split `runCatalogAction` from `runCatalog` so native actions reuse existing refresh/show output paths without touching app semantics. |
| Tests | Added native subtree/flag metadata tests, invalid-action usage classification test, and connection flag-form/unknown-flag behavior tests. |
| Docs/goldens | No help/docs/website/golden fixture changes; parity verified by existing golden docs test and runtime help checks. |

## Slice plan

1. Planning checkpoint ✅
   - Create phase artifacts and record adapter fallback, skills, scope, parity checklist, and verification plan.

2. Red tests
   - Add focused tests proving `catalog` is not yet native: catalog command should have `DisableFlagParsing=false`, native `refresh`/`show` subcommands, `--connection` as StringArray with `NoOptDefVal="true"`, and action subcommands should whitelist unknown flags.
   - Add behavior test for `pm catalog bogus --json` expecting usage exit 2 before `--connection` validation.
   - Capture exact red output in `TDD-LEDGER.md`.

3. Green implementation
   - Replace only the catalog legacy wrapper with a native Cobra subtree in `cobra_router.go`.
   - Add native `refresh` and `show` action commands with StringArray `--connection`, last-wins adaptation, `NoOptDefVal="true"`, and `FParseErrWhitelist{UnknownFlags:true}`.
   - Keep custom help/usage wired to the canonical catalog docs map; bare `pm catalog` writes contextual help exit 0.
   - Refactor `runCatalog` minimally so legacy semantics are shared by native action handler where useful.

4. Parity / golden check
   - Golden transcript changes expected: empty.
   - Docs/website updates expected: not applicable unless help output intentionally changes; docs-generate-diff test should remain green.
   - Verify runtime help: `pm help catalog`, bare `pm catalog`, `pm catalog --help`, JSON manual.

5. Full verification / PR
   - Run required focused and full gates.
   - Diff check against `origin/feat/cli-architecture-v2...HEAD`; `go.mod`/`go.sum` diff must be empty.
   - Commit/push green slice; open non-draft stacked PR to `feat/cli-architecture-v2` with `Refs #406` and `Refs #397`.
   - Do not request Claude/Copilot; record human/parent fallback pending per dispatch instruction.

## Planned tests / validations

- `go test ./internal/cli/ -run 'Catalog|CobraRouterShell|Golden' -count=1`
- `go test ./internal/cli/ -run Certify -count=1`
- `gofmt -w cmd internal`
- `go vet ./...`
- `go test ./...`
- `go build ./cmd/pm`
- `make verify`
- `git diff --check origin/feat/cli-architecture-v2...HEAD`
- `git diff -- go.mod go.sum`
- Runtime help parity after build: `./pm help catalog`, `./pm catalog`, `./pm catalog --help`, `./pm catalog --json`, plus docs/website grep/generator checks as applicable.

## Parity stance

This phase changes parser ownership only. Help text, docs, website, generated manuals, JSON envelopes, stdout/stderr discipline, global `--root`/`--json` late placement, fresh tree re-entrancy, and connector-catalog behavior should remain byte-identical. Invalid `pm catalog <bad-action>` usage classification may change to satisfy issue acceptance; no golden currently covers that path.
