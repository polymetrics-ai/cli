# Mailchimp official API audit trace

Fetched: 2026-07-31T13:41:16.440Z
Source: https://api.mailchimp.com/schema/3.0/Swagger.json
Swagger: 2.0; title: Mailchimp Marketing API; version: 3.0.91; host: server.api.mailchimp.com; basePath: /3.0
Root paths: 181; provider path refs fetched: 181; operations: 298

## Operations by method

- DELETE: 35
- GET: 149
- PATCH: 32
- POST: 75
- PUT: 7

## Operations by family

- lists: 73
- ecommerce: 60
- campaigns: 22
- reports: 22
- automations: 18
- reporting: 12
- file-manager: 11
- sms-campaigns: 10
- audiences: 8
- landing-pages: 8
- connected-sites: 7
- templates: 6
- batch-webhooks: 5
- campaign-folders: 5
- template-folders: 5
- verified-domains: 5
- batches: 4
- conversations: 4
- account-exports: 3
- authorized-apps: 2
- facebook-ads: 2
- /: 1
- activity-feed: 1
- customer-journeys: 1
- ping: 1
- search-campaigns: 1
- search-members: 1

## Red ledger comparison before implementation

Current api_surface rows: 9
Official operations missing from current api_surface: 290
Rows in current api_surface not present in official operation set: 1

This red evidence proves the pre-change Mailchimp connector ledger is incomplete relative to the current official Swagger operation inventory.
