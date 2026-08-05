# PLAN — issue 3595 icon registry single-source foundation

## Scope

- Parent issue: #3579
- Child issue: #3595
- Branch: `fix/3595-icon-registry-single-source`
- Base: `fix/3579-connector-path-ownership-guardrails`
- Stable decision: `icon-registry-single-source-bare-identifiers-20260802`

## GSD mode

- Repo-local GSD Core adapter checked with `scripts/gsd doctor`.
- Required programming-loop command probe: `scripts/gsd prompt programming-loop init --phase connector-guardrail-remediation-r1/workers/issue-3595 --dry-run` returned `unknown GSD command: programming-loop`; use the repo-local Pi programming-loop prompt (`.pi/prompts/pm-gsd-loop.md`) as the manual GSD fallback and record this explicitly in PR evidence.
- Legacy prompt trace remains available from the scaffold: `scripts/gsd prompt execute-phase connector-guardrail-remediation-r1 --dry-run`.
- Execution decision for this worker cycle: `local_critical_path` because the user explicitly prohibited subagents and this isolated worktree owns the issue #3595 write scope.
- Do not weaken TDD, review, or verification while using the manual fallback.

## Required skills / policy

Load and record: `gsd-core`, `no-mistakes`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-security`, `golang-safety`, `golang-error-handling`, `golang-lint`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-documentation`, and `javascript-testing-patterns`.

Safety gates: no secrets, no credentialed connector checks, no new dependencies without approval, no generic raw write tools, no reverse ETL execution, and no parent PR merge to `main`.

## Objective

Migrate connector icon mapping to one canonical bare-identifier registry. `internal/connectors/icon_data.json` becomes the only authored connector-to-icon registry. Canonical SVG assets live under `docs/connectors/icons/**`, including curated Simple Icons. `website/public/connectors/**` icons are generated/copied from that canonical docs asset tree; no mapping or SVG may exist only in the website output tree.

## Allowed write scope

- `internal/connectors/icon_data.json`
- `internal/connectors/icons.go`
- `cmd/iconregistrygen/**`
- `website/scripts/gen-connector-bundles.mjs`
- `website/scripts/fetch-simple-icons.mjs`
- deletion of `website/data/icon_overrides.json`
- canonical source assets under `docs/connectors/icons/**`
- focused tests for Go runtime, registry generation, website generation/fetching, and ownership consumers
- `internal/cli/connector_docs.go` narrow docs-generator copy fix required so nested canonical `icons/simple-icons/**` assets are included in generated connector docs and `pm docs validate` remains meaningful
- authoritative docs under `docs/architecture/**`, `docs/migration/**`, and this worker artifact directory

No direct PR #3590 R5/R6 response is in scope. PR #3590 remains parked until this foundation lands, then it is reconciled and freshly validated under native-Codex `gpt-5.6-sol` at `xhigh`.

## Implementation outline

1. Audit current icon mappings: prefixed keys, bare keys, source/destination collapses, website overrides, canonical docs assets, and website-only assets. Record current counts and all non-identical source/destination collapses.
2. Add red tests/proofs for prefixed-key rejection, ambiguous source/destination collapse rejection, Go exact bare lookup, website generation/fetch consumers using the canonical registry, and canonical source/generated-copy icon ownership helpers.
3. Migrate curated website Simple Icons mappings and fetch/review metadata into the canonical registry with bare keys, including `apple-search-ads -> icons/simple-icons/apple.svg`.
4. Convert legacy upstream `source-*`/`destination-*` registry entries to audited bare identifiers; keep reviewed local fallbacks distinct and ensure duplicate bare keys/collisions fail validation instead of being chosen by ordering.
5. Ensure generator emits bare keys and rejects prefixed keys, duplicate bare keys, ambiguous source/destination collapses, invalid paths, and unreviewed collisions.
6. Update Go runtime and website scripts to read only the canonical registry and exact bare keys. Remove website-only override authority.
7. Put canonical assets under `docs/connectors/icons/**`; generate/copy website public icons from that tree.
8. Add documentation describing canonical registry ownership and the audit/collision policy.
9. Run focused tests, full relevant gates, and comprehensive native-Codex `gpt-5.6-sol` no-mistakes validation at `xhigh` before integration.

## Review repair round

- F1 is an ownership and construction-validation defect: runtime consumers can synthesize an icon outside the canonical registry, and completed registries do not prove coverage. Make canonical lookup authoritative and validate every registered connector after construction.
- F2 is a collapse-invariant defect: equal IDs and paths are insufficient when source URLs differ. Reject the conflict before either candidate can win by iteration order.
- F3 is a generated-output reconciliation defect: copy-only generation cannot remove obsolete output. Rebuild only the bounded `website/public/connectors/icons/**` tree after validating all canonical inputs.
- F4 is a path-domain defect: registry paths are URL-style slash paths, not host filesystem paths. Validate them with `path.Clean` and `path.Ext`.
- Add focused regressions before production edits; run one combined focused Go and Node verification after the complete fix round.

## Review repair round 2

