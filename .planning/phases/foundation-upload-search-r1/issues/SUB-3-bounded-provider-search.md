# feat(connectors): add a bounded provider_search operation kind with enforceable list bounds

> **Draft — not created on GitHub.** See the parent draft for why (no `alfred-polymetrics-ai`
> credential in this environment; only the captain's own account is authenticated).

Parent: `.planning/phases/foundation-upload-search-r1/issues/PARENT-bounded-upload-and-provider-search-foundations.md`
Implements the captain decision recorded on **#2985**.

## Problem

A provider-search operation is a fixed POST body carrying a bounded list, executed as a read.
Freshchat's `POST /users/fetch` is the canonical shape: `ids[]`, bounded by the provider to 100.

Most of the transport already works. `OperationDirectRead` (`internal/connectors/engine/direct_read.go:32-129`)
executes POST reads today:

| requirement | already enforced |
| --- | --- |
| POST accepted | `direct_read.go:44-46` |
| `content_type` must be `application/json` | `:50-52` |
| `body_schema` required for POST | `:53-55` |
| body = declared `rest.body` + overrides, validated against `body_schema` | `operationReadBody`, `:229-247` |
| connector-relative path only | `:47-49`, and again at `:559-561` |
| endpoint declared in `api_surface` | `requireOperationDirectReadEndpoint`, `:214-227` |
| response bounded, clamped request → spec → ceiling | `clampOperationDirectReadMaxBytes`, `:257-269` |

The CLI input path is already typed rather than a raw body: flags declare `maps_to: body.<field>`
and are coerced per declared type (`commandrunner/runner.go:855-877`, `coerceFlagValue` at
`:1081-1124`). A `string_array` flag already yields `[]string` — the `ids[]` shape.

**Two things make it un-bounded, and therefore not shippable as a "bounded" capability:**

### 1. The bound cannot be expressed at all

The engine's schema dialect understands exactly these structural keywords
(`internal/connectors/engine/schema.go:60-72`):

```
type, required, properties, items, enum, pattern, minProperties,
additionalProperties, x-secret, x-primary-key, x-cursor-field
```

and **unknown keywords are a hard compile error** (`schema.go:104-110`, `compile schema: unknown
keyword %q`). So this declaration does not merely go unenforced — it **fails to load**:

```json
"ids": { "type": "array", "items": {"type": "string"}, "maxItems": 100 }
```

The CLI flag schema has the same hole. Flag objects in `cli_surface.schema.json` allow only
`name, type, summary, values, maps_to, format, allow_empty, required` —
no item bound — so a `string_array` flag accepts 10 000 values and nothing rejects them.

### 2. The body is open by default

A compiled schema node defaults to `additionalProperties: true` (`schema.go:108`). Of the **217**
declared `rest_read` POST operations, **169 do not set `additionalProperties: false`**. For general
reads that is the status quo. For a capability whose entire justification is "typed and bounded", an
open body root is precisely the escape hatch the captain banned — and it must be closed **by
construction for the new kind**, not retrofitted onto 169 existing rows, which would be an unrelated
migration riding a foundation PR.

## Design

### A distinct `provider_search` operation kind

#2985's recorded decision is that provider search/query is a **separate typed capability**, not a
global `capabilities.query=true` flip, and that `pm query` stays warehouse/materialized-data focused.
So: a new kind rather than a convention layered on `rest_read`, which makes the stricter rules
enforceable at load time and keeps the capability truthfully advertised in inspect/help.

Load-time validation, all of it refusing the bundle rather than warning:

- method `POST` only; path connector-relative; absolute URL rejected
- `content_type` must be `application/json`
- `body_schema` **required**, and its root must declare `additionalProperties: false`
- **every array-typed property in `body_schema` must declare `maxItems`** — this is what makes the
  kind bounded by construction. An unbounded list cannot ship.
- positive `max_bytes`
- `mutation_class` empty or `none` — it is a read, and must be rejected if it claims otherwise, the
  same way `rest_read` is at `bundle.go:1338-1340`

### Extend the schema dialect with `maxItems` / `minItems`

Add both to `structuralKeywords` and enforce them in array validation. Keep the dialect's existing
character: a bounded, explicit subset of draft-07 where unknown keywords stay a compile error. These
two keywords are added because a bound is now a load-time requirement, not because the dialect is
being opened up.

### Bound the CLI input too

Add `max_items` (and `min_items`) to the `cli_surface` flag object for `string_array` flags, and
enforce in `coerceFlagValue`. Two independent bounds — one on the wire contract, one on the input —
because the schema bound only fires after the flag has already been expanded, and the error a user
sees should name the flag they typed.

### Execution

Route `provider_search` through the existing bounded read path rather than a parallel executor: the
response bounding, clamping, redaction (`redactNamedJSONFields`, `direct_read.go:413-422`) and
output-policy handling are already correct and already tested. This sub-issue adds the stricter
front-half, not a second back-half.

## Explicitly out of scope

- Migrating the 169 existing `rest_read` POST operations to `additionalProperties: false`.
- Any generic SQL, GraphQL, HTTP, or arbitrary method/path/body surface. #2985 bans these outright
  and nothing here introduces one: the method is fixed by the kind, the path by the bundle, and the
  body by a closed schema whose every list is bounded.
- Editing `internal/connectors/defs/**` — connectors adopt this in their own lanes.

## Acceptance criteria

- A `body_schema` declaring `maxItems` compiles, where today it fails with `unknown keyword`.
- An array exceeding `maxItems` is rejected before the request is made.
- A bundle declaring `provider_search` with an unbounded array property **fails to load**.
- A bundle declaring `provider_search` without `additionalProperties: false` at the body root
  **fails to load**.
- A bundle declaring `provider_search` with a non-POST method, an absolute URL, a missing
  `body_schema`, or a mutating `mutation_class` **fails to load**.
- A `string_array` flag exceeding its declared `max_items` is rejected with an error naming the flag.
- A `provider_search` operation executes end-to-end against a real HTTP server in test, with the
  request body asserted.
- No caller can add a body key that the `body_schema` does not declare.

## Unblocks

- freshchat `POST /users/fetch` — currently `status: blocked`, risk medium, reason: "a fixed
  POST-body provider-search operation with ids[] bounded by the provider to 100 users, but executable
  provider_search/provider_query support is blocked on shared foundation #2985; no raw request
  body/query escape hatch is exposed"
- The wider gap #2985 records, which the parity audit measured at 310 operations across 3 connectors.
