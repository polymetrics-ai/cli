# Discussion log — #4366

## Inline/manual discussion fallback

The GSD `discuss-phase` prompt was resolved on 2026-08-27. This direct-PR task is already bounded by the launch brief and Batch 1 foundation plan, and the repository contract forbids spawning separate GSD roles for this shared foundation. The worker therefore applied the documented inline/manual fallback.

| Topic | Decision | Evidence |
| --- | --- | --- |
| Scope | Exactly the 608 `typed_input_schema` composition records: `oneOf` 572, `anyOf` 4, `allOf` 32. | Batch 1 plan P8 and manifest at `d842f739c`. |
| Provider cohort | Bitbucket 30, CircleCI 3, GitLab 542, Jira 1, Vercel 32. | Manifest query against `records[]`, each retaining source URL, SHA, bytes, and source location. |
| Admission | A closed contract alone is insufficient. Keep all records `missing_foundation` unless the lane, binding, runtime preflight, and command surface are already exact. | Declaration admission and commandrunner are the executable gate. |
| Security | Reject malformed, external, unresolvable, cyclic, duplicate/ambiguous, contradictory, or open composition locally. No credentials, provider calls, generic body, generic HTTP, or inferred fields. | Launch brief and AGENTS.md safety overlay. |
| Command proof | Count every newly credential-bound runnable command exactly; expected baseline is zero unless a current closed, pre-existing lane passes all admission checks. | Isolated project and built `pm` proof required for each nonzero promotion. |
