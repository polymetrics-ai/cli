<!-- fm-crisp-captain-policy-addendum-v1 -->
## Captain policy addendum — Crisp parity wave02 r1

This addendum preserves the existing r2 body content and count tables. It only clarifies execution policy for the current connector-local Crisp workstream.

- Every documented Crisp REST API operation remains in scope for the connector-owned ledger, including DELETE, destructive, sensitive, and admin actions previously classified unsafe. Do not blanket-exclude those operations as unsafe.
- An operation becomes executable only after connector-owned typed schemas, bounded flags/inputs, redaction, risk text, fixture/conformance evidence, and the correct safety path exist. Reverse ETL must remain plan -> preview -> explicit approval -> execute; destructive/admin actions additionally require typed destructive confirmation.
- Provider search/query, direct-read, binary/file, and changefeed operations that lack a shared safe execution contract or connector-owned fixed-target policy must stay represented as planned/blocked with source evidence. Do not expose raw method/path/body/query, generic HTTP write, generic SQL write, shell, arbitrary file, or passthrough API escape hatches.
- No live provider calls, credentials, writes, certification, VPS/Thaalam work, merges, pushes, or PR creation are authorized by this addendum.
