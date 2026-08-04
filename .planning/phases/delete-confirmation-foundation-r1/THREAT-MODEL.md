# Threat model — destructive-write confirmation gate

## Assets and trust boundary

The protected asset is provider data affected by destructive connector operations. Untrusted input
includes CLI arguments, local state files, connector records, command flags, and bundle metadata.
No live credentials or provider calls are part of foundation verification.

## Threats and mitigations

| Threat | Mitigation | Test evidence |
| --- | --- | --- |
| Execute directly without a plan | engine rejects missing plan identity/hash evidence before callback | direct-write RED/GREEN test |
| Skip preview | engine requires the exact dry-run digest; app requires persisted preview state | preview-order RED/GREEN test |
| Reuse approval for changed data | existing plan-hash/payload checks plus preview digest comparison | hash mismatch and digest mismatch tests |
| Free-text or unknown confirmation | closed JSON schema and typed parser accept only `destructive` | schema/parser tests |
| Strip a gate from local state | destructive intent is re-derived from the live action/operation metadata | stored-plan bypass test |
| Bulk a human-only action | existing `batchable: false` plan-time and run-time checks remain independent | existing plus focused regression tests |
| Add a synthetic command to evade policy | command mapping and operation ID stay canonical; no alias surface is added | commandrunner canonical mapping test |
| Future executor omits provider-specific safety | executor calls the generic gate wrapper before its closure | seam fixture test |

## Residual risk

Native connector implementations that bypass `internal/app` are internal Go call sites, not public
commands. Public reverse-ETL execution remains gated centrally. Native lanes should adopt the shared
evidence helper when they add destructive engine-style executor seams.
