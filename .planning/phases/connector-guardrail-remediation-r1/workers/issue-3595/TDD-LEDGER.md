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
- 2026-08-02: Required command probe `scripts/gsd prompt programming-loop init --phase connector-guardrail-remediation-r1/workers/issue-3595 --dry-run` failed with `unknown GSD command: programming-loop`; manual GSD fallback uses `.pi/prompts/pm-gsd-loop.md` and records `local_critical_path` because subagents are prohibited for this worker.
- 2026-08-02: GSD trace generated with `scripts/gsd prompt execute-phase connector-guardrail-remediation-r1 --dry-run` in the planning scaffold.
- 2026-08-02: Planning-only scaffold created before any production edits.
- 2026-08-02: Pre-edit audit evidence captured: `internal/connectors/icon_data.json` has 651 entries (590 `source-*`, 56 `destination-*`, 5 bare); 22 base identifiers currently collapse across multiple keys; non-identical collapses include `gcs`, `mssql`, and `file`; `website/data/icon_overrides.json` has 61 curated Simple Icons overrides; `docs/connectors/icons/**` has no Simple Icons while `website/public/connectors/icons/simple-icons/**` has 53 website-only SVG assets.
- 2026-08-02: Red tests to add before production code: Go runtime exact bare lookup rejects legacy prefixed fallback; icon registry builder rejects unreviewed source/destination collisions and prefixed emitted keys; website generator/fetcher reject `website/data/icon_overrides.json` authority; ownership/icon path helpers reject ambiguous, orphaned, duplicate, and undeclared canonical/generated icon paths.
