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
