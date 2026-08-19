# Verification — issue #4273 connector surface sweep batch 1

## Verdict

**Passed for the bounded declarative-surface batch and parity-accounting
revision.** Eight evidence-qualified connectors were materialized and gated;
twelve candidates were skipped with an exact provider-artifact reason. This
verifies declaration, preflight, and exhaustive non-delivery accounting only.
It does not certify live provider behavior, claim ETL/reverse-ETL transport, or
claim that any connector has met the >90% documented-operation parity bar.

## Batch evidence

- `batch plan --size 20` selected: `eventzilla`, `paypal-transaction`,
  `persistiq`, `docuseal`, `watchmode`, `avni`, `defillama`, `blogger`,
  `judge-me-reviews`, `oncehub`, `dockerhub`, `oura`, `flexmail`, `printify`,
  `finnworlds`, `perigon`, `nebius-ai`, `coin-api`, `pingdom`, and
  `alpaca-broker-api`.
- `batch materialize` included `avni`, `defillama`, `dockerhub`, `oura`,
  `flexmail`, `perigon`, `pingdom`, and `alpaca-broker-api` (381 declared
  provider operations). Its nonzero result is expected: it records the twelve
  named artifact/coverage drops in
  [`traces/batch-001-materialize.json`](traces/batch-001-materialize.json).
- The pipeline-generated survivor manifest then passed `batch gate`: 99
  commands reached the production `commandrunner.Preflight`, with no drop or
  provenance refusal. See
  [`traces/batch-001-gate.json`](traces/batch-001-gate.json).
- The durable progress ledger has 552 unique rows. Batch 1 is 8 `gated`, 12
  `skipped`, and 532 `pending`; its explicit resume pointer starts batch 2.
- The captain parity-bar revision measures 99 delivered operations of 780
  documented batch-1 operations (12.69% across measured candidates). All 20
  candidates have a coverage percentage; all 552 connectors have an explicit
  coverage value or `null`-while-unmeasured. No candidate is over 90%.
- [`docs/migration/batches/cli-surface-sweep-batches-r1-batch-001-rejections.json`](../../../docs/migration/batches/cli-surface-sweep-batches-r1-batch-001-rejections.json)
  records all 681 undelivered operations: 282 exact provider method/path rows
  and 12 exact source-ledger rows where materialization could not enumerate the
  operation inventory. Each record has a fixed-vocabulary reason, evidence,
  recoverability, and recovery path. G17--G19 are added to the foundation log;
  G13 is resolved by `31bfe62eb`.

## Checks

| Command | Result |
| --- | --- |
| `go run ./cmd/agentcontractgen check` | pass |
| `go run ./cmd/connectorgen batch plan ... --size 20 --min-operations 20 --max-operations 250` | pass; 20 candidates |
| `go run ./cmd/connectorgen batch materialize ...` | expected nonzero; 8 included and 12 named skips, preserved in trace and ledger |
| `go run ./cmd/connectorgen batch gate ...` | pass; 8 included, 381 operations, 99 preflight commands |
| `go run ./cmd/connectorgen validate internal/connectors/defs` | pass; 552 connectors, 0 findings |
| `go run ./cmd/connectorgen surface-sync --check` | pass; 552 scanned, 0 changes needed |
| `go run ./cmd/connectorgen surface-reconcile --check` | expected nonzero; 3,584 fleet-wide reclassifications; recorded as G16 and not applied outside this bounded batch |
| `go run ./cmd/connectorgen certification-candidates --connector <each survivor>` | expected nonzero; each survivor has no `certification.json`, so no certification claim |
| `go test -timeout 20m ./cmd/connectorgen -count=1` | pass; 104.433s |
| `go test -timeout 20m ./internal/connectors/commandrunner -run '^TestEveryImplementedCommandPassesRuntimePreflight$' -count=1` | pass |
| `make connector-runtime-preflight` | pass |
| `go test -timeout 20m ./internal/cli -count=1` | pass; 534.660s after the scoped root-manual snapshot refresh |
| `go vet ./...` and `go build ./cmd/pm` | pass |
| `./pm connectors inspect <each survivor> --json` | pass 8/8 without credentials |
| Built `pm` invoked for all 99 survivor `implemented`/`partial` command paths in an initialized no-credential project | pass; each stopped at the expected credential boundary or declared block |
| `make tidy-check`, `make lint`, `make docs-check-no-build`, `make smoke-no-build`, `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make connector-boundary`, `make release-workflow-check` | pass |
| `make verify` | pass; full repository test, generated-artifact, snapshot, boundary, canon, and release gates |
| `git diff --check` and 552-row ledger invariant | pass |
| `jq` coverage/rejection invariant (all 552 coverage fields; each measured connector has documented = delivered + rejected; 99 + 681 = 780; all rejection reasons in the fixed vocabulary; every foundation-gap rejection names an open gap) | pass |

## CLI/help/manual/website parity

- `pm help connectors` and `pm help docs` rendered successfully.
- The four newly command-surfaced connectors (`avni`, `oura`, `perigon`, and
  `pingdom`) have generator-produced `cli_surface.json`/`operations.json` and
  regenerated `MANUAL.md`/`SKILL.md`; website connector catalogs were regenerated.
- The changed root command manual has nine affected golden transcripts. The
  first broad CLI run made that RED failure visible; the scoped generator
  refresh and follow-up complete CLI suite are GREEN.
- Re-running `pm docs generate`, `pnpm --dir website run gen:website-data`,
  and `pnpm --dir website run gen:docs` produced no further changes.

## Constraints carried forward

- All eight accepted records still truthfully declare no five-class transport
  coverage: their legacy stream/operation metadata is present, but no valid
  `sync_transport.json` can be invented. G12, G14--G19 in the foundation log
  name the bounded fixes required before promotion; G13 is resolved on main.
- No credentials, provider API calls, reverse execution, or live certification
  were used. `gated` is not a certification or live-provider-success claim.
