# Verification — MySQL container harness R1

## Completed evidence

- [x] Focused unit tests for the harness, MySQL native connector, bundle loader, and changefeed contract.
- [x] The focused bundle-registry test asserts that the public `mysql` entry is the native connector
      and exposes its matching binlog changefeed executor.
- [x] `govulncheck ./...` (scanner v1.6.0, DB vuln.go.dev): **no vulnerabilities found** across the
      whole module, including `go-mysql` and its transitive set.
- [x] Measured binary size delta for the new dependency: 94,191,266 -> 100,101,650 bytes
      (+5,910,384; +5.64 MiB; +6.27 %), `go build -trimpath ./cmd/pm`, darwin/arm64.
- [x] Documented tagged MySQL **Podman** test on an explicit `--connection`, with the host-disk
      reclaim opt-in enabled. Passed in 33.8 s against MySQL 8.4.11, including four `sslmode`
      subtests asserted against the server's own negotiated `Ssl_cipher`.
- [x] The post-run Podman connection had **no containers, no volumes, and no images**. The live test
      reported the machine's sparse disk file moving by +0.2 MiB end to end.
- [x] Cleanup verified on the failure path too: two live runs that failed an assertion (before the
      reserved-word fixes) still tore down every resource and reclaimed disk.
- [x] `GOOS=windows go build ./...` passes; `-race` clean on the harness package.
- [x] `gofmt`, scoped `go vet`, focused MySQL/bundle/command-runner/engine tests, relevant
      connector-catalog CLI tests, and `go build ./cmd/pm`.
- [x] Full `go test ./internal/connectors/...` regression passed after the bundle-count update;
      `pnpm --dir website run test:scripts` passed after regenerating the website connector data.
- [x] `make` non-suite gates individually: tidy, lint, docs, smoke, agent contract, connector
      validation, surface sync, connector boundary, and release workflow. `tidy-check` passed on
      the clean post-commit module files.
- [x] Inline `verify-work` and code-review evidence recorded in `UAT.md` and `REVIEW.md`; compatible
      isolated GSD workers are unavailable/forbidden by the single-worker contract.
- [x] Committed on `fm/cli-database-container-test-harness-r1` and reran the clean-tree
      `make tidy-check` gate.

The saved Docker evidence exercised the native implementation directly. The outer test phase must
replay it after the final native registration, lifecycle reconciliation, deterministic pagination,
and row-ordinal changes.
