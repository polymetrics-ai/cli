# Plan — PM-only 957-case GitHub live lab

## Goal

Replace the preserved single-repository pre-skip boundary with a GitHub-only,
PM-driven lab framework. It must generate an exact 957-row plan from the frozen
current-head case ledger, refuse any fixture request outside run-owned immutable
targets, and convert cohorts to real provider attempts without leaking a secret or
weakening the existing reverse-ETL/confirmation contracts.

This is an incremental proof slice, not a claim that the 957 rows are already
proven. `LIVE-PROOF-CASES.json` and `LIVE-PROOF-REPORT.json` remain preserved
historical evidence until an increment runs and records its own terminal results.

## Locked safety contract

- Preserve branch `fm/cli-github-parity-extract-r1` at custody head
  `8ccd8dd6297914e9d2238e23c4cf6e5a880b6896`; do not reset, stash, discard, or
  create another worktree.
- Every fixture action is a built `pm github` invocation. `gh`, `gh-axi`, raw
  REST/GraphQL, browser controls, SDKs, and GitHub UI are prohibited for all lab
  setup, assertion, and cleanup. `gh-axi` remains delivery-PR-only.
- Before any provider write, validate a source-controlled boundary that defaults
  deny, rejects the `polymetrics-ai` owner and every worktree remote, requires
  exact owner/repository slugs plus immutable provider IDs, and rejects ambiguous
  resolution before a `pm` subprocess starts.
- The retained personal repository is read-only and mutation-ineligible until its
  proof-program provenance and a cleanup-ledger entry are demonstrated. If that
  cannot be shown, bootstrap a fresh private lab repository through the existing
  `pm github repo create` plan/preview/approval/execute contract. The preserved
  `repo view` control remains credential-pinned and has no target flags; retain
  its rejected owner/repo override as a sanitized regression record. Resolve a
  fresh bootstrap only through `pm github repos list-for-authenticated-user`,
  filtering the PM response in memory for one exact private generated slug under
  the authenticated immutable user, then bind that immutable repository ID.
- A write preserves plan → preview → approval → execute and any declared typed
  destructive acknowledgement. It must receive a separate PM read-back and an
  idempotent PM cleanup/neutralization record. Secret/dummy-secret values never
  enter arguments persisted to a file, reports, manifests, cleanup records, or
  commits.
- Organization deletion is last, ID-bound, typed-confirmed, and refuses to run
  until the cleanup ledger proves every referenced resource is run-owned.

## Manifest contract

Generate `GITHUB-LIVE-LAB-MANIFEST.json` from the preserved case ledger and the
current GitHub CLI surface. The JSON contains exactly 957 rows and a mechanically
derived class tally. Each row has:

1. stable command/case identity and current surface method/path/intent;
2. historical reason, target type, target-allowlist reference, credential class,
   plan/feature, and one of four factual cohorts;
3. PM-only setup, test, independent assertion, cleanup, and residual-state command
   templates (or an explicit `pm_surface_missing` prerequisite when GitHub exposes
   no PM-usable bootstrap path);
4. destructive acknowledgement requirement, cleanup behavior, and earliest
   divergence from the existing proven execution route; and
5. a named external prerequisite only where one is genuinely required.

The four exclusive cohorts are `personal_repo`, `sandbox_org_free`,
`github_app_or_marketplace`, and `unavailable_entitlement`. Their counts must sum
to 957. A plan/feature blocker remains a factual prerequisite, never a synthetic
`UNTESTABLE` result.

## TDD execution slices

### Slice 1 — source-derived manifest and default-deny boundary

**Red:** Add Node tests that fail when a generated manifest omits or duplicates a
historical pre-skipped command, produces any total other than 957, lacks a required
row field, uses a non-PM template, or classifies a row into zero/multiple cohorts.
Add fail-closed boundary tests for production-owner denial, working-repository
denial, missing immutable ID, slug/ID disagreement, ambiguous target resolution,
and a fake PM marker proving rejection happens before process launch.

**Green:** Add GitHub-only `scripts/github-live-lab.mjs` and a manifest generator
that export pure validation/classification functions. Write a zero-allowlist
boundary file and generate the 957-row manifest. The boundary validates before
credential inspection, plan creation, or provider dispatch.

### Slice 2 — append-only fixture and cleanup protocol

**Red:** Add tests for redaction, one terminal record per command, append-only
cleanup entries, intentional retention, idempotent cleanup replays, and typed
destructive acknowledgement propagation. The tests use synthetic non-secret IDs
and a fake PM only; they make no provider request.

