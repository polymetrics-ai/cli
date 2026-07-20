# pm Terminal UX/UI Design — the Interactive Layer

Status: DESIGN — companion to `docs/plans/cli-architecture-v2-improvement-plan.md` (Pillar B),
`docs/adr/0003-interactive-tui-layer.md`, and the evidence-backed cross-surface contract in
`docs/design/terminal-ui-research-and-design-system.md`. Library facts verified July 2026 against the
charm.land v2 line (bubbletea v2.0.8, bubbles v2.1.1, lipgloss v2.0.5, huh v2.0.3,
glamour v2.0.1, colorprofile v0.4.3, Evertras/bubble-table v0.22.3).

The subject: `pm` moves records through pipelines. Sources flow into a local warehouse,
queries shape them, actions push them back out. The design's job is to make that movement
visible — and to make creating a pipeline as safe as running one. The audience is two-fold
and equal: humans at a TTY, and agents reading JSON/NDJSON. Nothing here may cost the agent
contract anything.

---

## 1. Design language

### 1.1 One rule above all: two doors, one house

Every capability has two doors — a **flag door** (the API: scriptable, JSON, deterministic)
and a **TTY door** (the experience: interactive, styled, live). The flag door is load-bearing
and always exists first. The TTY door opens only when `ui.Detect` says so: stdin TTY ∧ stdout
TTY ∧ ¬`--json` ∧ ¬`--plain` ∧ ¬`--no-input` ∧ `PM_NO_TUI`/`CI` unset ∧ `TERM≠dumb`.
Everything behind it must degrade back to the flag door's output when stdin or stdout is piped.
Piped or non-TTY stdin is treated as scripted input for the flag/plain path only; TUI and Huh
prompt code must not consume it unexpectedly, must not hang waiting for interaction, and must not open `/dev/tty`
to bypass the gate. `--plain`, `--json`, and `--no-input` always stay on the flag door: no Bubble
Tea, no Huh, and no prompts. Explicit accessible mode may replace redraws with sequential prompts
only after the same stdin+stdout TTY gate passes and no bypass flag is set. A wizard's last act is
always to print the sanitized flag-door equivalent of what it just did, omitting secret values and
one-time authorization values.

### 1.2 Signature element: the pipeline rail

The one bold element, spent deliberately: a vertical rail that draws the flow DAG as a
living thing — step glyphs joined by `│ ├ └` connectors, records counting up beside the
active step. It appears, scaled, in three places and nowhere else:

- `pm flow run` / `pm etl run` — full rail with live counters (the dashboard).
- `pm flow create` — miniature rail growing in the preview pane as steps are added.
- `pm flow status` — static rail of checkpoint states.

Because the rail encodes something true (the actual DAG from `flow.BuildDAG`), it is
structure-as-information, not decoration. Everything around it stays quiet: no boxes for
the sake of boxes, no gradients, one accent color.

### 1.3 Palette

Semantic tokens only — components never name raw colors. Resolved at runtime by
`lipgloss.LightDark(isDark)` where `isDark` comes from `tea.BackgroundColorMsg`
(requested via `tea.RequestBackgroundColor` in `Init`) — never assume a dark terminal.

| Token | Meaning | Dark terminals | Light terminals | Paired glyph+word |
|---|---|---|---|---|
| `flow` | brand accent; running, selection, rail, progress fill | `#2DD4BF` teal | `#0F766E` teal | `● running` |
| `ok` | success | `#4ADE80` | `#15803D` | `✓ ok` |
| `warn` | attention, partial, rate-limited | `#FBBF24` amber | `#B45309` amber | `▲ warning` / `◐ partial` |
| `fail` | failure | `#F87171` | `#B91C1C` | `✗ failed` |
| `dim` | pending, skipped, chrome, help text | `#6B7280` | `#9CA3AF` | `○ pending` / `– skipped` |
| `ink` | primary text | terminal default fg | terminal default fg | — |

Rules:

- **Blue/teal–amber is the primary state axis.** Red/green appear only for terminal
  (finished) states and never alone — the glyph and the word carry the meaning; color is
  reinforcement. A colorblind user reading glyphs only loses nothing.
- All styled *static* output (piped-safe summaries, pretty tables) goes through lipgloss
  print helpers so colorprofile degrades TrueColor→256→16→none automatically and honors
  `NO_COLOR`, `CLICOLOR=0`, and `TERM=dumb`. Non-TTY output gets no ANSI at all
  (and remains under `safety.SanitizeTerminal` as today).
