# TDD ledger — Issue #3981 managed-target ownership and provisioning

| ID | Guarantee | RED evidence | GREEN evidence |
| --- | --- | --- | --- |
| R1 | Source-derived identity | The absent contract cannot create a typed owner/ref based on the warehouse source triple. | Owner/ref projections use only `warehouse.ArtifactIdentity`; test inputs containing display/credential-looking data cannot influence names or records. |
| R2 | Typed mutation authority | A fake can accept an untyped create request or a mismatched owner. | The only mutation port accepts a validated `ProvisioningPlan` with its asserted owner; invalid plans never call the fake. |
| R3 | Closed first-create admission | An absent namespace/control pair has no create-or-assert state transition. | The truth table creates once, then re-observes and stores/admits the exact record. |
| R4 | Exact repeat admission | Correct owner/control/native/schema state lacks an idempotent assertion path. | A repeat is admitted without a second create. |
| R5 | Foreign, missing, unreadable, and collision refusal | Foreign/missing/unreadable/colliding observations can be treated as a creatable target. | Each state returns a typed refusal and leaves fake mutation count unchanged. |
| R6 | Replacement and drift refusal | A changed native identity or schema hash/version can be admitted or evolved. | Moved/replaced and schema-drift table rows return typed refusal without mutation. |
| R7 | Cancellation and races fail closed | **Red:** An invoked create can return after cancellation or an error without a final owner/native/schema assertion. | **Green:** Every invoked create is re-observed through a cancellation-independent context while the typed target lock remains held; exact state is admitted, unsafe state wins over cancellation/driver errors, and concurrent callers across two provisioners observe one legitimate create. |
| R8 | Driver-neutral boundary | The shared contract requires a PostgreSQL/SQL implementation. | A pure in-memory fake proves all transitions; no driver, SQL, capability, or CLI files change. |

## Command record

**RED:** `go test -timeout 20m -count=1 ./internal/connectors/database -run '^TestManagedTargetProvisioningTruthTable$'`
failed to build because the deliberately absent `TargetOwner`, managed-target
ref/control/schema, plan, observation, and provisioning API were referenced.
The exact output is retained at `traces/managed-target-provisioning-red.txt`.

**GREEN (focused):** both the focused test and its `-race` run pass; exact
commands/output are retained in `traces/managed-target-provisioning-green.txt`.
Broader package, affected, static/build, and individual repository gates remain
in the verification checklist.

**Correction 1/5:** #4038 records the cross-provisioner gap found during the
required review before its fix. `traces/cross-provisioner-lock-red.txt` and
`traces/cross-provisioner-lock-green.txt` preserve the RED/GREEN evidence.

**Correction 2/5 (#4044):**
**Red:** Review finding R1 traced cancellation and driver-error exits after an
invoked create that skipped the required final observation.
**Green:** `go test -race -timeout 20m -count=1 ./internal/connectors/database -run '^TestManagedTargetProvisioningTruthTable$'` passed, including committed-create cancellation/error transitions and foreign-state classification.
