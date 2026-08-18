# Mutation certification batch 3

Commands `109`–`141` were attempted serially through the full pm lifecycle. Command `142` was reached but not executed because its hosted agent/model effect is a real-money escape.

## Finalized observations

- `109`, `110`, `113`–`115`, `128`–`130`, `135`, `137`–`141`: `class=integer_id_scientific_notation`; real webhook IDs were used for `128`–`130`, and every webhook fixture was directly deleted with an independent 404.
- `111`: `product_defect`; pm returned 422 for a private disposable user repository, raw GitHub `POST /user/migrations` returned 201 with the same values, the repository was directly deleted/read back 404, and the exported archive was directly deleted/read back 404.
- `112`: `product_defect`; pm returned 422 for a private disposable org repository, raw GitHub `POST /orgs/Polymetrics-Cert/migrations` returned 201 with the same values, and the repository/archive cleanup was independently proven.
- `124`–`127`: `entitlement`; pm and the raw provider PUT/GET control returned 405 for the repository interaction-limit surface.
- `131`: `product_defect`; the pm surface requires `insecure_ssl` to be structured JSON although the provider requires a scalar. Raw PATCH with the scalar succeeded and matched read-back; the hook was directly deleted/read back 404.
- `132`–`134`: `entitlement`; the enterprise-team endpoints returned 404/403 for the bounded nonexistent enterprise scope, and no object was created.

## Not banked pending fixture or captain decision

- `116`–`119`: the repository import surface returned 404; a contained import fixture/control is still required before classification.
- `120`–`123`: a valid OAuth-application fixture would create a publicly visible application under the organization and therefore needs captain authority before the invalid-ID observations can be replaced.
- `136`: the stack create request returned 422 with a fabricated PR ID; a real PR/stack fixture or raw control is required.
- `142`: `escape_needs_captain` pending decision because executing a hosted agent task can consume metered model service.

Commands `143`–`146` remain unattempted behind the command-142 escape gate.
