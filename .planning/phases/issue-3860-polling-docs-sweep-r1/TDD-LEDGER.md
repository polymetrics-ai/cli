# TDD Ledger — #3860 polling-watermark truth surfaces

## Slice 1 — preflight-derived surface eligibility

- Red: `go test -count=1 -timeout 20m ./internal/cli -run '^TestInspectPostgresKeepsPollingWatermarkPlannedUntilPreflightCanBindIt$'` failed because `pm connectors inspect postgres --json` omitted the polling-watermark declaration. The test observed the real CLI output and would also fail for an `implemented` or `change_capture` value.
- Green: after embedding PostgreSQL's planned declaration and projecting it through inspect/catalog, the focused CLI command passes. It observes `status=planned`, a non-empty preflight-block reason, no `implemented`/`change_capture` claim, and the same planned status in CDC-catalog output. Existing real-preflight coverage (`TestPollingModeEligibilityComesFromRealPreflight`) proves a valid registered declaration produces `implemented` rows without source/apply I/O; unsafe cursor validation, declared soft-delete delivery, and source-identity rebootstrap remain covered by the focused engine suites.
- Refactor: `PollingWatermarkDescriptor.MarshalJSON` now serializes a non-implemented declaration as status/reason only. The focused red test for that representation first observed fabricated empty `source`/`target` objects; the green test proves they are absent, so catalog users cannot mistake zero values for a binding.

## Slice 2 — native protocol surface and docs parity

- Red: `TestPostgresNativeAPISurfaceHasNoFabricatedRESTEndpoints` reads the actual PostgreSQL bundle and asserts the `endpoints` array is empty. It fails if any REST-shaped endpoint is introduced.
- Green: connector help, connection/ETL manuals, PostgreSQL generated manual/skill, PostgreSQL bundle docs, and website ETL/docs data now state the keyset ordering, durable-after-ack checkpoint, at-least-once replay, snapshot restriction, hard-delete limitation, cursor-advancing soft-delete condition, and explicit rebootstrap outcome. The help test asserts those exact operator-facing limitations.
- Refactor: built `./bin/pm` fresh before `./bin/pm docs generate --dir docs/cli`, then ran `npm run gen:website-data`. The generated diff was audited; only the planned PostgreSQL polling declaration/manual and its propagated docs changed. Normalized hashes of both website connector-data files match HEAD after blanking PostgreSQL's changed `docs_md`/`docsMd`, proving no other connector entry moved.

## Status

Focused red and green slices are complete. Non-container verification completed with the command results recorded in `VERIFICATION.md`. The live database lane was attempted once, then waived by the supervisor: the shared Docker endpoint stalled in `dbtest` image-store capacity validation and also stalled a read-only `docker info` probe under machine saturation. No retry is authorized for this documentation/surface issue. Review, rebase, and PR delivery remain.
