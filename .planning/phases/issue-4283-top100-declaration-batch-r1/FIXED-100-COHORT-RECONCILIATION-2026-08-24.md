# Fixed-100 cohort reconciliation — 2026-08-24

Scope: read-only reconciliation requested for the open
`dockerhub-open-object-structured-body` decision. No engine, selector, source
lock, declaration, or generated evidence artifact was changed.

## Correction

The previous open-object inventory incorrectly described the branch's current
fixed-100 file as an immutable constraint on Docker Hub descope. That premise
does not hold.

| Reference | Cohort composition |
| --- | --- |
| `origin/main` | 100 GitHub rows; 0 Docker Hub rows |
| branch parent of `36079c9c9` | 100 GitHub rows; 0 Docker Hub rows |
| `36079c9c9` and current branch | 37 GitHub, 33 Asana, 25 Docker Hub, 3 Jira, 1 Bitbucket, and 1 CircleCI row |

Commit `36079c9c9` (`fix(connectors): remove unsupported typed destinations`)
changes `internal/connectors/operation-evidence-fixed-100.json` by 571
insertions and 576 deletions. It replaces 63 of the 100 original GitHub rows.
Both Docker Hub SCIM rows enter in that unmerged branch re-selection; neither
is present in the `origin/main` baseline.

Accordingly, a partial/unsupported disposition fails this branch's generated
fixed-cohort check only because this branch put those rows in the cohort. It is
not evidence that a shipped baseline makes the operations immovable.

## 1. Authority to rewrite the cohort

I find **no explicit authority** for the cohort rewrite.

The issue delivery header scopes this lane to ten named connector bundles and
their source-backed declarations (`PLAN.md:6-9`); it neither requests nor
authorizes `operation-evidence --write-fixed-100`. The record at
`PLAN.md:673-683` says that the command regenerated the 5,903-row evidence and
the fixed reference after a stale-artifact failure, but records an action rather
than an approval for changing the baseline.

The only explicit fixed-cohort instruction located in the branch goes in the
opposite direction: the authorized downstream shared-test repair says the file
“must not be regenerated, re-selected, or rewritten”
(`PLAN.md:266-285`). That later instruction is scoped to its test-repair slice,
so it cannot retroactively establish authority for commit `36079c9c9`; it does
confirm that a selector run is not self-authorizing.

The generator exposes a technical write path at
`cmd/connectorgen/operationevidence.go:194-209`, but a command option is not a
delivery authorization. No task brief, repository procedure, GSD record, or
commit message inspected grants a decision to replace the main cohort.

## 2. Whether the two SCIM rows are a free selection choice

There are two distinct choices:

1. **Rebuilding the shipped cohort at all** was a discretionary branch action;
   no authority for it is recorded above. Restoring the untouched `origin/main`
   100 GitHub reference is the smallest way to avoid making the SCIM rows
   fixed-cohort anchors, without an engine change.
2. **Once the current generator is invoked over the branch artifact**, the two
   rows are not individually optional. They are deterministically selected by
   the current code, even though commandrunner preflight later rejects them.

The generator's actual candidate universe is also narrower than the 552
connector definition directories: it reads the **11** connectors that have
operation source locks in the current artifact. That artifact has 779 rows
meeting its eligibility proxy, including 276 `direct_write` rows. The six-class
counts are 38 ETL, 148 reverse ETL, 303 direct read, 276 direct write, 16 binary
download, and 0 binary upload.

The proxy does not call `commandrunner.Preflight`. It marks `runtime.enabled`
true when an operation has a target and an `availability: implemented` command
(`cmd/connectorgen/operationevidence.go:706-723`). As a result, both SCIM rows
look eligible to the evidence projector even though the actual preflight reaches
the closed-body refusal. This mismatch is the direct cause of their selection,
not a source or provider requirement.

## 3. Exact forcing rule

`cmd/connectorgen/operationevidence.go:351-356` sorts all projected rows by
connector then source ID. The class order is ETL, reverse ETL, direct read,
direct write, binary download, and binary upload at lines 32-47.

`buildOperationEvidenceFixed100` then takes the first 20 eligible, unselected
rows for each class (`cmd/connectorgen/operationevidence.go:1109-1145`).
`operationEvidenceFixedEligible` requires only the projector's absence/gap,
enabled, conformance, fixture, CLI, website, and classification fields
(`cmd/connectorgen/operationevidence.go:1153-1162`), not actual runtime
preflight.

After 20 ETL, 20 reverse-ETL, and 20 direct-read rows, the direct-write sample
starts at fixed-cohort positions 61–80. Sorting makes these its first two
candidates:

| Direct-write rank | Source ID |
| ---: | --- |
| 1 | `dockerhub.rest.post_/v2/scim/2.0/Users` |
| 2 | `dockerhub.rest.put_/v2/scim/2.0/Users/{id}` |
| 3 | `github.graphql.mutation.abortQueuedMigrations` |

Therefore, a current-generator rebuild forces the two rows through the
20-per-class sorted-sample rule. A conformant rebuilt cohort would need its
eligibility rule to include the real preflight result, or it would need an
explicitly authorized baseline selection that does not use this branch's stale
proxy. Neither action has been taken here.

## Outcome for the open-object decision

The source-faithfulness finding remains: Docker Hub's pinned SCIM request
objects enumerate fields but are open, so connector-local
`additionalProperties: false` would invent a provider restriction. That is
separate from the cohort error.

No engine foundation or captain decision is needed merely to keep these two
currently non-shipped SCIM operations out of the fixed-100 baseline. The
remaining decision is whether to restore the `origin/main` cohort as the branch
baseline or to authorize a separate, runtime-preflight-backed cohort-selection
change. This report makes no such change.
