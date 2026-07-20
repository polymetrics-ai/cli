# Terminal UI research and design system

Status: DESIGN GATE — issue #462, supporting the interactive phases of CLI Architecture v2.
Research and local interaction checks were performed 2026-07-19. The normative surface
wireframes remain in `docs/design/tui-ux-design.md`; this document freezes the cross-surface
interaction, layout, chart, and reference-application decisions that Pi/GSD workers must use.

## Decision in one sentence

Build a quiet **operator workspace** based structurally on LazyGit, combine it with fzf's
filter/list/preview loop, bpytop's exact telemetry density, and Gum's focused wizard cadence,
then express it in Polymetrics' pipeline-rail language through Bubble Tea v2. Do not make a
Vim clone: provide predictable Vim navigation inside explicit Normal, Filter, and Edit modes.

## Research method

The study used primary project repositories/documentation plus an isolated local terminal lab.
The requested applications were installed or built without modifying the Polymetrics module:

| Application | Evaluated version/path | Interaction exercised |
|---|---|---|
| bpytop | Homebrew 1.0.68 | overview, process focus, movement, help |
| Conky | 1.24.3-pre, source build with ncurses and GUI/network features disabled | text metrics, bars, process rows |
| CAVA | Homebrew 1.0.0, isolated FIFO synthetic-audio config | live spectrum, resize-safe terminal rendering |
| LazyGit | Homebrew 0.63.1, repository view | panel focus, `j/k`, `tab`, `?`, close help |
| LazyDocker | Homebrew 0.25.2, local Docker view | panel focus, logs, metrics, `j/k`, `?` |
| fzf | Homebrew 0.74.1, repository file list/preview | query, selection, preview scroll |
| Gum | Homebrew 0.17.0 | choose list, arrow movement, confirmation cadence |
| awesome-terminal-aesthetics | official repository | browsable catalog in a terminal pager |
| NTCharts | `v2` branch, isolated `examples/quickstart` | Bubble Tea v2 time-series rendering and axes |

The lab used isolated configuration and read-only data where possible. CAVA consumed a
synthetic local FIFO instead of a microphone. No credential, secret, remote mutation,
generic shell action, or repository write was exercised. ANSI captures and deterministic
1500×900 PNG renderings are stored outside the repository under
`~/.local/share/polymetrics-tui-reference/`; they are research evidence, not vendored assets.

Reproduction on macOS:

```bash
brew install bpytop cava lazygit lazydocker fzf gum tmux ffmpeg
git clone --depth 1 https://github.com/brndnmtthws/conky.git
git clone --depth 1 --branch v2 https://github.com/NimbleMarkets/ntcharts.git
```

Conky has no Homebrew formula in the evaluated environment, so use a local prefix and keep
X11/Wayland/network features disabled for a terminal-only study. Do not run reference-app
shell hooks or destructive Docker/Git actions merely to capture visuals.

## What each reference teaches us

### LazyGit: the primary structural reference

LazyGit keeps the selected object, its detail, contextual actions, current focus, and short
help visible simultaneously. Vim navigation is discoverable rather than required; arrows
still work. Its panel model maps directly to Polymetrics:

- runs/flows/connectors on the left;
- the primary rail, query result, log, or manual in the work pane;
- selected-object metadata or safe actions in a contextual detail pane;
- mode/status and short keys in a persistent footer.

Adopt the hierarchy and navigation. Do not adopt its generic command shell. Polymetrics must
never expose arbitrary shell, arbitrary HTTP write, or generic SQL write through a TUI.

### fzf: the browser interaction reference

fzf's query → filtered list → preview loop has low cognitive cost and excellent keyboard
throughput. It is the reference for connector browsing, manual selection, query table
selection, and searchable docs. Polymetrics previews must be internal, sanitized renderers;
fzf's shell-executed preview mechanism is explicitly out of scope.

### bpytop: the telemetry and chart-density reference

