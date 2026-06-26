# PROMPTS - Wave 1 HTTP API batch 100 (OpenCode subagents)

The orchestrator used the repo's GSD programming-loop discipline with OpenCode `Task` subagents.

Common builder prompt shape:

```text
Use the repo's GSD programming-loop discipline inside this subagent: inspect existing connector patterns, write red-first tests, implement, run package tests, and return concise evidence.

Implement native Go read-only HTTP API connector packages for the assigned catalog sources only. Each package must live under internal/connectors/<bare-name>/ and self-register with connectors.RegisterFactory("<bare-name>", New). Follow existing 10+ year Go quality: small explicit code, connsdk for HTTP, no new dependencies, no secrets in logs/tests, fixture mode must work credential-free, Check/Catalog/Read/Write implemented, Write returns connectors.ErrUnsupportedOperation unless a safe allow-listed write is implemented. Use httptest tests for auth/pagination/record mapping plus fixture/catalog/registration tests. Edit only assigned connector dirs. Do not edit go.mod, registry_gen.go, catalog_data.json, docs, shared packages, or other connector dirs.
```

Subagent slices:

- Group 1: `facebook-marketing`, `google-ads`, `iterable`, `marketo`, `mixpanel`, `recharge`, `salesforce`, `shopify`, `surveymonkey`, `asana`.
- Group 2: `facebook-pages`, `faker`, `hardcoded-records`, `posthog`, `smartsheets`, `yandex-metrica`, `zenloop`, `adjust`, `amazon-sqs`, `appsflyer`.
- Group 3: `avni`, `box-data-extract`, `braintree`, `cart`, `commcare`, `commercetools`, `convex`, `dynamodb`, `e2e-test`, `elasticsearch`.
- Group 4: `freshcaller`, `genesys`, `google-directory`, `gridly`, `jina-ai-reader`, `kyriba`, `kyve`, `linnworks`, `looker`, `microsoft-dataverse`.
- Group 5: `netsuite`, `okta`, `opuswatch`, `orb`, `oura`, `outbrain-amplify`, `outlook`, `outreach`, `oveit`, `pabbly-subscriptions-billing`.
- Group 6: `paddle`, `pagerduty`, `pandadoc`, `paperform`, `papersign`, `pardot`, `partnerize`, `partnerstack`, `payfit`, `pendo`.
- Group 7: `pennylane`, `perigon`, `perk`, `persistiq`, `persona`, `pexels-api`, `phyllo`, `picqer`, `pingdom`, `pipedrive`.
- Group 8: `pipeliner`, `pivotal-tracker`, `piwik`, `plaid`, `planhat`, `plausible`, `pocket`, `pokeapi`, `polygon-stock-api`, `poplar`.
- Group 9: `postmarkapp`, `pretix`, `primetric`, `printify`, `productboard`, `productive`, `public-apis`, `pylon`, `pypi`, `qonto`.
- Group 10: `qualaroo`, `quickbooks`, `railz`, `rd-station-marketing`, `recreation`, `recruitee`, `recurly`, `reddit`, `referralhero`, `rentcast`.
