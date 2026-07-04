# Overview

Gorgias is a Tier-1 declarative-HTTP wave2 fan-out migration. It reads Gorgias helpdesk tickets,
customers, messages, and satisfaction surveys through the Gorgias REST API. This bundle migrates
`internal/connectors/gorgias` (the hand-written connector); the legacy package stays registered and
unchanged until wave6's registry flip.

## Auth setup

Provide the Gorgias account email as `username` and the Gorgias API key as the `password` secret;
both flow into HTTP Basic auth (`Authorization: Basic base64(username:password)`) and the password
is never logged. `base_url` is required and must be the full API root, e.g.
`https://<domain>.gorgias.com/api`.

Legacy also accepted a bare `domain_name` (e.g. `acme`) and derived
`https://<domain>.gorgias.com/api` from it in code. The engine's `spec.json` `"default"` mechanism
only materializes a FIXED literal default, not one derived from another config value at read time,
so this bundle only accepts a fully-formed `base_url` — the `domain_name`-derivation convenience is
dropped as documented config-surface narrowing (never a data-parity change; once `base_url` is
correctly configured, every stream behaves identically to legacy).

## Streams notes

All 4 streams (`tickets`, `customers`, `messages`, `satisfaction-surveys`) share the identical
shape: `GET`, `records.path: "data"`, and the base's `cursor` pagination (`cursor_param: cursor`,
`token_path: meta.next_cursor`) — Gorgias's own `{"data":[...],"meta":{"next_cursor":"..."}}`
envelope. No `stop_path` is declared: Gorgias's only stop signal is `next_cursor` being absent/null,
which is the paginator's default stop-on-empty-token behavior; legacy also independently stops when
a page returns zero records, which the same default handles (an empty page combined with an absent
`next_cursor` is the terminal case). `limit` sends `config.page_size` (default `100`, matching
legacy's `gorgiasDefaultPageSize`).

**No `incremental` block is declared for any stream, matching legacy's real behavior exactly.**
Legacy's live `Read`/`harvest` path never sends a date-filter query parameter at all — every sync,
first or repeat, requests the exact same unfiltered list. `x-cursor-field` is still declared on each
schema (`updated_datetime`/`updated_datetime`/`created_datetime`/`created_datetime`, matching
legacy's catalog `CursorFields`) as informational catalog metadata; no `streams.json` `incremental`
block accompanies it, so no `incremental_append` sync mode is derived for these streams (by design —
see `conventions.md` §2's sync-mode-derivation rule). Declaring `client_filtered: true` was
considered and rejected: that mechanism drops already-seen records client-side by comparing against
a persisted lower bound, which is still MORE filtering than legacy performs (legacy has no
lower-bound concept whatsoever) and would silently diverge from parity.

## Write actions & risks

None. Gorgias is a read-only source in this connector (legacy `Capabilities.Write` is `false`); no
`writes.json` file is present.

## Known limits

- The `satisfaction-surveys` stream is renamed to `satisfaction_surveys` (underscore, not hyphen).
  The engine's `namePattern`/stream-name validation (`^[a-z][a-z0-9_]*$`, `conventions.md` §2)
  requires snake_case for every stream name; legacy's own catalog key used a hyphen. This is a
  catalog/identifier-naming-only change — the underlying endpoint (`GET /satisfaction-surveys`) and
  every emitted field are unchanged.

- Full Gorgias API surface (ticket create/update, tags, teams, views, macros, integrations, rules,
  etc.) is out of scope for wave2; see `api_surface.json`'s `excluded: {category: out_of_scope,
  reason: "not implemented in this bundle"}` entries. Only the 4 legacy-parity read streams are
  implemented.
- The `domain_name`-derived base URL convenience is dropped (see Auth setup above); `base_url` must
  be the fully-formed API root.
- No incremental sync mode is derived for any stream, matching legacy's real unfiltered-every-sync
  behavior — see Streams notes above for why the engine's `incremental`/`client_filtered` mechanisms
  were deliberately not used here.
