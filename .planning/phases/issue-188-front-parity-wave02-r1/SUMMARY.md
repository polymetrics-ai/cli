# Summary — issue #188 Front parity

Status: implemented locally; commit pending.

## Delivered

- Expanded `internal/connectors/defs/front/api_surface.json` to a 255-row operation-ledger v1 sourced from Front `llms.txt` and linked `reference/*.md` OpenAPI snippets.
- Added 115 Front read streams and generated permissive schemas while preserving the six legacy fixture-backed stream names.
- Added 119 fixed write actions with record schemas, path fields, redaction, risk text, `confirm: "destructive"`, and idempotent DELETE 404 handling.
- Added 119 sanitized write fixtures plus connector-local upload placeholder fixtures for multipart paths.
- Added `operations.json` for direct/provider-search reads, binary-download metadata, not-applicable/disallowed surfaces, and one duplicate docs row.
- Added `cli_surface.json` with 255 provider-style connector command metadata entries.
- Added `certification.json` that truthfully records fixture-only status; live Front certification remains unavailable and was not attempted.
- Updated Front `docs.md` and generated CLI golden transcripts for root help discovery.
- Added/updated GSD plan, TDD ledger, verification checklist, run state, and GSD prompt trace.
- Appended the idempotent captain-policy addendum to #188 and #189-#195 via `gh-axi`, preserving existing body text and count tables.

## Count reconciliation

- 255 documented operation rows total.
- 254 unique method/path pairs; `PATCH /channels/{channel_id}` appears in two Front docs pages and is recorded once as executable and once as a duplicate ledger row.
- Executable/fixed metadata split: 115 streams, 119 writes, 5 direct reads, 5 planned binary downloads, 10 disallowed/not-applicable rows, 1 duplicate row.
- Writes include 118 reverse-ETL operations plus the application event trigger surface; all 119 require destructive confirmation.

## Verification

Passed:

```bash
go run ./cmd/connectorgen validate
go test ./internal/connectors/conformance -run 'TestConformance/front' -count=1
go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1 -timeout=8m
go vet ./internal/connectors/... ./internal/cli/...
go build ./cmd/pm
make connector-boundary
git diff --check
```

Also passed targeted temp-root validation for only the Front bundle.

## Safety notes

No live Front calls, credentials, provider writes, reverse ETL execution, new dependencies, shared runtime/foundation edits, no-mistakes, push, PR, merge, VPS, or Thaalam work were performed.

Project-memory hook note: `/Users/karthiksivadas/karthik-agent-workspace/bin/fm-ensure-agents-md.sh .` was run as requested and returned `conflict: both AGENTS.md and CLAUDE.md are real files`; no project memory files were edited.

## Increment 2 — close the 94% → 97.6% CLI-reachability gap (updated bar)

Resumed after the `cli-inflight-connector-parity-audit-r1` measured this branch's ref at
239/255 (94%) CLI-reachable, with 16 documented endpoints undispositioned-as-executable:
5 binary attachment downloads (`binary_read`, genuinely blocked — no shared bounded
file-output execution contract exists in `internal/connectors/engine` today, confirmed by
code search: no executor reads `operations.json`'s `binary_download`/`op.Binary` kind
anywhere, and `output_policy: "binary_bounded_redacted"` is not in
`commandrunner/runner.go`'s `isSupportedDirectReadOutputPolicy`), 1 duplicate
`PATCH /channels/{channel_id}` documentation row (no new work needed), and 10
"application-channel/voice/plugin-adjacent... not applicable" rows previously marked
`operation.model: "disallowed"`.

Verified against Front's own public API reference docs (`dev.frontapp.com/reference/*`)
that 9 of those 10 rows carry no partner-only/marketplace-approval eligibility restriction
(only ordinary bearer-token scopes: `channels:write`, `messages:write`,
`message_templates:write`) and the connector already implements the surrounding custom
channel lifecycle (`create_a_channel`, `update_channel`, `validate_channel`) — so
"not applicable to this connector" was an over-classification, exactly the
deferred-but-buildable trap the updated task brief warns against. The 10th
(`add-call-recording`) requires a `multipart/form-data` file upload, which the engine DOES
already support (`body_type: "multipart"`, proven in production by `gong`'s
`upload_call_media`/`upload_crm_entities` actions) — also buildable, not blocked.

Implemented all 10 as real `writes.json` actions (`create_call`, `update_call`,
`add_call_recording` [multipart], `add_call_summary`, `add_call_transcript`,
`sync_application_message_template`, `update_application_message_template`,
`sync_inbound_message`, `sync_outbound_message`, `update_external_message_status`), each
behind the same plan → preview → explicit approval → execute path with
`confirm: "destructive"` as every other Front write. Added 9 `fixtures/writes/*.json`
dynamic-replay fixtures (all except the multipart action, which — like `gong`'s two
multipart writes — has no dynamic fixture since no bundle in this repo yet exercises
multipart write-shape conformance; its coverage rests on `connectorgen validate`'s static
schema/CLI-mapping checks). Flipped `api_surface.json`'s `covered_by` for all 10 endpoints,
removed their now-superseded `operations.json` blocked-ledger rows (kept the genuine
`update_channel` duplicate-documentation row), and replaced their `cli_surface.json`
`excluded`/`operation`-based command rows with `implemented`/`write`-based commands
(net command count unchanged: still 255).

Result: 249/255 (97.6%) documented operations executable as individual `pm front <command>`
invocations; the remaining 6 are honestly dispositioned (5 blocked with a named shared-
foundation dependency, 1 duplicate row) — zero undispositioned operations.

Also fixed two pre-existing `cli_surface.json` findings surfaced by a stricter
`cli_surface_safety` validate rule that landed on `main` after this branch's original
commit (`analytics reports create`/`analytics exports create` flags needed
`"required": true` to match their operation `body_schema`'s required fields) — unrelated to
the 10 new writes but blocking `connectorgen validate` after rebasing onto current `main`.

Rebased onto `origin/main` (9 commits: ashby/amazon-sqs staging, issueguard fix, Google Ads
v22, Zendesk Support, Asana, Xero, Freshchat, Bitbucket parity waves). Only conflict was a
shared generated golden-CLI-transcript fixture (`internal/cli/testdata/golden_transcripts.json`,
unrelated to Front), resolved by taking `main`'s regenerated version.

Updated `docs.md` (Overview counts, Write actions & risks, Known limits) for the new
disposition. Regenerated `docs/connectors/front/{MANUAL,SKILL}.md` via
`pm docs generate --dir docs/cli` and the website catalog via `pnpm run gen:website-data`;
both regenerations touch every connector's generated artifact by nature of being full-catalog
regenerators, so the diff was scoped surgically (byte-range splice of just the `front` entry/
row) to genuine Front-only changes in `docs/connectors/{README.md,catalog/all-connectors.{md,json}}`
and `website/{data/connectors.generated.json,lib/connectors.catalog.{generated.ts,data.generated.json}}`,
reverting unrelated drift discovered in unrelated connectors (e.g. `ashby`) caused by
generator-version skew already present on `main` before this rebase.
