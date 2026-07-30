# Homebrew release notification

Branch: `ci/homebrew-release-notification`
Target: `main`
Task: CLI-side dry-run notification to the merged Homebrew tap PM formula workflow after verified PM binary releases.

## GSD path

- `scripts/gsd doctor` passed on 2026-07-29.
- `scripts/gsd list` passed and did not expose a `programming-loop` command.
- `scripts/gsd prompt programming-loop init --phase homebrew-release-notification --dry-run` returned `unknown GSD command: programming-loop`; documented manual-GSD fallback is active for this phase.
- Required skills loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-documentation`, and `golang-continuous-integration`.

## Authoritative evidence read

- CLI release workflow: `.github/workflows/release.yml`.
- Website workflow isolation: `.github/workflows/website.yml`.
- Release asset/trust contract: `.goreleaser.yaml`, `scripts/verify-release-assets.sh`, and `docs/release-verification.md`.
- Repository conventions: `AGENTS.md` and `.planning/codebase/CONVENTIONS.md`.
- Merged tap contract from `polymetrics-ai/homebrew-tap` PR #7 and current workflow listing:
  - workflow path/name: `.github/workflows/pm-formula-update.yml` / `PM formula update`;
  - workflow inputs: `dispatch_schema`, `source_repo`, `tag`, `release_id`, `source_run_id`, `target_commitish_policy`, and `dry_run`;
  - approved schema/source/defaults: `pm-homebrew-formula/v1`, `polymetrics-ai/cli`, `target_commitish_policy=ignore`, `dry_run="true"`;
  - dry-run false is the tap's branch/PR mutation path and is out of scope for this initial rollout;
  - App identity evidence in the tap workflow: `pm-homebrew-pr-bot`, secret names `PM_HOMEBREW_PR_APP_ID` and `PM_HOMEBREW_PR_PRIVATE_KEY`.
- Secret-name evidence: `gh-axi secret list -R polymetrics-ai/cli` returned no visible configured repo secrets; implementation must fail explicitly when required App secrets are absent rather than falling back to PATs, `GITHUB_TOKEN`, or ambient credentials.

## Behavior contract

- PM release and website deployment stay independent; `.github/workflows/website.yml` must not mention or dispatch Homebrew automation.
- Notification runs only from the PM release workflow, only after `release-assets` succeeds and marks verified release assets ready.
- Top-level workflow permissions remain `contents: read`; the notification job grants only read permissions to the ordinary `GITHUB_TOKEN` and does not expose it as `GH_TOKEN`/`GITHUB_TOKEN`.
- The notification helper mints a GitHub App installation token from the approved App identity, validates the App slug and requested token permissions, scopes the token to `polymetrics-ai/homebrew-tap`, and passes it explicitly to GitHub's workflow-dispatch API.
- The only remote workflow call is `polymetrics-ai/homebrew-tap/.github/workflows/pm-formula-update.yml` on `main` with schema `pm-homebrew-formula/v1`, source repo `polymetrics-ai/cli`, strict stable `vX.Y.Z` tag, release id, source run id, `target_commitish_policy=ignore`, and `dry_run=true`.
- The initial rollout forbids `dry_run=false`; live branch/PR mutation requires a later code/documentation activation change.
- Duplicate release events are serialized by tag and dispatch identical dry-run requests; retries fail closed on malformed input, missing App credentials, wrong App/repository/workflow, wrong token permissions, or failed upstream verification.

## Implementation slices

1. **Planning checkpoint:** create this plan, TDD ledger, and verification checklist before production edits.
2. **Red assertions:** add focused static/functional tests that fail on the current release workflow because no Homebrew notification helper/job exists.
3. **Notification helper:** add a dependency-free script that validates inputs, rejects ambient credentials/live mutation, mints the approved GitHub App token, verifies the target workflow, and dispatches only the dry-run tap contract.
4. **Release workflow integration:** add release-asset outputs and a `notify-homebrew-tap` job that runs after verified release assets and uses the helper with explicit App secrets.
5. **Documentation:** update only release/operator docs needed to describe the dry-run contract and later activation boundary.
6. **Verification and commit:** run focused tests/static assertions, keep GSD artifacts current, commit the bounded implementation, then stop for firstmate validation.

## Explicit exclusions

- No live Homebrew formula branch or PR mutation.
- No manual or validation-time workflow dispatch to the tap.
- No PAT, personal account token, ambient `gh` credential, repository setting, or secret value access.
- No website deployment behavior changes.
- No connector packaging, broader release scope, Homebrew bottles/casks/notarization/signing, or deployment changes.
