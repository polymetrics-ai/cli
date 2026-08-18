# Code review — #4181

## Scope

Manual review of the transport approval, orchestration, declarative page-read,
run persistence, CLI help/documentation, and binary-scale-test changes.

## Findings and dispositions

| Finding | Disposition | Evidence |
| --- | --- | --- |
| A durable plan could be validated after its preview seal was removed. | Fixed. Managed-target pre-token validation now rejects a missing seal as `ErrPostgresManagedTargetApprovalStale`. | `TestPostgresManagedTargetDurableAuthorizationValidationDoesNotReusePreviewSealLifetime` failed first, then passed. |
| A destination read-back request-clone error could leave its unit timeout context uncancelled. | Fixed. A per-unit closure defers cancellation immediately after `context.WithTimeout`, so it releases before the next batch. | `go vet ./...` reported the path; vet and transport tests pass after the fix. |
| Phase-measurement finalization initially bypassed acknowledged-failure rebase recovery. | Fixed. The dispatch result retains its pending acknowledged stream state on an error before failure persistence. | `TestRunETLTransportAcknowledgedFailureAfterUnrelatedRevisionForAllModes`, `TestRunETLTransportAcknowledgedFailurePreservesSourceError`, and `TestRunETLTransportAcknowledgedFailureReturnsTruthfulPersistenceOutcome` failed first, then passed. |
| A long-run failure used to lose its three counts during cleanup. | Fixed. Terminal `Run.TransportPhaseMeasurement` is persisted in success and failure paths; the binary harness reopens state before cleanup. | Failure measurement and one-page/90k live proof tests. |
| Lifetime increase could regress to a 48-hour ephemeral evidence grant. | Rejected by design. The existing 15-minute preview grant remains one-time; this route converts it atomically into a duration-bound, revocable `AuthorizationRecord`, and unit deadlines remain short. | `postgres_transport_approval.go`, `transport_test.go`, live 90k run beyond 15 minutes. |

## Outcome

No open local findings. Automated review route is pending PR creation: this is
a non-default-base sub-PR, so the expected primary route is trusted-author
Claude auto review on PR open, with parent-PR fallback only if that review does
not materialize.
