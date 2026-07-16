# Verification Checklist

## Source Claims

- [x] Count connector definition directories.
- [x] Count API-surface endpoints by read, mutation, and delete method.
- [x] Count declared streams.
- [x] Inspect issue guard, GSD, verification, security, Claude, release, and website workflows.
- [x] Inspect current `main` branch-protection required checks.
- [ ] Re-run all inventory counts before final commit.

## Focused Checks

- [ ] Catalog unit test passes.
- [ ] Blog smoke test includes and opens every catalog post.
- [ ] `/blog` links to `/blog/human-harnesses`.
- [ ] `/blog/human-harnesses` renders the complete article.

## Website Gates

- [ ] `pnpm run typecheck`
- [ ] `pnpm run test:unit`
- [ ] Focused Playwright blog smoke tests.
- [ ] `pnpm run build`
- [ ] `git diff --check`

## Visual Checks

- [ ] Desktop article screenshot: header, body, left navigation, marginalia, and right rail do not
  overlap.
- [ ] Mobile article screenshot: title and prose fit, navigation remains usable, and no horizontal
  overflow appears.
- [ ] Blog index card text and metadata fit at desktop and mobile widths.

## Safety And Delivery

- [x] No secrets read or changed.
- [x] No runtime mutation or production deployment.
- [ ] Stacked PR targets `feat/293-blog-annotations` and uses `Refs`, not `Closes`.
- [ ] Automated review coverage or documented parent/human fallback is recorded.
- [ ] Parent merge into `main` remains human-gated.

