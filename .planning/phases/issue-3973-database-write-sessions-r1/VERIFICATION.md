# Verification checklist — Issue 3973: transactional database write sessions

## Behavioural proof

- [ ] Approval binds target identity, schema, mode, keys, count, and effects; mismatches have zero mutation calls.
- [ ] Approval is consumed before the first session mutation and cannot be replayed.
- [ ] Exactly one pinned session performs bounded batches; legacy connector writes stay at zero.
- [ ] Batch failure and cancellation trigger whole-session rollback and no receipt/checkpoint authority.
- [ ] Confirmed commit supplies a durable receipt before acknowledgement/checkpoint eligibility.
- [ ] Unknown commit is explicit, not retried, not labelled rolled back, and cannot checkpoint.
- [ ] `full_overwrite` requires atomic publish; append/upsert/dedupe use only their canonical closed session strategies.
- [ ] PostgreSQL remains descriptor-only and `write=false`.

## Local commands

- [ ] `go test -timeout 20m ./internal/connectors/database/... ./internal/app/...`
- [ ] `go test -race -timeout 20m ./internal/connectors/database -run 'TestDatabaseWriteExecutor' -count=1`
- [ ] `go vet ./...`
- [ ] `go build ./cmd/pm`
- [ ] `make tidy-check`
- [ ] `make lint`
- [ ] `make docs-check`
- [ ] `make smoke-no-build`
- [ ] `make agent-contract-check`
- [ ] `make connectorgen-validate`
- [ ] `make connectorgen-surface-sync`
- [ ] `make connector-boundary`
- [ ] `make release-workflow-check`
- [ ] Inline `verify-work` and `code-review`, including any GSD gap evidence.

## Deliberate not applicable

No live target/database session, PostgreSQL driver, DDL, SQL, credential,
connector capability, source checkpoint implementation, CLI/help/manual,
website, connector bundle, or generated artifact changes. Driver-native and
real-binary proof belongs to #3982/#3978.
