# Issue 4015 Sync Pipeline E2E — Code Review

## Scope

- Test-only change: add direct target content assertions and align the second-run expectation with checkpoint-skipping `incremental_upsert` behavior.
- Planning/evidence artifacts: describe the live commands, independent read-back, recovery finding, route verdicts, and cleanup.
- No production code, dependency, CLI surface, help, docs, website, or connector definition change.

## Findings

### Release blocker — PostgreSQL CDC restart is not resumable

The live child process fails after process death with `sync rebootstrap required: invalid_checkpoint: PostgreSQL polling checkpoint mechanism is not resumable`. This is recorded as a product finding and intentionally not fixed in this test-only task.

### Test expectation — resolved

The existing test expected an acknowledged no-change `incremental_upsert` rerun to extract/load all 1,001 rows. Live behavior correctly skips them after the polling watermark advances. The assertion now requires `0/0`, then independently proves the target count and sample are unchanged.

## Security and safety review

- The new SQL selects a fixed allowlisted relation identity discovered by the existing harness and parameterizes the row ID.
- Logs contain only schema/relation names, counts, and seeded non-secret sample values.
- No credential or approval token is logged.
- Cleanup targeted only the two exact run-resource identifiers created by this task.

## Verdict

The changed test and evidence are review-clean. The external release decision remains blocked on disposition of the recorded recovery finding.
