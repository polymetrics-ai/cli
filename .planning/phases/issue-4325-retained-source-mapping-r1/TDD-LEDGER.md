# TDD ledger — retained-source mapping-only bridge

## Red — observed

Planned first command:

```text
go test -timeout 20m ./cmd/connectorgen -run '^TestRetainedSourceMapping' -count=1
```

Observed command: `go test ./cmd/connectorgen -run '^TestRetainedSourceMappingCommandIsRegistered$' -count=1`.

Observed result: failed as intended because `connectorgen retained-source-mapping` was an unknown subcommand. No source-import behavior was treated as the mapping path.

## Green — observed

The implementation produces an in-memory retention-only mapping report for the frozen eight-connector/2,340-ID cohort, with all seven lanes and zero executable declarations. It rejects malformed evidence/matrix inputs fail-closed.

Observed command: `go test ./cmd/connectorgen -run '^TestRetainedSourceMapping' -count=1 -timeout 20m`.

Observed result: passed in 50.082s. The suite proves both source-lane matrix forms, CircleCI's `source_operation_id` alias, zero-GraphQL admission, missing/conflicting retained evidence rejection, exact ID reconciliation, deterministic source-only serialization, and no executable declaration claim.

The first full-cohort red attempt exposed that a descriptor/schema materializer rejects valid Docker Hub, Stripe, and Notion source facts for unrelated schema-resolution limits. The green bridge therefore verifies only retained document/operation identity facts structurally in memory; it intentionally does not run descriptor/schema materialization. This preserves mapping admission without changing runtime execution admission.

## Refactor — observed

- Make report and contract ordering deterministic.
- Keep provider source IDs opaque and do not derive partitions from HTTP methods.
- Preserve normal `canonical_evidence` importer admission behavior exactly.
- Run `gofmt`, `go vet`, focused/race tests, JSON checks, cohort check, agent-contract check, and `git diff --check`.

Observed race result: `go test -race ./cmd/connectorgen -run '^TestRetainedSourceMapping' -count=1 -timeout 20m` passed in 381.721s.
