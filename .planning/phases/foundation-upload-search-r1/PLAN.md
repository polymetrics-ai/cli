# Phase foundation-upload-search-r1 — bounded upload containment/media bounds and bounded provider-search

## GSD setup

- Branch: `fm/cli-freshchat-upload-foundations-r1`
- GSD preflight: `scripts/gsd doctor` passed 2026-08-04 (node v24.13.1, 69 commands registered).
- GSD prompt path attempted first:
  `scripts/gsd prompt programming-loop init --phase foundation-upload-search-r1 --dry-run`
  → `scripts/gsd: unknown GSD command: programming-loop`. The repo-local registry does not carry
  `programming-loop`, so this phase is an explicit **manual-GSD fallback** recorded per
  `.agents/agentic-delivery/references/gsd-pi-adapter.md`, matching the precedent in
  `.planning/phases/issue-3674-issueguard-checkpoint-links/PLAN.md`.
- Planning prompt fallback applied inline:
  `scripts/gsd prompt plan-phase foundation-upload-search-r1 --skip-research` (142 lines).
- Orchestration decision: `local_critical_path` — one shared-runtime slice in an isolated worktree;
  no mutating subagents spawned.

## Required skills loaded

- `golang-how-to` (orchestrator), `golang-security` (path containment, TOCTOU, bounded reads)
- Applied from routing (`.agents/agentic-delivery/references/required-skills-routing.md`,
  connector-runtime row): `golang-design-patterns`, `golang-structs-interfaces`,
  `golang-error-handling`, `golang-safety`, `golang-testing`, `golang-context`

## Issue tree

