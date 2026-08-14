# #4089 — TDD ledger

**Status:** green; implementation, focused regression, and CI-repair evidence complete.

| Checkpoint | Evidence | Result |
| --- | --- | --- |
| Plan | GSD prompts resolved inline; issue decisions, scope, parity, and safety checks recorded before production edits. | recorded |
| Red: bounded stdin carrier | `go test -count=1 -timeout 20m -run '^TestReverseETLApprovalUsesBoundedStdin$' -v ./internal/cli` built a fresh binary, observed a live argv with only the stdin marker, then failed the valid run with exit status 1 because the unchanged CLI ignored stdin. | reproduced |
| Green: generic wiring | Both request construction paths call `reverseApprovalTokenFromStdin`, which only validates the bare marker then delegates to `readApprovalTokenFromStdin(os.Stdin)` before the plan lookup or write dispatch. | passed |
| Green: binary regression | `go test -count=1 -timeout 20m -run '^TestReverseETLApprovalUsesBoundedStdin$' -v ./internal/cli` passed: it logged the live safe argv, independently checked argv/environment/project files/logs/receipt/evidence, rejected empty/oversized/multiline/valued/argv inputs with no receipt, and rejected replay. | passed |
| Refactor: docs and safety | Generated manuals, website data, transcripts, explicit stale-syntax scan, focused certification harness tests, and smoke run are green. | passed |

## Required red/green record

- Red: observed on 2026-08-14. The fresh binary process argv was recorded without the token; the valid stdin run failed at the unchanged argv-only request construction (exit 1). No production file had changed.
- Green: observed on 2026-08-14. The same selector exited 0. Its logged live command line was `pm reverse run <plan-id> --approval-token-stdin --root <temp-root> --json`; it contained no token. The test then asserted the token absent from argv, the command environment, durable project files, captured logs, the outbox receipt, and its emitted evidence record, while replay was rejected without another receipt.

## CI repair — 2026-08-14

- Red: `go test -count=1 -run '^TestCredentialCoordination_EmptyProjectOpenDoesNotRewriteState$' -v ./internal/app` reproduced the Verify failure: opening a fresh project changed revision `0` to `1`. CI also showed the same extra revision in the four `TestRunETLTransportAcknowledgedCompletionMissingRunIsTypedConflictForAllModes` cases.
- Red: CI `govulncheck` ran the repository-pinned Go `1.25.12` standard library and reported reachable GO-2026-6218, GO-2026-6091, GO-2026-6090, GO-2026-6089, GO-2026-6088, GO-2026-5972, and GO-2026-5026; all list Go `1.25.13` as the fixed version.
- Green: `go test -count=1 -timeout 20m -run '^(TestCredentialCoordination_EmptyProjectOpenDoesNotRewriteState|TestRunETLTransportAcknowledgedCompletionMissingRunIsTypedConflictForAllModes)$' -v ./internal/app` passed all eight relevant tests (seven ETL mode subtests plus the empty-project reopen regression), and `go test -count=1 -timeout 20m ./internal/app` passed.
- Green: `GOTOOLCHAIN=go1.25.13 go run golang.org/x/vuln/cmd/govulncheck@latest ./...` reported `No vulnerabilities found.` The module and every pinned GitHub Actions Go setup now use `1.25.13`.
- Required skills used for this CI repair: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-lint`, and `golang-documentation`.

## Rebased CI continuation — 2026-08-14

- Rebase: rebased the branch cleanly onto `origin/integration/4015-mvp-flat-r1`, the required PR base. The base already carries the Go `1.25.13` scanner remediation, and the branch retains its app-side CI repair.
- Red: the first rebased `go test -count=1 -timeout 20m ./internal/app` run failed `TestRunReverseETLRecoversCommittedApprovalConsumptionUnlockFailure`. Its `failAt: 2` injector now failed the read-only `UpdateAfterPreflight` unlock, not the later committed update unlock.
- Green: moved that test's explicit failure point to the third unlock (plan load, preflight read, committed update) so it again proves that a post-commit unlock failure returns `ApprovalConsumptionUncertainError` with a committed outcome and no external write.
- Green: both uncertainty-recovery selectors pass, the full `internal/app` package passes, the real-process stdin approval selector passes, and `GOTOOLCHAIN=go1.25.13 go run golang.org/x/vuln/cmd/govulncheck@latest ./...` reports no reachable vulnerabilities.

## Post-merge approval-invalidation repair — 2026-08-14

- Red: after merging `integration/4015-mvp-flat-r1` at `e28152c8466adc184b167746895bcc32bd62f69e`, `go test -count=1 -timeout 20m ./internal/app -run 'TestRunReverseETLRejectsPlanHashMismatchWhenRowsChange|TestMultipartDirectWriteCommandRejectsChangedPayloadBeforeNetwork'` failed. Both validations returned `reverse plan approval has already been consumed` instead of the expected changed-payload refusal.
- Cause: `invalidateReversePlan` used its state-mutating callback as the `UpdateAfterPreflight` observer. Because `state` contains a slice, that observer cleared the approval hash in the in-memory preflight snapshot before the single atomic update.
- Green: split `reversePlanInvalidationCandidate` into a pure observer and a single mutation path. The same selector now passes, as do the replay, cross-process atomic-consumption, stale-state, and legacy-lock selectors plus the raw-argv carrier selector in `internal/cli`.
