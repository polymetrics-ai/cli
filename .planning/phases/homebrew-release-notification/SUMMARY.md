# Implementation done — Homebrew release notification

## Scope completed

- Added CLI release workflow dry-run notification to the merged tap workflow only after `release-assets` verifies a PM release.
- Added `scripts/notify-homebrew-formula-update.sh` to validate inputs, reject ambient credentials/live mutation, mint the approved `pm-homebrew-pr-bot` App token, verify the tap workflow, and dispatch only `dry_run=true`.
- Added focused workflow/helper assertions under `scripts/tests/homebrew-release-notify.sh` and `make release-workflow-check`.
- Updated `docs/release-verification.md` with the dry-run notification contract and later activation boundary.

## Contract evidence

- Homebrew tap contract: `polymetrics-ai/homebrew-tap/.github/workflows/pm-formula-update.yml` with schema `pm-homebrew-formula/v1` and inputs `source_repo`, `tag`, `release_id`, `source_run_id`, `target_commitish_policy`, and `dry_run`.
- App identity/secret-name evidence from tap PR #7: `pm-homebrew-pr-bot`, `PM_HOMEBREW_PR_APP_ID`, `PM_HOMEBREW_PR_PRIVATE_KEY`.
- Initial rollout remains dry-run only; `dry_run=false` is rejected in the helper and absent from the release notification job.

## Verification run locally

- `scripts/tests/homebrew-release-notify.sh` — passed.
- `make release-workflow-check` — passed.
- `git diff --check` — passed.
- `shellcheck scripts/notify-homebrew-formula-update.sh scripts/tests/homebrew-release-notify.sh` — passed.
- Ruby YAML syntax load for `release.yml` and `website.yml` — passed.

## Not done / forbidden by task

- Did not dispatch the Homebrew workflow.
- Did not create a formula branch or PR.
- Did not publish a PM release.
- Did not access secret values or change repository/App settings.
- Did not change website deployment behavior.