Drafted, **not created on GitHub** — issue/PR creation for this program is supposed to use
`alfred-polymetrics-ai`, and no Alfred credential exists in this environment (`gh auth status`
reports only `karthik-sivadas`, the captain's own account). Authorship cannot be reattributed after
creation, so the tree is drafted to files:

- `.planning/phases/foundation-upload-search-r1/issues/PARENT-bounded-upload-and-provider-search-foundations.md`
- `.planning/phases/foundation-upload-search-r1/issues/SUB-1-bounded-multipart-file-upload.md`
- `.planning/phases/foundation-upload-search-r1/issues/SUB-2-upload-media-type-bound.md`
- `.planning/phases/foundation-upload-search-r1/issues/SUB-3-bounded-provider-search.md`

## The finding that reshaped this phase

The task framed the upload half as a missing capability. **It is not missing. It ships today, and it
is reachable from the CLI.** Established by execution, not by reading:

```
$ go build -o pm ./cmd/pm && ./pm gong calls upload-media --help
NAME
  pm gong calls upload-media - Add call media (/v2/calls/{id}/media)
INTENT        reverse_etl
AVAILABILITY  implemented
WRITE         upload_call_media
NOTES         Uses typed multipart write support; no generic upload command is exposed.
FLAGS
  --id (string): ... maps_to=record.id
  --media-file-path (string): Project-relative media file path to upload. maps_to=record.media_file_path
```

The live path is: `cli_surface.json` command (`intent: reverse_etl`, `write: upload_call_media`) →
`writes.json` action with `body_type: multipart` → `engine/write.go:430-436` →
`buildMultipartPayload` (`write.go:506-556`) → `resolveMultipartFilePath` (`write.go:558-596`) →
`connsdk.DoMultipart` (`connsdk/http.go:244-256`) → `snapshotApprovedMultipartFiles`
(`:291-343`) → `snapshotMultipartFile` (`:345-388`).

Gong declares two such actions (`internal/connectors/defs/gong/writes.json`: `upload_call_media`
1.5 GiB cap, `upload_crm_entities` 200 MiB cap). It is the **only** connector using it — a search of
every `internal/connectors/defs/*/*.json` for `"multipart"` returns exactly that one file. The path
is bounded, digest-approval-bound, record-driven (never argv), and behind
plan → preview → approval → execute.

So **building a second upload executor would duplicate working machinery** — exactly what the task
warned against. The `file_upload` *operation kind* (32 declarations: xero 22, zendesk-support 5,
bitbucket 4, asana 1) is a **dead parallel declaration**: validated at load
(`bundle.go:1368-1377`) but with no executor, hard-blocked at `commandrunner/runner.go:239-247`,
while the executable contract lives in `writes.json`. Reconciling those 32 declarations means editing
connector bundles, which the path-ownership guardrail forbids on this branch; it is recorded as a
follow-up for the connector lanes, not built here.

Freshchat's two upload rows say *"the current **Freshchat bundle** has no connector-local
binary/multipart execution contract"* — a statement about the bundle, not the runtime. Gong proves
the runtime contract exists. Freshchat adopts it in its own lane.

## What is genuinely missing

Three gaps survive that reading. All three are shared-runtime, all three are in scope
(`internal/connectors/engine/`, `internal/connectors/connsdk/`), and none of them duplicates
anything.

### Gap 1 — multipart file access is check-then-open (live security defect)

`resolveMultipartFilePath` (`engine/write.go:558-596`) validates once:
`safety.ValidateLocalWritePath` — **purely lexical, no symlink resolution at all**
(`internal/safety/safety.go:128-158`, verified: `filepath.Clean` + `filepath.Rel` + prefix compare)
— then `filepath.EvalSymlinks`, then `requireInsideRoot` (`write.go:604-613`).

After that single check the file is opened **three more times, by path, with no containment**:

| site | call |
| --- | --- |
| `connsdk/http.go:273` | `os.Stat(file.Path)` — the pre-read size check |
| `connsdk/http.go:346` | `os.Open(file.Path)` — the snapshot copy |
| `connsdk/http.go:475` | `os.Open(file.Path)` — the actual wire write |

Check-then-open with three subsequent opens is a textbook TOCTOU, and `golang-security` names the
fix directly: *"Path Traversal … Go 1.24+: use `os.Root`."* The download research reached the same
conclusion independently and noted the upload path "bolts symlink resolution on separately …
which only works for existing files and is TOCTOU-racy."

**Containment must live at the open, not before it.**

### Gap 2 — a declared media type is asserted, never verified

`MultipartPartSpec.ContentType` (`bundle.go:414-421`) is written straight into the part header by
`writeMultipartFile` (`http.go:462-472`). Nothing reads the bytes. A part declaring `image/png`
whose record points at a ZIP uploads the ZIP, labelled `image/png`.

This is precisely what separates Freshchat's `/images/upload` from its `/files/upload`: the image
operation's entire contract *is* the type bound, and today that contract cannot be expressed or
enforced. An upload is a provider mutation — shipping wrong bytes under a declared type is not
recoverable by retry.

### Gap 3 — a bounded list cannot be declared, let alone enforced

`OperationDirectRead` (`engine/direct_read.go:32-129`) already executes POST reads with a
`body_schema`-validated body, connector-relative paths, and response bounds. The CLI input is
already typed (`maps_to: body.<field>`, `coerceFlagValue` at `commandrunner/runner.go:1081-1124`),
and `string_array` already yields the `ids[]` shape.

The bound itself is impossible to state. The engine's schema dialect understands only
`type, required, properties, items, enum, pattern, minProperties, additionalProperties, x-secret,
x-primary-key, x-cursor-field` (`engine/schema.go:60-72`) and **unknown keywords are a hard compile
error** (`schema.go:104-110`). So

```json
"ids": { "type": "array", "items": {"type": "string"}, "maxItems": 100 }
```

**fails to load.** The CLI flag schema has the same hole — flag objects allow only
`name, type, summary, values, maps_to, format, allow_empty, required`, no item bound.

Second half: a compiled node defaults to `additionalProperties: true` (`schema.go:108`). Of 217
declared `rest_read` POST operations, **169 do not set `additionalProperties: false`** — an open body
root, which for a capability justified as "bounded and typed" is the escape hatch the captain banned.
It must be closed **by construction for a new kind**, not retrofitted onto 169 existing rows.

## Scope

### In scope — code

1. `internal/connectors/connsdk/http.go` — root-confined multipart source access; media-type
   sniffing during the existing snapshot copy.
2. `internal/connectors/engine/write.go` — open an `os.Root` at the project dir and hand it to
   connsdk instead of a pre-resolved absolute path.
3. `internal/connectors/engine/bundle.go` — `MultipartPartSpec.AllowedMediaTypes`; the
   `provider_search` operation kind and its semantic validation.
4. `internal/connectors/engine/schema.go` — `maxItems` / `minItems`.
5. `internal/connectors/engine/schema/operations.schema.json`,
   `internal/connectors/engine/schema/writes.schema.json`,
   `internal/connectors/engine/schema/cli_surface.schema.json` — declaration surfaces.
6. `internal/connectors/engine/direct_read.go` — accept `provider_search` on the bounded read path.
7. `internal/connectors/commandrunner/runner.go` — `max_items`/`min_items` enforcement on
   `string_array` flags; accept `provider_search`-backed commands so the validator and the runtime
   agree.
8. Tests alongside each.

### In scope — GSD planning evidence

9. `.planning/phases/foundation-upload-search-r1/{PLAN,TDD-LEDGER,VERIFICATION}.md`,
   `RUN-STATE.json`, and the four issue drafts. Required because `internal/**` changes and
   `scripts/verify-gsd-workflow` fails the `gsd-workflow-evidence` gate without changed
   `.planning/**` evidence. Documentation only; carries no behavior.

### Out of scope

- **Any file under `internal/connectors/defs/**`.** The path-ownership guardrail is live on `main`.
  Connectors adopt these capabilities in their own lanes. Verified clean with
  `go run ./cmd/connectorgen ownership . --base <merge-base>`.
- A second upload executor for the `file_upload` operation kind. The transport ships; duplicating it
  is the failure mode this phase exists to avoid.
- Reconciling the 32 `file_upload` operation-kind declarations with the live `writes.json` contract —
  requires bundle edits.
- Migrating the 169 `rest_read` POST operations to `additionalProperties: false` — an unrelated
  migration that must not ride a foundation PR.
- Resumability, archive extraction, and cross-host pre-signed fetches — download-lane concerns,
  and the research rejects the first two outright.

## Invariants this phase must not break

- **No raw request escape hatch.** After this phase, as before it, no caller can supply an arbitrary
  method, path, or body. The method is fixed by the kind or the action, the path by the bundle, the
  body by a closed schema. New declaration fields are bundle-authored, never caller-supplied.
- No local file path or credential crosses argv. The multipart path stays record-driven.
- No existing bundle stops loading. Every new declaration field is optional; every new constraint
  applies to the new kind or fires only when the new field is present. Verified by loading all
  bundles.
- No block gate, Connector Guard rule, or ownership guardrail is weakened.

## Task breakdown (TDD — red first for every behavior change)

| # | Slice | Red evidence to produce first |
| --- | --- | --- |
| 1 | connsdk root confinement | A test proving the current code uploads content from outside the intended root when the path is swapped after validation |
| 2 | engine hands an `os.Root` to connsdk | Existing multipart write tests stay green; a new test proves an escaping path is refused at the open |
| 3 | media-type sniff + allowlist | A test proving a ZIP is currently uploaded under a declared `image/png` |
| 4 | `allowed_media_types` declaration + load validation | A bundle with an unparseable / empty / non-member declaration currently loads |
| 5 | `maxItems`/`minItems` in the dialect | `CompileSchema` currently fails with `unknown keyword "maxItems"` |
| 6 | `provider_search` kind + strict load validation | An unbounded / open-bodied / non-POST / mutating declaration currently loads |
| 7 | `provider_search` executes | A `provider_search` operation currently errors `requires rest_read operation` |
| 8 | `string_array` item bounds | A flag currently accepts unbounded item counts |

## Verification gates

```bash
gofmt -l cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen ownership .            # path-ownership guardrail
scripts/verify-gsd-workflow origin/main          # gsd-workflow-evidence
```

Plus **execution, not reading** — the standing reason being the audit that found 174 commands
declaring `availability: implemented` while failing at runtime, caused by the validator exempting
operation-backed direct reads (`cmd/connectorgen/validate.go:1427-1430`) from a check the runtime
enforces (`commandrunner/runner.go:419-421`). Every claim in VERIFICATION.md is backed by a command
that was run and its actual output.

Website catalog: `.github/workflows/website.yml` triggers only on `website/**`,
`internal/connectors/icon_data.json`, `docs/connectors/icons/**`, `.github/workflows/website.yml`,
and `.gitlab-ci.yml`. This phase touches none of them, but the generator is run anyway and any diff
committed.
