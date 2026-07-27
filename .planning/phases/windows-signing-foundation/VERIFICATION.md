# Verification Checklist — Windows signing foundation

## Local required checks

- [ ] `go test ./build/windowsversion ./packaging/windows/winget`
- [ ] `gofmt -w build/windowsversion packaging/windows/winget`
- [ ] `go test ./...`
- [ ] `go vet ./...`
- [ ] `go build ./cmd/pm`
- [ ] `git diff --check`

## Windows PR-safe workflow checks

- [ ] Workflow is separate from production release workflow and runs only unsigned validation.
- [ ] Workflow has no secrets and no provider API calls.
- [ ] Builds Windows amd64 and arm64 `pm.exe` snapshots with generated VERSIONINFO.
- [ ] Builds WiX x64 and arm64 MSIs.
- [ ] Verifies VERSIONINFO fields for both executables.
- [ ] Verifies MSI metadata/structure for x64 and arm64.
- [ ] Installs/runs/uninstalls x64 MSI on `windows-latest`.
- [ ] Records arm64 native install/run/uninstall as deferred until a native Windows ARM64 runner is green.

## Documentation checks

- [ ] Code-signing policy is truthful about current unsigned state and future SignPath route.
- [ ] Policy includes SignPath attribution and publisher-display caveat.
- [ ] Policy documents HSM custody, release-only approvals, timestamping, no unsigned fallback, incident/revocation, and PM network behavior.
- [ ] README download section links policy.
- [ ] SECURITY.md links policy.
- [ ] WinGet docs/templates use `PolymetricsAI.PolymetricsCLI` and placeholders only.

## no-mistakes PR path

- [ ] Work committed on `build/windows-signing-foundation`.
- [ ] `no-mistakes axi` home checked before run.
- [ ] `no-mistakes axi run --intent "..."` driven to `checks-passed` or terminal blocker.
- [ ] PR exists with title `build(windows): prepare signed MSI releases`.
- [ ] PR body references #550, #554, #555 without closing them.
- [ ] Never merge.

## Results

Pending implementation.
