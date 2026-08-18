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
| 64 | `copilot add-copilot-seats-for-teams` | escape_needs_captain | A Copilot seat is a recurring per-user subscription, an order of magnitude above the captain's USD 2 per-operation ceiling and recurring rather than one-off. No provider request was issued. |

Batch 2 totals: `certified=6`, `no_object=1`, `wrong_credential=0`, `entitlement=2`, `not_implemented=0`, `product_defect=4`, `escape_needs_captain=1`; sum `14`.

Cumulative commands 1–64: `certified=33`, `no_object=11`, `wrong_credential=0`, `entitlement=8`, `not_implemented=0`, `product_defect=11`, `escape_needs_captain=1`; sum `64`.

## Live batch 3 — commands 65–86

| # | Command | Outcome | Independent observation |
| ---: | --- | --- | --- |
| 65 | `copilot add-copilot-seats-for-users` | escape_needs_captain | A Copilot seat is a recurring per-user subscription, an order of magnitude above the captain's USD 2 per-operation ceiling and recurring rather than one-off. No provider request was issued. |
| 66 | `copilot add-organizations-to-enterprise-coding-agent-policy` | escape_needs_captain | The mutation targets an enterprise policy outside `Polymetrics-Cert`; no provider request was issued. |
| 67 | `copilot cancel-copilot-seat-assignment-for-teams` | escape_needs_captain | This changes recurring paid-seat assignments. No provider request was issued under the paid-seat ruling. |
| 68 | `copilot cancel-copilot-seat-assignment-for-users` | escape_needs_captain | This changes recurring paid-seat assignments. No provider request was issued under the paid-seat ruling. |
| 69 | `copilot disable-coding-agent-for-selected-repository-for-organization` | product_defect | `class=integer_id_scientific_notation`; PM emitted the real repository ID in scientific notation. Raw DELETE with the exact integer returned `204`, collection read-back proved zero selected repositories, and the organization policy was restored to `all`. |
| 70 | `copilot enable-coding-agent-for-selected-repository-for-organization` | product_defect | `class=integer_id_scientific_notation`; PM emitted the real repository ID in scientific notation. Raw PUT with the exact integer returned `204`, read-back matched the repository, then raw DELETE and collection read-back proved absence before policy restoration. |
| 71 | `copilot remove-organization-from-enterprise-coding-agent-policy` | escape_needs_captain | The mutation targets an enterprise policy outside `Polymetrics-Cert`; no provider request was issued. |
| 72 | `copilot set-coding-agent-permissions-for-organization` | certified | Read-back matched the requested `selected` policy; direct provider restoration to `all` was independently read back. |
| 73 | `copilot set-coding-agent-selected-repositories-for-organization` | certified | Collection read-back matched the real fixture repository; direct provider clearing returned the collection to zero, followed by policy restoration. |
| 74 | `copilot set-content-exclusion-for-organization` | entitlement | The full-admin classic credential received `404` from the organization content-exclusion surface; the alternate measured credential matrix route also declares no access. |
| 75 | `copilot set-enterprise-coding-agent-policy` | escape_needs_captain | The mutation targets an enterprise policy outside `Polymetrics-Cert`; no provider request was issued. |
| 76 | `agents add-selected-repo-to-org-secret` | product_defect | `class=integer_id_scientific_notation`; PM emitted the real repository ID in scientific notation. Raw PUT with the exact integer returned `204`, read-back matched the repository, and direct secret DELETE plus `404` proved cleanup. |
| 77 | `agents add-selected-repo-to-org-variable` | product_defect | `class=integer_id_scientific_notation`; PM emitted the real repository ID in scientific notation. Raw PUT with the exact integer returned `204`, read-back matched the repository, and direct variable DELETE plus `404` proved cleanup. |
| 78 | `agents create-or-update-org-secret` | certified | Independent GET matched the unique secret name and `private` visibility; direct DELETE returned `204` and final GET returned `404`. |
| 79 | `agents create-org-variable` | certified | Independent GET matched the unique name, value, and `private` visibility; direct DELETE returned `204` and final GET returned `404`. |
| 80 | `agents delete-org-secret` | certified | The raw-created secret returned `404` after PM's approved deletion; an idempotent direct DELETE and final `404` read-back proved absence. |
| 81 | `agents delete-org-variable` | certified | The raw-created variable returned `404` after PM's approved deletion; an idempotent direct DELETE and final `404` read-back proved absence. |
| 82 | `agents remove-selected-repo-from-org-secret` | product_defect | `class=integer_id_scientific_notation`; PM reported success while the selected-repository collection still contained the large repository ID. Raw DELETE with the exact integer returned `204`, collection read-back proved zero, then direct secret DELETE and `404` proved cleanup. |
| 83 | `agents remove-selected-repo-from-org-variable` | product_defect | `class=integer_id_scientific_notation`; PM reported success while the selected-repository collection still contained the large repository ID. Raw DELETE with the exact integer returned `204`, collection read-back proved zero, then direct variable DELETE and `404` proved cleanup. |
| 84 | `agents set-selected-repos-for-org-secret` | product_defect | PM requested the nonexistent `/orgs/Polymetrics-Cert/agents/secrets/.../repositories` path and received `404`. Raw PUT to the provider's `/actions/secrets/.../repositories` endpoint returned `204` and read-back matched the exact repository ID; direct secret DELETE and `404` proved cleanup. |
| 85 | `agents set-selected-repos-for-org-variable` | product_defect | PM requested the nonexistent `/orgs/Polymetrics-Cert/agents/variables/.../repositories` path and received `404`. Raw PUT to the provider's `/actions/variables/.../repositories` endpoint returned `204` and read-back matched the exact repository ID; direct variable DELETE and `404` proved cleanup. |
| 86 | `agents update-org-variable` | product_defect | PM requested the nonexistent `/orgs/Polymetrics-Cert/agents/variables/...` path and received `404`. Raw PATCH to the provider's `/actions/variables/...` endpoint returned `204`, and read-back matched the requested value and `private` visibility before direct DELETE and `404`. |

