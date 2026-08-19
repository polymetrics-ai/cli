# Code review — issue #4273 connector surface sweep batch 1

## Scope reviewed

- Generated provider API/CLI/operation declarations for the eight gated
  connectors, plus their manuals and website catalog projections.
- The batch manifests, materialization/gate reports, 552-row resumable ledger,
  and foundation-gap tracker.
- The nine root CLI golden snapshots affected by generated connector entries.

## Review method

1. Checked changed paths: no Zoom bundle and no shared engine/classifier code.
2. Checked generated declaration consistency with `connectorgen validate`,
   `surface-sync --check`, the eight-survivor `batch gate`, and the all-bundle
   production runtime-preflight test.
3. Checked CLI/manual/website generation and the scoped snapshot RED/GREEN
   sequence, then ran the full `make verify` gate.
4. Inspected the ledger and reports for a named outcome for every planned
   candidate and for explicit, non-certifying language.
5. Re-reviewed the captain parity-bar addendum: coverage is exact
   implemented-command method/path coverage, not a claim that API-surface
   declaration alone works; every missing operation/ledger row is present in
   the rejection list with an allowed reason, evidence, and recoverability.

## Findings

No Critical, Warning, or actionable Info findings.

The following are deliberate, documented constraints rather than hidden
defects in this connector-local change:

- G12: `file_upload` has no engine executor.
- G13: DELETE action-kind propagation is owned by
  `cli-delete-action-kind-fix-r1`.
- G14: direct-read promotion is blocked by available output policy.
- G15: generic destination transport lacks an executable engine route.
- G16: `surface-reconcile` cannot scope its fleet-wide output to this batch.
- G17: provider callback/event operations cannot be represented separately
  from request paths.
- G18: no declarative SSE source executor exists.
- G19: parameter/body/fan-out schema cannot express the named provider inputs.

G13 is resolved by `31bfe62eb`; the parent and child branches are rebased onto
that commit so DELETE action-kind classification is current.

Every affected accepted connector remains `gated`, not certified, and carries
the actual empty parity-class/transport state in the progress tracker.

## Verdict

**PASS — ready for Firstmate integration of issue #4273 PR #4275 into the
stacked parent.** It is explicitly below the parity-completion bar (0/20 batch
candidates over 90%), which is visible rather than hidden by the `gated` state.
