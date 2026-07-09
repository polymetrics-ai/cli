# Plan: Chatwoot CLI Surface Metadata

Sub-issue: #149
Parent issue: #148
Parent branch: `feat/148-chatwoot-cli-parity`
Working mode: local critical path (Pi API session lacks `subagent` tool)

## Goal

Refresh Chatwoot connector metadata from the official Swagger source and add provider-shaped CLI surface metadata without enabling unsafe raw API access.

Official source: `https://raw.githubusercontent.com/chatwoot/chatwoot/develop/swagger/swagger.json`

## Scope

Allowed production files:

- `internal/connectors/defs/chatwoot/metadata.json`
- `internal/connectors/defs/chatwoot/api_surface.json`
- `internal/connectors/defs/chatwoot/cli_surface.json` (new)
- `internal/connectors/defs/chatwoot/docs.md`

Allowed planning files:

- `.planning/phases/issue-149-chatwoot-cli-surface-metadata/**`
- parent orchestration state updates under `.planning/phases/issue-148-chatwoot-cli-parity/**`

Out of scope for #149:

- Runtime help renderer changes (#150).
- Stream/schema/fixture expansion beyond metadata accounting (#151).
- Implementing direct-read execution (#153).
- Advanced body/query/binary execution (#154).
- Sensitive/admin reverse-ETL execution policy (#155).
- Any live Chatwoot API calls or credential use.

## Red / Green plan

1. Red validation: compare official Swagger operations (expected 144) to current `api_surface.json` entries (current 71). Expect failure before edits.
2. Generate/check official operation inventory with method split: POST 41, GET 62, PATCH 21, DELETE 18, PUT 2.
3. Update `api_surface.json` to operation-ledger mode with exactly 144 endpoint rows.
   - Preserve existing executable coverage for 7 streams and 6 write actions.
   - Classify non-executable official rows as blocked `operation` rows, not legacy `excluded` rows.
   - Keep direct-read, sensitive/admin reverse-ETL, destructive, binary/file-adjacent, duplicate, and disallowed candidates blocked by default.
4. Add `cli_surface.json` mapping Chatwoot-shaped commands to implemented streams/writes and planned/blocked safe intents.
   - Implemented reverse-ETL commands must include risk text, approval text, and required record-field flag mappings.
   - No `raw_api`, `direct_write`, generic HTTP write, generic SQL write, or shell escape hatch.
5. Update `metadata.json`/`docs.md` to describe full-surface accounting without claiming unsupported execution.
6. Green validation:
   - official surface count script exits 0;
   - `jq empty` passes;
   - `go test ./cmd/connectorgen -run CLISurface -count=1` passes;
   - `go test ./internal/connectors/engine -run CLISurface -count=1` passes;
   - `go run ./cmd/connectorgen validate internal/connectors/defs` passes (current validator expects the defs root; validating the connector subdirectory treats nested `fixtures/` and `schemas/` as connector dirs).

## CLI help/docs/website parity

Applies: yes, connector CLI surface metadata is CLI-visible once renderer slices consume it.

This issue only adds metadata and connector docs. Runtime help renderer, `docs/cli/**`, website docs, generated help/manual artifacts, and runtime `pm chatwoot` checks are deferred to #150 because no renderer behavior changes are in this slice.

## Safety checklist

- [ ] No secrets in files, prompts, logs, examples, or fixtures.
- [ ] No credentialed connector checks.
- [ ] No reverse ETL execution.
- [ ] No raw generic HTTP write or direct-write command.
- [ ] Sensitive/admin/destructive operations remain blocked-by-default metadata unless already represented by a safe typed write action.
- [ ] No new dependencies.
