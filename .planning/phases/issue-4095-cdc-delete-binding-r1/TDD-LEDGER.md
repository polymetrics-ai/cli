# TDD ledger — Issue 4095

| ID | Guarantee | Red assertion | Green proof |
| --- | --- | --- | --- |
| R1 | CDC delete is explicit | `TestCDCDeleteTombstone...` cannot convert a PostgreSQL `delete` event into a valid tombstone before the binding exists. | The converted envelope has a non-empty deterministic identity, source LSN position, and exactly the configured source keys; inserts, updates, missing LSNs, and missing keys refuse. |
| R2 | Mapping is the only key rename | A source-keyed tombstone cannot be projected by `MappingContractV1` before the method exists. | The resulting JSON has only declared target keys, preserves their values, and refuses undeclared/missing/extra source key payloads. |
| R3 | CDC delete closes history | The pre-binding live test uses a hand-authored target-key tombstone rather than a CDC-derived source tombstone. | A live target query finds the history versions still retained, with the current row closed (`_is_current=false`, `_valid_to` set). |
| R4 | Absence is not delete | The live workset proof does not exercise a CDC-derived explicit delete after the absence observation. | An absent row remains in the target; the subsequent CDC-derived tombstone deletes that exact keyed non-history row through the mapped envelope. |

## Recorded red result

**Red:** The planned package command failed at compilation because neither
`MappingContractV1.MapTombstone` nor `postgres.CDCDeleteTombstone` existed.
The exact error summary is retained in `traces/red-cdc-delete-binding.txt`.

## Recorded green result

**Green:** `MapTombstone` projects only declared source keys, and
`CDCDeleteTombstone` derives deterministic explicit source evidence from a
pgoutput delete. The targeted package command passed; the tagged PostgreSQL
dbtest then proved R3/R4 against actual target state. See
`traces/green-targeted.txt` and `traces/green-live-postgres.txt`.

## Red command

```sh
go test -timeout 20m -count=1 ./internal/connectors/database/... ./internal/connectors/native/postgres/...
```

## Green commands

```sh
go test -timeout 20m -count=1 ./internal/connectors/native/postgres/... ./internal/connectors/database/...
POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock go test -tags=databaseintegration -count=1 -timeout 20m -v ./internal/connectors/native/postgres
```
