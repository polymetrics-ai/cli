# PLAN: Issue 395 - Humans Need Harnesses Too

## Objective

Publish a repository-grounded blog essay that connects Polymetrics runtime mutation controls with
the issue, test, review, release, and deployment harnesses used by this repository.

## Delivery Context

- Issue: `#395`
- Parent issue: `#293`
- Parent PR: `#346`
- Base branch: `feat/293-blog-annotations`
- Work branch: `docs/395-human-harnesses`
- PR base: `feat/293-blog-annotations`

## Required Skills Used

- `ui-refactoring-docs`: preserve the existing information architecture and write for scanning.
- `vercel-react-best-practices`: keep the static catalog model and avoid new client-side work.
- `vercel-composition-patterns`: use the existing shared article renderer and data contract.
- `golang-documentation`: make only evidence-backed CLI and architecture claims.
- `golang-cli`: describe stdout, stderr, exits, and mutation boundaries as contracts.
- `golang-security`: avoid secrets and distinguish policy from enforced controls.
- `frontend-design` and `web-design-guidelines`: not installed in the current skill set; the
  repository design system, `ui-refactoring-docs`, and direct browser verification are the recorded
  fallback.

## GSD Activation

- `scripts/gsd doctor`: pass.
- `scripts/gsd prompt programming-loop init --phase issue-human-harnesses --dry-run`: failed because
  the repo-local registry does not expose `programming-loop`.
- Fallback: manual GSD loop using this plan, `TDD-LEDGER.md`, and `VERIFICATION.md`.

## Verified Source Set

- Runtime rules and connector inventory under `internal/connectors/defs`.
- Agent delivery contracts and Shepherd validator under `.agents/agentic-delivery`.
- GitHub workflows under `.github/workflows`.
- Current `main` branch-protection status queried through GitHub.
- Existing blog schema and article renderer under `website/lib/blog.ts` and `website/components/blog`.

## Tasks

1. Add a focused failing catalog test for the `human-harnesses` slug and key editorial contract.
2. Add the article to `BLOG_POSTS` without changing the renderer or route architecture.
3. Replace stale inventory figures with reproducible counts from the current parent branch.
4. Explain the repository harness as intent, evidence, deterministic checks, adversarial review,
   release, and production deployment layers.
5. State current gaps honestly: process-level human merge gate, optional review-bot availability,
   and production-environment approval not being equivalent to a required GitHub review count.
6. Run targeted tests, full website checks, build, and desktop/mobile browser verification.
7. Commit and push coherent plan, red, implementation, and review-fix checkpoints.
8. Open a stacked PR with `Refs #395` and `Refs #293`; do not merge the parent PR into `main`.

## Scope Boundaries

- Expected production edit: `website/lib/blog.ts` only.
- Expected test edit: one focused blog catalog unit test.
- Planning artifacts live only in `.planning/phases/issue-395-human-harnesses`.
- No workflow, runtime, auth, comment, database, generated website data, or dependency changes.

## Risks

- Future-dated content is immediately addressable because the catalog has no publication scheduler.
- Workflow prose can overstate checks that run but are not required by branch protection.
- Connector inventory numbers can drift as the parent branch advances.
- A long essay can become hard to scan in the existing article layout.

## Commit Checkpoints

1. Plan and manual-GSD evidence.
2. Red catalog test.
3. Green article implementation and updated verification evidence.
4. Review fixes, if any.

## Requested Narrative Revision

- Reframe the essay around the oversized, roughly million-line pull request that made review,
  conflict resolution, and integration confidence collapse into one merge event.
- Explain how that failure led to the parent issue, bounded sub-issue, isolated worktree, collision
  rule, stacked PR, review-coverage, and human-gated parent architecture.
- Keep the technical workflow evidence, but make the prose sound like a person telling the story
  rather than a repository audit report.
