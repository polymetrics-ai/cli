# TDD ledger — Issue #4293 mapping controls R2

## Red

Before production implementation:

```text
go test -timeout 20m ./cmd/connectorgen -run '^(TestBatch1SourceOperationMappingCohort|TestSourceOperationMapping)' -count=1
```

exited 1 after approximately 57 seconds with the expected missing-control
compile failure:

```text
cmd/connectorgen/sourceoperationmapping_test.go:194:17: undefined: sourceOperationMappingCohortPathCheck
cmd/connectorgen/sourceoperationmapping_test.go:252:19: undefined: sourceOperationMappingCohortPathCheck
```

The red tests now specify a missing/altered Batch R1 lock digest or source-ID
count, an omitted source ID, absent direct-write or reverse-ETL mutation
dispositions, a GraphQL mutation classified as non-mutating, a source-cited
non-mutating POST, omitted record shape, a fact cited to another operation
node, and contradictory ETL/binary/sync applicability.

## Green

The same focused command passed after the implementation and the final
negative-case additions:

```text
go test -timeout 20m ./cmd/connectorgen -run '^(TestBatch1SourceOperationMappingCohort|TestSourceOperationMapping)' -count=1
```

Result: exit 0 in 45.146 seconds.

The green cases include all ten tracked locks / 4,341 source rows through the
check-only CLI, altered lock and source-ID digests, altered counts and
membership, source-lock and connector-local matrix path traversal attempts,
REST and GraphQL mutation write-lane omissions, a source-cited non-mutating
POST, a GraphQL mutation-fact mismatch, omitted record shape, wrong-node fact
citations, and ETL / binary-download / binary-upload / sync contradictions.

## Refactor / review

The mapping validators were kept deterministic and authoring-only. The cohort
stores lock identity/count/digest anchors and canonical connector-local matrix
input paths, but neither source-operation rows nor lane cells. The source
mapping schema adds only cited mapping facts and no runtime fields.
