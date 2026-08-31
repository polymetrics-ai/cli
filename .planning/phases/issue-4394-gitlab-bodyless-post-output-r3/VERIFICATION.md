# Verification checklist — GitLab R3 bodyless POST source-output policy

## Required invariants

- [x] Exact frozen parent and R2 baseline recorded before correction.
- [x] GitLab retains 1,754 source rows and 12,278 lane cells; no R3 source lock,
  descriptor, matrix-ID/count, or enabled-contract-ID change.
- [x] Four exact status-source bodyless POST reads use `output_policy: none` in
  both operation and CLI declarations.
- [x] Four exact Conan JSON bodyless POST reads use `output_policy:
  json_redacted` in both declarations.
- [x] Source projection rejects status/JSON mismatch, policy drift, malformed
  policy, undeclared binding, and body-bearing POST admission.
- [x] Engine fixture proves zero-byte/no-content-type bodyless POST wire output,
  status no-body behavior, JSON decoding, nonempty-status rejection, and
  pre-I/O rejection of caller body/raw-body input.
- [x] Public CLI command table reaches typed credential boundary with zero
  provider requests for all eight exact source IDs.
- [x] Atlas record names the policy split and remains a reuse, not a new shared
  foundation.
- [x] No production engine, commandrunner, receiver, source lock/descriptor,
  generic runtime foundation, or unrelated connector diff.

## Commands to record at green

```sh
GOCACHE=/private/tmp/gocache-gitlab-r3 go test -count=1 ./cmd/connectorgen -run 'TestGitLabSourceProjection|TestGitLab.*Contract'
GOCACHE=/private/tmp/gocache-gitlab-r3 go test -count=1 ./internal/connectors/defs/gitlab -run 'TestGitLab(SourceBoundMaterializationCohort|SourceLaneMatrix|MissingFoundation)'
GOCACHE=/private/tmp/gocache-gitlab-r3 go test -count=1 ./internal/connectors/engine -run 'TestOperationDirectRead(ExecutesSourceBoundBodylessPOST|RejectsBodyForSourceBoundBodylessPOST|SupportsBoundedStatusAndTextResponses)'
GOCACHE=/private/tmp/gocache-gitlab-r3 go test -count=1 ./internal/cli -run 'TestGitLab(GeneratedDirectReadReachesCredentialBoundary|SourceLockedCommandsPassRuntimePreflight)'
GOCACHE=/private/tmp/gocache-gitlab-r3 go vet ./cmd/connectorgen ./internal/connectors/defs/gitlab ./internal/connectors/engine ./internal/cli
GOCACHE=/private/tmp/gocache-gitlab-r3 go run ./cmd/connectorgen validate internal/connectors/defs/gitlab
GOCACHE=/private/tmp/gocache-gitlab-r3 go run ./cmd/connectorgen surface-sync --check internal/connectors/defs/gitlab
jq empty docs/connector-canon/foundations/catalog.json internal/connectors/defs/gitlab/operations.json internal/connectors/defs/gitlab/cli_surface.json
jq -e '([.foundations[].id] | length) == ([.foundations[].id] | unique | length)' docs/connector-canon/foundations/catalog.json
git diff --check
```

## Recorded outcomes

- `go test -count=1 -timeout 180s ./cmd/connectorgen -run '^(TestSourceProjectionNonExecutableMutationDispositionAllowsOnlyClosedBodylessPOSTRead|TestGitLabSourceProjectionAdmitsOnlyRetainedBodylessSemanticPOSTReads|TestGitLabClosedBodylessPOSTReadsReachSurfaceSync)$'` — green (7.871s).
- `go test -count=1 -timeout 180s ./internal/connectors/defs/gitlab -run '^TestGitLabSourceBoundMaterializationCohort$'` — green (0.766s after the red proof).
- `go test -count=1 -timeout 180s ./internal/connectors/engine -run '^(TestOperationDirectReadExecutesSourceBoundBodylessPOST|TestOperationDirectReadRejectsNonemptyStatusResponseForSourceBoundBodylessPOST|TestOperationDirectReadRejectsBodyForSourceBoundBodylessPOST)$'` — green (1.011s).
- `go test -count=1 -timeout 180s ./internal/cli -run '^(TestGitLabGeneratedDirectReadReachesCredentialBoundary|TestGitLabBodylessPOSTReadCommandsReachCredentialBoundaryWithoutProviderIO)$'` — green (8.535s).
- `go test -race -count=1 -timeout 180s ./internal/connectors/engine -run '^(TestOperationDirectReadExecutesSourceBoundBodylessPOST|TestOperationDirectReadRejectsNonemptyStatusResponseForSourceBoundBodylessPOST|TestOperationDirectReadRejectsBodyForSourceBoundBodylessPOST)$'` — green (2.171s).
- `go vet ./cmd/connectorgen ./internal/connectors/engine ./internal/connectors/defs/gitlab ./internal/cli` — green.
- `go run ./cmd/connectorgen surface-sync --check --connector gitlab` — green: zero fields filled/corrected.
- `jq -e .` for the changed JSON documents and `git diff --check` — green.

`go run ./cmd/connectorgen validate --connector gitlab` remains red with 23 pre-existing broader GitLab findings: three legacy write coverage entries, six legacy CLI endpoint-coverage entries, and fourteen unrelated source-gap reads. This R3 diff changes only four `output_policy` values and their source-policy conformance proof; it leaves the source lock, descriptor, matrix, enabled contract, API surface, and legacy write views unchanged. The broader four-package `go test` aggregation was also stopped by unrelated baseline assertion/timeouts and is not a substitute for the focused gates above.

Race gate, if the shared disk/concurrency slot is healthy:

```sh
GOCACHE=/private/tmp/gocache-gitlab-r3 go test -race -count=1 ./internal/connectors/engine -run 'TestOperationDirectRead(ExecutesSourceBoundBodylessPOST|RejectsBodyForSourceBoundBodylessPOST|SupportsBoundedStatusAndTextResponses)'
```

## Review gate

- [x] Run the generated `gsd-code-review` work inline against the final changed
  files and record a zero-blocker result or exact findings.
- [ ] Parent assigns a fresh-context independent exact-SHA review after push.
- [ ] No integration or merge occurs in this task.

### Inline review result

`gsd-code-review` was rendered before execution; the active one-worker runtime
requires the documented inline fallback. Review of the final diff found no
blocker: the predicate is narrower than the pre-existing semantic POST-read
exception, the four declarations are changed symmetrically in operations and
CLI metadata, all eight source IDs are exercised through engine and public CLI
boundaries, and no source lock, descriptor, matrix, enabled contract, API
surface, commandrunner, receiver, or production engine file changed. A fresh
exact-SHA review remains required after push.
