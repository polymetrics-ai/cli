# TDD Ledger

Phase: `issue-395-human-harnesses`

## Baseline

- The parent catalog contains three posts and no `human-harnesses` slug.
- A live route baseline could not be captured because no process was listening on port 3100.
- The red test will therefore exercise the catalog directly and fail before the production edit.

## Red: Human Harnesses Catalog Contract

- Status: complete.
- Test: resolve `human-harnesses` through `getBlogPost` and require the verified inventory figures,
  delivery-harness section, current-limitations section, and August 4 publication metadata.
- Command: `npx -y pnpm@11.7.0 exec vitest run tests/blog-catalog.test.ts`
- Result: failed as expected because `getBlogPost('human-harnesses')` returned `undefined`.

## Green: Human Harnesses Catalog Contract

- Status: complete.
- Command: `npx -y pnpm@11.7.0 exec vitest run tests/blog-catalog.test.ts`
- Result: passed after adding the post to the existing `BLOG_POSTS` catalog.
- Scope: no renderer, route, component, dependency, auth, or database changes were needed.

## Refactor

- Status: complete.
- Replaced stale operation counts with a fresh API-surface inventory.
- Corrected the draft's Shepherd description: connector workers perform surface mapping; the
  Shepherd-style layer independently validates orchestration decisions.
- Distinguished workflows that run from status checks currently required by `main` branch
  protection.
- Reframed August 4 behavior as a release target and separated inventory, implementation,
  conformance, and live certification.
- Kept the existing renderer and static catalog architecture unchanged.

## Review Fix: Method Classification

- Parent-orchestrator audit identified that treating every non-GET/HEAD operation as a mutation
  mislabeled hook, wildcard, GraphQL, WebSocket, and composite-method rows.
- Updated the article and catalog contract to report 14,780 GET reads, 3 HEAD checks, 14,169
  explicit HTTP mutations, and 177 mixed/nonstandard rows.
- Tightened the article's PR Issue Guard and GSD workflow descriptions to their actual enforcement
  boundaries.
- Replaced an unsupported reference to signed release artifacts with the versioned GoReleaser
  artifacts the repository actually produces.
- Preserved the supplied draft's connection between the article annotation UI and a durable human
  feedback loop.
- Replaced the draft-style standalone approval command with the documented reverse ETL contract:
  human-readable plan output supplies the token consumed by `pm reverse run --approve`.
- Verification: catalog test, typecheck, six blog Playwright tests, production build, and
  `git diff --check` passed after the review fix.

## Revision: Million-Line PR Narrative

- Status: green complete.
- Red contract: require the merge-nightmare story, structural parent/sub-issue architecture, and a
  human star request; reject the repository/documentation/inventory footer requested for removal.
- Red command: `npx -y pnpm@11.7.0 exec vitest run tests/blog-catalog.test.ts`.
- Red result: failed because `The PR that ate the repository` and the revised narrative contract
  were absent.
- Green contract: preserve verified technical claims and CLI examples while the revised narrative
  passes the catalog, browser, typecheck, and production-build checks.
- Green result: the catalog contract passes with the million-line PR story, isolated-worktree
  architecture, star request, and removed footer; typecheck and `git diff --check` also pass.
- Regression result: all 64 website unit tests and all 6 focused blog Playwright tests pass.
- Browser result: desktop and mobile assertions find no horizontal overflow, show the revised first
  and final headings, include the star request, and exclude the removed footer.
- Build result: the production Next.js build passes and prerenders `/blog/human-harnesses`.
- Reading-time check: approximately 3,059 source words, recorded as a 14-minute read at 220 words
  per minute.

## Follow-Up: Current Publication Date

- Status: green complete.
- Red contract: expect only `human-harnesses` to publish and update on `2026-07-16` while preserving
  its existing content contract.
- Red command: `npx -y pnpm@11.7.0 exec vitest run tests/blog-catalog.test.ts`.
- Red result: failed with the catalog still returning `2026-08-04` for both fields, exactly isolating
  the requested metadata delta.
- Green result: the same catalog contract passes after changing only the two `human-harnesses`
  metadata fields to `2026-07-16`; website typecheck also passes.
- Render result: a headless Chromium assertion against the local article finds `2026-07-16` and
  rejects the previous `2026-08-04` placeholder.
- Build result: the production Next.js build passes and prerenders `/blog/human-harnesses`. The
  standalone build reports that `BETTER_AUTH_SECRET` was not supplied; the static article build is
  still successful and no secret value was read or printed.
- GSD activation: `scripts/gsd doctor` passed; `scripts/gsd prompt programming-loop init --phase
  issue-395-human-harnesses --dry-run` remains unavailable because `programming-loop` is absent
  from the repo-local registry, so the recorded manual-GSD loop continues.

## Follow-Up: Local Review And Repair Loop

- Status: red complete.
- Red contract: require a local exact-head review section, read-only reviewer, isolated repair
  worker, correction loop, and Shepherd next-post teaser; reject the stale Claude workflow copy and
  detailed Shepherd implementation description.
- Red command: `npx -y pnpm@11.7.0 exec vitest run tests/blog-catalog.test.ts`.
- Red result: failed because the catalog still contained `Reviewing untrusted code without trusting
  it` instead of `Review, fix, repeat, locally`, before reaching the new body assertions.
- Green target: preserve the surrounding verification, human merge, and deployment claims while
  accurately reflecting the current Pi local-review contract and active-session evidence.
