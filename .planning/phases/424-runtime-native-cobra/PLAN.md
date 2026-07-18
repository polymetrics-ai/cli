# Phase 424 Plan — Runtime native Cobra namespace

Issue: polymetrics-ai/cli#424
Umbrella: #407
Parent: #397 / parent PR #438
Branch: `refactor/424-runtime-native-cobra`
Base branch: `feat/cli-architecture-v2`
Base parent head at dispatch/planning: `56a7ecb08f755184af7b55318c3285582d5adfb7`
Execution decision: `local_critical_path` — fourth serialized Phase 9 namespace worker; cwd/branch isolated; worker has no subagent tool and must not delegate.

## Required reading complete

- Issue #424 body and acceptance criteria; umbrella #407 roster; parent #397 and draft parent PR #438 context.
- `AGENTS.md`; issue-agent, parent/subissue, automated-review, Claude-review, worker-handoff contracts/workflows.
- GSD universal runtime loop; `.planning/config.json`, `PROJECT.md`, `ROADMAP.md`, `STATE.md`, universal programming loop PRD/prompts.
- Required-skill routing, GSD Pi adapter, CLI help/docs/website parity, runtime/RLM/website integration reference.
- Runtime setup/source docs: `docs/architecture/runtime-dependencies.md`, `docs/runtime/SETUP.md`, `docs/cli/runtime.md`, `docs/cli/rlm.md`, `docs/cli/perf.md`, `docs/cli/agent.md`, `website/content/docs/architecture.mdx`, `website/content/docs/cli-reference.mdx`, `website/package.json`.
- CLI Architecture v2 plan §5/§9, execution prompt Stage 9, ADR 0002.
- Current runtime parser/handlers/tests/docs/goldens: `internal/cli/cobra_router.go`, `internal/cli/cli.go`, `internal/cli/runtime_helpers.go`, `internal/cli/docs.go`, `internal/cli/cli_test.go`, `internal/cli/config_migration_test.go`, `internal/cli/golden_transcript_test.go`, `docs/cli/runtime.md`, `website/content/docs/architecture.mdx`, `website/content/docs/cli-reference.mdx`, `internal/runtimecheck/**`.
- Phase 406 catalog, Phase 421 connections, Phase 422 query, and Phase 423 perf native-Cobra artifacts used as Stage 8/9 templates.

Current phase artifacts did not exist at kickoff; this checkpoint creates them before production edits.

## GSD adapter

- `scripts/gsd doctor` — pass.
- `scripts/gsd prompt plan-phase 424-runtime-native-cobra --skip-research >/tmp/gsd-plan-phase-424-runtime-native-cobra.prompt` — pass (10739 bytes).
- `scripts/gsd prompt programming-loop init --phase 424-runtime-native-cobra --dry-run >/tmp/gsd-programming-loop-424-runtime-native-cobra.prompt` — blocked: `scripts/gsd: unknown GSD command: programming-loop`; manual GSD fallback active using `.pi/prompts/pm-gsd-loop.md`, `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`, and the issue contract.
- Adapter/skill gap: `.pi/skills/go-implementation/SKILL.md` is required by worker instructions but missing in this checkout (`ENOENT`); global Go skills listed below are loaded and recorded.

## Required skills loaded

