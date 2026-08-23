# Twenty CRM all-ops recovery plan

## Task Delivery Header

- Issue: Refs #277 — Twenty CRM all-ops CLI parity.
- Base branch: `main`.
- Merges into: `main`.
- Delivery: PR #4298 is open against `main`; this reconciliation slice is
  committed, pushed, and its API-reported base is verified after the push.
- Working branch: `fm/cli-twenty-crm-gtm-parity-r1`.
- Task: Preserve all 168 Twenty commands as implemented and make the 28
  documented batch actions explicit, source-traced foundation-blocked rows.
  Captain option C defers only array-envelope *certification* to 0.3.1; it
  does not hide, disable, or mark a working command partial.
- Verification: run the connector-local TDD test, `connectorgen validate`,
  `surface-sync --check`, certification sweep, generated/doc checks, the
  connector boundary, and an actual commandrunner preflight sweep.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Batch actions are explicitly deferred, source-traced, and still reachable | live | The connector test reads every one of the 28 rows, matches its `POST` path to `writes.json`, requires the source URL/`records` envelope/60-record bound and 0.3.1 foundation gap, then proves the corresponding command remains `implemented`. |
| The connector is candidly non-certified | live | The seven-surface ledger and foundation-gap row state the 0.3.1 deferral while preserving all 168 commands; no live credential or runtime artifact is read in this slice. |

## 2026-08-23 captain decision — array-envelope certification

Option C is authoritative for this release: no array-envelope foundation is
built or awaited in this Twenty lane. The 28 exact provider batch actions stay
user-reachable as implemented reverse-ETL commands, but their application
batch certification is explicitly deferred to 0.3.1. This slice adds a
per-action source trace (published URL, exact method/path, envelope property,
and declared maximum) plus a stable foundation gap reference. It deliberately
does not claim a batch dispatch proof, alter foundation code, touch a live
credential, or recast any command as partial.

## GSD and skills

- Lifecycle: `discuss-phase` → `plan-phase --tdd` → `execute-phase` →
  `verify-work` → `code-review`, each resolved through `scripts/gsd prompt`.
- Inline/manual fallback: compatible isolated Pi workers are not available in
  this environment; the single connector worker executes the generated workflow
  inline and records evidence here.
- Required skills loaded: `golang-how-to`, `golang-cli`, `golang-testing`,
  `golang-error-handling`, `golang-security`, `golang-safety`,
  `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`,
  `golang-concurrency`, `golang-database`, `golang-graphql`, and
  `golang-documentation`.

## Foundation check

| Need | Expected proof | Stop condition |
| --- | --- | --- |
| Stream and direct-read runtime dispatch | `commandrunner.Preflight`, targeted command tests, and built CLI reach the Twenty request contract | Any required executor is missing: record `needs-decision` and do not alter foundation code. |
| Typed reverse ETL and delete safety | Generator/conformance tests plus live plan-preview-approval-execute proof | Any bypass of the typed lifecycle: stop. |
| Certification artifacts | `certification-sweep --check` passes for current allowed scope; assess whether Twenty is allowlisted | Adding Twenty to the allowlist would require a foundation decision. |
| Live provider | A dedicated disposable Twenty instance and its registered `pm-twenty` Keychain/secret-store reference authenticate the freshly built binary. The secret is read only in a direct stdin pipe to the CLI credential option; no argv, output, evidence, file, or App state may contain it. | After a genuine start and credential-reference attempt, document the exact missing service/reference or runtime failure and remain uncertified. Never substitute fixture proof or the captain's instance. |

## Reconciliation slice — typed destinations

1. Merge `origin/fm/cli-reverse-etl-destination-r1` without rewriting the
   existing Twenty history and retarget PR #4298 to that exact branch.
2. Audit source-locked operations across `binary_read`, `binary_write`,
   `direct_read`, `direct_write`, `etl`, `reverse_etl`, and executable CLI
   commands. The ledger must account for all 168 published REST operations;
   privileged or destructive operations remain reachable and safety-gated.
3. The refreshed #4304 head `d814875a902be684cb2a38b94f7a8077f66b70b1` is
   merged locally and its PR base was verified through GitHub's API. Its
   persisted selection does not close Twenty's requirement: the source-binding
   lookup is one mapping per source executor/stream, while all 56 record-shaped
   actions require their own exact `input_fields`. Stop the connector lane for
   a provider-neutral per-action source-binding decision; do not create a
   connector workaround or alter foundation code here. After that foundation
   lands, bind all 56 actions with their provider-owned name,
   `full_append`/`append` strategy, input fields, delivery/acknowledgement
   facts, and conformance evidence.
   `write_eligibility.json` must give each of the 112 typed actions an explicit
   disposition. A batch array envelope or a tombstone workset may be a semantic
   exclusion; safety, privilege, and destructive classification may not. No
   connector workaround, action omission, or command demotion is permitted.
4. Add red/green tests for declared destination coverage, an invalid/missing
   input binding, and the binary-file boundary. Regenerate connector-owned
   projections and run focused generator, binary, and live reversible proof.
   The declaration test is not application-dispatch proof: before final push,
   fetch and merge the latest #4304 head, prove it is an ancestor, and exercise
   the installed App/CLI path that selects the generic destination.
5. Record the seven-surface ledger, the exact PR dependency, command results,
   and the API read-back of PR #4298's base.

### Foundation-handoff witness (connector-local only)

Before a foundation lane changes a shared contract, keep a Twenty-local
behavioral witness that exercises the current public descriptor: extend its
declaration in memory with `update_companies`, select that action, and show
that `SourceBindingFor(declarative_stream_source, "companies")` still returns
the `create_companies` `name` → `name` mapping rather than the required
`id` → `id`, `name` → `name` mapping. The witness must cover the happy
selection, the bad mismatch before provider I/O, and the edge case where an
undeclared action is refused. It must not propose or carry an engine change.

