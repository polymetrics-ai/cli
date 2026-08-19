# Connector surface sweep progress

This is declaration and local runtime-preflight evidence, not provider-live certification.

The captain's parity bar is **more than 90% of documented provider operations per connector**. A `gated` state proves the static generator gate and the no-credential runtime-preflight boundary; it is not a claim that the connector has reached that bar.

| Measure | Total |
| --- | ---: |
| Connector rows | 552 |
| Pending / unmeasured | 532 |
| Batch-1 measured connectors | 20 |
| Connectors meeting the >90% bar | 0 |
| Documented batch-1 operations | 780 |
| Delivered batch-1 operations | 99 |
| Rejected batch-1 operations | 681 |
| Measured batch-1 coverage | 12.69% |
| Gated / skipped batch-1 connectors | 8 / 12 |
| Direct-read declarations | 23 |
| Direct-write declarations | 6 |
| ETL / reverse-ETL transport declarations | 2 / 2 |
| Binary declarations | 14 |
| `sync_transport.json` files | 2 |

## Batch 1 coverage

| Connector | Delivered / documented | Coverage | State |
| --- | ---: | ---: | --- |
| alpaca-broker-api | 10 / 52 | 19.23% | gated |
| avni | 7 / 31 | 22.58% | gated |
| defillama | 10 / 31 | 32.26% | gated |
| dockerhub | 4 / 54 | 7.41% | gated |
| flexmail | 5 / 43 | 11.63% | gated |
| oura | 35 / 75 | 46.67% | gated |
| perigon | 7 / 45 | 15.56% | gated |
| pingdom | 21 / 50 | 42.00% | gated |

The 12 materializer-dropped candidates are explicitly measured at 0% from their source-ledger operation counts; their operation inventory could not be projected without the named recovery. Every undelivered operation is accounted for in [the batch-1 rejection list](../../docs/migration/batches/cli-surface-sweep-batches-r1-batch-001-rejections.json): 282 exact method/path records plus 12 exact source-ledger rows. Reasons use the captain's fixed vocabulary and include evidence, recoverability, and a smallest recovery path.

Foundation-gap rejections update G12/G14/G15 and add G17 (provider callbacks), G18 (SSE sources), and G19 (declarative body/parameter/fan-out schema). G13 is resolved by main commit `31bfe62eb` and the parent stack has been rebased onto it.

## Resume pointer

After Firstmate integrates batch 1 into `chore/4277-connector-surface-sweep`, batch 2 must plan a fresh 20--40 connector candidate group from the committed ledger. It must measure a coverage percentage for every candidate and add a fixed-vocabulary rejection record for every undelivered operation or, when materialization cannot project an operation inventory, the exact source-ledger row.