**Green:** Implement PM invocation templates with transient process-only approval
values and an append-only cleanup ledger. The framework rejects non-PM programs,
unresolved/ambiguous targets, secret-shaped persisted data, and cleanup outside
the exact run-owned boundary.

### Slice 3 — personal-repository cohort increment

**Red:** Create a deterministic test for the earliest personal-cohort divergence:
the existing runner's global slug pin must not bypass the new ID-bound boundary or
skip the independent assertion/cleanup record.

**Green:** Build the current-head `pm`, reproduce the preserved credential-pinned
`repo view` control without invented owner/repo flags, and retain its malformed
targeting attempt as regression evidence. Only if the retained repository has
documented proof-program provenance may it enter the boundary. Otherwise bootstrap
a fresh private repository after Slice 1/2 tests pass, discover it with the
implemented PM authenticated-repository listing, and bind only one exact private
match under the authenticated immutable user ID. Do not rewrite stored credential
scope unless that PM discovery cannot prove the target. Exercise solely the rows
whose manifest cohort is `personal_repo`, one safe resource family at a time. Every
provider write uses the established PM plan/preview/approval/execute flow; each
result gains a new terminal record.

### Slice 3b — reversible repository-label family

This slice follows the external-bootstrap audit and is the next personal-repository
increment. It covers only the three historical rows `label create`, `label edit`,
and `label delete` on the already ID-bound private lab repository.

**Red:** Add a deterministic label resolver test that rejects an unscoped command,
an absent/duplicate generated label, and a malformed label-list envelope before any
write lifecycle. Strengthen the report/cleanup test to require a baseline PM list,
create read-back, edit read-back, typed-confirmed delete, and final absent read-back
for one exact generated label provider ID.

**Green:** Run PM `label list` with the immutable owner/repository config, fail if
the exact generated label name already exists, then use PM plan → preview → approval
→ execute for create and edit. Resolve the label only from a PM list response held
in memory, record its immutable ID in the append-only cleanup ledger, then use PM
typed-confirmed delete and a final PM list asserting absence. No label payload or
provider response body is persisted. Stop the family if preflight, scope binding,
or cleanup cannot be proved; do not substitute a provider tool.

### Slice 3c — editable issue plus comment lifecycle

This slice uses one fresh generated private-repository issue to cover the historical
`issue edit` and `issue comment` rows. It does not reuse the previously retained
closed issue, and it ends by PM-closing the new issue with the documented retention
rule because issue delete remains unsafe or disallowed.

**Red:** Add an in-memory PM issue-list resolver test that rejects a pre-existing,
absent, duplicate, or malformed generated issue. Require returned title/body/state
assertions across create and edit, a strictly increased returned comment count after
comment, and a final closed-state assertion before a terminal record is written.

**Green:** PM-list the exact bound repository baseline, then create one generated
issue through plan → preview → approval → execute. Resolve only its immutable ID
and issue number in memory; after an accepted write, retry only a successful
PM `issue list` envelope that has not yet reached the expected generated record,
with at most six PM-only attempts and five one-second visibility waits. Propagate
credential, entitlement, scope, or other PM read errors immediately rather than
masking them as eventual consistency. Edit through the same lifecycle and assert
returned title/body; add one PM comment and assert returned count increase; then
PM-close it and assert the closed state. The append-only ledger retains the closed
run-owned issue with the recorded issue-delete policy. No comment body or provider
response body is persisted.

### Slice 3d — disposable read-only deploy-key lifecycle

This slice covers the two historical personal-repository rows `repo deploy-key add`
and `repo deploy-key delete` without creating a production credential or retaining
a usable key. The runner generates one Ed25519 keypair in process memory, passes
only its public OpenSSH line to PM, discards the private half without writing it,
and records neither key material nor a provider payload.

**Red:** Add an in-memory PM deploy-key-list resolver test that rejects a wrong
command, absent/duplicate generated title, malformed/non-integer ID, or a returned
non-read-only record. Strengthen terminal/cleanup accounting to require exactly two
new proven rows plus one generated deploy-key fixture with create, immutable-ID
read-back, typed-confirmed deletion, and final PM-list absence.

