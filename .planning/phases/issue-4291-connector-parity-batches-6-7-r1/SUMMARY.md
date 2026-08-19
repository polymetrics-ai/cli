# Issue #4291 execution summary

## Reconciled enabled inventory

`enabled` means the ledger row carries an API-surface command binding or a typed write-action binding. Existing ETL stream bindings without either are recorded as `declaration-pending`.

| Connector | Documented | Enabled | Commands | Writes | Deletes | ENABLED% |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| close-com | 297 | 12 | 0 | 12 | 51 | 4.04 |
| outreach | 259 | 163 | 0 | 163 | 36 | 62.93 |
| salesloft | 12 | 0 | 0 | 0 | 0 | 0.00 |
| copper | 5 | 0 | 0 | 0 | 0 | 0.00 |
| zoho-bigin | 50 | 6 | 0 | 6 | 6 | 12.00 |
| klaviyo | 9 | 0 | 0 | 0 | 0 | 0.00 |
| braze | 95 | 29 | 0 | 29 | 4 | 30.53 |
| customer-io | 159 | 10 | 0 | 10 | 14 | 6.29 |
| intercom | 10 | 0 | 0 | 0 | 1 | 0.00 |
| freshdesk | 10 | 0 | 0 | 0 | 0 | 0.00 |
| segment | 188 | 0 | 0 | 0 | 27 | 0.00 |
| activecampaign | 61 | 0 | 0 | 0 | 9 | 0.00 |
| iterable | 4 | 0 | 0 | 0 | 0 | 0.00 |
| help-scout | 144 | 115 | 139 | 65 | 18 | 79.86 |
| gorgias | 114 | 104 | 108 | 61 | 18 | 91.23 |
| service-now | 22 | 6 | 0 | 6 | 3 | 27.27 |
| chatwoot | 148 | 94 | 101 | 60 | 18 | 63.51 |
| chargebee | 428 | 36 | 0 | 36 | 0 | 8.41 |
| square | 11 | 0 | 0 | 0 | 0 | 0.00 |
| braintree | 73 | 0 | 0 | 0 | 6 | 0.00 |

## Canonical directory names

- The issue's **close** map is in `internal/connectors/defs/close-com/`, the repository's canonical Close.com definition ID. Its public documentation capture succeeded; it is neither a retrieval failure nor a skipped connector.
- The issue's **servicenow** map is in `internal/connectors/defs/service-now/`, the repository's canonical ServiceNow definition ID. Its public documentation capture succeeded; it is neither a retrieval failure nor a skipped connector.

## Classification decisions

- Typed write actions are `direct_write`, and only a source row with such an action can be enabled as a write.
- Reverse ETL is represented under `declaration.reverse_etl`, never as an endpoint parity class. Every connector reports the corrected `generic-typed-destination-executor` foundation gap in its destination transport summary.
- Existing stream bindings remain ETL rows but are disabled `declaration-pending` until a runnable command or typed action binding exists.
