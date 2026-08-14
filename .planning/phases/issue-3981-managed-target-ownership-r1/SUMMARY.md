---
coverage:
  - id: D1
    description: "One owned namespace supports distinct immutable-stream relations"
    verification:
      - kind: unit
        ref: "TestManagedTargetProvisioningTruthTable/owned_namespace_allows_second_stream_relation"
        status: pass
      - kind: unit
        ref: "TestManagedTargetProvisioningConcurrentStreamsShareNamespaceOwner"
        status: pass
    human_judgment: false
  - id: D2
    description: "Stream identity survives persistence and rename"
    verification:
      - kind: unit
        ref: "TestStreamIDIsPersistedAndSurvivesStreamRename"
        status: pass
    human_judgment: false
---

# Summary — Issue 3981: managed-target ownership

Implemented the connector-neutral managed-target ownership kernel.

- The namespace now has its own database/native owner assertion and each stream
  relation has an independent control record.
- A second stream under an exact owned namespace creates a distinct relation
  rather than being rejected as a name collision.
- `StreamID` is allocated once, persisted/migrated once, collision-protected,
  and used in relation identity instead of the mutable source/table display text.
- Target-database and namespace-native identities are asserted by typed plans and
  controls, while physical names remain deterministic owner/stream hashes.

No DDL, driver, SQL, delivery ledger, schema evolution, mode application, or CLI
surface was added.
