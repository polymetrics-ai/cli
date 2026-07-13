# twenty S2 object schemas (#279) — verification

Branch-local mirror of the orchestrator VERIFY stage so the `gsd-workflow-evidence` gate travels with
the implementation. Full trace: `.planning/auto-loop/tasks/S2/VERIFICATION.md`.

- [x] 28 files under `internal/connectors/defs/twenty/schemas/`, one per object `namePlural`
- [x] per-object top-level properties count == `FIELD-MANIFEST.json` field_count; total = 546
- [x] every property name/type traces to `FIELD-MANIFEST.json` (no invented/dropped fields)
- [x] each file valid draft-07; `x-primary-key:["id"]` + `x-cursor-field:"updatedAt"` present
- [x] gofmt clean; `go build ./cmd/pm`; `go test ./...` GREEN uncached; `make verify` GREEN
- [x] `streams.json` still `streams:[]` (unchanged); no other def/file touched
- [x] committed + pushed on `feat/279-twenty-object-schemas` at `7a89f46c`

## Gate evidence (manual-GSD / TDD)
- Red: manifest/generator assertions authored first over `/tmp` output before production edits.
- Green: worker gates all PASS at `7a89f46c` (manifest assertion, draft-07 check, gofmt, build,
  `go test ./internal/connectors/...`, `go test ./...`, `make verify`).
- Green: orchestrator independent re-run of `make verify` post-worker = exit 0, tree clean.
- CI corroboration: PR #288 `verify` job on `7a89f46c` (auto-review errors on empty `ANTHROPIC_API_KEY`
  — infra, not a code review; dispositioned at REVIEW per S1 precedent).
