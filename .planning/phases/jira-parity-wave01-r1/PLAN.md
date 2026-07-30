# Jira Connector Parity Wave 01 R1

Issue: #81
Subissues: #104, #105, #106, #107, #108, #109, #110
Branch: `fm/cli-jira-parity-wave01-r1`

## GSD command path

- `scripts/gsd doctor` passed in this worktree.
- `scripts/gsd prompt plan-phase jira-parity-wave01-r1 --skip-research` rendered the repo-local Pi planning prompt.
- `scripts/gsd prompt programming-loop init --phase jira-parity-wave01-r1 --dry-run` failed because this adapter registry does not expose a `programming-loop` command in this checkout. Manual GSD fallback is active using `.pi/prompts/pm-gsd-loop.md` and `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`.

## Required skills loaded

- `gsd-core`
- `golang-how-to`
- `golang-cli`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-documentation`

## Goal

Refresh Jira's connector-local declarative bundle from the official Atlassian Jira Cloud OpenAPI v3 source so every documented operation has one explicit ledger disposition and every newly executable operation stays definition-owned, typed, bounded, redacted, and safety-gated.

## Scope

Allowed production paths:

- `internal/connectors/defs/jira/**`
- Jira-owned fixtures under that directory.
- Jira connector-local generated/manual docs and command metadata (`api_surface.json`, `operations.json`, `cli_surface.json`, `writes.json`, `certification.json`, `docs.md`, schemas/fixtures).

No shared runtime, `cmd/`, non-Jira connector, dependency, credential, live provider, VPS/Thaalam, merge, or certification execution work.

## Captain policy correction

The Jira parent and all seven subissues now carry an idempotent addendum stating that DELETE/destructive operations are in scope when implemented as named typed operations with typed `destructive` confirmation and the existing plan -> preview -> explicit approval -> execute path. The addendum does not change or fabricate implemented counts.

## Implementation slices

1. **Official inventory and red gate**
   - Fetch the official Jira OpenAPI source without credentials.
   - Record source hash and operation counts.
   - Capture a red validation asserting Jira currently lacks 616 documented operation rows and 296 reverse write actions.

2. **Operation ledger and executable surfaces**
   - Generate a complete `api_surface.json` with `operation_ledger_version: 1`.
   - Preserve existing stream coverage where valid and update only Jira-local stream metadata as needed.
   - Generate bounded `operations.json` for direct/provider search/read operations and blocked binary/download operations when no bounded binary executor exists.
   - Generate `cli_surface.json` command metadata with implemented bounded direct reads where the engine supports JSON direct reads; mark binary/unsupported shared-foundation gaps truthfully.

3. **Reverse ETL writes**
   - Generate Jira `writes.json` for non-deprecated write/mutation operations expressible by the existing declarative engine.
   - DELETE actions must use `kind: "delete"`, idempotent 404 semantics where safe, and `confirm: "destructive"`.
   - Other high-risk/destructive summaries also get `confirm: "destructive"` when risk text indicates irreversible/admin impact.
   - Unsupported raw-body/binary/multipart gaps are recorded as blocked operation rows, not faked.

4. **Fixtures, docs, and certification metadata**
   - Keep existing read fixtures sanitized and valid.
   - Add representative write fixtures for safe fixture replay; do not run provider calls.
   - Add certification metadata that remains fixture/replay-only and does not claim live certification.
   - Update `docs.md` with source hash, operation-ledger counts, destructive confirmation policy, and known shared-foundation blockers.

5. **Verification and evidence**
   - Validate JSON and connector loader.
   - Run Jira conformance and focused CLI/command metadata checks where feasible.
   - Record skipped/blocked certification truthfully; no live credentials or certification claims.

## Orchestration decision

`local_critical_path`: the task is a single connector-local Jira bundle in an isolated worktree. Mutating subworkers would collide on the same Jira files; read-only sidecars are optional but not required for this slice.

## Human gates / blockers

- New dependencies: human gate, none planned.
- Shared runtime changes: out of scope.
- Live credentials/provider calls/writes/certification: prohibited.
- If a documented operation requires a shared executor not present in the current engine (for example bounded binary download or raw arbitrary JSON scalar body write), record the exact dependency and keep the connector-local ledger honest instead of claiming implemented parity for that operation.
