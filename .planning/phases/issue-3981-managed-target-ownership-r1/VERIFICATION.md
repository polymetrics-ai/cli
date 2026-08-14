# Verification checklist — Issue 3981: managed-target ownership

## Behavioral proof

- [ ] Absent namespace and first create are typed-plan mutations followed by exact reassertion.
- [ ] Repeated create and correct namespace/relation owner are idempotent.
- [ ] Missing, unreadable, foreign, moved, replaced, name-colliding, orphaned,
      and schema-drifted targets are fail-closed with no mutation.
- [ ] The same connection provisions a second immutable stream as a different
      relation inside its existing namespace.
- [ ] A new connection gets a different namespace; an existing one reuses it.
- [ ] Stream display/map-key/destination-table rename preserves stream ID and
      exact physical namespace/relation.
- [ ] Driver fakes cover normal, post-create, canceled, and concurrent paths;
      no driver implementation, DDL, or generic SQL exists.
- [ ] Every mutation is a typed plan against an asserted owner/target/database
      identity; names contain no display name or credentials.

## Local commands

- [ ] `go test -timeout 20m ./internal/connectors/database -count=1`
- [ ] `go test -timeout 20m ./internal/app -count=1`
- [ ] `go test -race -timeout 20m ./internal/connectors/database ./internal/app -run 'TestManagedTarget|Test.*StreamID' -count=1`
- [ ] `go vet ./internal/connectors/database ./internal/app`
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
- [ ] Generated and manually executed `verify-work` / `code-review` prompts,
      including any required gap-cycle evidence.

## Deliberate not applicable

No CLI command, flag, help, manual, website page, connector definition, database
driver, DDL, or live credential flow changes. CLI/help/docs/website parity is not
applicable to this internal foundation slice.
