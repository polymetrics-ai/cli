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
| Review F1 canonical coverage | Missing registry rows can be masked by consumer-side fallback | constructed registries reject missing explicit rows and metadata consumers use canonical rows only |
| Review F2 source URL collision | Same ID/path with different source URLs collapses by iteration order | conflicting source URLs reject while identical source URL fixtures still collapse |
| Review F3 generated icon reconciliation | Deleted or renamed canonical icons leave stale website SVGs | a two-generation path-change fixture proves output-only SVGs are removed inside the bounded generated tree |
| Review F4 slash path portability | host `filepath` semantics can reject portable registry paths | nested forward-slash registry paths validate through slash-oriented `path` semantics |
| Review F5 generated subtree containment | `icons/../outside.svg` passes syntax validation and resolves above the generated icon root | non-clean, dot-segment, absolute, and escaped paths reject before any generated-tree mutation |
| Review F6 generated docs metadata | catalog, MANUAL, and SKILL icon paths can disagree with the canonical registry without failing docs validation | all three generated surfaces validate canonical paths; the exact 61 stale mappings are regenerated |
| Review F7 complete rendered metadata | matching paths mask missing review URLs and obsolete provenance in generated catalog, MANUAL, and SKILL icon blocks | serialized catalog icon objects and exact rendered guide blocks must equal their canonical projections, including optional-field omission |

## Actual evidence log

- 2026-08-02: GSD adapter checked with `scripts/gsd doctor` in isolated worktree.
- 2026-08-02: Required command probe `scripts/gsd prompt programming-loop init --phase connector-guardrail-remediation-r1/workers/issue-3595 --dry-run` failed with `unknown GSD command: programming-loop`; manual GSD fallback uses `.pi/prompts/pm-gsd-loop.md` and records `local_critical_path` because subagents are prohibited for this worker.
- 2026-08-02: GSD trace generated with `scripts/gsd prompt execute-phase connector-guardrail-remediation-r1 --dry-run` in the planning scaffold.
- 2026-08-02: Planning-only scaffold created before any production edits.
- 2026-08-02: Pre-edit audit evidence captured: `internal/connectors/icon_data.json` has 651 entries (590 `source-*`, 56 `destination-*`, 5 bare); 22 base identifiers currently collapse across multiple keys; non-identical collapses include `gcs`, `mssql`, and `file`; `website/data/icon_overrides.json` has 61 curated Simple Icons overrides; `docs/connectors/icons/**` has no Simple Icons while `website/public/connectors/icons/simple-icons/**` has 53 website-only SVG assets.
- 2026-08-02: Red tests added before production code: Go runtime exact bare lookup rejects legacy prefixed fallback; icon registry builder rejects unreviewed source/destination collisions and prefixed emitted keys; website generator/fetcher reject `website/data/icon_overrides.json` authority; ownership/icon path helpers reject ambiguous, orphaned, duplicate, and undeclared canonical/generated icon paths.
- 2026-08-02: RED `go test ./internal/connectors ./cmd/iconregistrygen` failed as expected before implementation: missing `IconSourceSimpleIcons`, `ConnectorIconOwnerForPath`, `ValidateConnectorIconOwnershipPaths`, and `buildOptions`; current generator signature cannot satisfy bare/collision tests.
- 2026-08-02: RED `node --test website/scripts/icon-registry.test.mjs` failed as expected before implementation: `destination-astra` prefixed registry key, `website/data/icon_overrides.json` still exists, and website scripts still reference `icon_overrides`/prefix stripping.
- 2026-08-02: GREEN `go test ./internal/connectors ./cmd/iconregistrygen` passed after implementation: exact bare lookup, Simple Icons mapping, collision rejection, and icon ownership helpers are covered.
- 2026-08-02: GREEN `node --test website/scripts/icon-registry.test.mjs` plus `node --check website/scripts/gen-connector-bundles.mjs` and `node --check website/scripts/fetch-simple-icons.mjs` passed after implementation: website consumers read the canonical registry only and source/generated icon assets exist under docs and website.
- 2026-08-02: Focused CLI docs generation exposed one necessary adjacent docs-generator path: `internal/cli/connector_docs.go` copied only top-level SVGs, so nested canonical `icons/simple-icons/**` assets were missing from temp docs. Implemented recursive SVG copy and verified with `go test ./internal/cli ./cmd/pm` and `make verify`.
- 2026-08-02: Review round 2 rechecked the repo-local GSD adapter; `scripts/gsd doctor` passed and the required `programming-loop` probe remained unavailable, so the recorded manual GSD fallback continues.
- 2026-08-02: Pre-edit F5 proof confirmed `validConnectorIconPath("icons/../outside.svg")` accepts an escaped path and output resolution is anchored at `website/public/connectors`, above the authorized icon subtree.
- 2026-08-02: Pre-edit F6 audit compared catalog, MANUAL, and SKILL path fields with `internal/connectors/icon_data.json` and found the same exact 61 stale connector mappings on every surface.
- 2026-08-02: Review round 3 rechecked the repo-local GSD adapter; `scripts/gsd doctor` passed and the required `programming-loop` probe remained unavailable, so the recorded manual GSD fallback continues.
- 2026-08-02: Pre-edit F7 audit reproduced 220 catalog JSON icon-object mismatches plus 216 MANUAL and 216 SKILL rendered-block mismatches against the canonical registry projection; `100ms` lacks its review URL and `convex` retains obsolete provenance despite matching paths.
