# TDD Ledger — Gong CLI Parity Parent (#133)

## GSD/TDD setup

- `scripts/gsd doctor`: pass.
- `scripts/gsd verify-pi`: pass.
- `scripts/gsd list --json`: pass.
- `scripts/gsd prompt plan-phase 133 --skip-research --tdd`: prompt rendered.
- `scripts/gsd prompt programming-loop init --phase issue-133-gong-cli-parity --dry-run`: failed with `scripts/gsd: unknown GSD command: programming-loop`.
- Manual-GSD fallback recorded: use `/pm-orchestrate` and `/pm-gsd-loop` prompt bodies plus `gsd-universal-runtime-loop.md`; do not skip red/green/refactor evidence.

## Parent red/green strategy

Parent #133 is orchestration scope. Each sub-issue owns its behavior tests.

| Lane | Red evidence owner | Green evidence target |
|---|---|---|
| #144 operation ledger | `go test ./cmd/connectorgen -run GongAPISurfaceOperationLedger -count=1` fails against current 10-entry surface | exact 67-operation ledger, no legacy exclusions |
| #141 CLI surface metadata | metadata validation/help discovery test fails until `cli_surface.json` exists | validated Gong CLI command surface metadata |
| #142 help renderer/docs | help/docs parity test fails until metadata renders | runtime help/docs/website parity |
| #143 stream runner | stream-backed command test fails until runner handles Gong streams | stream commands execute through generic runner |
| #145 direct reads | bounded direct-read tests fail until operations are implemented | max-bytes/redaction/output-policy enforced |
| #146 body/binary engine | body/binary shape tests fail until fixed schemas and bounds exist | no raw JSON/body escape hatch; bounded binary policy |
| #147 sensitive/admin policy | preview/redaction/confirmation tests fail until policy exists | typed reverse-ETL writes with risk/redaction/confirmation |

## Current red slice

Active local critical path: #144. See `.planning/phases/issue-144-gong-operation-ledger/TDD-LEDGER.md`.

#144 status: red captured, green targeted checks passed. Full parent verification remains pending.

## Refactor notes

- Parent orchestration docs may change as sub-issues land.
- Operation-ledger rows remain metadata-only until executor lanes implement typed surfaces.
