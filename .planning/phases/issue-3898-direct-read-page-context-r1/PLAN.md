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
- Cursor-surface regeneration changes only four derived `cli_surface.json`
  files: Amazon SQS (2 opaque tokens), Gong (8 cursors), Google Ads (1 page
  token), and Notion (6 start cursors). `connectorgen surface-sync --check`
  is green across all 551 connectors after regeneration; this intentional
  derived correction supersedes the earlier no-`defs/`-changes expectation.
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
12. Test the stale keep-alive replay hypothesis without weakening the
    non-idempotent-write contract: prove whether the JSON body reaches
    `http.NewRequest` as a concrete `*bytes.Reader`, then use a deterministic
    stale-idle write failure to distinguish ordinary Go replay from the strict
    no-replay mutation path. Do not restore `GetBody` for a non-idempotent
    action unless an idempotency contract makes replay safe.
13. Trace the live GitHub write with `net/http/httptrace` (connect, TLS,
    headers, request write, first response byte) using phase names and
    booleans only. Then send the same built `pm` binary through its normal
    GitHub reverse path to a local non-GitHub fixture; test TLS only through a
    trusted chain and never bypass certificate verification. Compare proxy
    variable presence as seen by curl and Go without revealing their values,
    and inspect VPN state read-only. Keep all no-replay safeguards unchanged.
14. Before attributing the remaining EOF to a tunnel, validate the credential
    path without logging secret material: `pm credentials test` must succeed
    against the private repository; the resolved configuration must select the
    bearer-token branch rather than GitHub App or public auth; and an internal
    hash-and-length-only round trip through the encrypted vault must preserve
    the `gh auth token` source with no leading, trailing, or newline
    whitespace. Confirm the authenticated GitHub identity has `push` access
    to the dedicated repository through a read-only permission response.
15. Complete the transport 2×2 without mutating repository state. Run an
    authenticated private-repository GET through the exact strict transport
    configuration (`noReplayClient`, no keep-alive, HTTP/1, one-use request)
    and run GitHub's non-mutating Markdown-render and read-only GraphQL POSTs
    through the ordinary transport. Also run those safe POSTs through strict
    transport so method, body, and `/issues` are not confounders. Do not alter
    the no-duplicate-write contract based on this diagnostic.
16. Add a RED ALPN invariant reproducing the strict transport's protocol
    mismatch with a local HTTP/2-capable TLS server. Keep the strict fresh
    connection/replay safeguards, but remove only the forced HTTP/1 setting so
    the cloned transport and TLS ALPN negotiation agree. Green the existing
    no-replay tests, the new ALPN test, and the live non-mutating matrix before
    any future captain-authorized mutation retest.
17. Captain completion validation found legacy direct-read cursor flags still
    present in generated command surfaces (for example Notion's
    `--start-cursor`), despite the documented single-navigation contract.
    Add a red `surface-sync` regression that preserves a page-size override
    but removes an opaque cursor override, then make the generator enforce
    that contract for both operation-backed and legacy direct-read commands
    and regenerate only the affected derived surfaces. Record the resulting
    live/help evidence before declaring the parameter proof complete.
18. Re-run the captain-authorized GitHub acceptance path through `pm` only:
    create an issue through reverse ETL, comment on it, and delete a file
    (creating the disposable deletion target through the same approved PM
    path when the private repository is empty). Independently verify each
    state transition with read-only `gh-axi`; then prove the ETL read-back and
    parameter/page contract against real returned records and local
    server-observed fixtures.
19. Audit the GitHub connector's declared local/binary surface without adding
    a generic shell or raw-HTTP escape hatch. Report clone support honestly and
    exercise each existing read-only file, release-asset, and archive path;
    do not extend the product surface without a captain decision.
20. Post-`main` refresh compatibility gap: the two branch-owned reverse-plan
    regressions added before the Parquet warehouse transition still hand-write
    root-level JSONL tables. Record their CI-reproduced RED failure, replace
    only those test fixtures with the existing real Parquet fixture helper,
    and prove the exact tests plus `internal/app` are green. This is test
    fixture alignment, not a change to the legacy-JSONL refusal contract.
