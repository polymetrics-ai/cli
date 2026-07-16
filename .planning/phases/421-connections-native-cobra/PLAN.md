# Phase 421 Plan — Connections native Cobra namespace

Issue: polymetrics-ai/cli#421
Umbrella: #407
Parent: #397 / PR #438
Branch: `refactor/421-connections-native-cobra`
Base branch: `feat/cli-architecture-v2`
Base parent head at dispatch: `1678f9ab`
Execution decision: `local_critical_path` — first serialized Phase 9 namespace worker; cwd/branch already isolated; worker has no subagent tool and must not delegate.

## Required reading complete

- Issue #421, umbrella #407, parent #397 bodies, and parent PR #438 context.
- `AGENTS.md`; issue-agent, parent-orchestrator, stacked PR, automated-review, Claude-review, and worker-handoff contracts/workflows.
- GSD universal runtime loop; `.planning/config.json`, `PROJECT.md`, `ROADMAP.md`, `STATE.md`, universal programming loop PRD/prompts.
- Required-skill routing, GSD Pi adapter, CLI help/docs/website parity.
- CLI Architecture v2 plan §5/§9, execution prompt Stage 9, ADR 0002.
- Current connections parser/tests/docs/goldens: `internal/cli/cobra_router.go`, `internal/cli/cli.go`, `internal/cli/parse.go`, `internal/cli/docs.go`, `internal/cli/*_test.go`, `internal/cli/testdata/golden_transcripts.json`, `docs/cli/connections.md`, `website/content/docs/cli-reference.mdx`, and website grep hits for `connections`.
- Phase 406 native catalog artifacts used as the nearest Stage 8/9 template.

## GSD adapter

- `scripts/gsd doctor` — pass.
- `scripts/gsd prompt plan-phase 421 --skip-research >/tmp/gsd-plan-phase-421.prompt` — pass.
- `scripts/gsd prompt programming-loop init --phase 421 --dry-run >/tmp/gsd-programming-loop-421.prompt` — blocked: `scripts/gsd: unknown GSD command: programming-loop`; manual GSD fallback active using `.pi/prompts/pm-gsd-loop.md`, `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`, and the issue contract.
- Adapter/skill gap: `.pi/skills/go-implementation/SKILL.md` was required by agent instructions but is missing in this checkout; global Go skills listed below are loaded and recorded.

## Required skills loaded

