# Batch 1 complete endpoint map

Captain’s 2026-08-19 classification correction: a provider mutation is a
`direct_write` endpoint. Reverse ETL is not an endpoint class and a typed write
action never establishes it. It is a separate eligibility attribute that needs a
destination transport binding, per-sync-mode apply strategies, and a durable
acknowledgement contract.

Every documented operation has exactly one endpoint class: direct read, direct
write, ETL, binary read, or binary write. ETL source is separately declared in
each connector’s `sync_transport.json` through the definition-owned
`declarative_api/declarative_stream_source` executor. All direct-write rows
carry `reverse_etl_eligibility`; it is false for every connector because the
only destination factory remains the issue-label-specific one.

| Connector | Documented | Enabled | Disabled | ENABLED% | Deletes | Endpoint classes: DR / DW / ETL / BR / BW | ETL source | Reverse-ETL eligibility | gap_ids | declaration_pending_ids |
| --- | ---: | ---: | ---: | ---: | ---: | --- | --- | --- | --- | --- |
| Docker Hub | 54 | 41 | 13 | 75.93 | 6 / 6 | 23 / 27 / 4 / 0 / 0 | declared (4 streams) | foundation-gap (0 eligible) | `head-response-less-operation-executor`, `operation-scoped-rest-pagination`, `generic-typed-destination-executor` | `declaration-pending-dockerhub` |
| Notion | 49 | 43 | 6 | 87.76 | 4 / 4 | 18 / 25 / 5 / 0 / 1 | declared (6 streams) | foundation-gap (0 eligible) | `generic-typed-destination-executor` | `command-availability-notion`, `cli-command-contract-notion` |
| Stripe | 589 | 8 | 581 | 1.36 | 1 / 32 | 254 / 326 / 5 / 4 / 0 | declared (5 streams) | foundation-gap (0 eligible) | `generic-typed-destination-executor` | `typed-operation-contract-stripe` |
| Bitbucket | 331 | 3 | 328 | 0.91 | 1 / 54 | 21 / 148 / 143 / 15 / 4 | declared (143 streams) | foundation-gap (0 eligible) | `generic-typed-destination-executor` | `cli-command-contract-bitbucket`, `typed-operation-contract-bitbucket` |
| GitLab | 1,755 | 4 | 1,751 | 0.23 | 0 / 212 | 699 / 1,006 / 4 / 46 / 0 | declared (4 streams) | foundation-gap (0 eligible) | `generic-typed-destination-executor` | `typed-operation-contract-gitlab` |
| CircleCI | 111 | 0 | 111 | 0.00 | 0 / 16 | 52 / 50 / 9 / 0 / 0 | declared (9 streams) | foundation-gap (0 eligible) | `generic-typed-destination-executor` | `cli-surface-missing-circleci` |
| Sentry | 223 | 0 | 223 | 0.00 | 0 / 35 | 117 / 103 / 3 / 0 / 0 | declared (4 streams) | foundation-gap (0 eligible) | `generic-typed-destination-executor` | `cli-surface-missing-sentry` |
| Vercel | 400 | 0 | 400 | 0.00 | 0 / 56 | 156 / 237 / 7 / 0 / 0 | declared (9 streams) | foundation-gap (0 eligible) | `generic-typed-destination-executor` | `cli-surface-missing-vercel` |
| Asana | 249 | 82 | 167 | 32.93 | 4 / 23 | 10 / 129 / 109 / 0 / 1 | declared (12 streams) | foundation-gap (0 eligible) | `generic-typed-destination-executor` | `command-availability-asana` |
| Jira | 617 | 295 | 322 | 47.81 | 0 / 89 | 292 / 319 / 3 / 3 / 0 | declared (3 streams) | foundation-gap (0 eligible) | `generic-typed-destination-executor` | `typed-operation-contract-jira` |

Totals: 4,378 documented and declared (100%); 476 command-backed enabled;
3,902 source operations disabled (3,894 declaration-pending, five
operation-level engine gaps, and three schema-incompatible). There are 2,370
direct-write endpoints and 118 enabled direct-write bindings. Reverse-ETL
eligibility is zero of 2,370 because
`generic-typed-destination-executor` is the remaining shared foundation gap.

The exact gap evidence is
`internal/app/issue_label_warehouse_transport.go:85-95`: only
`issue_label_destination` is registered and it calls the closed
`issueLabelTransportConnectorContract`. The recoverable minimum is a
connector-neutral typed destination `DefinitionFactory` selected by the
definition, with per-connector evidence, explicit source bindings,
acknowledgement, and per-mode apply strategies. No `transport_binding` action
or destination declaration is fabricated.
