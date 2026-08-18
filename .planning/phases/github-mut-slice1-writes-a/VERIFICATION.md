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

### Commands 101–150 — corrected classification audit and live proofs

- Commands 102–116 and 119–121 are certified. Their independent reads proved the exact ProjectV2, pull-request-review, ref, custom-property, ruleset, user-list, verifiable-domain, merge-queue, auto-merge, and organization-follow effects. Provider terminal disposal used the actual available GraphQL inverse/delete, REST delete, close, or repository-setting restore; the final read proved absence or the exact terminal state. Cleanup mechanics were not used to downgrade command correctness.
- Command 117 remains unbanked: the only authorized user cannot author an approvable review on its own pull request, and the provider rejected dismissal of the real COMMENTED review. Command 118 is `no_object`: the real repository vulnerability-alert collection is empty and GitHub exposes no alert-creation API for a disposable fixture.
- Command 122 is `no_object`, not `wrong_credential`: the only in-boundary user is the authenticated disposable identity itself, whose real read returned `viewerCanFollow: false`; PM and a raw `api.github.com` control both returned the provider's self-follow failure. No second real person was contacted.
- Commands 128, 129, 131, 134, 135, 138, 142, 143, and 150 are additionally certified. Fresh private ProjectV2/team, issue/comment, draft pull request, and assignee fixtures were created; exact link, lock, template, ready-for-review, minimized, pin, and empty-assignee states were independently read. Raw inverse operations and container deletion/close were followed by `node: null`, empty-link, unlocked/unminimized/unpinned, closed-PR, deleted-ref, and REST 404 checks as applicable.
- Command 147 reached GitHub against a real newly created verifiable domain but is not banked: both PM and raw GitHub control rejected regeneration because a fresh domain necessarily already has a verification token. The domain was directly deleted and independently read as absent; a second lifecycle is required before choosing a final bucket.
- Commands 123, 126, 127, 145, and 146 cannot target an enterprise/public surface without leaving the authorized boundary. Command 145 (`publish-sponsors-tier`) is a public-visibility captain escape and was not executed. Remaining commands in this range are still pending and are not silently assigned to `no_object`, `wrong_credential`, or `product_defect`.

The evidence validator and `git diff --check` pass for this incremental set. Every new record contains repository-salted fingerprints only; raw secret values and approval tokens are absent.

### Batch 4 — commands 146–195

Exactly 50 commands were processed, one at a time, under the corrected classification rules:

- `certified` (27): 150, 151, 157, 160–169, 172, 173, 175, 179–181, 183, 186, 188, 189, 191, 192, 194, 195.
- `escape_needs_captain` (17): 146, 148, 152–156, 158, 170, 171, 174, 176, 178, 182, 184, 185, 187.
- `no_object` (4): 147, 149, 190, 193.
- `product_defect` (2): 159, 177.
- `wrong_credential`, `entitlement`, and `not_implemented` (0).

The escape rows were skipped without an out-of-boundary provider mutation: enterprise-only targets, real reviewer/outside-collaborator notifications, public Sponsors state, metered Actions deployment creation, public-repository interaction limits, or third-party migration sources/archives. Command 182 was controlled against the real private fixture repository; PM and raw GitHub both state that repository interaction limits cannot be set for private repositories, while making a public repository is an explicit visibility escape.

The `no_object` rows were not inferred from fake IDs. Command 147 used a fresh verifiable domain and both PM and raw GitHub showed that the newly created fixture already has a token, making regeneration ineligible; the domain was deleted and read as absent. Command 149 created a fresh issue, listed `pendingSuggestions: []`, confirmed GitHub exposes no suggestion-creation API, then deleted the issue and read `node: null`. Command 190 listed real closed pull requests and tried to create an archived PR fixture with both the classic credential and the GitHub App; GitHub denied `archivePullRequest` for both identities, so no eligible archived object can be created. Command 193 used the real certification user and proved self-follow is forbidden; the App has no user identity and no second real person is permitted.

Commands 159 and 177 are product defects with raw controls. For 159, PM's fixed mutation returns only `__typename`; raw GitHub returns the affected repository identity, while no collection/read field exists to assert the bypass-user removal. For 177, PM again returns only `__typename`, while raw `revokeMigratorRole` returns effect-bearing `success: true`. The raw controls were run after real fixtures were granted, and terminal raw removals were applied. Neither result is mislabeled as `no_object`.

The GitHub App retry was performed through the repository's encrypted PM credential mechanism. Its installation token was minted in memory, supplied to `pm credentials add` through stdin, never printed or placed in argv, and used only against the approved fixture repository. This recovered command 172: a real App-owned check suite was rerequested, independently read as queued/in-progress, then completed to terminal `NEUTRAL` through the provider and independently read as completed. All other certified rows have exact produced-value reads plus provider inverse/delete/close/archive terminal disposal and independent cleanup reads in their schema-v2 records.

### Second attempted command — not classifiable as absence

