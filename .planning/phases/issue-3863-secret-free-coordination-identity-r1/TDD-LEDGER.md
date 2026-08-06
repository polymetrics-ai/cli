# TDD LEDGER — issue #3863 secret-free credential coordination identity

| ID | Enforcement | RED evidence | GREEN evidence | Refactor/verification |
| --- | --- | --- | --- | --- |
| R1 | One binding yields one opaque auth cohort; rate keys require declared scope | `go test ./internal/connectors -run '^TestCoordinationIdentity'` failed to compile because `CredentialBinding`, `CoordinationIdentity`, and `RateLimitScope` do not exist | Pending | Pending |
| R2 | Different declared rate subjects never share a budget, even with one binding | The same focused identity command failed before a rate projection implementation exists | Pending | Pending |
| R3 | The builder has no secret/revision input and never serializes a binding preimage | The same focused identity command failed before the opaque-only type/API exists | Pending | Pending |
| R4 | Explicit links require compatible provider family and auth profile; unlinked copies isolate | `go test ./internal/app -run '^TestCredentialCoordination'` failed to compile: the request/meta fields, `LinkCredential`, and runtime identity field do not exist | Pending | Pending |
| R5 | Existing credential migration and secret rotation preserve identity lifetime separation | The app red command failed before protected state migration/runtime handoff exists | Pending | Pending |
| R6 | CLI declarations/linking validate safely and documentation reflects the real surface | `go test ./internal/cli -run '^TestCredentialsCoordination'` ran and failed: add output omitted family/profile and link did not return a compatibility constraint | Pending | Pending |

## Red-test rule

Every listed contract must execute and fail against the unmodified relevant production code before
its production implementation changes. Tests must not contain, print, persist, or compare a secret
value; they model shared authentication through an explicit binding and non-secret scope subjects.
