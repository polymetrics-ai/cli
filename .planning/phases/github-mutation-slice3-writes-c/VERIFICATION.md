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
| 6 | `orgs block-user` | no_object | The member collection contains only the active certification admin. GitHub exposes no API for creating another user identity, so an eligible in-boundary block target cannot be created. |
| 7 | `orgs campaigns create-code-scanning` | no_object | The parent code-scanning alert collection is empty, and GitHub exposes no API for creating a synthetic code-scanning alert; no eligible campaign fixture can be created. |
| 8 | `orgs campaigns create-secret-scanning` | no_object | The parent secret-scanning alert collection is empty, and GitHub exposes no API for creating a synthetic secret-scanning alert; no eligible campaign fixture can be created. |
| 9 | `orgs cancel-invitation` | no_object | The invitation collection is empty and the only allowed user is already a member. GitHub exposes no API for creating a second disposable user identity, so no cancellable invitation can be created. |
| 10 | `orgs convert-member-to-outside-collaborator` | no_object | The member collection contains only the certification admin. No second disposable membership can be created without involving a real person outside the authorized identities. |
| 11 | `orgs create-artifact-deployment-record` | certified | A fresh record was independently read back by digest with the requested logical environment, cluster, and deployment name; the provider-native terminal disposal set `decommissioned`, and a later read-back proved the record was updated. |
| 12 | `orgs create-artifact-storage-record` | certified | A fresh record was independently read back by digest with its requested name, repository, and `active` status; provider-native disposal set `deleted`, and read-back proved the terminal status. |
| 13 | `orgs create-cluster-deployment-records-job` | certified | A real fixture repository was supplied in the deployment entry; digest read-back proved the background job created the requested cluster record. Provider-native `decommissioned` disposal was then applied and read back. |
| 14 | `orgs create-invitation` | no_object | The invitation collection is empty and the only allowed user is already the sole organization member. GitHub exposes no API for creating another disposable identity, so no eligible invitee can be created. |
| 15 | `orgs create-issue-field` | certified | Collection read-back matched the unique name and `text` type; direct DELETE returned `204` and object read-back returned `404`. |
| 16 | `orgs create-issue-type` | certified | Collection read-back matched the unique enabled purple type; direct DELETE returned `204` and object read-back returned `404`. |
| 17 | `orgs create-webhook` | product_defect | PM reported success but the hook collection remained empty; the equivalent raw GitHub POST returned `201`, and direct DELETE plus `404` read-back cleaned the control. |
| 18 | `orgs custom-properties-for-repos-create-or-update-organization-definition` | certified | Object read-back matched the unique property name and string type; direct DELETE returned `204`, followed by `404`. |
| 19 | `orgs custom-properties-for-repos-create-or-update-organization-definitions` | certified | Bulk-created property read-back matched the unique property; direct DELETE returned `204`, followed by `404`. |
| 20 | `orgs custom-properties-for-repos-create-or-update-organization-values` | certified | Repository property-values read-back matched the unique property/value; deleting the property definition removed the value and read-back proved absence. |
| 21 | `orgs custom-properties-for-repos-delete-organization-definition` | certified | A raw-created property returned `404` after the approved PM delete; a direct idempotent DELETE and another `404` proved cleanup. |
| 22 | `orgs delete` | no_object | The organization collection contains only the shared certification organization. GitHub exposes no API for creating a disposable organization container, so an independently cleanable target cannot be created. |
| 23 | `orgs delete-attestations-by-id` | entitlement | Full-admin organization attestations access returned `404`; no attestation identifier was obtainable. |
| 24 | `orgs delete-attestations-by-subject-digest` | entitlement | Full-admin organization attestations access returned `404`; no subject digest was obtainable. |
| 25 | `orgs delete-issue-field` | certified | A raw-created field returned `404` after the approved PM delete; direct idempotent DELETE returned `204` and final read-back stayed `404`. |
| 26 | `orgs delete-issue-type` | certified | A raw-created type returned `404` after the approved PM delete; direct idempotent DELETE returned `204` and final read-back stayed `404`. |
| 27 | `orgs delete-webhook` | product_defect | Reproduced #4221: PM reported success while independent GET remained `200`; raw DELETE returned `204`, followed by `404`. |
| 28 | `orgs disable-selected-repository-immutable-releases-organization` | product_defect | `class=integer_id_scientific_notation`; fleet raw control established that exact integer IDs succeed while PM serializes large path IDs in scientific notation. |
| 29 | `orgs enable-or-disable-security-product-on-all-org-repos` | product_defect | `provider_deprecated`; PM and raw POST both returned `404`, and independent repository read-back remained in its original enabled state. |
| 30 | `orgs enable-selected-repository-immutable-releases-organization` | product_defect | `class=integer_id_scientific_notation`; same fleet-controlled large repository-ID defect. |
| 31 | `orgs ping-webhook` | product_defect | PM exited `1` although a new delivery appeared; raw GitHub ping returned `204` with deliveries observed, then hook DELETE and `404` cleaned the fixture. |
| 32 | `orgs projects fields create-existing-issue-field` | certified | A fresh organization project and issue-field fixture were created; project-field read-back matched the issue field name and type. Direct issue-field DELETE and provider-native `deleteProjectV2`, each followed by `404`, proved cleanup. |
| 33 | `orgs projects fields create-iteration` | certified | A fresh organization project was created; field read-back matched the iteration name, type, and configuration. Provider-native `deleteProjectV2` and a REST `404` proved cleanup. |
| 34 | `orgs projects fields create-new-field` | certified | A fresh organization project was created; field read-back matched the requested text-field name and type. Provider-native `deleteProjectV2` and a REST `404` proved cleanup. |
| 35 | `orgs projects fields create-single-select` | certified | A fresh organization project was created; field read-back matched the single-select name, type, option name, and option color. Provider-native `deleteProjectV2` and a REST `404` proved cleanup. |
| 36 | `orgs projects items create-by-id` | certified | A fresh project and issue fixture were created; item read-back matched the issue ID, title, and content type. The project was deleted with `deleteProjectV2` and read back as `404`; the non-deletable issue was closed as `not_planned` and read back. |
| 37 | `orgs projects items create-by-repo-number` | certified | A fresh project and issue fixture were created; item read-back matched the repository issue title and content type. The project was deleted with `deleteProjectV2` and read back as `404`; the issue was closed as `not_planned` and read back. |
| 38 | `orgs redeliver-webhook-delivery` | product_defect | `class=integer_id_scientific_notation`; PM emitted scientific-notation hook and delivery IDs. Direct hook DELETE and `404` proved cleanup. |
| 39 | `orgs remove-member` | no_object | The member collection contains only the certification admin. GitHub exposes no API for creating a second user identity, so a removable membership cannot be created within the boundary. |
| 40 | `orgs remove-membership-for-user` | no_object | The membership collection contains only the certification admin. GitHub exposes no API for creating a second user identity, so a removable membership cannot be created within the boundary. |
| 41 | `orgs remove-outside-collaborator` | no_object | The outside-collaborator collection is empty. Creating one requires a second GitHub user identity, which GitHub exposes no API to create inside the authorized boundary. |
| 42 | `orgs remove-public-membership-for-authenticated-user` | certified | Raw fixture setup made the certification identity public; PM removed it, collection read-back proved absence, and direct DELETE plus final read-back confirmed cleanup. |
| 43 | `orgs remove-security-manager-team` | certified | PM removed the disposable assignment; direct role DELETE and collection read-back proved absence, then direct team DELETE and `404` removed the container. |
| 44 | `orgs review-pat-grant-request` | entitlement | Full-admin classic credential received `404` from the PAT request collection; no request ID was obtainable. |
| 45 | `orgs review-pat-grant-requests-in-bulk` | entitlement | Full-admin classic credential received `404` from the PAT request collection. |
| 46 | `orgs revoke-all-org-roles-team` | certified | Two raw-created role assignments disappeared from both role collections; direct all-role DELETE confirmed absence, then direct team DELETE and `404` removed the container. |
| 47 | `orgs revoke-all-org-roles-user` | certified | Two raw-created user role assignments disappeared from both role collections; direct all-role DELETE and collection read-back confirmed absence. |
| 48 | `orgs revoke-org-role-team` | certified | The assigned team disappeared from the role collection; direct role DELETE confirmed cleanup, then direct team DELETE and `404` removed the container. |
| 49 | `orgs revoke-org-role-user` | certified | The assigned user disappeared from the role collection; direct role DELETE and collection read-back confirmed absence. |
| 50 | `orgs set-cluster-deployment-records` | certified | A real fixture repository was supplied in the deployment entry; digest read-back matched the requested cluster and deployment. Provider-native `decommissioned` disposal was then applied and read back. |

