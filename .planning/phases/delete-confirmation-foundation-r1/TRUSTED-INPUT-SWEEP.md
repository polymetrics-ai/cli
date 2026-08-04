# Trusted-input sweep — destructive write approval

The execution gate accepts only facts recovered from a vault-authenticated grant and the current
engine-prepared write. State JSON carries the grant and lifecycle hints, but no mutable state field
can independently authorize dispatch.

| Fact | Read at authorization | Caller or agent writable? | Trusted source and tamper defense |
| --- | --- | --- | --- |
| Grant envelope | `WriteApprovalAuthority.VerifyWriteGrant` | The persisted JSON copy is writable | HMAC-SHA256 covers the complete grant; an altered envelope fails before consumption. |
| `Version` | grant verification | Yes, through state JSON | Covered by the grant MAC and checked against the closed runtime version. |
| `AuthorityID` | grant verification | Yes, through state JSON | Derived from the sealed project-vault root and covered by the grant MAC; caller-key authorities produce untrusted evidence that the engine refuses. |
| `PlanID` | app consumption and grant verification | Yes, through state JSON or request | Compared with the executing plan and covered by the grant MAC. |
| `PlanHash` | app recomputation and grant verification | Yes, through state JSON | Recomputed from the current command/bulk payload, covered by the plan seal, and covered again by the grant MAC. |
| `Mode` | app dispatch and grant verification | Yes, through state JSON | Covered by the plan seal and grant MAC; command and bulk modes cannot be substituted. |
| `PlanSealMAC` | grant issuance and verification | Yes, through state JSON | A production grant is issued only from a current authenticated plan seal; the seal MAC is then covered by the grant MAC. |
| `PreviewDigest` | app re-preview, grant verification, engine gate | Yes, through state JSON | Recomputed from every canonical request and complete definition/hook identity, then covered by the grant MAC. |
| `ApprovalTokenHash` | grant verification | The state copy is writable; the raw token is caller input | The authoritative hash is inside the grant MAC and is compared in constant time with the presented token. The state copy can only cause an early rejection. |
| `Nonce` | grant verification and vault consumption | Yes, through state JSON | Random at grant issuance, covered by the grant MAC, and persisted only as a vault-root HMAC inside the authenticated consumption record. |
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
| `ExpiresAt` | plan/grant verification and engine evidence | The persisted copy is writable | Plan and grant deadlines are authority-generated and MAC-bound. Grant lifetime is capped from trusted current time and the signed plan deadline; mutable `ReversePlan.ExpiresAt` is display-only. |
| Grant `Confirmation` | grant verification | The persisted copy is writable | Covered by the grant MAC and compared with the explicit typed `--confirm destructive` value. |
| Grant `MAC` | grant verification | The persisted copy is writable | Verified with the project-vault root before any fact is accepted. |
| Consumption record | grant issuance and verification before state replacement | Not writable through state mutations | One authenticated opaque marker per authority/plan ID/hash/mode is created with `O_CREATE|O_EXCL` before dispatch and contains the opaque consumed nonce identity. Grant issuance checks the same marker, so rollback cannot re-preview the consumed plan with a fresh nonce. |
| Authority/key | App open and authority authentication | No raw key is accepted from write callers or engine requests | `App.Open` receives an opaque root whose fields are constructible only by the opened project vault. Caller-key authorities are explicitly untrusted and their evidence is rejected. |

## Whole-state write audit

| Write path | Previous risk | Current rule |
| --- | --- | --- |
| `App.save` used by credentials, connections, catalogs, ETL runs, plans, and secret-field metadata | Replaced the complete file from a stale in-memory snapshot | Compares the loaded `state.Revision` under the state lock and rejects a stale revision before replacement. |
| Approval preview, consumption, completion, and invalidation | Locked reload/update but did not advance a version understood by other writers | All use `App.updateState`, which reloads under lock and advances `state.Revision`. |
| Project initialization | Creates the state file only when absent | Initializes revision zero and never overwrites an existing state file. |
| Vault consumption | Previously absent, so a valid older state snapshot restored a grant | Creates a separate authenticated append-only marker before state commit or provider dispatch. |

`Status`, `PreviewedAt`, the state copy of `ApprovalTokenHash`, `ApprovalConsumedAt`, `CreatedAt`, and
`ExpiresAt` remain useful lifecycle/display fields. They may reject an operation early, but none is
an affirmative authorization input.
