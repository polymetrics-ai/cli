# CLI Architecture v2 delivery and terminal research evidence

Accessed: 2026-07-22 UTC / 2026-07-23 local time. This artifact is dated because issue state,
branch heads, review coverage, and library repositories change. Re-query current sources before
scheduling or integration.

## Scope and method

Research used read-only `gh-axi` issue/PR/ref inspection, `chrome-devtools-axi` browser inspection of
official documentation and pinned repository files, parent/default branch code comparison, and the
independent #397 audit. No credential, runtime lifecycle, remote mutation, production TUI change, or
dependency addition occurred.

Pinned repository state at audit completion:

- `origin/main`: `873cd7b251f70c4a35a607a0d4e86051ea0fbd15`;
- `origin/feat/cli-architecture-v2`: `21d195aff0c7bd60b3bf54f14b1ce165cec9e03f`;
- `origin/feat/408-flow-etl-dashboards`: `6c643f5c971d1fac4a83e4ffe653b83847c2fceb`;
- parent PR #438: open, draft, implementation incomplete, and human-gated.

These identifiers are evidence for this date, not evergreen skill rules.

## Program gap finding

The parent branch already has the CLI Architecture v2 foundations: Cobra strangler routing, isolated
typed Viper configuration, bounded typed events and NDJSON, redacted slog, default-off OpenTelemetry
traces/metrics, native namespace migrations, TTY selection foundations, transcript tests, ADRs, and
the #462 design system. The detailed `bubble-tea-tui-design` skill already covers Bubble Tea models,
layout, forms, charts, accessibility, dual-TTY and bypass behavior, security, and deterministic
verification.

The remaining gap is delivery coordination, not generic TUI knowledge:

- open issues do not distinguish parent-integrated implementation from default-branch delivery;
- nested hierarchy, native/planning dependencies, and shared command/help/docs collisions differ;
- review and verification bind to exact stacked heads and become stale on drift;
- the local GSD registry lacks `programming-loop`, requiring a recorded manual universal-loop
  fallback;
- integrated review debt must not cause duplicate implementation or fabricated historical PRs;
- final parent readiness requires Shepherd, parent reruns, and a human-only merge.

Therefore the approved change adds `cli-architecture-v2-delivery`, keeps
`bubble-tea-tui-design` authoritative, and routes each phase to existing Go/domain skills.

## Dated issue-to-code map

| Surface | Parent truth at access time | Delivery gap |
|---|---|---|
| S0/P01-P06/P08 | implemented and credential-free verified | retain frozen; carry any exact-range review debt |
| P07 | stdout-only detector on parent | dual stdin+stdout correction exists only on #408 branch |
| P09 / #421-#437 | implementation integrated | parts of historical process/review evidence are incomplete; do not reimplement |
| P10/#408 | dashboard and dual-TTY work exists on its worker branch | synchronize parent, independent VERIFY/review/Shepherd, stacked PR, parent promotion |
| P12/P17 | default-off redacted OpenTelemetry traces and metrics integrated | preserve; no reopening |
| D-TUI/#462 | design docs and Bubble Tea skill integrated | exact-range review/process disposition remains a parent gate |
| P11, P13-P16, P18/P18B, P19-P20, P22, #463 | not implemented on parent | execute dependency/collision waves from current remote truth |
| P21/#419 | explicitly deferred by human | represent as `deferred_by_human`, not implemented or complete |

## TUI framework comparison

Framework choice remains use-case-specific. The comparison confirms rather than reopens ADR-0003
and #462.

| Project | Architecture and strengths | Fit for #397 |
|---|---|---|
| Bubble Tea v2 | Declarative `Model`/`Init`/`Update`/`View`; commands isolate effects; inline or full-window programs | Selected. Best match for deterministic state, event projection, cancellation, and testable UI/plain parity. |
| Bubbles v2 | List, table, viewport, help/key, text, progress, spinner, and file-picker components | Selected component layer; still requires explicit root ownership and safety tests. |
| Lip Gloss v2 | Declarative layout/style; ANSI/256/truecolor/ASCII profiles and automatic downsampling | Selected renderer; color remains reinforcement only. |
| Huh v2 | Grouped typed forms, validation, dynamic fields, first-class accessible standard-prompt mode | Selected for eligible dual-TTY wizards with complete flag alternatives. |
| Termenv | Capability/profile/theme detection and ANSI styling, not an application architecture | Useful lower-level evidence; already beneath the selected ecosystem, not a replacement. |
| tcell | Low-level cell screen, event loop, terminal portability, color/mouse/encoding control | Appropriate for custom engines; too much application/layout ownership for this program's selected model. |
| tview | Imperative widgets and layouts over tcell: forms, table, tree, list, grid/flex, modal | Strong batteries-included toolkit, but a different event/state architecture and dependency direction. |
| gocui | Layout callback, overlapping views, global/view key bindings, mouse, main loop | Proven by LazyGit; lower-level view/controller complexity offers no advantage over the approved stack. |
| termui | Dashboard widgets, grid/absolute layout, plots/tables, manual event/render loop | Useful for dashboards; weaker fit for the broader forms/browser/accessibility and machine-contract program. |
| go-prompt | Focused REPL/prompt completion, history, and Emacs-style editing | Good for interactive shells, not multi-pane dashboards or the complete #397 surface. |

