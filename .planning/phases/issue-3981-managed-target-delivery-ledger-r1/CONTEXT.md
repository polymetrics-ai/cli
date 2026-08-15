# Context — Issue 3981: target delivery ledger

## Task Delivery Header

- Issue: Closes #3981 — Postgres Parity: enforce managed-target ownership and provisioning
- Base branch: integration/4015-mvp-flat-r1
- Merges into: integration/4015-mvp-flat-r1 → main
- Delivery: Pull request open against `integration/4015-mvp-flat-r1` with its checks green; after opening, verify the API-reported base is exactly `integration/4015-mvp-flat-r1`.
- Working branch: fm/cli-3981-delivery-ledger-r1
- Task: Add only the Production-MVP amendment's driver-neutral durable target delivery ledger. Its identity must bind the asserted target owner, destination database identity, and immutable managed-target stream/relation identity; rename/restart and per-relation isolation must be proven.
- Verification: `go test -timeout 20m ./internal/connectors/database/... ./internal/app/...`, scoped race tests, `go vet ./...`, `go build ./cmd/pm`, and the individual `make verify` gates before the no-mistakes pipeline runs.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| The durable ledger key binds asserted owner, target database identity, and immutable managed-target stream/relation identity, never mutable display or table text. | fake | A driver-neutral in-memory durable-store fake is necessary because this issue explicitly excludes a native driver/DDL/SQL. Tests will assert the fake recorded an entry under the exact immutable key and rejects neither a renamed provenance artifact nor a restarted ledger instance. |
| A ledger record resolves after both a rename and a restart. | fake | No database target driver exists on this branch. The test will rebuild the target ref from a renamed artifact with the persisted StreamID, construct a fresh ledger against the same durable-store fake, and assert the previously written record is returned. |
| Two relations under one namespace owner retain independent ledger records and cannot read or overwrite one another. | fake | The existing fake-driver pattern is the only allowed executable target seam in this foundation. The test will write two distinct observable records, read each by its own immutable relation identity, and assert an overwrite of one leaves the other record unchanged. |

## Decision record

- The issue and its verification report identify the ledger as the sole residual. Existing managed-target provisioning types and tests are preserved.
- The ledger is a shared database-package foundation, not a PostgreSQL driver or connector-lane implementation. It must not add DDL, SQL, write sessions, sync modes, CLI surfaces, schema evolution, or reuse `CommittedTransactionStage` / `TransactionReceipt`.
- A target-native durable implementation belongs to a later driver issue. This slice establishes a typed durable-store port and exercises it with the existing fake-based seam, so the key/lookup contract is fixed before native storage exists.
