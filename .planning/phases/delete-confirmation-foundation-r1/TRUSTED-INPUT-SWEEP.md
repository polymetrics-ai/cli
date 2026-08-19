# Trusted-input sweep — destructive write approval

The execution gate accepts only facts recovered through App-private grant verification and the
current engine-prepared write. State JSON carries the grant and lifecycle hints, but no mutable
state field can independently authorize dispatch.

| Fact | Read at authorization | Caller or agent writable? | Trusted source and tamper defense |
| --- | --- | --- | --- |
| Grant envelope | `projectWriteApprovalAuthority.VerifyWriteGrant` | The persisted JSON copy is writable | HMAC-SHA256 covers the complete grant; App-private verification rejects an altered envelope before consumption. |
| `Version` | grant verification | Yes, through state JSON | Covered by the grant MAC and checked against the closed runtime version. |
| `AuthorityID` | grant verification | Yes, through state JSON | Derived during trusted `App.Open` from the project vault key and covered by the grant MAC; caller-key authorities produce only untrusted evidence that the engine refuses. |
| `PlanID` | app consumption and grant verification | Yes, through state JSON or request | Compared with the executing plan and covered by the grant MAC. |
| `PlanHash` | app recomputation and grant verification | Yes, through state JSON | Recomputed from the current command/bulk payload, covered by the plan seal, and covered again by the grant MAC. |
| `Mode` | app dispatch and grant verification | Yes, through state JSON | Covered by the plan seal and grant MAC; command and bulk modes cannot be substituted. |
| `PlanSealMAC` | grant issuance and verification | Yes, through state JSON | A production grant is issued only from a current authenticated plan seal; the seal MAC is then covered by the grant MAC. |
| `PreviewDigest` | app re-preview, grant verification, engine gate | Yes, through state JSON | Recomputed from every canonical request and complete definition/hook identity, then covered by the grant MAC; native preview and execution receive the same hook set. |
| `ApprovalTokenHash` | grant verification | The state copy is writable; the raw token is caller input | The authoritative hash is inside the grant MAC and is compared in constant time with the presented token. The state copy can only cause an early rejection. |
| `Nonce` | grant verification and project consumption | Yes, through state JSON | Random at grant issuance, covered by the grant MAC, and persisted only as an App-private root HMAC inside the authenticated consumption record. |
| `Target.Connector` | engine preview and engine gate | Bundle metadata is repository-authored; state is writable | Re-derived by the current writer, covered by the plan seal and grant MAC, and compared with opaque evidence. |
| `Target.Operation` | engine preview and engine gate | Bundle metadata is repository-authored; state is writable | Re-derived by the current writer, covered by the plan seal and grant MAC, and compared with opaque evidence. |
| `Target.Method` | engine preview and engine gate | Bundle metadata is repository-authored | Canonical prepared requests must match the normalized method; the value is covered by the grant MAC. |
| `Target.MutationClass` | engine preview and engine gate | Bundle metadata is repository-authored | Re-derived from live metadata and covered by the grant MAC; DELETE and destructive inference still fail closed. |
| `Target.TargetDigest` | engine preview and engine gate | Records/configuration are caller controlled | Hashes every concrete canonical request target/body/query and is covered by the grant MAC. |
| `Target.CredentialRevision` | credential resolution and engine gate | Credential secrets are vault-managed, not state JSON | Vault-root HMAC over credential ID plus sorted secret material; secret values never enter state or evidence. |
| `Target.ConfigurationDigest` | credential resolution and engine gate | Configuration values can originate in state/CLI | Vault-root HMAC over credential ID plus effective sorted configuration; the plan seal and grant MAC bind it. |
| `Target.Batchable` | live manifest, plan seal, engine gate | State JSON is writable; bundle metadata is repository-authored | Re-derived from the live action, then covered by the plan seal and grant MAC; confirmation cannot change `batchable:false`. |
| `Target.Scope` | app runtime and engine preview | Direct engine callers can request fixture scope | Production App always supplies `project`; fixture scope is accepted only for loopback prepared requests and cannot authorize a project target. The scope is MAC-bound. |
| `Target.Confirmation` | live metadata and engine gate | State JSON is writable | Re-derived as the closed destructive type and covered by the plan seal and grant MAC. |
| `IssuedAt` | plan/grant verification and engine evidence | The persisted copy is writable | Generated from process time inside the authority, covered by the MAC, and checked for future activation. |
| `ExpiresAt` | plan/grant verification, previewability, and engine evidence | The persisted copy is writable | Destructive plan/grant deadlines are authority-generated and MAC-bound, and mutable destructive `ReversePlan.ExpiresAt` is non-authoritative. Unsigned plans fail closed against their recorded creation and expiry times before preview or execution. |
| Grant `Confirmation` | grant verification | The persisted copy is writable | Covered by the grant MAC and compared with the explicit typed `--confirm destructive` value. |
| Grant `MAC` | grant verification | The persisted copy is writable | Verified with the project-vault root before any fact is accepted. |
| Project consumption record | grant issuance and verification before state replacement | Not writable through state mutations | App-private code creates one authenticated opaque marker per authority/plan ID/hash/mode with `O_CREATE|O_EXCL` before dispatch. Grant issuance checks the same marker, so rollback cannot re-preview the consumed plan with a fresh nonce. |
| Fixture consumption record | fixture grant verification | Test callers can retain or copy the authority value | A shared concurrency-safe registry is pointer-owned by the random fixture authority, so authority copies still atomically consume the same grant MAC/nonce only once. A second authority cannot verify the grant because its key differs. |
| Evidence origin | engine authorization | Direct callers can implement the public method shape | The connector wrapper accepts project evidence only when its concrete type is the unexported App evidence type; caller implementations and caller-key evidence are rejected before dispatch. |
| Authority/key | trusted App initialization and authority authentication | No raw key or production authority is accepted from write callers or engine requests | Production construction is unexported in `internal/app`; `App.Open` derives the signing root from the opened project vault, and the previous public process-authority/root API is absent. The exported keyed authority is permanently untrusted by the engine. |
| Redirect destination | shared transport policy under the authorized execution context | Provider responses control `Location` | The shared gate marks every authorized destructive execution context; `transportpolicy.HTTPClient` clones the client and rejects every redirect for both `connsdk` and native Amazon SQS before a second outbound request. The redirected target requires a new plan, preview, and approval. |
| Hook identity | prepared-write digest generation | Native connector code selects its hook implementation | Preview and execution pass the same native hooks, and the concrete hook type plus connector identity remain inside the digest. A mismatch fails before dispatch. |

