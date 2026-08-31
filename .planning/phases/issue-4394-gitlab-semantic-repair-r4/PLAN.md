# GitLab #4394 semantic POST mapping repair — plan

## Task Delivery Header

- Issue: Refs #4394 — GitLab complete connector-definition artifact projection.
- Base branch: `fm/cli-top100-declaration-batch-r1` at immutable commit `687eb1ded6b42cc456f8cc3c1e97f0a84fd042a8`.
- Merges into: candidate `codex/4394-gitlab-semantic-repair-r4` → `fm/cli-top100-declaration-batch-r1` (human-gated integration; never `main`).
- Delivery: a committed, normally pushed candidate with the listed focused no-provider-I/O checks green and an independent-review-ready scope report.
- Working branch: `codex/4394-gitlab-semantic-repair-r4`.
- Task: retain all eight source-semantic GitLab POST-read cells as `mapped_unproven`, correct the four Conan cells to state their retained JSON-body-without-schema gap exactly, and add source-lock-backed regression evidence that prevents a bodyless/raw-body or executable promotion.
- Verification: exact-base `connectorgen validate --connector gitlab`; GitLab matrix test; connector-shaped source projection/contract checks; JSON parse; agent-contract check; diff check; no provider I/O or credentials.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| The four Conan lookup rows remain retained, direct-read `mapped_unproven`, and non-executable | live | The matrix test reads the immutable lock prose, verifies each source ID and exact M-U disposition, and rejects an implemented/bodyless/raw-body-shaped drift. |
| The four status-only semantic POST reads remain source-consistent without a false direct-read JSON artifact | live | The matrix test checks their empty retained success-media set and verifies no implemented direct-read operation or CLI command claims one of those source IDs; `output_policy: none` is therefore not misrepresented as an implemented direct-read command. Existing legacy method-partition writes are not rewritten by this mapping-only slice. |
| Existing source IDs, lane counts, locks, descriptor, Atlas, and runtime remain unchanged | live | Matrix reconciliation, byte/path diff checks, and the exact-base validation compare actual tracked artifacts; no provider fixture is required because this is source-mapping evidence only. |

## GSD lifecycle and inline fallback

- Resolved with `scripts/gsd doctor` and `scripts/gsd sources` for `discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and `code-review`.
- Rendered the five corresponding `scripts/gsd prompt` commands for `issue-4394-gitlab-semantic-repair-r4`.
- Inline/manual fallback: this isolated issue slice is not a registered GSD roadmap phase, and the task contract prohibits spawning additional runtime agents. This plan, TDD ledger, and verification record execute the rendered prompts manually without weakening red/green/refactor or review requirements.

## Foundation Atlas check

| Need | Atlas lookup | Classification | Decision |
| --- | --- | --- | --- |
| Preserve source-backed direct-read mapping and exact source IDs | `source.projection-admission.v1` | reuse | Use the matrix's immutable source-lock backlink; do not alter locks/descriptors. |
| Execute a direct-read POST with a typed JSON body | `runtime.direct-execution.v1` | actual unsupported contract for this row | The foundation explicitly limits structured JSON flags to action-backed direct writes. Conan prose requires a JSON object but retains neither requestBody media nor a typed schema, so no typed contract can be produced and no new foundation is authorized. Keep the four rows M-U. |
| Execute a bodyless semantic POST read | no available selected foundation on this base | not pursued | The task forbids a new runtime foundation. Do not emit `rest.no_request_body`, a zero-byte body, or a raw-body escape. |

## Bounded cohort and decisions

The frozen source lock remains authoritative at
`internal/connectors/defs/gitlab/sources/gitlab-operation-source-lock.json`.
The matrix continues to retain 13 semantic POST direct-read rows and all seven
lane cells for every source row. This repair changes no source IDs or lane
counts.

1. Four Conan lookup IDs remain `mapped_unproven` because their retained
   descriptions require a JSON object containing per-file `name` and `size`,
   while the locked operation carries neither `requestBody` nor request media
   nor a closed object schema:
   - `gitlab.rest.postApiV4PackagesConanV1ConansPackageNamePackageVersionPackageUsernamePackageChannelPackagesConanPackageReferenceUploadUrls`
   - `gitlab.rest.postApiV4PackagesConanV1ConansPackageNamePackageVersionPackageUsernamePackageChannelUploadUrls`
   - `gitlab.rest.postApiV4ProjectsIdPackagesConanV1ConansPackageNamePackageVersionPackageUsernamePackageChannelPackagesConanPackageReferenceUploadUrls`
   - `gitlab.rest.postApiV4ProjectsIdPackagesConanV1ConansPackageNamePackageVersionPackageUsernamePackageChannelUploadUrls`
2. Four status-only source-semantic POST reads retain their source mapping,
   no declared successful response media, and no executable **direct-read**
   connector artifact:
   `postApiV4AiThirdPartyAgentsDirectAccess`,
   `postApiV4CodeSuggestionsConnectionDetails`,
   `postApiV4GeoNodeProxyIdGraphql`, and
   `postApiV4IntegrationsSlackOptions`.
   A future projected command would need `output_policy: none` only if a
   separately source-backed, executable request form exists; this change does
   not create that command or make any output-policy execution claim. Existing
   legacy method-partition write/reverse declarations, if any, remain untouched
   and do not change the source-semantic direct-read lane.
3. The remaining semantic POST rows retain their existing mapping facts and
   dispositions. No direct-write/reverse-ETL mutation reclassification, ETL,
   sync, binary, or declaration artifact is changed.

## Red → Green → Refactor plan

1. **Red:** add a source-lock-backed matrix test that fails on the base's
   generic Conan reason, and test mutated data for an `implemented`, bodyless,
   or raw-body-shaped Conan promotion. Add the status no-media/no-artifact
   assertion as a negative regression.
2. **Green:** change only the four matrix reason values through the existing
   deterministic expected-lane builder. The source text must state the missing
   typed JSON contract precisely; it must not call a bodyless request safe.
3. **Refactor:** keep the special case derived from the retained description,
   not an HTTP method or free-standing operation-ID allow-list; rerun full
   GitLab matrix reconciliation to prove every other row and count is intact.

## Scope and exclusions

- Allowed: `internal/connectors/defs/gitlab/sources/gitlab-source-lane-matrix.json`, its focused Go test, and this planning/verification evidence.
- Excluded: source lock, descriptor, crosswalk, runtime/engine/CLI behavior, connector artifacts, Atlas, transport, certification, credentials, network/provider I/O, generic raw or JSON request tools, and all other connectors.
- CLI help/manual/website parity: not applicable. No executable command, flag, help behavior, output contract, generated manual, or website surface changes.

## Commit and delivery checkpoints

1. Plan/TDD/verification evidence + red test recorded.
2. One green mapping-only commit after focused validations.
3. Fresh remote base check; normal candidate-only push and remote SHA verification.
4. Stop for independent review; do not integrate, open a PR, or merge.
