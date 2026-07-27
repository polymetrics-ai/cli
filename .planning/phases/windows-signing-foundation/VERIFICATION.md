# Verification Checklist — Windows signing foundation

## Local required checks

- [x] `go test ./build/windowsversion ./packaging/windows ./packaging/windows/winget`
- [x] `go test -count=1 ./cmd/prissueguard ./internal/coordination/issueguard`
- [x] PR #559 title/body through `go run ./cmd/prissueguard` — passed with 3 linked issues.
- [x] `GOTOOLCHAIN=go1.25.12 go run golang.org/x/vuln/cmd/govulncheck@latest ./...` — passed after the fixed `grpc` update.
- [x] `GOTOOLCHAIN=go1.25.12 go test -count=1 ./cmd/prissueguard ./internal/coordination/issueguard ./build/windowsversion ./packaging/windows ./packaging/windows/winget ./internal/runtimecheck ./internal/worker ./internal/connectors/native/postgres`
- [x] `GOTOOLCHAIN=go1.25.12 go vet ./cmd/prissueguard ./internal/coordination/issueguard ./build/windowsversion ./packaging/windows ./packaging/windows/winget ./internal/runtimecheck ./internal/worker ./internal/connectors/native/postgres`
- [x] `GOTOOLCHAIN=go1.25.12 go build ./cmd/pm`
- [x] `gofmt -w build/windowsversion packaging/windows`
- [x] `gofmt -w internal/coordination/issueguard/guard.go internal/coordination/issueguard/guard_test.go`
- [x] `go test ./...` — historical implementation run passed before the CI repair slice.
- [ ] `go test -timeout=4m ./...` — repair rerun timed out in existing `internal/cli` and `internal/connectors/certify` broad connector-bundle tests; not in modified/fix-adjacent packages.
- [x] `go vet ./...` — passed locally after implementation.
- [x] `go build ./cmd/pm` — passed locally after implementation.
- [x] `git diff --check` — passed locally after implementation.

## Windows PR-safe workflow checks

- [x] Workflow is separate from production release workflow and runs only unsigned validation.
- [x] Workflow has no secrets and no provider API calls.
- [x] Builds Windows amd64 and arm64 `pm.exe` snapshots with generated VERSIONINFO.
- [x] Builds WiX x64 and arm64 MSIs.
- [x] Verifies VERSIONINFO fields for both executables.
- [x] Verifies MSI metadata/structure for x64 and arm64.
- [x] Installs/runs/uninstalls x64 MSI on `windows-latest`.
- [x] Records arm64 native install/run/uninstall as deferred until a native Windows ARM64 runner is green.

## Documentation checks

- [x] Code-signing policy is truthful about current unsigned state and future SignPath route.
- [x] Policy includes SignPath attribution and publisher-display caveat.
- [x] Policy documents HSM custody, release-only approvals, timestamping, no unsigned fallback, incident/revocation, and PM network behavior.
- [x] README download section links policy.
- [x] SECURITY.md links policy.
- [x] WinGet docs/templates use `PolymetricsAI.PolymetricsCLI` and placeholders only.

## no-mistakes PR path

- [ ] Work committed on `build/windows-signing-foundation`.
- [ ] `no-mistakes axi` home checked before run.
- [ ] `no-mistakes axi run --intent "..."` driven to `checks-passed` or terminal blocker.
- [ ] PR exists with title `build(windows): prepare signed MSI releases`.
- [ ] PR body references #550, #554, #555 without closing them.
- [ ] Never merge.

## Results

Local focused verification for the CI repair passed. The actual unsigned MSI build/install path still requires the Windows PR workflow because this host is not a Windows runner with PowerShell, Windows SDK, Visual Studio Build Tools, WiX, and `msiexec.exe`.
