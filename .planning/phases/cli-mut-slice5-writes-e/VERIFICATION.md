# Verification — GitHub mutation certification slice 5 writes-e

## Planned checks

- `go build ./cmd/pm`
- `go test -timeout 20m ./internal/cli`
- `go run ./cmd/connectorgen certification-matrix --check`
- `git diff --check`
- GitHub API read-back of the opened PR base

## Safety checks

- Metered Codespaces create/start/resume operations are classified `escape_needs_captain` and are never sent.
- No certification pass is recorded without independent state-change and cleanup proof.
- Credential values are never written to the repository or command output.

## Results

In progress. `variable delete-2` has one schema-v2 record validated by
`go run ./cmd/connectorgen certification-matrix --check`. Product defects keep
their declared commands; their raw controls and cleanup proof are recorded in
the TDD ledger.
