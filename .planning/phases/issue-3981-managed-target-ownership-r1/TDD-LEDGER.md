# TDD ledger — Issue 3981: managed-target ownership

Manual inline GSD TDD execution. Red and green command output are retained in
`traces/` before and after production changes.

| ID | Guarantee | Red assertion | Green proof |
| --- | --- | --- | --- |
| R1 | Multi-stream namespace | A second stream in an owned connection namespace is classified as a collision. | An exact namespace owner plus absent requested relation/control calls the typed create path; it yields a distinct relation in the same namespace. |
| R2 | Namespace ownership | A relation create can proceed in a namespace with missing, unreadable, foreign, moved, or replaced ownership. | Each namespace-owner mismatch is a typed refusal before mutation. |
| R3 | Per-relation proof | A missing, foreign, colliding, orphaned, replaced, or drifted relation can be adopted/repaired. | The existing fail-closed relation assertions remain exact and are checked after every create. |
| R4 | Immutable stream identity | Relation address is derived from mutable table/name text. | `StreamID` is allocated/persisted once and managed address equality/hash use it; a display/map-key/table rename returns the identical address. |
| R5 | Connection isolation | New and existing connections can share a namespace. | Same connection reuses namespace; different opaque connection IDs derive different namespaces. |
| R6 | Typed mutation only | A driver can mutate from an unasserted/untyped target. | The driver receives only `ManagedTargetProvisioningPlan` with exact owner/target/database assertions. |
| R7 | Concurrent/cancel-safe provisioning | Two stream creates can race namespace initialization or canceled work mutates. | Namespace-scoped local/driver locking, post-create reassertion, race test, and cancellation tests pass with driver fakes. |
| R8 | Confidential structural names | Display, credential, mode, mapping, or target DB data affects/leaks through physical names. | Names remain fixed lower-case hashes of structural owner + stream ID only. |

## Red command

```sh
go test ./internal/connectors/database -run 'TestManagedTargetProvisioningTruthTable/owned_namespace_allows_second_stream_relation' -count=1
```

On the base this must fail because `assessManagedTargetObservation` returns
`ErrManagedTargetNameCollision` for namespace-present/relation-absent/control-
absent. The exact pre-change output is recorded as `traces/second-stream-red.txt`.

## Green commands

```sh
go test -timeout 20m ./internal/connectors/database -count=1
go test -timeout 20m ./internal/app -count=1
go test -race -timeout 20m ./internal/connectors/database -run 'TestManagedTargetProvisioning' -count=1
go test -race -timeout 20m ./internal/app -run '^Test(StreamIDIsPersistedAndSurvivesStreamRename|AllocateUniqueIdentityRetriesCollisions)$' -count=1
```

## Green evidence

The required second-stream command was rerun after the ownership split and
passed. The two-stream concurrency fake then passed under `-race`, as did the
persisted-stream-ID migration/rename and generated-ID collision retry tests.

```text
ok   polymetrics.ai/internal/connectors/database
ok   polymetrics.ai/internal/app
```

The green tests prove that a namespace-owner record is asserted before the
second relation is created, that the two relations differ while their namespace
does not, and that an unchanged `StreamID` keeps a renamed source artifact at
the exact same managed address.
