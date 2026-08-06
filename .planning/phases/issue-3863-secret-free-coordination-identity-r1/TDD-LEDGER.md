# TDD LEDGER — issue #3863 secret-free credential coordination identity

| ID | Enforcement | RED evidence | GREEN evidence | Refactor/verification |
| --- | --- | --- | --- | --- |
| R1 | One binding yields one opaque auth cohort; rate keys require declared scope | `go test ./internal/connectors -run '^TestCoordinationIdentity'` failed to compile because `CredentialBinding`, `CoordinationIdentity`, and `RateLimitScope` do not exist | `go test ./internal/connectors -run '^TestCoordinationIdentity' -count=1` passed | Typed auth/rate keys make an accidental interchange a compile error. |
| R2 | Different declared rate subjects never share a budget, even with one binding | The same focused identity command failed before a rate projection implementation exists | The focused connector suite passed for distinct subject, policy-kind, and auth/rate-domain tests | No default scope exists; unsupported/empty scope declarations fail closed. |
| R3 | The builder has no secret/revision input and never serializes a binding preimage | The same focused identity command failed before the opaque-only type/API exists | Connector identity tests plus `go test ./internal/app -run '^TestCredentialCoordination' -count=1` passed, including vault-file removal before identity derivation | `CredentialBinding` contains no secret/revision field; runtime identity retains opaque projections only. |
| R4 | Explicit links require compatible provider family and auth profile; unlinked copies isolate | `go test ./internal/app -run '^TestCredentialCoordination'` failed to compile: the request/meta fields, `LinkCredential`, and runtime identity field do not exist | `go test ./internal/app -run '^TestCredentialCoordination' -count=1` passed | Covers creation-time and existing-credential links, cross-connector compatibility, and unlinked isolation. |
| R5 | Existing credential migration and secret rotation preserve identity lifetime separation | The app red command failed before protected state migration/runtime handoff exists | The focused app suite passed migration and approval-revision separation tests | Migration creates isolated bindings without vault reads; rotation handling remains untouched. |
| R6 | CLI declarations/linking validate safely and documentation reflects the real surface | `go test ./internal/cli -run '^TestCredentialsCoordination'` ran and failed: add output omitted family/profile and link did not return a compatibility constraint | `go test ./internal/cli -run '^TestCredentialsCoordination' -count=1` passed | Embedded help, generated CLI docs, website reference/data, and golden transcripts are updated before final package gates. |

## Red-test rule

Every listed contract must execute and fail against the unmodified relevant production code before
its production implementation changes. Tests must not contain, print, persist, or compare a secret
value; they model shared authentication through an explicit binding and non-secret scope subjects.
