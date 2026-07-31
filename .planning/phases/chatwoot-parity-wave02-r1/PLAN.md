# PLAN — Chatwoot parity wave02 r1

## GSD path

- Required GSD adapter health: `scripts/gsd doctor` passed.
- Attempted required programming-loop command: `scripts/gsd prompt programming-loop init --phase issue-148-chatwoot-parity --dry-run` returned `unknown GSD command: programming-loop`.
- Fallback used: repo-local Pi programming-loop prompt `.pi/prompts/pm-gsd-loop.md` plus `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`; this is recorded here because the adapter registry lacks the documented alias while the Pi prompt exists.
- Execution decision cycle 1: `local_critical_path` — firstmate assigned one worker branch and one connector-owned write scope; no mutating subagents spawned into this checkout.

## Required skills loaded

- `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-documentation`, `golang-context`.
- Required routing/contracts loaded: required-skills routing, issue-agent contract, parent-orchestrator contract/workflow, stacked PR workflow, automated review routing, Claude review loop, CLI help/docs parity, GSD Pi adapter, connector migration handoff/conventions/design.

## Scope guard

Allowed production paths are Chatwoot-owned: `internal/connectors/defs/chatwoot/**`, `docs/connectors/chatwoot/**`, and phase artifacts under this directory. Do not edit shared runtime/foundation files, other connectors, generated hook/native sets, `go.mod`, or broad catalog artifacts.

## Source inventory

- Official OpenAPI sources fetched without credentials and indexed in context-mode. Processed operation inventory: `official-operations.json`.
- Counts from current official docs: total=145; by_group={'application': 115, 'client': 12, 'other': 1, 'platform': 17}; by_method={'GET': 61, 'PATCH': 22, 'POST': 42, 'DELETE': 18, 'PUT': 2}.
- Parent/subissue graph and preserved r2 count allocation: `issue-graph.json`.

## Implementation slices

1. Red ledger validation: prove current Chatwoot bundle is incomplete against official total (145) and still uses legacy `excluded` rows.
2. Connector-local source ledger: regenerate `api_surface.json` in operation-ledger mode with every official operation exactly once.
3. Bounded operation metadata: add `operations.json` for direct/provider search/report reads and planned write/read surfaces with fixed paths, bounded `max_bytes`, schemas, output policies, risk, and approval text.
4. Writes parity: expand `writes.json` for every safely expressible application/platform mutation through reverse ETL with typed record schemas; DELETE/destructive/admin mutations get `confirm: destructive`, idempotent delete notes where status 404 is safe, and risk text. Public/client or provider-absent operations that cannot safely execute with the existing base auth/config contract remain blocked in `api_surface.json`/`operations.json`.
5. CLI metadata parity: add `cli_surface.json` provider-style command metadata for implemented streams/writes and planned/blocked direct/admin/client surfaces without raw API escape hatches.
6. Fixture/docs/conformance: update Chatwoot fixtures/docs/manual Skill docs and certification metadata as connector-local evidence; no live provider calls or certification claims.
7. Verification: run targeted connectorgen/conformance/CLI docs checks, then local gates within time budget; record exact results in `VERIFICATION.md`.

## Safety decisions

- No credentials, live calls, live writes, live certification, VPS/Thaalam, push, PR, or no-mistakes pipeline.
- Reverse ETL execution is not run; only plan/preview metadata and dry-run/conformance fixtures are local.
- Destructive/admin operations are never represented as raw generic HTTP; executable rows must be named writes with path fields and `confirm: destructive`, or blocked with shared-foundation evidence.
