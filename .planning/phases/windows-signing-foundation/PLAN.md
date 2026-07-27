# Windows signing foundation (#554/#555)

Branch: `build/windows-signing-foundation`
Target: `main`
Issues: Refs #550, #554, #555. Does not close them.

## GSD path

- `scripts/gsd doctor` passed earlier in this session.
- Required command attempted: `scripts/gsd prompt programming-loop init --phase issue-554-555 --dry-run`.
- Adapter result: `scripts/gsd: unknown GSD command: programming-loop`.
- Manual-GSD fallback is active per `.agents/agentic-delivery/references/gsd-pi-adapter.md` and `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`.
- Design evidence: `/Users/karthiksivadas/karthik-agent-workspace/data/cli-windows-signpath-onboarding-r1/report.md`.
- Promotion instructions: `/Users/karthiksivadas/karthik-agent-workspace/data/cli-windows-signpath-onboarding-r1/promotion-instructions.md`.

## Required skills loaded

- `gsd-core`
- `golang-how-to`
- `golang-cli`
- `golang-testing`
- `golang-security`
- `golang-safety`
- `golang-lint`
- `golang-documentation`
- `golang-continuous-integration`
- `context-mode`
- `no-mistakes`

## Objective

Create one provider-inert foundation PR for Windows signing readiness without applying to SignPath, creating provider configuration, configuring secrets, signing artifacts, publishing release assets, or submitting WinGet manifests.

## Approved defaults from promotion

- Provider route default: SignPath Foundation, with Authenticode publisher display caveat.
- WinGet identifier: `PolymetricsAI.PolymetricsCLI`.
- MSI scope/path: machine-wide `%ProgramFiles%\Polymetrics\CLI` with machine `PATH` integration.
- arm64 WinGet gate: package structure now; require native Windows ARM64 install/run/uninstall before first arm64 WinGet publication.
- Public privacy/code-signing policy language must be truthful and captain-approved before provider submission.

## Implementation slices

1. **Planning checkpoint**
   - Create PLAN, TDD-LEDGER, VERIFICATION, PROMPTS, SUMMARY, and RUN-STATE before production edits.
   - Record manual-GSD fallback and local-critical-path spawn decision.
2. **Red/baseline checkpoint**
   - Add focused tests/validation for Windows version metadata normalization and WinGet identifier templates before implementation.
   - Capture missing current code-signing policy/versioninfo/WiX surfaces as baseline evidence.
3. **Policy/docs slice**
   - Add `docs/security/code-signing-policy.md` with SignPath-required attribution, publisher-display caveat, repository roles, release approvals, HSM custody, timestamping, no unsigned fallback, incident/revocation, and user-directed network behavior.
   - Link from `SECURITY.md` and README release/download section only; avoid website/generated churn.
4. **Windows VERSIONINFO slice**
   - Add deterministic Go RC generator for `pm.exe` VERSIONINFO metadata with version normalization.
   - Add Windows PowerShell script to compile RC to arch-specific `.syso` using Windows SDK `rc.exe`/`cvtres.exe` in CI.
   - Do not commit generated `.syso`, `.res`, `.msi`, or signed/unsigned release artifacts.
5. **WiX MSI slice**
   - Add WiX source/configuration for x64 and arm64 machine-scope installers under `%ProgramFiles%\Polymetrics\CLI`.
   - Include machine PATH integration, per-architecture stable UpgradeCode values, major upgrade/downgrade behavior, and clean uninstall authoring.
6. **PR-safe packaging validation slice**
   - Add a separate PR-only GitHub Actions workflow that builds unsigned Windows snapshot exes/MSIs, verifies VERSIONINFO and MSI structure, installs/runs/uninstalls x64 on `windows-latest`, and validates arm64 package structure without native install.
   - Ensure workflow never calls SignPath or any signing provider and has no secrets.
7. **WinGet docs/test slice**
   - Add manifest templates/checklist for `PolymetricsAI.PolymetricsCLI` with placeholders only, no unpublished hashes and no external PR submission.
   - Add tests to guard the approved ID and signed-installer hash placeholder behavior.
8. **Verification and delivery**
   - Run focused tests, gofmt, go test, go vet/build as practical, and no-mistakes full PR path.
   - Push branch and open PR title `build(windows): prepare signed MSI releases`.

## Security boundaries

- No SignPath application, provider account/project/policy, GitHub App installation, identity evidence, certificate, protected environment, secret, or production signing call.
- No PFX/private key, self-signed certificate, fake trusted signature, or unsigned fallback language.
- Do not modify `.github/workflows/release.yml`, `.goreleaser.yaml`, or `scripts/verify-release-assets.sh` in this PR unless unavoidable; current plan avoids them.
- Do not modify or replace `v0.1.0` assets.
- No WinGet external PR or hash generation for unpublished installers.

## Commit checkpoints

1. Planning artifacts.
2. Red tests/baseline validation.
3. Documentation/policy.
4. Versioninfo + tests.
5. WiX/scripts/workflow + validation.
6. WinGet templates/docs/tests.
7. Review/no-mistakes fixes if any.
