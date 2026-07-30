# Plan: HubSpot parity wave01 r1 (#132)

Parent issue: #132
Branch: `fm/cli-hubspot-parity-wave01-r1`
Scope: connector-local HubSpot definition bundle and GitHub issue policy addendum.

## GSD command path

- `scripts/gsd doctor` — pass.
- `scripts/gsd list` — pass; 69 commands available.
- `scripts/gsd prompt plan-phase 132 --skip-research --tdd` — rendered to `traces/gsd-plan-phase-132.prompt.md`.
- `scripts/gsd prompt gsd-quick --full "Implement connector-local HubSpot official API operation ledger and documented parity scaffolding for issue #132 without live credentials"` — rendered to `traces/gsd-quick-full.prompt.md`.
- `scripts/gsd prompt programming-loop init --phase issue-132 --dry-run` — unavailable in this adapter (`unknown GSD command: programming-loop`). Manual GSD fallback is recorded here; keep TDD/verification evidence in this phase.

## Required skills and references loaded

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
- Repo references: `required-skills-routing.md`, `gsd-pi-adapter.md`, `cli-help-docs-website-parity.md`, `docs/migration/HANDOFF-CODEX.md`, `docs/migration/conventions.md`, `docs/architecture/connector-architecture-v2-design.md`.

## Mission

Create `internal/connectors/defs/hubspot/` from the official HubSpot public OpenAPI spec collection at commit `2bebde2dca45eaa1792931089c4e441c8e377594` and preserve parent issue counts without fabricating implementation. The captain policy correction is part of the mission: DELETE/destructive operations are in scope when modeled with typed `destructive` confirmation and the plan -> preview -> explicit approval -> execute path; they must not be blanket-excluded as unsafe.

## Constraints

- No live HubSpot provider calls, credentials, writes, certification, VPS/Thaalam work, merges, or pushes to `main`.
- Keep provider behavior connector-owned under `internal/connectors/defs/hubspot/**`.
- Do not edit shared runtime/engine/CLI code. If a shared foundation is absent, record the dependency and keep the connector-local ledger truthful.
- Use `gh-axi` for GitHub issue updates.
- No secret-looking literals in docs or fixtures.

## Implementation slices

1. **Issue policy addendum** — append an idempotent addendum to #132 and subissues #134-#140 stating destructive/delete operations are included with typed confirmation and not blanket-excluded; preserve existing bodies and counts.
2. **Official inventory generation** — fetch/read the HubSpot spec collection, dedupe official operations by `(method, path)`, and generate a 3,118-row operation ledger matching parent total. Preserve deterministic source file and operation evidence.
3. **Connector-local bundle** — add metadata/spec/docs/empty streams plus `api_surface.json`, `operations.json`, and `cli_surface.json`. Use blocked operation rows for not-yet-executable operations and record open shared dependencies (#2985, #2986, #2988) without claiming completion.
4. **Safety/documentation** — document lane counts, delete/destructive inclusion, planned typed confirmation, no generic raw API/shell/file escape hatches, and fixture-only/unverified status.
5. **Validation** — run connector-targeted validation and conformance; run boundary/diff checks as feasible.

## Shared dependencies recorded

- #2985 — provider search/query foundation is open; direct/provider-query operations remain blocked/planned, not executable.
- #2986 and #2988 — CDC truth/lab foundations are open; HubSpot has no counted CDC operations in the parent lane, and no CDC capability is claimed.
- Destructive typed confirmation foundation exists in the current codebase via the merged GitHub parity correction, but each HubSpot destructive operation still needs a connector-owned named action before it can be executable.

## Verification plan

```bash
go run ./cmd/connectorgen validate internal/connectors/defs/hubspot
go test ./internal/connectors/conformance -run 'TestConformance/hubspot' -count=1
go vet ./internal/connectors/... ./internal/cli/...
go build ./cmd/pm
make connector-boundary
git diff --check
```

Broader `go test ./...` / `make verify` are left for no-mistakes unless targeted validation exposes connector-local defects.

## Implementation result

- Added `internal/connectors/defs/hubspot/` with `metadata.json`, `spec.json`, `streams.json`, `api_surface.json`, `operations.json`, `cli_surface.json`, and `docs.md`.
- Official source inventory: 524 OpenAPI files, 4,466 versioned operation entries, 3,118 unique `(method,path)` operations.
- Generated rows: 3,118 API-surface rows, 3,118 typed operation metadata rows, 3,118 planned fixed-target command metadata rows.
- Lane counts match #132: ETL/read 925, reverse ETL write 1,704, direct/provider-search/query 260, binary/file 229, CDC 0.
- DELETE/destructive operations are included: 315 DELETE rows and 522 operations marked destructive in `operations.json`; all remain blocked/planned until named actions add `confirm: "destructive"` and plan -> preview -> explicit approval -> execute evidence.
- No runtime capability is claimed (`read=false`, `write=false`, `query=false`, `cdc=false`) because this wave is documented ledger/metadata only.
