# Review — CI runner routing

## Scope review

- All Linux jobs now depend on and consume the one reusable selector output;
  there are 20 such consumer jobs.
- The selector itself is the only remaining Linux `ubuntu-latest` job. It is
  routing policy only; GitHub organization controls must secure fork workflow
  execution before any self-hosted runner is trusted.
- `windows-package-check.yml` was not altered.
- The `website.yml` deploy job now depends on and consumes the shared selector
  output with the website checks and image jobs.

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
- Website deployment retains its `main` plus feature-flag guard and resolves
  to the hosted fallback for its non-PR triggers.
- No token, runner registration key, credential, server path, or deployment
  configuration is added.

## Disposition

No actionable local review finding remains. The PR body carries the known Go
toolchain provisioning dependency and the required ephemeral-runner,
third-party-action SHA pinning, and runner-capacity hardening follow-ups.
