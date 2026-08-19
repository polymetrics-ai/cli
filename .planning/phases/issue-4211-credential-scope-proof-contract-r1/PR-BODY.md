Refs #4211
Refs #4015

## Summary

Accepted evidence is now schema v2 and carries both `credential_scope` and a
matching `credential_scope_proof`. The full-parity construction path takes an
executed `certify.Report` and refuses to construct a full claim unless its
`full_parity` stage passed. The normal proof writer publishes a truthful,
bounded `observed_operations` claim from the non-empty protocol-exchange
transcript it serializes. Callers no longer provide a scope attestation.

The pre-existing exact full-parity validator remains in use; the new validator
additionally binds each supported scope to its proof discriminator. Generated
matrix pointers and the agent-contract reader preserve and independently check
that discriminator.

## PostgreSQL evidence re-issue

All 14 PostgreSQL evidence records were re-issued from fresh opt-in live runs:
12 transport records from `TestPostgresCertificationProfileRunsBuiltBinaryLive`
and 2 CDC records from `TestPMBinaryDispatchesPostgresChangeCaptureToWarehouse`.
Each new record is v2 `observed_operations` with `protocol_exchanges` proof and
the explicit note that no broader credential scope is claimed.

The original records' full-parity claim was **unverified, not merely
unrecorded**. Their tests invoked `--full --write`; only `--full-parity` sets
`RequireFullParity`, so the report had no full-parity stage. The importer then
hard-coded the old `CredentialFullParity: true` caller assertion. No record was
re-stamped with that claim; none was dropped or silently downgraded—the fresh
run produced each replacement record.

## Guard demonstrated failing, then passing

- **Failing before the change:**
  `go test -timeout 20m ./cmd/connectorgen -run '^TestCertificationBoundedScopePublishesObservedOperations$' -count=1`
  failed with `completed live test did not use a full-parity credential`.
- **Passing after the change:**
  `go test -timeout 20m ./cmd/connectorgen -run '^(TestCertificationBoundedScopePublishesObservedOperations|TestCertificationFullParityScopeRequiresPassedReportStage|TestCertificationPublishesNarrowCredentialEvidence)$' -count=1`
  passed. `TestCertificationFullParityScopeRequiresPassedReportStage` first
  proves the missing-stage refusal, then proves a passed stage succeeds.

## Validation

- Passed: `go test -timeout 20m ./cmd/connectorgen`,
  `go test -timeout 20m ./internal/connectors/certify`,
  `go test -timeout 20m ./internal/cli`, and
  `go test -timeout 20m ./internal/agentcontract`.
- Passed: `go vet ./...`, `make fmt`, `make tidy-check`, `make build`,
  `make docs-check`, `make smoke-no-build`, `make lint`,
  `make connectorgen-validate`, `make connectorgen-surface-sync`,
  `make github-parity-artifacts-check`,
  `make connectorgen-certification-matrix`,
  `make connectorgen-certification-sweep`, `make connector-boundary`,
  `make connector-canon-check`, `make release-workflow-check`, and
  `go run ./cmd/agentcontractgen check`.
- PostgreSQL re-issue passed with the documented direct Colima Unix endpoint:
  transport 41.779s, CDC 30.230s. Matrix generation was run twice with matching
  SHA-256 `61e24fd58a9fcbc8debcd65c0258784ba020739ebd3a2a6e19cdb2d3ae4f7d70`.
- `make verify` was not run as one aggregate process because the repository's
  per-command-timeout instruction explicitly requires individual gates for the
  550+ connector suite; all relevant gates above were run individually. The
  identical base-branch `security/snyk` failure is pre-existing and ignored.

## Lifecycle and review

Used the required issue-first GSD lifecycle through the project adapter:
`discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and deep
`code-review`. The active runtime forbids GSD role spawning, so the documented
inline/manual fallback and its TDD, verification, and review evidence live in
`.planning/phases/issue-4211-credential-scope-proof-contract-r1/`.

Required skills used: `golang-how-to`, `golang-design-patterns`,
`golang-structs-interfaces`, `golang-error-handling`, `golang-security`,
`golang-safety`, `golang-testing`, `golang-database`, and `golang-lint`.

Automated-review route: `claude_auto`, pending the trusted-author PR-open
trigger. This is a stacked PR into `integration/4015-mvp-flat-r1`; parent PR
#4100 targets the default branch. Fallback: none unless the automatic route
fails or is unavailable.

## Rules for the rulebook

1. An evidence claim must be derived from an executed, persisted verifier fact;
   a caller flag, a writer constant, or a duplicated string comparison is an
   assertion, not proof.
2. A reader needs the claim *and* the discriminator for the fact that proved it;
   preserve both through every generated pointer and independent consumer.
3. When full scope was not verified, publish a bounded lower-bound claim tied to
   the transcript that was verified—never promote it or reject it solely because
   it is narrower.
4. Every guard needs a demonstrated failing state: a report without the relevant
   stage must be unable to construct the broader record before the passed-stage
   control can be trusted.
