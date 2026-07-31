# Help Scout parity wave02 r1 — GSD plan

## Scope

Parent issue: #212 (`help-scout`) with subissues #213-#219. Branch: `fm/cli-help-scout-parity-wave02-r1`.

Allowed write scope:

- `internal/connectors/defs/help-scout/**`
- Help Scout-owned tests/fixtures/CLI metadata/docs
- `.planning/phases/help-scout-parity-wave02-r1/**`
- GitHub issue bodies only for an idempotent captain-policy addendum via `gh-axi`

Out of scope/hard gates: no live Help Scout calls, no credentials, no new dependencies, no shared runtime/foundation edits, no generic raw HTTP/query/write escape hatches, no certification claims, no push/PR/no-mistakes until firstmate resumes.

## GSD command path

- Ran `scripts/gsd doctor`: pass.
- Ran `scripts/gsd list`: pass; adapter reports 69 commands.
- Attempted required `scripts/gsd prompt programming-loop init --phase issue-212-help-scout-parity --dry-run`: adapter returned `unknown GSD command: programming-loop` even though project docs name the command. Manual GSD fallback is active and recorded here per `.agents/agentic-delivery/references/gsd-pi-adapter.md`.

## Required skills loaded

- GSD/core: `gsd-core`, required-skills routing, GSD Pi adapter reference, universal runtime loop.
- Go/connector/CLI/docs: `golang-how-to`, `golang-cli`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-documentation`.
- Project references: issue-agent contract, parent-orchestrator contract/workflow, stacked PR workflow, automated review routing, Claude review loop, CLI help/docs/website parity reference, connector migration handoff, connector conventions, connector architecture v2 design.

## Source inventory

Official Help Scout source IDs from parent: `help_scout_llms`, `help_scout_mailbox_root`, `help_scout_mailbox_endpoint_pages`.

Local source artifacts:

- `.planning/phases/help-scout-parity-wave02-r1/github-issues/*.view.txt`: parent/subissue graph captured with `gh-axi`.
- `.planning/phases/help-scout-parity-wave02-r1/help-scout-official-scrape.json`: endpoint-page scrape from `https://developer.helpscout.com/mailbox-api/`.

Scrape reconciliation:

- Root endpoint nav links: 146.
- Parsed pages: 146.
- Raw unique method/path operations: 145.
- Canonical normalized operations: 144, matching parent #212 official total.
- Audit lane split from issue metadata: ETL read 43, reverse ETL write 60, direct/report read 32, binary/file 4, webhook/CDC-changefeed 5, excluded 0.
- Connector-local execution policy for this slice: make fixed-path webhook configuration reads/writes and attachment upload/delete executable through streams/reverse ETL safety gates without claiming CDC/binary-download execution. Final bundle targets stream=45, write=65, blocked operation=34.

## Implementation slices

1. **Red/contract tests** — add Help Scout-owned validation test asserting operation counts, command metadata safety, destructive confirmation, and foundation-blocked surfaces.
2. **Ledger generation** — rewrite `api_surface.json` with `operation_ledger_version: 1` and exactly 144 canonical rows.
3. **Read streams** — expand `streams.json`/schemas from 4 to 45 stream-backed GET operations, including fixed-path webhook configuration reads but not claiming CDC event consumption. Use bounded page-number pagination, optional config fan-out IDs for parameterized resources, and passthrough schemas where official response shape is large.
4. **Reverse writes** — add `writes.json` with 65 typed actions for every documented mutation, including attachment upload/delete and webhook admin create/update/delete. DELETE/destructive/admin actions use `confirm: "destructive"`, `body_type: "none"` where applicable, idempotent 404 notes, path field redaction, and reverse ETL risk text.
5. **Planned operations/CLI metadata** — add `operations.json` and `cli_surface.json` for blocked report/direct query reads and binary attachment downloads without claiming unsupported execution. Implemented ETL/reverse commands map only to streams/writes.
6. **Fixtures/docs** — keep sanitized stream fixtures, update check fixture for `/v2/mailboxes`, add write request-shape fixtures where practical, and update `docs.md` with safety/evidence/known limits.
7. **Issue addendum** — append idempotent captain-policy addendum to #212-#219 bodies with `gh-axi issue edit --body-file`, preserving existing text and count tables.
8. **Verification/commit** — run focused validation gates, update TDD/verification artifacts, commit the implementation, append `done:` status, stop.

## Shared-foundation dependencies to record, not edit

- #2985 provider search/query boundary: report/direct provider commands remain blocked/planned unless already supported by existing generic direct-read operation executor.
- #2986/#2988 CDC truth/state/lab: Help Scout webhooks are represented as blocked changefeed setup operations; no `metadata.capabilities.cdc=true` and no certification claim.
- Binary/file transfer execution for Help Scout attachments remains blocked/planned unless covered by existing bounded binary foundation; no raw file/body escape hatch.

## Orchestration decision log

- cycle=plan decision=local_critical_path reason=single disposable worker owns one connector directory; mutating subagent fan-out would collide inside the same Help Scout bundle and issue bodies are coordinator-owned.
