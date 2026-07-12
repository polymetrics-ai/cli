# Verification — #101 Linear direct read

## Focused evidence

- `go test ./internal/cli -run TestLinearCommandSurfaceRunsStreamBackedDirectRead -count=1`, `go test ./internal/connectors/engine -run TestLinearStreamsUseFixedGraphQLDocuments -count=1`.

## Parent gates also covering this slice

- `go vet ./...` — pass
- `go test ./...` — pass
- `go build ./cmd/pm` — pass
- `./pm docs validate --connectors-dir docs/connectors` — pass
- `npm --prefix website run gen:website-data` — pass
- `make verify` — pass
- `git diff --check` — pass

## Safety

No secrets were requested or printed. No live Linear credentials or network checks were used. Fixed GraphQL documents only; raw arbitrary GraphQL remains disallowed.
