# Issue tree — DRAFT, NOT YET CREATED

**Why this file exists.** The captain's standing order for foundation work is to create the GitHub
issue tree first, under the `alfred-polymetrics-ai` identity, as the canonical execution record.
That could not be done from this environment: no Alfred credential exists here (`gh auth` resolves
only to `karthik-sivadas`, one account in `hosts.yml`, no `ALFRED`/`GH_TOKEN` in the environment,
`.no-mistakes`, or workspace config). GitHub authorship is immutable, so creating the tree under a
different account could not be corrected afterwards.

The credential question is escalated to the captain and tracked as `cli-pipeline-alfred-identity-r1`
(the same gap previously produced three duplicate PRs under the wrong identity). Firstmate directed
this phase to proceed with GSD planning meanwhile and hold the drafted bodies here.

**To create the tree once identity is resolved:** create the parent from the first section, then
each sub-issue, then link them with `gh-axi issue subissue add <parent> <child> ...`. Each body
below is complete and ready to paste; the `<!-- ... -->` markers are the canonical identity comments
matching the convention in existing Alfred-authored issues.

**Sub-issue count changed from four to three.** A fourth sub-issue, "flip the binary certification
gate", was dropped as a blocking item after the premise was checked against the code and found not
to hold for this phase. The reasoning is preserved as a recommendation note at the end of this file.

---

## Parent issue

**Title:** `feat(connectors/engine): shared bounded binary download, write query params, and typed dynamic-key write bodies`

**Body:**

