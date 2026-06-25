# Local Verification

Phase: native-go-all-connectors

| Check | Status | Command | Notes |
| --- | --- | --- | --- |
| Format | passed | `make verify` | Runs `gofmt -w cmd internal`. |
| Vet | passed | `make verify` | Runs `go vet ./...`. |
| Tests | passed | `make verify` | Runs `go test ./...`. |
| Build | passed | `make verify` | Runs `go build ./cmd/pm`. |
| Docs validation | passed | `make verify` | Runs `./pm docs validate --connectors-dir docs/connectors`. |
| Smoke | passed | `make verify` | Runs init, credentials, connection, catalog, ETL, reverse ETL, and file assertions. |
| Install | passed | `make install` | Installed updated binary to `/Users/karthiksivadas/.local/bin/pm`. |

## Acceptance Probes

```text
pm connectors list --all --json | jq '{count, enabled: .summary.enabled, planned: .summary.planned_native_port, sources: .summary.sources, destinations: .summary.destinations}'
{
  "count": 647,
  "enabled": 1,
  "planned": 646,
  "sources": 591,
  "destinations": 56
}
```

```text
pm etl check --connector source-stripe --json
```

Returns a non-zero exit with `connector "source-stripe" not found`, because `source-stripe` is a planned native port and is not executable yet.

```text
pm connectors inspect destination-postgres
```

Shows `implementation_status: planned_native_port` and runtime capabilities `check=false`, `catalog=false`, `read=false`, `write=false`, `query=false`, `etl=false`, and `reverse_etl=false`.

```text
pm connectors inspect source-github
```

Shows `implementation_status: enabled`; the registry exposes `source-github` as an alias to the real built-in GitHub connector.
