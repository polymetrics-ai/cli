# Open-object structured REST body inventory — 2026-08-24

Scope: read-only evidence for the open decision
`dockerhub-open-object-structured-body`. No source lock, declaration, generated
artifact, or runtime code was changed for this inventory.

Docker Hub's public source was fetched only to compare its SHA-256 with the
pinned lock. The fetched SHA-256 is
`99d9d53c2d93656a3c66d604885abd153dc5df285abc0ecb13802a3bc53d0756`,
which exactly matches the 148,322-byte lock at
`internal/connectors/defs/dockerhub/sources/dockerhub-operation-source-lock.json:5-10`.

## 1. Blast radius

The declaration-wide, read-only scan found **118 operations in 6 connectors**
whose `rest_write` `body_schema` contains an object that trips the requested
open-object refusal: `additionalProperties` is not the literal boolean `false`.
There are **zero** matching operations that instead fail because `properties`
is absent.

| Connector | Operations with non-closed object bodies |
| --- | ---: |
| Asana | 43 |
| Docker Hub | 20 |
| Gong | 4 |
| HubSpot | 36 |
| Notion | 2 |
| Zendesk Support | 13 |
| **Total** | **118** |

Method: for every `internal/connectors/defs/*/operations.json`, select each
`rest_write` operation with a `rest.body_schema`; recursively examine object
nodes reached through `properties`, `items`, and `prefixItems`; and count an
object when it is expressed by `type: object`, an object-containing type union,
or an object-shaped `properties` map. The scan applies the two branches named
in the request, rather than claiming to count every distinct validation in the
full structured-body compiler.

The current installed-command exposure is materially narrower: just **2
implemented direct-write operations in 1 connector**, both Docker Hub SCIM
operations:

| Command | Source operation | Refusal |
| --- | --- | --- |
| `pm dockerhub scim user create` | `dockerhub.post__v2_scim_2.0_users` | root object lacks `additionalProperties: false` |
| `pm dockerhub scim user update` | `dockerhub.put__v2_scim_2.0_users__id_` | root object lacks `additionalProperties: false` |

The enforcement is the existing engine rule at
`internal/connectors/engine/structured_rest_body.go:1425-1443`: it rejects an
object unless `additionalProperties` is explicitly the boolean `false` and
then requires `properties`. This is an execution-bound rule, not a statement
that the provider source is malformed.

## 2. Fixed-100 descope availability

Docker Hub cannot retain fixed-100 membership while declaring either of those
two SCIM operations unsupported/partial under the current rule. Both exact
source IDs are rows in the immutable 100-row artifact
`internal/connectors/operation-evidence-fixed-100.json`:

| Source ID | CLI and website path | Fixture | Required class |
| --- | --- | --- | --- |
| `dockerhub.rest.post_/v2/scim/2.0/Users` | `scim user create` | `fixtures/check.json` | `direct_write` |
| `dockerhub.rest.put_/v2/scim/2.0/Users/{id}` | `scim user update` | `fixtures/check.json` | `direct_write` |

`cmd/connectorgen/operationevidence.go:1153-1162` makes a fixed-cohort row
eligible only when it has no gaps, is runtime-enabled and conformant, and has
fixture, CLI, website, and an enabled declared classification. Its validator
then rejects any fixed row whose runtime, gap, conformance, CLI, website,
fixture, or classification evidence regresses
(`cmd/connectorgen/operationevidence.go:1187-1207`). The check is executed by
`cmd/connectorgen/operationevidence.go:210-230`, is directly regression-tested
at `cmd/connectorgen/operationevidence_test.go:224-289`, and is mandatory in
`make verify` through `Makefile:108-112,169`.

Consequently, marking either SCIM action unsupported/partial leaves the rest of
Docker Hub declared but fails `connectorgen operation-evidence --check` with
that row's execution evidence regression. It is not a valid fixed-100 descope.
Only rewriting/replacing the fixed reference (the `--write-fixed-100` path at
`cmd/connectorgen/operationevidence.go:194-209`) could make that pass; that is
a cohort rebaseline, not a connector-local disposition, and has not been
authorized.

If such a rebaseline were explicitly permitted, Docker Hub would lose the two
real binary commands above, their website/help routes, the fixture-backed
direct-write classification, and its current complete execution proof for
those two documented SCIM mutations. Its other declared operations would
remain in the bundle, but the connector would no longer meet the existing
fixed-100 execution invariant.

## 3. Smallest source-faithful compliant path

There is no connector-local way to satisfy the present **closed**-object rule
from the pinned Docker Hub source without inventing a provider restriction.
The public contract does enumerate fields:

- `POST /v2/scim/2.0/Users` refers to `scim_create_user_request` at
  `latest.yaml:2297-2306`; that request body is a `type: object` with
  `schemas`, `userName`, and `name` properties at `latest.yaml:4322-4337`.
- `PUT /v2/scim/2.0/Users/{id}` refers to `scim_update_user_request` at
  `latest.yaml:2347-2356`; its body enumerates `schemas`, `name`, and
  `enabled` at `latest.yaml:4338-4356`.
- The nested `scim_user_name` object enumerates `givenName` and `familyName`
  at `latest.yaml:3921-3929`; the more complete `scim_user` model similarly
  enumerates its documented fields at `latest.yaml:3944-3986`.

None of those object nodes declares `additionalProperties: false`. Under the
source contract, absence of that keyword leaves additional properties allowed;
adding it in the declaration would make the CLI reject provider-allowed input.
Enumerating every listed field cannot supply the missing closure claim, and
using only the currently selected command fields would additionally narrow the
provider's documented object.

Therefore the smallest compliant resolution is not a source import or
connector-local transformation. It is a captain-approved shared execution
capability able to represent a source-traced open object within explicit,
bounded input semantics (or an updated provider source that explicitly closes
these objects). Until then, the exact engine refusal above remains the minimal
missing hook; no live credentials or provider calls are needed to establish it.
