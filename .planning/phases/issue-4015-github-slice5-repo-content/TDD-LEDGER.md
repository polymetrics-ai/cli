# Issue #4015 — GitHub slice 5 TDD ledger

## Cycle 1: produced-value certification

### Red

A command with no branch-owned schema-v2 record was uncertified even when the process exited successfully. Declared and agent-derived assertions were exercised against plausible wrong answers; a negative control that did not reject the wrong value was not acceptable evidence.

### Green

Twenty-five command executions produced branch-owned records with `status: passed`, an observed provider exchange, and a produced-value assertion. The 13 agent-derived assertions in the final pass rejected their negative controls. `go run ./cmd/connectorgen certification-matrix --check` accepted the evidence set.

### Refactor

No connector implementation was changed. Evidence filenames remained unique and run-scoped, and credentials and provider identifiers remained fingerprinted or redacted according to schema v2.

## Cycle 2: branch-local lifecycle evidence

### Red

At rebased head `86ff218aa`, this command failed with exit 1:

```bash
bash scripts/verify-gsd-workflow origin/integration/4015-mvp-flat-r1
```

The exact diagnostic was:

```text
verify-gsd-workflow: cmd/internal changed, but no GSD planning evidence changed.
```

### Green

The phase now contains `PLAN.md`, `TDD-LEDGER.md`, and `VERIFICATION.md`, documenting the inline manual-GSD fallback, the certification method, the red/green proof, and the exact verification results. The same workflow-evidence command is required to exit 0 before the fix is pushed.

### Refactor

The artifact names and location match the paths accepted by `scripts/verify-gsd-workflow`; no production or generated certification file was altered to satisfy the gate.
