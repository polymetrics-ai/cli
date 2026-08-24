# Context — issue 4336 provider dialect tolerance

## Task Delivery Header

- Issue: Refs #4336 — tolerate bounded provider OpenAPI dialect gaps
- Base branch: main
- Merges into: main
- Delivery: Pull request open against `main`, committed on
  `fm/cli-provider-dialect-tolerance-foundation-r1`, with required local checks
  green and GitHub API base read-back equal to `main`.
- Working branch: fm/cli-provider-dialect-tolerance-foundation-r1
- Task: Make the shared OpenAPI source importer accept legitimate provider
  dialect constructs, and retain malformed or safely bounded-unrepresentable
  constructs as source-traced gaps, without changing any connector definition.
- Verification: Red/green real-importer behavioral tests for each named
  provider document, byte-identical regression projection, a pathological depth
  refusal, focused and repository generation/static checks, and inline review.

## Decided handling

| Provider case | Decision | Reason |
| --- | --- | --- |
| Bitbucket pull-request comment response schema depth | raise a finite schema bound after measuring its real depth | The declared response is finite legitimate OpenAPI; the guard remains required against resource exhaustion. |
| Notion meeting-notes response schema depth | raise the same finite schema bound only if the measured maximum requires it | This is the same generic source-schema capability, not connector behavior. |
| Stripe `GET /v1/account` reference depth | raise a finite reference bound after measuring the real chain | It is a valid finite local reference chain; cycle and pathological checks remain active. |
| Vercel response `patternProperties` | support as a bounded schema construct | It is standard JSON Schema/OpenAPI schema syntax; descriptor retention must remain exact and bounded. |
| Docker Hub `#/components/responses/team_repo` dangling reference | retain and trace | The reference is malformed; the affected operation must survive but be merge-blocked with its pointer and response location. |
| Docker Hub SCIM `example` schema keyword | support as a schema annotation | `example` is an OpenAPI schema annotation and must no longer make a valid declarative schema unloadable. |
| GitLab `epic_issue_id` missing required path parameter | retain and trace | The provider path contract is malformed; preserve the operation and source evidence rather than inventing a parameter or dropping it. |

## Bound and safety constraints

- Measure the named provider paths first; record the old limit, observed depth,
  and proposed finite replacement before production code changes.
- Preserve existing limits for count, bytes, expansion, cycles, unsafe pointers,
  dynamic schemas, and semantic reference siblings.
- No connector-specific branches or edits under `internal/connectors/defs/`.
- Existing importable source projections must remain byte-identical.

## Inline GSD fallback

The adapter lifecycle prompts for `discuss-phase`, `plan-phase --tdd`,
`execute-phase`, `verify-work`, and `code-review` were resolved. This direct-PR
worker executes them inline because compatible isolated GSD roles are unavailable
and the repository delivery contract forbids role spawning in this lane.
