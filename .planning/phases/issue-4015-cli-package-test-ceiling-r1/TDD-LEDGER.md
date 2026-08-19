# TDD ledger — CLI package test-ceiling foundation

## Planned red/green evidence

| Slice | Red | Green | Refactor / verification |
| --- | --- | --- | --- |
| Shared executable fixture | A test requesting `buildTransportPM` twice receives different file paths because each request owns a separate `t.TempDir` build destination. | The same test receives one stable executable path built lazily once by the package fixture. | Full package preserves all test names and each existing real-binary proof retains its independent project root. |
| Suite headroom | Baseline evidence shows the base CI package has 19.018s of timeout margin and the current helper rebuilds its image at 18 call sites. | Changed verbose timing demonstrates the one-time binary build reduction; no `-timeout` value changes. | Record exact before/after package and wall timings, then run project gates. |

## Actual evidence

### 2026-08-17 — planning and baseline

- Red prerequisite measured: `/usr/bin/time -p go test -v -timeout 30m ./internal/cli` passed in 627.73s wall / 623.128s package duration. `MEASUREMENTS.md` retains the exact command and result.
- Inventory prerequisite: `go test -list '.*' ./internal/cli` produced 263 runnable test names. Before/after normalized inventories have the same SHA-256, recorded in `MEASUREMENTS.md`.
- Root cause: 18 tests call `buildTransportPM`; each executes the same `go build -o <new-tempdir>/pm ./cmd/pm` before the test's unique process-root assertions. The output hashes in the baseline are identical.
- Red: pending the focused path-identity test.
- Green: pending the lazy `sync.Once` package fixture and `TestMain` cleanup.

### 2026-08-17 — red fixture identity

- Red: `go test -timeout 20m ./internal/cli -run '^TestPMBinaryExecutesIssueLabelWarehouseTransportLifecycle$'` failed as intended. The existing real-binary proof requested `buildTransportPM` twice and received two different paths (`…/001/pm` and `…/002/pm`).
- Observable gap: the first helper invocation built an executable which no later real-binary test could reuse, even though all use the same immutable `./cmd/pm` source tree.
- The failure reported distinct `…/001/pm` and `…/002/pm` paths, recorded above; no test was changed merely to accommodate the failure.
- Green: pending implementation.

### 2026-08-17 — green package fixture

- Green: `go test -timeout 20m ./internal/cli -run '^TestPMBinaryExecutesIssueLabelWarehouseTransportLifecycle$'` passed. The existing proof now requests the package fixture twice, receives one path, and completes its original fresh-process lifecycle assertions.
- Green implementation: `pmTestBinaryFixture` is a lazy `sync.Once` fixture. It builds `./cmd/pm` once into a package temporary directory, returns only its read-only executable path, and `TestMain` removes the directory after `m.Run()` before its explicit exit.
- Green full suite: `/usr/bin/time -p go test -v -timeout 30m ./internal/cli` passed in **537.29s wall** / **532.694s package**, versus **627.73s** / **623.128s** before (90.434s package reduction; 14.5%). `MEASUREMENTS.md` retains the exact commands and results.
- Completeness: normalized before/after `go test -list '.*' ./internal/cli` sets are identical at **263 runnable names**. No test is renamed, removed, skipped, shortened, tagged out, or mode-excluded.
- Final normal-topology confirmation after rebasing: `make verify` passed and its unchanged `go test -timeout 20m ./...` stage reported `internal/cli` in **847.535s**, 352.465s (29.4%) below the ceiling. The final branch therefore meets the headroom requirement without raising the timeout.
