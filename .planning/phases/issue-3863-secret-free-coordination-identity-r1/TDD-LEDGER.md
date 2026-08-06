# TDD LEDGER — issue #3863 secret-free credential coordination identity

| ID | Enforcement | RED evidence | GREEN evidence | Refactor/verification |
| --- | --- | --- | --- | --- |
| R1 | One binding yields one opaque auth cohort; rate keys require declared scope | Pending: focused identity test must fail before the builder exists | Pending | Pending |
| R2 | Different declared rate subjects never share a budget, even with one binding | Pending: same-binding/different-subject test must fail before rate projection exists | Pending | Pending |
| R3 | The builder has no secret/revision input and never serializes a binding preimage | Pending: type/JSON/error containment test must fail before opaque builder exists | Pending | Pending |
| R4 | Explicit links require compatible provider family and auth profile; unlinked copies isolate | Pending: app integration test must fail before protected binding/link state exists | Pending | Pending |
| R5 | Existing credential migration and secret rotation preserve identity lifetime separation | Pending: migration/runtime test must fail before state migration/runtime handoff exists | Pending | Pending |
| R6 | CLI declarations/linking validate safely and documentation reflects the real surface | Pending: CLI/help/manual parity tests must fail before flags/action/docs exist | Pending | Pending |

## Red-test rule

Every listed contract must execute and fail against the unmodified relevant production code before
its production implementation changes. Tests must not contain, print, persist, or compare a secret
value; they model shared authentication through an explicit binding and non-secret scope subjects.
