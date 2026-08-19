# DISCUSSION LOG — issue #3775

Mode: `discuss-phase` inline/manual fallback using the authoritative parent issue and all four
child issues.

No unresolved gray area remains. Parent #3775 fixes the semantic split between required-property
presence and `min_items` cardinality; #3778 fixes the red contract; #3780 limits the mechanism to
the shared validator; #3781 requires public direct-read and reverse-ETL plan-path proof; and #3783
requires a durable real-runtime regression guard. The captain ruling preserves the Front schemas
without an invented minimum, so no provider or schema alternative is available to choose.

The only operational topology decision is also fixed by the task: serialize the four slices on this
branch and deliver one PR, with no spawned specialist, no provider request, and no reverse-ETL
execution.
