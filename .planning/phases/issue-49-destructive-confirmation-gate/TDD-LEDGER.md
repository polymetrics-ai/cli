# TDD Ledger — PR #49 destructive confirmation gate correction

| Slice | Test / evidence | Status | Notes |
| --- | --- | --- | --- |
| 1 | `go test ./internal/app -run 'TestRunReverseETL.*DestructiveConnectorCommand' -count=1` | Red confirmed | Failed to compile before production edits: `ReversePlan.ConfirmationChallenge undefined`; `RunReverseETLRequest.Confirmation` unknown. |
| 1 | `TestRunReverseETLRejectsDestructiveConnectorCommandWithoutConfirmation` | Green | Missing typed confirmation rejects before HTTP dispatch. |
| 1 | `TestRunReverseETLAcceptsDestructiveConnectorCommandWithMatchingConfirmation` | Green | Matching `Confirmation: "destructive"` allows approved dispatch. |
| 1 | `TestRunReverseETLRejectsGenericDestructiveActionWithoutConfirmation` | Green | Generic reverse ETL destructive action is gated, not just connector commands. |
| 2 | `TestGitHubDestructiveCommandRequiresTypedConfirmation` | Green | CLI rejects missing `--confirm`, preserves approval token, then succeeds with `--confirm destructive`. |
| 2 | `TestBuildWriteCommandCarriesDestructiveConfirmationChallenge` | Green | Commandrunner propagates write-action `confirm` metadata. |
| 3 | `TestWriteActionInventoryForPropagatesMissingWritesFile` | Green | Full-cert inventory read/parse failures now surface as errors. |
| 3 | `TestLiveStreamUnavailableClassifiesGitHubUnavailableErrors` | Green | Case/status variants classify as documented GitHub unavailable skips. |
| 4 | `TestScheduleCLI_Remove` via full `internal/cli` and `make verify` | Green | Schedule remove test no longer touches/hangs on real crontab; uses `PM_CRONTAB_FILE`. |
