# PLAN — Issue #3898: direct reads return the declared page and say so

## GSD execution record

**Manual-GSD fallback, recorded deliberately.** The GSD lifecycle commands
(`discuss-phase` → `plan-phase --tdd` → `execute-phase` → `verify-work`) were
not driven through `scripts/gsd` for this phase: the work began from a live
connector-validation report against a running defect rather than from a phase
prompt, and the fix had to ship as a release blocker alongside the warehouse
fix. AGENTS.md ("GSD Core Runtime For Agents") permits inline/manual execution
when the runtime cannot provide compatible isolated agents, provided the
fallback is recorded. This file is that record.

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
