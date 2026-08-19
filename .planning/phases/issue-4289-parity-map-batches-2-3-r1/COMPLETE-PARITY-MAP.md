# Source-accounted parity map — batches 2 and 3

All source artifacts were fetched credential-free from their recorded public documentation URLs. No provider operation, credential, write, or live certification was used. A connector marked `partial` is an explicit delivery hold, not a complete-source assertion.

| Connector | Operations found | Coverage confidence | Enabled | Disabled | Enabled % | Deletes | Foundation gaps |
| --- | ---: | ---: | ---: | ---: | ---: | --- |
| grafana | 314 | complete: provider-published machine-readable OpenAPI, Swagger, Discovery, or service-model document | 0 | 314 | 0.00 | 0/44 | generic-typed-destination-executor (reverse ETL) |
| trello | 261 | complete: provider-published machine-readable OpenAPI, Swagger, Discovery, or service-model document | 0 | 261 | 0.00 | 0/37 | generic-typed-destination-executor (reverse ETL) |
| slack | 174 | complete: provider-published machine-readable OpenAPI, Swagger, Discovery, or service-model document | 0 | 174 | 0.00 | 0/0 | generic-typed-destination-executor (reverse ETL) |
| n8n | 100 | complete: provider-published machine-readable OpenAPI, Swagger, Discovery, or service-model document | 26 | 74 | 26.00 | 0/15 | generic-typed-destination-executor (reverse ETL) |
| google-calendar | 38 | complete: provider-published machine-readable OpenAPI, Swagger, Discovery, or service-model document | 27 | 11 | 71.05 | 4/4 | generic-typed-destination-executor (reverse ETL) |
| gmail | 79 | complete: provider-published machine-readable OpenAPI, Swagger, Discovery, or service-model document | 56 | 23 | 70.89 | 9/10 | generic-typed-destination-executor (reverse ETL) |
| twilio | 197 | complete: provider-published machine-readable OpenAPI, Swagger, Discovery, or service-model document | 94 | 103 | 47.72 | 32/32 | generic-typed-destination-executor (reverse ETL) |
| amazon-sqs | 23 | complete: provider-published machine-readable OpenAPI, Swagger, Discovery, or service-model document | 22 | 1 | 95.65 | 0/0 | generic-typed-destination-executor (reverse ETL) |
| elasticsearch | 845 | complete: provider-published machine-readable OpenAPI, Swagger, Discovery, or service-model document | 0 | 845 | 0.00 | 0/72 | generic-typed-destination-executor (reverse ETL) |
| gong | 69 | complete: provider-published OpenAPI document | 57 | 12 | 82.61 | 3/3 | generic-typed-destination-executor (reverse ETL) |
| google-ads | 163 | complete: provider-published Google Discovery document | 28 | 135 | 17.18 | 0/1 | generic-typed-destination-executor (reverse ETL) |
| facebook-marketing | 36 | partial: official Meta business SDK code-generation source; complete Marketing API model traversal remains outstanding | 5 | 31 | 13.89 | 0/2 | generic-typed-destination-executor (reverse ETL) |
| linkedin-ads | 10 | partial: official Microsoft LinkedIn Marketing reference portal; complete versioned REST.li reference traversal remains outstanding | 0 | 10 | 0.00 | 0/0 | generic-typed-destination-executor (reverse ETL) |
| aircall | 93 | complete: provider's complete rendered API reference; Aircall publishes no public machine-readable specification | 20 | 73 | 21.51 | 6/13 | generic-typed-destination-executor (reverse ETL) |
| xero | 235 | complete: provider-published Xero Accounting OpenAPI document | 87 | 148 | 37.02 | 10/10 | generic-typed-destination-executor (reverse ETL) |
| paypal-transaction | 2 | complete: provider-published Transaction Search OpenAPI document | 0 | 2 | 0.00 | 0/0 | generic-typed-destination-executor (reverse ETL) |
| gocardless | 142 | complete: provider's complete rendered OpenAPI reference; no downloadable machine-readable artifact is published | 76 | 66 | 53.52 | 1/1 | generic-typed-destination-executor (reverse ETL) |
| amazon-seller-partner | 370 | complete: all provider-published Selling Partner OpenAPI model documents under models/ | 98 | 272 | 26.49 | 8/13 | generic-typed-destination-executor (reverse ETL) |
| miro | 197 | complete: provider-published Miro OpenAPI document linked by the API reference | 104 | 93 | 52.79 | 34/34 | generic-typed-destination-executor (reverse ETL) |

Total operations found across pinned source artifacts: **3348**. This is not a self-referential coverage percentage: the per-connector confidence and basis state whether the input is complete or partial. Un-authored endpoint declarations are `declaration-pending`; a typed write remains enabled `direct_write`, while its nested reverse-ETL eligibility records the real `generic-typed-destination-executor` foundation gap at `internal/app/issue_label_warehouse_transport.go:85-95`.

ETL source transport is declaration-pending until each connector authors `sync_transport.json` with exact source executor, delivery, and conformance evidence. Reverse-ETL eligibility for typed direct writes remains foundation-blocked: the current only destination DefinitionFactory enforces the GitHub issue-label contract, so no transport binding/action is invented.
