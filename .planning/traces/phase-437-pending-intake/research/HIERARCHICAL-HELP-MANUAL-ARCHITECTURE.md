# Hierarchical CLI Help and Manual Architecture Research

**Status:** Research and planning only; no implementation authorized.

**Recorded:** 2026-07-20

**Intended implementation owner:** Phase 19 / issue #417, `feat(cli): deepen help tree and generate
man pages`

**Repository examined:** PR #466 head `26f98a72419010b961b5b8378ef4a695b0c0a06f`

## Executive conclusion

Every visible command node should have its own focused manual. Parent commands should render their
own manual plus a compact table of immediate children, not concatenate every descendant's full
manual. Full subtree assembly should be an explicit renderer mode used for discovery, offline
reference generation, the documentation viewer, or an explicitly requested recursive view.

The recommended architecture is a hybrid:

1. Cobra remains authoritative for the command tree, canonical path, aliases, usage line, local
   and inherited flags, command groups, and child relationships.
2. A typed Polymetrics `ManualSpec` adds content Cobra cannot model reliably: output and JSON
   contracts, security constraints, credential behavior, approval gates, examples, exit meanings,
   related guides, and ordered extension sections.
3. A `ManualTree` is compiled from the fresh Cobra tree plus those specifications and validated
   before it is rendered.
4. One renderer family produces focused terminal text, JSON envelopes, Markdown pages, Section 1
   man pages, website data, subtree outlines, and combined reference books.
5. Dynamic connector manuals remain metadata providers attached to the help system; 547 connector
   definitions must not become 547 static Cobra subcommands.

This gives one semantic source of truth without forcing rich Polymetrics safety documentation into
raw Cobra templates or duplicating whole manual strings for every output format.

## User requirement captured

- Every root, parent, and leaf command has an individual manual.
- A parent manual remains useful by itself and lists its immediate children.
- A leaf or nested subcommand shows only its own contextual manual.
- Manuals can be assembled by subtree when explicitly needed.
- Generated terminal help, JSON, Markdown, man pages, website reference, and the future TUI docs
  viewer stay in parity.

## Current repository findings

The current system is intentionally namespace-oriented rather than command-tree-oriented:

- `internal/cli/docs.go` has 19 generated manual topics after excluding the two root aliases.
- `pm docs generate` iterates that flat map and writes one fenced Markdown file per topic.
- There are 76 production `setManualHelp` command bindings, while many child commands bind
  back to a parent topic. For example, connectors has six bindings to the single `connectors`
  topic; ETL, reverse, and credentials each have seven.
- `setManualHelp` accepts a free-form topic string rather than resolving the active Cobra command.
- The website CLI reference is maintained separately in MDX, then copied into generated website
  data; it is not derived from the full runtime command tree.
- The golden suite intentionally protects the existing byte output, including legacy cases where a
  child such as `connectors certify` renders the parent manual.

That structure explains both observations already recorded in this pending file:

1. the aggregate connectors manual looks disproportionately GitHub-specific; and
2. `pm connectors certify --help` is byte-identical to the entire connectors manual.

These are symptoms of the flat-topic model, not isolated certify defects.

## External architecture research

### Cobra's native model

Cobra already provides the structural half of the required design:

- every `cobra.Command` has `Use`, `Short`, `Long`, `Example`, aliases, annotations, flags, and a
  parent/child position;
- its default help command resolves a path to any nested command;
- command groups keep large parent help pages scannable;
- `cobra/doc` can walk a command and all descendants to create Markdown or man-page trees;
- generated man pages already include `SEE ALSO` links to the parent and immediate children.

The repository uses Cobra v1.10.2. In that version, `doc.GenManTreeFromOpts` traverses available
descendants, derives filenames from `CommandPath`, and renders local and inherited options
separately. It also exposes `CommandSeparator`.

There are two important constraints:

1. Cobra's built-in document generators primarily understand `Short`, `Long`, `Example`, flags,
   and tree relationships. They do not provide a typed Polymetrics model for security, output
   schemas, approval gates, or exit-code conditions.
2. Cobra documents a filename ambiguity when command names containing hyphens collide with a
   flattened descendant path. Generation therefore needs an explicit collision preflight; website
   and Markdown output should use path directories rather than relying only on flattened names.

### Patterns from mature CLIs

Git uses a focused-manual model: `git help <command>` resolves one command or guide, while root help
and `--all` are discovery indexes. Git does not print every command's full manual in normal root
help.

GitHub CLI follows the command path: `gh help [path to command]` retrieves full details for that
specific node. Its published reference can still assemble the complete tree offline.

Docker's reference distinguishes three useful levels:

- the root documents global behavior and lists management commands;
- a parent such as `docker container` describes the namespace and lists immediate subcommands;
- a leaf such as `docker container exec` owns detailed options and examples.

Kubernetes publishes a generated kubectl reference that exposes the hierarchy, supports expanding
or collapsing the combined reference, and links to focused command pages. This is a good website
analogue for "club the manuals when needed" without overwhelming normal terminal help.

The Linux man-page conventions reinforce individual Section 1 pages with a predictable section
order and `SEE ALSO` cross-references. That is a better Unix experience than one enormous page for
an entire deep CLI tree.

The Command Line Interface Guidelines recommend extensive help when `--help` is explicitly
requested, concise help for incomplete parent invocations, and help support on subcommands. This
supports separate focused and summary rendering modes.

## Approaches considered

| Approach | Strengths | Weaknesses | Verdict |
|---|---|---|---|
| One raw manual string per command path | Minimal change; exact prose control | Repeats usage, flags, children, links, and formatting; easy drift across 70+ nodes and output formats | Reject as final architecture |
| Cobra fields and `cobra/doc` only | Tree, flags, aliases, Markdown/man generation are automatic | Cannot enforce Polymetrics output, security, credential, approval, and exit contracts structurally | Useful base, insufficient alone |
| External Markdown/YAML is the only source | Writer-friendly; can render on website | Splits runtime command truth from docs; flags and paths can drift; adds parsing/embedding complexity | Reject as primary source |
| Typed manual registry independent of Cobra | Rich validation and multi-render support | Duplicates command path, usage, flags, aliases, and child relationships | Better, but still duplicates structure |
| Cobra tree + typed Polymetrics manual specifications | Uses live command structure, supports rich safety contracts, enables all renderers and validation | Requires deliberate migration and a small internal help subsystem | **Recommended** |

## Recommended semantic model

The model below is illustrative architecture, not production code:

```go
type ManualSpec struct {
    ID          string
    Description string
    Output      []Paragraph
    Examples    []Example
    Security    []Paragraph
    ExitStatus  []ExitStatus
    Related     []ManualRef
    Extra       []Section
}

