# Verification checklist — Issue #3976

## Acceptance proof

- [ ] Live typed catalog discovery proves two distinct PostgreSQL schema
      fixtures yield different correct catalogs by independent `pg_catalog`
      oracle inspection.
- [ ] Configured database/schema/table identity and deterministic relation,
      column, and key ordering are retained.
- [ ] Columns include nullability, native type identity/modifiers, supported
      logical type, ordinal, and ordered primary-key membership.
- [ ] Unsupported or unsafe native shapes fail closed with named secret-safe
      errors rather than coarse static/string fallback.
- [ ] The #4034 typed catalog foundation is used at PostgreSQL's shipping
      runtime boundary; no second disconnected static catalog remains.
- [ ] Static behavior owned by #3980/#3982/#3983/#3987 is recorded and not
      changed by this child.
- [ ] PostgreSQL `write`, `query`, and `cdc` capability truth remains unchanged.

## Planned local gates

- [ ] focused PostgreSQL catalog unit tests
- [ ] opt-in live `dbtest` PostgreSQL catalog tests with independent oracle,
      if local direct Podman is available
- [ ] `go test -race -timeout 20m ./internal/connectors/native/postgres`
- [ ] `go test -timeout 20m ./internal/cli -count=1`
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
- [ ] generated `verify-work` and `code-review` manual-inline records
- [ ] no-mistakes child pipeline without `--yes`, maximum five fresh correction
      loops

## Delivery holds

- [ ] Correctly stacked draft child PR exists with `Refs #3976` and `Refs #3972`.
- [ ] Parent integration remains held until corrected #4058 is green and merged.
- [ ] Automated review coverage is recorded and every actionable finding is
      dispositioned before parent-branch integration.

## CLI/docs/website applicability

No new command, flag, help topic, public output contract, generated manual, or
website page is planned. If the implementation alters any user-facing catalog
output, this checklist is amended before that change to include the repository
CLI help/docs/website parity protocol.
