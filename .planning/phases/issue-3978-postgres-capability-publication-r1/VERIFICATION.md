# Verification — Issue #3978: final PostgreSQL certification and publication

## Published shape

PostgreSQL publishes `write=false`, `cdc=true`, and `query=false`.

- `write=false` is deliberate: the public boolean represents a generic writer,
  while PostgreSQL deliberately exposes no direct `Connector.Write` and has
  only a closed managed destination.
- The narrower, live-proven contract stays in
  `sync_transport.destination_transport`: `postgres_managed_target`, six
  named modes, fixed apply strategies, bounded warehouse staging, and
  acknowledgement after receipt.
- `cdc=true` is separately backed by PostgreSQL 14+ pgoutput from a database
  source to the connection-owned warehouse. Neither a CDC-to-API route nor
  destination `change_capture` is claimed.
- `query=false` is unchanged.

The generated PostgreSQL matrix retains each exact `sync_mode` proof on its
own cell. It does not use an exact-mode record to certify a broad capability;
the generic `capability:write` cell is applicable but not declared,
implemented, or live-tested, with the concrete reason recorded by the shard.

## Issue-text reconciliation

1. The request to reject `incremental_dedupe_history` is stale. It is an
   executable declared mode and its PostgreSQL 16 real-binary evidence remains
   in its exact mode cells.
2. The Podman-only wording is stale. The shared harness supports Docker, and
   the historical live run used the explicit Colima Docker socket without
   restarting or reconfiguring Docker or Colima.
3. A prior revision attempted to promote generic write. The generic capability
   scope was not a valid description of the closed destination, so this final
   publication retracts that broad claim rather than weakening a guard.

## Red → green evidence

| Slice | Red observation | Green control |
| --- | --- | --- |
| Evidence scope | `sync_mode` evidence bound to `capability:write` was rejected by the certification gate at every protected transition. | Exact records remain exact; the relevant agent-contract tests pass and a bad GitHub baseline still rejects. |
| Generic write | `TestCertificationMatrixDoesNotTreatPostgresManagedTransportAsGenericWrite` failed with `PostgreSQL managed destination evidence promoted generic write capability`. | Generic `write` again follows the direct writer seam, which remains unsupported, so it is false. |
| Composition | `TestOpenRegistersDefinitionOwnedProductionTransports` rejected `write=true` because the destination is closed and managed. | Metadata, generated matrix, docs, and website data publish `write=false`; the narrow transport remains. |
| Exact destination proof | A broad capability cell would hide the actual contract. | `TestCertificationMatrixRetainsPostgresManagedDestinationEvidenceAtExactModeScope` requires every named `database_write_from_warehouse` mode cell to be declared, implemented, live-tested, and to retain its own proof. |
| CDC receipt | Acknowledging before warehouse receipt is invalid. | The receipt-backed `capability:cdc` and source `change_capture/database_read_into_warehouse` proof remain; API and destination CDC cells are non-pass. |

The earlier aggregate `capability:write` record was produced by a real
PostgreSQL 16 built-binary run after independent read-back of all six closed
managed-target modes. That work is sound evidence for those exact modes, but
it cannot honestly turn the generic public boolean true. The aggregate record
and its writer are therefore removed, not relabeled or re-scoped.

## Retained live PostgreSQL evidence

The recorded fresh profile built `pm`, ran the declared PostgreSQL transport,
and independently read every target back for all six modes. The source CDC
run built a fresh `pm`, started a PostgreSQL 16 pgoutput source, committed ID
901, independently read it from the connection-owned Parquet warehouse,
required a durable receipt with no pending transaction manifest and a complete
checkpoint, then verified `confirmed_flush_lsn` reached the post-insert
transaction LSN. The capability `cdc` record and the exact
`change_capture/database_read_into_warehouse` record remain.

The four program transfer directions are retained from merged #3987/#4195
evidence. This issue does not re-open the excluded GitHub or external-binary
lanes.

The required failure demonstration was retained: after schema compilation, a
scratch `sslmode=bananas` profile made the real PostgreSQL certification binary
exit 2; the profile was restored and the normal certification run passed. The
red trace is `traces/red-runtime-profile.txt`.

## Current local verification

```text
go test -timeout 20m ./cmd/connectorgen -count=1                                      PASS
go test -timeout 20m ./internal/app -count=1                                           PASS
go test -timeout 20m ./internal/connectors/native/postgres -count=1                   PASS
go test -timeout 20m ./internal/connectors/engine -count=1                            PASS
go test -timeout 20m ./internal/agentcontract -count=1                                PASS
go test -timeout 20m ./cmd/connectorgen \
  -run '^(TestCertificationMatrixDoesNotTreatPostgresManagedTransportAsGenericWrite|TestCertificationMatrixRetainsPostgresManagedDestinationEvidenceAtExactModeScope|TestPostgresPublishesOnlyGenericCapabilitiesWithMatchingLiveCertification|TestCertificationScopedSourceResolutionUsesScopedPostgresBundle)$' \
  -count=1 -v                                                                          PASS
go test -timeout 20m ./internal/app \
  -run '^TestOpenRegistersDefinitionOwnedProductionTransports$' -count=1 -v          PASS
go test -timeout 20m ./internal/agentcontract \
  -run '^(TestCertificationGateEnforcesEveryProtectedTransition|TestCertificationGateCurrentBaselineRejectsWithoutBreakingStructuralContractCheck|TestCertificationScopedSourceResolutionUsesScopedPostgresBundle|TestEvaluateCertificationGateGitHubBaselineAndGreenFixture)$' \
  -count=1 -v                                                                          PASS
go run ./cmd/connectorgen certification-matrix --check                                PASS
go vet ./...                                                                            PASS
go build -o pm ./cmd/pm                                                                PASS
make tidy-check; make lint; make docs-check-no-build                                   PASS
make smoke-no-build; make agent-contract-check                                         PASS
make connectorgen-validate; make connectorgen-surface-sync                             PASS
make connectorgen-certification-matrix; make connector-boundary                         PASS
make connector-canon-check; make release-workflow-check                                PASS
pnpm --dir website run lint; pnpm --dir website run typecheck                          PASS
pnpm --dir website run test:unit; pnpm --dir website run test:scripts                  PASS
pnpm --dir website run gen:website-data; pnpm --dir website run gen:docs (twice)      PASS
pnpm --dir website run build                                                           PASS
```

`pnpm --dir website run test:e2e` could not complete locally on its final
retry: Playwright's configured web server failed before tests with
`EADDRINUSE 127.0.0.1:3000`. Two earlier runs also failed inconsistently in
docs-smoke after transient server JSON parsing. This is shared local port
contention, not a source finding; no server or website state was changed. The
clean CI `Website checks` and `Website generated data` gates are authoritative
for this one unexecuted local check.

## Lifecycle and review

The canonical single-worker contract forbids lifecycle-role spawning. The GSD
adapter health/source checks and generated discuss, plan, execute, verify, and
review prompts were run inline/manual. This phase contains the required plan,
TDD ledger, verification report, run state, and red/green evidence. Inline
review found no unresolved Critical or Warning issue after the scope and
composition corrections.
