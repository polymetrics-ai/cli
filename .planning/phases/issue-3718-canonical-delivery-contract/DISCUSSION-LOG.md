# DISCUSSION LOG — issue #3718

Mode: `discuss-phase --auto` with authoritative issue decisions.

No unresolved gray areas were selected. Parent issue #3714 and sub-issue #3718 already lock the
state machine, connector insertion points, single-worker model, GSD/no-mistakes constraints,
authority boundary, Wayfinder rejection, and Wave 1 scope fences. Implementation choices are
limited to a deterministic generator/check boundary that future harness waves can consume without
copying the canonical prompt.
