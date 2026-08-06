# Rate-limit declaration population audit

## Authoritative source

`/Users/karthiksivadas/karthik-agent-workspace/data/cli-provider-artifact-sweep-r1/ledger.json`
is the sole candidate population. On 2026-08-06 it contained 359 records: 281 `done`, 71
`unknown`, and 7 `skipped`. This live read supersedes the earlier handoff snapshot (358 / 281 /
70 / 7).

## First-batch result

Every committed declaration matches a sweep record with `status: done`, a nonblank provider
artifact URL, and `scope_in_current_defs: true`:

`airtable`, `asana`, `auth0`, `basecamp`, `callrail`, `chargebee`, `clickup-api`, `clockify`,
`confluence`, `datadog`, `freshdesk`, `github`, `harvest`, `hubspot`, `intercom`, `jira`,
`klaviyo`, `monday`, `notion`, `okta`, `recurly`, `stripe`, `vercel`, `xero`, and
`zendesk-support`.

No declaration was removed by the population correction. Sweep `unknown` and `skipped` records
are not candidates for a rate-limit declaration, even if a matching connector definition exists.

## Deprecation candidates — separate from rate-limit unknowns

These sweep records show provider-surface retirement or replacement signals. They must be tracked
as connector lifecycle candidates and must not receive `rate_limits.json` files or rate-limit
`unknown` ledger records.

| Connector | Sweep evidence | Candidate disposition |
| --- | --- | --- |
| `breezometer` | Authoritative docs retired/redirected; no replacement provider artifact | Retired provider surface |
| `captain-data` | Named v3 documentation returned 404; published v1 is a distinct surface | Version-replacement review |
| `delighted` | Metadata documentation redirects to Qualtrics; former reference returned 410 | Retired provider surface |
| `goldcast` | Metadata-linked API docs and advertised API host returned provider Not Found | Provider-surface retirement review |
| `klaus-api` | Provider API host no longer resolves; company site redirects to Zendesk QA marketing | Retired/acquired provider review |
| `my-hours` | Provider API documentation returns Page not found and no developer reference remains | Provider-surface retirement review |
| `opsgenie` | Provider documentation records explicit end-of-support/migration status | Discontinued provider review |

`dolibarr`, `deputy`, `dwolla`, `gridly`, `instagram`, `kyve`, `mantle`, and `onfleet` remain
sweep-enumeration gaps rather than deprecation candidates: their records describe live providers
whose public operation artifacts are incomplete, deployment-specific, blocked, or non-enumerable.
