# Issue 196 Gorgias parity plan

## Scope

Parent issue: #196. Subissues: #197-#203.

Allowed write scope for this worker:

- `internal/connectors/defs/gorgias/**`
- Gorgias-owned fixtures and connector-local CLI metadata in that directory
- Gorgias planning artifacts under `.planning/phases/issue-196-gorgias-parity/**`

No shared engine/runtime/foundation files, no other connectors, no live provider calls, no credentials, no push/PR/no-mistakes run.

## GSD command path

- `scripts/gsd doctor` passed.
- `scripts/gsd prompt programming-loop init --phase issue-196-gorgias-parity --dry-run` failed because this repo-local adapter currently exposes no `programming-loop` command in `scripts/gsd list`.
- Manual GSD fallback used by applying `.pi/prompts/pm-gsd-loop.md` and `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md` directly in Pi, with this plan, `TDD-LEDGER.md`, `VERIFICATION.md`, `SUMMARY.md`, and `RUN-STATE.json` as the phase artifacts.

## Required skills loaded

- `gsd-core`
- `golang-how-to`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-testing`
- `golang-cli`
- `golang-documentation`

## Required policy references read

- `AGENTS.md`
- `.agents/agentic-delivery/references/required-skills-routing.md`
- `.agents/agentic-delivery/contracts/issue-agent-contract.md`
- `.agents/agentic-delivery/contracts/parent-orchestrator-contract.md`
- `.agents/agentic-delivery/workflows/parent-issue-orchestration-loop.md`
- `.agents/agentic-delivery/workflows/stacked-parent-subissue-workflow.md`
- `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`
- `.agents/agentic-delivery/references/gsd-pi-adapter.md`
- `docs/migration/HANDOFF-CODEX.md`
- `docs/migration/conventions.md`
- `docs/architecture/connector-architecture-v2-design.md`
- `docs/plans/universal-programming-loop-prd.md`
- `docs/prompts/universal-programming-loop-prompts.md`

## Parent graph inventory

- Parent #196: Gorgias official API parity, official lane counts 38 ETL reads, 63 reverse writes, 3 direct/provider search, 6 binary/file, 4 CDC/changefeed, total 114.
- Children:
  - #197 ledger
  - #198 ETL/read + CDC
  - #199 direct/provider-search/query + binary
  - #200 reverse ETL typed writes
  - #201 CLI/config/help
  - #202 fixtures/docs/conformance/Connector Guard
  - #203 certification/release evidence

## Implementation status

Completed locally: official ledger captured, captain-policy addendum applied to #196-#203, connector-local Gorgias bundle expanded, fixtures/docs/CLI/certification metadata added, and verification evidence recorded. No live provider calls, credentials, push, PR, certification, or no-mistakes run were performed.

## Implementation slices

1. Official source inventory
   - Fetch `https://developers.gorgias.com/llms.txt` and linked `reference/*.md` OpenAPI snippets.
   - Record exact operation ledger in `traces/official-operations.json` with method, official `/api/...` path, operationId, summary, tag, source URL, lane.
   - Add an idempotent captain-policy addendum to #196-#203 with `gh-axi issue edit`, preserving existing bodies and count tables.

2. Red validation artifact before production bundle edits
   - Run `go run ./cmd/connectorgen validate internal/connectors/defs/gorgias` against the current bundle and record the current incomplete 11-row surface as the baseline.

3. Connector-local bundle expansion
   - Update `metadata.json` to advertise write support and risk text.
   - Expand `spec.json` with optional path-parameter config keys needed by connector-local ETL detail streams while preserving existing auth and `base_url` API-root behavior.
   - Expand `streams.json` to all official ETL/read + CDC GET operations in lane counts.
   - Add stream schemas and sanitized fixtures for every declared stream; keep two-page cursor fixture for `tickets`.
   - Add `writes.json` for all lane write operations with typed schemas, path fields, risk text, destructive `confirm` where applicable, and idempotent delete notes.
   - Add `operations.json` for implemented direct-read operations and blocked binary/unsafe-without-foundation operations.
   - Add `cli_surface.json` documenting connector-local ETL, reverse ETL, direct-read, binary/planned commands.
   - Replace `api_surface.json` with a 114-row operation-ledger-mode surface: covered streams/writes/direct reads or blocked operation rows.
   - Update `docs.md` and add `certification.json` truthfully marking fixture-only status.

4. Green validation and conformance
   - Run focused validation and conformance.
   - Run connector boundary and local compile gates that are in scope and feasible without live services.
   - Record blockers for shared foundations (#2985, #2986, #2988) instead of editing shared code.

5. Commit only
   - Commit the connector bundle and planning artifacts on `fm/cli-gorgias-parity-wave02-r1`.
   - Stop after commit; do not no-mistakes, push, or open PR.

## Orchestration decision

Cycle `plan`: `local_critical_path` because firstmate launched this crewmate in an isolated worktree with a single connector-owned write scope; additional mutating subworkers would collide on the same `internal/connectors/defs/gorgias/**` files. Read-only sidecars are optional, but the official source inventory is scriptable and deterministic.

## Safety notes

- No secrets or credentialed checks.
- No live Gorgias calls; docs fetches only.
- Reverse ETL execution is not run. The connector only declares typed actions that are executed by the existing plan -> preview -> approval -> execute path.
- DELETE/destructive/admin actions use connector-owned typed schemas plus `confirm: destructive`; blocked direct/binary/admin rows name the missing safe contract instead of claiming unsupported execution.
