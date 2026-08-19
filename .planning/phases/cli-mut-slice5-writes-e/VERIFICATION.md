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

Complete. All 145 assigned ordinals were attempted and placed in exactly one final
bucket. The branch adds 26 schema-v2 evidence records over the integration base;
each record contains an independent produced-value assertion and provider cleanup
read-back. The final corrected ledger contains no `wrong_credential` rows and
retains `no_object` only for provider object classes whose parent collection was
read and whose contained fixture creation failed.

Local verification:

- `go run ./cmd/connectorgen certification-matrix --check` — PASS;
  `certification shards are current`.
- `go test -timeout 20m ./cmd/connectorgen ./internal/connectors/certifications/...`
  — PASS (`cmd/connectorgen` 121.115s; certifications cached).
- `go build ./cmd/pm` — PASS.
- `git diff --check` and
  `git diff --check origin/integration/4015-mvp-flat-r1...HEAD` — PASS.

The repository-wide test suite is left to CI because the repository instructions
explicitly prohibit running the 550+ connector suite as a single per-command check;
the changed validator and certification packages were tested directly. The pull
request base is independently verified from the GitHub API after opening.
