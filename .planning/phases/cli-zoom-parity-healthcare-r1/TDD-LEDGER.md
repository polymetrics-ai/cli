# TDD Ledger — Zoom Healthcare parity, R1

## Scope and safety

- Target connector: `zoom` only; provider module: `healthcare` (#3946).
- No live provider mutation, no credential creation, and no secret values in tests, logs, or this ledger.
- The only live request planned is a synthetic-token read reachability check after build. PATCH is
  exercised only against an in-process HTTP fixture through the existing reverse-ETL approval path.

## RED — captured before production bundle edits

1. Raise the inventory/reachability expectations from the current completed baseline:
   - covered rows: `9 → 12`
   - Zoom-local blocked rows: `1833 → 1830`
   - direct-read commands: `6 → 8`
   - reverse-ETL write commands: `0 → 1`
2. Add a Healthcare command execution test that expects both direct reads to resolve through the
   real commandrunner and the PATCH action to preflight/plan and accept `204 No Content` in an
   isolated fixture.
3. Run `go test -count=1 ./internal/connectors/defs/zoom/...` without changing any production
   bundle file. Capture the exact failing output below and commit/push this red checkpoint first.

### Red evidence

Command run on the unmodified production bundle:

```text
--- FAIL: TestProviderInventoryLedgerIsComplete (0.05s)
    command_surface_test.go:145: executable rows = 9, want 12
    command_surface_test.go:148: operations awaiting Zoom-local contracts = 1833, want 1830
--- FAIL: TestCoveredStreamsHaveReachableCommands (0.03s)
    command_surface_test.go:243: reachable direct_read operation commands = 6, want 8
    command_surface_test.go:244: reachable reverse_etl write commands = 0, want 1
--- FAIL: TestHealthcareClinicalNoteCommandsExecuteWithFixtures (0.10s)
    --- FAIL: TestHealthcareClinicalNoteCommandsExecuteWithFixtures/list (0.03s)
        command_surface_test.go:661: Run("healthcare clinical-notes list") = connector command "healthcare clinical-notes list" is blocked: unknown command
    --- FAIL: TestHealthcareClinicalNoteCommandsExecuteWithFixtures/get (0.03s)
        command_surface_test.go:661: Run("healthcare clinical-notes get") = connector command "healthcare clinical-notes get" is blocked: unknown command
    --- FAIL: TestHealthcareClinicalNoteCommandsExecuteWithFixtures/update_plans_before_mutation_and_accepts_no_content (0.03s)
        command_surface_test.go:707: BuildWriteCommand = connector command "healthcare clinical-notes update" is blocked: unknown command
FAIL
FAIL    polymetrics.ai/internal/connectors/defs/zoom    0.818s
FAIL
```

This red state contains only the test, synthetic Zoom-local fixtures, and delivery evidence. No
`operations.json`, `writes.json`, `cli_surface.json`, `api_surface.json`, metadata, generated
ledger, or connector documentation production file has changed.

## GREEN — pending

1. Add the two `rest_read` operation declarations and the typed `update_clinical_note` write action.
2. Add Healthcare command-surface declarations, leaving derivable metadata to `surface-sync`.
3. Add sanitized fixtures, update metadata/docs, run generation, then run the red test to green.
4. Record the exact green commands and output in this ledger and `VERIFICATION.md`.

## Refactor / review — pending

- Retain the generic command-count helper; do not add provider-specific branches in shared runtime code.
- Review declaration ownership, path/body mapping, clinical redaction, 204 acceptance, and generated-file scope.
