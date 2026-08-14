# Context — Issue #3973 mapping, tombstone, and delivery receipt contract

## Task Delivery Header

- Issue: Closes #3973 — Postgres Parity: bind database apply to transactional write sessions
- Base branch: `integration/4015-mvp-flat-r1`
- Merges into: `integration/4015-mvp-flat-r1 -> main`
- Delivery: direct PR against `integration/4015-mvp-flat-r1` with green checks
- Working branch: `fm/cli-3973-mapping-contract-r2`
- Scope: complete the undelivered shared contract only: `MappingContractV1`, an explicit tombstone envelope through the sealed write plan/session, and `DeliveryReceiptV1`.
- Verification: `go test -timeout 20m ./internal/connectors/database/... ./internal/synccontract/... ./internal/synctransport/...` plus scoped local gates and CI.

## Fixed decisions

- The existing `internal/connectors/database` package owns the shared contract because it already owns `LogicalType`, `TypePlan`, managed-target control, and the driver-neutral write session. #3980, #3982, and #3983 consume its exported values; no PostgreSQL-local mapping is introduced.
- `MappingContractV1` describes ordered source-field to target-column bindings, nullability, and an already-proven exact/lossless `TypePlan`. It is immutable by construction, validates duplicate/unknown identifiers, and is sealed into `DatabaseWritePlan` alongside keys.
- Mapping values are projected only through the declared lossless type plan. A value is copied into the declared target representation or refused; no generic string/JSON fallback or implicit semantic cast exists.
- A `DatabaseWriteInput` transports ordinary records and a typed `TombstoneEnvelope` to the same bounded `WriteBatch`. The plan seals both record and tombstone counts. No missing record is a delete signal; only validated `synccontract.Tombstone` values can reach a write session.
- `DeliveryReceiptV1` is the session-returned, plan-bound confirmed-commit proof. It is separate from the target-side `ManagedTargetDeliveryLedger`: the executor records its opaque delivery identity in that ledger before a downstream acknowledgement becomes available.
- No DDL, SQL, PostgreSQL implementation, capability change, physical-absence delete, source checkpoint advance, CLI/help/docs/website surface, or fixes for #4125, #4136, or #4090 are in scope.

## Evidence table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Typed/versioned mapping reaches a sealed write plan | fake | Plan mapping returns ordered target columns and type plans; an approval made for one mapping refuses a plan with a changed mapping before the fake driver begins a session. |
| Type mapping is lossless or rejected | fake | A source `int32` maps to a target `int64` and reverse-projects exactly; narrowing or a value outside its declared source representation is rejected before a target record is produced. A fake is necessary because this layer intentionally has no database engine. |
| Tombstones are explicit and bounded | fake | A fake target keeps a seeded row when it is absent from an ordinary batch, then removes that row only after a validated explicit tombstone reaches `ApplyWriteBatch`; malformed input and mismatched counts cause zero session mutations. |
| Typed durable receipt composes with the ledger | fake | The driver returns `DeliveryReceiptV1`; the result exposes it only after the fake ledger records its delivery ID. A ledger-store failure yields no acknowledgement. |

