# Design/UI rules for developer-docs sites (cited)

Sources: Web Interface Guidelines (Rauno Freiberg / vercel-labs), Refactoring UI, WCAG 2.2,
thoughtbot design practice, Stripe/Vercel docs teardowns, Baymard.

## Hierarchy & spacing

1. Hierarchy via weight and color before size. — Refactoring UI
2. Start with generous whitespace; remove to "just enough". — thoughtbot
3. Fixed spacing scale (Tailwind tokens), never arbitrary pixels. — Refactoring UI
4. Deliberate grid/baseline alignment; optical ±1px nudges allowed. — WIG
5. Design every density state: empty, sparse, dense, error. — WIG
6. Concentric radii: child radius ≤ parent radius. — WIG

## Color, contrast & themes

7. WCAG AA floors: 4.5:1 text, 3:1 UI components/focus indicators. — WCAG 2.2
8. Prefer APCA where tooling allows; increase contrast on hover/active/focus. — WIG
9. On tinted backgrounds, use low-saturation same-hue text, not flat gray. — Refactoring UI
10. `color-scheme` on `<html>` + matching `<meta name="theme-color">`. — WIG
11. Suppress transitions during explicit theme toggle. — WIG (rauno)
12. Never color-alone status; icon + text label always. — WIG
13. Layered shadows (ambient + direct) + semi-transparent borders for both themes. — WIG

## Accessibility

14. Full keyboard operability; visible `:focus-visible` everywhere; never bare `outline: none`. — WIG
15. Focus indicators ≥3:1 contrast, meaningful perimeter (WCAG 2.4.11/2.4.13). — WCAG 2.2
16. Honor `prefers-reduced-motion`; motion never the sole information carrier. — WCAG 2.2
17. Icon-only controls need `aria-label`; decorative elements get `aria-hidden`. — WIG (rauno)
18. Polite `aria-live` for toasts/async validation results. — WIG
19. Native semantics before ARIA (`<button>`, `<select>`, `<details>`). — WIG
20. Hit targets ≥24px, ≥44px on touch. — WIG
21. No tooltips on disabled buttons (unreachable); explain inline instead. — WIG (rauno)

## Navigation & search

22. Anchored, stratified primary nav (Stripe's nav/content/code three-column as reference). — Moesif teardown
23. URL state mirrors UI state: filters, tabs, pagination, expansion — shareable and
    back/forward-safe. — WIG
24. Back/forward restores scroll position. — WIG
25. Search tolerates partial matches and typos; instant results. — Moesif teardown
26. No dead ends: 404/empty/error states offer a next step. — WIG
27. All navigation is `<a>`/`<Link>` (cmd-click must work); never `<div onClick>`. — WIG

## Code presentation

28. Sans body + distinct monospace for code; `font-variant-numeric: tabular-nums` on numeric
    columns. — WIG; Vercel Geist
29. Code samples stay attached to the prose describing them; one-click language switch. — Moesif teardown
30. `translate="no"` on code tokens, brand/connector names, identifiers. — WIG
31. Flex text containers need `min-w-0` + truncation/line-clamp for long identifiers. — WIG

## Motion

32. Interaction animation ≤ ~200ms; animate only transform/opacity, never layout properties. — WIG
33. Animation is interruptible, input-driven, never autoplaying/unbounded. — WIG
34. `scroll-behavior: smooth` only for in-page anchors; no animation on frequent low-novelty
    actions. — WIG (rauno)

## Typography

35. Body line length 50–75ch (66 optimal), never >80ch — set `max-width` on prose columns. —
    Baymard; WCAG reflow

## Sources

- https://interfaces.rauno.me/ · https://github.com/raunofreiberg/interfaces
- https://github.com/vercel-labs/web-interface-guidelines
- https://www.sglavoie.com/posts/2023/09/09/book-summary-refactoring-ui/ · https://refactoringui.com/
- https://www.w3.org/TR/WCAG22/ · https://www.w3.org/WAI/WCAG22/quickref/
- https://thoughtbot.com/blog/whitespace · https://thoughtbot.com/blog/tags/design
- https://www.moesif.com/blog/best-practices/api-product-management/the-stripe-developer-experience-and-docs-teardown/
- https://vercel.com/geist/typography
- https://baymard.com/blog/line-length-readability
