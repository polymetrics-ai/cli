# Verification — PostgreSQL logical-replication CDC containment

## Current containment evidence

## Captain execution proof

- [x] The plan explicitly records that a GitHub → PostgreSQL product sync is
  impossible today because PostgreSQL `Write` is unsupported and no database
  write executor exists; rate-limit and CDC evidence are separate halves.
- [x] Real `pm` public `rails/rails` high-volume read: the freshly built
  Darwin/arm64 binary ran `etl read` for **2,000 pull-request records** in
  **28 seconds** with explicit `auth_type=public` and no secret field. The
  stream's declared `per_page=100` means the 2,000 returned records occupied
  **20 full pages**, an implied page pace of **0.71 requests/s** and record
  pace of **71.43 records/s**. GitHub returned neither 403 nor 429 (exit 0;
  no such status in the binary diagnostic channel). The GitHub bundle has
  neither a legacy `rate_limit` nor a declared `rate_limits.json`, so no
  connector limiter throttled this run. The public response payload is not
  included in this artifact or a commit.
- [ ] Docker-via-Colima PostgreSQL 14+ with `wal_level=logical`: integration
  test runs with `POLYMETRICS_INTEGRATION=1` and proves ordered insert/update/
  delete, byte-exact non-ASCII, restart/resume, source-bound durable checkpoint,
  and post-teardown absence of the derived replication slot.
- [ ] PostgreSQL CDC promotion only after the bounded protocol-v2 streamed
  transaction-stage contract, named quota failure, slot-health observability,
  and explicit teardown/rebootstrap procedure pass live conformance.

### Current blocking evidence

- [x] Historical containment evidence: Docker was available through the
  `colima` context, but no PostgreSQL container was started because the then
  merged `dbtest` harness was Podman-only. The #4083 foundation now supplies
  the explicit Docker-or-Podman direct-local-Unix harness path; its
  [maintainer guide](../../../internal/connectors/native/dbtest/README.md)
  owns the current Docker/Colima contract. This removes that foundation gap but
  does not make any PostgreSQL CDC proof current. A hand-authored Docker runner
  would still violate the harness's deterministic ownership and cleanup
  contract.
- [x] `POLYMETRICS_INTEGRATION=1 go test -count=1 -run
  '^TestHistoricalLogicalReplicationResumesAndCleansSlot$' -v
  ./internal/connectors/native/postgres` executed the named test and it
  **SKIPPED** at its unconditional planned-CDC guard before configuration or
  source access. This is not a live conformance pass.
- [x] `go test -count=1 -run
  '^TestCDCIsFailClosedUntilStreamedTransactionStagingExists$' -v
  ./internal/connectors/native/postgres` passes, proving the current reader
  remains non-executable before a source connection, slot creation, or LSN
  acknowledgement.
- [x] The shared `CDCReadRequest` only supplies per-record `emit` and an
  envelope-only checkpoint committer; it cannot establish the captain's
  whole-transaction durable downstream receipt. The required receipt port is
  a shared-runtime foundation change, not a native PostgreSQL-only diff.

- [x] Focused native PostgreSQL tests prove `ReadCDC` fails closed before a connection, non-ASCII values round-trip byte-exact, invalid UTF-8 is rejected, and replica-identity mode guards are named.
- [x] Focused CLI capability-projection tests prove `cdc=false`, PostgreSQL is absent from the CDC catalogue, and inspect reports `planned` with a reason.
- [x] Full `go test -count=1 -timeout 20m ./internal/cli` passes after containment (**662.780s**); scoped PostgreSQL tests, vet/build, connector validation/surface sync, connector docs validation, and `agentcontractgen check` also pass.
- [x] The integration test skips before source access while CDC is planned; it is not current conformance evidence.
- [x] Scoped `go vet` and `go build ./cmd/pm`.
- [x] `make verify` non-suite gates individually: tidy-check, lint, docs-check, smoke-no-build, agent-contract-check, connectorgen-validate, connectorgen-surface-sync, connector-boundary, release-workflow-check.
- [x] `pm help connectors`, `pm connectors`, `pm connectors inspect postgres --json`, and `pm connectors catalog --capability cdc --json` checked; website connector data regenerated.
- [x] Rebased onto `origin/main` `f96a47e80`; inspected the #3964 golden/help change and reran the real exit-code-sensitive CLI suite.
- [x] Final binary-size delta and dependency evidence recorded in `PR-BODY.md`.
- [x] Manual code-review record completed after implementation (`REVIEW.md`).

## Historical pre-containment evidence

The following records predate the captain's streamed-transaction ruling. They
are retained for provenance and dependency review only; they do not claim
current CDC, source progress, checkpoint, or slot lifecycle conformance.

## Recorded evidence

- `scripts/gsd doctor`: pass.
- Required GSD command sources: resolved.
- `go run ./cmd/agentcontractgen check`: pass.
- Rebased cleanly onto `origin/main` at `f96a47e80`; regenerated the website
  connector catalog rather than hand-merging it. Its only working-tree
  propagation is PostgreSQL CDC metadata/docs. The upstream PR #3964 golden
  transcript was inspected in the rebase, and the focused exit-code-sensitive
  CLI package suite passes: `go test -count=1 -timeout 20m ./internal/cli`
  (**599.878s**).
- Baseline `go build -trimpath -o <temporary>/pm ./cmd/pm` at `origin/main`:
  **145,079,490 bytes**. Post-change binary on the same Darwin/arm64 host and
  Go `go1.26.5`: **145,239,730 bytes** (+**160,240 bytes**).
- Historical live conformance: PostgreSQL **12.22**, isolated local data directory,
  non-default port, `wal_level=logical`, `max_replication_slots=8`, and
  `max_wal_senders=8`: pass. It proves selected-table filtering in a
  multi-table publication; insert/update/delete/truncate mapping; durable LSN
  restart without replaying prior records; missing-checkpoint rebootstrap
  refusal; active-slot teardown refusal; and inactive slot cleanup.
