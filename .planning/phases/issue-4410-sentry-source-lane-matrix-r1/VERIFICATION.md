# Verification plan — #4410 Sentry source-to-seven-lane matrix

## Executed checks

```sh
gofmt -d internal/connectors/defs/sentry/source_lane_matrix_test.go
jq empty internal/connectors/defs/sentry/sources/sentry-operation-source-lock.json
jq empty internal/connectors/defs/sentry/sources/sentry-source-lane-matrix.json
go vet ./internal/connectors/defs/sentry
go test ./internal/connectors/defs/sentry -run TestSentrySourceLaneMatrix -count=1
go test -race ./internal/connectors/defs/sentry -run TestSentrySourceLaneMatrix -count=1
go test ./internal/connectors/defs/sentry -count=1
```

All planned local checks passed. `jq empty` also passed for each copied source sidecar, and a sorted source-ID diff between the lock and matrix was empty with zero duplicate matrix IDs. `go run ./cmd/agentcontractgen check` passed.

The relevant current-main source-projection check was run deliberately:

```sh
go run ./cmd/connectorgen validate internal/connectors/defs/sentry
```

It failed before any matrix/runtime behavior with:

```text
sentry: sources/sentry-operation-source-lock.json: [source_projection] parse source lock: json: unknown field "source_operation"
connectorgen validate: 1 connector(s) checked, 1 finding(s)
```

This scoped delivery does not modify the shared importer or accept a dropped source row. No broad suite was run; no unrelated baseline failures are attributed to this change.
