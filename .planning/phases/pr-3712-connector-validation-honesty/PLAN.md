# PR 3712 — Connector Validation Honesty (`implemented` must mean executable)

## GSD setup

- PR: https://github.com/polymetrics-ai/cli/pull/3712
- Branch: `fm/cli-connector-validation-honesty-r1` (base `b5099e760`, the merged Freshchat/multipart
  upload foundation from PR #3701).
- GSD preflight: `scripts/gsd doctor` passed on 2026-08-04 (node v24.13.1, 69 commands registered).
- GSD prompt path: `scripts/gsd prompt programming-loop init --phase pr-3712-connector-validation-honesty --dry-run`
  was attempted first and the repo-local command registry returned
  `scripts/gsd: unknown GSD command: programming-loop`. `scripts/gsd` was therefore not used
  interactively to drive this slice, and this phase is an explicit manual-GSD fallback recorded per
  `.agents/agentic-delivery/references/gsd-pi-adapter.md`.
- Planning prompt fallback: `scripts/gsd prompt plan-phase pr-3712-connector-validation-honesty --skip-research`
  generated a 142-line prompt that was applied inline as the planning fallback.
- Orchestration decision, all cycles: `local_critical_path` — one connector-layer slice in an
  already isolated worktree, so no mutating subagents were spawned.

## Required skills loaded

- `gsd-core`
- `golang-how-to`
- `golang-testing`
- `golang-error-handling`
- `golang-design-patterns`
- `golang-safety`
- `golang-security`
- `golang-documentation`

## Goal

`availability: implemented` in a connector's `cli_surface.json` is a claim the runtime has to
honour. Before this PR, `cmd/connectorgen/validate.go` restated `commandrunner`'s rules by hand and
had drifted from them, so 174 commands validated clean and then blocked on every invocation. The
phase goal is that no command can claim `implemented` while failing at runtime — enforced by a test
that calls the *real* runtime rather than a second description of it.

## Scope boundaries

### In scope — code (landed in commits `e56a5950b`, `f846ad296`, `7a732e955`)

1. `cmd/connectorgen/validate.go` — close the `direct_read` escape that exempted operation-backed
   commands from the `api_surface` assertion, and treat an empty `output_policy` as a finding rather
   than a skip. Add the `binary_download` executable-intent rules mirroring
   `commandrunner.validateBinaryDownloadCommand` exactly.
2. `internal/connectors/commandrunner/runner.go` + `runner_test.go` — route the `binary_download`
   intent through `engine.OperationBinaryDownload`, and add
   `TestEveryImplementedCommandPassesRuntimePreflight`, which sweeps every bundle in `defs.FS`
   through the real `Preflight`. Because it calls the runtime instead of restating it, a future
   executor kind is covered the day it lands.
3. `cmd/connectorgen/surfacesync.go` (+ `orderedjson.go`, `main.go`, `surfacesync_test.go`) — a
   `surface-sync` subcommand that fills derivable command metadata (`api_surface`, flag `maps_to`,
   `output_policy`, `rest.max_bytes`) from the bundle's own `operations.json`, with `--check` wired
   into `make verify` as the `connectorgen-surface-sync` gate so drift fails CI.
4. `internal/connectors/engine/binary_read.go`, `connector.go`, `binary_download_flags.{go,json}`,
   `internal/cli/cli.go` — bounded binary download execution and the shared download flag surface.
5. `internal/connectors/defs/github/{cli_surface,operations}.json`,
   `internal/connectors/defs/gong/cli_surface.json` — regenerated, never hand-authored.

### In scope — this CI slice (`internal/connectors/conformance/static.go`)

6. `checkSurfaceComplete` builds its `directReads` coverage set from `cli_surface.json`. It admitted
   only `intent: direct_read`, so after `github`'s `artifact download` command became
   `intent: binary_download`, endpoint 7
   (`GET /repos/{owner}/{repo}/actions/artifacts/{artifact_id}/{archive_format}`) no longer resolved
   its `covered_by.direct_reads` target and `TestConformance/github` failed the `surface_complete`
   check — the `verify` job's only failure.

   This is the same defect class the PR exists to close: a second copy of the surface-coverage rule
   that did not move when the first one did. `cmd/connectorgen/validate.go:576-587` already widened
   its identical `directReads` map to `direct_read || binary_download`; conformance's copy at
   `internal/connectors/conformance/static.go:299-306` did not. The fix restores the two copies to
   parity, verbatim, with the same comment.

   The bookkeeping rule itself is unchanged and correct: an `api_surface` row records *which command*
   consumes an endpoint; which executor runs that command does not change who covers it. Availability
   is still enforced — a `planned` command covers nothing.

7. `.planning/phases/pr-3712-connector-validation-honesty/{PLAN,TDD-LEDGER,VERIFICATION}.md` and
   `RUN-STATE.json` — added because `cmd/**` and `internal/**` changed and
   `scripts/verify-gsd-workflow` fails the `gsd-workflow-evidence` gate without changed
   `.planning/**` evidence. These are documentation artifacts: they carry no behavior.

### Out of scope

- No connector definition changes in this CI slice. In particular, `github`'s `api_surface.json`
  keeps `covered_by.direct_reads: ["artifact download"]` on endpoint 7. Rewriting the bundle to
  route around a stale checker would have inverted the PR's own thesis — the checker was wrong, the
  bundle was not.
- No new `api_surface` endpoint is invented anywhere. Every endpoint referenced by a command that
  this PR promotes to `implemented` is substantiated by the connector's own `api_surface.json` and
  `operations.json`.
- No unification of the `cmd/connectorgen` and `internal/connectors/conformance` copies of
  `checkSurfaceComplete`. They are two deliberately separate corpora with separate testdata; merging
  them is a refactor well beyond this root-cause fix, and the standing invariant in `AGENTS.md`
  already names `TestEveryImplementedCommandPassesRuntimePreflight` as the guard that makes
  hand-copied rules detectable.
- The separate `cli_surface`/`streams.json` Tier-3 defect is left untouched and merely reported.
- No push to `main` and no merge; the parent gate stays human-owned.

## Verification plan

- Red: the new `TestCheckSurfaceComplete_BinaryDownloadSatisfiesDirectReadCoverage` must fail on the
  pre-fix checker with the exact CI error string, and its `planned` half must fail if availability
  is ever dropped from the condition.
- Green: `TestConformance/github` passes and the full `internal/connectors/conformance` package is
  green.
- Gates: every `make verify` step that CI never reached (it stopped at `test`) is run individually,
  since `verify` is serial.

See `TDD-LEDGER.md` for red/green evidence and `VERIFICATION.md` for the executed gate results.
