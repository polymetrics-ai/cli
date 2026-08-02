# Next.js / React / TypeScript rules (cited)

For website/**: Next.js 16 app router, React 19, Fumadocs MDX, Tailwind v4, Radix, vitest +
playwright, pnpm, committed generated artifacts.

## RSC / app router

1. Default to Server Components; `'use client'` only for state, handlers, effects, browser APIs. — nextjs.org
2. Push `'use client'` to the smallest interactive leaf — everything a client file imports joins
   the client bundle. — nextjs.org
3. Interleave server content into client shells via `children`/slots; never import server code
   from a client file. — nextjs.org
4. Guard server-only modules with the `server-only` package so accidental client imports fail the
   build. — nextjs.org
5. Wrap third-party client-only components in your own `'use client'` file before use in RSC. — nextjs.org
6. Mount providers as deep as possible, wrapping only `{children}`. — nextjs.org
7. Parallelize independent server fetches with `Promise.all`; sequential awaits recreate
   waterfalls. — Vercel Academy
8. `<Suspense>` around slow independent sections; stream instead of blocking the route. — nextjs.org
9. Static rendering by default; `force-dynamic` only for genuine per-request freshness. — Next 16 guides
10. `params`/`searchParams` are Promises in Next 15/16 — await them. — Next 16 guides
11. Rely on Turbopack's persistent cache; no hand-rolled build caching. — nextjs.org/blog

## Components / composition

12. Replace boolean-prop piles with compound components (context + subcomponents). — epicreact.dev
13. Principle of Least Privilege: components get only the props they strictly need. — Josh Comeau
14. Use Radix `asChild`/`Slot` to merge behavior onto the consumer's element — no wrapper divs. — radix-ui.com
15. One directory per component (`X.tsx`, helpers, types, `index.ts`). — Josh Comeau
16. `useEffect` only syncs with things OUTSIDE React; never state→state mirroring. — TkDodo
17. Reset-by-identity with `key`, not effects re-syncing on prop change. — TkDodo
18. Before `useState`: can it be derived? Derived values are computed, not stored. — TkDodo
19. A setter only called inside `useEffect` means the state shouldn't exist. — TkDodo

## TypeScript strictness

20. `strict: true` + `noUncheckedIndexedAccess` + `noImplicitOverride`. — Total TypeScript
21. No `enum`; union-of-literals or `as const` objects. — Total TypeScript
22. `unknown` over `any`; narrow before use. — typescriptlang.org
23. Narrowing (discriminated unions, `typeof`/`in`) over `as` assertions. — typescriptlang.org

## Styling / theming (Tailwind v4 + Radix)

24. Semantic color tokens per theme, not `dark:` sprinkled through components. — tailwindlabs
25. v4 dark mode via `@custom-variant dark (&:where(.dark, .dark *));` — no JS `darkMode` config. — tailwindcss.com
26. Inline blocking `<head>` script sets the theme class before first paint (no flash). — tailwindcss.com
27. Theme custom properties live in `@layer theme`, not `@layer base`. — tailwindcss.com
28. Interactive primitives follow Radix's compound pattern (Root context + consuming children). — radix-ui.com

## MDX / content (Fumadocs)

29. Content in `content/docs` collections; let the `docs` collection pair `meta` + `doc`. — fumadocs.dev
30. Schema-validate frontmatter; malformed `title`/`description` should fail the build. — fumadocs.dev
31. No hand-written `# H1` in MDX body; `frontmatter.title` is canonical. — fumadocs.dev
32. `mdx-components.tsx` is the single registry for MDX component overrides. — fumadocs.dev

## Generated-data hygiene

33. Generators are deterministic: stable key ordering, no timestamps/random IDs → byte-identical
    reruns. — idempotency practice
34. Co-version generator with output; CI regenerates and diffs (this repo's website.yml does). — codegen practice
35. Generated files clearly separated (`.generated.ts` suffix / `generated/` dir); never
    hand-edited. — codegen practice

## Testing (vitest + playwright)

36. Vitest for ~80% (utilities, transforms, generators, isolated components); Playwright only for
    critical e2e paths (nav, search, doc rendering). — strapi.io guide
37. Auto-retrying `expect(locator)` assertions; never `waitForTimeout` or sync `isVisible()`. — playwright.dev
38. `getByRole`/`getByText` locators, not CSS/XPath. — playwright.dev
39. Full test isolation (own storage/session per test) via fixtures. — playwright.dev
40. Debug CI failures with the trace viewer, not screenshots alone. — playwright.dev

## Sources

- https://nextjs.org/docs/app/getting-started/server-and-client-components · /docs/app
- https://nextjs.org/blog/next-16-3-turbopack
- https://vercel.com/academy/nextjs-foundations/data-fetching-without-waterfalls
- https://vercel.com/blog/understanding-react-server-components
- https://tkdodo.eu/blog/dont-over-use-state · /blog/putting-props-to-use-state
- https://www.epicreact.dev/soul-crushing-components · /myths-about-useeffect
- https://www.joshwcomeau.com/react/file-structure/ · https://www.joyofreact.com/
- https://www.totaltypescript.com/tsconfig-cheat-sheet · /why-i-dont-like-typescript-enums
- https://www.typescriptlang.org/docs/handbook/2/narrowing.html
- https://tailwindcss.com/docs/dark-mode · https://github.com/tailwindlabs/tailwindcss/discussions/15083
- https://www.radix-ui.com/primitives/docs/guides/composition
- https://www.fumadocs.dev/docs/mdx · /blog/new-conventions
- https://playwright.dev/docs/best-practices
- https://strapi.io/blog/nextjs-testing-guide-unit-and-e2e-tests-with-vitest-and-playwright
