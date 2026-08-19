# GitHub mutation certification — slice 5 writes-e

## Task Delivery Header

- Issue: Refs #4015 — Production MVP certification
- Base branch: `integration/4015-mvp-flat-r1`
- Merges into: `integration/4015-mvp-flat-r1` → `main`
- Delivery: A pull request from `fm/cli-mut-slice5-writes-e` is open against the exact base and contains only valid, independently supported certification evidence and this delivery record.
- Working branch: `fm/cli-mut-slice5-writes-e`
- Task: Attempt each of the 145 assigned GitHub reverse-ETL commands serially, execute only contained non-billing mutations, prove provider state changes and direct-API cleanup before recording any pass, and classify every non-pass honestly.
- Verification: `go run ./cmd/connectorgen certification-matrix --check`; targeted connector CLI/preflight checks; `git diff --check`; review the PR base with the GitHub API.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| A certified mutation changes a provider resource | live | Independent read-back observes the tagged mutation, which would be absent if `run` were a no-op. |
| Cleanup is complete | live | Direct GitHub API DELETE followed by a 404 or empty collection; mutation-command exit status is never used as cleanup evidence. |
| Metered compute is never provisioned | live | Codespaces create/start/resume commands are recorded `escape_needs_captain` without `plan`, `preview`, or `run`. |
| Published evidence is schema-valid | live | `certification-matrix --check` accepts only records that encode a provider-backed proof. |

## GSD and safety execution

Manual inline fallback: this runtime cannot run the adapter's Pi workflow. Generated prompts were resolved for `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review`; their required activities are recorded here and in the ledger.

Required skills loaded: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, and `golang-structs-interfaces`.

The firstmate decision excludes any operation that provisions/resumes Codespaces compute as `escape_needs_captain` (`metered compute - real money`). Metadata-only Codespaces operations remain eligible. The same containment rule excludes other external effects that leave the disposable organisation, fixture repository, or certification user.

## Red/green plan

1. **Red:** Before live writes, inspect the command ledger and certification validator to confirm the proposed evidence schema can represent a per-command mutation proof. Record any schema mismatch as a product defect rather than manufacturing an accepted record.
2. **Green:** For each eligible command, use `plan` → read its one-time token → `preview` → `run`; independently read provider state; direct-delete the disposable resource; independently prove absence.
3. **Refactor:** Publish only records accepted by the existing validator; do not regenerate shared matrices while parallel lanes are active.
4. **Review:** Independently validate classifications, safety escapes, committed file scope, and PR base.
