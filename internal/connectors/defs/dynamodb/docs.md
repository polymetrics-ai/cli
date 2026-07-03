# Overview

DynamoDB is a **Tier-3 native package** migration (`internal/connectors/native/dynamodb/`) of
`internal/connectors/dynamodb` (legacy, read-only reference until the wave6 registry flip). It reads
DynamoDB table items through the AWS JSON HTTP API: every request is a `POST` to the service root
(`/`) carrying an `X-Amz-Target: DynamoDB_20120810.Scan` header and a small JSON body
(`TableName`/`Limit`/`ExclusiveStartKey`), authenticated with hand-rolled **AWS Signature Version 4
(SigV4)** request signing (a canonicalized-request HMAC-SHA256 chain over date/region/service scope
— see `connection.go`'s `sign`). This package is engine-vs-legacy parity-tested against
`internal/connectors/dynamodb`. Read-only: legacy `dynamodb.go:128-130` always returns
`ErrUnsupportedOperation` from `Write`, and this package declares `capabilities.write: false` with no
`writes.json` to match.

## Why Tier 3, not Tier 1 or 2

Per `docs/migration/conventions.md` §6's escape-hatch decision tree: DynamoDB is genuinely
protocol-native, not a declarative-HTTP-shaped API. Two independent, compounding gaps make it
unexpressible even as a Tier-2 hook set:

1. **SigV4 signing has no declarative auth shape.** The engine's `AuthSpec` dialect covers
   `bearer`/`basic`/`api_key_header`/`api_key_query`/`oauth2_client_credentials`/`custom` (a named
   hook) — none of these is a canonicalized-request, date-scoped, HMAC-chained signature computed
   over the request's own method/URI/query/headers/body-hash. A single `AuthHook` COULD express the
   signing itself, but:
2. **The wire protocol itself is not REST.** Every operation is an identical `POST /` with an
   `X-Amz-Target` header selecting the RPC method and a JSON-RPC-style body — there is no per-stream
   `path`/`method`/query-param shape for `StreamSpec` to declare at all; the whole "declarative HTTP
   request" concept the engine's `read.go` builds (URL + method + query + path templating) does not
   apply to a fixed-endpoint, header-dispatched JSON-RPC protocol. This is the identical class of gap
   `docs/migration/conventions.md` §6 names DynamoDB itself as an example of ("SQL, message queues,
   filesystems, CDC... amazon-sqs" family) — combining a non-REST wire protocol with a non-declarative
   auth scheme means even a maximal 2-hook Tier-2 attempt (`AuthHook` for signing + `StreamHook` for
   the request shape) would need to reinvent the entire request lifecycle by hand inside hooks
   anyway, with no benefit over a plain Tier-3 package.

This bundle still ships `spec.json`/`streams.json`/`schemas/items.json`/`api_surface.json`/`docs.md`
so identity/spec/docs/schema stay uniform with every other connector (`native/dynamodb`'s
`connector.go` embeds `engine.Base`, built from this bundle, purely to serve
`Name()`/`Metadata()`/`Definition()`) — but `streams.json`'s `base`/`streams[]` entry is
**documentation/schema-reference only**: `Check`/`Catalog`/`Read` are hand-written Go
(`internal/connectors/native/dynamodb/{connection,reader,cataloger}.go`), never routed through
`engine.Check`/`engine.Read`. `metadata.json` does not set `capabilities.dynamic_schema: true`
(unlike the postgres Tier-3 golden, whose tables ARE discovered live from `information_schema`)
because DynamoDB's single generic `items` stream is a static constant, matching legacy's `Catalog`
exactly — no live `DescribeTable` discovery call is ever issued.

## Auth setup

Provide `access_key_id` and `secret_access_key` as secrets (both required), plus `region` (e.g.
`us-east-1`) and `endpoint` (the DynamoDB JSON HTTP API endpoint, e.g.
`https://dynamodb.us-east-1.amazonaws.com`; falls back to `base_url` when unset, matching legacy's
identical fallback). `connection.go`'s `sign` ports legacy `dynamodb.go`'s `sign` method verbatim:
every Scan POST is canonicalized (method, URI, query, signed headers, `SHA256(body)`), hashed into a
`stringToSign` scoped to `<date>/<region>/dynamodb/aws4_request`, and signed with an HMAC-SHA256 key
chain derived from `secret_access_key` — the secret itself never appears in a header or log line,
only the derived signature does.

## Streams notes

A single generic stream, `items` (`x-primary-key: ["pk"]`, matching legacy's own illustrative static
`Field{Name: "pk"}` catalog entry — the real partition key attribute name is whatever the target
table defines, unknowable ahead of time). `reader.go`'s `Read` issues a
`DynamoDB_20120810.Scan` POST per page (`TableName` from `config.table_name`, falling back to
`config.table`; `Limit` from `config.page_size`, default 100), passing the prior page's
`LastEvaluatedKey` back as `ExclusiveStartKey` until a page returns none (or `max_pages`, default
100, is reached — both bounds ported verbatim from legacy). `cataloger.go`'s `attribute` recursively
unwraps each item's DynamoDB `AttributeValue` envelope (`S`/`N`/`B` stringify — DynamoDB's own wire
convention represents numbers as decimal strings; `BOOL`/`NULL` map directly; `M`/`L` recurse into a
nested record/slice) into a plain `connectors.Record`, exactly matching legacy's `flattenItem`/
`attribute`. There is no incremental cursor: legacy never modeled one (a `Scan` has no
`updated_at`-style server-side filter), so every sync is a full snapshot, matching legacy exactly.

A `mode=fixture` config (`cfg.Config["mode"]=="fixture"`) short-circuits all network access,
mirroring legacy's identical `fixtureMode` gate: `Check` succeeds trivially, `Catalog` returns the
static `items` stream, and `Read` emits the same two deterministic fixture records legacy's
`readFixture` generates (`{"pk": "fixture#1", ...}`/`{"pk": "fixture#2", ...}`).

## Write actions & risks

None — DynamoDB is read-only here. `capabilities.write: false`, no `writes.json` file, matching
legacy's `ErrUnsupportedOperation` (`dynamodb.go:128-130`). **Note on the catalog label**: the
in-repo `catalog_data.json` carries BOTH a `"source-dynamodb"` and a `"destination-dynamodb"` slug
(stale Airbyte-slug residue predating this migration) — the `destination-` label does not reflect
any write capability legacy ever implemented; `internal/connectors/dynamodb/dynamodb.go`'s `Write`
method has always returned `ErrUnsupportedOperation` unconditionally, and its `Metadata()` has always
declared `Capabilities.Write: false`. This migration follows the legacy CODE (read-only), not the
catalog label, per the ground-truth rule.

## Known limits

- **No CDC.** DynamoDB Streams (a distinct, separate AWS API for change-data-capture) is not
  implemented — legacy never touched it either, so unlike `native/postgres` (which documents a
  genuine `pglogrepl`-gated CDC stub as recorded future scope), there is no `cdc.go` file at all in
  this package; a CDC stub would misrepresent a capability legacy never scoped, let alone attempted.
- **DynamoDB's `page_size`/`max_pages` ranges are not schema-validated.** Legacy's `intConfig`
  rejects a non-positive-integer value with a config error before the first request; this bundle's
  `spec.json` only types these as strings with a default — the same "range validation lives in Go,
  not the schema" shape every other native connector (postgres's `sslmode` enum aside) already
  accepts. Ported verbatim in `connection.go`'s `intConfig`, so the SAME validation legacy performed
  still runs — this is a `spec.json`-documentation note, not a behavior gap.
- **`endpoint`/`table_name` required-ness is Go-enforced, not schema-enforced.** Legacy accepts
  either `endpoint` or `base_url`, and either `table_name` or `table` (first-non-empty-wins
  fallback pairs) — JSON Schema's `required[]` cannot express an either-or constraint, so neither
  individual key is schema-required; `connection.go`'s `resolveEndpoint` and `reader.go`'s
  `tableName` enforce the real "at least one of the pair" rule at read/check time, exactly matching
  legacy's own runtime validation.
