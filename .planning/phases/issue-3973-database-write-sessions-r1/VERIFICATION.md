# Verification checklist — Issue 3973: transactional database write sessions

## Behavioural proof

- [x] Approval binds target identity, schema, mode, keys, count, and effects; mismatches have zero mutation calls.
- [x] Approval is consumed before the first session mutation and cannot be replayed.
- [x] Exactly one pinned session performs bounded batches; legacy connector writes stay at zero.
- [x] Batch failure and cancellation trigger whole-session rollback and no receipt/checkpoint authority.
- [x] Confirmed commit supplies a durable receipt before acknowledgement/checkpoint eligibility.
- [x] Unknown commit is explicit, not retried, not labelled rolled back, and cannot checkpoint.
- [x] `full_overwrite` requires atomic publish; append/upsert/dedupe use only their canonical closed session strategies.
- [x] PostgreSQL remains descriptor-only and `write=false`.

## Local commands

- [x] `go test -timeout 20m ./internal/connectors/database/...` and `go test -count=1 -timeout 20m ./internal/app/...` (run separately as required by the command timeout policy)
- [x] `go test -race -timeout 20m ./internal/connectors/database -run 'TestDatabaseWriteExecutor' -count=1`
- [x] `go test -timeout 20m ./internal/connectors/native/postgres -run '^TestWriteUnsupported$' -count=1`
- [x] `go vet ./...`
- [x] `go build ./cmd/pm`
- [x] `make tidy-check`
- [x] `make lint`
- [x] `make docs-check`
- [x] `make smoke-no-build`
- [x] `make agent-contract-check`
- [x] `make connectorgen-validate`
- [x] `make connectorgen-surface-sync`
- [x] `make connector-boundary`
- [x] `make release-workflow-check`
- [x] Inline `verify-work` and `code-review`; no implementation gap was found.

## Deliberate not applicable

No live target/database session, PostgreSQL driver, DDL, SQL, credential,
connector capability, source checkpoint implementation, CLI/help/manual,
website, connector bundle, or generated artifact changes. Driver-native and
real-binary proof belongs to #3982/#3978.
