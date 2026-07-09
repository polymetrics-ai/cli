# TDD Ledger: Help Scout CLI Parity Parent

Parent issue: #212

## Manual GSD Fallback

`programming-loop` is not present in `.gsd/commands.json`; `scripts/gsd prompt programming-loop init --phase issue-212-helpscout-cli-parity --dry-run` failed with `unknown GSD command: programming-loop`. Manual GSD/TDD loop is active: plan first, capture red validation/test evidence, implement smallest green slice, verify, update artifacts, commit/push green checkpoints.

## Red / Pre-Implementation Evidence

- Parent PR missing: `gh pr list --head feat/212-helpscout-cli-parity ...` returned `[]`.
- Existing canonical bundle differs from kickoff prompt: `internal/connectors/defs/help-scout/` exists; `internal/connectors/defs/helpscout/` does not. The parent plan records canonical `help-scout` to avoid duplicate connector exposure.
- Existing Help Scout `api_surface.json` is legacy narrow coverage: 4 covered streams and broad out-of-scope rows; full Inbox API crawl observed 146 official endpoint pages and 145 unique method/path pairs.

## Planned Red Gates

### #213

- Add `cli_surface.json` metadata and run validation before all mappings are fixed; expected initial failures should be limited to `cli_surface_*` or `surface_*` rules.
- Refresh `api_surface.json` from official docs and validate that all declared streams remain covered.
- If a new validator gap is discovered, add a focused failing test in `cmd/connectorgen` before changing validator code.

### Later lanes

- #214: help/docs renderer tests fail before renderer/docs updates.
- #215: stream/conformance fixture checks fail before new streams/fixtures pass.
- #216: operation ledger validation fails before exact classifications pass.
- #217/#218: direct-read/binary validation fails before bounded policies pass.
- #219: sensitive/admin policy validation fails before approval/redaction metadata passes.

## Green Evidence

Pending.

## Refactor Evidence

Pending.
