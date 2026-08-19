# Source-accounted parity map — batches 2 and 3

All source artifacts were fetched credential-free from their recorded public documentation URLs. No provider operation, credential, write, or live certification was used. A connector marked `partial` is an explicit delivery hold, not a complete-source assertion.

| Connector | Old api_surface | New api_surface | Operations found | Coverage confidence and basis | Enabled | Disabled | Enabled % | Deletes | Foundation gaps |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- | --- |
| grafana | 11 | 315 | 314 | complete: provider-published machine-readable OpenAPI, Swagger, Discovery, or service-model document | 0 | 314 | 0.00 | 0/44 | generic-typed-destination-executor (reverse ETL) |
| trello | 7 | 263 | 261 | complete: provider-published machine-readable OpenAPI, Swagger, Discovery, or service-model document | 0 | 261 | 0.00 | 0/37 | generic-typed-destination-executor (reverse ETL) |
| slack | 10 | 176 | 174 | complete: provider-published machine-readable OpenAPI, Swagger, Discovery, or service-model document | 0 | 174 | 0.00 | 0/0 | generic-typed-destination-executor (reverse ETL) |
| n8n | 67 | 108 | 100 | complete: provider-published machine-readable OpenAPI, Swagger, Discovery, or service-model document | 26 | 74 | 26.00 | 0/15 | generic-typed-destination-executor (reverse ETL) |
| google-calendar | 38 | 38 | 38 | complete: provider-published machine-readable OpenAPI, Swagger, Discovery, or service-model document | 27 | 11 | 71.05 | 4/4 | generic-typed-destination-executor (reverse ETL) |
| gmail | 79 | 79 | 79 | complete: provider-published machine-readable OpenAPI, Swagger, Discovery, or service-model document | 56 | 23 | 70.89 | 9/10 | generic-typed-destination-executor (reverse ETL) |
| twilio | 197 | 197 | 197 | complete: provider-published machine-readable OpenAPI, Swagger, Discovery, or service-model document | 94 | 103 | 47.72 | 32/32 | generic-typed-destination-executor (reverse ETL) |
| amazon-sqs | 23 | 23 | 23 | complete: provider-published machine-readable OpenAPI, Swagger, Discovery, or service-model document | 22 | 1 | 95.65 | 0/0 | generic-typed-destination-executor (reverse ETL) |
| elasticsearch | 5 | 818 | 845 | complete: provider-published machine-readable OpenAPI, Swagger, Discovery, or service-model document | 0 | 845 | 0.00 | 0/72 | generic-typed-destination-executor (reverse ETL) |
| gong | 69 | 69 | 69 | complete: provider-published OpenAPI document | 57 | 12 | 82.61 | 3/3 | generic-typed-destination-executor (reverse ETL) |
| google-ads | 164 | 162 | 163 | complete: provider-published Google Discovery document | 28 | 135 | 17.18 | 0/1 | generic-typed-destination-executor (reverse ETL) |
| facebook-marketing | 36 | 1446 | 1445 | complete: all current provider-published Facebook Business SDK code-generation API declarations; Graph routes are explicitly documented as object-ID node/edge templates, so totals count named owner-type/method/edge declarations rather than fabricated runtime identifiers | 3 | 1442 | 0.21 | 0/112 | generic-typed-destination-executor (reverse ETL) |
| linkedin-ads | 10 | 287 | 272 | complete: provider sitemap plus every current LinkedIn Marketing rendered-reference page | 0 | 272 | 0.00 | 0/35 | generic-typed-destination-executor (reverse ETL) |
| aircall | 93 | 123 | 93 | complete: provider's single complete rendered Public API reference; every rendered endpoint index entry is parsed, while the two literal example IDs are excluded in favour of their documented parameter templates | 19 | 74 | 20.43 | 6/13 | generic-typed-destination-executor (reverse ETL) |
| xero | 235 | 235 | 235 | complete: provider-published Xero Accounting OpenAPI document | 87 | 148 | 37.02 | 10/10 | generic-typed-destination-executor (reverse ETL) |
| paypal-transaction | 10 | 10 | 2 | complete: provider-published Transaction Search OpenAPI document | 0 | 2 | 0.00 | 0/0 | generic-typed-destination-executor (reverse ETL) |
| gocardless | 142 | 142 | 137 | complete: provider-published GoCardless OpenAPI document served to its public API reference | 75 | 62 | 54.74 | 1/1 | generic-typed-destination-executor (reverse ETL) |
| amazon-seller-partner | 353 | 370 | 370 | complete: all provider-published Selling Partner OpenAPI model documents under models/ | 98 | 272 | 26.49 | 8/13 | generic-typed-destination-executor (reverse ETL) |
| miro | 197 | 197 | 197 | complete: provider-published Miro OpenAPI document linked by the API reference | 104 | 93 | 52.79 | 34/34 | generic-typed-destination-executor (reverse ETL) |

Old api_surface counts are from `acb85dc03` (the current-main revision named in the transport correction); new counts are the source-derived projection at this revision. A new api_surface count can be lower than operations found when the provider publishes multiple operation declarations with the same normalized request shape. Total operations found across pinned source artifacts: **5014**. This is not a self-referential coverage percentage: the per-connector confidence and basis state whether the input is complete or partial. Un-authored endpoint declarations are `declaration-pending`; a typed write remains enabled `direct_write`, while its nested reverse-ETL eligibility records the real `generic-typed-destination-executor` foundation gap at `internal/app/issue_label_warehouse_transport.go:85-95`.

ETL source transport is declaration-pending until each connector authors `sync_transport.json` with exact source executor, delivery, and conformance evidence. Reverse-ETL eligibility for typed direct writes remains foundation-blocked: the current only destination DefinitionFactory enforces the GitHub issue-label contract, so no transport binding/action is invented.
