# Refs #4303 — preserved-history reconciliation

## Task Delivery Header

- Issue: Refs #4303 — make reverse ETL declarable by any connector.
- Base branch: `main` (frozen foundation rollup remains untouched).
- Merges into: additive layering onto the frozen foundation rollup → `main`.
- Delivery: one committed, unpushed reconciliation head; no remote update or rollup change.
- Working branch: `fm/cli-reverse-etl-destination-r1`.
- Task: retain the protected local, published remote, and failed-pipeline histories in one non-destructive merge; regenerate generated website data from canonical sources.
- Verification: ancestry checks, conflict-marker scan, website generator idempotence, and focused reverse-ETL/rate-parking regressions.

## Ancestry manifest

The guarded recovery run `01M0DYNQ9HSJBYS9YQ4MJR4JGR` preserved its unpublished head at `refs/no-mistakes/recover/01M0DYNQ9HSJBYS9YQ4MJR4JGR` (`5995473f`). Recovery correctly refused to move the diverged local branch. The local protected head `81666b92` already contained the remote evidence heads `d814875a` and `c8e75083`; merge commit `eb4fd0a2c` then joined that lineage with the preserved pipeline head without rewriting either parent.

| Acceptance criterion | Evidence | Observable assertion |
| --- | --- | --- |
| All protected histories remain reachable | live | `git merge-base --is-ancestor` succeeds for `81666b92`, `d814875a`, `c8e75083`, and `5995473f` against the reconciliation head. |
| Ledger evidence is not lost | live | The resolved ledger retains its initial-integration and pipeline-refinement Red/Green entries, with the later entry explicitly identified as effective behavior. |
| Website artifact is canonical | live | `cd website && pnpm run gen:website-data` followed by `git diff --exit-code` produces no diff. |
| Merge has no unresolved content | live | `git diff --check` and the conflict-marker scan produce no output. |
| Durable reverse-ETL recovery remains correct | live | Focused app, engine, coordination, and CLI tests exercise the persisted action dispatch and two-coordinator rate-parking boundaries. |

The branch is deliberately unpushed. No force-push, reset, rebase, squash, rollup mutation, or remote overwrite is authorized by this record.
