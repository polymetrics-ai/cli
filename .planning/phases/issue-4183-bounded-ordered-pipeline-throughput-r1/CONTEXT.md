# Context — bounded ordered pipeline throughput R1

## Task Delivery Header

- Issue: Refs #4183 — feat(postgres): add transformed full-overwrite binary-COPY fast path.
- Base branch: `integration/4015-mvp-flat-r1` at `404536538038e20c3010692ec8fb31e87b11f72f`.
- Merges into: `integration/4015-mvp-flat-r1 → main`.
- Delivery: Direct PR from `fm/cli-bounded-ordered-pipeline-throughput-r1`, API-read-back verified against `integration/4015-mvp-flat-r1` after opening.
- Working branch: `fm/cli-bounded-ordered-pipeline-throughput-r1`.
- Task: Overlap the PostgreSQL transformed full-overwrite Arrow extraction of batch N+1 with one ordered binary-COPY/apply of batch N. Add only the capability-bounded `--max-in-flight-batches` and connection-scoped `--target-copy-workers` surfaces required by the brief.
- Verification: TDD red/green tests, targeted race tests, production-binary PostgreSQL correctness tests, before/after quiet-host 5 GB proof, CLI/help/docs parity, and the repository verification gates recorded in `VERIFICATION.md`.

## Evidence table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Batch N+1 extraction overlaps batch N COPY in execution | fake + live | Deterministic fake blocks COPY and observes the second source extraction before release; the opt-in 5 GB binary proof reports measured stage timings. The fake is required to make scheduling deterministic. |
| Ordered run-scoped publication and one final checkpoint survive | live | Two-page PostgreSQL production-binary test independently reads the published target and asserts both pages, one receipt, and only the final checkpoint. |
| Mid-pipeline failure publishes nothing and advances no checkpoint | fake | A deterministic destination fails the second queued unit; it asserts abort, no publish/read-back, and no checkpoint. A fake is required to inject the exact interleaving before real I/O. |
| Capability refusal happens before endpoint I/O | fake | A source/destination without declared ordered-pipeline support returns the typed error while call counters remain zero. |
| `--max-in-flight-batches 1` is serial-compatible | fake | A controlled source observes no second extraction until the first COPY returns; fixture asserts the same ordered receipt/checkpoint result as the legacy controller. |
| Connection target copy-worker policy is bounded, persisted, and visible | live | App/CLI create-plan-preview path reads the stored value and the effective cap; invalid values fail before persistence or endpoint I/O. |
| Throughput claim is truthful | live | The retained 5,368,947,776-logical-byte baseline stays unchanged; a new adjacent report records host load, elapsed seconds, decimal MB/s, MiB/s, gate status, and stage timings. |

## Locked decisions

- The 5 GB baseline remains exactly `5,368,947,776` logical input bytes in `48.000221500 s` = `111.85` decimal MB/s (`106.67` MiB/s) on the quiet 8-CPU/16-GiB VM. It is never edited.
- The optimistic perfect-overlap ceiling (`~172.5 MB/s`) is explanatory only; it is not a result or a promise. The 200 MB/s / 25 s gate remains intact.
- `full_overwrite` keeps #4184's one shadow/publish receipt, independent read-back, and one final checkpoint. No per-page checkpoint may return.
- The controller is an ordered, bounded producer/consumer pipeline: one source extractor and one destination consumer. It is not source partitioning or concurrent COPY lanes.
- `--max-in-flight-batches` defaults to `2`, accepts `1..8`, and is refused before I/O unless both endpoint declarations support ordered pipelines. `1` selects the preserved serial controller.
- `--target-copy-workers` is connection-scoped, defaults to `2` only for a PostgreSQL full-overwrite binary-COPY target, is `1..min(8, declared target pool maximum)`, persists on the connection, and is reported by plan/preview. This slice does not add a second COPY lane.
- No `--workers`, generic `--parallel`, source-partition, target-connection, provider-rate, or flow-parallelism control is added.
- Required skills: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-context`, `golang-concurrency`, `golang-database`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-documentation`, `golang-performance`, and `golang-benchmark`.
- GSD is executed inline because this agent has no compatible isolated Pi role runtime and repository policy forbids role spawning. Generated prompts resolved: `discuss-phase 4183 --auto`, `plan-phase 4183 --tdd`, `execute-phase 4183`, `verify-work 4183`, and `code-review 4183`.
