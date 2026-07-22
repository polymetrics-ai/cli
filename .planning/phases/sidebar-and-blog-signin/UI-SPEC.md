# UI SPEC: Sidebar And Blog Sign-In

## References

- Production `https://cli.polymetrics.ai/`: sticky 256px sidebar inside the page flex row.
- Stripe/Vercel docs: fixed desktop navigation columns that reserve layout width.
- Linear: dense, quiet, work-focused navigation chrome.

## Visual Direction

Keep the existing boxy emerald language. The blog-only auth card belongs in the left sidebar footer, mirroring the `pm CLI` footer card with a mono eyebrow, compact pitch, square CTA, and account actions when signed in.

## Layout Rules

- Sidebar visibility uses explicit global CSS media classes.
- Sidebar must not intersect the main content bounding box when visible.
- Blog article layout must fit left sidebar, reading column, and marginalia rail without cramping.
- The auth card reserves height while session state loads to prevent layout shift.

## Accessibility Rules

- Account actions use semantic buttons or links.
- Icon-only controls keep `aria-label`.
- Hydration-ready markers are data attributes, not hidden unlabeled controls.