- Remove the closing repository, documentation, and dated inventory bullet list.
- End with a natural request to star the repository.
- Preserve the verified inventory numbers in the technical body where they provide necessary scale.

## Publication-Date And Deployment Follow-Up

- Change only the `human-harnesses` `publishedAt` and `updatedAt` values from the future placeholder
  to `2026-07-16`, the current date for this delivery.
- Extend the focused catalog contract first so any unchanged future date fails before the production
  metadata edit.
- Re-run the focused catalog test, typecheck, production build, and browser assertion that the
  article header renders `2026-07-16`.
- Document the existing deployment path rather than changing it: merge the stacked sub-PR into the
  parent branch, complete parent verification and human review, merge the parent PR into `main`,
  and let Website CI/CD publish and deploy the immutable GHCR image.
- No workflow, runner, VPS, OAuth, database, environment, or deployment mutation is part of this
  follow-up.

## Local Review Loop Follow-Up

- Replace the stale remote Claude-workflow section with the current local exact-head review gate.
- Describe the real local sequence observed in the active Pi sessions: verification, independent
  read-only review, disposition, isolated repair, focused verification, and re-review until clean.
- Explain that moved heads invalidate review evidence, accepted findings are fixed by a separate
  bounded worker, and the correction loop stops for a human when its cap is reached.
- Keep optional remote PR-bot review as supplemental shadow/canary input rather than the default
  delivery gate.
- Remove the detailed Shepherd implementation description. Mention Shepherd only as the next
  engineering story about supervising the loop itself.
- Update the focused catalog contract before editing the article, then run the focused test,
  typecheck, production build, and local browser assertions.
- Live evidence consulted read-only: the active Shepherd proof/recovery and CLI architecture Pi
  sessions, their worktree state, and the local-review contract/workflow artifacts. Warp UI access
  was blocked by the desktop safety layer; no live session was interrupted or mutated.

## Interactive GitHub Evidence Follow-Up

- Replace unsupported narrative-scale wording with canonical GitHub evidence for the abandoned
  precursor PR `#27`, the merged recovery PR `#29`, and merge commit `605b006`.
- Add typed, reusable evidence metadata to the blog catalog and an on-demand client preview that
  refreshes public GitHub metadata without credentials. Verified static values remain available
  when GitHub is unavailable or rate-limited.
- Keep the reader on the article while inspecting evidence in an accessible sheet; provide a
  separate canonical GitHub link for the full page because GitHub denies iframe embedding.
- Add evidence references to the issue-first and parent-orchestration PRs where the article claims
  those structural changes were implemented.
- Add a direct repository CTA. GitHub does not provide a safe GET URL that performs a star, so the
  CTA opens the repository in a new tab at the native Star control instead of pretending an action
  occurred.
- Provide a production-oriented ChatGPT Images brief with six prompts, exact aspect ratios,
  filenames, alt text, and varied top, left, right, between-paragraph, and below-section placements.
  Do not render empty image slots before approved assets exist.

### Design Direction

- Subject: an engineering post-mortem for developers evaluating an agent-assisted repository.
- Job: let a skeptical reader verify the story without losing their reading position.
- Palette: preserve the existing `surface-bg`, `surface-1`, `line-structure`, `line-cta`, and emerald
  action tokens; GitHub state labels use restrained neutral, red, and green semantic accents.
- Type: preserve Instrument Serif for the essay thesis, Chakra Petch for evidence labels, and the
  existing body face for long-form reading.
- Layout: quiet prose first; exact sourced phrases carry compact numbered references; a centered
  evidence dialog holds live metadata and the canonical external action.
- Signature: the source-record dialog behaves like an inspectable case file, with exact SHA,
  status, changed-line totals, and source link rather than a generic preview card.
- Motion: one dialog transition and a refresh-state icon only; respect reduced motion and avoid
  scroll-triggered decoration.

### Required Skills For This Follow-Up

