# CP11 F-01–F-08 TDD order correction

## Receipt and immutable boundaries

Firstmate correction `033.msg` was read on 2026-09-06 before any further production edit. The working tree was preserved exactly as found. Follow-up dependency clarification `034.msg` was read before running the prepared F-01 fixture.

- Active branch HEAD at receipt: `350831a44fb529dc6916551f1bfef3435342a315` (`docs(gsd): plan CP11 F01-F08 repair wave`).
- Required original-behaviour baseline: `7e014d00e2faf4ccf54e68b03dc1bb9c261463d3`, whose product source is equivalent to reviewed source `7294373166db75466e2c92269f7887f51ceaddc6`.
- Preserved uncommitted production/test patch identity: `dfaecb733728b9aacbd6ce0cdacc32574ddf10fe6a9326b2fc1541507cc0df55`, the SHA-256 of the exact binary diff command recorded in [PREMATURE-F01-EDIT-SNAPSHOT.patch](PREMATURE-F01-EDIT-SNAPSHOT.patch).
- The snapshot records 246 patch lines across `cmd/connectorgen/vnext_publication_repair.go` and `cmd/connectorgen/vnext_publication_repair_test.go`; no other source path was dirty at receipt.

## Honest chronology

1. The historical F-07 exact RED had already been recovered in the canonical ledger. It is relevant context, not new F-01/F-02/F-03 RED evidence.
2. The active worker inspected the F-01 pathname-identity/reopen sequence and changed production `vnext_publication_repair.go` to add `vNextPublicationOpenRecordedCaptureLocked` before executing a new F-01/F-02/F-03 negative regression.
3. It then added a direct substituted-capture test and a descriptor-bound snapshot rewrite, and ran an existing recovery matrix successfully. That later green run is not treated as a pre-edit RED or as acceptance of the repair.
4. No actual new F-01, F-02, or F-03 failure had executed before the first production change. The patch remains preserved but unaccepted and no further production repair is permitted until Group 1 original-behaviour RED evidence is recorded.

## Restored order and Group 1 protocol

Group 1 is F-04 safe snapshot observation followed by F-08 bounded real-process signal/harness cleanup. Their actual negative controls and coherent test-layer repairs are prerequisites for Group 2; F-01 is not a Group 1 regression and the prepared F-01 baseline fixture must not be run until those controls are recorded.

After Group 1, the first Group 2 regression is F-01: a recorded capture directory A is replaced by B after the original pathname identity check but before the original `openDirectory` call. The expected contract is a rejection that leaves B public and unmodified; the original code instead accepts B and can later use it as a control capture.

To demonstrate that behavior without discarding the active patch, a disposable hermetic fixture will be materialized from exact commit `7e014d00e2faf4ccf54e68b03dc1bb9c261463d3` under this worktree. Its sole overlay is an inert test seam at the historical check-to-open boundary plus the bounded test that coordinates A/B substitution. It is not a branch, writer, publication target, or fix; its complete source and overlay identities, command, failure output, and path limits will be appended here before the active repair resumes. The fixture remains available for later independent review.

F-02 and F-03 receive equivalent original-behaviour bounded negative controls before their still-unmodified production paths are changed. The fixture is prepared early solely to preserve the exact original baseline and is not evidence until the Group 1 F-04/F-08 records exist.
