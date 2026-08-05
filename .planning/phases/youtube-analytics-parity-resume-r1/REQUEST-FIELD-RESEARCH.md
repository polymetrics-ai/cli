# Provider request-field research matrix

This is the required research record while the shared machine-readable citation convention is landing. It is **not** a competing bundle format. When the convention reaches `main`, transfer these citations verbatim into its prescribed location and revalidate.

## Sources and method

Provider-owned Google reference pages were checked on 2026-08-05. Google provides prose parameter tables rather than OpenAPI request bodies for most of these APIs, so rows use tier 3 (`operation_reference`) unless explicitly marked tier 4 (`same_page_sibling`). The `media.download` row also cites Google's provider-owned Discovery document for its exact path parameter. Requiredness is provider wording, not inferred from the current bundle.

| Operation(s) | Declared request fields | Source URL / section | Evidence type | Confidence | Requiredness rationale |
| --- | --- | --- | --- | --- | --- |
| `groups.list` | query `mine`, `onBehalfOfContentOwner`, paginator `pageToken` | https://developers.google.com/youtube/analytics/reference/groups/list — Parameters | operation_reference | high | Google requires exactly one of `id` or `mine`, and `mine` must be `true`. The connector deliberately supports only the closed `mine=true` mode: its one-value enum is required and the stream cannot interpolate the request without it. Content-owner context and page token remain optional. |
| `groups.insert` | body `snippet.title`, `contentDetails.itemType` | https://developers.google.com/youtube/analytics/reference/groups/insert — Request body; https://developers.google.com/youtube/analytics/reference/groups — Resource representation | operation_reference | high | Google requires both fields when creating a group. The Group resource restricts item type to `youtube#channel`, `youtube#playlist`, `youtube#video`, or `youtubePartner#asset`; the write schema and typed CLI enum enforce those provider-owned values. |
| `groups.update` | body `id`, `snippet.title` | https://developers.google.com/youtube/analytics/reference/groups/update — Request body | operation_reference | high | The resource identifies the group and the provider documents title as the updatable property. |
| `groups.delete` | query `id` | https://developers.google.com/youtube/analytics/reference/groups/delete — Parameters | operation_reference | high | Provider marks the group identifier as required. |
| `groupItems.list` | query `groupId`, `onBehalfOfContentOwner` | https://developers.google.com/youtube/analytics/reference/groupItems/list — Parameters | operation_reference | high | Provider requires `groupId`; content-owner context is optional. |
| `groupItems.insert` | body `groupId`, `resource.id`, optional `resource.kind` | https://developers.google.com/youtube/analytics/reference/groupItems/insert — Request body; https://developers.google.com/youtube/analytics/reference/groupItem — Resource representation | operation_reference | high | Google requires `groupId` and `resource.id`. `resource.kind` is optional; when supplied, the provider resource type permits `youtube#channel`, `youtube#playlist`, `youtube#video`, or `youtubePartner#asset`. |
| `groupItems.delete` | query `id` | https://developers.google.com/youtube/analytics/reference/groupItems/delete — Parameters | operation_reference | high | Provider marks group-item ID required. |
| `reports.query` (planned) | query `ids`, `startDate`, `endDate`, `metrics` | https://developers.google.com/youtube/analytics/reference/reports/query — Parameters | operation_reference | high | Google requires channel/content-owner selection, dates, and at least one metric. It remains planned solely for issue #2985. |
| `jobs.list`; `reportTypes.list` | query `includeSystemManaged`, `onBehalfOfContentOwner`, `pageSize`, paginator `pageToken` | https://developers.google.com/youtube/reporting/v1/reference/rest/v1/jobs/list and https://developers.google.com/youtube/reporting/v1/reference/rest/v1/reportTypes/list — Parameters | operation_reference | high | All are optional filters or pagination controls; the connector intentionally sends its bounded page size. |
| `jobs.get` | path `jobId`, query `onBehalfOfContentOwner` | https://developers.google.com/youtube/reporting/v1/reference/rest/v1/jobs/get — Path and query parameters | operation_reference | high | The job identifier is the URL path key; content-owner context is optional. |
| `jobs.create` | body `reportTypeId`, `name` | https://developers.google.com/youtube/reporting/v1/reference/rest/v1/jobs/create — Request body and errors | operation_reference | high | Google states that the Job body sets both properties and lists 400 errors when either is absent. |
| `jobs.delete` | path `jobId` | https://developers.google.com/youtube/reporting/v1/reference/rest/v1/jobs/delete — Path parameters | operation_reference | high | Google marks the job ID mandatory. |
| `jobs.reports.list` | path `jobId`, query `createdAfter`, `onBehalfOfContentOwner`, `pageSize`, paginator `pageToken`, `startTimeAtOrAfter`, `startTimeBefore` | https://developers.google.com/youtube/reporting/v1/reference/rest/v1/jobs.reports/list — Parameters | operation_reference | high | Job ID selects the report collection; remaining fields are optional filters/continuation controls. |
| `jobs.reports.get` | path `jobId`, `reportId`, query `onBehalfOfContentOwner` | https://developers.google.com/youtube/reporting/v1/reference/rest/v1/jobs.reports/get — Path and query parameters | operation_reference | high | Both resource identifiers are path keys; content-owner context is optional. |
| `media.download` | path `resourceName` mapped safely to internal `path`, static query `alt=media` | https://www.googleapis.com/discovery/v1/apis/youtubereporting/v1/rest — `resources.media.methods.download`; https://developers.google.com/youtube/reporting/v1/reports — Step 6: Download the report | operation_reference + same_page_sibling | high | Google's Discovery document marks `resourceName` required and declares GET `v1/media/{+resourceName}` with `alt=media`; the provider guide directs an authorized GET to the returned download URL. The connector maps the value to safe multi-segment path expansion and never accepts an arbitrary URL. |

## Coverage accounting

- 16 documented operations: 15 reachable after `media.download` promotion; `reports.query` remains planned solely for typed provider-query foundation issue #2985.
- All declared request-field uses are covered by 15 provider-owned citation rows. Repeated shared fields cite their operation-table row once.
- Tier-5 deferrals: 0. The sole tier-4 supplement is `media.download`, explicitly flagged because Google's Reporting guide documents its download step on the report-retrieval page; the exact required path field is independently covered by the provider Discovery document.

## Mutation execution determination

Every documented mutation remains an executable, approval-gated `reverse_etl` command backed by a
specific `writes.json` action. None uses `rest_write`.

| Provider mutation | CLI command | Preserved write action | Determination |
| --- | --- | --- | --- |
| `jobs.create` | `jobs create` | `create_job` | Executable and preserved. |
| `jobs.delete` | `jobs delete` | `delete_job` | Executable and preserved. |
| `groups.insert` | `groups create` | `create_group` | Executable and preserved. |
| `groups.update` | `groups update` | `update_group` | Executable and preserved. |
| `groups.delete` | `groups delete` | `delete_group` | Executable and preserved. |
| `groupItems.insert` | `group-items create` | `create_group_item` | Executable and preserved. |
| `groupItems.delete` | `group-items delete` | `delete_group_item` | Executable and preserved. |
