# Slice 0 verification

## Result

Passed. Slice 0 remains declaration processing only: no credential was read,
no provider request or mutation ran, no fixture/container was created, and no
certification evidence was published.

## TDD evidence

- **RED:** `go test -count=1 -timeout 20m -run TestBuildGeneratedMutationCandidatesDerivesRESTCollectionCycleInURLDepthOrder -v ./cmd/connectorgen` failed after the test-only expectation was changed to child collection depth `4`: `collection depth = 5, want 4`.
- **GREEN:** restoring the correct expectation made the same command pass.
- **Classifier refusal:** `TestBuildGeneratedMutationCandidatesClassifiesEscapeAndFailsClosed` passes only when synthetic paid-seat, outside-invitation, public-publication, and third-party declarations receive their escape codes, while the unmatched declaration becomes `unassessed`, never `contained`.

## Generated artifacts and accounting

- `go run ./cmd/connectorgen certification-candidates --connector github` succeeded twice; the second run was byte-stable at SHA-256 `7f1608b713b08da357a9e7a5e6a894c93bfdd67442df6aa83d7cc6478d6f23e1`.
- `go run ./cmd/connectorgen certification-candidates --connector github --check` passed.
- `go run ./cmd/connectorgen certification-sweep --connector github` succeeded twice; the second run was byte-stable at SHA-256 `494421bdc4438b23e013fe3c1c7cd9175ad605c652eed02a26e10656c32cda3e`.
- `go run ./cmd/connectorgen certification-sweep --connector github --check` passed and reported 1,571 commands.
- Current measured artifact accounting: 856 candidates, all generated and zero manual; 279 `direct_write` + 577 `reverse_etl`; 793 contained + 15 real-money + 38 real-people + 10 public-visibility + 0 third-party + 0 unassessed = 856. Fixture provenance is 489 derived REST collection cycles + 367 named exceptions = 856.
- Current ledger calculation: 1,355 fixture-required commands have an API surface; 969 derive a same-collection POST/PUT cycle and 386 are named full-surface exceptions.

## Local commands

- `go test -count=1 -timeout 20m ./internal/connectors/engine` — passed (6.751s).
- Focused mutation projection/classifier/CRUD tests in `./cmd/connectorgen` — passed (1.154s).
- `go test -count=1 -timeout 20m -run '^TestCertificationMatrixRejectsDatabaseWriteStubs$' -v ./cmd/connectorgen` — passed (979.333s).
- First `make verify` attempt — failed under concurrent full-suite load: unrelated matrix, app, boundary, commandrunner, and conformance packages reached their 20-minute per-package timeouts. No test was modified or skipped.
- Second `make verify` attempt — passed end-to-end. It includes `go test -timeout 20m ./...`, with `cmd/connectorgen` (173.018s), `internal/app` (317.033s), `internal/cli` (674.978s), `internal/connectors/boundary` (301.218s), `commandrunner`, and `conformance` all passing; it also passed build, docs validation, smoke, lint, agent contract, connector validation, surface sync, GitHub ledger checks, matrix/candidate/sweep freshness, whole-tree boundary scan, connector canon, and release checks.
- `pnpm --dir website run gen:docs`, `pnpm --dir website run gen:website-data`, and `go run ./cmd/pm skills generate --dir docs/skills` — each run twice after rebase; combined diff was byte-stable at SHA-256 `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`.
- `git diff --check` — passed.

`security/snyk` was not run locally. The task brief identifies its base-branch failure as pre-existing; it is unrelated to this declaration-only change.
