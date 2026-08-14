# Plan — Issue #4090: PostgreSQL native database source transport

## Scope and ownership

Target connector: `postgres` only. Allowed production paths are
`internal/connectors/native/postgres/**` and
`internal/connectors/defs/postgres/**`; evidence is confined to this phase
directory. No shared engine, transport registry, bundle-schema, App
composition, target, polling, CDC, or certification report code is in scope.

## TDD slices

1. **Plan checkpoint (green):** record the required GSD commands, skills,
   boundaries, test matrix, and live-proof command before production edits.
2. **RED:** add focused PostgreSQL transport tests that require a definition
   descriptor and registration, reject the wrong family and missing executor
   before a counting connector can perform I/O, and require bounded typed
   full-page records/checkpoints. Capture the failing command output.
3. **GREEN:** add the definition-selected descriptor, PostgreSQL source
   executor, and registration helper. Reuse typed catalog/definition resources
   to select only safe relations/columns/order, construct a bounded snapshot,
   and emit deterministic records/checkpoints.
4. **GREEN hardening:** test full-append and full-overwrite; descriptor/family/
   registration preflight refusals; cancellation and invalid resume; type,
   identity, page-bound, and checkpoint invariants.
5. **Live green:** extend the existing dbtest PostgreSQL fixture and run it
   against an explicit local Docker-or-Podman Unix endpoint. Its output must
   display bounded row IDs/labels, catalog schema fingerprint/identity, and
   the emitted checkpoint fields.
6. **Refactor/review:** run focused tests, live proof, package race/test/vet,
   required individual repository gates, manual `verify-work`, then manual
   deep `code-review`. Record all dispositions.

## Design constraints

- The source must never take a raw SQL string, query fragment, target
  connector, destination credentials, or generic endpoint from a caller.
- Perform all non-I/O validation before opening the PostgreSQL pool. Registry
  preflight is the proof point for missing descriptor/family/executor errors.
- Every page is no larger than `SourceRequest.BatchSize` and the database
  definition resource maximum; an empty final snapshot remains checkpointed
  without falsely claiming incremental resume.
- Deterministic identity is assembled only from the typed discovered relation,
  schema fingerprint, closed executor id, and full-snapshot mechanism. It is
  not a stringified DSN, secret, raw row serialization, or app-composed value.
- No changes to `internal/connectors/engine/bundle.go`; its certification
  report contract remains exactly version 1.

## Validation commands

```sh
go test -timeout 20m -count=1 ./internal/connectors/native/postgres ./internal/synctransport
go test -race -timeout 20m -count=1 ./internal/connectors/native/postgres
go vet ./internal/connectors/native/postgres
go build ./cmd/pm
make tidy-check
make lint
make docs-check
make smoke-no-build
make agent-contract-check
make connectorgen-validate
make connectorgen-surface-sync
make connector-boundary
make release-workflow-check
```

The exact live command and real test output are recorded in
`traces/live-source-green.txt`. CI, not this task's per-command runner, owns
the full repository suite.