type ManualNode struct {
    Path          []string
    CanonicalPath string
    Aliases       []string
    UseLine       string
    Summary       string
    LocalFlags    []Flag
    Inherited     []Flag
    Children      []*ManualNode
    Spec          ManualSpec
    Kind          NodeKind // command, guide, dynamic
}
```

Do not repeat `Use`, aliases, flag names, inherited flags, or children in `ManualSpec`; derive them
from the live Cobra node. Bind the additional spec to the command with a stable annotation/ID or a
constructor helper, then compile the tree after all commands have been registered.

The command constructor should colocate user-facing metadata with behavior. A helper such as
`newManualCommand`/`bindManual` can populate Cobra's `Short`, `Long`, and `Example` fields and attach
the spec ID. The exact API should be chosen during the issue #417 RED/GREEN cycle.

## Rendering modes: how manuals are "clubbed"

Composition should be explicit and depth-aware:

| Mode | Contents | Intended surface |
|---|---|---|
| `node` | One command's complete focused manual; parent includes immediate child summaries | `pm <path> --help`, `pm help <path>`, bare namespace |
| `tree` | Breadcrumb plus descendant paths and one-line summaries, optionally depth-limited | interactive discovery and `pm help <path> --tree` |
| `subtree` | Full manuals for the selected node and descendants in deterministic preorder | explicit recursive help, pager/docs viewer |
| `all` | Root plus every available command and guide | generated combined reference/search index only |

The recommended terminal behavior is:

```text
pm connectors
  -> connectors parent manual + immediate child summary table

pm connectors certify --help
pm help connectors certify
  -> identical certify-only manual

pm help connectors --tree
  -> concise descendants outline, not every manual body

pm help connectors --recursive
  -> explicit combined subtree, preferably through a pager when interactive
