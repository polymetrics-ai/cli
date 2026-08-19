# Discussion log — #3987 four-path warehouse conformance

## Lifecycle

- Resolved command: `scripts/gsd prompt discuss-phase cli-3987-four-path-conformance-r1 --auto`.
- Manual inline fallback: the canonical single-worker contract forbids role spawning and this runtime has no compatible isolated Pi role executor. Decisions were therefore recorded directly from the issue, merged branch, and required shared context.

## Fixed decisions

| Area | Decision | Basis |
| --- | --- | --- |
| Stale mode clause | Exercise `incremental_dedupe_history` as executable where the branch admits it; do not reintroduce a refusal. | Shared captain context; merged #4187 and #4188. |
| Four directions | Treat each source/destination role pair as a separate named contract, not a roll-up of route tests. | #3987 and captain priority. |
| Mediator | Assert selection/persistence before either endpoint runs; source writes only its owned state and destination receives only a sealed/reopened workset. | #3987 warehouse-mediation invariant. |
| `change_capture` | Admit only PostgreSQL source work through its derived warehouse workset; refuse any destination-mode interpretation. | #3987 and current PostgreSQL declarations. |
| Certification truth | Never turn an unexecuted row into pass; fixture/matrix evidence stays distinct from #3978’s final live certification. | Shared context and #3978 exclusion. |
| Out-of-scope ownership | Do not touch `allStagesPassed`, PostgreSQL profile/transport adapter, GitHub read assertions, CLI dispatch, generic transport registration, or #3978. | Task brief. |

## Assumptions to validate in TDD

1. Existing narrow live/binary route tests cover individual directions but do not provide a single fail-closed conformance contract spanning the exact four canonical role pairs.
2. The matrix must derive the exact mode vocabulary from `synccontract.AllModes()` and descriptor admission, so new modes cannot be silently omitted.
3. A deliberate, temporary matrix-direction mismatch should fail the named direction after schema compilation, then be restored without committing the defect.
