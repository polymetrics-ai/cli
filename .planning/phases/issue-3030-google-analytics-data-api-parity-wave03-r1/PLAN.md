# Plan: Google Analytics Data API parity wave03 r1 (#3030-#3037)

Parent issue: #3030
Subissues: #3031, #3032, #3033, #3034, #3035, #3036, #3037
Branch: `fm/cli-google-analytics-data-api-parity-wave03-r1`
Scope: `google-analytics-data-api` connector bundle/native/hook/tests/fixtures/docs/generated connector surfaces, plus minimal generic tooling/native wiring only where required by the task gates.

## GSD command path

- `scripts/gsd doctor` — pass.
- `scripts/gsd list` — pass; adapter lists 69 commands.
- `scripts/gsd prompt programming-loop init --phase issue-3030-google-analytics-data-api-parity-wave03-r1 --dry-run` — unavailable (`unknown GSD command: programming-loop`). Manual GSD fallback is active for this worker; keep TDD and verification evidence in this phase.
- `scripts/gsd prompt plan-phase issue-3030-google-analytics-data-api-parity-wave03-r1 --skip-research` — rendered and used as the planning workflow prompt.

## Required skills and references loaded

- Skills: `gsd-core`, `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-cli`, `golang-documentation`, `golang-context`, `golang-concurrency`.
- Repo references: `AGENTS.md`, `.agents/agentic-delivery/references/required-skills-routing.md`, `.agents/agentic-delivery/references/gsd-pi-adapter.md`, `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`, `.agents/agentic-delivery/contracts/issue-agent-contract.md`, `.agents/agentic-delivery/contracts/parent-orchestrator-contract.md`, `docs/migration/HANDOFF-CODEX.md`, `docs/migration/conventions.md`, `docs/architecture/connector-architecture-v2-design.md`.
- Issue bodies read with `gh-axi`: #3030-#3037.

## Official re-audit baseline

- Re-audited official discovery source `https://analyticsdata.googleapis.com/$discovery/rest?version=v1beta` on 2026-07-31: discovery revision `20260729`, 11 v1beta REST methods.
- Re-audited official reference index `https://developers.google.com/analytics/devguides/reporting/data/v1/rest`.
- The issue baseline named 10 operations at revision `20260728`; this worker will not copy that count. The post-change ledger will report the current audited 11-operation inventory and explain the delta.
- Local `cli-official-api-parity-audit-r2/audit.json` is absent in this worktree, so the official live discovery/reference audit and issue bodies are the available evidence.

## Current constraints and implementation policy

- No live GA provider calls, credentials, live writes, certification claims, VPS/Thaalam changes, pushes, or PR work.
- Fixture-only validation. Any HTTP tests use local `httptest` or connector fixture mode only.
- No new dependencies.
- No generic raw API, shell, SQL, or unbounded HTTP passthrough surfaces.
- Reverse ETL remains unsupported unless a GA mutation can be modeled as a typed named action with schema, redaction, approval, and fixture evidence.
- POST read-query endpoints are not reverse ETL writes, but shared surface rules still treat POST as a mutation for `capabilities.write=false`; where that prevents truthful executable advertisement without a shared foundation change, keep the operation blocked/planned with precise evidence instead of faking write capability.

## Implementation slices

1. **TDD/baseline gates**
   - Record the existing exact connector validate gate failure (`connectorgen validate internal/connectors/defs/google-analytics-data-api` treats `fixtures/` and `schemas/` as bundle dirs).
   - Add/adjust a focused test for validating a single bundle dir before changing `cmd/connectorgen` tooling, because the task requires that exact command.
2. **Official ledger and bundle metadata**
   - Replace legacy HOOK-only `api_surface.json` with a current official v1beta ledger of 11 rows.
   - Add `operations.json` for every official row with fixed method/path, max-bytes, auth scope, output policy, and blocked/executable truth.
   - Preserve `operation_ledger_version: 1`, source URLs, review timestamp, and count evidence.
3. **Executable read coverage**
   - Preserve the five existing GA report streams backed by native fixture/live logic as the connector's ETL/report stream surface for the official `runReport` operation.
   - Add sanitized fixtures for every declared stream (not only the first stream), plus connector-owned tests that assert fixture reads for all streams and report pagination/request body behavior against `httptest`.
   - If feasible without shared runtime changes, expose GET direct reads for metadata and audience-export metadata/list via the native connector; otherwise keep them blocked with the precise native-wrapper/CLI-surface dependency recorded.
4. **Typed mutations/direct queries**
   - Inventory POST report/query/check/audience-export operations individually.
   - Implement only those that the current connector contract can execute safely with typed closed schemas and fixture evidence.
   - Keep POST provider query/search operations blocked/planned when the shared provider-query surface #2985 or POST-as-write guard prevents truthful executable exposure.
5. **CLI/docs/generated surfaces**
   - Add/update `cli_surface.json` with implemented stream commands and planned fixed-target operation commands; no raw API escape hatch.
   - Regenerate/update `docs/connectors/google-analytics-data-api/{MANUAL.md,SKILL.md}`, `docs/connectors/README.md`, and `docs/connectors/catalog/{all-connectors.json,all-connectors.md}`.
   - Update CLI golden transcripts only if runtime command help output changes.
6. **Issue addendum**
   - Append an idempotent captain-policy addendum to #3030-#3037 with actual post-change counts, explicit fixture-only/no-live evidence, and no certification claims.
7. **Verification and commit**
   - Run the task-required local gates exactly where possible; record and fix connector-local/generic-tooling failures.
   - Commit a clean green slice; do not invoke `/no-mistakes`, push, open/update PRs, or merge.

## Final disposition before commit

- Official inventory: 11 current v1beta REST methods from discovery revision `20260729`.
- Executable official methods in this slice: 4 (`runReport` via five fixed report streams, `getMetadata`, `audienceExports.list`, `audienceExports.get`).
- Blocked/planned official methods: 7 (`runRealtimeReport`, `runPivotReport`, `batchRunReports`, `batchRunPivotReports`, `checkCompatibility`, `audienceExports.create`, `audienceExports.query`).
- Fixture/conformance evidence: 5 sanitized stream fixtures, 3 sanitized direct operation fixtures, native fixture/live-httptest coverage for every executable surface, GA conformance pass. No live provider calls and no certification claims.
- Accidental generator cleanup: broad docs/skills churn was reverted; final generated diffs are limited to GA connector docs, connector catalog/README entries, CLI golden root command entries, and website generated connector data. No `*_gen.go` diffs remain.

## Target verification commands

```bash
go run ./cmd/connectorgen validate internal/connectors/defs/google-analytics-data-api
go test ./internal/connectors/conformance -run 'TestConformance/google-analytics-data-api' -count=1
go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1
go build ./cmd/pm
make connector-boundary
make verify
git diff --check
```

Additional focused tests before broad gates:

```bash
go test ./cmd/connectorgen -run TestValidateSingleBundleDir -count=1
go test ./internal/connectors/native/google-analytics-data-api ./internal/connectors/hooks/google-analytics-data-api -count=1
```
