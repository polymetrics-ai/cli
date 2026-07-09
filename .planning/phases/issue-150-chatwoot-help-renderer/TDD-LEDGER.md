# TDD Ledger: Chatwoot Help Renderer And Docs Parity

## Setup

- Issue: #150
- Parent: #148
- Branch: `feat/150-chatwoot-help-renderer`
- GSD: `scripts/gsd prompt quick --validate ...` succeeded; programming-loop prompt is unavailable and recorded as manual fallback.

## Red / green ledger

### Red 1 — checked-in connector manual lacks Chatwoot command surface

```bash
grep -q 'COMMAND SURFACE' docs/connectors/chatwoot/MANUAL.md
```

Expected before regeneration: fail.

### Red 2 — runtime renderer coverage for Chatwoot command surface

Add a Go test in `internal/connectors/bundleregistry/registry_test.go` requiring the Chatwoot manual to include:

- `COMMAND SURFACE`
- `Usage: pm chatwoot <command> <subcommand> [flags]`
- `conversation list` implemented stream mapping
- `message create` implemented reverse-ETL mapping and approval text
- planned/sensitive/destructive command rows such as `agent list`, `account update`, and `conversation delete`
- global `--json` flag guidance

Command:

```bash
go test ./internal/connectors/bundleregistry -run ChatwootGuide -count=1
```

Expected after test edit and before generated docs work: runtime renderer likely passes because #149 added `cli_surface.json`; if it passes, the red signal remains the checked-in docs/website generated artifacts.

### Red 3 — website generated data lacks Chatwoot cliSurface

Add a Vitest assertion in `website/tests/api/connector-data.test.ts` expecting Chatwoot `cliSurface` metadata in `/docs/connectors/chatwoot/data.json`.

Command:

```bash
( cd website && pnpm test:unit -- tests/api/connector-data.test.ts )
```

Expected before website data regeneration: fail because generated website connector data currently has `cliSurface: null` for Chatwoot.

## Green implementation notes

Pending.

## Refactor notes

Pending.
