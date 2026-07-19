# Phase 437 Plan — connectors and certify native Cobra

Issue: #437
Umbrella: #407
Parent: #397 / draft parent PR #438
Branch: `refactor/437-connectors-certify-native-cobra`
Base branch: `feat/cli-architecture-v2`
Exact starting HEAD: `6c038bb4ab4a5497fca28a0cab42d0a7fa4eb22b`
Invocation: `issue-437-pi-sol-high-20260719T095145Z`; profile `Sol`; thinking `high`.
Execution decision: `local_critical_path` — this is the final assigned serialized Phase 9 namespace in an isolated worktree; router changes collide with sibling units, no subagent tool is exposed, and the user requested implementation/commit/push with no PR/review.

## GSD route

- `scripts/gsd doctor` and `scripts/gsd list`: pass (69 commands).
- `scripts/gsd prompt plan-phase 437 --skip-research`: generated and executed inline.
- Required `programming-loop` command is absent from the adapter registry (`unknown GSD command`), so the manual fallback is `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md` with these six pre-production artifacts and strict RED → GREEN → refactor.

## Required context and skills

Read issue #437 and parents #407/#397; `AGENTS.md`; GSD/manual/issue/parent contracts; connector migration handoff, conventions, v2 design, certification design/contracts; CLI Architecture v2 plan/execution prompt; ADR-0002; CLI help/docs/website parity; current connectors/certify/router/golden/manual/website code.

Loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-context`, `golang-concurrency`, `golang-documentation`, `golang-spf13-cobra`.

## Scope

- Replace only the legacy top-level `connectors` wrapper and `runConnectors`/`runCertify` namespace parsers with native Cobra commands.
- Native connector actions: `list`, `catalog`, `inspect`, hidden positional `help`, and compatibility aliases `man`/`docs`; preserve current metadata-only behavior and output.
- Native nested certify actions: single connector, `--all`, and `--sweep`, declaring every currently consumed flag: `credential`, `from-env`, `config`, `stream`, `limit`, `modes`, `skip`, `rate-limit`, `budget`, `record`, `replay`, `live-all-modes`, `allow-production-writes`, `keep-workdir`, `write`, `full`, `credentials-file`, `parallel`, `resume`, `older-than`.
- Preserve legacy repeated-last, bare, assigned, space, operand, ignored trailing, unknown-flag, literal `--`, malformed unknown, action/operand discovery, invocation-global, text/JSON, and error-category behavior where current handlers consume it.
- Preserve `cli.Run` in-process re-entrancy, certify exit 0/1/2/3 mapping, context cancellation, bounded cross-connector concurrency, event sequence, telemetry span names/status, and secret/credential-value exclusion.
- Keep dynamic `pm <connector> <path...>` dispatch and its legacy `parseFlags` path exactly sanctioned and unchanged.
- Certify verification is fixture/replay/local only. No live credential checks or writes.

Excluded: connector defs, connector runtime behavior, new certify semantics/flags, live tests, credential values, external services, new dependencies, dynamic parser removal, other namespaces, Phase 16 dashboard, Phase 19 help-tree churn, PR/review.

## TDD and checkpoints

1. Commit/push these six planning artifacts before test or production edits.
2. RED: add focused tests for native tree shape, complete current flags, connector/certify actions and operands, bare/text/JSON/topic/positional/trailing help, literal separator/malformed unknown/action discovery/globals, exact outputs/errors, exits 0/1/2/3, re-entrancy, cancellation/concurrency/events/telemetry, and planted credential-value absence. Capture failure caused by absent native constructors/runtime seam; commit/push.
3. GREEN: introduce the smallest typed flag structs, native constructors, handler adaptation/runtime seam, and compatibility normalization. Remove only connectors/certify namespace `parseFlags` calls and connectors legacy registration.
4. Refactor: focused ×10, race, router/golden/full CLI/certify; exact-base differential; connector validation; docs/manual/website generation; runtime smoke; gofmt/vet/test/build/make verify.
5. Finalize six artifacts, commit/push, no PR/review.

## Parity and safety

Bare `pm connectors` must render the canonical manual and exit 0; invalid actions remain usage exit 2. Update the canonical connectors manual to document certify commands and 0/1/2/3 exits, regenerate `docs/cli/connectors.md`, and mirror the bounded surface in `website/content/docs/cli-reference.mdx`; generated website data follows existing scripts. Completion registration is unchanged and Phase 15/19 work is not pulled forward.

All certify command tests use sample fixtures, replay/local fakes, `t.TempDir`, injected runner/sweeper seams, synthetic credential variable names, and planted non-secret sentinel values solely to assert absence. Never print, summarize, or persist a credential value. No real env-secret resolution, live connector check, write, sweep against external systems, reverse ETL execution, model/runtime service, dependency, or broad path.

## Completion

Implemented through three test-first slices: absent native constructors, contextual trailing action help, and direct inspect help before private operand capture. Final implementation keeps dynamic connector parsing unchanged, passes exact-start operations 21/21, focused/repeated/race/router/golden/full CLI/certify gates, runtime help/docs/website parity, explicit local sample certify smoke, connector validation 547/0, and final `make verify`. Delivery remains commit/push only with no PR/review.

## Accepted review correction — exact HEAD `0d1792cec3ea829ceb6228fc600b6dc7bbd90eee`

Session `issue-437-review-correction-20260719T113319Z` reopens the phase for all five findings in `/tmp/pm-397-review-437.log`. The adapter still lacks `programming-loop`; `scripts/gsd doctor` passed and the manual universal runtime loop remains the recorded fallback. Execution decision is `local_critical_path`: one bounded safety-critical correction in the existing isolated issue worktree, no subagent capability, no services/credentials/PR/review.

Correction slices, in order:

1. **Artifacts/checkpoint:** mark verification false and record the accepted findings, skills, commands, and safety scope; commit and push before tests or production.
2. **RED:** add differential tests proving unsupported safety/mode controls cannot run as no-ops; single certify starts exactly one span before connector validation and option parsing, with connector validation precedence; batch loads the credential file before parsing `--parallel` and preserves byte-exact load/run wrappers; only exact connectors help tokens render manuals; docs distinguish CLI pre-report exits from completed-report exits. Commit/push failing tests before production.
3. **GREEN:** reject and unadvertise controls with no existing typed runner/stage support (`record`, `replay`, `allow-production-writes`, `rate-limit`, `budget`, `live-all-modes`); restore the legacy single and batch execution ordering/wrappers; tighten connectors-only help normalization; correct canonical/generated/website docs. Preserve write opt-in/cleanup gates and all existing redaction behavior.
4. **VERIFY:** focused differential, repeated/race, certify exits/redaction, unsupported replay no-live/runtime-call test, local sample smoke, docs/golden/website generation, full CLI/certify, gofmt/vet/test/build/`make verify`, and connector validation. Use fixture/temp inputs only.
5. **DELIVER:** finalize all six artifacts, coherent commits and pushes only; no PR or automated review per user instruction.

Correction completed at implementation head `a67d2ff9de84a2fabcd3b66097bf49518c1fa124`. All six controls without typed runner/stage support are hidden and fail closed before any runner invocation; replay therefore cannot touch credentials or live/write stages. Legacy single span and connector-validation-before-option precedence, batch file-before-parallel precedence and exact load/run wrappers, and exact-only connectors help behavior are restored. Canonical/generated/website docs now separate pre-report CLI exits from completed-report outcomes. Differential 5/5, focused/repeated/race, redaction/replay/exit, local sample, docs/golden/website, full CLI/certify, gofmt/vet/test/build, `make verify`, and connector validation all pass.

Loaded for the correction: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-documentation`, and `golang-spf13-cobra`.

## Second accepted safety correction — exact HEAD `0d743e54e06c9e27e550eacce9be7899a9e23d19`

Session `issue-437-second-safety-correction-20260719` reopens the phase for all three priority groups in `/tmp/pm-397-rereview-437.log`; every finding is accepted. `scripts/gsd doctor` and `scripts/gsd prompt plan-phase 437 --skip-research` passed. The required `programming-loop` command remains absent (`unknown GSD command`), so the manual universal runtime loop is the recorded fallback. Execution decision: `local_critical_path` because this is one safety-critical correction in the existing isolated issue worktree, no subagent tool is exposed, and the user prohibited credentials, services, dependencies, PR, and review.

Second-correction slices, in order:

