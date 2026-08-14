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

### 2026-08-14 — GraphQL declaration red test

- Red: `go test -timeout 20m ./internal/connectors/engine -run '^TestGitHubDeclaredRateLimits$'` failed because `graphql-authenticated-user`, `graphql-app-installation`, and `graphql-actions-token` were absent from GitHub's declaration.
- Observable gap: GitHub GraphQL traffic selected no provider-cited point budget, so returned `cost`, `remaining`, and `resetAt` could not affect a later admission.
- Green: pending the declaration plus typed response-observation implementation.

### 2026-08-14 — GraphQL admission and shared-certification green tests

- Green: `TestOperationGraphQLRateLimitResponseTightensNextAdmission` proves parsed GraphQL `cost`, `remaining`, and absolute `resetAt` control the next pre-send admission; each refused second request records zero provider sends. `TestOperationGraphQLRateLimitCostDoesNotDoubleChargeSecondaryBudget` proves the response cost is applied only to the GraphQL primary family.
- Green, independent families: `TestOperationGraphQLRateLimitPrimaryRemainingDoesNotBlockSecondaryBudget` proves a GraphQL primary `remaining=0` plus `resetAt` does not make the separately declared secondary budget wait. The second request reaches the provider, so no primary-body observation is silently applied to the wrong resource family.
- Green: `TestOperationDirectWriteUsesSharedApprovalForFixedGraphQLMutation` now makes a second separately approved mutation fail at rate admission with zero additional provider sends, proving mutation wiring as well as query wiring.
- Green: `TestGitHubCertificationRatePoliciesFailClosedBeforeSend` proves certification-tier account, installation, repository, IP, and GraphQL policies are `require_shared` and refuse before send when no coordinator is configured.
- Green, live coordinator: `POLYMETRICS_COORDINATION_INTEGRATION=1 go test -tags coordinationintegration -timeout 20m -v ./internal/connectors/engine -run '^TestGitHubCertificationSharedBudgetCoordinatesSeparateProcesses$'` passed. It launches two separate test processes against a capacity-one GitHub certification policy: the first sent once, the second waited until its deadline, and the server observed exactly one send.

### 2026-08-14 — admission-deadline and event green tests

- Red, real binary: the permitted public GitHub certification sweep, run with a local shared coordinator and no write flag, waited at rate admission without producing a deadline-bounded report. That exposed an in-scope runner gap: its admission context was unbounded.
- Green: `TestRequesterRateLimitEventsRecordDeadlineCutoffBeforeSend` gives a real requester a bounded admission context. It records `wait` then `not_sent` with `deadline_cutoff`, while the provider double records **zero** sends.
- Green: `TestRequesterRateLimitEventsRecordAttemptAndProviderReset` records an actual pre-send attempt and response reset observation. `TestCertificationRateLimitEventsCarryStage` is the focused report-projection double needed to prove event stage attribution; it asserts the serialized report carries no raw scope or credential material.
- Green: `TestRuntimeAppliesProjectRateLimitAdmissionTimeout` proves the certification runner's per-project configuration reaches the real engine requester. The runner defaults each admission wait to 30 seconds without imposing a whole-run timeout.
- Scope resolution: #4136 owns the invalid `sample` sweep target. The certification allowlist permits GitHub and PostgreSQL; this issue's live evidence and real-binary sweep work use GitHub only. A credentialed GitHub write/cleanup sweep remains an operator-authorized follow-up evidence run because this task has neither a sandbox credential nor mutation authorization.

### 2026-08-15 — inspection regression after integration rebase

- Red: protected PR Verify run `31830388406` failed `TestConnectorsInspectLabelsProcessLocalRateLimitProtection`: GitHub's certification-only `require_shared` policies made unconfigured/default inspection report `mixed`, hiding the actual process-local protection for ordinary traffic.
- Green: a certification-only shared overlay now preserves `process_local` inspection while the safe message explicitly states that certification refuses before send without shared coordination. `TestCertificationOnlyRequireSharedPolicyKeepsDefaultInspectionProcessLocal` proves the distinction; the CLI test verifies both JSON and human output disclose it without protected scope material.
