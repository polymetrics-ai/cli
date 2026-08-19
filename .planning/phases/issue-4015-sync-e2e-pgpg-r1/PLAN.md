# Issue 4015 Sync Pipeline E2E — Plan

## Objective

Produce live, independently read evidence for the PostgreSQL → PostgreSQL control pipeline and an explicit second-run `incremental_upsert` result, then record honest verdicts for all four requested routes without modifying product code.

## Lifecycle

- `scripts/gsd prompt discuss-phase issue-4015-sync-e2e-pgpg-r1 --auto`
- `scripts/gsd prompt plan-phase issue-4015-sync-e2e-pgpg-r1 --tdd --skip-research --auto`
- `scripts/gsd prompt execute-phase issue-4015-sync-e2e-pgpg-r1 --interactive --auto`
- `scripts/gsd prompt verify-work issue-4015-sync-e2e-pgpg-r1 --auto`
- `scripts/gsd prompt code-review issue-4015-sync-e2e-pgpg-r1 --depth=quick --files=.planning/phases/issue-4015-sync-e2e-pgpg-r1`

The commands are executed inline because spawning delivery roles is forbidden for this canonical single-worker task.

## Slice 1 — control environment and red evidence

1. Inspect Docker, Colima, and Podman read-only; do not start or restart an existing runtime.
2. Resolve `pm help etl`, `pm help etl transport`, `pm help connections`, and the PostgreSQL connector manifest before invoking the binary workflow.
3. Record the pre-run absence of the harness's unique database fixture resources. This is the certification Red state: no live destination receipt or read-back exists for this run yet.

## Slice 2 — PostgreSQL → PostgreSQL Green

1. Run the existing `databaseintegration` binary harness `TestPMBinaryExecutesPostgresWarehousePostgres` against the available explicit local container endpoint.
2. Require its live assertions: 1,001 source rows, batch size 1,000, completed warehouse-mediated load, durable target receipt, a second separately approved run, and unchanged target cardinality on replay.
3. Add an independently queried named content sample to the evidence. If the existing harness does not print it, query its target through a separate client path while the fixture is live or run an equivalent bounded standalone fixture.
4. Confirm the harness-owned container and volume are absent after cleanup.

## Slice 3 — optional GitHub routes

Only after Slice 2 is decisive, attempt routes in this order while time remains:

1. PostgreSQL → GitHub.
2. GitHub → PostgreSQL.
3. GitHub → GitHub.

GitHub credentials may only be read inline from the macOS Keychain into environment/stdin at point of use. Any created `pm-cert-` fixture is independently removed and its absence proved with a 404. If a complete safe proof cannot fit the remaining time, mark the route `not-attempted-and-why`.

## Slice 4 — verification and delivery

1. Write `VERIFICATION.md`, `SUMMARY.md`, and `REVIEW.md` with exact commands, red/green evidence, read-back results, incremental behavior, route verdicts, and cleanup.
2. Run targeted repository gates plus generated/planning checks; document any gate intentionally not run.
3. Commit, push only `fm/cli-sync-e2e-pgpg-r1`, open a Conventional Commit PR containing `Refs #4015`, and verify the API-reported base is exactly `integration/4015-mvp-flat-r1`.

## Checkpoints

- Plan checkpoint: context, plan, and TDD ledger committed before live execution.
- Green checkpoint: control evidence and cleanup recorded after live assertions pass or fail decisively.
- Review checkpoint: verification/review artifacts and local gates complete before push.
