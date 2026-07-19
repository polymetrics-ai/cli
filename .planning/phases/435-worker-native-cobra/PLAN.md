# Phase 435 Plan — worker native Cobra namespace

Issue: polymetrics-ai/cli#435
Umbrella: #407
Parent: #397 / draft parent PR #438
Branch: `refactor/435-worker-native-cobra`
Base branch: `feat/cli-architecture-v2`
Exact starting HEAD: `14c02d295065c3bf33c65eaac5f8d36642798f81`
Invocation session: `issue-435-pi-sol-high-20260719T064417Z`
Explicit invocation profile: `Sol`, `high`
Execution decision: `local_critical_path` — #435 is the assigned next serialized Phase 9 unit in an isolated worktree. Its central router scope collides with sibling migrations, this session exposes no subagent tool, and the user bounded delivery to implementation/commit/push with no PR or review.

## GSD route

- `scripts/gsd doctor` and `scripts/gsd list`: passed (69 commands).
- `scripts/gsd prompt plan-phase 435 --skip-research`: generated and executed inline.
- `scripts/gsd prompt programming-loop init --phase 435 --dry-run`: unavailable because the adapter registry has no `programming-loop` command.
- Manual fallback: `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`, with these six issue-local artifacts and strict RED → GREEN → refactor evidence.

## Required reading and skills

Read issues #435/#407/#397; `AGENTS.md`; GSD adapter/core/manual universal loop; parent-orchestrator and issue contracts; CLI help/docs/website parity; runtime-RLM integration reference and canonical runtime/RLM/website docs; architecture plan §5/§9; execution prompt Stage 9; ADR-0002; current worker CLI, Temporal worker, typed config, probe, golden, router, and adjacent native RLM/schedule code/tests.

Loaded: `gsd-core`; `golang-how-to`; `golang-cli`; `golang-testing`; `golang-error-handling`; `golang-security`; `golang-safety`; `golang-context`; `golang-concurrency`; `golang-documentation`; `golang-spf13-cobra`.

## Scope

- Replace only the hidden `worker` legacy wrapper/parser with a hidden native Cobra subtree for `status`, `serve`, and help compatibility.
- Keep the worker typed and RLM-workflow-only. Do not add a generic command, shell, HTTP, SQL, container, or workflow runner.
- Preserve `WorkerStatus` and `WorkerServe` text/JSON envelopes, task queue, readiness ordering, error taxonomy, stdout/stderr discipline, explicit Temporal configuration, RLM image/Podman activity configuration, and cancellation propagation.
- Native `worker` remains omitted from root command discovery. Direct bare/text/JSON/long/short/positional/trailing help becomes contextual and side-effect free per #435 acceptance; update only directly applicable worker manual/golden parity artifacts after reviewing fixture changes.
- Worker has no local flags today. Preserve invocation-global `--root`, `--json`, `--plain`, `--no-input`, and `--progress` placement/assignment behavior; preserve legacy operand ownership where it does not conflict with native help interception.
- Preserve strict first-action discovery: malformed/legal unknown flags, literal `--`, and operands must not reveal a later `status` or `serve`; invalid action heads remain usage errors. Trailing operands on a selected action remain ignored as before.
- Introduce invocation-local fake status/probe and serve seams. Every native parsing/help/error/config test uses fakes and asserts invalid/help paths start neither a probe nor a worker. No test may dial Temporal, start a Temporal worker, invoke Podman, open a network listener, access a database, or start runtime services.
- Verify config precedence for explicit Temporal address and RLM Podman/image settings (primary env > legacy alias > file > default, with worker requiring an explicit Temporal source), while proving help/errors/status do not disclose unrelated configuration canaries.
- Remove only the worker dispatcher and worker entry from `cobraLegacyCommands`. Dynamic connector parsing and all other namespace parsers remain untouched.

Excluded: Temporal workflow/activity redesign; RLM analyzer changes; Phase 16 dashboards; extract/connectors/certify migrations; connector bundles; dependencies; credentials; service-backed integration tests; broad generated churn; Phase 19 help-tree/man changes; PR/review.

## TDD slices and checkpoints

