---
phase: "3863"
status: passed
mode: inline_manual
---

# UAT — secret-free credential coordination identity

All deliverables are objectively covered by local automated tests and checks;
no browser or human-judgment acceptance step applies to this Go-only identity
foundation.

| ID | Deliverable | Automated evidence | Result |
| --- | --- | --- | --- |
| D1 | One explicit binding gives compatible credentials one opaque auth cohort | `TestCredentialCoordination_ExplicitLinkSharesOnlyOpaqueIdentity` | pass |
| D2 | Rate identity needs explicit policy, kind, and subject, and differs by scope | `TestCoordinationIdentity_DerivesDistinctOpaqueScopes` | pass |
| D3 | Binding/approval/secret preimages do not reach identity or inspection output | `TestCoordinationIdentity_ContainsNoBindingOrSecretInput`; app inspection assertion | pass |
| D4 | Identity derivation does not read the vault | `TestCredentialCoordination_IdentityDerivationDoesNotReadVault` | pass |
| D5 | Legacy state migrates and approval lifetime remains distinct | `TestCredentialCoordination_MigratesLegacyMetadataWithoutChangingApprovalLifetime` | pass |
| D6 | CLI linking and help/manual behavior are safe and discoverable | `TestCredentialsCoordinationLinkCLIAndHelp`; `TestCredentialsCoordinationLinkRejectsIncompatibleProfileWithoutEchoingValues` | pass |

No acceptance gap was found. Rate registries, fencing, parking, requester behavior, and transport
dispatch remain intentionally absent because they belong to #3754, #3865, #3867, and #3864.
