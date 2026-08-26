# Plan — issue 4325 declaration-admission foundation

## Task Delivery Header

- Issue: Refs #4325 — source-declaration admission certification.
- Base branch: main.
- Merges into: main.
- Delivery: Pull request #4351 is open against `main`; this expanded slice is
  committed, pushed to its existing branch, and locally verified.
- Working branch: fm/cli-declaration-admission-certification-r1.
- Task: Close the frozen R6 exact-SHA gaps by separating declaration mapping
  evidence from source-import retention metadata, validating effective
  config-carried declared values before App/credential construction with argv
  precedence, and migrating the independent admission inventory to schema v2.
- Verification: Focused red/green tests for absent and malformed retention
  metadata with strict mapping identity, unchanged strict source-import
  rejection, config enum/type/format/empty/byte-cap/range/cardinality plus argv
  precedence before credentials, inventory v2 acceptance/v1 rejection,
  formatting plus applicable generator/static/CI gates, and a fresh independent
  exact-head audit after push.

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Deferred source mapping has a CLI projection | fake | A hermetic cited bundle has a deferred delete command that resolves to a typed missing-foundation error. A provider call cannot establish this schema invariant. |
| Policy-only suppression is rejected | fake | A hermetic declaration bundle mutates only the foundation component to a policy value; it is rejected before any provider request. A provider call cannot establish this schema invariant. |
| Existing implemented delete remains runnable | live | GitHub's cited `label delete` declaration admits and the no-credential binary dispatch reaches `missing --credential`. |
| Deferred command cannot hide behind a policy or excluded endpoint | fake | A hermetic bundle tries to declare a deferred command against an excluded/policy endpoint. The shared runtime resolver refuses it before it can become `missing_foundation`; no provider call can establish this local structural invariant. |
| Deferred foundation is machine-verifiable and round-trippable | fake | A hermetic source cohort, declaration catalog, and command surface carry the same typed component, evidence, exact source/binding identity, destructive semantic, and canonical target. Tests mutate each independently and observe an admission/preflight failure. |
| Citation follows the existing safe source-publication policy | fake | A hermetic source row using HTTP, userinfo, a private literal, a fragment, or a credential-shaped query is rejected by the shared local citation validator. |
| Equivalent citation spellings cannot evade provider-operation uniqueness | fake | Hermetic authoring and production-ledger rows vary DNS host case, default HTTPS port, escaped path, and query order; noncanonical authored forms fail closed and the canonical identity still detects one provider operation claimed under different bindings. Provider I/O cannot prove this local identity invariant. |
| This foundation does not claim shipped Outreach commands | live | The committed PR delta leaves `internal/connectors/defs/outreach/**` unchanged and the corrected test/docs describe only the generic real-bundle resolver seam. Final merge validation is explicitly gated on a real combined-head Outreach mapping/pilot with committed discovery commands, source evidence, credential-boundary proof, zero transport calls, and a fresh audit after #4350 repair. |
| Mapping admission is independent of retained bytes and hashes | fake | A hermetic connector-owned source lock retains the exact URL/location/protocol/provider-ID/method/path row while omitting or corrupting only retention byte/hash fields; declaration admission accepts it, while the unchanged full source-import parser rejects it. Provider I/O cannot prove a repository schema boundary. |
| Config-carried declared values fail before App/credentials | live | Public no-credential connector invocations submit invalid enum/type/cardinality config values and observe validation errors rather than `missing --credential`; an explicit valid argv flag overriding an invalid config value reaches the credential boundary. |
| The independent selection denominator is schema v2 only | fake | Hermetic inventory validation accepts v2 and rejects legacy v1 before cross-catalog admission. A provider call cannot establish local schema compatibility. |

## Scope boundary

This is a shared tooling/schema PR. It does not convert a connector, refresh
a provider artifact, call a provider, add credentials, or perform a
write/delete. It must not weaken `commandrunner.Preflight`, `surface-sync`,
source-lock verification, runtime certification, or live certification.

