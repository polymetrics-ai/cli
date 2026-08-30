# TDD ledger — issue 4293 source-operation multi-lane manifest

## Red

Before implementation:

```text
go test -timeout 20m ./cmd/connectorgen -run '^TestSourceOperationMapping' -count=1
```

failed with `connectorgen: unknown subcommand "source-operation-mapping"`.
The passing-fixture test and all four independent mutations therefore failed at
the missing authoring command: duplicate source ID, locked source row absent
from the manifest, pageable row without an ETL disposition, and artifact link
to a nonexistent cell.

## Green

After the smallest authoring-only implementation:

```text
go test -timeout 20m ./cmd/connectorgen -run '^TestSourceOperationMapping' -count=1
```

passes. The suite also proves all four state forms, typed
`missing_foundation` reasons, source-evidenced `not_applicable`, deterministic
help, a v3 supplemental source lock for the same connector (3 source rows / 2
canonical operations), incompatible canonical-relation refusal, and GraphQL
root-field preservation. During refactor review, the focused artifact-path
traversal mutation was red (the checker accepted `../fixture/operations.json`)
and green after canonical-relative containment was added.

## Refactor / review

- `gofmt` on changed Go files and `git diff --check`: pass.
- `jq empty internal/connectors/engine/schema/source_operation_mapping.schema.json`: pass.
- `go vet ./cmd/connectorgen ./internal/connectors/engine`: pass.
- `go test -timeout 20m ./internal/connectors/engine -count=1`: pass.
- `go run ./cmd/connectorgen declaration-admission`: pass (`1 connector`, `1 source operation`, `0 findings`).
- `go run ./cmd/connectorgen operation-evidence --check`: existing generated-artifact drift; no rewrite was authorized or performed.
- `go run ./cmd/connectorgen validate internal/connectors/defs`: existing batch-parity findings (`553 connectors`, `50 findings`); no connector definition was touched.
- The review checked source-lock containment, duplicate JSON rejection through the existing strict decoder, deterministic findings, artifact non-creation, explicit supplemental lineage, state/reason semantics, and absence of calls into commandrunner, executors, credentials, transport, or certification.
