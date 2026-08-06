# TDD ledger — CI runner routing

| Requirement | Red evidence | Green evidence |
| --- | --- | --- |
| One fail-closed shared selector routes every eligible Linux job | `./scripts/tests/verify-ci-runner-routing.sh` failed before production workflow edits because `.github/workflows/runner-selection.yml` did not exist. | The same test passed after the shared reusable selector and all 19 Linux consumer jobs were wired. It asserts the structural same-repository condition, both approved logins, hosted fallback, and no remaining direct Linux hosted-job label. |
| Windows and dedicated website deployment remain on their existing runners | The source contract was written before workflow routing existed. | The green contract asserts `windows-latest` remains outside selector consumption and `website.yml` retains `[self-hosted, linux, tailscale, polymetrics-website]` for deploy. |
| The PR handoff records provisioning and hardening follow-ups | N/A: documentation is created with the implementation slice. | `PR-BODY.md` records the Go provisioning dependency and the three required hardening follow-ups without exposing credentials. |
