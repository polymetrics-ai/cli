# Issue #121 Verification

## Local Non-Credentialed Gates

Executed on the correctly stacked branch `feat/121-certify-full-sweep-stacked`:

```text
PASS: git diff --check
PASS: go test ./internal/connectors/certify/ -run 'TestSurfaceInventoryForGitHubAccountsForAllReviewedEndpoints|TestDirectReadCandidatesForGitHub|TestGithubWriteActionInventoryAccountsForAllDeclaredActions|TestFullSweepSourceStagesAgainstSample|TestSourceStagesAgainstSample|TestWriteStagesLedgerWrittenBeforeCreate' -count=1 -timeout=5m -v
PASS: go test ./internal/connectors/... -count=1
```

Executed on the clean branch before recreating the stacked PR:

```text
PASS: go vet ./...
PASS: go build ./cmd/pm
PASS: go test ./internal/connectors/... -count=1
```

## GSD Enforcement Gate

This PR adds:

```text
scripts/verify-gsd-workflow
.github/workflows/gsd-workflow.yml
```

Expected local use for stacked issue #121 work:

```bash
git fetch origin feat/44-github-cli-parity
scripts/verify-gsd-workflow origin/feat/44-github-cli-parity
```

## Known Caveats

- The token pasted into chat was not used and must be revoked/rotated.
- Full `go test ./...` timed out locally in host-crontab code without `PM_CRONTAB_FILE`; the isolated seam test passed previously.
- The certify package can be long-running because full-mode sample coverage exercises repeated in-process CLI flows.

## Live Test Findings

A live GitHub run was executed against disposable repo `karthik-sivadas/pm-cert-test-20260709025802` using an environment-provided token. The first run produced a report but did not pass.

Report path:

```text
/var/folders/tk/bmp_tx0976s4rkh1phvrpjlw0000gn/T/tmp.5VxmsRQj6x/.polymetrics/certifications/github.json
```

Key first-run findings:

- Surface accounting passed: 507 endpoints, 105 covered, 402 blocked.
- Catalog passed with 37 streams.
- Direct-read `repo read-dir` failed because the default path was `.`.
- Schedule stages failed for stream names with underscores because schedule names require lowercase alphanumeric plus hyphen.
- Write plan for `create_label` failed because generated record omitted required `color`.

Follow-up fixes were implemented with tests. A second live run is required after rebuilding `pm`.

## Live Test Command Template

Use only with a rotated token supplied through the local environment:

```bash
export PM_GITHUB_DEV_TOKEN='<rotated-token>'
export PM_GITHUB_DEV_OWNER='<owner-or-org>'
export PM_GITHUB_DEV_REPO='<disposable-dev-repo>'

go build ./cmd/pm
ROOT=$(mktemp -d)
./pm init --root "$ROOT" --json
./pm connectors certify github \
  --root "$ROOT" \
  --full \
  --write \
  --from-env token=PM_GITHUB_DEV_TOKEN \
  --config owner="$PM_GITHUB_DEV_OWNER" \
  --config repo="$PM_GITHUB_DEV_REPO" \
  --json
```

Do not paste or commit the token. Use a disposable/dev repo because `--write` can create labels, issues, and milestones as the live lifecycle is expanded.
