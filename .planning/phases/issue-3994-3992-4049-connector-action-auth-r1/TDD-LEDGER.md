# TDD ledger

The initial RED checkpoint (`95e2d60b1`) tested a per-fire authorization grant. The captain's
corrected scope forbids that design. Its compile failures remain historical evidence only; the
obsolete R2-R4 tests and implementation are removed and replaced by the corrected RED contract.

| ID | Guarantee | RED command/evidence | GREEN proof | Status |
| --- | --- | --- | --- | --- |
| R1 | prepared execution has a distinct payload-bound identity | original focused compile failed on absent prepare/execute identity; corrected tests retain the identity requirement without a grant | `TestAuthorizedFlowActionPreparedIdentityBindsPayloadAndReachesReceipt` binds resolved manifest digest, payload, firing, scope, destination, preview, and returns/persists the same `pex_` identity | green |
| R2 | flow creation accepts only existing jobs | `traces/corrected-red-run.txt`: absent `flow.JobReferenceError` / missing/malformed reason constants; tests assert no flow file | missing, malformed, unrecognised, and unapproved job tests receive `*flow.JobReferenceError`; `cli.Run` emits `flow_job_reference_refused`; no flow/provider write | green |
| R3 | action jobs are already approved and revalidated | corrected tests require an approved reverse plan; unapproved/revoked/expired/drifted plans return typed errors with zero writes/files/checkpoints | approved reverse-plan job is hydrated on every create/run; scope, credential revision/configuration, mapping, action, confirmation, expiry, and revocation tests refuse before dispatch and park scheduled fires | green |
| R4 | schedules accept only an existing valid flow | `traces/corrected-red-run.txt`: absent `schedule.FlowReferenceError` / missing reason; tests assert no schedule/crontab/sentinel write | create/install missing/malformed/invalid/ambiguous flow tests receive `*schedule.FlowReferenceError`; `cli.Run` emits `schedule_flow_reference_refused`; backend remains byte-identical | green |
| R5 | rendered firing contains no authority | renderer test requires exactly `pm --root <root> flow run <name> --json` plus sentinel and searches all state/output for token/credential/reference material | backend goldens plus fresh-binary installed firing assert direct payload and scan crontab/project artifacts for the one-time token and approval-carrier field names | green |
| R6 | scheduled terminal/parking state is durable | installed firing must execute the real flow result; overlap/crash/drift/partial write park or halt before cleanup and never advance checkpoint/replay | schedule fire-state/lock writes fsync before dispatch; overlap, process-death, cancellation, drift, partial/ambiguous write, cleanup, and parked-replay tests assert terminal state and no unsafe checkpoint/replay | green |
| R7 | unavailable shared coordination has the named SDK error/code | existing focused connsdk/engine/GitHub hook RED failed on absent named error/code and already asserts zero sends | connsdk/engine/GitHub hook tests and fresh `pm github issue list` return `*RateBudgetRefusalError` / JSON `shared_coordinator_unavailable` with zero HTTP sends | green |
| R8 | real binary reaches every production component | fresh `cmd/pm` tests/certification require the call chains recorded in the PR and observable terminal connector/schedule results | `TestPMBinaryExecutesInstalledApprovedJobFlow` builds `cmd/pm`, executes the exact installed argv, requires destination read-back, terminal schedule state, receipt, and prepared identity; separate fresh binary proves rate refusal | green |
| R9 | production certification carries no schedule authorization | CI and local `TestCertifyCLISingleConnectorPassExitsZero` failed because `stageScheduleRoundtrip` still passed `--authorization auth_0123456789abcdef` and required `authorization_reference` in list/install/remove | scripted driver rejects the obsolete flag; real and scripted create/list/install/remove envelopes plus installed crontab are scanned for authority carriers; the exact fresh-binary test and `certify-timing` pass with cleanup intact | green |

## Historical RED evidence

```text
internal/app/flow_action_test.go:218:21: a.PrepareAuthorizedFlowAction undefined
internal/app/flow_action_test.go:279:20: undefined: app.ExecutionGrantConsumedError
internal/connectors/connsdk/rate_budget_refusal_test.go:12:18: undefined: connsdk.RateBudgetRefusalError
internal/connectors/connsdk/rate_budget_refusal_test.go:13:19: undefined: connsdk.RateBudgetRefusalSharedCoordinatorUnavailable
```

The `ExecutionGrantConsumedError` assertion is explicitly superseded; it is not a requirement.

## Corrected RED commands

```sh
go test -count=1 -timeout 20m ./internal/app ./internal/flow ./internal/schedule ./internal/cli \
  -run 'Test(FlowCreate|FlowJob|ScheduleCreate|ScheduleInstall|InstalledSchedule|AuthorizedFlowActionPrepared)'
go test -count=1 -timeout 20m ./internal/connectors/connsdk ./internal/connectors/engine ./internal/connectors/hooks/github \
  -run 'Test(RateBudgetRefusal|RequireShared|GitHubWriteHook)'
```

## Observed GREEN commands

```text
go test -count=1 -timeout 20m ./internal/app ./internal/flow ./internal/schedule ./internal/cli \
  -run 'Test(FlowCreate|FlowJob|ScheduleCLI|ScheduleFire|InstalledSchedule|AuthorizedFlowAction|ConnectorFlowAction|PMBinary)'
ok internal/app, internal/flow, internal/schedule, internal/cli

go test -race -count=1 -timeout 20m ./internal/app \
  -run TestAuthorizedFlowActionConcurrentPreparedExecutionHasOneWinner
ok internal/app

go test -count=1 -timeout 20m ./internal/connectors/connsdk ./internal/connectors/engine \
  ./internal/connectors/hooks/github -run 'Test(RateBudgetRefusal|RequireShared|GitHubWriteHook)'
ok connsdk, engine, hooks/github
```
