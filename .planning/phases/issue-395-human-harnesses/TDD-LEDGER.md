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

- Status: green complete.
- Red contract: require a local exact-head review section, read-only reviewer, isolated repair
  worker, correction loop, and Shepherd next-post teaser; reject the stale Claude workflow copy and
  detailed Shepherd implementation description.
- Red command: `npx -y pnpm@11.7.0 exec vitest run tests/blog-catalog.test.ts`.
- Red result: failed because the catalog still contained `Reviewing untrusted code without trusting
  it` instead of `Review, fix, repeat, locally`, before reaching the new body assertions.
- Green target: preserve the surrounding verification, human merge, and deployment claims while
  accurately reflecting the current Pi local-review contract and active-session evidence.
- Green result: the catalog contract passes with `Review, fix, repeat, locally`, exact-head review,
  a read-only reviewer, bounded disposition and repair, a four-round cap, supplemental remote-bot
  language, and the Shepherd next-post teaser.
- Regression result: all 64 website unit tests, website typecheck, and `git diff --check` pass.
- Build result: the production Next.js build passes and prerenders `/blog/human-harnesses`; the
  standalone build again reports that no local `BETTER_AUTH_SECRET` was supplied without exposing a
  value.
- Browser result: desktop and mobile Chromium assertions render the new section, omit the stale
  Claude workflow language, and find no horizontal overflow.

## Follow-Up: Interactive GitHub Evidence

- Status: green complete.
- Red catalog contract: require canonical PR `#27`, PR `#29`, merge commit `605b006`, issue-first
  PR `#47`, parent-orchestrator PR `#51`, section-to-evidence references, and a repository star CTA.
- Red browser contract: require an evidence marker to open an in-page sheet with a verified
  fallback, canonical external link, close behavior, and no same-tab navigation.
- Green target: fetch public GitHub metadata only after opening a preview; preserve the static
  evidence snapshot on network or rate-limit failure; keep comments and text-selection behavior
  unchanged.
- Refactor target: isolate the evidence UI in one client component and keep the article renderer
  data-driven.
- Visual target: evidence remains subordinate to prose, works at 390px and 1440px, has visible
  keyboard focus, and introduces no horizontal overflow.
- Red result: the focused catalog test failed at the first new assertion because the post did not
  yet expose the repository CTA; no production code had been changed.
- Red command: `npx -y pnpm@11.7.0 exec vitest run tests/blog-catalog.test.ts`.
- Green result: the focused catalog contract passes with canonical PR, commit, workflow, section
  reference, and repository CTA metadata.
- Interaction result: configured Chromium passed the offline evidence-preview path, exact snapshot
  values, canonical new-tab link, Escape close, and same-tab preservation.
- Live-path evidence: GitHub's public PR endpoint returned HTTP 200 with
  `access-control-allow-origin: *`; the client refresh is cached per evidence item and uses no token.
- Regression result: all 64 website unit tests and website typecheck pass.
- Build result: the production Next.js build passes and prerenders `/blog/human-harnesses`. The
  standalone build reports the existing default-secret warning because no build-only auth secret
  was provided; no secret value was read or printed.
- Visual result: the desktop evidence sheet screenshot is clear, fully contained, and keeps the
  article visible behind it. The in-app browser refused a localhost refresh under its URL policy;
  remote Website CI passed the 390px containment and horizontal-overflow assertion.
- Remote correction round 1: Website CI reached the new evidence test and failed only the added
  focus-return assertion. The controlled sheet was not mounted through a Radix trigger, so its
  external marker remained inactive after Escape. The marker callback now records the activating
  button and the close path returns focus explicitly on the next animation frame. The same remote
  test passed focus return and then continued through the 390px containment assertion. Website
  checks and the website image build both passed; deployment was correctly skipped for the stacked
  pull request.

## Follow-Up: Inline Reference Preview

- Status: green complete.
- Catalog contract: every evidence source is referenced by an exact body phrase and block index;
  every phrase resolves in its paragraph; the section-level `evidenceIds` gallery contract is gone.
- Browser contract: no GitHub evidence trail is rendered; linked claim text and its numbered
  citation both open a centered dialog; the dialog retains verified fallback data, canonical source
  action, Escape close, focus return, and 390px containment.
- Regression contract: text annotations, repository star CTA, static prerendering, and public
  interaction-time GitHub refresh remain unchanged.
- Red command: `npx -y pnpm@11.7.0 exec vitest run tests/blog-catalog.test.ts`.
- Red result: failed at the assertion that section-level `evidenceIds` are absent; the current
  catalog still exposes the detached evidence-gallery model.
- Green result: every evidence source is attached to an exact body phrase and receives its stable
  evidence-order number; repeated website-workflow references reuse citation `[11]`.
- Interaction result: the linked phrase and superscript citation both open the same centered source
  dialog, Escape restores focus to the exact trigger, modified clicks retain native new-tab GitHub
  navigation, and no detached evidence trail remains.
- Visual result: desktop review shows readable inline citations and a compact source-record dialog;
  the browser contract confirms centered 390px containment and no horizontal overflow.
- Regression result: 64 website unit tests, all 7 blog Chromium tests, website typecheck, focused
  catalog test, production build, and `git diff --check` pass.
- Build result: `/blog/human-harnesses` prerenders successfully. Better Auth reports the existing
  default-secret warning during static collection because no build-only secret was supplied; no
  secret value was read or printed.

## Follow-Up: Review Loop Image

- Status: red captured; green pending.
- Catalog contract: the local review section no longer exposes its ASCII loop as `code`; it declares
  the approved review-repair image, intrinsic 3:2 dimensions, placement after paragraph three,
  descriptive alt text, and one-sentence caption.
- Browser contract: the image renders in reading order, remains contained at 390px, and the removed
  ASCII loop does not appear in the article.
- Red command: `npx -y pnpm@11.7.0 exec vitest run tests/blog-catalog.test.ts`.
- Red result: failed because the review section still returned the exact ASCII loop in `code`.
