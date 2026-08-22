# Foundation 0.3.0 release candidate r1 evidence

source_sha: 298a519de14b98cef5e1d8e20cd43b6d0ab42d84

**Implementation SHA I2:** `298a519de14b98cef5e1d8e20cd43b6d0ab42d84`.

The release candidate begins at exact canonical core `041d2ec7ed986aea15d2d3d64f2076b484c3f999`, preserves reverse action binding through merge `50e90fa854635b7c8b295b7090034b82a52e4e03`, and preserves public output hardening through merge `e18f4372f4f65ae5e42265f237abad79473a7425`. `input-manifest.json` is the immutable typed provenance record.

The 38 original blocker categories are historical: both authoritative intake audits found committed repairs, and this RC does not reimplement them. The implementation includes the prior `a5005fa…` test/source-lock alignment, the confirmed Recurly repair (five accidentally removed required path declarations restored through `operations.json` and `surface-sync`), the confirmed GitHub source-projection repair, and the confirmed RC-09 CodeQL repair. RC-09 removes unchecked allocation multiplication and two dead assignments without suppressing or baselining any alert.

`FND-B09` was reproduced before this closure: the inherited schema-2 manifest fails strict decode because `component_inputs` is an object, not the required typed list. The prior closures are superseded after CI directly proved generator-owned website data stale, the current branch reproduced the GitHub reachability regression, and CodeQL reproduced the three RC-09 alerts. This schema-3 manifest binds I2, its five typed component identities, the current certification subject fingerprint, and closure-artifact SHA-256 digests. The strict evidence gate is run only from the clean evidence commit because it verifies the commit-parent graph.

At I, the exact release-surface command reports identical results for `v0.2.1` and HEAD: `endpoints=1225`, `blocked=1`, and `direct_read_candidates=120`. The sole blocked identity is the same retired route, `POST /user/{user_id}/projectsV2/{project_number}/drafts`, whose operation remains duplicate-of `POST /graphql (github.graphql.mutation.add-project-v2-draft-issue)`.

No credential value was requested, read, printed, or stored. No approved non-secret credential reference was supplied to this isolated lane, so no new provider call was made. Recurly has local fixture proof only and is explicitly implemented-but-uncertified. GitHub has local behavioral proof only: a fresh binary executed all 97 generated candidates and all 633 released API-surface targets with emitted-request and declared-output assertions. RC-09 retains the 633-command fresh-binary proof (193.93s), cold full `cmd/connectorgen` (143.231s), and `go vet ./...`; fresh CodeQL remains an external gate. Local hermetic behavior proof, historical provider observations, and certification are intentionally separate in `RELEASE-CANDIDATE-R1-REVIEW.md`; the current certification matrix reports zero certified connector shards.
