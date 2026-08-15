# #4091 discussion log

## Resolved by issue and captain brief

| Question | Decision | Source |
| --- | --- | --- |
| Is approval per run? | No. The preview approval token is single-use, but it creates a durable authorization for the exact, content-free scope. | #4091 corrected approval model |
| Can a scope narrowing reuse authorization? | No. MVP policy requires re-approval for every bound-scope change. | #4091 corrected approval model |
| What is the authorization identity? | Reuse `internal/app/authorization.go`; do not derive a new identity from `reversePlanHash` or transport approval. | #4132 and #4091 |
| What safety evidence is required? | A refusal test records zero provider sends; changed scope is rejected before any request; authorized non-additive modes prove exact destination labels with read-back. | #4091 and task brief |
| Is live credentialed GitHub proof authorized? | No credential/runbook material was supplied to this worker. Record the gap and provide deterministic in-process provider evidence; a separately authorized operator may append live evidence to #4091. | #4091 live-proof section |