1. **Planning checkpoint** — commit/push these six artifacts before test or production edits.
2. **RED checkpoint** — add focused tests that specify:
   - a hidden native `worker` subtree owning `status`, `serve`, and hidden positional help, with no legacy wrapper;
   - bare/text/JSON/long/short/positional/trailing help, all side-effect free;
   - current no-local-flag surface, ignored trailing operands, strict first-action ownership, literal `--`, malformed/legal unknown flags, invalid actions, and no later action discovery;
   - assigned/space/repeated/late global flags and explicit config precedence;
   - fake status/serve routing, context propagation/cancellation, ready callback ordering, text/JSON output, exact error categories, one-envelope behavior, and nondisclosure;
   - no probe/worker/service start for help, malformed config, invalid globals, unknown actions, or missing explicit Temporal configuration.
   Capture focused failure before production edits; commit/push tests.
3. **GREEN checkpoint** — add the smallest native hidden worker command, bounded worker-only action normalization, invocation-local status/serve runtime seam, typed status/serve handlers, dedicated direct worker manual if required by acceptance, and remove only the worker legacy dispatcher.
4. **Refactor/parity checkpoint** — run focused/repeated/race worker, router/golden/full CLI, worker package fake tests, exact-start parser/output differential, config precedence/nondisclosure, runtime help, generated docs/website parity, formatting, vet, build, and scope/dependency guards.
5. **Final checkpoint** — run dependency-free full `make verify`; finalize all six artifacts; commit/push; no PR or review.

## CLI parity stance

The command stays hidden from root discovery and completion metadata. Direct worker help/bare invocation is intentionally normalized to contextual success by #435; status/serve output schemas and errors otherwise remain compatible. Review only worker-related golden/manual fixture deltas. Verify `pm help worker`, bare `pm worker`, `pm worker --help`, short/positional/trailing/JSON help, invalid actions, `docs/cli/worker.md` applicability, website CLI-reference/architecture text, generated docs, hidden discovery, and golden transcripts. Phase 16 owns dashboards; Phase 19 owns broad focused-help/man churn.

## Safety

Correction invariant: no secrets or config canaries in diagnostics and no Temporal dial/worker, network listener, Podman command, database, runtime service, credential, external connector, dependency, generic execution surface, destructive/admin action, reverse ETL, or production deployment. Corrected worker/config tests must use only invocation-local fakes and temp config roots. Cancellation remains propagated. The worker continues to serve only the typed RLM Temporal workflow and Podman activity.

## P2 test-isolation correction

Accepted at exact correction HEAD `f692225ab53a3c0467d42c0ac3e9810107d73a82` from `/tmp/pm-397-review-435.log`. The prior fake-only/no-dial verification was inaccurate: `TestWorkerStatusUsesExplicitConfigFileTemporalAddr` called `Run`, selected the production `temporalprobe.Probe`, and attempted a loopback Temporal dial.

Before any production edit, reset phase verification to pending. This bounded correction changes only test/evidence artifacts: first add a fake-status call/address assertion that fails while the test still uses `Run`, then route that config-migration status case through `runWorkerInvocation` and the invocation-local fake. Preserve the config-file precedence/address and unavailable-envelope assertions. Verify focused, repeated, and race worker/config tests; audit those tests for production network dial calls; run broader CLI only if needed; then run diff, gofmt, and vet. No services, dependencies, production behavior, PR, or review.

## Completion

Correction complete at test commit `01d70f55e755bd57b31662ccd333f34916de0563`. `TestWorkerStatusUsesExplicitConfigFileTemporalAddr` now invokes the native worker through `runWorkerInvocation`, asserts exactly one fake status call, the configured address, and config-file source, and retains the unavailable JSON assertions. Focused/repeated/race CLI worker/config and `internal/worker`/`internal/config` tests passed; source audits found no production `Run`/probe/dial path in worker CLI tests or the corrected status case; diff, gofmt, and `go vet ./...` passed. Full CLI was not needed for this test-only correction and would include unrelated runtime-probe tests, so no broader no-dial claim is made. No production file, service, dependency, PR, or review was used.
