# Discussion log — #4410 Sentry source-to-seven-lane matrix

The issue has no unresolved product choice: preserve the frozen source-lock denominator, keep execution unclaimed, retain DELETE and write rows, and create a connector-local matrix only. The implementation choice is fixed by the issue and parent brief:

- Current `origin/main` is the exact base and has no Sentry source directory.
- Batch R1 parent is evidence-only input; copy only missing retained source artifacts after exact byte/digest comparison.
- Use all seven lanes for every source row, with source-backed `mapped_unproven` or source-evidenced `not_applicable` states only unless the Atlas proves a real missing foundation.
- Do not reopen #4365 or use the Seer Models route override as source-membership or runtime authority.

No interactive question is needed because choosing a runtime interpretation would violate the scoped mapping-only task.
