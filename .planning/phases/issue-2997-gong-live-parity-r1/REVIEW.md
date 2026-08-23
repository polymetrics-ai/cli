# Code review: Gong release-0.3.0 parity reconciliation

## Scope reviewed

- Current official-source lock and the 69-row Gong disposition/API/CLI mapping.
- The 27 exact reverse-ETL write bindings, including the three multipart actions.
- Output-preservation declarations and generated manual/skill/website projections.
- Batch 2/3 gap-ledger cleanup after the merged generic multipart foundation.

## Findings and dispositions

1. **Fixed — stale partial write labels.** Twenty-four actions remained `partial` after the
   merged generic structured-body and reverse-action foundation. Each is now an implemented,
   exact named write binding; the focused test requires a one-to-one implemented CLI/write mapping.
2. **Fixed — mismatched declaration field shapes.** Six array/object fields and the CRM
   `selected_fields` array were not representable by their old string/no-flag CLI declarations.
   They are now schema-bound JSON flags and pass runtime command preflight.
3. **Fixed — stale F4 bookkeeping.** The Batch 2/3 generator fabricated Gong multipart gap rows
   after focused generic conformance passed. The obsolete Gong-specific generator branch was
   removed; the regenerated 19-connector check contains zero Gong gap rows.
4. **Open, shared foundation — provider-value collision masking.** Gong read declarations have no
   field-redaction policy and now require ordinary provider values to remain exact even when a
   value equals configured credential material. The current shared output boundary violates that
   rule by collision-masking values, headers, raw receipts, and cursors. Foundation issue #4321
   owns the generic red/green correction; the Gong PR contains no provider-specific exception.
5. **Open, shared foundation — source-import common-input preflight.** #4335 is merged and the
   v3 `gong-v2` declaration now makes the importer fetch/parse the fixed official query-bearing
   artifact. It then rejects `GET /v2/all-permission-profiles` parameter 0 for a missing
   `maxLength` before any descriptor can be projected. The generic importer must retain or emit a
   typed gap for such common-bound provider inputs; no Gong maximum or connector exception was
   added. See `SOURCE-AUDIT.md` for the exact failure.
6. **Open — source-projected required input.** The current source says `GET /v2/targets` requires
   `workspaceId`, while the generated direct command has not yet received that marker. The missing
   canonical descriptor prevents `surface-sync` from deriving it. The live HTTP-400 classification
   is recorded without payload data; no hand-authored flag or Gong-named runtime branch is allowed.
7. **Open — full live parity.** Scoped persisted-App authentication, bounded ETL, typed direct read,
   input validation, pagination validation, and external proof pass. Five ETL append cells are
   still uncertified; there are no self-cleaning Gong write pairings; and the two paid agentic
   source rows were deliberately not called. No customer payload, identifier, credential, or
   browser-session evidence was used.
8. **Fixed — unsafe direct-read certification candidate.** The first declaration selected
   `targets list`, which needs a provider-required workspace ID and cannot be safely inferred by
   the harness. The red test now requires the sole live candidate to be the bounded, ordinary
   typed `users extensive` command. It passed against the actual disposable account; no agentic
   endpoint is in the candidate set.
9. **Fixed — multipart request parameters missing from operation declarations.** The official
   parameter importer added the exact path/query declarations for call media, CRM entity schema,
   and target assignments. Its post-generation check is clean. The review verified that these are
   connector-owned JSON declarations and use no Gong-named runtime branch.
10. **Verified — evidence boundaries.** The committed evidence contains counts, classifications,
    source hashes, and one external-proof hash only. The local proof, credential store, responses,
    and account data stay outside the repository. `git diff --check`, focused conformance,
    commandrunner preflight, generated-candidate/sweep/subject checks, lint, boundary, and release
   workflow gates pass; the repository-wide test and `make verify` remain blocked by unrelated
   generated-skill drift.
11. **Verified — captain collision-policy declaration coverage.** The strengthened focused test
    first found three direct-read descriptions that omitted the policy. All 30 direct reads,
    metadata, generated manual/skill, and catalog now state that an ordinary provider value stays
    intact when it equals configured credential material. The review traced the contrary runtime
    behavior to shared output projections and recorded provider-neutral #4321 instead of adding a
    Gong exception. Declared-secret masking remains the only permitted provider-output masking.

## Review result

No remaining connector-local correctness or safety finding was identified beyond the source
projection gap above. The branch remains draft/non-merge-ready until generic source-import
common-input preflight, foundation #4321 output-policy correction, #4337 external-proof privacy,
and full live certification gates are resolved.
