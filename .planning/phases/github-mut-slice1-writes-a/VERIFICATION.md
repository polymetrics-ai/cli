# Verification — GitHub mutation certification slice 1 writes-a

## Lifecycle

- `scripts/gsd doctor` — passed.
- `scripts/gsd sources discuss-phase|plan-phase|execute-phase|verify-work|code-review` — resolved through the pinned adapter.
- `scripts/gsd prompt discuss-phase github-mut-slice1-writes-a --auto`, `plan-phase github-mut-slice1-writes-a --tdd --skip-research`, `execute-phase github-mut-slice1-writes-a --interactive`, `verify-work github-mut-slice1-writes-a`, and `code-review github-mut-slice1-writes-a --depth=quick` — generated and recorded as the inline/manual lifecycle fallback.
- `go run ./cmd/agentcontractgen check` — passed.

## Required skills

- `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, and `golang-structs-interfaces` were loaded for this Go connector/CLI certification work.

## Live batch results

### Batch 1 — commands 1–50

- All first 50 assigned GraphQL mutation paths were attempted serially through plan → preview → one-time approval-token execution.
- `abort-queued-migrations` and `change-user-status` returned provider HTTP 200. The status mutation was subsequently rerun with an exact tagged value, independently read through raw GitHub GraphQL, cleared through the provider's own `changeUserStatus` mutation, and independently read back as `status: null`.
- `abort-repository-migration` had an independently queried empty `repositoryMigrations` collection for `Polymetrics-Cert`, so no eligible contained fixture exists.
- Most remaining attempts reached GitHub and rejected the intentionally non-resolving bounded fixture ID with an exact GraphQL node-resolution error. `add-pull-request-review-comment` additionally exposed missing required provider arguments absent from its generated input schema; `add-pull-request-review-thread` returned the certification identity's exact permission denial; `clone-project` returned GitHub's Projects Classic retirement message; and `cancel-sponsorship` returned its exact required-target message.
- Classifications and captured provider messages are retained in the local run state only pending a record shape that can express non-passing command outcomes. Published schema-v2 accepted evidence remains reserved for real, cleanup-proven passes.

### Surface finding — approval transport

- Attempted: `graphql mutation abort-queued-migrations`.
- Plan and preview: the fixed command produced `rplan_d4fdeac5613a853c` and a human preview minted an approval token (redacted; never persisted or placed on argv).
- The checked-out integration base rejects `--approve` and requires `--approval-token-stdin`; the token was passed as one bounded stdin line. This is a local one-time approval grant, not a GitHub credential; the Keychain PAT never reached argv, a file, command output, or evidence.
- Captain decision: approval-token transport is a surface design choice, not a provider product defect.

### Product defect — abort-queued-migrations proof is absent

- Executed: `graphql mutation abort-queued-migrations` using the disposable `Polymetrics-Cert` owner. The approval-bound run completed with provider HTTP 200 and `records_succeeded=1`.
- Provider control: raw GitHub GraphQL introspection returned `AbortQueuedMigrationsPayload` fields `clientMutationId` and `success`.
- Defect: the fixed declaration in `internal/connectors/defs/github/operations.json` selects only `abortQueuedMigrations { __typename }`. It omits provider `success`, and the command has no independent read surface for queued migration state. The returned typename is compatible with both a true effect and a no-op/wrong effect, so it fails the task's required observable assertion and cannot support cleanup proof.
- Classification: `product_defect`, not `certified` and not `no_object`. The live request reached GitHub; its effect cannot be honestly proven from the generated fixed response contract. The separate raw provider collection read was empty, so there was no disposable migration object to create a stronger control without expanding the fixture contract.

### Completed proof — change-user-status

- A fresh plan → preview → approval-token-stdin → run completed for `graphql mutation change-user-status` with an exact `pm-cert-status-*` message on the disposable certification user.
- Independent raw GitHub GraphQL read-back asserted the viewer login was `polymetrics-ai-certification` and the returned status message exactly matched the tag. This agent-derived assertion rejects both an absent status and a plausible wrong status.
- GitHub exposes no DELETE for a user status. The direct provider cleanup used raw GraphQL `changeUserStatus(message: "")`; the independent final raw read returned `status: null`. Cleanup is therefore accurately classified `contained_closed`, with the disposable certification user as the container, never `verified_absent`.
- The proof constructor wrote a schema-v2 record with four sanitized exchanges and a Keychain-PAT fingerprint. `go run ./cmd/connectorgen certification-matrix --check` then correctly reported only that the shared GitHub certification shard would drift. The task contract reserves shared artifact regeneration to the integration pass, so the record was deleted rather than leaving a failing branch. This is an integration-artifact ownership constraint, not a missing live proof.

### Completed provider lifecycle — create-branch-protection-rule

- `graphql mutation create-branch-protection-rule` created a uniquely tagged `pm-cert-branch-rule-*` rule in `Polymetrics-Cert/pm-cert-3993-20260810-wz0fru` through plan → preview → approval → run.
- Raw GraphQL collection read-back asserted the exact tagged pattern and rejected a missing or wrong pattern. Raw provider `deleteBranchProtectionRule` cleanup then removed it; the final independent collection read excluded both the exact rule ID and pattern (`verified_absent`).
- A temporary proof-materialization harness encountered a raw GitHub 502, then revealed its own malformed cleanup JSON on the one permitted retry. Both created tagged rules were subsequently removed with raw provider deletes and a collection absence read-back. No evidence record was retained from that flawed harness attempt; this is a local harness defect, not a provider result.

### Batch 2 partial — commands 51–79

- Commands 52–78 were attempted serially through the fixed plan → preview → one-time approval-token-stdin → run lifecycle. Commands 52 and 53 reached GitHub and returned the provider's exact GitHub-App-only authentication requirement. Commands 54–66, 69–75, 77, and 78 reached GitHub with bounded non-resolving node fixtures and returned exact node-resolution errors without creating provider state.
- Command 67 first enforced its declaration-owned `env_only` input carrier, then enforced typed destructive confirmation. The bounded retry supplied both through the declared channels and reached GitHub, which rejected the non-resolving owner ID. No credential or approval token was printed, persisted in evidence, or placed on argv.
- Command 68 (`create-project`) returned GitHub's Projects Classic retirement message. A separate raw `curl` GraphQL control to `api.github.com` returned HTTP 200 with the same `NOT_FOUND` provider message. Classification: `provider_deprecated`, not `product_defect`; PM matched the provider control and the command remains declared and implemented per the captain's ruling.
- Command 76 (`create-repository`) completed twice with uniquely tagged private repositories owned by `polymetrics-ai-certification`. The proof replay independently queried exact `name`, `nameWithOwner`, `isPrivate: true`, and owner login; that assertion rejects an absent repository, a public repository, or a plausible wrong owner/name. Raw provider DELETE returned 204 and an independent REST read returned 404 (`verified_absent`).
- `internal/connectors/certifications/evidence/github-agent-derived-graphql-mutation-create-repository-20260818.json` records the fixed GraphQL mutation, agent-derived produced-value read-back, raw provider DELETE, and independent 404 cleanup exchange using repository-salted fingerprints only. `go run ./cmd/connectorgen certification-matrix --check` accepted the record without shared-artifact regeneration.
- Command 79 (`create-sponsors-listing`) was not executed. It creates public visibility under the organization/account, one of the four explicit captain escapes. The captain explicitly declined it; classification is permanently `escape_needs_captain` with no provider request issued.

### Classification re-audit

- Fake or wrong-type node IDs do not prove `no_object`. Commands 54–66, 67, 69–75, 77, and 78 remain incomplete pending a real parent collection and disposable fixture lifecycle. None is banked as `no_object`.
- Commands 52 and 53 remain incomplete pending the mandated GitHub App credential retry. Their classic-PAT App-only response is not banked as `wrong_credential`.
- Cleanup proves containment, while the mutation read-back proves correctness. A successful mutation is not downgraded because the direct terminal disposal reports the object already absent.
- `product_defect` is reserved for PM behavior that disagrees with a raw `api.github.com` control. Matching provider retirement or node-resolution behavior is not a product defect.

### Commands 80–100

- Commands 80–82 reached GitHub but are not banked: the Sponsors tier depends on the explicitly declined public Sponsors listing, while both sponsorship creates require a real sponsorable target and remain pending a contained fixture/control analysis.
- Command 83 (`create-user-list`) created an exact private `pm-cert-` list for `polymetrics-ai-certification`; a raw node query proved name, description, privacy, and owner. Raw `deleteUserList` plus an independent `node: null`/`NOT_FOUND` read proved cleanup. The schema-v2 record passes `certification-matrix --check`.
- Command 84 (`decline-topic-suggestion`) is `no_object`: the real repository's topic-suggestion collection is empty, GitHub exposes no API to create a topic-suggestion fixture, and the fixed command plus raw provider behavior report that topic suggestions are unsupported. This is not inferred from a fake ID.
- Commands 85–97 are certified using fresh provider-created objects: branch protection rule, deployment, discussion, discussion comment, environment, disabled IP allow-list entry, issue, issue comment, issue field, issue field value, issue type, label, and linked branch. Each PM mutation was followed by an independent exact absence/collection read, provider terminal disposal, and a second absence read. Disposable parent containers and temporary repository settings were also removed or restored.
- Command 98 (`delete-package-version`) is `no_object` for the GraphQL command's eligible object type. All real organization package collections were first listed. A private GHCR `pm-cert-` OCI package/version was then created, but GitHub's GraphQL `PackageType` surface exposes only `DEBIAN` and `PYPI`; the real GHCR REST version is not a GraphQL node and both PM and raw node resolution reject its numeric REST ID. The raw provider package DELETE removed the fixture and independent reads returned 404 for both package and version.
- Commands 99 and 100 (`delete-project`, `delete-project-card`) are `provider_deprecated`: the real Projects Classic parent collection and a raw `createProject` fixture attempt both return GitHub's Projects Classic retirement response. PM and raw controls agree that the bounded organization ID is not a Project/ProjectCard node. No product defect is claimed.

Schema-v2 passing evidence was written immediately for commands 85–97. `go run ./cmd/connectorgen certification-matrix --check` and `git diff --check` both pass for the accumulated evidence set.

### Second attempted command — not classifiable as absence

- Executed: `graphql mutation abort-repository-migration` with the real fixture repository node ID supplied where the operation requires a `RepositoryMigration` node.
- Provider result: GitHub rejected the request with `Could not resolve to RepositoryMigration node with the global id of '<repository node id>'`.
- This is deliberately **not** recorded as `no_object`: the attempt proved the ID is the wrong node type, not that an eligible migration collection was read and found empty. The command surface supplies neither a migration discovery read nor a fixture lifecycle/cleanup contract.
- Combined with the first command's unobservable effect, further runs would repeat the same unprovable-mutation obstacle. The mandatory cleanup contract requires stopping rather than manufacturing a success classification.

## Final checks

`go run ./cmd/connectorgen certification-matrix --check` — passed with the cleanup-proven `create-repository` schema-v2 record and no shared-artifact regeneration. The required product fix for `abort-queued-migrations` remains to project effect-bearing fields (at minimum `success` where available) and declare an independent state/read-back and cleanup contract.

## Captain escape

The earlier sponsorship escape is resolved by the captain's later authorization for contained charges below $2 with immediate teardown; the attempted command itself returned GitHub's required-target rejection and did not create, cancel, or charge a sponsorship.
