# GSD plan — issue #79 Bitbucket parity

## Command path and fallback

- GSD adapter preflight: `scripts/gsd doctor` and `scripts/gsd list` passed on this worktree.
- Required programming-loop command attempted: `scripts/gsd prompt programming-loop init --phase issue-79-bitbucket-parity --dry-run`.
- Adapter result: `unknown GSD command: programming-loop`; `scripts/gsd help programming-loop` and `scripts/gsd sources pm-gsd-loop` were also unavailable.
- Manual-GSD fallback is active for this worker. This file records the plan, TDD ledger, verification checklist, and execution decisions before connector production edits.
- Merge-readiness resume command: `scripts/gsd prompt progress --do "resume issue 79 Bitbucket connector parity through no-mistakes validation, PR creation, and green checks without live-provider work or merge"`.

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
- `golang-spf13-cobra`
- `golang-lint`
- `no-mistakes`
- `.agents/agentic-delivery/references/required-skills-routing.md`
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`
- `docs/migration/HANDOFF-CODEX.md`
- `docs/migration/conventions.md`
- `docs/architecture/connector-architecture-v2-design.md`

## Scope guard

Initial connector-bundle edits for this worker stayed connector-local under `internal/connectors/defs/bitbucket/`.
Review follow-up shared runtime edits were limited to reverse-plan redaction metadata and generated CLI/docs parity.
No hooks, natives, command runners, go.mod/go.sum, live credentials, provider calls, VPS/Thaalam work, merges, or certification claims were added.
GitHub issue-body updates use `gh-axi` only and preserve existing issue body counts.

## Captain policy addendum requirement

Append an idempotent addendum to #79 and #90-#96 stating that destructive/delete operations are in scope when implemented as typed actions with `confirm: "destructive"` and the existing plan -> preview -> explicit approval -> execute path. The addendum must not change existing count tables or fabricate implemented counts.

Evidence: `gh-axi issue edit` updated #79 and #90-#96 with marker `captain-policy-addendum-bitbucket-destructive-confirmation-r1`; verification found exactly one marker in each body.

## Implementation slices

1. **Policy and provenance**
   - Read parent/subissue bodies and official Bitbucket OpenAPI source.
   - Append the idempotent captain-policy addendum with `gh-axi`.
   - Record the official source URL and observed OpenAPI operation totals in connector docs.

2. **Definition bundle scaffold**
   - Create `metadata.json`, `spec.json`, `streams.json`, `schemas/`, `writes.json`, `api_surface.json`, `operations.json`, `cli_surface.json`, `certification.json`, fixtures, and `docs.md`.
   - Use only synthetic fixture values and no secret-shaped literals.

3. **Operation ledger completeness**
   - Generate one `api_surface.json` row for every official OpenAPI operation.
   - Implement durable GET operations as declarative streams where the generic engine can safely express the route.
   - Implement mutating operations as typed reverse-ETL actions with schemas, path fields, risk text, and destructive confirmation for DELETE/destructive routes.
   - Classify direct/binary/CDC/foundation-dependent surfaces truthfully in connector-local ledgers when shared execution support is absent; do not claim live certification.

4. **Fixtures and docs**
   - Add sanitized check, read, and write fixtures sufficient for conformance.
   - Document auth setup, streams, write risks, known limits, dependency blockers, and evidence.

Implemented connector-local inventory summary: `api_surface.json` has 331 official rows (GET 179, POST 50, PUT 48, DELETE 54); 143 rows are declarative streams, 54 rows are closed-schema reverse-ETL writes, and 134 rows are blocked connector-local operation metadata for untyped JSON-body, schema-ambiguous user-email reads, binary/direct-read/provider-query/upload foundations. All executable DELETE writes carry `confirm: "destructive"`.

5. **Verification and commit**
   - Red: focused validation fails before bundle files exist.
   - Green target: `go run ./cmd/connectorgen validate internal/connectors/defs/bitbucket`.
   - Run focused conformance if the bundle reaches dynamic readiness; otherwise record exact blocker.
   - Run `git diff --check`, connector boundary if available, and commit the coherent connector-local slice.

6. **Merge-readiness gate**
   - Invoke `/no-mistakes` as `no-mistakes axi run --intent ...` on the committed feature branch.
   - Let the no-mistakes pipeline own review, fixes, tests, docs, lint, push, PR creation, and CI gates while it is active.
   - Do not run live-provider or credentialed Bitbucket checks, do not merge, and preserve typed destructive confirmation and fixtures.
   - Stop only at `checks-passed`/green PR checks or an explicit blocker that needs human input.

## TDD ledger

| Step | Command/evidence | Result |
| --- | --- | --- |
| Red validation | `go run ./cmd/connectorgen validate internal/connectors/defs/bitbucket` before bundle files | command reported 0 checked/0 findings because an empty connector dir is skipped by connectorgen |
| Red conformance | `go test ./internal/connectors/conformance -run 'TestConformance/bitbucket' -count=1` before bundle files | failed as expected: `load bundle bitbucket: missing required file metadata.json` |
| Bundle validation | `go run ./cmd/connectorgen validate internal/connectors/defs` after bundle generation | passed: `550 connector(s) checked, 0 findings` |
| Conformance | `go test ./internal/connectors/conformance -run 'TestConformance/bitbucket' -count=1` | passed |
| CLI/golden targeted | `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1` and `go test ./internal/cli -run 'TestConnectorCatalogCLIJSON|TestGoldenTranscripts|TestGoldenDocsGenerateMatchesTrackedCLIManuals|TestDocsGenerateIncludesConnectorCatalog' -count=1` | passed after generated golden/catalog docs updates |
| Focused connector packages | `go test ./internal/connectors/engine ./internal/connectors/conformance -count=1` | passed |
| Runtime help/inspect spot checks | `pm connectors inspect bitbucket --json`, `pm bitbucket --help`, `pm bitbucket repositories delete --help` via `go run ./cmd/pm` | passed before review follow-up; generated metadata now reports 143 streams and 54 write actions after blocking schema-ambiguous user-email reads, delete help shows destructive confirmation path |
| Broad package attempts | `go test ./internal/connectors/...` timed out at 300s; `go test ./internal/cli -count=1` timed out at 240s with no failure output before timeout | incomplete broad gates; focused connector/CLI gates above passed |
| Vet | `go vet ./internal/connectors/... ./internal/cli/...` | passed |
| Build | `go build ./cmd/pm` | passed |
| Boundary | `make connector-boundary` | passed with clean outcome |
| Review finding regression | JSON inspection of Bitbucket metadata | red state confirmed issue import, issue attachment, repository `/src` multipart uploads, untyped JSON-body mutations, and schema-ambiguous user-email reads were exposed; fixed state has 54 closed-schema writes, 134 blocked rows, 4 `file_upload` operations, 94 blocked `rest_write` operations, blocked user-email `rest_read` evidence, no removed upload writes, and `downloads get` references `get_repositories_workspace_repo_slug_downloads_filename` with `binary_file_bounded` |
| Focused review validation | copied Bitbucket definition into a single-connector temp root and ran `go run ./cmd/connectorgen validate <temp-root> --json` | passed with `connectors_checked: 1`, no findings, no warnings |
| Conformance retry | `go test ./internal/connectors/conformance -run 'TestConformance/bitbucket' -count=1` | timed out after 120s with no failure output; broader Go test processes were already active in the worktree |
| Diff hygiene | `git diff --check` | passed |
| Scope guard | `git status --short` and path review | working tree changes are the Bitbucket defs bundle, generated Bitbucket/catalog docs, CLI golden/catalog metadata required by the new connector, and this connector-local GSD artifact |
| Merge-readiness plan | `scripts/gsd prompt progress --do "resume issue 79 Bitbucket connector parity through no-mistakes validation, PR creation, and green checks without live-provider work or merge"` | prompt generated; proceeding with no-mistakes AXI gate on the committed feature branch |

## Verification checklist

- [x] `go run ./cmd/connectorgen validate internal/connectors/defs`
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/bitbucket' -count=1`
- [x] `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1`
- [x] `go vet ./internal/connectors/... ./internal/cli/...`
- [x] `go build ./cmd/pm`
- [x] `make connector-boundary`
- [x] `git diff --check`
- [ ] `no-mistakes axi run --intent "Implement complete documented Bitbucket connector parity for issue #79 and subissues #90-#96, preserving typed destructive confirmation, sanitized fixtures, issue policy addenda, no live-provider work, and no merge."`
- [ ] PR opened/updated and checks green

## Project memory hook

Ran `/Users/karthiksivadas/karthik-agent-workspace/bin/fm-ensure-agents-md.sh .` because repo instructions require it when `AGENTS.md`/`CLAUDE.md` exist. The script exited with `conflict: both AGENTS.md and CLAUDE.md are real files`; no project guidance was edited because this Bitbucket bundle adds no durable cross-project rule that belongs in AGENTS/CLAUDE.

## Orchestration decisions

- cycle=planning decision=local_critical_path reason=single connector-owned write scope; subissue lanes share the same `defs/bitbucket` files, so spawning mutating workers would collide.
- cycle=execute decision=local_critical_path reason=generated connector-local bundle, fixtures, command metadata, and docs in one isolated worktree; mutating fanout would collide on `defs/bitbucket`.
- cycle=verify decision=local_critical_path reason=focused validation, conformance, CLI golden/catalog, vet, build, boundary, and diff checks run locally.
