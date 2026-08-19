# Complete parity map — batches 2 and 3

All source artifacts were fetched credential-free from their recorded public documentation URLs. No provider operation, credential, write, or live certification was used.

| Connector | Documented | Enabled | Disabled | Enabled % | Deletes | Foundation gaps |
| --- | ---: | ---: | ---: | ---: | ---: | --- |
| grafana | 314 | 0 | 314 | 0.00 | 0/44 | generic-typed-destination-executor (reverse ETL) |
| trello | 261 | 0 | 261 | 0.00 | 0/37 | generic-typed-destination-executor (reverse ETL) |
| slack | 174 | 0 | 174 | 0.00 | 0/0 | generic-typed-destination-executor (reverse ETL) |
| n8n | 100 | 26 | 74 | 26.00 | 0/15 | generic-typed-destination-executor (reverse ETL) |
| google-calendar | 38 | 27 | 11 | 71.05 | 4/4 | generic-typed-destination-executor (reverse ETL) |
| gmail | 79 | 56 | 23 | 70.89 | 9/10 | generic-typed-destination-executor (reverse ETL) |
| twilio | 197 | 94 | 103 | 47.72 | 32/32 | generic-typed-destination-executor (reverse ETL) |
| amazon-sqs | 23 | 22 | 1 | 95.65 | 0/0 | generic-typed-destination-executor (reverse ETL) |
| elasticsearch | 845 | 0 | 845 | 0.00 | 0/72 | generic-typed-destination-executor (reverse ETL) |
| gong | 69 | 57 | 12 | 82.61 | 3/3 | generic-typed-destination-executor (reverse ETL) |
| google-ads | 164 | 28 | 136 | 17.07 | 0/1 | generic-typed-destination-executor (reverse ETL) |
| facebook-marketing | 36 | 5 | 31 | 13.89 | 0/2 | generic-typed-destination-executor (reverse ETL) |
| linkedin-ads | 10 | 0 | 10 | 0.00 | 0/0 | generic-typed-destination-executor (reverse ETL) |
| aircall | 93 | 20 | 73 | 21.51 | 6/13 | generic-typed-destination-executor (reverse ETL) |
| xero | 235 | 87 | 148 | 37.02 | 10/10 | generic-typed-destination-executor (reverse ETL) |
| paypal-transaction | 10 | 0 | 10 | 0.00 | 0/0 | generic-typed-destination-executor (reverse ETL) |
| gocardless | 142 | 76 | 66 | 53.52 | 1/1 | generic-typed-destination-executor (reverse ETL) |
| amazon-seller-partner | 353 | 98 | 255 | 27.76 | 8/13 | generic-typed-destination-executor (reverse ETL) |
| miro | 197 | 104 | 93 | 52.79 | 34/34 | generic-typed-destination-executor (reverse ETL) |

Total documented operations: **3340**. Un-authored endpoint declarations are `declaration-pending`; a typed write remains enabled `direct_write`, while its nested reverse-ETL eligibility records the real `generic-typed-destination-executor` foundation gap at `internal/app/issue_label_warehouse_transport.go:85-95`.

ETL source transport is declaration-pending until each connector authors `sync_transport.json` with exact source executor, delivery, and conformance evidence. Reverse-ETL eligibility for typed direct writes remains foundation-blocked: the current only destination DefinitionFactory enforces the GitHub issue-label contract, so no transport binding/action is invented.
