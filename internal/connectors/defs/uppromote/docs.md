# Overview

UpPromote reads affiliates from the UpPromote affiliate marketing API (`GET
{base_url}/api/affiliates`). This bundle migrates the hand-written `internal/connectors/uppromote`
legacy package to a declarative Tier-1 defs bundle at capability parity; the legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

Requires one secret: `api_key` (UpPromote API key), sent as a Bearer token on every request via
`streams.json` `base.auth`'s `bearer` mode — matching legacy's `connsdk.Bearer(apiKey)` exactly.
`base_url` defaults to `https://api.uppromote.com` (legacy's `defaultBaseURL`), overridable for
tests or proxies.

## Streams notes

One stream: `affiliates` (`GET api/affiliates`, records at `affiliates`). No pagination — legacy's
`Read` issues a single unpaginated request and emits every record from the one response, so this
bundle declares no `pagination` block either (parity, not an omission).

The optional `start_date` config value is sent as the `start_date` query parameter only when set,
using the opt-in optional-query object dialect (`{"template": "{{ config.start_date }}",
"omit_when_absent": true}`) — matching legacy's own conditional `query.Set("start_date", start)`
(only called `if start := strings.TrimSpace(req.Config.Config["start_date"]); start != ""`).
`start_date` is a plain static passthrough config value, not a stateful incremental cursor: legacy
never tracks or advances a read cursor across syncs, so this bundle deliberately declares no
`incremental` block on the stream (declaring one would add cursor-based `incremental_append` sync
modes legacy never supported — an accepted-input-behavior change the migration meta-rule forbids).
The schema still declares `x-cursor-field: created_at` for manifest-surface parity with legacy's
`CursorFields: []string{"created_at"}` (`uppromote.go:128`) — this is purely descriptive metadata;
sync-mode derivation is gated on the stream's `incremental` block, not on `x-cursor-field` alone, so
no new sync mode is introduced by declaring it.

## Write actions & risks

None. This is a read-only connector; `capabilities.write` is `false` and this bundle ships no
`writes.json`, matching legacy's `Write` stub (`connectors.ErrUnsupportedOperation`).

## Known limits

None beyond the `start_date`/`incremental` framing above: every legacy read/config/auth behavior is
reproduced exactly (single unpaginated request, optional `start_date` filter, Bearer auth, base URL
override).
