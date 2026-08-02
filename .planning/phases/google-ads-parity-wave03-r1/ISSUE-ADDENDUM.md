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
- Fixed direct reads: 21
- Guarded reverse/write actions: 7
- Sanitized write fixtures: 7
- Blocked/planned rows: 133 (`disallowed=98`, `destructive_action=4`, `admin_reverse_etl=10`, `direct_read=20`, `duplicate=1`)
- Excluded/N/A rows: 0 in v2 operation-ledger mode

**Safety disposition**
- No secrets requested, printed, stored, or added to fixtures.
- No live Google Ads calls, writes, or certification performed.
- No generic GAQL/search passthrough, generic HTTP write, generic SQL write, shell, raw request body, or raw request escape hatch exposed.
- Reverse/write actions remain plan → preview → explicit approval → execute and carry destructive confirmation metadata where applicable.
- Reserved-expansion resource-name methods and direct reads with required complex request bodies remain blocked because executing them without shared reserved-expansion or typed body-field support would be untruthful.

**Local verification passed**
- `go run ./cmd/connectorgen validate internal/connectors/defs/google-ads`
- `go test ./internal/connectors/conformance -run 'TestConformance/google-ads' -count=1`
- `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1`
- `go build ./cmd/pm`
- `make connector-boundary`
- `make verify`
- `git diff --check`
