# Plan: GitLab parity wave02 r1 (#78)

Parent issue: #78
Subissues: #83, #84, #85, #86, #87, #88, #89
Branch: `fm/cli-gitlab-parity-wave02-r1`
Scope: connector-local GitLab definition bundle, GitLab-owned fixtures/docs/metadata, planning artifacts, and GitHub issue policy addendum.

## GSD command path

- `scripts/gsd doctor` — pass.
- `scripts/gsd list` — pass; 69 commands available.
- `scripts/gsd prompt plan-phase 78 --skip-research --tdd` — rendered to `traces/gsd-plan-phase-78.prompt.md`.
- `scripts/gsd prompt gsd-quick --full "Implement connector-local GitLab official API operation ledger and documented parity scaffolding for parent issue #78 without live credentials"` — rendered to `traces/gsd-quick-full.prompt.md`.
- `scripts/gsd prompt programming-loop init --phase gitlab-parity-wave02-r1 --dry-run` — unavailable in this adapter (`unknown GSD command: programming-loop`). Manual GSD fallback is recorded here; follow `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`, keep TDD/verification evidence in this phase, and do not skip test-first validation.

## Required skills and references loaded

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
- Repo references: `AGENTS.md`, `required-skills-routing.md`, `gsd-pi-adapter.md`, `issue-agent-contract.md`, `parent-orchestrator-contract.md`, `parent-issue-orchestration-loop.md`, `cli-help-docs-website-parity.md`, `gsd-universal-runtime-loop.md`, `docs/migration/HANDOFF-CODEX.md`, `docs/migration/conventions.md`, `docs/architecture/connector-architecture-v2-design.md`, `.planning/{config,PROJECT,ROADMAP,STATE}.md/json`, `docs/plans/universal-programming-loop-prd.md`, and `docs/prompts/universal-programming-loop-prompts.md`.

## Mission

Replace the legacy 11-row GitLab surface snapshot with a complete official GitLab OpenAPI v2 operation ledger from the pinned source commit `9cd04099eb59d87335798e4f57a2bc5a2622e4cc`. Preserve the four existing fixture-backed streams, add connector-local typed operation/CLI/certification metadata, and record the captain policy correction that DELETE/destructive/admin operations are in scope when modeled with typed destructive confirmation plus the existing plan -> preview -> explicit approval -> execute path.

## Constraints

- No live GitLab connector execution, credentials, writes, certification, VPS/Thaalam work, pushes, PRs, merges, or shared daemon restarts.
- Keep provider behavior under `internal/connectors/defs/gitlab/**`; do not edit shared engine/runtime/CLI/foundation files or other connectors.
- If a shared foundation is absent, record the dependency and keep connector-local rows blocked/planned rather than claiming unsupported execution.
- Do not expose generic HTTP method/path/body, arbitrary GraphQL, shell, file, SQL write/read, extension, or raw passthrough operations.
- No secret-looking literals in docs, fixtures, operations, or issue comments.
- Use `gh-axi` for GitHub issue body updates; preserve existing bodies and counts.

## Implementation slices

1. **Issue policy addendum** — append an idempotent addendum to #78 and #83-#89 stating destructive/delete/admin operations remain in scope with typed confirmation and plan -> preview -> explicit approval -> execute; preserve existing body text and count tables.
2. **Official inventory generation** — fetch/read the pinned OpenAPI YAML, dedupe official operations by `(method,path)`, classify rows into the parent lanes (ETL/read, reverse ETL write, direct/provider search/query, binary/file, CDC/changefeed, excluded/not-applicable), and generate deterministic operation evidence matching the parent total of 1,146.
3. **Connector-local bundle** — update `metadata.json`, `api_surface.json`, `operations.json`, `cli_surface.json`, `certification.json`, and `docs.md`; keep existing streams/schemas/fixtures unless validation requires connector-local fixture adjustments.
4. **Safety/documentation** — document lane counts, destructive/admin inclusion, typed confirmation requirements, shared foundation dependencies (#2985, #2986, #2987, #2988), no generic escape hatches, and fixture-only/uncertified status.
5. **Validation and commit** — run connector-targeted validation/conformance plus boundary/diff checks as feasible, update this phase's TDD/verification artifacts with exact outcomes, and commit the connector bundle. Stop after commit.

## Shared dependencies recorded

- #2985 — provider search/query foundation; direct/provider-query rows remain blocked/planned unless already stream-backed.
- #2986 and #2988 — CDC truth/lab foundations; GitLab changefeed/audit/webhook rows remain planned/blocked without live certification.
- #2987 — binary/direct-file transfer safety foundation; binary/file rows remain planned/blocked unless a future named bounded command proves execution.
- Existing reverse ETL supports plan -> preview -> approval -> execute and `confirm: "destructive"`, but this wave does not generate hundreds of executable write actions without per-action fixtures and runner evidence; write rows remain planned/blocked unless explicitly covered by a named action.

## Spawn decision

- Cycle `plan`: `local_critical_path` — firstmate launched this worker in an isolated disposable worktree for one connector scope; mutating sub-workers would share the same connector directory and risk generated-file collisions. Read-only recon was performed inline with deterministic scripts.

## Verification plan

```bash
python3 .planning/phases/issue-78-gitlab-parity-wave02-r1/traces/gitlab_inventory_check.py
python3 .planning/phases/issue-78-gitlab-parity-wave02-r1/traces/gitlab_surface_counts.py
go run ./cmd/connectorgen validate internal/connectors/defs --json
go test ./internal/connectors/conformance -run 'TestConformance/gitlab' -count=1
go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1 -timeout=10m
go vet ./internal/connectors/... ./internal/cli/...
go build ./cmd/pm
make connector-boundary
git diff --check
```

Broader `go test ./...` and `make verify` are reserved for firstmate/no-mistakes after this worker is resumed for shipping, unless targeted validation exposes connector-local defects that need immediate repair.
