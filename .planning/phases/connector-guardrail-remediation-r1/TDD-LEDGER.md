# TDD LEDGER — connector-guardrail-remediation-r1

## Global TDD rule

Behavior-changing sub-issues must capture red test or validation evidence before production edits. Documentation-only guidance changes may use a docs-only exemption only when no executable rule exists; prefer grep/schema/validator tests.

## Planned red/green evidence

| Slice | Red evidence before production edit | Green evidence | Status |
| --- | --- | --- | --- |
| 1a icon registry foundation | Failing tests/proofs for legacy prefixed icon mappings, website-only overrides/assets, ambiguous source/destination collapses, and ownership rejection of canonical icon assets | canonical bare registry, docs-icon asset authority, website generated copies, and consumer tests pass under #3595 | Draft sub-PR #3596 open |
| 1 target-scope core validator | Failing tests for exactly-one slug, auto-detected connector scope, shared runtime rejection, unrelated connector rejection, unrelated generated rejection, label omission bypass | `go test ./internal/connectors/boundary ./cmd/connectorgen` plus CLI fixture runs | Blocked on #3595 before #3590 reconciliation |
| 2 Actions/local/remote gate | Failing workflow/hook invocation test or guard command fixture showing CI non-green on violation | workflow YAML lint/grep, local hook invokes same command, GitHub required-check read-back | Planned |
| 3 PM/no-mistakes integration | Failing docs/config/grep test or docs-only exemption | guidance files updated; grep proves connector lane says stop and split foundation PR | Provisionally integrated via #3588; parent final review/gate pending |
| 4 HubSpot/Bitbucket remediation | Failing regression/ledger check for audited shared path dispositions | focused tests for preserved foundations; ledger rows complete | Planned |
| 5 Stripe/Freshchat/Google Ads shared remediation | Failing engine/runner/connectorgen regression for preserved or corrected shared behavior | focused package tests; ledger rows complete | Planned |
| 6 Zendesk/Google Ads generated remediation | Failing guard fixture for unrelated connector docs/generated churn | generator/guard tests; stale unrelated outputs removed or justified | Planned |
| 7 audit ledger/proof | Failing ledger schema/completeness or proof fixture | all eight dispositions and end-to-end guard proof pass | Planned |

## Actual evidence log

- 2026-08-02: GSD adapter healthy (`scripts/gsd doctor`).
- 2026-08-02: `scripts/gsd prompt programming-loop init --phase connector-guardrail-remediation-r1 --dry-run` unavailable (`unknown GSD command: programming-loop`); fallback recorded in PLAN. Use available GSD `execute-phase` prompt plus manual universal loop.
- 2026-08-02: #3583 / #3588 red/green evidence lives in `.planning/phases/connector-guardrail-remediation-r1/workers/issue-3583/TDD-LEDGER.md`; no-mistakes run `01KZ0SEAKBB9TG7N3SMG97XKJS` passed at current head `0c321595d7ae4852550a5012a895c3e11f7e8298`.
- 2026-08-02: #3588 integrated into parent branch as `86b91fc40f46b8653538531fc40c183913676f05`; parent ledger restoration/update is docs/state-only and verified with JSON parse plus diff/status checks before push.
- 2026-08-02: #3595 planning scaffold created before production edits; captain addendum requires canonical SVG assets under `docs/connectors/icons/**` and website public icons generated/copied from that tree.
