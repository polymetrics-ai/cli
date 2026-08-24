# Retained-artifact lane adoption handoff

This is the concrete post-foundation handoff for the four blocked lanes. It
does **not** import their locks or provider bytes into this foundation branch.
Each owner merges the landed foundation, acts only under its own
`internal/connectors/defs/<connector>/` boundary, and commits the raw artifact,
manifest, lock/report evidence, and generated projection together.

## Common invariant and commands

`source-retain` is an explicit maintenance read, not a build step. It fetches
only the URL already named by a valid source-import lock, proves the returned
raw body against that preexisting lock, and writes:

- `sources/artifacts/<lowercase-lock-sha256>.artifact`;
- `sources/<connector>-retained-artifacts.json`.

Normal verification is offline:

```sh
go run ./cmd/connectorgen source-import <connector> --check
```

Before changing an untracked lock, make a durable copy inside the lane planning
evidence and prove it is byte-identical. Substitute the lane's phase directory
and lock path exactly; the copy is deliberately outside the untracked
`sources/` directory.

```sh
lock=internal/connectors/defs/<connector>/sources/<connector>-operation-source-lock.json
backup=.planning/phases/<lane-phase>/evidence/<connector>-operation-source-lock.pre-retain.json
mkdir -p "$(dirname "$backup")"
cp -p "$lock" "$backup"
cmp -s "$lock" "$backup"
shasum -a 256 "$lock" "$backup"
```

For a still-valid lock, acquire and then verify it with the following exact
shape. `undetermined` is permitted only when the lane records why a more
specific license/terms statement could not be established; use the known
provider text when it is available.

```sh
go run ./cmd/connectorgen source-retain <connector> \
  --retrieved-at "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  --license 'undetermined' --terms 'undetermined'
go run ./cmd/connectorgen source-import <connector> --check
```

If the second command reports descriptor/projection drift **after** retained
verification succeeds, regenerate through the canonical importer, never by
editing a generated operation/projection manually:

```sh
go run ./cmd/connectorgen source-import <connector>
go run ./cmd/connectorgen source-import <connector> --check
go run ./cmd/connectorgen surface-sync --check
```

Every lane report must record the connector, source URL, source kind, previous
lock bytes/SHA-256, fetched HTTP status and observed bytes/SHA-256, retained
file path, provenance license/terms, timestamp, and the hermetic
`source-import --check` result. A source-retain failure is not permission to
edit a pin: it is terminal until a Firstmate-authorized real-document re-pin
report exists.

## `cli-batch1-repair-r1` — Docker Hub

**Known state:** current `main` has no
`internal/connectors/defs/dockerhub/sources/dockerhub-operation-source-lock.json`.
There is therefore no preexisting byte/hash identity for `source-retain` to
preserve, and this command is intentionally expected to fail today:

```sh
go run ./cmd/connectorgen source-retain dockerhub \
  --retrieved-at "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  --license 'undetermined' --terms 'undetermined'
```

Do not replace that failure by retaining a current web page or a derived API
surface. First create and review a connector-owned canonical source-import
lock that names the actual provider artifact and its independently obtained
bytes/SHA-256; preserve that new lock with the common copy command. Then run:

```sh
go run ./cmd/connectorgen source-retain dockerhub \
  --retrieved-at "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  --license '<recorded Docker Hub license/status>' \
  --terms '<recorded Docker Hub terms/status>'
go run ./cmd/connectorgen source-import dockerhub --check
```

Record the canonical-source selection and first-pin evidence separately from a
re-pin report. **Terminal cause if this cannot be done:** `no canonical
machine-readable/rendered/bundle source lock`; Docker Hub remains
source-unavailable/merge-blocked rather than claiming a retained provider pin.

## `cli-map-batch23-r1` — Elasticsearch

The lane's existing v2 lock is directly compatible. Preserve
`internal/connectors/defs/elasticsearch/sources/elasticsearch-operation-source-lock.json`
first using the common copy command. Its prior identity is exactly
`6,458,869` bytes / `9b2ad824cfc8c8c4ea1b942ef6797da63ac6306c69f7e30e108b11a1e527dde9`.

