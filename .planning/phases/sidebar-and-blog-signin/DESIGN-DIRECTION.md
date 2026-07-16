# DESIGN DIRECTION: Sidebar And Blog Sign-In

## References

- `https://cli.polymetrics.ai/` production markup: sticky 256px sidebar inside the page flex row, not overlayed; right TOC appears at wider breakpoints.
- Stripe/Vercel documentation: persistent left navigation and article columns reserve space rather than floating over content.
- Linear application chrome: dense, quiet utility navigation with compact action blocks.

## Direction

Keep the existing “boxy emerald” identity and avoid introducing a new visual language. The distinctive blog sign-in placement should feel like a sibling of the `pm CLI` footer card: a bordered, square, mono-labeled block with one clear call to join the discussion.

## Constraints

- No rounded corners.
- Token classes only for new colors.
- Geist Sans body, Geist Mono eyebrows, Chakra Petch labels/CTAs, Instrument Serif only for display headings.
- Motion must respect reduced-motion.
