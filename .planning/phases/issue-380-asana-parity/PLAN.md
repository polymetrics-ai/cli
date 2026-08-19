# Asana connector parity plan (#380)

Parent issue: #380
Subissues: #381, #382, #383, #384, #385, #386, #387
Branch: `fm/cli-asana-parity-wave02-r1`

## Scope

Allowed production paths for this worker slice:

- `internal/connectors/defs/asana/**`
- Asana connector-owned fixtures under `internal/connectors/defs/asana/fixtures/**`
- Asana generated/manual connector docs and command metadata (`docs/connectors/asana/**`, `cli_surface.json`, `operations.json`, `certification.json`)
- GSD/planning artifacts under `.planning/phases/issue-380-asana-parity/**`

Do not edit shared runtime/foundation files, other connector definitions, generated hook/native import sets, `go.mod`, or `go.sum`. If a shared foundation is absent, record it in Asana-owned docs/ledgers and keep unsupported execution blocked/planned.

## GSD command path and fallback

- `scripts/gsd doctor` — passed.
- `scripts/gsd list` — passed; 69 commands available.
- `scripts/gsd prompt plan-phase 380 --skip-research` — rendered to `traces/gsd-plan-phase-380.prompt.md`.
- `scripts/gsd prompt gsd-quick --full "Implement connector-local Asana official API parity ..."` — rendered to `traces/gsd-quick-full.prompt.md`.
- Required implementation command attempted: `scripts/gsd prompt programming-loop init --phase issue-380-asana-parity --dry-run`; adapter returned `unknown GSD command: programming-loop`.
- Manual GSD fallback is active using `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`, `.pi/prompts/pm-gsd-loop.md`, and `docs/prompts/universal-programming-loop-prompts.md`.

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
- `golang-context`, `golang-concurrency` for connector runtime/cancellation review context
- `golang-code-style`, `golang-lint`, `golang-naming`
- `context-mode`
- Repo references: `AGENTS.md`, required-skill routing, parent orchestrator contract/workflow, issue-agent contract, stacked workflow, automated/Claude review routing, CLI help/docs/website parity, GSD Pi adapter, migration handoff, migration conventions, connector architecture v2 design, GSD planning files, universal programming loop PRD/prompts.

## Official source inventory

Authoritative source for this slice:

- `asana_openapi_pinned`: `https://raw.githubusercontent.com/Asana/openapi/56796a67a3c093eedf55fd9682357957a2ebfd85/defs/asana_oas.yaml`
- OpenAPI evidence from the pinned source: OpenAPI `3.0.0`, `info.version=1.0`, `249` HTTP operations (`GET=119`, `POST=81`, `PUT=26`, `DELETE=23`).

Parent #380 operation-lane counts remain authoritative: `etl_read=111`, `reverse_etl_write=125`, `direct_read_query_search=3`, `file_upload=1`, `cdc_changefeed=8`, `excluded_not_applicable=1`, total `249`.

## Implementation slices

1. **Issue policy addendum**
   - Append an idempotent captain-policy addendum to #380 and #381-#387 through `gh-axi`.
   - Preserve existing bodies and count tables.

2. **Complete official operation ledger**
   - Replace `api_surface.json` with an `operation_ledger_version: 1` ledger covering exactly the 249 pinned official operations once each.
   - Keep already executable Asana streams and writes covered.
   - Convert previous broad `excluded` rows into source-linked `operation` blocked/planned rows; DELETE/destructive/admin operations are in scope, not blanket unsafe exclusions.
   - Represent `/batch` as disallowed/not-applicable because it is a generic subrequest wrapper; every underlying operation is represented individually.

3. **Typed operation and command metadata**
   - Add `operations.json` for every official operation with bounded fixed-target metadata, risk, approval, mutation class, destructive flag, and source evidence.
   - Add `cli_surface.json` so provider-style command/help metadata is definition-owned. Planned commands must not expose raw method/path/body/query passthrough.
   - Record open shared-foundation dependencies (#2985, #2986, #2988) for provider search/direct and CDC/changefeed execution.

4. **Write safety hardening for existing executable actions**
   - Preserve the 13 existing Asana write actions and fixtures.
   - Ensure destructive delete actions use `confirm: "destructive"`, idempotent 404 semantics, and path field redaction.
   - Update risk text so every write documents reverse ETL plan -> preview -> explicit approval -> execute; destructive actions also require typed `--confirm destructive`.

5. **Docs and certification truthfulness**
   - Update `docs.md` with official source/count evidence, destructive/delete policy correction, blocked/planned shared dependencies, fixture-only uncertified status, and no-live-call/non-generic-tool safety notes.
   - Add a minimal connector-local certification contract that declares fixture source defaults only; no live certification claim.
   - Regenerate/update Asana connector manual/skill docs after metadata changes.

6. **Verification**
   - Run credential-free inventory comparison, `connectorgen validate`, Asana conformance, CLI/help checks, targeted vet/build, connector-boundary, and diff checks as time permits.
   - Do not run live provider calls, credentialed certification, no-mistakes, push, PR, or merge.

## Orchestration decision

`local_critical_path`: the task assigns one connector-owned write scope in one isolated worktree. Mutating subagents would collide on the same `internal/connectors/defs/asana/**` files and issue body addendum state. Read-only recon is handled inline with context-mode tooling.

## Implementation outcome

- Issue addendum marker added exactly once to #380 and #381-#387.
- `api_surface.json` now tracks exactly 249 pinned official Asana OpenAPI operations with no `/users/me` extra alias and no legacy blanket `excluded` rows.
- `operations.json` and `cli_surface.json` now provide fixed-target metadata for all 249 operations; planned/blocked commands document shared-foundation gaps instead of exposing raw passthrough.
- Existing executable surface remains 12 read streams and 13 reverse-ETL writes.
- Destructive implemented deletes use typed `confirm: "destructive"`, redacted path fields, idempotent 404 semantics, and plan -> preview -> approval -> execute documentation.
- Asana connector docs, generated manual/skill docs, certification truthfulness metadata, and CLI golden fixture were updated.

## Human gates / non-goals

- No secrets, credential requests, live Asana calls, provider writes, live certification, VPS/Thaalam work, no-mistakes shipping flow, pushes, PRs, merges, or default-branch operations.
- No new dependencies.
- No shared runtime/foundation edits.
- Reverse ETL remains plan -> preview -> explicit approval -> execute.
- Generic shell, raw HTTP write, raw SQL write/read, arbitrary GraphQL, file, binary, and unrestricted passthrough tools remain disallowed.
