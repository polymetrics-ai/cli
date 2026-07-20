# Testing and accessibility

## Test pyramid

1. **Pure state tests:** mode transitions, key conflicts, focus order, resize class,
   selection, filtering, pagination, cancellation, and error state.
2. **Component tests:** Bubbles models receive only expected messages; disabled key
   bindings stay inert and absent from short help.
3. **Headless frame tests:** stable frames at 160×45, 100×30, 80×24, compact, and below
   minimum for success/failure/empty/loading/cancelled states.
4. **Command integration:** TTY projection uses the same service and exit status as plain;
   Bubble Tea/Huh activate only on stdin+stdout TTY activation and no bypass flag. Add RED tests
   for `stdin-piped+stdout-TTY` fallback, `stdout-piped`, `CI=1`, `PM_NO_TUI=1`, `TERM=dumb`,
   `--json`, `--plain`, and `--no-input`; each must bypass the TUI/prompt path without consuming
   scripted stdin, without hanging, and without using `/dev/tty`. Sequential prompting is allowed
   only in explicit accessible mode after the same gate passes.
5. **Manual terminal matrix:** dark/light, 24-bit/256/16/no color, ASCII, resize, paste,
   narrow Unicode, mouse optional, screen reader/reduced-motion transcript.

Avoid brittle full-frame goldens for timestamps, random IDs, spinners, or live counters.
Inject a clock/ID source and assert semantic regions when appropriate.

## Required scenarios

- printable `q`, `j`, `/`, and `?` edit text in Filter/Edit mode;
- `esc` exits the child mode before navigation/back;
- `q` quits only in Normal mode; `ctrl+c` cancels and produces the truthful final state;
- every focusable control is reachable without a mouse;
- resize never panics or produces negative widths;
- empty, loading, partial, failure, and stale states are distinguishable in words;
- dynamic control characters and secret-like values are sanitized/redacted before View;
- query/chart data is bounded and reports downsampling/missing values;
- JSON remains one deterministic envelope and NDJSON progress remains on stderr;
- final inline dashboard output remains useful in scrollback.

## Accessibility checklist

- Provide a no-redraw sequential/plain experience for screen readers.
- Announce `Step N of M`, mode, focus, status, result count, and selection in text.
- Pair every color with a word/glyph and validate 4-bit and `NO_COLOR` output.
- Avoid braille-only state indicators. If braille charts are enabled, retain axes, exact
  selected values, and the text summary/table fallback.
- Respect reduced motion: no spinner/animated chart; bounded periodic status lines.
- Use ASCII fallbacks for borders, rail, selection, and charts.
- Do not make mouse, OSC52, Kitty graphics, or progressive keyboard protocols necessary.
- Keep error recovery beside the error: what happened, safe next step, and plain command.

## Review evidence

Record exact commands and results in `TDD-LEDGER.md`/`VERIFICATION.md`, including focused
RED, focused GREEN, repeated/race tests as applicable, full repository gates, help/manual/
website parity, and the manual terminal matrix. Screenshots are supporting evidence, not a
substitute for behavioral tests or accessible transcripts.