- GSD/status: `gsd-core`, `caveman`.
- Go/CLI/runtime: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-documentation`, `golang-spf13-cobra`, `golang-security`, `golang-safety`, `golang-context`, `golang-concurrency`, `golang-code-style`.
- Skill rule anchors for handoff: go-how-to CLI routing table; CLI exit-code/stdout-stderr/testing rules; testing best-practices #1, #3, #5; error-handling #1, #2, #7, #9; documentation writing principles and application CLI help; cobra best practices #1, #3, #4, #5 plus `StringArray`/`NoOptDefVal`/unknown-flag guidance; security trust-boundary questions #1-#3/no secrets/command args untrusted; safety #2 and #10; context rules #1/#3/#5; concurrency rules #1/#7; code-style early returns and small focused helpers.

## Scope / exclusions

Allowed:

- Top-level `pm runtime` Cobra node, native `doctor` action, declared runtime flags, and minimal handler adaptation under `internal/cli/**`.
- Runtime-focused tests and unchanged golden/docs parity checks.
- Directly applicable runtime help/manual/website/generated artifacts only if output intentionally changes.
- Issue-local `.planning/phases/424-runtime-native-cobra/**` artifacts.

Excluded:

- Other namespace migrations, connector dynamic dispatch, connector bundles, runtime service implementation, RLM/worker/flow/schedule/ETL/reverse behavior, telemetry spans, parent state/roadmap/PR body, go.mod/go.sum, and other worker branches.
- Completion implementation beyond preserving no-file fallback seams for later Phase 15.
- Credentials/secrets, credentialed connector checks, external runtime service startup, destructive actions, dependency changes, and `main` merge.

## Current behavior notes before production edits

- `runtime` is still a legacy Cobra wrapper with `DisableFlagParsing: true` and no native `doctor` subcommand.
- `runRuntime` returns `errUsage` for bare namespace and invalid actions; the legacy wrapper currently intercepts bare namespace and writes contextual help exit 0.
- `runRuntime` accepts only `doctor`; it does not call `parseFlags` because `doctor` has no runtime-specific flags beyond global `--json`.
- Legacy compatibility to preserve: bare `pm runtime` help exit 0, `pm runtime --help` docs-map help exit 0, `pm runtime --json` JSON `CommandManual`, `pm runtime doctor --json`, late global `--json`, late global `--root`, unknown flags/extra args after `doctor` ignored, invalid actions usage exit 2, JSON/stderr contract unchanged.
- `runtime doctor` routes through `runtimecheck.Doctor(ctx, runtimecheck.FromConfig(cfg))`; tests must use loopback endpoints/config and must not start Podman/PostgreSQL/DragonflyDB/Temporal.
- Help/docs currently come from canonical `docs` map and checked-in `docs/cli/runtime.md`; website runtime docs already describe runtime doctor. No intentional help text change planned.

## Slice plan

1. Planning checkpoint
   - Create phase artifacts and record adapter fallback, missing repo Go skill file, loaded skills, scope, parity checklist, runtime safety stance, and verification plan.
   - Commit/push planning checkpoint after no production files are touched.

2. Red tests
   - Add focused failing tests proving `runtime` is not yet native: top-level command should have `DisableFlagParsing=false`, native `doctor` subcommand should exist, unknown-flag whitelist should be set, and no-file completion fallback seam should be present.
   - Add behavior tests for runtime doctor compatibility: `doctor --json`, `doctor --unknown ignored --extra`, late global `--json`, late global `--root`, config-file endpoints, bare namespace help, `--help`, JSON manual, and invalid action usage.
   - Capture exact red output in `TDD-LEDGER.md` before production code.

3. Green implementation
   - Add native `runtime` subtree in `cobra_router.go` with docs-map help/usage.
   - Add native `runtime doctor` subcommand using declared Cobra node and `FParseErrWhitelist{UnknownFlags: true}`.
   - Adapt runtime handler to a `runRuntimeDoctor(ctx, cfg, stdout, jsonOut)` helper and keep `runRuntime` compatibility only if still used by tests/legacy references; remove `runtime` namespace from legacy wrapper list.
   - Preserve existing output envelopes, redaction, runtime config usage, stdout/stderr discipline, and exit taxonomy.

4. Parity / golden check
   - Expect golden transcript changes: none.
   - Docs/website updates: not applicable if help/output unchanged; verify by docs generator diff and website docs generator.
   - Verify runtime help: `pm help runtime`, bare `pm runtime`, `pm runtime --help`, JSON manual, invalid action JSON usage error, representative runtime doctor JSON with loopback config endpoints.

5. Full verification / PR
   - Run required focused and full gates.
   - Check `git diff --check origin/feat/cli-architecture-v2...HEAD` and `git diff -- go.mod go.sum`.
   - Commit/push coherent checkpoints: planning, red tests when useful, green implementation, verification/PR artifacts.
   - Open non-draft stacked PR against `feat/cli-architecture-v2` with `Refs #424`, `Refs #407`, `Refs #397`.
   - Record automated-review route status; do not request redundant Claude/Copilot review unless fallback conditions apply.

## Planned tests / validations

- `gofmt -w cmd internal`
- `go test ./internal/cli/... -run 'Runtime|CobraRouterShell|Golden' -count=1`
- `go vet ./...`
- `go test ./...`
- `go build ./cmd/pm`
- `make verify`
- `git diff --check origin/feat/cli-architecture-v2...HEAD`
- `git diff -- go.mod go.sum`
- Runtime help/parity after build: `./pm help runtime`, `./pm runtime`, `./pm runtime --help`, `./pm runtime --json`, `./pm runtime bogus --json`, `./pm runtime doctor --json`, `./pm --root "$root" --json runtime doctor`, docs generator diff, docs validate, website docs generator.

## Parity stance

This phase changes parser ownership only. Help text, docs, website, generated manuals, golden stdout/stderr/exit, JSON envelopes, stdout/stderr discipline, global late flags, fresh-tree re-entrancy, completion metadata, runtime service optionality, endpoint redaction, and runtime doctor behavior should remain byte-identical unless an intentional reviewed change is recorded. No intentional user-facing output change is planned.
