---
coverage:
  - id: D1
    description: "Target delivery keys bind the asserted source owner, destination database, and immutable StreamID-derived relation, never a mutable artifact table."
    verification:
      - kind: unit
        ref: "TestManagedTargetDeliveryLedgerKeyBindsOwnerAndTargetDatabase"
        status: pass
      - kind: unit
        ref: "TestManagedTargetDeliveryLedgerRejectsInvalidKeyBeforeStoreMutation"
        status: pass
    human_judgment: false
  - id: D2
    description: "A ledger entry survives source-artifact rename and fresh-ledger restart."
    verification:
      - kind: unit
        ref: "TestManagedTargetDeliveryLedgerRenameAndRestart"
        status: pass
    human_judgment: false
  - id: D3
    description: "Sibling immutable relations under one owner namespace retain separate delivery records."
    verification:
      - kind: unit
        ref: "TestManagedTargetDeliveryLedgerSeparatesRelations"
        status: pass
    human_judgment: false
key_files:
  created:
    - internal/connectors/database/managed_target_delivery_ledger.go
    - internal/connectors/database/managed_target_delivery_ledger_test.go
  modified: []
---

# Summary — Issue 3981: durable target delivery ledger

Implemented the Production-MVP amendment's remaining shared foundation.

- `ManagedTargetDeliveryLedgerKey` derives its address only from a validated
  `ManagedTargetControlRecord`: asserted source owner, target database identity,
  and immutable StreamID-derived namespace/relation. It retains no mutable
  warehouse artifact table or display name.
- `ManagedTargetDeliveryLedger` validates keys and records before invoking a
  typed, driver-owned durable-store port, and fails closed on invalid identity,
  malformed stored evidence, or store error.
- Driver-neutral tests use a durable-store fake shared by fresh ledger instances
  to observe rename/restart persistence, owner/database isolation, sibling
  relation isolation, and zero writes on invalid identity.

No target driver, DDL, SQL, write session, source checkpoint, transaction-stage
reuse, mode application, CLI surface, or schema evolution was added.
