# SUMMARY — issue 3595 icon registry single-source foundation

## Status

Implementation and local verification are green in the isolated worker worktree.

## Completed slice

- Migrated `internal/connectors/icon_data.json` to canonical bare connector identifiers only.
- Moved curated Simple Icons mappings/fetch metadata into the canonical registry and deleted `website/data/icon_overrides.json`.
- Added canonical Simple Icons SVG assets under `docs/connectors/icons/simple-icons/**`; website copies are generated from that docs tree.
- Updated Go runtime lookup to exact bare-key lookup only.
- Added icon ownership helpers that map canonical docs assets and generated website copies to bare connectors while rejecting ambiguous, orphaned, duplicate, and undeclared icon paths.
- Updated iconregistrygen to emit bare keys, scope to implemented definitions, preserve curated canonical metadata, add reviewed fallbacks, and reject ambiguous source/destination collapses.
- Updated website generation/fetch scripts to read only the canonical registry.
- Updated docs generation to copy nested icon assets recursively for Simple Icons.
- Added migration documentation with the source/destination collapse audit and #3590 catalog allowance handoff.
- F15 fix: `loadCuratedIconEntries` now rejects empty or `source-`/`destination-`-prefixed curated connector keys with an error naming the key and file, instead of silently dropping and backfilling them from upstream/fallback data; raw upstream prefix collapse is unaffected.
- Diagnosed and confirmed resolved the PR #3596 `Website checks` CI failure: the failing run targeted the stale scaffold-only commit, predating the six pipeline review-fix commits already reconciling generated website icon data.
- Review round 8: no-mistakes run `01KZ2661QR8MTV33B5WDB7S1HV` fixed all 8 review-round findings (curated-builtin merge-layer silent drop, Go/JS icon-ownership divergence, unwired Node test suite, panic-on-icon-coverage-drift, shared-path empty-source-URL false conflict, plus 3 mechanical cleanups) and 1 follow-up CI path-filter finding; document gate approved as-is with doc-migration-dir-consolidation deferred to `cli-docs-migration-dir-consolidation-r1`.
- CodeQL fix: added `website/scripts/lib/simple-icons.mjs` (`validSimpleIconSlug`, `resolveSimpleIconRequest`) so `fetch-simple-icons.mjs` validates the registry-authored slug and output path immediately before the CDN `fetch` and the `docs/connectors` filesystem write, closing both CodeQL alerts without weakening the existing SVG-payload/duplicate-path guards.
- Review round 9: `resolveSimpleIconRequest` also rejects a non-string path and any in-tree destination outside `icons/simple-icons/<name>.svg`, so containment alone can no longer authorize overwriting a non-icon file; `validSimpleIconPath` became the single shape guard shared with `fetch-simple-icons.mjs`; `assertInside` is now imported by `gen-connector-bundles.mjs` instead of duplicated; and `Registry.MustValidateIconCoverage` is memoized with invalidation on `Register` so layered constructors stop re-scanning while post-registration drift still aborts.
- CodeQL alert #93 fix: traced the alert to a distinct dataflow (fetched SVG response body flowing to `writeFileSync`, not the destination path) that no prior round validated. Added `verifyFetchedIconDigest`/`sha256Hex`/`readSimpleIconsLockfile`/`writeSimpleIconsLockfile` to `website/scripts/lib/simple-icons.mjs` and a new `website/data/simple-icons.lock.json`, keyed by connector so connectors sharing one icon (confirmed for real in the registry: `ebay-finance`/`ebay-fulfillment`, and 8 `zoho-*` connectors) each verify independently with expected duplicate digests. `fetch-simple-icons.mjs` gained a documented `--update-lockfile` regeneration mode; default mode verifies before every write and fails loudly naming the connector plus expected/received digests. Verified end-to-end against the live Simple Icons CDN: real regeneration, real re-verification, and a real tamper test proving a corrupted digest blocks the write with the on-disk file left unchanged.

## GSD / TDD

- `scripts/gsd doctor`: pass.
- `scripts/gsd prompt programming-loop init --phase connector-guardrail-remediation-r1/workers/issue-3595 --dry-run`: unavailable (`unknown GSD command: programming-loop`), so manual GSD fallback used `.pi/prompts/pm-gsd-loop.md`.
- Execution decision: `local_critical_path` because the user prohibited subagents and this isolated worktree owns issue #3595.
- Red tests captured before implementation; green focused and broad gates are recorded in `TDD-LEDGER.md` and `VERIFICATION.md`.

## Review / integration handoff

- PR #3596 remains draft and targets `fix/3579-connector-path-ownership-guardrails`.
- PR #3590 remains parked; R5's second-registry approach is rejected/superseded, and R6 catalog allowances must be reconciled after this foundation lands.
- Parent PR #3580 remains human-gated and must not be merged by this worker.
