# TDD Ledger — Zoom Healthcare parity, R1

## Scope and safety

- Target connector: `zoom` only; provider module: `healthcare` (#3946).
- No live provider mutation, no credential creation, and no secret values in tests, logs, or this ledger.
- The only live request planned is a synthetic-token read reachability check after build. PATCH is
  exercised only against an in-process HTTP fixture through the existing reverse-ETL approval path.

## RED — pending

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

Pending — no production bundle file has been edited.

## GREEN — pending

1. Add the two `rest_read` operation declarations and the typed `update_clinical_note` write action.
2. Add Healthcare command-surface declarations, leaving derivable metadata to `surface-sync`.
3. Add sanitized fixtures, update metadata/docs, run generation, then run the red test to green.
4. Record the exact green commands and output in this ledger and `VERIFICATION.md`.

## Refactor / review — pending

- Retain the generic command-count helper; do not add provider-specific branches in shared runtime code.
- Review declaration ownership, path/body mapping, clinical redaction, 204 acceptance, and generated-file scope.
