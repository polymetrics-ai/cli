# Verification: GitLab CLI Surface Metadata (#83)

## Planned Red Command

```bash
go test ./internal/connectors/engine -run TestBundleLoadEmbeddedGitLabCLISurface -count=1
```

## Planned Focused Green Commands

```bash
go test ./internal/connectors/engine -run TestBundleLoadEmbeddedGitLabCLISurface -count=1
go test ./cmd/connectorgen ./internal/connectors/engine -run 'CLISurface|GitLab' -count=1
go run ./cmd/connectorgen validate internal/connectors/defs --json
```

## Planned Broader Gates Before Handoff

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

## Results

Pending. This file must be updated with exact command results after red and green runs.

## CLI Help / Docs / Website Parity

- #83 is metadata-only. Runtime help/manual/website parity is planned for #84 unless this slice changes runtime behavior.
- If generic website bundle generation consumes `cli_surface.json`, run the applicable generation check and record it here.
