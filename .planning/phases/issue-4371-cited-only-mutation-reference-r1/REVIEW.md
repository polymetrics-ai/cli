# Inline GSD code review — issue 4371

Reviewed under the generated `scripts/gsd prompt code-review 4371` contract,
inline because the canonical single-worker delivery contract forbids spawning a
review role.

## Scope

- `sourceProjectionApplyNonExecutableMutationDispositions`
- `sourceProjectionApplyPartialMutationCoverageDispositions`
- `sourceProjectionApplyWriteDisabledMutationArtifacts`
- cited-only/ordinary/disposition regression coverage

## Findings and disposition

| ID | Severity | Finding | Disposition |
| --- | --- | --- | --- |
| F1 | high | Explicit non-executable disposition changed a cited-only descriptor before strict validation. | Fixed by pre-mutation closed-reference rejection; red/green test covers it. |
| F2 | high | Partial-coverage disposition could make the same illegal second-gap change. | Fixed by the shared pre-mutation rejection; red/green test covers it. |
| F3 | high | Automatic write-disabled mutation artifact could reintroduce F1 after explicit applications. | Fixed by skip when the sole `source_contract_unavailable` gap exists; red/green test covers it. |
| R1 | none | No additional correctness, security, race, output, source-provenance, command-reachability, or CLI/docs parity finding. | Accepted: zero blockers at this review SHA. |

## Review checks

- Citation validation remains before the new guard, so invalid/unknown/
  duplicate/mismatched/non-mutating dispositions keep their existing
  fail-closed behavior.
- Contract-complete OpenAPI/Swagger operations lack the cited-only foundation
  and retain the existing mutation-disposition path byte-for-byte.
- The strict descriptor validator was not weakened; it still compares a
  cited-only descriptor with the lock-owned exact reference projection.
- No user-controlled transport, secret, generic HTTP write/body, provider I/O,
  action, command, or credential path was introduced.
- Fresh exact-head independent review remains required after the final evidence
  commit and before any merge consideration.
