# Plan — Zoom Auto Dialer documented-operation parity, R1

## Delivery record

- Parent: [#3915](https://github.com/polymetrics-ai/cli/issues/3915); provider-owned slice:
  [#3940](https://github.com/polymetrics-ai/cli/issues/3940).
- Scope: Zoom's published **Auto Dialer** artifact only: all sixteen documented actions, exact
  typed CLI paths, fixture-backed lifecycle checks, generated Zoom docs/site output, endpoint-ledger
  reconciliation, and this phase evidence.
- Required skills carried by the parent delivery: `golang-how-to`, `golang-cli`,
  `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`,
  `golang-design-patterns`, `golang-structs-interfaces`, and `golang-documentation`.
- GSD provenance: `scripts/gsd sources` resolved `discuss-phase`, `plan-phase`,
  `execute-phase`, `verify-work`, and `code-review` on 2026-08-08. The provider-category phase is
  not registered by the official runtime and the parent contract forbids role spawning, so this is
  the documented inline manual-GSD fallback with explicit discussion, plan, RED, GREEN,
  verification, and review evidence.

## Inline discuss-phase record

Zoom's own Auto Dialer artifact is a coherent provider category. Its Call History, Call List
Management, Prospect Management, and report-tagged endpoints belong to the one published Auto
Dialer document; they are not an invented delivery category. The category is independent of landed
Virtual Agent except for the shared declarative executor and existing narrow named-root-object body
foundation.

## Live artifact audit — completed before RED

The source was fetched afresh instead of trusting the inherited ledger.

| Item | Evidence |
| --- | --- |
| API URL | `https://developers.zoom.us/docs/api/auto-dialer.md` |
| Retrieval | `2026-08-08T14:45:08Z` |
| HTTP / bytes | `200` / `80,801` |
| SHA-256 | `2ca270a6dc2ac5bb72cf1ce7e6684785d5df21285affaf272c46bf8fbf127f61` |
| Artifact | OpenAPI `3.1.1`, API server `https://api.zoom.us/v2` |
| Ledger delta | `0` — exactly sixteen `provider_module=auto-dialer` rows already match method, path, title, and source URL |

The live artifact contains exactly these sixteen operations:

| Method | Path | Provider title |
| --- | --- | --- |
| GET | `/dialer/call-histories/{callHistoryId}` | Get call history by ID |
| GET | `/dialer/call-history` | Get Call History |
| GET | `/dialer/reports/call-history` | Get Report Call History |
| GET | `/dialer/reports/seller-productivity` | Get Seller Productivity Report |
| GET | `/dialer/call-lists` | List Call Lists |
| POST | `/dialer/call-lists` | Create Call List |
| GET | `/dialer/call-lists/{callListId}` | Get Call List by ID |
| DELETE | `/dialer/call-lists/{callListId}` | Delete Call List |
| PATCH | `/dialer/call-lists/{callListId}` | Update Call List |
| GET | `/dialer/call-lists/{callListId}/prospects` | List All Prospects in Call List |
| POST | `/dialer/call-lists/{callListId}/prospects` | Create Prospect |
| PATCH | `/dialer/call-lists/{callListId}/prospects` | Update Prospects batch |
| POST | `/dialer/call-lists/{callListId}/prospects/batch` | Create Prospects batch |
| DELETE | `/dialer/call-lists/{callListId}/prospects/{prospectId}` | Delete Prospect |
| PATCH | `/dialer/call-lists/{callListId}/prospects/{prospectId}` | Update Prospect |
| GET | `/dialer/prospects/{prospectId}` | Get Prospect by ID |

The published document gives article-style response paging fields only in response schemas and no
request-parameter section for these operations. No `page`, `per_page`, `limit`, token, date, or
guessed query flag will be hand-authored. The documents state `204 No Content` for both deletes,
call-list update, and single-prospect update; those actions must assert status only. Complex
prospect and batch request schemas stay fixed-operation typed JSON-object bodies under the existing
narrow named-root-object foundation; this is not a generic raw-body feature.

## Locked decisions

- Implement all sixteen operations: eight bounded `rest_read` / `direct_read` commands and eight
  `rest_write` / `direct_write` commands. There are no `unsafe_or_disallowed` rows and no duplicate
  exclusions.
- Command paths mirror the provider category/resource language: `auto-dialer call-histories`,
  `auto-dialer call-history`, `auto-dialer reports`, `auto-dialer call-lists`, and
  `auto-dialer prospects`.
- Call-list creation uses four documented typed scalar fields. Update and prospect/batch payloads
  use named JSON-object flags bound to one declared operation schema, including all documented
  nested prospect fields and enum constraints. No generic JSON/HTTP input is introduced.
- Every mutation keeps plan → no-network preview → single-use approval → execute. Both DELETE
  actions also require typed destructive confirmation. All documented 204 responses have
  `output_policy: none` and status-only assertions.
- Use the normal `/v2` Zoom bearer transport. Call history, call list, prospect, contact, company,
  communication, CRM, transcript, report, and generic token values are redacted in previews,
  errors, and output. Fixtures are synthetic; no credential or token-derived value is printed or
  recorded.
- No foundation is anticipated: named root JSON-object input and literal member redaction already
  shipped in the SCIM2 slice. If a remaining declared contract exposes a missing engine capability,
  stop connector authoring, add the narrow reusable foundation in its own RED/GREEN commits, and
  state which connectors it unblocks.
- Run `connectorgen surface-sync`; never hand-author derived command metadata. Generate docs/site
  data repository-wide, restore unrelated generated output, and retain only Zoom-specific output.

## TDD execution

1. **Plan checkpoint** — commit this source audit, manual-GSD fallback, root-object reuse decision,
   target accounting, and verification plan before test or production changes.
2. **RED checkpoint** — add only Zoom command-surface tests first; run and capture their actual
   failure before changing connector declarations. They must require `51 → 67` executable rows,
   `1,791 → 1,775` Zoom-local rows, `30 → 38` direct reads, and `16 → 24` direct writes, and prove
   all sixteen exact paths are currently unknown to the real commandrunner.
3. **GREEN connector** — declare typed commands/operations, complete schemas, redaction, fixtures,
   metadata, docs intent, and reconcile only Auto Dialer rows to direct read/write coverage.
4. **Verify/review** — run fixture lifecycle tests, surface/ledger checks, generated docs/site
   checks, fresh binary base/group/every-command help, and scoped review. Confirm ledger change is
   confined to the sixteen Auto Dialer endpoints.

## Target accounting

| Measure | Before | After |
| --- | ---: | ---: |
| Zoom executable operations | 51 | 67 |
| Zoom-local implementable rows | 1,791 | 1,775 |
| Direct reads | 30 | 38 |
| Direct writes | 16 | 24 |
| Reverse-ETL writes | 2 | 2 |
| `unsafe_or_disallowed` Zoom rows | 0 | 0 |

## Verification plan

- RED/GREEN commandrunner preflight plus fixture lifecycle tests for all sixteen actions.
- Exact method/path/body/auth/status assertions, no-body DELETE semantics, destructive
  confirmation, response redaction, no invented paging inputs, and no leaked fixture-sensitive
  strings.
- `go run ./cmd/connectorgen surface-sync --check`, full connector validation, and scoped
  `surface-reconcile --check --notes-contains provider_module=auto-dialer`.
- Fresh binary `pm help zoom`, bare `pm zoom`, bare `pm zoom auto-dialer`, and each exact command
  `--help`.
- Scoped CI-equivalent vet/lint/docs/website/CLI/contract/surface/boundary/release gates from
  `AGENTS.md`; the full suite remains CI-owned.

## Canonical references

- `AGENTS.md`
- `docs/migration/HANDOFF-CODEX.md`
- `docs/migration/conventions.md`
- `docs/architecture/connector-architecture-v2-design.md`
- `.agents/agentic-delivery/contracts/issue-agent-contract.md`
- `.agents/agentic-delivery/contracts/parent-orchestrator-contract.md`
- `.agents/agentic-delivery/references/gsd-pi-adapter.md`
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`
- `https://developers.zoom.us/docs/api/auto-dialer.md`
