# Phase: engine-request-shaping-foundations-r1

**GSD command:** `/gsd-plan-phase engine-request-shaping-foundations-r1`, generated through the
repo-local Pi adapter with `scripts/gsd prompt plan-phase engine-request-shaping-foundations-r1`
(`scripts/gsd doctor` green, 69 commands registered). The generated prompt is recorded at
`.planning/traces/gsd-plan-phase-engine-request-shaping-foundations-r1-prompt.md`.

**Runtime fallback:** the adapter's `/gsd-plan-phase` expects Pi runtime subagents. This session runs
in Claude Code, so the documented inline/manual fallback was taken: the official workflow was
executed inline by the single session agent instead of via spawned Pi researcher/planner/checker
subagents. AGENTS.md permits this and requires it be recorded — it is recorded here. The TDD
lifecycle is **not** waived: every behaviour-adding task starts red, and the red/green evidence is
in `TDD-LEDGER.md`.

**Required skills loaded** (per `.agents/agentic-delivery/references/required-skills-routing.md`,
"Connector runtime and architecture"): `golang-how-to` (orchestrator), `golang-security`,
`golang-safety`. Applied concretely: `os.Root` for path containment (the skill's stated Go 1.24+
answer to path traversal — "never rely on `filepath.Clean` + `strings.HasPrefix` alone"), bounded
reads that reject rather than truncate, explicit numeric bounds and non-negative bound checks,
fail-closed stop conditions in pagination, `defer`-in-loop avoidance, and no nil-map writes.

## Scope

Foundation phase. Four shared request-shaping capabilities that connector lanes cannot add
themselves. All four are strictly additive and opt-in per bundle: a bundle that declares none of them
behaves byte-for-byte as it does today.

**Paths owned by this phase:** `internal/connectors/engine/**` and its schemas plus tests.
**No connector bundle is modified** — the path-ownership guardrail is live on `main` and would
reject bundle edits from a foundation branch. `internal/connectors/connsdk/**`,
`internal/connectors/certify/**`, `internal/connectors/commandrunner/**`, `internal/app/**` and
`cmd/**` are out of scope.

**Known parallel lane:** `cli-engine-bounded-binary-read-r1` is building the download direction on
branch `engine-shared-capabilities-r1` and also edits `internal/connectors/engine/**` (its tasks:
`connsdk` streaming, a binary-download executor, write-action query params, dynamic-key bodies).
File-level overlap is expected in `bundle.go`, `write.go` and `schema/writes.schema.json`; the two
lanes touch disjoint *concerns* inside those files, so conflicts should be additive-adjacent rather
than semantic. This phase reuses that lane's established **rules** (not its code, which is unmerged):
`os.Root` containment, read one past the limit and reject, clamp request → spec → ceiling, digest
during the copy, flat records, never trust a provider-claimed content type.

## Verified starting state

Every claim below was read directly out of this repo, or computed from the Airtable ledger, before
planning.

| Claim | Evidence |
| --- | --- |
| 30 Airtable endpoints are blocked; 29 name one of four foundations | `git show origin/fm/cli-airtable-parity-wave03-r1:internal/connectors/defs/airtable/api_surface.json`, grouped by the `airtable-*-foundation` token in each `operation.reason` |
| Split is 25 / 2 / 1 / 1, plus one non-foundation CSV-body exclusion | same, counted |
| The engine dialect has no `minItems`; unknown keywords are a hard compile error | `engine/schema.go:62-74` (`structuralKeywords`), `schema.go:100-106` |
| The gap is already documented as a limitation by other connectors | `defs/drip/writes.json` risk text; `defs/zoho-bigin/writes.json` `data` description |
| `record_schema` is compiled and validated on every write record | `engine/write.go:42-81` |
| `json_array` `body_schema` is validated per record | `engine/write.go:489-504` |
| Operation `rest.body_schema` is validated for POST direct reads | `engine/direct_read.go:229-247` |
| Six pagination strategies exist; none is 1-based-with-total | `engine/paginate.go:41-78` |
| `newPaginator` returns an error for an unknown type (so a new type is additive and fail-closed) | `paginate.go:75-77` |
| Loop-guard + sticky `Err()` is the established paginator convention | `paginate.go:351-365`, `470-484` |
| Operation query is the merge of `rest.query` and caller query, with no cardinality rule | `engine/direct_read.go:67-77` |
| Write body types are json/form/none/graphql/json_array/multipart | `engine/write.go:400-451` |
| Multipart file parts already bind an approved payload digest | `engine/write.go:534-537`, `internal/app/util.go:190-205` |
| The approval digest is keyed off record fields whose name contains `file_path` | `internal/app/util.go:233-236` |
| `safety.ValidateLocalWritePath` is purely lexical | `internal/safety/safety.go` |
| Go toolchain is 1.25.4 — `os.Root` (1.24+) is available | `go.mod:3` |
| Website CI only triggers on `website/**`, `icon_data.json`, `docs/connectors/icons/**` | `.github/workflows/website.yml:4-9` |
| GSD evidence gate fires whenever `cmd/` or `internal/` changes | `scripts/verify-gsd-workflow` |

