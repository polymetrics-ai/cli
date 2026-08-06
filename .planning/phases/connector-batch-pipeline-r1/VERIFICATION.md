# Verification — connector batch pipeline r1

Status: planned; update with actual command output before handoff.

## Evidence to collect

- Ledger field validation and deterministic manifest tests.
- Individual gate continuation/drop behavior with isolated temporary bundles.
- Real `commandrunner.Preflight` invocation for every implemented command in a
  gated bundle.
- `surface-sync --check` equivalence through `syncBundle(..., true)`.
- Generated five-candidate manifest exact totals and citations.
- Targeted Go package, static, and repository gates from `PLAN.md`.

## External wait

The manifest and gate can be completed on the current base. Actual bundle
authoring remains intentionally paused pending merged #3869 (v2 provenance),
#3870 (non-redacting output policies), and associated #3868 verification.