bpytop demonstrates that a terminal can carry graphs, exact metrics, tables, and selected
detail together when the focal region and scales remain visible. Adopt:

- sparklines/graphs paired with exact numbers;
- units, min/current/max, and selected-row detail;
- restrained semantic color and terminal-profile fallbacks;
- mouse as an optional accelerator only.

Avoid turning every metric into a graph, relying on braille/color alone, or giving all panes
equal emphasis.

### LazyDocker: the operational dashboard reference

Its resource list + logs + ASCII metrics + contextual actions is a useful shape for flow,
ETL, certify, and RLM dashboards. Polymetrics changes the action model: writes and dangerous
operations live behind an action menu, preview, explicit labels, and the existing approval
gate. A single unlabeled destructive key is not permitted.

### Gum: the wizard reference

Gum's beauty comes from focus: one decision, strong prompt, immediate validation, minimal
chrome. Use that cadence for each Huh form group and confirmation, while retaining the
growing pipeline preview. A multi-pane operational dashboard should not be decomposed into
an endless chain of prompts.

### CAVA and Conky: visual rhythm and composition references

CAVA proves smooth terminal motion can remain legible with constrained color and simple
geometry. Use this only for truthful bounded activity/rate feedback, and stop it in reduced-
motion mode. A decorative waveform is not a data chart.

Conky shows the value of user-selectable metric modules and compact scalar + small-graph
pairings. That is a later dashboard-personalization idea, not permission to support arbitrary
scripts or a desktop-overlay layout in `pm`.

### awesome-terminal-aesthetics: a catalog, not a specification

Use the catalog to discover and compare patterns. A screenshot is insufficient evidence for
a dependency or interaction choice; implementation decisions still require official API
documentation, an isolated spike, accessibility review, and the human dependency gate.

## Normative interaction model

### Modes

Every place-like surface begins in **NORMAL** mode. The footer displays both mode and focus,
for example `NORMAL · results`.

| Mode | Meaning | Printable keys | `esc` | `q` |
|---|---|---|---|---|
| Normal | navigate/select/action | commands | back one surface | quit the browser |
| Filter | edit a search/filter | insert text | leave filter, retain or restore documented query | insert `q` |
| Edit | edit SQL/name/path/form field | insert text | cancel field edit | insert `q` |
| Confirm | review an explicitly named mutation | only documented challenge/input | cancel confirmation | input or no-op, never approve |
| Help | complete keys for current mode | search when provided | close help | close help |

Vim-style navigation is supported where its spatial meaning is clear:

- `j/k` and down/up select rows;
- `gg/G` and home/end select first/last;
- `ctrl+u/ctrl+d` and page-up/page-down move by half/full page;
- `/` enters Filter, `n/N` visits matches;
- `tab/shift+tab` always changes focus; `h/l` may move between spatial panes/tabs;
- `enter` activates, space toggles a checkbox only;
- `?` shows complete current-mode help;
- `esc` unwinds one layer;
- `q` quits only in Normal;
- `ctrl+c` requests cancellation and still permits cleanup/truthful final state.

Do not implement Vim command-line mode, macros, registers, operator-pending grammar, or
undocumented key chords. Familiarity is the goal; emulation is not.

### Focus and action design

- One accent border/title/selection identifies focus. Never make the user infer focus from
  color alone.
- Ordinary bare namespace commands render contextual help and exit 0. On an eligible dual-TTY,
  bare `pm query` and bare `pm reverse` are the deliberate human-first exceptions and enter the
  same workspace as `pm query grid` and `pm reverse guide`; those explicit aliases remain supported
  for documentation, scripts, and direct navigation. Help flags and every bypass/non-TTY path
  render deterministic contextual help, never a TUI. Invalid actions remain usage errors.
- The footer is contextual: disabled bindings are absent, and `?` matches actual behavior.
- Printable keys belong to focused inputs. Global navigation must not steal `j`, `q`, `/`,
  or `?` while Filter/Edit owns focus.
