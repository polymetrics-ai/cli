# Threat model — destructive-write confirmation gate

## Assets and trust boundary

The protected asset is provider data affected by destructive connector operations. Untrusted input
includes CLI arguments, local state files, connector records, command flags, and bundle metadata.
No live credentials or provider calls are part of foundation verification.

## Threats and mitigations

| Threat | Mitigation | Test evidence |
| --- | --- | --- |
| Execute directly without a plan | production evidence requires the project-vault authority plus an authenticated plan seal; caller-key evidence is untrusted | caller-selected-authority RED/GREEN test |
| Skip preview | engine requires the exact prepared-request digest; app requires persisted preview state | preview-order RED/GREEN test |
| Reuse or copy approval evidence | revision CAS, create-exclusive vault nonce marker, and a shared one-shot in-memory marker reject stale save, state rollback, process, and engine replay | command, bulk, rollback, stale-save, and copied-evidence tests |
| Modify token hash or grant in state | the HMAC binds token hash, plan, target, digest, confirmation, and expiry | state-tamper and grant-authentication tests |
| Change credentials or secret-derived target | HMAC credential revision and concrete target digest change without exposing secrets | credential/target drift tests |
| Change effective configuration or batchability | plan seal and grant target bind configuration HMAC plus the live batchable flag | configuration/batchability drift tests |
| Extend mutable state expiry | plan and grant deadlines come from authority time and are MAC-bound; state timestamps cannot extend them | extended-deadline regression test |
| Reuse approval for changed data | plan/payload checks plus all-record canonical request digest comparison | hash mismatch and digest mismatch tests |
| Free-text or unknown confirmation | closed JSON schema and typed parser accept only `destructive` | schema/parser tests |
| Strip a gate from local state | destructive intent is re-derived from the live action/operation metadata | stored-plan bypass test |
| Bulk a human-only action | existing `batchable: false` plan-time and run-time checks remain independent | existing plus focused regression tests |
| Add a synthetic command to evade policy | command mapping and operation ID stay canonical; no alias surface is added | commandrunner canonical mapping test |
| Native or future executor omits provider-specific safety | native SQS and the future executor seam call `ExecutePreparedWrite` before their closures | SQS and rest-write seam tests |

## Residual risk

Future native write implementations must construct `PreparedWrite` and dispatch through
`ExecutePreparedWrite`; Amazon SQS demonstrates the required provider-neutral integration.
