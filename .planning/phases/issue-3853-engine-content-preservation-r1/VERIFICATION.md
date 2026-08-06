# VERIFICATION — issue #3853 engine content preservation

Status: passed by automated execution, with the explicitly safety-excluded smoke target recorded below.

## Checklist

- [x] Preview warnings retain top-level, nested, and config-secret interpolation content.
- [x] Existing `redact_fields` declarations remain load-compatible without rewriting bundles.
- [x] Direct-read, operation-direct-read, and binary-download failure messages retain bounded
  provider diagnostic content.
- [x] Error-map class/hint, request bounds, redirect protections, preview digest, typed
  confirmation, approval evidence, and no-retry behavior remain unchanged.
- [x] No #3771 command-runner, #3852 enum, successful-output policy, binary-record, generic
  source-table, provider, credential, or reverse-execution scope leakage occurs.
- [x] Runtime help, manual, golden transcript, and website documentation agree about the
  connector-engine complete-content boundary and approval-token handling.
- [x] Targeted tests, package tests, static/build checks, individual repository gates, and inline
  GSD verification/review are recorded with exact command output.

## Executed verification

- RED: `go test ./internal/connectors/engine -run '^(TestDryRunWritePreviewResolvedMethodPathPreservesSecretValues|TestDryRunWritePreviewResolvedPathPreservesConfiguredRecordFields|TestDryRunWritePreviewResolvedPathPreservesNestedRecordFields|TestDirectReadPreservesHTTPErrorText|TestOperationDirectReadPreservesHTTPErrorText|TestBinaryDownloadPreservesHTTPErrorTextAndLeavesNoFile)$' -count=1 -v` failed before the production change: the preview replaced values with `***` or `redacted`, and the three HTTP errors removed query/body diagnostics.
- Focused green: the same six-test command passed after the engine change.
- Package regressions: `go test ./internal/connectors/engine` and `go test ./internal/cli` passed. The latter was rerun warm to capture its exit status after the initial long package run completed.
- CLI/docs: `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test ./internal/cli -run '^TestGoldenTranscripts$' -count=1` and `go test ./internal/cli -run '^(TestGoldenTranscripts|TestGoldenDocsGenerateMatchesTrackedCLIManuals)$' -count=1` passed. `go run ./cmd/pm docs generate --dir docs/cli --connectors-dir docs/connectors` regenerated the manual; unrelated Amazon SQS generator drift was reverted rather than absorbed.
- Runtime help/manual parity: `go build -o pm ./cmd/pm`, `./pm reverse`, `./pm help reverse`, and `./pm reverse --help` all passed and displayed the contextual reverse-ETL manual. `./pm docs validate --connectors-dir docs/connectors` passed.
- Static/build: `gofmt -d` over changed Go files produced no output; `go vet ./internal/connectors/engine`, `go vet ./internal/cli`, and `go build ./cmd/pm` passed.
- Individual repository gates: `make tidy-check`, `make lint` (0 issues), `make docs-check`, `make agent-contract-check`, `make connectorgen-validate` (550 connectors, 0 findings), `make connectorgen-surface-sync` (550 connectors, 0 drift), `make connector-boundary` (clean), and `make release-workflow-check` all passed.
- `git diff --check` passed.

## Inline GSD verification and review

`scripts/gsd prompt verify-work 3853` and `scripts/gsd prompt code-review 3853` were resolved and executed inline. The canonical no-spawn task contract precluded the workflow's optional reviewer role. The coverage evidence is in `SUMMARY.md`; the manual review result is in `REVIEW.md`.

## Intentionally excluded

- Whole-repository `go test ./...` and `make verify` monolith: CI-owned under the repository's
  connector-suite timeout guidance.
- `make smoke-no-build`: it creates credentials and executes a local `pm reverse run`, while this
  issue forbids credentialed checks and reverse-ETL execution.
- Live provider calls and credentialed checks: prohibited by issue scope.

## Project-memory helper

`/Users/karthiksivadas/karthik-agent-workspace/bin/fm-ensure-agents-md.sh .` was invoked as required.
It refused to reconcile the repository's pre-existing distinct real `AGENTS.md` and `CLAUDE.md` files.
No task-specific, project-intrinsic agent-memory change was discovered, so that unrelated reconciliation
was not absorbed into issue #3853.
