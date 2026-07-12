# UI SPEC: Reader Notes And Header Clarity

## Reference Inputs

- User screenshots from 2026-07-12 showing flat replies, a cursor-escaping hover card, clipped header controls, and difficult-to-scan marginalia.
- Existing Polymetrics square emerald visual language.
- Reddit-style parent/child thread rails, adapted to an editorial reading surface.

## Interaction Contract

- Hovering or focusing highlighted text shows a lightweight preview.
- Clicking or pressing Enter/Space pins the complete note card; moving the pointer does not dismiss it.
- The pinned card exposes explicit `Open thread` and `Close` controls and closes with Escape.
- Replies render beneath their direct parent. A reply to a reply is never flattened into the root list.
- Every visible message includes author, time, body, and available actions in the same order.
- At desktop widths, Product, Docs, Blog, Changelog, GitHub, search, Get Started, and Get Demo must not overlap or clip.

## Visual Contract

- Root notes use a quiet framed surface; replies use a left rail and compact vertical rhythm.
- Message text remains the highest-contrast content inside a thread.
- Metadata is secondary but legible; actions use icons plus specific labels.
- Thread depth uses 12px indentation up to depth 4, then continues with rails without further horizontal compression.
- Pinned note cards stay within a 16px viewport gutter and choose above/below placement based on available space.

## Accessibility

- Highlight controls remain keyboard operable with visible focus.
- Pinned notes use dialog semantics and receive an accessible name.
- Icon-only close controls have labels.
- Reply textareas have associated accessible labels and character guidance.
- Dynamic post/delete status remains announced through the provider live region.
