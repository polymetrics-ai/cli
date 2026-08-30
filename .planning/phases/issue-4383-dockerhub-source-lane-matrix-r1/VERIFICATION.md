# Verification plan — #4383 Docker Hub source-to-seven-lane matrix

## Planned checks

```sh
gofmt -d internal/connectors/defs/dockerhub/source_lane_matrix_test.go
jq empty internal/connectors/defs/dockerhub/sources/dockerhub-operation-source-lock.json
jq empty internal/connectors/defs/dockerhub/sources/dockerhub-source-lane-matrix.json
go vet ./internal/connectors/defs/dockerhub
go test ./internal/connectors/defs/dockerhub -run TestDockerHubSourceLaneMatrix -count=1
go test -race ./internal/connectors/defs/dockerhub -run TestDockerHubSourceLaneMatrix -count=1
go test ./internal/connectors/defs/dockerhub -count=1
go run ./cmd/agentcontractgen check
```

The generic source/import/declaration checks will be run only to record their current-main result. Any retained source-lock compatibility rejection is a source-preserving mapping restriction; it will not be repaired in this scoped connector-local slice.

## Observed scoped results

- Red: the first focused run compiled the test and failed only because
  sources/dockerhub-source-lane-matrix.json did not yet exist.
- Green and edge: go test ./internal/connectors/defs/dockerhub -count=1 and
  go test -race ./internal/connectors/defs/dockerhub -run
  TestDockerHubSourceLaneMatrix -count=1 pass. The decoded-matrix adversarial
  cases cover hidden/duplicate source rows, invalid source backlinks,
  nonexistent cell backlinks, omitted fact citations, missing ETL, missing
  direct-write/reverse-ETL, and source-count mismatch.
- Formatting/structure: gofmt -d emitted no diff; jq empty passed for the four
  byte-verified sidecars and the new matrix; go vet
  ./internal/connectors/defs/dockerhub passed.
- Contract registry: go run ./cmd/agentcontractgen check passed.
- Connector-local reconcile: go run ./cmd/connectorgen surface-reconcile
  internal/connectors/defs/dockerhub --check --json exited zero and reported
  50 retained runtime-ledger records as refused, with no source mutation.

## Preserved source-import restriction

The current generic importer and validator do not admit the immutable Docker
Hub schema-v2 lock because its source rows retain the legacy source_operation
member:

~~~
go run ./cmd/connectorgen source-import dockerhub --defs internal/connectors/defs --out /private/tmp/dockerhub-source-import-check.json --check
connectorgen source-import: parse source lock: json: unknown field "source_operation"

go run ./cmd/connectorgen validate internal/connectors/defs/dockerhub
dockerhub: sources/dockerhub-operation-source-lock.json: [source_projection] parse source lock: json: unknown field "source_operation"
~~~

This is recorded as a mapping/admission restriction, not a missing runtime
foundation or a reason to drop rows. The connector-local validator reads the
immutable lock and crosswalk directly, keeps every one of the 54 source IDs
visible, and does not change importer, runtime, generator, certification, or
source bytes to work around the restriction.
