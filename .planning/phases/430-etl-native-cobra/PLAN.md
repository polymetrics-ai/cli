# Phase 430 Plan — ETL native Cobra namespace

Issue: polymetrics-ai/cli#430
Umbrella: #407
Parent: #397 / draft parent PR #438
Branch: `refactor/430-etl-native-cobra`
Base branch: `feat/cli-architecture-v2`
Exact starting parent HEAD: `6c94754c58185df5aac53bd97587603c3154b1d5`
Invocation session: `issue-430-pi-openai-codex-gpt-5.6-sol-high-20260718T225346Z`
Explicit invocation profile: `model=openai-codex/gpt-5.6-sol`, `thinking=high`
Execution decision: `local_critical_path` — this is the next serialized Phase 9 namespace in an isolated worktree; central router writes collide with sibling units, no subagent tool is exposed, and the user bounded delivery to #430 with no PR/review/services.

## GSD route

- `scripts/gsd doctor` and `scripts/gsd list`: passed.
- `scripts/gsd prompt plan-phase 430 --skip-research`: generated for inline execution.
- `scripts/gsd prompt programming-loop init --phase 430 --dry-run`: unavailable (`unknown GSD command: programming-loop`).
- Manual fallback: `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`, with strict RED → GREEN → refactor and six issue-local artifacts.

## Required reading and skills

Read issue #430, `AGENTS.md`, GSD/issue/orchestration contracts, CLI parity policy, Stage 9, ADR-0002, ETL router/app/tests/manual/website surfaces, and events/telemetry contracts.

Loaded: `gsd-core`; `golang-how-to`; `golang-cli`; `golang-testing`; `golang-error-handling`; `golang-security`; `golang-safety`; `golang-context`; `golang-concurrency`; `golang-documentation`; `golang-spf13-cobra`.

## Scope

- Replace only the `etl` legacy wrapper with a native Cobra subtree for `check`, `catalog`, `read`, `run`, `status`, and hidden positional `help` compatibility.
- Declare every current ETL flag with typed repeated pflags and legacy bare/assigned/repeated behavior: direct `--connector`, `--config`; read `--stream`, `--limit`; run `--connection`, `--stream`, `--batch-size`, `--runtime`.
- Adapt ETL handlers to typed flag values and remove only ETL's `parseFlags` calls. Keep `directConnector` compatibility for non-ETL callers only if still needed.
- Preserve bounded batches/default 1000, configured sync-mode validation, cancellation, event/telemetry propagation, stdout/stderr separation, one JSON envelope, runtime opt-in semantics, and dependency-free default.
- Use fixture/local temporary connectors and roots only.

Excluded: other namespaces; dynamic connector parser; connector definitions; dependencies; optional services; credentialed connector checks; reverse execution; Phase 15 completion values; Phase 19 help/man churn; PR/review.

## TDD slices and checkpoints

1. **Planning checkpoint** — commit/push these six artifacts before test or production edits.
2. **RED checkpoint** — focused tests specify:
   - native tree/action ownership, current flag types, repeated/bare/assigned forms, fresh tree, and no legacy ETL wrapper;
   - check/catalog/read/run/status behavior using sample and local temporary project fixtures;
   - batch-size integer/default/bounded flush behavior and configured sync-mode cursor/primary-key validation;
   - bare/text/JSON/long/short/positional help; trailing help and literal `--` compatibility;
   - unknown flags, invalid actions, assigned global booleans, and fail-closed action discovery;
   - cancellation propagation and event/telemetry-compatible stdout/stderr/single-envelope behavior.
   Capture focused failure before production edits, then commit/push tests.
3. **GREEN checkpoint** — add the smallest ETL command/typed handler implementation, remove ETL from `cobraLegacyCommands`, add only ETL normalization needed to preserve legacy token semantics, and remove only ETL parser calls.
4. **Refactor/parity checkpoint** — focused/repeated/race/router/golden/full CLI and app tests; built binary help/error/global checks; temporary docs and website generation checks; exact-start differential for preserved cases.
5. **Final checkpoint** — gofmt, vet, full tests, build, `make verify`, scope/dependency guards; finalize artifacts; commit/push; no PR.

## CLI parity stance

The command names, flags, canonical manual text, output envelopes, docs, website page, and generated/golden fixtures should remain byte-identical. Runtime help and temporary generators prove parity. Checked-in docs/website/golden changes are not applicable unless tests reveal a real existing mismatch; Phase 19 owns deliberate help-tree churn.

## Safety

No secrets, external connector checks, optional runtime services, dependency additions, standalone reverse execution, destructive/admin actions, generic write tools, production deployment, or broad generated churn. Tests use only built-in sample/local connectors and `t.TempDir`; `--runtime` parsing is exercised without crossing into service-backed recording. The required `make verify` gate ran its existing temporary-root local smoke, including its built-in approval-gated reverse step; no external or user-data reverse operation ran.

## Completion

Completed and verified at implementation head `fc88f1694ee73593f1130f866bd6166be18eb661`. Strict initial RED and a test-first local-review correction preceded production fixes. Focused/repeated/race/router/golden/full CLI/app/repository, 20-case exact-start differential, runtime help, generated docs/website parity, gofmt, vet, build, scope/dependency guards, and `make verify` passed. No PR or external review was created.
