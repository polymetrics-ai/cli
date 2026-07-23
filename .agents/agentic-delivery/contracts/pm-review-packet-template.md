# Canonical PM Review Packet and Response Contract

The parent orchestrator compiles packets with `scripts/pm-review-system.py compile` before spawning
fresh-context local Codex review. Packets are bounded review inputs, not independent lifecycle
owners or verdicts. The PM synthesizes exactly one local-Codex disposition. Shepherd remains a
separate downstream trajectory/evidence validator.

## Compiled packet

```json
{
  "schema_version": "polymetrics.ai/pm-review-packet/v1",
  "packet_id": "architecture_reference-01",
  "role": "architecture_reference",
  "exact_base_sha": "<40 hex>",
  "exact_head_sha": "<40 hex>",
  "changed_files": [],
  "closure_files": [],
  "authority_files": [],
  "invariants": [],
  "context": {
    "target_tokens": 30000,
    "estimated_tokens": 0,
    "overflow": false,
    "truncated": false
  },
  "required_response_fields": [
    "reviewed_files",
    "closure_files",
    "authority_files",
    "invariants",
    "unreviewed_files",
    "findings",
    "context"
  ]
}
```

## Reviewer response

Store each raw response outside the tracked worktree or under a Git-ignored evidence location. Do
not commit a response after exact-head review because that would change the reviewed head.

```json
{
  "schema_version": "polymetrics.ai/pm-review-packet-response/v1",
  "packet_id": "architecture_reference-01",
  "exact_base_sha": "<same exact base>",
  "exact_head_sha": "<same exact head>",
  "status": "clean",
  "reviewed_files": [],
  "closure_files": [],
  "authority_files": [],
  "invariants": [
    {
      "id": "active_reference_closure",
      "status": "pass",
      "evidence_paths": []
    }
  ],
  "unreviewed_files": [],
  "findings": [],
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

`status` is exactly `clean`, `findings`, or `blocked`. Finding count is unlimited. Every finding
uses severity, category, path/line evidence, impact, and smallest safe correction. Missing token,
cost, or latency data is `null`, never invented.

## Synthesis rules

`scripts/pm-review-system.py synthesize` blocks rather than returning clean when:

- a packet response is absent;
- exact base/head differ from the compiled candidate;
- any assigned changed, closure, authority, or invariant item is undeclared;
- any response declares an unreviewed file;
- overflow or truncation is true;
- a response is blocked or malformed.

Any finding produces `findings_correction_required`. Only complete clean responses with no finding
produce one PM-owned `clean` synthesis. Any changed base/head invalidates all packet responses and
Shepherd evidence.

After clean synthesis, run independent Shepherd validation against the same exact identities.
Neither packet review, synthesis, nor Shepherd grants merge authority.
