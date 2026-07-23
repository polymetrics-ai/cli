---
name: design-ui
description: Design/UI quality standards for the docs website — hierarchy/spacing, contrast/themes, WCAG 2.2 accessibility, navigation/search UX, code presentation, motion, typography. Load BEFORE any website UI/UX or component styling work; rules cited to Web Interface Guidelines (Rauno/Vercel), Refactoring UI, WCAG 2.2, thoughtbot, and Stripe-docs teardowns.
---

# Design/UI standards (developer-docs website)

Load before styling or reviewing UI under `website/**`. Full cited rules:
[references/design-rules.md](references/design-rules.md). Cite rule numbers in
findings/dispositions.

## Repo overlay

- **Data-heavy pages are the norm here** (548-connector catalog): design every density state
  (#5), URL-reflected filter state (#23), scroll restoration on back-nav (#24), truncation rules
  for long connector/field names (#31), tabular numerals for counts (#28).
- **Both themes always** (dark/light): semantic tokens (ts-website #24), `color-scheme` +
  `theme-color` on `<html>` (#10), no transition storm on theme toggle (#11), layered shadows
  that survive dark mode (#13).
- **Non-negotiable floors**: WCAG AA contrast 4.5:1 text / 3:1 UI (#7), keyboard operability with
  visible focus (#14–15), `prefers-reduced-motion` honored (#16), real `<a>`/`<Link>` for all
  navigation (#27), hit targets ≥24px (#20).
- **Status is never color-alone** (#12): connector health/availability badges carry icon + text.
- **Prose columns cap at ~66ch** (#35); interaction animation caps at ~200ms, transform/opacity
  only (#32).

## Verification before handoff

Check both themes; keyboard-walk new interactive elements; run the catalog page at sparse/dense
extremes; confirm filters round-trip through the URL; `pnpm test:e2e` when nav/search touched.
