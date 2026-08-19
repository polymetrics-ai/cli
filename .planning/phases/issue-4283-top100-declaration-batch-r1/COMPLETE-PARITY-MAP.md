# Batch 1 complete six-class map

Captain order 2026-08-19: map every source-locked operation before any
certification runs. Docker Hub at `3ee815c01` remains the accepted reference;
the other nine connector-local disposition and crosswalk artifacts now use the
same source-lock basis. `ENABLED%` is source operations with an implemented
CLI command, not an operation-inventory or declared-coverage number.

Every listed connector has no `sync_transport.json`. Its ETL and reverse-ETL
transport result is therefore `declaration-pending`, assessed against the
definition-owned contract in `docs/sync-transport-definition.md` from PR
#4286. The missing descriptor/evidence is connector declaration work, not a
foundation-lane request. None is inferred from provider REST documentation.

| Connector | Documented | Enabled | Disabled | ENABLED% | Deletes | Primary source classes: DR / DW / ETL / RETL / BR / BW | ETL transport | Reverse-ETL transport | gap_ids | declaration_pending_ids |
| --- | ---: | ---: | ---: | ---: | ---: | --- | --- | --- | --- |
| Docker Hub | 54 | 41 | 13 | 75.93 | 6 / 6 | 23 / 7 / 4 / 20 / 0 / 0 | declaration-pending | declaration-pending | `head-response-less-operation-executor`, `operation-scoped-rest-pagination` | `declaration-pending-dockerhub`, transport descriptors |
| Notion | 49 | 43 | 6 | 87.76 | 4 / 4 | 18 / 3 / 5 / 22 / 0 / 1 | declaration-pending | declaration-pending | none | `command-availability-notion`, `cli-command-contract-notion`, transport descriptors |
| Stripe | 589 | 8 | 581 | 1.36 | 1 / 32 | 254 / 323 / 5 / 3 / 4 / 0 | declaration-pending | declaration-pending | none | `typed-operation-contract-stripe`, transport descriptors |
| Bitbucket | 331 | 3 | 328 | 0.91 | 1 / 54 | 21 / 94 / 143 / 54 / 15 / 4 | declaration-pending | declaration-pending | none | `cli-command-contract-bitbucket`, `typed-operation-contract-bitbucket`, transport descriptors |
| GitLab | 1,755 | 4 | 1,751 | 0.23 | 0 / 212 | 699 / 1,006 / 4 / 0 / 46 / 0 | declaration-pending | declaration-pending | none | `typed-operation-contract-gitlab`, transport descriptors |
| CircleCI | 111 | 0 | 111 | 0.00 | 0 / 16 | 52 / 43 / 9 / 7 / 0 / 0 | declaration-pending | declaration-pending | none | `cli-surface-missing-circleci`, transport descriptors |
| Sentry | 223 | 0 | 223 | 0.00 | 0 / 35 | 117 / 103 / 3 / 0 / 0 / 0 | declaration-pending | declaration-pending | none | `cli-surface-missing-sentry`, transport descriptors |
| Vercel | 400 | 0 | 400 | 0.00 | 0 / 56 | 156 / 222 / 7 / 15 / 0 / 0 | declaration-pending | declaration-pending | none | `cli-surface-missing-vercel`, transport descriptors |
| Asana | 249 | 82 | 167 | 32.93 | 4 / 23 | 10 / 0 / 109 / 129 / 0 / 1 | declaration-pending | declaration-pending | none | `command-availability-asana`, transport descriptors |
| Jira | 617 | 295 | 322 | 47.81 | 0 / 89 | 292 / 319 / 3 / 0 / 3 / 0 | declaration-pending | declaration-pending | none | `typed-operation-contract-jira`, transport descriptors |

The Jira source lock contains 617 operations. The order's parenthetical 590
matches the pre-existing CLI-surface count, not the pinned-document count, so
the map uses the locked 617 as its documented denominator.

`DR`, `DW`, `ETL`, `RETL`, `BR`, and `BW` mean direct read, direct write,
ETL, reverse ETL, binary read, and binary write. Each source-operation row in
the connector-local declaration disposition has exactly one primary class and
names either its present foundation, a `declaration-pending` record, or one of
the five evidence-backed engine gaps.
Live credentialed certification remains pending for every connector.
