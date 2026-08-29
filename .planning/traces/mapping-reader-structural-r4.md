# Mapping-reader v3 event-schema inventory R4

## Delivery header

- Scope: the source-lock-first Asana v3 migration and the mapping-only
  declaration-admission reader that consumes the resulting closed wire.
- Base chain: `14aa19c76c327617216a891f394c9a658819208f` through the reviewed R2/R3
  mapping-reader commits `aedc7a97d` and `72bac8256`.
- Asana v3/import/projection code SHA:
  `b4d01f5603f54afd960a7a43592cf40f5a1a7d7e`.
- R4 mapping-reader code SHA:
  `b0d17554dcdb563c8c3e75ff6a5a6db3852c5133`.
- Delivery: local commits only; no push, PR creation, merge, runtime change,
  or certification behavior change.

## Source-lock-first migration

The former Asana v2 lock mixed an OpenAPI operation inventory with an embedded
event-source contract. The historical provider-evidenced v3 lock is the
canonical source representation: one `asana-openapi` document, the exact
Asana artifact, and 249 operation identities. Its retained identity is:

| Field | Value |
| --- | --- |
| Lock schema / SHA-256 | `3` / `af9c24a7d64a923af3a6b508b69d0cef3c413058545356d6c5aff73e108d2e72` |
| Document ID | `asana-openapi` |
| Artifact URL | `https://raw.githubusercontent.com/Asana/openapi/56796a67a3c093eedf55fd9682357957a2ebfd85/defs/asana_oas.yaml` |
| Artifact SHA-256 / bytes / form | `cb3b90f4e0af56035eab0c648974f625b942a28a7144aa6c2326e38ca0bb3d56` / `3066750` / OpenAPI `3.0.0` |
| Operation tuple count | `249` |

The exact tuple `(id, protocol, method, path, operation_id, source_location)`
matches historical canonical commit `b6b41bd9d`. The v3 lock retains an
`event_schema_inventory` as selectors, not copied provider schema:

```json
{
  "source_document": "asana-openapi",
  "schemas": [
    {"name":"AsanaNamedResource","source_location":"components.schemas[\"AsanaNamedResource\"]"},
    {"name":"EventResponse","source_location":"components.schemas[\"EventResponse\"]"},
    {"name":"NextPage","source_location":"components.schemas[\"NextPage\"]"},
    {"name":"TaskCompact","source_location":"components.schemas[\"TaskCompact\"]"},
    {"name":"TaskResponse","source_location":"components.schemas[\"TaskResponse\"]"}
  ]
}
```

Strict source import validates the selector's document, canonical schema
location, unique sorted names, and resolution against the pinned retained
artifact. The event contract test derives its event proof from those selectors
and bytes. No compatibility parser preserves an independently invented v2
event contract.

## R4 mapping-only repair

The mapping reader is intentionally independent of retained hashes, byte
counts, capture representation, certification, and opaque provider leaves.
It still needs a closed wire, however. After the Asana v3 migration, mapping
projection rejected the valid lock before it could review an operation because
`rest.event_schema_inventory` was unknown.

R4 adds the optional inventory to both mapping reader wire stages:

1. the initial v3 closed wire uses typed closed inventory and selector structs,
   so unknown root or nested selector fields fail; and
2. the later lossy projection wire keeps the admitted payload as
   `json.RawMessage`, so it has no mapping, runtime, or certification effect.

It does not make mapping admission resolve provider schema selectors. That
remains strict source-import behavior against the retained artifact.

## Red / green evidence

The R4 red checkpoint temporarily omitted the new closed-wire fields:

```text
go test ./cmd/connectorgen -run '^TestDeclarationAdmissionMappingV3EventSchemaInventoryIsClosedAndIgnored/valid_source_selector_is_ignored_after_structural_admission$' -count=1 -v
```

It failed exactly with:

```text
parse source lock mapping evidence: json: unknown field "event_schema_inventory"
```

The restored R4 test matrix proves a valid selector is accepted and ignored
after structural admission, while unknown inventory and nested selector fields
are rejected. It also parses the retained Asana lock through mapping admission
and asserts all 249 identities survive.

The direct-read projection repair is additive: valid stream ownership remains
in `api_surface.json` and receives a `direct_read` binding alongside it. The
Asana lane matrix records its twelve such rows as
`implemented_direct_read_etl`; no user-facing `direct_read` intent is replaced
by `etl`.

## Verification

```text
go test ./cmd/connectorgen -run '^(TestDeclarationAdmissionMappingV3EventSchemaInventoryIsClosedAndIgnored|TestDeclarationAdmissionMappingReaderAcceptsRetainedAsanaV3EventSchemaInventory|TestRetainedAsanaSourceImportRejectsReadProjectionDrift|TestSourceProjectionMaterializesSourceBoundGETReadsWithoutInventingETL|TestSourceProjectionMergeDirectReadSurfaceCoveragePreservesOnlyValidStream)$' -count=1
# PASS

go run ./cmd/connectorgen source-import asana --defs internal/connectors/defs --read-projection-only --check
# asana, 249 operation(s), 0 inbound event(s) verified

go run ./cmd/connectorgen source-import asana --defs internal/connectors/defs --check
# asana, 249 operation(s), 0 inbound event(s) verified

# Isolated first materialization followed by byte comparison and second check:
connectorgen source-import: asana, 249 operation(s), 0 inbound event(s) imported; source projection updated writes=0 cli=0
connectorgen source-import: asana, 249 operation(s), 0 inbound event(s) verified

go test ./internal/connectors/defs/asana -count=1
# PASS (23.993s)

go test ./cmd/connectorgen -count=1
# PASS (333.924s)

git diff --check
# PASS
```

## Fresh review requirement

The trace is documentation-only and binds the two code SHAs above. A reviewer
must review the exact ordered code pair `b4d01f560` then `b0d17554d` together:
the latter closes the mapping reader for the former's valid source lock. Any
later code-bearing commit requires a fresh exact-SHA review.
