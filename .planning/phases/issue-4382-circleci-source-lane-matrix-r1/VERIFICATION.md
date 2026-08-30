# Verification — CircleCI source-lane matrix R1

Status: scoped verification passed. No runtime/executor behavior was changed.

Planned evidence:

- strict JSON parse of the lock and materialized matrix;
- source-ID set equality (111 retained source rows);
- exactly seven cited cells per operation;
- required lane-state/accounting counts;
- deliberate hidden-source-row and invalid-artifact-backlink rejection;
- documented paging and mutation candidate checks;
- source-lock/map-focused Go test, gofmt, diff check, and declaration-admission check.

Focused green result:

```text
go test ./internal/connectors/defs/circleci -run 'TestCircleCISourceLaneMatrix' -count=1
ok   polymetrics.ai/internal/connectors/defs/circleci
```

The validator proves the matrix retains the lock's exact 111 IDs and citations, has
777 cells, accepts no runtime state, identifies nine source-documented cursor candidates,
and independently retains all 50 source mutation candidates. It also validates that the
linked `streams.json` and `writes.json` records exist and cannot point to an unknown
source row or absent lane cell.

Additional scoped checks passed:

```text
go test ./cmd/connectorgen -run 'Test(ParseSourceImportLockRetainsBatchOneLegacyProviderFacts|SourceProjectionMarksDeclaredCircleCIWebhookSecretsEnvOnly)$' -count=1
ok   polymetrics.ai/cmd/connectorgen

go vet ./internal/connectors/defs/circleci
go test ./internal/connectors/defs/circleci -count=1
ok   polymetrics.ai/internal/connectors/defs/circleci

jq empty internal/connectors/defs/circleci/sources/circleci-operation-source-lock.json
jq empty internal/connectors/defs/circleci/sources/circleci-source-lane-matrix.json
jq -e '{rows:(.operations|length), cells:([.operations[].cells[]]|length), paging:([.operations[] | select(.facts.pagination.kind == "cursor")]|length), mutations:([.operations[] | select(.method != "GET")]|length)} == {rows:111,cells:777,paging:9,mutations:50}' internal/connectors/defs/circleci/sources/circleci-source-lane-matrix.json
true

git diff --check
go run ./cmd/agentcontractgen check
agentcontractgen: canonical contract and registered projections are current
```

Current non-scoped baseline checks (not repaired in this source/matrix-only task):

- `connectorgen declaration-admission internal/connectors/defs/circleci --json` exits
  1 because the pre-existing `declaration_admission_inventory.json` is absent.
- `connectorgen validate internal/connectors/defs/circleci --connector circleci --json`
  exits 1 with one pre-existing `source_projection` finding: canonical source descriptor
  `sources/circleci-operation-descriptor.json` is absent.

Neither failure is a reason to create runtime declarations, a source descriptor, or a
Foundation Atlas gap in this delivery.

Out of scope by design: provider calls, credentials, executor behavior, generated CLI
surfaces, runtime ETL/reverse-ETL/sync execution, and modifications to shared
foundation code.
