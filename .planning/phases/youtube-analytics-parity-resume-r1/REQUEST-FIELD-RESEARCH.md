# Provider request-field research matrix

This is the required research record while the shared machine-readable citation convention is landing. It is **not** a competing bundle format. When the convention reaches `main`, transfer these citations verbatim into its prescribed location and revalidate.

## Sources and method

Provider-owned Google reference pages were checked on 2026-08-05. Google provides prose parameter tables rather than OpenAPI request bodies for these APIs, so rows use tier 3 (`operation_reference`) unless explicitly marked tier 4 (`same_page_sibling`). Requiredness is provider wording, not inferred from the current bundle.

| Operation(s) | Declared request fields | Source URL / section | Evidence type | Confidence | Requiredness rationale |
| --- | --- | --- | --- | --- | --- |
| `groups.list` | query `id`, `mine`, `onBehalfOfContentOwner`, paginator `pageToken` | https://developers.google.com/youtube/analytics/reference/groups/list — Parameters | operation_reference | high | `id`, `mine`, and content-owner context are optional selectors; page token is optional continuation state. |
| `groups.insert` | body `snippet.title` | https://developers.google.com/youtube/analytics/reference/groups/insert — Request body | operation_reference | high | Provider Group resource requires a title when creating a group. |
| `groups.update` | body `id`, `snippet.title` | https://developers.google.com/youtube/analytics/reference/groups/update — Request body | operation_reference | high | The resource identifies the group and the provider documents title as the updatable property. |
| `groups.delete` | query `id` | https://developers.google.com/youtube/analytics/reference/groups/delete — Parameters | operation_reference | high | Provider marks the group identifier as required. |
| `groupItems.list` | query `groupId`, `onBehalfOfContentOwner` | https://developers.google.com/youtube/analytics/reference/groupItems/list — Parameters | operation_reference | high | Provider requires `groupId`; content-owner context is optional. |
| `groupItems.insert` | body `groupId`, `resource.kind`, `resource.id` | https://developers.google.com/youtube/analytics/reference/groupItems/insert — Request body | operation_reference | high | A group item must identify its destination group and embedded resource. |
| `groupItems.delete` | query `id` | https://developers.google.com/youtube/analytics/reference/groupItems/delete — Parameters | operation_reference | high | Provider marks group-item ID required. |
| `reports.query` (blocked) | query `ids`, `startDate`, `endDate`, `metrics` | https://developers.google.com/youtube/analytics/reference/reports/query — Parameters | operation_reference | high | Google requires channel/content-owner selection, dates, and at least one metric. It remains blocked on issue #2985. |
| `jobs.list`; `reportTypes.list` | query `includeSystemManaged`, `onBehalfOfContentOwner`, `pageSize`, paginator `pageToken` | https://developers.google.com/youtube/reporting/v1/reference/rest/v1/jobs/list and https://developers.google.com/youtube/reporting/v1/reference/rest/v1/reportTypes/list — Parameters | operation_reference | high | All are optional filters or pagination controls; the connector intentionally sends its bounded page size. |
| `jobs.get` | path `jobId`, query `onBehalfOfContentOwner` | https://developers.google.com/youtube/reporting/v1/reference/rest/v1/jobs/get — Path and query parameters | operation_reference | high | The job identifier is the URL path key; content-owner context is optional. |
| `jobs.create` | body `reportTypeId`, `name` | https://developers.google.com/youtube/reporting/v1/reference/rest/v1/jobs/create — Request body and errors | operation_reference | high | Google states that the Job body sets both properties and lists 400 errors when either is absent. |
| `jobs.delete` | path `jobId` | https://developers.google.com/youtube/reporting/v1/reference/rest/v1/jobs/delete — Path parameters | operation_reference | high | Google marks the job ID mandatory. |
| `jobs.reports.list` | path `jobId`, query `createdAfter`, `onBehalfOfContentOwner`, `pageSize`, paginator `pageToken`, `startTimeAtOrAfter`, `startTimeBefore` | https://developers.google.com/youtube/reporting/v1/reference/rest/v1/jobs.reports/list — Parameters | operation_reference | high | Job ID selects the report collection; remaining fields are optional filters/continuation controls. |
| `jobs.reports.get` | path `jobId`, `reportId`, query `onBehalfOfContentOwner` | https://developers.google.com/youtube/reporting/v1/reference/rest/v1/jobs.reports/get — Path and query parameters | operation_reference | high | Both resource identifiers are path keys; content-owner context is optional. |
| `media.download` | path `resourceName` mapped safely to internal `path`, static query `alt=media` | https://developers.google.com/youtube/reporting/v1/reports — Step 6: Download the report | same_page_sibling | high | The provider guide directs an authorized GET to the report download URL. Google client-library reference names required `resourceName`; connector maps it to safe multi-segment path expansion and never accepts an arbitrary URL. |

## Coverage accounting

- 16 documented operations: 15 reachable after `media.download` promotion; `reports.query` remains blocked solely on typed provider-query foundation issue #2985.
- 41 request-field uses are covered by 15 citation rows. Repeated shared fields cite their operation-table row once.
- Tier-5 deferrals: 0. The sole tier-4 citation is `media.download`, explicitly flagged because Google's Reporting guide documents its download step on the report-retrieval page rather than in a dedicated REST operation page.
