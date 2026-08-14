# TDD Ledger: Issue #3990

## Planned red/green evidence

| Slice | Red | Green | Refactor/verification |
| --- | --- | --- | --- |
| GraphQL policy and observation | A GraphQL request has no matched GitHub budget and the response metadata is output-only. | Query and mutation admissions are policy-selected before send; a parsed body observation controls the next request. | Run focused engine tests and assert independent scopes still send. |
| Shared certification selection | Certification has no required-shared GitHub route and absence of the coordinator is not a certification pre-send refusal. | Missing coordinator produces typed refusal and zero sends; selected scope identity is opaque. | Run focused certification and coordination tests. |
| Multi-process budget | Two workers with isolated registries can both send against a capacity-one budget. | Shared workers coordinate: first sends once, second waits/refuses, and second sends zero times. | Run opt-in integration test with a deterministic local coordinator. |
| Deadline/events/ledger | A queued wait has no structured not-sent or deadline event. | Certification reports attempts, waits/resets and not-sent deadline cutoff, with complete cleanup ledger state. | Verify report JSON carries no raw scope or credential material. |

## Actual evidence

### 2026-08-14 — planning checkpoint

- Red: pending implementation. The authoritative #3990 audit records GraphQL as explicitly excluded from all GitHub policies and response metadata as output-only.
- Green: pending implementation.
- Manual GSD fallback: performed inline because isolated GSD worker worktrees are unavailable and this task prohibits role spawning; all required lifecycle prompts were resolved first.
