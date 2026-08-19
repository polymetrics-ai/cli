# Discussion log — Issue 3985 connector canon

## GSD discussion path

- Generated with `scripts/gsd prompt discuss-phase cli-canon-cleanup-and-connector-procedure-r1`.
- Executed inline because this task is governed by the canonical single-worker contract and the
  active runtime does not supply compatible isolated GSD roles. No role was spawned.
- The captain's task and accepted reports already resolve the product choices; no human decision
  was reopened.

## Decisions recorded

| Area | Decision | Source |
|---|---|---|
| Delivery shape | Every connector flow passes through the warehouse; no direct source-to-destination route. | Captain ruling; accepted CDC/bidirectional designs |
| Database architecture | Typed framework, PostgreSQL reference driver; no generic SQL executor, extra mode enum, or new repository. | Database framework design |
| CDC | PostgreSQL 14+ `pgoutput` v2; bounded stage; receipt before acknowledgement; quota fails closed; cursor fallback rejected. | Large-transaction strategy |
| Reverse delivery | Inbound transactions and outbound Parquet/DuckDB deltas are different producers for one delivery contract; tombstones explicit. | Bidirectional design |
| Certification | All applicable cells need accepted live evidence; current baseline is zero. | Captain ruling; daily-use report |
| Command claims | `availability: implemented` must pass real runtime preflight; surface metadata is never standalone proof. | Commandrunner test; task |
| Archives | Preserve originals and replace only entry-point files with clear pointers; flag wrong material plainly. | Captain task |
| Remote use | Copy accepted source reports into the repository with hash provenance; list any environment prerequisites honestly. | Task |

## Assumptions checked

- The target reports existed only in the shared captain workspace, not in the isolated Git checkout.
- The existing runtime sweep is present and runs through the actual `commandrunner.Preflight`
  entry point.
- The existing `make verify` suite runs all Go tests but did not expose that sweep as a named
  standalone gate.
- The repository contains a current 15-entry `docs/migration/quarantine.json`; no evidence
  substantiated the repeated 195-provider statement.
