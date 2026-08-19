# YouTube Analytics connector parity plan (#3456-#3463)

## Scope

Implement connector-local YouTube Analytics (`youtube-analytics`) documented official API parity on branch `fm/cli-youtube-analytics-parity-wave03-r1`.

Allowed production paths for this wave:

- `internal/connectors/defs/youtube-analytics/**`
- `internal/connectors/hooks/youtube-analytics/**` only if the existing OAuth refresh hook needs connector-local coverage updates
- YouTube Analytics connector-owned generated docs and data surfaces: `docs/connectors/youtube-analytics/**`, `docs/connectors/catalog/**`, `docs/connectors/README.md`, `website/data/connectors.generated.json`, `website/lib/connectors*.generated.*`
- Targeted CLI/help/golden artifacts only if generated tests report drift

Do not edit shared connector runtime/engine/CLI behavior. If a direct-read, binary, cross-host, or write-query capability needs shared runtime work, keep that operation blocked with exact evidence.

## GSD command path and fallback

- Ran `scripts/gsd doctor`: passed.
- Ran `scripts/gsd list`: passed, 69 commands listed.
- Attempted required implementation command `scripts/gsd prompt programming-loop init --phase issue-3456-youtube-analytics-parity --dry-run`: failed with `scripts/gsd: unknown GSD command: programming-loop`.
- Rendered `scripts/gsd prompt plan-phase issue-3456-youtube-analytics-parity --skip-research` for official command provenance.
- Manual GSD fallback is active for this phase using `.pi/prompts/pm-gsd-loop.md`, `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`, and `docs/prompts/universal-programming-loop-prompts.md`.

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
- CLI help/docs/website parity reference

## Official source inventory

Official sources re-audited for this slice:

- `youtube_analytics_discovery_v2`: `https://youtubeanalytics.googleapis.com/$discovery/rest?version=v2`; title `YouTube Analytics API`, version `v2`, revision `20260729`, base URL `https://youtubeanalytics.googleapis.com/`; 8 operations.
- `youtube_reporting_discovery_v1`: `https://youtubereporting.googleapis.com/$discovery/rest?version=v1`; title `YouTube Reporting API`, version `v1`, revision `20260729`, base URL `https://youtubereporting.googleapis.com/`; 8 operations.
- Official docs pages fetched/indexed: Analytics `reports.query`, `groups`, `groupItems`; Reporting REST root, `jobs`, `reportTypes`, `jobs.reports`, and bulk-report overview.

Official operation inventory (16 total):

| Source | Operation | Method/path | Initial disposition |
| --- | --- | --- | --- |
| YouTube Analytics v2 | `youtubeAnalytics.groups.list` | `GET /v2/groups` | implement stream `groups` |
| YouTube Analytics v2 | `youtubeAnalytics.groupItems.list` | `GET /v2/groupItems` | implement stream `group_items` |
| YouTube Reporting v1 | `youtubereporting.jobs.list` | `GET /v1/jobs` | keep stream `jobs` |
| YouTube Reporting v1 | `youtubereporting.jobs.get` | `GET /v1/jobs/{jobId}` | implement stream `job` |
| YouTube Reporting v1 | `youtubereporting.reportTypes.list` | `GET /v1/reportTypes` | keep stream `report_types` |
| YouTube Reporting v1 | `youtubereporting.jobs.reports.list` | `GET /v1/jobs/{jobId}/reports` | keep stream `reports` |
| YouTube Reporting v1 | `youtubereporting.jobs.reports.get` | `GET /v1/jobs/{jobId}/reports/{reportId}` | implement stream `report` |
| YouTube Reporting v1 | `youtubereporting.jobs.create` | `POST /v1/jobs` | implement write `create_job` |
| YouTube Reporting v1 | `youtubereporting.jobs.delete` | `DELETE /v1/jobs/{jobId}` | implement destructive write `delete_job` |
| YouTube Analytics v2 | `youtubeAnalytics.groups.insert` | `POST /v2/groups` | implement write `create_group` |
| YouTube Analytics v2 | `youtubeAnalytics.groups.update` | `PUT /v2/groups` | implement write `update_group` |
| YouTube Analytics v2 | `youtubeAnalytics.groups.delete` | `DELETE /v2/groups` | implement destructive write `delete_group` |
| YouTube Analytics v2 | `youtubeAnalytics.groupItems.insert` | `POST /v2/groupItems` | implement write `create_group_item` |
| YouTube Analytics v2 | `youtubeAnalytics.groupItems.delete` | `DELETE /v2/groupItems` | implement destructive write `delete_group_item` |
| YouTube Analytics v2 | `youtubeAnalytics.reports.query` | `GET /v2/reports` | block as `direct_read` pending provider query/cross-host direct-read foundation #2985 |
| YouTube Reporting v1 | `youtubereporting.media.download` | `GET /v1/media/{+resourceName}` | block as `binary_read` pending bounded binary executor |

## Implementation slices

1. **Operation ledger parity**
   - Convert `api_surface.json` to `operation_ledger_version: 1`.
   - Represent all 16 official discovery operations exactly once with `covered_by` or blocked `operation` rows.
   - Remove the previous absolute extra `reports:query` path row and use the discovery path `/v2/reports`.

2. **Read stream expansion**
   - Add streams/schemas/fixtures for `job`, `report`, `groups`, and `group_items`.
   - Preserve existing `jobs`, `report_types`, and `reports` stream names.
   - Keep JSON records distinct from raw report bytes: `reports`/`report` expose report metadata including `download_url`; `media.download` remains blocked binary.

3. **Write action expansion**
   - Add `writes.json` with 7 typed actions: Reporting job create/delete plus Analytics group/group-item create/update/delete.
   - Use closed record schemas with required fields, `additionalProperties: false`, path-field redaction on destructive IDs, and `confirm: "destructive"` for deletes.
   - Do not broaden default OAuth scope; document that write execution requires caller-provided credentials/scopes adequate for the provider operation.
   - Do not claim idempotent delete unless official docs state missing deletes are safe.

4. **CLI/help/docs metadata**
   - Add `cli_surface.json` commands for implemented streams and write-plan commands.
   - Add planned/blocked docs-only commands for `reports query` and report binary download without making them executable.
   - Regenerate connector manuals/catalog and website connector generated data.

5. **Fixtures/conformance/certification truthfulness**
   - Add sanitized stream and write fixtures for every executable stream/action.
   - Keep dynamic conformance skip truthful if the existing custom OAuth refresh hook prevents fixture replay from resolving auth; static checks and hook tests remain the substitute proof.
   - Do not run live provider calls, credentials, provider writes, or certification.

6. **Issue addendum**
   - Append the established captain-policy addendum idempotently to #3456-#3463 with actual post-change counts.
   - Use `gh-axi issue comment`; do not edit issue bodies or fabricate certification.

## Orchestration decision

`local_critical_path`: firstmate launched this isolated crewmate to own the coupled parent+seven subissues for one connector. Mutating subagents are not spawned because the write scope is one connector-owned bundle/docs surface in this worktree and parallel mutators would collide. Read-only official-source recon was performed inline with context-mode and gh-axi.

## Human gates / non-goals

- No secrets, credential requests, live YouTube/Google API calls, provider writes, certification, VPS, Thaalam, Herdr lifecycle, merges, or pushes.
- No new dependencies.
- No shared runtime/engine/CLI behavior edits.
- Reverse ETL remains plan → preview → explicit approval → execute.
- Direct provider query and binary bytes stay blocked unless an existing bounded executor can support them without shared changes.
