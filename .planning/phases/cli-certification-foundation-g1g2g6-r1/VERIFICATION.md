# Verification: connector certification foundation G1/G2/G6

## Required checks

- [x] Focused `cmd/connectorgen` classification, sweep, proof, and evidence tests (including race/concurrent reader tests).
- [x] `go run ./cmd/connectorgen certification-sweep --connector github --check`
- [x] `go run ./cmd/connectorgen certification-matrix --check`
- [x] `go run ./cmd/connectorgen surface-sync --check`
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs`
- [x] Changed package tests and `internal/cli` in separate timeout-bounded commands.
- [x] `go vet ./...`, `go build ./cmd/pm`, generated/snapshot checks, and repository verification gates.
- [x] `git diff --check`; automated review and PR API base read-back remain pending PR opening.

## Results

- `go test -timeout 20m ./cmd/connectorgen` — PASS, including G1/G2 table,
  generated-artifact, scoped-check, prevalidated-batch, no-replace,
  concurrent-reader, and runner-order regressions.
- `go test -timeout 20m ./internal/cli` and a timeout-bounded batched full
  suite (`go list ./...` grouped into `go test -timeout 20m` invocations) —
  PASS.
- `go run ./cmd/connectorgen certification-sweep --connector github` and its
  `--check` — PASS; 1,571 rows now each retain `operation_kind` and `op_class`.
- `go run ./cmd/connectorgen certification-matrix --connector github --check`
  and global `go run ./cmd/connectorgen certification-matrix --check` — PASS.
- `go run ./cmd/connectorgen certification-candidates --connector github --check`,
  `go run ./cmd/connectorgen surface-sync --check`,
  `go run ./cmd/connectorgen validate internal/connectors/defs`, and
  `go run ./cmd/agentcontractgen check` — PASS.
- `go vet ./...`, `go build ./cmd/pm`, `make lint`, `make docs-check`,
  `make smoke-no-build`, `make connector-boundary`, and the connector canon,
  pinned dependency, Homebrew notification, release-target, and GitHub ledger
  Node checks — PASS.
- Real GitHub proof — PASS, using the Keychain reference only at process start:
  `PM_CERT_GITHUB_TOKEN="$(security find-generic-password -s pm-cert-classic -w)" node scripts/certify-connector-live.mjs github --pm .tmp/pm-live-certification --credential-env PM_CERT_GITHUB_TOKEN --credential-field token --credential-config owner=Polymetrics-Cert --credential-config repo=pm-cert-3993-20260810-wz0fru --credential-config rate_limit_account=polymetrics-ai-certification --limit 1`.
  The read-only `repo read-file` run returned HTTP 200, published schema-v2
  accepted evidence, regenerated GitHub's scoped shard, and passed its scoped
  check. The runner removed only its sanitized draft and the ephemeral project;
  its accepted evidence record remained present. No credential value is stored
  here.
