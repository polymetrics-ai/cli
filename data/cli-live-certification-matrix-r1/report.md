# Live certification matrix r1

## Purpose and first slice

Certification is a generated, proof-bearing quality signal. It does **not**
hide an uncertified connector: `pm connectors inspect <name>` shows either
`CERTIFIED`, or `COMMUNITY BUILD, UNCERTIFIED` with a plain warning. The status
is embedded in the CLI from a generated projection and is what a website should
render. Connector availability is unchanged.

A normal user sees a CERTIFICATION section in the connector inspect command;
automation receives the same certification object from JSON output. The website
renders that generated label and warning directly in its connector reference:
there is no third status vocabulary to reinterpret.

The first shippable slice is deliberately small: a generator, checker,
publishable-proof boundary, status renderer, and an all-red honest baseline. It
enables the next GitHub and PostgreSQL live runs without pretending either has
happened in this PR. It defers broad live coverage, a database write executor,
API-destination durability, and GitHub resumable delivery.

## Certification record and proof

Only `internal/connectors/certifications/evidence/*.json` is accepted live
evidence. The generator reads that record first, validates its embedded proof,
and only then derives and cross-checks the matrices against code. Existing
`certification.json` files remain untouched and are expressly listed as legacy
non-evidence (11 bundle contracts and 6 fixture/schema files in this checkout).

The following is a **schema example, not evidence**: no live GitHub run is
claimed by this branch. An accepted record has the same shape, but can only be
written through the completed-live-run boundary after a real `pm` invocation.
The fake fingerprints demonstrate the publishable format rather than claiming
a request occurred.

```json
{
  "schema_version": 1,
  "scope": "capability",
  "status": "passed",
  "credential_scope": "full_parity",
  "credential_note": "Certification used a full-parity credential; a narrower credential exposes a subset of this certified surface.",
  "connector": "github",
  "function_kind": "operation:rest_write",
  "provider": "github",
  "executed_at": "2026-08-10T12:00:00Z",
  "run_id": "github-full-parity-example",
  "proof": {
    "redaction_strategy": "repository_salted_hmac_sha256_v1",
    "pm_binary_sha256": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
    "pm_command_fingerprint": "{{pmcertfp:v1:1111111111111111111111111111111111111111111111111111111111111111}}",
    "credential_fingerprints": ["{{pmcertfp:v1:2222222222222222222222222222222222222222222222222222222222222222}}"],
    "http_exchanges": [{
      "operation": "github.issues.create",
      "request": {
        "method": "POST",
        "target": "{{pmcertfp:v1:3333333333333333333333333333333333333333333333333333333333333333}}",
        "query": [],
        "headers": [
          {"name": "authorization", "value": "{{pmcertfp:v1:4444444444444444444444444444444444444444444444444444444444444444}}{{pmcertfp:v1:2222222222222222222222222222222222222222222222222222222222222222}}"},
          {"name": "content-type", "value": "{{pmcertfp:v1:5555555555555555555555555555555555555555555555555555555555555555}}"}
        ],
        "body": {"encoding": "json", "value": {"title": "{{pmcertfp:v1:6666666666666666666666666666666666666666666666666666666666666666}}"}, "original_bytes": 31, "truncated": false}
      },
      "response": {
        "status": 201,
        "headers": [{"name": "x-github-request-id", "value": "{{pmcertfp:v1:7777777777777777777777777777777777777777777777777777777777777777}}"}],
        "body": {"encoding": "json", "value": {"id": "{{pmcertfp:v1:8888888888888888888888888888888888888888888888888888888888888888}}"}, "original_bytes": 18, "truncated": false}
      }
    }],
    "database_exchanges": []
  }
}
```

The complete request/response transcript is represented: method, exact target,
query/header names and values, body shape and values, status, response headers,
and response body. Values are HMAC fingerprints unless their safety is proven;
the transcript remains locally re-verifiable and publishable while unknown
response-body data cannot leak. A normal user supplies their credential for
replay, fingerprints it with this checkout's salt, and compares it to
`credential_fingerprints` or the corresponding transcript value.

