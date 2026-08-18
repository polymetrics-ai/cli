# Mutation certification batch 2

This receipt covers the 50-command live execution batch `58`–`93` and `95`–`108`; command `94` was already certified. Every listed execution used plan, preview, and the one-time token through `--approval-token-stdin`. This is an intermediate receipt: rows explicitly listed as fixture-audit pending are not banked into a final bucket.

## Certified in this batch

Commands `69`, `71`, `79`, `80`, `83`, `92`, `93`, `95`, `102`, `103`, and `104` have validated schema-v2 records. Their assertions respectively proved runner-group repository access, selected variable repositories, variable update, runner-group update, pull-request creation, issue close/comment/delete, private user/org repository creation, and organization ruleset creation. Each owned container was directly deleted (or, for non-REST-deletable issues/PRs, provider-terminally deleted/closed) and independently read back absent/contained.

The empty-parent re-audit also converted commands `11`, `19`, and `20` from `no_object` to certified with freshly created sealed-secret/variable fixtures.

## Finalized non-pass observations

- `59`, `60`, `61`, and `76`: provider entitlement responses (`404`/`402`) with no state change.
- `64`, `78`, `81`, `82`, `84`–`87`, `89`–`91`, `96`–`98`, `100`, `101`, and `106`–`108`: `class=integer_id_scientific_notation`; the provider URL contained scientific notation for the large numeric path ID. The fleet-wide exact-integer raw control applies.

## Fixture audit still required before final classification

- `62`, `63`, `65`–`68`, and `75`: repeat with a deliberately different provider pre-state so a no-op cannot satisfy the assertion.
- `70`, `72`–`74`: create/read a real secret, selected-repository mode, or runner fixture before classifying.
- `77`: obtain the required raw provider control for the missing-body OIDC surface.
- `88` and `99`: use real contained PR/issue GraphQL node IDs.
- `105`: create a private disposable template repository, generate from it, and delete both repositories.

No pending row above is counted as `no_object` or `product_defect`.
