# Provider-neutral foundation gaps — issue #4291

Machine-readable rows: `FOUNDATION-GAPS.json`. Every row remains an open, not-enabled operation
for the affected surface and cannot support a merge-ready verdict.

| Stable gap id | Affected operations | Connectors | Owner | Status | Closure proof |
| --- | ---: | ---: | --- | --- | --- |
| `declarative-typed-destination-app-dispatch` | 9 | 9 | #4304 / `fm/cli-reverse-etl-destination-r1` | open | Merge latest foundation; definition validation, surface sync, runtime preflight, and installed App/CLI fixture path. |
| `declarative-typed-destination-exact-action-selection` | 369 | 9 | #4304 / `fm/cli-reverse-etl-destination-r1` | open | Prove one exact eligible action can be selected without connector dispatch, then run definition/fixture/App-CLI checks. |
| `declarative-typed-destination-camelcase-input-identifier` | 1 | 1 | #4304 / `fm/cli-reverse-etl-destination-r1` | pending publish | Regression test accepts only schema-known `conversationId`, then validate Help Scout, surface sync, preflight, and fixture App/CLI path. |

Batch 6 currently fans out through Close, Outreach, Zoho Bigin, Braze, and Customer.io. Batch 7
currently fans out through Gorgias, ServiceNow, Chatwoot, Chargebee, and the blocked Help Scout
candidate. The JSON rows carry the exact provider operation and URL/version/hash trace for each.

No shared gap is represented as disabled, `N/A`, or a connector-specific workaround. `N/A` is
allowed only when a source-locked provider operation proves the corresponding capability absent.
