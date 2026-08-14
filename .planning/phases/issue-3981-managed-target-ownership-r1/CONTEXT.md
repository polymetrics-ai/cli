# Issue 3981 — managed-target ownership and provisioning: context

**Gathered:** 2026-08-14
**Status:** Ready for TDD implementation
**Source:** GitHub issue #3981, parent #3972, and `data/cli-reverse-etl-managed-table-design-r1/report.md`.

## Locked decisions

- Keep the existing length-framed SHA-256 name derivation and opaque physical
  names. Namespace identity is workspace + source connector + source connection;
  relation identity is that same owner plus immutable stream ID.
- `ArtifactRef.Table()` remains source provenance only. It is not managed-target
  identity, so a stream/table display rename cannot move a target relation.
- Replace the ambiguous single control observation with a namespace-owner record
  and a per-relation control record. An exact owned namespace with both the
  requested relation and its control absent is a legitimate create.
- Retain fail-closed refusal for missing/unreadable/foreign owner, collision,
  orphaned control, replacement, moved target, and schema drift. Never adopt an
  arbitrary relation and never evolve schema automatically.
- Persist a `StreamID` once per configured stream. Allocate it on connection
  creation and migrate legacy persisted streams once during state normalization;
  it survives map-key/display/table renames. Generated connection and stream IDs
  are checked for collisions before persistence.
- Control and plan assertions include opaque target-database and namespace-native
  identities, but neither is a physical-name input. There is no PostgreSQL
  literal, connector branch, DDL, SQL, write session, delivery ledger, mapping,
  mode application, or schema evolution in this issue.
- All mutations remain behind the existing typed provisioning plan and a
  connector-neutral driver port. Fakes prove state transitions; no generic SQL is
  exposed.

## Observed defect

`managed_target.go:65-72` derives one namespace for a connection but derives
relations from `artifact.Table()`. `managed_target_provisioning.go:296-304`
refuses a namespace-present/relation-absent/control-absent observation. Thus,
after the first stream creates its namespace, a distinct second stream under the
same connection necessarily reaches and is refused by that state.

## Scope

Production files are limited to `internal/connectors/database/**` and the
persisted `StreamID` migration/allocation in `internal/app/**`, plus focused unit
tests and this evidence. This is shared foundation work under #3981, not a
connector-specific implementation lane.

## Deferred

- Database-specific DDL, database-driver implementation, delivery ledger,
  write sessions, mode algorithms, mappings, target schema creation/evolution,
  and destination apply/read-back belong to their dedicated successor issues.
- User-facing CLI/docs/website surface is unchanged; parity is not applicable.
