# Plan — Zoom documented-operation parity, Wave 2 (qss module), R1

## Delivery record

- Phase: `cli-zoom-parity-wave2-qss-r1`
- Scope owner: `internal/connectors/defs/zoom/**` (operations.json, cli_surface.json,
  api_surface.json, docs.md, metadata.json, fixtures/direct_reads/**,
  command_surface_test.go) plus scoped `docs/connectors/zoom/{MANUAL.md,SKILL.md}` and
  `docs/connectors/README.md` regeneration, and this phase trace only.
- Parent program: zoom full documented-operation parity (1,913 ops; see
  `.planning/phases/cli-zoom-parity-implementation-r1/` for Wave 0/Wave 1 — the 3 existing
  stream-backed reads this wave builds on top of).
- Required skills loaded: `golang-how-to`, `golang-cli`, `golang-testing`,
  `golang-error-handling`, `golang-safety`, `golang-design-patterns`,
  `golang-structs-interfaces`, `golang-documentation`, plus
  `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`.
- GSD note: the repo-local Pi adapter has no `phase_found` record for this slice name (same
  documented gap as Wave 1's PLAN.md). This directory is the recorded inline/manual fallback,
  per AGENTS.md's "Where the lifecycle genuinely cannot run, record an explicit manual-GSD
  fallback with its red/green evidence in the planning file."

## Live-artifact re-verification (methodology step 1)

`internal/connectors/defs/zoom/api_surface.json` (`reviewed_at: 2026-08-05`) already records the
`qss` provider module with 3 operations, sourced from
`https://developers.zoom.us/docs/api/qss.md`. Re-fetched directly (not trusted) on 2026-08-07:

```
curl -sS -o /tmp/zoom_qss.md -w 'http_code=%{http_code} bytes=%{size_download}\n' \
  https://developers.zoom.us/docs/api/qss.md
=> http_code=200 bytes=28282
```
Retrieved at 2026-08-07T21:44:22Z. The fetched document's `## Operations` section lists exactly
three operations, matching the recorded ledger exactly:

| Method | Path | Ledger operation_title |
| --- | --- | --- |
| GET | `/metrics/meetings/{meetingId}/participants/qos_summary` | List meeting participants QoS Summary |
| GET | `/metrics/webinars/{webinarId}/participants/qos_summary` | List webinar participants QoS Summary |
| GET | `/videosdk/sessions/{sessionId}/users/qos_summary` | List session users QoS Summary |

Delta vs. recorded ledger: **0**. The document has **no "Parameters"/"Query Parameters" heading**
for any of the three operations — the only query-shaped fields present (`next_page_token`,
`page_size`, `type`) appear solely inside each operation's **response body** schema (`## Responses`
→ `### Status: 200` → `**All of:**`), not as documented request inputs. This is the reason the
commands below expose only the required path parameter and no query flags (see "Locked decisions").

Also spot-checked two smaller Wave-3-candidate modules from the same artifact family for
provenance evidence (not implemented this wave): `my-notes.md` (200, 6739 bytes) and
`ai-companion.md` (200, 5156 bytes).

## Locked decisions and scope fences

- This wave implements exactly the 3 `qss` module operations as `direct_read` commands. It does
  not touch any other of the 35 provider modules, does not add `writes.json` (capabilities.write
  stays `false`), and does not modify the shared engine/commandrunner packages.
- Model: `direct_read` (matches the already-recorded ledger `operation.model`, not re-derived).
  `output_policy: json_redacted` on all three (the only policy in
  `supportedDirectReadOutputPolicies` that fits a non-clinical, non-repository JSON body).
- **Do not hand-author `next_page_token`/`page_size` request-query flags.** The live artifact
  documents no request-parameters section for these operations (see above); inventing accepted
  query parameters the artifact does not document would misrepresent the API. Each command takes
  only its one required path-parameter flag (`--meeting-id`, `--webinar-id`, `--session-id`).
  Revisit only against a provider artifact that documents request parameters explicitly.
- Discovered, not pre-planned: the shared `json_redacted` policy's `shouldRedactJSONField`
  (`internal/connectors/engine/direct_read.go`) redacts any field whose name contains `token` —
  including the response's own `next_page_token` pagination cursor, which is not a credential.
  This is existing shared-engine behavior; it is documented in `docs.md`'s new qss subsection
  rather than changed (changing shared redaction heuristics is out of this connector-scoped lane
  per AGENTS.md's "connector implementation lanes" rule and would need its own foundation issue).
- `api_surface.json`: the 3 `qss` endpoint rows move from `operation` (blocked) to
  `covered_by.direct_read = "<command path>"`. No other row changes.
- `operations.json` did not exist before this wave; it is created with exactly these 3 entries.

## Execution slices

1. **RED** — extend `command_surface_test.go`'s existing ledger/reachability assertions
   (`covered` 3→6, `implementable_now` 1839→1836, exact command count 3→6) and add a new
   `TestQSSOperationDirectReadCommandsExecuteWithFixtures` test, run against the **unmodified**
   production JSON (qss still blocked) to capture a genuine failing baseline. Fixture files for
   the new direct reads are authored and committed at this step so the green commit only needs to
   touch `operations.json`/`cli_surface.json`/`api_surface.json`/`docs.md`/`metadata.json`.
2. **GREEN** — author `operations.json` (3 `rest_read` entries), extend `cli_surface.json` (new
   `qss` group + 3 commands, `api_surface`/`output_policy` filled by
   `connectorgen surface-sync`, never hand-authored), flip the 3 `api_surface.json` rows to
   `covered_by`, update `docs.md`/`metadata.json` description and counts. Regenerate the runtime
   `operation_endpoint_ledger.json` projection via `surface-sync`.
3. **Binary verification** — build `./cmd/pm`, run `pm zoom qss --help` and each command's
   `--help`, then execute all three commands with a synthetic (non-real) credential against the
   **live** Zoom host to prove each resolves to the correct method/path (expect a structured 401
   from Zoom itself, not a client-side "unknown command" — the exact failure mode a prior
   connector wave shipped undetected).
4. **Scoped docs/CLI parity** — hand-patch `docs/connectors/zoom/{MANUAL.md,SKILL.md}` and the
   `docs/connectors/README.md` zoom line to match what `pm docs generate` would produce for the
   zoom-specific delta only. `pm docs generate` was run once read-only to diff against; its
   repo-wide output was discarded because it also rewrote all 550 other connectors' docs with an
   unrelated pre-existing field-type-annotation generator change — out of this lane's scope per
   AGENTS.md's "connector implementation lanes" rule. `pm docs validate` passes against the
   hand-patched result.
5. **Verification and handoff** — run the full local gate list scoped to zoom + connectorgen +
   conformance + commandrunner + internal/cli, commit, push, update the parent issue tree and
   status file.

## Issue tree

Supersedes the stale zoom issue tree (#3110 parent, #3111–#3117 children), which categorized by
invented cross-cutting lanes (`etl_cdc`, `direct_binary`, `reverse`, ...) rather than the
provider's own operation categories, and carried a stale reconciled total (184, not 1,913). New
tree: one parent + one sub-issue per Zoom API module (matching the dockerhub pilot's per-category
issue pattern). This wave's sub-issue is the `qss` module (3 ops). See the parent issue body for
the full 35-module breakdown and per-module status.

## Canonical references

- `AGENTS.md`, `docs/migration/conventions.md`
- `.planning/phases/cli-zoom-parity-implementation-r1/` (Wave 0/Wave 1 baseline)
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`
- `internal/connectors/engine/schema/{operations,cli_surface,api_surface}.schema.json`
