# Plan — Zoom AI Services documented-operation parity, R1

## Delivery record

- Parent: [#3915](https://github.com/polymetrics-ai/cli/issues/3915); provider-owned slice:
  [#3935](https://github.com/polymetrics-ai/cli/issues/3935).
- Scope: every operation in Zoom's published **AI Services** category: Scribe, Summarizer, and
  Translator batch-job endpoints, synchronous actions, and the documented Live Scribe WebSocket
  upgrade. There are no exclusions and no `unsafe_or_disallowed` rows.
- Required skills loaded: `golang-how-to`, `golang-cli`, `golang-testing`,
  `golang-error-handling`, `golang-security`, `golang-safety`,
  `golang-design-patterns`, `golang-structs-interfaces`, and `golang-documentation`.
- GSD provenance: `scripts/gsd doctor`, `scripts/gsd sources`, and generated prompts for
  `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` ran on
  2026-08-08. This provider-category phase is not registered by the official runtime and the
  parent delivery contract forbids role spawning, so the documented inline/manual GSD fallback
  records discussion, plan, RED, GREEN, verification, and review evidence here.

## Inline discuss-phase record

AI Services is Zoom's own documented operation category, not a locally invented grouping. Its
twenty-two source operations are all callable provider contracts: twelve HTTP JSON reads, nine
approval-gated JSON mutations, and one `101 Switching Protocols` Live Scribe session. The source
states that Live Scribe takes a fixed `live-asr` WebSocket subprotocol, a first `session.update`
text frame, PCM16 16 kHz mono signed-little-endian binary audio frames, and returns the documented
bidirectional frame protocol. It is neither an inbound event nor a deferrable pseudo-endpoint.

The batch submit schemas can contain provider-hosted storage credentials and webhook signing
material. Commands therefore accept only typed, provider-closed JSON objects and use the existing
redacted output/preview policy. Tests use only synthetic values and never put credentials,
token-derived values, signed URLs, or transcripts in a diagnostic assertion.

## Live artifact audit — completed before RED

The provider artifact was fetched afresh; the inherited operation ledger was then compared rather
than trusted.

| Item | Evidence |
| --- | --- |
| API URL | `https://developers.zoom.us/docs/api/ai-services.md` |
| Retrieval | `2026-08-08T20:56:08Z` |
| HTTP / bytes | `200` / `87,750` |
| SHA-256 | `154631ef97c292468c81a79dc50cd51ea142d18f1f9fab060622215ddf3ba367` |
| Artifact | OpenAPI `3.1.1`, API version `2`, server `https://api.zoom.us/v2` |
| Ledger comparison | exactly 22 local `provider_module=ai-services` rows; method, path, title, and source URL match (delta `0`) |

The source declares exactly these operations:

| Method | Path | Provider title |
| --- | --- | --- |
| GET | `/aiservices/scribe/jobs` | List Batch Jobs |
| POST | `/aiservices/scribe/jobs` | Submit Batch Scribe Job |
| GET | `/aiservices/scribe/jobs/{jobId}` | Get Batch Job Status |
| DELETE | `/aiservices/scribe/jobs/{jobId}` | Cancel Batch Job |
| GET | `/aiservices/scribe/jobs/{jobId}/files` | List Batch Job Files |
| GET | `/aiservices/scribe/jobs/{jobId}/files/{fileId}` | Get Batch Scribe Job File |
| GET | `/aiservices/scribe/live` | Live Scribe WebSocket |
| POST | `/aiservices/scribe/transcribe` | Scribe (Synchronous) |
| GET | `/aiservices/summarizer/jobs` | List Batch Summarizer Jobs |
| POST | `/aiservices/summarizer/jobs` | Submit Batch Summarizer Job |
| GET | `/aiservices/summarizer/jobs/{jobId}` | Get Batch Summarizer Job |
| DELETE | `/aiservices/summarizer/jobs/{jobId}` | Cancel Batch Summarizer Job |
| GET | `/aiservices/summarizer/jobs/{jobId}/files` | List Batch Summarizer Job Files |
| GET | `/aiservices/summarizer/jobs/{jobId}/files/{fileId}` | Get Batch Summarizer Job File |
| POST | `/aiservices/summarizer/summarize` | Summarize (Synchronous) |
| GET | `/aiservices/translator/jobs` | List Batch Translator Jobs |
| POST | `/aiservices/translator/jobs` | Submit Batch Translator Job |
| GET | `/aiservices/translator/jobs/{jobId}` | Get Batch Translator Job |
| DELETE | `/aiservices/translator/jobs/{jobId}` | Cancel Batch Translator Job |
| GET | `/aiservices/translator/jobs/{jobId}/files` | List Batch Translator Job Files |
| GET | `/aiservices/translator/jobs/{jobId}/files/{fileId}` | Get Batch Translator Job File |
| POST | `/aiservices/translator/translate` | Translate (Synchronous) |

The artifact mentions `page_size` and `next_page_token` only as provider pagination semantics for
the batch listing responses. No command-specific `page`, `per_page`, `limit`, `page_size`, or
`next_page_token` flag will be hand-authored; any allowable navigation comes from declared paging
and its shared runtime flags.

### Continuation re-fetch — 2026-08-09

The paused phase was independently re-fetched before any continuation work. This confirms that the
same official provider artifact—not the 426-sweep generated surface—is still the source of the
next implementation decision.

| Item | Evidence |
| --- | --- |
| API URL | `https://developers.zoom.us/docs/api/ai-services.md` |
| Retrieval | `2026-08-09T20:57:31Z` |
| HTTP / bytes | `200` / `87,750` |
| SHA-256 | `154631ef97c292468c81a79dc50cd51ea142d18f1f9fab060622215ddf3ba367` |
| Result | byte-for-byte identical to the 2026-08-08 audited artifact above |

The current branch still has exactly 22 `provider_module=ai-services` rows. The continuation does
not import the 426-sweep surface: that branch's 1,913 rows remain a checklist only because it has
only three implemented operations.

### Active consumer re-fetch — 2026-08-10 (IST)

Immediately before declaration authoring, the official artifact was fetched again from
`https://developers.zoom.us/docs/api/ai-services.md` at `2026-08-09T23:06:39Z` (2026-08-10 IST).
It returned HTTP `200`, exactly `87,750` bytes, and SHA-256
`154631ef97c292468c81a79dc50cd51ea142d18f1f9fab060622215ddf3ba367`. This is byte-for-byte
identical to both earlier audited retrievals. The consumer continues from the same 22 source
operations; it does not import the 426-sweep generated Zoom surface.

## Locked implementation decisions

1. The twelve ordinary HTTP reads become bounded `rest_read` operations with fixed paths,
   `json_redacted` output, exact path-variable flags, and generated endpoint/output metadata.
2. The six submits/synchronous calls become closed-schema `rest_write` operations; the three
   `DELETE` cancellation calls are status-only, plan/preview/approval-gated direct writes with
   typed destructive confirmation. No generic JSON, body, header, URL, shell, or SQL input is
   exposed.
3. Live Scribe becomes a fixed, declaration-owned WebSocket session operation. It admits exactly
   the documented relative endpoint, `live-asr` subprotocol, closed `session.update` JSON schema,
   PCM16 file input, fixed client frame sequence, finite output and input bounds, a declaration-owned
   capped session lifetime, normal connector authentication/rate-limit admission, and redacted result
   framing. It has no caller-selected
   origin, protocol, header, arbitrary initial frame, arbitrary frame type, or raw HTTP escape.
4. The WebSocket runtime/schema/CLI route is reusable shared runtime work. The connector-lane
   ownership contract therefore required foundation [#3963](https://github.com/polymetrics-ai/cli/issues/3963),
   now present in stacked PR #3965, before this AI Services bundle declaration proceeds. Its closed
   schema acceptance test is green on the consumer base. It uses the Go standard library and adds no
   dependency; the consumer must stay closed enough to make the documented Zoom operation executable
   without exposing caller-selected transport, origin, protocol, header, or frame controls.
5. `surface-sync`, `surface-reconcile`, documentation/manual, and website catalogs are generated
   normally. The recorded mechanical retention trace may restore non-Zoom aggregate catalog rows;
   generated Zoom data is never hand-merged.

## TDD execution

1. **Plan checkpoint** — retain this fresh provenance, exact source list, source-to-ledger audit,
   GSD fallback, WebSocket design constraints, and target accounting before bundle changes.
2. **RED checkpoint** — add only the AI Services command-surface contract and target-count bumps;
   run it against CRC-complete HEAD and commit/push the captured failure before declaration or
   foundation edits.
3. **Foundation RED/GREEN** — introduce a closed fixed-WebSocket operation contract, schema,
   command preflight/run path, runtime help/manual/website support, and loopback protocol tests in
   separate commits. Test protocol rejection, bounded frames, cancellation, handshake auth,
   status 101, redaction, and no generic transport escape first.
4. **Connector GREEN** — author the 22 operation declarations and CLI commands, reconcile exactly
   the matching ledger rows, mechanically regenerate derived files, and add local fixture lifecycle
   coverage for every HTTP read/write and the live session.
5. **Verify/review** — run every route through a newly built binary, validate generated/docs/site
   locality, execute scoped gates, record inline manual `verify-work` and `code-review`, then
   commit/push the coherent provider-category slice and update/close the child issue.

## Target accounting

| Measure | Before | After |
| --- | ---: | ---: |
| Zoom executable endpoints | 143 | 165 |
| Zoom-local implementable rows | 1,699 | 1,677 |
| Conventional direct reads | 70 | 82 |
| Direct writes | 69 | 78 |
| Fixed WebSocket sessions | 0 | 1 |
| Binary downloads | 1 | 1 |
| Reverse-ETL writes | 2 | 2 |
| `unsafe_or_disallowed` Zoom rows | 0 | 0 |

## Verification plan

- Real `commandrunner.Preflight` coverage for all 22 exact `ai-services …` command paths.
- Loopback fixture coverage for 12 REST reads, six JSON writes, three destructive status-only
  deletes, and the fixed Live Scribe handshake/session framing; outputs are redacted and fixtures
  contain only synthetic data.
- A freshly built binary must run `pm help zoom`, bare `pm zoom`, bare `pm zoom ai-services`, and
  every exact route's `--help` in bounded batches.
- `go run ./cmd/connectorgen surface-sync --check`, full connector validation, and scoped
  `surface-reconcile --check --notes-contains provider_module=ai-services`.
- Scoped tests, vet, build, individual non-full-suite `make` gates, docs generator/check, website
  typecheck, endpoint-ledger delta locality, and website catalog locality. CI owns the full suite.

## Canonical references

- `AGENTS.md`
- `docs/migration/HANDOFF-CODEX.md`
- `docs/migration/conventions.md`
- `docs/architecture/connector-architecture-v2-design.md`
- `.agents/agentic-delivery/contracts/issue-agent-contract.md`
- `.agents/agentic-delivery/contracts/parent-orchestrator-contract.md`
- `.agents/agentic-delivery/references/gsd-pi-adapter.md`
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`
- `https://developers.zoom.us/docs/api/ai-services.md`