- Actions open a labelled menu or button row. A mutation states target, record count, and
  reversibility before approval.
- The existing reverse-ETL plan → preview → approval → execute sequence is preserved exactly.
  Approval tokens are sensitive one-time authorization values: a guided flow may carry them only
  ephemerally in memory through that seam and must never render them in final frames, transcripts,
  logs, screenshots, accessibility output, JSON, shell-equivalent command text, or fixtures.

## Layout and visual system

### Responsive classes

| Class | Width | Shape |
|---|---:|---|
| Wide | 120+ | list/navigation + primary work + contextual detail when justified |
| Standard | 80–119 | at most two panes; detail stacks or opens on demand |
| Compact | 60–79 | one pane at a time, breadcrumb + explicit pane switching |
| Guard | below 60×18 | measured-size message + recommended size + plain command |

Some feature designs retain an 80×24 enhanced-layout minimum. That is compatible with this
system: the 60–79 class may show a focused compact view or the size guard, but never a broken
three-pane frame. Resize is derived from `tea.WindowSizeMsg`, not constants.

### Hierarchy

1. Quiet title/status line.
2. One dominant work area: pipeline rail, result table/chart, log, or document.
3. Optional context/detail area.
4. Persistent short-help/status footer.

Use the terminal default background and semantic tokens. Ask Bubble Tea for background/
color-profile information and let Lip Gloss downsample. Reserve gradients for bounded,
nonessential accents; keep the pipeline rail as the signature visual. Whitespace and a
single focus treatment are more useful than a box around every region.

### Motion

- Animate only a real state change, live rate, or bounded progress.
- Throttle rendering separately from lifecycle-event delivery.
- Avoid flashing and rapid color cycling.
- Reduced-motion/accessibility mode replaces spinners/animated charts with periodic text.
- Inline dashboards freeze a truthful final frame in scrollback.

## Query charts and terminal dashboards

### Chart grammar

| Data/question | Default | Required companion |
|---|---|---|
| ordered timestamp + numeric value | line/time series | start/end/current/min/max + unit |
| tiny trend inside a row/card | sparkline | current + delta |
| category + numeric value | sorted horizontal bars | exact value and unit on every bar |
| one numeric distribution | histogram | bucket labels/counts + sample count |
| two numeric fields | scatter | x/y names, units, ranges, sample count |
| matrix/grid | heatmap | labelled scale + selected cell exact value |
| exact records or many text columns | table | headers, page, total/limit |

No pie charts, 3D charts, unlabeled waveforms, dual axes, or color-only heatmaps. When axes,
units, and values do not fit, switch to a table/text summary rather than rendering a
misleading miniature.

### Query interaction

Phase #411 delivers the read-only query workspace, entered by bare `pm query` on an eligible
dual-TTY or by the explicit `pm query grid` alias. On bypass/non-TTY paths, bare `pm query` renders
contextual help and exits 0. Dedicated child issue #463 may then add:

- `v` toggles Table/Chart without re-running a write or changing the underlying rows;
- a labelled chart setup view chooses chart type, X, Y, aggregation, unit, and sort from
  validated result metadata;
- selection/crosshair reports exact values in text;
- the table remains available as the accessible and exact-data representation;
- export serializes the underlying rows, never the glyph rendering, and is a typed read-only
  export path rather than a generic file writer;
- export defaults to a project-scoped directory such as `.polymetrics/query-exports/`, resolves
  and cleans the requested path, confines it to the project, rejects control characters,
  traversal, absolute or broad paths, symlink targets/final-component races, and overwrites by
  default, then requires confirmation only when stdin and stdout are TTYs or noninteractive
  `--output <project-relative-path>` plus `--force`;
- `--no-input` without a preapproved export path fails with
  `query grid export requires explicit output — pass --output <project-relative-path> and --force, or run without --no-input to confirm interactively`;
