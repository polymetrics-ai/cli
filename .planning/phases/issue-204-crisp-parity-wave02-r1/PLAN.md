# PLAN — issue #204 Crisp parity wave02 r1

## Scope

Parent issue: #204. Subissues inventoried: #205 ledger, #206 ETL/read + changefeed, #207 direct/query/binary, #208 reverse ETL, #209 CLI/config/help, #210 fixtures/docs/conformance/Connector Guard, #211 certification/release evidence.

Allowed write scope for this worker:

- `internal/connectors/defs/crisp/**`
- Crisp-owned planning artifacts under `.planning/phases/issue-204-crisp-parity-wave02-r1/**`
- GitHub issue body addenda for #204-#211 via `gh-axi`, append-only with counts preserved

Forbidden: shared runtime/foundation files, other connectors, live provider calls, credentials, writes/certification, no-mistakes pipeline, push, PR, merges.

## Required context loaded

- `AGENTS.md` project instructions (provided by Pi)
- `.agents/agentic-delivery/references/required-skills-routing.md`
- `.agents/agentic-delivery/contracts/parent-orchestrator-contract.md`
- `.agents/agentic-delivery/workflows/parent-issue-orchestration-loop.md`
- `.agents/agentic-delivery/contracts/issue-agent-contract.md`
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`
- `.agents/agentic-delivery/references/gsd-pi-adapter.md`
- `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`
- `docs/migration/HANDOFF-CODEX.md`
- `docs/migration/conventions.md`
- `docs/architecture/connector-architecture-v2-design.md`
- `.planning/config.json`, `.planning/PROJECT.md`, `.planning/ROADMAP.md`, `.planning/STATE.md`
- `docs/plans/universal-programming-loop-prd.md`
- `docs/prompts/universal-programming-loop-prompts.md`

## Required skills loaded

`gsd-core`, `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-cli`, `golang-documentation`.

## GSD command path

- `scripts/gsd doctor` passed.
- `scripts/gsd list` showed 69 commands.
- `scripts/gsd prompt programming-loop init --phase issue-204-crisp-parity --dry-run` failed because this repo-local adapter currently has no `programming-loop` prompt registered. Manual GSD fallback is active using `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`; this fallback is recorded here before production edits.

## Orchestration decision

Cycle `plan`: `local_critical_path` — firstmate launched this worker into an isolated disposable worktree with the full Crisp parent/subissue scope; mutating sub-workers are not spawned because the assigned write scope is one connector bundle plus Crisp-owned artifacts and all subissue work collides on `internal/connectors/defs/crisp/**`.

## Implementation slices

1. **Issue graph and source inventory**
   - Capture parent/subissue graph with `gh-axi`.
   - Fetch official Crisp REST API V1 docs without provider credentials.
   - Generate a deterministic operation inventory with method/path/title/category/source URL and lane classification.

2. **Safety-policy addendum**
   - Append an idempotent captain-policy addendum to #204-#211 via `gh-axi issue edit` only when the marker is absent.
   - Preserve existing issue bodies and count tables verbatim.

3. **Connector-local bundle**
   - Add `internal/connectors/defs/crisp/{metadata.json,spec.json,streams.json,api_surface.json,operations.json,cli_surface.json,certification.json,docs.md}`.
   - Use blocked/planned operation rows for operations that lack an executable connector-local safe contract in this slice.
   - Do not expose raw method/path/body/query escape hatches.

4. **Evidence artifacts**
   - Write `SOURCES.md` with official source inventory, counts, parser evidence, and shared-foundation blockers.
   - Update `TDD-LEDGER.md`, `VERIFICATION.md`, `SUMMARY.md`, and `RUN-STATE.json` with red/green/refactor evidence.

5. **Local gates and commit**
   - Targeted: `go run ./cmd/connectorgen validate internal/connectors/defs/crisp`; `go test ./internal/connectors/conformance -run 'TestConformance/crisp' -count=1`; `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1`; `go vet ./internal/connectors/... ./internal/cli/...`; `go build ./cmd/pm`; `make connector-boundary`; `git diff --check`.
   - Broader gates as time allows: `gofmt -w cmd internal` only if Go files changed; `go test ./...`; `make verify` if practical.
   - Commit the connector bundle and Crisp-owned artifacts. Stop; do not run no-mistakes, push, or open a PR.

## Safety decisions

- No live provider calls, credentials, certification, or write execution.
- DELETE/destructive/admin operations are included in the ledger and command metadata; executable reverse ETL requires named actions, schemas, redaction, plan -> preview -> explicit approval -> execute, and `confirm: destructive` before any future execute path can exist.
- Provider search/query/direct/binary operations remain planned/blocked when shared foundation (#2985) or connector-owned fixed-target policies are absent.
- CDC/changefeed operations remain planned/blocked when CDC foundation (#2986/#2988) or webhook/runtime intake contracts are absent.

## Completion update

- Crisp bundle and planned CLI surface generated under `internal/connectors/defs/crisp/**`.
- Connector docs/catalog and website generated connector data refreshed for Crisp.
- Targeted gates, full `go test -timeout 20m ./...`, and `make verify` are green; see `VERIFICATION.md`.
- Stop before no-mistakes, push, or PR per user scope.
