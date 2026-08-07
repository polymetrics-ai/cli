# Gmail documented-operation parity — plan

Part of `cli-top50-fixed-schema-sweep-r1`. One connector, one PR.

## Goal

Bring `internal/connectors/defs/gmail` from **45 covered / 34 legacy-`excluded`** to a surface where
all **79** documented operations carry a real disposition (`executable`,
`blocked-with-named-dependency`, or `unsupported-with-source-citation` — **never `excluded`**) and
every executable one is individually reachable as `pm gmail <command>`.

## Operation surface, derived before authoring

Artifact: Google's official Discovery document
`https://gmail.googleapis.com/$discovery/rest?version=v1` — HTTP 200, 217,687 bytes,
`kind: discovery#restDescription`, `id: gmail:v1`, **`revision: 20260803`**, fetched 2026-08-07.
Public, no auth.

**79 methods: GET 30, POST 28, DELETE 10, PUT 8, PATCH 3.**

Counted by walking `resources.*.methods.*` **recursively** — Gmail nests resources
(`users.messages.attachments`, `users.settings.sendAs.smimeInfo`), so a flat walk undercounts.
`0` top-level `methods` outside `resources`.

**The ledger's 79 reconciles exactly.** The count is not the problem here.

## Baseline

| | State |
| --- | --- |
| `api_surface.json` | 79 rows, method split matches the artifact exactly |
| Dispositions | 10 `covered_by.stream` · 35 `covered_by.write` · **34 `excluded`** · 0 blank |
| `operation_ledger_version` | **unset** — the v2 provenance ledger is missing |
| `cli_surface.json` | **ABSENT** — no direct-read command surface can exist without it |
| `operations.json` | **ABSENT** |
| `streams.json` / `writes.json` | 10 streams / 35 actions |

## Adjudication of the 34 `excluded` rows — the real work

`excluded` is a legacy category that this sweep's bar does not accept. Each row is re-dispositioned
below. **Two of these exclusions directly contradict standing captain rulings.**

| # | Family | Rows | Disposition |
| --- | --- | ---: | --- |
| 1 | Single-resource detail GETs | 8 | **PROMOTE → direct read** |
| 2 | Settings singleton GETs | 5 | **PROMOTE → direct read** |
| 3 | Bulk `batchDelete` / `batchModify` | 2 | **blocked-with-named-dependency** |
| 4 | S/MIME certificate surface | 5 | **PROMOTE — the stated block is false** |
| 5 | CSE (Client-Side Encryption) | 11 | **blocked-with-named-dependency** |
| 6 | `attachments.get` | 1 | **binary download** |
| 7 | `watch` / `stop` | 2 | **PROMOTE — webhook management is in scope** |

### 1 — Detail GETs (8): promote

`drafts/{id}`, `labels/{id}`, `messages/{id}`, `threads/{id}`, `settings/delegates/{delegateEmail}`,
`settings/filters/{id}`, `settings/forwardingAddresses/{forwardingEmail}`,
`settings/sendAs/{sendAsEmail}`.

Current reason is "identical resource shape the list stream already paginates through in full". That
conflates **ETL coverage** with **command reachability**. The brief names *"Direct read — single-resource
reads reachable as their own command"* as its own required scope. Precedent in this repo:
**gong ships `calls get` as a direct read alongside the `calls` ETL stream** — the exact shape.

### 2 — Settings singleton GETs (5): promote

`autoForwarding`, `imap`, `language`, `pop`, `vacation`. Current reason is "reads back the singleton
this bundle already writes". "We already write it" is not a reason it cannot be read. A bounded
get-after-write read is a legitimate direct read.

### 3 — Bulk variants (2): blocked with a NAMED dependency

`messages/batchDelete`, `messages/batchModify` take an `ids[]` array. The engine's write dialect is
one-request-per-record. This is a **real shared-engine gap with an existing named issue —
#514 "WhatsApp: schema-gated top-level JSON array request bodies"**. Disposition
`blocked-with-named-dependency`, dependency **#514**. Not silently excluded, and **not** worked
around by widening the engine from inside a connector PR.

### 4 — S/MIME (5): the stated block is FALSE — promote

`sendAs/{}/smimeInfo` (GET/POST), `smimeInfo/{id}` (GET/DELETE), `smimeInfo/{id}/setDefault` (POST).

The exclusion reason says S/MIME "requires the narrow `https://mail.google.com/` full-mailbox scope
beyond gmail.settings.basic/gmail.readonly". **gmail's own `spec.json` already declares
`https://mail.google.com/`**, and documents why: *"since this bundle now declares mutating write
actions (send/modify/trash/delete/settings) that gmail.readonly cannot authorize"*.

