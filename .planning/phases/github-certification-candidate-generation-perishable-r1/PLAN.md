# GitHub certification candidate generation — plan

**Issue link:** Refs #4015

## Scope and ownership

The target connector is GitHub. Shared tooling remains provider-neutral; all
provider fixture values and candidate-selection policy live in
`internal/connectors/defs/github/`. The concurrently owned
`cmd/connectorgen/certificationevidence.go` is excluded.

## TDD plan

1. **RED — generated candidate projection.** Add a focused connectorgen test
   requiring an implemented declared direct-read command to receive a generated
   candidate, typed argument atoms, and a non-empty produced-value assertion.
   It must reject a candidate that only asserts exit status/envelope metadata.
2. **GREEN — generic projection.** Implement the provider-neutral projection
   from CLI command/operation metadata plus connector-owned fixture bindings;
   preserve explicitly hand-authored cases and report them by named reason.
3. **RED — perishable inventory.** Add a definition/test fixture that requires
   the trial families to resolve to exactly 192 declared commands, with no
   duplicate command; generate only their 97 direct-read commands and defer
   their 95 reverse-ETL commands to the mutation-fixture lifecycle.
4. **GREEN — GitHub binding and generated artifacts.** Add the GitHub-owned
   bindings, regenerate the certification sweep, and prove the status buckets
   still sum to 1,571.
5. **Live proof — serialized/resumable.** Build the real binary, use only the
   authorized disposable identity, execute the 97 generated perishable reads
   with produced-value assertions, then record executed pass, provider result,
   product defect, and fixture-blocked outcomes. Do not claim certification or
   evidence publication; the latter is owned by the concurrent importer lane.
6. **Failure demonstration.** Deliberately break one certified candidate's
   own produced-value assertion after candidate compilation, run its own case
   to RED, restore the assertion, and run it GREEN.

## Verification plan

- `go test -timeout 20m ./cmd/connectorgen` (the consumer of certification
  definitions), targeted certification packages, and `go build ./cmd/pm`.
- Generated and connector gates: `go run ./cmd/connectorgen certification-sweep
  . --connector github --check`, `go run ./cmd/connectorgen surface-sync
  --check`, `make connectorgen-validate`, `make connectorgen-surface-sync`, and
  `make connector-boundary`.
- Repository checks individually under the per-command timeout: tidy, lint,
  docs, smoke-no-build, agent-contract, and release-workflow checks. Record any
  base-branch `internal/cli` timeout separately rather than weakening tests.
- Run relevant generators twice and confirm the second run is byte-stable.
- Run `make verify` only if the environment permits a full-suite command;
  otherwise record each exact gate and why the full suite cannot complete.

## Commit checkpoints

1. Planning artifact checkpoint.
2. Failing test checkpoint.
3. Generator, definition, regenerated-artifact green checkpoint.
4. Live result and verification checkpoint.
5. Review/fix checkpoint.

## Pre-implementation measurement — resolved

The current generated sweep proves that a direct-read-only candidate projection
cannot satisfy the assigned live proof by itself:

- `fixture_required` is 1,377 commands: 495 `direct_read`, 279
  `direct_write`, 577 `reverse_etl`, 15 `etl`, and 11 `binary_download`.
  All 495 direct reads lack an output assertion.
- 228 `reverse_etl` commands have no `api_surface` entry at all. Their current
  reason explicitly requires fixture, plan, preview, approval, execution,
  independent read-back, and cleanup. The declared CLI surface cannot synthesize
  that ceremony or its fixture values.
- The only currently executable candidate type is
  `CertificationCommandCandidate`, consumed by `stageDirectReadSweep`; it
  requires a `ConnectorCommandDirectRead` result and a direct-read response
  envelope. It cannot execute a reverse-ETL or direct-write contract.
- `cli_surface.json` and `certification-sweep.json` have no `trial_family`,
  entitlement, or perishable-cohort field. The stated 50/50/46/46 slice cannot
  be selected deterministically from those inputs without a new connector-owned
  cohort declaration.

**Resolved scope:** the user narrowed this PR to generic direct-read candidate
generation plus connector-owned cohort declarations. The result must state the
97-read/95-reverse-ETL split and must not imply certification of the whole 192.
The mutation-candidate and fixture-lifecycle contract remains a separate,
designed delivery.