**Green:** First PM-list the immutable-bound private repository and refuse a
pre-existing generated title. Generate an Ed25519 pair in process memory; run PM
plan → preview → approval → execute for `repo deploy-key add` with only the public
line and explicit `--read-only`. Resolve only the returned immutable ID/title/read-only
facts from a PM list held in memory. Then run typed-confirmed PM `repo deploy-key
delete` for that exact ID and require final PM-list absence before recording terminal
results. A PM read failure is terminal evidence, not a reason to use a provider
fallback or retry an unrelated operation.

### Slice 4 — external cohorts and escalation

**Red:** Tests require every non-personal cohort row to name a concrete smallest
external prerequisite and a least-privilege permission/plan list, never a generic
skip.

**Green:** Execute sandbox-org then App/Marketplace cohorts only after their actual
IDs are in the allowlist. Report only exact missing organization, App/installation,
or entitlement prerequisites with affected command counts. GitHub.com bootstrap
impossibilities (for example enterprise-only organization mutation or interactive
App registration) remain PM-surface findings; no fallback provider tool is used.

### Slice 4a — PM-only organization and App/Marketplace bootstrap probes

This sub-slice is ordered ahead of further personal-repository expansion. It is a
bootstrap audit, not permission to create a production or unbound provider resource.

**Red:** Add deterministic harness tests that reject an organization delete before
one run-owned immutable organization target and its cleanup provenance are present;
that refuse an account-level probe carrying repository selectors, a connection
override, or a write command; and that prove the only permitted account probes are
the exact PM direct reads named in the source-derived probe record. Add a source-map
test which fails if its command/path counts drift from the current GitHub
`cli_surface.json`/`api_surface.json`.

**Green:** Derive and record one compact, source-cited bootstrap-probe record before
any provider mutation. The organization audit must distinguish a runnable `orgs
delete` command from the absence or presence of a PM organization-create bootstrap
command; it may not invoke delete without an ID-bound, run-owned organization and a
complete cleanup ledger. The App audit must distinguish an executable PM command
that consumes a manifest conversion code from a PM command that can obtain that
code, and separately map App-authenticated versus Marketplace-user reads. Execute
only planned PM direct reads with the approved user credential, retain no response
body, and record only command, endpoint, HTTP/result class, and a sanitized
entitlement diagnosis. If the GitHub command surface cannot bootstrap an
organization/App/Marketplace fixture, record the precise missing command or
provider entitlement and stop that external cohort; never substitute UI, `gh`, a
raw API, or a production resource.

## Cohort recovery and rate limits

- Execute one cohort/resource family at a time; append a terminal record before
  advancing. Resume reads the ledger and refuses duplicate non-idempotent writes.
- Use the existing GitHub declared limiter through `pm`; record sanitized request
  counts, client stop/wait/refusal, and a PM `rate-limit get` read-back showing
  remaining provider headroom. The earlier curl instruction is superseded by the
  captain's PM-only constraint.
- On a provider rejection, retain `failed` with the sanitized provider diagnosis.
  On a genuine missing plan/permission/App prerequisite, retain the exact factual
  blocker and stop that cohort rather than relabeling it proven or pre-skipped.

## Required files and verification

Planned files are GitHub-specific harness/evidence only:

- `scripts/github-live-lab.mjs`, `scripts/github-live-lab-manifest.mjs`, and
  their `scripts/tests/` coverage;
- `scripts/github-live-cases.mjs` / `scripts/github-live-proof-sweep.mjs` only
  for boundary-compatible parameterization;
- `.planning/phases/github-parity-extract-r1/GITHUB-LIVE-LAB-{MANIFEST,BOUNDARY}.json`
  plus an append-only `GITHUB-LIVE-LAB-CLEANUP.jsonl` and the source-derived,
  response-body-free `GITHUB-LIVE-LAB-BOOTSTRAP-PROBES.json`.

Run before the first provider write:

```bash
node --test scripts/tests/github-live-lab.test.mjs
node --test scripts/tests/github-live-proof-sweep.test.mjs
node scripts/github-live-lab-manifest.mjs --check
node scripts/github-live-lab.mjs --check-boundary \
  --boundary .planning/phases/github-parity-extract-r1/GITHUB-LIVE-LAB-BOUNDARY.json
go build ./cmd/pm
```

Then run `pm help github`, `pm github <command> --help` for every newly invoked
command template, and scoped connector/CLI tests. No new no-mistakes run begins
until this framework and a coherent live-proof increment are committed.

## GSD / skills record

