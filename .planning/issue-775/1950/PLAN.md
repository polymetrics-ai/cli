# GSD Plan — Issue #1950 Lucid ELD Operation Ledger

## Scope

- Branch: `feat/1950-lucid-eld-operation-ledger`
- Parent issue/branch: #775 / `feat/775-lucid-eld-full-parity`
- Allowed production file: `internal/connectors/defs/lucid-eld/api_surface.json`
- Allowed evidence/planning: `.planning/issue-775/1950/**`
- Do not edit `.planning/issue-775/STATE.md`, other connector defs, shared Go, hooks, natives, docs, or website files.

## GSD mode

- `scripts/gsd doctor` passed.
- Required adapter command attempted: `scripts/gsd prompt programming-loop init --phase issue-775-1950-lucid-eld-operation-ledger --dry-run`.
- Result: adapter returned `scripts/gsd: unknown GSD command: programming-loop`; manual GSD fallback active per `gsd-pi-adapter.md` and issue-agent contract.
- Execution decision: `local_critical_path` (already-isolated worker context; no recursive subagents available/allowed).

## Required skills loaded

- `gsd-core`
- `caveman`
- Go: `golang-how-to`, `golang-cli`, `golang-spf13-cobra`, `golang-spf13-viper`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-documentation`
- Design/website: not applicable; no `website/**` or UI scope.

## Authoritative sources to reconcile

1. Official OpenAPI 2.0 JSON: `https://api.drivehos.app/partner/swagger/doc.json`
2. Official Swagger UI: `https://api.drivehos.app/partner/swagger`
3. WithTerminal provider page: `https://docs.withterminal.com/providers/tsp/lucid-eld`

Ground rule: OpenAPI JSON is wire-schema ground truth; WithTerminal names are capability hints only.

## Slice plan

### Slice 1 — Planning, official-source evidence, and red fixture

- Save public official OpenAPI JSON plus source summary under `.planning/issue-775/1950/evidence/`.
- Add planning-only validator/fixtures under `.planning/issue-775/1950/fixtures/`.
- Red evidence: intentionally incomplete fixture must fail because at least one known OpenAPI endpoint is absent.
- Commit checkpoint: planning/red fixture after `git diff --check`.

### Slice 2 — Endpoint ledger

- Create `internal/connectors/defs/lucid-eld/api_surface.json`.
- Include current `reviewed_at`, authoritative API/version, source URL, and exact 8 OpenAPI operations.
- Classify list/history endpoints as planned streams; singleton/latest endpoints as planned direct reads; no writes/binary unless official evidence adds mutations/media.
- No wildcard/raw passthrough paths; no duplicate method/path rows.

### Slice 3 — Negative fixtures and green planning validation

- Add negative fixtures for duplicate coverage, unknown targets, invalid exclusion category, wildcard path, and stale review metadata.
- Run planning validator against final ledger and negative fixtures.
- Run required repo gates, record exact outputs in `VERIFICATION.md`.

### Slice 4 — PR/handoff

- Commit green implementation.
- Push only `feat/1950-lucid-eld-operation-ledger`.
- Open non-draft sub-PR to `feat/775-lucid-eld-full-parity` with `Refs #1950` and `Refs #775`.
- Run `~/.local/bin/gh-axi pr checks <num>`.
- Do not request `@claude review`; record auto-review status.

## Acceptance checklist

- [x] Full body of #1950 and parent #775 read.
- [x] Required repo docs and connector architecture docs read.
- [x] Required skills recorded in TDD ledger and handoff.
- [x] Public OpenAPI JSON fetched directly and reconciled.
- [x] Every documented operation appears once in `api_surface.json`.
- [x] `reviewed_at` current.
- [x] No generic/raw HTTP, wildcard, shell, SQL, GraphQL escape hatch.
- [x] No credentials, real driver/vehicle/customer data, or authenticated calls.
- [x] Required verification run and exact results recorded.
