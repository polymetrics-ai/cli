# GitHub command certification product-defect fixes

## Task Delivery Header

- Issue: Refs #4015 — Production MVP; Refs #4221 — reverse delete false success.
- Base branch: `integration/4015-mvp-flat-r1`.
- Merges into: `integration/4015-mvp-flat-r1` → `main`.
- Delivery: Open a direct pull request from `fm/cli-parity-fix-defects` against the stated base with the census, red/green evidence, live conversion evidence, and repository checks committed; then verify the API-reported base.
- Working branch: `fm/cli-parity-fix-defects`.
- Task: Classify all 277 supplied GitHub product-defect commands, fix exact integer preservation, provider-required request bodies, invalid GitHub endpoint declarations, and false-success write accounting at their shared sources, then prove at least three real command conversions per fixed class without retaining credentials or fixtures.
- Verification: Targeted Go tests, GitHub bundle validation and generated-surface checks, executable command preflight, CLI/help/docs parity checks, repository non-suite gates, live provider read-back and cleanup proof, secret scan, automated code review, and API read-back of the pull request base.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Every supplied product defect is assigned one primary class | live | `CENSUS.md` accounts for exactly 277 unique slice/command pairs and its five buckets sum to 277. |
| Integer identifiers survive plan persistence exactly | live | A persisted command containing the exact integer `9007199254740993` reaches a fake provider path as those decimal digits; the pre-fix path is rounded/scientific. Three real GitHub commands then read back their expected provider effects. |
| Required GitHub bodies are expressible and transmitted | live | Fake-provider tests reject an empty body and assert typed required fields for each repaired action; three real commands return provider-created objects whose fields match the requested values. |
| Wrong GitHub paths are removed from executable declarations | live | Bundle tests assert the corrected operation, write, and command endpoints; three real commands mutate/read back state through the corrected paths. |
| Missing provider responses cannot be counted as writes | live | A fake provider 404 produces zero successful writes, a failed record, and a non-zero command result; real create/delete commands are accepted only when independent read-back matches. |
| Live fixtures and credentials are contained | live | Every `pm-cert-` object created by this task is deleted through the provider and independently returns 404/absence; a branch scan finds no credential material. |

## GSD lifecycle and required skills

The repository-local adapter resolved `discuss-phase`, `plan-phase --tdd`,
`execute-phase`, `verify-work`, and `code-review`; `scripts/gsd doctor` and
`go run ./cmd/agentcontractgen check` passed before production work. This lane
uses the documented inline/manual GSD fallback because the launch contract
requires one autonomous direct-PR worker and the current runtime does not
provide compatible project-local Pi workers. The lifecycle remains explicit in
this plan, `TDD-LEDGER.md`, `RUN-STATE.json`, `VERIFICATION.md`, and `REVIEW.md`.

Required skills reviewed: `golang-how-to`, `golang-cli`, `golang-testing`,
`golang-error-handling`, `golang-security`, `golang-safety`,
`golang-design-patterns`, and `golang-structs-interfaces`. Repository routing,
GSD adapter, issue-agent, connector canon, and CLI/help/docs parity references
were also reviewed.

## Discuss-phase decisions

- The 77% precision-loss figure from the 39 reasoned rows is a sample, not the census. The command-level census is completed and frozen before production edits.
- Classification is by the first independently supported root cause. Recorded live reasons outrank static declaration analysis. Overlaps are retained in notes but counted once.
- Exact integer preservation belongs at JSON state decode, the place typed command records currently become `float64`; call-site ID special cases are forbidden.
- Missing bodies belong in the GitHub write schemas so ordinary surface derivation creates typed flags and every caller shares the contract.
- Provider paths are corrected across runtime declarations and their derived command/API artifacts. Declared commands remain implemented.
- A provider `missing_ok_status` response is an idempotent no-op, not a successful write. It must never increment `RecordsWritten`; command execution must surface the incomplete acknowledgement.
- Live writes remain inside `Polymetrics-Cert/pm-cert-3993-20260810-wz0fru`. Credentials are read from Keychain only into process stdin/environment at point of use, never argv or evidence.

## TDD execution plan

1. Commit this plan, census, red/green ledger, verification checklist, and run state.
2. Add red tests for exact persisted integer transport, six required-body command contracts, corrected paths, and missing-delete accounting. Run each narrow test and record its failure.
3. Make the smallest shared-source and GitHub bundle changes; regenerate only canonical derived artifacts using repository generators.
4. Run each narrow test green, then the changed packages, `internal/cli`, bundle validation, surface sync, certification matrix, docs/help parity, and remaining repository gates.
5. Build `pm`; perform bounded live conversion probes for at least three commands in each fixed class. Assert created/updated/deleted provider state independently, clean every fixture, and assert 404/absence.
6. Scan the branch for credentials and uncontained `pm-cert-` artifacts, run the code-review lifecycle, disposition findings, commit, push, open the direct PR, and verify `.base.ref` through the GitHub API.
