# Context: Foundation public-output repair r1

## Locked scope

- Repair only `FND-B10` through `FND-B14` and `FND-W02` from the immutable
  review commit `c9824b5837f487acaa2c2a39126d29cf401d7fb5`.
- Preserve ordinary provider values, including IDs, occurrence IDs, and
  token-shaped values. Public masking has authority only for configured
  secrets and their concrete printable/encoded representations.
- Provider I/O must be exercised only through hermetic provider doubles. No
  test obtains, prints, persists, or commits a real credential.
- Malformed GitHub App restrictions and undeclared/unsafe binary or status
  parameter bindings must fail before authenticated provider I/O.
- Do not edit source-import/certification findings `FND-B01`–`FND-B09` or
  `FND-W01`, nor reverse-ETL action-binding behavior. A necessary shared
  utility conflict is a `needs-decision` stop, not authorization to widen the
  change.

## GSD discussion record

`scripts/gsd prompt discuss-phase 4307`, `plan-phase 4307 --tdd`,
`execute-phase 4307`, `verify-work 4307`, and `code-review 4307` were
resolved through the project-local adapter. The immutable launch brief supplies
the otherwise interactive decisions, and the environment prohibits spawning
separate delivery roles, so the lifecycle is executed inline and recorded in
this phase.
