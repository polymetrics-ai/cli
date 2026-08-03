# Issue tree draft — engine request-shaping foundations

**Status: CREATED.** Parent #3686; sub-issues #3690 (array cardinality, unblocks 25), #3691
(start_index pagination, unblocks 2), #3692 (required_query, unblocks 1), #3693 (bounded base64
upload, unblocks 1). All four are linked as sub-issues of #3686.

The captain authorised a bounded exception after this file was first written: because no Alfred
credential exists, the tree was created under `karthik-sivadas`. The exception covers **issue
creation only** — approvals, merges and branch-protection changes remain the captain's, and
firstmate still owns every merge. The original blocker analysis is kept below for the record.

Issue and PR creation is supposed to use the Alfred account (`alfred-polymetrics-ai`). No Alfred
credential exists in this environment — verified directly in this lane, not assumed:

```
$ gh auth status
  ✓ Logged in to github.com account karthik-sivadas
$ env | grep -iE '^(GH_TOKEN|GITHUB_TOKEN|ALFRED)'   # empty
$ grep -c 'user:' ~/.config/gh/hosts.yml             # 1  (karthik-sivadas)
```

GitHub authorship is immutable — an issue created as `karthik-sivadas` cannot be reattributed to
Alfred afterwards. Per the captain's standing order these bodies are drafted here and **not**
created. Sibling lane `cli-engine-bounded-binary-read-r1` hit the identical block and was directed
to do the same (tracked as `cli-pipeline-alfred-identity-r1`).

Each body below is ready to paste into `gh issue create` verbatim.

---

## PARENT

**Title:** `feat(engine): request-shaping foundations for documented array, pagination, query and upload constraints`

**Labels:** `connector-engine`, `foundation`

**Body:**

Airtable's operation ledger (`internal/connectors/defs/airtable/api_surface.json`, on branch
`fm/cli-airtable-parity-wave03-r1`) has 30 blocked endpoints. 29 of them name one of four missing
shared-engine capabilities in their own `operation.reason` text. None of the four is
Airtable-specific — each is a request-shaping rule the declarative engine cannot express today, and
at least two are already documented as known limitations by *other* connectors.

Counted directly from the ledger:

| Dependency named in the ledger | Blocked operations |
| --- | ---: |
| `airtable-array-cardinality-foundation` | 25 |
| `airtable-scim-pagination-foundation` | 2 |
| `airtable-required-query-foundation` | 1 |
| `airtable-bounded-base64-upload-foundation` | 1 |
| (not a foundation — CSV body, deliberately disallowed) | 1 |

This parent tracks the four shared capabilities. **No connector bundle is modified by this work** —
the path-ownership guardrail is live on `main` and would reject bundle edits from a foundation
branch. Airtable, and any other connector with the same need, adopts these in its own lane
afterwards.

### Sub-issues

- [ ] #3690 — array cardinality (`minItems`/`maxItems`) in the engine schema dialect — unblocks 25
- [ ] #3691 — `start_index` pagination strategy (SCIM RFC 7644 §3.4.2.4) — unblocks 2
- [ ] #3692 — `required_query` any-of constraint on operations — unblocks 1
- [ ] #3693 — bounded base64 upload write body — unblocks 1

### Standing constraints

- No raw request escape hatch. Every capability stays typed and inspectable; none of them lets a
  caller supply arbitrary method, path, or body structure.
- Bounds, containment and safety rules for the upload direction are kept consistent with the
  download direction being built in `cli-engine-bounded-binary-read-r1` (`os.Root` containment,
  read one byte past the limit and reject, clamp request → spec → ceiling, SHA-256 during the
  copy). See `docs/decisions` research note dated 2026-08-04.

---

## SUB 1 — array cardinality

**Title:** `feat(engine): enforce minItems/maxItems in the connector schema dialect`

**Parent:** parent issue above

**Body:**

### What is blocked

**25 Airtable operations**, every one of which carries this reason verbatim in
`internal/connectors/defs/airtable/api_surface.json`:

> blocked until `airtable-array-cardinality-foundation` can enforce non-empty documented request
> arrays without shared minItems/runtime changes

The full list (method + path), from the ledger:

