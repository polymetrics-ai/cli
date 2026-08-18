# Issue #3993 — GitHub live certification harness and inbound warehouse proof

**Issue:** #3993 (child of #3988)  
**Target connector:** GitHub only  
**GSD path:** `discuss-phase` → `plan-phase --tdd` → `execute-phase` → `verify-work` → `code-review`, executed inline.

## Manual GSD fallback

`scripts/gsd prompt discuss-phase 3993 --auto` was resolved and its workflow was read. Its initialization rejects `3993` because the repository roadmap does not have that numeric phase. The issue is nevertheless the canonical delivery contract, so this directory is the explicit inline/manual-GSD artifact set required by the delivery contract. No role agents are spawned: this dispatch explicitly prohibits delegation.

## Locked decisions

- The only live owner is `Polymetrics-Cert`; `polymetrics-ai`, personal accounts, and the historical pinned repository remain denied.
- The run boundary is supplied as immutable input and must be explicitly run-owned before any provider mutation. Every setup mutation requires independent provider read-back, inverse cleanup, and final residue scan.
- Authentication uses only a GitHub App installation token minted from the existing private key without printing, logging, or committing secret material; temporary run state is destroyed after the evidence check. The revoked fine-grained token is never probed.
- Applicable cases release together from one barrier. Quota exhaustion and admission outcomes are evidence, never retried or hidden; this issue does not implement #3990 shared REST/GraphQL coordination.
- The present warehouse proof ends after GitHub source → flow sync → Parquet warehouse → DuckDB read-back. Warehouse → GitHub is documented as waiting for #3994; no GitHub-specific approval/action implementation belongs here.

## TDD slices

1. **Boundary-derived case classification (Red → Green).** Add a deterministic script test proving an organization/repository input replaces frozen constants in generated arguments and classification. Verify reason-family movement is reported against the frozen baseline.
2. **Whole-surface sweep mode (Red → Green).** Add a deterministic script test proving non-historical boundaries are admitted after boundary validation, launch is barrier-released rather than sequential, and every terminal record carries independent read-back status.
3. **Self-check drift repair (Red → Green).** Make the manifest and bootstrap inventories source-derived for the current ledger, regenerate checked-in artifacts, and prove both `--check` commands pass.
4. **Live evidence (Green).** Build `pm`, mint the App token without disclosure, create only run-owned resources in `Polymetrics-Cert`, run the applicable surface, cleanup, and capture sanitized result numbers. Execute and independently query the inbound warehouse proof. Record the outbound dependency refusal without attempting a write.

## Verification plan

- Deterministic Node tests for parameter parsing, boundary enforcement, barrier behavior, read-back/result shape, and stale-ledger rejection.
- `node scripts/github-live-lab-manifest.mjs --check` and `node scripts/github-live-bootstrap-probes.mjs --check`.
- Targeted Go tests for changed/used connector and flow/warehouse packages, plus `go build ./cmd/pm`.
- `go run ./cmd/connectorgen surface-sync --check`; scoped `make verify` component gates per AGENTS.md.
- Credentialed live proof uses only the captain-authorized organization, with secret-free evidence and an empty residue scan.

## CLI help/docs/website parity

Not applicable: this change alters certification tooling and evidence, not the user-facing `pm` command surface. The built binary is used as the subject under test; no command, flag, help, manual, website page, or completion changes are planned.

## Required skills

- `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`.
- Final-gate verification/review additionally used `golang-lint`,
  `github-issue-first-delivery`, and `no-mistakes`.
- Manual inline GSD lifecycle because role spawning is prohibited by the dispatch.

## Fresh current-SHA delivery lineage (captain-authorized 1/5)

This historical plan and its completed correction ledger are preserved as the
record of the prior measurement. On 2026-08-11, the captain authorized a
separate fresh delivery lineage capped at five loops for current-SHA live
closure. Its context and TDD plan are `3993-CONTEXT.md` and `3993-01-PLAN.md`;
it begins at **1/5**, not correction 6. The new line uses the built in-process
certification path as the only candidate for shared-rate-coordinator evidence,
treats the external per-operation Node runner as non-equivalent, and records
#4059/#3994/#3992 dependencies without modifying their code.
