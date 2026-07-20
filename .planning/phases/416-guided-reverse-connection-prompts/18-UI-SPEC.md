---
phase: 18
issue: 416
child_issue: 469
slug: guided-reverse-connection-prompts
status: draft
shadcn_initialized: false
preset: not-applicable-terminal-ui
created: 2026-07-20
gsd_path: adapter-prompt-with-inline-ui-check
---

# Phase 18 — Guided reverse ETL and progressive setup UI design contract

> Implementation-ready visual, interaction, safety, copy, and parity contract for reverse issue
> [#416](https://github.com/polymetrics-ai/cli/issues/416) and setup child
> [#469](https://github.com/polymetrics-ai/cli/issues/469). This contract is a TTY-gated
> projection over the existing CLI and application services; flags, plain output, JSON, and NDJSON
> remain the API.

## Status and provenance

`scripts/gsd doctor` passed with 69 registered commands on 2026-07-20 and
`scripts/gsd prompt ui-phase 18 --text` generated the project-local Phase 18 UI workflow prompt.
The roadmap lookup used by the native SDK still reports this issue-backed phase as non-native:
`gsd-sdk query init.plan-phase 18` returned `phase_found: false`: CLI Architecture v2 records Phase
18 in `.planning/ROADMAP.md` as issue #416 with setup child #469 rather than as a native GSD roadmap phase. This document
therefore applies the generated workflow inline. No user questions were needed because the human-
first entry decision and every safety boundary were explicitly locked by the user.

The Phase 18 planner and executor must load and record these skills before implementation:

- `bubble-tea-tui-design`
- `golang-how-to`
- `golang-cli`
- `golang-testing`
- `golang-security`
- `golang-safety`
- `golang-error-handling`
- `golang-context` and `golang-concurrency` for cancellation or asynchronous metadata loading
- `golang-documentation` for help/manual/website parity

## Phase outcome

Phase 18 specifies three related, bounded surfaces split across two implementation issues:

1. `pm credentials add [name]` asks only for missing non-secret input on a genuine interactive
   terminal. It may collect secret-source metadata, never plaintext secret values.
2. `pm connections create [name]` derives valid choices and defaults from existing credentials,
   connector definitions, catalogs, and sync metadata, then shows a review before creation.
3. Bare `pm reverse` on an eligible dual-TTY opens the guided reverse workspace; the explicit
   `pm reverse guide` alias opens the same model. Both project the existing plan -> preview ->
   approval -> execute workflow without exposing the approval token or weakening typed confirmation.

Issue #469 owns surfaces 1–2. Issue #416 owns surface 3. Each implementation uses one bounded GSD
loop and one stacked PR; neither issue may absorb the other's write scope.

The bare namespaces `pm credentials` and `pm connections` continue to render contextual help/
subcommand summaries and exit 0. Bare `pm reverse` is the narrow human-first exception: it opens
the guided workspace only when the dual-TTY gate passes. On every bypass/non-TTY path it renders
deterministic contextual help and exits 0; help flags always render help.

## Non-goals and hard boundaries

- Do not add a global `--agent-mode`. `pm query run --agent-mode summary|stream` is a
  query-specific result-shaping option; it is not a universal prompt or safety mode.
- Do not add interactive secret entry, password fields, masked secret prompts, clipboard secret
  intake, a secret file flag, or a second vault path. Existing named `--from-env field=ENV` and
  controlled `--value-stdin field` intake remain authoritative.
- Do not add connection update, replace, overwrite, or remove behavior. A duplicate name never
  becomes an implicit update.
- Do not add a new reverse mutation path. Existing plan, preview, approval, typed confirmation,
  execute, status, exit codes, and approval-token handling remain authoritative.
- Do not add generic shell, generic HTTP write, generic SQL write, or generic file-write actions.
- Do not absorb connector list/inspect/browser/capability behavior from
  [#411](https://github.com/polymetrics-ai/cli/issues/411), help-tree/man-page work from
  [#417](https://github.com/polymetrics-ai/cli/issues/417), or connector/certify namespace
  nativization from [#437](https://github.com/polymetrics-ai/cli/issues/437). Phase 18 may link to
  those follow-ups, but it owns only validation and recovery needed by these three guided flows.
- Do not request or add a dependency. At research time `go.mod` contains `golang.org/x/term` but not
  the Charm v2 suite. Phase 18 depends on #409 and #462/D-TUI; it must reuse the approved Bubble Tea
  v2, Bubbles v2, Lip Gloss v2, and Huh v2 versions integrated upstream. If those dependencies have
  not landed, stop at the dependency gate instead of editing `go.mod` or `go.sum`.

## Design system

This is a terminal UI, so the React/shadcn initialization and registry gates are not applicable.
The design system is the existing Polymetrics terminal contract in
`docs/design/tui-ux-design.md`, `docs/design/terminal-ui-research-and-design-system.md`, ADR-0003,
and `internal/ui/styles`.

| Property | Contract |
|---|---|
| Tool | Manual Polymetrics terminal design system; no shadcn |
| Framework | Bubble Tea v2 model/update/view, reusing the upstream-approved phase dependency |
| Form library | Huh v2 groups and dynamic fields; `WithAccessible(true)` only in explicit accessible mode |
| Components | Bubbles v2 `key`, `help`, `list`, `textinput`, `viewport`, and existing approved primitives as needed |
| Styling | Lip Gloss v2 plus `internal/ui/styles` semantic tokens and glyph fallbacks |
| Font | User-configured terminal monospace; Phase 18 must not select or download a font |
| Icons | Existing semantic glyphs with `PM_ASCII=1` and no-TTY fallbacks; no icon dependency |
| Screen model | Alt-screen for the multi-step wizard; sanitized inline outcome after exit |
| Registry | Not applicable; no third-party component registry or block |

### Component inventory

| Component | Responsibility | Must not own |
|---|---|---|
| Guided command preflight | Parse supplied fields, determine completeness, evaluate activation gate | Prompts, view rendering, business writes |
| Wizard root model | Surface, step, mode, focus, layout class, cancellation, sanitized outcome | Connector discovery or blocking I/O in `Update`/`View` |
| Step header | Title plus `Step N of M: label` | Hidden state conveyed only by color |
| Huh form group | One focused decision; dynamic options and same validators as flag path | Plaintext secret value |
| Review panel | Sanitized, complete configuration and mutation impact | Secret values or approval token |
| Duplicate recovery menu | `Inspect existing`, `Choose a new name`, `Cancel connection setup` | Update, remove, replace, overwrite |
| Error panel | What happened, affected field/target, exact safe next step | Raw HTTP bodies, headers, URLs with query strings, secret-bearing errors |
| Outcome frame | Truthful saved/created/tested/cancelled result plus safe next command | Approval token or shell command containing a secret value |
| Contextual footer | Mode, focus, valid short keys, progress/error status | Bindings disabled in the current mode |
| Size guard | Measured size, recommended size, deterministic flag command | Damaged or clipped form controls |

## Spacing scale

The GSD logical scale remains multiples of four. Terminal rendering maps it to indivisible
monospace cells; pixel values document hierarchy only and are not emitted as terminal font/layout
settings.

| Token | Logical value | Terminal mapping | Usage |
|---|---:|---:|---|
| xs | 4px | 1 cell | Glyph-to-label gap, checkbox marker gap |
| sm | 8px | 1 cell | Field label/value gap, compact row padding |
| md | 16px | 2 cells | Default horizontal inset and group spacing |
| lg | 24px | 3 cells | Separation between form and review regions |
| xl | 32px | 4 cells | Wide-layout pane gap |
| 2xl | 48px | 6 cells | Major blank state or guard separation |
| 3xl | 64px | 8 cells | Reserved upper bound; do not create decorative whitespace beyond this |

Exceptions: terminal cells are the implementation unit. There are no pointer-only controls or
44px touch-target exceptions; every action is keyboard reachable and text-labelled.

## Typography

Only two weights are allowed: regular and bold. Actual point size is owned by the user's terminal;
the sizes below are semantic equivalents used by the UI checker and documentation previews.

| Role | Semantic size | Weight | Terminal rendering | Line height |
|---|---:|---:|---|---:|
| Label / chrome | 14px equivalent | 400 | normal or faint for nonessential chrome | 1.4 |
| Body / field | 16px equivalent | 400 | terminal default | 1.5 |
| Heading / active item | 20px equivalent | 700 | bold, sentence case | 1.2 |
| Display / final status | 28px equivalent | 700 | one bold outcome line only | 1.2 |

Never use ALL CAPS as a substitute for hierarchy. The mode token remains uppercase because it is a
compact status value: `NORMAL`, `EDIT`, `CONFIRM`, `HELP`.

## Color contract

Terminal-default foreground/background remain the dominant surface. Color reinforces words and
glyphs; it never determines state or selection alone.

| Role | Share | Value | Reserved usage |
|---|---:|---|---|
| Dominant | 60% | terminal default background + `ink` foreground | Form surface, review text, outcomes |
| Secondary | 30% | `dim`: dark `#6B7280`, light `#9CA3AF` | Separators, descriptions, inactive help, metadata provenance |
| Accent | 10% | `flow`: dark `#2DD4BF`, light `#0F766E` | Current focus, selected option, active step, primary CTA, active reverse stage only |
| Success | semantic | `ok`: dark `#4ADE80`, light `#15803D` | `✓ saved`, `✓ created`, `✓ validated`, completed stage only |
| Warning | semantic | `warn`: dark `#FBBF24`, light `#B45309` | Partial outcome, pending external test, destructive impact warning |
| Destructive/failure | semantic | `fail`: dark `#F87171`, light `#B91C1C` | Validation failure, failed operation, destructive reverse confirmation only |

Accent is reserved for current focus, selected list row, active step label, primary CTA, and the
active segment of the reverse rail. It is not applied to every field, border, key hint, link, or
successful result.

`NO_COLOR`, `CLICOLOR=0`, and `TERM=dumb` remove ANSI styling. `PM_ASCII=1` selects text glyphs.
Accessible colors use the existing ANSI-16 mapping. `NO_COLOR` and `PM_ASCII` do not themselves
disable an otherwise eligible TUI; `TERM=dumb` does.

## Exact activation and bypass matrix

The implementation must decide between direct execution, guidance, accessible sequential prompts,
help, and deterministic error before initializing Bubble Tea/Huh. A command is **complete** only
when all inputs required by the selected connector/schema/action are present. A complete but invalid
invocation returns its normal validation error; it does not open a wizard to repair the input.

The gate is:

```text
interactive = stdinTTY && stdoutTTY
           && !json && !plain && !noInput
           && PM_NO_TUI is empty
           && CI is empty
           && lower(trim(TERM)) != "dumb"
```

| Invocation / environment | Complete action | Incomplete `credentials add` / `connections create` | Prompt/TUI contract |
|---|---|---|---|
| Bare `pm credentials` or `pm connections` | N/A | N/A | Contextual help, exit 0, never TUI |
| `--help` or help route | N/A | N/A | Help, exit 0, never TUI |
| Both stdin and stdout TTY, gate clear | Execute directly | Bubble Tea/Huh guidance by default | Ask only missing fields |
| Same, explicit accessible mode | Execute directly | Static sequential Huh prompts | No redraw; announce step/mode/focus |
| `--json` | Direct JSON execution | One JSON `Error` envelope | No TUI/Huh/prompt/ANSI |
| `--plain` | Direct deterministic plain execution | Actionable plain error | No TUI/Huh/prompt/ANSI |
| `--no-input` | Direct selected output mode | Actionable usage/validation error | No TUI/Huh/prompt |
| non-empty `PM_NO_TUI` | Direct deterministic path | Actionable usage/validation error | No TUI/Huh/prompt |
| non-empty `CI` (including `CI=0`) | Direct deterministic path | Actionable usage/validation error | No TUI/Huh/prompt |
| `TERM=dumb` | Direct deterministic path | Actionable usage/validation error | No TUI/Huh/prompt/ANSI |
| stdin piped/non-TTY, stdout TTY | Direct deterministic path | Error before consuming scripted stdin | Never open or read `/dev/tty` |
| stdin TTY, stdout piped/non-TTY | Direct deterministic path | Actionable usage/validation error | No redraw, ANSI, or prompt |
| both stdin and stdout non-TTY | Direct deterministic path | Actionable usage/validation error | Script-safe |
| explicit `--value-stdin field` on an otherwise complete credential command | Read only that field after all non-secret validation passes | Error before reading stdin | Existing vault intake only |
| fully specified invocation with invalid supplied value | Return same direct validation error | N/A | Never switch to TUI as an error-repair fallback |

The reverse workspace uses the same gate with a deliberately different bare-command entry policy:

| Invocation / environment | Result |
|---|---|
| Bare `pm reverse`, stdin+stdout TTY, gate clear | Open the guided reverse workspace |
| `pm reverse guide`, stdin+stdout TTY, gate clear | Open the same guided reverse model |
| Bare `pm reverse` on any bypass/non-TTY path | Deterministic contextual help, exit 0; no Bubble Tea/Huh |
| `pm reverse --help` or `pm reverse guide --help` | Focused help, exit 0; never TUI |
| Invalid reverse action | Usage error; never reinterpret as workspace entry |

Additional invariants:

- Resolve the gate from explicit invocation facts; tests must not depend on process-global stdio.
- Piped stdin must not be consumed merely to discover that another required field is missing.
- The TUI must not open `/dev/tty` or use Bubble Tea v2's automatic TTY input behavior to override
  the product gate.
- `--json`, `--plain`, and `--no-input` always bypass even when explicitly set in a real TTY.
- Cancellation before the final submit transition performs no save, create, approval, or execute.

## Shared interaction model

### Modes and keys

| Mode | Entry | Keys | `Esc` | `q` |
|---|---|---|---|---|
| Normal | Surface/step start or completed edit | arrows, `j/k`, `gg/G`, `ctrl+u/d`, tab focus, Enter activate, `?` help | Previous step or prior surface | Cancel wizard after a labelled confirmation if state would be lost |
| Edit | Activate a text field | printable input and standard text editing | Cancel field edit, restore Normal | Inserts `q` |
| Confirm | Review/save/create/reverse approval | labelled choices; typed challenge where required | Back to Review | Never approves or quits implicitly |
| Help | `?` | Scroll/search current-mode bindings | Close Help only | Close Help only |

`Esc` unwinds exactly one layer. For example: Help -> prior mode; Edit -> Normal; Confirm -> Review;
Review -> previous form step; first Normal step -> cancel decision. It never jumps from a field to
process exit. `Ctrl+C` requests cancellation everywhere, waits for in-flight command cleanup, and
renders a truthful cancelled outcome. A printable key owned by Edit is never intercepted as a
global binding.

Focus order is header/progress (not focusable), primary field/options, secondary field/options,
review/action row, footer help. `Tab`/`Shift+Tab` are universal. `h/l` may switch spatial panes in a
wide review but never triggers a mutation. Mouse support is optional and adds no exclusive action.

### Responsive layouts

| Class | Size | Contract |
|---|---|---|
| Wide | width >=120 | Form/options left, sanitized review/context right; one accent focus |
| Standard | width 80-119 | One form group plus a compact review summary below; at most two regions |
| Compact | width 60-79 and height >=18 | One region at a time, breadcrumb and explicit next/back |
| Guard | width <60 or height <18 | Measured size, `60x18` recommendation, safe flag-path command; no clipped inputs |

The model recomputes layout from `tea.WindowSizeMsg`. Resize cannot reset entered non-secret data,
selection, step, mode, or cancellation state. The guard must never display secret-source values
beyond field name and environment-variable name.

## Credential guidance contract

### Data and safety rules

- Pre-populate valid values supplied by arguments/flags and ask only for missing values.
- Connector definitions/manifests drive connector choice, auth mode, required/optional non-secret
  config fields, defaults, descriptions, and required secret field names.
- Required config is validated before any external credential check. Missing GitHub owner/repo must
  not reach a runtime `internal_error`; a literal documentation placeholder such as `OWNER`,
  `REPO`, `<owner>`, or `<repo>` must produce field-specific validation instead of an opaque 404.
- Secret-source selection stores metadata only:
  - **Environment variable**: collect and display the variable name; never read the value into the
    Bubble Tea/Huh model. The command layer resolves it through the existing `--from-env` path.
  - **Controlled stdin**: collect the one field that existing `--value-stdin` will read. The wizard
    exits without saving and prints a sanitized handoff command for redirection. It never converts
    the interactive TTY into a plaintext secret prompt.
- A review may show secret field names, `from environment GITHUB_TOKEN`, or `from controlled stdin`.
  It must never show whether a value resembles a token, its length, a prefix/suffix, or any raw
  value.
- The submit actions are `Save credential`, `Save and test`, and `Cancel credential setup` when all sources are
  environment-backed or no secret is required. When controlled stdin is selected, replace save
  actions with `Show save command` and `Change source`; no credential is persisted in that session.
- `Save and test` saves through the existing vault flow, then calls the existing credential-test
  service. If the test fails, say explicitly that the credential was saved but validation failed;
  do not silently roll back or remove it.

### Credential state machine

```text
ENTRY
  -> PREFLIGHT supplied fields + gate
  -> NAME (only if missing)
  -> CONNECTOR (only if missing)
  -> SCHEMA / AUTH MODE
  -> NON-SECRET CONFIG (missing/invalid fields only)
  -> SECRET SOURCE METADATA (required fields only; never values)
  -> REVIEW
       -> environment/no-secret: SAVE -> [optional TEST] -> OUTCOME
       -> controlled stdin: HANDOFF COMMAND -> OUTCOME (not saved)
       -> Esc: previous step
       -> Ctrl+C/q from Normal: CANCELLED (not saved)
```

Before `SAVE`, revalidate name uniqueness, schema, source metadata, and environment-variable names.
Read environment values only inside the existing command/application seam and keep them out of UI
messages. Sanitize/redact every returned error before it reaches model state.

### Credential wireframes

Wide/standard interactive review (values are illustrative metadata, never real secrets):

```text
  Add credential                              Step 4 of 5: secret sources
  ------------------------------------------------------------------------------
  Connector   GitHub                         Review
  Name        github-prod                    Connector   github
                                               Repository  octocat/Hello-World
  Token source                                Token       environment: GITHUB_TOKEN
  > Environment variable                     Plaintext values are never shown.
    Controlled stdin

  Variable name
  [ GITHUB_TOKEN                         ]

  NORMAL · secret source     enter select · tab focus · esc back · ? help
```

Review and save:

```text
  Add credential                                      Step 5 of 5: review
  ------------------------------------------------------------------------------
  Name          github-prod
  Connector     github
  Repository    octocat/Hello-World
  Secret source token <- environment GITHUB_TOKEN

  No secret values will be displayed or logged.

  > Save credential     Save and test     Cancel credential setup
  CONFIRM · save credential          enter choose · esc edit · ? help
```

Controlled-stdin outcome:

```text
  Credential not saved
  This secret must enter through controlled stdin.

  Run:
    pm credentials add github-prod --connector github \
      --config owner=octocat --config repo=Hello-World \
      --value-stdin private_key < app-private-key.pem

  The command contains field and file names only; it never contains the secret value.
```

Accessible sequential form:

```text
  Add credential. Step 3 of 5: non-secret configuration.
  Focus: Repository owner.
  Enter the repository owner, or type "back" to return.
```

The accessible transcript never repeats a secret value and never asks the user to type one.

## Connection guidance contract

### Data and eligibility rules

- Pre-populate supplied, valid values and ask only for missing fields.
- Check the proposed name against `ListConnections` before expensive metadata/catalog work and
  again immediately before submission.
- Source and destination choices come from existing credential metadata joined to connector
  capabilities/definitions. Do not offer a credential that cannot serve the selected role.
- Stream choices come from source service/catalog metadata. Loading, empty, partial, and failure
  states are textual. Raw request/response data never enters the view.
- Sync-mode choices are the valid modes derived from source stream and destination service
  metadata, not a hard-coded universal list when the metadata is narrower.
- Cursor and primary-key controls appear only when required by the selected mode. Defaults come
  from stream metadata. A default is labelled `default from <stream>` and remains editable.
- Destination table defaults to the selected stream name and is shown explicitly in Review.
- Source/destination endpoint config overrides are shown only when their metadata declares a
  relevant missing non-secret value.
- Review must state source, destination, stream, sync mode, cursor, primary key(s), destination
  table, and which defaults were accepted. `Create connection` is the only mutation.

### Duplicate-name recovery

When a name exists at entry, review, or final submit:

```text
  Connection "github-to-warehouse" already exists.
  No changes were made.

  > Inspect existing
    Choose a new name
    Cancel connection setup
```

- `Inspect existing` opens a read-only sanitized detail view and returns to the recovery menu.
- `Choose a new name` returns to Name with other valid selections preserved.
- `Cancel connection setup` exits with a truthful `Connection not created` outcome.
- There is no update, remove, replace, merge, `--force`, or implicit overwrite action.

### Connection state machine

```text
ENTRY
  -> PREFLIGHT supplied fields + gate
  -> NAME + early duplicate check
       -> DUPLICATE RECOVERY -> inspect | rename | cancel
  -> SOURCE CREDENTIAL
  -> DESTINATION CREDENTIAL
  -> LOAD SERVICE/CATALOG METADATA
  -> STREAM
  -> SYNC MODE
  -> CONDITIONAL CURSOR / PRIMARY KEY
  -> DESTINATION TABLE / declared endpoint config
  -> REVIEW
  -> FINAL duplicate + validation check
  -> CREATE
  -> OUTCOME
```

Any metadata selection change invalidates downstream derived values that are no longer eligible;
preserve only still-valid values. Cancellation before `CREATE` has no state write. A final
duplicate race returns to duplicate recovery instead of an `internal_error`.

### Connection wireframes

Stream and mode selection:

```text
  Create connection                          Step 4 of 7: stream and sync
  ------------------------------------------------------------------------------
  Source       github:github-prod            Metadata
  Destination  warehouse:warehouse-local    Stream       pull_requests

  Stream                                      Cursor       updated_at (default)
  > pull_requests                             Primary key  node_id (default)
    issues                                    Modes        3 eligible
    comments

  Sync mode
  > incremental_append_deduped
    incremental_append
    full_refresh_overwrite

  NORMAL · stream         j/k or arrows · enter select · esc back · ? help
```

Review:

```text
  Create connection                               Step 7 of 7: review
  ------------------------------------------------------------------------------
  Name          github-to-warehouse
  Source        github:github-prod
  Destination   warehouse:warehouse-local
  Stream        pull_requests
  Sync mode     incremental_append_deduped
  Cursor        updated_at        default from pull_requests
  Primary key   node_id           default from pull_requests
  Table         github_pull_requests

  > Create connection     Back to connection details     Cancel connection setup
  CONFIRM · create connection          enter choose · esc edit · ? help
```

Empty metadata state:

```text
  No eligible source credentials.
  Add one with: pm credentials add <name> --connector <connector>
  Then rerun: pm connections create github-to-warehouse
```

## Guided reverse ETL contract

Bare `pm reverse` is the human-first surface on an eligible dual-TTY. `pm reverse guide` is an
explicit alias into the same state model. Both use the same activation gate as the other wizards;
bypass/non-TTY bare invocations render contextual help and exit 0, while explicit guide bypasses
return actionable guidance to the canonical `reverse plan`, `reverse preview`, and `reverse run`
flag path.

### Stages

1. **Plan**: select an existing connection/destination, source table, write action, mappings, and
   bounded record limit from existing service metadata. Run the existing plan service.
2. **Preview**: show sanitized destination, action, mapped fields, record count, and bounded sample.
3. **Approve**: state target/count/impact and require the existing typed challenge when policy says
   so. Carry the approval token ephemerally in memory, never in model display data.
4. **Execute**: call the existing run seam once after approval. Show progress through existing
   events when available; final frame contains run ID/status command, never token.

```text
  Reverse ETL guide                                 Step 3 of 4: approve
  ------------------------------------------------------------------------------
  Destination  github:github-prod
  Action       create_issue
  Source       issue_candidates
  Records      12
  Impact       Creates 12 external GitHub issues; not automatically reversible.

  Type issue_candidates to approve:
  [                                      ]

  EDIT · approval challenge       enter validate · esc preview · ? help
```

The approval token is forbidden in the terminal view, final frame, accessible transcript, stdout,
stderr, JSON, NDJSON, logs, telemetry, screenshots, golden fixtures, copied command, and help text.
The final safe teaching line is `Next: pm reverse status <run-id> --json`; it must not print a
`reverse run --approve ...` equivalent.

`Ctrl+C` before execution produces `Reverse ETL cancelled — no records were written.` During
execution it requests cancellation through the existing context and waits for a truthful final
state. It never reports cancellation if a mutation actually completed.

## Copywriting and error contract

### Primary and state copy

| Element | Exact copy or pattern |
|---|---|
| Credential primary CTA | `Save credential` |
| Credential test CTA | `Save and test` |
| Controlled-stdin CTA | `Show save command` |
| Connection primary CTA | `Create connection` |
| Reverse approval CTA | `Approve and execute` only after Preview and any typed challenge |
| Credential empty heading | `No connectors are available.` |
| Credential empty body | `Inspect installed connectors with: pm connectors list --json` |
| Connection empty heading | `No eligible source credentials.` or `No eligible destination credentials.` |
| Connection empty body | `Add one with: pm credentials add <name> --connector <connector>` |
| Catalog empty | `No streams are available for <source>. Inspect the connector or choose another source.` |
| Credential saved | `✓ Credential <name> saved for <connector>.` |
| Credential saved/tested | `✓ Credential <name> saved and validated.` |
| Credential partial | `▲ Credential <name> was saved, but validation failed: <redacted reason>.` |
| Connection created | `✓ Connection <name> created.` |
| Cancelled credential | `Credential not saved.` |
| Cancelled connection | `Connection not created.` |
| Reverse complete | `✓ Reverse ETL run <run-id> completed — <count> records.` |

### Error-copy rules

Every error contains, in this order: what happened; affected field/target; whether any state was
written; the exact safe next step. Never begin with `internal_error`, a raw Go type, an HTTP status
alone, `something went wrong`, or an apology.

| Condition | Category / exit | Exact message contract |
|---|---|---|
| Missing positional name in bypass path | `usage` / 2 | `credentials add needs a name — pass pm credentials add <name>, or rerun in an interactive terminal.` (same pattern for `connections create`) |
| Missing required flags/config in bypass path | `validation` / 3 | `Credential <field> is required — pass <exact flag>, or rerun in an interactive terminal.` |
| Missing GitHub owner/repo | `validation` / 3 | `GitHub owner and repository are required — pass --config owner=<owner> --config repo=<repo>.` |
| Literal placeholder | `validation` / 3 | `GitHub <field> cannot be "<placeholder>" — replace the documentation placeholder with the real <field>.` |
| Empty environment source | `validation` / 3 | `Environment variable <ENV> for <field> is empty — set it or choose a different secret source.` |
| Controlled stdin selected in wizard | success, no write | `Credential not saved — run the shown --value-stdin command with redirected input.` |
| Credential name exists | `validation` / 3 | `Credential "<name>" already exists — choose a new name or inspect the existing credential.` |
| Connection name exists | `validation` / 3 | `Connection "<name>" already exists — no changes were made. Inspect it, choose a new name, or cancel.` |
| No eligible credential | `validation` / 3 | `No eligible <source|destination> credential is available — add one with pm credentials add, then retry.` |
| Metadata/catalog load fails | typed connector/runtime category | `<Connector> metadata could not be loaded: <redacted reason>. Retry metadata loading, choose another credential, or use pm connectors inspect <connector> --json.` |
| Cursor required | `validation` / 3 | `Sync mode <mode> needs a cursor — choose one of: <sanitized fields>.` |
| Primary key required | `validation` / 3 | `Sync mode <mode> needs a primary key — choose at least one field.` |
| Save succeeded, test failed | typed connector/runtime category with saved state | `Credential "<name>" was saved, but validation failed: <redacted reason>. Inspect metadata or retry pm credentials test <name>.` |
| Reverse challenge mismatch | `validation` / 3, no execute | `Approval text did not match <challenge> — no records were written.` |
| Reverse cancel before execute | cancellation, no execute | `Reverse ETL cancelled — no records were written.` |

JSON uses the existing single `polymetrics.ai/v1` `Error` envelope with existing `category`,
`code`, and redacted `message`. Do not add a second stdout object, progress text, prompt, or ANSI.
Expected user-correctable missing/invalid/duplicate conditions must be wrapped in the existing usage
or validation error constructors so they never fall through to `category=internal` /
`code=internal_error`.

## Agent-safe invocation profile

The universal agent-safe profile is explicit and composable:

```text
--json --no-input
```

For long-running commands, add:

```text
--progress ndjson
```

NDJSON progress stays sanitized on stderr; the final JSON envelope remains the only stdout object.
Secret material enters only through a named environment source or controlled stdin:

```bash
pm credentials add github-prod \
  --connector github \
  --config owner=octocat \
  --config repo=Hello-World \
  --from-env token=GITHUB_TOKEN \
  --json --no-input

pm connections create github-to-warehouse \
  --source github:github-prod \
  --destination warehouse:warehouse-local \
  --stream pull_requests \
  --sync-mode incremental_append_deduped \
  --cursor updated_at \
  --primary-key node_id \
  --table github_pull_requests \
  --json --no-input
```

`pm query run --agent-mode summary|stream` remains separate: it controls query result shape only.
It does not imply `--json`, `--no-input`, prompt suppression, secret safety, or NDJSON progress, and
must not be documented as a global mode.

## Red -> green -> refactor test matrix

Production edits start only after the RED assertions below fail for the expected missing behavior.
Use named table-driven cases, injected stdin/stdout TTY facts, deterministic clocks/IDs, and no
real credentials or external mutations.

| ID | RED assertion | GREEN contract | REFACTOR guard |
|---|---|---|---|
| GATE-01 | Incomplete action with stdin+stdout TTY does not launch guidance | Launch once and ask only missing fields | Shared gate helper; no per-command drift |
| GATE-02 | Fully specified TTY invocation launches a TUI | Execute directly with byte/semantic parity | Completeness is schema-aware and tested |
| GATE-03 | `stdin-piped+stdout-TTY` prompts or consumes input | Deterministic error; zero reads unless complete explicit `--value-stdin` | Reader spy; no `/dev/tty` |
| GATE-04 | stdout piped activates or emits ANSI | Deterministic plain/error path | Snapshot stdout/stderr separately |
| GATE-05 | `--json`, `--plain`, or `--no-input` prompts | Each bypasses Bubble Tea/Huh and returns deterministic output/error | One table for all three flags |
| GATE-06 | `CI`, `PM_NO_TUI`, or `TERM=dumb` prompts | Each bypasses, including non-empty `CI=0` | Explicit env table |
| GATE-07 | Bare credentials/connections launch UI or fail | Contextual help, exit 0 | Golden help fixtures |
| GATE-07A | Eligible dual-TTY bare reverse does not enter guidance, or differs from `reverse guide` | Bare and alias reach the same guided state model | Paired route/model assertions |
| GATE-07B | Bare reverse bypass/non-TTY initializes UI or fails | Deterministic contextual help, exit 0 | Full bypass matrix plus Bubble Tea init spy |
| GATE-07C | Reverse help flag opens guidance | Focused help, exit 0 | Golden help fixture plus init spy |
| GATE-08 | Accessible mode bypasses gate or redraws | Sequential mode only after dual-TTY gate | Static transcript assertions |
| CRED-01 | Missing schema config reaches connector runtime | Inline/schema validation before save/test | Same validator for flag and TUI paths |
| CRED-02 | `OWNER`/`REPO` placeholder reaches opaque 404 | Field-specific validation error | Placeholder table, no real network |
| CRED-03 | UI model receives a secret value | Model contains field/source metadata only | Secret-marker scan of structs, frames, logs |
| CRED-04 | Controlled stdin choice reads interactive TTY | No save; sanitized handoff command | Reader spy and vault-state assertion |
| CRED-05 | Complete `--value-stdin` reads before non-secret validation | Validate first, then read one declared field | Malformed config + reader spy |
| CRED-06 | `Save and test` test failure looks like total failure or rollback | Explicit saved-but-test-failed outcome | Persisted metadata exists; no secret output |
| CRED-07 | Credential duplicate is `internal_error` or overwrites | Validation error; no state change | Before/after state equality |
| CONN-01 | Wizard offers ineligible credential/stream/mode | Options derive from service metadata | Fake service with narrow capabilities |
| CONN-02 | Cursor/PK/table defaults are guessed or hidden | Defaults derive from metadata and appear in Review | Change selection invalidates stale defaults |
| CONN-03 | Duplicate connection is `internal_error` or overwrite | Recovery menu; validation category; no write | Entry + final-submit race cases |
| CONN-04 | Metadata loading blocks `Update` or leaks raw error | `tea.Cmd` emits typed sanitized event | Cancellation and delayed command test |
| CONN-05 | Create failure claims success or loses edit state | Truthful error with `Retry connection creation`, `Edit connection`, and `Cancel connection setup` | State transition test |
| MODE-01 | `j`, `q`, `/`, `?` are stolen from Edit | Printable input wins while field focused | Key conflict table |
| MODE-02 | `Esc` exits multiple layers | Exactly one-layer unwind | Transition table for every mode/step |
| MODE-03 | Help lists inactive action | Disabled binding absent and inert | `bubbles/help` semantic test |
| VIEW-01 | Resize panics/clips/resets form | Wide/standard/compact/guard preserve state | 160x45, 100x30, 80x24, 60x18, 59x17 frames |
| VIEW-02 | Color/glyph is sole state carrier | Glyph + word; NO_COLOR and ASCII readable | ANSI-stripped transcript comparison |
| SEC-01 | Control chars or secret-like marker reaches View | Sanitize/redact before model/view state | ESC/token marker scan |
| REV-01 | Guide skips plan/preview or executes on cancel | Exact four-stage seam; no execute before approval | Service call-order spy |
| REV-02 | Approval token appears in any output/artifact | Token exists only in ephemeral command seam | Scan stdout/stderr/JSON/NDJSON/frame/log/fixture/command echo |
| REV-03 | Typed challenge mismatch executes | Validation error and zero write calls | Mutation spy |
| REV-04 | `Ctrl+C` lies about outcome or leaks goroutine | Context cancellation plus truthful final state | `go test -race`; leak/cancel test |
| PARITY-01 | TUI and direct path use different service/result | Same service and equivalent exit/result semantics | Paired TTY/plain/JSON cases |
| PARITY-02 | JSON writes prompt/progress to stdout | One deterministic final envelope only | JSON decoder asserts EOF after one object |
| DOC-01 | Help/manual/website/agent examples diverge | All parity checklist items updated | Generator + grep/golden checks |

Refactor only while all focused GREEN tests remain green. Then run focused repeated tests, focused
race tests, full `go test ./...`, `go vet ./...`, `go build ./cmd/pm`, and `make verify`. Headless
frames assert semantic regions rather than animation timestamps. Manual terminal checks cover
dark/light, TrueColor/256/16/no color, ASCII, paste, resize, screen reader, and reduced motion.

## Help, manual, website, completion, and agent-documentation parity

The implementation PR is incomplete until every applicable item is checked or explicitly marked
not applicable with a reason.

### Runtime help and discovery

- [ ] `pm credentials` and `pm connections` print contextual help and exit 0 in TTY and non-TTY
  contexts.
- [ ] Eligible dual-TTY bare `pm reverse` opens the guided workspace; bypass/non-TTY bare
  `pm reverse` prints deterministic contextual help and exits 0.
- [ ] Invalid namespace actions remain usage errors; help behavior does not hide them.
- [ ] `pm help credentials`, `pm help connections`, `pm help reverse`, `pm credentials --help`,
  `pm connections --help`, and `pm reverse --help` describe the prompt gate and bypass flags.
- [ ] `pm credentials add --help` documents missing-field guidance, schema-driven non-secret
  fields, `--from-env`, `--value-stdin`, and the no-plaintext rule.
- [ ] `pm connections create --help` documents metadata-derived choices/defaults, duplicate-name
  recovery, and no overwrite.
- [ ] `pm reverse --help` and `pm reverse guide --help` document the human-first entry, explicit
  alias, four stages, typed confirmation, bypass behavior, and approval-token secrecy.
- [ ] Shell completions/discovery include `reverse guide` and any approved accessible option, but
  do not invent connector/credential values that require secret access.

### CLI manual and generated artifacts

- [ ] Update `docs/cli/credentials.md`, `docs/cli/connections.md`, `docs/cli/reverse.md`, and
  `docs/cli/config.md` where the global gate/agent profile is described.
- [ ] Regenerate the repository's help/manual/golden artifacts; do not hand-edit generated output
  when a source generator exists.
- [ ] Add exact error, bypass, stdout/stderr, and exit-category examples without real secrets.
- [ ] Preserve one-envelope JSON and stderr-only NDJSON progress language.

### Website

- [ ] Update the canonical website source pages/examples that teach credential and connection
  setup and reverse ETL; then regenerate `website/lib/docs.generated.ts` through the existing
  generator.
- [ ] Examples use `connector:credential`, real-looking non-secret placeholders that are clearly
  marked for replacement, and named env variables without values.
- [ ] Website text distinguishes the interactive convenience path from the agent-safe flag path.
- [ ] Website checks/generation leave no stale generated diff.

### Agent documentation

- [ ] Document `--json --no-input` as the universal agent-safe invocation profile.
- [ ] Document `--progress ndjson` for long-running commands and stderr routing.
- [ ] Document secret intake only as named `--from-env` or controlled `--value-stdin`.
- [ ] State explicitly that query `--agent-mode summary|stream` is query-specific and is not a
  replacement for `--json --no-input`.
- [ ] Update applicable `docs/skills/**`, recipes, help examples, and agent contract/golden tests.
- [ ] No example contains a secret, approval token, generic write command, or automatic reverse
  execution.

### Required parity evidence

```bash
pm help credentials
pm credentials
pm credentials add --help
pm help connections
pm connections
pm connections create --help
pm help reverse
pm reverse
pm reverse --help
pm reverse guide --help
rg -n "credentials add|connections create|reverse guide|--no-input|--agent-mode" docs/cli website docs/skills
```

Run these against a temporary project using fixture-only metadata. Do not use a real credential and
do not execute an external reverse ETL mutation for documentation verification.

## Accessibility and transcript contract

- Explicit accessible mode is static/sequential and activates only after the same dual-TTY gate.
- Announce `Step N of M`, mode, focus, validation status, option count, selected option, and outcome
  in text.
- Every focusable action is reachable with basic keyboard input; arrows and Vim navigation both
  work, with Tab as the universal fallback.
- Reduced-motion/accessibility mode replaces spinners with bounded text status lines.
- Review content follows logical reading order without relying on pane position.
- Error recovery sits immediately after the error and includes a plain command alternative.
- No secret value or approval token appears in accessible prompts or transcripts.
- Screen-reader mode must not be inferred solely from `NO_COLOR`; use the existing explicit
  accessibility configuration/env contract from the terminal design system.

## Security and privacy checklist

- [ ] UI data types cannot carry plaintext secret values or approval tokens into renderable fields.
- [ ] Dynamic connector names, descriptions, config, catalog values, errors, and IDs are sanitized
  before styling.
- [ ] Environment variable names and secret field names are validated identifiers before display
  or use.
- [ ] `--value-stdin` consumes only after all other validation and only on a complete explicit
  direct invocation.
- [ ] Review, command echoes, telemetry, logs, screenshots, tests, and fixtures contain metadata
  only.
- [ ] Duplicate names fail closed without overwrite.
- [ ] Reverse call order is plan -> preview -> approval/challenge -> execute, with zero alternate
  write seam.
- [ ] Cancellation and metadata I/O use context; no unmanaged goroutine or blocking model update.
- [ ] No dependency, generic write tool, credentialed connector check, or production external
  mutation is introduced by this phase.

## Sources and decisions used

### Repository and issue sources

- `AGENTS.md` — GSD, skill routing, secret handling, reverse workflow, and CLI parity rules.
- `.planning/ROADMAP.md` — Phase 18/#416 position and #409/#462 dependencies.
- [Issue #416](https://github.com/polymetrics-ai/cli/issues/416) — objective, acceptance criteria,
  phase gate, and no-interactive-secret scope.
- `docs/design/tui-ux-design.md` — two-door gate, palette, layout, key model, reverse guide, and
  credentials/connections direction.
- `docs/design/terminal-ui-research-and-design-system.md` — operator workspace, responsive classes,
  accessibility, and verification matrix.
- `docs/adr/0003-interactive-tui-layer.md` — accepted Charm v2 architecture, import boundary,
  prompt gate, and token secrecy.
- `.agents/skills/bubble-tea-tui-design/` — mandatory implementation, layout, testing,
  accessibility, and inspiration contracts.
- `.planning/phases/462-terminal-ui-design-research/` — #462 design-gate corrections, especially
  dual-TTY activation and prompt-bypass evidence.
- `internal/ui/detect.go` and `internal/ui/styles/styles.go` — current gate/style seams. The current
  detection source still exposes stdout-only facts; Phase 18 tests must require the accepted
  dual-TTY contract before UI activation.
- `internal/cli/credentials_cli.go`, `internal/cli/cli.go`, `internal/cli/cobra_router.go`,
  `internal/cli/errors.go`, and `internal/app/app.go` — existing flags, vault intake, connection
  defaults/validation, duplicate errors, and error taxonomy.

### Verified primary references (accessed 2026-07-20)

- [Bubble Tea v2 upgrade guide](https://github.com/charmbracelet/bubbletea/blob/main/UPGRADE_GUIDE_V2.md)
  — v2 uses `tea.View`, `tea.KeyPressMsg`, declarative terminal options, and explicit commands for
  asynchronous work. Phase 18 follows those APIs without allowing Bubble Tea's input behavior to
  override the product's dual-TTY gate.
- [Huh v2 official repository](https://github.com/charmbracelet/huh) — forms consist of groups and
  fields, dynamic properties use `Func` forms, and `WithAccessible(true)` replaces the TUI with
  standard prompts. Phase 18 uses dynamic metadata-backed choices and the explicitly gated
  accessible sequential mode.
- [Bubbles v2 official repository](https://github.com/charmbracelet/bubbles) — shared primitives
  cover help, keymaps, lists, inputs, and viewports. Phase 18 reuses them instead of building a
  competing component system.
- [GitHub CLI `gh repo create`](https://cli.github.com/manual/gh_repo_create) — incomplete
  interactive invocation prompts while fully specified flags run non-interactively. Phase 18
  adopts the same progressive-enhancement shape.
- [GitHub CLI environment contract](https://cli.github.com/manual/gh_help_environment) —
  `GH_PROMPT_DISABLED` disables prompts, while accessible prompter/colors and spinner settings are
  explicit. This supports separate prompt-disable, accessible, color, and motion controls.
- [GitHub CLI accessibility guide](https://accessibility.github.com/documentation/guide/cli/) —
  numbered static prompts, 4-bit accessible colors, and text progress are screen-reader/motion
  patterns adopted by the Polymetrics design system.
- [Command Line Interface Guidelines: interactivity](https://clig.dev/#interactivity) — prompts
  must never be required; non-TTY stdin and `--no-input` must fail with the flag needed, and users
  must have a clear escape. Phase 18 strengthens this with a dual-TTY gate.
- [Pulumi CLI global options](https://www.pulumi.com/docs/iac/cli/commands/) —
  `--non-interactive` is a global prompt-disable contract.
- [Pulumi structured stdout pattern](https://www.pulumi.com/docs/iac/cli/direct-resource-operations/)
  — structured JSON stays on stdout while progress/prompts use stderr. Phase 18 retains the same
  machine-readable separation through `--json --no-input` and stderr-only NDJSON progress.

## Checker sign-off

- [x] Dimension 1 Copywriting: PASS — human-first entries, aliases, help-only routes, bypass copy,
      reverse stages, and setup recovery actions are explicit and non-generic.
- [x] Dimension 2 Visuals: PASS — workspace hierarchy, focus/modes, responsive frames, states, and
      semantic terminal wireframes remain implementation-ready for both entry routes.
- [x] Dimension 3 Color: PASS — semantic tokens degrade through 4-bit/no-color and pair every
      state with glyph plus text; the entry revision introduces no color-only meaning.
- [x] Dimension 4 Typography: PASS — terminal monospace hierarchy and text labels remain explicit;
      no font download or unsupported typographic dependency is introduced.
- [x] Dimension 5 Spacing: PASS — the terminal-cell spacing scale and wide/standard/compact/guard
      layout rules cover the revised workspace entries without adding a competing layout.
- [x] Dimension 6 Registry Safety: PASS — not applicable; terminal UI uses the approved local
      design system and this revision adds no registry or dependency.

**Approval:** UI-SPEC VERIFIED by inline GSD checker pass on 2026-07-20. The runtime skill normally
spawns a checker, but the active Codex policy permits subagents only when the user asks for them;
the same six-dimension gate was therefore applied inline and recorded here.
