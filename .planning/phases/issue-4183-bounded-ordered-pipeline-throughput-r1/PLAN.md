# Plan — bounded ordered pipeline throughput R1

## Scope

Target connector: PostgreSQL. Shared production code remains connector-neutral: it obtains ordered-pipeline capability from endpoint declarations and contains no PostgreSQL names, SQL, or pgx types. PostgreSQL alone declares and composes the binary-COPY Arrow endpoints.

## TDD delivery slices

1. **R1 — capability contract and serial preservation.** Red tests define an endpoint declaration that omits ordered-pipeline support and prove a typed pre-I/O refusal; define depth-one sequencing and prove source page two cannot begin before COPY page one ends. Green adds validated declaration capability and routes depth one through the unchanged serial Arrow controller.
2. **R2 — bounded ordered Arrow controller.** Red tests block COPY page one and assert source page two has begun at depth two; prove the admitted count never exceeds requested depth or byte credit; prove output apply order, deterministic first-error cancellation, record release, and goroutine completion. Green implements one producer plus one consumer with context cancellation, explicit ownership transfer, bounded queue, and no shared mutable result races.
3. **R3 — full-overwrite atomicity failures.** Red tests inject stage/COPY failure and cancellation while work is queued. They assert no target publication/read-back/receipt and no checkpoint; retain the existing two-page overwrite and #4184 data-loss regression. Green keeps publication and checkpoint outside the producer/consumer stages and commits only after final read-back.
4. **R4 — CLI/config contract.** Red parser/application tests cover `--max-in-flight-batches` defaults/range, typed unsupported pre-I/O refusal, and connection `--target-copy-workers` default/range/persistence/plan/preview output. Green threads values through the hand parser, app request, connection state, transport request, runtime help, manual docs, website documentation, and generated/golden surfaces.
5. **R5 — proof and review.** Run the pre-change and post-change opt-in 5 GB binary proof on a quiet host, preserving both reports. If 200 MB/s remains missed, record the measured result and stage bottleneck; do not weaken the harness. Run `verify-work`, review the changed cross-package paths inline, and record findings/dispositions.

## Safety and concurrency invariants

- The source owns its Arrow record until the emit callback returns. The queue receives an independently retained record, and exactly the consumer releases it.
- The producer owns source counters; the consumer owns transform/segment/COPY counters; aggregation occurs only after both complete or through synchronization that passes `-race`.
- The producer blocks on both bounded batch depth and byte credit. It selects on cancellation and consumer failure before trying more source I/O.
- A consumer failure cancels the producer and prevents any later apply. Source failure closes the queue; the consumer stops before publication. The abort defer remains armed until publication succeeds.
- Receipts remain appended in consumer order. The final source checkpoint is the candidate of the highest successfully processed contiguous batch and is committed after publish/read-back only.

## CLI parity checklist

- [x] `pm etl`, `pm help etl`, and `pm etl run --help` include `--max-in-flight-batches`.
- [x] `pm connections`, `pm help connections`, and `pm connections create --help` include `--target-copy-workers`.
- [x] `docs/cli/etl.md`, `docs/cli/connections.md`, website ETL/connections/reference pages, generated docs, and golden transcripts match the runtime.
- [x] Bare namespaces still exit zero with contextual help; invalid actions remain usage errors.
- [x] PR body records actual help/docs/generator verification plus named happy, bad, and edge cases and an explicit pending benchmark statement.

## Verification plan

Run affected package tests and `internal/cli` with `-timeout 20m`, `-race` for the pipeline package, `go vet`, `go build ./cmd/pm`, all individual `make verify` gates listed in `VERIFICATION.md`, databaseintegration live proofs under the explicit shared Docker endpoint, and the mandatory GSD inline verification/review fallback.
