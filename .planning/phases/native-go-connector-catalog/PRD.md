# Native Go Connector Catalog PRD

## Goal

Expose the full validated public connector catalog through `pm` as native-Go-only metadata, documentation, and agent-readable JSON.

## Success Criteria

- `pm connectors list --all --json` returns exactly 647 catalog entries.
- `pm connectors catalog --type source|destination --stage <stage> --json` filters catalog entries deterministically.
- `pm connectors inspect <slug>` works for implemented and catalog-only connectors.
- Catalog-only connector manuals include docs URL, configuration schema summary, secret fields, sync support, implementation status, runtime kind, and native Go support path.
- Generated docs include `docs/connectors/catalog/all-connectors.json`, `docs/connectors/catalog/all-connectors.md`, and connector manual/skill files for every catalog entry.
- No connector image execution bridge is introduced.

## Non-Goals

- Do not execute upstream connector containers.
- Do not mark a connector enabled until a native Go implementation exists and passes conformance tests.
- Do not add third-party dependencies for catalog parsing or YAML.
