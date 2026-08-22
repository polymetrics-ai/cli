# Refs #4303 — Action-binding canonical repair R2 discussion log

## Delivery mode

`/gsd-discuss-phase 4303 --auto` was generated and the auto-mode source was read. The registered roadmap has no phase `4303`, so the issue-local manual fallback is used. The Firstmate launch brief explicitly authorizes autonomous execution; no product choice is left open because the immutable final review supplies every required outcome.

## Auto-resolved areas

| Area | Selected decision | Authority |
| --- | --- | --- |
| Physical action disclosure | Persist and digest-bind all executable actions, including tombstone delete semantics, before approval. | AB-B01 |
| Action-read-back compatibility | Bind policy and maximums to action; reject mismatches before I/O. | AB-B02, AB-B04 |
| Receipt and output safety | Size sealed units before I/O and retain sanitized provider output after local post-success failures. | AB-B03, AB-B05–AB-B08 |
| Identity and declaration closure | Bind idempotency to durable worksets and make optional/schema/strategy/binding contracts closed. | AB-B09, AB-W01–AB-W04 |
| Operator parity | Update help, manual, website, and generated surfaces together. | AB-W05 |
