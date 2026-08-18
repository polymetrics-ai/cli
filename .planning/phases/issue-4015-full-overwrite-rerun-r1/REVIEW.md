# Code review — Full-overwrite rerun correctness

Review mode: inline standard review. The repository's canonical single-worker contract forbids spawning the GSD reviewer/fixer roles, so the official GSD review prompt was executed through the documented inline/manual fallback.

## Scope reviewed

- Source request construction in generic, run-scoped overwrite, serial Arrow, and pipelined Arrow paths.
- Six-mode semantic boundary test and fake request capture.
- Live full-overwrite mutation/replacement proof and dedicated incremental no-regression proof.
- Planning, TDD, verification, UAT, and trace evidence.

## Findings

No unresolved critical, warning, or informational finding.

Reviewed specifically:

1. The selector is based on canonical source semantics represented by mode: only `full_append` and `full_overwrite` discard the saved position. All incremental modes and change capture take the default resumable path.
2. The saved checkpoint is not mutated. Existing clone boundaries still defensively copy a retained checkpoint before executor dispatch.
3. `Resume` continues to carry source/account/generation identity. The change removes only the prior source position for full refreshes.
4. Destination planning, run-scoped overwrite publication, durable receipt read-back, final candidate selection, downstream acknowledgement, and checkpoint commit ordering are unchanged.
5. Both previously duplicated Arrow full-overwrite guards now use the shared rule; serial and pipelined behavior remain equivalent.
6. The mode-matrix unit test fails only the two intended full-refresh cases before the fix and proves every requested incremental mode retains its prior checkpoint afterward.
7. Live assertions use connector-owned/sanitized relation identifiers and fixed test-only SQL. Logs contain only generated relation names, counts, IDs, and non-secret seeded labels; no credential or approval token is exposed.
8. The database harness remains opt-in, pins the existing direct local Unix endpoint, owns its run-specific resources, and performs unconditional cleanup. No shared runtime was restarted.
9. The task adds no dependency, command/flag/output change, connector surface, generic shell/HTTP/SQL capability, or release-branch change.

## Static analysis note

The repository Makefile lint gate passed. A supplemental whole-package lint invocation over `internal/synctransport` and `internal/cli` surfaced 16 pre-existing findings; diff-scoped lint from the target branch reported `0 issues`. No suppression was added and no unrelated cleanup entered this PR.

## Verdict

Review-clean and ready for the PR/automatic-review gate.
