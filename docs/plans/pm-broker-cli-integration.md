# PM Broker CLI Integration Parent

Status: integration-parent seed for CLI parent issue [#563](https://github.com/polymetrics-ai/cli/issues/563).
Branch: `integration/pm-broker-production-program` from `origin/main`.

This file is the minimal parent-branch tracking artifact for the PM Broker production program. It records the issue map, merge rules, GSD/TDD evidence, and live-deployment blockers for parallel CLI sub-issue implementation.

## GSD, TDD, and skills evidence

- GSD command path used before this planning artifact: `scripts/gsd doctor`, `scripts/gsd list`, and `scripts/gsd prompt plan-phase pm-broker-cli-parent-integration --skip-research`.
- Required routing read: `.agents/agentic-delivery/references/required-skills-routing.md`, `.agents/agentic-delivery/references/gsd-pi-adapter.md`, parent-orchestrator workflows, automated-review routing, and no-mistakes guidance.
- Skills loaded for this docs-only parent setup: `gsd-core`, `no-mistakes`, `golang-how-to`, `golang-cli`, `golang-documentation`, `golang-security`, `golang-testing`.
- TDD ledger: docs-only coordination seed; no Go behavior changed. Red-test gate is not applicable. Validation is source/issue mapping plus no-mistakes for the parent PR.
- Verification checklist: `git diff --check`, committed branch diff review, parent PR checks, and no-mistakes CI-ready outcome.

## Parent and cross-repository links

- CLI parent: [polymetrics-ai/cli#563](https://github.com/polymetrics-ai/cli/issues/563)
- PM Broker parent: [polymetrics-ai/pm-broker#1](https://github.com/polymetrics-ai/pm-broker/issues/1)
- PM Broker architecture-unblocked marker: [pm-broker#25](https://github.com/polymetrics-ai/pm-broker/issues/25)
- Exact single-VPS deployment-input decision: [pm-broker#32](https://github.com/polymetrics-ai/pm-broker/issues/32)
- Future public authentication registry: [pm-broker#31](https://github.com/polymetrics-ai/pm-broker/issues/31), [cli#592](https://github.com/polymetrics-ai/cli/issues/592)

## Integration rules

Final merge from this parent PR to the repository default branch is captain-only.

Validated sub-issue PRs may be merged autonomously only into `integration/pm-broker-production-program`, never into `main`, and only when all approved gates hold:

1. Accepted sub-issue scope is complete.
2. Relevant local unit, integration, race, security, and contract tests pass.
3. The current-code-matched no-mistakes run completed review, fixes, tests, documentation, push, PR, and CI.
4. All required available PR checks are green.
5. No unresolved security-sensitive, destructive, irreversible, credential, privacy, or production-resource decision remains.
6. API/schema assumptions match the pinned parent contract and dependencies are satisfied.
7. The guarded Firstmate PR merge helper and approved merge method are used.
8. No force-push, reset, discard, hand-splice, or history rewrite is used.

After each integration merge, run combined parent-branch tests and update the parent PR evidence. If integration fails, revert through an ordinary reviewed change on the integration branch; do not rewrite shared history.

## Live-deployment blockers

Exact single-VPS pilot inputs in [pm-broker#32](https://github.com/polymetrics-ai/pm-broker/issues/32) block only live deployment and host-specific rollout. They do not block code implementation, tests, fake fixtures, docs, packaging, deployment guardrail scaffolding, or CLI UX that refuses live rollout until inputs are approved.

Still blocked before live rollout: exact VPS/host, network/ingress posture, domain/TLS ownership, workload-identity trust anchor, metadata design, backup destination, audit destination, maintenance window, and recovery operator. Do not assume a host name, create production resources, create credentials, create GCP project/billing, configure host-specific deployment, or merge to `main`.

## CLI dependency map

| CLI issue | Scope | Primary PM Broker counterpart(s) | Dependency / parallelism note |
|---|---|---|---|
| [#564](https://github.com/polymetrics-ai/cli/issues/564) | ADR and command schema | [#3](https://github.com/polymetrics-ai/pm-broker/issues/3), [#4](https://github.com/polymetrics-ai/pm-broker/issues/4), [#8](https://github.com/polymetrics-ai/pm-broker/issues/8), [#24](https://github.com/polymetrics-ai/pm-broker/issues/24) | Foundation lane; publish command/schema contract early. |
| [#565](https://github.com/polymetrics-ai/cli/issues/565) | OIDC auth commands | [#12](https://github.com/polymetrics-ai/pm-broker/issues/12), [#17](https://github.com/polymetrics-ai/pm-broker/issues/17), [#18](https://github.com/polymetrics-ai/pm-broker/issues/18), [#22](https://github.com/polymetrics-ai/pm-broker/issues/22) | May proceed against fake/loopback auth once API fixtures are pinned. |
| [#566](https://github.com/polymetrics-ai/cli/issues/566) | Organization/Workspace/Environment context | [#8](https://github.com/polymetrics-ai/pm-broker/issues/8), [#10](https://github.com/polymetrics-ai/pm-broker/issues/10), [#12](https://github.com/polymetrics-ai/pm-broker/issues/12), [#24](https://github.com/polymetrics-ai/pm-broker/issues/24) | Parallel with auth/profile work after safe metadata contract is pinned. |
| [#567](https://github.com/polymetrics-ai/cli/issues/567) | Setup storage profiles and downgrade refusal | [#9](https://github.com/polymetrics-ai/pm-broker/issues/9), [#11](https://github.com/polymetrics-ai/pm-broker/issues/11), [#20](https://github.com/polymetrics-ai/pm-broker/issues/20), [#21](https://github.com/polymetrics-ai/pm-broker/issues/21), [#26](https://github.com/polymetrics-ai/pm-broker/issues/26) | Must not silently fall back or grant writes. |
| [#568](https://github.com/polymetrics-ai/cli/issues/568) | Local OS/local-encrypted storage UX | [#11](https://github.com/polymetrics-ai/pm-broker/issues/11), [#21](https://github.com/polymetrics-ai/pm-broker/issues/21), [#22](https://github.com/polymetrics-ai/pm-broker/issues/22), [#26](https://github.com/polymetrics-ai/pm-broker/issues/26) | Independent local-storage lane; no plaintext export. |
| [#569](https://github.com/polymetrics-ai/cli/issues/569) | Compatibility adapters | [#8](https://github.com/polymetrics-ai/pm-broker/issues/8), [#15](https://github.com/polymetrics-ai/pm-broker/issues/15), [#23](https://github.com/polymetrics-ai/pm-broker/issues/23), [#24](https://github.com/polymetrics-ai/pm-broker/issues/24), [#26](https://github.com/polymetrics-ai/pm-broker/issues/26) | Depends on migration and typed-auth assumptions; can build adapters against pinned fake fixtures. |
| [#570](https://github.com/polymetrics-ai/cli/issues/570) | Versioned config and migration ledger | [#11](https://github.com/polymetrics-ai/pm-broker/issues/11), [#15](https://github.com/polymetrics-ai/pm-broker/issues/15), [#21](https://github.com/polymetrics-ai/pm-broker/issues/21), [#22](https://github.com/polymetrics-ai/pm-broker/issues/22), [#26](https://github.com/polymetrics-ai/pm-broker/issues/26) | Migration path must be resumable and reversible until verification completes. |
| [#571](https://github.com/polymetrics-ai/cli/issues/571) | Opaque refs and typed execution | [#5](https://github.com/polymetrics-ai/pm-broker/issues/5), [#7](https://github.com/polymetrics-ai/pm-broker/issues/7), [#8](https://github.com/polymetrics-ai/pm-broker/issues/8), [#21](https://github.com/polymetrics-ai/pm-broker/issues/21), [#23](https://github.com/polymetrics-ai/pm-broker/issues/23), [#24](https://github.com/polymetrics-ai/pm-broker/issues/24), [#27](https://github.com/polymetrics-ai/pm-broker/issues/27) | Critical contract consumer; wait for pinned OpenAPI/domain fixtures before broad fanout. |
| [#572](https://github.com/polymetrics-ai/cli/issues/572) | Connection routing and ExecutionPlan UX | [#7](https://github.com/polymetrics-ai/pm-broker/issues/7), [#8](https://github.com/polymetrics-ai/pm-broker/issues/8), [#19](https://github.com/polymetrics-ai/pm-broker/issues/19), [#20](https://github.com/polymetrics-ai/pm-broker/issues/20), [#23](https://github.com/polymetrics-ai/pm-broker/issues/23), [#24](https://github.com/polymetrics-ai/pm-broker/issues/24) | Depends on policy/grant/approval-mode contracts. |
| [#573](https://github.com/polymetrics-ai/cli/issues/573) | Policy approval and audit UX | [#7](https://github.com/polymetrics-ai/pm-broker/issues/7), [#18](https://github.com/polymetrics-ai/pm-broker/issues/18), [#19](https://github.com/polymetrics-ai/pm-broker/issues/19), [#20](https://github.com/polymetrics-ai/pm-broker/issues/20), [#22](https://github.com/polymetrics-ai/pm-broker/issues/22) | Can proceed with deterministic denial/approval fixtures. |
| [#574](https://github.com/polymetrics-ai/cli/issues/574) | Read-only Shepherd/GSD validator | [#13](https://github.com/polymetrics-ai/pm-broker/issues/13) | Independent safety/orchestration lane; no merge/history rewrite authority. |
| [#575](https://github.com/polymetrics-ai/cli/issues/575) | CLI certification tests | [#4](https://github.com/polymetrics-ai/pm-broker/issues/4), [#5](https://github.com/polymetrics-ai/pm-broker/issues/5), [#8](https://github.com/polymetrics-ai/pm-broker/issues/8), [#16](https://github.com/polymetrics-ai/pm-broker/issues/16), [#21](https://github.com/polymetrics-ai/pm-broker/issues/21), [#23](https://github.com/polymetrics-ai/pm-broker/issues/23), [#24](https://github.com/polymetrics-ai/pm-broker/issues/24) | Runs alongside implementation; final matrix depends on all CLI slices under test. |
| [#576](https://github.com/polymetrics-ai/cli/issues/576) | Docs/release rollout | [#25](https://github.com/polymetrics-ai/pm-broker/issues/25), [#31](https://github.com/polymetrics-ai/pm-broker/issues/31), [#32](https://github.com/polymetrics-ai/pm-broker/issues/32) | Docs can proceed after command schema; rollout wording must preserve future/deployment blockers. |
| [#577](https://github.com/polymetrics-ai/cli/issues/577) | Cross-repo dependency tracker | [#5](https://github.com/polymetrics-ai/pm-broker/issues/5), [#6](https://github.com/polymetrics-ai/pm-broker/issues/6), [#7](https://github.com/polymetrics-ai/pm-broker/issues/7), [#8](https://github.com/polymetrics-ai/pm-broker/issues/8), [#12](https://github.com/polymetrics-ai/pm-broker/issues/12), [#15](https://github.com/polymetrics-ai/pm-broker/issues/15), [#16](https://github.com/polymetrics-ai/pm-broker/issues/16), [#19](https://github.com/polymetrics-ai/pm-broker/issues/19), [#20](https://github.com/polymetrics-ai/pm-broker/issues/20), [#21](https://github.com/polymetrics-ai/pm-broker/issues/21), [#22](https://github.com/polymetrics-ai/pm-broker/issues/22), [#23](https://github.com/polymetrics-ai/pm-broker/issues/23) | This artifact surfaces the initial tracker map for worker routing. |
| [#578](https://github.com/polymetrics-ai/cli/issues/578) | Membership RBAC metadata UX | [#10](https://github.com/polymetrics-ai/pm-broker/issues/10), [#12](https://github.com/polymetrics-ai/pm-broker/issues/12), [#17](https://github.com/polymetrics-ai/pm-broker/issues/17), [#18](https://github.com/polymetrics-ai/pm-broker/issues/18), [#22](https://github.com/polymetrics-ai/pm-broker/issues/22) | Safe metadata only; no admin bypass. |
| [#579](https://github.com/polymetrics-ai/cli/issues/579) | Standing write grant UX | [#7](https://github.com/polymetrics-ai/pm-broker/issues/7), [#18](https://github.com/polymetrics-ai/pm-broker/issues/18), [#19](https://github.com/polymetrics-ai/pm-broker/issues/19), [#20](https://github.com/polymetrics-ai/pm-broker/issues/20), [#22](https://github.com/polymetrics-ai/pm-broker/issues/22) | Requires human-authorized grant lifecycle; no self-approval. |
| [#580](https://github.com/polymetrics-ai/cli/issues/580) | Agent write approval modes | [#7](https://github.com/polymetrics-ai/pm-broker/issues/7), [#19](https://github.com/polymetrics-ai/pm-broker/issues/19), [#20](https://github.com/polymetrics-ai/pm-broker/issues/20), [#21](https://github.com/polymetrics-ai/pm-broker/issues/21), [#22](https://github.com/polymetrics-ai/pm-broker/issues/22) | Connection creation and write authority stay separate. |
| [#581](https://github.com/polymetrics-ai/cli/issues/581) | No plaintext secret export | [#6](https://github.com/polymetrics-ai/pm-broker/issues/6), [#11](https://github.com/polymetrics-ai/pm-broker/issues/11), [#15](https://github.com/polymetrics-ai/pm-broker/issues/15), [#21](https://github.com/polymetrics-ai/pm-broker/issues/21), [#22](https://github.com/polymetrics-ai/pm-broker/issues/22), [#26](https://github.com/polymetrics-ai/pm-broker/issues/26) | Hard security invariant across migration/export UX. |
| [#582](https://github.com/polymetrics-ai/cli/issues/582) | Audit retention UX | [#7](https://github.com/polymetrics-ai/pm-broker/issues/7), [#18](https://github.com/polymetrics-ai/pm-broker/issues/18), [#19](https://github.com/polymetrics-ai/pm-broker/issues/19), [#20](https://github.com/polymetrics-ai/pm-broker/issues/20), [#22](https://github.com/polymetrics-ai/pm-broker/issues/22) | Remote 365-day/local 30-day redacted evidence UX. |
| [#583](https://github.com/polymetrics-ai/cli/issues/583) | GCP provider CLI contract | [#5](https://github.com/polymetrics-ai/pm-broker/issues/5), [#6](https://github.com/polymetrics-ai/pm-broker/issues/6), [#21](https://github.com/polymetrics-ai/pm-broker/issues/21), [#22](https://github.com/polymetrics-ai/pm-broker/issues/22), [#27](https://github.com/polymetrics-ai/pm-broker/issues/27), [#28](https://github.com/polymetrics-ai/pm-broker/issues/28) | Use fake/synthetic data by default; live non-production GCP evidence remains protected. |
| [#584](https://github.com/polymetrics-ai/cli/issues/584) | Generic typed auth compatibility | [#5](https://github.com/polymetrics-ai/pm-broker/issues/5), [#6](https://github.com/polymetrics-ai/pm-broker/issues/6), [#8](https://github.com/polymetrics-ai/pm-broker/issues/8), [#12](https://github.com/polymetrics-ai/pm-broker/issues/12), [#16](https://github.com/polymetrics-ai/pm-broker/issues/16), [#21](https://github.com/polymetrics-ai/pm-broker/issues/21), [#22](https://github.com/polymetrics-ai/pm-broker/issues/22), [#23](https://github.com/polymetrics-ai/pm-broker/issues/23), [#24](https://github.com/polymetrics-ai/pm-broker/issues/24) | Built-in/internal typed auth only; public registry is future work. |
| [#585](https://github.com/polymetrics-ai/cli/issues/585) | OpenAPI /v1 HTTP/JSON client and fake fixtures | [#8](https://github.com/polymetrics-ai/pm-broker/issues/8), [#24](https://github.com/polymetrics-ai/pm-broker/issues/24) | Contract-owner lane; pins the typed client/transport contract, accepted fixtures, and fake broker for downstream CLI work. |
| [#586](https://github.com/polymetrics-ai/cli/issues/586) | Architecture unblocked marker | [#25](https://github.com/polymetrics-ai/pm-broker/issues/25), [#31](https://github.com/polymetrics-ai/pm-broker/issues/31), [#32](https://github.com/polymetrics-ai/pm-broker/issues/32) | Implementation unblocked; public registry and exact deployment inputs are non-code blockers. |
| [#587](https://github.com/polymetrics-ai/cli/issues/587) | Legacy vault retirement UX | [#11](https://github.com/polymetrics-ai/pm-broker/issues/11), [#15](https://github.com/polymetrics-ai/pm-broker/issues/15), [#21](https://github.com/polymetrics-ai/pm-broker/issues/21), [#22](https://github.com/polymetrics-ai/pm-broker/issues/22), [#26](https://github.com/polymetrics-ai/pm-broker/issues/26) | Two-stage human-controlled migration; no plaintext export. |
| [#588](https://github.com/polymetrics-ai/cli/issues/588) | Provider SDK boundary in CLI | [#5](https://github.com/polymetrics-ai/pm-broker/issues/5), [#6](https://github.com/polymetrics-ai/pm-broker/issues/6), [#27](https://github.com/polymetrics-ai/pm-broker/issues/27) | No vendor leakage, runtime plugins, or auto-merge of provider APIs into CLI contracts. |
| [#589](https://github.com/polymetrics-ai/cli/issues/589) | Provider certification fixtures | [#5](https://github.com/polymetrics-ai/pm-broker/issues/5), [#6](https://github.com/polymetrics-ai/pm-broker/issues/6), [#22](https://github.com/polymetrics-ai/pm-broker/issues/22), [#27](https://github.com/polymetrics-ai/pm-broker/issues/27), [#28](https://github.com/polymetrics-ai/pm-broker/issues/28) | Synthetic every PR; sanitized non-production GCP evidence only when separately protected. |
| [#590](https://github.com/polymetrics-ai/cli/issues/590) | Remote outage UX | [#9](https://github.com/polymetrics-ai/pm-broker/issues/9), [#22](https://github.com/polymetrics-ai/pm-broker/issues/22), [#24](https://github.com/polymetrics-ai/pm-broker/issues/24), [#29](https://github.com/polymetrics-ai/pm-broker/issues/29) | No silent fallback; idempotent retry and reconciliation evidence. |
| [#591](https://github.com/polymetrics-ai/cli/issues/591) | Single-VPS pilot CLI UX | [#25](https://github.com/polymetrics-ai/pm-broker/issues/25), [#30](https://github.com/polymetrics-ai/pm-broker/issues/30), [#32](https://github.com/polymetrics-ai/pm-broker/issues/32) | Limited-availability UX may proceed; #32 blocks only live rollout. |
| [#592](https://github.com/polymetrics-ai/cli/issues/592) | Future auth contract registry | [#31](https://github.com/polymetrics-ai/pm-broker/issues/31) | Future-only; do not block initial production-pilot implementation. |

## Parallel implementation feed

- Start contract-foundation lanes first: #564, #577, #585, and #586, coordinated with PM Broker #3/#4/#8/#24/#25.
- With #585 publishing typed `/v1` fixtures plus the endpoint-safe HTTP client foundation, auth/context/storage/policy/audit UX lanes can proceed in parallel when write scopes do not collide: #565, #566, #567, #568, #573, #578, #579, #580, #581, #582, #590, #591.
- Migration/compatibility lanes #569, #570, and #587 depend on PM Broker #15/#26 and no-raw-export #21.
- Provider/GCP lanes #583, #588, and #589 depend on PM Broker #5/#6/#27/#28 and must stay synthetic unless live non-production evidence is separately approved.
- #575 and #576 run continuously against accepted slices, then become final certification/docs gates after implementation lands.

## CLI-side contract fixture package

The first CLI-side synthetic contract foundation for #585 lives in `internal/pmbroker/contract/v1`.
Later profile, context, and execution CLI lanes should consume `NewSyntheticBroker().NewClient()`,
`NewHTTPClient()`, and `AcceptedSyntheticFixtures()` for deterministic tests of `/v1` compatibility
negotiation, opaque connector references, and typed execution-plan requests. The exact endpoint,
auth, correlation, pagination, and response-bounds contract is owned by
`internal/pmbroker/contract/v1` package documentation and code. Tests stay network-free through the
fake broker transport, while production-bound HTTP remains a typed endpoint-safe foundation with
explicit auth and correlation seams. The package does not expose arbitrary request methods,
caller-supplied headers, generic JSON/body execution, provider SDKs, raw-secret retrieval/export,
public gRPC, divergent socket semantics, SQL, shell, or runtime plugins. `auth_registry_mode`
remains pinned to `internal_experimental`; this package does not claim a stable public
authentication registry.