Fingerprints use deterministic `HMAC-SHA-256(repository-local-salt, exact
prepared value)`. The salt is generated with `crypto/rand` under
`internal/connectors/certifications/.fingerprint-salt`, mode `0600`, and is
git-ignored. Every checkout gets a different salt; the salt and raw credential
never serialize. Sanitization receives the actual prepared values, substitutes
before a JSON serializer or output file is opened, fingerprints all unproved
scalar request and response values, and never relies on a keyword list.

## Matrix shape

`cmd/connectorgen/certificationallowlist.go` declares the connectors for which
this repository currently makes a proof-bearing claim: initially GitHub and
PostgreSQL. `go run ./cmd/connectorgen certification-matrix --connector <name>`
generates the one matching, non-embedded
`internal/connectors/defs/<name>/certification-matrix.json` shard; it does not
rewrite another connector's shard or the compact runtime status projection.
`--all` deliberately refreshes all allowlisted shards and that projection, and
`--check` reconstructs the aggregate view in memory for the drift gate.

Each shard discovers all capability fields and engine operation kinds from
source for its connector. A direct unsupported method such as PostgreSQL
`Write` is applicable with `implemented=false`, never hidden as N/A. A new
engine operation kind enters the inventory automatically; without an executor
annotation it remains honestly unimplemented. Source anchors use
`relative/path.go:Symbol`, never a line number.

Each shard carries its connector's final certification obligations:

- Workflows are source-discovered from real command handlers: `etl`,
  `reverse_etl`, `flow_authoring`, and `schedule`. All applicable workflows
  need their own fixture/live proof; command reachability does not complete
  flow authoring or scheduling.
- Sync modes come from `internal/synccontract/mode.go` and cross every
  connector with all four stable warehouse-facing primitives:
  `api_read_into_warehouse`, `api_write_from_warehouse`,
  `database_read_into_warehouse`, and `database_write_from_warehouse`.
  Every mode/primitive combination has an explicit cell. `change_capture` is
  applicable only to `database_read_into_warehouse`; every other combination
  carries a named machine-readable non-applicability reason.
- Flows are exact source/destination pairs within the active allowlist with mandatory
  `local_parquet_warehouse` mediation: API→warehouse→API,
  API→warehouse→database, database→warehouse→API, and
  database→warehouse→database. The in-memory aggregate uses non-overlapping
  compressed pair sets for validation, but no shard stores them; its resolver
  proves every exact pair has one cell and evidence overrides apply to one exact
  pair only. Thus
  `github → warehouse → github` is valid and distinct.

Flow proof embeds both independent warehouse and destination readback
operations. It additionally records delivery facts—resumable, receipt-backed,
checkpointed, replay identity, and provider idempotency key—with a named
limitation for every false fact. A working one-shot GitHub reverse mutation can
therefore remain explicitly non-resumable rather than be advertised as durable.

An allowlisted connector is `CERTIFIED` only if every applicable capability, workflow,
sync-mode/warehouse-primitive, and source-or-destination flow-pair cell is
declared, implemented, fixture-tested, and backed by accepted live proof.

## Historical aggregate baseline

Before connector-local sharding, the aggregate generator's baseline was intentionally red: 556 connectors, 21 discovered
function kinds, zero capability-complete connectors, and zero finally certified
connectors. All live-tested totals are zero because this change runs no provider
and accepts no invented evidence.

GitHub's declared `operation:graphql_query` now correctly reports one
implemented cell through the bounded GraphQL direct-read executor. Its declared
`operation:local_git` remains one applicable but unimplemented cell. Neither
had fixture or live proof, so neither changed the historical zero-certification baseline.

The historical aggregate flow-kind totals are 309,136 exact pairs each. Their applicable /
implemented / complete counts are:

| Flow | Applicable | Implemented | Complete |
| --- | ---: | ---: | ---: |
| API → API | 301,401 | 0 | 0 |
| API → database | 2,196 | 0 | 0 |
| database → API | 2,196 | 0 | 0 |
| database → database | 16 | 0 | 0 |

The historical aggregate sync scoreboard is zero complete for all 28 mode × primitive cells. API
writes are red because the engine refuses an API destination without a durable
acknowledgement; database writes are red because no database write executor
exists. PostgreSQL and MySQL change-capture reads are code-present but have no
fixture or live proof, so they remain incomplete rather than being silently
promoted. Those failures are the useful baseline for the PostgreSQL waves, not
a reason to omit cells.
