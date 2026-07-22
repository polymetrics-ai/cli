# PLAN: Reader Notes And Header Clarity

## Delivery Path

- GSD command attempted: `scripts/gsd prompt programming-loop init --phase comments-thread-ux --dry-run`
- Adapter result: `programming-loop` is not registered by the local shell adapter.
- Fallback: manual GSD programming loop with test-first evidence recorded in this phase.
- Skills: `frontend-design`, `web-design-guidelines`, `vercel-react-best-practices`, `vercel-composition-patterns`, `e2e-testing-patterns`.

## Design Direction

Build a quiet editorial conversation layer inspired by durable annotation tools and Reddit's readable thread structure. Notes remain visually native to the existing square emerald system: compact author metadata, full message bodies, persistent click-open cards, and vertical thread rails that make parentage unmistakable without turning the article into a social feed.

## Tasks

1. Add failing browser coverage for persistent note previews, true nested replies, and non-overlapping header controls.
2. Replace the hover-only note preview with a hover preview plus click/keyboard-pinned note card.
3. Render replies recursively from direct parent relationships and preserve reply/delete/composer behavior at every depth.
4. Clarify the notes-sheet hierarchy, labels, controls, and empty/loading states.
5. Remove the navbar's phantom symmetric grid column and make search/navigation breakpoints collision-free.
6. Verify typecheck, unit tests, focused Playwright flows, desktop/mobile screenshots, keyboard access, and overflow.

## Constraints

- No new dependencies.
- Preserve existing API and database contracts.
- Keep deep threads readable inside a narrow sheet by capping visual indentation while retaining semantic nesting.
- Preserve reduced-motion behavior and visible keyboard focus.
