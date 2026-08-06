# Review — CI runner routing

## Scope review

- All prior Linux `ubuntu-latest` jobs now depend on and consume the one
  reusable selector output; there are 19 such consumer jobs.
- The selector itself is the only remaining Linux `ubuntu-latest` job. It is
  routing policy only; GitHub organization controls must secure fork workflow
  execution before any self-hosted runner is trusted.
- `windows-package-check.yml` was not altered.
- The existing `website.yml` deploy job remains on its dedicated label. Only
  the non-deploy website checks/image jobs adopt shared routing.

## Safety review

- The self-hosted branch requires both the same-repository structural condition
  and exactly `karthik-sivadas` or `alfred-polymetrics-ai`; otherwise it
  returns GitHub-hosted `ubuntu-latest`.
- Push, tag, release, schedule, issue-comment, and workflow-dispatch events
  lack a qualifying `pull_request` event and therefore use the hosted fallback.
- The selector does not run for an ineligible Claude review trigger, so
  unrelated comments do not consume a runner slot.
- `docs/security/self-hosted-ci-runner-policy.md` requires all-external-fork
  approval and a dedicated, repository-scoped CLI runner group. Routing remains
  unsafe until those GitHub-side controls are applied.
- Website deployment remains the documented trusted-ref non-PR exception and
  retains its original `main` plus feature-flag guard.
- No token, runner registration key, credential, server path, or deployment
  configuration is added.

## Disposition

No actionable local review finding remains. The PR body carries the known Go
toolchain provisioning dependency and the required ephemeral-runner,
third-party-action SHA pinning, and runner-capacity hardening follow-ups.
