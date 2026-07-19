# Phase 434 Plan — RLM native Cobra namespace

Issue: polymetrics-ai/cli#434
Umbrella: #407
Parent: #397 / draft parent PR #438
Branch: `refactor/434-rlm-native-cobra`
Base branch: `feat/cli-architecture-v2`
Exact starting HEAD: `2ac457a163cbd7bc9a3708da88b03d375ec5e952`
Invocation session: `issue-434-pi-sol-high-20260719T053630Z`
Explicit invocation profile: `Sol`, `high`
Execution decision: `local_critical_path` — #434 is the assigned next serialized Phase 9 unit in an isolated worktree. Its router scope collides with sibling migrations, this Pi session exposes no subagent tool, and the user bounded delivery to implementation/commit/push with no PR or review.

## GSD route

- `scripts/gsd doctor` and `scripts/gsd list`: passed (69 commands).
- `scripts/gsd prompt plan-phase 434 --skip-research`: generated and executed inline.
- `scripts/gsd prompt programming-loop init --phase issue-434 --dry-run`: unavailable because the adapter registry has no `programming-loop` command.
- Manual fallback: `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`, with six issue-local artifacts and strict RED → GREEN → refactor evidence.

## Required reading and skills

Read issues #434/#407/#397; `AGENTS.md`; GSD adapter/core/manual universal loop and active parent-orchestration contracts; issue contract; CLI help/docs/website parity; runtime-RLM integration reference and all canonical runtime/RLM/website docs; architecture plan §5/§9; execution prompt Stage 9; ADR-0002; current RLM CLI/analyzers/spec/fixture/model/agent/worker tests; typed RLM/runtime config; golden transcripts; and adjacent native flow/schedule patterns.

Loaded: `gsd-core`; `golang-how-to`; `golang-cli`; `golang-testing`; `golang-error-handling`; `golang-security`; `golang-safety`; `golang-context`; `golang-concurrency`; `golang-documentation`; `golang-spf13-cobra`.

## Scope

- Replace only the `rlm` legacy wrapper with a native Cobra subtree for the current public `run` action and hidden positional `help` compatibility.
- Declare the complete current local flag surface: `--spec`, `--in`, `--out`, `--mode`, `--dry-run`, and `--request`; preserve repeated last-value, bare value `true`, assigned/space values, ignored unknown flags, ignored operands, and invocation-global `--root`, `--json`, `--plain`, `--no-input`, and `--progress` placement/assignment behavior.
- Adapt RLM execution to typed flag values and remove only RLM's `parseFlags` call site plus `runRLM` dispatcher. Dynamic connector dispatch and other namespace parsers remain untouched.
- Preserve deterministic, fixture, model-stub, and optional agent routing; spec parsing; warehouse path selection; dry-run behavior; exact text/JSON result shapes; error categories; stdout/stderr discipline; context propagation; and optional runtime configuration aliases.
- Add an invocation-local RLM analyzer factory seam so tests inject deterministic fakes for all four routes. Tests may use only `t.TempDir` specs/warehouses and existing hermetic fake runner paths. They must never call a model, Temporal, Podman, a worker service, or another external service.
- Assert the request string is routed only to the agent factory and is absent from output/errors used by the success contract. Preserve the typed RLM analyzer boundary; do not add a generic command/model/shell runner.

Excluded: Phase 16 RLM viewer/dashboard; worker/extract/connector migrations; new RLM modes/actions; model implementation; Temporal/Podman/service invocation; generic runner surfaces; connector bundles; dependencies; credentials; broad generated churn; Phase 19 focused help/man changes; PR/review.

## TDD slices and checkpoints

1. **Planning checkpoint** — commit/push these six artifacts before test or production edits.
2. **RED checkpoint** — add focused tests that specify:
   - native `rlm` ownership for run/help, all current pflags/NoOpt behavior/completion seams, and absence from legacy wrappers;
   - repeated/bare/assigned/space forms for every current flag, ignored operands, and unchanged request/mode selection;
   - deterministic/fixture/model/agent factory routing with injected fakes, context propagation, close behavior, spec/warehouse request contents, and no model/Temporal/Podman/service calls;
   - bare/text/JSON/long/short/positional help; trailing help; literal `--`; malformed/legal unknown flags; invalid action heads; no later-action or operand discovery; and assigned/late global flags;
   - exact usage/internal error mapping, stdout/stderr and one-envelope JSON behavior, no request leakage, and no generic runner/viewer action.
   Capture focused failure before any production edit; commit/push tests.
