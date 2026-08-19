Refs #3978

## Summary

- Publish PostgreSQL `write=false`, `cdc=true`, and `query=false`.
- Keep the independently certified PostgreSQL managed destination as its existing narrow `sync_transport.destination_transport` declaration: exact `postgres_managed_target` executor, six named modes, fixed apply strategies, warehouse receipt, and acknowledgement boundary.
- Retain the PostgreSQL 14+ pgoutput source-to-connection-owned-warehouse proof. No CDC-to-API or destination `change_capture` route is implied.

## Stale issue-text reconciliation

1. `incremental_dedupe_history` is executable; all six declared managed-target modes remain live-proven.
2. Docker through the explicit Colima socket is valid for `native/dbtest`; Docker and Colima were not restarted or reconfigured.
3. The issue's typed/non-executable CDC wording is stale. Current CDC is executable only as database source -> connection-owned warehouse.

## Certification-gate and composition corrections

The first revision was halted correctly when it bound `sync_mode` proof to `capability:write`: a mode record cannot certify a broad capability. The gate still rejects that mismatch and still rejects the current bad GitHub baseline.

The next revision produced a fresh PostgreSQL 16 aggregate `capability:write` proof after independent read-back of every managed-target mode. That live work is sound, but the composition guard then correctly rejected its publication: `Capabilities.Write=true` means a generic writer, while PostgreSQL has only a closed managed destination and its direct `Connector.Write` deliberately returns unsupported.

The schema has no narrower `Capabilities` variant. The truthful narrower shape already exists in `sync_transport.destination_transport`, so this PR retracts the generic capability record/binding and publishes `write=false`. The twelve exact `sync_mode` records continue to certify the six closed managed-destination modes; they are not relabeled, re-scoped, or used to satisfy a generic capability cell. `cdc=true` remains independently receipt-backed; `query=false` remains explicit.

## Cases exercised

- Happy: every closed managed-target mode independently read back its PostgreSQL target after the fresh built-binary profile.
- Bad: mode evidence cannot satisfy `capability:write`; a declared generic write remains false when direct `Connector.Write` is unsupported; acknowledgement before warehouse receipt is rejected.
- Edge: schema-valid `sslmode=bananas` still makes the real certification binary fail after schema compilation, then passes after restoration; startup replication-slot LSN is not accepted as the transaction acknowledgement.

## Verification

```text
go test -timeout 20m ./cmd/connectorgen -count=1                                      PASS
go test -timeout 20m ./cmd/connectorgen \
  -run '^(TestCertificationMatrixDoesNotTreatPostgresManagedTransportAsGenericWrite|TestCertificationMatrixRetainsPostgresManagedDestinationEvidenceAtExactModeScope|TestPostgresPublishesOnlyGenericCapabilitiesWithMatchingLiveCertification|TestCertificationScopedSourceResolutionUsesScopedPostgresBundle)$' \
  -count=1 -v                                                                          PASS
go test -timeout 20m ./internal/app -run '^TestOpenRegistersDefinitionOwnedProductionTransports$' -count=1 -v PASS
go test -timeout 20m ./internal/connectors/native/postgres -count=1                  PASS
go test -timeout 20m ./internal/connectors/engine -count=1                           PASS
go test -timeout 20m ./internal/agentcontract -run '^(TestCertificationGateEnforcesEveryProtectedTransition|TestCertificationGateCurrentBaselineRejectsWithoutBreakingStructuralContractCheck|TestCertificationScopedSourceResolutionUsesScopedPostgresBundle|TestEvaluateCertificationGateGitHubBaselineAndGreenFixture)$' -count=1 -v PASS
go run ./cmd/connectorgen certification-matrix --check                               PASS
go run ./cmd/pm docs generate --dir docs/cli --connectors-dir docs/connectors         PASS
pnpm --dir website run gen:website-data; pnpm --dir website run gen:docs (twice)      PASS
pnpm --dir website run build                                                          PASS
go vet ./...; make agent-contract-check; make connectorgen-certification-matrix; make docs-check-no-build PASS
```

`pnpm --dir website run test:e2e` was not recorded as passing: its final
local retry failed before tests because Playwright's web server could not bind
the shared `127.0.0.1:3000` (`EADDRINUSE`). Earlier retries also hit different
transient docs-smoke JSON parse timeouts with no source change. No shared
server or website state was touched; clean CI `Website checks` and `Website
generated data` remain authoritative for that environmental-only check.

The mandatory GSD lifecycle was run inline/manual because the repository's single-worker contract forbids role spawning. The plan, TDD ledger, verification report, and red/green traces are included in this PR.
