# Phase: engine-shared-capabilities-r1

**GSD command:** `/gsd-plan-phase engine-shared-capabilities-r1`, generated through the repo-local Pi
adapter with `scripts/gsd prompt plan-phase engine-shared-capabilities-r1`. The generated prompt is
recorded at `.planning/traces/gsd-plan-phase-engine-shared-capabilities-r1-prompt.md`.

**Runtime fallback:** the adapter's `/gsd-plan-phase` expects Pi runtime subagents. This session runs
in Claude Code, so the documented inline/manual fallback was used: the official workflow was executed
inline by the single session agent rather than via spawned Pi researcher/planner/checker subagents.
This is the manual-GSD fallback AGENTS.md permits, recorded here explicitly. The TDD lifecycle in
`TDD-LEDGER.md` is not skipped — every behaviour-adding task starts red.

**Required skills loaded** (per `.agents/agentic-delivery/references/required-skills-routing.md`,
"Connector runtime and architecture"): `golang-how-to` (orchestrator), `golang-security`,
`golang-safety`. Applied throughout: `os.Root` for path containment (the skill's stated Go 1.24+
answer to path traversal), fail-closed origin checks for SSRF/credential leakage, bounded reads,
`defer`-in-loop avoidance, and explicit numeric bounds on byte limits.

## Scope

Foundation phase. Three shared connector-engine capabilities that connector lanes cannot add
themselves, all strictly additive and opt-in per bundle.

**Paths owned by this phase:** `internal/connectors/engine/**`, `internal/connectors/connsdk/**`,
and their tests. **No connector bundle is modified.** `internal/connectors/certify/**` and
`internal/connectors/commandrunner/**` are explicitly out of scope.

**Baseline recorded before edits:** `go run ./cmd/connectorgen boundary . --json` → `outcome: clean`,
`findings: 0`, `warnings: 0`, 132 files checked, 550 connectors loaded.

## Verified starting state

Every claim below was read directly out of this repo before planning.

| Claim | Evidence |
| --- | --- |
| `binary_download` is declared, validated, and deliberately unexecutable | schema enum in `schema/operations.schema.json`; block map `bundle.go:1313`; `BinaryOperationSpec` `bundle.go:543-549`; GET-only + positive `max_bytes` `bundle.go:1361-1367` |
| 83 operations already declared | counted in `defs/*/operations.json`: hubspot 32, xero 26, bitbucket 15, github 9, zendesk-support 1 |
| No executor exists; commands hard-blocked | `commandrunner/runner.go:239-247` |
| `connsdk` cannot stream | every response buffered into `Response.Body []byte`, cap 64 MiB (`http.go:27,574`) |
| Retry restarts the whole request | `doWithBody` retries up to 5 times (`http.go:524-590`) |
| No redirect policy anywhere | no `CheckRedirect` in the repo; `resolveURL` accepts absolute URLs (`http.go:174-179`); `Auth.Apply` unconditional (`http.go:547-552`) |
| Reusable origin guard exists | `checkOrigin` `paginate.go:389-407` |
| `WriteAction` has no query; all six write branches pass `nil` | `bundle.go:379-397`; `write.go:391-451` |
| A working query resolver already exists | `resolveQueryParams` `read.go:707-726`, shared by stream and check reads |
| `download_url` is silently auto-redacted | `shouldRedactJSONField` `direct_read.go:451-469` |
| Records are flat `map[string]any` | `connectors.go:41` |
| `binary` schema block is `additionalProperties: false` | new fields must be added to the schema to be usable |

## Task 1 — `connsdk` bounded streaming with a redirect origin policy

New file `internal/connectors/connsdk/stream.go`.

`DoStream(ctx, method, path, query, StreamOptions) (*StreamResponse, error)` returns an **open**
`io.ReadCloser` instead of a buffered `[]byte`.

- **Retry safety by construction.** The body is never read inside the retry loop. On a transport
  error or a retryable status the response body is closed and discarded before the next attempt, so
  partial bytes can never concatenate into a corrupt file. The body is handed to the caller only on
  the terminal successful attempt.
- **`CheckRedirect` on a cloned client.** The shared `r.Client` is never mutated. Every hop re-runs
  the same scheme+host equality check `checkOrigin` implements, fail-closed on unparseable or
  host-less URLs.
- **Credential stripping is fail-closed and general.** Rather than enumerating which header a
  connector uses for auth, `DoStream` snapshots the header keys before and after `Auth.Apply` to
  learn exactly which keys auth contributed, and on any cross-origin hop strips those keys plus
  every `DefaultHeaders` key, leaving only `Accept` and `User-Agent`. This covers the 71 connectors
  authenticating via a custom header that Go does **not** strip.
- Cross-origin is refused outright unless the operation declares it, in which case the hop proceeds
  **with no credentials at all**.

## Task 2 — engine bounded binary download executor

New file `internal/connectors/engine/binary_read.go`, parallel to the bounded-JSON path.

Order of operations, each a hard error:

1. operation exists, `kind == "binary_download"`, `op.Binary != nil`
2. method is GET
3. **`extract_archives: true` is refused at execution time.** It is already declared `true` on
   `github.tarball_ref` and `github.zipball_ref`; because this phase modifies no connector bundle,
   refusing at execution rather than at bundle-validation keeps the github bundle valid while
   guaranteeing extraction never happens. Zip-slip is a separate capability, not a flag.