```

`--tree` and `--recursive` are proposed names. Issue #417 may select different names, but it should
preserve the distinction between a compact outline and full recursive bodies.

Normal `--help` must never concatenate all descendants. Full aggregation grows too quickly,
duplicates content, harms scanning, and recreates the exact certify problem at larger scale.

## Manual section contract

Use a stable Section 1-style order, omitting sections that do not apply:

1. `NAME`
2. `SYNOPSIS`
3. `DESCRIPTION`
4. `COMMANDS` for parents (immediate children only)
5. `ARGUMENTS`
6. `OPTIONS`
7. `INHERITED OPTIONS`
8. `OUTPUT` and JSON envelope/report kinds
9. `ENVIRONMENT` and `FILES` when relevant
10. `EXAMPLES`
11. `SECURITY` and mutation/credential safety
12. `EXIT STATUS`
13. `SEE ALSO` (parent, children, related guides)

The standard ordering follows man-page conventions. `OUTPUT` and `SECURITY` are deliberate
Polymetrics additions because agentic JSON contracts, credential boundaries, and approval-gated
writes are product behavior, not optional prose.

## Resolution rules

- `pm help <path...>` resolves a multi-segment canonical command path.
- `pm <path...> --help` resolves the same node and must be byte-identical in the same output mode.
- Aliases resolve to the canonical node; the manual can list the alias but must not fork content.
- Hidden commands are omitted from parent discovery unless explicitly requested and documented.
- Additional conceptual guides such as future `a11y` documentation are registered as `guide`
  nodes, not fake runnable commands.
- Dynamic connector paths use a `ManualProvider` that returns metadata-only manuals. Help and
  inspection must not resolve credentials or perform network calls.
- Bare namespace commands continue to render their own node manual and exit 0. This policy does not
  automatically convert a missing required argument on every runnable leaf into success.
- Invalid actions remain usage errors and must not fall back to parent help.

## Renderer and generator architecture

```text
fresh Cobra command tree
        +
typed manual specifications / dynamic providers
        |
        v
compile + validate ManualTree
        |
        +--> terminal text / pager
        +--> CommandManual or HelpTree JSON
        +--> docs/cli tree (one page per node)
        +--> Section 1 man-page tree
        +--> website navigation + searchable combined reference
        +--> TUI docs viewer