The handoff contains the exact `writes.json`, `write_eligibility.json`, and
`sync_transport.json` hashes; the descriptor and App call-path source lines;
the 55/28/28 dependency membership selectors; and closure tests for a separate
provider-neutral foundation lane. A source hash mismatch is a red ledger
failure, not a reason to reduce the Twenty acceptance criteria.

## Captain hard certification gate — dynamic workspace inventory

Static Twenty REST declarations remain necessary provenance but are not final
certification evidence. After the non-empty credential foundation and final
#4304/#4305 heads are integrated, dynamically inventory the existing dedicated
Twenty workspace through documented authenticated provider surfaces. For every
provider-defined object, field, operation, and executable application command
that results from that inventory, record its provider source/schema trace and
reconcile it to the checked-in declaration, generated CLI command, help/manual,
and website documentation.

The reconciliation ledger has these six mandatory capability surfaces:

1. ETL source reads.
2. Reverse-ETL typed actions.
3. Direct reads.
4. Direct writes.
5. Binary downloads.
6. Binary uploads.

Every discovered operation must be enabled and remain reachable through the
installed CLI and documentation surface, or be represented as an exact,
provider-sourced capability fact where the provider has no operation for that
surface. Scope, tier, privilege, destructive classification, confirmation, and
authorization are execution metadata only; none may demote, hide, or disable a
provider operation. The binary ledger must distinguish documented transfer
operations from attachment JSON metadata and must never infer a transfer route.

The live gate uses only records/resources created by this lane. It must prove
the supported CRUD/application paths with independent provider read-back and
cleanup, preserving ordinary provider output fields in retained local evidence
while continuing to exclude credentials, transport secrets, session material,
and identifiers from durable artifacts. No merge-ready claim is permitted until
the dynamic inventory and all applicable six-surface proofs are complete.

`FOUNDATION-GAPS.json` is the machine-readable companion to this gate. It
deduplicates every currently known provider-neutral missing capability, locks
each affected Twenty operation set to a source-file hash and selector, and
contains per-batch plus portfolio rollups. A row whose status is open makes all
of its selected operations ineligible for a merge-ready verdict; it is never a
reason to hide an operation as disabled or N/A.

## TDD slices

1. Add a Twenty-local source contract test that fails while the recovered bundle
   is absent and later asserts all 168 commands are executable/fully classified.
2. Cherry-pick the recovered bundle only, run the red structural/generator tests,
   and repair current-main declaration drift inside `defs/twenty`.
3. Add or tighten connector-local tests for representative list/get/pagination,
   write validation, typed deletes, invalid inputs, and edge pagination/output
   cases. Make each red result durable in `TDD-LEDGER.md` before its green fix.
4. Regenerate only connector-owned derived files and run targeted gates.
5. Build `pm`; establish or use the dedicated disposable self-hosted Twenty;
   obtain the registered `pm-twenty` secret only at process execution and pipe
   it to the CLI stdin-secret option. Prove authenticated pagination/list/get,
   lane-owned create/read-back/update/delete/cleanup, ETL execution,
   reverse-ETL plan/apply/acknowledgement with independent API read-back, and
   a binary round-trip only if Twenty documents a binary endpoint. The old
   live run is historical only and does not satisfy this post-#4304 gate.
6. Run the required verification and review workflow. Before the final push,
   fetch and merge the latest #4304 head without rewriting history, prove that
   exact SHA is an ancestor, exercise its installed App/CLI dispatch path, push
   the stacked branch, and read PR #4298's #4304 base back through the GitHub
   API.
7. Audit the current live-read HTTP 403 without changing the disposable API
   key: prove the encrypted-vault token's length/hash is identical at the
   actual runtime bearer header, inspect only non-secret API-key role/expiry
   state through the official UI, and compare Twenty's pinned-source JWT/API
   key guards plus boolean-only self-host configuration. If the header path is
   connector-owned, add a failing focused test then fix it locally; if it is a
   shared-engine gap, record the provider-neutral dependency and stop.
8. After the shared credential-input and final destination foundations publish,
   run the captain hard gate: dynamically inventory the disposable Twenty
   workspace schema, attach a provider source/schema trace for every discovered
   operation, and reconcile all six capability surfaces to declarations,
   executable CLI paths, help/manual output, and website docs. Exercise every
   applicable supported CRUD/application command against lane-owned records
   with independent provider read-back and cleanup. Safety/tier/scope metadata
   governs execution but never removes a discovered provider operation.

## CLI/docs parity checklist

- [ ] `pm twenty` contextual help exits successfully.
- [ ] `pm help twenty` and representative `pm twenty <object> <action> --help`
  reflect the recovered surface.
- [ ] Generated CLI/manual/docs checks pass; any required non-definition output
  is treated as a foundation/scope decision rather than silently omitted.
- [ ] JSON output, credential safety, pagination, and reverse-ETL safety gates
  are represented in connector documentation and live evidence.

## Planned gates

```text
go test -timeout 20m ./internal/connectors/defs/twenty
go test -timeout 20m ./internal/connectors/conformance -run 'TestConformance/twenty'
go run ./cmd/connectorgen validate internal/connectors/defs
go run ./cmd/connectorgen surface-sync --check
go run ./cmd/connectorgen certification-sweep --connector twenty --check
make connector-runtime-preflight
make connector-canon-check
make connector-boundary
go build ./cmd/pm
make lint
make verify
```
