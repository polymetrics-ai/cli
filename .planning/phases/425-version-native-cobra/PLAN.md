# Phase 425 Plan — Version native Cobra namespace

Issue: polymetrics-ai/cli#425
Umbrella: #407
Parent: #397 / draft parent PR #438
Branch: `refactor/425-version-native-cobra`
Base branch: `feat/cli-architecture-v2`
Exact starting/base HEAD: `479a62f930e7c8a9a51ba0b3deb088bf3aad3ecc`
Invocation session: `issue-425-pi-openai-codex-gpt-5.6-sol-high-20260718T095316Z`
Explicit invocation profile: `model=openai-codex/gpt-5.6-sol`, `thinking=high`
Execution decision: `local_critical_path` — fifth serialized Phase 9 namespace unit is already assigned to this isolated worktree/branch. Central router write scopes collide with later namespace units, and the user limited this run to #425 with no PR or external review request.

## Required reading complete

- Issue #425 via `gh`; parent #397; Phase 9 umbrella #407; adjacent completed namespace issues/patterns #423 and #424; draft parent PR #438.
- `AGENTS.md`; issue-agent and parent-orchestrator contracts; parent/subissue workflow; worker handoff template; GSD universal runtime loop.
- `.planning/config.json`, `PROJECT.md`, `ROADMAP.md`, `STATE.md`; universal programming-loop PRD and prompt library.
- Required skills routing, GSD Pi adapter, CLI help/docs/website parity policy.
- CLI Architecture v2 improvement plan §5/§9; execution prompt Stage 9; ADR 0002.
- Current version router/handler/tests/help/goldens/manual and adjacent native Cobra implementations/fixes for perf/runtime.

## GSD adapter and fallback

- `scripts/gsd doctor` — pass.
- `scripts/gsd list` — pass; 69 commands listed.
- Explicit planning invocation: `scripts/gsd prompt plan-phase 425 --skip-research --model=openai-codex/gpt-5.6-sol --thinking=high` — prompt generated successfully.
- Required programming-loop probe: `scripts/gsd prompt programming-loop init --phase 425 --dry-run --model=openai-codex/gpt-5.6-sol --thinking=high` — blocked exactly by `scripts/gsd: unknown GSD command: programming-loop` (exit 1).
- Manual fallback: follow `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md` inline with strict red → green → refactor evidence and issue-local phase artifacts. The adapter itself is healthy; only the known absent programming-loop command requires fallback.

## Required skills loaded

- `gsd-core`.
- `golang-how-to` first, then `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-documentation`, `golang-spf13-cobra`, and `golang-security`.
- Applied constraints: fresh Cobra tree per invocation; `RunE`; Cobra argument validators; injected output writers; stable exit taxonomy/stdout/stderr; observable contract tests; concise canonical help; untrusted argument handling; no secrets/dependencies/external I/O.

## Scope and exclusions

Allowed:

- Register native top-level `version` Cobra command and positional `version help` compatibility alias.
- Minimal `runVersion` adaptation after native Cobra owns argument parsing.
- Focused version/router tests and issue-local phase artifacts.
- Directly applicable docs/help/golden/website outputs only if behavior intentionally changes.

Excluded:

- Other namespace migrations; connector dynamic dispatch/parser; connector bundles; completion implementation; help-tree deepening/man pages; dependencies; runtime services; credentials; ETL/reverse ETL; parent/shared orchestration artifacts; PR creation/review.

## Existing contract to preserve

- `cli.Run(args, stdout, stderr) int` remains unchanged and fresh-tree/re-entrant.
- `pm version` prints exactly deterministic plain metadata; `pm version --json` emits the established `Version` envelope.
- `pm help version`, `pm version --help`, `pm version -h`, and positional `pm version help` render the canonical manual; JSON help emits `CommandManual`.
- Unknown flags and invalid positional actions remain usage errors (version never used `parseFlags`; its legacy handler rejected every residual argument).
- Global `--json` remains accepted in any position because `parseGlobal` owns it.
- No version-specific repeated or bare-boolean local flags exist; those ADR conventions are not applicable. Unknown-flag compatibility means preserving rejection, not introducing tolerance.

## TDD slices

1. **Planning checkpoint**
   - Create PLAN, TDD-LEDGER, VERIFICATION, PROMPTS, RUN-STATE, SUMMARY before production edits.
   - Commit and push the planning-only checkpoint.

2. **RED tests first**
   - Router registration: `version` must be native (`DisableFlagParsing=false`) and no longer appear in legacy wrappers.
   - Bare output/help: deterministic bare plain output; `--help`, `-h`, and positional `help` match `pm help version`.
   - JSON: deterministic `Version` output and JSON `CommandManual` for flag/positional help.
   - Compatibility: unknown flag and invalid action exit 2 with usage classification and never render a manual.
   - Run focused tests and capture exact RED before production edits; commit/push the red checkpoint if coherent.

3. **Smallest GREEN implementation**
   - Remove `version` from `cobraLegacyCommands`.
   - Register `newVersionCobraCommand` from `newRootCmd` with native Cobra parsing, `cobra.NoArgs`, canonical help/usage, and `RunE` delegating to version output.
   - Add a hidden native `version help` alias to preserve positional help and JSON manual behavior.
   - Remove the now-unused version handler argument check/signature.

4. **Refactor/parity**
   - Keep output/help bytes and golden fixtures unchanged.
   - Run version/router/golden and full `internal/cli` tests.
   - Build `pm`; compare `pm help version`, bare `pm version`, `pm version --help`, positional help, JSON output/manual, unknown flag, and invalid action.
   - Generate CLI docs to a temp directory and diff `docs/cli`; run docs validation and website docs generation, then prove no tracked docs/website/generated delta.

5. **Full gates and delivery**
   - `gofmt -w cmd internal`; `go vet ./...`; `go test ./...`; `go build ./cmd/pm`; `make verify`.
   - Safety/scope checks: no `go.mod`/`go.sum`, connector-def, unrelated namespace, docs, website, or golden delta unless explicitly reviewed.
   - Update all phase artifacts truthfully, commit coherent green/verification checkpoints, push this branch, and create no PR.

## CLI help/docs/website parity stance

Parser ownership changes only. Runtime help text, plain/JSON outputs, checked-in `docs/cli/version.md`, website docs, generated docs data, completion/discovery metadata, and golden fixtures are expected unchanged. Each unchanged surface will be marked N/A with proof from byte comparisons, docs generation diff, website generator/diff, and focused/golden tests. Bare `pm version` is an operational leaf command, so it must continue printing version metadata rather than namespace help; contextual help is covered by all help aliases.

## Safety

No secrets, credentials, services, reverse ETL, dependencies, generic write tools, destructive/admin actions, quality-gate reductions, external review requests, PR, or merge. Commands use only local source/tests/build/docs generation and non-credentialed GitHub metadata reads.
