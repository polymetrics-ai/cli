# UAT — issue-3753-rate-limit-enforcement-r1

Local acceptance evidence:

1. A declared endpoint/tier/auth policy attaches admission only to matching traffic and parses its declared actual-cost header.
2. Unknown, not-applicable, absent, and unmatched declarations leave the requester unchanged.
3. Direct reads, operation reads, check, stream/fan-out pagination, form/multipart and operation writes, binary downloads, and whole-connector hook access are backed by path tests that require a rate-limit wait.
4. Linked credential bindings share the opaque budget across credential revisions; an unlinked binding does not.
5. Legacy page-loop throttling remains active beside, rather than replaced by, declared requester admission.
