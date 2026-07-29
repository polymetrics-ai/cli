# TDD ledger — Homebrew release notification

## Red / baseline

Baseline facts before production edits:

- `.github/workflows/release.yml` has `release-please`, `package-check`, and `release-assets`; it has no Homebrew notification job after `release-assets`.
- `.github/workflows/website.yml` is independent and scoped to `website/**`, its own workflow file, and manual website dispatch.
- The merged tap contract exposes `PM formula update` at `.github/workflows/pm-formula-update.yml` with inputs `dispatch_schema`, `source_repo`, `tag`, `release_id`, `source_run_id`, `target_commitish_policy`, and `dry_run`.
- Tap PR #7's approved App identity evidence is `pm-homebrew-pr-bot` with secret names `PM_HOMEBREW_PR_APP_ID` and `PM_HOMEBREW_PR_PRIVATE_KEY`.
- `gh-axi secret list -R polymetrics-ai/cli` showed no visible configured repo secrets; missing authentication must therefore fail explicitly.

Planned red checks:

- Static workflow assertions should fail before implementation because `notify-homebrew-tap` and the helper script are absent.
- Functional helper tests should fail before implementation because there is no notification helper to validate dry-run inputs or reject unsafe auth/mutation paths.

## Green

Pending.

## Refactor / hardening

Pending.

## Verification evidence

Pending.

## Skills

`gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-documentation`, `golang-continuous-integration`.
