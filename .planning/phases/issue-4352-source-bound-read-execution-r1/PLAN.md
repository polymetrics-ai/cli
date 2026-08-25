# Source-bound read execution foundation

## Task Delivery Header

- Issue: Closes #4352 — add source-bound read execution foundation.
- Base branch: main.
- Merges into: main.
- Delivery: Pull request open against `main` with the shared foundation committed, pushed, focused verification green, and the GitHub API-reported base verified as `main`.
- Working branch: `fm/cli-source-bound-read-execution-r1`.
- Task: Add the smallest declarative generator/runtime extension that makes a source-locked, non-mutating provider operation callable through an exact, closed operation binding. A command must reach the normal credential boundary without provider I/O, live certification, a source hash requirement, or a connector-local shim. It must distinguish bounded direct reads from genuinely stream-capable reads and refuse missing shared foundation before I/O.
- Verification: Focused red/green generator and command-runner tests; source/generator checks; built `pm` preflight for a representative read in an isolated, credential-free project; targeted regressions for direct writes, reverse ETL, binary, and delete behaviour; `go vet ./...`, `go build ./cmd/pm`, relevant individual `make verify` gates, and `git diff --check`.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Source projection considers non-mutating source operations | live | A generated-source fixture preserves a GET operation in the source-bound declaration path while an existing mutation control retains its write behavior. |
| Generated reads remain source-bound and closed | live | Tests assert the command's locked operation identity and exact method/path; a substituted route is refused before a requester can observe I/O. |
| Read semantics are honest | live | Tests classify a source-backed singleton/direct response separately from a record/pagination-backed collection and retain a named missing foundation where the source contract cannot establish a stream. |
| Missing inputs or foundation fail before provider I/O | live | Requester spies observe zero calls while the stable actionable preflight error names the missing source-bound contract/foundation. |
| Valid generated read reaches credentials | live | A credential-free built-binary invocation of the selected command stops at `missing --credential`, rather than unknown command, unsupported preflight, or provider I/O. |
| Existing safety behaviour is unchanged | live | Focused existing direct-write, reverse-ETL, binary, and delete controls continue to pass. |

## GSD discussion record

The launch brief resolves the relevant design choices: this is a shared foundation issue rather than Batch-1 connector work; its command surface is operation-identity-bound and has no arbitrary request escape hatch; direct reads are one bounded response; ETL promotion is allowed only with source-backed pagination and record semantics; materialization and runtime certification remain independent. The canonical isolated-worker runtime is unavailable and the repository delivery contract forbids role spawning, so the resolved GSD prompts are executed inline/manual.

## Required skills and CLI parity

Loaded: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, and `golang-structs-interfaces`.

This changes generated connector command execution, not a hand-authored top-level `pm` help tree. Before handoff, record the applicable generated command/help behavior and inspect `docs/cli/**`, `website/**`, and generated manual/help artifacts; explicitly mark any unchanged surface as not applicable with the reason.

## TDD slices

1. **Projection and declaration admission.** Red: add a source-projection fixture with a non-mutating operation and assert the current generator leaves it declaration-pending or otherwise lacks a closed execution binding. Green: generate only source-locked GET/read bindings with exact identity/method/path and typed inputs; leave non-supported contracts in a named declaration/foundation state. Preserve mutation/write controls.
2. **Runtime preflight and dispatch.** Red: prove a generated read has no credential-bound execution path and that a route substitution/missing contract can reach no requester. Green: resolve the command only through the embedded source-bound operation ledger, validate identity/method/path and typed inputs, then use the existing credential/auth preflight. Do not add arbitrary method/path/header/body input.
3. **Honest read class.** Red: exercise a singleton/direct-response operation, a genuinely source-backed paginated/record collection, and a path-parameter one-object operation. Green: dispatch a bounded direct read or existing stream only where its source contract proves semantics; otherwise return stable `missing_foundation` before I/O.
4. **Real locked controls.** Red: identify the Asana `getAccessRequests` source row plus available collection/pagination and path-parameter source rows, then add tests that fail on the current foundation. Green: prove the materialized commands reach the credential boundary without a provider request.
5. **Regression and verification.** Run focused cross-package tests plus existing direct-write, reverse-ETL, binary, and delete controls. Review the diff for input-boundary escape hatches, error leakage, and behavior regression.

## Expected change boundaries

- `cmd/connectorgen/sourceprojection.go` and its tests
- `internal/connectors/{engine,commandrunner}/` only where a source-bound operation/read contract requires it, with focused tests
- Generated, source-backed fixture/artifact files only when required to demonstrate the three named read shapes
- This phase's `PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, `RUN-STATE.md`, `SUMMARY.md`, and `REVIEW.md`

No connector-wide Asana operation materialization, certification policy, live provider calls, generic request interface, or edits to another worker's branch belong in this PR.
