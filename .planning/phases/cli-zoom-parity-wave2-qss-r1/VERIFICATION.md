# Verification Checklist — Zoom documented-operation parity, Wave 2 (qss module), R1

## Live-artifact re-verification (before any implementation)

- [x] Re-fetched `https://developers.zoom.us/docs/api/qss.md` directly with `curl`:
  `http_code=200 bytes=28282`, retrieved `2026-08-07T21:44:22Z`.
- [x] The document's 3 operations (List meeting participants QoS Summary, List webinar
  participants QoS Summary, List session users QoS Summary) match the recorded
  `api_surface.json` ledger's `qss` module rows exactly (method, path, title). Delta: 0.
- [x] Confirmed the document has no "Parameters"/"Query Parameters" section for any of the three
  operations — locked the scope decision to path-parameter-only commands (see PLAN.md).

## RED

- [x] `command_surface_test.go` extended with the Wave 2 target counts and a new
  `TestQSSOperationDirectReadCommandsExecuteWithFixtures` test, run against the unmodified
  production JSON, and its failure captured verbatim (see TDD-LEDGER.md).
- [x] Committed the red state (`test(connectors): add failing zoom QSS direct-read coverage
  (red)`) before any production JSON changed.

## GREEN

- [x] `go test ./internal/connectors/defs/zoom/... -count=1` passes (4 tests, including the new
  `TestQSSOperationDirectReadCommandsExecuteWithFixtures`).
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs/zoom` passes.
- [x] `go run ./cmd/connectorgen validate` (all 551 connectors) passes.
- [x] `go run ./cmd/connectorgen surface-sync internal/connectors/defs --check` passes with zero
  drift (api_surface/output_policy for the 3 new commands were filled by a live, non-`--check`
  surface-sync run, then re-verified idempotent).
- [x] `go run ./cmd/connectorgen boundary` passes (183 files checked, 0 findings).
- [x] `go test ./internal/connectors/conformance/... -count=1` passes.
- [x] `go test ./internal/connectors/commandrunner/... -count=1` passes.
- [x] `go test -timeout 20m ./internal/cli/... -count=1` passes (after regenerating the 9 affected
  golden CLI transcripts with `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1`; verified programmatically
  that every changed line in every changed transcript entry mentions "zoom" — no unrelated
  drift).
- [x] `gofmt -l` and `go vet` clean on all touched Go files.

## Binary verification (not just reading the diff)

- [x] `go build -o /tmp/pm_zoom ./cmd/pm` succeeds.
- [x] `pm zoom --help` lists the new `qss` command group.
- [x] `pm zoom qss --help` lists all 3 new commands with `[availability=implemented]`.
- [x] `pm zoom qss meeting-participants list --help` shows the correct operation id, output
  policy, and required `--meeting-id` flag.
- [x] Each of the 3 commands executed with a real (synthetic, non-functional) credential against
  the **live** `https://api.zoom.us` host: all three reached the correct resolved path
  (`https://api.zoom.us/v2/metrics/{meetings,webinars}/.../qos_summary`,
  `https://api.zoom.us/v2/videosdk/sessions/.../qos_summary`) and received a structured 401
  `{"code":124,"message":"Invalid access token."}` from Zoom itself — proof the command resolves
  and dispatches correctly, not a client-side "unknown command" (the exact failure mode a prior
  connector wave shipped undetected per the task brief).
- [x] `pm docs validate --connectors-dir docs/connectors` passes.

## Scope discipline

- [x] `pm docs generate` was run once, read-only, to diff against; its full repo-wide output (551
  connectors, an unrelated pre-existing field-type-annotation generator drift) was discarded.
  `docs/connectors/zoom/{MANUAL.md,SKILL.md}` and the zoom line of `docs/connectors/README.md`
  were hand-patched to match exactly what the generator would have produced for the zoom-specific
  delta only (diffed against the generator's own output to confirm — the only remaining diff is
  the unrelated field-type drift, confirmed out of scope).
- [x] `git status --short` confined to: `internal/connectors/defs/zoom/**`,
  `docs/connectors/zoom/{MANUAL.md,SKILL.md}`, `docs/connectors/README.md` (1 line),
  `internal/connectors/defs/operation_endpoint_ledger.json` (generated projection, zoom rows
  only), `internal/cli/testdata/golden_transcripts.json` (9 zoom-only entries), and this phase's
  `.planning/phases/cli-zoom-parity-wave2-qss-r1/**`.
- [x] Zero `unsafe_or_disallowed` rows introduced or present.
- [x] No hand-authored `page`/`per_page`/`limit` flags.
- [x] No shared engine/commandrunner code changed; the `next_page_token` redaction behavior
  discovered during testing is documented, not worked around or silently patched.

## Issue tree

- [x] Stale zoom issue tree (#3110 parent, #3111-#3117 children — invented cross-cutting
  categories, stale 184-op total) closed with pointer comments.
- [x] New parent #3915 created with the full 35-module breakdown (1,913 ops, reconciled exactly).
- [x] 35 module sub-issues created (#3916-#3950), one per Zoom provider documentation module,
  attached to #3915.
- [x] This wave's module sub-issue (#3947, `qss`) implemented in full; PR closes it.

## Handoff

- 6 of 1,913 zoom operations now implemented (3 existing streams + 3 new `qss` direct reads).
- 34 module sub-issues remain open under parent #3915, ranging from 1 op
  (`customer-managed-keys-hybrid`, `ai-companion`) to 360 ops (`phone`). Recommended next slices
  for a follow-up worker: `my-notes` (#3948, 2 ops), `healthcare` (#3946, 3 ops), `chatbot`
  (#3944, 4 ops), `cobrowse-sdk` (#3945, 4 ops) — same small-all-GET-or-simple-write shape as this
  wave.
- `customer-managed-keys-hybrid` (#3950) needs a product/design decision before implementation:
  its one operation targets a customer-hosted `{keyConnectorLb}` server, not `api.zoom.us` — see
  the issue body and PLAN.md's Non-goals section.
