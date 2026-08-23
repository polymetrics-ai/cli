# Verification — issue 4329

## Status

Implementation is green for the focused engine, source-projection, and
operation-evidence behavior tests. Final repository and generated-artifact
verification remains pending.

## Required final checks

- `GOFLAGS=-p=3 go test -count=1 -timeout 20m ./internal/connectors/engine -run '^TestParseAPISurfaceOperationModelEnumRemainsClosedWithReadOnly$'` — passed
- `GOFLAGS=-p=3 go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestSourceProjectionRequiresExplicitReadOnlyNonMutationDeclaration|TestOperationEvidenceSeparatesDeclaredReadOnlyFromFoundations)$'` — passed
- `go run ./cmd/connectorgen operation-evidence --check`
- `wc -c` and `shasum -a 256` for the frozen GitHub lock and descriptor
- `make verify`

The Sentry/Vercel source locks are intentionally no longer in the production
tree after the source-lock embed slimming work, so source-import checks for
them are not locally runnable on this branch. Their historical source evidence
is used only to hand off actual mutation findings; it is not restored or
modified by this issue.
