# Verification checklist — cli-unknown-subcommand-false-success-r1

Status: **local verification complete**.

- [x] Focused red regression test failed against the unchanged resolver; retained in `TDD-LEDGER.md`.
- [x] `go test -timeout 20m ./internal/cli -count=1` passes.
- [x] `go vet ./internal/cli` passes.
- [x] `go build ./cmd/pm` passes after the change.
- [x] Golden transcript regeneration is inspected; only the new deep-invalid-help JSON entry changed.
- [x] CLI parity checks pass: connector root, root help flag, real deep help, invalid deep help,
  connector-level unknown, and JSON structured error.
- [x] The canonical website CLI reference and agent guide describe the resolved-help requirement;
  `website/lib/docs.generated.ts` was regenerated from those source documents.
- [x] `scripts/verify-gsd-workflow origin/main` accepts the committed implementation and its GSD/TDD evidence.
- [x] Individual relevant verification gates pass: `agentcontractgen check`, `connectorgen validate`,
  `connectorgen surface-sync --check`, `connectorgen boundary . --json`, and
  `scripts/tests/release-target-parity.sh`. CI owns the monolithic full suite under the repository timeout policy.

## Commands run

```text
go test -timeout 20m ./internal/cli -count=1
go vet ./internal/cli
go build ./cmd/pm
go run ./cmd/agentcontractgen check
go run ./cmd/connectorgen validate internal/connectors/defs
go run ./cmd/connectorgen surface-sync --check
go run ./cmd/connectorgen boundary . --json
scripts/tests/release-target-parity.sh
node website/scripts/gen-docs-data.mjs
```