- scripted command echoes are sanitized and project-relative;
- `--plain`, non-TTY stdin, or non-TTY stdout prints deterministic table + numeric summary when
  required flags are present, or the exact required-flag error asserted by the implementation
  issue. `--json` emits only documented machine data/schema. `--no-input` never prompts; any JSON
  chart-spec requires a separately documented stable schema.

Rendering is bounded. Cap points, apply deterministic bucketing/downsampling, and disclose it
(`2,000 rows · 120 plotted · min/max buckets`). Report ignored missing/non-numeric values.
Charts use only the existing `QuerySQL`/`validateSelectOnly` path and cannot synthesize or
execute SQL.

### Dashboard composition

The first frame answers: what is selected/running, is it healthy/safe, what changed, and what
can I do next? Use one primary chart or pipeline rail plus a few scalar facts. Secondary
charts belong behind tabs/drill-down; a bright tile wall obscures priority.

Useful initial compositions:

- flow/ETL: pipeline rail + rows/rate/elapsed + selected-step log/detail;
- query: result table or one selected chart + table list/metadata + exact selected value;
- certify: status table + pass/fail/partial bars + selected finding;
- RLM: state/heartbeat/attempt + bounded time-series only when the source is truthful;
- connector browser: filtered list + manual/capability preview, no chart by default.

## Bubble Tea and chart dependency direction

Use the Charm v2 line already selected by ADR-0003:

- Bubble Tea v2 for declarative model/update/view and `tea.View` terminal options;
- Bubbles v2 for key/help/list/table/viewport/text components;
- Lip Gloss v2 for semantic styling, layout, and color-profile degradation;
- Huh v2 for accessible wizard groups;
- Glamour v2 for manual/document rendering.

NTCharts' `v2` branch is the best current chart candidate because it targets Bubble Tea v2
and provides time-series, line, bar, heatmap, scatter, streaming, and sparkline models. The
maintainers explicitly say its v2 designation is Bubble Tea compatibility and its own API
may still change. Therefore dedicated chart issue #463, after #411, must:

1. obtain the human dependency approval before editing `go.mod`;
2. pin and record the reviewed version;
3. wrap it behind a small local chart interface;
4. test resize, large-result bounds, no-color/ASCII/text fallback, and accessible summaries;
5. retain a minimal internal sparkline/horizontal-bar fallback if approval is withheld.

The isolated quickstart rendered successfully with Bubble Tea v2 during this study. That is
compatibility evidence, not dependency approval.

## Bubble Tea implementation rules

### Progressive action commands and agent invocations

Ordinary bare namespaces remain concise help. The human-first query/reverse workspaces use their
bare commands on eligible dual-TTYs and retain explicit aliases. Action
commands use the mature CLI pattern of progressively filling missing inputs: after the dual-TTY
gate passes, incomplete `pm credentials add [name]` and `pm connections create [name]` invocations
ask only for missing fields, while fully specified invocations execute directly. This matches
GitHub CLI's interactive-without-arguments/noninteractive-with-flags pattern and the CLI Guidelines'
rule that prompts must always have a complete flag alternative.

The machine/agent profile is the existing `--json --no-input` pair, with `--progress ndjson` for
long-running commands. It is intentionally not named `--agent-mode`, because query already owns
`--agent-mode summary|stream` for result shaping. Missing input under the agent profile is a single
structured, actionable usage/validation envelope; prompting, ANSI, and unexpected stdin reads are
forbidden.

Credential guidance handles only non-secret config and secret-source metadata. Environment-backed
secrets flow through existing `--from-env`; a controlled-stdin choice emits a sanitized
`--value-stdin` handoff and performs no save. Connection guidance derives choices from service
metadata and treats duplicate names as no-write recovery states, never implicit updates.

- `Update` is deterministic; blocking I/O and timers return `tea.Cmd`.
- Root models own mode, focus, layout class, key bindings/help, cancellation, and child models.
- Send key messages to the focused child; broadcast only resize/theme/event/cancel messages
  that truly apply globally.
