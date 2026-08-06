# Verification — MySQL container harness R1

## Completed evidence

- [x] Focused unit tests for the harness, MySQL native connector, bundle loader, and changefeed contract.
- [x] `go mod verify`; `govulncheck -show verbose ./internal/connectors/native/mysql` is clean after
      upgrading the transitive `github.com/klauspost/compress` to v1.18.7.
- [x] Documented tagged MySQL Docker/Colima test on the explicit `colima` context, with the opt-in
      Colima reset enabled. It passed in 53.48 seconds against MySQL 8.4.11.
- [x] The post-run Docker context had no containers, volumes, or images. The live test reported
      before=84802707456 and after=84784943104 free bytes with `colima_reset=true`, a 17.8 MB
      decrease that is within the test's 128 MB ordinary-build noise allowance.
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
