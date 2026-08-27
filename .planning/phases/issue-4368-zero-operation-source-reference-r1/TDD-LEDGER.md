# TDD LEDGER — issue #4368 zero-operation source-reference foundation

| ID | Enforcement | RED evidence | GREEN evidence | Refactor/verification |
| --- | --- | --- | --- | --- |
| R1 | An intentionally empty rendered coverage document needs an explicit closed marker and all normal retained provenance/identity evidence. | Pending: add strict fixture cases before validator production edits. | Pending. | Marker must not alter OpenAPI, source-reference, bundle, or non-empty rendered lock wire semantics. |
| R2 | Missing, malformed, duplicate, unverified, accidental-empty, and mixed-invalid documents fail closed at the offending document/location. | Pending: table-driven strict fixtures. | Pending. | Validate before descriptor/projection side effects or provider I/O. |
| R3 | A valid coverage-only document is integrity-checked but emits no executable descriptor operation. | Pending: importer/projection fixture. | Pending. | Preserve deterministic ordering and reject mixed invalid inventories. |
| R4 | Amplitude 187, Dremio 49, Ashby 193, Workable 84, and HiBob 207 reconcile to 720 exact source-cited deferred rows. | Pending: exact fan-in test/invariant check. | Pending. | Prove one source identity, citation, provider ID, lane, and `missing_foundation` disposition per row; no fabricated action. |
| R5 | Deferred registry commands return structured missing-foundation before credential/provider work, while untouched runnable commands retain missing-credential behavior. | Pending: commandrunner boundary test. | Pending. | Preflight uses the real registry and records zero transport/record/mutation witnesses. |
