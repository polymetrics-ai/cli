# Phase 423 Plan — Perf native Cobra namespace

Issue: polymetrics-ai/cli#423
Umbrella: #407
Parent: #397 / PR #438
Branch: `refactor/423-perf-native-cobra`
Base branch: `feat/cli-architecture-v2`
Base parent head at dispatch/planning: `6fbff849932e891a8184000fb677e1b6fca7f6d4`
Execution decision: `local_critical_path` — third serialized Phase 9 namespace worker; cwd/branch isolated; worker has no subagent tool and must not delegate.

## Required reading complete

- Issue #423 body and acceptance criteria; umbrella #407 roster; parent #397 and draft parent PR #438 context.
- `AGENTS.md`; issue-agent, parent/subissue, automated-review, Claude-review, worker-handoff contracts/workflows.
- GSD universal runtime loop; `.planning/config.json`, `PROJECT.md`, `ROADMAP.md`, `STATE.md`, universal programming loop PRD/prompts.
- Required-skill routing, GSD Pi adapter, CLI help/docs/website parity.
- CLI Architecture v2 plan §5/§9, execution prompt Stage 9, ADR 0002.
- Current perf parser/handlers/tests/docs/goldens: `internal/cli/cobra_router.go`, `internal/cli/cli.go`, `internal/cli/parse.go`, `internal/cli/docs.go`, `internal/cli/cli_test.go`, `internal/cli/config_migration_test.go`, `internal/cli/golden_transcript_test.go`, `docs/cli/perf.md`, `website/content/docs/perf.mdx`, `website/lib/docs.generated.ts`, `internal/perf/perf.go`.
- Phase 406 catalog, Phase 421 connections, and Phase 422 query native-Cobra artifacts used as Stage 8/9 templates.

Current phase artifacts did not exist at kickoff; this checkpoint creates them before production edits.

## GSD adapter

- `scripts/gsd doctor` — pass.
- `scripts/gsd prompt plan-phase 423 --skip-research >/tmp/gsd-plan-phase-423.prompt` — pass.
- `scripts/gsd prompt programming-loop init --phase 423 --dry-run >/tmp/gsd-programming-loop-423.prompt` — blocked: `scripts/gsd: unknown GSD command: programming-loop`; manual GSD fallback active using `.pi/prompts/pm-gsd-loop.md`, `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`, and the issue contract.
- Adapter/skill gap: `.pi/skills/go-implementation/SKILL.md` is required by worker instructions but missing in this checkout (`ENOENT`); global Go skills listed below are loaded and recorded.

## Required skills loaded

