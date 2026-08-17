# Plan — Issue 4211 provable credential-scope contract

**Status:** implemented; pending direct-PR review

## Task delivery header

- **Issue:** #4211, `fix(certification): verify accepted full-parity credential scope`
- **Parent:** #4015
- **Branch/base:** `fm/cli-credential-scope-proof-contract-r1` from
  `origin/integration/4015-mvp-flat-r1` at `db494bc8fba63023b9ca79b022c2b9dd638aaf76`
- **Delivery mode:** direct PR to `integration/4015-mvp-flat-r1`
- **Isolation:** `/Users/karthiksivadas/.treehouse/cli-83d592/14/cli` verified by
  `pwd -P` and `git rev-parse --show-toplevel`
- **Template fallback:**
  `.agents/agentic-delivery/contracts/task-delivery-header-template.md` is absent
  from the base; this section is its recorded manual equivalent.
- **PR-base check:** after opening the PR, read the API-reported base and record it
  in this phase's verification artifact and PR body.

## Lifecycle and skills

- GSD commands resolved and generated inline: `discuss-phase 4211 --auto`,
  `plan-phase 4211 --tdd --skip-research`, `execute-phase 4211 --interactive`,
  `verify-work 4211 --auto`, and `code-review 4211 --depth=deep`.
- Inline/manual fallback: the active runtime forbids role spawning. The generated
  workflow's analysis, TDD, verification, and review gates are performed in this
  one isolated worktree and recorded here, in `TDD-LEDGER.md`, `VERIFICATION.md`,
  and `REVIEW.md`.
- Required skills loaded: `golang-how-to`, `golang-design-patterns`,
  `golang-structs-interfaces`, `golang-error-handling`, `golang-security`,
  `golang-safety`, `golang-testing`, `golang-database`, and `golang-lint`.

## Confirmed facts and scope fence

1. `newProofBearingEvidence` unconditionally serializes `full_parity` and its
   note after accepting a caller-supplied `CredentialFullParity` boolean.
2. `validateFullParityCredential` only compares the serialized strings to the
   writer's constants, so it cannot establish that a run verified the claim.
3. A `DirectReadOnly` report returns before the full-parity stage and
   `Report.FullParityVerified()` is therefore false.
4. The existing PostgreSQL evidence was produced by `--full --write`, not
   `--full-parity`; `RequireFullParity` was false, so the report omitted the
   full-parity stage. Its importer hard-coded `CredentialFullParity: true`.

Captain decision `cli-existing-evidence-reissue-decision-r1`: re-issue all 14
PostgreSQL records. They must visibly carry an honestly bounded scope unless a
fresh report proves full parity; no old `full_parity` claim is restamped.

This issue owns the shared evidence scope contract and PostgreSQL re-issuance.
It does not redesign the in-flight generic importer; only its compile-required
removal of the obsolete caller attestation is in scope.

## Contract

New evidence records are schema v2. The writer derives exactly one scope:

- `full_parity`, only through a passed `Report.FullParityVerified()` result;
  the serialized proof discriminator is `full_parity_stage`.
- `observed_operations`, the publishable bounded scope derived from the
  non-empty, sanitized protocol exchanges that this evidence record carries;
  its discriminator is `protocol_exchanges` and its note explicitly disclaims a
  broader credential claim.

The reader sees both `credential_scope` and `credential_scope_proof`; no caller
passes a credential-scope boolean or a free-form scope. The old full-parity
validator remains in the full-parity path and the new validator additionally
requires the matching proof discriminator. Schema-v1 records are replaced in
this change, not silently tolerated as proof under the new contract.

## TDD slices

1. **Guard failure (red).** Add a test that supplies a passed, direct-read-like
   report with no full-parity stage to the full-parity construction path. It must
   fail with the verified-stage refusal; the prior implementation would have
   emitted `full_parity` when a boolean was set.
2. **Derived scopes (green).** Add the report-derived full-parity construction
   path and a bounded default construction path. Assert the former emits
   `full_parity/full_parity_stage` only after a passed stage, and the latter
   emits `observed_operations/protocol_exchanges` and validates.
3. **Reader and validator (red/green).** Require `credential_scope_proof` in a
   v2 full-parity record, reject a mismatched/missing discriminator, and repeat
   it in generated evidence pointers. Retain and call the existing exact
   full-parity validator rather than removing it.
4. **Re-issue (green).** Run the opt-in PostgreSQL live certification evidence
   writer after the contract lands. It must write all 14 named evidence records
   with bounded scope and proof discriminator, then regenerate the PostgreSQL
   certification matrix twice to prove byte stability.

## Verification plan

- `go test -timeout 20m ./cmd/connectorgen` including the red/green tests.
- `go test -timeout 20m ./internal/connectors/certify` and
  `go test -timeout 20m ./internal/cli` as changed-package consumers.
- Opt-in PostgreSQL live re-issue with the direct Colima endpoint documented in
  `internal/connectors/native/dbtest/README.md`; no shared runtime is restarted.
- `gofmt`, `go vet`, lint, all individual `make verify` non-suite gates,
  `go run ./cmd/connectorgen certification-matrix --check`,
  `go run ./cmd/connectorgen surface-sync --check`, and `make connector-boundary`.
- CLI help/docs/website parity is not applicable: no `pm` command, flag, help,
  docs source, or website source changes. The evidence generator's own package
  is explicitly included.