4. endpoint declared in `api_surface` (reuses the direct-read allowlist helper)
5. path is connector-relative unless the operation declares cross-host
6. `max_bytes` clamped request → spec → ceiling
7. destination root opened with **`os.Root`** — closes traversal, symlink escape and the TOCTOU race
   in one primitive, which `safety.ValidateLocalWritePath` (purely lexical, `safety.go:128-158`)
   does not
8. stream into a temp file **inside the destination directory**, `io.MultiWriter` to SHA-256, reading
   `max+1` and rejecting overflow — a truncated PDF looks like a valid write
9. first 512 bytes captured for `http.DetectContentType`; the provider's claim is recorded, never
   trusted, never used to reject
10. `f.Sync()` then rename within the root; files `0o600`, dirs `0o700`; `allow_overwrite` honoured
    via `O_CREATE|O_EXCL`

Filename derivation: `mime.ParseMediaType` → read `params["filename"]` (never `filename*`; the
decoded RFC 5987 value lands under the unstarred key) → strip both `/` and `\` (RFC 6266 counts both)
→ `filepath.Base` → `filepath.Localize` → `filepath.IsLocal` assertion. Falls back to a
connector-controlled identifier when the provider supplies nothing usable.

Flat record, no inlined bytes:
`file_path, file_name, file_size_bytes, file_sha256, content_type, content_type_sniffed,
source_operation, source_ref, downloaded_at, truncated`. `source_ref` **not** `download_url`, which
`shouldRedactJSONField` would silently turn into `download_url_redacted: true`.

New optional `BinaryOperationSpec` fields (schema is `additionalProperties: false`, so each must be
added there too): `allow_cross_host`, `allowed_hosts`, `content_types`, `stall_timeout_seconds`.

**Captain decisions — built configurable and surfaced, not silently picked:** destination root and
whether absolute paths outside it are ever allowed; whether cross-host pre-signed fetches are
permitted at all (without them HubSpot private files and Stripe file links are undownloadable); the
default size ceiling (declarations say 100 MiB, current buffer cap is 64 MiB); whether downloads need
approval treatment (existing declarations disagree — GitHub says "explicit destination policy",
Xero's 26 say `approval: none`); and whether `golang.org/x/sys` for a disk-space check is acceptable.
**No new dependency is added in this phase** — the disk-space pre-check is deliberately omitted
rather than pulling in `x/sys` unilaterally.

Deliberately **not** built: resumability (Airbyte has none; the cap is 10–100 MiB so a full retry
costs seconds; size and digest are recorded so a future `Range` resume has something to validate
against).

## Task 3 — query parameters on write actions

- `bundle.go`: `Query map[string]QueryParam` on `WriteAction`. Parsing comes free — the existing
  `QueryParam.UnmarshalJSON` (`bundle.go:310-330`) already accepts both the plain-string and object
  forms.
- `schema/writes.schema.json`: `"query": { "type": "object" }`, matching `streams.schema.json:101`.
- `write.go`: resolve **once** in `executeWriteRecord` via the existing `resolveQueryParams` and
  thread the result through all six body-type branches. No second helper.

Record fields are exposed to query templates on the same `Vars` the path already interpolates from.
A write action with no `query` sends no query string, exactly as today.

## Task 4 — typed dynamic-key write bodies

An optional `dynamic_fields` block declaring **one** record field as a dynamic-key region. Everything
about the region is bundle metadata; only keys and scalar values come from the caller.

| Attack | Why it is closed |
| --- | --- |
| Inject request structure | `value_types` admits scalars only; a nested object or array is a hard error |
| Inject arbitrary keys | Every key must match the bundle-declared `key_pattern` |
| Shadow declared fields | Collision with `path_fields`/`body_fields`/`body_field` or a template-set key is rejected |
| Reach URL / method / headers | The region is merged into the JSON body only, after path interpolation |
| Unbounded growth | `max_keys` and `max_value_bytes` |
| Bypass the closed schema | `record_schema` stays closed; the region is validated separately, for one declared field |

This is a typed primitive, not an escape hatch. If implementation shows it cannot hold that line, the
correct outcome is to report that and ship nothing for this task.

## Out of scope, with reasons

- **Flipping the binary certification gate.** Verified unnecessary for this phase: only github
  declares a binary certification candidate (`defs/github/certification.json:61-69`), it targets
  `release download` (`intent: local_workflow`), and that is blocked in `commandrunner` — outside
  this phase's paths. The gate keeps passing unchanged; only its reason text goes stale. Recommended
  to flip with first lane adoption so the assertion always asserts something true. Recorded in
  `ISSUE-TREE-DRAFT.md`.
- **Normalizing `output_policy` to `binary_file_bounded`.** Would require migrating 67 non-conforming
  declarations across connector bundles, which this phase must not touch.
- **Connector adoption.** The 83 declared operations are not a trustworthy work list — HubSpot's 32
  are auto-generated from OpenAPI and misclassified (JSON folder listings and search endpoints typed
  as `binary_download`). Adoption is per-lane and human-reviewed.

## Definition of done

- All three capabilities additive and opt-in; existing JSON direct-read path, current write
  behaviour, and closed record schemas unchanged.
- `gofmt`, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify` green.
- `connectorgen boundary` still `clean`.
- CLI help/docs/website parity assessed against
  `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`.