```
POST   /scim/v2/Groups
PATCH  /scim/v2/Groups/{groupId}
PUT    /scim/v2/Groups/{groupId}
POST   /scim/v2/Users
PATCH  /scim/v2/Users/{userId}
PUT    /scim/v2/Users/{userId}
POST   /v0/meta/bases
POST   /v0/meta/bases/{baseId}/collaborators
POST   /v0/meta/bases/{baseId}/interfaces/{pageBundleId}/collaborators
POST   /v0/meta/bases/{baseId}/tables
POST   /v0/meta/enterpriseAccounts/{enterpriseAccountId}/moveGroups
POST   /v0/meta/enterpriseAccounts/{enterpriseAccountId}/moveWorkspaces
POST   /v0/meta/enterpriseAccounts/{enterpriseAccountId}/personalAccessTokens/revoke
PATCH  /v0/meta/enterpriseAccounts/{enterpriseAccountId}/users
POST   /v0/meta/enterpriseAccounts/{enterpriseAccountId}/users/claim
POST   /v0/meta/enterpriseAccounts/{enterpriseAccountId}/users/grantAdminAccess
POST   /v0/meta/enterpriseAccounts/{enterpriseAccountId}/users/revokeAdminAccess
POST   /v0/meta/enterpriseAccounts/{enterpriseAccountId}/workspaceAiAllowlist
POST   /v0/meta/workspaces/{workspaceId}/collaborators
DELETE /v0/{baseId}/{tableIdOrName}
PATCH  /v0/{baseId}/{tableIdOrName}
POST   /v0/{baseId}/{tableIdOrName}
PUT    /v0/{baseId}/{tableIdOrName}
POST   /v0/{enterpriseAccountId}/{dataTableId}/deleteRecords
PUT    /v0/{enterpriseAccountId}/{dataTableId}/upsertRecords
```

### Why it is not Airtable-shaped

Two connectors already ship a written apology for this exact gap:

- `internal/connectors/defs/drip/writes.json` — *"the engine's schema dialect has no
  minItems/maxItems keyword to mechanically cap it to exactly one element, so callers are expected
  to supply exactly one"*
- `internal/connectors/defs/zoho-bigin/writes.json` — *"the engine's draft-07 subset has no
  minItems/maxItems keyword to enforce cardinality here, see docs.md Known limits"*

### The gap

`internal/connectors/engine/schema.go` implements a deliberately minimal draft-07 subset. Unknown
keywords are a **compile error** (`schema.go:105`), so a bundle that writes `"minItems": 1` today
fails to load at all. There is no way to say "this array is required and must contain at least one
element", so a connector cannot declare such an operation executable without risking a malformed
request.

### Acceptance

- `minItems` and `maxItems` are accepted structural keywords in the engine dialect.
- Both are enforced only against array instances, per draft-07 applicability — "required and
  non-empty" is `required` + `minItems: 1`, exactly as it is in real draft-07.
- Compile-time rejection of a negative bound, a non-integer bound, and `maxItems < minItems`.
- Enforcement reaches every existing compile site with no per-site change: write `record_schema`,
  `json_array` `body_schema`, operation `rest.body_schema`, stream schemas, and `spec.json`.
- Error text names the JSON-pointer-ish path, matching the dialect's existing convention.
- No connector bundle is edited in this issue.

---

## SUB 2 — SCIM startIndex pagination

**Title:** `feat(engine): add start_index pagination strategy (SCIM RFC 7644)`

**Parent:** parent issue above

**Body:**

### What is blocked

**2 Airtable operations**, both `model: direct_read`:

```
GET /scim/v2/Groups
GET /scim/v2/Users
```

Reason, verbatim from `internal/connectors/defs/airtable/api_surface.json`:

> blocked until `airtable-scim-pagination-foundation` can derive SCIM startIndex pagination from
> Resources, startIndex, itemsPerPage, and totalResults without using a nonexistent
> nextStartIndex token

### The gap

`internal/connectors/engine/paginate.go` implements six strategies — `none`, `link_header`,
`page_number`, `offset_limit`, `cursor` (two token sources), `next_url`. None of them models SCIM
2.0 list pagination, which is 1-based and carries its own total:

- request: `startIndex` (1-based), `count`
- response: `{ "totalResults": N, "itemsPerPage": M, "startIndex": S, "Resources": [...] }`

There is no next-page token to read, so `cursor` cannot express it; `offset_limit` is 0-based and
has no total to stop against, so it would either over- or under-run the last page.

### Acceptance

- New pagination `type: "start_index"`, sitting alongside the existing strategies in
  `newPaginator`, not special-cased anywhere.
