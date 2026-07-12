# Plan: Gorgias CLI Parity Parent Orchestration

Parent issue: #196  
Parent branch: `feat/196-gorgias-cli-parity`  
Parent PR: https://github.com/polymetrics-ai/cli/pull/229 (draft)  
Default branch: `main`

## GSD command path

- Adapter health: `scripts/gsd doctor` — passed.
- Pi verification: `scripts/gsd verify-pi` — passed.
- Command inventory: `scripts/gsd list --json` — ran; output large/truncated by harness.
- Planning prompt: `scripts/gsd prompt plan-phase 196 --skip-research` — generated prompt and used for this parent plan.
- Execution prompt: `scripts/gsd prompt execute-phase issue-196-gorgias-cli-parity --dry-run` — generated prompt and used as phase-execution preflight.
- Required programming-loop attempt: `scripts/gsd prompt programming-loop init --phase issue-196-gorgias-cli-parity --dry-run` — blocked because this adapter registry has no `programming-loop` command. Manual GSD universal runtime fallback is active per `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md` and `.pi/prompts/pm-gsd-loop.md`.

## Required skills loaded

- `gsd-core`
- `caveman`
- `golang-how-to`
- `golang-cli`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-context`
- `golang-concurrency`
- `golang-documentation`
- `golang-lint`
- `golang-spf13-cobra` if CLI command-tree edits become necessary

## Scope

Coordinate parent issue #196 and sub-issues #197-#203 for Gorgias connector CLI parity. Preserve stacked-PR model: sub-issue branches target `feat/196-gorgias-cli-parity`; parent PR targets `main` and remains human-gated.

## Sub-issue dependency order

1. #197 — CLI surface metadata and authoritative source refresh. Critical path.
2. #200 — operation ledger, full official operation classification, depends on #197 baseline.
3. #199 — stream runner/read coverage, depends on #197/#200 classifications.
4. #201 — bounded direct reads, depends on #197/#200 and runner support.
5. #202 — provider-specific query/body and binary policies, depends on #197/#200; may unblock #201/#199 details.
6. #203 — sensitive/admin/write policy hardening, depends on #200 and write action inventory.
7. #198 — help/docs rendering parity, depends on metadata produced by #197 and command surfaces from #199/#201/#203.

## Slice boundaries

- Parent slice: planning, orchestration state, parent PR #229 seed, review routing, final integration readiness.
- #197 slice: `internal/connectors/defs/gorgias/metadata.json`, `api_surface.json`, optional `cli_surface.json`, and issue-specific planning/verification artifacts. No runtime command behavior beyond validation unless tests prove metadata loader gaps.
- Later slices own streams, direct reads, binary policy, write actions, help/docs rendering, and sensitive/admin safeguards.

## Execution decision

Current Pi session lacks the `subagent` tool, so no mutating worker can be spawned. Decision for this cycle: `local_critical_path` for #197 after parent PR seed; blocker for parallel spawn: `not_spawned_runtime_capability_missing`.

## Human gates

- Parent PR merge into `main`.
- New dependencies.
- Auth scope changes or `gh auth refresh`.
- Secrets, credentials, live credential checks, or credential printing.
- Destructive external actions or production deployment.
- Reverse ETL execution outside plan → preview → approval → execute.
- Generic shell, generic HTTP write, generic SQL write, raw generic GraphQL mutation, or unrestricted raw API tooling.
- Quality gate reductions.

## Verification plan

Parent planning checkpoint:

```bash
jq empty .planning/phases/issue-196-gorgias-cli-parity/RUN-STATE.json .planning/phases/issue-196-gorgias-cli-parity/ORCHESTRATION-STATE.json
scripts/gsd doctor
scripts/gsd verify-pi
```

Issue #197 targeted verification will be recorded in `.planning/phases/issue-197-gorgias-cli-surface-metadata/VERIFICATION.md` before production edits.
