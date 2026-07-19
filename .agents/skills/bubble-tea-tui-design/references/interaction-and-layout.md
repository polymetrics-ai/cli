# Interaction and layout contract

## Modes

Use explicit, shallow modes rather than turning the CLI into a full Vim clone.

| Mode | Enter | Keys | Exit |
|---|---|---|---|
| Normal | surface start, completed edit, `esc` from a child mode | navigation, pane focus, action menu, help | `/`, `i`, `e`, `enter`, or `q` |
| Filter | `/` or focus a filter field | printable input edits; arrows move cursor/history | `enter` applies, `esc` restores Normal |
| Edit | `i`, `e`, or a wizard field | printable input edits; standard text editing keys | `enter` accepts, `esc` cancels field edit |
| Confirm | selected mutation after preview | full action labels; typed challenge when destructive | approve, `esc` cancels |
| Help | `?` | scroll/search current-mode bindings | `?`, `esc`, or `q` closes help only |

Rules:

- Show mode and focus in text, for example `NORMAL · results` or `FILTER · connectors`.
- `esc` unwinds one layer: edit/filter/help → Normal → prior surface. It never jumps
  directly out of a multi-step workflow.
- `q` quits only from Normal mode. In an input it types `q`.
- `ctrl+c` requests cancellation everywhere; it does not skip cleanup/final-state rendering.
- Use `h/l` for sibling tabs or spatial panes, not for actions. Use `tab/shift+tab` as the
  universal focus fallback.
- `enter` activates the focused item. Avoid overloaded single-letter mutations.

## Key bindings

Define bindings with `bubbles/key` and expose them through `bubbles/help`. Disable bindings
that are not valid in the current mode so help and behavior cannot diverge.

| Intent | Primary | Alternative |
|---|---|---|
| next/previous row | `j` / `k` | down / up |
| first/last row | `gg` / `G` | home / end |
| half-page | `ctrl+d` / `ctrl+u` | page down / page up |
| next/previous pane | `tab` / `shift+tab` | `h` / `l` when spatial |
| filter/search | `/` | `ctrl+f` where terminal-safe |
| next/previous match | `n` / `N` | — |
| activate | `enter` | space only for toggles |
| contextual actions | `a` | labelled action menu |
| help | `?` | — |
| back/cancel layer | `esc` | — |
| quit browser | `q` in Normal | — |
| force cancellation request | `ctrl+c` | — |

## Layout classes

- **Wide (>=120 columns):** navigation/list left, primary content center, contextual detail
  right when useful. Do not fill all width with boxes.
- **Standard (80–119):** two panes at most; detail stacks below or opens as a focused view.
- **Compact (>=60 and below 80):** one pane at a time with a breadcrumb and explicit pane
  switching. Tables hide optional columns before truncating identifiers.
- **Below 60×18:** render the measured size, recommended size, and a plain-command fallback.

The feature-specific design may keep the historic 80×24 enhanced minimum, but it must
still provide a readable compact or size-guard state rather than a damaged frame.

## Visual hierarchy

1. A quiet top line names the surface and truthful status.
2. One visually dominant work area holds the current task.
3. Focus uses one accent border/title/selection treatment, never several competing ones.
4. A contextual detail pane explains the selected object; it is not another navigation
   source unless explicitly focused.
5. A persistent footer shows mode, short bindings, progress/error, and plain fallback.

Use semantic tokens, terminal-default foreground/background, restrained borders, and
spacing. CAVA-like gradients and rapid motion are inspiration only for bounded progress,
never decoration. Prefer the information density of bpytop with the navigational clarity
of LazyGit.

## Bubble Tea v2 mechanics

- Represent terminal modes in `tea.View` fields (alt screen, mouse, cursor, title) instead
  of imperative scattered setup.
- Treat `tea.KeyPressMsg` and window/background/color-profile messages as state changes.
- Return `tea.Cmd` for I/O, timers, event subscriptions, clipboard, and cancellation.
- Batch independent commands deliberately; do not spawn unmanaged goroutines from models.
- Pass messages only to the focused child unless a global message (resize, cancel, theme,
  event) must reach more than one child.
- Keep final inline dashboard frames in scrollback. Use alt screen only for place-like
  browsers, grids, and pagers.
