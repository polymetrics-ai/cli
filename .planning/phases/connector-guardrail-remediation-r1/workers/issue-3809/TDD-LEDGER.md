# TDD LEDGER — issue 3809 curated icon-collapse authority

## Red-first contract

| Case | Before implementation | After implementation |
| --- | --- | --- |
| Curated bare row plus conflicting upstream URLs | `buildIconEntries` aborts at `mergeCollapsedIconEntry` before it can apply the curated row | The authored bare row is retained, no upstream URL is silently selected, and generation proceeds. |
| No curated row plus conflicting upstream URLs | Generator rejects the collapse, but its diagnostic lacks both raw record names, both URLs, and an actionable key | Generator still rejects; the error names source record, destination record, both URLs, and the exact bare curated key to author. |
| Existing policy guarantees | Existing tests protect bare keys, reviewed attribution, orphan curated owners, reviewed fallback rows, and shared asset path URL identity | All remain unchanged and green. |
| Generated registry and website invariant | The generator cannot complete against the current upstream registry | `make icons-generate` regenerates `internal/connectors/icon_data.json`; `pnpm run test:scripts` still passes exact registry-to-lockfile coverage. |

## Evidence log

- 2026-08-06: #3809 read in full with `gh-axi issue view 3809 --full`; it specifies the reproduction, no-goals, required collision diagnostic, generator-only registry update, and guarantees to preserve.
- 2026-08-06: `gh-axi issue subissue list 3809` returned no sub-issues.
- 2026-08-06: Pre-edit code audit found the fatal URL comparison in `cmd/iconregistrygen/main.go:257-259`, before curated rows are processed at `cmd/iconregistrygen/main.go:163-186`. The current `customer-io` canonical row is bare and implemented at `internal/connectors/icon_data.json:1070`.
- Pending red-test evidence: add and run the curated-conflict regression before production code.
- Pending green evidence: focused generator tests, public-artifact regeneration, coverage/website checks, scoped verification, and review.
