Refs #3866
Refs #3855
Refs #4015

## Intent

Add the final deterministic shared-transport conformance matrix for four
**family half-paths**: API source → warehouse, database source → warehouse,
warehouse → API destination, and warehouse → database destination.

## What changed

- Added a private fake/fixture matrix at the shared transport seam. Each of
  the four named families exercises every `synccontract.AllModes()` member and
  asserts produced record IDs, descriptor-owned strategy, durable
  acknowledgement, and committed checkpoint values.
- Added named refusal and edge proofs for unbound source preflight (typed error
  and zero source/stage/plan/apply/commit calls), missing acknowledgement, and
  cancellation after stage.
- Added a secret-free, in-memory coordination proof for verified-auth fencing,
  rate parking, durable conflict, cancellation, restart, and exact
  no-replay checkpoint resume.

## Proof classification and non-overlap

**Every assertion in this PR is simulated, not live certification.** The far
sides are local fakes and memory stores; this PR does not register a production
connector, call a provider/database, alter a certification matrix/capability,
or create a protocol executor.

Merged PR #4195 already proves the different connector-route axis:
API→warehouse→API, API→warehouse→PostgreSQL,
PostgreSQL→warehouse→API, and PostgreSQL→warehouse→PostgreSQL, including
sealed-workset ordering. This PR deliberately does not re-claim that route
coverage. Existing live route proofs remain their original owners' evidence.

## Happy, bad, and edge cases

- Happy: every named family/mode row asserts records staged/applied, selected
  apply strategy, durable acknowledgement, and the exact committed checkpoint.
- Bad: an unbound source returns
  `*DestinationSourceIneligibleError` before any source/stage/plan/apply/commit
  operation; a missing acknowledgement returns
  `ErrDownstreamAcknowledgementRequired` and makes zero checkpoint commits;
  a durable conflict returns `ErrRateParkingConflict`.
- Edge: verified authentication alone fences the cohort; parking resumes only
  at reset from the exact committed checkpoint without replay; cancellation
  suppresses destination apply or scheduled resume as applicable.
- Sensitivity: after `connectorgen validate` passed, a scratch substitution of
  the API-source fixture with a database source made only
  `api_source_to_warehouse/full_append` fail. Restoring the binding made the
  same named test pass. The scratch edit was not committed.

## TDD / GSD

- Red: `go test -count=1 -timeout 20m ./internal/synctransport -run '^TestTransportFamilyHalfPathConformance$'` initially failed because the private family matrix helper was absent.
- Green/refactor: focused shared-transport and coordination packages pass; the
  ledger records the schema-valid named failure demonstration and restoration.
- Lifecycle: `scripts/gsd prompt discuss-phase`, `plan-phase --tdd`,
  `execute-phase --interactive`, `verify-work`, and `code-review` were executed
  inline under the canonical no-role-spawn delivery contract.
- Skills: `golang-how-to`, `golang-design-patterns`,
  `golang-structs-interfaces`, `golang-error-handling`, `golang-security`,
  `golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`,
  `golang-database`, `golang-code-style`, and `golang-lint`.

## Verification

- `go test -count=1 -timeout 20m ./internal/synctransport ./internal/coordination ./internal/synccontract`
- `go test -count=1 -timeout 20m ./internal/app`
- `go test -count=1 -timeout 20m ./internal/cli`
- `go test -count=1 -timeout 20m ./cmd/connectorgen`
- `go test -race -count=1 -timeout 20m ./internal/synctransport ./internal/coordination`
- `go vet ./internal/synctransport ./internal/coordination ./internal/synccontract ./internal/app ./internal/cli ./cmd/connectorgen`
- `go build ./cmd/pm`
- Individual `make verify` gates: format, tidy, lint, docs, smoke,
  agent-contract, generated connector checks, GitHub artifacts, certification
  matrix, connector boundary/canon, and release workflow.
- `pnpm --dir website run gen:docs` twice, with a clean generated diff after
  each run.

The aggregate `go test -timeout 20m ./...` and `make verify` were not run as
single commands because repository policy identifies command-harness timeout
ambiguity; changed packages, consumers, `internal/cli`, `cmd/connectorgen`, and
every non-test verify gate were run individually.

## Review coverage

Inline GSD review found no unresolved findings. The PR's automatic Claude
review route will be checked after opening; any actionable finding will receive
a reasoned disposition before resolution.
