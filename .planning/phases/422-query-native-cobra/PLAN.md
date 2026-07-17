# Phase 422 Plan — Query native Cobra namespace

Issue: polymetrics-ai/cli#422
Umbrella: #407
Parent: #397 / PR #438
Branch: `refactor/422-query-native-cobra`
Base branch: `feat/cli-architecture-v2`
Base parent head at dispatch: `e6faecfb` (#421 squash); rebased to latest parent ledger checkpoint `f12d573b` before edits.
Execution decision: `local_critical_path` — second serialized Phase 9 namespace worker; cwd/branch isolated; worker has no subagent tool and must not delegate.

## Required reading complete

- Issue #422, umbrella #407, parent #397 bodies, and parent PR #438 context.
- `AGENTS.md`; issue-agent, stacked parent/subissue, automated-review, Claude-review, worker-handoff contracts/workflows.
- GSD universal runtime loop; `.planning/config.json`, `PROJECT.md`, `ROADMAP.md`, `STATE.md`, universal programming loop PRD/prompts.
- Required-skill routing, GSD Pi adapter, CLI help/docs/website parity.
- CLI Architecture v2 plan §5/§9, execution prompt Stage 9, ADR 0002.
- Current query parser/handlers/tests/docs/goldens and SQL safety guards: `internal/cli/cobra_router.go`, `internal/cli/cli.go`, `internal/cli/parse.go`, `internal/cli/docs.go`, `internal/cli/agentmode_query_cli_test.go`, `internal/cli/validation_cli_test.go`, `internal/cli/golden_transcript_test.go`, `internal/cli/testdata/golden_transcripts.json`, `internal/app/query_engine_default.go`, `internal/app/query_engine_duckdb.go`, `internal/app/util.go`, `docs/cli/query.md`, `website/content/docs/query.mdx`, `website/lib/docs.generated.ts`.
- Phase 406 catalog and Phase 421 connections native-Cobra artifacts used as Stage 8/9 templates.

## GSD adapter

- `scripts/gsd doctor` — pass.
- `scripts/gsd prompt plan-phase 422 --skip-research >/tmp/gsd-plan-phase-422.prompt` — pass.
- `scripts/gsd prompt programming-loop init --phase 422 --dry-run >/tmp/gsd-programming-loop-422.prompt` — blocked: `scripts/gsd: unknown GSD command: programming-loop`; manual GSD fallback active using `.pi/prompts/pm-gsd-loop.md`, `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`, and the issue contract.
- Adapter/skill gap: `.pi/skills/go-implementation/SKILL.md` is required by worker instructions but missing in this checkout (`ENOENT`); global Go skills listed below are loaded and recorded.

## Required skills loaded

- GSD/status: `gsd-core`, `caveman`.
- Go/CLI/database: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-documentation`, `golang-spf13-cobra`, `golang-security`, `golang-safety`, `golang-database`.
- Skill rule anchors for handoff: go-how-to CLI/database routing table; CLI exit-code/stdout-stderr/testing rules; testing best-practices #1, #3, #5; error-handling #1, #2, #7, #9; documentation writing principles and application CLI help; cobra best practices #1, #3, #4, #5 plus StringArray/NoOptDefVal/unknown-flag guidance; security trust-boundary questions #1-#3/no secrets/SQL injection; safety #2 and #10; database #2 parameterized/read-only SQL safety, #3 context, #5 rows closed, #14 no schema writes.

## Scope / exclusions

Allowed:

- Top-level `pm query` Cobra node, native `run` action, declared query flags, and minimal handler adaptation under `internal/cli/**`.
- Query-focused tests and unchanged golden/docs parity checks.
- Directly applicable query help/manual/website/generated artifacts only if output intentionally changes.
- Issue-local `.planning/phases/422-*` artifacts.

Excluded:

- Other namespace migrations, connector dynamic dispatch, connector bundles, app/query-engine behavior, SQL grammar changes, events/logging/telemetry/runtime/RLM/worker/flow/schedule/ETL/reverse behavior, parent state/roadmap/PR body, go.mod/go.sum, and other worker branches.
- Completion implementation beyond preserving declared flag metadata / no-file fallback seams for later Phase 15.
- Credentials/secrets, credentialed connector checks, external services, destructive SQL, dependency changes, and `main` merge.

## Current behavior notes before production edits

- `query` is still a legacy Cobra wrapper with `DisableFlagParsing: true` and no native `run` subcommand or pflag declarations.
- `runQuery` requires `args[0] == "run"`, then parses `run` flags using `parseFlags(args[1:])`.
- Legacy flag behavior: `--table`, `--sql`, `--limit`, `--fields`, `--agent-mode`, and `--sample` support `--flag value`, `--flag=value`, repeated last-wins for scalar flags, repeated `--fields` accumulation with comma splitting, bare `--flag` becomes `"true"`, and unknown flags are ignored.
- Legacy extra positional args after `run` are ignored by `parseFlags`.
- Bare `pm query` already renders contextual help exit 0 through wrapper help interception.
- Invalid action currently returns usage; native Cobra must keep invalid action as usage error with mapped exit 2 while avoiding `.polymetrics` project open before action recognition.
- `--sql` routes to `App.QuerySQL`; default JSONL engine only supports `SELECT * FROM <table> [LIMIT n]`, DuckDB-tagged engine has strict read-only `validateSelectOnly` guards. This phase must not loosen either guard or expose generic SQL write.
- Help/docs currently come from canonical `docs` map and checked-in `docs/cli/query.md`; website query docs already describe read-only SQL and agent-mode output. No intentional help text change planned.

## Delivered implementation matrix

| Scope | Delivery |
|---|---|
| Native query Cobra node | `query` removed from legacy wrapper list and registered as a native command with custom docs-map help/usage. |
| Action | Added native `run` subcommand. |
| Flag parity | Added pflag `StringArrayVar` declarations for `--table`, `--sql`, `--limit`, `--fields`, `--agent-mode`, and `--sample`; set `NoOptDefVal="true"`; preserved unknown-flag tolerance and optional-value normalization for legacy space-form values. |
| Handler adaptation | Replaced `runQuery`/query `parseFlags` path with `runQueryRun` receiving parsed native flag values while preserving validation order, repeated flag semantics, output envelopes, and app/query-engine behavior. |
| Bare/invalid behavior | Bare namespace help exits 0; invalid action exits 2 through `mapCobraErr` without opening `.polymetrics` first. |
| SQL safety | Read-only SQL guards remain in app query engines; no generic SQL write, no SQL grammar expansion, no app/query-engine behavior changes. |
| Completion metadata | Native flags are declared; `query run` has a no-file completion seam. Phase 15 completion implementation deferred. |
| Docs/goldens | No help/docs/website/golden fixture changes expected; focused golden test passed. |

## Slice plan

1. Planning checkpoint ✅
   - Created phase artifacts and recorded adapter fallback, missing repo Go skill file, loaded skills, scope, parity checklist, SQL safety stance, and verification plan.

2. Red tests ✅
   - Added focused tests proving `query` was not yet native: top-level command should have `DisableFlagParsing=false`, native `run` subcommand should exist, declared flags should have legacy-compatible metadata (`StringArray` + `NoOptDefVal="true"`), unknown-flag tolerance, and no-file completion fallback.
   - Added behavior tests for query flag forms: equals, space, repeated scalar last-wins, repeated/comma `--fields` accumulation, bare bool sentinel preservation, unknown flag tolerance, extra args tolerance, late global `--root`/`--json`, SQL last-wins, invalid action usage-before-project-open, and read-only SQL rejection.
   - Captured exact red output in `TDD-LEDGER.md` before production code.

3. Green implementation ✅
   - Added native `query` subtree in `cobra_router.go` with docs-map help/usage.
   - Added native `query run` declared flags and normalization for optional value flags so pflag `NoOptDefVal` does not swallow legacy space-form values.
   - Added a query handler adapter that accepts parsed flag values and preserves existing validation order, output envelopes, agent mode behavior, and SQL read-only guard routing.
   - Removed the `query` namespace legacy wrapper/`parseFlags` call site.

4. Parity / golden check ✅
   - Golden transcript changes: empty.
   - Docs/website updates: not applicable because help/output did not change.
   - Verified runtime help: `pm help query`, bare `pm query`, `pm query --help`, JSON manual, invalid action JSON usage error, representative query run JSON, and read-only rejection.
   - Verified generated docs/manual and website docs generator/diff.

5. Full verification / PR
   - Required focused and full gates passed.
   - `git diff --check origin/feat/cli-architecture-v2...HEAD` passed and `go.mod`/`go.sum` diff empty.
   - Commit/push checkpoints completed: planning, red tests, green implementation. PR still pending.
   - Do not request Claude/Copilot; record human/parent fallback pending per dispatch instruction.

## Planned tests / validations

- `gofmt -w cmd internal`
- `go test ./internal/cli/... -run 'Query|CobraRouterShell|Golden' -count=1`
- `go test ./internal/cli/ -run Certify -count=1`
- `go vet ./...`
- `go test ./...`
- `go build ./cmd/pm`
- `make verify`
- `git diff --check origin/feat/cli-architecture-v2...HEAD`
- `git diff -- go.mod go.sum`
- Runtime help/parity after build: `./pm help query`, `./pm query`, `./pm query --help`, `./pm query --json`, invalid action JSON usage error, query JSON local fixture, read-only SQL rejection, docs generator diff, docs validate, website docs generator.

## Parity stance

This phase changes parser ownership only. Help text, docs, website, generated manuals, golden stdout/stderr/exit, JSON envelopes, stdout/stderr discipline, global late flags, fresh-tree re-entrancy, completion metadata, app/query-engine behavior, and SQL read-only guards should remain byte-identical unless an intentional reviewed change is recorded. No intentional user-facing output change is planned.