Batch 1 totals after fixture and classification re-audit: `certified=27`, `no_object=10`, `wrong_credential=0`, `entitlement=6`, `not_implemented=0`, `product_defect=7`, `escape_needs_captain=0`; sum `50`.

## Live batch 2 — commands 51–64

| # | Command | Outcome | Independent observation |
| ---: | --- | --- | --- |
| 51 | `orgs set-immutable-releases-settings` | certified | Read-back proved enforcement changed to `all`; direct provider restoration returned it to `none`, and a second read-back proved the restored value. |
| 52 | `orgs set-immutable-releases-settings-repositories` | certified | After a contained `selected` setup, collection read-back matched the real fixture repository. Direct provider clearing plus restoration to `none` made the repository collection return `409`, proving absence. |
| 53 | `orgs set-membership-for-user` | certified | The only in-boundary membership was reused safely with role `admin`; object read-back matched `active/admin`, and direct provider restoration plus read-back preserved that strongest benign state. |
| 54 | `orgs set-public-membership-for-authenticated-user` | certified | The certification identity appeared in the public-member collection; direct DELETE returned `204`, and collection read-back proved zero matching members. |
| 55 | `orgs unblock-user` | no_object | The blocked-user collection is empty. GitHub exposes no API for creating a second disposable user identity, and blocking the sole organization admin would destroy the only authorized credential path. |
| 56 | `orgs update` | certified | Organization read-back proved `members_can_create_public_repositories=false`; direct PATCH restored `true`, and a second read-back proved restoration. |
| 57 | `orgs update-issue-field` | product_defect | `class=integer_id_scientific_notation`; PM emitted the real issue-field ID as scientific notation in the path. Raw PATCH with the exact integer returned `200`; direct DELETE and `404` proved cleanup. |
| 58 | `orgs update-issue-type` | product_defect | `class=integer_id_scientific_notation`; PM emitted the real issue-type ID as scientific notation in the path. Raw PUT with the exact integer returned `200`; direct DELETE and `404` proved cleanup. |
| 59 | `orgs update-membership-for-authenticated-user` | certified | Object read-back matched the only in-boundary membership as `active/admin`; direct provider restoration and a second read-back proved the benign terminal state. |
| 60 | `orgs update-pat-access` | entitlement | The full-admin classic credential received `404` from the parent fine-grained PAT collection, so no accessible PAT ID could be selected or created through GitHub's API. |
| 61 | `orgs update-pat-accesses` | entitlement | The full-admin classic credential received the same `404` from the parent fine-grained PAT collection; no accessible PAT IDs were available. |
| 62 | `orgs update-webhook` | product_defect | `class=integer_id_scientific_notation`; PM emitted the fresh hook ID in scientific notation. Raw PATCH with the exact integer updated active/events, then direct DELETE and `404` proved cleanup. |
| 63 | `orgs update-webhook-config-for-org` | product_defect | `class=integer_id_scientific_notation`; PM emitted the fresh hook ID in scientific notation. Raw PATCH with the exact integer updated the config, then direct DELETE and `404` proved cleanup. |
| 64 | `copilot add-copilot-seats-for-teams` | escape_needs_captain | Adding Copilot seats is a paid-seat mutation and therefore crosses the brief's explicit real-money escape boundary. It was not executed. |

Batch 2 totals: `certified=6`, `no_object=1`, `wrong_credential=0`, `entitlement=2`, `not_implemented=0`, `product_defect=4`, `escape_needs_captain=1`; sum `14`.

Cumulative commands 1–64: `certified=33`, `no_object=11`, `wrong_credential=0`, `entitlement=8`, `not_implemented=0`, `product_defect=11`, `escape_needs_captain=1`; sum `64`.
