# Verification checklist — Issue 156 Zendesk Support parity

## Required local commands

- [x] `go run ./cmd/connectorgen validate internal/connectors/defs` — passed, 548 connectors, 0 findings.
- [x] Temp single-connector validation (`cp -R internal/connectors/defs/zendesk-support /tmp/zendesk-support-defs-validate/ && go run ./cmd/connectorgen validate /tmp/zendesk-support-defs-validate`) — passed, 1 connector, 0 findings. Note: the issue-listed direct connector-dir form currently treats `fixtures/` and `schemas/` as connectors and fails; no shared tool edit was made.
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/zendesk-support' -count=1` — passed.
- [x] `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1` — passed.
- [x] `go test ./cmd/connectorgen -count=1` — passed.
- [x] `go vet ./internal/connectors/... ./internal/cli/...` — passed.
- [x] `go build ./cmd/pm` — passed.
- [x] `go run ./cmd/pm docs validate --connectors-dir docs/connectors` — passed.
- [x] `make connector-boundary` — clean.
- [x] `git diff --check` — passed.
- [x] Read-only reviewer subagent re-review — no critical/warning findings after fixes.

## Non-live safety verification

- [x] No provider credentials requested or read.
- [x] No live Zendesk provider calls made.
- [x] No live writes, certification, VPS/Thaalam, merges, or dependency changes.
- [x] No shared runtime files edited.
- [x] Secret-shaped literal scan remains clean through connectorgen/conformance.
- [x] Destructive/delete operations are represented as in-scope blocked/typed metadata or existing typed write actions with `confirm: "destructive"`, not blanket-excluded as unsafe.

## CLI/help/docs parity notes

- [x] `cli_surface.json` validates as connector-owned command metadata.
- [x] `docs.md` records operation-ledger provenance, safety gates, blocked dependencies, and fixture-only certification status.
- [x] Generated Zendesk connector docs (`docs/connectors/zendesk-support/MANUAL.md`, `SKILL.md`) and `internal/cli/testdata/golden_transcripts.json` were updated for the connector command surface. Broad docs generator drift for unrelated connectors was reverted.
