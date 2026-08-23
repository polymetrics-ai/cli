# Google Ads v22 source audit

- Source: `https://googleads.googleapis.com/$discovery/rest?version=v22`
- Version: `v22`
- Revision: `20260817`
- Canonical source SHA-256: `c14a489015a3a4664addc58fa429c05b3bce26adc2a519a3a5469d475c18f8f8`
- Canonical source bytes: `2243707`
- Raw source bytes: `2937930`
- Raw methods: `163` ({'DELETE': 1, 'GET': 11, 'POST': 151})
- Schemas: `1363`
- Path variable counts: `{'adGroupAd': 1, 'campaignDraft': 1, 'customerId': 131, 'experiment': 2, 'name': 5, 'resourceName': 13}`
- Classification counts: `{'blocked_duplicate': 1, 'blocked_raw_query': 1, 'blocked_reserved_path': 22, 'direct_read': 33, 'stream': 2, 'write': 104}`
- Generated counts: `{'api_surface_rows': 164, 'streams': 3, 'direct_reads': 33, 'write_actions': 104, 'write_fixtures': 104, 'blocked_rows': 24}`

The `api_surface.json` ledger has one extra row relative to raw discovery because `customers.googleAds.search` backs two fixed connector streams (`campaigns`, `ad_groups`).