The redirect audit found no destructive provider fixture or connector test that requires redirects;
existing matches describe read resources or metadata named “redirect,” so declarative and native
writers need no provider exception.

## Whole-state write audit

| Write path | Previous risk | Current rule |
| --- | --- | --- |
| `App.save` used by credentials, connections, catalogs, ETL runs, plans, and secret-field metadata | Replaced the complete file from a stale in-memory snapshot | Compares the loaded `state.Revision` under the state lock and rejects a stale revision before replacement. |
| Approval preview, consumption, completion, and invalidation | Locked reload/update but did not advance a version understood by other writers | All use `App.updateState`, which reloads under lock and advances `state.Revision`. |
| Project initialization | Creates the state file only when absent | Initializes revision zero and never overwrites an existing state file. |
| Project approval consumption | Previously absent, so a valid older state snapshot restored a grant | App-private authority code creates a separate authenticated append-only marker before state commit or provider dispatch. |

`Status`, `PreviewedAt`, the state copy of `ApprovalTokenHash`, and `ApprovalConsumedAt` remain
useful lifecycle fields that may reject early but never authorize. Signed destructive lifetime is
accepted only from the authenticated seal/grant; recorded `CreatedAt` and `ExpiresAt` are enforced
only for unsigned plans and likewise cannot affirmatively authorize dispatch.
