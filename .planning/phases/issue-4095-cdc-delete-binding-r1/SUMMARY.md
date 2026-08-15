---
coverage:
  - id: D1
    description: PostgreSQL CDC deletes become explicit source-keyed tombstones and map through the sealed shared contract.
    verification:
      - kind: unit
        ref: internal/connectors/native/postgres/cdc_tombstone_test.go; internal/connectors/database/mapping_contract_test.go
        status: pass
    human_judgment: false
  - id: D2
    description: An explicit CDC tombstone closes a PostgreSQL history row, while source absence remains non-destructive.
    verification:
      - kind: integration
        ref: TestPostgresManagedTargetWorksetDeliveryLive; TestPostgresManagedTargetIncrementalDedupeHistoryLive
        status: pass
    human_judgment: false
---

# Summary — Issue 4095

Implemented the narrow PostgreSQL CDC-delete binding. `CDCDeleteTombstone`
creates deterministic source-keyed delete evidence from a pgoutput event; the
sealed `MappingContractV1.MapTombstone` is now the shared key-vocabulary
projection used by both direct history test delivery and immutable worksets.

**Red:** the first targeted test run failed because neither binding existed.

**Green:** targeted packages and the complete tagged PostgreSQL dbtest run
passed. The live workset proof observes that absence leaves a row present,
then its CDC-derived tombstone removes that row. The live history proof
observes retained versions whose current row becomes closed rather than
physically deleted.

No CLI, help, manual, website, generated connector, destination CDC mode, or
unrelated issue path changed.