```

Recommended components:

- `model`: typed specs, nodes, sections, flags, exit statuses, examples, and references;
- `registry`: binds stable IDs to additional specifications and dynamic providers;
- `compiler`: traverses a fresh Cobra tree and derives structural metadata;
- `resolver`: exact path, alias, guide, and dynamic-provider resolution;
- `render/text`: stable plain output with no ANSI/control leakage;
- `render/json`: structured envelope plus the canonical rendered manual for compatibility;
- `render/markdown`: individual pages, parent indexes, and combined reference;
- `render/man`: Cobra `doc.GenManTreeFromOpts` or a thin adapter after specs enrich the tree;
- `validate`: full command coverage, links, collisions, safety sections, deterministic order, and
  output drift.

## Generated file layout

Use hierarchical paths for Markdown and website pages:

```text
docs/cli/index.md
docs/cli/connectors/index.md
docs/cli/connectors/list.md
docs/cli/connectors/inspect.md
docs/cli/connectors/certify.md
```

Section 1 man pages remain conventionally flat:

```text
docs/man/pm.1
docs/man/pm-connectors.1
docs/man/pm-connectors-certify.1
```

Before generating man pages, compute every output filename and fail on a collision. Use deterministic
headers: set `DisableAutoGenTag`, provide a stable release/source date, or use `SOURCE_DATE_EPOCH`.
Checked-in docs must not change solely because the generator ran on a different day.

The website should consume generated navigation/search data derived from `ManualTree`. A combined
reference can offer expand/collapse behavior like kubectl, while each command retains a stable
focused URL. Hand-written tutorials remain separate and link to generated reference nodes.

## Cobra `doc` usage recommendation

Use `cobra/doc`, but do not make its default Markdown output the only canonical representation.

- `GenManTreeFromOpts` is a strong backend for recursive Section 1 pages and already creates parent
  and child `SEE ALSO` links.
- Its local/inherited flag separation is useful and should be preserved.
- A preflight must detect flattened filename collisions.
- Rich Polymetrics sections should be rendered into the document input in a controlled manner, or a
  thin custom man adapter should surround Cobra's structural information.
- Custom Markdown/JSON renderers should consume `ManualTree` directly so directory hierarchy,
  frontmatter, website links, and machine-readable safety metadata remain under project control.

This uses the already approved Cobra dependency and does not require a new runtime dependency.

## Validation and test architecture

Issue #417 should make help coverage mechanically complete:

1. Traverse a freshly constructed root and snapshot every available canonical command path.
2. Require exactly one resolved manual spec for every visible node, with explicit exceptions for
   hidden/internal commands.
3. Table-test `pm help <path>` and `pm <path> --help` byte parity for every node.
4. Test bare parent help exits 0 and lists only immediate available children.
5. Test leaf help excludes sibling and parent-only sections.
6. Test tree/subtree traversal ordering, depth, hidden commands, aliases, and guides.
7. Assert help is effect-free before configuration, logger, telemetry, credential loading,
   workspace creation, runners, live services, or external writes/sweeps.
8. Assert local and inherited flags are accurate and not hand-maintained duplicates.
9. Validate all `SEE ALSO` targets and website links.
10. Preflight Markdown/man filenames for collisions and path traversal.
11. Use fixed generation metadata and rerun generators twice to prove deterministic bytes.
12. Diff runtime text, JSON, Markdown, man, website navigation, and golden transcripts in one parity
    gate.
13. Keep control-character sanitation and the no-ANSI agent contract.

The two pending observations from PR #466 should be the first acceptance examples:

- connectors parent help identifies GitHub only as an example; and
- certify help resolves the certify node rather than the connectors parent node.

## Migration strategy

This is centralized help infrastructure and should have one mutating owner. Parallel read-only
inventory/review is safe; parallel writers touching the registry, tree compiler, docs generator,
goldens, or website index would have a high collision rate.

Recommended serial waves inside issue #417:

1. Inventory and RED coverage matrix for every existing command/help form.
2. Add the typed model/compiler/resolver behind the current output as a compatibility seam.
3. Pilot individual parent/child manuals with `connectors` and `connectors certify`.
4. Migrate remaining static namespaces in stable groups while keeping the full inventory gate green.
5. Add dynamic connector/guide providers.
6. Add explicit tree/subtree modes and structured JSON.
7. Generate hierarchical Markdown, deterministic man pages, website data, and combined reference.
8. Remove the flat topic map only after every current topic is represented and parity exemptions are
   reviewed.
9. Run full local gates, code review, generated diff review, and human UX review.

## Risks and mitigations

| Risk | Mitigation |
|---|---|
| 19 pages expand into many generated files | Treat expansion as deliberate Phase 19 churn; review manifest and generated diff together |
| Flags or aliases drift from manuals | Derive them from the live Cobra node, never duplicate them in specs |
| Rich safety text is lost by generic Cobra docs | Keep typed Polymetrics sections and validate required content by command risk class |
| Parent help becomes another wall of text | Include immediate summaries only; make full subtree rendering explicit |
| Help triggers credential/runtime effects | Resolve and render before pre-run/run hooks; use no-effect tests for all nodes |
| Man filenames collide | Precompute filenames and fail generation before writing |
| Generated dates create noisy diffs | Disable auto tags or set deterministic source/release dates |
| Dynamic connector catalog explodes the Cobra tree | Use metadata-only dynamic providers and keep connector docs/catalog separate |
| Website and CLI diverge | Generate navigation/reference data from the same `ManualTree` |
| Huge one-shot migration becomes unreviewable | Use serial TDD waves and coherent commits within the dedicated issue #417 worktree |

## Final recommendation

Approve this as the design basis for issue #417:

- one focused manual per command node;
- one compact parent summary per namespace;
- explicit outline and recursive composition modes;
- Cobra for structure and flags;
- typed Polymetrics specifications for safety/output/exit semantics;
- one compiled manual tree feeding terminal, JSON, Markdown, man, website, and TUI renderers;
- deterministic generation and exhaustive tree coverage tests.

Do not implement it on PR #466. PR #466 should carry only the pending intake/research artifacts
until the user authorizes implementation and the Phase 19 dependency gate is satisfied.
