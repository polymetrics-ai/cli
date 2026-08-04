# TDD Ledger — destructive-write confirmation foundation

Phase: delete-confirmation-foundation-r1

Red evidence is recorded here before any production edit. Each slice later receives its exact green
command and broader verification result.

## task:tdd-plan-schema

Status: green-verified

- RED: `go test ./internal/connectors/engine -run 'TestBundleLoadAcceptsClosedDestructive(Write|Operation)Confirmation' -count=1`
  failed because both schemas rejected `/confirmation` as an additional property.
- GREEN: the same command passed after adding the shared closed `ConfirmationSpec` and strict
  `{"kind":"destructive"}` schema to writes and operations.
- REFACTOR: typed decode assertions plus rejection coverage for unknown kinds and free-text fields.
- EXECUTED VERIFICATION: `go test ./internal/connectors/engine` passed, including both closed
  schemas and their unknown-kind/additional-field rejection cases.

## task:tdd-preview

Status: green-verified

- RED: `go test ./internal/app -run 'Test(RunReverseETLRejectsDestructiveConnectorCommandWithoutPreview|DestructiveConnectorCommandMintsApprovalOnlyAfterPreview)' -count=1`
  showed a destructive command executing without preview and a token minted during planning.
- RED: `go test ./internal/app -run TestGenericDestructivePlanMintsApprovalOnlyAfterPreview -count=1`
  failed to compile because generic reverse plans had no real preview operation.
- RED: `go test ./internal/app -run TestDestructiveCanonicalCommandPreviewProducesApprovablePlan -count=1`
  showed canonical `--preview` returning a non-approvable `planned` plan.
- GREEN: the targeted app tests pass with persisted digest/status/time and approval minted only
  after a no-network dry run for connector-command and generic reverse plans.
- REFACTOR: connector and generic preview paths share `persistDestructivePreview`; canonical
  command `--preview` uses the same lifecycle.
- EXECUTED VERIFICATION: `go test ./internal/app` and the canonical CLI lifecycle test passed;
  re-preview after execution is rejected so it cannot mint a replay token.

## task:tdd-approval

Status: green-verified

- RED: `go test ./internal/connectors/engine -run TestWriteRejectsDestructiveActionWithoutTypedApprovalEvidence -count=1`
  dispatched a DELETE when `confirm` was omitted and no approval evidence existed.
- RED: `go test ./internal/connectors -run TestParseWriteConfirmationIsClosed -count=1`
  failed to compile because no closed confirmation parser existed.
- RED: the app compile check rejected `WriteConfirmation` because `RunReverseETLRequest` still
  accepted a free-form string.
- GREEN: the same targeted commands pass. HTTP DELETE, delete/destructive mutation classes, and
  typed declarations normalize to one gate; CLI input is parsed to the closed enum before app use.
- REFACTOR: approval tokens stay in the app and are consumed into non-secret, plan/preview-bound
  `WriteApprovalEvidence` for executors.
- RED: `go test ./internal/connectors/engine -run TestWriteFailsClosedForUnknownConfirmationDeclaration -count=1`
  dispatched a POST carrying an unknown non-empty confirmation declaration.
- GREEN: the same command passed after normalized targets treated every non-empty declaration as
  gated; the typed evidence still accepts only `destructive`.
- EXECUTED VERIFICATION: connector, engine, app, commandrunner, and CLI package checks passed.

## task:tdd-execute-seam

Status: green-verified

- RED: `go test ./internal/connectors/engine -run TestRestWriteOperationUsesSharedDestructiveExecutionGate -count=1`
  failed to compile with `undefined: DestructiveTargetForOperation`.
- GREEN: the test passes: a typed `rest_write` operation normalizes into the shared target and its
  approved closure executes through `GateDestructiveExecution` unchanged.
- REFACTOR: provider-specific dispatch/results stay inside a closure; the shared gate has no REST
  request implementation dependency.