- Business packages emit typed events and never import `internal/ui`.
- Sanitize/redact dynamic data before it reaches styles/View; approval tokens, credentials,
  headers, request bodies, query strings, and secret-bearing errors never reach final frames,
  transcripts, logs, screenshots, accessibility output, JSON, shell-equivalent command text, or
  fixtures.
- Inline mode is the default for run commands; alt screen is reserved for browsers/grids/
  pagers where entering/leaving a place is expected.
- The command layer enters Bubble Tea/Huh only when stdin and stdout are TTYs and no `--json`,
  `--plain`, or `--no-input` bypass flag is set. Piped or non-TTY stdin always takes the
  deterministic plain/noninteractive path; models/prompts must not consume scripted stdin
  unexpectedly, must not hang waiting for a user, and must not open `/dev/tty` to bypass the gate.
  Sequential prompting is allowed only for explicit accessible mode after this same gate passes.
- Mouse, OSC52, Kitty graphics, and progressive keyboard protocols are optional accelerators.
  Every operation remains possible with basic keyboard input and glyph/text rendering.

## Verification contract for every TUI phase

Before production edits, the GSD plan must define RED tests for:

- Normal/Filter/Edit mode and printable-key conflicts;
- arrows + Vim key equivalence, focus order, help accuracy, and one-layer `esc` behavior;
- wide/standard/compact/guard rendering;
- loading/empty/partial/failure/cancel/final frames;
- no-color, ASCII, reduced-motion, and accessible/plain transcripts;
- control-character sanitation and secret redaction;
- bounded chart/query data and truthful units/downsampling labels;
- TUI/prompt gate matrix: stdin+stdout TTY activation, `stdin-piped+stdout-TTY` fallback,
  `stdout-piped` fallback, `CI=1`, `PM_NO_TUI=1`, `TERM=dumb`, `--json`, `--plain`, and
  `--no-input`;
- unchanged exit code/stdout/stderr/one-envelope contracts;
- cancellation and goroutine cleanup under `-race`.

Use headless semantic/golden frames at 160×45, 100×30, 80×24, compact, and below minimum.
Screenshots support the visual review; they never replace key/state/accessibility tests.

## Primary sources

- [Bubble Tea](https://github.com/charmbracelet/bubbletea),
  [v2 upgrade guide](https://github.com/charmbracelet/bubbletea/blob/main/UPGRADE_GUIDE_V2.md),
  and [official examples](https://github.com/charmbracelet/bubbletea/tree/main/examples)
- [Bubbles](https://github.com/charmbracelet/bubbles) and
  [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- [GitHub CLI accessibility guide](https://accessibility.github.com/documentation/guide/cli/)
  and [GitHub's CLI accessibility engineering article](https://github.blog/engineering/user-experience/building-a-more-accessible-github-cli/)
- [GitHub CLI interactive repository creation](https://cli.github.com/manual/gh_repo_create),
  [prompt-disable environment](https://cli.github.com/manual/gh_help_environment),
  [Command Line Interface Guidelines](https://clig.dev/), and
  [Pulumi non-interactive CLI conventions](https://www.pulumi.com/docs/iac/cli/commands/pulumi_do/)
- [bpytop](https://github.com/aristocratos/bpytop),
  [Conky](https://github.com/brndnmtthws/conky), and
  [CAVA](https://github.com/karlstav/cava)
- [LazyGit](https://github.com/jesseduffield/lazygit) and its
  [keybinding reference](https://github.com/jesseduffield/lazygit/blob/master/docs/keybindings/Keybindings_en.md)
- [LazyDocker](https://github.com/jesseduffield/lazydocker),
  [fzf](https://github.com/junegunn/fzf), and [Gum](https://github.com/charmbracelet/gum)
- [awesome-terminal-aesthetics](https://github.com/kud/awesome-terminal-aesthetics)
- [NTCharts](https://github.com/NimbleMarkets/ntcharts)
