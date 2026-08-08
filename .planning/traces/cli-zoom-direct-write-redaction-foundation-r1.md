# Direct-write redaction foundation trace — Zoom CMK Hybrid R1

## Why this foundation belongs in this slice

Zoom's Customer Managed Keys Hybrid archival API is a documented POST whose successful response
can contain plaintext key material. The connector may not classify it as unavailable or unsafe;
it must use a typed declarative operation with a redacted output policy. Before this slice, the
`rest_write` executor accepted `json_redacted` but returned full decoded response content, and
direct-write plan samples ignored declared sensitive fields. That violates the operation contract.

## Bounded change

The foundation is limited to declared `rest_write` execution and plan presentation:

- honor redacting direct-write output policies for generic and operation-declared fields;
- redact declared request literals from direct-write error text;
- make direct-write plan samples use declared operation-sensitive fields while retaining the raw
  private execution record; and
- convert `sensitive_policy.approval_mode=typed_confirmation` into the existing closed
  confirmation gate for secret-sensitive operations.

No generic HTTP write, arbitrary host discovery, token synthesis, secret logging, or connector
specific condition is introduced. The existing normalized path/base URL, conditional auth, preview,
single-use approval, and destructive transport policy are reused.

## Consumers unblocked

Any future connector with a legitimate secret-returning declared `rest_write` response can use
`json_redacted` plus `sensitive_policy.redact_fields` without exposing response or echoed-request
secret material. Zoom CMK Hybrid is the first consumer.

## TDD evidence

- RED: `5a0172053`, captured in the CMK Hybrid phase ledger before any foundation or Zoom
  production declaration.
- GREEN: targeted engine, commandrunner, and app lifecycle tests pass. The foundation commit ID
  is recorded after its separate commit is created.
