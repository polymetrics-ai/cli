# Connector Inspect Experience — Research and Proposed Architecture

**Status:** Research complete; planning only. No production implementation or GitHub issue edit is
authorized yet.

**Recorded:** 2026-07-20

**Applies to:** CLI Architecture v2 issues #411, #412, and the later help-parity issue #417.

## Executive decision

`pm connectors inspect <name>` should stop being synonymous with “print the entire generated
manual.” Inspection and reference reading are different jobs.

Use three progressive layers built from one canonical connector presentation model:

1. **Overview:** `pm connectors inspect <name>` prints a useful one-screen summary and safe next
   commands.
2. **Focused inspection:** section and item selectors reveal one subject at a time, such as
   authentication, streams, one stream, write actions, one action, or security.
3. **Full reference:** `--full`, `pm connectors man <name>`, the #411 browser detail view, and the
   #412 docs pager expose the complete generated manual with search and paging.

The Bubble Tea plan partially covers layer 3 through the #411 filter/list/manual-preview browser and
the #412 Glamour/viewport pager. It does not currently define layers 1 and 2, nor require the direct
`inspect` command to become concise.

## Repository evidence

At PR #466 head `26f98a72419010b961b5b8378ef4a695b0c0a06f`:

- `pm connectors inspect github` is 1,583 lines and approximately 150 KB.
- GitHub expands 37 ETL streams and 231 reverse-ETL actions, including every stream field and every
  action's endpoint, required fields, optional fields, and risk.
- Across the generated connector manuals, the median is 81 lines, the mean is approximately 146,
  and the largest is 3,039 lines. This is therefore a connector-wide information-architecture
  problem, not a GitHub-only exception.
- `newConnectorsInspectCobraCommand` sends `inspect`, `man`, and `docs` through the same
  `RenderConnectorManual` function.
- `RenderGuideManual` eagerly renders every `ConnectorGuide.Sections` line. The underlying data is
  already structured (`ConnectorGuide`, `GuideSection`, and `Manifest`), so progressive views do not
  require parsing the rendered text or changing connector bundles.

## What the existing plans cover

### Issue #411 — connector browser

Already planned:

- explicit `pm connectors browse` Bubble Tea surface;
- fzf-style fuzzy connector filtering;
- list + connector manual preview;
- viewport scrolling and full-screen promotion;
- Vim and arrow navigation;
- plain/JSON parity and non-interactive CI/piped paths.

Gap: the preview is currently specified as the complete `RenderConnectorManual` output. Selecting
GitHub would therefore place a 1,583-line document in a viewport without first helping the user
understand it.

### Issue #412 — terminal docs viewer

Already planned:

- searchable Glamour-rendered command and connector manuals;
- light/dark rendering and responsive wrapping;
- viewport paging;
- plain piped output identical to the canonical manual source.

This is the right home for the **full reference reader**, but a beautiful pager does not by itself
make the first 24 terminal lines meaningful.

### Issue #417 — help/manual tree

This later phase owns focused command help, generated man pages, goldens, docs, and website parity.
It should consume the same connector section model, but it should not delay the connector-specific
overview needed by #411.

## Research-derived interaction principles

1. **Progressive disclosure:** default output answers the common questions; explicit selectors
   expose details; `--full` preserves the exhaustive reference.
2. **Search before scroll:** large collections need filter/list/preview rather than one linear wall.
3. **Structure before decoration:** counts, sections, short summaries, and next commands matter more
   than borders or gradients.
4. **TTY and flag doors remain peers:** Bubble Tea is a projection. Plain and JSON remain usable
   without a terminal.
5. **One canonical data model:** plain overview, focused output, TUI panes, Markdown, generated docs,
   and website pages must not each infer their own connector semantics.
6. **No hidden data loss:** changing the default human view is acceptable only when `--full`, focused
   selectors, and the unchanged JSON form make every existing fact discoverable.

## Recommended command contract

### Concise default

```bash
pm connectors inspect github
```

Print approximately one terminal screen:

```text
GitHub  github
Read and write GitHub repository, issue, pull-request, release, Actions,
security, webhook, environment, and ruleset data.

Type          api
Capabilities  check yes  catalog yes  read yes  write yes  query no
Auth modes    public, token, github_app
Configuration 13 fields · 3 secret
Data          37 streams · 231 write actions
Sync          full refresh, incremental

Streams
  repository, issues, pull_requests, branches, commits, +32 more

Write actions
  create_issue, update_issue, comment_issue, close_issue, +227 more
  Safety: external writes require plan -> preview -> approval -> execute.

Explore
  pm connectors inspect github --section auth
  pm connectors inspect github --section streams
  pm connectors inspect github --stream issues
  pm connectors inspect github --section actions
  pm connectors inspect github --action create_issue
  pm connectors inspect github --full
  pm connectors inspect github --json
```

