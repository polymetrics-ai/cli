# TDD ledger — Homebrew release notification

## Red / baseline

Baseline facts before production edits:

- `.github/workflows/release.yml` has `release-please`, `package-check`, and `release-assets`; it has no Homebrew notification job after `release-assets`.
- `.github/workflows/website.yml` is independent and scoped to `website/**`, its own workflow file, and manual website dispatch.
- The merged tap contract exposes `PM formula update` at `.github/workflows/pm-formula-update.yml` with inputs `dispatch_schema`, `source_repo`, `tag`, `release_id`, `source_run_id`, `target_commitish_policy`, and `dry_run`.
- Tap PR #7's approved App identity evidence is `pm-homebrew-pr-bot` with secret names `PM_HOMEBREW_PR_APP_ID` and `PM_HOMEBREW_PR_PRIVATE_KEY`.
- `gh-axi secret list -R polymetrics-ai/cli` showed no visible configured repo secrets; missing authentication must therefore fail explicitly.

Red checks:

- `scripts/tests/homebrew-release-notify.sh` exited 1 before implementation with `notification helper is missing`, proving the new static/functional gate fails on the current release workflow because there is no helper/job path.
- The test file now covers the planned behavior surfaces: authorized dry-run dispatch, missing/ambient credentials, wrong upstream verification ordering, duplicate dry-run payloads, malformed tags, wrong repository/workflow, website prohibition, and dry-run-only rollout.

## Green

- Added `scripts/notify-homebrew-formula-update.sh`, a dependency-free shell helper using `curl`, `openssl`, and `python3` to validate the release notification contract, reject ambient credentials, mint an explicit `pm-homebrew-pr-bot` installation token, verify the active tap workflow, and send only the `dry_run=true` `workflow_dispatch` payload.
- Added `notify-homebrew-tap` to `.github/workflows/release.yml` after `release-assets`; the job needs the verified release-assets output, keeps ordinary `GITHUB_TOKEN` permissions at `contents: read`, uses only `PM_HOMEBREW_PR_APP_ID` / `PM_HOMEBREW_PR_PRIVATE_KEY`, serializes duplicate tag notifications, and passes the exact tap workflow inputs.
- Added release metadata outputs (`tag_name`, `release_version`, `release_id`, `source_run_id`, `homebrew_notification_ready`) so notification is gated on a successfully verified release-assets job.
- Added `scripts/tests/homebrew-release-notify.sh` and `make release-workflow-check` to cover authorized dry-run dispatch against a fake GitHub API, missing App credentials, ambient credential refusal, upstream verification ordering, duplicate dry-run payload determinism, malformed tags, wrong repo/workflow, website prohibition, and live-mutation refusal.
- Updated `docs/release-verification.md` with the dry-run notification contract and later activation boundary.

## Refactor / hardening

- The helper never writes the installation token to `GITHUB_OUTPUT`, never invokes `gh`, and fails if `GITHUB_TOKEN` or `GH_TOKEN` is present.
- No-mistakes review flagged that `notify-homebrew-tap` must not check out the release tag before exposing App secrets; the checkout now pins to `${{ github.workflow_sha }}` and the static assertions require that trusted workflow-source commit.
- The initial rollout has no `dry_run=false` path; activation requires a later code and operator-doc change.

## Verification evidence

- `scripts/tests/homebrew-release-notify.sh` passed.
- `make release-workflow-check` passed.
- `shellcheck scripts/notify-homebrew-formula-update.sh scripts/tests/homebrew-release-notify.sh` passed.
- `git diff --check` passed.
- `ruby -e 'require "yaml"; YAML.load_file(".github/workflows/release.yml"); YAML.load_file(".github/workflows/website.yml"); puts "yaml syntax ok"'` passed.

## Skills

`gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-documentation`, `golang-continuous-integration`.
