# UAT — API → API GitHub route proof

- D1 PASS: the fresh `pm` binary completed one `issues` source record through
  the connection-owned WAL/Parquet receipt and `full_append` / `append` /
  `add_issue_labels` destination route.
- D2 PASS: `gh-axi` independently observed exactly the requested label on
  destination issue `#2`; a persisted acknowledged checkpoint followed. The
  explicit cleanup independently returned that issue to no labels.
- D3 PASS: focused regressions cover the typed ineligible-stream refusal,
  unsupported canonical mode before source I/O, explicit-only zero results,
  absent configured issue mapping, no duplicate label after replay, and
  explicit missing-label cleanup.

No human judgment remains for the route proof. The retained repository and its
two run-owned sentinel issues remain available for independent inspection.
