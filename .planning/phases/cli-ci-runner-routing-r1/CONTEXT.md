# Context — CI runner routing

## Fixed decisions

- PR code may use the self-hosted `polymetrics-website` label only when both
  `github.event.pull_request.head.repo.full_name == github.repository` and the
  author login is `karthik-sivadas` or `alfred-polymetrics-ai`.
- Fork PRs and every non-PR event deliberately resolve to `ubuntu-latest`.
- `windows-latest` remains GitHub-hosted. The existing website deployment job
  retains its dedicated self-hosted label and is not part of general routing.
- The decision lives in one reusable workflow; callers only consume its output.

## Constraints and follow-ups

- The known online self-hosted runner has `polymetrics-website`; no server or
  deployment files are in scope.
- Go tooling must be provisioned on that runner before this routing can serve
  Go jobs. The shipping PR must state this dependency and the required
  ephemeral-runner, action-SHA-pinning, and concurrency hardening follow-ups.
