# Summary — Windows signing foundation

Status: Implemented locally; no-mistakes/PR CI pending.

This phase implements a provider-inert Windows signing foundation PR for #554/#555 using the completed scout report as design evidence. It does not apply to SignPath, configure signing secrets, sign artifacts, publish releases, or submit WinGet manifests.

Implemented:

- Public Windows code-signing/privacy policy linked from README and SECURITY.
- Deterministic Windows VERSIONINFO RC generator and Windows SDK compilation script.
- WiX MSI source for machine-scope x64/arm64 installers under `%ProgramFiles%\Polymetrics\CLI` with machine PATH integration and stable per-architecture UpgradeCode values.
- PR-only unsigned Windows package validation workflow with no secrets/provider calls.
- WinGet templates/docs/tests for `PolymetricsAI.PolymetricsCLI` with placeholder hashes only.

Local verification passed: focused tests, `go test ./...`, `go vet ./...`, `go build ./cmd/pm`, and `git diff --check`.
