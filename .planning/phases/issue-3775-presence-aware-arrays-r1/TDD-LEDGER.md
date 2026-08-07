# TDD LEDGER — issue #3775 presence-aware required string arrays

| ID | Enforcement | RED evidence | GREEN evidence | Refactor/verification |
| --- | --- | --- | --- | --- |
| R1 | Required array raw presence is distinct from cardinality | `go test ./internal/connectors/commandrunner -run '^TestValidateRequiredCommandFlagsPreservesStringArrayPresence$'` failed on the unchanged runner: both `[]string{""}` and `[]string{", ,"}` returned `missing required flag --items for command "widgets create"` | The same focused test passed after `commandValueEmpty` retained post-coercion array presence while `coerceFlagValue` continued to enforce bounds | Included in `go test ./internal/connectors/commandrunner` (passed) |
| R2 | Explicit blank direct-read array reaches `body.items` as literal `[]string{}` | After the required-presence correction, `TestRunOperationDirectReadPreservesExplicitEmptyRequiredStringArray` failed because `json.Marshal(body)` returned `{"items":null}` rather than `{"items":[]}` | Initializing the coercion result as `make([]string, 0)` made the public `Run` path pass with typed `[]string{}` and JSON `{"items":[]}` | Focused test and package test passed |
| R3 | Explicit blank reverse-ETL array reaches planned `record.items`, retains approval, and never executes | The same focused run failed in `TestBuildWriteCommandPreservesExplicitEmptyRequiredStringArray`: `json.Marshal(record)` returned `{"items":null}` rather than `{"items":[]}` | The public `BuildWriteCommand` path now passes with typed `[]string{}`, JSON `{"items":[]}`, `ApprovalRequired: true`, validation only, and no `Write` call | Focused test and package test passed |
| R4 | Omitted/raw-empty/min-items-one controls remain rejected before executor dispatch | Pending until public-path table cases were added | Both public-path tables assert the errors and prove the fake operation/Write path was not reached | Focused test and package test passed |
| R5 | Existing max-items and scalar-blank behavior are unchanged | Existing `TestCoerceFlagValueBoundsStringArrayItems` covers bounds; new table adds scalar blank | `TestValidateRequiredCommandFlagsPreservesStringArrayPresence` accepts neither required scalar blank nor `MaxItems: 2` input with three values | Focused test and package test passed |

## Red-test rule

The explicit-empty acceptance assertion must run and fail on the current baseline before changing
production code. Test removals, skips, or weakened assertions are not permitted.
