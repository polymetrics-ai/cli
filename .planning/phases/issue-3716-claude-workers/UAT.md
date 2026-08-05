# UAT — issue #3716 clean project-local Claude workers

The coverage-aware `SUMMARY.md` classifies all four deliverables as automated/non-judgmental.

| Deliverable | Evidence | Result |
|---|---|---|
| D1 generated workers | frontmatter/isolation unit test and canonical check | pass |
| D2 blocked delegation | scoped-skill/denylist tests and prior trusted-home forced-Agent smoke | pass |
| D3 canonical ownership | deterministic renderer test and make gate | pass |
| D4 documented/scope boundary | generated policy and changed-path audit | pass |

The clean-home runtime smoke is **NOT PERFORMED**. The unauthenticated fixture did not run a model
and is not selection evidence. `VERIFICATION.md` separates that missing runtime criterion from the
static generated-file, documented-precedence, scoped-skill, and drift-enforcement proof.

Verdict: passed under the captain-authorized clean-home smoke deferral. No implementation gap
required `plan-phase --gaps` or `execute-phase --gaps-only`.
