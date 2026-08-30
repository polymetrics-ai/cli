# Verification — CircleCI semantic lane repair R2

Status: local verification complete and the scoped repair is pushed to `origin/fix/4382-circleci-semantic-repair-r2`.

Executed commands:

```text
go test ./internal/connectors/defs/circleci -run 'TestCircleCI(SourcePagingEvidenceUsesSourceSemanticsNotHTTPMethod|SourcePagingEvidenceResolvesReferencedResponseSchemas|SourcePagingEvidenceRejectsRequestOnlyAndInlineFalsePositives|SourceResponseTokensWithoutContinuationAreReconciled|SourceWebhookRegistrationIsTheOnlySyncCandidate|WebhookRegistrationRequiresDeliverySemantics|SourceLaneMatrixRejectsPagingAsSyncTransport|SourceLaneMatrixRejectsWebhookRegistrationWithoutNamedGap)' -count=1
go test ./internal/connectors/defs/circleci -count=1
go test -race ./internal/connectors/defs/circleci -count=1
go vet ./internal/connectors/defs/circleci
jq empty internal/connectors/defs/circleci/sources/circleci-source-lane-matrix.json
jq -e '(.rest.source_url == "https://circleci.com/api/v2/openapi.json") and (.rest.sha256 == "61c6ce11e8de509948aa3d53dcd0169913f52de20920b130b6a85dea41d66d07") and ((.rest.operations | length) == 111)' internal/connectors/defs/circleci/sources/circleci-operation-source-lock.json
jq -e '(.source_lock.source_url == "https://circleci.com/api/v2/openapi.json") and (.source_lock.sha256 == "61c6ce11e8de509948aa3d53dcd0169913f52de20920b130b6a85dea41d66d07") and (.source_lock.source_operation_count == 111)' internal/connectors/defs/circleci/sources/circleci-source-lane-matrix.json
go run ./cmd/agentcontractgen check
git diff --check
```

All commands passed. The unchanged retained source lock and lane matrix bind provider source `https://circleci.com/api/v2/openapi.json`, SHA-256 `61c6ce11e8de509948aa3d53dcd0169913f52de20920b130b6a85dea41d66d07`, for 111 operations. The cohort ledger's `06527e…` value is the serialized lock-file byte digest, not the provider-source digest used by this verification.

Matrix reconciliation is 111 source rows × 7 lanes = 777 cells: 175 `mapped_unproven`, 2 `missing_foundation`, and 600 `not_applicable`. The 14 ETL cells include the five repaired response-reference paginators. The only non-NA sync cells are `createWebhook` and `updateWebhook`, both source-backed `missing_foundation`; none are executable claims.

Residual restriction: CircleCI source documents outbound registration but this Track A patch does not add a CircleCI-specific inbound webhook receiver, HMAC signing-secret verification, or delivery/replay conformance proof. That runtime work requires a separately approved foundation task; it is not a test or source-lock defect in this repair.
