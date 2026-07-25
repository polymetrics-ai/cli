# Plan — Bahmni rename and official-operation parity follow-up

## Objective

Apply the captain-authorized follow-up on branch `feat/bahmni-docker-connector`:

1. Rename the connector identity from `bahmni-docker` / "Bahmni Docker" to `bahmni` / "Bahmni" across the declarative bundle, generated docs/catalogs, website data, help examples, tests, and GitHub issue/PR surfaces. Preserve `Bahmni/bahmni-docker` only when referring to the deployment repository/setup.
2. Fix the four authorized no-mistakes review findings:
   - redact patient-document file paths by using a `file_path`-marked field name;
   - disable inherited `offset_limit` pagination for the four root-array streams;
   - replace `/ws/rest/v1/session` as the credential check with an authenticated data endpoint;
   - correct the `drug-orders` vs `drug_orders` group/command mismatch.
3. Verify official-operation parity from Bahmni/OpenMRS sources rather than from the current issue list alone, and record an operation-to-subissue matrix with implementation/exclusion/test evidence.
4. Run the repository's current PM/no-mistakes orchestrator/review workflow on the exact candidate head, resolve every authorized in-scope finding, and update PR #533 without merging.

## GSD runtime

`scripts/gsd doctor` is healthy in this worktree. The documented `programming-loop` shell command is not registered by this branch's adapter (`scripts/gsd: unknown GSD command: programming-loop`), so this follow-up records an explicit manual-GSD fallback under the official adapter: plan before edits, TDD/verification ledger, and review/workflow evidence.

## Required skills / references loaded

- `gsd-core`
- `.agents/agentic-delivery/references/required-skills-routing.md`
- `.agents/agentic-delivery/references/gsd-pi-adapter.md`
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`
- `golang-how-to`, `golang-cli`, `golang-testing`, `golang-security`, `golang-safety`, `golang-documentation`, `golang-design-patterns`, `golang-error-handling`
- `no-mistakes` skill was already loaded earlier in this branch session.

## Safety boundaries

- Do not decide the two captain-owned findings: global PHI field-redaction semantics and nullable `diagnoses.existingObs` primary-key semantics.
- Do not claim runtime PHI field redaction exists where the engine does not enforce it.
- No live credentials or credentialed Bahmni checks.
- No generic raw HTTP/write escape hatch.
- No reverse-ETL execution outside plan → preview → approval → execute.
- Do not merge.
