# Issue #4292 seven-surface reconciliation

Generated entirely from the pinned source ledgers and existing connector-owned stream/action schemas. `declared-static` means structural, credential-free contract validation only; application-level generic destination dispatch remains pending the latest #4304 foundation integration and its installed App/CLI proof.

| Batch | Connector | Provider operations | Direct reads | Direct writes | Binary R/W | Streams | Selected destination proof | Eligible typed actions | CLI implemented/partial/declared |
| --- | --- | ---: | ---: | ---: | --- | ---: | --- | ---: | ---: |
| 8 | brex | 108 | 43 | 49 | 0/0 | 16 | update_vendor | 13 | 14/0/14 |
| 8 | zoho-books | 838 | 86 | 579 | 0/0 | 174 | update_bank_account | 338 | 10/559/569 |
| 8 | testrail | skipped | 0 | 0 | 0/0 | 13 | add_project | 9 | 0/10/10 |
| 8 | amplitude | 187 | 78 | 99 | 0/0 | 10 | update_annotation | 7 | 8/4/12 |
| 8 | posthog | 1943 | 807 | 1134 | 0/0 | 2 | — | 0 | 0/0/0 |
| 8 | metabase | 634 | 305 | 329 | 0/0 | 5 | — | 0 | 0/0/0 |
| 8 | dbt | 52 | 16 | 26 | 0/0 | 10 | create_job | 9 | 0/13/13 |
| 8 | looker | 433 | 206 | 222 | 0/0 | 5 | — | 0 | 0/0/0 |
| 8 | mode | 94 | 49 | 45 | 0/0 | 5 | — | 0 | 0/0/0 |
| 8 | dremio | 49 | 17 | 31 | 0/0 | 5 | create_user | 9 | 9/2/11 |
| 9 | coda | 124 | 59 | 58 | 0/0 | 7 | — | 0 | 0/8/8 |
| 9 | clickup-api | 173 | 58 | 107 | 0/0 | 8 | create_task | 16 | 0/20/20 |
| 9 | calendly | 61 | 27 | 22 | 0/0 | 12 | create_share | 1 | 6/2/8 |
| 9 | greenhouse | skipped | 0 | 0 | 0/0 | 69 | add_application_to_candidate_prospect | 20 | 127/0/127 |
| 9 | lever-hiring | 104 | 30 | 51 | 0/0 | 25 | reactivate_user | 2 | 60/0/60 |
| 9 | ashby | 193 | 9 | 113 | 0/0 | 71 | create_candidate | 6 | 173/14/187 |
| 9 | workable | 84 | 45 | 39 | 0/0 | 42 | create_department | 13 | 0/38/38 |
| 9 | recruitee | 938 | 370 | 563 | 0/0 | 5 | — | 0 | 0/0/0 |
| 9 | hibob | 207 | 65 | 142 | 0/0 | 3 | — | 0 | 0/0/0 |
| 9 | factorial | 155 | 61 | 94 | 0/0 | 5 | — | 0 | 0/0/0 |
| 10 | datadog | 1739 | 735 | 989 | 0/0 | 15 | create_monitor | 20 | 13/14/27 |
| 10 | pagerduty | 465 | 207 | 254 | 0/0 | 4 | — | 0 | 0/0/0 |
| 10 | auth0 | 469 | 192 | 270 | 0/0 | 7 | update_user | 7 | 6/2/8 |
| 10 | okta | 734 | 6 | 444 | 0/0 | 284 | create_api_v1_behaviors | 47 | 105/324/429 |
| 10 | firehydrant | 373 | 13 | 204 | 0/0 | 205 | create_connection | 25 | 0/244/244 |
| 10 | adobe-commerce-magento | dynamic | 0 | 0 | 0/0 | 10 | update_product | 4 | 0/0/0 |
| 10 | commercetools | 821 | 463 | 355 | 0/0 | 3 | — | 0 | 0/0/0 |
| 10 | recharge | 123 | 44 | 76 | 0/0 | 3 | — | 0 | 0/0/0 |
| 10 | docuseal | 23 | 3 | 16 | 0/0 | 4 | archive_submission | 5 | 9/0/9 |
| 10 | eventbrite | skipped | 0 | 0 | 0/0 | 5 | — | 0 | 0/0/0 |

Each typed write action has a machine-readable `reverse_etl_eligibility` disposition and `direct_write_cli_status` in `SEVEN-SURFACE-LEDGER.json`. `partial-blocked` has a directly invokable CLI path plus an exact closed-contract blocker; `declaration-pending-cli-binding` is an unfinished reachability obligation, not a safety exclusion. When more than one action is structurally representable, the declaration lists every eligible action but the current closed destination multiplicity selects one action per mode; unselected actions are explicitly pending the foundation's multi-action selection capability. Semantic exclusions name the exact record-schema incompatibility and remain subject to direct CLI reachability.
