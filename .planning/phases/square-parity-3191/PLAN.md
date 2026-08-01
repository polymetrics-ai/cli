# Square parity #3191 — GSD plan

## Scope

Final-wave Square connector-local parity for parent #3191 and children #3192-#3198. Stay inside `internal/connectors/defs/square/**`, Square-owned generated surfaces, dedicated Square fixtures/docs, and GSD phase artifacts. No live Square calls, no credentials, no provider writes, no new dependencies, no shared runtime edits, no PR/push/no-mistakes pipeline.

## GSD command path and skills

- Ran `scripts/gsd doctor` successfully.
- Attempted required `scripts/gsd prompt programming-loop init --phase square-parity-3191 --dry-run`; adapter returned `unknown GSD command: programming-loop`.
- Manual GSD fallback: following `.pi/prompts/pm-gsd-loop.md` and `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md` in this phase artifact.
- Required skills loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-documentation`; CLI/help parity reference loaded.

## Plan slices

1. **Inventory/red validation**
   - Fetch Square official OpenAPI from developer.squareup.com docs endpoints only.
   - Enumerate all 346 operations with method/path/operationId/tag/source URL and classify into stream/write/direct/binary/changefeed/blocked models.
   - Capture pre-change Square validation/conformance counts and failing parity gaps.
2. **Definition generation**
   - Update Square metadata/spec for full-surface defaults and write capability.
   - Generate `streams.json`, schemas, stream fixtures for implemented fixture-backed read/search/retrieve operations that the engine can execute.
   - Generate `writes.json` and write fixtures for typed reverse-ETL operations with closed schemas, path-field redaction for destructive/sensitive actions, and idempotent delete semantics where available.
   - Generate `operations.json` + `cli_surface.json` for bounded direct reads and blocked binary/unsupported surfaces without generic request escape hatches.
   - Generate complete `api_surface.json` ledger: each official operation exactly once, covered or blocked with official source URL/reason.
   - Update Square docs/certification evidence truthfully.
3. **Validation and repair**
   - Run focused inventory/validation, connector conformance, CLI/golden/docs checks, build, connector-boundary, `git diff --check`, and `make verify`.
   - Repair connector-local failures only; shared runtime gaps become documented blockers.
4. **Final evidence**
   - Update parent #3191 and children #3192-#3198 once via `gh-axi issue comment` with truthful counts and verification evidence.
   - Commit locally on `fm/cli-square-parity-wave05-r1` and append final status with hash/counts/blockers.

## Safety boundaries

- No Square credentials, no live provider API calls, no live writes.
- Reverse ETL remains the existing plan → preview → approval → execute path; generated write actions only describe typed request shapes and fixture replay.
- No arbitrary method/path/body/raw query/shell/file/passthrough operation.
- Destructive/admin/elevated operations either have closed schemas + `confirm: destructive`/redaction/idempotent delete semantics or remain blocked with source evidence/shared-runtime dependency.
