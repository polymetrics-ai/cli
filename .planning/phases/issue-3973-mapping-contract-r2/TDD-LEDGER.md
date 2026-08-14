# TDD ledger — Issue #3973 mapping contract completion

| ID | Guarantee | Red assertion | Green proof |
| --- | --- | --- | --- |
| R1 | Shared mapped target schema | A plan exposes only keys and a driver cannot inspect ordered target columns/types. | `MappingContractV1` returns copied ordered bindings and `DatabaseWritePlan.Mapping()` preserves them; a mapping mismatch consumes no approval/session. |
| R2 | Lossless values only | A mapped value can be narrowed, coerced to string, or silently accepted outside its logical source representation. | `int32 -> int64` target/source projection round trips exactly; narrowing plan and invalid value requests return a typed refusal with no projected record. |
| R3 | Explicit tombstones only | A fake target may delete a row merely because a record is absent from a batch. | The seeded fake row remains after a non-tombstone batch and is deleted only after `TombstoneEnvelope` delivers a validated key delete; rejected counts produce zero session calls. |
| R4 | Plan/approval binds delete authority | A plan can be approved with one tombstone count and executed with another. | Count mismatch is rejected before `BeginDatabaseWrite`; valid tombstones are bounded alongside records in `WriteBatch`. |
| R5 | Named durable receipt remains separate from ledger | A committed session receipt can become checkpoint authority without ledger persistence. | `DeliveryReceiptV1` is the session return type; a ledger-store failure leaves `DatabaseWriteResult.Receipt()` unavailable. |

## Red command

```sh
go test -timeout 20m ./internal/connectors/database -run 'TestMappingContractV1|TestDatabaseWriteExecutor.*Tombstone|TestDeliveryReceiptV1' -count=1
```

The pre-production source has no `MappingContractV1`, `TombstoneEnvelope`, or
`DeliveryReceiptV1`; the expected compile failure is retained at
`traces/red-run.txt`.

## Green commands

```sh
go test -timeout 20m ./internal/connectors/database/... -count=1
go test -timeout 20m ./internal/synccontract/... -count=1
go test -timeout 20m ./internal/synctransport/... -count=1
go test -race -timeout 20m ./internal/connectors/database -run 'Test(MappingContractV1|DatabaseWriteExecutor.*Tombstone)' -count=1
```