No alternative supplied evidence that justified replacing the already integrated Bubble Tea v2
design contract. No chart library is approved by this comparison; #463 retains its explicit human
dependency gate and internal-renderer fallback.

## Production application lessons

| Application | Evidence used | Adopt / avoid |
|---|---|---|
| LazyGit | gocui event loop with contexts, controllers, views, key maps, async tasks, and end-to-end tests | Adopt explicit focus/actions and layered UI boundaries; avoid generic command execution and God-struct drift. |
| K9s | Dense resource browser, filter/navigation/help, readonly mode, context actions | Adopt visible modes and readonly safety; avoid unsafe single-key remote mutations. |
| gh-dash | Bubble Tea sections, configurable Vim-style keys, preview/detail workflow | Adopt composable section models and keyboard throughput; keep Polymetrics' stricter write and machine contracts. |
| aerc | tcell-based tabs, context commands, key bindings, viewer/composer modes | Adopt mode discoverability; avoid embedded arbitrary command/shell facilities. |
| Glow | TUI browser plus deterministic CLI/pager path | Strong evidence for interactive/plain sibling surfaces. |
| Gum | Focused choose/input/write/filter/confirm/table/pager commands with shell composition | Adopt one-decision wizard cadence and exit semantics, not an entire shell-driven architecture. |
| fzf | Query/list/preview keyboard loop and explicit no-color support | Adopt safe internal filter/list/preview; never adopt shell-backed previews. |
| LazyDocker | Resource list, logs, metrics, contextual actions | Adopt operational dashboard hierarchy; label and gate every mutation. |

The #462 design study additionally covered bpytop, Conky, CAVA, awesome-terminal-aesthetics, and
NTCharts. Its existing reference files remain the implementation authority.

## Source ledger

All repository pins below were inspected through direct official URLs. Short pins identify the
reviewed snapshot; use the full object from the repository before dependency or API decisions.

| Source | Pin / URL | Evidence scope |
|---|---|---|
| Bubble Tea | `fc707bb7`, https://github.com/charmbracelet/bubbletea | model/update/view, commands, inline/full-window rendering |
| Bubbles | `b52e21a6`, https://github.com/charmbracelet/bubbles | component inventory and key/help patterns |
| Lip Gloss | `5696b280`, https://github.com/charmbracelet/lipgloss | layout and profile downsampling |
| Huh | `e0035498`, https://github.com/charmbracelet/huh | grouped forms, validation, accessible mode |
| Termenv | `368a3572`, https://github.com/muesli/termenv | terminal/color profiles, `NO_COLOR`, theme detection |
| tcell | `45d70ee4`, https://github.com/gdamore/tcell | cell/event abstraction, portability and capabilities |
| tview | `63ee97f9`, https://github.com/rivo/tview | widget/layout/application inventory |
| gocui | `0e75b37a`, https://github.com/jroimartin/gocui | view/layout/keybinding/main-loop model |
| termui | `3ee54a07`, https://github.com/gizak/termui | dashboard widgets and event/render loop |
| go-prompt | `82a91227`, https://github.com/c-bata/go-prompt | prompt completion/history scope |
| LazyGit | `d8b07ee4`, https://github.com/jesseduffield/lazygit | production layering and end-to-end test boundaries |
| K9s | `5fedc440`, https://github.com/derailed/k9s | resource browser, help, filters, readonly mode |
| gh-dash | `118eef11`, https://github.com/dlvhdr/gh-dash | Bubble Tea dashboard composition |
| aerc | `1e048622`, https://github.com/rjarry/aerc | custom terminal UI modes and command surface |
| Glow | `53788271`, https://github.com/charmbracelet/glow | TUI/CLI/pager sibling paths |
| Gum | `716d8b5d`, https://github.com/charmbracelet/gum | focused interactive command composition |
| fzf | `235a726f`, https://github.com/junegunn/fzf | filter/list/preview and no-color behavior |
| LazyDocker | `7e7aadc2`, https://github.com/jesseduffield/lazydocker | operational dashboard composition |
| Cobra flags/completion | https://cobra.dev/docs/how-to-guides/working-with-flags/ and https://cobra.dev/docs/how-to-guides/shell-completion/ | local/persistent flags, validation, dynamic completion |
| Viper | https://github.com/spf13/viper | precedence, isolated instances, environment/flag binding |
| OpenTelemetry Go | https://opentelemetry.io/docs/languages/go/ | traces/metrics stable; logs beta at access time |
| GitHub subissues | https://docs.github.com/en/issues/tracking-your-work-with-issues/using-issues/browsing-sub-issues | hierarchy and machine-readable parent/subissue fields |
| GitHub CLI accessibility | https://accessibility.github.com/documentation/guide/cli/ | non-color semantics, prompts, assistive technology expectations |
| `NO_COLOR` | https://no-color.org/ | user-controlled color suppression convention |

## Not implemented by this task

This skill slice adds no runtime dependency, CLI command, TUI model, dashboard, wizard, chart,
completion, telemetry exporter, generated help/manual/website artifact, credential behavior, or
connector behavior. It does not arbitrate the live queue, promote #408, change #419's human decision,
mark parent PR #438 ready, or merge any branch.
