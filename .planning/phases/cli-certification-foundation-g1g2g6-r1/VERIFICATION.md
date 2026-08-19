# Verification: connector certification foundation G1/G2/G6

## Required checks

- [x] Focused `cmd/connectorgen` classification, sweep, proof, and evidence tests (including race/concurrent reader tests).
- [x] `go run ./cmd/connectorgen certification-sweep --connector github --check`
- [x] `go run ./cmd/connectorgen certification-matrix --check`
- [x] `go run ./cmd/connectorgen surface-sync --check`
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs`
- [x] Changed package tests and `internal/cli` in separate timeout-bounded commands.
- [x] `go vet ./...`, `go build ./cmd/pm`, generated/snapshot checks, and repository verification gates.
- [x] `git diff --check`; PR #4259 API base read-back returned `integration/4015-mvp-flat-r1`. Automated review remains pending CI.

## Results

- `go test -count=1 -timeout 20m ./cmd/connectorgen` — PASS, including G1/G2 table,
  generated-artifact, scoped-check, prevalidated-batch, no-replace,
  concurrent-reader, and runner-order regressions.
- `go test -count=1 -timeout 20m ./...` — PASS on the rebased integration
  base (including the changed generator package and `internal/cli`).
- `go run ./cmd/connectorgen certification-sweep --connector github` and its
  `--check` — PASS; 1,575 rows (1,571 CLI commands plus four generated
  declaration rows) each retain `operation_kind` and `op_class`.
- Rebase reconciliation — PASS: `go test -count=1 -timeout 20m
  ./cmd/connectorgen -run 'TestCertificationSweepForGitHubIsSurfaceDerivedAndExhaustive|TestCertificationSweepEmitsDeclaredManagedDestinationProjection|TestCertificationSweepProjectsDeclaredAPIReadCapabilities'`
  proves the actual PostgreSQL managed destination becomes a reverse-ETL row,
  while Zoom/GitLab `capability read` rows become `rest_read/direct_read`.
- `go run ./cmd/connectorgen certification-matrix --connector github --check`
  and global `go run ./cmd/connectorgen certification-matrix --check` — PASS.
- `go run ./cmd/connectorgen certification-candidates --connector github --check`,
  `go run ./cmd/connectorgen surface-sync --check`,
  `go run ./cmd/connectorgen validate internal/connectors/defs`, and
  `go run ./cmd/agentcontractgen check` — PASS.
- `go vet ./...`, `go build ./cmd/pm`, and `make tidy-check lint docs-check
  smoke-no-build agent-contract-check connectorgen-validate
  connectorgen-surface-sync connector-boundary release-workflow-check` — PASS;
  the connector canon, pinned dependency, Homebrew notification, release-target,
  and GitHub ledger Node checks also passed.
- Real GitHub proof — PASS, using the Keychain reference only at process start:
  `PM_CERT_GITHUB_TOKEN="$(security find-generic-password -s pm-cert-classic -w)" node scripts/certify-connector-live.mjs github --pm .tmp/pm-live-certification --credential-env PM_CERT_GITHUB_TOKEN --credential-field token --credential-config owner=Polymetrics-Cert --credential-config repo=pm-cert-3993-20260810-wz0fru --credential-config rate_limit_account=polymetrics-ai-certification --limit 1`.
  The read-only `repo read-file` run returned HTTP 200, published schema-v2
  accepted evidence `github-direct_read_sweep_repo_read_file-github-1787097821338-6301903bc006.json`, regenerated GitHub's scoped shard, and passed its scoped
  check. The runner removed only its sanitized draft and the ephemeral project;
  its accepted evidence record remained present. No credential value is stored
  here.
