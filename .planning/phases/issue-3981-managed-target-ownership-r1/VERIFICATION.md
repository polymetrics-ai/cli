# Verification checklist — Issue 3981: managed-target ownership

## Behavioral proof

- [x] Absent namespace and first create are typed-plan mutations followed by exact reassertion.
- [x] Repeated create and correct namespace/relation owner are idempotent.
- [x] Missing, unreadable, foreign, moved, replaced, name-colliding, orphaned,
      and schema-drifted targets are fail-closed with no mutation.
- [x] The same connection provisions a second immutable stream as a different
      relation inside its existing namespace.
- [x] A new connection gets a different namespace; an existing one reuses it.
- [x] Stream display/map-key/destination-table rename preserves stream ID and
      exact physical namespace/relation.
- [x] Driver fakes cover normal, post-create, canceled, and concurrent paths;
      no driver implementation, DDL, or generic SQL exists.
- [x] Every mutation is a typed plan against an asserted owner/target/database
      identity; names contain no display name or credentials.

## Local commands

- [x] `go test -timeout 20m ./internal/connectors/database -count=1`
- [x] `go test -timeout 20m ./internal/app -count=1`
- [x] `go test -race -timeout 20m ./internal/connectors/database -run 'TestManagedTargetProvisioning' -count=1`
- [x] `go test -race -timeout 20m ./internal/app -run '^Test(StreamIDIsPersistedAndSurvivesStreamRename|AllocateUniqueIdentityRetriesCollisions)$' -count=1`
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
- [x] Generated and manually executed `verify-work` / `code-review` prompts,
      including any required gap-cycle evidence.

## Deliberate not applicable

No CLI command, flag, help, manual, website page, connector definition, database
driver, DDL, or live credential flow changes. CLI/help/docs/website parity is not
applicable to this internal foundation slice.

## Scoped lint note

`golangci-lint run ./internal/app/... ./internal/connectors/database/...` reports
three unchanged findings outside this issue: unchecked read-only closes in
`internal/app/query_engine_duckdb.go:81,96` and unused
`internal/app/util.go:569`. None appears in this branch's changed-file list;
the repository's required `make lint` gate passed. They are not altered here to
avoid absorbing unrelated lint remediation into the ownership kernel.
