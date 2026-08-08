# Foundation trace — declarative date-only CLI flags

## Why this foundation belongs in the Zoom Cobrowse SDK slice

Zoom's live Cobrowse SDK artifact explicitly declares optional monthly `from`/`to` **date** query
inputs. `cli_surface.json` and `commandrunner` accepted only `date-time`, so declaring the provider
format truthfully failed validation/runtime preflight. This is a shared declarative CLI capability,
not a reason to omit or weaken Cobrowse's documented inputs.

The foundation extends the existing date-time flag-format validation to date-only (`YYYY-MM-DD`),
updates the CLI surface schema and `connectorgen` validator consistently, and supplies an exact
runtime test. It unblocks any connector whose provider declares a date-only command input; no
connector-specific HTTP behavior or policy is introduced.

## RED — captured before foundation implementation

```text
$ go test -count=1 -run '^TestValidateFlagFormatDate$' ./internal/connectors/commandrunner
--- FAIL: TestValidateFlagFormatDate (0.00s)
    runner_test.go:2247: validateFlagValue valid ISO date = connector command "unknown" is blocked: flag --from has unsupported format "date", want accepted
FAIL
FAIL    polymetrics.ai/internal/connectors/commandrunner    0.783s
FAIL
```

The red test is committed before the schema/validator/runtime implementation. Green evidence is
added in the follow-up foundation commit.
