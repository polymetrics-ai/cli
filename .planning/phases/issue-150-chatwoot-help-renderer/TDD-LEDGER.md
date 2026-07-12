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

Result before regeneration: failed with exit code 1. This proved checked-in connector docs had not consumed #149 `cli_surface.json`.

### Red 2 — runtime renderer coverage for Chatwoot command surface

Added `TestChatwootGuideIncludesCLISurfaceHelp` in `internal/connectors/bundleregistry/registry_test.go` requiring the Chatwoot runtime manual to include:

- `COMMAND SURFACE`
- `Usage: pm chatwoot <command> <subcommand> [flags]`
- `conversation list` implemented stream mapping
- `message send` implemented reverse-ETL mapping and approval text
- planned/destructive/admin rows such as `agent list`, `account update`, and `platform account create`
- global `--json` flag guidance

Command:

```bash
go test ./internal/connectors/bundleregistry -run ChatwootGuide -count=1
```

Result: passed immediately because #149 added valid `cli_surface.json` and the generic renderer already loaded it. The missing coverage was checked-in docs and website generated data.

### Red 3 — website generated data lacks Chatwoot cliSurface

Added a Vitest assertion in `website/tests/api/connector-data.test.ts` expecting Chatwoot `cliSurface` metadata in `/docs/connectors/chatwoot/data.json`.

Initial command using pnpm could not run because `node_modules` was absent and frozen pnpm install is currently blocked by a lockfile override mismatch:

```bash
( cd website && pnpm test:unit -- tests/api/connector-data.test.ts )
# sh: vitest: command not found

( cd website && pnpm install --frozen-lockfile )
# ERR_PNPM_LOCKFILE_CONFIG_MISMATCH
```

Installed from the checked-in npm lockfile for local test execution:

```bash
( cd website && npm ci )
```

Then ran the red test:

```bash
( cd website && npm run test:unit -- tests/api/connector-data.test.ts )
```

Result before regeneration: failed because `json.cliSurface` was `null` for Chatwoot.

### Red 4 — command-surface flag formatting punctuation

Added a stricter Chatwoot guide assertion for the global connection flag:

```text
--connection (string): Use a saved Chatwoot connector credential and account scope.; maps_to=connection
```

Result before renderer refactor: failed because the renderer emitted `scope.: maps_to=connection`.

## Green implementation notes

- Regenerated Chatwoot connector manual/skill from the current bundle so checked-in docs include `COMMAND SURFACE`.
- Regenerated website connector bundle/catalog data so Chatwoot `/data.json` includes `cliSurface`.
- Updated `website/content/docs/cli-reference.mdx` and `website/lib/docs.generated.ts` to mention Chatwoot command metadata.
- Fixed `renderCommandSurfaceFlag` punctuation so metadata extras render as semicolon-delimited annotations.
- Fixed empty stream/write descriptions in runtime guide sections so generated lines end at `name:` instead of `name: `.

## Refactor notes

- Kept command dispatch unchanged; this remains docs/help metadata only.
- Avoided committing noisy full connector-doc regeneration by regenerating into a temporary directory and copying only `docs/connectors/chatwoot/{MANUAL.md,SKILL.md}`.