- `frontend-design`: evidence-led editorial direction and one intentional signature interaction.
- `vercel-react-best-practices`: fetch only after interaction, keep static fallbacks, and avoid
  adding route-level data waterfalls.
- `vercel-composition-patterns`: keep evidence preview state in one reusable component rather than
  adding renderer mode flags.
- `web-design-guidelines`: unavailable locally; keyboard, focus, labeling, external-link, mobile,
  and reduced-motion checks are the recorded manual fallback.

## Inline Reference Preview Follow-Up

- Replace the detached, section-level GitHub evidence rails with references attached to the exact
  phrases they support. The phrase and its numbered marker must both be interactive links.
- Assign citation numbers from the stable article evidence order. Repeated evidence keeps the same
  number instead of creating a second source identity.
- Open evidence in a compact centered dialog that traps focus, closes on Escape, restores focus to
  the activating phrase or marker, and exposes the canonical GitHub URL as an explicit new-tab link.
- Keep the on-demand public GitHub refresh and verified static fallback; do not add credentials,
  route-level fetching, or iframe embedding.
- Preserve annotation behavior. If a reader annotation overlaps cited text, its note interaction
  owns the text while the adjacent citation marker remains available for evidence.
- Remove the visual evidence gallery entirely so sourcing reads as editorial typography, not a
  second card layout inside the article.

### Reference Research

- MediaWiki Page Previews keeps contextual exploration attached to the source link and avoids
  forcing navigation away from the article.
- Wikimedia Reference Previews attaches detail to inline footnote markers. Its June 2026 click
  experiment keeps the preview persistent and requires a separate action to visit the reference
  list, which maps to the requested modal plus explicit GitHub action.
- The implementation intentionally uses a dialog rather than hover-only UI so it works on touch,
  keyboard, and pointer input without accidental moving popovers.

### Required Skills For This Follow-Up

- `frontend-design`: editorial citation hierarchy and a restrained evidence dialog.
- `vercel-react-best-practices`: retain interaction-time fetching and static fallback data.
- `vercel-composition-patterns`: extend the existing text renderer and keep one shared dialog state.
- `web-design-guidelines`: unavailable locally; keyboard, focus, label, mobile containment, and
  reduced-motion assertions remain the documented fallback.

## Review Loop Image Follow-Up

- Replace the ASCII review/fix loop in `Review, fix, repeat, locally` with the approved 3:2 asset
  from `~/Downloads/harness blog/8c6bb99b-7ee5-4d41-ae8e-40c25243a8d2.png`.
- Export it as `website/public/blog/human-harnesses/04-review-repair-loop.webp` at quality 88 without
  cropping, recoloring, or adding text.
- Render the figure after body paragraph three and before paragraph four, matching the existing
  image-placement brief and preserving annotation block indexes.
- Use intrinsic dimensions, responsive width, descriptive alt text, and a one-sentence editorial
  caption. Do not add a card, rounded frame, overlay copy, or decorative motion.
- Add the catalog contract before production edits, then verify typecheck, unit tests, focused
  Chromium behavior, production build, and desktop/mobile containment.

## Complete Article Image Set Follow-Up

- Use macOS creation time to map the six UUID-named Downloads PNGs to prompts `01` through `06`.
  Creation order and visual content agree, so no filename guessing or image-content substitution is
  required.
- Preserve every generated composition and export WebP quality 88 using the prompt filenames.
- Place image `01` below the article header; float image `02` left after the first paragraph of `The
  tool after the fire`; float image `03` right after the first paragraph of `The repository became
  a harness`; retain image `04` after paragraph three of the local-review section; place image `05`
  before the release section heading; and place image `06` between the Shepherd teaser and star ask.
- Floats apply only from `md` upward. Mobile keeps every image full width in reading order with its
  original aspect ratio, intrinsic dimensions, alt text, and one-sentence caption.
- Convert the body stack to a flow-root paragraph stream so desktop floats can wrap prose while
  annotation block indexes and exact text anchors remain unchanged.
