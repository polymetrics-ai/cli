# TDD Ledger — Windows signing foundation

## Rules

- Start with red tests or explicit validation artifacts before production edits.
- Do not call signing providers or require secrets in tests.
- Unsigned PR snapshots are validation artifacts only and are never releasable.

## Red / baseline evidence

| Slice | Red/baseline command | Expected red/baseline | Status |
|---|---|---|---|
| Current policy surface | `test -f docs/security/code-signing-policy.md` | Missing before docs slice | Pending |
| Version normalization | `go test ./build/windowsversion ./packaging/windows/winget` after adding tests before implementation | Failed: undefined `Version`, `NormalizeVersion`, `RenderRC` | Red captured |
| WinGet ID templates | `go test ./build/windowsversion ./packaging/windows/winget` after adding tests before templates | Failed: manifest templates missing | Red captured |
| Windows package CI | PR workflow dry-run on GitHub | Fails if unsigned build/package/install validation regresses | Pending |
| CI govulncheck | `GOTOOLCHAIN=go1.25.12 go run golang.org/x/vuln/cmd/govulncheck@latest ./...` | Failed with GO-2026-6061 in reachable `google.golang.org/grpc` v1.79.3; fixed in v1.82.1 | Red captured |
| PR issue guard | PR #559 `require-linked-issue` with narrative body references `issues #554 and #555` plus `reference #550/#554/#555` | Failed because only `Refs #N`/closing keywords and single issue tokens were recognized | Red captured from CI/PR metadata |
| Windows SDK VERSIONINFO object | PR #559 `unsigned-msi-snapshot` | Failed in `GOOS=windows GOARCH=amd64 go build` with `sectnum < 0!` after `cvtres.exe` generated `cmd\pm\pm_windows_amd64.syso` | Red captured from CI |
| Direct VERSIONINFO `.syso` linkability | `GOTOOLCHAIN=go1.25.12 go run ./build/windowsversion -version 0.0.0 -goarch amd64 -out cmd/pm/pm_windows_amd64.syso`, then `GOTOOLCHAIN=go1.25.12 GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build ... ./cmd/pm` | Must pass locally without Windows SDK tools | Red captured from CI; green captured locally |
| MSI ProductName verifier | PR #559 `unsigned-msi-snapshot`; then static verifier-normalization test | Failed with MSI `ProductName` got `' Polymetrics CLI '`, want `Polymetrics CLI` | Red captured |
| Snyk OpenTelemetry finding | PR #559 `security/snyk`; public Snyk advisory `SNYK-GOLANG-GOOPENTELEMETRYIOOTELPROPAGATION-17054905` | Branch introduced OpenTelemetry `v1.43.0`; fixed floor is `v1.44.0` | Red captured |

## Green evidence

| Slice | Command | Expected | Result |
|---|---|---|---|
| Version generator | `go test ./build/windowsversion` | PASS | Passed via focused package run |
| WinGet templates | `go test ./packaging/windows/winget` | PASS | Passed via focused package run |
| WiX source guard | `go test ./packaging/windows` | PASS | Passed via focused package run |
| Go tests | `go test ./build/windowsversion ./packaging/windows ./packaging/windows/winget` | PASS | Passed |
| Formatting | `gofmt -w build/windowsversion packaging/windows` | no diff after gofmt | Passed |
| Broad Go gates | `go test ./...`, `go vet ./...`, `go build ./cmd/pm` | PASS or documented blocker | Passed locally |
| CI govulncheck fix | `GOTOOLCHAIN=go1.25.12 go run golang.org/x/vuln/cmd/govulncheck@latest ./...` | PASS | Passed after `google.golang.org/grpc` v1.82.1 update |
| PR issue guard fix | `go test -count=1 ./cmd/prissueguard ./internal/coordination/issueguard`; PR #559 title/body through `go run ./cmd/prissueguard` | PASS | Passed; actual PR body reports 3 linked issues |
| Direct VERSIONINFO `.syso` fix | `GOTOOLCHAIN=go1.25.12 go test -count=1 ./build/windowsversion ./packaging/windows ./packaging/windows/winget ./cmd/prissueguard ./internal/coordination/issueguard` plus local Windows cross-builds for amd64 and arm64 with generated `.syso` files | PASS | Passed |
| CI repair vulnerability scan | `GOTOOLCHAIN=go1.25.12 go run golang.org/x/vuln/cmd/govulncheck@latest ./...` | PASS | Passed; no vulnerabilities found |
| CI repair vet/build | `GOTOOLCHAIN=go1.25.12 go vet ./build/windowsversion ./packaging/windows ./packaging/windows/winget ./cmd/prissueguard ./internal/coordination/issueguard`; `GOTOOLCHAIN=go1.25.12 go build ./cmd/pm` | PASS | Passed |
| CI repair package set | `GOTOOLCHAIN=go1.25.12 go test -count=1 ./cmd/prissueguard ./internal/coordination/issueguard ./build/windowsversion ./packaging/windows ./packaging/windows/winget ./internal/runtimecheck ./internal/worker ./internal/connectors/native/postgres` | PASS | Passed |
| Full suite rerun | `go test -timeout 20m ./...` | PASS | Passed |
| Full repo verification | `make verify` | PASS | Passed |
| MSI ProductName verifier fix | `GOTOOLCHAIN=go1.25.12 go test -count=1 ./packaging/windows` | PASS | Passed; verifier trims Windows Installer COM scalar padding for MSI properties while executable VERSIONINFO checks remain exact |
| Snyk OpenTelemetry floor | `GOTOOLCHAIN=go1.25.12 go list -m go.opentelemetry.io/otel go.opentelemetry.io/otel/metric go.opentelemetry.io/otel/sdk go.opentelemetry.io/otel/sdk/metric go.opentelemetry.io/otel/trace google.golang.org/grpc`; `GOTOOLCHAIN=go1.25.12 go run golang.org/x/vuln/cmd/govulncheck@latest ./...` | OpenTelemetry selected at `v1.44.0`; vulnerability scan clean | Passed |
| PR-safe Windows package workflow | GitHub Actions `Windows Package Check` | PASS | Pending on PR CI |
| no-mistakes | `no-mistakes axi run --intent ...` | `checks-passed` | Pending |

## Deferred production-signing evidence

- SignPath signing request IDs, Authenticode verification, RFC 3161 timestamp verification, release upload, and WinGet external PR validation are intentionally deferred to provider-accepted follow-up work.
