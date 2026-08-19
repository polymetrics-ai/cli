# Discussion log — API → API GitHub transport live proof

## Fixed inputs

- Warehouse mediation is mandatory: there is no GitHub-to-GitHub direct hop.
- The exact binary carrier is the existing `pm etl transport github-issue-label`
  plan/preview/approval lifecycle followed by ordinary `pm etl run`.
- The selected definition route is `issues` + `full_append` + `append` +
  `add_issue_labels`.
- `deletes: not_available` describes forward transport deletion behavior. It
  must be reported as honoured: the route does not infer an absent source row
  as a provider delete. The separate cleanup action is an explicit typed
  inverse, not delete propagation.

## Acceptance decisions

| Case | Evidence required |
| --- | --- |
| Happy | Binary JSON reports one read and one loaded record; an independent GitHub `issues` read confirms the exact run-owned label on the target. Durable WAL, Parquet, manifest, acknowledgement and checkpoint facts are recorded without payload or secret material. |
| Bad | Existing typed preflight tests prove both unsupported action/strategy and ineligible source stream are rejected before executor/provider I/O. The live run also uses only the supported action. |
| Edge: zero records | An empty GitHub source result completes without stage, destination write, or checkpoint effects; this is covered by the explicit-empty-source orchestration regression. |
| Edge: absent mapping | A source record without its required positive issue `number` is rejected before destination execution. |
| Edge: replay | Repeat the same approved route and independently confirm exactly one effective label; GitHub label addition is set-like and the keyed route must not duplicate or corrupt state. |
| Edge: deletes | Source deletion is not converted to a GitHub write because the destination declares `deletes: not_available`; separately exercise the declared typed cleanup twice so GitHub's missing-label 404 is accepted rather than silently hidden. |

## Manual GSD fallback

The repository's numeric-roadmap GSD runtime cannot generate an isolated
worker for this named direct-PR proof phase. Its mandatory generated prompts
were resolved and are executed inline. This is a manual runtime fallback, not
a waiver of discuss → TDD plan → execute → verify → code review.
