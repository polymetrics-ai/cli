# Plan — CI toolchain and integration branch guard

## Goal

Restore programme-wide CI by moving every active Go pin to the minimum scanner-clean patch and by admitting only correctly structured integration branches.

## Allowed scope

- `go.mod`
- `.github/workflows/{connector-boundary,pr-issue-guard,release,security,verify,conventions}.yml`
- `.planning/phases/ci-toolchain-and-branchname-r1/**`

## Explicit exclusions

- Application, connector, test, dependency, and generated-artifact changes.
- Vulnerability suppressions, scanner exemptions, disabled/non-blocking checks, or a one-branch exception.
- Rewriting historical planning records that mention the old toolchain.

## TDD sequence

1. **Red:** Reproduce `govulncheck` under the current Go 1.25.12 and evaluate the current branch rule against `integration/4015-mvp-flat-r1`; record both failures.
2. **Green:** Verify Go 1.25.13 resolves the scanner findings before changing pins. Update every active workflow pin and `go.mod` together.
3. **Green:** Add a separately anchored integration branch pattern that requires a positive issue number and the existing slug grammar, leaving all current exceptions and conventional branch validation untouched.
4. **Regression:** Run `govulncheck` under the selected toolchain, validate representative accepted and rejected branch names with the workflow expression, run syntax/configuration gates, and inspect the exact diff.
5. **Verify/review:** Complete the checklist and manual security review; open a direct PR only after local checks pass.