1. **Artifacts/checkpoint:** mark verification false and record accepted P1/P2/P3 findings, required skills, strict safety scope, RED/GREEN boundaries, full flag audit, verification matrix, and commit/push checkpoints. Commit and push before tests or production edits.
2. **RED/checkpoint:** add effect-recorder tests proving batch `--write=false` and `--skip=write` dominate credential entries with `write: true`; configured credential-file sandbox/rate/budget (and every other unsupported accepted control discovered by the audit) fail before runner effects; unsupported or mode-inapplicable flags and skip values fail before effects; every declared certify flag is either used or rejected. Add docs/help claim tests. Run focused tests to capture failure, then commit and push before production edits.
3. **GREEN P1/P2:** preserve safe base behavior while adding batch write-disable overrides; fail closed rather than discard unsupported credential-file constraints; hide/reject unsupported single `--credential`, `--limit`, and `--modes`; allow only implemented `--skip=write`; reject flow/schedule/unknown skip values and all mode-inapplicable controls before runner effects. Keep dynamic dispatch, certify exit mapping, telemetry, and redaction unchanged.
4. **GREEN P3:** remove stale rejected-option examples and claims from certification architecture and the universal-loop PRD; correct the connector-help name claim to match the namespace manual behavior; regenerate CLI docs, golden/help artifacts, and website generated data through repository generators only.
5. **VERIFY:** focused effect/no-op and flag-audit tests; repeated and race runs; local sample fixture smoke; full CLI/certify/docs/website gates; gofmt, vet, full test, build, `make verify`, and `connectorgen validate`. Fixture/temp inputs only; no credentials, live services, or external writes.
6. **DELIVER:** finalize all six artifacts, commit coherent green/verification checkpoints, and push only the active issue branch. No PR or review.

Loaded for this cycle: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-documentation`, and `golang-spf13-cobra`.

Second correction completed at implementation head `7b6eaa58`. Batch write-disable controls now dominate credential-file write settings; sandbox gates enabled writes; unsupported rate, budget, and read-limit constraints fail before runner construction. Unsupported single controls are hidden/rejected, skip is allowlisted to write, and every mode rejects inapplicable declared controls before runtime effects. Architecture/PRD/help claims and generated CLI/website data are accurate. Focused, repeated, race, no-op/effect, local sample, full CLI/certify, docs/website, gofmt/vet/test/build, `make verify`, and connectorgen gates pass without credentials, services, dependencies, PR, or review.

## Third accepted safety/correctness correction — exact HEAD `437d13cf`

Session `issue-437-third-safety-correction-20260719`; all findings in `/tmp/pm-397-rereview2-437.log` are accepted. `scripts/gsd doctor` and `scripts/gsd list` passed, but the required `scripts/gsd prompt programming-loop ...` route remains absent (`unknown GSD command`), so the manual universal runtime loop is the recorded fallback. Execution decision: `local_critical_path` because this is one bounded correction in the existing isolated issue worktree, no subagent tool is exposed, and the user prohibited credentials, external commands, services, dependencies, PR, and review.

Third-correction slices, in order:

1. **Artifacts/checkpoint:** set verification/state to honestly pending and record accepted findings, skills, strict fixture/temp safety scope, RED/GREEN boundaries, verification matrix, and commit/push checkpoints before tests or production edits.
2. **RED/checkpoint:** add native-Cobra subtree tests proving every unknown flag—including write-like typos—returns usage before credential loading, runner, or sweep effects; add strictly-positive/reasonably-bounded `--older-than` no-effect cases; add credential-file `exec` rejection before runner effects; add a two-run batch effect-recorder proving ordinary completed reports resume while incomplete reports rerun. Update the flag/docs audit assertions. Capture and commit the failures before production edits.
3. **GREEN safety:** remove certify subtree unknown-flag whitelisting while preserving non-certify connector behavior only where required; validate sweep age before `runtime.Sweep`; reject any credential-file `exec` entry before batch effects; remove the generic external execution implementation and claims.
4. **GREEN correctness/docs:** make ordinary `--resume` reuse a valid completed prior report without timestamp fabrication, while incomplete/failed report artifacts rerun; correct CLI usage-exit docs, stage examples to `ga`, resume/sweep safety wording, canonical/generated/website docs, and architecture examples.
5. **VERIFY:** focused and repeated unknown-flag/no-effect/exec/resume/sweep tests; race; full flag/docs audit; docs/golden/website generation and drift; local credential-free sample smoke; full CLI and certify suites; gofmt, vet, full tests, build, `make verify`, and connectorgen validation. Fixture/temp inputs only; no external command invocation, credentials, services, or writes.
6. **DELIVER:** finalize all six artifacts with truthful terminal state, commit coherent GREEN/verification checkpoints, and push only the active issue branch. No PR or review.

Loaded for this cycle: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-concurrency`, `golang-context`, `golang-code-style`, `golang-naming`, `golang-documentation`, `golang-spf13-cobra`, and `golang-lint`.
