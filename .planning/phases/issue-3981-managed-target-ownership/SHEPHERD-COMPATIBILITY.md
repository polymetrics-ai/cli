# #3995 Shepherd-compatible verdict — Issue #3981

## Bounded result

`SHEPHERD-COMPATIBILITY.json` uses the version-1 #3995 verdict fields:
`schema_version`, `connector`, `transition`, `decision`, and `failures`. It is
**not** an automatic #3995 approval. The #3995 implementation at
`ef66991350a5de137ab17e492acf2640feecf5c3` is not an ancestor of this child or
`origin/feat/3972-postgres-parity`, so no evaluator was imported, replayed, or
claimed as having reviewed this change.

The bounded compatibility decision is `RETRY` only for the future
`integrate_sub_pr` certification transition. PostgreSQL remains `COMMUNITY
BUILD, UNCERTIFIED` with `capability_complete=false`; #3981 supplies only the
shared owner/provisioning contract and must not change that status. This result
does not reject the unmerged F2 child, nor does it authorize a merge or a
capability promotion.

The record is read-only and uses the final generated certification checks. It
contains no credential, provider interaction, or live database proof. #4038 is
the separate in-scope review correction already counted in this issue; the
non-integrated #3995 compatibility result does not consume another correction
round.
