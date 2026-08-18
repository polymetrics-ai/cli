# Discussion log — GitHub certification suite r1

The autonomous launch brief and the live-measurement design settle the delivery
decisions: GitHub is the sole connector, `cli_surface.json` is the exhaustive
command inventory, every command must have exactly one non-ambiguous outcome,
and a successful exit is not evidence. The suite stays serial and resumable.
Provider refusals are terminal non-pass observations; declaration/runtime
disagreements are product defects and receive a separate findings list.

PR #4198 is open, so its `http_exchanges` capture is unavailable. This slice
therefore builds and checks candidate generation, accounting, defect discovery,
and the read/write sweep contract without emitting accepted evidence records.
The evidence writer is a named follow-up once #4198 is merged.

No product decision remains open. This is the documented inline/manual GSD
fallback: the task forbids waiting and the canonical delivery contract forbids
spawning GSD roles, so the single worker runs the prompted lifecycle inline.
