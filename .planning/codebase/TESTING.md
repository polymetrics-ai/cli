# Testing

**Analysis Date:** 2026-07-08
**Generated via:** Upstream `/gsd:map-codebase` workflow shape, issue #122 prompt.

## Primary Gates

Project-level gates from `AGENTS.md` and `Makefile`:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
```

`make verify` composes:

- `fmt`
- `tidy-check`
- `vet`
- `test`
- `build`
- `docs-check`
- `smoke`
- `lint`
- `connectorgen-validate`

## Connector-Specific Gates

- `go run ./cmd/connectorgen validate internal/connectors/defs`
- `go test ./internal/connectors/engine/...`
- `go test ./internal/connectors/conformance/...`
- `go test ./internal/connectors/certify/...`
- Per-connector conformance subtests where applicable.
- Certification replay/fixture gates for no-secret verification.

## Planning-Only Issue #122 Gates

Issue #122 must not edit Go source. Required verification:

```bash
node -e "JSON.parse(require('fs').readFileSync('.planning/config.json','utf8')); console.log('config ok')"
test -f .planning/PROJECT.md
test -f .planning/REQUIREMENTS.md
test -f .planning/ROADMAP.md
test -f .planning/STATE.md
test -d .planning/codebase
rg -n "connector parity|reverse ETL|binary|direct-read|GraphQL|XML|SOAP|CDC|queue|human-gated|certification|conformance" .planning

git diff --check
git diff --name-only -- cmd internal
```

Expected output for the final source guard:

```text
# no output
```

## Runtime-Backed Optional Gates

Optional and credential/service-gated:

```bash
scripts/runtime.sh doctor
scripts/runtime.sh up
POLYMETRICS_INTEGRATION=1 go test ./...
scripts/runtime.sh down
```

Do not run these for issue #122 unless explicitly requested by a human.

## TDD / Evidence Pattern

For behavior-changing issues, GSD requires red-first tests before implementation. Issue #122 is planning-only; its red/green evidence is document/state validation:

- Red: legacy `.planning/` exists, codebase maps absent, stale phase tree present.
- Green: upstream-shaped `.planning/` exists, codebase maps exist, connector parity language covers non-REST surfaces, config parses, no Go source changed.

---
*Testing analysis: 2026-07-08*
