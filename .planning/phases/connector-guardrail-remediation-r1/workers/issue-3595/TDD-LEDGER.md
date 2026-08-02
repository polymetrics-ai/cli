# TDD LEDGER — issue 3595 icon registry single-source foundation

## Red-first expectations

| Area | Red evidence before production edit | Green evidence |
| --- | --- | --- |
| Canonical key policy | Test/proof fails when registry contains `source-*` / `destination-*` keys or ambiguous collapses | generator/validation rejects prefixed keys and audited collisions; canonical registry uses bare keys |
| Go runtime lookup | Existing exact/source/destination fallback behavior is demonstrated | Go runtime resolves exact bare identifiers only |
| Website mapping authority | Website-only `website/data/icon_overrides.json` masks canonical registry gaps | website scripts read only `internal/connectors/icon_data.json`; override file removed |
| Asset authority | Website-only SVGs can exist without canonical docs source assets | canonical SVGs live under `docs/connectors/icons/**`; website public assets are generated/copied copies |
| Ownership consumer | Ownership cannot resolve examples like `apify-dataset -> icons/apify.svg` and `apple-search-ads -> icons/simple-icons/apple.svg` through one registry | ownership maps source assets and generated website copies to the canonical bare connector; ambiguous/orphan/duplicate paths reject |
| Catalog allowance handoff | PR #3590 still needs exact catalog allowances | handoff records exact `docs/connectors/catalog/all-connectors.json` and `.md` allowance for PR #3590 reconciliation |

## Actual evidence log

- 2026-08-02: GSD adapter checked with `scripts/gsd doctor` in isolated worktree.
- 2026-08-02: GSD trace generated with `scripts/gsd prompt execute-phase connector-guardrail-remediation-r1 --dry-run`.
- 2026-08-02: Planning-only scaffold created before any production edits.