- F5 is a generated-subtree boundary defect: syntax-only validation accepts non-clean dot-segment paths, and output resolution is anchored above the authorized icon directory. Reject non-clean, absolute, and escaped paths before mutation, then resolve and contain generated files against `website/public/connectors/icons/**` itself.
- F6 is a generated-metadata validation defect: checked-in catalog, MANUAL, and SKILL icon paths can drift from the canonical registry while docs validation still passes. Validate every generated icon path against registry-backed definitions and regenerate only the 61 affected connector mappings.
- Preserve all non-icon generated content and keep the generated website reconciliation bounded to its icon subtree.
- Run one combined focused Node/Go/docs-validation command after both fixes and all generated metadata updates are complete.

## Review repair round 3

- F7 is a canonical-projection defect: generated validation compares icon paths while allowing the rest of the rendered canonical icon record to drift. Compare complete serialized catalog icon objects and exact generated MANUAL/SKILL icon blocks, including deterministic optional-field omission.
- Add same-path regressions for a missing canonical review URL (`100ms`) and obsolete provenance (`convex`) on catalog, MANUAL, and SKILL surfaces.
- Regenerate exactly the existing metadata-only drift: 220 catalog JSON icon objects, 216 MANUAL icon blocks, and 216 SKILL icon blocks. Preserve every byte outside those icon blocks.
- Run one focused Go/docs-validation command after implementation, generated updates, and complete diff review.

## Review repair round 4

- F8 is a canonical projection abstraction defect: the registry decoder retains the curated metadata, but `ConnectorIcon` drops title, Simple Icons identity/color, license/attribution, and match metadata before catalog and guide rendering.
- Define one docs projection on `ConnectorIcon` for ID, path, title, Simple Icons slug/hex, source, license, optional attribution, review status/URL, and match/matched-by metadata. Keep connector identity, implementation state, and fallback disposition as registry control fields rather than icon metadata.
- Render every projected field in deterministic order and omit only absent optional fields. Add catalog, MANUAL, and SKILL regressions for ID, license/attribution, additional curated metadata, and omission on non-curated entries.
- Generate into a bounded worktree-local staging directory, transplant only icon objects/blocks while preserving all current non-icon bytes, then run one focused Go/docs-validation command.
- Required skills: `gsd-programming-loop`, `no-mistakes`, `golang-how-to`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-lint`, `golang-cli`, `golang-documentation`, `golang-structs-interfaces`, and `golang-design-patterns`.

## Review repair round 5

- F9 is a generated-output propagation defect: `website/data/connectors.generated.json` has the canonical website icon projection, but 230 downstream catalog icon objects, 14 connector-list icons, and 61 obsolete public SVG copies are stale. Regenerate those three bounded output surfaces, prove every non-icon catalog/list field is unchanged, and require a second generation run to be byte-stable.
- F10 is an icon-section parsing defect: unanchored first-substring matching can select `FAVICON` or an embedded `## Icon` fragment and does not reject duplicate sections. Match headings only at exact line boundaries, require exactly one occurrence, and return deterministic missing/duplicate errors.
- Add F10 regression cases before production edits. The active review-fix contract permits only one focused verification after all source and generated-output fixes, so the RED expectation is recorded without an interim test run.
- Recheck `scripts/gsd doctor` and the required programming-loop probe; continue the recorded manual GSD fallback and `local_critical_path` execution because the adapter still lacks the command and subagents are prohibited in this review phase.
- Required skills: `gsd-programming-loop`, `no-mistakes`, `golang-how-to`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-lint`, `golang-cli`, `golang-documentation`, and `javascript-testing-patterns`.

## Review repair round 6

- F11 is a section-parser abstraction defect: exact heading identity is coupled to the blank-line separator that follows it. Count exact heading lines independently, then enforce the expected section separator only after uniqueness is established.
- F12 is a registry-validation ownership defect: invalid icon paths can disappear through implementation filtering. Validate and collect every canonical registry row before connector mapping.
- F13 is a curated-precedence defect: duplicate curated keys can overwrite each other before one entry overrides upstream metadata. Reject duplicate curated connector keys before inclusion and precedence logic.
- F14 is a generated-asset ownership defect: different connectors can assign different source URLs to the same destination path. Require one identical source URL identity for every shared path before constructing the download list.
- Add focused regressions before production edits, run the website bundle generator twice, and use one combined focused Go and Node verification after every fix is applied.
- Recheck `scripts/gsd doctor` and the required programming-loop probe; continue the recorded manual GSD fallback and `local_critical_path` execution because the adapter still lacks the command and subagents are prohibited in this review phase.
- Required skills: `gsd-programming-loop`, `no-mistakes`, `golang-how-to`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-lint`, `golang-cli`, `golang-documentation`, and `javascript-testing-patterns`.

## Integration ordering

This child PR must land into the parent branch before PR #3590 is reconciled. After integration, rebase/reconcile PR #3590 so ownership reads only the canonical bare registry, permits the two exact docs catalog outputs, and includes positive/negative collision, orphan, generated-copy, and source-asset tests.
