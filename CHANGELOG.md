# Changelog

## [0.2.0](https://github.com/polymetrics-ai/cli/compare/v0.1.1...v0.2.0) (2026-08-18)


### ⚠ BREAKING CHANGES

* **warehouse:** materialize tables as Parquet and make DuckDB the query engine ([#3903](https://github.com/polymetrics-ai/cli/issues/3903))
* **warehouse:** The local warehouse layout now nests every materialized table under the connection that produced it, at <workspace-id>/<connector>/<connection-id>/tables/<table>.jsonl, so two connections can no longer overwrite each other's rows. A warehouse written by the earlier flat layout, in which every connection shared one table file, is refused on read and on write rather than migrated: which connection owned a flat table is unknowable, so guessing would recreate the data loss this fixes. Delete the warehouse directory and re-run your syncs to rebuild it. pm will not rewrite or delete warehouse data on your behalf.

### Features

* **app:** add secret-free credential coordination identities ([#3875](https://github.com/polymetrics-ai/cli/issues/3875)) ([bf54aa4](https://github.com/polymetrics-ai/cli/commit/bf54aa4d48542552a9cea6c5eb54cd67495c96f7))
* batch compatible connector and website updates ([#3891](https://github.com/polymetrics-ai/cli/issues/3891)) ([8a2971a](https://github.com/polymetrics-ai/cli/commit/8a2971a0c88ac5e980413c48529f714cccbb0a44))
* **certification:** generate proof-bearing connector matrix ([#3999](https://github.com/polymetrics-ai/cli/issues/3999)) ([815dc1a](https://github.com/polymetrics-ai/cli/commit/815dc1ab65380e03f6e0c078ba36030baaec21ea))
* **connectorgen:** add connector boundary guard ([#605](https://github.com/polymetrics-ai/cli/issues/605)) ([787547a](https://github.com/polymetrics-ai/cli/commit/787547a78d10e36a86c52d5417c517e8a80770dd))
* **connectors:** add API-surface provenance evidence ([#3869](https://github.com/polymetrics-ai/cli/issues/3869)) ([5da7555](https://github.com/polymetrics-ai/cli/commit/5da7555964a87f4921013fded59df83a67a78e56))
* **connectors:** add Asana documented operation parity ([#3538](https://github.com/polymetrics-ai/cli/issues/3538)) ([fc07dc8](https://github.com/polymetrics-ai/cli/commit/fc07dc8303427b781b86f942f2bc8fe756bd8d45))
* **connectors:** add Bitbucket declarative connector parity ([#3531](https://github.com/polymetrics-ai/cli/issues/3531)) ([bfe7854](https://github.com/polymetrics-ai/cli/commit/bfe785464d04fd73dba0c4a70f36e23dd84da3d0))
* **connectors:** add dynamic schema discovery ([#3892](https://github.com/polymetrics-ai/cli/issues/3892)) ([c057bb8](https://github.com/polymetrics-ai/cli/commit/c057bb81d174a515eb611f7713acc3caea59125c))
* **connectors:** add fixture-only Google Ads v22 parity ([#3535](https://github.com/polymetrics-ai/cli/issues/3535)) ([5d61794](https://github.com/polymetrics-ai/cli/commit/5d61794f76c46ca256280eb4166c5285c4b68731))
* **connectors:** add Freshchat connector parity ([#3536](https://github.com/polymetrics-ai/cli/issues/3536)) ([b053dc4](https://github.com/polymetrics-ai/cli/commit/b053dc4a3ad7930738cd719357178e90e8dda333))
* **connectors:** add HubSpot operation ledger ([#3529](https://github.com/polymetrics-ai/cli/issues/3529)) ([41a0039](https://github.com/polymetrics-ai/cli/commit/41a00398a88db809b4e799a59fea381ace5cc06e))
* **connectors:** add MySQL container test harness ([#3952](https://github.com/polymetrics-ai/cli/issues/3952)) ([a30dbd3](https://github.com/polymetrics-ai/cli/commit/a30dbd3a06e1be1013899094458d473629e21d50))
* **connectors:** add polling-watermark changefeed executor ([#3880](https://github.com/polymetrics-ai/cli/issues/3880)) ([dc1a6a7](https://github.com/polymetrics-ai/cli/commit/dc1a6a7171ca901c8dbaf8cd528f67b18e57d9bb))
* **connectors:** add PostgreSQL logical replication CDC foundation ([#3967](https://github.com/polymetrics-ai/cli/issues/3967)) ([c69e58a](https://github.com/polymetrics-ai/cli/commit/c69e58ab813914bd892f43ff4c060b8ed0faa3bc))
* **connectors:** add provenance-backed batch authoring pipeline ([#3881](https://github.com/polymetrics-ai/cli/issues/3881)) ([cb1d3c4](https://github.com/polymetrics-ai/cli/commit/cb1d3c45fe08afaf51cb684284669d588e7b1d30))
* **connectors:** add rate-limit declaration and admission support ([#3874](https://github.com/polymetrics-ai/cli/issues/3874)) ([3311a1c](https://github.com/polymetrics-ai/cli/commit/3311a1c6d14e360cad836c6bb38d361a7a6950e4))
* **connectors:** add source-pinned GitHub operation parity ([#3970](https://github.com/polymetrics-ai/cli/issues/3970)) ([4df0b04](https://github.com/polymetrics-ai/cli/commit/4df0b0416e46958d9acb1b02708464570c070e0f))
* **connectors:** add Stripe official API parity ([#3530](https://github.com/polymetrics-ai/cli/issues/3530)) ([86d5109](https://github.com/polymetrics-ai/cli/commit/86d510927a05aa56b184bf5a8778b5444c69b9b1))
* **connectors:** add Zendesk Support operation ledger parity ([#3532](https://github.com/polymetrics-ai/cli/issues/3532)) ([e99d6f1](https://github.com/polymetrics-ai/cli/commit/e99d6f1193814d00bd1b0c09fc092639d4fd8c54))
* **connectors:** bound multipart uploads and provider search ([#3701](https://github.com/polymetrics-ai/cli/issues/3701)) ([b5099e7](https://github.com/polymetrics-ai/cli/commit/b5099e760a16f80520bec51cea38a8cec4e2b9be))
* **connectors:** bring gong to documented-operation parity ([#3895](https://github.com/polymetrics-ai/cli/issues/3895)) ([d808203](https://github.com/polymetrics-ai/cli/commit/d8082031e40323e9cffc24d3395a1cf5f1c320bb)), closes [#2997](https://github.com/polymetrics-ai/cli/issues/2997) [#2998](https://github.com/polymetrics-ai/cli/issues/2998)
* **connectors:** bring gorgias to documented-operation parity ([#3896](https://github.com/polymetrics-ai/cli/issues/3896)) ([1cecbb9](https://github.com/polymetrics-ai/cli/commit/1cecbb9352c038eae7b74155fb52590729083802)), closes [#196](https://github.com/polymetrics-ai/cli/issues/196)
* **connectors:** bring notion to documented-operation parity ([#3894](https://github.com/polymetrics-ai/cli/issues/3894)) ([d71c620](https://github.com/polymetrics-ai/cli/commit/d71c6206b44455f463e64b454862e85419963698))
* **connectors:** complete Google Search Console operation parity ([#3731](https://github.com/polymetrics-ai/cli/issues/3731)) ([d30dd49](https://github.com/polymetrics-ai/cli/commit/d30dd490570b29daa167f3c0ebefcd4cee8179ca))
* **connectors:** complete Hubplanner operation parity ([#3559](https://github.com/polymetrics-ai/cli/issues/3559)) ([7774bf4](https://github.com/polymetrics-ai/cli/commit/7774bf4012599c6e6683cf8edb59f2213f94ebfa))
* **connectors:** complete Mailchimp operation parity ([#3562](https://github.com/polymetrics-ai/cli/issues/3562)) ([5d43d7c](https://github.com/polymetrics-ai/cli/commit/5d43d7c00b56ddfc077b0a730028b74c5daf5a10))
* **connectors:** complete Recurly operation parity ([#3736](https://github.com/polymetrics-ai/cli/issues/3736)) ([4871df2](https://github.com/polymetrics-ai/cli/commit/4871df2f82777bba69e46f0a53fb1267f3bda70f))
* **connectors:** complete Xero Accounting parity ([#3537](https://github.com/polymetrics-ai/cli/issues/3537)) ([2685c37](https://github.com/polymetrics-ai/cli/commit/2685c3712a9677cf0f74864edf79474b823db55e))
* **connectors:** declare and enforce non-batchable write actions ([#3698](https://github.com/polymetrics-ai/cli/issues/3698)) ([c5ce52f](https://github.com/polymetrics-ai/cli/commit/c5ce52fcf62e37362d9c8a7d4e84a0d8264ac384))
* **connectors:** declare supported CLI output policies ([#3870](https://github.com/polymetrics-ai/cli/issues/3870)) ([ee26d20](https://github.com/polymetrics-ai/cli/commit/ee26d20fc5c4df573be2de3656c0b0f2180b137a))
* **connectors:** enforce direct-read runtime preflight ([#3890](https://github.com/polymetrics-ai/cli/issues/3890)) ([e01f9e1](https://github.com/polymetrics-ai/cli/commit/e01f9e18b5103610a6f941c4f4d1a8a6494b4737))
* **connectors:** engine shared capabilities for bounded binary download, write queries, and dynamic fields ([#3699](https://github.com/polymetrics-ai/cli/issues/3699)) ([504a7c0](https://github.com/polymetrics-ai/cli/commit/504a7c07a9c03bacbe7ad00a4664df063b103d73))
* **connectors:** execute declared REST writes safely ([#3739](https://github.com/polymetrics-ai/cli/issues/3739)) ([83affbf](https://github.com/polymetrics-ai/cli/commit/83affbf5b730619e819ca6f7cf2907ac122560b7))
* **connectors:** expose GitLab stream read commands ([#3760](https://github.com/polymetrics-ai/cli/issues/3760)) ([de5ebb5](https://github.com/polymetrics-ai/cli/commit/de5ebb55b1a1b4de44e3d920b7104670e2d63b02))
* **connectors:** expose Zoom read commands ([#3759](https://github.com/polymetrics-ai/cli/issues/3759)) ([02b0f32](https://github.com/polymetrics-ai/cli/commit/02b0f3267c2dd37647b8c6321fdefff018d3466a))
* **connectors:** extract seven completed connector surfaces ([#3961](https://github.com/polymetrics-ai/cli/issues/3961)) ([0600dfe](https://github.com/polymetrics-ai/cli/commit/0600dfefba4fc5c1c274755c8a8c6a5e1221a9bb))
* **connectors:** genericize repository read policies ([#609](https://github.com/polymetrics-ai/cli/issues/609)) ([7f4b9f3](https://github.com/polymetrics-ai/cli/commit/7f4b9f31189ba68221a9b70347cd0f3702f889a3))
* **connectors:** land production MVP certification and parity ([#4250](https://github.com/polymetrics-ai/cli/issues/4250)) ([e8baeb3](https://github.com/polymetrics-ai/cli/commit/e8baeb353180ee533fc9bb840c3bddb7ed600396))
* **connectors:** load certification contracts from definitions ([#610](https://github.com/polymetrics-ai/cli/issues/610)) ([c85740b](https://github.com/polymetrics-ai/cli/commit/c85740b6f52ff53abbf521b631ed9d10209d567b))
* **connectors:** require preview-bound approval for destructive writes ([#3730](https://github.com/polymetrics-ai/cli/issues/3730)) ([451a21d](https://github.com/polymetrics-ai/cli/commit/451a21da0e4cdcc2fa1db5254aa48f4ce0a9d85a))
* **connectors:** restore Google Calendar documented operation parity ([#3725](https://github.com/polymetrics-ai/cli/issues/3725)) ([c0d95ce](https://github.com/polymetrics-ai/cli/commit/c0d95ce1282fd66b12bcc9a006b80cd03c070b8b))
* **connectors:** restore YouTube Analytics operation parity ([#3726](https://github.com/polymetrics-ai/cli/issues/3726)) ([09ecec8](https://github.com/polymetrics-ai/cli/commit/09ecec8f6ab90a305a7fe25e41ea325a2f9a03a4))
* **connectors:** support typed multipart rest writes ([#3871](https://github.com/polymetrics-ai/cli/issues/3871)) ([4d77ef3](https://github.com/polymetrics-ai/cli/commit/4d77ef3ed2bcb2b358be7fd5ac3b3a0971c3dbc2))
* **connectors:** unblock reverse-ETL writes for zendesk-support and asana ([#3711](https://github.com/polymetrics-ai/cli/issues/3711)) ([36b431c](https://github.com/polymetrics-ai/cli/commit/36b431cf152eca02c782cf84ff155c33c2175007))
* **connectors:** validate declared credential configuration constraints ([#3738](https://github.com/polymetrics-ai/cli/issues/3738)) ([56051ea](https://github.com/polymetrics-ai/cli/commit/56051eada51ac83cc815ff17d39728347609c239))
* **engine:** add oauth2_refresh_token auth mode with vault-backed rotation ([#3709](https://github.com/polymetrics-ai/cli/issues/3709)) ([7dffe9f](https://github.com/polymetrics-ai/cli/commit/7dffe9ff3c5bdd89a9d333f3a168bf2c834e8609))
* **engine:** add request-shaping foundations for arrays, pagination, query and upload ([#3700](https://github.com/polymetrics-ai/cli/issues/3700)) ([48e9283](https://github.com/polymetrics-ai/cli/commit/48e92839831b1214b854402ef089c1b041938701))
* **synccontract:** add durable database sync contract ([#3882](https://github.com/polymetrics-ai/cli/issues/3882)) ([da45c63](https://github.com/polymetrics-ai/cli/commit/da45c632811b960018a16baacfd07087483b7a18))
* **synctransport:** add GitHub issue-label warehouse transport ([#4082](https://github.com/polymetrics-ai/cli/issues/4082)) ([2df18ee](https://github.com/polymetrics-ai/cli/commit/2df18ee3a083fe507cbe1c07e0270e82c5ab0182))
* **warehouse:** materialize tables as Parquet and make DuckDB the query engine ([#3903](https://github.com/polymetrics-ai/cli/issues/3903)) ([08cc41c](https://github.com/polymetrics-ai/cli/commit/08cc41c872c42f79fb69d688a507926d24e3c903))


### Bug Fixes

* **agents:** reject drifted clean-project worker projections ([#3742](https://github.com/polymetrics-ai/cli/issues/3742)) ([7d34a07](https://github.com/polymetrics-ai/cli/commit/7d34a07949146d2b099667d8214209312f9166d2))
* **app:** preserve sync state and warehouse durability ([#3885](https://github.com/polymetrics-ai/cli/issues/3885)) ([b06230d](https://github.com/polymetrics-ai/cli/commit/b06230d2e5a7c3458c1e64129188a2aef9821811))
* **certify:** bound certification harness cost ([#3878](https://github.com/polymetrics-ai/cli/issues/3878)) ([2afd1a5](https://github.com/polymetrics-ai/cli/commit/2afd1a529a8fc24375bb0f7cf92b951133088874))
* **ci:** pin immutable build dependencies ([#4000](https://github.com/polymetrics-ai/cli/issues/4000)) ([07e35e1](https://github.com/polymetrics-ai/cli/commit/07e35e1fe105fd1a2129929eb8e8cafed592ce7d))
* **cli:** reject unresolved connector help paths ([#3964](https://github.com/polymetrics-ai/cli/issues/3964)) ([f96a47e](https://github.com/polymetrics-ai/cli/commit/f96a47e801b89f25386c33951a53a93d1a4c7c8d))
* **commandrunner:** preserve connector command content ([#3868](https://github.com/polymetrics-ai/cli/issues/3868)) ([50deaad](https://github.com/polymetrics-ai/cli/commit/50deaade988f777a519e1b6cceec550d5ab7f64e))
* **commandrunner:** preserve explicit empty string arrays ([#3851](https://github.com/polymetrics-ai/cli/issues/3851)) ([2847356](https://github.com/polymetrics-ai/cli/commit/28473561f89270e59e180a728522b865cad46adb))
* **connectors:** allow sensitive policies without redaction ([#3743](https://github.com/polymetrics-ai/cli/issues/3743)) ([268fb11](https://github.com/polymetrics-ai/cli/commit/268fb11c50399be58e9c6de4243de3a2a5241442))
* **connectors:** block hollow union write commands ([#3737](https://github.com/polymetrics-ai/cli/issues/3737)) ([3511aa1](https://github.com/polymetrics-ai/cli/commit/3511aa1fa0f27d9c985eaf79f48c3a4bcd4687fc))
* **connectors:** derive CDC catalog eligibility from changefeed contracts ([#3861](https://github.com/polymetrics-ai/cli/issues/3861)) ([d215d96](https://github.com/polymetrics-ai/cli/commit/d215d96363a1bcfefd3da9140f09afe713cc9a88))
* **connectors:** derive direct-read page context and command parameters from connector declarations ([#3902](https://github.com/polymetrics-ai/cli/issues/3902)) ([d453fbe](https://github.com/polymetrics-ai/cli/commit/d453fbe256eb22d90ea77dbed634b245bd6e795b))
* **connectors:** enforce connector path ownership guardrails ([#3580](https://github.com/polymetrics-ai/cli/issues/3580)) ([ed7a694](https://github.com/polymetrics-ai/cli/commit/ed7a6945419b14441d252f8e2a8ea04f97b52a6b))
* **connectors:** make implemented commands executable and route binary downloads ([#3712](https://github.com/polymetrics-ai/cli/issues/3712)) ([e2e1393](https://github.com/polymetrics-ai/cli/commit/e2e13934034d0936d6c0e411d89d3bffc49d4140))
* **connectors:** recognize unvalidated-cloud-checkpoint issue links in issueguard ([#3675](https://github.com/polymetrics-ai/cli/issues/3675)) ([61142ce](https://github.com/polymetrics-ai/cli/commit/61142ce25189cdcb53387a9d5ab5334ca2f6ae24))
* **engine:** preserve complete connector content ([#3872](https://github.com/polymetrics-ai/cli/issues/3872)) ([85557fd](https://github.com/polymetrics-ai/cli/commit/85557fd7f4c2260cbd1eae857e1159720a263788))
* **iconregistrygen:** honor curated icon registry entries ([#3854](https://github.com/polymetrics-ai/cli/issues/3854)) ([b0cd6cb](https://github.com/polymetrics-ai/cli/commit/b0cd6cbcdb8484062362443a2570f69a1eb5c4f8))
* **warehouse:** nest materialized tables under their owning connection ([#3901](https://github.com/polymetrics-ai/cli/issues/3901)) ([fcff76a](https://github.com/polymetrics-ai/cli/commit/fcff76a7305fe469c3903c33a89bc47912852ac6))
* **website:** restore lint command ([#542](https://github.com/polymetrics-ai/cli/issues/542)) ([dc75308](https://github.com/polymetrics-ai/cli/commit/dc753087d3ec7cbb7d317869a26550db5de25cd2))
* **website:** wait for profile settings data ([#540](https://github.com/polymetrics-ai/cli/issues/540)) ([446039f](https://github.com/polymetrics-ai/cli/commit/446039fa0fe07d6e0a301f38ced1cdc444377bcc))

## [0.1.1](https://github.com/polymetrics-ai/cli/compare/v0.1.0...v0.1.1) (2026-07-28)


### Bug Fixes

* **connectors:** add Gong calls list date filters ([#597](https://github.com/polymetrics-ai/cli/issues/597)) ([b41dc61](https://github.com/polymetrics-ai/cli/commit/b41dc611ccfadc1477cab60aa309fa39446b21cd))

## [0.1.0](https://github.com/polymetrics-ai/cli/compare/v0.1.0...v0.1.0) (2026-07-27)


### ⚠ BREAKING CHANGES

* complete connector architecture v2 migration ([#29](https://github.com/polymetrics-ai/cli/issues/29))

### Features

* add flow RLM and agent query output ([#19](https://github.com/polymetrics-ai/cli/issues/19)) ([68db8a1](https://github.com/polymetrics-ai/cli/commit/68db8a1e15c51749edb956d8960ce08d394a15b9))
* **agentic:** add issue-first delivery system ([#47](https://github.com/polymetrics-ai/cli/issues/47)) ([7eb144f](https://github.com/polymetrics-ai/cli/commit/7eb144f59b548dcab4831bbf9df14cea45775985))
* **agentic:** add parent issue orchestrator ([#51](https://github.com/polymetrics-ai/cli/issues/51)) ([484c3dc](https://github.com/polymetrics-ai/cli/commit/484c3dcad57e8e5c099e1ad1e85438006f39167a))
* complete connector architecture v2 migration ([#29](https://github.com/polymetrics-ai/cli/issues/29)) ([605b006](https://github.com/polymetrics-ai/cli/commit/605b006e5aa1adae697d5b7dd26ec485c570c250))
* **config:** add typed Viper configuration ([#441](https://github.com/polymetrics-ai/cli/issues/441)) ([d863025](https://github.com/polymetrics-ai/cli/commit/d863025406e104d5161f56fd8a3e911df68518e3))
* **connectors:** add config-driven bahmni-docker clinical EMR connector ([cefb6b1](https://github.com/polymetrics-ai/cli/commit/cefb6b17953dac562c3674da4a4082749ae2e12f)), closes [#516](https://github.com/polymetrics-ai/cli/issues/516) [#517](https://github.com/polymetrics-ai/cli/issues/517) [#518](https://github.com/polymetrics-ai/cli/issues/518) [#519](https://github.com/polymetrics-ai/cli/issues/519) [#520](https://github.com/polymetrics-ai/cli/issues/520) [#521](https://github.com/polymetrics-ai/cli/issues/521) [#522](https://github.com/polymetrics-ai/cli/issues/522) [#523](https://github.com/polymetrics-ai/cli/issues/523) [#524](https://github.com/polymetrics-ai/cli/issues/524) [#525](https://github.com/polymetrics-ai/cli/issues/525) [#526](https://github.com/polymetrics-ai/cli/issues/526)
* **github:** complete CLI parity roadmap ([#49](https://github.com/polymetrics-ai/cli/issues/49)) ([ae01c3f](https://github.com/polymetrics-ai/cli/commit/ae01c3f962fe089fc26e274fba1f9bbad540f7dd))
* **gong:** complete CLI parity surface ([#232](https://github.com/polymetrics-ai/cli/issues/232)) ([873cd7b](https://github.com/polymetrics-ai/cli/commit/873cd7b251f70c4a35a607a0d4e86051ea0fbd15))
* **pi:** autonomous resumable multi-model orchestration loop ([#272](https://github.com/polymetrics-ai/cli/issues/272)) ([f356bea](https://github.com/polymetrics-ai/cli/commit/f356bea11c04e094cc9ce7dfcad59080d392a1db)), closes [#271](https://github.com/polymetrics-ai/cli/issues/271)
* **pi:** Claude-orchestrated driver + Shepherd validator layer ([#276](https://github.com/polymetrics-ai/cli/issues/276)) ([cab8f3d](https://github.com/polymetrics-ai/cli/commit/cab8f3dfeb954a45db8d567bb87c5eec6c4b034e))
* **pi:** Codex-only Shepherd-supervised driver + working subagent stack ([38debf1](https://github.com/polymetrics-ai/cli/commit/38debf1f7ae9dfe204be7cffb7e8e90f62e71cda))
* **pi:** web-research stage + explicit PR stages for the autonomous loop ([#274](https://github.com/polymetrics-ai/cli/issues/274)) ([12a9e12](https://github.com/polymetrics-ai/cli/commit/12a9e1249efab6760215d389379d43b186deebe7)), closes [#273](https://github.com/polymetrics-ai/cli/issues/273) [#271](https://github.com/polymetrics-ai/cli/issues/271)
* **release:** add binary release automation ([d63b9d5](https://github.com/polymetrics-ai/cli/commit/d63b9d5b0f9408bfa773b69a3a5676a1bc87e53b))
* **website:** add blog release and SEO surfaces ([e0f5284](https://github.com/polymetrics-ai/cli/commit/e0f5284ff1ea953febf1b6f5d48e17450a5a5721))
* **website:** add creator profile links ([9592199](https://github.com/polymetrics-ai/cli/commit/959219919ff5a2d1b70bf064f966613410848e77))
* **website:** add creator social icons ([795abc5](https://github.com/polymetrics-ai/cli/commit/795abc5e14b86fe44884685dd4c152554be76592))
* **website:** refresh logo and typography ([6035064](https://github.com/polymetrics-ai/cli/commit/603506456c2834cb7f7d3c18a104431c6aa2b20c))
* **website:** ship blog annotations and reader discussions ([#346](https://github.com/polymetrics-ai/cli/issues/346)) ([c3e6244](https://github.com/polymetrics-ai/cli/commit/c3e624486031008b6017ab758d0226c78b3019f7))


### Bug Fixes

* **ci:** stop Claude review cancelling itself via concurrency ([#270](https://github.com/polymetrics-ai/cli/issues/270)) ([d864aa9](https://github.com/polymetrics-ai/cli/commit/d864aa93e8692676442e76ab65ade4b85d63da4e)), closes [#269](https://github.com/polymetrics-ai/cli/issues/269)
* **cli:** expose Bahmni command help parity ([d4e9a22](https://github.com/polymetrics-ai/cli/commit/d4e9a228cdb342b8f26213e33a0162d6a812ac64))
* **cli:** preserve Gong behavior in Cobra foundation ([a07d3f7](https://github.com/polymetrics-ai/cli/commit/a07d3f7b73f97fcfbd4048017c7c6283999d4b9b))
* **connectors:** freeze verified Bahmni write scope ([902fd94](https://github.com/polymetrics-ai/cli/commit/902fd9403c16958816ee9b744646058cc5cbdc31))
* **connectors:** redact Bahmni write error identifiers ([103152f](https://github.com/polymetrics-ai/cli/commit/103152fe672f5db2241ff1a01405b54b9c71a1b9))
* **connectors:** rename bahmni connector and close parity gaps ([c231d96](https://github.com/polymetrics-ai/cli/commit/c231d9653d99b54e2e4419f4a3afbeeb6db64708)), closes [#516](https://github.com/polymetrics-ai/cli/issues/516)
* **pi:** kill stale Shepherd turns with live children ([fca27dd](https://github.com/polymetrics-ai/cli/commit/fca27ddb670ec741d03dbc8704de66fd4368bd1a))
* **pi:** use session event time for stall guard ([6f6b3e4](https://github.com/polymetrics-ai/cli/commit/6f6b3e4cad59168f71e4acd52684799dbd0fb86b)), closes [#354](https://github.com/polymetrics-ai/cli/issues/354)
* **readme:** replace retired report badge ([29d7088](https://github.com/polymetrics-ai/cli/commit/29d708895a6803f944d3bf53d8e456ec8468c63e))
* **readme:** use profile card asset ([c18b43c](https://github.com/polymetrics-ai/cli/commit/c18b43cb25703000fb9e9f752f7fb2a020d06188))
* remove crontab schedules reliably ([#20](https://github.com/polymetrics-ai/cli/issues/20)) ([6314ad4](https://github.com/polymetrics-ai/cli/commit/6314ad498c1852a1bde36db905f681100d784e5a))
* **website:** force https redirects ([82cb4c7](https://github.com/polymetrics-ai/cli/commit/82cb4c76456018081bd0ec80b9dbf11dc2dd8c4f))


### Miscellaneous Chores

* release 0.1.0 ([#543](https://github.com/polymetrics-ai/cli/issues/543)) ([871b103](https://github.com/polymetrics-ai/cli/commit/871b1035a58cf8b69308c201ec8413a1e21e0f8c))
