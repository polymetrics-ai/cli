# Issue #4288 — certify batch-1 implemented capability cells

## Task Delivery Header

- Issue: Refs #4288 — feat(certification): certify already-implemented batch-1 capability cells
- Base branch: main
- Merges into: main
- Delivery: A direct PR from `fm/cli-cert-batch1-live-r1` is open against `main`, the API-reported base is `main`, and every accepted evidence record plus its required local verification is committed.
- Working branch: fm/cli-cert-batch1-live-r1
- Task: Add Jira, Asana, and Notion to the existing reviewable certification scope; generate their certification artifacts; provision one secret-safe, free credential per connector; certify only pre-existing implemented capability cells and record each concrete non-pass; publish the normal accepted evidence with no connector-definition or engine changes.
- Verification: Targeted `connectorgen` unit tests; scoped and final certification matrix/sweep/candidate checks; the repository certification harness; per-record matrix checks; generated-file checks; targeted CLI/build checks; all applicable repository gates, including detached `connector-boundary`; `git diff --check`; and API read-back of the opened PR base.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| The certification path accepts precisely the three issue targets | live | Before the change each `certification-matrix --connector <target> --check` invocation refuses the connector; after it, source-derived shard generation and validation accepts that target. |
| Generated scope remains truthful | live | Regenerated matrix/status artifacts name the three targets and are current under `certification-matrix --check`; no connector definition is changed. |
| Existing implemented cells receive accepted evidence only after a real response proves their declared assertion | live | The normal harness persists one accepted secret-free record and its immediate matrix check succeeds; zero-provider-result executions cannot create an accepted record. |
| Cells that cannot be certified remain explicit | live | The per-connector sanitized receipt lists each non-pass capability with its concrete foundation, fixture, provider, entitlement, or product-defect reason. |
| Secrets and third-party data stay out of the repository | live | The normal evidence scanner accepts written artifacts and review confirms only classifications, counts, and salted fingerprints are committed. |

## Scope boundary

Production files allowed in this issue are the certification-scope allowlist and its generator-owned
artifacts/evidence for Jira, Asana, and Notion. No `internal/connectors/defs/{jira,asana,notion}/**`
file, shared engine/runtime code, source declaration, or one of the seven connectors owned by the
parallel declaration sweep may be changed. If a result needs such a change, record the exact
response and leave that cell uncertified.

## Lifecycle and inline fallback

The required GSD adapter checks and generated prompts are being executed inline: `discuss-phase`,
`plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review`. The direct-PR brief and
canonical single-worker contract prohibit spawning compatible GSD lifecycle agents, so this is an
explicit manual fallback—not a lifecycle exemption.

Required skills loaded: `golang-how-to`, `golang-cli`, `golang-testing`,
`golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, and
`golang-structs-interfaces`. The CLI help/manual/website parity reference was reviewed; it is not
applicable because this changes no `pm` command, flag, output contract, connector command surface,
manual, or website document.

## TDD slices and checkpoints

1. **Certification scope (red → green).**
   - Red: `go run ./cmd/connectorgen certification-matrix --connector jira --check`, and the same
     for Asana and Notion, each refuse an unallowlisted connector.
   - Green: add only these three names to the reviewable allowlist, run `certification-matrix --all`,
     and prove scoped and full checks succeed.
   - Commit/push checkpoint: plan and green generator scope.
2. **Capability inventory and credentials.**
   - Derive each connector's pre-existing cells from its normal certification matrix/sweep/candidates;
     record no new connector declaration.
   - Create the connector mailbox and live credential strictly under provisioning runbook §§0–16.
     Stop on a defined human gate, rather than weaken the claim.
3. **Serial live certification.**
   - Run the normal harness against one connector/cell at a time in issue order. Persist an accepted
     record only after the declared observable assertion and immediate matrix validation; write
     sanitized receipts for every failure or missing foundation.
   - Reads first; mutations require a scratch resource created by this run plus an independent
     cleanup read-back. Never run reverse-ETL without its plan/preview/approval/execute contract.
4. **Verification and review.**
   - Regenerate/check certification artifacts, run scoped repository gates and review the diff for
     secret or ownership leakage. Rebase onto current `main` immediately before each push. Open the
     direct PR, confirm its base with the GitHub API, and use the documented automated-review route.

## Safety and reporting

Credential values, verification codes, account identifiers, raw provider responses, and personal data
must never enter an argument vector, log, evidence record, commit, or PR body. Receipts report only
the operation/cell, HTTP/status classification, safe provider reason, and redacted/salted response
fingerprint. Every status line uses `Refs #4288`; completion reports certified cells, non-certified
cells with exact reasons, and the responsible follow-up where one exists.
