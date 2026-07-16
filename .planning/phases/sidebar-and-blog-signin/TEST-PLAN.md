# TEST PLAN: Sidebar And Blog Sign-In

## Red Tests

- `blog-smoke.spec.ts`: navbar keeps `Get Demo` on blog routes and blog auth appears only in the sidebar.
- `blog-annotations.spec.ts`: hydration/profile actions use the blog sidebar auth placement instead of `header`.
- `docs-smoke.spec.ts`: homepage hydration waits no longer rely on a header sign-in button.

## Verification Commands

```bash
npx -y pnpm@11.7.0 typecheck
DATABASE_URL=postgres://website:dev@localhost:55432/website npx -y pnpm@11.7.0 test:unit
PLAYWRIGHT_BASE_URL=http://localhost:3100 DATABASE_URL=postgres://website:dev@localhost:55432/website npx -y pnpm@11.7.0 test:e2e
npx -y pnpm@11.7.0 build
```

## Visual / Browser Checks

- Use `agent-browser` at widths `1560`, `1440`, `1280`, `1024`, `768`, and `390`.
- Capture `/`, `/docs`, `/blog`, `/blog/one-cli-to-rule-them-all`, and `/bookmarks`.
- Assert sidebar/content bounding boxes do not intersect where the sidebar is visible.
- Assert navbar CTAs and blog-only auth presence.
- Sign up through the test API, reload a blog route, and confirm account card state.
