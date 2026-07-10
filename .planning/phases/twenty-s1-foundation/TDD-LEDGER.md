# twenty S1 foundation (#278) — TDD ledger

Strict red→green per slice. Manual-GSD fallback (adapter lacked `programming-loop`); recorded here on
the branch so the gsd-workflow-evidence gate travels with the implementation. Full trace:
`.planning/auto-loop/tasks/S1/TDD-LEDGER.md`.

| # | Slice | Test / gate (RED first) | GREEN when | Status | Commit |
|---|-------|-------------------------|------------|--------|--------|
| 1 | bundle + metadata loads | `connectorgen validate .../twenty` fails: dir/`metadata.json` absent | `metadata.json` authored, loader parses capability flags | done | ac9f39c9 |
| 2 | spec valid draft-07 + secret marked | spec-validation over `spec.json` fails: absent | `spec.json` draft-07, `required:["api_key"]`, `x-secret` | done | ac9f39c9 |
| 3 | loader-valid bundle (no panic) | `go test ./...` panics: `docs.md`/`streams.json` missing under tree-wide embed | `docs.md` (5 headings) + `streams.json` stubs added | done | 7b5bac92 |
| 4 | conformant empty-surface skeleton | conformance `surface_complete`/`docs_present` RED; count tests 547/551 | `api_surface` endpoints=[], counts bumped 548/552 | done | 7b5bac92 |
| 5 | generated docs regenerated | `pm docs validate` RED: catalog 551 want 552 | catalog + `docs/connectors/twenty/` manual regenerated | done | c769719e |

## Notes
- Red: partial bundle broke the tree-wide `//go:embed` load-all → repo-wide `go test ./...` panic.
- Red: bidirectional `checkSurfaceComplete` rejects a full covered_by api_surface against empty
  streams/writes → S1 must ship an empty-surface skeleton; S3/S4/S5 re-materialize the 168 ops.
- Green: `make verify` exit 0, uncached, at `c769719e`; CI `verify` run 29116868346 agrees.
