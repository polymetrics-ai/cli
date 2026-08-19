# Batch 1 complete six-class map

Captain order 2026-08-19: map every source-locked operation before any
certification runs. Docker Hub at `3ee815c01` remains the accepted reference;
the other nine connector-local disposition and crosswalk artifacts now use the
same source-lock basis. `ENABLED%` is source operations with an implemented
CLI command, not an operation-inventory or declared-coverage number.

Every listed connector has no `sync_transport.json`. Its ETL and reverse-ETL
transport result is therefore `gap`, assessed against the definition-owned
contract in `docs/sync-transport-definition.md` from PR #4286. The two
transport gaps require a connector-owned exact executor, conformance evidence,
and (for destinations) typed action bindings, acknowledgement and per-mode
apply strategies. None is inferred from provider REST documentation.

| Connector | Documented | Enabled | Disabled | ENABLED% | Deletes | Primary source classes: DR / DW / ETL / RETL / BR / BW | ETL transport | Reverse-ETL transport | gap_ids |
| --- | ---: | ---: | ---: | ---: | ---: | --- | --- | --- | --- |
| Docker Hub | 54 | 41 | 13 | 75.93 | 6 / 6 | 17 / 20 / 4 / 0 / 0 / 0 | gap | gap | `sync-transport-source-definition-dockerhub`, `sync-transport-destination-definition-dockerhub`, plus the five source-recorded Docker executor/secret gaps |
| Notion | 49 | 43 | 6 | 87.76 | 4 / 4 | 18 / 3 / 5 / 23 / 0 / 0 | gap | gap | `command-availability-notion`, `cli-command-contract-notion`, `sync-transport-source-definition-notion`, `sync-transport-destination-definition-notion` |
| Stripe | 589 | 8 | 581 | 1.36 | 1 / 32 | 254 / 323 / 5 / 3 / 4 / 0 | gap | gap | `typed-operation-contract-stripe`, `sync-transport-source-definition-stripe`, `sync-transport-destination-definition-stripe` |
| Bitbucket | 331 | 3 | 328 | 0.91 | 1 / 54 | 21 / 98 / 143 / 54 / 15 / 0 | gap | gap | `cli-command-contract-bitbucket`, `typed-operation-contract-bitbucket`, `sync-transport-source-definition-bitbucket`, `sync-transport-destination-definition-bitbucket` |
| GitLab | 1,755 | 4 | 1,751 | 0.23 | 0 / 212 | 699 / 1,006 / 4 / 0 / 46 / 0 | gap | gap | `typed-operation-contract-gitlab`, `sync-transport-source-definition-gitlab`, `sync-transport-destination-definition-gitlab` |
| CircleCI | 111 | 0 | 111 | 0.00 | 0 / 16 | 52 / 43 / 9 / 7 / 0 / 0 | gap | gap | `cli-surface-missing-circleci`, `sync-transport-source-definition-circleci`, `sync-transport-destination-definition-circleci` |
| Sentry | 223 | 0 | 223 | 0.00 | 0 / 35 | 117 / 103 / 3 / 0 / 0 / 0 | gap | gap | `cli-surface-missing-sentry`, `sync-transport-source-definition-sentry`, `sync-transport-destination-definition-sentry` |
| Vercel | 400 | 0 | 400 | 0.00 | 0 / 56 | 156 / 222 / 7 / 15 / 0 / 0 | gap | gap | `cli-surface-missing-vercel`, `sync-transport-source-definition-vercel`, `sync-transport-destination-definition-vercel` |
| Asana | 249 | 82 | 167 | 32.93 | 4 / 23 | 10 / 1 / 109 / 129 / 0 / 0 | gap | gap | `command-availability-asana`, `sync-transport-source-definition-asana`, `sync-transport-destination-definition-asana` |
| Jira | 617 | 295 | 322 | 47.81 | 0 / 89 | 292 / 319 / 3 / 0 / 3 / 0 | gap | gap | `typed-operation-contract-jira`, `sync-transport-source-definition-jira`, `sync-transport-destination-definition-jira` |

The Jira source lock contains 617 operations. The order's parenthetical 590
matches the pre-existing CLI-surface count, not the pinned-document count, so
the map uses the locked 617 as its documented denominator.

`DR`, `DW`, `ETL`, `RETL`, `BR`, and `BW` mean direct read, direct write,
ETL, reverse ETL, binary read, and binary write. Each source-operation row in
the connector-local declaration disposition has exactly one primary class and
names either its present foundation or a recoverable, evidence-backed gap.
Live credentialed certification remains pending for every connector.
