# Verification — CI toolchain and integration branch guard

## Automated checks

- [x] **Red:** `GOTOOLCHAIN=go1.25.12 go run golang.org/x/vuln/cmd/govulncheck@latest ./...` reported seven reachable standard-library vulnerabilities: GO-2026-6218, GO-2026-6091, GO-2026-6090, GO-2026-6089, GO-2026-6088, GO-2026-5972, and GO-2026-5026. Every report named `Fixed in: ...@go1.25.13`.
- [x] **Green:** `GOTOOLCHAIN=go1.25.13 go version` reported `go version go1.25.13 darwin/arm64`; the same scan reported `No vulnerabilities found.` and `Your code is affected by 0 vulnerabilities.`
- [x] All 11 active explicit workflow references and `go.mod` now match Go 1.25.13 (12 active pins total). The two `go-version-file: go.mod` uses in `claude-review.yml` inherit the same directive. The only remaining 1.25.12 mentions are historical planning records (and unrelated icon bytes), which are intentionally preserved.
- [x] The exact workflow shell passes `fm/cli-ci-toolchain-and-branchname-r1`, `integration/4015-mvp-flat-r1`, and `feat/github-connector`; it rejects `integration/0-mvp-flat`, `integration/abc-mvp-flat`, `integration/4015-Invalid`, and `proposal/mvp`.
- [x] Ruby Psych parsed every changed workflow YAML. `make lint`, `make agent-contract-check`, `make release-workflow-check`, `make docs-check-no-build`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make connector-boundary`, `make github-parity-artifacts-check`, `make connectorgen-certification-matrix`, and `make connector-canon-check` passed.
- [x] `GOTOOLCHAIN=go1.25.13 go test -timeout 20m ./cmd/prissueguard` and `GOTOOLCHAIN=go1.25.13 go build ./cmd/pm` passed.
- [x] `git diff --check` passes.

`GOTOOLCHAIN=go1.25.13 make tidy-check` intentionally exits after showing the requested `go.mod` toolchain diff against `HEAD`; its preceding `go mod tidy` made no further `go.mod` or `go.sum` changes.

## Manual review focus

- [x] No `govulncheck` suppression, exemption, skip, or non-blocking conversion.
- [x] No special-case branch name and no relaxation of existing branch classes: existing exception and conventional patterns are unchanged; the new integration arm is separately anchored.
- [x] Historical evidence remains unchanged.
