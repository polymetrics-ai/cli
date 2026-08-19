# Verification — #3977 pgoutput v2 CDC

## Evidence recorded

- [x] Red focused test failed against the planned reader for the intended missing-machine reason: `traces/red-v2-machine.txt`.
- [x] Focused PostgreSQL v2/recovery/version tests pass: `traces/green-v2-machine.txt`.
- [x] The explicit Docker/Colima PostgreSQL `databaseintegration` command passes without a skip: `traces/live-postgres-dbtest.txt`.
- [x] `go vet ./...` and `go build ./cmd/pm` pass.
- [x] `tidy-check`, `lint`, docs validation, smoke, agent-contract, connectorgen validation/surface-sync/certification, connector boundary/canon, GitHub parity artifacts, and release-workflow gates pass individually.
- [x] Local code review findings are recorded and dispositioned: `REVIEW.md`.

## CI follow-up

- [x] The first Verify run exposed stale CLI and engine assertions that expected
  the deliberately planned PostgreSQL descriptor. They were updated to the proven
  pgoutput v2 contract and their focused suites pass; the replacement CI run is
  pending after the review-fix commit.
