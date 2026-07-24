# Canonical PM Review Packet, Impact, Hypothesis-Lab, and Response Contract

The parent orchestrator compiles packets with `scripts/pm-review-system.py compile` before spawning
fresh-context local Codex review. Packets are bounded analytical inputs, not lifecycle owners or
independent verdicts. The PM synthesizes one disposition. Shepherd remains a separate downstream
trajectory/evidence validator and never edits code.

## Versioned contracts

- review policy/scope: `polymetrics.ai/pm-review-system/v4` and `polymetrics.ai/pm-review-scope/v1`
- compile manifest: `polymetrics.ai/pm-review-compile/v4`
- practical impact graph: `polymetrics.ai/pm-review-impact-graph/v3`
- packet: `polymetrics.ai/pm-review-packet/v4`
- packet response: `polymetrics.ai/pm-review-packet-response/v4`
- hypothesis-lab request/evidence: `polymetrics.ai/pm-review-lab-request/v2` and
  `polymetrics.ai/pm-review-lab-evidence/v3`
- synthesis: `polymetrics.ai/pm-review-synthesis/v4`

Incompatible v3 and earlier compile/config/packet/response contracts do not upgrade implicitly. Synthesis returns an explicit migration
blocker. Any exact base/head/tree change invalidates the manifest, packet responses, lab evidence,
synthesis, and Shepherd evidence.

## Compiled packet

```json
{
  "schema_version": "polymetrics.ai/pm-review-packet/v4",
  "packet_id": "impact_graph-01",
  "role": "impact_graph",
  "exact_base_sha": "<40 hex>",
  "exact_head_sha": "<40 hex>",
  "exact_head_tree": "<40 hex>",
  "changed_files": [],
  "changed_file_slices": [],
  "closure_files": [],
  "authority_files": [],
  "context_file_slices": [],
  "impact_files": [],
  "impact_edge_ids": [],
  "impact_edges": [],
  "impact_file_slices": [],
  "edge_context_files": [],
  "edge_context_slices": [],
  "invariants": [],
  "context": {
    "target_tokens": 30000,
    "input_token_upper_bound": 0,
    "response_reserve_tokens": 8000,
    "context_window_tokens": 100000,
    "total_context_upper_bound": 0,
    "rendered_prompt_bytes": 0,
    "slice_payload_bytes": 0,
    "slice_separator_bytes": 0,
    "bytes_per_token_upper_bound": 1,
    "estimation": "complete_rendered_prompt_one_token_per_byte",
    "overflow": false,
    "truncated": false
  }
}
```

Compilation reads policy, per-run scope, closure, authority state, and file context only from detached
exact-commit snapshots. Reviewers receive only prompts emitted by
`scripts/pm-review-system.py render --manifest <manifest> --packet-id <id>`; rendering authenticates
the clean candidate and exact blob/slice bindings, then emits the contract, compact packet envelope,
and canonical-order exact slice payloads. Hand-built or augmented packet prompts are prohibited. The manifest binds exact config/scope hashes and independently reconstructible
canonical hashes for coverage, packets, and complete semantics; synthesis recompiles and compares it.
Impact context is discovered completely under the configured typed relation policy before packets
are partitioned. Changed/context/impact blobs are assigned exact revision/blob-bound byte-and-line
slices; edge packets carry every endpoint plus exact bounded provenance slices. Every impact edge has stable id, source, target, relation, base direction, parser,
provenance reason/line, certainty (`active`, `inactive`, or `unknown`), traversal directions, and
minimum depth. Graph/index/traversal/packet bounds block; packet limits never silently trim impact.
This is practical deterministic file/package context, not symbol-level call/data-flow coverage.

## Observable reviewer behavior

Before judging individual lines, each packet reviewer must:

1. build the assigned impact model;
2. trace upstream, downstream, lateral, and temporal/state paths;
3. inspect relevant history when behavior/ownership/compatibility is ambiguous;
4. compare divergent siblings/variants when present;
5. state falsifiable hypotheses and strongest alternatives;
6. seek disconfirming evidence;
7. use the smallest discriminating counterfactual experiment when static evidence is insufficient;
8. report limitations and unreviewed impact rather than substitute a confidence statement.

These are observable behaviors. They are not a claim of equivalence to an ideal or maximally
efficient human reviewer.

## Disposable hypothesis lab

The canonical candidate is read-only. A reviewer requests one experiment through
`scripts/pm-review-lab.py run`; the tool clones the exact head into a private per-packet temporary
root, applies bounded replacements only there, runs one allowlisted local command under a proven OS
sandbox, captures bounded evidence, proves the candidate unchanged, kills descendants, and destroys
the whole lab.

```json
{
  "schema_version": "polymetrics.ai/pm-review-lab-request/v2",
  "hypothesis_id": "H1",
  "claim": "the candidate accepts an unknown current-schema gate kind",
  "alternative": "the behavior is legacy-only and cannot occur for current schema",
  "impact_edges_examined": ["edge-..."],
  "temporary_change": "replace the fixture gate kind with an unknown value",
  "changes": [
    {"path": "relative/file", "find": "exact once", "replace": "temporary text"}
  ],
  "command": ["python3", "scripts/<targeted-test>.py"],
  "expected_discriminator": {"exit_code": 1},
  "limits": {"timeout_seconds": 10}
}
```

