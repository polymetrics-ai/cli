# Carried corrections139 — local implementation evidence

Source base2e364954; artifact-only ledger10ed8370749a743f9d1e331f9f8f9268a852cc01 preceded all behavior edits. CP11/CP12 remain unaccepted; this is one local correction slice inside CP13, not a separate closure-review request. All nine original CP13 obligations and combined independent review remain pending.

## Changes and observable contracts

- CR-12: `sourceResponseStatus` admits exact200–299/2XX success authority. Shape, collection/interpretation scopes, required binding coverage and no-body checking share it. Malformed scopes remain unknown, genuine non-success/default preserve existing exclusion from successful response scope, valid sibling bindings remain intact, unresolved siblings cannot confer complete proof. Scope pointers use JSON Pointer escaping. Initial69 permanent assertions were identical through RED/GREEN (testSHA51731eab495ab9fdd1a2f112ee1930bbc69ad3e259321fce4e3d1b31bcc83566). Later8 interpretation-authority events compare actual valid source citations with malformed scopes through the independent validator; these are later controls, not original pre-edit RED.
- WR-138A-01: the three real child producers now use a common test-only private-file Write/Close/Rename readiness publisher. Permanent deterministic barrier invokes actual parents/readers and actual children; three before-write negatives reproduced the race, three completed controls passed. Regression function bytes remained2257/SHA4f6f4688ff7baf977c2206c8bccd0c3f63f03c53b1370c6cc086f4e35f00412f across RED/GREEN. Initial pre-repair instrumentation preserved defective create-before-write ordering and is not relabeled original138 source. No production publication changes. Original producer line locations101/151/180 and all old caller assertion lines remain unchanged; original119 IDs/28groups/history preserved. New helper file and its dependency are disclosed. Lint found an unchecked cleanup result; final helper joins meaningful cleanup errors while tolerating the renamed-away temporary path. The final18 caller/regression events were rerun normal and race after this change.
- WR-138B-01: published optional graphql member now accepts null or the same closed graphql_refs object, matching existing reader. Literal11-case fixture drives real Go reader and installed Python jsonschema4.25.1 Draft202012Validator; null alone failed schema before edit and all11 agree afterward. No Go/runtime dependency, GraphQL capability or provider input changed. Python validator is an explicit installed external validation prerequisite.

## Actual receipt index

Every private receipt directory is under `data/cli-batch1-pi-takeover/receipts-cp13-139/` in the Firstmate home. Each retains actual command, immutable source/test/JSON/module snapshots, start/end/exit/raw hash and changed-input check. A capture-parent-directory failure happened before a test process existed and remains separately recorded as nonsemantic setup failure. No failed command was deleted.

| Receipt | Exit | Passed/selected Go test events | Wall seconds | Raw SHA256 |
| --- | --- | --- | --- | --- |
| corrections-final-vet-01 | 0 | 0/0 | 1.095871 | e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 |
| corrections-focused-01 | 0 | 314/314 | 98.446439 | 5ac9be37892ff00d3d2b6eb931aa5e64645b71b649b1c514d423e13edbd99112 |
| corrections-lint-01 | 1 | 0/0 | 3.087095 | 9febc18e014ae6caa98f6fd00bc76c55371d10e1350818dcd3d1b65a64f97a93 |
| corrections-lint-02 | 0 | 0/0 | 4.405700 | e92606b0bf483111dff0a120c315ea165821348f31365020e2468a0059095c47 |
| corrections-race-01 | 0 | 107/107 | 31.768358 | 70be451b7a44f5f8b43edf08b87b23bcc895367aef486dfb8223f00008d635dc |
| corrections-vet-01 | 0 | 0/0 | 1.389084 | e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 |
| cr12-green-01 | 0 | 69/69 | 4.868450 | 27f1da443469dee6ed4ad14ba064c750b70604a8a6073498c83a7b87bcb3a9e9 |
| cr12-red-01 | 1 | 35/69 | 5.423587 | 69b799e93822526e41e27040403e4382eada3811f9156115820eae8b4ef65741 |
| graphql-green-01 | 0 | 12/12 | 3.579356 | e47d8a3baa890b6be191a207c3f0fe2e882a743b4283f09166b8324c62f5409a |
| graphql-red-01 | 1 | 12/12 | 4.896609 | e0db7bb1efa0dbb6ec23199631f8c0ab5ae8be75fd56756920b0a9962248df06 |
| readiness-final-normal-01 | 0 | 18/18 | 8.740283 | 01cb3b7263752249b6d1acf3d77fbba18f3d93eddd1bd1860b9cf93ab18b079a |
| readiness-final-race-01 | 0 | 18/18 | 27.933209 | 486b26671365d54d30e30ee9cd678bb5d096edb823c5564879c5745f2847d682 |
| readiness-green-01 | 0 | 7/7 | 6.175048 | 45ef5103b9d3f267ac7d560c60692df5be6629ed54f4ca91b6c5d58e4beee7c9 |
| readiness-red-01 | 1 | 3/7 | 6.217894 | 26fd6d3451ac9bb58133edc0c643eaae39d147455503c9a2313584fe693721a7 |

The107-event race and314-event focused groups precede only the final readiness cleanup error handling. Their unrelated source-lane tests remain source-current under that explicit dependency exclusion; the18 affected handshake events were independently rerun normal/race. This is not an aggregate whole-package run. The final lint/vet receipts bind the corrected helper. Zero selected tests for vet/lint are command checks, not behavioral proof. Original137 broad results remain historical and are not transferred as current-SHA proof.

## Evidence table

| Criterion | Evidence | Actual observable assertion or individual fake reason |
| --- | --- | --- |
| Malformed status authority and valid siblings | fake provider, real retained filesystem/builder/normalizer/admission | Synthetic source.a/source.b rows allow adversarial status keys without provider mutation; asserts exact keys/14cells, per-lane applicability, scoped diagnostics and retained accepted target identity. |
| Independent interpretation authority | fake provider, real source citation/independent validator | Literal synthetic envelope citations allow valid200/299/2XX controls and forged malformed scopes without editing retained provider evidence. |
| Readiness visibility and exact child behavior | live hermetic subprocess/filesystem | Actual original parent/child paths, creation-before-write barrier, no readiness name before completion, complete JSON/identity validation, direct Wait and original bounded cleanup. |
| Published schema/reader parity | fake annotation, real published schema and Go decoder | Literal optional/null/malformed cases are unavailable as all variants in current provider annotations; both real validators must match each independent boolean. |

## Remaining work and consequential boundary

No final generated manifest, Atlas or proof receipt inputs were changed; full final generation/census/determinism/check preservation awaits integrated CP13 implementation, not a claim that old bytes are current proof. Source status correction may affect retained malformed-key rows, so final corpus verification remains required. No PM/App/runtime command surface changed; docs describe authoring annotation parity. No provider-live, credential, customer database, receiver, publication or main merge occurred.

CP13 proof admission has a concrete boundary documented privately in `cp13-139-foundation-proof-boundary.md`: Atlas shared-foundation tests include top-level tests without connector targets, while existing sourceLaneProofShape requires exact source/lane/admitted generation targets and a selected subtest, and assessSourceLaneProof promotes that lane to implemented. A Firstmate-authored bounded Astra/xhigh design must select the authoring-only foundation proof/adopter assessment contract without weakening current lane proof or treating Atlas examples as an exhaustive adopter set. No generic review restart or CP12-only review is requested. CP13 source/receiver reconciliation and all nine obligations remain required; CP14 remains gated by combined closure.
