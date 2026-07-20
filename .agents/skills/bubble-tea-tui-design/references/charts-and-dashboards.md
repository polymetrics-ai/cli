# Charts and dashboards

## Chart selection

Choose the chart from the analytical question, never from available visual effects.

| Question/data | Terminal chart | Required text companion |
|---|---|---|
| change over ordered time | line/time-series | start, end, min, max, current, unit |
| compact trend in a row/card | sparkline | current and delta |
| compare categories | horizontal bars | exact value per bar, sort order, unit |
| distribution | histogram | bucket ranges/counts, sample count |
| relationship of two numeric fields | scatter | x/y units, count, ranges |
| dense matrix | heatmap | legend scale, min/max, selected-cell value |
| exact records | table, not a chart | headers, pagination, result count |

Do not use pie charts, unlabeled decorative waveforms, 3D effects, color-only heatmaps, or
dual axes. A chart that cannot explain its scale in the available width becomes a table or
summary instead.

## Query chart contract

- Charts consume an already returned, read-only `QuerySQL` result. They do not generate SQL
  and do not introduce write statements or a generic SQL console.
- The user chooses chart type, X field, Y field(s), aggregation when needed, unit, and sort
  from validated result metadata. Never infer a destructive or remote operation.
- Cap points before rendering. Use deterministic sampling/bucketing and disclose it in the
  footer, for example `2,000 rows · 120 plotted · min/max bucketed`.
- Missing/non-numeric/infinite values get a textual count and deterministic handling.
- Preserve the data table as a sibling pane/toggle. `v` switches chart/table; export works
  from the underlying rows, not from screen glyphs. Export is a typed read-only path only: default
  to project-scoped output, resolve/clean/confine the path, reject control characters, traversal,
  broad roots, symlink targets/final-component races, and overwrites by default, require
  confirmation only when both stdin and stdout are TTYs or noninteractive
  `--output <project-relative-path> --force`, echo only sanitized commands, and fail `--no-input`
  or non-TTY stdin without a preapproved path with exact flag guidance.
- A `--plain` or non-TTY path prints the numeric summary/table. JSON emits data or a stable
  chart-spec object only when that schema is explicitly designed and documented.
- Screen-reader/accessibility mode renders the text summary plus ordered values/buckets,
  with no redraw requirement.

## Dashboard composition

Dashboards answer, in order:

1. What is running or selected?
2. Is it healthy/safe?
3. What changed and how much?
4. What can I do next?

Limit the first view to one primary chart or pipeline rail, a few high-value scalar facts,
and a selected-item detail. More charts are reachable by tabs or drill-down. A screen full
of equally bright tiles is not a hierarchy.

For progress/event dashboards:

- throttle visual updates while retaining all lifecycle events;
- show count, rate, elapsed, and last-update time with units;
- freeze a truthful final frame on success, failure, or cancellation;
- use reduced-motion mode with periodic text updates;
- never imply precision beyond the source data.

## Renderer decision

`github.com/NimbleMarkets/ntcharts/v2` is the leading candidate because its current `v2`
branch targets Bubble Tea v2 and supplies time-series, bar, heatmap, scatter, streaming,
and sparkline models. Its maintainers state that the v2 tag indicates Bubble Tea API
compatibility and that the NTCharts API may still change. Therefore:

- evaluate it in an isolated spike;
- pin an exact reviewed version if approved;
- wrap it behind a small `internal/ui/chart` interface;
- test resize, Unicode-width, no-color, ASCII/text fallback, and large-data bounds;
- do not add it to `go.mod` until the issue names it and a human approves the dependency.

A minimal internal sparkline/horizontal-bar renderer is the fallback when the dependency
gate is not approved. Keep the interface small enough to replace either implementation.
