## Summary

- Route same-repository PRs by `karthik-sivadas` and
  `alfred-polymetrics-ai` to the existing `polymetrics-cli` self-hosted
  runner label through one shared selector.
- Route fork PRs and all non-PR triggers to GitHub-hosted `ubuntu-latest`.
- Keep Windows on `windows-latest` and route website deployment through the
  shared selector.

## Required GitHub configuration before enabling routing

The selector alone is not a security boundary against a fork-controlled
workflow. CI routing is unsafe until an organization owner applies and verifies
the fork-approval and runner-group controls in
[`docs/security/self-hosted-ci-runner-policy.md`](../../../docs/security/self-hosted-ci-runner-policy.md).

## Runner provisioning

Two online `polymetrics-cli` runners now provide Go 1.25.12, so the prior
runner-provisioning dependency is satisfied. This PR does not touch the server.

## Required hardening follow-ups (not implemented here)

1. Make runners ephemeral so job state cannot persist between runs.
2. Require third-party actions to be pinned by immutable SHA. A compromised
   action executes on the runner independently of the PR-author gate.
3. Add runner capacity: one runner serializes a repository with roughly twenty
   jobs per PR.

## Verification

- `./scripts/tests/verify-ci-runner-routing.sh`
- YAML parse check for `.github/workflows/*.yml`
- `make release-workflow-check`
- `go run ./cmd/agentcontractgen check`

## Delivery notes

- GSD prompts for discuss, TDD planning, execute, verify, and code review were
  generated with `scripts/gsd prompt ...` and completed through the documented
  inline/manual fallback; role spawning is prohibited by the worker brief.
- No Go, CLI, docs, or website UI behavior changes are included.
- No credentials, registration keys, server access, or deployment changes are
  included.
