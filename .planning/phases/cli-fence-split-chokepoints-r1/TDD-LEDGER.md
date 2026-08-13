# TDD ledger — Fence split chokepoints r1

| ID | Guarantee | Red evidence | Green proof |
| --- | --- | --- | --- |
| R1 | Connector-local certification truth | A GitHub/PostgreSQL shard-reconstruction test cannot compile because no sharding API exists. | The two generated shards reconstruct exactly their former aggregate payload; any baseline is derived in memory rather than committed. |
| R2 | Stable source anchors | A source-anchor test finds `file:line` anchors in discovered function/workflow inventories. | Every generated Go source anchor is `file:Symbol`; line insertion before an anchor leaves its shard byte-identical. |
| R3 | Honest scoped drift detection | A changed/missing shard could evade the old aggregate-only checker. | `certification-matrix --check` validates and byte-compares allowlisted shards, so changed source constructs and connector rows fail while non-allowlisted connectors stay silent. |
| R4 | App split is mechanical | Existing app ETL and Open-path tests are the baseline behavioural contract. | The same tests pass after moving only composition and dispatch blocks to package-local helpers. |
| R5 | PostgreSQL capability parity | Existing PostgreSQL tests require CDC to remain false. | The table applies the same `CDC=false` override and all native package tests pass. |
| R6 | Incremental lane isolation | A scoped generator run has no API or rewrites every shard. | Generating GitHub leaves PostgreSQL's committed shard byte-identical, and conversely. |

## Red command

```sh
go test -count=1 ./cmd/connectorgen -run 'TestCertification(ShardsRoundTripGeneratedMatrices|ScopedGenerationLeavesOtherShardByteIdentical|SourceAnchorsUseSymbols)'
```

The pre-implementation failure is retained in `traces/red-run.txt`. The identical focused command
must pass after implementation.