```markdown
<!-- cli-engine-shared-capabilities-r1:canonical -->
# Shared connector-engine capabilities (foundation)

Status: **implementation-ready**. Foundation work scoped to `internal/connectors/engine/` and
`internal/connectors/connsdk/` plus tests. No connector bundle is modified by this tree; connector
lanes adopt these capabilities afterwards in their own lanes.

## Objective

Build three shared connector-engine capabilities that connector lanes cannot add themselves, plus a
deliberate certification-gate transition. Each capability is strictly additive and opt-in per bundle;
every existing caller keeps working unchanged.

## Sub-issues

| # | Capability | Unblocks | Verifiable on `main`? |
| --- | --- | --- | --- |
| 1 | Bounded binary/file download executor | 83 declared operations | yes — counted in `defs/` |
| 2 | Query parameters on write actions | 11 marketo operations | no — lane-reported |
| 3 | Typed dynamic-key write bodies | 1 marketo operation | no — lane-reported |
| 4 | Flip the binary certification gate | gate semantics | yes |

### Operation counts, verified

Sub-issue 1's 83 operations were counted directly from `internal/connectors/defs/*/operations.json`
on `main` (`kind == "binary_download"`):

| Connector | Operations |
| --- | --- |
| hubspot | 32 |
| xero | 26 |
| bitbucket | 15 |
| github | 9 |
| zendesk-support | 1 |
| **total** | **83** |

Plus front (5) and recurly (3) reported in flight — both connectors exist on `main` but declare no
`binary_download` operations there yet, so those 8 are not included in the 83.

**Sub-issues 2 and 3 counts are lane-reported and NOT verifiable on `main`**: marketo on `main` is
read-only (`api_surface.json`, `streams.json`, `spec.json` only — no `writes.json`, no
`operations.json`). The 11 query-parameter operations and `sync_program_member_data` are declared in
the marketo lane's own branch. The engine capability is built ahead of that adoption, which is the
point of a foundation PR.

## Design research

Full design research with `file:line` evidence:
`data/decisions/cli-binary-download-design-research-2026-08-04.md` (agent workspace). Every claim in
it was re-verified against this repo before implementation began.

Key finding: **binary download is not greenfield.** The kind is already in the schema enum
(`schema/operations.schema.json`), the block map (`bundle.go:1313`), `BinaryOperationSpec`
(`bundle.go:543-549`) and GET-only + positive-`max_bytes` semantic validation
(`bundle.go:1361-1367`). Only the executor is missing.

## Security constraint that shapes sub-issue 1

71 connector definitions authenticate with a non-`Authorization` custom header. Go strips only
`Authorization`, `WWW-Authenticate` and `Cookie` across cross-domain redirects — **custom headers are
forwarded** — and download endpoints redirect to CDNs constantly. There is no `CheckRedirect` policy
anywhere in the repo, `resolveURL` accepts absolute URLs as-is, and `Auth.Apply` runs
unconditionally.

Today this is safe only by accident: all five connectors declaring `binary_download` use
bearer/basic. The capability must close this hazard as it ships, by reusing `checkOrigin`
(`engine/paginate.go:389-407`) and installing an explicit `CheckRedirect` that re-runs the origin
check on every hop.

## Bounded by construction

Explicit size limits, no unbounded read into memory, no raw request escape hatch, and no archive
extraction. These are invariants of the tree, not per-issue choices.

## Path ownership bounds

`internal/connectors/engine/**`, `internal/connectors/connsdk/**`, `internal/connectors/certify/**`
(sub-issue 4 only), and their tests. No connector bundle edits.

## Acceptance criteria

- GSD plan, TDD ledger, and verification checklist under `.planning` before production edits.
- Every capability additive and opt-in; existing JSON direct-read path, current write behaviour, and
  closed record schemas unchanged.
- No secrets, no raw request escape hatch, no archive extraction.
- Local gates green: `gofmt`, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify`.
- Docs/help/website parity updated or explicitly marked not applicable.
```

---

## Sub-issue 1

**Title:** `feat(connectors/engine): bounded binary/file download executor`

**Body:**

```markdown
<!-- cli-engine-shared-capabilities-r1:sub1 -->
# Bounded binary/file download executor

Parent: shared connector-engine capabilities (foundation).

## Objective

Implement the missing executor for the already-declared, already-validated, deliberately
unexecutable `binary_download` operation kind — bounded by construction, streaming to disk, and
closing the custom-auth-header cross-domain redirect hazard as it ships.

## Unblocks 83 declared operations

Counted from `internal/connectors/defs/*/operations.json` on `main`:

| Connector | Operations |
| --- | --- |
| hubspot | 32 |
| xero | 26 |
| bitbucket | 15 |
| github | 9 |
| zendesk-support | 1 |
| **total** | **83** |

Plus front (5) and recurly (3) in flight, not yet declared on `main`.

**The declared corpus is not a trustworthy work list.** HubSpot's 32 are auto-generated from OpenAPI
and misclassified — several are JSON folder listings and search endpoints typed as
`binary_download`. This issue ships the executor; it does not light up 83 endpoints. Adoption is
per-lane, human-reviewed, starting with genuinely-binary operations.

## What already exists (do not rebuild)

- kind in the schema enum (`schema/operations.schema.json`) and block map (`bundle.go:1313`)
- `BinaryOperationSpec` with `method`, `path`, `max_bytes`, `allow_overwrite`, `extract_archives`
  (`bundle.go:543-549`)
- GET-only and positive-`max_bytes` semantic validation (`bundle.go:1361-1367`)
- commands hard-blocked at `commandrunner/runner.go:239-247`

## Security requirement: custom-header credential leak across redirects

71 connector definitions authenticate with a non-`Authorization` custom header (`X-API-Key`,
`Circle-Token`, `DOLAPIKEY`, `Ocp-Apim-Subscription-Key`, …). Go strips only `Authorization`,
`WWW-Authenticate` and `Cookie` across cross-domain redirects; **custom headers are forwarded**.
There is no `CheckRedirect` policy anywhere in the repo. `resolveURL` accepts absolute URLs as-is
(`connsdk/http.go:174-179`) and `r.Auth.Apply` runs unconditionally (`connsdk/http.go:547-552`), so a
foreign pre-signed URL through the existing `Requester` attaches the connector credential to a
third-party host.

**Required fix**: reuse `checkOrigin` (`engine/paginate.go:389-407`) — same-origin sends credentials,
cross-origin requires an explicit per-operation allowlist and sends none — and install an explicit
`CheckRedirect` that re-runs that check on every hop.

## Bounding requirements

- **Read one byte past the limit, then reject.** A truncated PDF looks like a valid write. Mirrors
  the existing JSON idiom (`DoLimited` passes `maxBytes+1`; callers reject overflow at
  `direct_read.go:110-112,181-183`).
- **Clamp request → spec → ceiling**, as `clampOperationDirectReadMaxBytes` does
  (`direct_read.go:257-269`).
- **Endpoint allowlisting**, as `requireOperationDirectReadEndpoint` does (`direct_read.go:214-227`).
- `snapshotMultipartFile` (`connsdk/http.go:345-388`) is the shape template.

**Blocking gap**: `connsdk` buffers every response into `Response.Body []byte` with a 64 MiB cap
(`http.go:27,574`), so a declared 100 MiB `max_bytes` cannot be satisfied. A new streaming method
returning an open `io.ReadCloser` is unavoidable.

**Retry hazard**: `doWithBody` retries the whole request up to 5 times (`http.go:524-590`). A retry
after partial bytes must restart from zero, or partial bodies concatenate into a corrupt file.

## Filesystem safety

- Use **`os.Root`** for containment. `safety.ValidateLocalWritePath` is purely lexical
  (`safety.go:128-158`) with no `EvalSymlinks`, and the upload path bolts symlink resolution on
  separately (`write.go:560-613`), which only works for existing files and is TOCTOU-racy. `os.Root`
  closes traversal, symlink escape and the race in one primitive. `os.Root.Rename`/`MkdirAll` are Go
  1.25+; this repo is on go 1.25.4.
- Files `0o600`, dirs `0o700` — downloaded content is often invoices and identity documents.
- Honour `allow_overwrite` with `O_CREATE|O_EXCL`.
- Provider-supplied filenames: `mime.ParseMediaType` → strip both `/` and `\` → `filepath.Base` →
  `filepath.Localize` → strict charset. Read `params["filename"]`, never `params["filename*"]`.
- `f.Sync()` before `os.Rename`; temp file in the destination directory.

## Explicitly out of scope

- **`extract_archives` must never extract.** Zip-slip and decompression-bomb territory, a separate
  capability. It is already declared `true` on `github.tarball_ref` and `github.zipball_ref`. Because
  this tree modifies no connector bundle, it is enforced as a **hard execution-time error** rather
  than a bundle-validation error (which would invalidate the github bundle).
- **Resumability.** Airbyte has none either; the cap is 10–100 MiB so a full retry costs seconds, and
  resume multiplies failure surface. Size and digest are recorded so a future `Range` resume has
  something to validate against.

## Record shape

Records are flat `map[string]any` (`connectors.go:41`) and pass through schema projection, so a
nested object will not survive. Bytes are never inlined.

    file_path, file_name, file_size_bytes, file_sha256, content_type,
    content_type_sniffed, source_operation, source_ref, downloaded_at, truncated

**Trap**: `shouldRedactJSONField` (`direct_read.go:451-469`) auto-redacts `download_url`, `content`,
`body`, `payload`, `raw`, and anything containing both "download" and "url". Use `source_ref`.

Field vocabulary intentionally overlaps Airbyte's verified `file_reference` names
(`staging_file_url`, `source_file_relative_path`, `file_size_bytes`) for cheap interoperability.
Airbyte's **path handling is not copied** — it does `lstrip("/")` and no `..` containment at all.

## Content type and integrity

Never trust `Content-Type` and never infer from the URL path — Marketo serves CSV bytes from a path
ending `.json`. Record both the provider's claim and `http.DetectContentType` of the first 512 bytes;
surface the mismatch, do not reject on it. Always compute SHA-256 during the copy via
`io.MultiWriter`.

## Captain decisions — surfaced, built configurable, not silently picked

1. Where downloads land by default, and whether an absolute path outside that root is ever permitted.
2. Whether cross-host pre-signed fetches are allowed at all. Without them, HubSpot private files and
   Stripe file links cannot be downloaded; with them, the "connector-relative URLs only" invariant
   (`direct_read.go:47-49,139-141`) gains a deliberate exception.
3. The default size ceiling — declarations say 100 MiB, the current buffer cap is 64 MiB.
4. Whether downloads need approval/plan-preview treatment. Existing declarations disagree: GitHub's
   says "filesystem writes require an explicit destination policy", Xero's 26 say `approval: none`.
5. Whether adding `golang.org/x/sys` for a disk-space check is acceptable in a dependency-light CLI.

## Acceptance criteria

- Bounded streaming executor with overflow-byte rejection, clamped limits, and endpoint allowlisting.
- `CheckRedirect` origin policy; credentials never attached to a non-owned host.
- `os.Root` containment; `0o600`/`0o700`; `allow_overwrite` honoured.
- `extract_archives: true` is a hard execution-time error.
- Flat record, no inlined bytes, `source_ref` not `download_url`.
- Existing JSON direct-read path byte-for-byte unchanged in behaviour.
```

