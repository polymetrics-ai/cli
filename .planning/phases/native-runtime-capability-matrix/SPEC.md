# Native Runtime Capability Matrix Spec

## User Stories

- As an agent, I can run `pm connectors inspect <slug> --json` and determine whether check, catalog, read, write, query, ETL, or reverse ETL is supported.
- As a human, I can read `pm connectors inspect <slug>` and understand whether a connector is enabled or planned.
- As an engineer, I can regenerate the public catalog and preserve the capability contract.

## Acceptance Criteria

- `pm connectors list --all --json` returns 647 connectors and every connector has `runtime_capabilities`.
- `pm connectors inspect source-github --json` reports native GitHub read and write support.
- `pm connectors inspect destination-postgres --json` reports metadata-only support with a non-empty unsupported reason.
- `pm connectors inspect destination-postgres` renders a "RUNTIME CAPABILITIES" manual section.
- `pm docs validate --connectors-dir docs/connectors` validates generated catalog connector docs.
