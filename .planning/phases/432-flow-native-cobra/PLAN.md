# Phase 432 Plan — Flow native Cobra namespace

Issue: polymetrics-ai/cli#432
Umbrella: #407
Parent: #397 / draft parent PR #438
Branch: `refactor/432-flow-native-cobra`
Base branch: `feat/cli-architecture-v2`
Exact starting HEAD: `ec12c1729e0aaf233a853eff5c6291885f910b15`
Invocation session: `issue-432-pi-openai-codex-gpt-5.6-sol-high-20260719T034344Z`
Explicit invocation profile: `model=openai-codex/gpt-5.6-sol`, `thinking=high`
Execution decision: `local_critical_path` — #432 is the assigned next serialized Phase 9 unit in an isolated worktree. Its central router scope collides with sibling migrations, this runtime exposes no subagent tool, and the user explicitly bounded delivery to implementation/commit/push with no PR or review.

## GSD route

- `scripts/gsd doctor` and `scripts/gsd list`: passed (69 commands).
- `scripts/gsd prompt plan-phase 432 --skip-research`: generated and is executed inline.
- `scripts/gsd prompt programming-loop init --phase 432 --dry-run`: unavailable because the adapter registry has no `programming-loop` command.
- Manual fallback: `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`, with six issue-local artifacts and strict RED → GREEN → refactor evidence.

## Required reading and skills

Read issues #432/#407/#397; `AGENTS.md`; GSD adapter/core/manual universal loop; issue and parent-orchestrator contracts; CLI help/docs/website parity; architecture plan §5/§9; Stage 9 execution prompt; ADR-0002; existing flow CLI, manifest/DAG/engine/action/checkpoint/events/telemetry/ledger tests; event and telemetry packages; generated flow manual and website references; and adjacent native ETL/reverse Cobra patterns.

Loaded: `gsd-core`; `golang-how-to`; `golang-cli`; `golang-testing`; `golang-error-handling`; `golang-security`; `golang-safety`; `golang-context`; `golang-concurrency`; `golang-documentation`; `golang-spf13-cobra`.

## Scope

- Replace only the `flow` legacy wrapper with a native Cobra subtree for `plan`, `preview`, `run`, `status`, `list`, and hidden positional `help` compatibility.
- Declare the complete existing local flag surface with native pflags while preserving current repeated, bare, assigned, ignored-unknown, malformed, and operand behavior where applicable: `--file`, `--force`, and `--flows-dir`, plus invocation-global `--root`, `--json`, `--plain`, `--no-input`, and `--progress` behavior.
- Adapt only flow handlers to typed values and remove only flow's bespoke parser/dispatcher. Dynamic connector dispatch remains on `parseFlags`; other namespaces remain untouched.
- Preserve exact manifest parsing and relative RLM spec resolution, DAG order, flow directory defaults, named-run resolution, checkpoint location/resume/force behavior, operation ledger behavior, event order/redaction, telemetry spans/metrics/redaction, cancellation propagation, exit taxonomy, stdout/stderr discipline, JSON envelopes, and deterministic list/plan/status/run output.
- Use only temporary flow manifests, temporary project roots, in-memory fakes, and existing dependency-free local paths.

Excluded: other namespaces; connector bundles; dynamic connector parser; new dependencies; services; credentials; external HTTP/SQL writes; reverse ETL execution; Phase 10 flow dashboards; Phase 11 flow-create wizard; Phase 19 focused help/man churn; PR/review.

## TDD slices and checkpoints

1. **Planning checkpoint** — commit/push these six artifacts before test or production edits.
2. **RED checkpoint** — focused tests specify:
   - native flow ownership for plan/preview/run/list/status/help, native flag type/NoOpt/unknown tolerance/completion seams, and absence from legacy wrappers;
   - every flag and operand form, repeated last-value behavior, bare `--force`, global assigned booleans, strict first positional ownership, and unchanged flow directory defaults;
   - bare/text/JSON/long/short/positional help; trailing help; literal `--`; malformed/unknown flags; invalid actions; no later-action discovery;
   - plan/preview/run/list/status text and JSON output, exact usage/validation/runtime categories, one-envelope output, and deterministic ordering;
   - cancellation propagation, sanitized deterministic lifecycle events, telemetry span/metric preservation and redaction, checkpoint resume/force semantics, and ledger order/status using temp state and fakes only.
   Capture focused failure before any production edit; commit/push tests.
3. **GREEN checkpoint** — add the smallest native flow command and typed handlers; add only flow-specific normalization/private operand state needed for exact compatibility; remove flow from `cobraLegacyCommands`; delete only `runFlow`/`parseFlowFlags` once unused.
4. **Refactor/parity checkpoint** — run focused/repeated/race flow/router/golden/full CLI and flow-package tests, exact-start parser/output differential where useful, runtime help, generated docs/website checks, formatting, vet, build, and scope/dependency guards.
5. **Final checkpoint** — run the established full `make verify`; finalize all six artifacts; commit/push; no PR or review.

## CLI parity stance

Public command names, flags, manual bytes, output schemas, docs, website content, generated artifacts, and golden fixtures should remain unchanged. Checked-in docs/website/golden edits are not applicable unless verification finds a real mismatch. Phase 10 owns dashboards, Phase 11 owns `flow create`, and Phase 19 owns deliberate focused-help/man churn. Verify `pm help flow`, bare `pm flow`, `pm flow --help`, short/positional/JSON manual routes, invalid actions, `docs/cli/flow.md`, website CLI-reference/architecture pages, generated docs, completion discovery, and golden transcripts.

## Safety

No secrets or approval values, external connectors, credentialed checks, optional services, dependencies, unrestricted writes, destructive/admin actions, reverse ETL, or production deploys. Tests use temp manifests and temp roots only. Action-step execution is excluded; no generic HTTP/SQL write path is exercised or added. Context cancellation must remain propagated, event/telemetry values sanitized, and all state effects bounded to temporary directories.
