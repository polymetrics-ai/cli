# Verification — Issue 4166 Live Write Certification

**Current status:** the repository-safe wave is complete as far as the current
disposable credential can independently prove it: 25 actions live-pass and
three commit-comment actions are blocked/non-certified. The separately
recorded browser-only provisioning decision remains open for the larger waves.

- [x] Isolated worktree and required base branch verified.
- [x] Binding context, issue #4166, GSD adapter, CLI parity reference, and
  required skills reviewed.
- [x] All 607 declared GitHub write actions classified by exhaustive selector;
  25 repository actions are live-proven, three have the exact Metadata-read
  permission blocker, and every remaining action has a concrete disposable
  fixture, permission, App/OAuth, enterprise, secondary-user, notification,
  or sacrificial-credential prerequisite.
- [x] Rate/runtime and resumability requirements recorded.
- [x] Captain decision authorizes the repository-safe resource boundary.
- [x] Report truthfulness, full-parity refusal, and durable write lifecycle
  checkpoint tests are green.
- [x] Red/green implementation and live provider evidence: production
  `--write-only` run has 25 independent read-back/verified-cleanup passes;
  commit comments are `blocked`, never pass, and the deliberately broken
  post-schema `update_issue` request fails certification (exit 2) before the
  scratch change is restored.
- [x] Repository-wave resume recovery refuses unowned resources, removes only
  tagged named resources, and restores an exact ledger-captured topic baseline.
- [x] Connector boundary correction: `write_wave` and `write_inventory` keep
  provider-specific fixture, action, pairing, tag, and read-back-blocker data
  in the GitHub certification definition; the shared sweep names no provider.
  Focused engine/certify tests, `make connector-boundary`, and a byte-stable
  GitHub certification-matrix regeneration passed.
- [x] Certification timing correction: CI's 409.546s result was repeated full
  bundle loading, not provider work or an accepted slower budget. The retained
  five affected tests are 0.13–0.21s in a cold focused run, and `make
  certify-timing` passes at 33.473s total (8.902s certify, 24.571s CLI) against
  the unchanged 210s budget. The lightweight discovery fails if it finds no
  declaration or any declaration cannot load through `certificationWriteWaveFor`.
- [x] Round-two rebase/matrix check: absorbed `ff6a8710`; the CI report that
  named malformed MySQL actually logged a stale Postgres certification shard.
  The unchanged exact test passes at the original timing commit, its parent,
  and the rebased head; the rebased `cmd/connectorgen` package and
  certification-matrix drift check pass. The full `make test` command reached
  the unrelated CLI package timeout under whole-suite concurrency, while its
  prescribed separate `go test -timeout 20m ./internal/cli` fallback passes.
- [x] Full-parity now implies `--full --write`, rejects `--skip write`, and a
  fresh external HTTPS child refuses to write a proof artifact for incomplete
  write coverage.
- [x] GSD code review completed by documented inline/manual fallback; no
  findings (`REVIEW.md`).
- [x] Local repository gates passed:
  `gofmt -w cmd internal`; `go test -timeout 20m ./internal/connectors/certify`
  (pass); `go test -timeout 20m ./internal/cli` (pass, 553.529s);
  `go vet ./...`; `go build ./cmd/pm`; `make tidy-check lint
  docs-check-no-build smoke-no-build agent-contract-check
  connectorgen-validate connectorgen-surface-sync
  github-parity-artifacts-check connectorgen-certification-matrix
  connector-boundary connector-canon-check release-workflow-check`; `pnpm --dir
  website run gen:docs` twice; `scripts/verify-gsd-workflow`; and `git diff
  --check` all passed. `go test -timeout 20m ./...` / aggregate `make verify`
  were not run as one local command because the project verification contract
  explicitly directs per-command agents to run changed packages plus
  `internal/cli` separately and leave the full 550+-connector suite to CI.
- [ ] Commit, push, direct PR, and API base verification.
