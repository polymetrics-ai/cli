# Phase 428 Plan — Agent native Cobra namespace

Issue: polymetrics-ai/cli#428
Umbrella: #407
Parent: #397 / draft parent PR #438
Branch: `refactor/428-agent-native-cobra`
Base branch: `feat/cli-architecture-v2`
Exact starting parent HEAD: `235233f7cfde4a24612be6b0f95fb37a412d388a`
Invocation session: `issue-428-pi-openai-codex-gpt-5.6-sol-high-20260718T124925Z`
Explicit invocation profile: `model=openai-codex/gpt-5.6-sol`, `thinking=high`
Execution decision: `local_critical_path` — eighth serialized Phase 9 namespace unit is assigned to this isolated branch/worktree. Central router files collide with later units, the runtime exposes no subagent tool, and the user bounded this run to #428 with no PR or external review.

## Required reading complete

- Issue #428 via `gh`; parent #397; umbrella #407; adjacent #426/#427 native Cobra implementation and artifact patterns.
- `AGENTS.md`; issue-agent contract; worker handoff; GSD Pi adapter and universal/manual programming loop.
- `.planning/config.json`, `PROJECT.md`, `ROADMAP.md`, `STATE.md`; universal programming-loop PRD/prompt library.
- Required-skills routing; runtime/RLM integration reference; CLI help/docs/website parity.
- CLI Architecture v2 improvement plan §5/§9; execution prompt Stage 9; ADR 0002.
- Canonical runtime dependencies/setup, agent CLI manual, website architecture/CLI reference, current agent/image handlers, router/global parser, golden and adjacent focused tests.

## GSD adapter and fallback

- `scripts/gsd doctor` — pass.
- `scripts/gsd list` — pass; 69 commands listed.
- `scripts/gsd sources programming-loop` — unavailable because the registry has no `programming-loop` command.
- `scripts/gsd prompt plan-phase 428 --skip-research` — generated a 10668-byte prompt and executed inline.
- `scripts/gsd prompt programming-loop init --phase 428 --dry-run` — unavailable: `scripts/gsd: unknown GSD command: programming-loop` (exit 1).
- Manual fallback: execute `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md` inline with strict RED → GREEN → refactor evidence and all six issue-local artifacts.

## Required skills loaded

- `gsd-core`.
- `golang-how-to` first, then `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-context`, `golang-documentation`, and `golang-spf13-cobra`.
- Applied constraints: fresh Cobra trees; typed repeated pflags; `RunE`; stable exit/stdout/stderr contracts; context propagation; injected image-runtime fakes; control/path/image validation; no shell concatenation, secrets, dependencies, services, credentials, image pull/build, or worker behavior changes.

## Scope and exclusions

Allowed:

- Register native `agent`, `agent plan`, `agent image`, `agent image build|pull|ensure`, and positional `agent help` nodes.
- Declare current `plan --request` flag with legacy repeated/bare/unknown semantics; adapt only the agent handler away from `parseFlags`.
- Add an injected image-runtime seam so every image action is tested without Podman/Docker or external execution.
- Add bounded request, build-path, Podman-bin, and image-reference validation at the agent CLI boundary.
- Preserve agent-scoped trailing help and literal `--` behavior required by the legacy dispatcher.
- Focused agent/router tests and issue-local phase artifacts.

Excluded:

- Worker/RLM behavior; Temporal/Podman/Docker execution; image pulls/builds/publishing; credentials; other namespaces; dynamic connector parsing; connector bundles; dependencies; Phase 15 completion implementation; Phase 19 focused-help/man churn; shared parent artifacts; PR or external review.

## Existing contract to preserve

- Bare `pm agent`, `pm help agent`, `pm agent --help`, `pm agent -h`, and positional `pm agent help` return the canonical manual; JSON routes return `CommandManual/agent`.
- `plan` accepts spaced/assigned/repeated-last-wins/bare `--request`, unknown flags, extra positional values, and deterministic text/JSON output.
- `image` actions are `build`, `pull`, and `ensure`; build uses `<root>/build/agent/Containerfile`; pull/ensure use configured `rlm.image` and `rlm.podman_bin`; output kinds/status remain unchanged.
- Invalid actions remain usage errors. Global `--root`, `--json`, `--plain`, and `--no-input` work before/after the namespace in spaced/assigned forms, including configured JSON overridden by `--json=false`.
- Agent action-tail `--help`/`-h` remain ordinary ignored legacy inputs rather than Phase 19 focused help. Literal `--` does not stop later legacy flag parsing where the prior `parseFlags` continued.

## TDD slices and checkpoints

1. **Planning checkpoint** — create PLAN/TDD-LEDGER/VERIFICATION/PROMPTS/RUN-STATE/SUMMARY before production edits; commit and push.
2. **RED tests first** — specify native registration/tree/flags/completion seam; all plan forms and deterministic output; all image actions through injected fakes; ensure present/pull branches; bare/text/JSON/positional help; unknown/invalid/global assigned booleans; request/path/image validation; trailing help/literal separator compatibility. Capture exact RED before production edits; commit/push.
3. **Smallest GREEN** — remove `agent` from `cobraLegacyCommands`; add native nodes; typed `StringArray --request` with `NoOptDefVal=true`, unknown whitelist, no-file completion; agent-only pre-Cobra normalizer; typed plan handler; injected image runtime; bounded validation.
4. **Refactor/parity** — focused agent/router/golden/full CLI; base-vs-head differential matrix for preserved legacy cases; built binary help/plan/error/global checks only (never image actions); docs temp generation/diff, website generator/diff, runtime dependency-free tests.
5. **Full gates/delivery** — `gofmt -w cmd internal`, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify`; scope/dependency/diff guards; finalize artifacts; coherent commits/push, no PR.

## CLI help/docs/website parity stance

Parser ownership changes but command names, flags, output envelopes, canonical manuals, checked-in `docs/cli/agent.md`, website command map/examples, generated help, and goldens should remain unchanged. Runtime/bare/text/JSON/positional help, temporary docs generation diff, website generation diff, and golden tests prove parity. Focused action help remains intentionally deferred to Phase 19; completion implementation remains Phase 15.

## Safety

No secrets, credentials, services, dependencies, generic write tools, destructive/admin actions, production deployment, image pull/build, Podman/Docker invocation, quality-gate reduction, external review, PR, or merge. Image tests use injected fakes and temporary directories only. Worker/RLM execution is untouched. The required `make verify` may run only its existing dependency-free/local sample smoke under its established safety gates.

## Completion note

The bounded slice completed within scope. Native Cobra owns `agent`, `plan`, `image`, all three image actions, and positional help; `plan --request` is typed and only the agent handler's `parseFlags` call is removed. A context-aware injected runtime covers build/pull/ensure without executing Podman/Docker, while bounded request/root/Podman-bin/image validation runs before external lookup/execution. Agent-scoped trailing-help and literal-separator compatibility remains exact.

Focused RED preceded production edits. Focused/router/golden/full CLI and repository tests, a 25-case exact legacy differential, built-binary help/plan/error checks, runtime dependency-free tests, docs/website/generated parity, gofmt, vet, build, and `make verify` passed. No dependency, worker/RLM, connector, checked-in docs/website/golden, or unrelated namespace delta exists. Coherent checkpoints were pushed without PR or external review.
