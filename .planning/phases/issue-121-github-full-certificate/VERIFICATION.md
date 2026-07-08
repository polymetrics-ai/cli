# Issue #121 Verification

## Local Non-Credentialed Gates

Executed on the correctly stacked branch `feat/121-certify-full-sweep-stacked`:

```text
PASS: git diff --check
PASS: go test ./internal/connectors/certify/ -run 'TestDefaultStreamName|TestFullSweepSourceStagesAgainstSample|TestSourceStagesAgainstSample|TestWriteStagesLedgerWrittenBeforeCreate|TestSweepPairingsForGithubHasMultiple|TestDirectReadCandidateForGitHub|TestBinaryDownloadCandidateForGitHub|TestFullSweepFlowAndScheduleNamesAreStreamScoped' -count=1 -timeout=4m -v
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

- The full GitHub certificate has not been live-tested.
- The token pasted into chat was not used and must be revoked/rotated.
- Full `go test ./...` timed out locally in host-crontab code without `PM_CRONTAB_FILE`; the isolated seam test passed previously.
- The certify package can be long-running because full-mode sample coverage exercises repeated in-process CLI flows.

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
