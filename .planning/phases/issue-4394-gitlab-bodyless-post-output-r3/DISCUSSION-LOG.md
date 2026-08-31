# GitLab R3 source-output policy discussion

## Fixed decision

The reviewed GitLab R2 candidate correctly identified eight exact retained,
bodyless semantic `POST` reads. Its materialization did not preserve the
provider response shape for four of them: it declared `json_redacted` even
though the retained source records a status response with no response media
type. The correction is deliberately narrow:

- The four status-only source operations must use `output_policy: "none"` in
  both the operation and generated CLI declaration.
- The four Conan lookup source operations retain `output_policy:
  "json_redacted"`, because their retained source records
  `application/json` success media.
- The source lock, descriptor, source IDs, matrix rows/counts, enabled
  contract IDs, legacy write views, and runtime execution code are not changed
  by this R3 correction.

## Source-backed cohort

### Status-only → `none`

1. `gitlab.rest.postApiV4AiThirdPartyAgentsDirectAccess` — `POST /api/v4/ai/third_party_agents/direct_access`, success `201`, no response media.
2. `gitlab.rest.postApiV4CodeSuggestionsConnectionDetails` — `POST /api/v4/code_suggestions/connection_details`, success `201`, no response media.
3. `gitlab.rest.postApiV4GeoNodeProxyIdGraphql` — `POST /api/v4/geo/node_proxy/{id}/graphql`, success `200`, no response media.
4. `gitlab.rest.postApiV4IntegrationsSlackOptions` — `POST /api/v4/integrations/slack/options`, success `201`, no response media.

### Documented JSON → `json_redacted`

1. `gitlab.rest.postApiV4PackagesConanV1ConansPackageNamePackageVersionPackageUsernamePackageChannelPackagesConanPackageReferenceUploadUrls`.
2. `gitlab.rest.postApiV4PackagesConanV1ConansPackageNamePackageVersionPackageUsernamePackageChannelUploadUrls`.
3. `gitlab.rest.postApiV4ProjectsIdPackagesConanV1ConansPackageNamePackageVersionPackageUsernamePackageChannelPackagesConanPackageReferenceUploadUrls`.
4. `gitlab.rest.postApiV4ProjectsIdPackagesConanV1ConansPackageNamePackageVersionPackageUsernamePackageChannelUploadUrls`.

All eight are exact source-bound, no-request-body `POST` direct reads. The
distinction is response policy, not method, source identity, or lane.

## Foundation Atlas result

R2 introduces the reusable direct-read form
`runtime.source-bound-bodyless-post-read.v1`. Its existing engine symbols
already execute `output_policy: "none"` and reject a nonempty status-only
response. R3 only tightens source-to-declaration conformance and corrects the
consumer evidence. This is a **reuse**, not a new generic foundation,
connector-specific runtime hook, or engine change.

## Explicit non-goals

- No credentialed GitLab request or provider write.
- No runtime executor, commandrunner, receiver, source-importer, certification,
  source-lock, descriptor, matrix-ID/count, or enabled-contract-ID change.
- No generic raw HTTP, raw response, or raw JSON capability.
- No source row is suppressed or promoted beyond its existing R2 disposition.

## GSD execution mode

The adapter was checked with `scripts/gsd doctor`, its registered sources were
resolved, and `go run ./cmd/agentcontractgen check` passed. The canonical
`gsd-discuss-phase`, `gsd-plan-phase --tdd`, `gsd-execute-phase`,
`gsd-verify-work`, and `gsd-code-review` prompts were rendered. This runtime
cannot launch the Pi/GSD role agents without violating the active single-worker
restriction, so the prompts are executed inline and their red/green evidence
is recorded in this phase directory.
