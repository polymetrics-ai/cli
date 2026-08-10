---
phase: issue-3985-connector-canon-r1
depth: standard
mode: inline-manual-fallback
status: clean
---

# Code review — Issue 3985 connector canon

The repository contract disallows spawning a code-review role for this single-worker delivery, so the
standard review was completed inline after the rebased implementation and focused validation.

## Scope reviewed

- `Makefile` and `scripts/tests/connector-canon.sh` for an executable, real-runtime guard rather
  than duplicated validation.
- Canon source pins, archives, index links, current correction markers, and r2/r1 supersession.
- The rebase resolution against #3970 to preserve source-pinned GitHub parity data while removing
  only stale certification claims.
- REST change evidence for #3972, #3978, #3987, and the amended child scopes.
- Generated website/catalog artifacts after their official generators ran.

## Findings

| Severity | Count | Disposition |
| --- | ---: | --- |
| Critical | 0 | None. |
| Warning | 0 | None. |
| Info | 1 | GitHub REST later returned a rate-limit response during a nonessential attempted P6 prose refinement; the already-created/attached P6 body and all required corrections were read back successfully, so no state claim relies on that failed refinement. |

## Review conclusions

- `connector-runtime-preflight` invokes `TestEveryImplementedCommandPassesRuntimePreflight` through
  `commandrunner`, so new executor kinds are covered by the real runtime admission path.
- `connector-canon-check` is deterministic and hermetic: it verifies source hashes and stable
  content/location markers without provider credentials or GitHub API access.
- The r2 report makes the only dependency change explicit: #3987 gates #3978 but does not affect
  active #3974. r1 is preserved byte-for-byte in the archive.
- Current GitHub source inventory remains generated from the merged source lock; the script no
  longer mistakes valid generated coverage rows for the void historical report.

## Verdict

**PASS.** No fix is required before verification/hand-off.
