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
- Fixed lane allocation remains: ETL/read 19, direct/provider query 10, reverse ETL write 31, binary 0, CDC 0, excluded 0.

## Implementation slices

1. Red/contract tests and fixtures:
   - Add connector-local tests proving 60 operation rows, 19 streams, 10 direct-read operations, 31 write actions, no generic escape hatches, and native/header request dispatch.
   - Add sanitized replay fixtures for check, every stream, and every write action.
2. Definition bundle:
   - Replace legacy hook-only surface with exact AWS operation ledger in `api_surface.json`.
   - Add `operations.json`, `writes.json`, `cli_surface.json`, `certification.json`.
   - Expand `streams.json`, stream schemas, metadata, and docs for the 19/10/31 lane split.
3. Native/hook runtime:
   - Generalize AWS CloudTrail JSON-RPC execution with per-action `X-Amz-Target`, bounded POST body, SigV4, max pages, and read/query/write dispatch.
   - Preserve fixture mode and no-live-provider tests.
   - Add small shared support needed by promoted definition-backed natives: expose bundle-derived manifests/command surfaces, forward optional direct-read/dry-run/validation interfaces, cache parsed bundles for repeated CLI test/runtime registry construction, and allow `connectorgen validate <connector-dir>` for focused gates.
4. Generated surfaces:
   - Regenerate connector docs/skills and website connector data without unrelated churn.
5. Issue addendum:
   - Append idempotent captain-policy addendum to #3142-#3149 with truthful counts and gates.

## Safety invariants

- No live AWS API calls, no credential requests, no secret values in files/logs/issues.
- No new dependencies.
- No generic HTTP method/path/body/query, shell, SQL write, local file, binary, browser, or passthrough escape hatches.
- Reverse ETL remains plan -> preview -> approval -> execute; destructive/admin writes are typed and redacted.
- AWS event-record fields are modeled as payload/schema fields where returned by AWS; they are not counted as CDC operations.
