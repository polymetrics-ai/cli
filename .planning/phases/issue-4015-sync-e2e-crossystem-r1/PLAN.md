# Issue 4015 Cross-System Pipeline Certification — Plan

## Objective

Add one opt-in, fresh-binary live certification harness and use it to produce bounded, independently read evidence for the three previously unattempted warehouse-mediated routes plus an installed flow/schedule firing. Ship only test and lifecycle evidence; do not change product behavior.

## Lifecycle

- `scripts/gsd prompt discuss-phase issue-4015-sync-e2e-crossystem-r1`
- `scripts/gsd prompt plan-phase issue-4015-sync-e2e-crossystem-r1 --tdd --skip-research --auto`
- `scripts/gsd prompt execute-phase issue-4015-sync-e2e-crossystem-r1`
- `scripts/gsd prompt verify-work issue-4015-sync-e2e-crossystem-r1`
- `scripts/gsd prompt code-review issue-4015-sync-e2e-crossystem-r1 --depth=quick --files=internal/cli/cross_system_live_certification_integration_test.go,.planning/phases/issue-4015-sync-e2e-crossystem-r1`

The generated workflows are executed inline because the canonical single-worker contract forbids delivery-role spawning.

## Slice 1 — environment and Red

1. Inspect the local container/runtime state read-only. Do not start, stop, restart, or update it.
2. Inspect `pm help etl`, `pm help reverse`, `pm help flow`, `pm help schedule`, and both connector manifests before invoking unfamiliar commands.
3. Add an opt-in live test constrained to the named GitHub fixture repository and explicit local container endpoint.
4. Red is the absence of run-specific live destination state and 404 cleanup evidence for these routes. No product code is changed to manufacture a test failure.

## Slice 2 — PostgreSQL → warehouse → GitHub

1. Seed one PostgreSQL row describing a uniquely prefixed label update.
2. Materialize it with `incremental_upsert` into a connection-owned warehouse table.
3. Create, preview, approve through stdin, and run a typed `update_label` reverse-ETL plan.
4. Independently read the GitHub label and assert exact name, color, and description.
5. Re-run the incremental ETL and a newly approved reverse plan; assert the source skips acknowledged rows, the GitHub update remains idempotent, and exactly one label exists.

## Slice 3 — GitHub → warehouse → PostgreSQL

1. Reuse the still-live task label as a named GitHub source record and independently count the repository labels.
2. Run the `labels` stream through a `full_overwrite` PostgreSQL managed-target connection with plan, preview, stdin approval, and execute.
3. Reopen PostgreSQL through pgx, assert the exact independent source count and the named label's content.
4. Re-run under the durable standing authorization; assert a second full overwrite replaces rather than duplicates and preserves exact content.

## Slice 4 — GitHub → warehouse → GitHub

1. Create one uniquely prefixed issue comment after recording an exact `since` boundary; assert the bounded source returns only that comment.
2. Materialize the `issue_comments` stream through `full_refresh_overwrite` into the warehouse.
3. Create, preview, approve through stdin, and run a `--limit 1` typed `update_issue_comment` plan mapping `id → comment_id` and `repository → body`.
4. Independently read the comment and assert its exact changed body and singular ID.
5. Re-run through the scheduled flow in Slice 5 and verify the update is idempotent.

## Slice 5 — flow and installed schedule

1. Persist a flow that references the already approved API → API action job.
2. Run `pm flow plan`, `pm flow preview`, and `pm flow create` to prove the stored definition is honored.
3. Run `pm schedule create` and `pm schedule install --crontab` against a task-local crontab file.
4. Execute the exact installed entry point: `pm --root <root> flow run <flow> --json`.
5. Inspect schedule state for terminal success, prepared execution identity, and receipt; independently read the comment and assert exactly one object with unchanged desired content.

## Slice 6 — cleanup, verification, and delivery

1. Delete every task-created label/comment through an independent bounded GitHub client in unconditional cleanup.
2. Assert each exact resource endpoint returns HTTP 404. Confirm no credential material exists under the project tree.
3. Confirm task-owned database/container resources are absent after harness cleanup.
4. Record exact commands, results, route verdicts, incremental semantics, and optional scale disposition in `VERIFICATION.md` and `SUMMARY.md`.
5. Run targeted tests and the repository's non-full-suite gates separately per `AGENTS.md`, commit/push coherent checkpoints, open the direct PR with `Refs #4015`, and verify its API-reported base.

## TDD boundaries

- Red: no current live test can satisfy the authorized repository boundary and all four acceptance rows.
- Green: the opt-in harness observes every destination independently and always cleans task-owned GitHub fixtures to 404.
- Refactor: consolidate only test-local helpers; no product refactor or defect fix is allowed.

## Checkpoints

- Plan checkpoint: context, plan, TDD ledger, and run state committed before live execution.
- Live checkpoint: all attempted route outcomes and cleanup are committed after the credentialed run.
- Verification checkpoint: review artifacts and local gates committed before final push/PR.

## Skills and parity

Required skills loaded: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-database`, `golang-design-patterns`, and `golang-structs-interfaces`.

CLI help/manual/website parity is not applicable unless the lane unexpectedly changes a command, flag, output contract, connector surface, docs, or website behavior. Such a change is outside this PR and must be split.
