# Binary size analysis after Wave 6 legacy removal

Date: 2026-07-06
Branch: `feat/connector-architecture-v2`

## Finding

The post-removal binary size increase was not caused by deleting legacy connector
Go code. It was caused by the Wave 6 registry flip making `cmd/pm` import
`internal/connectors/defs.FS`; the previous `//go:embed all:*` directive compiled
the entire definition tree into the production CLI, including conformance-only
fixtures and API coverage manifests.

`go tool nm` showed the increase as static string/rodata, and the raw defs tree
confirmed the same pattern: most of the added bytes were JSON/Markdown assets,
not executable code.

## Measured binary sizes

| Build | Bytes | Approx | Notes |
| --- | ---: | ---: | --- |
| `origin/main` default | 68,512,000 | 65 MiB | Pre-v2 runtime baseline |
| Pre-deletion PR head default | 68,943,616 | 66 MiB | Before production defs embed mattered |
| Wave 6 all-defs embed default | 106,028,688 | 101 MiB | Embedded full defs tree |
| Optimized runtime embed default | 74,255,760 | 71 MiB | Current branch after this change |
| `origin/main` stripped | 49,599,344 | 47 MiB | `-trimpath -ldflags="-s -w"` |
| Pre-deletion PR head stripped | 49,917,248 | 48 MiB | Same stripped mode |
| Wave 6 all-defs embed stripped | 90,945,616 | 87 MiB | Embedded full defs tree |
| Optimized runtime embed stripped | 59,172,752 | 56 MiB | Current branch after this change |

The optimization removes about 31,772,928 bytes from the default build and about
31,772,864 bytes from the stripped build compared with the all-defs embed.

## Compiled-file mapping

Production `cmd/pm` now embeds only the files needed to construct runnable
connectors:

| Defs artifact | Files | Raw bytes | Compiled into `cmd/pm` | Runtime purpose |
| --- | ---: | ---: | --- | --- |
| `metadata.json` | 547 | 499,606 | yes | Connector identity, capability flags, catalog metadata |
| `spec.json` | 547 | 826,631 | yes | Config schema and secret partitioning |
| `streams.json` | 543 | 3,001,184 | yes | Declarative check/read request behavior |
| `writes.json` | 223 | 6,787,703 | yes | Reverse ETL write actions |
| `schemas/*.json` | 6,791 | 7,402,300 | yes | Stream catalog, projection, primary keys, cursors, sync modes |
| `docs.md` | 547 | 5,304,487 | yes | Human/agent connector manuals and docs checks |
| `api_surface.json` | 547 | 7,045,880 | no | Authoring/conformance coverage only |
| `fixtures/**` | 13,367 | 22,496,157 | no | Conformance replay only |

Runtime raw embedded defs data is now about 23.82 MB. Conformance-only data left
out of `cmd/pm` is about 29.54 MB. The full on-disk defs tree remains available
for validation and tests.

## Validation boundary

The engine loader now treats `api_surface.json` as optional at runtime. When the
file is present, the loader still parses and validates it.

`connectorgen validate internal/connectors/defs` remains strict: a bundle missing
`api_surface.json` is still reported as an authoring/conformance failure. The
conformance package now loads the real on-disk `internal/connectors/defs` tree so
fixtures and API-surface manifests stay covered without compiling them into the
shipping CLI.

## Speed effect

The same `connectors list --json` command produced byte-identical output before
and after the change, with 551 runtime catalog entries.

On the local smoke run, warm wall-clock timings improved from roughly
0.83-0.92s with the all-defs embed to roughly 0.65-0.69s with the runtime-only
embed. User CPU time dropped from roughly 1.08-1.19s to 0.85-0.89s. This is
expected because startup no longer walks/parses the API-surface manifests or
carries fixture payloads in the executable image.

## Further reduction options

1. Keep using release flags: `go build -trimpath -ldflags="-s -w" ./cmd/pm`.
   This produces the smallest normal Go release artifact measured here: 56 MiB.
2. Consider excluding `docs.md` from production embed if connector manuals can
   be generated from metadata/spec/schema summaries or loaded from an external
   docs pack. Potential raw saving: about 5.3 MB. This needs a UX decision
   because `pm connectors help` is user-facing.
3. Generate a compact runtime registry artifact from validated bundles. This
   would reduce JSON parser work and could improve startup speed, but it adds a
   build-generation step and must preserve reviewable source bundles.
4. Lazy-load full bundle details for `inspect`, `read`, and `write` instead of
   loading every stream/write/schema for commands that only need catalog-level
   metadata. This is the best speed-oriented follow-up but touches registry
   contracts more broadly.
5. Keep conformance fixtures out of production permanently. Re-embedding
   `fixtures/**` would add about 22.5 MB of raw JSON with no runtime value.

