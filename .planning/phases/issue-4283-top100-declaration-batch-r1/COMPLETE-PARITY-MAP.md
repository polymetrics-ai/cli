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

| Connector | Operations found / mapped | Input confidence | Enabled | Disabled | ENABLED% | Deletes | Endpoint classes: DR / DW / ETL / BR / BW | ETL source | Reverse-ETL eligibility | gap_ids | declaration_pending_ids |
| --- | ---: | --- | ---: | ---: | ---: | ---: | --- | --- | --- | --- | --- |
| Docker Hub | 54 / 54 | high — OpenAPI | 45 | 9 | 83.33 | 6 / 6 | 22 / 27 / 4 / 1 / 0 | declared (4 streams) | foundation-gap (0 eligible) | `operation-kind-loader-registration`, `typed-action-content-type`, `generic-typed-destination-executor` | `declaration-pending-dockerhub` |
| Notion | 49 / 49 | high — OpenAPI | 43 | 6 | 87.76 | 4 / 4 | 18 / 25 / 5 / 0 / 1 | declared (6 streams) | foundation-gap (0 eligible) | `generic-typed-destination-executor` | `command-availability-notion`, `cli-command-contract-notion` |
| Stripe | 589 / 589 | high — OpenAPI | 8 | 581 | 1.36 | 1 / 32 | 254 / 326 / 5 / 4 / 0 | declared (5 streams) | foundation-gap (0 eligible) | `generic-typed-destination-executor` | `typed-operation-contract-stripe` |
| Bitbucket | 331 / 331 | high — OpenAPI | 3 | 328 | 0.91 | 1 / 54 | 21 / 148 / 143 / 15 / 4 | declared (143 streams) | foundation-gap (0 eligible) | `generic-typed-destination-executor` | `cli-command-contract-bitbucket`, `typed-operation-contract-bitbucket` |
| GitLab | 1,755 / 1,755 | high — OpenAPI | 4 | 1,751 | 0.23 | 0 / 212 | 699 / 1,006 / 4 / 46 / 0 | declared (4 streams) | foundation-gap (0 eligible) | `generic-typed-destination-executor` | `typed-operation-contract-gitlab` |
| CircleCI | 111 / 111 | high — OpenAPI | 0 | 111 | 0.00 | 0 / 16 | 52 / 50 / 9 / 0 / 0 | declared (9 streams) | foundation-gap (0 eligible) | `generic-typed-destination-executor` | `cli-surface-missing-circleci` |
| Sentry | 223 / 223 | high — OpenAPI | 0 | 223 | 0.00 | 0 / 35 | 117 / 103 / 3 / 0 / 0 | declared (4 streams) | foundation-gap (0 eligible) | `generic-typed-destination-executor` | `cli-surface-missing-sentry` |
| Vercel | 400 / 400 | high — OpenAPI | 0 | 400 | 0.00 | 0 / 56 | 156 / 237 / 7 / 0 / 0 | declared (9 streams) | foundation-gap (0 eligible) | `generic-typed-destination-executor` | `cli-surface-missing-vercel` |
| Asana | 249 / 249 | high — OpenAPI | 82 | 167 | 32.93 | 4 / 23 | 10 / 129 / 109 / 0 / 1 | declared (12 streams) | foundation-gap (0 eligible) | `generic-typed-destination-executor` | `command-availability-asana` |
| Jira | 617 / 617 | high — OpenAPI | 295 | 322 | 47.81 | 0 / 89 | 292 / 319 / 3 / 3 / 0 | declared (3 streams) | foundation-gap (0 eligible) | `generic-typed-destination-executor` | `typed-operation-contract-jira` |

Totals: 4,378 operations found in complete provider-published OpenAPI documents
and 4,378 mapped connector-locally; input confidence is high for every row.
This is not expressed as a declaration percentage. There are 480 command-backed
enabled operations,
3,898 source operations disabled (3,894 declaration-pending and four
operation-level engine gaps). There are 2,370 direct-write endpoints and 120
enabled direct-write bindings. Reverse-ETL
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

Docker Hub reverse-ETL preparation is recorded in
`internal/connectors/defs/dockerhub/sources/dockerhub-reverse-etl-action-audit.json`:
all 27 direct writes are identified, 20 already have source-backed typed
actions, two SCIM writes require a typed-action request-content-type extension
before they can be actions, and five credential lifecycle/session exchanges are
not reverse-ETL targets. No destination declaration is implied or authored.

Docker Hub repair after PR #4297: operation-scoped REST pagination enabled
`GET /v2/auditlogs/{account}` and `GET /v2/scim/2.0/Users`; the closed SCIM
media type enabled `POST /v2/scim/2.0/Users` and `PUT /v2/scim/2.0/Users/{id}`.
The three HEAD checks and CSV export do not flip yet: #4297 supplied their
executors, but the operation-kind loader/validation path at
`internal/connectors/engine/bundle.go:2451,2676,2705,2733` omits
`rest_status` and `text_export`. This is recorded as recoverable
`operation-kind-loader-registration`, not as a connector gap.
