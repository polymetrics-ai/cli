# Jira Track A discussion log

| Topic | Decision | Evidence / boundary |
| --- | --- | --- |
| Denominator | Keep all 617 locked REST source IDs as matrix rows. | Lock `counts.total` and `rest.operations` are both 617. |
| Source reconciliation | Keep the crosswalk as a 617-to-617 exact method/path backlink. | `jira-operation-crosswalk.json.accounting` reports zero source-only and surface-only rows. |
| ETL cohort | Use only GET plus exact retained `maxResults` query parameter. | 95 lock rows satisfy the predicate; the matrix preserves all other paging/cursor facts without promoting them. |
| Legacy streams | Backlink issues, projects, users to their exact existing stream paths only. | `streams.json` has three entries; all remain mapped-unproven. |
| Binary candidates | Use three exact avatar response-media rows and four exact upload media/text rows. | The selected IDs and media/text are locked and validator-checked. |
| Sync | Record dynamic webhook registration as an actual inbound-receiver gap. | Source request includes `url` and `webhooks[].events`; Atlas sync contract requires an already-registered executor. |
| Shared runtime | Do not implement the receiver. | Track A must await captain approval for any shared foundation delivery. |

No unresolved decision blocks the mapping artifact. The recorded webhook foundation demand blocks only a future runtime implementation, not source accounting.
