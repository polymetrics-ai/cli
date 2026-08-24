# TDD ledger — issue 4347 source-lock usability

| Slice | Red | Green | Refactor | Evidence |
| --- | --- | --- | --- | --- |
| Retain-specific validation | Planned: a no-form v3 lock fails current retain because import inventory validation runs before fetch | Pending | Pending | `go test -timeout 5m ./cmd/connectorgen -run TestSourceRetain...` |
| Canonical JSON identity | Planned: reordered JSON fails byte comparison although semantic JSON is equal | Pending | Pending | same focused suite |
| Wrong-source classification | Planned: 403 and drastic undersize are indistinguishable from generic source failure/drift | Pending | Pending | same focused suite |
| Live exact retention | Planned after unit green: built utility cannot currently discover parity locks | Pending | Pending | one built-binary command per exact connector |

The first implementation test run must replace each applicable `Planned` entry with an exact `Red:` command/result before production edits. Green evidence must assert retained artifact/manifest state, not merely exit status.