The exact counts and sample names come from the manifest. Do not invent “popular” or “recommended”
items unless future connector metadata explicitly declares them.

### Focused sections

Recommended selectors:

```bash
pm connectors inspect github --section overview
pm connectors inspect github --section auth
pm connectors inspect github --section config
pm connectors inspect github --section streams
pm connectors inspect github --section actions
pm connectors inspect github --section sync
pm connectors inspect github --section security
pm connectors inspect github --section examples
pm connectors inspect github --section links
pm connectors inspect github --section metadata
```

Use stable lowercase selector names and completion. Accept clear aliases such as
`authentication -> auth`, `configuration -> config`, and `writes -> actions`, but render one
canonical spelling in help.

### Focused items

```bash
pm connectors inspect github --stream issues
pm connectors inspect github --action create_issue
```

One stream view should show its description, primary key, cursor, fields, sync compatibility, and a
safe next read/catalog command. One action view should show its description, endpoint metadata,
required/optional fields, risk, confirmation semantics, and the mandatory reverse-ETL workflow.

Selectors must validate against manifest names, offer spelling suggestions/completion, and return a
usage or validation error without reading credentials or contacting the connector.

### Full reference

```bash
pm connectors inspect github --full
pm connectors man github
pm docs view github
```

- `--full` preserves the current exhaustive plain manual for explicit use and pipes.
- `connectors man` should mean full connector reference rather than silently duplicating the concise
  inspector.
- `pm docs view github` is the styled/searchable #412 pager over the same full content.
- Automatic external paging, if retained, must occur only on an eligible TTY and respect the
  repository's safe pager boundary. Piped output must never invoke a pager.

### JSON compatibility

`pm connectors inspect <name> --json` must retain its existing full structured envelope, field
names, values, ordering guarantees, stdout/stderr behavior, and exit status. Scripts should never
have to scrape the human overview.

Filtered JSON is not needed to solve the human UX problem. If added later, it requires an explicit
typed schema and must not silently change the existing unfiltered envelope.

## Bubble Tea connector inspector

The #411 connector browser should preview the concise overview first, not the full manual.

### Wide layout (120+ columns)

```text
Connectors / github                              NORMAL · overview
┌─ Sections ───────────┐  GitHub  github
│ > Overview          │  API connector
│   Authentication  3 │
│   Configuration  13 │  ✓ check   ✓ catalog   ✓ read   ✓ write   – query
│   Streams         37 │  37 streams · 231 write actions
│   Write actions  231 │
│   Sync modes       5 │  Auth: public, token, github_app
│   Security           │  Writes require plan -> preview -> approval -> execute
│   Examples           │
│   Links              │  Streams: repository, issues, pull_requests, +34 more
└──────────────────────┘  Actions: create_issue, update_issue, +229 more

/ search · j/k move · h/l section · enter open · tab focus · ? help · q quit
```

Behavior:

- left pane: section index with truthful counts;
- right pane: selected overview/section/item in a `viewport`;
- `/`: search within the active collection or document;
- `n/N`: next/previous match;
- `j/k`, arrows, `gg/G`, `ctrl+u/ctrl+d`: navigation;
- `h/l` or `tab/shift+tab`: pane/section movement;
- `enter`: open the selected stream/action or promote a document to the full pager;
- `esc`: unwind search/item/full-document one layer at a time;
- `?`: contextual key help;
- `q`: quit only in Normal mode.

### Standard layout (80–119)

Use two panes only when both remain legible. Otherwise show the section list above a focused detail
viewport. Keep the title, current section, result count, mode, and footer visible.

### Compact layout (60–79)

Show one pane at a time with a breadcrumb such as `github / streams / issues`; `tab` or `h/l`
switches between section index and content. Hide optional columns before truncating names.

### Below 60x18

Render the measured size, the recommended size, and these working fallbacks:

```text
pm connectors inspect github --section overview --plain
pm connectors inspect github --full | less
```

### Accessible path

- `--accessible`, `PM_ACCESSIBLE_PROMPTER`, screen-reader mode, `--plain`, CI, and pipes use a static,
  sequential hierarchy rather than redraw-based panes.
