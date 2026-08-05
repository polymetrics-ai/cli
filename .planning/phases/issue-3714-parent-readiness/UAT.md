# UAT — issue #3714 parent readiness

Automated acceptance evidence:

- The parent contains current `origin/main` (`git merge-base --is-ancestor origin/main HEAD`).
- Canonical projected worker files are current (`go run ./cmd/agentcontractgen check`).
- Focused tests and scoped validation gates passed as listed in `VERIFICATION.md`.

Human-only acceptance boundary:

- The captain reviews PR #3723 and alone decides whether to merge it. This phase may mark the PR
  ready only after remote checks and automated review coverage are green; it never merges the PR.
