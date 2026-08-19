# Reverse-ETL readiness — issue #4292

This is a preparation audit, not a destination declaration. Issue #4303 must first add the connector-neutral typed destination factory. Until then, every source-backed `direct_write` below retains `generic-typed-destination-executor`; a named typed action is necessary but insufficient for reverse ETL.

A future connector-owned destination must additionally declare per-connector conformance evidence, explicit source-to-destination field bindings, acknowledgement, and per-sync-mode apply strategies. No `transport_binding` is declared here.

## Batch 8–10 readiness

| Batch | Connector | Provider source state | Direct-write operations | Action-backed rows | Unique typed actions | Typed-action authoring pending | Destination now |
| --- | --- | --- | ---: | ---: | ---: | ---: | --- |
| 8 | brex | mapped | 49 | 14 | 14 | 35 | none — #4303 foundation gap |
| 8 | zoho-books | mapped | 579 | 562 | 569 | 17 | none — #4303 foundation gap |
| 8 | testrail | skipped: no public API description | — | — | — | — | none — #4303 foundation gap |
| 8 | amplitude | mapped | 99 | 12 | 12 | 87 | none — #4303 foundation gap |
| 8 | posthog | mapped | 1134 | 0 | 0 | 1134 | none — #4303 foundation gap |
| 8 | metabase | mapped | 329 | 0 | 0 | 329 | none — #4303 foundation gap |
| 8 | dbt | mapped | 26 | 13 | 13 | 13 | none — #4303 foundation gap |
| 8 | looker | mapped | 222 | 0 | 0 | 222 | none — #4303 foundation gap |
| 8 | mode | mapped | 45 | 0 | 0 | 45 | none — #4303 foundation gap |
| 8 | dremio | mapped | 31 | 10 | 10 | 21 | none — #4303 foundation gap |
| 9 | coda | mapped | 58 | 8 | 8 | 50 | none — #4303 foundation gap |
| 9 | clickup-api | mapped | 107 | 20 | 20 | 87 | none — #4303 foundation gap |
| 9 | calendly | mapped | 22 | 8 | 8 | 14 | none — #4303 foundation gap |
| 9 | greenhouse | skipped: no public API description | — | — | — | — | none — #4303 foundation gap |
| 9 | lever-hiring | mapped | 51 | 14 | 14 | 37 | none — #4303 foundation gap |
| 9 | ashby | mapped | 113 | 98 | 98 | 15 | none — #4303 foundation gap |
| 9 | workable | mapped | 39 | 0 | 0 | 39 | none — #4303 foundation gap |
| 9 | recruitee | mapped | 563 | 0 | 0 | 563 | none — #4303 foundation gap |
| 9 | hibob | mapped | 142 | 0 | 0 | 142 | none — #4303 foundation gap |
| 9 | factorial | mapped | 94 | 0 | 0 | 94 | none — #4303 foundation gap |
| 10 | datadog | mapped | 989 | 27 | 27 | 962 | none — #4303 foundation gap |
| 10 | pagerduty | mapped | 254 | 0 | 0 | 254 | none — #4303 foundation gap |
| 10 | auth0 | mapped | 270 | 8 | 8 | 262 | none — #4303 foundation gap |
| 10 | okta | mapped | 444 | 429 | 429 | 15 | none — #4303 foundation gap |
| 10 | firehydrant | mapped | 204 | 190 | 190 | 14 | none — #4303 foundation gap |
| 10 | adobe-commerce-magento | dynamic: instance-dependent | — | — | — | — | none — #4303 foundation gap |
| 10 | commercetools | mapped | 355 | 0 | 0 | 355 | none — #4303 foundation gap |
| 10 | recharge | mapped | 76 | 0 | 0 | 76 | none — #4303 foundation gap |
| 10 | docuseal | mapped | 16 | 6 | 6 | 10 | none — #4303 foundation gap |
| 10 | eventbrite | skipped: no public API description | — | — | — | — | none — #4303 foundation gap |
| **Total mapped** | **26 connectors** | **4 skipped/dynamic** | **6311** | **1419** | **1426** | **4892** | **none** |

`Action-backed rows` are direct-write operations whose pinned route is already bound to a named typed `writes.json` action. `REVERSE-ETL-TYPED-ACTION-INVENTORY.json` preserves the exact source ID, route, source location, and action IDs for each of those 1,419 rows. `Typed-action authoring pending` is the remaining source-backed direct-write inventory; it is connector-local contract/safety/fixture work, not a shared engine gap. Neither artifact decides whether an operation is product-safe for reverse ETL.

## Zoom critical-path preparation

The captain directed that Zoom's 204-action destination cohort must gain destination declarations after #4303. That cohort is outside #4292's batch source inventory, so this report records **204 as a captain-directed planning target, not a provider-source operation total**. Current Zoom bundle evidence is read-only: `internal/connectors/defs/zoom/docs.md` records no declared Zoom write action. Once #4303 supplies the declaration schema, the Zoom lane must first identify the exact named typed action IDs and their pinned operations, then add the required per-connector evidence, explicit bindings, acknowledgement, and mode strategies. It must not add a `transport_binding` in advance.

## Regeneration and assertions

```bash
node .planning/phases/issue-4292-parity-batches-8-10-r1/traces/audit-reverse-etl-readiness.mjs
```

The script fails if a source-backed direct write lacks the locked foundation gap, an enabled direct write lacks a named typed action, any row uses primary `reverse_etl`, or a `transport_binding` appears.
