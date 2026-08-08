# Plan — Zoom Virtual Agent documented-operation parity, R1

## Delivery record

- Parent: [#3915](https://github.com/polymetrics-ai/cli/issues/3915); provider-owned slice:
  [#3941](https://github.com/polymetrics-ai/cli/issues/3941).
- Scope: Zoom's published **Virtual Agent** artifact only: all thirteen documented actions, exact
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

The provider's own Virtual Agent artifact is a bounded, coherent category. It contains the
Knowledge Management article and sync endpoints plus the provider-tagged Report endpoints, all
under one published source document. This introduces no invented resource category and is
independent of the landed SCIM2 slice except for the shared declarative executor.

## Live artifact audit — completed before RED

The source was fetched afresh instead of trusting the inherited ledger.

| Item | Evidence |
| --- | --- |
| API URL | `https://developers.zoom.us/docs/api/virtual-agent.md` |
| Retrieval | `2026-08-08T14:20:14Z` |
| HTTP / bytes | `200` / `60,147` |
| SHA-256 | `5be404cc4cbcf03736914f52ad9e50dc4a17ebfbc104db9e20bf7d31a1fb6436` |
| Artifact | OpenAPI `3.1.1`, API server `https://api.zoom.us/v2` |
| Ledger delta | `0` — exactly thirteen `provider_module=virtual-agent` rows match method, path, title, and source URL |

The live artifact contains exactly these thirteen operations:

| Method | Path | Provider title |
| --- | --- | --- |
| GET | `/km/kbs/{kbId}/articles` | Get articles |
| POST | `/km/kbs/{kbId}/articles` | Create article |
| GET | `/km/kbs/{kbId}/articles/{articleId}` | Get article |
| PUT | `/km/kbs/{kbId}/articles/{articleId}` | Update article |
| DELETE | `/km/kbs/{kbId}/articles/{articleId}` | Delete article |
| POST | `/km/kbs/{kbId}/sync` | Create sync request |
| GET | `/km/kbs/{kbId}/sync/{syncId}` | Get sync |
| GET | `/virtual_agent/report/engagements` | Get ZVA engagements |
| GET | `/virtual_agent/report/engagements/query_details` | Get ZVA query details |
| GET | `/virtual_agent/report/engagements/variables` | Get ZVA variable details |
| GET | `/virtual_agent/report/surveys` | Get ZVA Surveys |
| GET | `/virtual_agent/report/transcripts` | Get ZVA transcripts |
| GET | `/ai_studio/reports/operation_logs` | List AI Management operation logs |

The article create/update request schemas require `content`, `exclude`, and `title`; optional
declared fields are `category`, `external_id`, `language`, and `url`. `Create sync request` has no
request body. Delete returns documented `204 No Content` and must assert status only. The source
shows paging values only in response schemas; it declares no request parameter section for these
operations, so no `page`, `per_page`, `limit`, token, or guessed date flag is hand-authored.

## Locked decisions

- Implement all thirteen operations: nine bounded `rest_read` / `direct_read` commands and four
  `rest_write` / `direct_write` commands. There are no `unsafe_or_disallowed` rows and no duplicate
  exclusions.
- Command paths mirror the provider category and resource terminology: `virtual-agent
  knowledge-bases articles …`, `virtual-agent knowledge-bases sync …`, and `virtual-agent reports
  …`. The exact fixed endpoints remain declaration-owned.
- Article inputs are explicit documented typed fields; no generic JSON/HTTP input is added. The
  sync action has no body. Every mutation keeps plan → no-network preview → single-use approval →
  execute; the DELETE action additionally requires typed destructive confirmation and emits no
  invented response body.
- Use the normal `/v2` Zoom bearer transport. Sensitive article, knowledge-base, engagement,
  transcript, survey, variable, operator, and report fields are redacted in previews, errors, and
  output. Fixtures are synthetic; no credential or token-derived value is printed or recorded.
- No foundation is anticipated. If a declared contract exposes a missing engine capability, stop
  connector authoring, add the narrow reusable foundation in its own RED/GREEN commits, and state
  which connectors it unblocks.
- Run `connectorgen surface-sync`; never hand-author derived command metadata. Generate docs/site
  data repository-wide, restore unrelated generated output, and retain only Zoom-specific output.

## TDD execution

1. **Plan checkpoint** — commit this source audit, manual-GSD fallback, target accounting, and
   verification plan before test or production changes.
2. **RED checkpoint** — add only Zoom command-surface tests first; run and capture their actual
   failure before changing connector declarations. They must require `38 → 51` executable rows,
   `1,804 → 1,791` Zoom-local rows, `21 → 30` direct reads, and `12 → 16` direct writes, and prove
   all thirteen exact paths are currently unknown to the real commandrunner.
3. **GREEN connector** — declare typed commands/operations, schemas, redaction, fixtures,
   metadata, docs intent, and reconcile only Virtual Agent rows to direct read/write coverage.
4. **Verify/review** — run fixture lifecycle tests, surface/ledger checks, generated docs/site
   checks, fresh binary base/group/every-command help, and scoped review. Confirm ledger change is
   confined to the thirteen Virtual Agent endpoints.

## Target accounting

| Measure | Before | After |
| --- | ---: | ---: |
| Zoom executable operations | 38 | 51 |
| Zoom-local implementable rows | 1,804 | 1,791 |
| Direct reads | 21 | 30 |
| Direct writes | 12 | 16 |
| Reverse-ETL writes | 2 | 2 |
| `unsafe_or_disallowed` Zoom rows | 0 | 0 |

## Verification plan

- RED/GREEN commandrunner preflight plus fixture lifecycle tests for all thirteen actions.
- Exact method/path/body/auth/status assertions, no-body POST/DELETE semantics, destructive
  confirmation, response redaction, no invented paging inputs, and no leaked fixture-sensitive
  strings.
- `go run ./cmd/connectorgen surface-sync --check`, full connector validation, and scoped
  `surface-reconcile --check --notes-contains provider_module=virtual-agent`.
- Fresh binary `pm help zoom`, bare `pm zoom`, bare `pm zoom virtual-agent`, and each exact command
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
- `https://developers.zoom.us/docs/api/virtual-agent.md`
