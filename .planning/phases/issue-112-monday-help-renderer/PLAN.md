# Plan — issue #112 Monday help renderer/docs

## Objective

Render Monday connector command help/docs from `cli_surface.json` without overclaiming executable direct reads or writes.

## GSD mode

- `scripts/gsd prompt plan-phase issue-112-monday-help-renderer --skip-research` generated the planning prompt.
- `programming-loop` prompt is unavailable; manual GSD/TDD fallback active.

## Slice

1. Red test: Monday rendered connector manual includes command surface, implemented stream commands, approval notes, and no duplicate command entries from overlapping groups.
2. Green: adjust Monday `cli_surface.json` grouping if needed and update `docs.md` with CLI command surface notes.
3. Verify connector manual output via `pm connectors inspect monday`; `pm help monday` / `pm monday --help` remain not applicable unless the CLI has a dynamic help topic route for connector aliases.

## Safety

No live Monday calls, no credentials, no raw API examples, no executable write/direct-read overclaiming.

## Verification

```bash
go test ./internal/connectors/bundleregistry -run 'TestMondayGuideIncludesCLISurfaceHelp' -count=1
go run ./cmd/pm connectors inspect monday
rg -n "command surface|pm monday|board list" internal/connectors/defs/monday/docs.md
```
