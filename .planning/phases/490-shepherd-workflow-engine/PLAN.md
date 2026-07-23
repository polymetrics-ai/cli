# Plan: Shepherd bounded workflow-engine adoption (#490)

## Objective

Adopt the approved, exact `pi-workflow-engine@0.12.0` package for project-local bounded analysis and review while making Shepherd compatible with Pi `0.80.10` and removing raw AgentSession event order as success authority. Shepherd remains the durable controller and `ProductionAgentSessionPort` remains the production boundary.

Parent: #471. Branch: `refactor/490-shepherd-workflow-engine`. PR base: `feat/471-pi-agent-session-shepherd`. Parent PR: #472. Never merge or push to `main`.

## GSD and skills

- `scripts/gsd doctor`: pass (69 registered commands).
- Requested mandatory command: `scripts/gsd prompt programming-loop init --phase 490 --dry-run`.
- Result: unavailable once with `unknown GSD command: programming-loop`; no repeated retry. Manual-GSD fallback follows `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`.
- Loaded skill: `gsd-core`.
- Required references loaded: skill routing, runtime/RLM/Pi integration, GSD Pi adapter, universal runtime loop, and issue-agent contract.
- No Go or website task-specific skill applies. TypeScript/Pi behavior is governed by the issue, #471 phase contracts, and the package's local public documentation.

## Package provenance and adoption decision

- Approved dependency: `npm:pi-workflow-engine@0.12.0` (the user explicitly requested this exact adoption).
- Resolved local package: `.pi/npm/node_modules/pi-workflow-engine/package.json`, version `0.12.0`.
- Registry tarball: `https://registry.npmjs.org/pi-workflow-engine/-/pi-workflow-engine-0.12.0.tgz`.
- npm integrity: `sha512-DX+e2U03raK8o8YbwnDUcAQSKNZm0v1J6jWS+bk2j2kEFihLmZCf0sUlrHWou1kWC3Zw+CA4HCgqpjLWlmtcRg==`.
- Local package metadata says Pi `0.80.10` or newer, but its peer ranges are wildcards and its package root publishes no stable host SDK export for embedding.
- Decision: **partial adoption**. Do not deep-import engine internals and do not implement it as `ProductionAgentSessionPort`. Its documented API lacks the complete caller-claimed cwd/workspace receipt, per-session host capability binding, typed failure union, caller-owned cancellation operation, binding receipt, and typed abort/join receipt required by Shepherd.
- Use the engine only as trusted project-local analysis/review orchestration. `.pi/.workflow-runs` is non-authoritative private run state and never replaces Shepherd persistence or journals.
- Reuse the documented event-agnostic Pi AgentSession pattern behind Shepherd's existing port: await `prompt()`, derive authority from the terminating typed handoff, and use event callbacks only for bounded non-authoritative progress.

Analysis workflow evidence: `43dc4d81-66bf-4642-bed2-207a00d5fec0` (resumed from budget-exhausted `4cb6791a-828a-4cdc-aebc-281d2c415cf1`), four read-only lanes plus typed xhigh synthesis on `openai-codex/gpt-5.6-sol`.

## Architecture invariants

1. Shepherd owns issue/sub-issue state, dependency scheduling, collisions, persistent worktrees, branches, commits, pushes, PR publication, exact-head review, external-effect receipts/journal, recovery, retries, human gates, and the no-main-merge rule.
2. `ProductionAgentSessionPort` remains `run`/`abort`/optional `close`; production code does not import `pi-workflow-engine`.
3. The compatibility policy is explicit and bounded: `>=0.80.10 <0.80.11`, therefore only stable `0.80.10`; prereleases, mixed family versions, and future versions fail closed.
4. `prompt()` settlement plus one validated typed terminal handoff is success authority. Unknown/reordered raw events cannot create or revoke a valid handoff.
5. Claimed cwd, exact active tools, scoped capabilities, binding fields, cancellation, abort/wait/unsubscribe/dispose, and lease release remain fail-closed.
6. All 17 production-matrix rows remain green.

## Declared write scope

- `.pi/settings.json` (pre-existing issue setup: exact package registration only)
- `.pi/extensions/shepherd/agent-session-runtime.ts`
- `.pi/extensions/shepherd/agent-session-runtime.test.ts`
- `.pi/extensions/shepherd/pi-compatibility.ts` and `.test.ts` if a shared policy is needed
- `.pi/extensions/shepherd/sdk-runner.ts` and `.test.ts` only to consume the shared Pi policy
- `.pi/extensions/shepherd/index.ts` only if the injected policy constant changes
- `.pi/extensions/shepherd/role-prompts.ts` and `.test.ts` only if the typed terminal tool contract requires it
- `.pi/extensions/shepherd/tool-policy.ts` and `.test.ts` only if Pi 0.80.10 public tool types require a compatibility alignment
- `.github/scripts/verify-shepherd-pi-runtime.mjs`
- `.github/scripts/verify-shepherd-workflow-engine.mjs` and `verify-shepherd-offline-rpc.mjs` for review-disposition reproducibility
- `.github/workflows/shepherd.yml`
- `.gitignore` for workflow-engine's local non-authoritative run records
- `.pi/README.md`
- `.planning/phases/490-shepherd-workflow-engine/**`

Forbidden without a new human decision: controller/state/journal/recovery/Git/GitHub/decision behavior, Go/connector code, credentials, deployment, quality-gate reductions, parent/default-branch integration.

## TDD slices and checkpoints

### Slice 1 — planning/provenance

Create plan, TDD ledger, verification checklist, partial-adoption ADR, and run state. Commit and push the plan checkpoint.

### Slice 2 — RED compatibility and terminal authority

Before production edits, add focused executable tests for:

- Pi `0.80.10` acceptance and bounded rejection of `0.80.9`, `0.80.11`, prerelease, malformed, and mixed required/runtime versions;
- harmless new/unknown AgentSession events;
- unknown events not invalidating a successful typed terminal result;
- prompt completion without a typed terminal handoff failing closed;
- duplicate/malformed/binding-mismatched terminal handoffs failing closed;
- retained claimed cwd, exact scoped tools/capabilities, cancellation, abort/join, and terminal validation contracts.

Run focused tests and record assertion-level RED while production blobs remain unchanged. Commit/push the RED checkpoint if useful.

### Slice 3 — GREEN compatibility and event-agnostic completion

- Add one tested compatibility policy for stable Pi `0.80.10` only.
- Update the production runtime and legacy SDK runner to consume it.
- Await prompt settlement; collect only bounded non-authoritative progress; validate one typed terminal handoff; ignore harmless unknown events.
- Preserve active-tool/cwd/capability binding and cleanup ownership.
- Keep workflow-engine outside production imports.
- Update CI, docs, and partial-adoption ADR.

Run focused tests until green, then commit/push one coherent implementation slice.

### Slice 4 — one bounded review and disposition

After local implementation gates, freeze the exact head and run exactly one comprehensive workflow-engine/Codex `gpt-5.6-sol` xhigh code-review round. Disposition every actionable finding once. Apply only necessary fixes, run affected tests, then run the final full Shepherd gate once. Commit/push review fixes if any.

## Verification boundary

Follow `.planning/phases/471-pi-agent-session-shepherd/PROMPTS.md`: focused child tests, complete sequential Shepherd TypeScript suite, strict typecheck against exact Pi `0.80.10`, offline Shepherd RPC/canary, workflow-engine registration/co-load RPC, `git diff --check`, and exact changed-path scope. Do not run Go, connector, certification, runtime-service, or root `make verify` gates in this child worktree.
