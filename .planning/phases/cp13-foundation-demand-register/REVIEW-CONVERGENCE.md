# CP13 canonical complete carried finding ledger139

Frozen original reviewed source2e364954a6253442f50d86e0eb5e755b30b1978b/tree24d78d74914da61cea7e70ff3b2d93dcb2a8b137. Full138A72884bytes/SHA9407ccdb0aaeb017a0e523e30981811b3c48013e6c974497fdb5f01650981255; full138B118993bytes/SHA52b2c878e1d1513d3637853e92cd6f9d4d69b66845610ca86b7f172df853f458. Complete original reports/journals/probes/native finals remain immutable in canonical home. Three distinct invariants: CR-12 medium blocker; WR-138A-01/WR-138B-01 low warnings. All three are required corrections under139. CR-02 is an umbrella linked toCR-12, not a duplicate. CR-11/136 is independently closed; CP11/CP12 checkpoint acceptance remains carried.

The complete final finding narratives below are retained verbatim from the sealed reports. Their requested fixes are now authorized by139; their original candidate verdicts are not overwritten. No repair has begun at this ledger commit.

## Narrative Findings (AI reviewer)

### WR-138A-01 — Publish readiness JSON completely before exposing its pathname

**Classification: WARNING. Direct impact: low; test reliability. Final finding disposition: confirmed, should fix.**

**File:** `/Users/karthiksivadas/.treehouse/cli-6bae67/2/cli/cmd/connectorgen/vnext_publication_lock_contention_test.go:217-226`; producer at line 180, public caller at line 117.

**Reachable path and violated contract:** `TestConnectorgenMainSignalsOnlyAfterExactLockContention` starts the real-main child; after real `LOCK_NB` returns contention for the held directory inode, the child publishes its device/inode acknowledgment with `os.WriteFile`. The parent waits for pathname existence, then reads and unmarshals once. The F08R contract requires a bounded complete acknowledgment of that exact lock before signaling. `os.WriteFile` creates/truncates its destination before writing the JSON. If the child is descheduled between those operations, the parent reads an empty file and fails with `unexpected end of JSON input`, before its sound exact-inode check. A correct child can therefore spuriously fail the mandatory signal proof.

**Sibling invariant:** complete publication of interprocess readiness data. The same schedule exists in `cmd/connectorgen/vnext_publication_observation_test.go:101` → `:128-136` and `cmd/connectorgen/vnext_publication_witness_observation_test.go:151` → `:174-183`. Both parse once on successful ReadFile. The result read after child Wait does not share this readiness race. These siblings are one finding, not additional counts.

**Causality:** original CP11 miss, fix-created in the earlier F08R repair `69246943bdcb5c3cdc39c08a7cf1664f4af811aa`, as shown by blame on the actual producer/reader. It predates 094 and is unchanged by CP12/137. The underlying contention condition, real signal assertion and production lock behavior remain valid; this does not reopen their original production bugs.

**Independently expected/observed evidence:** one private overlay kept the real acknowledgment reader and split the real create/write sequence with a channel barrier. The created-but-unwritten schedule fails on the expected EOF; an actual complete `os.WriteFile` positive passes. The writer is released and joined by defer even on Fatal. This demonstrates a possible schedule, not its frequency or an assertion that any accepted historical receipt encountered it.

```text
go test -overlay /Users/karthiksivadas/pm-cli-agent-workspace/data/review-runs/cp11-joint-138a-2e364954-20260907T080056Z/probe-overlay.json -json -count=1 -timeout 20m ./cmd/connectorgen -run '^TestCP11Review138AAcknowledgmentPublication$'
```

Terminal exit 1; Go package elapsed 0.998s; events 2026-09-07T08:17:13.849204Z–08:17:14.847512Z. Raw evidence `probe-contention-output.jsonl`: 3,423 bytes, SHA256 `c4c1f1520e68b15011235704efc9c949893dd1a83e24294a604aa2608063176c`. Overlay: 270 bytes, SHA256 `86d4b325bc1870247e0b5b5bf0343e6ee3f6283ac80b2e1928ecab94c282feb2`. Overlaid test: 10,264 bytes, SHA256 `cf93189c18b9ac6f1001e88d959e38472b1e677e94d852cc381960f9a4d0e3e3`. All are under the exclusive run directory. This probe is not a whole-package gate.

**Concrete fix and regression:** write each acknowledgment to a completed sibling temporary file and rename it to the ready pathname, or send a complete record over a pipe. Preserve the exact device/inode comparison, real LOCK_NB cut, bounded lifecycle and signal assertions. A deterministic regression should pause between temporary creation and complete publication and prove the parent waits and then accepts the correct complete payload. Apply the same protocol to the three siblings. No source fix was made during this frozen review.


## Narrative Findings (AI reviewer)

No structural pre-pass was supplied. Findings below follow direct source review, retained-source counterexamples and independent evidence reconciliation.

### CR-12 — BLOCKER: malformed response keys create successful-response ETL authority

**Impact:** medium. **Acceptance blocking:** yes. **Attribution:** initial_snapshot_miss in273644db, preserved through137; not caused by the local-shape helper. Related original umbrella: CR-02. This is one finding with four witnessed malformed spellings.