---

## Sub-issue 2

**Title:** `feat(connectors/engine): query parameters on write actions`

**Body:**

```markdown
<!-- cli-engine-shared-capabilities-r1:sub2 -->
# Query parameters on write actions

Parent: shared connector-engine capabilities (foundation).

## Objective

Add an optional typed query map to write actions, shaped exactly like the `query` object
`streams.json` already uses. **Wiring only — the helper already exists.**

## Unblocks

11 marketo operations (lane-reported). No other connector documents this gap today.

**Not verifiable on `main`**: marketo on `main` is read-only — `api_surface.json`, `streams.json`,
`spec.json` only, with no `writes.json` and no `operations.json`. The 11 operations are declared in
the marketo lane's own branch. The engine capability is built ahead of that adoption.

## Current state, verified

- `WriteAction` (`bundle.go:379-397`) has no query field.
- `executeWriteRecord` (`write.go:391-451`) passes `nil` for query in **all six** body-type branches:
  `form`, `graphql`, `none` (twice), `json_array`, `multipart`, and the default `json` path.
- `connsdk`'s `Requester.Do`, `DoForm`, `DoMultipart` all already accept a query `url.Values`.
- `resolveQueryParams` (`read.go:707-726`) is a working, documented helper already shared by
  `buildInitialQuery` (stream reads) and `buildCheckQuery` (check reads).

## Requirement

Reuse `resolveQueryParams` rather than writing a second one. The per-entry dialect must be identical
to the one `stream.Query` has always used and that `QueryParam.UnmarshalJSON` (`bundle.go:310-330`)
already implements:

- a **plain string** entry — `Template` is that string, `OmitWhenAbsent` false, `Default` empty — and
  an unresolved config/secrets key is **always a hard error** (zero migration risk);
- an **object** entry `{"template": "...", "omit_when_absent": true, "default": "..."}` — the
  explicit opt-in dialect.

This is deliberately not a blanket absent-key-falsy change, which would silently convert a
missing required key from a fail-loud error into a silently-unfiltered request.

## Files

- `bundle.go` — add `Query map[string]QueryParam` to `WriteAction` (parsing comes free from the
  existing `QueryParam.UnmarshalJSON`).
- `schema/writes.schema.json` — add `"query": { "type": "object" }`, matching how
  `streams.schema.json:101` types the same construct.
- `write.go` — resolve once in `executeWriteRecord` and thread the result through all six branches.

## Acceptance criteria

- A write action with no `query` behaves **exactly** as before — `nil`/empty query, no query string.
- Record fields are available to query templates on the same `Vars` the path already interpolates
  from.
- The stream/check dialect is reused verbatim; no second helper.
- Existing write tests pass unchanged.
```