- RED: `go test ./internal/connectors/engine -run TestRestWriteDestructiveFlagCannotBeOverriddenByMutationClass -count=1`
  showed `destructive:true` could be hidden by a simultaneous `mutation_class:admin`.
- GREEN: the normalized target retains the independent destructive boolean; the seam test proves
  an unapproved callback is not invoked and an approved callback runs exactly once.
- EXECUTED VERIFICATION: `go test ./internal/connectors/engine` passed.

## task:tdd-bypass

Status: green-verified

- RED: `go test ./internal/app -run TestRunReverseETLRejectsPreviewDigestDriftBeforeNativeWrite -count=1`
  showed a native writer could change its preview digest and still receive the destructive write.
- RED: `go test ./internal/cli -run TestGitHubDestructiveCommandRequiresTypedConfirmation -count=1`
  showed a destructive canonical plan had no usable preview-to-approval transition.
- RED: `go test ./internal/app -run TestExecutedDestructivePlanCannotBeRepreviewedForReplay -count=1`
  showed an executed plan could be previewed back into an approvable state.
- RED: commandrunner/native-manifest tests showed an HTTP DELETE with omitted legacy `confirm`
  could be classified as safe before method-based inference.
- RED: the Asana and Zendesk Support execution fixtures failed when their existing DELETE actions
  reached the engine without approval evidence; captain decision `defs-delete-fixtures` approved
  bounded test-only updates.
- GREEN: app replays the no-network dry run immediately before dispatch, compares the digest, then
  the engine independently compares it again; the canonical public `github repo deploy-key delete`
  plan/preview/confirm/run fixture passes without provider calls.
- GREEN: the Asana and Zendesk Support fixtures obtain authenticated grants through the real app
  plan, preview, confirmation, and execution lifecycle while preserving every HTTP request
  assertion.
- EXISTING GUARD: `batchable:false` remains checked at bulk plan time, persisted-plan execution
  time, and is intentionally allowed only for the single-record canonical command path.
- EXECUTED VERIFICATION: `go test ./internal/app`, `go test ./internal/connectors/conformance`,
  `go test ./internal/connectors/defs/asana ./internal/connectors/defs/zendesk-support`, and the
  focused canonical CLI test all passed.

## task:tdd-authenticated-grant-hardening

Status: green-verified

- MANUAL-GSD FALLBACK: `scripts/gsd prompt programming-loop init --phase delete-confirmation-foundation-r1 --dry-run` exited 1 with `unknown GSD command: programming-loop`; the repository-approved manual universal loop remains active.
- RED: `go test ./internal/app ./internal/connectors/engine ./internal/connectors/native/amazon-sqs -run 'Test(GateRejectsForgedAndReplayedDestructiveEvidence|DryRunWriteDigestBindsCanonicalRequestAndCredentialRevision|DestructiveWriteUsesPreparedPreviewAndSharedGate|RunReverseETLRejectsApprovalHashStateTamper|RunReverseETLConsumesApprovalAtomicallyAcrossProcesses|PreviewReversePlanRejectsExpiredGenericPlan)$' -count=1` exited 1: forged evidence executed, copied evidence replayed twice, token-hash tamper executed, secret-derived target digest stayed unchanged, expired generic preview minted approval, and SQS returned no digest then dispatched without approval.
- RED: `go test ./internal/app -run TestRunReverseETLConsumesApprovalAtomicallyAcrossProcesses -count=1` exited 1 with two destructive requests after the deterministic stale second process began only after the first process had consumed its stale state and entered provider dispatch.
- GREEN: one external-key authenticated grant and one exact prepared-write seam are shared by
  declarative and native executors; command and bulk paths atomically reload and consume the grant
  before dispatch, and copied engine evidence remains one-shot.
- EXECUTED VERIFICATION: `go test ./internal/connectors ./internal/vault ./internal/connectors/engine ./internal/connectors/hooks/github ./internal/connectors/conformance ./internal/connectors/native/amazon-sqs ./internal/app ./internal/connectors/defs/asana ./internal/connectors/defs/zendesk-support -count=1` passed.