## Task 1 — array cardinality: `minItems` / `maxItems` (unblocks 25)

`internal/connectors/engine/schema.go`.

Add `minItems` and `maxItems` to `structuralKeywords` and to `schemaNode`, each with an explicit
`has*` flag so an explicit `0` is distinguishable from absence (the same pointer-vs-zero-value
problem `PaginationSpec.StartPage` already solves one layer up, solved here with a bool because the
node is package-internal).

Enforcement lives in `schemaNode.validate`, in the existing `arrayElements(v)` branch, so it applies
**only to array instances** — which is what draft-07 says and what makes "required and non-empty"
the composition `required: ["x"]` + `properties.x.minItems: 1`. Enforcing on a missing value instead
would silently change the meaning of every existing optional array field.

Compile-time rejection:

- a non-integer bound (`json.Unmarshal` into `int` already fails; the error is wrapped with the
  keyword name to match the file's convention)
- a negative bound — draft-07 requires a non-negative integer
- `maxItems < minItems` when both are declared, which is unsatisfiable and is always an authoring
  mistake rather than an intent

Because both keywords land in the shared dialect, every compile site inherits them with no per-site
change: write `record_schema`, `json_array` `body_schema`, operation `rest.body_schema`, stream
record schemas, `spec.json`. That is what makes one rule unblock 25 operations.

`maxItems` is included rather than deferred because it is the same fifteen lines, it is the other
half of the keyword pair every author will reach for, and there is already an in-repo connector
(`zoho-bigin`, a documented 1–100 cap) waiting for it. It is not speculative surface.

**Deliberately not done:** an `x-` extension keyword. `minItems` is standard draft-07 and is the
literal name the Airtable ledger asks for; inventing a synonym would make bundles non-portable for
no gain.

## Task 2 — `start_index` pagination (unblocks 2)

`internal/connectors/engine/paginate.go`, `bundle.go`, `schema/streams.schema.json`.

New strategy `type: "start_index"`, added to `newPaginator`'s switch alongside the other six.

Named for the mechanism, not the standard: SCIM 2.0 (RFC 7644 §3.4.2.4) is the motivating case and
supplies every default, but any 1-based `startIndex` + total API is served by the same code.
Defaults — `start_index_param: "startIndex"`, `count_param: "count"`,
`total_path: "totalResults"`, `items_per_page_path: "itemsPerPage"`,
`start_index_path: "startIndex"`, first index `1` — mean a SCIM stream declares only
`{"type": "start_index", "page_size": N}`.

Walk:

1. `Start()` requests `startIndex = start_index_base` (default 1) and `count = page_size`.
2. `Next(resp, recordCount)` derives the next index from **`recordCount`** — the records the engine
   actually extracted at the stream's own `records.path` (`Resources` for SCIM) — never from the
   server's claimed `itemsPerPage`. A server that lies about its page size cannot desynchronise the
   walk. `itemsPerPage` is read only as a fallback base when the body echoes no usable
   `startIndex`.
3. Stop conditions, all fail-closed:
   - `recordCount == 0`
   - `totalResults` present and `next > totalResults`
   - the computed next index does not strictly advance → sticky `Err()`, matching
     `tokenPathCursor` and `nextURL`. A non-advancing index is the shape a hostile or buggy API
     would use to loop pagination forever.
4. `page_size <= 0` disables advancement entirely rather than sending `count=0`.

`PaginationSpec` gains `StartIndexParam`, `CountParam`, `StartIndexBase *int`, `TotalPath`,
`ItemsPerPagePath`, `StartIndexPath`. `StartIndexBase` is a pointer for the same reason `StartPage`
is: an explicit `0` (a 0-based server that still reports a total) must be distinguishable from
"absent, default to 1". Both the `base` and per-stream `pagination` blocks in
`schema/streams.schema.json` get the new properties — they are duplicated in that file today, and
leaving one out would make a base-level declaration fail schema validation.

`bundle.go` validation: `start_index` requires a positive effective page size, and rejects a
negative `start_index_base`.

## Task 3 — `required_query` any-of groups (unblocks 1)

`internal/connectors/engine/bundle.go`, `direct_read.go`, `schema/operations.schema.json`.

`RESTOperationSpec` gains:

```json
"required_query": [ { "any_of": ["email", "id"] } ]
```

A list of groups; **every** group must be satisfied by **at least one** of its named parameters. One
group is Airtable's case. Multiple groups express "at least one of A and at least one of B", which
is a real API shape (a required time window plus a required subject filter) and costs nothing beyond
the loop that is already there.

Enforcement in `OperationDirectRead`, immediately after the operation `rest.query` / caller
`req.Query` merge and before any network call. A parameter counts as present only when its value is
non-empty after trimming — an empty string in a query map is exactly the "unfiltered" request the
constraint exists to prevent. A value hardcoded in the operation's own `rest.query` satisfies the
group: the constraint is about the request that goes on the wire, not about who supplied the value.

Error text names the group's parameters, so an operator who omits the filter is told what to add.

Bundle-load validation rejects a group with no `any_of`, and an `any_of` containing a blank name —
both would otherwise be silently unenforceable, which is worse than a hard failure.

**Deliberately not done:** the same constraint on streams. The blocked operation is a direct read,
and its ledger reason explicitly rules out "claiming an unfiltered executable stream" — a stream is
the wrong surface for this endpoint, so adding stream support would be surface nobody asked for.

## Task 4 — `base64_upload` write body (unblocks 1)

`internal/connectors/engine/write.go` (new helpers), `bundle.go`, `schema/writes.schema.json`.

New `body_type: "base64_upload"` with a deliberately small spec:

```json
"body_type": "base64_upload",
"base64_upload": {
  "source": "path",                 // "path" (default) | "base64"
  "source_field": "file_path",
  "content_field": "file",
  "max_decoded_bytes": 3932160,
  "max_encoded_bytes": 5242880      // optional; defaults to the base64 length of max_decoded_bytes
}
```

The body is built by the **existing** rules — `body_fields` if declared, otherwise every record field
minus `path_fields` — and then exactly two things happen: `source_field` is **removed** and
`content_field` is set to the validated base64 string. Everything else (Airtable's `filename` and
`contentType`) is an ordinary record field already governed by the action's closed `record_schema`.
That is what keeps this typed: no record can influence method, path, or body *structure*.

Removing `source_field` from the body is a correctness requirement, not tidiness — in `path` mode it
holds a local filesystem path, and transmitting that to the provider would leak the operator's
directory layout.

Two source modes, converging on the same validated string:

- **`path`** — the file is opened under `os.Root` rooted at the project directory. `os.Root` closes
  path traversal, symlink escape and the stat-then-open TOCTOU race in one primitive;
  `safety.ValidateLocalWritePath` is purely lexical and the existing multipart path bolts
  `EvalSymlinks` on separately, which only works for files that already exist. Absolute paths are
  resolved against, and required to stay inside, the root. The handle must be a regular file. The
  read is `io.LimitReader(f, max+1)` and **rejects** on overflow rather than truncating — a
  truncated attachment is a silently corrupt upload, the same reason the download direction reads
  one past its limit.
- **`base64`** — decoded with `base64.StdEncoding.Strict()`. Strict mode is what "official base64"
  means in the ledger's words: canonical padding, no trailing garbage bits, no embedded newlines, no
  URL-safe alphabet. The decoded length is then bounded identically, and the transmitted string is
  **re-encoded from the decoded bytes** so what goes on the wire is canonical regardless of what
  came in.

Bounds: `max_decoded_bytes` is required and positive at bundle load, and clamped at load time
against a hard engine ceiling of 16 MiB — the same ceiling `maxOperationDirectReadBytes` already
uses for the inbound direction. Encoded size is checked too, because that is the limit real APIs
document (Airtable's is 5 MB *encoded*).

Approval: when `cfg.ApprovedPayloadSHA256` is non-nil the source field must have an approved digest
and the bytes read must match it, mirroring the multipart file part exactly. The digest is computed
upstream for record fields whose name contains `file_path` (`internal/app/util.go:233-236`), so a
bundle naming its field accordingly inherits plan → preview → approve → execute binding with no
change outside the engine. When the map is nil (no approval flow in play, e.g. a direct
`engine.Write` call in a test) the check is skipped, which is the existing multipart contract.

`DryRunWrite` must not read the file: the preview stays a resolved request line, unchanged.

Bundle-load validation: `base64_upload` present iff `body_type` is `base64_upload`; `source_field`
and `content_field` required and non-blank; `source` in `{"", "path", "base64"}`; `max_decoded_bytes`
positive; `max_encoded_bytes`, when declared, at least the encoded length of `max_decoded_bytes`
(otherwise the pair is unsatisfiable).

**Deliberately not done:** an archive/extraction mode, a provider-supplied filename derivation, a
content-type allowlist mechanism (a `record_schema` `enum` already expresses it), and any streaming
upload path. A base64 body is inherently buffered; bounding it at 16 MiB is the containment.

## Threat model

| Threat | Mitigation |
| --- | --- |
| Path traversal / symlink escape on upload | `os.Root` confinement; no lexical prefix check anywhere |
| TOCTOU between validation and read | `os.Root` opens the handle; the size check is on the open handle, not a separate `os.Stat` of a path |
| Memory exhaustion via a huge upload | `max_decoded_bytes` required, clamped to a 16 MiB engine ceiling, read bounded at `max+1` |
| Silent truncation producing a corrupt upload | overflow is an error, never a truncation |
| Local filesystem path leaking to the provider | `source_field` is deleted from the body before transmission |
| Malformed base64 accepted and transmitted | `base64.StdEncoding.Strict()`, then re-encode from decoded bytes |
| Approved-payload substitution between approve and execute | digest verified against `cfg.ApprovedPayloadSHA256`, mirroring multipart |
| Infinite pagination loop from a hostile/buggy API | strictly-advancing index required; non-advance is a sticky `Err()` |
| Over-run past the last page leaking unrelated records | stop when `next > totalResults`, and on any empty page |
| Unfiltered enterprise-wide listing | `required_query` enforced before the request is issued |
| Raw request escape hatch | none added: method, path and body structure remain bundle metadata in all four capabilities |

## Verification plan

Local gates, all run before the branch is handed off:

```
gofmt -l cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
go run ./cmd/connectorgen validate . --json      # 0 findings
go run ./cmd/connectorgen boundary . --json      # clean
make verify
```

Plus, per the captain's "verify by executing, not by reading" rule, the built `pm` binary is run
against a real bundle to prove that 553 connectors still load and that no existing declaration
regressed — reading a passing test is not the same as watching the binary work.

## Definition of done

- Four capabilities implemented, each test-first, each with red evidence recorded.
- No connector bundle modified; `connectorgen boundary` clean.
- All local gates green.
- Issue tree drafted to `ISSUE-TREE-DRAFT.md` (creation blocked on Alfred identity — recorded, not
  skipped).
