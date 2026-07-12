# SPEC: Sidebar And Blog Sign-In

## Scope

Fix the website home sidebar layout and auth placement on branch `fix/sidebar-and-blog-signin`.

## Requirements

- The left home sidebar renders as a real docked column on every eligible route, matching the production `cli.polymetrics.ai` behavior: no absolute overlay and no intersection with the main content.
- The sidebar is present on `/`, `/docs`, `/blog`, `/blog/[slug]`, and `/bookmarks`, and hidden at narrow widths by explicit global CSS media rules.
- The blog sidebar keeps the latest-posts group and adds one blog-only auth placement in the sidebar footer.
- The navbar always shows `Get Started` and `Get Demo`; it no longer renders `UserMenu`.
- Blog-only auth uses the existing sign-in dialog, profile dialog, bookmarks link, and sign-out action. It reserves layout while session state loads.
- The article body keeps enough width for reading content plus the marginalia rail.
- Annotation sign-in flows continue to call `requestSignIn()` and open the existing dialog.

## Non-Goals

- No API, auth config, database, or anchor algorithm changes.
- No new dependencies.
- No production deploy.

## Required Skills Loaded

- `frontend-design`
- `web-design-guidelines`
- `vercel-react-best-practices`
- `vercel-composition-patterns`
- `agent-browser`
- `gsd-programming-loop`

## GSD Command Path

- `scripts/gsd doctor` passed.
- `scripts/gsd prompt programming-loop init --phase sidebar-and-blog-signin --dry-run` failed because the adapter does not expose `programming-loop`.
- Manual fallback: `node /Users/karthiksivadas/.codex/skills/gsd-programming-loop/scripts/programming-loop.mjs init --phase sidebar-and-blog-signin`, then `run --phase sidebar-and-blog-signin --mode auto`.
