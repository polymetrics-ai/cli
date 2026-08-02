# PLAN — issue 3595 icon registry single-source foundation

## Scope

- Parent issue: #3579
- Child issue: #3595
- Branch: `fix/3595-icon-registry-single-source`
- Base: `fix/3579-connector-path-ownership-guardrails`
- Stable decision: `icon-registry-single-source-bare-identifiers-20260802`

## GSD mode

- Repo-local GSD Core adapter checked with `scripts/gsd doctor`.
- Prompt trace: `scripts/gsd prompt execute-phase connector-guardrail-remediation-r1 --dry-run` saved under this worker trace directory.
- Use manual GSD universal loop if a specific programming-loop alias is unavailable; do not weaken TDD, review, or verification.

## Required skills / policy

Load and record: `gsd-core`, `no-mistakes`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-security`, `golang-safety`, `golang-error-handling`, `golang-documentation`, and task-specific website/JS test guidance as needed.

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
- authoritative docs under `docs/architecture/**`, `docs/migration/**`, and this worker artifact directory

No direct PR #3590 R5/R6 response is in scope. PR #3590 remains parked until this foundation lands, then it is reconciled and freshly validated under native-Codex `gpt-5.6-sol` at `xhigh`.

## Implementation outline

1. Audit current icon mappings: prefixed keys, bare keys, source/destination collapses, website overrides, canonical docs assets, and website-only assets.
2. Add red tests/proofs for prefixed-key rejection, duplicate/ambiguous collapse rejection, Go exact bare lookup, website generation/fetch consumers using the canonical registry, and ownership mapping of canonical source assets plus generated website copies.
3. Migrate curated website Simple Icons mappings and fetch/review metadata into the canonical registry with bare keys.
4. Ensure generator emits bare keys and rejects prefixed keys, duplicate bare keys, ambiguous source/destination collapses, invalid paths, and unreviewed collisions.
5. Update Go runtime and website scripts to read only the canonical registry and exact bare keys. Remove website-only override authority.
6. Put canonical assets under `docs/connectors/icons/**`; generate/copy website public icons from that tree.
7. Add documentation describing canonical registry ownership and the audit/collision policy.
8. Run focused tests, full relevant gates, and comprehensive native-Codex `gpt-5.6-sol` no-mistakes validation at `xhigh` before integration.

## Integration ordering

This child PR must land into the parent branch before PR #3590 is reconciled. After integration, rebase/reconcile PR #3590 so ownership reads only the canonical bare registry, permits the two exact docs catalog outputs, and includes positive/negative collision, orphan, generated-copy, and source-asset tests.