Batch 3 totals through command 86: `certified=6`, `no_object=0`, `wrong_credential=0`, `entitlement=1`, `not_implemented=0`, `product_defect=9`, `escape_needs_captain=6`; sum `22`.

Cumulative commands 1–86: `certified=39`, `no_object=11`, `wrong_credential=0`, `entitlement=9`, `not_implemented=0`, `product_defect=20`, `escape_needs_captain=7`; sum `86`.

## Live batch 4 — commands 87–100

| # | Command | Outcome | Independent observation |
| ---: | --- | --- | --- |
| 87 | `apps add-repo-to-installation-for-authenticated-user` | product_defect | `class=integer_id_scientific_notation`; PM emitted both the real installation and repository IDs in scientific notation. The organization installation collection independently showed the contained certification App installation and its `all` repository selection. |
| 88 | `apps create-from-manifest` | no_object | A manifest conversion code can only be minted by GitHub's interactive manifest flow; GitHub exposes no REST fixture-creation API, and no independently deletable one-time code was available. The invalid fixture code reached GitHub and returned `404`. |
| 89 | `apps delete-installation` | no_object | The parent collection contains only the shared `polymetrics-cert-app` installation. GitHub exposes no API to create a disposable installation, so deleting the shared credential container could not supply an independently replaceable fixture. |
| 90 | `apps redeliver-webhook-delivery` | no_object | Retried with the correct App JWT: `/app` returned `200`, while the parent `/app/hook/deliveries` collection returned `404 Not Found`. This certification App deliberately has webhooks inactive, and GitHub exposes no REST operation to create a delivery without a configured webhook, so no real delivery ID can be created inside the fixture boundary. |
| 91 | `apps remove-repo-from-installation-for-authenticated-user` | product_defect | `class=integer_id_scientific_notation`; PM used scientific-notation installation and repository path IDs and reported deletion success. Organization installation read-back remained `repository_selection=all`, so no repository selection was removed. |
| 92 | `apps revoke-installation-access-token` | certified | Retried with a freshly minted token from App installation `152693166`; PM plan, preview, and destructive run completed one record. An independent `/installation/repositories` read-back with that same token returned `401 Bad credentials`, rejecting the plausible wrong result that the token remained valid. Revocation is the provider-native terminal disposal. |
| 93 | `apps scope-token` | no_object | The operation requires an OAuth/GitHub App client ID and compatible user token. GitHub exposes no API to create that disposable OAuth client/token fixture; an inert `pm-cert-` placeholder reached the provider and returned `404`. |
| 94 | `apps suspend-installation` | product_defect | `class=integer_id_scientific_notation`; the real large installation ID is serialized into the path through the known fleet defect. The only available classic credential is also rejected for App-JWT-only installation administration. |
| 95 | `apps unsuspend-installation` | product_defect | `class=integer_id_scientific_notation`; the real large installation ID is serialized into the path through the known fleet defect. Credential health was independently repaired before the attempt; the App-JWT-only operation still rejected the classic credential. |
| 96 | `apps update-webhook-config-for-app` | no_object | Retried through PM with the correct App JWT: plan and preview succeeded, but PATCH returned `404`. The exact raw PATCH control and independent parent GET also returned `404 Not Found`; `/app` remained `200` and reported no configured events. The deliberately inactive App webhook has no configuration object, and GitHub exposes update/read but no REST create operation for it. |
| 97 | `projects create-draft-item-for-authenticated-user` | product_defect | PM requested the nonexistent singular `/user/{login}/projectsV2/.../drafts` path and returned `404`. Raw GraphQL `addProjectV2DraftIssue` succeeded, independent node read-back matched the title, and `deleteProjectV2` plus a null project read-back proved cleanup. |
| 98 | `projects create-draft-item-for-org` | certified | REST collection and GraphQL read-back matched the draft title and body in the fresh organization project; provider-native `deleteProjectV2` followed by a null project read-back proved cleanup. |
| 99 | `projects create-view-for-org` | certified | GraphQL view read-back matched the unique name, `BOARD_LAYOUT`, and filter; provider-native `deleteProjectV2` followed by a null project read-back proved cleanup. |
| 100 | `projects create-view-for-user` | certified | GraphQL view read-back matched the unique name, `ROADMAP_LAYOUT`, and filter; provider-native `deleteProjectV2` followed by a null project read-back proved cleanup. |

