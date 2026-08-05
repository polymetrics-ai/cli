# DISCUSSION LOG — issue-3013-shopify-parity resume

`scripts/gsd prompt discuss-phase issue-3013-shopify-parity --auto` was used for this resumed security slice.

Resolved decision: enforce only `^[a-z0-9][a-z0-9-]*\\.myshopify\\.com$` for Shopify Admin API hosts. The captain explicitly declined an allow-list. No unresolved product, safety, or provider-behavior decision remains.
