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

## SE-R1-001 semantic repair verification

The independent audit of `1e66668ecb3a12756fee8644df60e56182a35fd0` found that ETL and `sync_transport` had both been derived from a cursor/array collection heuristic. The repair ran only against the Sentry matrix/test/evidence slice.

Commands that passed:

```sh
gofmt -w internal/connectors/defs/sentry/source_lane_matrix_test.go
go test ./internal/connectors/defs/sentry -run 'TestSentrySourceLaneMatrix|TestSentrySemanticLanePredicatesUseSourceContracts' -count=1
go vet ./internal/connectors/defs/sentry
go test ./internal/connectors/defs/sentry -count=1
go test -race ./internal/connectors/defs/sentry -run 'TestSentrySourceLaneMatrix|TestSentrySemanticLanePredicatesUseSourceContracts' -count=1
jq empty internal/connectors/defs/sentry/sources/sentry-operation-source-lock.json
jq empty internal/connectors/defs/sentry/sources/sentry-source-lane-matrix.json
go run ./cmd/agentcontractgen check
git diff --check
```

A JSON set reconciliation independently reports 223 lock IDs, 223 matrix IDs, exact equality, and zero duplicate matrix IDs.

The matrix now has 45 ETL candidates: 43 cursor facts plus two SCIM `startIndex` facts whose provider descriptions explicitly say pagination. Seventeen JSON-array reads with no continuation mechanism and the `per_page`-only session aggregate remain `not_applicable` for ETL. `sync_transport` is one `mapped_unproven` provider contract, `Register a New Service Hook`, because its request source requires a webhook callback URL and event selector; listing service hooks is not sync. Direct read/write/reverse mapping uses provider action language plus documented success response; binary mapping still uses provider-published media only.

The shared source-projection check remains a known non-green residual and is intentionally unchanged:

```text
sentry: sources/sentry-operation-source-lock.json: [source_projection] parse source lock: json: unknown field "source_operation"
connectorgen validate: 1 connector(s) checked, 1 finding(s)
```

It is a shared importer compatibility gap, not a runtime foundation need and not justification to remove a source row or make a source-only cell executable.