- Named generically (any 1-based `startIndex` + total API), with SCIM's own names as the defaults
  so a SCIM stream declares only `{"type": "start_index", "page_size": N}`.
- The next index is derived from the records actually extracted this page, not from a claimed
  `itemsPerPage` — a server that lies about the page size cannot desynchronise the walk.
- Stops on: zero records this page, `startIndex + count > totalResults`, and a non-advancing
  index. The non-advancing case is a sticky `Err()`, matching `nextURL`/`tokenPathCursor`, so a
  hostile or buggy API cannot loop pagination forever.
- Declared in `schema/streams.schema.json` in both the `base` and per-stream `pagination` blocks
  (they are duplicated today).
- No connector bundle is edited in this issue.

---

## SUB 3 — required query parameters

**Title:** `feat(engine): express required-any query parameter groups on operations`

**Parent:** parent issue above

**Body:**

### What is blocked

**1 Airtable operation**, `model: direct_read`:

```
GET /v0/meta/enterpriseAccounts/{enterpriseAccountId}/users
```

Reason, verbatim from `internal/connectors/defs/airtable/api_surface.json`:

> blocked until `airtable-required-query-foundation` can require at least one documented email[] or
> id[] query value without claiming an unfiltered executable stream

### The gap

`OperationDirectRead` (`internal/connectors/engine/direct_read.go:67-77`) merges the operation's
declared `rest.query` with the caller's `req.Query` and issues the request. There is no way to
declare that the merged result must contain at least one of a named set. An endpoint that returns
an error (or, worse, an unbounded enterprise-wide listing) when unfiltered therefore cannot be
declared executable.

### Acceptance

- `rest.required_query` on an operation: a list of groups, each `{"any_of": [names...]}`, all of
  which must be satisfied. One group expresses Airtable's case; multiple groups express
  "at least one of A **and** at least one of B", which is a real API shape and costs nothing.
- A parameter counts as present only when its value is non-empty, and a value hardcoded in the
  operation's own `rest.query` satisfies the requirement — the constraint is about the wire
  request, not about who supplied it.
- Enforced before the request is issued, with an error naming the group.
- Bundle-load validation rejects an empty group and an empty parameter name.
- Generic: nothing in the rule mentions email, id, or Airtable.
- No connector bundle is edited in this issue.

---

## SUB 4 — bounded base64 upload

**Title:** `feat(engine): add bounded base64 upload write body type`

**Parent:** parent issue above

**Body:**

### What is blocked

**1 Airtable operation**, `model: sensitive_reverse_etl`:

```
POST /v0/{baseId}/{recordId}/{attachmentFieldIdOrName}/uploadAttachment
```

Reason, verbatim from `internal/connectors/defs/airtable/api_surface.json`:

> blocked on `airtable-bounded-base64-upload-foundation` until an Airtable-owned executor validates
> official base64 encoding and decoded-size bounds before transmission

### The gap

`internal/connectors/engine/write.go` supports `json`, `form`, `none`, `graphql`, `json_array` and
`multipart` bodies. An API that wants a base64-encoded payload **inside a JSON body** has no typed
route: the only way to produce one today would be to let the caller hand the engine a raw body,
which is the one thing that is banned outright.

### Acceptance

- New `body_type: "base64_upload"` with a small declared spec: which record field is the source,
  which JSON body property receives the encoded content, and the decoded-size bound.
- Two source modes, both ending in the same validated base64 string:
  - `path` — a local file, read under `os.Root` containment (closes traversal, symlink escape and
    the TOCTOU race in one primitive), bounded read of one byte past the limit, rejected on
    overflow.
  - `base64` — an already-encoded string, decoded with `base64.StdEncoding.Strict()`, which is
    what "official base64" means: canonical padding, no line breaks, no URL-safe alphabet.
- The source field never reaches the wire. A local filesystem path in particular must not be
  transmitted to the provider.
- Bounds are consistent with the download direction (`cli-engine-bounded-binary-read-r1`): clamp
  against a hard engine ceiling, reject rather than truncate, digest during the read.
- Honours the existing approved-payload digest contract (`cfg.ApprovedPayloadSHA256`) exactly as
  the multipart file part does, so plan → preview → approve → execute still binds the bytes that
  were approved.
- No raw method, path or body structure is reachable from a record.
- No connector bundle is edited in this issue.
