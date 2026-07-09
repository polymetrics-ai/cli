# Research: Features

**Generated via:** official GSD Core Pi adapter command path.

## Product Feature Areas

- Connector inspection and manifest/catalog discovery.
- ETL reads and local warehouse/output loading.
- Reverse ETL writes with plan, preview, approval, execute.
- Local warehouse queries.
- Credential reference management without printing secrets.
- Scheduling and optional runtime-backed execution.
- Connector conformance and certification reporting.

## Connector Surface Feature Areas

- Durable record streams.
- Report/export ingestion.
- Direct-read commands.
- Binary transfer commands.
- Reverse ETL actions.
- Native protocol access for databases, CDC, queues, file/object systems, and product-specific protocols.
- Typed exclusions for unsafe or out-of-scope operations.

## Planning/Agent Feature Areas

- Official GSD command registry under `.gsd/commands.json`.
- Pi interactive command aliases through `.pi/extensions/gsd/index.ts`.
- Default Pi skill behavior through `.pi/skills/gsd-core/SKILL.md`.
- Deterministic shell prompt generation through `scripts/gsd prompt`.

---
*Feature research refreshed: 2026-07-08.*
