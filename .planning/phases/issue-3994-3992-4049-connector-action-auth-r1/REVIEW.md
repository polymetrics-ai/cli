# Code review

Verdict: PASS — no unresolved actionable findings.

The inline/manual review followed the repository's single-worker fallback and covered the changed
production call paths, typed errors, filesystem durability, cancellation/replay behavior, secret
boundaries, and CLI classification.

| Area | Result | Evidence |
| --- | --- | --- |
| Authorization | Pass | Flow action scope is derived from the stored reverse job and revalidated immediately before dispatch. |
| Replay/races | Pass | Exclusive durable prepared marker and schedule fire lease; selected race test permits one writer. |
| Failure ordering | Pass | Pre-dispatch refusals release the marker; possibly-sent outcomes retain it and park without checkpoint advance. |
| Rate budgets | Pass | Shared-coordinator absence becomes the SDK refusal type/code before requester transport. |
| Secret handling | Pass | No approval token/auth carrier in crontab, argv, flow manifest, schedule manifest, fire state, or CLI JSON. |
| Filesystem safety | Pass | Validated names, atomic temp-file rename, file sync, and parent-directory sync on durable state. |
| Certification correction | Pass | Scripted driver rejects `--authorization`; production glue scans schedule envelopes/crontab for authority carriers and retains direct-flow and byte-identical cleanup assertions. |

Non-finding: the schedule-to-flow uniqueness check is not a substitute for authority; it only makes
the direct installed `flow run` association deterministic. The job's standing authorization remains
the sole approval source.

Corrective review after rebase found no unresolved actionable issue. The raw root check accepts
shell quoting by matching the rooted path and direct-flow fragments independently; secret-value
redaction remains covered by the harness-wide live scan, while carrier checks use exact flag/field
markers to avoid false positives from benign path names.
