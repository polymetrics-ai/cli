# Local exact-head review gate

The local reviewer is read-only and independent. Its verdict binds candidate/base SHA, policy hash,
observed validator model/thinking, independent checks, finding dispositions, and creation time.
Moved heads invalidate the verdict. GitHub automated review stays enabled during shadow/canary and
may be removed only in the approved cutover slice.
