# TDD Ledger — issue #112 Monday help renderer/docs

Manual GSD programming-loop fallback is active because the adapter registry has no `programming-loop` command.

| Step | Evidence | Result |
| --- | --- | --- |
| Red | `go test ./internal/connectors/bundleregistry -run 'TestMondayGuideIncludesCLISurfaceHelp' -count=1` fails because overlapping Monday groups render `board list` twice. | Captured |
| Green | Adjusted Monday command groups to avoid duplicate prefix rendering and added docs.md CLI command surface notes. `go test ./internal/connectors/bundleregistry -run 'TestMondayGuideIncludesCLISurfaceHelp' -count=1` passes. | Captured |
| Refactor | `go run ./cmd/pm connectors inspect monday` shows command surface; `go run ./cmd/connectorgen validate internal/connectors/defs --json` passes with 547 connectors, 0 findings, 0 warnings. | Captured |
