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

Live GitHub runs were executed against disposable repo `karthik-sivadas/pm-cert-test-20260709025802` using an environment-provided token. Early runs produced reports but did not pass; the final run passed.

Final report path:

```text
/var/folders/tk/bmp_tx0976s4rkh1phvrpjlw0000gn/T/tmp.5VxmsRQj6x/.polymetrics/certifications/github.json
```

Final passing run:

```text
started_at: 2026-07-08T23:11:29.08724Z
completed_at: 2026-07-08T23:17:34.330273Z
passed: true
stage_count: 925
failed_count: 0
skipped_count: 129
```

Key final evidence:

- Surface accounting passed: 507 endpoints, 105 covered, 402 blocked.
- Catalog passed with 37 streams.
- Direct-read sweep passed with 2 stages.
- Binary download surface remained safely blocked.
- Flow and schedule capabilities passed; schedule residue was false.
- Secret redaction and JSON contract passed.
- Write action inventory accounted for 67 actions: 1 live pass (`create_label`), 10 untested pairings, 56 blocked.
- `create_label` write lifecycle passed with read-back verification and cleanup; residue check found 0 `pm-cert-github-*` labels remaining.

Early-run findings fixed with tests:

- Schedule names used underscores even though `pm schedule` only accepts lowercase alphanumeric plus hyphen.
- `repo read-dir` used `--path .`, which the direct-read policy rejects.
- GitHub `create_label` record generation omitted required `color`.
- Streams without cursor/primary-key metadata were forced through unsupported sync modes.
- Optional GitHub security/project streams returned permission/config availability errors and now record documented skips.
- Reverse writes needed GitHub's default `base_url` stored with the credential.

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