Batch 4 totals: `certified=4`, `no_object=5`, `wrong_credential=0`, `entitlement=0`, `not_implemented=0`, `product_defect=5`, `escape_needs_captain=0`; sum `14`.

Cumulative commands 1–100: `certified=43`, `no_object=16`, `wrong_credential=0`, `entitlement=9`, `not_implemented=0`, `product_defect=25`, `escape_needs_captain=7`; sum `100`.

## Live batch 5 — commands 101–146

| # | Command | Outcome | Independent observation |
| ---: | --- | --- | --- |
| 101 | `projects delete-item-for-org` | product_defect | `class=integer_id_scientific_notation`; PM reported success while the real item remained `200`. Raw DELETE with exact ID returned `204`, item read-back returned `404`, and `deleteProjectV2` removed the project. |
| 102 | `projects delete-item-for-user` | product_defect | `class=integer_id_scientific_notation`; PM reported success while the item remained. Exact raw DELETE returned `204`, item read-back returned `404`, and the user project was deleted. |
| 103 | `projects update-item-for-org` | product_defect | `class=integer_id_scientific_notation`; PM used scientific notation for the item path ID. Exact raw PATCH returned `200`, read-back matched the new title, then direct item DELETE and project deletion proved cleanup. |
| 104 | `projects update-item-for-user` | product_defect | `class=integer_id_scientific_notation`; PM used scientific notation for the item path ID. Exact raw PATCH returned `200`, read-back matched the new title, then direct item DELETE and project deletion proved cleanup. |
| 105 | `activity delete-thread-subscription` | product_defect | `class=integer_id_scientific_notation`; PM reported success but the real subscription remained. Exact raw DELETE returned `204` and subscription read-back returned `404`. |
| 106 | `activity mark-notifications-as-read` | certified | A real in-boundary thread changed from unread to read; direct provider terminal disposal marked it done, and the unread collection independently excluded it. |
| 107 | `activity mark-thread-as-done` | product_defect | `class=integer_id_scientific_notation`; the large thread ID is malformed in PM's path. Exact raw DELETE returned `204` and unread-collection read-back proved terminal disposal. |
| 108 | `activity mark-thread-as-read` | product_defect | `class=integer_id_scientific_notation`; PM received `404` on the scientific-notation path. Exact raw PATCH returned `205`, read-back showed `unread=false`, and direct terminal DELETE left no unread thread. |
| 109 | `activity set-thread-subscription` | product_defect | `class=integer_id_scientific_notation`; PM received `404` on the scientific-notation path. Exact raw PUT returned `200`, followed by direct subscription DELETE and `404`. |
| 110 | `activity star-repo-for-authenticated-user` | certified | Independent star read-back returned `204`; direct DELETE returned `204` and final read-back returned `404`. |
| 111 | `activity unstar-repo-for-authenticated-user` | certified | A raw-created star returned `404` after PM unstarred it; idempotent direct DELETE returned `204` and final read-back stayed `404`. |
| 112 | `environments deployment-branch-policies create` | product_defect | PM exposes no required policy name/type fields and GitHub returned `422`. Raw POST with those fields succeeded, collection read-back matched it, then direct policy and environment DELETEs produced zero matches and `404`. |
| 113 | `environments deployment-branch-policies delete` | product_defect | `class=integer_id_scientific_notation`; exact raw DELETE returned `204`, collection read-back proved zero, and direct environment DELETE returned `404` on read-back. |
| 114 | `environments deployment-branch-policies set` | product_defect | `class=integer_id_scientific_notation`; PM used scientific notation and received `404`. The exact-ID provider control reached the resource; the command also exposes no update body fields. Direct policy/environment cleanup returned `404`. |
| 115 | `environments deployment_protection_rules create` | no_object | Both the rule collection and eligible-App collection were empty. GitHub exposes no API to create an eligible deployment-protection App fixture, so no rule object could be created; the disposable environment was deleted and read back `404`. |
| 116 | `environments deployment_protection_rules delete` | no_object | The parent collection was empty and no eligible protection-rule App exists or can be created by API, so a deletable rule fixture cannot be created. |
| 117 | `oidc create-oidc-custom-property-inclusion-for-enterprise` | escape_needs_captain | Enterprise-scoped state leaves the `Polymetrics-Cert` boundary; no provider request was issued. |
| 118 | `oidc create-oidc-custom-property-inclusion-for-org` | certified | A fresh custom-property definition backed the inclusion; collection read-back matched it, direct inclusion DELETE removed it, and direct definition DELETE returned `404` on read-back. |
| 119 | `oidc delete-oidc-custom-property-inclusion-for-enterprise` | escape_needs_captain | Enterprise-scoped state leaves the authorized boundary; no provider request was issued. |
| 120 | `oidc delete-oidc-custom-property-inclusion-for-org` | certified | A raw-created inclusion disappeared after PM deletion; idempotent direct DELETE returned `404`, collection read-back excluded it, and the property definition returned `404` after cleanup. |
| 121 | `oidc update-oidc-custom-sub-template-for-org` | certified | After an explicit opposite-state fixture, read-back matched `repo`, `context`, and immutable-subject false; raw PUT restored immutable-subject true and read-back proved restoration. |
| 122 | `enterprise-team-memberships add` | escape_needs_captain | Enterprise membership affects an enterprise outside the authorization boundary; no provider request was issued. |
| 123 | `enterprise-team-memberships bulk-add` | escape_needs_captain | Enterprise membership affects external scope and people; no provider request was issued. |
| 124 | `enterprise-team-memberships bulk-remove` | escape_needs_captain | Enterprise membership affects external scope and people; no provider request was issued. |
| 125 | `enterprise-team-memberships remove` | escape_needs_captain | Enterprise membership affects an enterprise outside the boundary; no provider request was issued. |
| 126 | `release create` | certified | The draft release existed in the provider collection; object read-back matched tag/name/body/draft, and direct DELETE plus object `404` and zero collection matches proved cleanup. |
| 127 | `release delete` | product_defect | `class=integer_id_scientific_notation`; PM used scientific notation for the real release ID. Exact raw DELETE returned `204` and read-back returned `404`. |
| 128 | `release delete-asset` | product_defect | `class=integer_id_scientific_notation`; PM reported success while the real asset remained `200`. Exact raw DELETE returned `204`, asset read-back returned `404`, then release DELETE/read-back cleaned the container. |
| 129 | `release edit` | product_defect | `class=integer_id_scientific_notation`; PM received `404` on the scientific-notation release ID. Exact raw PATCH returned `200`, read-back matched name/body, and direct DELETE plus `404` cleaned the release. |
| 130 | `billing create-organization-budget` | escape_needs_captain | Organization budget controls affect real-money and metered usage; no provider request was issued. |
| 131 | `billing delete-budget-org` | escape_needs_captain | Organization budget controls affect real-money and metered usage; no provider request was issued. |
| 132 | `billing update-budget-org` | escape_needs_captain | Organization budget controls affect real-money and metered usage; no provider request was issued. |
| 133 | `hosted-compute create-network-configuration-for-org` | escape_needs_captain | Hosted-compute networking requires metered compute and third-party network-setting resources; no provider request was issued. |
| 134 | `hosted-compute delete-network-configuration-from-org` | escape_needs_captain | Hosted-compute networking crosses the metered/third-party boundary; no provider request was issued. |
| 135 | `hosted-compute update-network-configuration-for-org` | escape_needs_captain | Hosted-compute networking crosses the metered/third-party boundary; no provider request was issued. |
| 136 | `run cancel` | escape_needs_captain | The run collection is empty; creating a cancellable run would start metered GitHub Actions compute. No mutation request was issued. |
| 137 | `run delete` | escape_needs_captain | The run collection is empty; creating a deletable run would start metered GitHub Actions compute. No mutation request was issued. |
| 138 | `run rerun` | escape_needs_captain | The run collection is empty; creating or rerunning one starts metered GitHub Actions compute. No mutation request was issued. |
| 139 | `immutable-releases delete` | certified | A raw-enabled repository setting read `enabled=true`; PM disabled it, read-back showed false, and direct idempotent DELETE plus final read-back preserved false. |
| 140 | `immutable-releases set` | certified | Read-back changed from `enabled=false` to true; direct DELETE returned `204` and final read-back restored false. |
| 141 | `vulnerability-alerts delete` | certified | A raw-enabled setting returned `204`; PM deletion changed read-back to `404`, and direct idempotent DELETE plus final `404` proved cleanup. |
| 142 | `vulnerability-alerts set` | certified | Read-back changed from `404` to `204`; direct DELETE returned `204` and final read-back returned `404`. |
| 143 | `code-quality setup update` | escape_needs_captain | PM's empty-body attempt returned `422` without effect. A valid setup request would enable metered Code Quality analysis, so no valid provider mutation/control was issued. |
| 144 | `discussion create` | certified | After enabling discussions inside the private fixture, GraphQL read-back matched title/body/category; `deleteDiscussion` made node read-back null, and repository discussions were restored disabled. |
| 145 | `project create` | certified | Organization project collection read-back matched the unique title; provider-native `deleteProjectV2` followed by null project read-back proved cleanup. |
| 146 | `workflow run` | escape_needs_captain | The workflow collection is empty and creating/running a fixture workflow would start metered GitHub Actions compute; no mutation request was issued. |

Batch 5 totals: `certified=13`, `no_object=2`, `wrong_credential=0`, `entitlement=0`, `not_implemented=0`, `product_defect=14`, `escape_needs_captain=17`; sum `46`.

Final slice totals: `certified=56`, `no_object=18`, `wrong_credential=0`, `entitlement=9`, `not_implemented=0`, `product_defect=39`, `escape_needs_captain=24`; sum `146`.

## Final verification

- `git diff --check` — passed.
- `go run ./cmd/connectorgen certification-matrix --check` — passed: certification shards current.
- `go test -timeout 20m ./cmd/connectorgen -run '^TestCertification' -count=1` — passed in 75.219s.
- `go test -timeout 20m ./internal/agentcontract -run 'Certification|WorkflowEvidence' -count=1` — passed in 4.855s.
- Evidence records added over `origin/integration/4015-mvp-flat-r1`: `56`, matching the certified bucket.
- Verification rows: `146`, sequential and unique from 1 through 146.
- The full 550+ connector suite and monolithic `make verify` were not run locally because `AGENTS.md` explicitly directs per-command-timeout agents to scope local runs and let CI carry the full suite; this evidence-only change ran the relevant certification and workflow-evidence gates instead.
