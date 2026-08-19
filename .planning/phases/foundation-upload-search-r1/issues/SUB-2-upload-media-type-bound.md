# feat(connectors): enforce a declared media-type bound on bounded uploads

> **Draft — not created on GitHub.** See the parent draft for why (no `alfred-polymetrics-ai`
> credential in this environment; only the captain's own account is authenticated).

Parent: `.planning/phases/foundation-upload-search-r1/issues/PARENT-bounded-upload-and-provider-search-foundations.md`
Depends on: sub-issue 1 (`SUB-1-bounded-multipart-file-upload.md`)

## Problem

Freshchat declares **two** blocked upload operations, not one: `POST /files/upload` (any file) and
`POST /images/upload` (image only). They are the same transport with a different input contract. The
difference is a **type bound**, and today there is no way to declare or enforce one.

The transport itself is not missing — see sub-issue 1, which establishes by execution that bounded
multipart upload already ships (`pm gong calls upload-media`, `availability: implemented`). What is
missing is any check that the bytes match the declared type.

`MultipartPartSpec.ContentType` (`engine/bundle.go:414-421`) is asserted by the bundle and written
straight into the part header by `writeMultipartFile` (`connsdk/http.go:462-472`). **Nothing reads
the bytes.** A part declaring `image/png` whose record points at a ZIP uploads the ZIP, labelled
`image/png`. For an image-upload operation that unchecked assertion *is* the entire contract.

This matters beyond correctness. Uploads are a **write to the provider**: shipping the wrong bytes
under a declared type is a mutation the user did not approve, and it is not recoverable by retry.

## Design

### `allowed_media_types` on the multipart part spec

A declared, closed list on `MultipartPartSpec`, alongside the existing `content_type`:

```json
"allowed_media_types": ["image/png", "image/jpeg", "image/gif", "image/webp"]
```

- Validated at bundle load: each entry must parse as a media type (`mime.ParseMediaType`); an empty
  array is a load error, not "allow everything" — absent means unconstrained, present means bounded,
  and the two must not be confusable.
- When the spec also declares `content_type`, it must itself be a member of the list.

### Enforce against the bytes, not the label

Sniff with `http.DetectContentType` over the first 512 bytes **during the existing snapshot copy** —
the same pass that already computes SHA-256 via `io.MultiWriter` (`connsdk/http.go:378`), so no extra
I/O and no second read that could race the first.

**Upload rejects on mismatch. This is deliberately the opposite of the download lane**, which the
research says should *record and surface* a `Content-Type` mismatch rather than reject, because
there the provider is the one making the claim and providers lie (Marketo serves CSV bytes from a
`.json` path). On upload **we** are the party making the claim, so an unsatisfiable claim is our bug
and must fail closed.

Sniffing caveats to encode in tests rather than discover later:

- `http.DetectContentType` returns `application/octet-stream` for content it cannot classify, and
  never returns an error — treat "unclassifiable" as a distinct rejection with its own message, not
  as a silent pass.
- It returns parameters for text types (`text/plain; charset=utf-8`), so compare on the parsed media
  type, not on raw string equality.
- It classifies by sniff table, so a valid-but-unusual image encoding can legitimately sniff as
  `application/octet-stream`. The declared allowlist is the contract; the error message must say
  which type was sniffed so a connector author can widen the declaration deliberately.

### Where the check lives

In `snapshotApprovedMultipartFiles` / `snapshotMultipartFile`, which today run only when a file
carries an approved digest (`connsdk/http.go:302`). The gate widens to "a digest **or** a media
bound is declared", so an operation may bound its media type without also requiring digest approval.
Files with neither still take the existing no-snapshot path unchanged.

## Acceptance criteria

- A bundle whose `allowed_media_types` contains an unparseable entry, or is present but empty, fails
  to load.
- A bundle whose declared `content_type` is not in its own `allowed_media_types` fails to load.
- A test demonstrates the current defect first (red): a ZIP uploads successfully under a declared
  `image/png` part.
- Uploading a PNG to a part allowing `image/png` succeeds.
- Uploading a ZIP to that same part is **refused before any request is made**, with an error naming
  both the declared allowlist and the sniffed type.
- A file whose content sniffs as `application/octet-stream` against a constrained allowlist is
  refused with a distinct, non-generic message.
- An operation with no `allowed_media_types` still uploads (unconstrained remains a valid, explicit
  choice).
- No new I/O pass over the file — sniffing rides the existing snapshot copy.

## Unblocks

- freshchat `POST /images/upload` — currently `status: blocked`, risk high, reason: "multipart/form-data
  with local image input; the current Freshchat bundle has no connector-local binary/multipart
  execution contract without shared binary/file foundation or an approved hook"
- Tightens the 32 declared `file_upload` operations across xero/zendesk-support/bitbucket/asana,
  several of which are attachment endpoints with provider-side type restrictions that are currently
  undeclarable.
