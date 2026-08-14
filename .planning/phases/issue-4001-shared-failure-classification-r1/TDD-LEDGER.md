# TDD ledger — Issue 4001: shared connector failure classification

| ID | Guarantee | Red assertion | Green proof |
| --- | --- | --- | --- |
| R1 | Closed domains | There is no contract rejecting an empty or unknown domain. | `Classification.New` and JSON decode accept only configuration/system/transient. |
| R2 | Retry safety | Configuration errors can be retried by default. | `Retryable` is true only for transient; configuration and system are false. |
| R3 | Stable dispatch vocabulary | Dispatch kinds are strings that callers can misspell. | The five required kinds validate; unknown kinds and non-system dispatch classifications fail. |
| R4 | Safe diagnostic separation | User JSON can contain a Go cause, or callers must parse error text. | JSON contains domain/code/message/path/kind/references only; `Cause`/`Unwrap` retain the private cause. |
| R5 | Exact path | A field path is absent or an arbitrary unsafe string, including malformed Unicode from direct input or JSON decoding. | JSON Pointer validation accepts escaped pointers and rejects malformed syntax, invalid UTF-8, and unpaired UTF-16 escapes before or during JSON decoding. |
| R6 | Database configuration consumer | The generic validation boundary cannot preserve a typed non-retryable configuration failure. | Focused `internal/connectors` test observes the exact classification and its cause unchanged. |
| R7 | Engine configuration consumer | Schema constraints produce only text errors. | Focused engine test observes configuration domain, code, exact field path, and private cause. |
| R8 | Engine dispatch consumer | Commandrunner has no common carrier for #3991's result, or an absent optional classification appears as a typed error. | `BlockedCommandError` carries and unwraps every valid dispatch classification while an absent classification unwraps to nil. |
| R9 | Certification consumer | Certification must invent a local untestable-reason enum or serialize a cause. | `CapabilityResult.untestable_reason` uses the common JSON object and omits cause text. |
| R10 | Stacked replay integrity | Red: the current child has no verified copy of the seven-patch source series on the parent branch. | Green: `git range-diff` proves the seven source patches replay in order onto `5996a8a2a5e99c8aa8eb5a8603ecb1f6bba21f12`; current focused and repository gates prove that transplant has not changed behavior. |

## Red command

```sh
go test ./internal/failures ./internal/connectors ./internal/connectors/engine ./internal/connectors/commandrunner ./internal/connectors/certify -run 'Test(Classification|ValidateConfigurationPreserves|SchemaValidateConfigurationReturns|BlockedCommandErrorCarries|CapabilityResultSerializes)' -count=1
```

The exact initial output is retained at `traces/red-run.txt`. The package does not exist at red
time, so compilation failure is the expected red signal. The same command must pass after green.

## Green evidence

Passed after implementation:

```sh
go test -timeout 20m ./internal/failures ./internal/connectors ./internal/connectors/engine ./internal/connectors/commandrunner ./internal/connectors/certify
go vet ./internal/failures ./internal/connectors ./internal/connectors/engine ./internal/connectors/commandrunner ./internal/connectors/certify
go build ./cmd/pm
```

`internal/failures/classification_test.go` proves R1-R5, including malformed Unicode JSON Pointer rejection at construction and raw JSON decoding boundaries while preserving valid UTF-16 surrogate pairs. The generic configuration boundary test
proves R6 without changing the in-flight PostgreSQL driver. Engine validation proves R7, the
commandrunner carrier proves R8, including the absent-classification nil boundary, and certification report JSON proves R9. The second green run
also pins unknown domain/dispatch JSON rejection and RFC 6901 escaping for declaration keys.

## Stacked-delivery note

R10 is an ancestry-only delivery check, not a new production behavior slice. The original Red and
Green evidence for R1-R9 remains unchanged in `traces/red-run.txt` and the original delivery
commits. The replay's current-base Green evidence is recorded in `VERIFICATION.md` after local
gates and review complete.

### R10 Green evidence

`git range-diff origin/main..origin/fm/cli-cert-shared-foundations-r1
origin/docs/4015-connector-release-certification..HEAD` reported all seven corresponding patches
as `=`. A second stable `git patch-id` comparison matched each source/replay pair in order. The
focused and changed-package tests listed in `VERIFICATION.md` passed on the replay head.
