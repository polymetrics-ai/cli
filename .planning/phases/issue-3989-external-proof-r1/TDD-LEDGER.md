# #3989 TDD ledger

| Slice | Red | Green | Refactor | Evidence |
| --- | --- | --- | --- | --- |
| External ephemeral run | **Red:** `go test -v -count=1 -run '^TestCertificationEphemeralCredentials' ./internal/app` failed because `app.BeginCertificationEphemeralCredentials` and `app.EphemeralCredential` were undefined. | **Green:** `go test -v -count=1 -run '^TestCertificationEphemeralCredentials' ./internal/app` passed. The test resolves the in-memory secret and an overridden config through the normal runtime path while asserting zero persisted credentials plus the absence of `.polymetrics/vault` and `vault/key`. | Pending | Must assert absent canary and present fingerprints. |
| Evidence refusal boundary | Pending — add failing/no-artifact cases first | Pending | Pending | Each refusal asserts zero artifact writes. |
| Bounded protocol observation | **Red:** `go test -timeout 20m ./internal/connectors/certify/... -run '^TestObservedTransport' -count=1` failed: `undefined: certify.NewObservedTransport` in both canary and bounded-response tests. | **Green:** `go test -v -count=1 -run '^TestObservedTransport' ./internal/connectors/certify` passed. It asserts one TLS exchange's exact request/response values and a 65-byte error body retained as a 32-byte marked truncation while the caller still receives all 65 bytes. | Pending | Assert count, bytes, truncation, and preserved child response. |
| Artifact/smoke | Pending — external stdout and filesystem-audit tests first | Pending | Pending | Assert observed process output equality and no vault key/profile. |

No production edit has occurred before this ledger.
