---
name: ts-website
description: TypeScript/Next.js standards for the website — RSC/app-router boundaries, component composition, TS strictness, Tailwind v4 theming, Fumadocs MDX, generated-data hygiene, vitest/playwright testing. Load BEFORE any website/** work; rules cited to Next.js/Vercel, TkDodo, Epic React, Total TypeScript, Tailwind, Radix, Fumadocs, Playwright docs.
---

# Website implementation standards (Next.js 16 + Fumadocs)

Load before implementing or reviewing anything under `website/**`. Full cited rules:
[references/ts-rules.md](references/ts-rules.md). Cite rule numbers in findings/dispositions.

## Repo overlay

- **Generated-data rules are gates here** (#33–35): `pnpm run gen:website-data` must be
  byte-idempotent (CI regenerates and fails on diff over the 6 `*.generated.*` artifacts +
  icons). Generators stay deterministic — stable ordering, no timestamps. Never hand-edit a
  `.generated.ts` file.
- **CI runs typecheck + vitest + playwright + build — but NOT `pnpm lint`**; run it locally
  before handoff, treat its findings as blocking.
- **No new frontend deps without human approval** (parity-contract hard stop). Radix/shadcn +
  Tailwind v4 + Fumadocs are the sanctioned stack.
- **Server-first** (#1–2): docs pages are static; `'use client'` only at interactive leaves
  (search, theme toggle, catalog filters). Catalog filter state belongs in the URL
  (see design-ui skill #23).
- **MDX/frontmatter** (#29–32): content under the Fumadocs collections; schema-validated
  frontmatter; no in-body H1; component overrides only via `mdx-components.tsx`.

## Verification before handoff

`pnpm typecheck` · `pnpm lint` · `pnpm test:unit` · `pnpm run gen:website-data` twice with no
diff · `pnpm build` when routing/layout changed. Record results in the handoff.