**File:** [source_lane_rules.go:401](/Users/karthiksivadas/.treehouse/cli-6bae67/2/cli/cmd/connectorgen/source_lane_rules.go:401), within401–421; classifier144–174. Affected sibling scope checks: [source_lane_collection.go:129](/Users/karthiksivadas/.treehouse/cli-6bae67/2/cli/cmd/connectorgen/source_lane_collection.go:129), [source_lane_bindings.go:269](/Users/karthiksivadas/.treehouse/cli-6bae67/2/cli/cmd/connectorgen/source_lane_bindings.go:269) and2082.

**Issue:** sourceResponseShape treats every response member beginning with `2` as successful. A retained read whose sole response key is `2invalid`, `20`, `2000` or `2ab`, with an otherwise ordinary object-array schema, reaches the actual inventory/facts/classifier/builder and advertises `applicable/mapped_unproven/source_record_collection`. Those keys cannot establish a successful response. The existing collection-scope and binding-coverage consumers use a different three-character test, which still admits `2ab`. Thus classification and scope accounting also disagree on malformed lengths.

| Sole response member | Independently expected ETL | Actual current builder |
| --- | --- | --- |
| `200` | applicable; positive control | applicable |
| `2XX` | applicable; supported range control | applicable |
| `2invalid` | undetermined | applicable |
| `20` | undetermined | applicable |
| `2000` | undetermined | applicable |
| `2ab` | undetermined | applicable |

Every case retained both literal keys `(fixture,primary,source.a)` and `(fixture,primary,source.b)`, observed normalized facts,14 cells in the literal seven-lane order, and applicable direct-read siblings. No accepted target, behavioral proof or implemented cell was created. This is a false authoring claim, not evidence of a runtime corruption incident.

**Evidence:** [private test](/Users/karthiksivadas/pm-cli-agent-workspace/data/review-runs/cp12-joint-138b-2e364954-20260907T080056Z/probes/status138b_test.go), [complete native output](/Users/karthiksivadas/pm-cli-agent-workspace/data/review-runs/cp12-joint-138b-2e364954-20260907T080056Z/probes/status138b-raw.jsonl), [command and custody receipt](/Users/karthiksivadas/pm-cli-agent-workspace/data/review-runs/cp12-joint-138b-2e364954-20260907T080056Z/probes/status138b-receipt.json). The overlay adds one nonexistent virtual test and replaces no production source. Actual Go exit1:7 run events,2 pass,5 fail including the parent. All four negative children fail the intended ETL assertion after the real builder frontier. Go package elapsed1.021s; the yielding transport did not separately measure total command wall time, so none is invented. All199 source hashes, HEAD/tree and tracked state match before/after.

**Fix:** use one closed successful-response-key predicate across shape, collection scopes, required coverage and no-body checks. Preserve exact200–299 and the already supported2XX range. Malformed/unsupported keys must not establish positive or exclusion authority; preserve anchored rows and visible uncertainty. Add permanent actual-builder/validator controls for malformed lengths/nondigits, valid200/299/2XX, non-success responses, binary/scalar consumers and malformed siblings beside valid success. Trace binding/proof scope consumers rather than repairing only the first prefix check.

**Current corpus limit:** the demonstrated malformed keys are private retained fixtures. No affected real provider row or runtime incident is claimed. The current structural-census results remain valid for their actual retained inputs.

### WR-138B-01 — WARNING: optional GraphQL null selector disagrees with its annotation schema

**Impact:** low. **Acceptance blocking:** not independently blocking; Firstmate owns disposition. **Attribution:** fix_created:481795060ab5fa619d87f03d540ee241ee146002, which introduced the annotation-envelope schema. The explicit null-positive reader test predates it at61d674d4. This is separate from137 and CR-11.

**File:** [source-lane-manifest.schema.json:1167](/Users/karthiksivadas/.treehouse/cli-6bae67/2/cli/docs/connector-canon/source-lane-manifest.schema.json:1167), referring to the object-only definition at1030; [source_lane_rules_test.go:434](/Users/karthiksivadas/.treehouse/cli-6bae67/2/cli/cmd/connectorgen/source_lane_rules_test.go:434).

**Issue:** the documented annotation envelope rejects `graphql:null`, although the optional Go pointer and permanent TestSourceLaneGraphQLCitationShapes deliberately accept that representation as carrying no facts. The same actual annotation passes the independent Draft202012 schema with the member absent or `{}`, but fails with null at `annotations/0/graphql`: `None is not of type 'object'`. The current normal receipt includes the passing permanent reader-shape test. No narrower canonical-input restriction is documented at SOURCE-LANE-MANIFEST.md:199–217.

**Fix:** allow null alongside the existing closed graphql_refs object; retain rejection of malformed citations and unknown fields. Add schema/reader parity checks for absent/null/empty/valid/malformed forms.

**Limit:** all six tracked annotations omit this selector; the complete current manifest and annotation envelope validate. This is authoring schema interoperability, not provider behavior, proof promotion or runtime data loss.