`scripts/gsd doctor`, command-source resolution, `discuss-phase --auto`, and
`plan-phase --tdd --skip-research` ran in this worktree. This is an inline/manual
GSD fallback because isolated GSD roles are not authorized here. Required skills
loaded for the GitHub connector/proof work: `golang-how-to`, `golang-cli`,
`golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`,
`golang-design-patterns`, `golang-structs-interfaces`, `golang-graphql`, and
`golang-documentation`. The framework adds no CLI surface, so help/manual/website
generation is not applicable; existing command help is verified before live use.

## Captain-approved continuation — deploy-key false-success safety gate (2026-08-09)

The captain approved the complete REST + GraphQL continuation, but provider activity is
paused behind this narrower first gate. The preserved lab has two typed-confirmed PM
`repo deploy-key delete` executions whose independent PM list assertion still found the
same generated immutable key. No third delete, new fixture write, or bulk cohort run may
begin until this gate is green.

### Diagnosis contract

Work locally and sequentially from the symptom:

1. Reproduce the plan → preview → typed confirmation → execute path against a local
   HTTP server that returns `404` for a correctly addressed deploy-key DELETE. Capture
   only method, route shape, and status in the test; no credential, key, repository, or
   provider response is retained.
2. Prove or disprove configuration/record mapping loss by asserting the persisted
   plan's owner, repository, and integer key ID reach the expected local request path.
   The plan stores `DestinationConfig` and the current CLI reuses it for `--plan`, so
   this is a concrete counterfactual rather than an assumption.
3. Compare that result with the proven label-delete control. Its path/record mapping
   must reach the same local server, while a `404` remains an error because it has no
   `missing_ok_status` exemption.
4. Inspect the originating declaration and history. The current GitHub
   `delete_deploy_key` declaration uniquely allows `missing_ok_status: [404]`; the
   generic engine counts a matching error as `RecordsWritten=1`. Separate this masking
   condition from the still-unknown provider reason for the returned `404`.

### Red/green decision

**Red:** an app-level regression must show the exact deploy-key DELETE route and then
fail because a visible local fixture plus a `404` cannot be reported as a completed
cleanup. This intentionally fails on the current declaration, establishing the same
false-success class without a provider call. A companion label control proves the
comparison is not a configuration, record-flag, or confirmation-lifecycle regression.

**Green:** make the smallest generated-source or runtime change that stops an attempted
GitHub deploy-key delete from converting a `404` into a successful write. Do not weaken
generic idempotent-delete behavior without evidence; regenerate every owned GitHub
artifact instead of hand-editing derived output. Keep the independent list-based absence
assertion as a second defense: successful transport alone never proves cleanup.

### Evidence-state rule

The preserved report must not leave this executed pair unaccounted for merely because cleanup did
not converge. Its add operation is proven by the created fixture and independent immutable-ID
read-back; its delete operation is `failed` until an independent PM list proves absence. The
append-only ledger records the two completed-but-not-absent executions as `cleanup_failed`, which is
explicitly nonterminal for the fixture. That distinction keeps the existing final-absence test
strict while allowing the boundary/read-back suite to be green on the truthful current state; it
does not authorize a third delete.

### Required local gates before provider activity resumes

```bash
go test -timeout 20m ./internal/app/ -run 'TestGitHubDeployKeyDelete'
go test -timeout 20m ./internal/connectors/engine/ -run 'TestWriteDelete'
node --test scripts/tests/github-live-lab.test.mjs
node --test scripts/tests/github-live-proof-sweep.test.mjs
node scripts/github-live-lab-manifest.mjs --check
node scripts/github-live-lab.mjs --check-boundary \
  --boundary .planning/phases/github-parity-extract-r1/GITHUB-LIVE-LAB-BOUNDARY.json
go run ./cmd/connectorgen validate internal/connectors/defs
go run ./cmd/connectorgen surface-sync --check
```

The inline/manual GSD fallback is intentional: this project contract prohibits role
spawning for the parent lane, while `scripts/gsd doctor`, all five command-source
resolutions, `agentcontractgen check`, and generated discuss/plan/execute/verify/review
prompts ran in this worktree. The captain's written continuation resolves the product
decision; the focused behavior change starts with the red test above. Skills loaded for
this slice: `golang-how-to`, `golang-troubleshooting`, `golang-cli`, `golang-testing`,
`golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`,
`golang-structs-interfaces`, and `golang-graphql`. No command/flag/help surface changes
in this safety gate, so CLI manual and website generation are not applicable.