The first retention attempt is an evidence command and is expected to refuse
the observed 32-byte drift:

```sh
go run ./cmd/connectorgen source-retain elasticsearch \
  --retrieved-at "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  --license 'undetermined' --terms 'undetermined'
```

Only after Firstmate's existing narrow authorization is applied to a recorded
HTTP-200 real OpenAPI document may the lane change that one lock identity. Its
`REPIN-REPORT.json` must state the source URL, retrieval time, HTTP 200/OpenAPI
classification, prior identity above, observed new `6,458,837` bytes and its
computed SHA-256, decision authority, and the pre-change backup path. It then
runs:

```sh
go run ./cmd/connectorgen source-retain elasticsearch \
  --retrieved-at "<recorded-RFC3339-repin-time>" \
  --license '<recorded Elastic license/status>' \
  --terms '<recorded Elastic terms/status>'
go run ./cmd/connectorgen source-import elasticsearch --check
```

**Terminal cause if classification, authorization, or exact retention fails:**
`source-lock drift`; retain neither the mismatched bytes nor an error response,
and leave import terminal rather than silently fetching live content.

## `cli-map-batch8910-r1` — 15 form pins and 12 descriptors

Preserve every lane-owned `*-operation-source-lock.json` before changing its
untracked `sources/` directory. The branch must normalize its older
`rest.documents` records into the v3 `rest.source_documents` contract without
changing an already valid artifact URL, byte count, SHA-256, source kind, or
operation inventory. This is source-lock normalization, not a re-pin.

For each of the 15 machine-readable OpenAPI/Swagger documents whose normalized
lock validates, run the common `source-retain <connector>` and
`source-import <connector> --check` commands. `source-retain` enumerates every
artifact in that connector lock; do not pass an artifact URL or path yourself.

For each of the 12 rendered-reference or zip/gzip bundle documents, put a v3
document in that same connector lock with its actual `kind`
(`rendered_reference` or `bundle`), exact `artifact` and `published_source`
identity, normalized content type, and reviewed operation inventory. Then use
the same commands:

```sh
go run ./cmd/connectorgen source-retain <connector> \
  --retrieved-at "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  --license '<recorded provider license/status>' \
  --terms '<recorded provider terms/status>'
go run ./cmd/connectorgen source-import <connector>
go run ./cmd/connectorgen source-import <connector> --check
go run ./cmd/connectorgen surface-sync --check
```

Record one row per document, not one aggregate lane result: source kind,
canonical descriptor path, old/new identities (identical for preserved pins),
raw retained path, and check result. **Terminal causes:** a machine-readable
body that differs is `source-lock drift` pending a Firstmate re-pin report; a
rendered/bundle source with no independently captured raw bytes and canonical
descriptor is `canonical descriptor unavailable`. Neither condition permits a
live fallback or an error-page pin.

## `cli-zoom-parity-lane-r1` — Zoom accounts

**Do not run `source-retain zoom` for the accounts source and do not re-pin it.**
The historical identity is `805,789` bytes /
`d8d650237496719594ca93a5aecacf368e71c4e30ac17eba46bd6f676a98319a`; the
dated upstream URL returned HTTP 404 with an 8,329-byte error body, and the
historic raw bytes were absent from the Git-object search. The error body is
not provider-document evidence and must never enter `sources/artifacts/`.

The lane must record this exact unavailable source in its connector-owned v3
source lock/evidence: document `kind: unavailable`, a stable accounts document
ID, empty operations, and a non-empty `unavailable_reason` naming the URL,
HTTP 404, 8,329-byte error body, historic bytes/SHA-256, and failed historic
blob search. Do not fabricate `artifact` or `published_source` bytes for that
unavailable document. The required terminal proof is:

```sh
go run ./cmd/connectorgen source-import zoom --check
```

It must fail terminally with `source document <accounts-id> is unavailable:`
and the recorded reason. Retain only other Zoom source documents that can pass
the common exact-byte command. **Terminal cause:** `irrecoverable historical
accounts artifact (HTTP 404; no verified historic copy)`; there is no
workaround, re-pin, cache fallback, or recovery claim.