So the bundle already holds the scope the exclusion claims it lacks. The stated dependency does not
exist. These 5 are ordinary typed operations — 2 reads, 3 writes, one of which
(`smimeInfo` POST) carries a certificate blob and needs the bounded-input/redaction treatment
`upload_call_media` already models in gong.

### 5 — CSE (11): genuinely blocked, named dependency

9 `requires_elevated_scope` + `keypairs` POST (`binary_payload`) + `keypairs/{id}:obliterate`
(`destructive_admin`).

Unlike S/MIME, CSE is **not** a scope problem — it is a **Google Workspace Enterprise Plus /
Education Plus add-on gated behind organization-level admin enablement**. No OAuth scope the
connector can declare unlocks it. That is a real, external, **named** dependency and the correct
disposition is `blocked-with-named-dependency` citing Google's CSE documentation.

**But note `obliterate` is currently excluded as `destructive_admin`, which contradicts the
captain's standing policy** (recorded on the gong issues, 2026-07-30): *"every documented operation
is in scope for connector parity, including DELETE, destructive, admin, file-upload … modeled only
as fixed-target typed connector operations with schema-bounded inputs, redaction, typed `destructive`
confirmation where applicable, and the existing reverse ETL plan → preview → approval → execute
safety path."* **Risk is not a reason to exclude.** It is blocked on the CSE entitlement, and must
say so — not "destructive_admin".

### 6 — `attachments.get` (1): the binary download scope

Returns base64url-encoded attachment bytes. The brief names **binary** as a required scope
("uploads and downloads through the existing engine capability, not ad-hoc HTTP"). The engine
declares `base64_upload` (`engine.Base64UploadSpec`); **whether a matching bounded *download*
capability exists must be checked before authoring**. If it exists → executable. If not →
`blocked-with-named-dependency` naming that missing engine capability. **Do not invent one, and do
not mark it implemented unless the command genuinely runs.**

### 7 — `watch` / `stop` (2): webhook MANAGEMENT — promote

Currently excluded as `non_data_endpoint` ("registers a Cloud Pub/Sub push-notification
subscription; a control-plane side effect").

**This contradicts the captain's standing webhook ruling** (finding F2, point 5): *"Webhook
**management** endpoints stay fully in scope. Create/list/update/delete of a webhook **subscription**
is an ordinary REST operation, is already inside the counts, and is **not** deferred. Only webhook
**events** are deferred."*

`watch()` registers a push subscription and `stop()` cancels it — that is subscription lifecycle
management, precisely what stays in scope. Promote as typed writes.

## Planned partition after the change

| Bucket | Now | Planned |
| --- | ---: | ---: |
| ETL streams | 10 | 10 |
| Direct reads | 0 | **~15** (8 detail + 5 settings + 2 S/MIME reads) |
| Writes | 35 | **~40** (+3 S/MIME writes, +2 watch/stop) |
| Binary | 0 | **1** (`attachments.get`), pending the capability check |
| Blocked with named dependency | 0 | **13** (11 CSE + 2 bulk-array on #514) |
| Legacy `excluded` | 34 | **0** |
| **Total** | **79** | **79** |

Bucket sizes are planned, not asserted. The red test asserts the **derived** total and the
zero-`excluded` / zero-blank partition; bucket counts settle during authoring.

## Webhook inventory (input for `cli-webhook-surface-sweep-r1`)

Gmail's Discovery document declares **no webhook event catalogue** — push notifications are
delivered via Cloud Pub/Sub, and the *events* are Pub/Sub messages, not documented API operations.
**0 webhook events.** The 2 webhook **management** operations (`watch`, `stop`) are in scope and
counted, per the ruling above.

## TDD sequence

1. **RED** — add `cmd/connectorgen/gmail_api_surface_test.go` asserting 79 rows, the method split,
   `operation_ledger_version` set, **zero `excluded` rows**, zero blank dispositions, and no
   duplicate endpoint key. Must fail against today's 34-excluded/unset-ledger bundle.
2. **GREEN** — author `cli_surface.json` and `operations.json` from scratch, re-disposition all 34.
3. **REFACTOR** — docs, catalogs, ledger resync.
4. Gates, then no-mistakes.

**Check for a second surface test before authoring (finding F5)** —
`ls cmd/connectorgen/ | grep '^gmail_'` and run the whole `cmd/connectorgen` package, never a
targeted `-run`.

## Safety notes

- Do not loosen `connectorgen validate`, the connector boundary gate, `certify`, or
  `TestEveryImplementedCommandPassesRuntimePreflight` to make this pass.
- Nothing is marked `implemented` unless its command runs; every block carries a **named** dependency.
- Do not widen the shared engine from inside this connector PR — the array-body gap belongs to #514.
- No credential or token-derived value is ever emitted.
- Keep the diff scoped to gmail; revert the repo-wide docs generator drift (finding F4).
