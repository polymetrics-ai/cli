# Issue #4329 — source-cited read-only declarations

## Discuss-phase decision record

The source-projection validator currently rejects every source-cited read
without an executable route unless its source descriptor carries a missing
foundation gap. That combines an implementation limitation with a deliberate
product boundary and makes an intentionally read-only source operation
impossible.

The required declaration is endpoint-local, in the existing `api_surface`
operation ledger: `model: "read_only"`, `status: "blocked"`,
`blocked_by_default: true`, a non-empty reason, and the exact named-policy note
`Named policy: source-cited-read-only-operations-r1`. It is effective only
when source evidence identifies the same non-mutating method/path. It is not
inferred from an exclusion, a human-readable reason, or a missing runtime
declaration. The declaration is invalid when the source operation is POST,
PUT, PATCH, or DELETE, when it has an executable route, or when the source
descriptor carries a foundation gap.

The mutation case is explicitly out of scope: a cited POST/PUT/PATCH/DELETE
without a complete action remains a source-projection finding for
`cli-mutation-disposition-foundation-r1`. Naming one `read_only` would hide a
real write from both coverage and the write-safety boundary.

The operation-evidence artifact must project a distinct `read_only` state and
a separate rollup from `missing_foundations`; otherwise its existing
`runtime_reachability`/CLI gaps would incorrectly make an intentional refusal
look like an unimplemented foundation. Sentry's and Vercel's source-cited
mutations receive no `read_only` declaration in this issue; their handoff list
is recorded for the sibling foundation.

## Inline GSD fallback

`scripts/gsd` resolved `discuss-phase`, `plan-phase --tdd`, `execute-phase`,
`verify-work`, and `code-review`. This direct-PR worker cannot use compatible
isolated GSD roles, so the lifecycle is being executed inline with durable
phase artifacts.
