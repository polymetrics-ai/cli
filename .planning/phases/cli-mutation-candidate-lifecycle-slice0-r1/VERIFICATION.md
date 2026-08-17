# Slice 0 verification

## Result

Passed. Slice 0 remains declaration processing only: no credential was read,
no provider request or mutation ran, no fixture/container was created, and no
certification evidence was published.

## TDD evidence

- **RED:** `go test -count=1 -timeout 20m -run TestBuildGeneratedMutationCandidatesDerivesRESTCollectionCycleInURLDepthOrder -v ./cmd/connectorgen` failed after the test-only expectation was changed to child collection depth `4`: `collection depth = 5, want 4`.
- **GREEN:** restoring the correct expectation made the same command pass.
- **Classifier refusal:** `TestBuildGeneratedMutationCandidatesClassifiesEscapeAndFailsClosed` passes only when synthetic paid-seat, outside-invitation, public-publication, and third-party declarations receive their escape codes, while the unmatched declaration becomes `unassessed`, never `contained`.
- **Runtime-cost RED → GREEN:** `TestGitHubMutationInventoryIsNotEmbeddedInTheRuntimeBundle` first failed with `runtime bundle mutation candidates = 856, want 0`. The 856 generated rows now live in the generator-only `certification-mutation-candidates.json`; the green guard proves the runtime bundle has zero rows and `defs.FS` does not embed the sidecar. The exhaustive equality assertion remains in `./cmd/connectorgen`.

## Generated artifacts and accounting

- `go run ./cmd/connectorgen certification-candidates --connector github` succeeded twice; the runtime declaration SHA-256 was `28f0db953b3c4379b37eacd8bb467e4a223b44397eae148db20212bb18aeacd6` and the generated mutation inventory SHA-256 was `40abd6620f84533181b11662cdd9a7249a1eb25fe32fd999f92b8de6db900fdd` on both runs.
- `go run ./cmd/connectorgen certification-candidates --connector github --check` passed.
- `go run ./cmd/connectorgen certification-sweep --connector github` succeeded twice; the second run was byte-stable at SHA-256 `494421bdc4438b23e013fe3c1c7cd9175ad605c652eed02a26e10656c32cda3e`.
- `go run ./cmd/connectorgen certification-sweep --connector github --check` passed and reported 1,571 commands.
- Current measured artifact accounting: 856 candidates, all generated and zero manual; 279 `direct_write` + 577 `reverse_etl`; 793 contained + 15 real-money + 38 real-people + 10 public-visibility + 0 third-party + 0 unassessed = 856. Fixture provenance is 489 derived REST collection cycles + 367 named exceptions = 856.
- Current ledger calculation: 1,355 fixture-required commands have an API surface; 969 derive a same-collection POST/PUT cycle and 386 are named full-surface exceptions.

## Local commands

- `go test -count=1 -timeout 20m ./internal/connectors/engine` — passed (6.751s).
- `go test -count=1 -timeout 20m ./internal/connectors/engine` after the runtime-cost repair — passed (5.551s).
- `go test -count=1 -timeout 20m ./cmd/connectorgen` after the runtime-cost repair — passed (83.320s).
- `go test -timeout 20m ./internal/cli` under the unchanged 1200s ceiling — passed before the repair (510.076s) and after it (503.941s).
- `make verify` after the runtime-cost repair — passed end-to-end, including the full Go suite, build, docs, smoke, lint, agent contract, connector validation, candidate/sweep freshness, whole-tree boundary scan, connector canon, and release checks.
- Focused mutation projection/classifier/CRUD tests in `./cmd/connectorgen` — passed (1.154s).
- `go test -count=1 -timeout 20m -run '^TestCertificationMatrixRejectsDatabaseWriteStubs$' -v ./cmd/connectorgen` — passed (979.333s).
- First `make verify` attempt — failed under concurrent full-suite load: unrelated matrix, app, boundary, commandrunner, and conformance packages reached their 20-minute per-package timeouts. No test was modified or skipped.
- Second `make verify` attempt — passed end-to-end. It includes `go test -timeout 20m ./...`, with `cmd/connectorgen` (173.018s), `internal/app` (317.033s), `internal/cli` (674.978s), `internal/connectors/boundary` (301.218s), `commandrunner`, and `conformance` all passing; it also passed build, docs validation, smoke, lint, agent contract, connector validation, surface sync, GitHub ledger checks, matrix/candidate/sweep freshness, whole-tree boundary scan, connector canon, and release checks.
- `pnpm --dir website run gen:docs`, `pnpm --dir website run gen:website-data`, and `go run ./cmd/pm skills generate --dir docs/skills` — each run twice after rebase; combined diff was byte-stable at SHA-256 `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`.
- `git diff --check` — passed.

`security/snyk` was not run locally. The task brief identifies its base-branch failure as pre-existing; it is unrelated to this declaration-only change.
