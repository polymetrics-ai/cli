# Twenty CRM all-ops recovery plan

## Task delivery header

- Primary issue: #277 — Twenty CRM all-ops CLI parity.
- Active branch / PR: `fm/cli-twenty-crm-gtm-parity-r1` / #4298, stacked on
  `fm/cli-reverse-etl-destination-r1` / #4304.
- Scope: connector-local Twenty definitions, generated projections, and the
  dedicated disposable `pm-twenty-cert-r1` live-certification evidence. No
  production GTM workspace, generic writer, foundation bypass, force-push, or
  merge is permitted.
- Delivery mode: direct PR; status evidence is appended only at supervisor
  checkpoints. Every commit and the PR body carry `Refs #277`.
- Current live-certification gate: diagnose the bounded companies-list HTTP
  403 read-only. First prove the built CLI and an independent bearer request
  return HTTP 200 before any mutation or reverse-ETL proof.

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
3. After the refreshed #4304 head is merged, bind all 56 record-shaped actions
   (the existing `create_companies` proof plus all 55 eligible actions) through
   its persisted per-action selection model. Every binding must retain the
   provider-owned action name, `full_append`/`append` strategy, exact
   `input_fields`, delivery/acknowledgement facts, and conformance evidence.
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
