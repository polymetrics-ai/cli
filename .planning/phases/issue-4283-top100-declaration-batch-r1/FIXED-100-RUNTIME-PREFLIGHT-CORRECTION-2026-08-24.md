# Fixed-100 runtime-preflight correction — 2026-08-24

Decision source: Firstmate inbox item `010.msg`.

Scope: `cmd/connectorgen/operationevidence.go` and its focused unit test only.
Docker Hub declarations, source locks, generated checked-in operation evidence,
the checked-in fixed-100 reference, and the closed-body engine rule are
unchanged. No credential or provider request was used.

## Defect and correction

The selector previously set `runtime.enabled` when an operation had a target
and an `availability: implemented` command. That was metadata evidence, not the
binary's executable contract. It admitted Docker Hub's SCIM create/update rows
even though their command paths stop in the runtime's closed structured-body
preflight.

`operationEvidenceRuntimePreflight` now creates the declarative connector with
`engine.New(bundle, engine.HooksFor(bundle.Name))` and delegates each matching
implemented command to `commandrunner.Preflight`. It reports a
`runtime_reachability` gap rather than accepting a metadata-only claim. This is
the exact production command-runner entry point; it introduces no parallel
schema/body validation.

## Red / green evidence

Red command:

```sh
go test -timeout 20m -count=1 -run '^TestOperationEvidenceFixed100UsesRuntimePreflightForDockerHubSCIMWrites$' ./cmd/connectorgen
```

Before the correction it failed at the create row:

```text
Docker Hub SCIM row "dockerhub.rest.post_/v2/scim/2.0/Users" is runtime-enabled despite commandrunner preflight refusal
```

After the correction the same test passes. It proves both source IDs have
disabled runtime and `runtime_reachability`, then builds an in-memory
prospective fixed cohort and requires neither ID appears.

## Prospective cohort only

The corrected deterministic selector would produce exactly 100 rows:

| Connector | Rows |
| --- | ---: |
| Asana | 33 |
| Bitbucket | 1 |
| CircleCI | 1 |
| Docker Hub | 23 |
| GitHub | 39 |
| Jira | 3 |
| **Total** | **100** |

The two Docker Hub SCIM writes fall out by normal eligibility, not a
hand-authored exclusion. Their two direct-write slots are filled by the next
eligible sorted GitHub rows.

## Deliberate hold

The checked-in branch fixed reference still contains the two SCIM rows. Thus:

```sh
go test -timeout 20m -count=1 -run '^TestOperationEvidenceFixed100' ./cmd/connectorgen
```

correctly fails with
`dockerhub.rest.post_/v2/scim/2.0/Users execution evidence regressed`. This is
expected evidence that the old unmerged selection was unsound. Per the decision,
no regenerated cohort is written or committed; the captain decides whether to
restore the shipped all-GitHub baseline or approve another baseline change.
