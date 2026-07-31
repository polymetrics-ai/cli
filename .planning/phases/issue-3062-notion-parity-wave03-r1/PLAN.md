# PLAN — issue-3062 Notion parity wave03 r1

## Scope

Parent issue: #3062. Subissues: #3063-#3069. Connector: `internal/connectors/defs/notion` plus connector-owned Notion hook/tests/fixtures/docs/generated surfaces.

## GSD path

- Ran `scripts/gsd doctor` successfully.
- Attempted documented `scripts/gsd prompt programming-loop init --phase issue-3062-notion-parity-wave03-r1 --dry-run`; adapter returned `unknown GSD command: programming-loop` even though doctor/list are healthy.
- Manual GSD fallback is active for this phase, following `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md` and `.pi/prompts/pm-gsd-loop.md`.

## Required skills loaded

`gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-documentation`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-spf13-cobra`, `golang-lint`, `context-mode`; plus `cli-help-docs-website-parity.md`.

## Orchestration decision

`local_critical_path`: all executable work is in one connector's definition/hook/docs surface; splitting mutating workers would collide on the same Notion bundle and generated docs. Read-only recon is inline via context-mode, not spawned.

## Implementation slices

1. Re-audit official Notion OpenAPI from `https://developers.notion.com/openapi.json`, record source metadata, operation inventory, and count evidence in `traces/notion-official-openapi-audit.md`. **Done.**
2. Replace the partial Notion bundle with a complete operation ledger, expanded streams, typed writes, sanitized stream/write fixtures, docs, and certification metadata. Keep direct/search/query/file/binary/OAuth operations blocked only when the exact shared runtime or safety dependency is named. **Done.**
3. Extend the Notion Tier-2 StreamHook only as needed for Notion's body-cursor/read pagination and schema projection; no shared runtime behavior changes except the validation-tool compatibility fix required by the mandated exact connectorgen command. **Done.**
4. Update generated MANUAL/SKILL and website/catalog surfaces with the repo generators. **Done; unrelated connector-doc churn was reverted, retaining Notion docs only.**
5. Append the captain destructive/delete policy addendum idempotently to #3062-#3069 with actual post-change counts and no certification claim. **Done via `gh-axi issue edit --body-file`.**
6. Run required local gates and commit a clean slice; do not push, open/update PRs, merge, invoke `/no-mistakes`, or drive live/provider checks. **Verification passed; local commit pending.**

## Safety gates

- No live Notion calls beyond public OpenAPI fetch; no credentials; no provider writes.
- Reverse ETL remains typed plan → preview → explicit approval → execute. This task only authors schemas/fixtures and local dry-run/conformance evidence.
- Destructive/delete-like operations use explicit action schemas, risk text, redaction where identifiers appear in paths, `confirm: "destructive"`, and idempotency notes where supported.
- The official multipart `upload_file` byte-transfer operation remains blocked/planned with exact reason: shared binary payload approval/conformance runner is required to provide approved digests and live-safe redacted artifacts without broad local file exposure.

## Final post-change inventory

Official Notion OpenAPI operation set audited from `https://developers.notion.com/openapi.json`:

| total official ops | implemented | fixture-tested | blocked/planned | excluded/not-applicable | certified |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 49 | 45 | 45 | 1 | 3 | 0 |

Connector-surface addition: two legacy-compatible object-filtered `/search` stream variants (`databases`, `pages`) remain fixture-tested and are represented as extra connector-surface rows, not additional official OpenAPI operations.

Blocked/planned official operation: `POST /file_uploads/{file_upload_id}/send` (`upload-file`) pending shared binary payload approval/conformance runner.

Excluded/not-applicable operations: `POST /oauth/token`, `POST /oauth/revoke`, `POST /oauth/introspect` (OAuth token lifecycle, not connector data operations).
