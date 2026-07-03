# Overview

Amazon SQS is a Tier-3 native source connector (design §B.7) — catalog-labeled
native/destination, but that label is WRONG on both counts: legacy
`internal/connectors/amazon-sqs/amazon_sqs.go` is read-only (`Write` always returns
`ErrUnsupportedOperation`, `capabilities.write: false`), never a destination. It IS genuinely
Tier-3, though, for a different reason: it signs AWS SQS Query API requests with hand-rolled
AWS SigV4 (canonical request construction, HMAC-SHA256 signing-key derivation) and decodes an XML
`ReceiveMessageResponse` body — neither of which the engine's declarative HTTP dialect (bearer/
basic/api-key/OAuth2 `auth` modes, JSON body typing) can express; SigV4 is exactly the "signature
auth (SigV4, HMAC)" example the tier ladder names as a legitimate protocol-native trigger, and it
recurs alongside an XML (not JSON) response envelope, a second engine-inexpressible shape. This
package speaks the SQS Query API directly, following `internal/connectors/native/postgres/`'s
component split as the golden pattern. The legacy package stays registered and unchanged until
wave6's registry flip; this bundle is parity-tested against it in fixture mode.

## Auth setup

Provide `queue_url` (the full SQS queue URL) and `region` in config, and an AWS `access_key`/
`secret_key` pair as secrets (never logged); an optional `session_token` secret is sent as the
`X-Amz-Security-Token` header for temporary/STS credentials. Every request is signed with AWS
SigV4 (`AWS4-HMAC-SHA256`), computed fresh per request from the current UTC time (or an injectable
clock in tests) — there is no token exchange or refresh; the same access/secret key pair signs
every request for the life of the connector.

## Streams notes

A single `messages` stream (primary key `message_id`) calls `ReceiveMessage` in a bounded poll
loop: each iteration requests `max_batch_size` messages (default 10, clamped 1-10),
`max_wait_time` long-poll seconds (default 0, clamped 0-20), an optional `visibility_timeout`
(clamped 0-43200) when configured, and `attributes_to_return` (default `All`) as
`MessageAttributeName`s; `AttributeName.1=All` is always sent to request every system attribute.
The loop runs up to `max_polls` times (default 1, clamped 1-100) and stops early the moment a poll
returns zero messages — matching legacy's exact loop-and-early-stop shape
(`amazon_sqs.go` `Read`). Every emitted record carries `message_id`, `md5_of_body`,
`receipt_handle`, and a `body` field that is JSON-decoded when the message body parses as JSON
(else left as the raw string) — matching legacy's `parseBody`. Every SQS system Attribute and
MessageAttribute is projected onto the record under its own snake_cased name (e.g.
`SentTimestamp` -> `sent_timestamp`), identical to legacy's `messageRecord`/`snake` helpers. There
is no incremental cursor: SQS's `ReceiveMessage` has no timestamp/offset-based server-side filter,
and this connector never deletes messages, so repeated syncs may re-observe the same message
according to ordinary SQS visibility-timeout/redelivery semantics — this matches legacy's behavior
exactly (no `InitialState`/cursor tracking in the legacy package either).

`capabilities.dynamic_schema: true` is set for the same structural reason `native/faker`'s docs
explain: this bundle ships no `streams.json` (the Tier-3 loader requires `dynamic_schema: true`
whenever `streams.json` is absent), even though `messages` is a fixed, hand-written single-stream
catalog, not a runtime schema-discovery target the way postgres's tables are. There is no
declarative equivalent for SigV4 signing or XML-envelope decoding, so a real `streams.json` was
never an option regardless of how static the catalog actually is.

## Write actions & risks

None. This is a read-only source connector — `capabilities.write` is `false` and `Write` always
returns `ErrUnsupportedOperation`, matching legacy exactly. Note `ReceiveMessage` itself is not
side-effect-free per SQS semantics: receiving a message can still change its visibility (making it
temporarily invisible to other consumers) even though this connector never explicitly deletes or
mutates a message — this is documented in both the legacy package's doc comment and this bundle's
`metadata.json.risk.read` field, not a new risk introduced by migration.

## Known limits

- No CDC/streaming support: SQS is polled, not subscribed to, and legacy implements no CDC path
  either — there is no `ReadCDC`/`CDCReader` implementation, unlike postgres's documented stub;
  `capabilities.cdc` is `false`.
- `SendMessage`/`DeleteMessage`/`PurgeQueue` and every other SQS action are unimplemented (legacy
  never implemented them either). `api_surface.json` declares zero endpoints (matching postgres's/
  faker's Tier-3 minimal-surface pattern) rather than listing SQS actions as `covered_by`/`excluded`
  entries, since there is no `streams.json` for `covered_by.stream` to resolve against once a
  bundle ships none.
- Message re-delivery/at-least-once semantics are governed entirely by the target queue's own
  visibility timeout and redrive policy — this connector does not deduplicate, track per-message
  state, or delete messages after reading, identical to legacy.
- A `mode=fixture` config value short-circuits all network access (Check succeeds, Read emits two
  canned messages) — this is a test/conformance-harness affordance only and must never be set in
  production config, matching the postgres/faker goldens' identical fixture-mode convention.
