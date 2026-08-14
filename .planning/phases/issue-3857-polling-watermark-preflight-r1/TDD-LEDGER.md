# TDD ledger — #3857 polling-watermark preflight

Each production behavior starts with a compiling, focused failing assertion.
No test calls a database, uses a credential, or treats a skipped test as proof.

| ID | Acceptance criterion | Red evidence | Green evidence |
| --- | --- | --- | --- |
| R1 | Correct declaration resolves registered executors and can proceed | preflight API is absent and the guarded fake cannot read | resolved preflight authorizes exactly one fake source read and one emitted record |
| R2 | Absent executor is refused before source I/O | absent registration can reach the guarded sync | exact missing-source/apply error and read counter remains `0` |
| R3 | Cursor and ordering safety are mandatory | lossy/null/precision-coercing codec or missing tie-breaker is admitted | exact policy error and zero reads |
| R4 | Traversal and mutation safety are mandatory | unstable page/keyset or unsafe commit/overlap policy is admitted | exact error and zero reads |
| R5 | Target compatibility is mandatory | incompatible strategy or unsafe history target is admitted | exact target/history error and zero reads |
| R6 | Modes/delete truth derive from runtime preflight | unsupported canonical mode or delete-aware hard-delete polling is eligible | exact preflight error and no source read |
| R7 | Immutable corpus remains the contract | an executor with stale/missing corpus proof is admitted | exact immutable-corpus error and zero reads |

## Evidence classification

| Criterion | Evidence | Live/fake | Why this evidence is sufficient |
| --- | --- | --- | --- |
| Happy: valid declaration passes and sync proceeds | preflight returns a resolved result; a guarded source fake increments `reads` and records one emitted row only after that result | Fake | #3857 expressly excludes a native engine declaration, driver, credential, and live database; #3858/#3859 own real source/apply execution. The counter and emitted row are observable changes that a no-op cannot satisfy. |
| Sad: every distinct invalid declaration is refused before source I/O | table-driven exact-error tests with guarded source `reads == 0` | Fake | A fake is required to observe and prove the forbidden I/O boundary without violating the issue's no-live-database restriction. It is not a no-op assertion: any premature read increments the counter and fails. |
| Edge: null, empty, type boundaries, corpus-admitted cursor types | immutable v1 corpus-derived inputs assert the exact accepted/rejected policy and source-read counter | Fake | The immutable corpus is the executable shared contract. Its JSON samples include null, nanosecond timestamp, large numeric tie-breaker, and float coercion; no database implementation exists in this slice. |

## Run log

### Red

Pending. The next implementation action is to add the tests described above
and run their focused package command before adding production code.

### Green

Pending.
