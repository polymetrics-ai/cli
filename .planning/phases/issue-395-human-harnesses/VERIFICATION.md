# Verification Checklist

## Source Claims

- [x] Count connector definition directories.
- [x] Count API-surface endpoints by read, mutation, and delete method.
- [x] Count declared streams.
- [x] Inspect issue guard, GSD, verification, security, Claude, release, and website workflows.
- [x] Inspect current `main` branch-protection required checks.
- [x] Re-run all inventory counts before final commit.

## Focused Checks

- [x] Catalog unit test passes.
- [x] Blog smoke test includes and opens every catalog post.
- [x] `/blog` links to `/blog/human-harnesses`.
- [x] `/blog/human-harnesses` renders the complete article.

## Website Gates

- [x] `pnpm run typecheck`
- [x] `pnpm run test:unit` - 64 tests passed.
- [x] Focused Playwright blog smoke tests - 6 tests passed.
- [x] `pnpm run build` - passed and prerendered `/blog/human-harnesses`.
- [x] `pnpm run gen:website-data` - passed with no generated drift.
- [x] `git diff --check`

## Visual Checks

- [x] Desktop article screenshot: header, body, left navigation, marginalia, and right rail do not
  overlap.
- [x] Mobile article screenshot: title and prose fit, navigation remains usable, and no horizontal
  overflow appears.
- [x] Blog index card text and metadata fit at desktop and mobile widths.

## Safety And Delivery

- [x] No secrets read or changed.
- [x] No runtime mutation or production deployment.
- [x] Stacked PR `#396` targets `feat/293-blog-annotations` and uses `Refs`, not `Closes`.
- [x] Automated review is unavailable; parent/human fallback is recorded in the PR body.
- [x] Parent merge into `main` remains human-gated.

## Residual Parent-Branch Findings

- The blog smoke suite passes, but navigating to bookmarks can log an existing auth-card hydration
  mismatch. Issue 395 does not change that component.
- Mobile floating note controls can temporarily cover the lowest line at the viewport edge. The
  article itself has no horizontal overflow; changing the shared annotation controls is outside this
  content-only issue.
- The post is dated August 4 but is immediately visible because the current static catalog has no
  publication scheduler. This matches the requested local preview and must be considered when the
  parent PR is merged.
- The final build loaded the existing local environment without printing any value and completed
  without auth warnings.

## Requested Narrative Revision

- [x] Opens with the roughly million-line, single-PR merge failure that motivated the architecture.
- [x] Connects that failure to parent issues, bounded sub-issues, isolated worktrees, stacked PRs,
  layered verification, and a human-gated parent merge.
- [x] Keeps the technical claims and CLI examples while using a more conversational, lightly funny
  narrative voice.
- [x] Removes the repository, documentation, and inventory-snapshot footer.
- [x] Ends with a direct request to star the repository.
- [x] Updates the catalog contract and reading time to 14 minutes.
- [x] Passes 64 unit tests, 6 focused blog Playwright tests, typecheck, and production build.
- [x] Passes desktop/mobile content assertions with no horizontal overflow.

## Publication-Date Follow-Up

- [x] Focused catalog test fails against the future placeholder date.
- [x] Only `human-harnesses` uses `publishedAt: '2026-07-16'` and `updatedAt: '2026-07-16'`.
- [x] Focused catalog test and website typecheck pass.
- [x] Production build prerenders `/blog/human-harnesses`.
- [x] Local article header renders `2026-07-16` and does not render `2026-08-04`.
- [x] Deployment remains human-gated and no production environment is mutated by this change.

## Deployment Readiness Snapshot

- [x] `WEBSITE_DEPLOY_ENABLED` currently evaluates to `true` without exposing variable contents.
- [x] The `polymetrics-origin-website` runner is online and has the required self-hosted, Linux,
  Tailscale, and website labels.
- [x] The `website-production` environment exists; it currently has no environment protection
  rules, so parent-PR human review remains a process gate rather than a GitHub environment gate.
- [x] The expected website deployment variable names exist.
- [x] OAuth and database configuration remains in the VPS-side mode-0600 environment files, not in
  the GitHub repository or Actions secret store.

## Local Review Loop Follow-Up

- [x] Focused catalog test fails while the stale remote-review section remains.
- [x] Article describes exact-head local review, read-only findings, disposition, isolated fixes,
  verification, and re-review.
- [x] Remote PR-bot review is supplemental rather than the default delivery gate.
- [x] Shepherd implementation detail is removed and only the next-post teaser remains.
- [x] Focused catalog test, typecheck, production build, and `git diff --check` pass.
- [x] Local article renders the new section and omits the stale Claude workflow language.
- [x] Desktop and mobile Chromium assertions find no horizontal overflow.
- [x] No Warp/Pi session, GitHub workflow, runtime, or production environment is mutated.

## Interactive GitHub Evidence Follow-Up

- [x] Catalog contract covers canonical PR, commit, structural-harness, and repository URLs.
- [x] Focused unit test fails before the evidence model is implemented and passes afterward.
- [x] Evidence markers open an in-page sheet and preserve the article scroll position.
- [x] The sheet exposes exact status/SHA/change totals and a canonical external GitHub link.
- [x] Public GitHub refresh has a verified static fallback and does not use credentials.
- [x] Repository CTA opens `polymetrics-ai/cli` in a new tab at GitHub's native Star control.
- [x] Escape, close button, focus return, and external-link labels pass remote Website CI after the
  explicit controlled-sheet focus fix.
- [x] Remote Website CI passes the new 390px sheet containment and overflow assertion.
- [x] Reduced-motion rendering remains usable through `motion-reduce:animate-none`.
- [x] Image prompt guide records six varied placements without rendering broken placeholders.
- [x] Focused e2e passes remotely after the focus-return fix; 64 unit tests, typecheck, production
  build, and `git diff --check` pass across the local and remote gates.

## Inline Reference Preview Follow-Up

- [x] Focused catalog test fails before the reference model replaces section evidence IDs.
- [x] Every reference phrase resolves inside its declared article paragraph.
- [x] No detached GitHub evidence trail remains in rendered article markup.
- [x] Linked phrase and numbered citation open the same centered evidence dialog.
- [x] Dialog retains verified fallback, live refresh, canonical GitHub link, Escape close, and focus
  restoration for both trigger forms.
- [x] Desktop and 390px layouts have no horizontal overflow.
- [x] Focused unit test, full website unit tests, typecheck, production build, and `git diff --check`
  pass.

## Review Loop Image Follow-Up

- [x] Focused catalog test fails before the ASCII loop is replaced.
- [x] Approved Downloads asset is exported to WebP at its original 3:2 composition.
- [x] Figure renders after paragraph three with intrinsic dimensions, alt text, and caption.
- [x] ASCII review loop is absent from catalog and rendered article.
- [x] Desktop and 390px layouts preserve reading order and avoid horizontal overflow.
- [x] Unit tests, typecheck, focused Chromium test, production build, and `git diff --check` pass.

## Complete Article Image Set Follow-Up

- [x] Creation-time order and visual subjects map UUID assets to prompts `01` through `06`.
- [x] Focused catalog test fails while only image `04` is declared.
- [x] Six WebP assets exist at their prompt filenames with original dimensions.
- [x] Lead, left float, right float, review loop, release lead, and teaser placements match the brief.
- [x] All six images decode at 390px, 768px, 1024px, and 1440px with alt text and captions.
- [x] Annotation block indexes and inline evidence interactions remain unchanged.
- [x] Unit tests, typecheck, blog Chromium suite, production build, and `git diff --check` pass.
