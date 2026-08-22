# Foundation 0.3.0 release-candidate r1 review record

source_sha: a5005fae7f1e92d19ef4b8e82d514050227bec38

This is the final code-and-generated implementation SHA **I**. The evidence closure commit is deliberately above it: a committed evidence file cannot truthfully self-name its own Git object. The branch ref and PR record provide the immutable evidence-commit SHA after the strict graph gate passes.

## Intake and composition crosswalk

| Item | Disposition and current witness |
| --- | --- |
| Original 38 blocker categories | Historical and repaired: the two authoritative read-only audits trace every category to committed repairs. No category was reimplemented here. |
| Exact core input | `041d2ec7ed986aea15d2d3d64f2076b484c3f999`; origin predecessor `524850d…` was an ancestor, non-force advanced, and exact-verified with `git ls-remote`. |
| Exact reverse input | `487ec14e01c90f31a71b1cb5de060b8c66a203e9`, preserved by merge `50e90fa854635b7c8b295b7090034b82a52e4e03`. |
| Exact public input | `7fdef00d7e758cb4a3c413a16f8452ee0615f0d5`, preserved by merge `e18f4372f4f65ae5e42265f237abad79473a7425`; origin predecessor `2d6143b…` was non-force advanced and remote-verified. |
| Already-inherited components | Source import `19a32bd…`, closed operation `3c768ca…`, structured body `66b4c3c…`, and reverse destination `d814875…` are ancestors; none was remerged or cherry-picked. |
| RC-05 current-code witness | The exact release run exposed a stale certification-sweep test and stale embedded source-lock checksum. `surface-sync --check` directly established the declaration is `implemented`, so the narrow correction updates only tests to match the current source-authoritative surface. It does not change runtime behavior, authority, or an output claim. |

## Conflict-invariant preservation

| Seam | Preserved invariant | Evidence |
| --- | --- | --- |
| Core + reverse: App, Postgres transport, sync transport | Core empty-publication, authentication, deadline, and acknowledgement fences coexist with reverse action-owned receipt/readback binding. | Full `internal/app`, race-enabled Postgres, `synctransport`, `commandrunner`, and connector cohorts pass from I. |
| Core + public: connector registry, engine direct-read/headers, SQS, output tests | Exact configured-secret masking is retained without corrupting ordinary keys, identifiers, raw receipt bytes, or declaration-owned parameter authority. | Engine, connector, commandrunner, SQS, and full CLI/App tests pass; App regression preserves raw provider receipt bytes except configured secret masking. |

## FND-B09 and FND-W01

- **FND-B09 — reproduced and closed by this evidence-only commit:** before closure, the inherited manifest deterministically failed:
  `parse evidence manifest: json: cannot unmarshal object into Go struct field foundationEvidenceManifest.component_inputs of type []main.foundationEvidenceInput`.
  The schema-3 manifest binds I, the typed current component list, subject fingerprint, and digests. The strict gate is run from the clean closure commit.
- **FND-W01 — deferred / non-reproducing:** current `sourceimport` flow and the hermetic `TestSourceImport.*` cohort were inspected and passed. No deterministic externally observable publication-order failure was established. No speculative transaction/publication rewrite was made.

## Behavioral proof, provider proof, and certification claim

| Command or surface | Local behavioral / fixture proof | Current real-provider proof | Certification claim |
| --- | --- | --- | --- |
| Foundation App/transport/reverse/public-output seams | Full hermetic App, CLI, engine, commandrunner, sync transport, SQS, Postgres-race, smoke, and output cohorts pass. | No new provider execution: no approved non-secret credential reference was supplied; none was inferred or probed. | Not a connector certification claim. |
| `github actions fork-pr-contributor-approval view` | Source-sync and certification-sweep generation preserve its declared `implemented` surface and observed 422 refusal record. | The retained historical 422 observation is not a new, I-bound live proof. No live call was executed. | Implemented but **uncertified**. |
| Every other source-mapped GitHub command/surface | Declaration validation, surface-sync, candidate, sweep, and binary-subject checks pass; this is not individual behavioral/provider proof. | No current real-provider evidence was run in this lane. | Each remains only its declared implemented/partial state; none is certified by this RC. |

`make connectorgen-certification-matrix` reports `connectors=3 capability_complete=0 certified=0`. Thus no command is labeled certified on the strength of source mapping, a fixture, or an old provider observation.

## Commands and results from I

| Command | Result |
| --- | --- |
| `go test -timeout 20m ./cmd/connectorgen ./internal/connectors/defs -count=1` | PASS; validates the RC-05 red/green correction. |
| `go test -timeout 20m ./internal/app -count=1` | PASS (279.841s). |
| `go test -timeout 20m ./internal/cli -count=1` | PASS (854.639s). |
| `go test -timeout 20m ./internal/connectors/engine ./internal/connectors ./internal/connectors/commandrunner ./internal/synctransport ./internal/connectors/native/amazon-sqs -count=1` | PASS. |
| `go test -race -timeout 20m ./internal/connectors/native/postgres -count=1` | PASS. |
| `gofmt -d cmd internal`, `go mod tidy -diff`, `go vet ./...`, `git diff --check` | PASS; no formatting, module, vet, or whitespace drift. |
| `go build -o ./pm ./cmd/pm`, `make docs-check-no-build`, `make smoke-no-build` | PASS. |
| `make lint`, `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync` | PASS; 552 connectors and zero validator findings/surface corrections. |
| `make github-parity-artifacts-check`, certification subject/matrix/candidates/sweep, connector-boundary, connector-canon, release-workflow-check, website docs test | PASS before the test-only I correction; regenerated/surface/certification checks were repeated from I and pass. |
| `go test -timeout 20m ./...` on the predecessor I | RED outside this RC: certification-harness count expectations and a Recurly binary-download test. Recurly is explicitly nonblocking under the later exact-I independent green review evidence; the certification count failures remain a separate non-certified-surface concern, not a pass or waiver. |

## Merge recommendation

Do **not** merge to `main` from this record. The Foundation conflict-adjacent and release-local evidence is green, but final merge authorization still needs the strict evidence gate on E, remote/PR identity verification, CI/no-mistakes review, and an explicit disposition of the remaining aggregate certification-harness count failures. The PR remains unmerged and targets `fm/cli-current-foundations-main-integration-r1`.
