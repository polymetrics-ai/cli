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

## GSD / TDD

- `scripts/gsd doctor`: pass.
- `scripts/gsd prompt programming-loop init --phase connector-guardrail-remediation-r1/workers/issue-3595 --dry-run`: unavailable (`unknown GSD command: programming-loop`), so manual GSD fallback used `.pi/prompts/pm-gsd-loop.md`.
- Execution decision: `local_critical_path` because the user prohibited subagents and this isolated worktree owns issue #3595.
- Red tests captured before implementation; green focused and broad gates are recorded in `TDD-LEDGER.md` and `VERIFICATION.md`.

## Review / integration handoff

- PR #3596 remains draft and targets `fix/3579-connector-path-ownership-guardrails`.
- PR #3590 remains parked; R5's second-registry approach is rejected/superseded, and R6 catalog allowances must be reconciled after this foundation lands.
- Parent PR #3580 remains human-gated and must not be merged by this worker.
