# Verification — connector batch pipeline r1

Status: final scoped verification passed on `origin/main`
`de5ebb55b1a1b4de44e3d920b7104670e2d63b02`.

## Pipeline and runtime evidence

- `go test ./cmd/connectorgen -count=1`: passed.
- `go test ./internal/connectors/engine -count=1`: passed.
- `go test ./internal/connectors/commandrunner -run
  '^TestEveryImplementedCommandPassesRuntimePreflight$' -count=1`: passed.
  This is the real runner preflight sweep, not a copied generator rule.
- `go test ./internal/cli -count=1`: passed in 529.383s after the rebase. The
  root-help golden transcript includes the newly reachable batch namespaces.
- `gofmt -d` over the changed `cmd/connectorgen` Go files: no output.
- `go vet ./...`, `go build ./cmd/connectorgen`, and `go build ./cmd/pm`:
  passed.

## Batch 001 re-gate

On the same base, `connectorgen batch gate --connector` passed separately for:

- `docuseal`: 23 declared operations;
- `defillama`: 31;
- `dockerhub`: 54;
- `flexmail`: 43;
- `alpaca-broker-api`: 52.

The final five-candidate gate passed with zero drops and records 203 declared
operations: **39 executable / 27 provider-blocked / 137 excluded**. The
gate reports are retained under `docs/migration/batches/`.

The help traversal invoked every implemented path as
`pm <connector> <command> --help` and reached all **38** commands. The batch
gate separately ran real runtime preflight for every one of those executable
claims. Per-connector ownership checks over each changed `metadata.json`
reported `outcome: clean`.

## Provider evidence and Top-50 planning evidence

- Re-read the live external provider-artifact ledger at
  `/Users/karthiksivadas/karthik-agent-workspace/data/cli-provider-artifact-sweep-r1/ledger.json`.
  Its schema is `cli-provider-artifact-sweep-r1`; all required record fields
  are present. At final read it contained 338 records and 50,334 summed
  documented operations. Batch 001 retains its immutable 235-record source
  snapshot rather than silently changing surveyed evidence under the branch.
- Re-read and JSON-validated the external Top-50 audit. Its fixed cohort is 50
  providers, 46 defensibly measured providers, 13,761 operations (6,106 reads
  / 7,655 writes), and four unknown/not-applicable providers. The checked-in
  size-tier map's 46 named measured providers were compared as an exact set to
  the audit JSON and preserve those totals; it holds invalid or unknown
  inventories out of authoring manifests.

## Repository gates

Each was run individually, avoiding the timeout-prone monolithic suite:

- `make tidy-check`
- `make lint`
- `make docs-check-no-build`
- `make smoke-no-build`
- `make agent-contract-check`
- `make connectorgen-validate` — 550 connectors checked, 0 findings
- `make connectorgen-surface-sync` — 550 scanned, 0 changes
- `make connector-boundary` — clean
- `make release-workflow-check`
- `git diff --check`

No shared schema, engine, or command-runner implementation was changed. No
live provider API call or credential was used; materialization fetched only the
five cited public documentation artifacts and recorded their provenance.

## Next delivery gate

The implementation and Top-50 plan are ready to commit on the branch. The
next Top-50 authoring batch must first obtain a current provider-artifact
manifest through `connectorgen batch plan`; this planning map intentionally
does not make an executable connector claim by itself.