- Announce connector, section, count, selection, and mode in text.
- Pair color with a glyph and word; support `NO_COLOR`, 4-bit colors, and ASCII.
- Mouse, OSC52, and terminal hyperlinks are optional accelerators only.

## Presentation architecture

Introduce a connector presentation layer derived from `Manifest`/`ConnectorGuide`, conceptually:

```text
ConnectorPresentation
  identity and summary
  capability facts
  auth mode summaries
  configuration summaries
  stream summaries -> stream detail
  action summaries -> action detail
  sync/pagination/security
  examples and links
  metadata/provenance
```

Requirements:

- derive it once at the CLI boundary from the existing structured connector types;
- do not parse `RenderConnectorManual` text;
- keep business packages independent of `internal/ui`;
- let plain, Bubble Tea, Glamour/Markdown, generated manuals, and website adapters render the same
  presentation nodes;
- keep source ordering deterministic and copy slices/maps defensively where a view model outlives its
  input;
- sanitize and redact every dynamic display string before styling;
- perform no connector I/O, credential resolution, filesystem mutation, or network access;
- do not add charts: connector metadata is categorical/reference information, and charts would add
  decoration rather than insight.

## RED -> GREEN -> REFACTOR plan

### RED — behavior contracts

1. Default human `inspect` contains identity, summary, capabilities, auth/config/stream/action counts,
   safety, and next commands, and stays within a defined semantic/line budget for large fixtures.
2. Default output does not contain the complete field/action expansion.
3. `--full` contains all currently required manual sections and representative late-document items.
4. Every section selector returns only the requested section plus minimal identity/context.
5. `--stream` and `--action` select exact manifest items; unknown names fail safely with suggestions.
6. `--json` without selectors remains byte/schema compatible.
7. Inspection never reads credentials or starts connector/network/runtime work.
8. Control characters and secret-like values are sanitized/redacted.
9. Test generic fixtures: no-auth read-only connector, authenticated read/write API, query connector,
   local file connector, zero-stream connector, hundreds-of-actions connector, and incomplete optional
   metadata.
10. Headless TUI tests cover 160x45, 100x30, 80x24, compact, and size-guard layouts; Normal/Filter/
    Help key ownership; focus; search; resize; empty/loading/error states; and plain fallback.

### GREEN — smallest useful slices

1. Build the read-only presentation model and overview renderer.
2. Add `--full`, `--section`, `--stream`, and `--action` using the same model.
3. Point #411's preview at the overview/section model and promote full documents to a viewport.
4. Let #412 render the full reference with Glamour and search.

### REFACTOR — remove drift

1. Make existing full manual generation an adapter over the same presentation hierarchy.
2. Reuse completion/discovery metadata for section, stream, and action selectors.
3. Update runtime help, `docs/cli/**`, connector generated manuals, website reference, goldens, and
   tests together.
4. Document that `inspect` is concise, `man`/`--full` is exhaustive, `browse` is interactive, and
   `--json` is the stable machine contract.

## Issue ownership recommendation

- **#411 owns** the connector presentation model, concise default inspector, focused selectors, and
  Bubble Tea browser/preview because these must be designed and tested together.
- If #411 becomes too large, create one child issue, `feat(ui): add progressive connector inspector`,
  and make it land immediately before or as the first connector slice of #411. Do not create several
  workers that concurrently edit the connector command and presentation files.
- **#412 owns** the full Glamour/viewport docs reader consuming that model.
- **#417 owns** final focused-help/man-page/docs/website consolidation after command surfaces settle.
- #411's production UI gate remains #462 plus #409 integration. The design commits have landed in the
  parent branch, but live issue state and parent-orchestrator readiness must be reconciled before a Pi
  worker starts.

## Acceptance criteria to add when authorized

- [ ] Default `pm connectors inspect <name>` is a concise, useful overview for both small and very
      large connectors.
- [ ] `--section`, `--stream`, and `--action` provide focused discovery with completion and safe
      validation.
- [ ] `--full` and `connectors man` preserve the exhaustive reference.
- [ ] #411 starts connector preview on the overview and provides searchable section/item drill-down.
- [ ] #412 provides the searchable, wrapped, light/dark full manual pager.
- [ ] Plain/CI/piped/accessibility paths are static and ANSI-free; JSON is unchanged.
- [ ] No inspection path reads credentials, invokes connectors, or exposes secret values.
- [ ] All connectors render safely, including empty, incomplete, and extremely large manifests.

