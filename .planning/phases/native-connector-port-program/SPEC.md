# Native Connector Port Program Spec

## CLI

- `pm connectors port-plan --all --json` returns all native port plans and summary counts.
- `pm connectors port-plan <slug> --json` returns one plan.
- `pm connectors port-plan <slug>` renders a manual with family, wave, ETL, reverse ETL, CDC, and conformance sections.

## Catalog Manuals

`pm connectors inspect <slug>` includes a native port plan section for catalog-only connectors.

## Acceptance

- All 647 catalog connectors produce a native port plan.
- `source-postgres` plan includes logical replication CDC setup.
- `source-mysql` plan includes binlog/GTID CDC setup.
- MongoDB source plans include change streams and resume tokens.
- Planned connectors remain runtime-disabled until implementation status changes.
