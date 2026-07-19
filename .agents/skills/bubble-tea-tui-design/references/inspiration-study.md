# Reference application study

Use the references as a vocabulary. Do not clone their branding or copy unsafe actions.

| Reference | Adopt | Adapt for Polymetrics | Avoid |
|---|---|---|---|
| [bpytop](https://github.com/aristocratos/bpytop) | dense telemetry, visible scale, selected-process detail, 256-color fallback | one dominant pipeline/query chart with exact values | filling every cell, braille as the only carrier |
| [Conky](https://github.com/brndnmtthws/conky) | configurable metric modules and compact scalar/graph pairing | user-selectable dashboard sections later | desktop-overlay assumptions and unrestricted scripting/config execution |
| [CAVA](https://github.com/karlstav/cava) | legible motion, gradient restraint, dumb-terminal fallback | bounded progress/rate animation with reduced-motion path | decorative animation or waveform-shaped analytical claims |
| [LazyGit](https://github.com/jesseduffield/lazygit) | focused panels, persistent keys, `?` help, Vim navigation, context actions | operator workspace for runs/browse/query | generic shell command and unconfirmed mutation shortcuts |
| [LazyDocker](https://github.com/jesseduffield/lazydocker) | resource list + logs + ASCII metrics + actions in one workspace | run/certify/RLM dashboard drill-down | unlabeled destructive single-key actions |
| [fzf](https://github.com/junegunn/fzf) | query/list/preview composition, instant feedback, keyboard ergonomics | safe internal fuzzy browser and preview | shell-backed preview execution |
| [Gum](https://github.com/charmbracelet/gum) | one focused interaction at a time; excellent shell/plain composability | wizard field and confirmation cadence | treating a multi-pane dashboard as a sequence of disconnected prompts |
| [awesome-terminal-aesthetics](https://github.com/kud/awesome-terminal-aesthetics) | catalog for comparative evaluation | periodically re-check relevant Go TUI patterns | selecting a dependency on screenshots alone |
| [NTCharts](https://github.com/NimbleMarkets/ntcharts) | native Bubble Tea v2 chart models and examples | dependency-gated renderer behind a local interface | assuming its evolving API is stable without a pin/wrapper |

The primary structural reference is **LazyGit's operator workspace**, combined with
**fzf's filter/list/preview interaction**, **bpytop's exact telemetry/chart density**, and
**Gum's focused wizard cadence**. CAVA supplies motion restraint; Conky supplies optional
metric composition. Polymetrics keeps its own quiet palette, pipeline rail, safety gates,
and plain/JSON contract.

Primary implementation references:

- [Bubble Tea v2](https://github.com/charmbracelet/bubbletea) and its
  [v2 upgrade guide](https://github.com/charmbracelet/bubbletea/blob/main/UPGRADE_GUIDE_V2.md)
- [Bubbles](https://github.com/charmbracelet/bubbles)
- [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- [GitHub CLI accessibility guide](https://accessibility.github.com/documentation/guide/cli/)
- [Building a more accessible GitHub CLI](https://github.blog/engineering/user-experience/building-a-more-accessible-github-cli/)
