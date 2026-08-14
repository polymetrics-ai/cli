Refs #4097

## Intent

Unblock the Production MVP integration stack without weakening security scanning or branch-name enforcement.

## What changed

- Updated the active Go toolchain pin from 1.25.12 to the verified minimum clean patch, 1.25.13, in `go.mod`, `connector-boundary.yml`, `pr-issue-guard.yml`, all four `release.yml` uses, all three `security.yml` uses (including `GOTOOLCHAIN`), and both `verify.yml` uses.
- Kept `claude-review.yml` on `go-version-file: go.mod`, which now resolves to Go 1.25.13.
- Extended only the `branch-name` job in `.github/workflows/conventions.yml` with an anchored integration class requiring a positive issue number and the existing slug grammar.

## `branch-name` rule — `.github/workflows/conventions.yml`

Before:

```bash
case "$HEAD_REF" in
  dependabot/*|release-please--branches--*|fm/*|connector-architecture-v2)
    exit 0
    ;;
esac

pattern='^(feat|fix|docs|chore|ci|test|refactor|perf|build|release|revert|deps)/[a-z0-9][a-z0-9._-]*$'
if [[ ! "$HEAD_REF" =~ $pattern ]]; then
  {
    echo "Invalid branch name: $HEAD_REF"
    echo "Use <type>/<description>, for example feat/github-connector or fix/stripe-pagination."
    echo "Allowed types: feat, fix, docs, chore, ci, test, refactor, perf, build, release, revert, deps."
  } >&2
  exit 1
fi
```

After (the prior `case` and conventional pattern are verbatim):

```bash
case "$HEAD_REF" in
  dependabot/*|release-please--branches--*|fm/*|connector-architecture-v2)
    exit 0
    ;;
esac

integration_pattern='^integration/[1-9][0-9]*-[a-z0-9][a-z0-9._-]*$'
if [[ "$HEAD_REF" =~ $integration_pattern ]]; then
  exit 0
fi

pattern='^(feat|fix|docs|chore|ci|test|refactor|perf|build|release|revert|deps)/[a-z0-9][a-z0-9._-]*$'
if [[ ! "$HEAD_REF" =~ $pattern ]]; then
  {
    echo "Invalid branch name: $HEAD_REF"
    echo "Use <type>/<description>, for example feat/github-connector or fix/stripe-pagination."
    echo "Allowed types: feat, fix, docs, chore, ci, test, refactor, perf, build, release, revert, deps."
  } >&2
  exit 1
fi
```

The new class accepts `integration/4015-mvp-flat-r1`; it rejects zero, non-numeric, or missing issue numbers and invalid slugs. No named-branch exemption, existing exception, existing conventional class, or job behavior changed.

## Red / green evidence

**Red:** Go 1.25.12 produced seven reachable standard-library findings (GO-2026-6218, GO-2026-6091, GO-2026-6090, GO-2026-6089, GO-2026-6088, GO-2026-5972, and GO-2026-5026). Each reports `Fixed in: ...@go1.25.13`.

**Green — real `govulncheck` output under Go 1.25.13:**

```text
=== Symbol Results ===

No vulnerabilities found.

Your code is affected by 0 vulnerabilities.
This scan also found 1 vulnerability in packages you import and 0
vulnerabilities in modules you require, but your code doesn't appear to call
these vulnerabilities.
Use '-show verbose' for more details.
```

## Verification

- `GOTOOLCHAIN=go1.25.13 go version` → `go version go1.25.13 darwin/arm64`
- `GOTOOLCHAIN=go1.25.13 go run golang.org/x/vuln/cmd/govulncheck@latest ./...` → clean (output above)
- Exact workflow shell: passed this PR's `fm/cli-ci-toolchain-and-branchname-r1`, `integration/4015-mvp-flat-r1`, and `feat/github-connector`; rejected malformed integration and unsupported branches.
- `GOTOOLCHAIN=go1.25.13 go test -timeout 20m ./cmd/prissueguard`
- `GOTOOLCHAIN=go1.25.13 go build ./cmd/pm`
- Workflow YAML parsing plus `make lint`, `make agent-contract-check`, `make release-workflow-check`, `make docs-check-no-build`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make connector-boundary`, `make github-parity-artifacts-check`, `make connectorgen-certification-matrix`, and `make connector-canon-check`.
- `git diff --check`

`make tidy-check` ran `go mod tidy` cleanly; it then correctly detected this PR's intentional `go.mod` toolchain diff against `HEAD`.

## Delivery record

- GSD: resolved `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` with `scripts/gsd prompt`; manual-inline fallback recorded because this direct-PR slice is non-numbered and compatible isolated Pi workers are unavailable.
- Skills: `golang-how-to`, `golang-continuous-integration`, `golang-security`, `golang-lint`, `golang-testing`.
- No CLI surface/docs parity applies. No credentials, provider calls, dependencies, connector/application code, test changes, suppressions, or non-blocking checks were introduced.
- Commits: `498fc3a6a` (plan checkpoint), `9e3d60afb` (implementation and verification).

## Automated review

Primary route: `claude_auto` on PR creation by the trusted author. This is a stacked PR into `integration/4015-mvp-flat-r1`; parent-review fallback applies only if the automatic route does not produce review coverage.
