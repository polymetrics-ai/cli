# Review — issue-3752-rate-limit-admission-r1

Manual-GSD code-review fallback: Pi cannot run the official reviewer role and the canonical
single-worker contract forbids role spawning. Review scope was the complete `origin/main...HEAD`
diff, with focus on bundle-schema validation, requester retry/admission behavior, error chains,
secret handling, and deferred-slice boundaries.

## Dispositions

| Severity | Finding | Disposition | Evidence |
| --- | --- | --- | --- |
| warning | Selector paths accepted outer whitespace, which could make a declared endpoint selector ambiguous. | accepted | Reject outer whitespace; malformed declaration test passes. |
| warning | A cost `response_header` accepted non-HTTP-field syntax. | accepted | Validate with the existing HTTP header-name pattern; malformed declaration test passes. |
| info | A first production `rate_limits.json` could be added without expanding the explicit production embed pattern. | accepted | `TestProductionDefinitionsEmbedEveryRateLimitDeclaration` now fails that migration until `defs.FS` includes it. |
| info | Provider artifact anchors are valid documentation citations. | verified | Anchor URL loader test passes without relaxing HTTPS or userinfo rejection. |

No critical or unaddressed warning findings remain. No external Claude/Copilot review is claimed:
the task explicitly stops before firstmate directs no-mistakes/PR creation. That future PR must use
the repository’s automatic Claude route and record the resulting coverage separately.
