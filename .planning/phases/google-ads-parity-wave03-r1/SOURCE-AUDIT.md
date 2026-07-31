# Google Ads v22 source audit

- Source: `https://googleads.googleapis.com/$discovery/rest?version=v22`
- Version: `v22`
- Revision: `20260721`
- Raw methods: `163` ({'DELETE': 1, 'GET': 11, 'POST': 151})
- Schemas: `1363`
- Path variable counts: `{'adGroupAd': 1, 'campaignDraft': 1, 'customerId': 131, 'experiment': 2, 'name': 5, 'resourceName': 13}`
- Classification counts: `{'blocked_duplicate': 1, 'blocked_reserved_path': 22, 'direct_read': 34, 'stream': 2, 'write': 104}`
- Generated counts: `{'api_surface_rows': 164, 'streams': 3, 'direct_reads': 34, 'write_actions': 104, 'write_fixtures': 104, 'blocked_rows': 23}`

The `api_surface.json` ledger has one extra row relative to raw discovery because `customers.googleAds.search` backs two fixed connector streams (`campaigns`, `ad_groups`).
