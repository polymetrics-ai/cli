---
coverage:
  - id: AC1
    description: PostgreSQL history apply creates exact validity windows and a retained soft-delete close.
    requirement: "#4094"
    verification:
      - kind: integration
        ref: internal/connectors/native/postgres/managed_target_integration_test.go:TestPostgresManagedTargetIncrementalDedupeHistoryLive
        status: pass
    human_judgment: false
  - id: AC2
    description: Unsupported history routes return typed reasons before session or ledger mutation.
    requirement: "#4094"
    verification:
      - kind: unit
        ref: internal/connectors/database/history_route_test.go:TestIncrementalDedupeHistoryRefusesEachNonPostgresRouteBeforeSessionMutation
        status: pass
    human_judgment: false
  - id: AC3
    description: CDC deletes map to keyed target effects and history close after durable receipts.
    requirement: "#4095"
    verification:
      - kind: integration
        ref: internal/connectors/native/postgres/managed_target_integration_test.go:TestPostgresManagedTargetIncrementalDedupeHistoryLive
        status: pass
      - kind: integration
        ref: internal/connectors/native/postgres/managed_target_integration_test.go:TestPostgresManagedTargetWorksetDeliveryLive
        status: pass
      - kind: unit
        ref: internal/connectors/native/postgres/cdc_v2_test.go:TestPGOutputV2StreamCommitReceiptsBeforeCheckpointAndAcknowledgement
        status: pass
    human_judgment: false
  - id: AC4
    description: The shared polling apply adapter can construct the required history plan from definition-owned source and destination identities.
    requirement: "#3859 residual"
    verification:
      - kind: integration
        ref: internal/connectors/native/postgres/managed_target_integration_test.go:TestPostgresManagedTargetIncrementalDedupeHistoryLive
        status: pass
    human_judgment: false
---

# Summary — PostgreSQL apply/history club

The already-staged history target, CDC tombstone mapping, and shared keyed apply
path are preserved. This change closes the audited adapter residual: history
mode now obtains the immutable database definition from the exact registered
polling source, checks it against the declaration, seals it with the target
definition, and supplies that route to `NewDatabaseWritePlan`.

The production change is limited to the polling apply adapter. The live
PostgreSQL history proof now exercises that adapter for v1, v2, restart replay,
and a CDC-derived tombstone, requiring both real row-state changes and durable
delivery evidence. Non-history planning remains unchanged, unsupported history
routes retain their typed pre-I/O refusals, and no transport-registry or public
CLI surface changed.

Official GSD commands were resolved through the project-local adapter and run
inline because the repository contract forbids spawning lifecycle roles for
this combined delivery. The plan, red/green ledger, verification, and deep code
review capture that fallback and its evidence.
