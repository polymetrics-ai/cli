# VERIFICATION — issue #4368 zero-operation source-reference foundation

## Status

Planning complete; implementation and RED evidence pending.

## Acceptance checklist

- [ ] Explicit zero-operation rendered coverage validates only with retained bytes, source lock, manifest provenance and a closed marker.
- [ ] Missing, malformed, unverified, accidental-empty, duplicate and mixed-document variants fail closed with location.
- [ ] Non-empty rendered-reference and OpenAPI v1-v3 source-lock/import behavior has unchanged byte/count checks.
- [ ] Exactly 720 source-cited deferred rows reconcile as 187 Amplitude + 49 Dremio + 193 Ashby + 84 Workable + 207 HiBob.
- [ ] Rows preserve exact source citation, provider operation ID, stable identity, applicable lane, and exactly one `missing_foundation` disposition.
- [ ] Deferred command preflight stops before credential/transport/record/mutation; retained runnable command boundary still reaches missing credential.
- [ ] Generator/projection/evidence/surface-sync/validate and bounded JSON duplicate invariants pass.
- [ ] Formatting, vet, build, diff, individual repository gates, rebase, exact-head review, and PR API base verification pass.

## CLI documentation parity

Not applicable to user-facing help/manual/website: this changes the internal `connectorgen` locked-source import contract. The final verification nevertheless builds `pm` and runs relevant help checks to prove no accidental public CLI regression.
