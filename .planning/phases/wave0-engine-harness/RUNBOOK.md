# RUNBOOK — wave0-engine-harness

## 1. Running the gates locally

Existing baseline (unchanged): `make verify` = `fmt tidy-check vet test build docs-check smoke`
(`Makefile:60`).

Wave0 additions (after task B-18 lands):

```sh
make lint                  # golangci-lint run (config: .golangci.yml)
make connectorgen-validate # go run ./cmd/connectorgen validate internal/connectors/defs
make verify                # extended: ... smoke lint connectorgen-validate
```

Targeted suites:

```sh
go test ./internal/connectors/engine -v                 # engine unit tests
go test -cover ./internal/connectors/engine             # coverage gate (>=85%)
go test ./internal/connectors/conformance -run TestConformance -v   # per-bundle subtests
go test ./internal/connectors/conformance -run 'TestConformance/stripe' -v
go test ./internal/connectors/certify -run TestSourceStages -v      # sample end-to-end
go test ./internal/connectors/engine -run 'TestParity' -v           # 3 golden parity suites
go run ./cmd/connectorgen validate internal/connectors/defs
go run ./cmd/inventorygen                                # regenerates docs/migration/inventory.json
```

Notes:
- golangci-lint is not currently installed on this machine; the `lint` target uses the
  coordinator-decided acquisition path (SPEC §5). Everything else is stdlib `go` + `make`.
- No runtime services are required (no Podman/Docker); postgres golden tests run in fixture mode.
- Certify tests create ephemeral roots via `os.MkdirTemp` and delete them; a leftover
  `pm-certify-*` temp dir after a green run is a bug.

## 2. How the coexistence works (operational view)

There is **no runtime flag**. Wave0 code is reachable only through:
1. Tests (engine connectors constructed from `defs.FS` in `_test.go` files),
2. Dev tools (`cmd/connectorgen`, `cmd/inventorygen`) run explicitly.

The production registry path is byte-identical to before the phase:
`internal/cli/cli.go:983 appRegistry() → registryset.New() → connectors.NewLiveRegistry()`, with
legacy `internal/connectors/{stripe,searxng,postgres}` still self-registering via `init()`.
`cmd/registrygen` skips the new dirs (`defs`, `engine`, `hooks`, `native`, `conformance`,
`certify`) so regeneration cannot pull them in. The registry flip to bundles is wave6, behind the
roadmap HUMAN GATE.

## 3. Rollback

Wave0 is purely additive plus two small shared-file edits (Makefile targets, registrygen skip
map). To roll back:

```sh
git log --oneline            # identify wave0 commits (one per PLAN.md wave-close)
git revert <range>           # or: git revert --no-commit <c1>..<cN> && git commit
make verify                  # must be green — legacy paths were never modified
```

Verification after rollback: `pm connectors list` output unchanged; `make smoke` green;
`git status --porcelain` clean. Because no legacy file's behavior changed and nothing imports the
new packages outside their own tests/tools, reverting cannot strand references. Partial rollback
(e.g. keep engine, drop a golden bundle) is safe at directory granularity for
`internal/connectors/defs/<name>/` + its parity test file.

Data: no schema/state migrations in this phase; saved sync state (`connection:stream` keys,
`internal/app/sync_modes.go:96`) untouched.

## 4. Failure triage

| Symptom | Likely cause | Action |
|---|---|---|
| `connectorgen validate` fails on a golden | bundle drift vs meta-schema or spec keys | read finding (file+rule), fix bundle; never relax the validator |
| `TestConformance/<name>` dynamic failure | fixture/request-key mismatch | check replay envelope request `path+query` vs engine request build |
| Parity test diff | engine feature gap | file typed `ENGINE_GAP` blocker; do NOT special-case the golden |
| `registryset` diff after `go run ./cmd/registrygen` | skip map incomplete | must be byte-identical in wave0; stop + escalate (T-16 guard) |
| Coverage < 85% on engine | untested branch in read/write | add table rows, don't lower the gate (quality-gate reduction = human gate) |
| certify sample stage red | CLI flag drift | fix stage driver against `internal/cli` docs; cli edits are out of scope — blocker if truly required |

## 5. Wave-close procedure (orchestrator only, per wave in PLAN.md)

```sh
go run ./cmd/registrygen                      # must be a no-op diff in wave0
go build ./... && go test ./internal/connectors/...
make lint 2>/dev/null || true                 # once B-18 has landed: required
git status --porcelain                        # path-guard vs the wave's assigned files
git add -A && git commit -m "wave0 <wave-letter>: <summary>"
```
