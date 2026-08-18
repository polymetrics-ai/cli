# TDD Ledger — GitHub mutation certification slice 5 writes-e

## Manual GSD fallback

The repository-local Pi adapter prompts were generated and reviewed. This isolated terminal runtime cannot run Pi subagents, so the lifecycle is completed inline.

## Cycle 1 — certification evidence representability

**Red:** Validate the integration base before writing. The designated base adds schema-v2 `observed_operations` records with salted fingerprints and command-specific `function_kind` values; the pre-rebase checkout did not contain those records.

**Green criterion:** Persist a record only when the command has a real provider proof and validator acceptance. Non-passes are recorded as classified attempt outcomes, not fabricated `status: passed` evidence.

## Cycle 2 — live mutation containment

**Red criterion:** A successful CLI exit alone is insufficient because issue #4221 demonstrates delete success can leave the object present.

**Green criterion:** A passed command must include an independent provider read-back of its state change and a direct-provider deletion followed by a 404 or empty collection.

## Attempt ledger (redacted)

| Ordinal | Command | Result | Evidence |
| ---: | --- | --- | --- |
| 1 | `codespaces add-repository-for-secret-for-authenticated-user` | `product_defect` | `class=integer_id_scientific_notation`; the repository ID is rendered in exponential notation. Fleet raw control already proved the exact-integer request succeeds and cleans up. |
| 2 | `codespaces add-selected-repo-to-org-secret` | `product_defect` | `class=integer_id_scientific_notation`; PM requested `/repositories/1.330114309e%2B09`. |
| 3 | `codespaces create` | `entitlement` | Plan and preview completed. Run reached GitHub and returned HTTP 400. A raw GitHub provider control using the smallest tier (`basicLinux32gb`) returned the same 400 and created no Codespace; no cleanup was required. |
| 4 | `codespaces create-from-pull-request` | `escape_needs_captain` | Excluded by the resolved metered-compute ruling; no request was sent. |
| 5 | `codespaces create-from-repository` | `escape_needs_captain` | Excluded by the resolved metered-compute ruling; no request was sent. |
| 6 | `codespaces create-or-update-org-secret` | `wrong_credential` | Classic PAT reached GitHub, but the organisation Codespaces-secret endpoint returned 404; the raw collection/key reads returned the same access result. |
| 7 | `codespaces create-or-update-secret-for-authenticated-user` | `no_object` | GitHub rejected the deliberately non-secret invalid cipher fixture with 422 and independent secret listing remained empty. |
| 8 | `codespaces delete-codespaces-access-users` | `product_defect` | `class=delete_false_success`; PM reported one success while the raw DELETE control returned 404. |
| 9 | `codespaces delete-for-authenticated-user` | `product_defect` | `class=delete_false_success`; PM reported one success while the raw DELETE control returned 404. |
| 10 | `codespaces delete-from-organization` | `product_defect` | `class=delete_false_success`; PM reported one success while the raw DELETE control returned 404. |
| 11 | `codespaces delete-org-secret` | `product_defect` | `class=delete_false_success`; PM reported one success while the raw DELETE control returned 404. |
| 12 | `codespaces delete-secret-for-authenticated-user` | `product_defect` | `class=delete_false_success`; PM reported one success while the raw DELETE control returned 404. |
| 13 | `codespaces export-for-authenticated-user` | `no_object` | The disposable identity has no Codespace; GitHub returned 404 and no export was created. |
| 14 | `codespaces publish-for-authenticated-user` | `no_object` | A private publish was attempted against the absent tagged Codespace; GitHub returned 404 and no repository was created. |
| 15 | `codespaces remove-repository-for-secret-for-authenticated-user` | `product_defect` | `class=integer_id_scientific_notation,delete_false_success`; PM reported success with a large repository ID while the exact-ID raw DELETE returned 404. |
| 16 | `codespaces remove-selected-repo-from-org-secret` | `product_defect` | `class=integer_id_scientific_notation,delete_false_success`; PM reported success with a large repository ID while the exact-ID raw DELETE returned 404. |
| 17 | `codespaces set-codespaces-access` | `wrong_credential` | GitHub returned 404; independent raw read of the org access endpoint returned the same result. |
| 18 | `codespaces set-codespaces-access-users` | `wrong_credential` | GitHub returned 404; independent raw read of selected access users returned the same result. |
| 19 | `codespaces set-repositories-for-secret-for-authenticated-user` | `no_object` | The tagged user secret did not exist; GitHub returned 404 and independent secret listing was empty. |
| 20 | `codespaces set-selected-repos-for-org-secret` | `wrong_credential` | GitHub returned 404 and the org secret collection is inaccessible to the measured credential. |
| 21 | `codespaces start-for-authenticated-user` | `escape_needs_captain` | Excluded by the resolved metered-compute ruling; no request was sent. |
| 22 | `codespaces stop-for-authenticated-user` | `no_object` | The tagged Codespace did not exist; GitHub returned 404. |
| 23 | `codespaces stop-in-organization` | `no_object` | The tagged member Codespace did not exist; GitHub returned 404. |
| 24 | `codespaces update-for-authenticated-user` | `no_object` | The tagged Codespace did not exist; GitHub returned 404. |
| 25 | `copilot-spaces add-collaborator-for-org` | `no_object` | The tagged space number did not exist; GitHub returned 404. |
| 26 | `copilot-spaces add-collaborator-for-user` | `no_object` | The tagged space number did not exist; GitHub returned 404. |
| 27 | `copilot-spaces create-for-org` | `certified` | Independent collection read-back observed space 2 with the exact tagged name, description, and base role; a plausible wrong name failed. Direct DELETE returned 204 and GET returned 404. Published record: `github-manual-slice5-copilot-spaces-create-for-org-rrun-3bb41358830d7492.json`. |
| 28 | `copilot-spaces create-for-user` | `certified` | Independent collection read-back observed space 8 with the exact tagged name, description, and base role; a plausible wrong name failed. Direct DELETE returned 204 and GET returned 404. Published record: `github-manual-slice5-copilot-spaces-create-for-user-rrun-0b6cfe3b7e44c04c.json`. |
| 29 | `copilot-spaces create-resource-for-org` | `no_object` | The tagged space number did not exist; GitHub returned 404 and no resource was created. |
| 30 | `copilot-spaces create-resource-for-user` | `no_object` | The tagged space number did not exist; GitHub returned 404 and no resource was created. |
| 31 | `copilot-spaces delete-for-org` | `product_defect` | `class=delete_false_success`; PM reported one success while the raw DELETE control returned 404. |
| 32 | `copilot-spaces delete-for-user` | `product_defect` | `class=delete_false_success`; PM reported one success while the raw DELETE control returned 404. |
| 33 | `copilot-spaces delete-resource-for-org` | `product_defect` | `class=delete_false_success`; PM reported one success while the raw DELETE control returned 404. |
| 34 | `copilot-spaces delete-resource-for-user` | `product_defect` | `class=delete_false_success`; PM reported one success while the raw DELETE control returned 404. |
| 35 | `copilot-spaces remove-collaborator-for-org` | `product_defect` | `class=delete_false_success`; PM reported one success while the raw DELETE control returned 404. |
| 36 | `copilot-spaces remove-collaborator-for-user` | `product_defect` | `class=delete_false_success`; PM reported one success while the raw DELETE control returned 404. |
| 37 | `copilot-spaces update-collaborator-for-org` | `no_object` | The tagged space/collaborator did not exist; GitHub returned 404. |
| 38 | `copilot-spaces update-collaborator-for-user` | `no_object` | The tagged space/collaborator did not exist; GitHub returned 404. |
| 39 | `copilot-spaces update-for-org` | `no_object` | The tagged space did not exist; GitHub returned 404. |
| 40 | `copilot-spaces update-for-user` | `no_object` | The tagged space did not exist; GitHub returned 404. |
| 41 | `copilot-spaces update-resource-for-org` | `no_object` | The tagged space/resource did not exist; GitHub returned 404. |
| 42 | `copilot-spaces update-resource-for-user` | `no_object` | The tagged space/resource did not exist; GitHub returned 404. |
| 43 | `issues approve-suggestion` | `no_object` | Tagged issue 32 existed, but suggestion 999 did not; GitHub returned 404. |
| 44 | `issues blocked_by create` | `product_defect` | PM POSTed no body and GitHub returned 422. Raw POST with issue ID 5177928589 returned 201, read-back contained issue 33, direct DELETE returned 200, and read-back was empty. |
| 45 | `issues blocked_by delete` | `product_defect` | `class=integer_id_scientific_notation`; PM requested the issue ID as `5.177928589e%2B09`. The exact-ID raw DELETE control for ordinal 44 succeeded and read-back was empty. |
| 46 | `issues dismiss-suggestion` | `no_object` | Tagged issue 32 existed, but suggestion 999 did not; GitHub returned 404. |
| 47 | `issues issue-field-values create` | `no_object` | The fixture repository has no configured issue field/value object; GitHub returned 422 and no value was created. |
| 48 | `issues issue-field-values delete` | `no_object` | Tagged issue 32 existed, but issue field 999 did not; GitHub returned 404. |
| 49 | `issues issue-field-values set` | `no_object` | The fixture repository has no configured issue field/value object; GitHub returned 422 and no value was set. |
| 50 | `issues labels delete` | `certified` | A tagged label was assigned to issue 32. PM removed it; independent GET returned `[]` and the plausible label-present assertion failed. Direct label DELETE returned 204 and GET returned 404; issues 32 and 33 were closed and independently read back as `contained_closed`. Published record: `github-manual-slice5-issues-labels-delete-rrun-223f7f08c2000902.json`. |
| 55 | `issues reactions create-2` | `product_defect` | PM POSTed a bodyless reaction and GitHub returned 422. Raw GitHub POST with the required `content` succeeded (201), independently read back as present, then direct DELETE returned 204 and read back absent. |
| 57 | `issues reactions delete-2` | `product_defect` | `class=integer_id_scientific_notation`; PM serialized the reaction identifier in exponential notation and GitHub returned 404. Direct DELETE using the provider-issued identifier returned 204 and independent collection read-back proved absence. |
| 99 | `variable delete-2` | `certified` | PM DELETE returned a completed one-record run; independent raw GET returned 404. Published schema-v2 record: `github-manual-slice5-variable-delete-2-rrun-5282e7b8218c.json`. |
| 102 | `variable update` | `product_defect` | PM PATCH had no value payload and GitHub returned 422. Raw GitHub PATCH with the required value returned 204; direct DELETE and independent GET then proved the fixture absent. |

Batch 1 classification totals for ordinals 1-50: `certified=3`, `no_object=22`,
`wrong_credential=4`, `entitlement=1`, `not_implemented=0`, `product_defect=17`,
`escape_needs_captain=3`; sum `50`.

## Contained fixture cleanup

GitHub does not offer deletion for issue resources. The disposable issue used for
the reaction controls was directly PATCHed to `state=closed` after use and an
independent provider read-back confirmed `closed`; it is recorded as
`contained_closed`, never as `verified_absent`. The enclosing private fixture
repository is the disposable container.
