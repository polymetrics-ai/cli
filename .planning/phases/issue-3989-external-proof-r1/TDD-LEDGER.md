# #3989 TDD ledger

| Slice | Red | Green | Refactor | Evidence |
| --- | --- | --- | --- | --- |
| External ephemeral run | Pending — add canary external-process test first | Pending | Pending | Must assert absent canary and present fingerprints. |
| Evidence refusal boundary | Pending — add failing/no-artifact cases first | Pending | Pending | Each refusal asserts zero artifact writes. |
| Bounded protocol observation | **Red:** `go test -timeout 20m ./internal/connectors/certify/... -run '^TestObservedTransport' -count=1` failed: `undefined: certify.NewObservedTransport` in both canary and bounded-response tests. | Pending | Pending | Assert count, bytes, truncation, and preserved child response. |
| Artifact/smoke | Pending — external stdout and filesystem-audit tests first | Pending | Pending | Assert observed process output equality and no vault key/profile. |

No production edit has occurred before this ledger.
