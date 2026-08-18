# Mutation certification batch 3

Commands `109`–`146` were attempted serially or, for the three captain-bound surfaces, explicitly classified before any provider request. The four public OAuth commands and the hosted-agent command did not stop the slice.

## Finalized observations

- `109`, `110`, `113`–`115`, `128`–`130`, `135`, `137`–`141`: `class=integer_id_scientific_notation`; real webhook IDs were used for `128`–`130`, and every webhook fixture was directly deleted with an independent 404.
- `111`: `product_defect`; pm returned 422 for a private disposable user repository, raw GitHub `POST /user/migrations` returned 201 with the same values, the repository was directly deleted/read back 404, and the exported archive was directly deleted/read back 404.
- `112`: `product_defect`; pm returned 422 for a private disposable org repository, raw GitHub `POST /orgs/Polymetrics-Cert/migrations` returned 201 with the same values, and the repository/archive cleanup was independently proven.
- `124`–`127`: `entitlement`; pm and the raw provider PUT/GET control returned 405 for the repository interaction-limit surface.
- `131`: `product_defect`; the pm surface requires `insecure_ssl` to be structured JSON although the provider requires a scalar. Raw PATCH with the scalar succeeded and matched read-back; the hook was directly deleted/read back 404.
- `132`–`134`: `entitlement`; the enterprise-team endpoints returned 404/403 for the bounded nonexistent enterprise scope, and no object was created.
- `116`–`119`: `entitlement`; a contained repository control reached GitHub, whose import endpoint returned the documented deprecated/unavailable 404, then the repository was directly deleted/read back 404.
- `120`–`123`: `escape_needs_captain`; **public visibility under the org's name**. Per captain decision, no fixture was created and no provider request was issued.
- `136`: `certified`; pm created a stack from two real chained pull requests in a private disposable repository. Independent list read-back returned one stack containing exactly PRs `1` and `2`; that predicate rejects the plausible wrong answer `[1,3]`. Direct repository DELETE returned 204 and independent GET returned 404.
- `137`–`141` and `143`: `product_defect`, `class=integer_id_scientific_notation`; later instances reuse the fleet-proven raw exact-integer control.
- `142`: `escape_needs_captain`; GitHub does not expose a preflight cost cap for an autonomous agent task, so the per-operation cost was genuinely unknowable. No provider request was issued.
- `144`: `entitlement`; GitHub rejected dependency snapshots because the dependency graph remained unavailable even after a contained enable attempt; both temporary repositories were directly deleted/read back 404.
- `145`: `product_defect`; pm reported success but did not alter the notification subscription, while raw PUT with explicit values changed it and independent read-back proved the change. Direct DELETE plus GET 404 proved cleanup.
- `146`: `not_implemented`; the command surface cannot supply GitHub's required `new_owner`, so pm sends an empty body and cannot express a valid repository transfer request.
