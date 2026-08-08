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

## GREEN — captured

1. Added two `rest_read` declarations with the existing `clinical_json_redacted` response policy,
   and the typed `update_clinical_note` PATCH action with a closed two-field record schema.
2. Added the three Healthcare commands. `surface-sync` generated the direct-read endpoint/output
   metadata and runtime endpoint projection; the scoped reconciler generated only the two
   `covered_by.direct_read` rows. The PATCH ledger row is linked to
   `covered_by.write=update_clinical_note`.
3. Kept the synthetic fixture-only execution test: direct reads assert redaction and documented
   request inputs, while PATCH asserts exact JSON body/path and a successful `204 No Content`.
4. Generated the connector documentation, retained the Zoom/catalog delta only, and verified the
   binary routes using a temporary local credential with a synthetic environment token. No token
   value was printed or retained.

### Green evidence

```text
$ go test -count=1 ./internal/connectors/defs/zoom/...
ok  	polymetrics.ai/internal/connectors/defs/zoom	0.859s

$ go run ./cmd/connectorgen surface-sync --check
connectorgen surface-sync: 551 connector(s) scanned, 0 field(s) filled and 0 field(s) corrected across 0 connector(s)

$ go run ./cmd/connectorgen validate internal/connectors/defs/zoom
connectorgen validate: 1 connector(s) checked, 0 findings

$ go run ./cmd/connectorgen surface-reconcile internal/connectors/defs/zoom --notes-contains provider_module=healthcare
zoom: reconciled covered=2 blocked=0 unchanged=0 refused=1
```

Built-binary evidence (`go build -o <temporary>/pm ./cmd/pm`): `pm help zoom`, bare `pm zoom`,
bare `pm zoom healthcare`, and all three Healthcare route helps resolved. With a synthetic token
stored only through `--from-env`, list and get each reached Zoom and returned HTTP `401` (exit `1`),
not `unknown command`. The update command returned a typed preview (exit `0`) and made no live PATCH.

## Refactor / review — captured

Inline `verify-work` and `code-review` are complete under the documented manual-GSD fallback.
The final focused test, both surface checks, full connector validation, and diff-scope check passed
after the implementation. See `VERIFICATION.md` and `REVIEW.md` for commands and dispositions.

- Retain the generic command-count helper; do not add provider-specific branches in shared runtime code.
- Review declaration ownership, path/body mapping, clinical redaction, 204 acceptance, and generated-file scope.
