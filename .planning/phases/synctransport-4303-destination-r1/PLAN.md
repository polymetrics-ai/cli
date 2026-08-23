## Task Delivery Header

- Issue: Refs #4303 — feat(synctransport): compose connector-neutral typed reverse-ETL destinations.
- Base branch: main
- Merges into: main.
- Delivery: A pull request from `fm/cli-reverse-etl-destination-r1` to `main`, with local verification, live GitHub parity evidence, generated checks, and review complete.
- Working branch: fm/cli-reverse-etl-destination-r1
- Task: Replace the GitHub-only destination registration with a definition-selected, connector-neutral typed-action destination adapter. Prove two synthetic declarations compose through App and the orchestrator, retain GitHub issue-label behavior through the new path, fail closed before I/O, publish the mechanical declaration contract, and make every declared typed action selectable only through a stable persisted connection stream identity.
- Verification: Targeted red/green tests in `internal/app`, `internal/connectors`, and `internal/synctransport`; the purpose-built real GitHub proof; `go test -timeout 20m` for affected packages plus `internal/cli`; `go vet ./...`; `go build ./cmd/pm`; all `make verify` gates including detached/polled `connector-boundary`; generated/snapshot checks; and an inline code review.

## GSD execution record

The required adapter was validated with `scripts/gsd doctor`, `scripts/gsd sources` for `discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and `code-review`, and generated prompts for the same commands (with `plan-phase --tdd`). `go run ./cmd/agentcontractgen check` passed. This direct-PR worker follows the lifecycle inline because compatible isolated workers are unavailable and the canonical single-worker contract forbids role spawning.

Required skills: `golang-how-to`, `golang-cli`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, and `golang-documentation`.

## TDD slices

1. **Red — generic selection:** Add a synthetic declarative API connector whose destination describes a typed named action, source field bindings, acknowledgment, and closed strategies. Show App composition rejects it because the destination factory accepts only `issue_label_destination`.
2. **Green — generic typed adapter:** Make production factory selection definition-derived for a new declarative typed-destination executor. Build the same typed adapter for every declaring connector, derive its contract only from that connector’s declaration and action metadata, and register it once.
3. **Red/Green — tenancy and refusals:** Add a second synthetic destination with distinct evidence and action metadata. Assert each gets its own evidence admission and action contract; malformed action contract, unknown executor/evidence, wrong source binding/role, and `change_capture` are refused before reads, plans, writes, or read-back.
4. **Parity/refactor:** Route GitHub’s issue-label contract through the same definition-evidence/factory loop while retaining its specialized typed adapter and provider-state read-back. Delete the bespoke composition branch only after the existing approval, recovery, and real-provider proof pass.
5. **Documentation/review:** Update the declaration guide with exact typed action and acknowledgement requirements, no-generic-writer boundary, mode/action strategies, source binding constraints, and evidence admission. Run full local verification and record review disposition.
6. **Red/Green — persisted multi-action dispatch:** Add a production-shaped synthetic connector with two declared actions in one mode and another connector with one. Show the persisted `StreamConfig.destination_action` selects the exact descriptor action through `App.RunETL`; no runtime request selects an action. Reject omitted, foreign, malformed, and cross-connector selections before source or provider I/O. Extend the application command manual to name the `pm etl transport declarative-typed-destination` plan/preview path and `pm etl run --approval-plan …` execute path.
7. **Red/Green — exact action-schema source fields:** Add cross-connector synthetic snake_case and camelCase actions. Accept an `input_fields` name only when the exact selected action's top-level `record_schema` property exists; reject empty, malformed, unknown/cross-action, runtime-selected, generic/shell/http, and undeclared names before I/O, with no provider-specific branch.
8. **Red/Green — complete persisted reverse result:** Add a provider-shaped synthetic typed-action response and show the persisted App run and CLI JSON projection retain ordinary response status, headers, nested fields, and tier-specific fields. Provider-returned fields, keys, and values remain verbatim even when equal to configured credential bytes; system-generated plans, logs, request diagnostics, and synthetic errors remain secret-taint-safe.

## Commit checkpoints

1. Plan evidence (`Refs #4303`).
2. Red synthetic composition test (`Refs #4303`) if a standalone red checkpoint is safe to commit.
3. Green implementation, regression tests, documentation, and generated evidence (`Refs #4303`).
4. Review fixes and final verification evidence (`Refs #4303`).

