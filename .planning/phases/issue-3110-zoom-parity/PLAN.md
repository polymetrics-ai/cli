# Zoom connector parity plan (#3110)

## Scope

Implement connector-local Zoom official API parity evidence for parent #3110 and subissues #3111-#3117 on branch `fm/cli-zoom-parity-wave04-r1`.

Allowed production paths for this slice:

- `internal/connectors/defs/zoom/**`
- Zoom connector-owned generated/manual docs under `docs/connectors/zoom/**`
- CLI/help/golden/catalog and website generated surfaces only when regeneration reflects the Zoom connector metadata change
- Issue bodies for #3110-#3117 through `gh-axi` for the required idempotent captain-policy addendum

No shared runtime/engine/CLI Go behavior changes are planned for connector semantics. During verification, the recursive certification harness exposed a fixed 20m package-timeout blocker from repeatedly reparsing the embedded connector bundle catalog after adding the Zoom parity bundle. A minimal shared performance unblock may cache the immutable embedded bundle load while still returning a fresh registry per `New()` call; this must not change connector behavior, safety policy, or runtime surfaces.

## GSD command path and fallback

- `scripts/gsd doctor`: passed.
- `scripts/gsd list`: passed.
- Required implementation command attempt `scripts/gsd prompt programming-loop init --phase issue-3110-zoom-parity --dry-run`: failed because this adapter does not expose `programming-loop`.
- `scripts/gsd prompt quick "Implement Zoom connector parity issue 3110"`: generated an official GSD quick-task prompt.
- Manual GSD fallback is active for programming-loop semantics using `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`: plan before production edits, TDD/ledger evidence, and verification checklist kept in this phase directory.

## Required skills loaded

- `gsd-core`
- `golang-how-to`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-testing`
- `golang-context`
- `golang-concurrency`
- `golang-cli`
- `golang-spf13-cobra`
- `golang-documentation`
- `context-mode`
- CLI help/docs/website parity reference
- Connector migration conventions and connector architecture v2 design docs

## Source inventory

Official source for this slice:

- Zoom Developer Docs Meeting API reference: `https://developers.zoom.us/docs/api/rest/reference/zoom-api/methods/`
- Fetched evidence in `.tmp/zoom_official.html`: HTTP 200, ETag `"08322c4f0fa086914cd5d144268b61bf"`, Last-Modified `Fri, 31 Jul 2026 17:15:56 GMT`, Next build id `2026-07-31T11-05-34-06-00`.
- Extracted packed OpenAPI evidence in `.tmp/zoom_openapi.json`: OpenAPI `3.0.0`, `info.version=2`, title `Zoom Meeting API`, 129 paths, 184 HTTP operations (`GET=93`, mutations=91), decompressed bytes `1187667`, SHA256 `7490bae1a0815d82721af85b894e90f5662489c0aacaa2300277758562213ab9`.

Landed audit record from #3110 remains the dispatch baseline: rank 49, source id `zoom-meeting-api-openapi-embedded`, 184 total operations with lane counts `etl_read=48`, `reverse_etl_write=78`, `direct_read_query_search=29`, `binary_file=29`, `cdc_changefeed=0`, `excluded_not_applicable=0`. The current official source still has 184 operations, so this slice preserves the landed lane allocation while updating source freshness metadata in Zoom-owned docs.

## Implementation slices

1. **Operation ledger parity**
   - Replace partial `api_surface.json` with an operation-ledger-version 1 ledger covering all 184 official operations exactly once.
   - Preserve the landed audit lane counts with explicit classifier evidence.
   - No generic HTTP, query, SQL, shell, file, GraphQL, or passthrough escape.

2. **ETL/read streams**
   - Declare 48 stream-backed operations with bounded page-size defaults and fixture replay.
   - Use connector-relative paths and declared config path parameters; do not require live credentials.

3. **Direct and binary surfaces**
   - Declare fixed-target `operations.json` and `cli_surface.json` command metadata for 29 direct/provider reads plus the 29 binary/file-lane operations.
   - Execute JSON GET direct reads through existing `rest_read`/`OperationDirectRead` only where supported.
   - Binary/file operations that are provider-state mutations are exposed through typed write actions; pure binary download/file-transfer rows record the existing binary executor boundary truthfully.

4. **Typed reverse ETL writes**
   - Declare named write actions for all 91 mutating official operations (the 78 reverse lane mutations plus 13 binary/file-lane mutations), with closed record schemas, path-field redaction, risk text, destructive confirmation for DELETE/destructive updates, and idempotent delete 404 handling where provider semantics support already-gone success.
   - Fixtures are synthetic and cover request shape for each write action.

5. **Docs/config/certification/generated surfaces**
   - Update Zoom `docs.md`, generated `MANUAL.md`/`SKILL.md`, CLI surface metadata, certification metadata, docs/website generated catalogs, and golden help transcripts only when touched by Zoom metadata.
   - Certification remains fixture-only and uncertified (`certified=0`) because no live provider call is authorized.

6. **Issue policy addendum**
   - Append/update the captain destructive/delete scope addendum idempotently on #3110-#3117 with actual post-change counts.

## Orchestration decision

`local_critical_path`: the task assigns one connector directory and seven coupled subissues to this isolated worker branch. Mutating subagents were not spawned because the write scope is one connector-owned tree and additional mutating workers would collide in the same worktree. Read-only recon ran inline with context-mode tooling.

## Human gates / non-goals

- No secrets, credential requests, live Zoom calls, provider writes, certification claims, VPS/Thaalam work, PR creation, pushes, merges, or no-mistakes invocation.
- No new dependencies.
- No shared runtime behavior edits.
- Reverse ETL remains plan → preview → explicit approval → execute.
