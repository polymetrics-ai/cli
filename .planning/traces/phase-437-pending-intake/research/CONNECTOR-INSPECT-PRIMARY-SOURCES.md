# Connector Inspect UX — Primary Sources and Evidence

**Recorded:** 2026-07-20

Only primary project documentation/repositories and the local accepted design contracts were used
for implementation-facing conclusions.

## External primary sources

### Command Line Interface Guidelines

- Source: <https://clig.dev/>
- Relevant guidance: distinguish TTY human output from machine output; keep successful output brief;
  retain `--plain`/`--json` forms; do not emit color/animation when non-interactive; use a pager for
  large text only on an interactive terminal; suggest useful next commands.
- Applied decision: make the default inspector concise, preserve stable JSON, keep a static plain
  form, and reserve paging for explicit/full long-reference use.

### Bubbles

- Source: <https://github.com/charmbracelet/bubbles>
- Relevant components: `list` includes fuzzy filtering, pagination, help, and status messages;
  `viewport` provides scrollable content and pager bindings; `help` and `key` keep displayed bindings
  aligned with behavior.
- Applied decision: use a section/item list plus viewport rather than scrolling one pre-rendered
  150-KB document from its first line.

### fzf

- Source: <https://github.com/junegunn/fzf/blob/master/README.md>
- Relevant pattern: query/list/preview, adaptive preview placement, wrapping/visibility controls,
  match navigation, and efficient keyboard-first traversal.
- Applied decision: filtering and preview are the primary interaction vocabulary. Polymetrics uses an
  internal sanitized renderer, not fzf's shell-backed preview command.

### Glamour

- Source: <https://github.com/charmbracelet/glamour>
- Relevant capability: deterministic terminal Markdown rendering, explicit word-wrap width, styles,
  and Lip Gloss color downsampling.
- Applied decision: use Glamour for the explicit full-reference reader under #412, not as a reason to
  dump the entire manual by default.

### kubectl explain

- Source: <https://kubernetes.io/docs/reference/kubectl/generated/kubectl_explain/>
- Relevant pattern: the default explains a resource, a path targets a particular nested field, and
  `--recursive` explicitly opts into the larger expansion.
- Applied decision: default overview + focused stream/action selectors + explicit `--full` is the
  connector equivalent of overview/path/recursive disclosure.

### AWS CLI output and paging

- Sources:
  - <https://docs.aws.amazon.com/cli/latest/userguide/cli-usage-output.html>
  - <https://docs.aws.amazon.com/cli/latest/userguide/cli-usage-output-format.html>
  - <https://docs.aws.amazon.com/cli/latest/userguide/cli-usage-pagination.html>
- Relevant pattern: human table output, structured formats, filtering, bounded item display, and a
  configurable client-side pager that can be disabled.
- Applied decision: presentation format, filtering, and paging are separate concerns. A pager is a
  safety net for long output, not a substitute for useful summary and selection.

### GitHub CLI accessibility guidance

- Source: <https://accessibility.github.com/documentation/guide/cli/>
- Relevant guidance: redraw-heavy selectors need a static numbered/sequential alternative; accessible
  colors use terminal-adaptive 4-bit palettes; motion/spinners must be disableable.
- Applied decision: connector sections and items need a static accessible/plain traversal, textual
  mode/focus/count announcements, and no color-only meaning.

## Local authoritative sources

- `docs/design/tui-ux-design.md` section 2.5: #411 connector browser, list custom delegate, fuzzy
  filtering, manual preview, viewport, full-screen promotion, accessible/plain fallback, and unchanged
  connector JSON.
- `docs/design/tui-ux-design.md` section 2.6: #412 Glamour/viewport docs pager and piped plain output.
- `docs/design/terminal-ui-research-and-design-system.md`: LazyGit operator-workspace hierarchy,
  fzf filter/list/preview loop, responsive layout classes, Normal/Filter/Edit modes, accessibility,
  and test matrix.
- `docs/adr/0003-interactive-tui-layer.md`: affirmative TTY gate, plain/JSON contract, import direction,
  dependency decisions, and accessibility launch requirements.
- `internal/connectors/guide.go`: canonical `ConnectorGuide`/`GuideSection` model and eager full-manual
  renderer.
- `internal/connectors/manifest.go`: structured auth, config, stream, action, sync, pagination, and risk
  data available to a universal connector presentation model.
- `internal/cli/connectors_cli.go`: current `inspect`, `man`, and `docs` actions all call
  `RenderConnectorManual` directly.

## Rejected approaches

- **Only add color/borders:** decorates the wall of text without changing its information hierarchy.
- **Only pipe to `less`:** makes scrolling possible but leaves common answers buried.
- **Only build the #411 browser:** direct `inspect`, pipes, accessibility mode, and users who already
  know the connector name remain underserved.
- **Remove detailed content:** breaks offline reference and discoverability.
- **Parse the rendered manual into sections:** fragile and unnecessary because the source model is
  already typed.
- **Add charts:** streams/actions are categorical reference data; charts would be decorative and less
  precise than counts, lists, and focused detail.
- **Use a shell-backed preview:** violates the repository's safe internal-preview rule and expands the
  command-injection surface.

