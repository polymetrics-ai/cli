## Intent

Refs #4183. This direct PR targets `integration/4015-mvp-flat-r1` and makes the PostgreSQL transformed `full_overwrite` Arrow path a bounded, ordered producer/consumer pipeline. It overlaps source extraction of batch N+1 with the single ordered binary-COPY/apply of batch N.

## Throughput measurement — explicitly PENDING

The overlap is implemented, but its throughput effect is **not yet measured**. This PR makes no speedup claim.

The last recorded quiet-host result remains unchanged: **5,368,947,776 logical input bytes in 48.000221500 s = 111.85 decimal MB/s (106.67 MiB/s)** on the 8-CPU/16-GiB VM. Its input-byte definition is pre-transform projected source Arrow buffers, excluding Parquet, pgwire, target storage, and checkpoint bytes.

`~172.5 MB/s` is only the audit's optimistic perfect-overlap upper bound from the prior stage timings; it is not a result or forecast. The post-change 5 GB proof is PENDING a quiet host: run the existing explicit database/performance opt-in with its Docker Unix socket endpoint, record load before each baseline/post-change invocation, preserve the reports adjacent to the historical result, and retain the harness's 200 MB/s / 25 s failure gate. No contended measurement, estimate, or extrapolation will be substituted.

## What changed

- Added declaration-gated ordered-pipeline capability to the source and destination transport descriptors. Only the PostgreSQL definition declares it; shared Go remains connector-neutral.
- Added a depth-bounded Arrow producer/consumer controller. It retains records before the synchronous source callback returns, charges byte credit before enqueue, applies receipts serially in source order, drains retained work on cancellation/error, and does not add a second COPY lane.
- Kept full-overwrite's run-scoped lifecycle intact: all immutable segments apply to the shadow run; exactly one publication receipt and read-back occur after the pipeline drains; exactly one final checkpoint is committed after that read-back. There is no per-page checkpoint.
- Added `pm etl run --max-in-flight-batches <1..8>`. Its admitted default is 2 only for the transformed full-overwrite Arrow route when both endpoints declare support. `1` retains the existing serial controller; any explicit value is refused with `OrderedPipelineUnsupportedError` before source or destination I/O unless that exact route can execute the pipeline, so the flag is never silently ignored.
- Added `pm connections create --target-copy-workers <n>` as stored PostgreSQL immutable-COPY connection policy. It defaults to 2 and is bounded by `min(8, target-declared pool maximum)`; plan and preview display it. It is intentionally not a run-level workers knob. This release still has one COPY consumer; a second lane remains a separately measured follow-on.
- Updated executable subcommand help, manuals, generated CLI docs, website docs, golden transcripts, the PostgreSQL certification matrix, and GSD/TDD evidence.

## Safety proof

The following are assertions, not inferred behavior:

- **Happy:** depth 2 deterministically starts page-2 extraction while page-1 COPY is blocked, holds two 32-byte credit leases at peak, refuses to begin page 3 until a slot advances, and COPY-applies `[1, 2, 3]` in source order.
- **Edge:** `--max-in-flight-batches 1` keeps page 2 from extracting until page-1 COPY returns and finishes with one publish, one read-back, and one final checkpoint.
- **Bad:** an endpoint without both declarations produces `OrderedPipelineUnsupportedError` before extractor, destination-plan, or begin calls.
- **Bad:** an injected second queued COPY failure causes one abort and zero publish/read-back/checkpoint calls.
- **Edge:** cancellation after page 2 is admitted cancels/drains the queue, then aborts; it makes no publish/read-back/checkpoint call.
- **Regression:** the existing orchestrator two-page full-overwrite test proves all source pages are applied before one publication/read-back/final checkpoint. App stale-writer, source-failure, cancellation-before-publish, final-checkpoint, and unrelated-state rebase regressions preserve PR #4184's exactly-once publish-then-checkpoint contract.
- **Production binary:** the tagged PostgreSQL two-page overwrite test independently reads `[1,2,3]`, and the transformed Arrow-COPY test independently reads the filtered, typed target with one full-overwrite receipt.

## CLI / docs parity

Verified successful contextual help for `pm help etl`, bare `pm etl`, `pm etl run --help`, `pm help connections`, bare `pm connections`, and `pm connections create --help`; each changed subcommand help exposes its new flag. `docs/cli`, website docs, generated website docs, and golden transcripts were regenerated and validated.

## GSD / skills

Manual inline GSD fallback: the compatible isolated Pi role runtime is unavailable and repository policy forbids role spawning. Resolved `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` prompts; the plan, Red/Green ledger, verification checklist, and review dispositions are recorded in `.planning/phases/issue-4183-bounded-ordered-pipeline-throughput-r1/`.

Skills used: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-context`, `golang-concurrency`, `golang-database`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-documentation`, `golang-performance`, and `golang-benchmark`.

## Verification

See the phase `VERIFICATION.md` for exact commands and results. The 5 GB performance proof is deliberately listed there as PENDING, not green. The remaining repository gate results and final API-reported PR base are appended before opening this PR.
