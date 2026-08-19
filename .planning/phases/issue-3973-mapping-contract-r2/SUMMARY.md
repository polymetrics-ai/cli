---
coverage:
  - id: D1
    description: Sealed V1 source-to-target mapping exposes typed target columns through the write plan.
    verification:
      - kind: unit
        ref: internal/connectors/database/mapping_contract_test.go:TestDatabaseWritePlanSealsMappingBeforeSessionMutation
        status: pass
    human_judgment: false
  - id: D2
    description: Mapping converts only exact/lossless values and refuses an unrepresentable type/value.
    verification:
      - kind: unit
        ref: internal/connectors/database/mapping_contract_test.go:TestMappingContractV1ProjectsLosslessValuesAndRoundTrips
        status: pass
      - kind: unit
        ref: internal/connectors/database/mapping_contract_test.go:TestMappingContractV1RefusesUnrepresentableMappingsAndValues
        status: pass
    human_judgment: false
  - id: D3
    description: Only a validated explicit tombstone can delete a target row; ordinary absence cannot.
    verification:
      - kind: unit
        ref: internal/connectors/database/mapping_contract_test.go:TestDatabaseWriteExecutorDeletesOnlyExplicitTombstones
        status: pass
      - kind: unit
        ref: internal/connectors/database/mapping_contract_test.go:TestDatabaseWriteExecutorRefusesTombstoneMismatchBeforeSessionMutation
        status: pass
    human_judgment: false
  - id: D4
    description: DeliveryReceiptV1 remains session durability evidence until the separate ledger records it.
    verification:
      - kind: unit
        ref: internal/connectors/database/database_write_session_test.go:TestDatabaseWriteExecutorConsumesApprovalBeforeOnePinnedBoundedSession
        status: pass
    human_judgment: false
---

# Summary — Issue #3973 mapping contract completion

`DatabaseWritePlan` now seals an immutable `MappingContractV1` and exact
explicit-tombstone count alongside its existing target/mode/key/batch/effect
binding. The contract uses the established closed `LogicalType`/`TypePlan`
vocabulary and projects values only when the declared mapping is exact or
lossless.

`DatabaseWriteInput` and `TombstoneEnvelope` add the only delete path to the
pinned session. A record being absent from its `Records` projection carries no
delete meaning; only a validated row-delete tombstone is delivered in a bounded
`WriteBatch`. The legacy `Execute` method deliberately builds an empty envelope
and remains non-delete-capable; consumers with deletes use `ExecuteInput`.

`DeliveryReceiptV1` is now the session return type. It remains plan-bound and
becomes downstream acknowledgement evidence only after the pre-existing
managed-target delivery ledger records its opaque identity.

The GSD lifecycle ran inline/manual because this issue is outside numbered
roadmap phases and the canonical worker forbids role spawning. Red and green
proof are retained under `traces/`.
