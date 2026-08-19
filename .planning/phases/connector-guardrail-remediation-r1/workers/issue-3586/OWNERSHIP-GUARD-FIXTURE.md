# Ownership Guard Fixture — issue 3586

This proof fixture is an issue-owned input for #3581/#3587. It captures the #3532/#3535 generated/unrelated path classes that must be rejected by the target-aware connector ownership guard.

Machine-readable fixture: `ownership-guard-fixture.json`.

## Expected guard behavior

- A `zendesk-support` connector lane must reject unrelated connector docs/manuals/skills such as:
  - `docs/connectors/bahmni/MANUAL.md`
  - `docs/connectors/bitbucket/SKILL.md`
  - `docs/connectors/gong/MANUAL.md`
  - `docs/connectors/hubspot/SKILL.md`
  - `docs/connectors/xero/MANUAL.md`
- A `google-ads` connector lane must reject unrelated connector definitions such as:
  - `internal/connectors/defs/gong/cli_surface.json`
- A connector lane must reject shared website source/tooling/test foundation paths such as:
  - `website/app/api/raw/[...slug]/route.ts`
  - `website/app/api/search/route.ts`
  - `website/app/docs/connectors/[slug]/page.tsx`
  - `website/scripts/lib/cli-surface.mjs`
  - `website/tests/api/connector-data.test.ts`
- A connector lane may allow target generated docs and narrow shared generated indexes only with explicit target/generated evidence, for example:
  - `docs/connectors/google-ads/MANUAL.md`
  - `docs/connectors/google-ads/SKILL.md`
  - `docs/connectors/catalog/all-connectors.json`
  - `website/data/connectors.generated.json`

## Current branch status

`go run ./cmd/connectorgen ownership --help` currently fails with `connectorgen: unknown subcommand "ownership"` because #3581 has not integrated the executable validator in this worker base. This fixture is the green artifact this sub-issue can own without touching shared guard code.
