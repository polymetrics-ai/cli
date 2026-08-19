## Task Delivery Header

- Issue: Refs #4303 — feat(synctransport): compose connector-neutral typed reverse-ETL destinations.
- Base branch: main
- Merges into: main.
- Delivery: A pull request from `fm/cli-reverse-etl-destination-r1` to `main`, with local verification, live GitHub parity evidence, generated checks, and review complete.
- Working branch: fm/cli-reverse-etl-destination-r1
- Task: Replace the GitHub-only destination registration with a definition-selected, connector-neutral typed-action destination adapter. Prove two synthetic declarations compose through App and the orchestrator, retain GitHub issue-label behavior through the new path, fail closed before I/O, and publish the mechanical declaration contract.
- Verification: Targeted red/green tests in `internal/app`, `internal/connectors`, and `internal/synctransport`; the purpose-built real GitHub proof; `go test -timeout 20m` for affected packages plus `internal/cli`; `go vet ./...`; `go build ./cmd/pm`; all `make verify` gates including detached/polled `connector-boundary`; generated/snapshot checks; and an inline code review.

## GSD execution record

The required adapter was validated with `scripts/gsd doctor`, `scripts/gsd sources` for `discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and `code-review`, and generated prompts for the same commands (with `plan-phase --tdd`). `go run ./cmd/agentcontractgen check` passed. This direct-PR worker follows the lifecycle inline because compatible isolated workers are unavailable and the canonical single-worker contract forbids role spawning.

Required skills: `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`, and `golang-documentation`.

## TDD slices

1. **Red — generic selection:** Add a synthetic declarative API connector whose destination describes a typed named action, source field bindings, acknowledgment, and closed strategies. Show App composition rejects it because the destination factory accepts only `issue_label_destination`.
2. **Green — generic typed adapter:** Make production factory selection definition-derived for a new declarative typed-destination executor. Build the same typed adapter for every declaring connector, derive its contract only from that connector’s declaration and action metadata, and register it once.
3. **Red/Green — tenancy and refusals:** Add a second synthetic destination with distinct evidence and action metadata. Assert each gets its own evidence admission and action contract; malformed action contract, unknown executor/evidence, wrong source binding/role, and `change_capture` are refused before reads, plans, writes, or read-back.
4. **Parity/refactor:** Route GitHub’s issue-label contract through the same definition-evidence/factory loop while retaining its specialized typed adapter and provider-state read-back. Delete the bespoke composition branch only after the existing approval, recovery, and real-provider proof pass.
5. **Documentation/review:** Update the declaration guide with exact typed action and acknowledgement requirements, no-generic-writer boundary, mode/action strategies, source binding constraints, and evidence admission. Run full local verification and record review disposition.

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

- `make verify` passed locally, including the full Go suite, generated parity
  checks, documentation validation, lint, and connector-boundary.
- The separately detached-and-polled `make connector-boundary` gate completed
  cleanly: 552 connectors loaded, 293 files checked, no findings.
- The opt-in real API proof
  `TestPMBinaryExecutesLivePostgresWarehouseGitHubIssueLabels` passed with its
  dedicated proof repository. It independently confirmed add, set, keyed
  replay, destination read-back, acknowledgement, and checkpoint behavior.
- The required code-review stage was performed inline because compatible
  isolated reviewer agents are unavailable and the canonical single-worker
  contract forbids spawning them. Its clean disposition is in `REVIEW.md`.
