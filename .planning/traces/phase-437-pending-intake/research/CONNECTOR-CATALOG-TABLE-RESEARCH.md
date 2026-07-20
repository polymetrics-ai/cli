# Connector Catalog Table Research

**Status:** Research and acceptance design recorded; implementation is not authorized.

**Recorded:** 2026-07-20

**Applies to:** `pm connectors list --all`, `pm connectors catalog`, and the issue #411 connector
browser.

## Finding

The earlier pending table plan mentioned `list --all`, but only required reuse of a shared renderer.
That was incomplete. The complete catalog has information that the ordinary runtime list does not
currently communicate: release stage, stream count, and write-action count.

At PR #466 head `26f98a72419010b961b5b8378ef4a695b0c0a06f`:

- `pm connectors list --all` prints 551 rows and approximately 26 KB;
- it has no header or result summary;
- each row prints only connector name, integration type, and `read=true write=true query=false`;
- the underlying `connectors.Definition` already includes display name, description, release stage,
  all five capabilities, streams, write actions, risk, docs, and icon metadata;
- default `pm connectors list --json` and `pm connectors list --all --json` currently both contain
  551 connectors, although the `--all` form returns the richer `ConnectorCatalog` definition shape;
- the catalog currently contains 486 GA, 25 beta, 36 alpha, and 4 unspecified-stage connectors;
- stream counts range from 0 to 284 and write-action counts from 0 to 569, so counts materially help
  users distinguish simple and very broad connectors.

## External design evidence

- [GitHub CLI formatting](https://cli.github.com/manual/gh_help_formatting) uses vertically aligned
  line-based tables for default human output while keeping explicit JSON fields, jq, and templates
  for machine/custom output.
- [AWS CLI output formats](https://docs.aws.amazon.com/cli/latest/userguide/cli-usage-output-format.html)
  distinguish human table output from tab-delimited text and structured JSON; the documented table
  examples preselect meaningful columns instead of dumping whole objects.
- [Command Line Interface Guidelines](https://clig.dev/) recommend TTY-aware human output, stable
  plain/JSON forms, one record per line for composability, intentional color, and terminal-width
  awareness for tables.
- The local `bubble-tea-tui-design` contract requires a TTY-only interactive projection with
  deterministic plain/JSON siblings, semantic glyph+word status, responsive layouts, and no ANSI in
  pipes or CI.

## Recommended table

The complete catalog should show decision-making fields rather than internal metadata:

```text
CONNECTOR                 TYPE      STAGE  READ  WRITE  QUERY  STREAMS  ACTIONS
100ms                     api       ga     yes   yes    no           8        7
7shifts                   api       ga     yes   yes    no          65       76
activecampaign            api       ga     yes   no     no          11        0
acuity-scheduling         api       ga     yes   yes    no           8        5
adjust                    api       ga     yes   no     no           1        0
adobe-commerce-magento    api       ga     yes   yes    no          10        4
```

Column rules:

- `CONNECTOR`: exact stable slug accepted by other commands; left aligned and never silently changed
  to display name.
- `TYPE`: integration type; left aligned.
- `STAGE`: `ga`, `beta`, `alpha`, or `-` when unspecified. Do not invent a release stage.
- `READ`, `WRITE`, `QUERY`: portable `yes`/`no`; color may reinforce but never replace the words.
- `STREAMS`, `ACTIONS`: exact counts, right aligned.
- `CHECK` and `CATALOG` remain useful in the runtime `connectors list` table but are low-value,
  near-uniform columns in the full catalog. They may appear in a wide interactive detail pane rather
  than displacing stage/count fields from the catalog table.
- Description, docs URL, risk text, icon provenance, schemas, stream names, and action names belong in
  `inspect`/browser detail, not this 551-row overview.

## Command relationships

### `pm connectors list`

Runtime-oriented inventory with the five runtime capabilities:

```text
CONNECTOR  TYPE  CHECK  CATALOG  READ  WRITE  QUERY
```

### `pm connectors list --all`

Convenience route to the complete catalog table:

```text
CONNECTOR  TYPE  STAGE  READ  WRITE  QUERY  STREAMS  ACTIONS
```

### `pm connectors catalog`

The authoritative full-catalog command. With no filters, its human rows and ordering should equal
`connectors list --all`. Its `--capability` and `--stage` flags filter the same table rather than
switching to another format.

Help must explain that `list` is runtime-oriented while `list --all`/`catalog` exposes catalog
metadata. Because both currently contain 551 names, tests and documentation must protect semantic
shape rather than falsely claiming a membership difference that does not exist today.

## Output modes

- **TTY human:** aligned table; optional restrained header summary such as
  `551 connectors · 486 ga · 25 beta · 36 alpha · 4 unspecified`. No heavy box around 551 rows.
- **Plain/piped/CI:** deterministic header and one connector per line, no ANSI, wrapping, animation,
  or paging side effect.
- **JSON:** existing `ConnectorCatalog` envelope and complete definitions unchanged.
- **Interactive:** `pm connectors browse` shows the same catalog with fuzzy name/description search,
  stage/capability filters, counts, and the progressive inspect preview. It is a separate explicit
  command; `list --all` does not unexpectedly enter an alt screen.

Do not add a new table dependency solely for this. A small shared renderer using existing approved
facilities or the standard library is sufficient unless issue #411's approved dependency set already
provides the needed primitive.

## Responsive behavior

- Wide: show all recommended columns and optional display name in the TUI detail/list delegate.
- Standard: show the eight canonical table columns above.
- Compact TUI: keep connector, stage, read/write/query, and counts; open type/details in preview.
- Plain/piped: retain the canonical fixed columns and one-row contract rather than changing shape
  based on an unavailable or misleading terminal width.

## RED -> GREEN -> REFACTOR acceptance

### RED

- Require the complete-catalog header, alignment, stage, capability words, stream counts, and action
  counts for representative GA/beta/alpha/unspecified, read-only, writable, query, zero-count, and
  high-count connectors.
- Require numeric columns to be right aligned and connector/type/stage columns left aligned.
- Assert no `read=true`, `write=false`, Go struct dump, ANSI, wrapping, or duplicated rows on the
  piped/plain path.
- Assert deterministic slug ordering.
- Assert unfiltered `connectors list --all` and `connectors catalog` use the same human table rows.
- Protect the existing `ConnectorCatalog` JSON envelope and definition fields.
- Record and document that ordinary list and full catalog currently both have 551 names; do not make
  a brittle test requiring them to differ.

### GREEN

- Introduce one sanitized catalog-row view model and one shared plain table renderer.
- Render `list --all` and `catalog` from the same rows; apply catalog filters before rendering.
- Reuse those rows in the #411 Bubble Tea browser instead of inventing another capability/count
  projection.

### REFACTOR

- Update focused command help, generated CLI docs, website reference, golden transcripts, and shell
  completion/filter documentation together.
- Keep catalog data selection separate from human/plain/JSON rendering.
- Verify terminal, pipe, `--plain`, `--no-input`, CI, `TERM=dumb`, `NO_COLOR`, and JSON paths.

## Ownership

Issue #411 remains the correct owner because it already owns the 551-connector browser and the plain
list cleanup. When authorized, its acceptance criteria should explicitly incorporate this complete-
catalog table contract along with the ordinary `connectors list` table contract.

