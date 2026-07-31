# AWS CloudTrail parity wave04 plan

## Scope

Parent issue #3142 with subissues #3143-#3149. Branch `fm/cli-aws-cloudtrail-parity-wave04-r1` in isolated worktree `/Users/karthiksivadas/.treehouse/cli-83d592/43/cli`.

## GSD path and fallback

- `scripts/gsd doctor` passed for the repo-local Pi adapter.
- `scripts/gsd prompt plan-phase aws-cloudtrail-parity-wave04 --skip-research` generated the planning prompt and was followed manually.
- `scripts/gsd prompt programming-loop init --phase aws-cloudtrail-parity-wave04 --dry-run` failed because this adapter registry has no `programming-loop`/`gsd-programming-loop` command. Manual GSD programming-loop fallback is active: plan first, red/contract tests, implementation, local verification, update ledger, commit.

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
- `golang-context`
- `golang-concurrency`

## Authoritative evidence read

- `AGENTS.md`
- `docs/migration/HANDOFF-CODEX.md`
- `docs/migration/conventions.md`
- `docs/architecture/connector-architecture-v2-design.md`
- `.agents/agentic-delivery/references/required-skills-routing.md`
- `.agents/agentic-delivery/references/gsd-pi-adapter.md`
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`
- Issue bodies for #3142-#3149 via `gh-axi issue view --full --comments`.
- Landed audit record rank 53 from `/Users/karthiksivadas/karthik-agent-workspace/data/cli-official-api-parity-audit-r2/audit.json`.
- Fresh official AWS CloudTrail API reference actions and event-record contents were audited; parsed operation/field ledgers are stored under `traces/`.

## Re-audit result

- Official AWS CloudTrail API action inventory remains 60 actions.
- Event record contents page still reports eventVersion current version 1.11 and 31 top-level record fields; those fields are schema/data fields, not separate connector operations.
- Final scope-corrected lane allocation: implemented ETL/read 19, implemented direct/provider query 0, implemented reverse ETL write 0, binary 0, CDC 0, excluded 0, blocked/planned 41 (10 direct/provider query + 31 write/admin requiring shared promoted-native forwarding).

## Implementation slices

1. Red/contract tests and fixtures:
   - Connector-local tests prove 60 operation rows remain inventoried, 19 streams are implemented, 41 direct/write operations are blocked/planned, no generic escape hatches exist, and native read dispatch uses fixed CloudTrail targets.
   - Sanitized replay fixtures cover check and every implemented stream.
2. Definition bundle:
   - `api_surface.json` inventories all 60 official actions exactly once: 19 covered read streams and 41 blocked/planned operation rows.
   - `operations.json` and `writes.json` intentionally contain zero executable actions in this connector-local corrective head.
   - `cli_surface.json` is removed because `pm aws-cloudtrail ...` is not executable without shared promoted-native command-surface forwarding.
   - `streams.json`, stream schemas, metadata, and docs reflect the implemented 19/0/0 + 41 blocked/planned split.
3. Native/hook runtime:
   - Keep AWS CloudTrail JSON-RPC read streams with per-action `X-Amz-Target`, bounded POST body, SigV4, max pages, fixture mode, and no-live-provider tests.
   - Revert all shared runtime edits; direct-read and write maps are empty so hidden native paths do not expose blocked operations.
4. Generated surfaces:
   - Regenerate root help golden transcripts, connector catalog/docs, and website connector data so CloudTrail has no command surface and no write actions.
5. Issue addendum:
   - Update captain-policy addendum for #3142-#3149 with final implemented/blocked counts.

## Safety invariants

- No live AWS API calls, no credential requests, no secret values in files/logs/issues.
- No new dependencies.
- No generic HTTP method/path/body/query, shell, SQL write, local file, binary, browser, or passthrough escape hatches.
- Reverse ETL remains plan -> preview -> approval -> execute; CloudTrail write/admin actions are not exposed in this corrective head.
- AWS event-record fields are modeled as payload/schema fields where returned by AWS; they are not counted as CDC operations.

## Scope correction 2026-08-01

User correction: commit `3dd65b20d` included six shared-file edits outside the connector-local brief. Corrective slice uses repo-local GSD via `scripts/gsd prompt gsd-fast "scope correction for aws cloudtrail parity wave04: revert six shared files only and verify"`; `scripts/gsd prompt programming-loop init --phase aws-cloudtrail-parity-wave04-scope-correction --dry-run` remained unavailable in this adapter, so manual GSD fast/programming-loop fallback is active.

Required skills for this corrective slice: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`.

Plan:

1. Preserve CloudTrail commit history; do not amend/reset/rebase.
2. Restore only these shared files to `3dd65b20d^` contents in the working tree: `cmd/connectorgen/main.go`, `cmd/connectorgen/validate.go`, `internal/cli/cli_test.go`, `internal/connectors/bundleregistry/registry.go`, `internal/connectors/engine/connector.go`, `internal/connectors/native/nativeset/promoted.go`.
3. Keep CloudTrail-owned/generated surfaces and dedicated tests intact.
4. Reclassify any CloudTrail command depending on reverted shared forwarding as blocked/planned rather than restoring shared runtime or leaving stale direct/write claims.
5. Rerun focused CloudTrail checks and `make verify`.
6. Commit the corrective scope revert as a new commit and append the clean corrected head to the external status file.

Shared-runtime dependency documented by this correction:

- Future enablement of the 10 provider query/direct-read actions and 31 write/admin actions needs a separate shared-runtime slice for promoted-native `Manifest`, `CommandSurface`, `OperationDirectRead`, `WriteValidator`, and `DryRunWriter` forwarding plus focused connector-dir validation. This corrective commit does not include that shared work.
