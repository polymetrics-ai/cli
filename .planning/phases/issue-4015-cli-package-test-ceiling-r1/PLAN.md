# Issue #4015: CLI package test-ceiling foundation

## Task Delivery Header

- Issue: Refs #4015 — Production MVP certification foundation
- Base branch: `integration/4015-mvp-flat-r1`
- Merges into: `integration/4015-mvp-flat-r1` → `main`
- Delivery: Pull request open against `integration/4015-mvp-flat-r1`, API-read-back verified, with recorded local verification and review evidence.
- Working branch: `fm/cli-cli-package-test-ceiling-r1`
- Task: Eliminate repeated production-binary linking in `internal/cli` while retaining every real-binary proof and every runnable test name.
- Verification: Compare before/after test inventories; run baseline and changed verbose packages; prove a repeated fixture lookup shares an executable path; run focused, package, consumer, and required verification gates.

## Evidence table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Real-binary tests execute the production entry point | live | The unchanged tests still invoke the package fixture's `pm` executable against their own temp project roots and assert their existing provider/warehouse/approval outcomes. |
| The shared fixture links the executable no more than once per test binary | live | A red/green fixture test asks twice for the binary and requires the same executable path; the pre-change helper created two distinct paths and fails. |
| Tests do not disappear | live | Normalized `go test -list '.*' ./internal/cli` before/after sets are identical. |
| Package timing leaves headroom | live | Changed verbose package timing is measured and compared to the 20-minute ceiling; the documented integration baseline is 1180.982s and the local baseline is retained for like-for-like comparison. |

## Measured decision

### Baseline

`/usr/bin/time -p go test -v -timeout 30m ./internal/cli` completed successfully:

- Go package duration: **623.128s**
- Process wall clock: **627.73s**
- Runnable test names: **263** (`TestMain` is the 264th top-level `Test…` function but is not a runnable test name).
- Integration base evidence supplied by the task: **1180.982s**, only 19.018s below the Makefile's 1200s per-binary timeout.

The compact command, inventory, and timing evidence is in `MEASUREMENTS.md`. The per-test ranking is a long tail — 126.44s, 44.10s, 40.55s, 39.78s, 32.70s, and 30.40s — rather than an individual failing test. `TestBahmniDeclaredCommandMatrixIsRecognizedOrExplicitlyBlocked` completes in 39.78s and is not changed.

### Execution audit

- `TestMain` (`certify_cli_test.go`) installs the in-process certification CLI runner and tracks its process-wide real-invocation budget. It does not use `t.Parallel()`.
- No `internal/cli` top-level test calls `t.Parallel()`.
- Multiple tests use process-global `t.Setenv`, and `config_test.go` plus `cobra_router_test.go` use process-global `t.Chdir`; Go intentionally disallows these mutators in parallel tests. A package-wide parallelization refactor would need a broad isolation rewrite and is rejected as a disproportionate blast radius.
- The real-binary tests use distinct `t.TempDir` project roots and ephemeral HTTP listeners, so their *workspaces* are isolated, but every call currently rebuilds an identical immutable `pm` image.
- `buildTransportPM` has **18 call sites** across the real-binary proofs. It invokes `go build -o <test-tempdir>/pm ./cmd/pm` once per caller; no test depends on changing the executable.
- `.github/workflows/verify.yml` runs `make verify`, whose `test` target is exactly `go test -timeout 20m ./...`; Go can schedule packages but `internal/cli` remains one serial test binary. The hosted runner therefore cannot reclaim the repeated linking inside that binary.

The final rebased aggregate `make verify` command directly proves the normal `-timeout 20m` package result at 847.535s — 29.4% below the ceiling — rather than relying solely on an extrapolation from the diagnostic run.

### Selected mechanism: lazy package-scoped binary fixture

Use a `sync.Once` fixture owned by `TestMain`: the first real-binary test creates a package-owned temporary directory, builds `./cmd/pm` once, and later callers receive that read-only path. `TestMain` removes the directory only after `m.Run()` completes, before its existing explicit `os.Exit`.

This is the smallest generic mechanism that addresses the measured repeated operation and preserves each test's production entry point, args, project root, environment, and assertions. It needs no new dependency, no connector-specific logic, no test exclusion, no timeout increase, and no CI partition that could omit a test.

### Rejected mechanisms

| Direction | Rejection reason |
| --- | --- |
| Package-wide `t.Parallel()` | Unsafe without a broad rewrite because the package uses process-global environment and current-working-directory mutation. |
| CI sharding | It would distribute the symptom but retain eighteen cold links per binary per shard, increase total compilation, and require a test-name partition guard. The measured direct cause has a lower-risk local remedy. |
| Move heavy tests | Would alter package topology and timeout accounting without fixing repeated image construction; the heavy real-binary proofs remain intentionally intact. |
| Raise `-timeout` | Prohibited as a standalone response and unnecessary after removing the directly measured repeated work. |

## TDD slices

1. **Red — shared executable identity.** Add a focused fixture test that requests `buildTransportPM` twice and requires a stable path. On the baseline it fails because each request builds into a distinct `t.TempDir` path.
2. **Green — package-owned one-time fixture.** Make the fixture lazy, synchronized, and cleanup-safe under `TestMain`, then pass the focused test and full CLI suite without changing any existing assertion.
3. **Proof — completeness and timing.** Compare test-name sets, collect the changed verbose trace, establish the measured reduction/headroom, and run all required verification gates including consumer tests and connector-boundary validation.

## Lifecycle and skills

- Generated prompts resolved: `discuss-phase`, `plan-phase --tdd`; `execute-phase`, `verify-work`, and `code-review` remain required after the implementation slice.
- Manual GSD fallback is necessary because this worker is outside Pi and this task prohibits spawning the GSD role set. The lifecycle artifacts are created and executed inline.
- Skills loaded: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-benchmark`, `golang-performance`, and `golang-troubleshooting`.
- CLI help/manual/website parity is not applicable: runtime command behavior, command surface, help text, and documentation do not change; only test fixture process construction changes.

## Commit checkpoints

1. Planning and red-test evidence.
2. Green fixture implementation and targeted package evidence.
3. Full verification, review disposition, push, and PR base read-back.
