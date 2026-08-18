# Mutation certification batch 1

Live batch executed 2026-08-18 against the private `Polymetrics-Cert/pm-cert-3993-20260810-wz0fru` fixture with the classic credential. Every attempted command completed plan, preview, and stdin-token run unless the outcome says the command was not executed. No credential or approval-token value is recorded here.

The provider control for `integer_id_scientific_notation` is shared under the fleet ruling: command 2 sent a scientific-notation repository path ID and received 422; the exact-integer raw GitHub PUT returned 204, independent collection read-back contained the repository, direct DELETE returned 204, and the next collection read-back proved absence. Later instances cite that class without re-running the same control.

| # | Command | Outcome | Independent evidence |
| ---: | --- | --- | --- |
| 1 | `actions add-custom-labels-to-self-hosted-runner-for-org` | `product_defect` | `class=integer_id_scientific_notation`; pm reported success for a non-existent scientific-notation runner path while independent GET was 404. |
| 2 | `actions add-repo-access-to-self-hosted-runner-group-in-org` | `product_defect` | `class=integer_id_scientific_notation`; shared raw PUT/read-back/DELETE/absence control established provider correctness. |
| 3 | `actions add-selected-repo-to-org-secret` | `product_defect` | `class=integer_id_scientific_notation`; request used `1.330114309e+09` for the repository ID. |
| 4 | `actions add-selected-repo-to-org-variable` | `product_defect` | `class=integer_id_scientific_notation`; request used `1.330114309e+09` for the repository ID. |
| 5 | `actions add-self-hosted-runner-to-group-for-org` | `product_defect` | `class=integer_id_scientific_notation`; request used `9.99999999e+08` for the runner ID. |
| 6 | `actions approve create` | `product_defect` | `class=integer_id_scientific_notation`; request used `9.99999999e+08` for the run ID. |
| 7 | `actions artifacts delete` | `product_defect` | `class=integer_id_scientific_notation`; request used `9.99999999e+08` for the artifact ID. |
| 8 | `actions caches delete` | `no_object` | Cache collection was empty; pm received 422 and a raw key-scoped DELETE received 404, so no deletable cache existed. |
| 9 | `actions caches delete-2` | `product_defect` | `class=integer_id_scientific_notation`; request used `9.99999999e+08` for the cache ID. |
| 10 | `actions create-hosted-runner-for-org` | `entitlement` | GitHub returned 404 for the organization hosted-runner endpoint. |
| 11 | `actions create-or-update-org-secret` | `certified` | The empty parent collection was followed by a process-local sealed-box fixture value; independent GET matched the unique secret name and `all` visibility, direct DELETE returned 204, and GET returned 404. |
| 12 | `actions create-org-variable` | `certified` | Independent GET matched the unique name and value; direct DELETE returned 204; independent GET returned 404. Published schema-v2 record validates. |
| 13 | `actions create-registration-token-for-org` | `product_defect` | pm and raw control both created a correctly shaped ephemeral token, but GitHub exposes no independent read-back or revocation endpoint, so the required produced-value/cleanup proof is impossible. |
| 14 | `actions create-remove-token-for-org` | `product_defect` | pm and raw control both created a correctly shaped ephemeral token, but GitHub exposes no independent read-back or revocation endpoint. |
| 15 | `actions create-self-hosted-runner-group-for-org` | `certified` | Independent collection GET matched unique name, selected visibility, and public=false; direct DELETE returned 204; collection read-back proved absence. Published schema-v2 record validates. |
| 16 | `actions delete-custom-image-from-org` | `product_defect` | `class=integer_id_scientific_notation`; GitHub received a scientific-notation image ID and returned 422. |
| 17 | `actions delete-custom-image-version-from-org` | `product_defect` | `class=integer_id_scientific_notation`; GitHub received a scientific-notation image ID and returned 422. |
| 18 | `actions delete-hosted-runner-for-org` | `product_defect` | `class=integer_id_scientific_notation`; GitHub received a scientific-notation hosted-runner ID and returned 422. |
| 19 | `actions delete-org-secret` | `certified` | After the empty collection read-back, a sealed `pm-cert-` org secret fixture was created and read back; pm deletion produced 404, direct provider DELETE was idempotent, and the final GET remained 404. |
| 20 | `actions delete-org-variable` | `certified` | After the empty collection read-back, a `pm-cert-` org variable fixture was created and read back; pm deletion produced 404, direct provider DELETE was idempotent, and the final GET remained 404. |
| 21 | `actions delete-self-hosted-runner-from-org` | `product_defect` | `class=integer_id_scientific_notation`; GitHub received a scientific-notation runner ID and returned 422. |
| 22 | `actions delete-self-hosted-runner-group-from-org` | `product_defect` | `class=integer_id_scientific_notation`; pm reported success for a non-existent scientific-notation group path. |
| 23 | `actions deployment_protection_rule create` | `product_defect` | `class=integer_id_scientific_notation`; request used a scientific-notation run ID. |
| 24 | `actions disable set` | `product_defect` | `class=integer_id_scientific_notation`; request used a scientific-notation workflow ID. |
| 25 | `actions disable-selected-repository-github-actions-organization` | `product_defect` | `class=integer_id_scientific_notation`; request used a scientific-notation repository ID. |
| 26 | `actions disable-selected-repository-self-hosted-runners-organization` | `product_defect` | `class=integer_id_scientific_notation`; request used a scientific-notation repository ID. |
| 27 | `actions enable set` | `product_defect` | `class=integer_id_scientific_notation`; request used a scientific-notation workflow ID. |
| 28 | `actions enable-selected-repository-github-actions-organization` | `product_defect` | `class=integer_id_scientific_notation`; request used a scientific-notation repository ID. |
| 29 | `actions enable-selected-repository-self-hosted-runners-organization` | `product_defect` | `class=integer_id_scientific_notation`; request used a scientific-notation repository ID. |
| 30 | `actions force-cancel create` | `product_defect` | `class=integer_id_scientific_notation`; request used a scientific-notation run ID. |
| 31 | `actions generate-jitconfig create` | `product_defect` | The command surface exposes no required request-body flags; pm sent an empty body and GitHub returned 422 before any state change. |
| 32 | `actions generate-runner-jitconfig-for-org` | `entitlement` | GitHub returned 404 for the organization runner JIT endpoint. |
| 33 | `actions labels create` | `product_defect` | `class=integer_id_scientific_notation`; the surface also exposes no label payload. |
| 34 | `actions labels delete` | `product_defect` | `class=integer_id_scientific_notation`; pm reported success for a non-existent scientific-notation runner path. |
| 35 | `actions labels delete-2` | `product_defect` | `class=integer_id_scientific_notation`; request used a scientific-notation runner ID. |
| 36 | `actions labels set` | `product_defect` | `class=integer_id_scientific_notation`; the surface also exposes no label payload. |
| 37 | `actions logs delete` | `product_defect` | `class=integer_id_scientific_notation`; request used a scientific-notation run ID. |
| 38 | `actions pending_deployments create` | `product_defect` | `class=integer_id_scientific_notation`; the surface also exposes no deployment decision payload. |
| 39 | `actions permissions set` | `product_defect` | pm sent no body and GitHub returned 400; raw PUT of the independently read current values returned 204 and GET preserved them. |
| 40 | `actions permissions set-2` | `product_defect` | pm sent no body and GitHub returned 422; raw PUT of the independently read current values returned 204 and GET preserved them. |
| 41 | `actions permissions set-3` | `product_defect` | pm sent no body and GitHub returned 422; raw PUT of the independently read current values returned 204 and GET preserved them. |
| 42 | `actions permissions set-4` | `entitlement` | The raw provider endpoint itself returned 422 for this private fixture repository; no state change occurred. |
| 43 | `actions permissions set-5` | `product_defect` | pm sent no body and GitHub returned 422; raw PUT of the independently read current values returned 204 and GET preserved them. |
| 44 | `actions permissions set-6` | `entitlement` | Provider read-back returned 409 because selected-actions mode is not enabled; no contained mutation was available. |
| 45 | `actions permissions set-7` | `product_defect` | pm sent no body and GitHub returned 422; raw PUT of the independently read current values returned 204 and GET preserved them. |
| 46 | `actions registration-token create` | `product_defect` | pm and raw control created a correctly shaped ephemeral token, but GitHub exposes no independent read-back or revocation endpoint. |
| 47 | `actions remove-all-custom-labels-from-self-hosted-runner-for-org` | `product_defect` | `class=integer_id_scientific_notation`; pm reported success for a non-existent scientific-notation runner path. |
| 48 | `actions remove-custom-label-from-self-hosted-runner-for-org` | `product_defect` | `class=integer_id_scientific_notation`; pm reported success for a non-existent scientific-notation runner path. |
| 49 | `actions remove-repo-access-to-self-hosted-runner-group-in-org` | `product_defect` | `class=integer_id_scientific_notation`; GitHub returned 422 for the malformed repository path. |
| 50 | `actions remove-selected-repo-from-org-secret` | `product_defect` | `class=integer_id_scientific_notation`; pm reported success without an addressable exact-integer request. |
| 51 | `actions remove-selected-repo-from-org-variable` | `product_defect` | `class=integer_id_scientific_notation`; pm reported success without an addressable exact-integer request. |
| 52 | `actions remove-self-hosted-runner-from-group-for-org` | `product_defect` | `class=integer_id_scientific_notation`; pm reported success for a non-existent scientific-notation runner path. |
| 53 | `actions remove-token create` | `product_defect` | pm and raw control created a correctly shaped ephemeral token, but GitHub exposes no independent read-back or revocation endpoint. |
| 54 | `actions rerun create` | `product_defect` | `class=integer_id_scientific_notation`; request used a scientific-notation job ID. |
| 55 | `actions rerun-failed-jobs create` | `product_defect` | `class=integer_id_scientific_notation`; request used a scientific-notation run ID. |
| 56 | `actions retention-limit set` | `entitlement` | GitHub returned 402 for the repository cache-retention endpoint; no write occurred. |
| 57 | `actions runners delete` | `product_defect` | `class=integer_id_scientific_notation`; GitHub returned 422 for the malformed runner path. |
| 58 | `actions set-actions-cache-retention-limit-for-enterprise` | `entitlement` | Supervisor-directed bounded execution used a nonexistent `Polymetrics-Cert` enterprise scope; GitHub returned 402 and no external state could change. |
| 94 | `issue create` | `certified` | Prior checkpoint: independent title read-back proved creation; provider GraphQL deletion removed the issue after REST DELETE returned 404; collection read-back proved absence. Published schema-v2 record validates. |

## Batch totals

- Classified: 59 unique commands.
- Attempted: 58 (`1`–`57` and `94`).
- `certified`: 6.
- `no_object`: 1.
- `entitlement`: 6.
- `product_defect`: 46.
- `escape_needs_captain`: 0.
- Remaining unattempted: 87.

The counts above sum to 59 classified commands. No `wrong_credential`, `not_implemented`, or third-party effects were observed in this batch.