---

## Sub-issue 3

**Title:** `feat(connectors/engine): typed dynamic-key write bodies`

**Body:**

```markdown
<!-- cli-engine-shared-capabilities-r1:sub3 -->
# Typed dynamic-key write bodies

Parent: shared connector-engine capabilities (foundation).

## Objective

Let a write action accept tenant-defined custom fields that have no fixed, enumerable official set —
**without** becoming a raw-body escape hatch.

## Unblocks

1 marketo operation: `sync_program_member_data` (lane-reported, not verifiable on `main` — marketo
on `main` is read-only with no `writes.json`).

## The constraint

Every write action validates records against `record_schema` (`write.go:42-66`,
`bundle.go:392`). Marketo's other writes use a closed schema with `additionalProperties: false`,
which deliberately excludes tenant-defined fields — so `sync_program_member_data` is fully blocked
today.

**The captain forbids a raw request escape hatch, and that rule stands.** The deliverable is a
*typed* dynamic-key primitive. If it cannot be built without effectively becoming an escape hatch,
the correct outcome is to report that and ship nothing.

## Design: a declared dynamic-key region, scalars only

An optional `dynamic_fields` block on the write action declares **one** named record field as the
dynamic-key region. Everything about the region is bundle metadata; only the keys and scalar values
inside it come from the caller.

```json
"dynamic_fields": {
  "field": "custom_fields",
  "key_pattern": "^[A-Za-z][A-Za-z0-9_]{0,63}$",
  "max_keys": 100,
  "value_types": ["string", "number", "boolean", "null"],
  "max_value_bytes": 4096,
  "target": "inline"
}
```

Why this is a typed primitive and not an escape hatch:

| Attack | Why it is closed |
| --- | --- |
| Inject request structure | `value_types` admits **scalars only**. A nested object or array is a hard validation error, so no caller input can ever become request structure. |
| Inject arbitrary keys | Every key must match the bundle-declared `key_pattern`. The pattern is metadata, never caller input. |
| Collide with declared fields | A dynamic key equal to any `path_fields` / `body_fields` / `body_field` entry, or to a key the body template already sets, is rejected — dynamic keys can never shadow structural ones. |
| Influence URL / method / headers | The region is merged into the JSON body only, after path interpolation. It reaches no other part of the request. |
| Unbounded growth | `max_keys` and `max_value_bytes` bound key count and per-value size. |
| Bypass the closed schema | `record_schema` stays closed. The dynamic region is validated **separately** and only for the one declared field. |

`target` selects whether the region is merged at the body root (`inline`) or kept nested under its
own field (`nested`), because providers differ; both are declarative, neither is caller-controlled.

## Files

- `bundle.go` — `DynamicFieldsSpec` struct + validation (declared field must not collide with
  `path_fields`/`body_fields`; `key_pattern` must compile; `value_types` from a fixed enum).
- `schema/writes.schema.json` — the `dynamic_fields` object, `additionalProperties: false`.
- `write.go` — validate and merge the region during body construction.

## Acceptance criteria

- A write action with no `dynamic_fields` behaves **exactly** as before; closed schemas stay closed.
- Non-scalar values, unmatched keys, colliding keys, over-count and over-size are all hard errors.
- No path exists from caller input to request structure, headers, method, or endpoint.
- Redaction (`redact_fields`) still applies to dynamic values.
```

