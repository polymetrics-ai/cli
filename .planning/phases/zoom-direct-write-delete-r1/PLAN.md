# Plan — Zoom direct-write and delete proving cohort

Issue: #4268
Base: `origin/main` at `31bfe62eb` (`fix(certification): derive operation write action kinds`)

## Scope

The intended source cohort has no valid post-foundation endpoint coverage representation. All 61 non-upload REST writes, including 18 operation-backed deletes, are therefore deferred as recoverable `foundation-gap` entries pending the firstmate-owned coverage foundation lane. Eight load-bearing upload/multipart operations are separately recorded as `schema-incompatible` plus a recoverable foundation gap. This wave changes no production command surface.

Source inspection shows 18 write declarations carry the retired operation-scoped `rest.auth`/`rest.base_url` shape and six declare multipart content. A straight copy is invalid: retained REST declarations must use the existing Zoom credential contract without a shared-auth change; multipart/upload work stays explicitly pending the binary executor gap.

### Standing source-compatibility rule

For every unsupported legacy source field: (1) drop it and keep the operation only when operation semantics remain honest; document the dropped field in the PR; (2) otherwise reject the exact operation as `schema-incompatible`, with field, validation evidence, recoverability, and recovery condition; (3) also fan out a `foundation-gap` record when the missing schema support is a genuine engine capability, naming the smallest unimplemented change. Never invent a replacement field, coerce the field, or patch shared foundation code in this lane.

The deleted-operation selector must be exercised against the merged mutation-class projection fix. Destructive commands stay implemented-and-uncertified unless a run-owned create/readback/cleanup ledger proves live safety.

## Expanded parity contract

The five proving waves are not the delivery endpoint. The parent must ultimately implement over 90% of documented Zoom rows, or list every remaining row in a connector-local rejection record with its exact operation, fixed-vocabulary reason, evidence, recoverability, and recovery condition. `foundation-gap` entries also feed the foundation-gap log. This wave will establish the generated/listable accounting shape while long-tail waves populate it.

## TDD slices

1. **Red:** a connector-local test expects 69 `rest_write` operations, 18 deletes whose generated sweep action kind is `delete`, and the compatible write fixture corpus; baseline has none.
2. **Green:** port only Zoom-local declarations/fixtures, synchronize generated surfaces, and prove preflight, plan/preview/approval, destructive selection, and loopback fixture behavior.
3. **Certification honesty:** fixture proof is recorded; live mutations require a disposable create/readback/cleanup ledger and are otherwise pending certification.
4. **Long-tail accounting:** derive the remaining documented operations into a rejection list, with evidence/recovery and foundation-gap fan-out, before the parent declares parity.

## Skills and fallback

Used: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-documentation`.

The canonical lifecycle is executed inline because the delivery contract forbids role delegation in this parent-owned stacked workflow. Red/green evidence, verification, and review will be retained in this phase directory.
