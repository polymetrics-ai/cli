# Retained-artifact adoption handoff

This foundation PR deliberately retains only the two GitHub artifacts whose
lock is on current `main`. Firstmate inbox 004 prohibits importing unmerged
connector files from another lane. The shared contract is now:

- raw bytes: `sources/artifacts/<lowercase-lock-sha256>.artifact`;
- provenance: `sources/<connector>-retained-artifacts.json`;
- offline verifier: `connectorgen source-import <connector> --check`;
- explicit acquisition: `connectorgen source-retain <connector> --retrieved-at
  <RFC3339> --license <text> --terms <text>`.

`source-retain` never modifies a lock. It fetches only lock-addressed public
artifact URLs, validates the old bytes/SHA-256 before writing, and rejects
redirects, non-200 responses, and byte/hash drift. Builds and
`source-import` never use it.

## Batch 8–10

The owning Batch 8–10 lane must preserve each currently untracked source lock
before editing it, rebase on this foundation, and retain each real
machine-readable, rendered-reference, or bundle payload beside its own lock.
For every preserved pin, run `source-retain` and then `source-import --check`.

The branch's older `rest.documents` wire form is not the source-import v3
`rest.source_documents` wire contract. Before it can use `source-retain`, the
owning lane must convert its already-reviewed document entries to the standard
v3 lock representation without changing a valid artifact's URL, byte count,
or SHA-256. For any current response whose bytes differ, use Firstmate's
real-document classification/re-pin process and report the old/new identities
before `source-retain`; never let the retention command rewrite it.

## Elasticsearch (Batch 2–3)

Elasticsearch's v2 lock is directly compatible with `source-retain`. Preserve
the untracked lock first. The prior pin is `6,458,869` bytes / SHA-256
`9b2ad824cfc8c8c4ea1b942ef6797da63ac6306c69f7e30e108b11a1e527dde9`;
the observed real OpenAPI response is `6,458,837` bytes. Under Firstmate's
authorization, record the HTTP-200/OpenAPI classification and an explicit
old/new report, update only the corresponding lock identity, then run
`source-retain elasticsearch` and `source-import elasticsearch --check`.

## Zoom

Do **not** re-pin the accounts URL. Its historic lock is `805,789` bytes /
SHA-256 `d8d650237496719594ca93a5aecacf368e71c4e30ac17eba46bd6f676a98319a`,
while the dated upstream URL returned HTTP 404 with an 8,329-byte error body
and no matching Git blob was found. That response is not
provider-document evidence and must not enter `sources/artifacts/`. The Zoom
lane must record the accounts source as irrecoverable/unavailable (including
the exact URL, 404, byte count, and failed historic-blob search) and retain
only independently verifiable non-error sources. Its source-import check must
remain terminal for an unavailable source rather than silently refetching or
claiming accounts recovery.
