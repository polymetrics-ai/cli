# PLAN — Issue #3898: direct reads return the declared page and say so

## GSD execution record

**Restarted manual-GSD fallback, recorded deliberately.** This recovery run
resolved the installed adapter and generated the `discuss-phase` and
`plan-phase --tdd` prompts through `scripts/gsd`; `doctor` and
`go run ./cmd/agentcontractgen check` passed. The canonical issue-agent
contract forbids spawning planner/reviewer roles for this one-worker issue, so
the generated prompts are executed inline and their decisions are recorded in
this phase directory. The remaining generated prompts (`execute-phase`,
`verify-work`, then `code-review`) are run in that order before handoff.

Required skills loaded for the restart: `golang-how-to`, `golang-cli`,
`golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`,
`golang-design-patterns`, `golang-structs-interfaces`,
`golang-documentation`, and `golang-lint`.

TDD was followed for real: every behaviour below was written as a failing test
first, and the red output is retained verbatim in TDD-LEDGER.md.

## Goal

A direct read must never report a completeness it cannot prove.

## Problem

`OperationDirectRead` and `DirectRead` issued exactly one HTTP request with no
page-size parameter, so the provider applied its own default page (GitHub's is
30). The result was returned with `status: 200`, exit 0, and no field capable of
saying more records existed — `DirectReadResult` held one `Status` and one
`Body`.

Measured against the live GitHub API: `code-scanning analyses view` returned 30
of 1466 rows (97.9% discarded); `hooks deliveries view` 30 of 931; `pulls files
view` 30 of 55.

## Scope, established before any fix

Fixture-backed runs of the real `pm` binary against github, gong and notion:
each issued exactly one request, none sent a page-size parameter, all returned
30 of 120 available records. The fault is in the shared executor, not GitHub —
362 implemented `direct_read` commands across 13 connectors. Confirmed
structurally: 0 of 1964 `rest_read` operations across all 551 bundles declare
any page-size parameter.

The restart independently repeated the execution proof against the current
binary and actual connector bundles, pointed at a local 120-record fixture:

- Gong's legacy `DirectRead` command (`pm gong logs list`) returned four
  cursor-addressable pages of 30; page one reported `complete: false` and the
  final page reported `complete: true`.
- Notion's `OperationDirectRead` command (`pm notion comment list`) returned
  the declared 100-row page followed by 20 rows through `--page-cursor`; it
  reported the same explicit incomplete/complete transition.

Those are real command-surface runs, not a structural inference. Together
with the retained red fixture evidence for GitHub, they establish a shared
executor release blocker rather than a GitHub-only defect.

## Design, as ruled by the project owner

A direct read is page-wise **exploration**, not bulk extraction — the ETL path
stores what it reads, a direct read does not. It therefore stays ONE request and
deliberately does **not** paginate to completion. An earlier
paginate-to-completion implementation was written and discarded on that ruling.

What changes: the page returned is the connector's **declared** page, and the
result carries the context needed to reach the next one.

## Constraints honoured

- Paging derives from each connector's already-declared pagination spec through
  the existing `engine/paginate.go` strategies. No second pagination
  implementation.
- **Zero files under `internal/connectors/defs/` change.**
  `connectorgen surface-sync --check` reports 0 fields changed across 551
  connectors, so the in-flight authoring sweep is not invalidated.
- No test was weakened, skipped or deleted.

## Tasks

1. Reproduce the truncation as failing tests asserting returned record counts.
2. Add `DirectReadPage` (strategy, records, size, number, has_more,
   next_number/next_cursor, complete, reason) to `DirectReadResult`.
3. Route both executors through one shared pager built on `newPaginator`.
4. Add `--page` / `--page-cursor`, declared once in
   `direct_read_page_flags.json` and rendered by all three renderers.
5. Regenerate the generated docs/website artifacts for parity.
6. Recovery validation: prove the GitHub command surface against observable
   data, including parameter validation and page context; then run ETL and
   reverse ETL only through `pm github` against one captain-authorized private
   disposable repository. Reverse ETL remains plan → preview → approval →
   execute. No write may target `polymetrics-ai/cli` or any other repository.
7. Record each live operation's outcome, returned record count, and any exact
   failure before the verify-work and code-review gates. The private test
   repository is deliberately retained for the captain to delete.
8. Captain-required live reverse validation exposed an internal reverse-plan
   consistency defect before dispatch: a plan limited to one row re-read one
   additional row during preview/run and then rejected its own unchanged
   source hash. Add a red regression that plans, previews, and runs a
   one-row slice from a multi-row fixture; preserve the source-drift rejection
   test, then make preview/run hash and dispatch exactly the planned slice.
9. The first actual GitHub `create_issue` execution displayed EOF against a
   seemingly malformed `.../issues%22` target. Reproduce the exact create-issue
   request through a local GitHub bundle fixture before any retry against the
   private repository. If the fixture isolates an implementation defect, add
   its red request-target assertion and correct it; if it passes, trace the
   exact argv, stored plan/config fields, request URL immediately before send,
   and error-redaction path before assigning a live transport cause.
10. Add a red safety regression for Go's quoted HTTP transport error form.
    `RedactErrorText` must remove URL query/fragment secrets without absorbing
    the closing quote delimiter into the URL and percent-encoding it as `%22`.
11. If the plain live `EOF` remains after that formatter correction, isolate it
    on the captain-authorized repository with equivalent curl and `gh` POST
    controls before changing transport behavior. Capture PM's method,
    Content-Length, body-byte count, and header names only; never emit a
    credential or header value. Treat a PM-only failure as a PM transport-path
    defect (including a VPN/proxy interaction specific to that path), not as a
    GitHub-wide outage.
