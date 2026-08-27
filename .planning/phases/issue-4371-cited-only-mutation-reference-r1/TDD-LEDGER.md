# TDD ledger — issue 4371

## Planned slices

| Slice | Red assertion before production edit | Green assertion | Refactor guard |
| --- | --- | --- | --- |
| Cited-only non-executable mutation | Existing application adds a second runtime gap and then conflicts with strict closed-reference validation. | Incompatible disposition is rejected before output with the exact cited source ID/location; descriptor bytes cannot change. | Normal contract-complete non-executable dispositions stay byte-stable. |
| Cited-only partial mutation coverage | Partial disposition likewise attempts to add metadata/gap to an exact cited-only descriptor. | It is rejected before output and does not make the validator accept extra gaps. | Existing partial-coverage match/duplicate/mutation checks stay fail-closed. |
| Cohort and runtime truth | Salesloft/Copper source-reference evidence cannot reconcile after the contradiction. | Each retained operation is still visible exactly once with the sole source-contract-unavailable foundation state; no generated command claims implementation. | `missing_foundation` command behavior, if an existing unavailable command is present, precedes credential/provider I/O. |

## Red

Pending. Production code has not been edited. The exact failing command and
observed failure will be recorded before the green implementation.

## Green

Pending red evidence.

## Refactor

Pending green evidence.
