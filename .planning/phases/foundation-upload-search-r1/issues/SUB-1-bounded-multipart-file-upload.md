# fix(connectors): confine multipart upload file access with os.Root

> **Draft — not created on GitHub.** See the parent draft for why (no `alfred-polymetrics-ai`
> credential in this environment; only the captain's own account is authenticated).

Parent: `.planning/phases/foundation-upload-search-r1/issues/PARENT-bounded-upload-and-provider-search-foundations.md`

## First, the thing that is *not* broken

This issue was scoped expecting to build a bounded multipart upload executor. **That executor already
exists and ships.** Established by execution:

```
$ go build -o pm ./cmd/pm && ./pm gong calls upload-media --help
NAME          pm gong calls upload-media - Add call media (/v2/calls/{id}/media)
INTENT        reverse_etl
AVAILABILITY  implemented
WRITE         upload_call_media
NOTES         Uses typed multipart write support; no generic upload command is exposed.
FLAGS
  --id (string): ... maps_to=record.id
  --media-file-path (string): Project-relative media file path to upload. maps_to=record.media_file_path
```

The chain is: `cli_surface.json` command → `writes.json` action with `body_type: multipart` →
`engine/write.go:430-436` → `buildMultipartPayload` (`write.go:506-556`) →
`resolveMultipartFilePath` (`write.go:558-596`) → `connsdk.DoMultipart` (`connsdk/http.go:244-256`)
→ `snapshotApprovedMultipartFiles` (`:291-343`) → `snapshotMultipartFile` (`:345-388`).

Gong is the only adopter — searching every `internal/connectors/defs/*/*.json` for `"multipart"`
returns exactly `gong/writes.json`, which declares `upload_call_media` (1.5 GiB cap) and
`upload_crm_entities` (200 MiB cap). The path is bounded, digest-approval-bound, record-driven
(never argv), and gated behind plan → preview → approval → execute.

Two consequences:

1. **Do not build a second executor.** Duplicating this is the specific failure this program was
   told to avoid.
2. The `file_upload` **operation kind** (32 declarations: xero 22, zendesk-support 5, bitbucket 4,
   asana 1) is a **dead parallel declaration** — load-validated at `bundle.go:1368-1377`, hard-blocked
   at `commandrunner/runner.go:239-247`, no executor — while the executable contract lives in
   `writes.json`. Reconciling the two requires editing connector bundles, which the path-ownership
   guardrail forbids on a foundation branch. Tracked as a follow-up for the connector lanes.

Freshchat's blocked rows say *"the current **Freshchat bundle** has no connector-local
binary/multipart execution contract"* — a statement about the bundle, not the runtime. Gong proves
the runtime contract exists; Freshchat adopts it in its own lane.

## The actual defect

**Multipart file access is check-then-open, and the file is re-opened three times after the check.**

`resolveMultipartFilePath` (`engine/write.go:558-596`) validates once:

1. `safety.ValidateLocalWritePath` — **purely lexical**. Verified at
   `internal/safety/safety.go:128-158`: `RejectDangerousChars`, `filepath.Clean`, `filepath.Rel`,
   prefix compare. **No symlink resolution whatsoever.**
2. then `filepath.EvalSymlinks`
3. then `requireInsideRoot` (`write.go:604-613`)

Then the resolved path string is handed to `connsdk`, which opens it **three more times, by path,
with nothing confining it**:

| site | call | purpose |
| --- | --- | --- |
| `connsdk/http.go:273` | `os.Stat(file.Path)` | pre-read size rejection |
| `connsdk/http.go:346` | `os.Open(file.Path)` | snapshot copy + SHA-256 |
| `connsdk/http.go:475` | `os.Open(file.Path)` | the actual wire write |

Between the check and each of those opens the path can be replaced. Nothing re-validates. `os.Stat`
and `os.Open` both follow symlinks. The digest-approval binding narrows but does not close this: it
only applies when `ExpectedSHA256 != ""` (`http.go:302`), and the third open — the one that puts
bytes on the wire — happens after the digest comparison, against the path again.

`golang-security` names the fix without qualification: *"Path Traversal … Go 1.24+: use `os.Root`."*
The download research reached the same conclusion independently, observing that the upload path
"bolts symlink resolution on separately … which only works for existing files and is TOCTOU-racy",
and that `os.Root` "closes traversal, symlink escape and the race in one primitive". This module is
on `go1.25.4`, so the full method set is available.

## Design

**Containment belongs at the open, not before it.**

- `MultipartFile` gains a root handle plus a root-relative path. When set, **every** `Stat` and
  `Open` of the source file goes through `os.Root` — validation, snapshot, and wire write alike.
  A path that escapes is refused by the kernel-level containment on each access, not by a string
  comparison performed once, minutes earlier.
- `engine/write.go` opens the `os.Root` at the project directory, converts the record-supplied path
  to a root-relative form, and hands both to connsdk; it closes the root when the request finishes.
- The lexical `safety.ValidateLocalWritePath` check stays as a cheap first filter — defense in
  depth, and it produces the better error message for the common typo — but it is no longer the
  thing load-bearing for containment.
- The temp file produced by `snapshotMultipartFile` lives outside the root by design, so the root
  handle is cleared on the prepared copy once a snapshot exists; the subsequent open targets the
  snapshot, which is the correct behaviour and is what makes digest approval meaningful.

Nothing about the request shape changes. No new caller input, no new flag, no method/path/body
surface.

## Acceptance criteria

- A test demonstrates the current defect first (red): source content is swapped for content outside
  the intended root after validation, and the outside content reaches the wire.
- After the fix, that same swap is refused, and refused by `os.Root` rather than by a lexical
  comparison.
- A root-relative path traversing out (`../…`) is refused at every one of the three access sites,
  each covered independently.
- Existing multipart write behaviour is unchanged: `gong upload_call_media` / `upload_crm_entities`
  still build the same request, per-file and aggregate caps still fire, the one-byte-past-the-limit
  overflow check still fires, digest-approval mismatch still refuses, temp files are still removed
  on every failure path.
- No connector bundle changes. `go run ./cmd/connectorgen ownership .` clean.

## Unblocks

- freshchat `POST /files/upload` (documented 25 MB cap) — adoptable in the Freshchat lane as a typed
  `writes.json` multipart action once file access is soundly confined.
- Hardens the two Gong upload actions that are live on `main` today.
