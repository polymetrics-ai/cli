# Inline code review — issue 4336

## Method

`scripts/gsd sources code-review` resolved the official GSD code-review command
on 2026-08-23. Its normal reviewer subagent is unavailable in this task's
single-worker runtime, so this is the documented inline/manual fallback.

Scope reviewed:

- `e338cd301..318fb58e8` production and targeted-test diff;
- the follow-up test expectation correction and planning evidence in the
  working tree;
- the relevant source-import call paths, schema compiler/validator, and the
  complete changed-package test result.

## Findings

### Resolved before review completion

- **Warning — stale negative expectations:** the full
  `./cmd/connectorgen` suite initially showed that the generic invalid-form
  test still required an abort for an unresolved **response-schema** reference
  and a missing path parameter. That would have described the old contract,
  not the new retain-and-trace contract. The test now requires one preserved,
  merge-blocked operation, a source-located gap, the exact foundation, and the
  original provider error text. Focused and package tests pass.

### No unresolved findings

- Retention is gated by typed errors whose expected grammar is `schema`, and
  is used only in response preflight/response descriptor handling; external
  references and non-schema malformed positions still return errors.
- A malformed path receives no invented binding: the descriptor is preserved,
  its runtime is merge-blocked, and its gap is source-located.
- The raised depth limit is finite (64); a 65-level schema remains refused.
- `patternProperties` compiles sorted regular expressions, validates every
  matching schema, and retains the `additionalProperties: false` rejection for
  unmatched keys. `example` remains a non-validating annotation.
- No production connector-name branch or connector-definition edit was added.

## Verdict

No unresolved critical, warning, security, or quality finding remains in this
scope. The full provider-artifact import is deliberately **not** claimed green:
`REQUEST-CONTRACT-INVENTORY.md` records its separately scoped blockers.
