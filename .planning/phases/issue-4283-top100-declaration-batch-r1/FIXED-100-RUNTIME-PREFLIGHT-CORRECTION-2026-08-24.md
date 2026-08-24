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

## Historical prospective cohort only

Before the later current-main merge, the corrected deterministic selector
produced this 100-row diagnostic snapshot:

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

The post-merge regression projects the full source-locked corpus and asserts
the invariant that matters here—both SCIM rows are absent—without treating this
historical composition as a replacement baseline.

## Superseded hold and restoration

The earlier branch-only reference was not the shipped baseline. Firstmate inbox
item `012.msg` directed its restoration, not a replacement. Commit `4ad21d771`
restored current `origin/main` bytes exactly; the source and restored file both
have SHA-256
`c0d600d323e7effb15c1e092dce6fb590193f23613b17a51917af79e0d74812f`.

After non-rewriting current-main merge `8b6abbf7b`, the normal gate passes:

```sh
go test -timeout 20m -count=1 -run '^TestOperationEvidenceFixed100' ./cmd/connectorgen
go run ./cmd/connectorgen operation-evidence --check .
```

The focused Docker Hub test now projects the full local source-locked corpus,
not a fixture inferred from the restored GitHub-only fixed reference. It still
proves that neither SCIM row is eligible under real preflight. No
`--write-fixed-100` operation was run. Any broader cohort is additive evidence
requiring separate review, never a replacement for the accepted baseline.
