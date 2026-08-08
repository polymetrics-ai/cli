# PersistIQ artifact materialization pilot - TDD ledger

> This ledger is maintained before and during the pilot. It must contain
> observed Red and Green evidence, not plan-shaped claims.

## Scope lock

- One connector: `persistiq`.
- No credentials, provider data, or live API exercise.
- 21 ledger operations; pre-fetch mapping: ETL 11, direct_read 1,
  reverse_etl 7, direct_write 2, binary_download 0, unclassified 0.

## RED

- Status: confirmed before production bundle edits.
- Test: build/run the real `pm` binary against the baseline PersistIQ bundle;
  assert the pilot command surface is not fully reachable.
- Command: `.planning/phases/persistiq-artifact-materialize-pilot-r1/pm-baseline persistiq leads list --help`
- Captured output (verbatim):

  ```text
  error: unknown command "persistiq"
  ```

  This is the baseline red: the current binary has no reachable PersistIQ
  command namespace before materialization.

## GREEN

- Status: pending staged materialization and gates.
- Test: run the same real-binary sweep against the generated PersistIQ bundle,
  plus validation, surface-sync check, runtime preflight, and batch gate.
- Captured output: to be filled with verbatim results; no credentials.

## Refactor / safety

- No new generator, generic write tool, credentials, or gate weakening.
- Any failed batch stage remains a failed pilot result; it is not papered over
  by editing the test or source artifact.
