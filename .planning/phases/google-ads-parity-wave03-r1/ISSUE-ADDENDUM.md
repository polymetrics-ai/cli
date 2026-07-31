<!-- google-ads-parity-wave03-r1-captain-policy-addendum -->
### Captain-policy addendum — Google Ads parity wave03-r1

Fixture-only local implementation complete on branch `fm/cli-google-ads-parity-wave03-r1`; no PR/push/merge performed.

**Official audit source**
- Google Ads REST discovery: `https://googleads.googleapis.com/$discovery/rest?version=v22`
- Version/revision: `v22` / `20260721`
- Raw discovery inventory: 163 methods (`POST=151`, `GET=11`, `DELETE=1`)

**Implemented local parity counts**
- Local operation-ledger rows: 164 (one extra row because `customers.googleAds.search` backs two fixed streams: `campaigns` and `ad_groups`)
- Streams: 3 (`accessible_customers`, `campaigns`, `ad_groups`)
- Fixed direct reads: 34
- Guarded reverse/write actions: 104
- Sanitized write fixtures: 104
- Blocked/planned rows: 23 (`destructive_action=4`, `admin_reverse_etl=10`, `direct_read=8`, `duplicate=1`)
- Excluded/N/A rows: 0 in v2 operation-ledger mode

**Safety disposition**
- No secrets requested, printed, stored, or added to fixtures.
- No live Google Ads calls, writes, or certification performed.
- No generic GAQL/search passthrough, generic HTTP write, generic SQL write, shell, or raw request escape hatch exposed.
- Reverse/write actions remain plan → preview → explicit approval → execute and carry destructive confirmation metadata where applicable.
- Reserved-expansion resource-name methods remain blocked because existing connector-local path interpolation URL-encodes slashes; executing them without shared reserved-expansion support would be untruthful.

**Local verification passed**
- `go run ./cmd/connectorgen validate internal/connectors/defs/google-ads`
- `go test ./internal/connectors/conformance -run 'TestConformance/google-ads' -count=1`
- `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1`
- `go build ./cmd/pm`
- `make connector-boundary`
- `make verify`
- `git diff --check`
