# DISCUSSION LOG — issue #3863

The GSD `discuss-phase` prompt was generated with `scripts/gsd prompt discuss-phase 3863` and
executed inline. No interactive product decision remains: the issue, parent, and architecture report
fix the required distinctions.

| Question | Resolved decision | Source |
| --- | --- | --- |
| Is one credential key sufficient for both concerns? | No. One builder yields a binding-only authentication cohort and a policy/scope-specific rate key. | #3863; architecture report, “Credential identity: one builder, two correct scopes”. |
| Can identity be inferred from secret material, credential ID, or revision? | No. It is an explicit durable non-secret binding; approval revisions remain secret-derived rotation evidence only. | #3863; `internal/connectors/approval.go:181-205`. |
| May a rate key be created without a provider declaration? | No. There is no fallback key; rate scope requires policy ID, supported kind, and non-secret subject. | #3863; #3754. |
| How can an operator link credentials without revealing a binding preimage? | Link by credential name after exact provider-family/auth-profile compatibility checks. | #3863 required contract. |
| May this slice implement registry, requester, transport, fence, or parking behavior? | No. Those remain #3754, #3752/#3753, #3864, #3865, and #3867. | Parent #3862 and child issues. |

The answer is intentionally secret-free: neither this log nor the plan carries a credential value,
secret-derived equality, or binding preimage.
