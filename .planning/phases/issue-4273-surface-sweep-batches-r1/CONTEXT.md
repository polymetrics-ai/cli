# Context — issue #4273 surface-sweep batches r1

## Locked decisions

- Deliver one reviewable 20-connector batch on `fm/cli-surface-sweep-batches-r1`, targeting `main` through one direct PR. The PR references #4273; it never merges itself.
- Use only `connectorgen batch plan`, `batch materialize`, `batch gate`, `surface-sync`/`surface-reconcile`, `validate`, and certification-candidate commands. Provider retrieval is through the materializer's cited public artifact or supplied offline corpus; Chrome is a missing/stale-artifact fallback only.
- Read the immutable corpus at `/private/tmp/claude-501/-Users-karthiksivadas-karthik-agent-workspace/3e026f04-5b2c-4ada-bc62-2a28ceeef040/scratchpad/parked-slot41-scratch/.regen-command-contracts.ye3jWI/`; do not modify it.
- Definitions stay under `internal/connectors/defs/<connector>/`. This lane owns no engine, schema, classifier, or command-runner code and does not touch `zoom`.
- A declaration is not live proof. The batch report may record a structurally reachable command after runtime preflight, but no connector is marked certified without accepted live evidence.
- ETL/reverse-ETL transport and binary coverage must be supported by exact runtime executors and conformance. Where the generator or runner cannot establish that contract, retain the truthful absence and record a foundation gap.

## Autonomous discussion result

`discuss-phase 4273 --auto` was resolved inline because the Firstmate brief explicitly authorizes autonomous execution. The canonical contract forbids GSD-role spawning in this environment, so the discussion and planning artifacts are manual inline equivalents.

## Baseline observed on origin/main `454ea182d`

- 552 definition directories; 548 `streams.json`, 240 `writes.json`, 36 `cli_surface.json`, 30 `operations.json`, 12 `certification.json`, and 2 `sync_transport.json`.
- Current direct operation scan: 23 connectors declare a direct-read operation, 6 direct-write, and 14 binary operation; these counts are declarations, not proof.
- The external provider-artifact survey has 526 records and 109 eligible public, versioned OpenAPI/Swagger candidates in the 20–250-operation interval.
