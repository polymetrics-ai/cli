# Threat model — destructive-write confirmation gate

## Assets and trust boundary

The protected asset is provider data affected by destructive connector operations. Untrusted input
includes CLI arguments, local state files, connector records, command flags, and bundle metadata.
No live credentials or provider calls are part of foundation verification.

## Threats and mitigations

| Threat | Mitigation | Test evidence |
| --- | --- | --- |
| Execute directly without a plan | engine accepts only opaque evidence authenticated from a vault-derived external key | forged-evidence RED/GREEN test |
| Skip preview | engine requires the exact prepared-request digest; app requires persisted preview state | preview-order RED/GREEN test |
| Reuse or copy approval evidence | atomic state consumption plus a shared one-shot in-memory marker rejects both process and engine replay | command, bulk, and copied-evidence tests |
| Modify token hash or grant in state | the HMAC binds token hash, plan, target, digest, confirmation, and expiry | state-tamper and grant-authentication tests |
| Change credentials or secret-derived target | HMAC credential revision and concrete target digest change without exposing secrets | credential/target drift tests |
| Reuse approval for changed data | plan/payload checks plus all-record canonical request digest comparison | hash mismatch and digest mismatch tests |
| Free-text or unknown confirmation | closed JSON schema and typed parser accept only `destructive` | schema/parser tests |
| Strip a gate from local state | destructive intent is re-derived from the live action/operation metadata | stored-plan bypass test |
| Bulk a human-only action | existing `batchable: false` plan-time and run-time checks remain independent | existing plus focused regression tests |
| Add a synthetic command to evade policy | command mapping and operation ID stay canonical; no alias surface is added | commandrunner canonical mapping test |
| Native or future executor omits provider-specific safety | native SQS and the future executor seam call `ExecutePreparedWrite` before their closures | SQS and rest-write seam tests |

## Residual risk

Future native write implementations must construct `PreparedWrite` and dispatch through
`ExecutePreparedWrite`; Amazon SQS demonstrates the required provider-neutral integration.