- `ui accessible_colors` setting (config + `PM_ACCESSIBLE_COLORS`) remaps every token to
  the standard 16 ANSI colors so the user's own terminal palette controls contrast
  (gh CLI pattern).

### 1.4 Typography of a monospace grid

The terminal's "type system" is weight, spacing, and case — used intentionally:

- **Bold** = the one thing on screen the user acts on (selected row, active field, final
  status line). Never bold whole paragraphs.
- **Faint/dim** = chrome: keybinding footers, column rules, timestamps.
- Column alignment is a grid: numbers right-aligned, names left-aligned, counters use
  thousands separators (`12,480`), durations compact (`00:42`, `3m12s` past a minute).
- Sentence case everywhere. No ALL-CAPS headers except the man-page surfaces that already
  use them.

### 1.5 Glyph vocabulary

Single-cell, screen-reader-tested, always paired with a word. ASCII fallback column applies
when the color profile is `Ascii`/`NoTTY` or `PM_ASCII=1`:

| State | Glyph | ASCII | Rail connector | ASCII |
|---|---|---|---|---|
| ok | `✓` | `[ok]` | `│` | `|` |
| failed | `✗` | `[x]` | `├` | `+` |
| running | `●` (+ spinner) | `[*]` | `└` | `\`` |
| pending | `○` | `[ ]` | | |
| partial | `◐` | `[~]` | | |
| warning | `▲` | `[!]` | | |
| skipped | `–` | `[-]` | | |

