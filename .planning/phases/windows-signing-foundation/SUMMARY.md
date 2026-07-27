# Summary — Windows signing foundation

Status: Implemented locally; CI repair for PR #559 in progress.

This phase implements a provider-inert Windows signing foundation PR for #554/#555 using the completed scout report as design evidence. It does not apply to SignPath, configure signing secrets, sign artifacts, publish releases, or submit WinGet manifests.

Implemented:

- Public Windows code-signing/privacy policy linked from README and SECURITY.
- Deterministic Windows VERSIONINFO RC and `.syso` generator.
- WiX MSI source for machine-scope x64/arm64 installers under `%ProgramFiles%\Polymetrics\CLI` with machine PATH integration and stable per-architecture UpgradeCode values.
- PR-only unsigned Windows package validation workflow with no secrets/provider calls.
- WinGet templates/docs/tests for `PolymetricsAI.PolymetricsCLI` with placeholder hashes only.

CI repair:

- Updated `google.golang.org/grpc` to v1.82.1 to clear reachable GO-2026-6061 under CI's Go 1.25.12 `govulncheck`.
- Fixed PR issue guard parsing so non-closing narrative references like `issues #554 and #555` and `reference #550/#554/#555` satisfy the linked-issue gate without closing those issues.
- Replaced the Windows SDK `cvtres.exe` VERSIONINFO `.syso` path with direct Go COFF generation after CI showed Go rejecting the `cvtres.exe` object with `sectnum < 0!`.

Local verification passed for the CI repair checks, including CI-equivalent `govulncheck`, issue-guard tests, actual PR #559 title/body validation, focused Windows packaging Go tests, direct `.syso` Windows cross-link checks for amd64/arm64, full `go test -timeout 20m ./...`, `go vet ./...`, `go build ./cmd/pm`, `git diff --check`, and `make verify`. Full unsigned MSI build/install verification remains on the Windows PR workflow.
