# Mutation certification batch 2

This receipt covers the 50-command live execution batch `58`–`93` and `95`–`108`; command `94` was already certified. Every listed execution used plan, preview, and the one-time token through `--approval-token-stdin`. The later fixture audit completed every initially pending row.

## Certified in this batch

Commands `63`, `65`–`71`, `73`–`75`, `79`, `80`, `83`, `88`, `92`–`95`, and `102`–`105` have validated schema-v2 records. Their independent assertions proved exact organization settings, secret/variable repository selection, runner-group state, pull-request and issue effects, private repository creation, organization ruleset creation, template generation, and revert state as applicable. Each owned container was directly deleted (or, for settings and non-REST-deletable issue/PR state, provider-terminally restored/deleted/closed) and independently read back absent or restored.

The empty-parent re-audit also converted commands `11`, `19`, and `20` from `no_object` to certified with freshly created sealed-secret/variable fixtures.

## Finalized non-pass observations

- `59`, `60`, `61`, and `76`: provider entitlement responses (`404`/`402`) with no state change.
- `64`, `78`, `81`, `82`, `84`–`87`, `89`–`91`, `96`–`98`, `100`, `101`, and `106`–`108`: `class=integer_id_scientific_notation`; the provider URL contained scientific notation for the large numeric path ID. The fleet-wide exact-integer raw control applies.
- `62`: `product_defect`; pm double-encoded the allowed-actions array, while a raw exact-array PUT succeeded and read back correctly.
- `72`: `product_defect`; pm reported success but left the selected-repository collection empty, while the raw exact-ID PUT succeeded and read back correctly.
- `77`: `product_defect`; pm sent no usable OIDC body, while raw PUT changed `use_default` and `include_claim_keys`, read-back proved the change, and raw restore read-back proved containment.
- `99`: `product_defect`; pm reported success without moving the issue, while raw GraphQL `transferIssue` moved an equivalent contained issue and independent destination read-back proved it.
