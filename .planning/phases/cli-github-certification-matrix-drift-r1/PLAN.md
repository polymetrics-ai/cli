---
phase: cli-github-certification-matrix-drift-r1
type: tdd
base_branch: main
files_modified:
  - .planning/phases/cli-github-certification-matrix-drift-r1/
  - internal/connectors/defs/github/certification-matrix.json
autonomous: true
---

# TDD Plan: Refresh the GitHub certification-matrix shard

## Task Delivery Header

- Issue: No dedicated issue — Firstmate requested a shared-base generated-artifact unblocker. This work does not claim or close #4302, whose engine scope is unrelated.
- Base branch: main
- Merges into: main
- Delivery: A dedicated, unmerged pull request against `main` with the canonical GitHub matrix regenerated only if the generator inputs prove it stale, all required local verification recorded, and a no-mistakes delivery run when Firstmate requests it.
- Working branch: fm/cli-github-certification-matrix-drift-r1
- Task: Reproduce `make connectorgen-certification-matrix` drift at the current `origin/main`, identify its authoritative inputs, and refresh only `internal/connectors/defs/github/certification-matrix.json` through its canonical generator when the data (not the generator) is stale. Preserve certification claims, operation mappings, source provenance, runtime behavior, and every unrelated connector artifact.
- Verification: Record expected RED and passing GREEN matrix checks; run focused generator/GitHub checks; prove semantic and byte-stable regeneration; run connector validation, surface sync, boundary, generated artifact/docs, and repository verification gates; review the final diff and commit only the matrix plus its delivery evidence.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| A stale GitHub matrix is rejected | live | The canonical `make connectorgen-certification-matrix` check exits non-zero before regeneration and names the stale GitHub artifact. |
| The checked-in GitHub matrix equals the authoritative generated artifact | live | The same canonical check exits zero after generation; a second generator run produces no file diff. |
| The correction does not alter certification truth or non-GitHub scope | live | Semantic JSON comparison and `git diff --name-only` show only generator-derived GitHub matrix fields and this phase's evidence. |
| Generator drift protection covers happy, stale, and deterministic paths | live | Existing focused generator tests assert a matching artifact succeeds, a mismatched artifact is rejected, and repeated generation is byte-stable; add a focused regression only if the existing coverage lacks one of those observations. |

## Manual-GSD execution record

Resolved commands:

- `scripts/gsd prompt discuss-phase cli-github-certification-matrix-drift-r1 --auto`
- `scripts/gsd prompt plan-phase cli-github-certification-matrix-drift-r1 --tdd`
- `scripts/gsd prompt execute-phase cli-github-certification-matrix-drift-r1`
- `scripts/gsd prompt verify-work cli-github-certification-matrix-drift-r1`
- `scripts/gsd prompt code-review cli-github-certification-matrix-drift-r1`

The requested shared-base cleanup is not an active numbered roadmap phase, and this runtime cannot use the GSD role agents without violating the repository's single-worker contract. The worker therefore executes the generated GSD prompts inline and records each stage in this phase directory.

Required skills loaded: `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, and `golang-lint`.

## TDD slices

### 1. RED — prove and preserve the stale-artifact failure

At the fetched `origin/main` base, run `make connectorgen-certification-matrix` before any generator invocation. Capture its exit status and non-secret diagnostic. Inspect the generator source and the GitHub bundle inputs to identify the exact authoritative reason for the mismatch. If any input or generator behavior is inconsistent, stop instead of changing generated output.

### 2. GREEN — regenerate and audit only GitHub

Use the repository's certification-matrix generation command with its supported GitHub-only scope. Do not hand-edit JSON. Compare the checked-in and regenerated JSON structurally: verify stable ordering, declared source provenance, every operation's reachability and certification status, and absence of changed claims. Immediately rerun the canonical check and a second generation; the latter must be byte-stable.

### 3. Verification and review

Run the focused generator tests covering matching, mismatch, and deterministic generation; GitHub connector checks; `make connectorgen-certification-matrix`; connector validate and surface-sync; connector boundary; generated docs/artifact checks; and the repository's scoped full-verification alternatives required by `AGENTS.md`. Record all results in `TDD-LEDGER.md` and `VERIFICATION.md`. Complete inline `verify-work` and code review after the diff is final.

## Scope fence

No source code, runtime behavior, connector operation mapping, source-lock contract, certification claim, CLI surface, documentation, website, credential, network I/O, or non-GitHub artifact may change. CLI help/manual/website parity is not applicable: this task changes no command surface.
