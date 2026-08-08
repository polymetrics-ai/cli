# TDD ledger — seven connector extraction r1

## Plan checkpoint

- **GSD:** `discuss-phase --auto` and `plan-phase --tdd` prompts generated through the project
  adapter; execution is inline/manual for the documented single-worker reason in `PLAN.md`.
- **Red:** pending — import the seven connector-specific acceptance tests from `c28bc75a3` while
  retaining current-main bundle inputs, then capture their focused failure output.
- **Green:** pending — import exactly the seven authored bundle deltas and regenerate derived
  surfaces/docs/data.
- **Refactor:** pending — review generated diff for allowlist compliance and generator drift.

## Required red command

```bash
go test -timeout 20m ./cmd/connectorgen -run 'Test(Chatwoot|Gmail|Greenhouse|HelpScout|Jira|LeverHiring|WorkdayRest)'
```

The red result must be recorded verbatim enough to distinguish an assertion failure from a missing
test, compilation failure, or unrelated baseline failure. No production bundle input may change
before that capture.