## Evidence table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| A synthetic definition declares and runs a typed destination through generic App/orchestrator composition | fake | A synthetic in-memory bundle is necessary to prove arbitrary declaration shape without modifying connector definitions owned by other workers; it asserts read, stage, plan, apply, read-back, and commit each occur once. |
| A second destination remains independent | fake | A second in-memory bundle is necessary for the same ownership reason; it has distinct evidence and action metadata and is admitted only through its own factory evidence. |
| Malformed, unknown, wrong-role, and capture destination declarations fail closed | fake | In-memory declarations can assert zero source reads, destination plans, applies, and read-backs after the refusal. |
| GitHub issue labels retain observable write/reconciliation behavior | live | The purpose-built real GitHub API proof executes the named action through plan, preview, approval, run, acknowledgement, and read-back; the independently read label is the required effect. |
| Connector authors can declare destinations mechanically | live | `docs/sync-transport-definition.md` states the exact accepted fields, typed-action boundary, evidence rule, source binding rule, acknowledgement, and mode strategy requirement. |

## Completion evidence

The earlier generic-adapter evidence remains valid. The application-dispatch
reconciliation is complete: its red/green evidence includes persisted exact
multi-action selection, selected-action schema spelling, complete verbatim provider output and secret-safe system diagnostics,
generated CLI documents and transcripts, a clean full `make verify`, and fresh
inline review. The prior real-provider GitHub proof remains the compatibility
evidence: this reconciliation keeps its specialized adapter and read-back
unchanged.

## Follow-up foundation r1 — action-owned mappings, batch, and tombstones

### Task Delivery Header

- Issue: Refs #4303 — feat(synctransport): compose connector-neutral typed reverse-ETL destinations.
- Base branch: `fm/cli-current-foundations-postfix-fix-wave-r1` at `c3f83cbf6eabbae00219566fb02719ca2d6c480d`.
- Merges into: `fm/cli-current-foundations-postfix-fix-wave-r1` → `main` (human-gated); this follow-up remains a separate stacked PR until Firstmate supplies the final base and merge instruction.
- Delivery: A committed, non-force-pushed `fm/cli-reverse-etl-action-binding-foundation-r1` branch, ready for the Firstmate-directed no-mistakes PR/CI gate. No connector definition changes and no merge to `main`.
- Working branch: `fm/cli-reverse-etl-action-binding-foundation-r1`.
- Task: Extend the existing connector-neutral declarative typed-destination foundation with declaration-owned per-action source mappings plus closed, receipt-complete bounded batch and tombstone semantics. Preserve GitHub and all existing declarations without provider- or route-specific shared Go.
- Verification: behavioral red/green tests in schema, connector, engine, synctransport, App, and CLI; generator validation and surface check; race checks for changed packages; binary/list-plan-apply-resume persisted-path assertions; repository boundary/canon gates; scoped local gates and, where the command timeout permits, full verification. Record exact commands and outcomes in `VERIFICATION.md`.

### GSD execution record

`scripts/gsd doctor`, every required `scripts/gsd sources` lookup, all five
generated lifecycle prompts, and `go run ./cmd/agentcontractgen check` passed
on this branch before planning. This is the required inline/manual fallback:
the current runtime cannot host compatible isolated GSD workers and the
canonical single-worker contract forbids role spawning. The repo-local adapter
has no `programming-loop` command, so the mandatory lifecycle is executed as
`discuss-phase` → `plan-phase --tdd` → `execute-phase` → `verify-work` →
`code-review` with artifacts updated inline.

