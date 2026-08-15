# Code review — PostgreSQL production transport wiring club

Verdict: pass; no unresolved findings.

## Findings and dispositions

| Severity | Finding | Disposition |
| --- | --- | --- |
| Critical | PostgreSQL's bounded source initially declared `full_overwrite`, but a destination replace applied once per page would retain only the last page. | Fixed: the PostgreSQL source now declares only `full_append` and `incremental_upsert`. The destination still declares all five driver modes, and its safe one-record API overwrite route is proven by the authenticated GitHub binary test. |
| High | The single-use approval was consumed during destination planning but its expiry was not rechecked after source read and warehouse reopen. | Fixed: approval evidence now has a non-consuming binding/expiry validation operation, and PostgreSQL revalidates it immediately before every apply. An expired refusal asserts the typed approval error and an empty side-effect directory. |
| Medium | A failure while finalizing an already-acknowledged transport plan discarded the acknowledged `etlExecutionResult`, weakening recovery evidence. | Fixed: the acknowledged result is returned with the finalization error so `failAcknowledgedTransportRun` retains the durable checkpoint witness. |
| Medium | The inherited one-record GitHub batch rule was keyed only from the destination, so GitHub→PostgreSQL could accept an invalid larger batch. | Fixed: batch policy now considers both resolved endpoints; GitHub on either side requires one, while PostgreSQL→PostgreSQL remains bounded by the caller batch. |
| Low | Comments and refusal text described all approvals as GitHub issue-label-only after the managed-target route was introduced. | Fixed: wording now describes closed definition-selected transports and the PostgreSQL plan binding accurately. |

## Scope and parity review

- PostgreSQL public `write` remains false.
- Destination `change_capture` and `incremental_dedupe_history` are not advertised.
- No raw SQL, arbitrary target identifier, generic HTTP write, or caller-selected warehouse path was introduced.
- The existing #4125 and #4158 defects were not changed. The exact #4158 history/out-of-order live assertion remains disclosed in the PR edge table rather than claimed as green.
- Generated catalog, website data, manuals/skills, CLI docs, and golden transcripts were regenerated in one final pass and their drift checks passed.

## Review route

The official `code-review` prompt was resolved through the project-local GSD adapter and executed inline because the issue-worker contract forbids spawning review roles. GitHub's automatic Claude review is expected after the PR opens; Copilot was not requested because no fallback blocker exists at handoff.
