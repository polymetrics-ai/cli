# PR 528 require-linked-issue repair

Branch: `fm/cli-release-and-connector-issues-r1`
PR: #528
Target: `main`
Primary issue reference: #67

## GSD path

- `scripts/gsd doctor` passed on 2026-07-26.
- `scripts/gsd prompt programming-loop init --phase pr-528-require-linked-issue --dry-run` returned `unknown GSD command: programming-loop`; the documented manual-GSD fallback is active.
- Required skills loaded: `github-issue-first-delivery`, `gsd-programming-loop`, `no-mistakes`, `golang-how-to`, `golang-testing`, and `golang-cli`.

## Objective

Repair the failed `require-linked-issue` CI gate on PR #528 without dropping any existing
no-mistakes fix commits, weakening the issue-first guard, cutting a release, deploying the website,
or merging the PR.

## Diagnosis

The repo-native `.github/workflows/pr-issue-guard.yml` job passes the PR title and body to
`cmd/prissueguard`. The current PR body includes prose references such as PR #16 and commit/issue
#29, but it has no explicit issue-first footer. A local run of the same guard against the live PR
body produced:

```text
issueguard: blocked
- PR body must reference an issue with Closes #123 for completed work or Refs #123 for stacked/incremental work
```

## Repair plan

1. Keep all existing commits on `fm/cli-release-and-connector-issues-r1`.
2. Add this bounded planning/TDD/verification trace for the CI repair.
3. Update PR #528's body with `Refs #67`, because issue #67 covers docs and website updates for
   GitHub parity, operation-model documentation, and agent learnings, and this PR is incremental
   recovery rather than completion of that issue.
4. Re-run `cmd/prissueguard` against the edited PR body.
5. Commit and push the planning trace only; do not merge PR #528.

## Human gates

- No merge to `main`.
- No release tag, GitHub release, prerelease, or website deployment.
- No guard weakening or branch exemption for `fm/*`.
- No secret access or credentialed connector checks.