Required skills loaded: `golang-how-to`, `golang-cli`,
`golang-documentation`, `golang-design-patterns`,
`golang-structs-interfaces`, `golang-error-handling`, `golang-security`,
`golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`,
and `no-mistakes`.

### Frozen contract and TDD matrix

1. **Red — action-owned source bindings.** A synthetic destination with three
   selected actions over one source must use three different, exact input
   mappings; a fourth action in a second connector must remain independently
   selectable. A binding is keyed by its exact destination action and source
   executor/stream, has no action-independent fallback, and is sealed by the
   persisted connector/source/destination/action/mapping/mode/strategy/
   credential-revision/configuration-digest/evidence/workset/approval identity.
   Unknown, duplicate, malformed, stale, cross-action, cross-source, and
   cross-connector selections refuse before stage, transport, plan, write, or
   read-back I/O.
2. **Red — closed batch disposition.** Each selected action declares its
   bounded disposition. The adapter sends deterministic, ordered bounded units
   only, retains one stable idempotency identity per sealed record/action,
   returns every 2xx/4xx/5xx provider receipt, and preserves successful
   partial evidence without treating a partial workset as acknowledged. Retry,
   cancellation, duplicate records, and reordered input may not replay an
   already acknowledged provider write.
3. **Red — explicit tombstone routing.** A tombstone can select only an
   action the same declaration identifies as delete-capable. It maps the
   declaration-owned tombstone image/key to that delete action, never to a
   create/update payload; its acknowledgement and read-back are included in
   the same sealed workset evidence. A source that emits tombstones without
   this exact disposition refuses before provider I/O.
4. **Green — smallest neutral model.** Add only closed shared declarations,
   schema/loader validation, persisted-plan sealing, and typed adapter/orchestrator
   support required by the red tests. Keep existing source bindings backward
   compatible only when unambiguous and preserve GitHub's specialized adapter
   unchanged. Do not add a generic action, HTTP, route, method, body, mapping,
   or credential parameter.
5. **Refactor and integration.** Exercise connection listing, planning,
   preview/approval, applying, durable terminal persistence, and resuming via
   the real CLI/App path with synthetic bundles. Regenerate tracked artifacts
   once after green behaviour, then run generator, race, boundary, and
   verification gates. Update the declaration guide solely for the new shared
   foundation; connector adoption and provider definitions stay out of scope.

### Follow-up evidence table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Three actions sharing one source use distinct definition-owned mappings | fake | Synthetic bundles are required because production connector definitions are expressly out of scope. Tests assert each declared request body and action name differs exactly by selected binding, and a mismatched mapping produces zero transport calls. |
| A second connector cannot borrow a binding, evidence, or action | fake | A separate synthetic definition and recording provider prove only its declared action is planned/applied/read back; foreign selection leaves all I/O counters at zero. |
| Batch and retry semantics are bounded, deterministic, and receipt-complete | fake | Deterministic in-memory provider fixtures return ordered 2xx, then 4xx/5xx/cancellation. They record stable keys, exact attempted records, retained receipts, and no re-attempt after an acknowledged record. |
| Tombstones reach only declared delete actions | fake | Synthetic create/update/delete action schemas and recording provider prove tombstone data cannot materialize a non-delete request and undeclared cases make zero calls. |
| Persisted CLI/App lifecycle seals and resumes the route | fake | Real project/CLI tests create, list, plan, preview, approve, apply, and resume a saved connection; observable stored run/receipt/checkpoint state proves no caller-shaped route exists. |
| Existing declarations and GitHub behavior remain executable | live | Existing declaration/transport regression suites plus generator checks exercise the unchanged GitHub specialized contract; no production bundle is edited. |

### Commit checkpoints

1. This frozen plan, TDD ledger, and verification checklist (`Refs #4303`).
2. Red behavioral tests only (`Refs #4303`).
3. Green neutral implementation plus documentation and one artifact regeneration (`Refs #4303`).
4. Verification/review remediation (`Refs #4303`).