- GSD/status: `gsd-core`, `caveman`.
- Go/CLI: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-documentation`, `golang-spf13-cobra`, `golang-security`, `golang-safety`, `golang-code-style`.
- Skill rule anchors for handoff: go-how-to CLI routing table; CLI exit-code/stdout-stderr/testing rules; testing best-practices #1, #3, #5; error-handling #1, #2, #7, #9; documentation writing principles and application CLI help; cobra best practices #1, #3, #4, #5 plus `StringArray`/`NoOptDefVal`/unknown-flag guidance; security trust-boundary questions #1-#3/no secrets/command args untrusted; safety #2 and #10; code-style early returns and clear small helpers.

## Scope / exclusions

Allowed:

- Top-level `pm perf` Cobra node, native `compare` and `sync-modes` actions, declared perf flags, and minimal handler adaptation under `internal/cli/**`.
- Perf-focused tests and unchanged golden/docs parity checks.
- Directly applicable perf help/manual/website/generated artifacts only if output intentionally changes.
- Issue-local `.planning/phases/423-perf-native-cobra/**` artifacts.

Excluded:

- Other namespace migrations, connector dynamic dispatch, connector bundles, runtime/RLM/worker/flow/schedule/ETL/reverse behavior, telemetry spans, parent state/roadmap/PR body, go.mod/go.sum, and other worker branches.
- Completion implementation beyond preserving declared flag metadata / no-file fallback seams for later Phase 15.
- Credentials/secrets, credentialed connector checks, external runtime service startup, destructive actions, dependency changes, and `main` merge.

## Current behavior notes before production edits

- `perf` is still a legacy Cobra wrapper with `DisableFlagParsing: true` and no native `compare` or `sync-modes` subcommands.
- `runPerf` returns `errUsage` for bare namespace; the wrapper currently intercepts bare namespace and writes contextual help exit 0.
- `runPerf` parses action flags with `parseFlags(args[1:])` for `compare` and `sync-modes`.
- Legacy flag behavior: `--iterations`, `--runtime`, and `--records` support `--flag value`, `--flag=value`, repeated last-wins for scalar flags, bare `--flag` becomes `"true"`, and unknown flags/extra positional args are ignored.
- `compare --runtime` routes through `runtimecheck.Doctor` using `runtimecheck.FromConfig(cfg)`; this phase must preserve typed config usage and avoid starting runtime services.
- Help/docs currently come from canonical `docs` map and checked-in `docs/cli/perf.md`; website perf docs are generated/linked from canonical docs. No intentional help text change planned.

## Slice plan

1. Planning checkpoint
   - Create phase artifacts and record adapter fallback, missing repo Go skill file, loaded skills, scope, parity checklist, perf/runtime safety stance, and verification plan.
   - Commit/push planning checkpoint after no production files are touched.

2. Red tests
   - Add focused failing tests proving `perf` is not yet native: top-level command should have `DisableFlagParsing=false`, native `compare`/`sync-modes` subcommands should exist, declared flags should have legacy-compatible metadata (`StringArray` + `NoOptDefVal="true"`), unknown-flag whitelist, and no-file completion fallback.
   - Add behavior tests for perf flag forms: equals, space, repeated scalar last-wins, bare bool/value sentinel preservation, unknown flag tolerance, extra args tolerance, late global `--root`/`--json`, runtime config endpoints, invalid action usage, and bare namespace help.
   - Capture exact red output in `TDD-LEDGER.md` before production code.

3. Green implementation
   - Add native `perf` subtree in `cobra_router.go` with docs-map help/usage.
   - Add native `perf compare` and `perf sync-modes` declared flags and normalization for optional value flags so pflag `NoOptDefVal` does not swallow legacy space-form values.
   - Add perf handler adapters that accept parsed native flag values and preserve existing validation order, output envelopes, runtime config usage, and benchmark behavior.
   - Remove the `perf` namespace legacy wrapper/`parseFlags` call site.

4. Parity / golden check
   - Expect golden transcript changes: none.
   - Docs/website updates: not applicable if help/output unchanged; verify by docs generator diff and website docs generator.
   - Verify runtime help: `pm help perf`, bare `pm perf`, `pm perf --help`, JSON manual, invalid action JSON usage error, representative perf JSON outputs, and runtime compare config use with local-loopback endpoints.

5. Full verification / PR
   - Run required focused and full gates.
   - Check `git diff --check origin/feat/cli-architecture-v2...HEAD` and `git diff -- go.mod go.sum`.
   - Commit/push coherent checkpoints: planning, red tests when useful, green implementation, verification/PR artifacts.
   - Open non-draft stacked PR against `feat/cli-architecture-v2` with `Refs #423`, `Refs #407`, `Refs #397`.
   - Record automated-review route status; do not request redundant Claude/Copilot review unless fallback conditions apply.

## Planned tests / validations

- `gofmt -w cmd internal`
- `go test ./internal/cli/... -run 'Perf|CobraRouterShell|Golden' -count=1`
- `go vet ./...`
- `go test ./...`
- `go build ./cmd/pm`
- `make verify`
- `git diff --check origin/feat/cli-architecture-v2...HEAD`
- `git diff -- go.mod go.sum`
- Runtime help/parity after build: `./pm help perf`, `./pm perf`, `./pm perf --help`, `./pm perf --json`, invalid action JSON usage error, representative `perf compare`/`perf sync-modes` JSON, docs generator diff, docs validate, website docs generator.

## Parity stance

This phase changes parser ownership only. Help text, docs, website, generated manuals, golden stdout/stderr/exit, JSON envelopes, stdout/stderr discipline, global late flags, fresh-tree re-entrancy, completion metadata, runtime service safety, and perf benchmark behavior should remain byte-identical unless an intentional reviewed change is recorded. No intentional user-facing output change is planned.

## Review-fix cycle — PR #458 head `3a50385714c52c7483e6a137ced3285f73f2b929`

Execution decision: `local_critical_path` — review-fix is in the existing isolated worker cwd/branch; no subagent tool available; keep same PR/branch and do not reset.

Accepted findings to fix:

1. Perf manual/help/goldens omit exit 3 even though invalid numeric flags return validation errors.
2. Perf SECURITY text says counts/durations only, but `--runtime --json` includes `runtime_report` check metadata.
3. `perf compare --iterations` and `perf sync-modes --records` are unbounded workload controls.
4. Runtime-backed perf error strings need the shared redaction path before storage/printing.
5. Dragonfly/Temporal endpoints in `runtime_report` are topology metadata, not credentials; document this rather than suppressing them.
6. go-redis diagnostics can bypass CLI formatting on runtime health failures; route them through the redacting logger and keep terminal output quiet when possible.

Review-fix scope:

- Allowed production scope: `internal/cli/**` perf parsing/docs/tests, `internal/perf/**` benchmark guards/cleanup/redaction, `internal/runtimecheck/**` runtime health diagnostic routing, `docs/cli/perf.md`, website docs/generated data for perf/runtime metadata parity, and this phase directory.
- Excluded: #410 telemetry, new dependencies, runtime service startup, credentialed checks, parent ledgers/PR merge state, broad connector/generated rewrites outside the named perf/docs surfaces, and `main` push.

TDD red plan before production edits:

- Add CLI tests for invalid and oversized `--iterations`/`--records` returning exit 3 with JSON `validation_error`, while small/default values still pass.
- Add docs/manual parity tests that perf help mentions exit 3 and runtime metadata/redaction/topology semantics.
- Add internal perf tests for direct API max guards and redaction of runtime-backed error strings before `RuntimeBacked.Error` is exposed.
- Add runtimecheck test that a failing Dragonfly health check no longer writes raw `redis:` library diagnostics to process stderr.
- Capture exact failing output in `TDD-LEDGER.md` before editing production code.

Minimal implementation plan:

1. Define bounded perf workload defaults/maxima in `internal/perf`; enforce direct API guards and CLI validation with exit 3 for oversized/invalid values.
2. Redact runtime-backed errors in `perf.Compare` using `internal/logging` before storing them in results; keep runtime report check errors on the existing redaction path.
3. Add safe temp-root cleanup for generated perf benchmark roots with `defer os.RemoveAll(root)`.
4. Route go-redis internal diagnostics through the shared logger from runtimecheck, at a quiet level, so they inherit redaction and stop bypassing CLI/test formatting.
5. Update canonical perf manual text in `internal/cli/docs.go`, regenerate `docs/cli/perf.md`, update golden transcripts, update website content/docs/generated data where applicable, and document Dragonfly/Temporal endpoints as non-secret topology metadata.

Review-fix verification plan:

- Focused red/green: `go test ./internal/cli/ -run 'Perf|GoldenDocs|GoldenTranscripts' -count=1`, `go test ./internal/perf -count=1`, `go test ./internal/runtimecheck -count=1`.
- Docs/website: `go run ./cmd/pm docs generate --dir docs/cli --connectors-dir "$TMP_CONNECTORS"`, `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test ./internal/cli/ -run TestGoldenTranscripts -count=1`, `npm --prefix website run gen:docs` (or `gen:website-data` if docs data requires connector data refresh).
- Required gates after green: `go test ./internal/cli/...`, `gofmt -w cmd internal`, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify`, plus diff guards for go.mod/go.sum and CLI docs/website generated parity.

Review-fix result: implemented and locally verified. All accepted findings are dispositioned in-scope; runtime-backed integration services were not started by task constraint, and loopback-only runtime metadata checks were used.