Captain clarification 007 supersedes the prior Stripe conversion direction.
The uncommitted mapping work at
`internal/connectors/defs/stripe/cli_surface.json` and
`internal/connectors/defs/stripe/sources/stripe-declaration-admission.json`
is intentionally preserved for `cli-batch1-repair-r1`; it must not be staged
or committed by this PR.

## Audit repair context (2026-08-25)

Captain inbox 008 carries an independent audit of frozen commit `3d39cc1fc`.
DA-001 through DA-007 were confirmed, but DA-008 found that a deferred command
could bypass exact API-surface/executor mapping through an excluded or policy
row, and DA-009 found that admission citations used a weaker URL policy than
source import. This repair remains a generic foundation change: it adds no
provider mapping and leaves the preserved Stripe files untouched.

The normal GSD adapter path was resolved with `scripts/gsd doctor`,
`scripts/gsd sources`, and generated `discuss-phase`, `plan-phase --gaps`,
`execute-phase --gaps-only`, `verify-work`, and `code-review` prompts. This
non-Pi single-worker session cannot run the Pi role runtime, and the canonical
contract forbids role spawning here, so the discuss/plan/execute/verify/review
work is recorded and performed inline. Required skills loaded: `golang-how-to`,
`golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`,
`golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, and
`golang-documentation`, plus `golang-lint` for final static verification.

## Audit repair waves (TDD)

1. **Red — independent denominator and command identity:** Extend the
   hermetic catalog tests to reject duplicate exact provider operation
   identities, noncanonical/non-round-trippable command paths, and weak or
   secret-shaped citation URLs. The source row itself remains the independent
   completeness denominator; no source lock, imported bytes, or hash is read.
2. **Red — shared deferred resolver:** Add adversarial bundles whose deferred
   target is excluded, policy-only, mismatched, or multi-endpoint. The current
   resolver returns `missing_foundation` before inspecting any target, so these
   tests must fail first.
3. **Green — typed target and foundation evidence:** Carry a typed foundation
   component/evidence and exactly one target through the declaration catalog,
   CLI schema, engine projection, and commandrunner. The engine's no-network
   deferred preflight validates its target against a blocked ledger operation;
   `connectorgen declaration-admission` calls the real `commandrunner.Preflight`
   with canonical command segments and accepts only typed
   `system/missing_foundation` for a valid deferred row.
4. **Green — semantic destructive metadata:** Require DELETE rows to declare
   `kind=delete`; require any other exact target classified as
   `destructive_action` to carry destructive metadata; reject a non-destructive
   target falsely labelled delete; and require an implemented row's exact
   write/operation binding to retain the declared destructive semantics.
   Deferred metadata remains a declaration, not a runtime claim.
5. **Refactor/document/verify:** Reuse the source-import publication URL
   validator for admission citations, document all three certificates, run
   the focused red/green tests and repository gates, then conduct the inline
   code review and update the PR body.

## TDD execution slices

1. **Red — admission contract:** Add focused table-driven tests for a cited
   runnable read; deferred reverse-ETL write/delete; deferred binary
   download/upload; importer/descriptor gap; missing/duplicate/stale/base-path
   mismatch; false implementation; and an all-deferred zero-runnable bundle.
   The tests should fail because `declaration-admission` and its schema/type do
   not exist.
2. **Green — shared declaration checker:** Add the required, versioned
   `declaration_admission_sources.json` independent cohort and separate
   `declaration_admissions.json` catalog plus a
   `connectorgen declaration-admission [defs-dir] [--json]` command. It checks
   a nonzero expected source denominator, deterministically cross-links source identity, lane,
   canonical endpoint, command, destructive/delete metadata, and runtime
   state. It never fetches provider data or requires source artifact bytes,
   hashes, request schemas, or fixtures.
3. **Green — explicit deferred command state:** Extend the command surface’s
   shared deferred/foundation metadata only as needed so an admitted deferred
   command stays discoverable and `commandrunner` returns a typed
   missing-foundation refusal before any executor. Keep implemented preflight
   rules unchanged.
4. **Refactor/document/gate:** Add the Make target and concise certification
   design/canon documentation distinguishing declaration admission from
   runtime and live certificates. Run formatting, targeted tests, relevant
   generator/check targets, review, and full feasible local verification.
5. **Red/green — enforceable deferred mapping:** Require every deferred
   declaration to name an evidenced missing implementation component rather
   than a method, risk, confirmation, approval, blocked-by-default, source
   retention/hash, or live-certification policy. Prove a hermetic deferred
   delete command returns `system/missing_foundation`, while GitHub's `label
   delete` remains implemented and reaches the no-credential boundary.

## Exact-SHA re-audit repair (2026-08-26)

The independent Codex audit of `683a3c76e` reopened DA-001 through DA-006,
DA-010, and DA-011. The repair uses two required root catalogs rather than
optional connector sidecars; binds source, declaration, command, and runtime by
stable binding identity rather than method/path alone; derives destructive
semantics from the independent source row; and moves implemented binding
resolution into one engine function shared with runtime preflight. Deferred
targets now include source/binding identity and use the compact source cohort
embedded in `defs.FS`, so production preflight does not depend on the omitted
`api_surface.json`. Typed `missing_foundation` survives oversized metadata,
the App plan path preflights before credential resolution, and the public CLI
preserves its machine-readable code. Structural endpoint checks apply to new
source/deferred identities without retroactively changing legacy implemented
provider-specific command paths.

## Exact-SHA R2 gap closure (2026-08-26)

The independent Codex R2 audit of `c99e40a315b20b776a3b8653b54fc682a8469844`
found four remaining High blockers. This gap wave stays within the shared
admission/runtime seam and preserves every uncommitted Stripe and Docker Hub
path byte-for-byte.

1. **Red — provenance-only source uniqueness (DA-002):** duplicate one exact
   provider citation under different local source IDs and different runtime
   bindings. Both the authoring check and compact production ledger must reject
   it before declaration or runtime evaluation. The source key comprises only
   canonical source URL, exact document location, protocol, raw provider
   operation ID, method, and canonical provider endpoint identity; binding
   uniqueness remains a separate invariant.
2. **Red — fail-closed canonical endpoint equivalence (DA-003):** reject
   same-binding wrong-method/wrong-path aliases for stream, write, REST, and
   binary bindings. Add a locked real-bundle census for the 243 non-GraphQL
   aliases and an end-to-end GitHub `discussion list` admission case. The green
   resolver may prove only named transformations: declared base-URL path
   composition, absolute-transport normalization, query separation, positional
   placeholder equivalence, the closed operation-variant annotation grammar,
   a registered hook transport route, and GraphQL operation identity to
   `POST /graphql` transport. It must retain canonical and transport endpoints
   separately.
3. **Red — reject runnable commands relabelled deferred (DA-004):** exercise
   the actual implemented `commandrunner` preflight against a synthesized
   implemented form of each deferred command. Every foundation component must
   fail for the runnable GitHub label delete and GraphQL read/write controls.
   Remove the idempotency component because no runtime executor requires it,
   and make response-descriptor evidence REST/binary-specific.
4. **Red — exact production ledger inventory (DA-012):** require
   `declaration_admission_sources.json` to have one exact root-artifact class,
   byte attribution, and deterministic inventory while continuing to reject
   the declaration catalog, API surfaces, fixtures, and non-exempt source
   locks.
5. **Green/refactor/gate:** implement the smallest shared changes, correct the
   stale Makefile wording, update the certificate-separation docs and this
   ledger, run focused packages plus every applicable generated/static gate in
   a clean committed checkout, then execute `verify-work` and `code-review`
   inline. The generated GSD gap prompts are executed inline because the
   canonical single-worker contract and launch brief forbid role spawning.
6. **Real-bundle resolver compatibility (captain steer 014, claim corrected by
   R3):** load the real Outreach stream/write shapes and synthesize only the
   absent discovery projection in memory. This proves generic admission,
   canonical/transport resolution, and no-I/O commandrunner preflight only. It
   is not shipped CLI, source-evidence, credential-boundary, or zero-transport
   proof. The real combined-head Outreach mapping/pilot remains a final merge
   gate after #4350 repair. No provider request or connector definition change
   is permitted in this PR.

Required skills for this gap wave: `golang-how-to`, `golang-cli`,
`golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`,
`golang-design-patterns`, `golang-structs-interfaces`, `golang-graphql`,
`golang-documentation`, and `golang-lint`.

Closure at R2: all six steps completed on clean code SHA `f97dede07`. The fleet
census proved 243 non-GraphQL and 4 GraphQL aliases; the Outreach test proved
only synthetic-discovery real-bundle resolver compatibility, as corrected by
R3. All 477 connectorgen tests/examples and every directly integrated Go
package passed, and the whole-tree boundary scan reported 0 findings without a
new exception. `VERIFICATION.md` and `REVIEW.md` contain the exact commands and
R2 disposition plus the R3 correction.

## Exact-SHA R3 gap closure (2026-08-26)

The independent R3 audit of `9c779a14218d00857c587da6f3499d6b9546c445`
found one valid High implementation gap and one invalid foundation-level
claim. This single-worker gap wave uses the generated GSD
`plan-phase --gaps` / `execute-phase --gaps-only` prompts inline; role spawning
is forbidden by the canonical contract and launch brief.

1. **Red — canonical provider citation identity (DA-002):** add authoring and
   production-ledger cases for uppercase DNS hosts, explicit HTTPS `:443`,
   unstable query ordering, non-normalized escaped paths, and one canonical
   provider operation claimed under different bindings. The current raw URL
   key must fail these tests.
2. **Green — one fail-closed shared canonicalizer:** add the shared URL seam at
   `internal/safety/provider_citation.go`. It enforces public absolute HTTPS,
   rejects userinfo/fragments/private literals/ambiguous or trailing-dot DNS
   hosts/credential-shaped queries, lowercases DNS hosts, removes `:443`,
   normalizes escaped paths, and stably encodes bounded single-valued query
   keys. Both `cmd/connectorgen/declarationadmission.go` and
   `internal/connectors/engine/declaration_target_ledger.go` require the stored
   URL to equal the canonical result and key uniqueness by that result.
3. **Schema seam:**
   `internal/connectors/engine/schema/declaration_admission_sources.schema.json`
   already represents the citation as a bounded `source_url` string. Schema
   version 1 remains unchanged; canonicality is a semantic validator invariant,
   not a rewritten value or a new connector-owned field.
4. **Outreach claim correction:** rename/reword the in-memory real-bundle test
   as generic resolver compatibility only. It is not actual Outreach CLI,
   credential-boundary, or zero-transport proof. Do not add or synthesize a
   committed Outreach discovery surface in this foundation PR. Record the
   required combined-head Outreach mapping/pilot gate after #4350 repair.
5. **Refactor/gate/review:** preserve the 243+4 endpoint-equivalence census,
   six lanes, no-hash/no-provider-I/O admission rule, and all Stripe/Docker Hub
   handoffs. Run focused packages and the applicable generated/static CI gates,
   update verification/review/PR evidence, push #4351, then request a fresh
   independent exact-head audit. Do not merge.

Required skills loaded for this wave: `golang-how-to`, `golang-testing`,
`golang-error-handling`, `golang-security`, `golang-safety`,
`golang-design-patterns`, `golang-structs-interfaces`, and
`golang-documentation`, plus the `gsd-ns-workflow` routing skill.

## Exact-SHA R5 gap closure (2026-08-26)

The independent R5 audit of `51367fc7e97705046ce6274125e9cee1d6e1e365`
found four remaining High blockers. Inbox 019 supersedes the prior audit wait
and authorizes a generic-only repair. The generated `discuss-phase --auto` and
`plan-phase --gaps` prompts are executed inline because the canonical
single-worker contract and launch brief forbid role spawning. Required skills
loaded for this wave: `golang-how-to`, `golang-cli`, `golang-testing`,
`golang-error-handling`, `golang-security`, `golang-safety`, `golang-lint`,
`golang-design-patterns`, `golang-structs-interfaces`,
`golang-documentation`, and `gsd-ns-workflow`.

1. **Red — reviewed source-lock identity (DA-001/DA-002):** mutate a source
   row to an unrelated lock, nonexistent lock operation, alias location,
   operation ID, method, or endpoint and require admission to fail without
   provider I/O. The source catalog names an exact operation in a
   connector-owned reviewed source lock; the checker compares the row's URL,
   location, protocol, provider operation ID, and endpoint byte-for-byte to
   that selected inventory record. It reads no provider and does not make
   retained bytes, artifact hashing, or live certification an admission step.
2. **Red — independent denominator (DA-012):** create a separately controlled
   declaration-admission inventory that selects connector/source-lock operation
   identities. Delete a source row and declaration while attempting to adjust
   legacy adjacent counts; certification must still fail because the inventory
   entry remains. Expected-count fields are removed from the mutable catalogs.
3. **Red — provider-evidenced unsupported disposition:** cover ETL, reverse
   ETL, direct read, direct write, binary download, and binary upload with an
   exact source-backed command projection whose availability is
   `unsupported_with_provider_evidence`. It remains denominator-visible and
   discoverable, claims no executor or missing foundation, and returns a typed
   unsupported refusal distinct from `missing_foundation`.
4. **Red — input-aware credential-free public validation:** prove `--help`
   returns first; bare GitHub `label delete` reports missing `--name`; adding
   `--name bug` reaches missing credential; and unknown, enum, minimum, and
   direct secret/env-only carrier defects all fail before `withApp` credential
   resolution. The public CLI and runtime execution share one command flag
   validator rather than restating accepted input rules.
5. **Green/refactor/gate:** migrate only the shared schemas and root catalogs,
   reuse the existing connector source-lock parser offline, add the unsupported
   projection/refusal, invoke input preflight before credential resolution,
   update certificate-separation docs, and run focused plus repository gates in
   a clean exact-head checkout. Do not modify, stage, or normalize any Stripe,
   Docker Hub, Outreach, or other connector-owned production definition.

### R5 code and schema seam

- `cmd/connectorgen/declarationadmission.go` consumes a new independently
  controlled `internal/connectors/defs/declaration_admission_inventory.json`,
  resolves each selected operation through the existing source-lock parser,
  and cross-links the compact source and declaration catalogs.
- `internal/connectors/engine/schema/declaration_admission_*.schema.json` and
  `declaration_admission_inventory.schema.json` express the lock reference,
  count-free catalogs, and provider-evidenced unsupported state; the compact
  production ledger retains admitted targets but does not perform provider I/O
  or re-certify retained artifacts.
- `internal/connectors/commandrunner` exposes one request-input preflight that
  shares runtime flag validation, and `internal/cli/runMaybeConnectorCommand`
  invokes it after help parsing but before `withApp` and credential lookup.

## CLI docs parity

`connectorgen declaration-admission` is an internal generator command, not a
new `pm` command. The applicable docs are its `connectorgen` usage and the
connector certification/design canon. `pm help`, bare namespace behavior,
`docs/cli/**`, website pages, manual generation, and shell completion are not
applicable. Deferred connector commands retain normal `cli_surface.json`
discovery and are covered by commandrunner tests.

## Commit checkpoints

- Plan/TDD evidence checkpoint before production changes.
- Red-test checkpoint when the repository’s test convention permits it.
- Green implementation/documentation checkpoint after targeted gates.
- Review-fix checkpoint if inspection finds an actionable defect.

## Exact-SHA R6 gap plan (2026-08-26)

The independent R6 audit of exact head
`ab2c5e3933e0dc1355948d3585b269c46f75754d` is the frozen repair ledger. This
wave uses `scripts/gsd prompt plan-phase issue-4325-declaration-admission-r1
--gaps` and `execute-phase ... --gaps-only` inline because the canonical
single-worker contract and launch brief forbid role spawning. Required skills:
`golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`,
`golang-security`, `golang-safety`, `golang-design-patterns`,
`golang-structs-interfaces`, `golang-documentation`, and `golang-lint`, plus
`gsd-ns-workflow` and `gsd-ns-review` for the inline lifecycle and review.

1. **Red — mapping evidence is not retention evidence (DA-R6-001):** mutate
   only `bytes`/`sha256` in the connector-owned reviewed source lock. Admission
   must accept an absent or malformed retention pair while retaining strict
   version, connector ownership, URL, location, protocol, provider operation
   ID, method/path, counts, and duplicate checks. The ordinary source-import
   parser must continue to reject the same mutations.
2. **Green — admission-only source-lock reader:** add one narrow reader beside
   the source-import lock types and call it only from
   `declarationAdmissionReviewedSourceFindings`. It performs strict JSON and
   mapping-inventory validation but never validates or reads retained bytes,
   hashes, capture metadata, or certificates. Do not weaken
   `parseSourceImportLock`.
3. **Red — effective config validation before credentials (DA-R6-002):** add
   commandrunner tests for enum, boolean/integer, date-time, empty, encoded-byte
   cap, minimum/maximum, and string-array cardinality on `config.*` mappings;
   add public no-credential Freshchat/config regressions; and prove valid argv
   overrides an invalid config value.
4. **Green — one shared configured-value validator:** replace the numeric-only
   helper with coercion through the declaration-owned command flag validator.
   Validate only effective config values, skip a config target when an explicit
   argv value wins, and return redacted config-key/flag-context errors before
   App construction. Runtime override validation consumes the same helper.
5. **Red/green — inventory schema v2 (DA-R6-003):** update the inventory
   metaschema, committed root inventory, authoring constant/check, and hermetic
   fixtures to v2; add an exact legacy-v1 rejection. No compatibility fallback
   or connector-specific migration is permitted.
6. **Refactor/document/gate:** keep declaration admission, source-import
   retention, runtime preflight, and live certification distinct in the canon;
   run focused red/green packages, full changed packages, applicable generated
   and snapshot checks, then `verify-work` and `code-review` inline. Preserve
   every unstaged Stripe/Docker Hub handoff path byte-for-byte and do not run a
   provider request, credentialed check, write, delete, or broad connector
   regeneration.

## R9 exact-head re-audit repair plan (2026-08-26)

`REVIEW-CONVERGENCE.md` freezes the independent R8 audit at
`92b2c495f45fbc5d011fcd40cdf4ab51178ddc39`. This is a coordinated shared
foundation wave, not connector mapping work. The inline/manual GSD fallback
uses the already-resolved `discuss-phase` → `plan-phase --tdd` →
`execute-phase` → `verify-work` → `code-review` sequence because the direct-PR
contract forbids lifecycle-role spawning.

1. **Red F1 (App):** demonstrate that an invalid effective command request
   reaches `missing --credential` before input validation. The direct App
   regression must then require the validation error and zero vault/credential
   reads for requiredness plus declared flag/config coercion.
2. **Red F2 (CLI plan):** demonstrate that unknown argv and malformed config
   on an existing command plus `--plan` reach plan lookup. The CLI and
   built-binary tests must require request-validation errors before `withApp`,
   state lookup, credential resolution, or preview; valid argv overrides still
   win over an invalid config value.
3. **Red F3 (production ledger):** load a schema-valid compact ledger whose
   object repeats `source_url`. It must fail at the real production loader, not
   solely in a `connectorgen` duplicate reader.
4. **Green shared boundaries:** make `PlanConnectorCommand` and the CLI
   `--plan` path use `commandrunner.PreflightRequest` with the same effective
   request semantics. Add duplicate-object-member detection before production
   schema/decode. Do not alter source locks, mapping-only admission, runtime
   executor selection, or connector declarations.
5. **Verify and re-audit:** run focused packages and the listed serial gates;
   build the binary to prove invalid-before-credential and valid GitHub
   label-delete-to-missing-credential behavior; review the final diff inline,
   push only a normal fast-forward to PR #4351, verify API base/head, and leave
   the new exact SHA open for a fresh independent Codex re-audit.
