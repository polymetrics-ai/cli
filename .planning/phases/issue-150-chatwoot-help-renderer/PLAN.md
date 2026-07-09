# Plan: Chatwoot Help Renderer And Docs Parity

Issue: #150
Parent issue: #148
Parent PR: https://github.com/polymetrics-ai/cli/pull/223
Branch: `feat/150-chatwoot-help-renderer`
Base: `feat/148-chatwoot-cli-parity`

## Required skills used

- gsd-core
- golang-how-to
- golang-cli
- golang-testing
- golang-documentation
- golang-security
- golang-safety
- frontend-design / website docs parity guidance (website generated data only; no UI redesign)

## Goal

Expose Chatwoot provider-style command/help documentation from `internal/connectors/defs/chatwoot/cli_surface.json` across runtime connector manuals, checked-in connector docs, and website generated connector data without enabling unsafe generic dispatch.

## Scope

- Add Chatwoot-specific renderer/data tests so the generic command-surface path is locked for this connector.
- Regenerate checked-in connector docs so `docs/connectors/chatwoot/{MANUAL.md,SKILL.md}` include the Chatwoot command surface.
- Regenerate website connector data so `/docs/connectors/chatwoot/data.json` includes `cliSurface` metadata and the connector page can render it.
- Update CLI reference website copy to mention Chatwoot alongside GitHub as connector command metadata.
- Preserve safety language: secrets only from env/stdin, reverse ETL plan/preview/approval/run, no generic raw HTTP/write tools.

## Non-goals

- Do not add `pm chatwoot ...` runtime command dispatch in this slice.
- Do not add new direct-read or write implementations; those belong to #151-#155.
- Do not call live Chatwoot APIs or require credentials.
- Do not add dependencies.
- Do not expose raw generic shell, SQL write, HTTP write, or arbitrary binary download tools.

## TDD plan

1. Red: prove checked-in `docs/connectors/chatwoot/MANUAL.md` currently lacks `COMMAND SURFACE` and Chatwoot command metadata.
2. Red: add Go test for `connectors.RenderConnectorManual(chatwoot)` containing usage, groups, implemented ETL/write mappings, planned/sensitive/destructive classifications, and approval text.
3. Red: add website data test expecting Chatwoot `cliSurface` metadata in `data.json`.
4. Green: regenerate connector docs and website generated data from the existing Chatwoot `cli_surface.json`.
5. Refactor: keep tests connector-specific and avoid renderer behavior changes unless tests expose a real bug.

## CLI help/docs/website parity

Applies: yes. This issue changes CLI-visible connector help/docs generated from `cli_surface.json`.

Checklist:

- Runtime help/manual: `pm connectors inspect chatwoot` should include `COMMAND SURFACE`.
- Bare namespace behavior: no namespace behavior changes; existing `pm connectors` behavior remains applicable.
- `docs/connectors/chatwoot/MANUAL.md` and `SKILL.md`: regenerated.
- Website data: `website/lib/connectors.catalog.data.generated.json` includes Chatwoot `cliSurface`.
- Website docs: `website/content/docs/cli-reference.mdx` mentions Chatwoot command metadata.
- Generated website artifacts: regenerate connector catalog/data files.
- Tests: Go renderer test and Vitest connector data test.

## Verification checklist

```bash
./pm help docs
./pm connectors inspect chatwoot | grep -E 'COMMAND SURFACE|Usage: pm chatwoot|conversation list|message create'
./pm docs generate --dir docs/cli --connectors-dir docs/connectors
( cd website && pnpm gen:website-data )
go test ./internal/connectors/bundleregistry -run ChatwootGuide -count=1
( cd website && pnpm test:unit -- tests/api/connector-data.test.ts )
./pm docs validate --connectors-dir docs/connectors
make verify
( cd website && pnpm test:unit -- tests/api/connector-data.test.ts )
git diff --check
```

## Manual GSD fallback

`scripts/gsd prompt quick --validate ...` generated the quick-task prompt. `scripts/gsd prompt programming-loop ...` is not registered in this adapter, so this phase uses the manual programming loop with explicit plan, red tests, green implementation, refactor, and verification artifacts.