No braille spinners (they confuse screen readers — gh's finding); the spinner is
`bubbles/spinner` `Dot` on TTY, and a static `● running` when spinners are disabled.

### 1.6 Layout rules

- **Inline mode by default** for run commands — the final frame persists in scrollback as
  the human summary; no alt-screen, so a `pm flow run` in a CI-attached TTY still leaves a
  readable transcript. Alt-screen is reserved for browsers and pagers (connectors browse,
  docs view, query grid) where the user expects to "enter" and "leave" a place.
- **Help footer on every screen** (`bubbles/help` + `bubbles/key`): short, current-mode
  help always visible; `?` toggles the complete binding list. Keys work in both dialects —
  arrows *and* Vim (`j/k/gg/G/ctrl+d/ctrl+u`), `tab`/`shift+tab` between panes,
  `enter` activates, `esc` steps back one layer, `q` quits only in Normal mode, and
  `ctrl+c` requests cancellation while allowing a truthful final frame.
- **Responsive classes**: wide (120+ columns), standard (80–119), compact (60–79), and a
  measured size guard below 60×18. Feature layouts may require 80×24 for enhancement, but
  compact terminals receive a one-pane view or useful guard rather than a broken layout.
  Standard layouts stack detail before truncating identifiers.
- Wide layouts use `lipgloss.JoinHorizontal`; every pane owns its width from
  `tea.WindowSizeMsg` minus frame sizes — no hardcoded widths.

### 1.7 Copy voice

Words are design material (they make the interface usable, nothing else):

- Labels name what the user controls: "Connection", "Streams", "Runs at" — never
  "endpoint descriptor" or "cron spec".
- Actions say what happens and keep their name through the flow: the button that says
  **Create flow** produces `✓ Flow created`.
- Errors state what happened and the exact fix, in the interface's voice, no apology:
  `cron needs 5 fields — try "0 2 * * *" or choose a preset`.
- Empty states are invitations: `No flows yet. Create one: pm flow create`.
- Every wizard's final frame teaches the flag door:
  `Next: pm flow run likely-customers` plus the full equivalent command.

### 1.8 Interaction modes and focus

The primary structural reference is LazyGit's operator workspace, combined with fzf's
filter/list/preview interaction, bpytop's exact telemetry density, and Gum's focused wizard
cadence. Polymetrics keeps its own restrained palette, pipeline rail, safety gates, and
plain/JSON contract.

- Surfaces start in **NORMAL** mode. `/` enters **FILTER**; `i`/`e` or a focused form field
  enters **EDIT**. Mutations enter an explicitly labelled **CONFIRM** view only after
  preview. `?` opens **HELP** for the current mode.
- The footer displays mode and focus (`NORMAL · results`). One accent treatment identifies
  focus in addition to text; color alone is insufficient.
- Printable keys belong to focused Filter/Edit inputs. In those modes `j`, `q`, `/`, and
  `?` insert text or perform the input component's documented edit behavior; they are not
  stolen by global navigation.
- `esc` unwinds exactly one layer: Help/Edit/Filter → Normal → prior surface. `q` quits only
  in Normal. `ctrl+c` requests cancellation everywhere and does not bypass cleanup.
- Vim familiarity stops at navigation. Do not implement registers, macros, operator grammar,
  or command-line mode. Every Vim key has an arrow/home/page/tab alternative.

---

## 2. Surface designs

Each surface lists: wireframe (80-col), components, keys, accessible-mode behavior, and
agent parity. All dynamic strings pass `safety.SanitizeTerminal` + `RedactErrorText`
before styling (view hygiene — the TUI-path equivalent of the plain path's sanitizer).

### 2.1 `pm flow run` / `pm etl run` — the run dashboard (flagship)

```
  Flow likely-customers                                      elapsed 00:42
  ──────────────────────────────────────────────────────────────────────
  ✓ sync-hubspot     sync    hubspot-prod: contacts, companies
  │                          12,480 read → 12,480 written        00:28
  ● score-contacts   query   ▓▓▓▓▓▓▓▓░░░░░░  6,214 rows          00:14
  │
  ○ export-scored    action  waiting on score-contacts
  ──────────────────────────────────────────────────────────────────────
  ctrl+c cancel (checkpoints kept) · ? help
```

- Final frame on success:
  `✓ Flow likely-customers finished — 3 steps, 12,480 records, 00:51`
  On cancel: `– Cancelled after sync-hubspot. Resume: pm flow run likely-customers`
  On failure: `✗ score-contacts failed — <redacted error>. Logs: .polymetrics/logs/<run-id>.jsonl`
- Components: pipeline rail (custom), `spinner`, `progress` (only when a step reports a
  known total; otherwise counter-only), `help`. Inline mode.
- Events: model subscribes to `internal/events` `Chan`; `ctrl+c` cancels the **engine
  context**, the model waits for `DoneMsg` before quitting so the checkpoint file is
  flushed and the final frame is truthful.
- Accessible/plain: with spinners disabled or `--accessible`, emits one plain line per
  state change (`step score-contacts running…`, `step score-contacts ok 6,214 rows 00:14`)
  — same data, sequential, no redraws.
- Agent parity: `--json` unchanged (final envelope); `--progress ndjson` streams the same
  events to stderr.

### 2.2 `pm flow create` — the flow wizard (new command)

Two-pane alt-screen: huh form (left) + growing rail preview (right).

```
  Create flow                                          Step 2 of 4: steps
  ┌─ Add step ────────────────────────────┐  Preview
  │ Kind                                  │
  │ > sync      pull a stream in          │  ✓ sync-hubspot   sync
  │   query     shape tables with SQL     │  │   out: contacts, companies
  │   rlm       score records             │  ● (adding step 2…)
  │   action    push records out          │
  │                                       │
  │ Connection                            │
  │ > hubspot-prod   hubspot: contacts…   │
  │                                       │
  │ Streams (space to toggle)             │
  │ [x] contacts   [x] companies  [ ] deals│
  └───────────────────────────────────────┘
  enter next · esc back · ? help
```

- huh v2 groups: (1) name + description (`Input` with `ValidateNotEmpty` + the
  `^[A-Za-z0-9_-]+$` rule from `flow.ParseManifest`), (2) repeating step loop, (3) review,
  (4) confirm. Kind-dependent pages via `WithHideFunc`; option sets via `OptionsFunc`:
  - **sync** → Connection `Select` (`App.ListConnections`), Streams `MultiSelect`
    (`App.ShowCatalog`); `out` tables default to stream names.
  - **query** → SQL `Text` field; a `Note` lists tables available *so far* (upstream
    `out` sets) — the reference panel.
  - **rlm** → spec `FilePicker` (`*.json`), mode `Select` (deterministic/fixture/model/agent).
  - **action** → destination connector/credential `Select`s, action verb
    (upsert/create/delete), source-table `Select` **restricted to upstream `out` tables**.
  - **`in` wiring is structural**: the `in` picker only offers tables some earlier step
    produced — the whole class of hand-wired DAG errors (mismatched `in`/`out` names)
    cannot be expressed.
- Finish: round-trip through `flow.ParseManifest` (the same validator `run` uses), write
  `.polymetrics/flows/<name>.json`, final frame:

```
  ✓ Flow created  .polymetrics/flows/likely-customers.json
    3 steps · sync-hubspot → score-contacts → export-scored
  Next: pm flow run likely-customers
  Scripted equivalent: pm flow plan --file .polymetrics/flows/likely-customers.json
```

- Accessible mode: the entire form runs through huh `WithAccessible` — numbered sequential
  prompts, no redraw, no preview pane (the review page prints the rail as indented text).
- Agent parity: agents keep writing manifest JSON directly (schema documented in
  `docs/cli/flow.md`); the wizard's output artifact is that same file. `--no-input` on
  `flow create` errors: `flow create is interactive — write the manifest directly and run: pm flow plan --file <path>`.

### 2.3 `pm schedule create` — schedule wizard (existing flags remain the API)

Prompts only for what's missing; with all of `--name --cron --flow` present it behaves
exactly as today.

```
  Create schedule                                    Step 2 of 3: timing
  Flow            likely-customers  (picked from .polymetrics/flows)
  Runs at
  > Nightly at 02:00          0 2 * * *
    Hourly                    0 * * * *
    Every 15 minutes          */15 * * * *
    Weekdays at 09:00         0 9 * * 1-5
    Custom (5-field cron)…
  Next runs   Jul 17 02:00 · Jul 18 02:00 · Jul 19 02:00
  Backend     launchd (auto-detected — darwin)
```

- Flow picker validates existence (fixing today's unvalidated `--flow`); the custom-cron
  input validates with `schedule.ParseCron` and shows the parse error inline in copy-voice;
  **next-3-fire-times preview** uses the existing dead-code `schedule.Next`
  (`internal/schedule/cron.go:132`). Backend `Select` defaults to what `SelectBackend`
  would pick, showing why. Final `Confirm`: install now?
- Final frame: `✓ Schedule created — nightly-leads runs likely-customers at 0 2 * * *`
  plus `Scripted equivalent: pm schedule create --name nightly-leads --cron "0 2 * * *" --flow likely-customers && pm schedule install nightly-leads`.
- Accessible: huh accessible mode; the fire-times preview prints as plain lines.
- Agent parity: flags unchanged; `--json` envelope unchanged.

### 2.4 Query experience

**`pm query tables` (new plain command — agents and wizard share it):** enumerates
`<projectDir>/warehouse/*.jsonl` (same derivation as the DuckDB engine's `registerViews`)
with row counts and sizes. Plain: `name\trows\tbytes`; `--json`: `{"kind":"WarehouseTables",...}`.
This closes the "users must guess `--table`" gap for everyone, not just the TUI.

**Bare namespace:** `pm query` with no action selected renders contextual help and the
subcommand summary, then exits 0 in both TTY and non-TTY contexts. It never starts the grid.
Invalid actions still return usage errors.

**`pm query grid` (explicit TTY subcommand):** alt-screen browser.

```
  Query  warehouse: 4 tables
  ┌─ SQL ────────────────────────────────────────────────────────────┐
  │ SELECT city, count(*) FROM contacts GROUP BY city ORDER BY 2 DESC│
  └──────────────────────────────────────────────────────────────────┘
  ┌────────────────┬──────────┐   contacts     12,480 rows
  │ city           │ count(*) │   companies     1,733 rows
  ├────────────────┼──────────┤   scored_leads  6,214 rows
  │ Berlin         │    2,110 │   runs             12 rows
  │ Amsterdam      │    1,874 │
  │ Austin         │    1,551 │   page 1/7 · sorted by count(*) ↓
  └────────────────┴──────────┘
  enter run · tab focus · s sort · / filter · e export · q quit
```

- Components: `textinput` with table-name autocompletion (bubbletea autocomplete pattern),
  **Evertras/bubble-table** grid (pagination via LIMIT/OFFSET re-query, column sort,
  built-in filter), table list pane fed by the `query tables` enumerator.
- Read-only safety unchanged: everything runs through the existing `QuerySQL` path and its
  `validateSelectOnly` guard; the TUI adds no SQL write path.
- Export (`e`): a typed read-only result export, not a generic file-write tool. It writes only
  JSONL/CSV rows returned by `QuerySQL` to a default project-scoped output directory such as
  `.polymetrics/query-exports/`. The path validator rejects control characters, traversal,
  absolute or broad roots, empty/bare directory targets, and paths outside the project after
  resolve/clean. It must confine the final clean path to the project, reject symlink targets and
  final-component symlink races, and open with exclusive create so existing files are not
  overwritten by default. Export asks for explicit confirmation only when both stdin and stdout are
  TTYs; noninteractive export, including piped/non-TTY stdin, requires both an explicit
  `--output <project-relative-path>` and `--force`. The scripted
  equivalent echo is sanitized and project-relative, for example
  `pm query run --sql "<sanitized select>" --limit 500 --output .polymetrics/query-exports/result.jsonl --format jsonl --force`.
  With `--no-input` and no preapproved output, fail exactly:
  `query grid export requires explicit output — pass --output <project-relative-path> and --force, or run without --no-input to confirm interactively`.
- Bypass and accessibility: `pm query grid --plain`, `--json`, and `--no-input` never run a
  sequential SQL prompt. They bypass Bubble Tea, Huh, and prompts, then produce deterministic
  table/summary output when required flags such as `--sql` are present, or the exact required-flag
  error asserted by the implementation issue. Explicit accessible mode may use a sequential SQL
  prompt only when stdin and stdout are TTYs and none of `--plain`, `--json`, or `--no-input` is
  set; piped/non-TTY stdin uses the deterministic plain path and requires flags rather than
  consuming scripted input. Bare `pm query` remains contextual help, not a grid launcher.
- Agent parity: `pm query run` untouched; `query tables` is plain/JSON first.

**Query charts (dependency-gated child issue #463):** after #411 lands the result grid may
toggle (`v`) between
the exact table and one selected visualization over the already-returned, read-only result.
The chart setup chooses type, X, Y, aggregation, unit, and sort from validated result
metadata; it does not synthesize or execute SQL. Initial grammar: time series, sparkline,
sorted horizontal bars, histogram, scatter, and heatmap only when the data/question fits.
Every chart shows axes/units and selected exact values, retains the table/text summary,
bounds and deterministically downsamples points, and discloses sampling/missing values.
No pie charts, 3D effects, dual axes, decorative waveforms, or color-only meaning.

`github.com/NimbleMarkets/ntcharts/v2` is the leading Bubble Tea v2-compatible candidate,
but its own API is still described by its maintainers as subject to change. Adding it to
`go.mod` in #463 requires the isolated spike evidence, exact pin, local wrapper, and explicit human
dependency approval. A minimal internal sparkline/horizontal-bar renderer is the fallback.

### 2.5 `pm connectors browse` — the 551-connector browser

TTY-enhanced replacement for reading a 551-row dump. Alt-screen, split view.

```
  Connectors  551 available                       filter: "hub"  3 matches
  ┌──────────────────────────┬───────────────────────────────────────┐
  │ > hubspot      crm       │  hubspot                              │
  │   hubplanner   planning  │  CRM platform. read ✓  write ✓        │
  │   githubhub…   devtools  │  Streams: contacts, companies, deals… │
  │                          │  Auth: private app token (secret)     │
  │                          │  ─ from MANUAL, scrolls ─             │
  └──────────────────────────┴───────────────────────────────────────┘
  / filter · r/w/q capability · s stage · enter manual · c copy setup · q quit
```

- Components: `bubbles/list` with a custom delegate (name + integration type + capability
  glyphs), instant fuzzy filter with match highlighting (`DefaultFilter` +
  `MatchesForItem`), capability/stage toggles re-filter the item set; preview pane =
  `viewport` rendering `RenderConnectorManual` output; `enter` promotes the preview to a
  full-screen pager; `c` copies `pm credentials add <name>-cred --connector <name>` via
  OSC52.
- Plain path fix (piped or `--json` unchanged shape, human format improved in the same
  phase): the `%+v` capabilities dump in `pm connectors list` becomes the existing
  `read=… write=… query=…` columns already used by `--all`.
- Accessible: browse may use a sequential `filter? capability?` flow only in explicit accessible
  mode after the stdin+stdout TTY gate passes and no `--plain`, `--json`, or `--no-input` bypass
  flag is set. Bypass and piped paths print deterministic plain/list output — or users simply use
  `connectors list`, which stays first-class.
- Agent parity: `connectors list/catalog/inspect --json` untouched.

### 2.6 `pm docs view` — the docs pager

Glow pattern: glamour v2 `TermRenderer` (auto light/dark style from the detected
background, word wrap to pane width) inside a `viewport` pager.

- `pm docs view` → topic browser (command manuals from the docs map + connector docs).
- `pm docs view etl` / `pm docs view hubspot` → straight into the pager.
- Piped: renders plain text exactly as `pm help <topic>` does today. `$PAGER` respected
  with `--pager`; `q`/`esc` leave; `/` searches within the page.
- Sources need zero new plumbing: command manuals are the docs-map strings; connector
  manuals come from `RenderConnectorManual`/`GuideOf` (already clean sectioned text).

### 2.7 `pm connectors certify` — batch dashboard

```
  Certifying 12 connectors                               8 done · 00:03:12
  ────────────────────────────────────────────────────────────────────────
  ✓ airtable      passed        00:12      ✓ asana       passed     00:09
  ✗ amplitude     read failed   00:31      ● hubspot     running…
  ◐ braintree     partial       00:22      ○ intercom    queued
  ────────────────────────────────────────────────────────────────────────
```

- `bubbles/table` (or the run-dashboard row style) updated concurrently from the events
  bus (`certify.RunBatch` worker pool emits per-connector events); summary line carries the
  rolled-up exit code; certify's 0/1/2/3 exit contract unchanged.

### 2.8 `pm rlm --mode agent` — agent run viewer

Workflow ID, state, heartbeat age (from the Temporal `DescribeWorkflowExecution` poller),
attempt count, elapsed, and a liveness pulse. `● running · heartbeat 2s ago · attempt 1`.
When later phases add a Temporal Query handler with iteration payloads, the same view gains
`iteration 3/10` — additive, no redesign.

### 2.9 Guided reverse-ETL session (`pm reverse guide`)

Bare `pm reverse` renders contextual help and the subcommand summary, then exits 0. It never
starts the guided session. Invalid reverse actions still return usage errors.

`pm reverse guide` walks the existing gate — the gate itself is untouched:

1. **Plan** — pick connection/table/action (huh), runs `reverse plan`, shows the non-sensitive
   plan ID.
2. **Preview** — records table (bubble-table) of what would be written.
3. **Approve** — the approval token is a sensitive one-time authorization value. The guided
   flow may carry it ephemerally in memory only long enough to pass through the existing
   plan → preview → approval → execute seam. It is never printed in final frames, transcripts,
   logs, screenshots, accessibility output, JSON, shell-equivalent command text, or phase
   fixtures. Destructive actions still surface the typed-confirmation challenge exactly as the
   flag flow does (`--confirm` semantics):
   `Type the table name to approve writing 214 records to hubspot contacts:`.
4. **Run** — live progress; final frame prints run ID + `reverse status` command, with no token.

The session teaches the 4-command flag flow with sanitized commands and IDs only. Agents keep the
canonical plan → preview → approval → execute path; the TUI does not weaken typed approval or expose
approval tokens.

### 2.10 `credentials add` / `connections create` prompting

`connections create` gains missing-input prompts (source/destination pickers from
credentials list, stream picker from catalog, sync-mode select where
`RequiresCursor`/`RequiresPrimaryKey` drive follow-up fields).
**`credentials add` stays non-interactive for secret values** — pickers may help with
connector/field *names*, but values arrive only via `--from-env`/`--value-stdin`
(no interactive secret entry; a second, less-auditable intake path is a non-goal).

---

## 3. Accessibility spec (becomes the `pm a11y` help topic)

1. **Prompt accessibility**: every huh flow honors `--accessible`,
   `PM_ACCESSIBLE_PROMPTER`, and generic `ACCESSIBLE` only after stdin and stdout are TTYs and no
   `--plain`, `--json`, or `--no-input` bypass flag is set → sequential numbered prompts, no
   redraws (huh `WithAccessible`). Bypass flags and non-TTY stdin/stdout never prompt.
2. **Reduced motion**: `ui.spinner: disabled` config (and implied by accessible mode) →
   static `● running` lines with action-specific text; progress becomes periodic plain
   updates.
3. **Color**: NO_COLOR / CLICOLOR / TERM=dumb honored everywhere (colorprofile);
   `accessible_colors` remaps tokens to the 16-color terminal palette; color is never
   the only carrier (glyph + word always present).
4. **Screen-reader-safe structure**: positional context is textual
   (`Step 2 of 4: steps`), no braille spinners, tables render with header words not just
   rules.
5. **Keyboard**: everything reachable without a mouse; arrows and vim keys; `esc` back,
   `q` quit, `ctrl+c` force-quit, `?` help — consistent on every screen.
6. **Size**: minimum-size notice instead of broken layouts; panes stack before truncating.
7. **Windows Terminal**: inline-mode defaults, single-cell glyphs with ASCII fallbacks,
   Bubble Tea v2 renderer handles console modes; `PM_NO_TUI=1` is the universal escape.
8. **Non-TTY stdin or stdout**: never a TUI, never ANSI, never Huh prompts; identical to today's
   deterministic plain/noninteractive output.

## 4. Agent UX parity (first-class, not a footnote)

- `--json` envelopes: byte-identical everywhere; one envelope per invocation preserved.
- `--progress ndjson`: live sanitized events on **stderr** for any long-running command —
  agents get what the TUI gets, in their dialect, without polling checkpoint files.
- Every TUI surface has a plain sibling: browse↔`connectors list`, query grid↔`query run`,
  docs pager↔`help/man`, wizards↔flags+manifest files, `query tables` is plain-first.
- Wizards emit machine artifacts (manifest JSON at a documented path) and echo the sanitized
  scripted equivalent — TTY sessions teach the API without printing secrets or one-time
  authorization values.
- `--json`, `--plain`, `--no-input`, `CI`, and non-TTY stdin or stdout hard-disable prompting; an
  interactive-only path errors by naming the flag/file to provide instead (clig.dev rule).
- New enumerators added for the TUI (`query tables`) land as plain/JSON commands first, so
  the agent surface grows with the human one.

## 5. Component map, dependencies, testing

| Surface | Components | New deps (first use) |
|---|---|---|
| Run dashboards (2.1, 2.7, 2.8) | events bus, pipeline rail, spinner, progress, help/key | bubbletea/bubbles/lipgloss v2 |
| Wizards (2.2, 2.3, 2.10) | huh v2 groups (embedded as `tea.Model`), rail preview | huh v2 |
| Query grid + charts (2.4) | textinput+autocomplete, bubble-table, viewport; dependency-gated chart renderer | evertras/bubble-table; proposed ntcharts/v2 requires separate human approval |
| Connectors browser (2.5) | list (custom delegate), viewport preview, OSC52 copy | — |
| Docs pager (2.6) | glamour TermRenderer in viewport | glamour v2 |
| Gate & styles | `ui.Detect` (x/term), `ui/styles` tokens, colorprofile | golang.org/x/term |

Import law (CI-enforced with a `go list` check): business packages
(`internal/{app,flow,connectors,worker,rlm,schedule}`) may import `internal/events` and
nothing under `internal/ui`; `internal/ui/**` imports events/safety/charm and no business
packages — models consume `events.Event` plus small planned-shape structs passed by the
command layer; only the command layer branches plain-vs-TUI.

Testing:

- **Models**: teatest/v2 golden frames per model (happy path, failure, cancel, narrow
  terminal), driven headlessly — no TTY in CI. Cover 160×45, 100×30, 80×24, compact, and
  below-minimum frames plus Normal/Filter/Edit mode conflicts and one-layer `esc` behavior.
- **Gate**: `ui.Detect` table tests for stdin+stdout TTY activation, `stdin-piped+stdout-TTY`
  fallback, `stdout-piped` fallback, `CI`, `PM_NO_TUI`, `TERM=dumb`, `--json`, `--plain`,
  `--no-input`, and proof that no `/dev/tty` bypass is used.
- **Contract**: existing agentic contract suite runs unchanged (plain path is the default
  of the untouched `cli.Run`); one added test per TUI-enabled command asserting
  `stdin-piped+stdout-TTY`, `stdout-piped`, `CI=1`, `PM_NO_TUI=1`, `--json`, `--plain`, and
  `--no-input` force plain/noninteractive output and `--progress ndjson` writes nothing to stdout
  beyond the single envelope.
- **Wizard outputs**: round-trip written manifests through the same parser `run` uses
  (`flow.ParseManifest`) — the wizard cannot produce a manifest the engine rejects.
- **Charts**: chart selection/units, bounded data, deterministic bucketing, missing values,
  exact selected-value text, table fallback, no-color/ASCII/accessibility transcripts, and
  resize behavior are test contracts before a renderer is wired.
- **View hygiene**: red tests feeding `\x1b[31m` and `token=…` strings through step
  names/errors into views, asserting sanitized/redacted rendering.
