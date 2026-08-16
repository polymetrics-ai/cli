# #4181 GitHub rails/rails 90k commit transport scale proof and two-clock approval repair

## Delivery header

- Issue: Closes #4181 and #4171 — prove the declared `commits` source can complete the shipped GitHub → warehouse → PostgreSQL transport at real scale, then repair the globally scoped approval timeout that prevented the 90k proof.
- Base / head: `integration/4015-mvp-flat-r1` → `fm/cli-github-rails-90k-commits-postgres-scale-r1`.
- Required GSD route: `discuss-phase` → `plan-phase --tdd` → `execute-phase` → `verify-work` → `code-review`.
- Manual fallback: the task phase is not present in the current roadmap (`gsd-sdk query init.phase-op github-rails-90k-commits-postgres-scale-r1` returned `phase_found:false`), the project-local task-delivery header template is absent, and this non-Pi runner cannot provide the compatible isolated roles. This phase record, TDD ledger, run-state, verification, and review record execute the required route inline.

## Skills loaded

`golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-context`, `golang-concurrency`, `golang-database`, `golang-benchmark`, `golang-performance`, and `golang-troubleshooting`.

## Scope and decisions

- Target connector: `github` only. The declared `commits` stream and its transport eligibility must be proven through the fresh `pm` binary, never inferred from bundle metadata.
- Target route: authenticated public `rails/rails` reads with a disposable, harness-owned PostgreSQL target and a fresh project directory. The only credential input is the already-authorized environment reference; no token may enter logs, evidence, commits, or PR text.
- First run is intentionally one declared GitHub page (100 commits) to prove the route before using the 90,000-record budget. The scale run uses 900 pages at the connector's declared `per_page=100` and must independently read PostgreSQL with `SELECT count(*)` after reconnecting.
- Any test-only configurability must preserve the existing unlimited certification default and reject invalid configuration before the PostgreSQL harness or binary starts.
- Disk is a hard safety threshold: stop the scale attempt cleanly if free host space falls below 4 GiB; clean the disposable database/harness after every attempt and never reconfigure the shared runtime.
- Captain decision, 2026-08-16: the run-scoped 15-minute evidence lifetime is removed from this transport path. A one-time preview token instead creates a durable `AuthorizationRecord` with a shape-bound, revocable 24-hour scope; identical no-token runs reuse the standing scope. The project retains the short per-page HTTP deadline and adds a short per-batch apply deadline. A timed-out unit preserves its last acknowledged checkpoint so a later same-scope run resumes under the same standing authorization.
- Durable authorization is not a longer `WriteApprovalEvidence`: the original token remains one-time; the record excludes payload; every PostgreSQL batch rechecks the stored record immediately before mutation; and the source page request itself is bounded independently.
- Measurement is durable application state, not a final successful-test log. Each closed transport run records extracted, Parquet-staged, and PostgreSQL-applied counts plus phase elapsed durations before terminal success or failure is persisted. The binary test reads that state before deferred cleanup and prints a redacted durable measurement on either path.

## TDD slices

1. **Red (complete):** add named configuration tests for the default full certification, invalid/zero page limits rejected before I/O, and the one-page/900-page scale boundaries. **Green (complete):** permit only `unlimited` or a positive page count and derive an exact expected record count from the connector's fixed 100-record page size.
2. **Red (complete):** use the bounded configuration in the existing built-binary commits route and assert the independent Parquet and separate PostgreSQL read-back do not merely echo the pipeline result. **Green (complete):** retain the current unlimited full proof and make the one-page/800-page routes assert their intended count, receipt, and checkpoint.
3. **Red (complete):** named tests failed without authorization-lifetime validation, durable no-token continuation, token-replay refusal, durable-seal separation, and per-unit revocation rejection. **Green (complete):** the one-time token atomically creates an `AuthorizationRecord`, bound to the PostgreSQL transport shape and rechecked just before each staged destination unit. A same-shape fresh binary continuation uses no token; replay is typed and side-effect-free.
4. **Red (complete):** named HTTP-page and destination-batch timeout tests failed without a unit deadline. **Green (complete):** declarative HTTP requests and apply/read-back units have one-minute default deadlines, while the process-wide authority remains day-scale. The second timed-out unit preserves first-unit counts/checkpoint; the unit reauthorization guard refuses a revoked second unit before warehouse/PostgreSQL side effects.
5. **Red (complete):** the failed-transport measurement test initially could not compile because terminal `Run` state had no phase measurement. **Green (complete):** `failRun` and acknowledged failure state persist partial counts/timings atomically; the commits binary harness reopens persisted state and logs the redacted terminal counts before its deferred cleanup.
6. **Measurement (complete):** re-ran one page, then exactly 900 pages. Peak sampled RSS, free disk, staged Parquet bytes, target relation bytes, independently queried PostgreSQL count, receipt/checkpoint, rate-limit headers, and wall time were captured outside the child process. The report distinguishes source-page, stage, and PostgreSQL apply timings from setup/plan time.

## Acceptance evidence

| Criterion | Observable evidence |
| --- | --- |
| Binary route is actually reachable | Fresh `pm` builds, creates the GitHub/PostgreSQL connection, plan/preview/approval, and returns a completed run. |
| Small feasibility | One API page produces exactly 100 commits in connection-owned Parquet and a separately reconnected PostgreSQL `SELECT count(*)`. |
| 90k scale | 900 GitHub pages produce exactly 90,000 records and the same independent read-back count; receipt and checkpoint are present, without a run-scoped evidence expiry. |
| Two-clock safety | One-time token consumption creates one day-scale durable shape authorization; each source/batch unit has a short deadline; revocation is observed before the next PostgreSQL apply; a failed unit resumes from its committed checkpoint. |
| Failure evidence | Terminal failed as well as completed run state contains extracted, staged, applied, and phase timing measurements before caller cleanup. |
| Resource safety | Peak child RSS, disk before/after, Parquet bytes, relation bytes, Docker VM 2 CPU / 2 GiB, and cleanup outcome are recorded without secrets. |
| Bottleneck attribution | Request count/rate-limit headroom plus recorded timing observations identify the limiting component, or explicitly state why the streaming pipeline cannot be separated without new production telemetry. |

## Verification

- Compile and run focused `databaseintegration` tests with the supplied container endpoint and GitHub environment reference.
- Re-run the full existing unlimited commits proof only if the 90k bounded proof passes and disk remains safely above the threshold.
- Run focused unit/configuration tests before any credentialed execution, then `gofmt`, affected-package tests, `go vet`, `go build ./cmd/pm`, `go run ./cmd/agentcontractgen check`, and appropriate `make verify` component gates.
- The production PostgreSQL transport plan now exposes `--authorization-lifetime` (24h default; 24h–48h accepted). Runtime help, `docs/cli/etl.md`, website ETL documentation, generated website docs data, and CLI bad-path/help tests are updated together.