Denied: candidate/outside/symlink writes; generic shell; network; commit/push/PR/remote mutation;
dependency installation; credentials/live connectors; deployment/production access; destructive
external effects; and unavailable/policy-only sandbox fallback. The environment is allowlisted and
uses private HOME/TMP/cache roots with no ambient secret/config values. Limits cover request/change,
time, process count, disk, output, file descriptors, and CPU. Any denial, ambiguity, timeout, bound,
Git mutation, process residue, cleanup failure, or candidate identity drift is blocked evidence.

A performed experiment records argv, stdout, stderr, exit, duration, expected/observed discriminator,
examined edge ids, temporary diff hash/stat/paths, sandbox/limits, exact candidate identity, and
cleanup proof. The response author determines whether it supports the claim, supports the
alternative, or is inconclusive. Inconclusive performed evidence cannot support `clean`. If no
experiment is necessary, `no_experiment_reason` explains why static evidence is decisive.

## Reviewer response

Store each raw response and lab evidence outside the tracked worktree or under a Git-ignored
private evidence root. Do not commit it after exact-head review.

```json
{
  "schema_version": "polymetrics.ai/pm-review-packet-response/v4",
  "packet_id": "impact_graph-01",
  "exact_base_sha": "<same exact base>",
  "exact_head_sha": "<same exact head>",
  "exact_head_tree": "<same exact tree>",
  "status": "clean",
  "reviewed_files": [],
  "changed_file_slices": [],
  "closure_files": [],
  "authority_files": [],
  "context_file_slices": [],
  "impact_files": [],
  "impact_edge_ids": [],
  "impact_file_slices": [],
  "edge_context_files": [],
  "edge_context_slices": [],
  "invariants": [
    {"id": "impact_complete", "status": "pass", "evidence_paths": []}
  ],
  "unreviewed_files": [],
  "review_behaviors": {
    "impact_model_built_first": true,
    "directions_traced": ["upstream", "downstream", "lateral", "temporal"],
    "history_inspected": {"status": "inspected", "reason": "compatibility behavior changed"},
    "sibling_paths_compared": {"status": "not_needed", "reason": "no sibling path exists"},
    "hypotheses": [{
      "id": "H1",
      "claim": "the exact assigned coverage is complete",
      "strongest_alternative": "an assigned slice or edge is omitted",
      "falsifier": "any response assignment differs field-for-field",
      "evidence_paths": ["scripts/tests/pm-review-system.sh"]
    }],
    "disconfirming_evidence": "the strongest legacy-only alternative was tested"
  },
  "experiments": [
    {
      "hypothesis_id": "H1",
      "claim": "...",
      "alternative": "...",
      "impact_edges_examined": ["edge-..."],
      "temporary_change": "...",
      "command": ["python3", "scripts/<targeted-test>.py"],
      "expected_discriminator": {"exit_code": 1},
      "observed": {
        "exit_code": 1,
        "limit_hit": null,
        "process_residue_detected": false,
        "processes_remaining": 0,
        "sandbox_denial_observed": false
      },
      "supports": "claim",
      "candidate_unchanged": true,
      "lab_cleanup_verified": true,
      "lab_evidence_path": "pm-review-evidence.tmp/<head>/<packet>/H1.json",
      "lab_evidence_sha256": "<64 hex>"
    }
  ],
  "no_experiment_reason": null,
  "findings": [],
  "residual_risk": [],
  "context": {
    "input_tokens": null,
    "output_tokens": null,
    "cost": null,
    "overflow": false,
    "truncated": false
  },
  "wall_clock_ms": null
}
```

Every hypothesis is an exact object with `id`, `claim`, `strongest_alternative`, `falsifier`, and
non-empty `evidence_paths`; identifier-only strings are incompatible. Every performed response
experiment matches its hashed v3 lab artifact field-for-field for hypothesis, claim, alternative,
edges, temporary change, argv, expected discriminator, and observation. `status` is exactly `clean`, `findings`, or `blocked`. Finding count is unlimited. Every finding
uses severity, category, path/line evidence, impact, and smallest safe correction. Missing token,
cost, or latency data is `null`, never invented.

## Synthesis rules

`scripts/pm-review-system.py synthesize` blocks instead of returning clean when:

- a response is absent or on an incompatible schema without explicit migration;
- exact base/head/tree differs;
- any assigned changed, closure, authority, impact-file, impact-edge, or invariant item is omitted;
- directional behavior, reasoned history/sibling handling, or disconfirming evidence is absent;
- no experiment and no decisive-static reason is supplied;
- a performed experiment is incomplete, inconclusive, blocked, unsafe, or lacks candidate/cleanup proof;
- any response declares unreviewed context, overflow, truncation, or blocked status.

Any finding produces `findings_correction_required`. A failed assigned invariant paired with a valid finding enters that correction state; it is not misclassified as missing coverage. Only complete clean responses with no finding
produce one PM-owned `clean` synthesis. After clean synthesis, independent Shepherd validates the
same exact identities and trajectory. Neither packet review, lab evidence, synthesis, nor Shepherd
grants merge authority.
