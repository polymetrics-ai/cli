# TDD LEDGER — issue-3062 Notion parity wave03 r1

## Red / baseline evidence

- `go run ./cmd/connectorgen validate internal/connectors/defs/notion` failed before production edits because `connectorgen validate` treated `fixtures/` and `schemas/` as bundle directories when pointed at a single bundle directory.
- `go test ./internal/connectors/conformance -run 'TestConformance/notion' -count=1` passed on the baseline, but dynamic stream coverage was skipped by Notion metadata, so this was insufficient for expanded parity.
- Baseline Notion bundle had partial read-only coverage: `databases`, `pages`, `users`; no `writes.json`; stale docs advertised read-only support.

## Red/green ledger

| Slice | Red / failing condition | Green condition | Evidence |
| --- | --- | --- | --- |
| Official ledger | Existing `api_surface.json` had only partial rows and legacy exclusions | Official OpenAPI audit recorded; 49 official ops mapped, 45 implemented, 1 blocked, 3 excluded; validate clean | `traces/notion-official-openapi-audit.md`; `connectorgen validate .../notion` exit 0 |
| Streams | Notion conformance dynamic stream checks were skipped | Metadata dynamic skip removed; 25 executable streams have sanitized fixtures and schemas | `go test ./internal/connectors/conformance -run 'TestConformance/notion' -count=1` exit 0 |
| Writes | No `writes.json`; `metadata.capabilities.write=false` | 22 typed write actions with fixtures pass request-shape and delete semantics conformance | same conformance gate exit 0 |
| Hook | Hook only knew `databases`, `pages`, `users` | Hook handles all executable Notion GET/POST read routes, cursor pagination, path/query/body templating, single-object streams, repeated-cursor fail-closed behavior, stream-specific POST pagination keys, and projection | `go test ./internal/connectors/hooks/notion -count=1` exit 0 |
| Docs/generated | Manual/SKILL said read-only and listed 3 streams | Generated Notion MANUAL/SKILL and website catalog data describe expanded read/write support truthfully | `./pm docs validate --connectors-dir docs/connectors` exit 0; website generation completed |
| Required exact gate | `connectorgen validate internal/connectors/defs/notion` failed on single-bundle path | `validatePath` accepts a single bundle directory, including a malformed single bundle missing metadata, and validates exactly one connector | `go test ./cmd/connectorgen -run 'TestValidatePath|TestValidate_AcceptsGoodBundle' -count=1` exit 0; exact validate command exit 0 |
| End-to-end local gates | Full repo verify initially timed out in long-running existing certify/CLI packages under parallel load | Re-ran package tests with matching cache, then `make verify` completed cleanly | final `make verify` exit 0 |

## Verification ledger

```bash
go run ./cmd/connectorgen validate internal/connectors/defs/notion                         # exit 0; 1 connector checked, 0 findings
go test ./internal/connectors/conformance -run 'TestConformance/notion' -count=1           # exit 0
go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1                            # exit 0
go build ./cmd/pm                                                                          # exit 0
make connector-boundary                                                                    # exit 0; outcome clean
git diff --check                                                                           # exit 0
./pm docs validate --connectors-dir docs/connectors                                        # exit 0
make verify                                                                                # exit 0 after rerun/cached long-running packages
```

Notes:

- Local reviewer findings were addressed before final verification: archive/trash-capable Notion update actions now require destructive confirmation, meeting-note pagination uses `limit` instead of an added `page_size`, repeated cursors fail closed, malformed single-bundle validation is covered, and generated catalog/index Notion rows are current.
- `make verify` was attempted multiple times. Earlier attempts timed out in unrelated long-running full-package tests (`internal/cli` and then `internal/connectors/certify`) under the fixed package `-timeout 20m`; focused reruns of those tests/packages passed, and the final required `make verify` run passed cleanly after review fixes.
- No live Notion workspace API calls, credentials, provider writes, live file uploads, certification, `/no-mistakes`, push, PR, merge, VPS, or Thaalam actions were performed.
