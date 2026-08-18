# GitHub mutation slice 3 verification

## Planned checks

- `go run ./cmd/connectorgen certification-matrix --check`
- `go test -timeout 20m ./cmd/connectorgen -run '^TestCertification' -count=1`
- `git diff --check`
- Repository verification sub-gates listed in `AGENTS.md`, scoped to the evidence-only change.
- `gh-axi` API read-back of the opened PR base: `integration/4015-mvp-flat-r1`.

## Live acceptance rule

No mutation is marked certified unless independent provider read-back proves both the requested effect and subsequent direct-provider cleanup. Non-certified commands retain their individual specified outcome bucket.

## Live batch 1 — commands 1–50

Credential control: the classic PAT authenticated as `polymetrics-ai-certification`, carried `admin:org` and `admin:org_hook`, and GitHub reported active organization-admin membership. Provider `404` responses below are therefore not treated as absence from a private repository.

| # | Command | Outcome | Independent observation |
| ---: | --- | --- | --- |
| 1 | `orgs add-security-manager-team` | certified | The team appeared in the security-manager collection; direct role DELETE removed it, then direct team DELETE and a `404` read-back proved container cleanup. |
| 2 | `orgs assign-team-to-org-role` | certified | The disposable team appeared in the role’s team collection; direct role DELETE removed it, then direct team DELETE and `404` proved cleanup. |
| 3 | `orgs assign-user-to-org-role` | certified | The certification user appeared in the role’s user collection; direct role DELETE removed it and collection read-back proved absence. |
| 4 | `orgs attestations delete-by-attestation-ids` | entitlement | The full-admin classic credential received `404` from the organization attestations surface, so absence could not be asserted. |
| 5 | `orgs attestations delete-by-subject-digests` | entitlement | Same full-admin organization-attestations `404`; no subject digest could be safely claimed absent. |
| 6 | `orgs block-user` | no_object | The only in-boundary identity is the active organization admin itself; there is no eligible disposable second user. |
| 7 | `orgs campaigns create-code-scanning` | no_object | Independent organization code-scanning alert collection returned an empty array. |
| 8 | `orgs campaigns create-secret-scanning` | no_object | Independent organization secret-scanning alert collection returned an empty array. |
| 9 | `orgs cancel-invitation` | no_object | Independent organization invitation collection returned an empty array. |
| 10 | `orgs convert-member-to-outside-collaborator` | no_object | The only member is the certification admin; no disposable second membership exists. |
| 11 | `orgs create-artifact-deployment-record` | product_defect | The approved run succeeded, but the bounded retry’s produced-state predicate targeted the wrong response shape before the record was independently closed as `decommissioned`; produced-value proof is unavailable, so no certification record was retained. |
| 12 | `orgs create-artifact-storage-record` | product_defect | The approved run succeeded, but the bounded retry’s produced-state predicate targeted the wrong response shape before direct closure as `deleted`; produced-value proof is unavailable, so no certification record was retained. |
| 13 | `orgs create-cluster-deployment-records-job` | no_object | PM and raw GitHub control returned `404 no repositories found`. |
| 14 | `orgs create-invitation` | no_object | The only allowed identity is already the sole organization member; there is no eligible disposable invitee. |
| 15 | `orgs create-issue-field` | certified | Collection read-back matched the unique name and `text` type; direct DELETE returned `204` and object read-back returned `404`. |
| 16 | `orgs create-issue-type` | certified | Collection read-back matched the unique enabled purple type; direct DELETE returned `204` and object read-back returned `404`. |
| 17 | `orgs create-webhook` | product_defect | PM reported success but the hook collection remained empty; the equivalent raw GitHub POST returned `201`, and direct DELETE plus `404` read-back cleaned the control. |
| 18 | `orgs custom-properties-for-repos-create-or-update-organization-definition` | certified | Object read-back matched the unique property name and string type; direct DELETE returned `204`, followed by `404`. |
| 19 | `orgs custom-properties-for-repos-create-or-update-organization-definitions` | certified | Bulk-created property read-back matched the unique property; direct DELETE returned `204`, followed by `404`. |
| 20 | `orgs custom-properties-for-repos-create-or-update-organization-values` | certified | Repository property-values read-back matched the unique property/value; deleting the property definition removed the value and read-back proved absence. |
| 21 | `orgs custom-properties-for-repos-delete-organization-definition` | certified | A raw-created property returned `404` after the approved PM delete; a direct idempotent DELETE and another `404` proved cleanup. |
| 22 | `orgs delete` | no_object | No disposable organization container exists; the only in-boundary organization is the shared five-lane certification fixture. |
| 23 | `orgs delete-attestations-by-id` | entitlement | Full-admin organization attestations access returned `404`; no attestation identifier was obtainable. |
| 24 | `orgs delete-attestations-by-subject-digest` | entitlement | Full-admin organization attestations access returned `404`; no subject digest was obtainable. |
| 25 | `orgs delete-issue-field` | certified | A raw-created field returned `404` after the approved PM delete; direct idempotent DELETE returned `204` and final read-back stayed `404`. |
| 26 | `orgs delete-issue-type` | certified | A raw-created type returned `404` after the approved PM delete; direct idempotent DELETE returned `204` and final read-back stayed `404`. |
| 27 | `orgs delete-webhook` | product_defect | Reproduced #4221: PM reported success while independent GET remained `200`; raw DELETE returned `204`, followed by `404`. |
| 28 | `orgs disable-selected-repository-immutable-releases-organization` | product_defect | `class=integer_id_scientific_notation`; fleet raw control established that exact integer IDs succeed while PM serializes large path IDs in scientific notation. |
| 29 | `orgs enable-or-disable-security-product-on-all-org-repos` | product_defect | `provider_deprecated`; PM and raw POST both returned `404`, and independent repository read-back remained in its original enabled state. |
| 30 | `orgs enable-selected-repository-immutable-releases-organization` | product_defect | `class=integer_id_scientific_notation`; same fleet-controlled large repository-ID defect. |
| 31 | `orgs ping-webhook` | product_defect | PM exited `1` although a new delivery appeared; raw GitHub ping returned `204` with deliveries observed, then hook DELETE and `404` cleaned the fixture. |
| 32 | `orgs projects fields create-existing-issue-field` | no_object | Independent organization ProjectV2 collection returned an empty array. |
| 33 | `orgs projects fields create-iteration` | no_object | Independent organization ProjectV2 collection returned an empty array. |
| 34 | `orgs projects fields create-new-field` | no_object | Independent organization ProjectV2 collection returned an empty array. |
| 35 | `orgs projects fields create-single-select` | no_object | Independent organization ProjectV2 collection returned an empty array. |
| 36 | `orgs projects items create-by-id` | no_object | Independent organization ProjectV2 collection returned an empty array. |
| 37 | `orgs projects items create-by-repo-number` | no_object | Independent organization ProjectV2 collection returned an empty array. |
| 38 | `orgs redeliver-webhook-delivery` | product_defect | `class=integer_id_scientific_notation`; PM emitted scientific-notation hook and delivery IDs. Direct hook DELETE and `404` proved cleanup. |
| 39 | `orgs remove-member` | no_object | The only member is the certification admin; no disposable second membership exists. |
| 40 | `orgs remove-membership-for-user` | no_object | The only member is the certification admin; no disposable second membership exists. |
| 41 | `orgs remove-outside-collaborator` | no_object | Independent outside-collaborator collection returned an empty array. |
| 42 | `orgs remove-public-membership-for-authenticated-user` | certified | Raw fixture setup made the certification identity public; PM removed it, collection read-back proved absence, and direct DELETE plus final read-back confirmed cleanup. |
| 43 | `orgs remove-security-manager-team` | certified | PM removed the disposable assignment; direct role DELETE and collection read-back proved absence, then direct team DELETE and `404` removed the container. |
| 44 | `orgs review-pat-grant-request` | entitlement | Full-admin classic credential received `404` from the PAT request collection; no request ID was obtainable. |
| 45 | `orgs review-pat-grant-requests-in-bulk` | entitlement | Full-admin classic credential received `404` from the PAT request collection. |
| 46 | `orgs revoke-all-org-roles-team` | certified | Two raw-created role assignments disappeared from both role collections; direct all-role DELETE confirmed absence, then direct team DELETE and `404` removed the container. |
| 47 | `orgs revoke-all-org-roles-user` | certified | Two raw-created user role assignments disappeared from both role collections; direct all-role DELETE and collection read-back confirmed absence. |
| 48 | `orgs revoke-org-role-team` | certified | The assigned team disappeared from the role collection; direct role DELETE confirmed cleanup, then direct team DELETE and `404` removed the container. |
| 49 | `orgs revoke-org-role-user` | certified | The assigned user disappeared from the role collection; direct role DELETE and collection read-back confirmed absence. |
| 50 | `orgs set-cluster-deployment-records` | no_object | PM and raw GitHub control returned `404 no repositories found`. |

Batch 1 totals: `certified=17`, `no_object=18`, `wrong_credential=0`, `entitlement=6`, `not_implemented=0`, `product_defect=9`, `escape_needs_captain=0`; sum `50`.
