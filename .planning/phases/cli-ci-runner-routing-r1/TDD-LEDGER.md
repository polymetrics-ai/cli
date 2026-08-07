# TDD ledger — CI runner routing

| Requirement | Red evidence | Green evidence |
| --- | --- | --- |
| One fail-closed shared selector routes every eligible Linux job | `./scripts/tests/verify-ci-runner-routing.sh` failed before production workflow edits because `.github/workflows/runner-selection.yml` did not exist. | The same test passed after the shared reusable selector and all 20 Linux consumer jobs were wired. It asserts the structural same-repository condition, both approved logins, hosted fallback, and no remaining direct Linux hosted-job label. |
| Windows remains separate and website deployment consumes shared routing | The source contract was written before workflow routing existed. | The green contract asserts `windows-latest` remains outside selector consumption and `website.yml` deployment depends on and consumes the selector output without a self-hosted runner label. |
| Fork routing has an external security boundary and Claude ignores ineligible comments | N/A: GitHub organization settings cannot be changed from this source slice. | The green contract requires the persistent fork-approval and runner-group runbook, documents that routing is unsafe until applied, and checks that Claude's selector repeats the auto/on-demand review eligibility conditions. |
| The PR handoff records provisioning and hardening follow-ups | N/A: documentation is created with the implementation slice. | `PR-BODY.md` records the Go provisioning dependency and the three required hardening follow-ups without exposing credentials. |
