# Context — CI runner routing

## Fixed decisions

- PR code may use the self-hosted `polymetrics-cli` label only when both
  `github.event.pull_request.head.repo.full_name == github.repository` and the
  author login is `karthik-sivadas` or `alfred-polymetrics-ai`.
- Fork PRs and every non-PR event deliberately resolve to `ubuntu-latest`.
- `windows-latest` remains GitHub-hosted. Website deployment consumes the
  shared selector and therefore resolves to `ubuntu-latest` for its non-PR
  triggers.
- The decision lives in one reusable workflow; callers only consume its output.
- The workflow condition is routing policy rather than a durable fork-security
  boundary. The required GitHub approval and runner-group controls are in
  `docs/security/self-hosted-ci-runner-policy.md`; routing is unsafe until an
  organization owner applies them.

## Constraints and follow-ups

- Two known online self-hosted runners carry `polymetrics-cli` with Go 1.25.12
  installed; no server or deployment files are in scope.
- The shipping PR must record that the runner-provisioning dependency is now
  satisfied, along with the required ephemeral-runner, action-SHA-pinning, and
  concurrency hardening follow-ups.
