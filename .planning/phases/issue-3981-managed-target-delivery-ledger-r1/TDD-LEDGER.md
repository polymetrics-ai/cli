# TDD ledger — Issue 3981: durable target delivery ledger

Manual inline GSD TDD execution. Red and green command output is retained in
`traces/` before and after production changes.

| ID | Guarantee | Red assertion | Green proof |
| --- | --- | --- | --- |
| R1 | Immutable target delivery identity | A ledger lookup is absent because no ledger/key derives owner + target database + persisted relation identity. | A record written with a target ref rebuilt after source artifact rename is read through the same owner/target database/StreamID address. |
| R2 | Restart durability contract | A fresh ledger instance cannot observe an existing delivery record. | A fresh ledger instance using the same durable-store port retrieves the exact recorded value. |
| R3 | Per-relation isolation | Records under one owner/namespace can collapse to one mutable table/display key. | Two StreamIDs write/read distinct records; updating relation A preserves relation B byte-for-byte. |
| R4 | Fail-closed identity assertion | A mismatched owner/ref or invalid database identity can cause a store mutation. | Validation rejects the request and fake-store write count stays zero. |
| R5 | No transaction-spool substitution | A `CommittedTransactionStage` or source checkpoint can act as target delivery authority. | The ledger API depends only on its typed target identity/store port and tests use an independent fake store. |

## Red command

```sh
go test -timeout 20m ./internal/connectors/database -run 'TestManagedTargetDeliveryLedgerRenameAndRestart' -count=1
```

The base must fail because no `ManagedTargetDeliveryLedger` implementation or
test exists yet. The exact pre-change command result is retained at
`traces/ledger-red.txt`.

## Green commands

```sh
go test -timeout 20m ./internal/connectors/database -count=1
go test -timeout 20m ./internal/app -count=1
go test -race -timeout 20m ./internal/connectors/database -run 'TestManagedTargetDeliveryLedger' -count=1
```

The green proof must show non-empty observable records after rename/restart,
separate records for sibling StreamIDs, and zero persistence calls for invalid
requests. The exact focused green output is retained at
`traces/ledger-green.txt`.
