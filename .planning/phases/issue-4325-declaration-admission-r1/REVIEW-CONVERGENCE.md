# Review convergence — issue 4325 declaration-admission foundation, R9

## Task Delivery Header

- Issue: Refs #4325 — declaration-admission certification foundation; repair existing PR #4351.
- Base branch: `main` at `b33983927d863032dac8220949990506e812937d`.
- Merges into: `fm/cli-declaration-admission-certification-r1` → `main` (human-gated).
- Delivery: PR #4351 remains open on its existing branch with all frozen R8 findings repaired, behaviorally verified, non-force pushed, and ready for a fresh independent Codex re-audit.
- Working branch: local `repair/4351-declaration-admission-r9`, tracking `origin/fm/cli-declaration-admission-certification-r1`; every publish is an explicit normal fast-forward to that existing PR branch.
- Task: Repair input-validation ordering at the App and CLI `--plan` boundaries, and reject duplicate JSON members in the production declaration-target ledger without coupling admission to retained source bytes or certificates.
- Verification: exact red/green regressions in `internal/app`, `internal/cli`, and `internal/connectors/engine`; focused affected packages; formatter, lint, build, generator/admission/surface-sync/runtime-preflight/canon checks; built-binary ordering probes; remote ref and PR-base read-back.

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Invalid App command input cannot cross the credential/vault boundary | live | A direct `PlanConnectorCommand` call with missing or malformed declared input returns the request-validation refusal while its vault spy observes zero reads. |
| Invalid CLI `--plan` input cannot open App state or resolve credentials | live | CLI and built-binary invocations with unknown argv and malformed config return request-validation errors rather than `reverse plan ... not found` or `missing --credential`. |
| A duplicate production-ledger member is refused | live | The real `loadDeclarationTargetLedgers` rejects a schema-valid compact JSON fixture with repeated `source_url`; a last-member-wins decoder would accept it. |
| Valid GitHub label delete remains executable through the closed path | live | Built `pm github label delete --name bug` in an initialized no-credential project stops at `missing --credential`, after valid request preflight. |
| Admission remains retention-independent and all lanes/states remain intact | fake | This repair changes neither source lock bytes/capture/certificates nor connector definitions; existing admission and six-lane/runtime-preflight gates prove the unchanged repository contract without provider I/O. |

## Freeze

- Frozen independent audit: `/Users/karthiksivadas/karthik-agent-workspace/data/cli-declaration-admission-reaudit-codex-r8/report.md`.
- Audit SHA: `92b2c495f45fbc5d011fcd40cdf4ab51178ddc39`.
- Base SHA: `b33983927d863032dac8220949990506e812937d`.
- Remote verification before branch transition: `origin/main` and `origin/fm/cli-declaration-admission-certification-r1` resolved exactly to those SHAs; the latter is a descendant of the former.
- Discovery set is frozen to F1–F3 below. Green CI at the audit SHA is evidence only; it does not close these behavioral failures.
- The current worktree was clean at base and contains no preserved Stripe/Docker Hub handoff. The existing separate worktree that owns the literal PR branch was not touched; Git's single-worktree branch constraint is handled by the local tracking branch named above.

## Frozen finding ledger

| ID | Severity | Frozen reachability / violated invariant | Repair and regression |
| --- | --- | --- | --- |
| F1 | High | `(*App).PlanConnectorCommand` uses path-only `commandrunner.Preflight`, then `ResolveConnectorCredential`/vault before `BuildWriteCommand` validates effective flags and config. Invalid untrusted input must stop before credential, provider, or protected application state. | Call the same `PreflightRequest` request contract before credential resolution. Add a direct App regression that observes validation error and zero vault reads. |
| F2 | High | CLI `--plan` special-cases path-only preflight before `withApp`; unknown argv/config defects can reach plan/state lookup and preview/credential work. The plan continuation must use the identical request preflight before any App construction. | Resolve environment-only values, build effective declared flags/config, and invoke `PreflightRequest` for `--plan`. Add CLI and built-binary unknown/malformed-config ordering regressions. |
| F3 | Medium | `loadDeclarationTargetLedgers` validates/schema-decodes with standard JSON decoding, accepting repeated object names as last-member-wins. The compact production ledger is an authorization input and must be unambiguous itself. | Add duplicate-aware JSON-object validation to the production loader before schema/decode. Add a direct production-loader repeated-member regression; generator-only checking is insufficient. |

## Cross-lens convergence

| Lens | Required invariant | R9 decision |
| --- | --- | --- |
| Architecture / ordering | Untrusted request input stops before App state, credential/vault, or provider work. | One declaration-owned `PreflightRequest` is the boundary at both App and CLI plan entry points. |
| CLI/App parity | `--plan` is not a validation bypass and config/argv precedence stays shared. | Preserve help-first behavior and the existing environment-only resolution; validate effective values before App. |
| Ledger integrity | Production authorization JSON cannot be ambiguous. | Reject duplicate object members before `ValidateDeclarationAdmissionSources` and `strictDecode`; do not rely on the generator reader. |
| Admission separation | Mapping needs identity, not byte/hash/capture/certificate proof. | Do not alter source locks, retained artifacts, certificate metadata, or the admission mapping reader. |
| Reachability | Six lanes and `implemented`, `missing_foundation`, and `unsupported_with_provider_evidence` stay visible. | Do not relabel commands or edit connector definitions; run existing runtime/admission/canon proof after the shared fix. |
| Security | Values and credentials must not be exposed or read prematurely. | Tests assert boundary behavior without real credentials or provider I/O; errors retain existing redaction rules. |

## Coordinated TDD wave

1. **Red F1:** direct App request with invalid declared input must currently return `missing --credential`; the regression asserts instead for request validation and zero vault access.
2. **Red F2:** CLI and built `pm ... --plan` with `--bogus` and invalid declared `--config` must currently reach `reverse plan ... not found`; regressions require typed request-validation errors before App state.
3. **Red F3:** direct compact-ledger fixture with duplicate `source_url` must currently load; regression requires a duplicate-member refusal.
4. **Green:** route App and CLI plan execution through shared `PreflightRequest`, preserving explicit argv-over-config precedence and help behavior; install a bounded duplicate-aware validation at the production loader.
5. **Refactor / verify:** keep mapping-only and source-lock certificate paths untouched; run all listed package, generator, runtime-preflight, canon, and binary witnesses; perform inline `verify-work` then deep code review; request a fresh independent exact-head audit.

## GSD and skills

The installed adapter was checked with `scripts/gsd doctor`; the generated
`discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and
`code-review` prompts were resolved. The canonical direct-PR/single-worker
contract forbids lifecycle-role spawning, so these stages are executed inline
and recorded here. Required skills loaded: `golang-how-to`, `golang-cli`,
`golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`,
`golang-design-patterns`, `golang-structs-interfaces`, and `golang-lint`.

## CLI help/manual/website parity

This changes only validation order for existing connector command arguments;
it adds no command, flag, output shape, help topic, manual text, generated
reference, or website content. Runtime `--help`, bare namespace behavior, and
existing command output must remain unchanged; the applicable CLI parity proof
is therefore help and existing command behavior, with docs/website changes
explicitly not applicable.

## Final review gate

The repair necessarily produces a new SHA. It is not merge-ready until the
new exact SHA has local evidence, required CI, and a fresh independent Codex
re-audit. No provider request, credential use, write/delete operation, source
lock rewrite, connector-definition edit, force-push, or merge is authorized.
