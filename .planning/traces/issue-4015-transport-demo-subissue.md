# Objective

Deliver the smallest demonstrable production Transport construction slice: one closed GitHub source → connection-owned DuckDB/Parquet warehouse → independent workset reopen → typed GitHub destination path, executed by one freshly built `pm` binary. The path must stage durably, read the destination workset only by immutable handle, independently read back the provider state, advance the source checkpoint only after a durable destination receipt, and complete a typed inverse with zero residue.

# Architectural background

The accepted Transport spine closes descriptor admission, source/destination dispatch, acknowledgement ordering, and checkpoint CAS, but deliberately leaves production construction fail closed: `App.Open` supplies an unavailable verifier, no production registrations, and no `WarehouseStage`. Existing JSONL WAL, DuckDB materialization, one-file Parquet layout, GitHub declarative reads, and typed approval-gated writes are substrate—not a Transport composition. This child adapts that substrate without a new framework, direct connector hop, generic HTTP writer, generic SQL writer, or descriptor self-certification.

# Exact scope

- One explicit production/demo composition root with non-nil connection-owned `WarehouseStage`, read-only accepted-evidence `ConformanceVerifier`, and exact GitHub source/destination registrations.
- A durable stage adapter over the existing owner-scoped JSONL WAL, DuckDB materialization, and Parquet table layout. `Stage` returns only after durability and returns an immutable receipt/handle, never a source-owned record slice.
- `Reopen(handle)` validates owner, generation, manifest, and content identity, then produces bounded records from durable storage after original page/workset references are discarded.
- GitHub source and destination adapters that consume the existing declarative read and typed plan → preview → approval → execute write paths. The destination returns a durable acknowledgement/receipt before checkpoint CAS.
- A bounded manually invocable real-binary demo against a faithful GitHub test server, with sanitized machine-readable evidence and typed inverse cleanup.

# Exclusions

- PostgreSQL source/destination adapters and the remaining three pairings.
- Schedules, flows, query, CDC, cross-process auth/rate coordination, broad seven-mode certification, `full_overwrite` GitHub resource design, generic writers, connector-to-connector calls, and certification-generator expansion.
- Shepherd #3995/#4062 and any merge to `main`.

# Acceptance criteria

1. Production/default construction remains fail closed when verifier, stage, or exact GitHub registrations are absent.
2. The admitted GitHub path runs only as source → durable connection-owned warehouse → `Reopen(handle)` → destination; raw source-page application is impossible.
3. `Stage` fsyncs the WAL and produces an owner-scoped DuckDB-generated Parquet artifact before returning its immutable handle/receipt.
4. `Reopen` rejects owner, generation, manifest, or content-hash tampering and does not expose original source-page aliases.
5. A freshly built `pm` binary reads one bounded run-owned GitHub fixture page, stages/reopens it, performs one typed approval-bound destination mutation on a separate run-owned fixture, independently reads back the expected state, records durable destination receipt before checkpoint CAS, applies a typed inverse, retries cleanup safely, and proves zero residue.
6. Destination failure and checkpoint-store failure preserve replay safety and never falsely advance the checkpoint.
7. Record isolation includes #4077's closed mutable-value correction; no credential value, raw provider payload, or generic write surface is added or emitted.

# RED / GREEN TDD plan

1. **Red:** commit focused regressions showing empty registry/unavailable verifier and nil stage reject before GitHub I/O; raw-page application or in-memory alias is rejected; owner/hash tampering fails reopen; destination failure preserves checkpoint; checkpoint-store failure deterministically replays; and source records remain isolated under the #4077 mutable-value boundary.
2. **Green:** add the minimum composition root, durable stage/reopen adapter, exact GitHub adapters, receipt-before-CAS path, and bounded test-server demo necessary to turn the exact binary run green.
3. **Refactor:** keep construction explicit, ownership checks centralized in existing warehouse primitives, typed errors contextual, evidence sanitized, and all defaults fail closed.

# Verification

- `go test -timeout 20m ./internal/synctransport`
- focused `go test -timeout 20m ./internal/app` and new warehouse/demo packages; targeted `-race` where shared mutable state or durability is exercised
- `go vet ./...`, `go build ./cmd/pm`, and the individual repository verification gates required by `AGENTS.md`
- `scripts/verify-gsd-workflow <combined-base>`
- build a fresh `pm`, record exact commit, SHA-256, and byte size, then run the bounded sanitized demo and assert returned record counts, receipt-before-CAS ordering, Parquet/DuckDB reopen, provider read-back, cleanup retry, and zero residue
- `scripts/gsd` lifecycle: `discuss-phase` → `plan-phase --tdd` → `execute-phase` → `verify-work` → `code-review`, with GSD/TDD evidence committed before production edits
- child no-mistakes pipeline without `--yes`, maximum five correction loops, exact-head CI/review disposition

# Provider and credential safety

Use only the approved GitHub App through PM's normal encrypted credential mechanism and only a disposable `Polymetrics-Cert` resource for any real-provider run. Do not use `polymetrics-ai`, personal repositories, a revoked PAT, direct secret discovery, secret argv, or printed credentials. Begin read-only; mutate only after local fail-closed tests and typed inverse cleanup are green. If the approved boundary does not resolve normally, stop before provider I/O with a fixed blocker code; a faithful local test-server proof remains local proof, not live certification.

# Cleanup

Every mutation has a typed inverse before execution. Independently read back both the mutation and restoration; retry cleanup idempotently; retain only sanitized machine-readable evidence, never provider payloads or credential values. Tampering and destination/checkpoint failures preserve replayable artifacts rather than deleting or advancing state.

# Dependencies and topology

- Parent: #4015; add this as its direct sub-issue and keep the combined parent draft/human-gated.
- Transport substrate: #3862, closed #3864/#4059, and #4077/#4079 record isolation.
- Required implementation base: remote `docs/4015-connector-release-certification` only after it contains corrected Transport parent `feat/3862-any-to-any-transport` at `aaf288d069adc1b67a09500afcca4be4a6d1bab3` (PR #4079 integrated into #4019). Until then this issue permits planning only.
- The child branch is `feat/<issue>-warehouse-mediated-transport-demo`; its draft PR targets `docs/4015-connector-release-certification`, never `main`.

# Source links

- https://github.com/polymetrics-ai/cli/issues/4015
- https://github.com/polymetrics-ai/cli/issues/3862
- https://github.com/polymetrics-ai/cli/issues/3864
- https://github.com/polymetrics-ai/cli/issues/4077
- https://github.com/polymetrics-ai/cli/pull/4019
- https://github.com/polymetrics-ai/cli/pull/4079
- `docs/migration/HANDOFF-CODEX.md`, `docs/migration/conventions.md`, and `docs/architecture/connector-architecture-v2-design.md`
- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-connector-release-certification-r1/FINISH-AND-PARALLELIZATION-PLAN.md`
- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-two-stage-mvp-closure-plan-r1/report.md`
- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-transport-mvp-architecture-audit-r1/report.md`
- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-cert-4015-production-interface-audit-r1/report.md`
