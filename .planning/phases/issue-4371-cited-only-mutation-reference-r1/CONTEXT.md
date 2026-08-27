# Context — issue 4371 cited-only mutation reference closure

## Decided contract

- Issue #4371 is a shared `cmd/connectorgen` source-import/projection repair
  linked to Batch 6–7 parent #4291. It is not a Salesloft or Copper connector
  capability implementation.
- A cited-only source-reference operation has no retained request/response
  execution contract. Its output is the closed reference projection: exact
  source identity/provenance, provider operation identifier, method/path when
  cited, `merge_blocked: true`, and exactly one
  `source_contract_unavailable` gap.
- Applying either a non-executable mutation disposition or a partial mutation
  coverage disposition to that closed descriptor is incompatible. The narrow
  repair rejects that input before descriptor output is written, naming the
  cited source operation. It must not relax strict descriptor validation.
- Ordinary retained OpenAPI/Swagger operations with a request/response
  contract retain their existing disposition behavior exactly. Existing
  malformed, unknown, duplicate, mismatched, or non-mutating citations remain
  fail-closed.
- No fake stream, action, generic HTTP route/body escape, credential path, or
  executable command may be introduced. Existing runnable commands must not
  be downgraded. Expected usable CLI-surface delta is zero.

## Scope and source evidence

- Initial local and `origin/main` SHA: `cf29d302c13f7fcd340d31ad6dc27872880ccf42`.
- Target shared seams: `importSourceLockResult`,
  `sourceProjectionApplyNonExecutableMutationDispositions`, and
  `sourceProjectionApplyPartialMutationCoverageDispositions`.
- Target cohorts: the preserved Salesloft and Copper cited-only mutation
  references. Their source locks are intentionally not present in this clean
  worktree; tests will use their real retained locks only if a tracked or
  supplied copy is available. Otherwise small explicit fixtures are limited to
  the irreducible cited-only incompatibility and do not claim provider import.

## Retained cohort inventory observed without materialization

The local Git object for the non-main Batch 6–7 cohort commit
`84255801b` preserves source-lock inventories but not their ignored retained
artifact copies in this worktree. It reports 211 Salesloft REST operations
from `https://developers.salesloft.com/docs/api/` and 89 Copper REST operations
from `https://developer.copper.com/`; both have zero GraphQL operations. The
respective source identities begin `salesloft.rest.delete.V2AccountsId0` and
`copper.rest.delete.V1EntityEntityIdRelated0`. This is read-only scope evidence
for #4291, not a current-main output, source re-pin, or regeneration claim.

## GSD execution mode

The canonical single-worker contract forbids role spawning. The generated
`discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and
`code-review` prompts are therefore executed inline, with this phase record
serving as the durable manual fallback.