3. **GREEN checkpoint** — add the smallest native RLM command, typed flags, bounded RLM-only normalization, injected runtime seam, and typed run handler; remove RLM from `cobraLegacyCommands`; delete only `runRLM` and RLM's `parseFlags` use.
4. **Refactor/parity checkpoint** — run focused/repeated/race RLM, analyzer/router/golden/full CLI, worker fake, exact-start parser/output differential, runtime help, generated docs/website, formatting, vet, build, and scope/dependency guards.
5. **Final checkpoint** — run dependency-free full `make verify`; finalize all six artifacts; commit/push; no PR or review.

## CLI parity stance

Public command names, flags, manuals, result schemas, docs, website content, generated artifacts, and golden fixtures should remain unchanged. Checked-in docs/website/golden edits are not applicable unless verification finds an actual mismatch. Phase 16 owns the RLM viewer; Phase 19 owns deliberate focused-help/man churn. Verify `pm help rlm`, bare `pm rlm`, `pm rlm --help`, short/positional/JSON routes, invalid actions, `docs/cli/rlm.md`, website CLI-reference/architecture pages, generated docs, completion discovery, and golden transcripts.

## Safety

No secrets/request leakage, model calls, Temporal, Podman, optional services, worker daemon, credentials, dependencies, external connectors, generic execution surfaces, unrestricted writes, destructive/admin actions, reverse ETL, or production deployment. Tests use only temp specs/warehouses and injected analyzers/hermetic fakes. Context cancellation remains propagated. Agent mode stays opt-in; the default deterministic/fixture paths stay dependency-free.

## Review correction at `92f26587`

Review finding: `runRLMRun` currently forwards `--request` to every analyzer factory even though the phase contract permits request content only at the agent-mode factory boundary. The built-in deterministic, fixture, and model factories ignore it, so public CLI output and mode behavior are unchanged; the seam is nevertheless broader than required.

Correction slice:

1. Update all six phase artifacts before tests or production edits; record manual GSD fallback because `scripts/gsd prompt programming-loop ...` remains absent from the adapter registry.
2. RED first: change the injected-factory contract so deterministic, fixture, and model receive an empty request while agent receives the parsed request. Keep request values out of test diagnostics and verify text/JSON output compatibility through hermetic fakes only.
3. GREEN: add the smallest mode gate at the analyzer-factory call site. Do not alter analyzer implementations, mode selection, pflags, help, output, errors, service wiring, dependencies, docs, website, or golden fixtures.
4. Verify focused and race tests, the reviewer's 1,984-case exact-start parser/output differential, full RLM and CLI suites, request non-disclosure, gofmt, vet, build, and diff/scope guards. Do not call a model, Temporal, Podman, worker service, or another optional service.
5. Finalize artifacts, commit, and push to the existing issue branch. No PR or review.

Correction execution decision: `local_critical_path` — one bounded test/seam correction in the existing isolated issue worktree; no subagent tool is exposed and the user prohibited PR/review/services/dependencies.

## Completion

Completed and verified at implementation head `633f1e21`. Native RLM owns run/help and every current local flag; only the RLM wrapper/dispatcher/`parseFlags` call were removed. Injected analyzer factories verify deterministic/fixture/model/agent routing, context, closure, request isolation, spec/warehouse mapping, and outputs without model, Temporal, Podman, worker service, or other external calls. Exact-start differential matched 24/24 cases. Focused/repeated/race/router/golden/full CLI, RLM, worker-fake, runtime help, generated docs/website, gofmt, vet, full repository tests, build, scope/dependency guards, and `make verify` pass. Public help/docs/website/golden bytes are unchanged; Phase 16 viewer work remains deferred. Implementation and planning checkpoints are pushed; no PR or review was created.
