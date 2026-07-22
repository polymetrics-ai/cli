# PLAN: Sidebar And Blog Sign-In

## Design Direction

Use the existing boxy emerald system, production sidebar structure, and documentation-site patterns from Stripe/Vercel/Linear: persistent left navigation at desktop widths, dense groups, explicit content columns, and a compact account block inside navigation chrome rather than the top bar.

## Tasks

1. Add failing e2e coverage for the new navbar/sidebar/auth requirements.
2. Fix sidebar primitive/CSS so desktop visibility comes from explicit media classes and stat rows cannot collide.
3. Re-add the sidebar shell to blog index, article, and bookmarks views.
4. Move auth out of the navbar and into a blog/bookmarks-only sidebar footer card.
5. Rebalance article layout so content and marginalia fit with the left sidebar.
6. Update hydration/profile e2e selectors to target the new blog auth placement.
7. Verify with typecheck, unit tests, e2e tests, build, and agent-browser screenshots/DOM assertions.
8. Follow-up: mount the right "On this page" rail alongside the existing left rail on every public page shell, including blog, article, bookmarks, changelog, and patterns.
9. Follow-up: move the blog discussion/sign-in card out of the left sidebar and into the right rail above the footer, then make the footer discussion link the page-specific GitHub Discussion action.

## Files Expected To Change

- `website/components/home/home-sidebar.tsx`
- `website/components/ui/sidebar.tsx`
- `website/components/home/navbar.tsx`
- `website/components/home/page-aside.tsx`
- `website/components/blog/bookmarks-view.tsx`
- `website/components/blog/article-body.tsx`
- `website/app/(home)/blog/page.tsx`
- `website/app/(home)/blog/[slug]/page.tsx`
- `website/app/globals.css`
- `website/tests/e2e/blog-smoke.spec.ts`
- `website/tests/e2e/blog-annotations.spec.ts`
- `website/tests/e2e/docs-smoke.spec.ts`

## Risks

- Tailwind v4 cascade can override responsive display utilities.
- Blog article may become cramped at 1280px if the margin rail appears too early.
- Hydration waits need a real client-rendered marker now that the navbar auth slot is gone.
