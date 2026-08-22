# Foundation 0.3.0 release candidate r1 evidence

**Implementation SHA I:** `368f21f4e874a262cf2b4fc8d13186474dfd2159`.

The release candidate begins at exact canonical core `041d2ec7ed986aea15d2d3d64f2076b484c3f999`, preserves reverse action binding through merge `50e90fa854635b7c8b295b7090034b82a52e4e03`, and preserves public output hardening through merge `e18f4372f4f65ae5e42265f237abad79473a7425`. `input-manifest.json` is the immutable typed provenance record.

The 38 original blocker categories are historical: both authoritative intake audits found committed repairs, and this RC does not reimplement them. The implementation includes the prior `a5005fa…` test/source-lock alignment plus the confirmed Recurly repair: five accidentally removed required path declarations are restored through `operations.json` and `surface-sync`. This changes no runtime authority and makes no provider or certification claim.

`FND-B09` was reproduced before this closure: the inherited schema-2 manifest fails strict decode because `component_inputs` is an object, not the required typed list. The prior closure is superseded after CI directly proved generator-owned website data stale; this schema-3 manifest binds the exact regenerated website implementation, its five typed component identities, the current certification subject fingerprint, and closure-artifact SHA-256 digests. The strict evidence gate is run only from the clean evidence commit because it verifies the commit-parent graph.

No credential value was requested, read, printed, or stored. No approved non-secret credential reference was supplied to this isolated lane, so no new provider call was made. Recurly has local fixture proof only and is explicitly implemented-but-uncertified. Local hermetic behavior proof, historical provider observations, and certification are intentionally separate in `RELEASE-CANDIDATE-R1-REVIEW.md`; the current certification matrix reports zero certified connector shards.
