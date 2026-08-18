# Issue 4015: Full-overwrite re-run correctness — Context

**Gathered:** 2026-08-18
**Status:** Ready for planning

## Task Delivery Header

- Issue: Refs #4015 — Production MVP certification
- Base branch: `integration/4015-mvp-flat-r1`
- Merges into: `fm/cli-full-overwrite-fix-r1` → `integration/4015-mvp-flat-r1` → `main`
- Delivery: Pull request open against `integration/4015-mvp-flat-r1`, with its API-reported base verified and required local checks green.
- Working branch: `fm/cli-full-overwrite-fix-r1`
- Task: Reproduce and repair the production data-correctness defect where a second `full_overwrite` run reports `records_read=0`, `records_loaded=0`, and leaves stale destination rows. Prove replacement by independent database read-back, and preserve checkpoint skipping for incremental modes.
- Verification: Red/green unit tests at the source-refresh/checkpoint boundary; opt-in PostgreSQL binary integration proof with exact first/second-run counts and named rows; incremental unchanged-source replay proof; targeted Go tests, vet/build, repository generated/snapshot gates, and final review.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| A second full-overwrite run re-reads the full source | live | A real PostgreSQL source is changed between binary runs; the second run reports the exact changed-source count rather than `0/0`. |
| Full overwrite replaces destination contents | live | An independent PostgreSQL query proves the target count and named sample changed, and a source row deleted before run two is absent afterward. |
| Full-refresh and incremental semantics stay distinct across the mode matrix | live | An orchestrator boundary test proves prior checkpoints are withheld only for `full_append` and `full_overwrite`, while every incremental mode receives its committed checkpoint. |
| Incremental replay continues to skip unchanged input | live | The existing real PostgreSQL binary flow performs a second `incremental_upsert` run and asserts `records_read=0`, `records_loaded=0` with unchanged independent target state. |

## Phase Boundary

Repair source-refresh checkpoint handling in the shared transport boundary. Full-refresh modes must start source extraction without a previous position; incremental modes and change capture retain durable resumption. Destination behavior remains governed by its declared mode, including run-scoped publication for overwrite.

## Locked Decisions

- The firstmate hypothesis is evidence to test, not an assumed diagnosis.
- The fix belongs at the source-semantics/checkpoint boundary if reproduction confirms the checkpoint causes the skip; it must not be a PostgreSQL-only or `full_overwrite`-only workaround.
- `full_append` and `full_overwrite` both use full-refresh source semantics and therefore re-read the whole source on every run.
- `incremental_append`, `incremental_dedupe`, `incremental_dedupe_history`, and `incremental_upsert` retain prior checkpoints and unchanged-source skip behavior.
- `full_overwrite` must publish one complete replacement generation; exact destination state, not exit status, is the proof.
- No Colima or Docker restart. No secret values may enter commands, artifacts, logs, commits, or the PR.

## Canonical References

- `AGENTS.md` — repository delivery lifecycle, transport correctness, safety, and verification rules.
- `.agents/agentic-delivery/contracts/issue-agent-contract.md` — issue-first TDD and PR evidence contract.
- `.agents/agentic-delivery/references/gsd-pi-adapter.md` — required GSD command resolution and inline fallback rules.
- `.agents/agentic-delivery/references/required-skills-routing.md` — required Go/runtime skill routing.
- `internal/app/sync_modes.go` — public mode to source/destination semantic mapping.
- `internal/synctransport/orchestrator.go` — shared source request and checkpoint boundary.
- `internal/synctransport/arrow_fast_path_pipeline.go` and `internal/synctransport/arrow_fast_path_controller.go` — transformed full-overwrite source request paths.
- `internal/cli/postgres_transport_binary_integration_test.go` — real-binary PostgreSQL transport and incremental replay proof harness.

## Existing Code Insights

- `RunRequest` carries both a resume identity and a prior checkpoint; only the checkpoint contains the source position that causes a polling executor to seek past previously read rows.
- The Arrow full-overwrite paths already suppress the prior checkpoint locally, which is precedent for replacement-generation behavior but duplicates the rule.
- The regular orchestrator passes the prior checkpoint unchanged to both generic and run-scoped full-overwrite source reads.
- PostgreSQL's run-scoped full-overwrite destination publishes via a shadow relation and independent receipt read-back, including a zero-row replacement; destination publication should not own source refresh policy.
- The existing opt-in binary harness already proves `incremental_upsert` run two is `0/0` and independently checks unchanged destination rows.

## GSD Runtime Note

The repository-local adapter was validated with `scripts/gsd doctor`, all five lifecycle commands were resolved with `scripts/gsd sources`, and `scripts/gsd prompt discuss-phase issue-4015-full-overwrite-rerun-r1 --auto` was executed inline. The task is a dispatched bug slice rather than a numbered `.planning/ROADMAP.md` phase, and the canonical single-worker contract forbids role spawning, so this phase uses the documented inline/manual GSD fallback without weakening TDD or evidence gates.

## Deferred Ideas

None. Release `0.2.0`, PR #4250, broader connector certification, and unrelated transport refactors remain out of scope.