---

## Recommendation note (dropped as a blocking sub-issue)

Retained verbatim as the drafted note. Raise this as its own issue at first lane adoption, not now.

```markdown
<!-- cli-engine-shared-capabilities-r1:sub4 -->
# Flip the binary certification gate deliberately

Parent: shared connector-engine capabilities (foundation).

## Objective

Move the binary certification stage from "**must stay blocked** because no executor exists" to
"**must stay bounded**", as a deliberate, recorded act rather than a side effect.

## Current state, verified

`certify/stages_binary.go:30-41` **fails the certificate if a binary command exits 0**:

> `binary command unexpectedly ran; operation-backed binary executors must stay blocked until an
> explicit bounded file policy is implemented`

It additionally requires stderr to contain `"operation"` and `"executor is not implemented"`, and
reports `Result: "blocked"` with the reason "bounded binary executor remains a future implementation
gate".

The same "stays blocked" claim is stated in `docs/connectors/github/SKILL.md:33` and
`MANUAL.md:31`.

## Sequencing finding — this gate does NOT break when sub-issue 1 lands

Verified by reading the actual wiring:

- Only **github** declares a binary certification candidate
  (`defs/github/certification.json:61-69`); `xero`'s `binary_candidates` is `[]` and no other
  connector has one.
- That candidate is `github release download`, which is declared `intent: local_workflow`,
  `availability: unsupported_local`.
- `resolvePreflightCommand` (`commandrunner/runner.go:239-247`) blocks it on the **first** branch —
  `cmd.Operation != "" && cmd.Intent != "direct_read"` — producing exactly the
  `"operation %s executor is not implemented in this slice"` string the certify stage asserts on.

Sub-issue 1 is scoped to `internal/connectors/engine/` and `internal/connectors/connsdk/` and does
**not** touch `commandrunner`. So after it lands, `github release download` still exits non-zero with
the same stderr, and **the gate keeps passing unchanged**.

The gate therefore does not have to flip for the foundation PR to be correct or green. What does go
stale is the gate's *reason text*, which will assert a future-implementation gap that no longer
exists.

## Recommended sequencing — captain's call

Flipping the assertion from "must stay blocked" to "must stay bounded" only becomes *meaningful* once
a connector lane actually wires a binary command to the executor, because only then can a binary
command exit 0. Two options:

1. **Flip with first adoption** (recommended). The gate stays a real assertion at every point: today
   it asserts "still blocked", and it becomes "ran, and stayed within its byte cap and its
   destination root" in the same PR that first makes that reachable.
2. **Flip now**, in or beside the foundation PR — updating the reason text and the two GitHub docs to
   say "bounded" while the command is still unreachable. Honest about the executor existing, but the
   assertion has nothing to bound yet.

This issue exists so the choice is recorded and made deliberately, in line with the captain's
instruction, rather than being decided silently by whoever touches `certify` next.

## Files

- `internal/connectors/certify/stages_binary.go`
- `docs/connectors/github/SKILL.md:33`
- `docs/connectors/github/MANUAL.md:31`

## Acceptance criteria

- The gate asserts a property that is true and meaningful at the time it lands.
- No certificate silently degrades from "asserts something" to "asserts nothing".
- The GitHub SKILL/MANUAL statements match the gate's actual semantics.
```
