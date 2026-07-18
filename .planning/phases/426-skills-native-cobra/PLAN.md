# Phase 426 Plan — Skills native Cobra namespace

Issue: polymetrics-ai/cli#426
Umbrella: #407
Parent: #397 / draft parent PR #438
Branch: `refactor/426-skills-native-cobra`
Base branch: `feat/cli-architecture-v2`
Exact starting parent HEAD: `54bfcbab5a997c81676b286fe78de00a109f5fba`
Invocation session: `issue-426-pi-openai-codex-gpt-5.6-sol-high-20260718T104457Z`
Explicit invocation profile: `model=openai-codex/gpt-5.6-sol`, `thinking=high`
Execution decision: `local_critical_path` — sixth serialized Phase 9 namespace unit is assigned to this isolated branch/worktree. Central router files collide with later units, and the user explicitly bounded this run to #426 with no PR or external review.

## Required reading complete

- Issue #426 via `gh`; parent #397; umbrella #407; adjacent #424 runtime and #425 version implementation/artifact patterns; draft parent PR #438.
- `AGENTS.md`; issue-agent and parent-orchestrator contracts; parent orchestration loop; GSD universal runtime/manual programming-loop fallback.
- `.planning/config.json`, `PROJECT.md`, `ROADMAP.md`, `STATE.md`; universal programming-loop PRD and prompt library.
- Required-skills routing, GSD Pi adapter, CLI help/docs/website parity.
- CLI Architecture v2 improvement plan §5/§9; execution prompt Stage 9; ADR 0002.
- Current skills handler/generator, Cobra router, global/config parsing, focused/agentic/golden tests, embedded help, checked-in CLI manual, and website references.

## GSD adapter and fallback

- `scripts/gsd doctor` — pass.
- `scripts/gsd list` — pass; 69 commands listed.
- `scripts/gsd prompt plan-phase 426 --skip-research` — prompt generated successfully.
- Required programming-loop probe: `scripts/gsd prompt programming-loop init --phase 426 --dry-run` — unavailable with `scripts/gsd: unknown GSD command: programming-loop` (exit 1).
- Manual fallback: execute `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md` inline with strict RED → GREEN → refactor evidence and all six issue-local artifacts. The adapter is healthy; only the absent programming-loop command requires fallback.

## Required skills loaded

- `gsd-core`.
- `golang-how-to` first, then `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-documentation`, `golang-spf13-cobra`, and `golang-security`.
- Applied constraints: fresh Cobra trees; `RunE`; declared pflags and argument nodes; injected writers; stable exit taxonomy/stdout/stderr; observable table tests; canonical help; untrusted path/argument handling; no secrets/dependencies/external I/O.

## Scope and exclusions

Allowed:

- Register native top-level `skills` namespace, native `generate` action and declared `--dir` flag, plus positional `skills help` compatibility.
- Minimal skills handler adaptation from `parseFlags` to typed values while preserving generator/path/security behavior and output bytes.
- Skills-focused/router tests and issue-local phase artifacts.
- Directly applicable help/manual/website/generated artifacts only if behavior intentionally changes.

Excluded:

- Other namespace migrations; dynamic connector parsing; connector bundles; generator content changes; path-policy changes; completion implementation; help-tree/man-page churn; dependencies; services; credentials; ETL/reverse ETL; shared parent artifacts; PR creation or external review.

## Existing contract to preserve

- Bare `pm skills`, `pm help skills`, `pm skills --help`, `pm skills -h`, and positional `pm skills help` render the canonical manual and exit 0; JSON variants emit `CommandManual/skills`.
- Only action is `generate`; invalid actions remain usage exit 2. Missing/empty `--dir` remains validation exit 3.
- `--dir value`, `--dir=value`, repeated `--dir` (last wins), and bare `--dir` (`true`) preserve legacy behavior, including comma-containing/path values without splitting.
- Unknown flags remain tolerated for `generate`; their values and extra positional arguments do not alter declared values.
- Global `--json` and `--root` remain accepted before/after the namespace in spaced/assigned forms; `--json=true|false` remains connected to config precedence, including config-file JSON overridden by assigned false.
- Skill generation remains metadata-only; generated filenames remain fixed internal names, and existing filesystem errors/output are unchanged. No secret or credential resolution is introduced.

## TDD slices and checkpoints

1. **Planning checkpoint** — create PLAN/TDD-LEDGER/VERIFICATION/PROMPTS/RUN-STATE/SUMMARY before production edits; commit and push.
2. **RED tests first** — specify native registration, all action/flag forms, bare/text/JSON/positional help, unknown-flag compatibility, invalid action/error categories, config/global positioning/assignment forms, generated output/path safety, and fresh-tree completion seam. Capture exact RED before production edits; commit/push test checkpoint.
3. **Smallest GREEN** — remove `skills` from `cobraLegacyCommands`; add native namespace/generate/help nodes; declare `--dir` with `StringArrayVar` + `NoOptDefVal=true` and unknown-flag whitelist; normalize spaced `--dir`; adapt `runSkills` to typed dir and remove its sole `parseFlags` use.
4. **Refactor/parity** — run focused skills/router/golden/full CLI tests; verify built-binary text/JSON/help/error/global forms; generate CLI docs to temp and diff; run docs validation and website generator/diff; expect no checked-in help/docs/golden delta.
5. **Full gates/delivery** — `gofmt -w cmd internal`, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify`; scope/dependency/diff checks; finalize artifacts; commit coherent green and verification checkpoints; push branch without PR.

## CLI help/docs/website parity stance

This changes parser ownership only. Embedded help, `docs/cli/skills.md`, website skill references, generated docs data, completion/discovery names, output envelopes, and goldens should remain unchanged. Bare namespace help, all text/JSON help routes, invalid actions, generated temp docs diff, website generator diff, and golden tests provide proof. Completion implementation remains Phase 15; only no-file fallback metadata is tested here.

## Safety

No secrets, credentials, services, dependencies, generic write tools, destructive/admin actions, quality-gate reductions, external reviews, PR, or merge. Tests write only to `t.TempDir`; no broad or traversal paths are introduced. The required `make verify` may run its existing local sample reverse-ETL smoke only under plan → preview → approval → execute and without external writes.

## Completion note

The bounded slice completed exactly within scope. Native Cobra now owns `skills`, `skills generate`, positional help, and typed/repeated/bare `--dir`; the skills legacy wrapper and handler `parseFlags` call are gone. Exact RED preceded production edits. Focused/router/golden/full CLI and repository gates passed, including `make verify`; built-binary help/output/error/config/global forms, generated CLI docs, website generation, dependencies, connector definitions, and golden fixtures remained in parity. Coherent plan, RED, implementation, and verification checkpoints were pushed. No PR or external review was created.
