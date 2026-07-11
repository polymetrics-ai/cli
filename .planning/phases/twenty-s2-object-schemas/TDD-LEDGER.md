# twenty S2 object schemas (#279) — TDD ledger

Manual-GSD fallback (repo-local adapter lacked `programming-loop`); recorded here on the branch so the
`gsd-workflow-evidence` gate travels with the implementation. Full trace:
`.planning/auto-loop/tasks/S2/TDD-LEDGER.md`.

Data-only slice: the 28 schema JSON files are declarative connector data. The executable gate is the
standing loader/conformance suite (`go test ./...`) plus `connectorgen validate`, which the schemas
must not break — no new Go test file is warranted for pure schema data.

| # | Slice | Test / gate (RED first) | GREEN when | Status | Commit |
|---|-------|-------------------------|------------|--------|--------|
| 1 | 28 schemas match manifest | manifest/output assertion RED: files absent, counts 0/546 | 28 files emitted from authoritative `FIELD-MANIFEST.json`, per-object counts match, total = 546 | done | 7a89f46c |
| 2 | draft-07 valid | `Draft7Validator.check_schema` RED against missing/invalid schemas | all 28 valid draft-07; `x-primary-key:["id"]`, `x-cursor-field:"updatedAt"` present | done | 7a89f46c |
| 3 | loader/conformance stays green | `go test ./internal/connectors/...` would break if schema shape invalid | full `go test ./...` GREEN uncached; `streams.json` unchanged (`streams:[]`) | done | 7a89f46c |

## Notes
- Red: before production edits, manifest shape/counts and generator output validated in `/tmp`
  (throwaway generator, not committed); existing loader/conformance/full tests are the green gate for
  this data-only slice.
- Green: `make verify` exit 0, uncached, at `7a89f46c` — `gofmt -w`, `go mod tidy` diff check,
  `go vet ./...`, `go test -timeout 20m ./...`, `go build ./cmd/pm`, docs validate, smoke,
  golangci-lint scoped run, `connectorgen validate internal/connectors/defs` (548 connectors, 0 findings).
  Orchestrator re-ran `make verify` independently post-worker (worker pid dead, tree clean) — exit 0,
  corroborating CI `verify`.
- Scope: commit contains only `internal/connectors/defs/twenty/schemas/*.json`; no other def/file touched.
