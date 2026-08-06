# Review — CI runner routing

## Scope review

- All prior Linux `ubuntu-latest` jobs now depend on and consume the one
  reusable selector output; there are 19 such consumer jobs.
- The selector itself is the only remaining Linux `ubuntu-latest` job. It is
  intentionally hosted so untrusted event metadata is evaluated before any
  self-hosted job is requested.
- `windows-package-check.yml` was not altered.
- The existing `website.yml` deploy job remains on its dedicated label. Only
  the non-deploy website checks/image jobs adopt shared routing.

## Safety review

- The self-hosted branch requires both the same-repository structural condition
  and exactly `karthik-sivadas` or `alfred-polymetrics-ai`; otherwise it
  returns GitHub-hosted `ubuntu-latest`.
- Push, tag, release, schedule, issue-comment, and workflow-dispatch events
  lack a qualifying `pull_request` event and therefore use the hosted fallback.
- No token, runner registration key, credential, server path, or deployment
  configuration is added.

## Disposition

No actionable local review finding remains. The PR body carries the known Go
toolchain provisioning dependency and the required ephemeral-runner,
third-party-action SHA pinning, and runner-capacity hardening follow-ups.
