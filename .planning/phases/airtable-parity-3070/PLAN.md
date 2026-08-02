# Airtable official API parity — issue #3070 wave 03

## GSD command path

- `scripts/gsd doctor` — passed in this worktree.
- `scripts/gsd prompt programming-loop init --phase airtable-parity-3070 --dry-run` — adapter returned `unknown GSD command: programming-loop`; recorded manual-GSD fallback for the missing command name only.
- `scripts/gsd prompt plan-phase airtable-parity-3070 --skip-research` — generated the planning prompt used for this artifact.
- `scripts/gsd prompt execute-phase airtable-parity-3070 --dry-run` — generated the execution prompt used for the local critical path.

Manual-GSD fallback is limited to executing the generated prompts with Pi tools in this isolated worker because the repo-local shell adapter has no `programming-loop` command entry.

## Required skills loaded

- `gsd-core`
- `golang-how-to`
- `golang-cli`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-context`
- `golang-documentation`
- `context-mode`
- References: `required-skills-routing.md`, `cli-help-docs-website-parity.md`, `gsd-pi-adapter.md`, `issue-agent-contract.md`, parent/subissue orchestration references.

## Source contract

- Parent: #3070.
- Subissues read with `gh-axi --full`: #3071, #3072, #3073, #3074, #3075, #3076, #3077.
- Official source re-audited from `https://airtable.com/developers/web/api/introduction`.
- Embedded OpenAPI version: `3.1.0`.
- OpenAPI SHA-256: `f1506571034500ffb0887e7244f770c0b060a4205e734881bd3de6fa20d6f6b8`.
- Fetch evidence: 103 operations, methods `GET=31`, `POST=33`, `DELETE=19`, `PATCH=15`, `PUT=5`.
- Lane counts reproduced: `etl_read=27`, `reverse_etl_write=69`, `direct_read_query_search=1`, `binary_file=1`, `cdc_changefeed=5`, total `103`.

## Orchestration decision

`local_critical_path`: all seven subissues touch the same connector bundle (`internal/connectors/defs/airtable/**`) plus generated docs/catalog surfaces, so mutating worker fan-out would collide. Read-only subagent review may be used after implementation, but production edits stay local in this isolated worktree.

## Implementation slices

1. **Audit and ledger slice**
   - Replace `api_surface.json` with all 103 official operations exactly once.
   - Preserve source links and classify remaining blocked work only with exact evidence or engine/runtime dependency.

2. **Read/changefeed stream slice**
   - Update Airtable base URL/path model to cover both `/v0/**` and `/scim/**` official endpoints.
   - Add declarative streams, schemas, and sanitized fixtures for all executable GET/read/changefeed operations.
   - Keep bounded pagination where official list endpoints paginate; use single-object streams for detail endpoints.

3. **Reverse/binary mutation slice**
   - Add typed JSON write actions for supportable POST/PATCH/PUT/DELETE operations.
   - Keep `upload_attachment` blocked on `airtable-bounded-base64-upload-foundation` until an Airtable-owned executor can validate base64 encoding and decoded size before transmission.
   - Mark destructive/admin deletes and destructive replacement PUTs with `confirm: "destructive"` and idempotent 404 handling where the API supports missing-resource tolerance.
   - Final implementation partition from the audited spec: 28 executable read streams, 44 executable typed write actions, 1 executable direct-read CLI operation, and 30 blocked operations.
   - Keep Sync API CSV import blocked because the shared write runtime cannot emit `text/csv` bodies without a raw upload escape hatch.

4. **Direct/provider query slice**
   - Add `operations.json` + `cli_surface.json` for the HyperDB `getRecords` POST direct read, bounded by declared body schema, fixed path, redaction, and `max_bytes`.
   - Add CLI golden/test coverage for the new Airtable direct read command.

5. **Docs/generated surfaces slice**
   - Regenerate connector manuals/skills/catalog data and website generated connector data.
   - Update Airtable `docs.md`, generated `MANUAL.md`, generated `SKILL.md`, all-connectors catalogs, and website generated JSON/TS surfaces.

6. **Verification and issue addendum slice**
   - Run required gates from the task.
   - Issue-comment addenda are not executed in this local-only resume because the requested stopping point is the tested local commit with no shipping/noise actions; actual post-change counts are preserved in this phase artifact for a coordinator to serialize externally if desired.
   - Commit the clean slice locally; do not push, invoke `/no-mistakes`, open/update PRs, or merge.

7. **CI issue-link guard repair slice**
   - Capture PR #3540's generated unvalidated-checkpoint body as a regression test before changing guard behavior.
   - Recognize canonical GitHub issue URLs only inside the generated `Canonical issue links preserved from the task record` section of an explicitly completed-task checkpoint; keep standalone bare URLs and vague `Issue`/`References` wording rejected.
   - Verify the focused guard packages and the exact CI-shaped `prissueguard` invocation locally; leave push, PR mutation, and pipeline control to the outer no-mistakes executor.

## Final implementation state

- `api_surface.json`: 103 official operations tracked exactly once.
- Executable partition: 28 stream-backed GET/read/changefeed operations, 44 typed write actions, and 1 HyperDB direct-read CLI operation.
- Blocked partition: 30 operations, including Sync API CSV import and attachment upload, remain blocked on named typed runtime foundations rather than exposing unsafe or unenforceable request shapes.
- Comments use the official exact per-record endpoint and require `record_id`; the stream no longer bulk-fans out while ignoring the narrowing flag.
- Validation artifacts: `VERIFICATION.md` distinguishes prior full verification from the focused review-fix gate owned by this phase.

## Safety constraints

- No live Airtable calls, credentials, provider writes, secret values, VPS, Thaalam, or runtime service changes.
- No new dependencies.
- No shared runtime behavior edits unless validation reveals an existing generated-surface issue that must be recorded instead of changed.
- No generic HTTP/read/write/query escape hatches.
- Reverse ETL remains plan -> preview -> explicit approval -> execute.
