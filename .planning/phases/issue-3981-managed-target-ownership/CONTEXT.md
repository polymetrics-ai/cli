# Context — Issue #3981 managed-target ownership and provisioning

## Locked delivery decisions

- This is the F2 shared, driver-neutral contract only. It has no PostgreSQL DDL,
  target driver, generic SQL, direct write/query surface, CDC, transport, or
  capability promotion.
- The managed-target owner derives only from the source-owned warehouse artifact
  identity: workspace ID, source connector ID, and source connection ID. It must
  reuse `warehouse.ArtifactIdentity`; credentials and display names are not part
  of the identity or physical target naming.
- A typed owner, managed-target reference, expected native relation identity, and
  schema hash/version form the complete immutable control record. A driver fake
  implements the narrow observation/create port; this phase does not select a
  real driver or persistent control-table representation.
- Planning is non-mutating. A mutation can happen only through a validated typed
  provisioning plan carrying the asserted owner. The executor re-observes while
  holding a target-scoped lock and fails closed on cancellation, races, or any
  unprovable state.
- The admission table is closed: create only when both target and control record
  are absent; admit only an exact owner/ref/native-ID/schema record; refuse
  missing, unreadable, foreign, colliding, moved/replaced, drifted, or orphaned
  targets. It never adopts arbitrary customer tables or evolves a managed table.

## Inputs reviewed

- #3981, #3972, #3974 topology and child/parent PR relationship through
  `gh-axi`; F1 is integrated into `feat/3972-postgres-parity` at
  `be561871e6bb7d1a5b54d7687743ef8396a2cafe`.
- The supplied PostgreSQL parity topology report and the shared implementation
  brief.
- `docs/migration/HANDOFF-CODEX.md`, connector canon, migration conventions,
  architecture design, issue-first delivery contract, and required routing
  references.

## Non-applicable integrations

No CLI command, generated help/manual, website, runtime service, credential, or
live database integration changes are in this contract slice. Driver fakes give
the required proof, so Podman/PostgreSQL integration is intentionally not run.
