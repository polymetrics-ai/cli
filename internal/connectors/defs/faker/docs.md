# Overview

Sample Data (`faker`) is a Tier-3 native source connector (design §B.7): it generates
deterministic, in-process sample `users`, `purchases`, and `products` records without making any
network calls at all. The legacy package (`internal/connectors/faker`) is not connsdk-HTTP-based —
it has no `connsdk.Requester`, no auth, no wire protocol whatsoever, just a pure Go loop deriving
field values from `count`/`seed`/a loop index — so per `docs/migration/conventions.md`'s tier
ladder (§6 item 3, "the protocol is not HTTP/REST at all") this migrates to a Tier-3 native
component split, following `internal/connectors/native/postgres/` as the golden pattern for a
non-REST connector, rather than a declarative Tier-1 bundle: there is no HTTP request/response
shape, path, pagination, or records-at-path for `streams.json` to declare, since nothing is ever
requested over a wire. `capabilities.dynamic_schema: true` is set for the same structural reason
postgres sets it — this bundle ships **no `streams.json`** — even though, unlike postgres, this
connector's 3 streams are not actually schema-discovered at runtime from an external source; they
are simply a fixed, hand-written Go catalog with no declarative equivalent to express (a
"generate N records from a counter/seed formula" primitive does not exist anywhere in the engine
dialect). This bundle is parity-tested against the legacy package, which stays registered and
unchanged until wave6's registry flip.

## Auth setup

None. This connector requires no credentials, tokens, or connection details of any kind — it never
leaves the local process. `spec.json` declares no `x-secret` fields.

## Streams notes

Three streams, matching legacy's `Catalog` exactly:

- `users` — primary key `id`, cursor field `updated_at`. Emits `count` records (config `count`,
  default 1000): `id`/`name` are zero-padded `user_%03d` / `User %03d` on `seed + i`, `email` is
  `user%03d@example.com`, `updated_at` cycles through `2026-01-01`..`2026-01-28` by `(i-1) % 28`.
- `purchases` — primary key `id`, cursor field `updated_at`. Emits `count` records tying each
  purchase to a generated user (`user_id`) and one of 10 generated products (`product_id`, cycling
  `(i % 10) + 1`); `amount` is `((i % 10) + 1) * 9.99`.
- `products` — primary key `id`, no incremental cursor (matches legacy: legacy's `Catalog` declares
  no `CursorFields` for this stream). Always emits exactly 10 records regardless of `count`
  (`sku`/`name`/`price` derived from a fixed 1-10 loop), matching legacy's `Read`'s `"products"`
  case verbatim (`faker.go:84-92`).

`count` defaults to 1000 and must be a positive integer (legacy: `positiveInt`, `faker.go:103-112`);
`seed` defaults to 0, must parse as an integer, and negative values are silently clamped to 0
(legacy: `faker.go:56-62`) — both behaviors are ported rule-for-rule into `reader.go`'s config
parsing, not re-derived.

## Write actions & risks

None. `capabilities.write` is `false` and `Write` always returns
`connectors.ErrUnsupportedOperation`, matching legacy exactly.

## Known limits

- **No fixture-mode short-circuit exists, unlike the HTTP-based Tier-1/Tier-2 goldens.** This
  connector never makes a network call in the first place — `Check`/`Catalog`/`Read` are
  identically pure/deterministic in every runtime, so there is no live-vs-fixture branch to model
  (mirrored in the parity/unit test suite, which exercises the connector directly with no replay
  server). This bundle's `api_surface.json` declares zero endpoints for the same reason
  `native/postgres`'s does: there is no REST surface, so neither `covered_by` nor `excluded`
  entries apply.
- **`dynamic_schema: true` is a structural label, not a literal claim of runtime schema
  discovery.** Unlike postgres (which genuinely queries `information_schema` at runtime), this
  connector's 3-stream catalog is fixed and hand-written; the flag is set purely because the
  Tier-3 loader requires it whenever a bundle ships no `streams.json` (`bundle.go`'s `loadStreams`
  only tolerates a missing `streams.json` when `dynamic_schema` is true) — there is no
  declarative-JSON equivalent for a synthetic per-record generation formula, so a real
  `streams.json` was never an option here regardless of how static the catalog actually is.
