# Sources: Front CLI Surface Metadata

## Issues and PRs

- Parent issue #188: https://github.com/polymetrics-ai/cli/issues/188
- Sub-issue #189: https://github.com/polymetrics-ai/cli/issues/189
- Draft parent PR #224: https://github.com/polymetrics-ai/cli/pull/224

## Official Front sources

- Front public docs index: https://dev.frontapp.com/llms.txt
- Front API docs index in current bundle: https://dev.frontapp.com/reference/introduction

## Existing Front bundle files

- `internal/connectors/defs/front/metadata.json`
- `internal/connectors/defs/front/spec.json`
- `internal/connectors/defs/front/streams.json`
- `internal/connectors/defs/front/api_surface.json`
- `internal/connectors/defs/front/docs.md`

## Reference implementation sources

- `internal/connectors/defs/github/cli_surface.json`
- `cmd/connectorgen/main_test.go` (`CLISurface` validation tests)
- `cmd/connectorgen/validate.go` (`checkCLISurface` rules)
- `internal/connectors/engine/bundle.go` (`CLISurface`, `CLICommand`, `CLIFlag` structs)
- `docs/architecture/connector-operation-kernel.md`
- `.planning/phases/github-cli-surface-metadata/**`
