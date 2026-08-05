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
- 2026-08-06: RED `go test ./cmd/iconregistrygen -run TestBuildIconEntriesAllowsCuratedRowToResolveConflictingSourceURLs -count=1` failed before production code with `ambiguous source/destination icon collapse for "demo": conflicting source URLs`.
- 2026-08-06: GREEN `go test ./cmd/iconregistrygen -count=1` passed after the curated-key index became authoritative and the unresolved diagnostic gained raw upstream record identity and an operator-curated-key remedy.
- 2026-08-06: The required public-artifact regeneration initially passed `customer-io` and reached the independent shared path guard for `facebook-marketing`/`facebook-pages`. Treating the existing canonical registry as authored authority (rather than only special-review rows) lets both established registry rows settle that upstream disagreement while retaining the no-curated shared-path rejection test.
- 2026-08-06: GREEN `PM_ICON_REGISTRY_SOURCE='https://connectors.airbyte.com/files/registries/v0/oss_registry.json' make icons-generate` completed: `generated 554 connector icon entries and 5 SVG assets`. A key-sorted JSON comparison against the pre-generation registry had no semantic data changes; the generated file changes only its canonical field ordering.
- Pending green evidence: runtime coverage/website checks, scoped verification, and review.