- GSD/status: `gsd-core`, `caveman`.
- Go/CLI: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-documentation`, `golang-spf13-cobra`, `golang-security`, `golang-safety`.
- Skill rule anchors for handoff: go-how-to CLI routing table; CLI exit-code/stdout-stderr/testing rules; testing best-practices #1, #3, #5; error-handling #1, #2, #7, #9; documentation writing principles and application CLI help; cobra best practices #1, #3, #4, #5 plus StringArray/NoOptDefVal/unknown-flag guidance; security trust-boundary questions #1-#3/no secrets; safety #2 and #10.

## Scope / exclusions

Allowed:

- Top-level `pm connections` Cobra node, native `create`/`list` declared flags, and minimal handler adaptation under `internal/cli/**`.
- Connections-focused tests and unchanged golden/docs parity checks.
- Directly applicable connections help/manual/website/generated artifacts only if output intentionally changes.
- Issue-local `.planning/phases/421-*` artifacts.

Excluded:

- Other namespace migrations, connector dynamic dispatch, connector bundles, app/domain behavior, events/logging/telemetry/runtime/RLM/worker/flow/schedule/ETL/reverse behavior, parent state/roadmap/PR body, go.mod/go.sum, and other worker branches.
- Completion implementation beyond preserving a connection-name completion compatibility seam for later Phase 15.
- Credentials/secrets, credentialed connector checks, external writes beyond local test fixtures, dependency changes, and `main` merge.

## Current behavior notes before production edits

- `connections` is still a legacy Cobra wrapper with `DisableFlagParsing: true` and no native subcommands or pflag declarations.
- `runConnections` parses `create|list` using `parseFlags`.
- Legacy `create` behavior: missing action/name returns usage; `--source`, `--destination`, and endpoint validation are handled in order; repeated singleton flags use last value; repeated `--primary-key` accumulates; bare `--flag` becomes `true`; `--flag value` and `--flag=value` both work; unknown flags are ignored.
- Legacy `list` behavior: ignores unknown flags and extra positional args, then lists sorted connections.
- Bare `pm connections` already renders contextual help exit 0 through wrapper help interception.
- Invalid action currently returns usage; native Cobra must keep invalid action as usage error with mapped exit 2 while avoiding production app/domain side effects before action recognition.
- Help/docs currently come from canonical `docs` map and generated `docs/cli/connections.md`; website CLI reference mirrors examples. No intentional help text change planned.

## Delivered implementation matrix

| Scope | Planned delivery |
|---|---|
| Native connections Cobra node | Remove `connections` from legacy wrapper list and register native command with custom docs-map help/usage. |
| Actions | Add native `create` and `list` subcommands. |
| Flag parity | Use pflag declarations with `StringArrayVar` where repetition matters or last-wins must be recoverable; set `NoOptDefVal="true"`; whitelist unknown flags; normalize optional-value flags so legacy `--flag value` keeps working. |
| Handler adaptation | Split `runConnectionsAction`/request construction so native actions reuse existing app behavior without changing domain semantics. |
| Bare/invalid behavior | Bare namespace help exit 0; invalid action usage exit 2 through `mapCobraErr`. |
| Completion seam | Provide a package-level connection-name completion seam returning no completions/NoFileComp until Phase 15 wires project-aware completion. |
| Docs/goldens | Expect no golden/docs/website/generated changes; verify docs-generate-diff/goldens and runtime help. |

## Slice plan

1. Planning checkpoint ✅
   - Create phase artifacts and record adapter fallback, missing repo Go skill file, loaded skills, scope, parity checklist, and verification plan.

2. Red tests
   - Add focused tests proving `connections` is not yet native: command should have `DisableFlagParsing=false`; native `create`/`list` subcommands; `create` flag metadata for repeated/listable and singleton flags; `NoOptDefVal="true"`; unknown-flag whitelist; list unknown-flag tolerance; completion seam exists.
   - Add behavior tests for create flag forms: equals, space, repeated last-wins, repeated primary key accumulation, bare bool preservation, unknown flag tolerance, extra args tolerance, and late global `--root`/`--json` via existing global parser.
   - Add invalid-action usage test if not already covered by goldens.
   - Capture exact red output in `TDD-LEDGER.md` before production code.

3. Green implementation
   - Add native `connections` subtree in `cobra_router.go`, with action subcommands and custom docs-map help/usage.
   - Add normalization for `connections create` optional-value flags so pflag `NoOptDefVal` does not swallow legacy space-form values.
   - Add `runConnectionsCreate` / `runConnectionsList` helper(s) that accept already-parsed values while preserving current validation/order/output behavior.
   - Remove `connections` namespace `parseFlags` call site when native path no longer needs it.

4. Parity / golden check
   - Golden transcript changes expected: empty.
   - Docs/website updates expected: not applicable unless help output intentionally changes; generated docs check should remain green.
   - Verify runtime help: `pm help connections`, bare `pm connections`, `pm connections --help`, JSON manual, and representative `connections create/list --json` local fixture commands.

5. Full verification / PR
   - Run required focused and full gates.
   - Diff check against `origin/feat/cli-architecture-v2...HEAD`; `go.mod`/`go.sum` diff must be empty.
   - Commit/push green slice; open non-draft stacked PR to `feat/cli-architecture-v2` with `Refs #421`, `Refs #407`, and `Refs #397`.
   - Do not request Claude/Copilot; record human/parent fallback pending per dispatch instruction.

## Planned tests / validations

- `gofmt -w cmd internal`
- `go test ./internal/cli/... -run 'Connections|CobraRouterShell|Golden' -count=1`
- `go test ./internal/cli/ -run Certify -count=1`
- `go vet ./...`
- `go test ./...`
- `go build ./cmd/pm`
- `make verify`
- `git diff --check origin/feat/cli-architecture-v2...HEAD`
- `git diff -- go.mod go.sum`
- Runtime help parity after build: `./pm help connections`, `./pm connections`, `./pm connections --help`, `./pm connections --json`, invalid action JSON usage error, and docs/website grep/generator checks.

## Parity stance

This phase changes parser ownership only. Help text, docs, website, generated manuals, golden stdout/stderr/exit, JSON envelopes, stdout/stderr discipline, global late flags, fresh-tree re-entrancy, and app/domain behavior should remain byte-identical unless an intentional reviewed change is recorded. No intentional user-facing output change is planned.