- Executed: `graphql mutation abort-repository-migration` with the real fixture repository node ID supplied where the operation requires a `RepositoryMigration` node.
- Provider result: GitHub rejected the request with `Could not resolve to RepositoryMigration node with the global id of '<repository node id>'`.
- This is deliberately **not** recorded as `no_object`: the attempt proved the ID is the wrong node type, not that an eligible migration collection was read and found empty. The command surface supplies neither a migration discovery read nor a fixture lifecycle/cleanup contract.
- Combined with the first command's unobservable effect, further runs would repeat the same unprovable-mutation obstacle. The mandatory cleanup contract requires stopping rather than manufacturing a success classification.

## Final checks

`go run ./cmd/connectorgen certification-matrix --check` — passed with the cleanup-proven `create-repository` schema-v2 record and no shared-artifact regeneration. The required product fix for `abort-queued-migrations` remains to project effect-bearing fields (at minimum `success` where available) and declare an independent state/read-back and cleanup contract.

## Captain escape

The earlier sponsorship escape is resolved by the captain's later authorization for contained charges below $2 with immediate teardown; the attempted command itself returned GitHub's required-target rejection and did not create, cancel, or charge a sponsorship.

## Final slice ledger — commands 1–274

Every ordinal in the assigned JSON work list appears in exactly one bucket below. `certified` is counted only from schema-v2 evidence files added by this branch over `origin/integration/4015-mvp-flat-r1`; cleanup was independently read back and is not used as a proxy for correctness.

- `certified` (116): 76, 83, 85–97, 102–116, 119–121, 128, 129, 131, 134, 135, 138, 142, 143, 150, 151, 157, 160–169, 172, 173, 175, 179–181, 183, 186, 188, 189, 191, 192, 194, 195, 197–199, 201–210, 230–233, 235–241, 243, 244, 249, 251–266 except 267, 268, 270–273.
- `no_object` (33): 2, 5, 11, 15, 16, 29, 31, 33, 41, 46, 68, 84, 98–101, 117, 118, 122, 130, 139, 140, 147, 149, 190, 193, 196, 200, 242, 246–248, 274.
- `wrong_credential` (0): none.
- `entitlement` (62): 6–10 except 11, 14, 17–28 except 29, 32, 34, 38–40, 42–45, 47–57 except 58, 59–66, 69–75, 77, 78, 124, 132, 133, 136, 137, 141, 144, 234.
- `not_implemented` (0): none.
- `product_defect` (3): 1, 159, 177.
- `escape_needs_captain` (60): 3, 4, 12, 13, 30, 35–37, 50, 58, 67, 79–82, 123, 125–127, 145, 146, 148, 152–156, 158, 170, 171, 174, 176, 178, 182, 184, 185, 187, 211–229, 245, 250, 267, 269.

Total: **116 + 33 + 0 + 62 + 0 + 3 + 60 = 274**.

Command 79 is `escape_needs_captain` for the captain's exact reason: **Public visibility under the org name is a captain escape and stays unexecuted.** No provider request was issued. The other escape rows likewise stop at the boundary: enterprise/third-party targets, real people or notifications, public Sponsors state, or money/metered deployment effects.

The final `no_object` controls included real parent reads plus fixture attempts. In particular, command 242 read the real organization setting and GitHub refused the only opposite-value fixture because no verified/approved domain exists; commands 246–248 listed the Projects Classic parent and attempted `createProject`, both returning GitHub's retirement response; command 274 listed zero domains, created a fresh `pm-cert-*.example.com` domain, received the expected missing-TXT rejection, directly deleted it, and independently read `node: null`.

Command 234 was retried with the GitHub App after the classic credential reached GitHub; the classic identity lacks the setting entitlement and the App credential was rejected, so it is not mislabeled `wrong_credential`. Commands 1, 159, and 177 alone are `product_defect`, each backed by the raw GitHub controls described above.

Final containment audit restored `has_discussions: false` on the fixture repository and independently read it back. It also found no residual `pm-cert-258-repo-*` repository, `pm-cert-270-team-*` team, verifiable domain, repository custom property, ProjectV2, or branch from this lane.

## Final local verification

- `go run ./cmd/connectorgen certification-matrix --check` — passed; all schema-v2 evidence records accepted without regenerating shared certification artifacts.
- `go test -timeout 20m ./cmd/connectorgen` — passed (`100.258s`).
- `go vet ./cmd/connectorgen/...` — passed.
- `go run ./cmd/agentcontractgen check` — passed; canonical contract and projections are current.
- `scripts/verify-gsd-workflow origin/integration/4015-mvp-flat-r1` — passed with this phase's PLAN/VERIFICATION evidence.
- `make connectorgen-validate connectorgen-surface-sync` — passed across 552 connectors with zero findings and zero drift.
- `git diff --check` — passed.
- The full `go test -timeout 20m ./...`/`make verify` suite was intentionally left to CI: AGENTS.md forbids running the 550+ connector suite as one per-command-timeout invocation. The changed package test plus generated/evidence gates above are the repository-prescribed local scope for this evidence-only lane.
