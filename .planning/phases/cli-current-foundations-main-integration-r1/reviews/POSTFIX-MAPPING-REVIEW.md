---
phase: cli-current-foundations-main-integration-r1
reviewed: 2026-08-21T02:59:22Z
depth: deep
diff_base: e62ae21d428f0d27225f9bff564dc2cd797f6b65
source_sha: 8a8a866ff6d5282c28bda12acceed8a624218f01
files_reviewed: 774
files_reviewed_list:
  - "AGENTS.md"
  - "REVIEW-CONVERGENCE.md"
  - "cmd/connectorgen/batch_test.go"
  - "cmd/connectorgen/certificationcandidates_test.go"
  - "cmd/connectorgen/certificationmatrix.go"
  - "cmd/connectorgen/certificationsweep_test.go"
  - "cmd/connectorgen/evidencegate.go"
  - "cmd/connectorgen/evidencegate_test.go"
  - "cmd/connectorgen/main.go"
  - "cmd/connectorgen/main_test.go"
  - "cmd/connectorgen/paramsimport.go"
  - "cmd/connectorgen/paramsimport_test.go"
  - "cmd/connectorgen/sourceimport.go"
  - "cmd/connectorgen/sourceimport_test.go"
  - "cmd/connectorgen/sourceprojection.go"
  - "cmd/connectorgen/sourceprojection_test.go"
  - "cmd/connectorgen/surfacesync.go"
  - "cmd/connectorgen/testdata/sourceimport/alpha/alpha-openapi.yaml"
  - "cmd/connectorgen/testdata/sourceimport/alpha/alpha-operation-source-lock.json"
  - "cmd/connectorgen/testdata/sourceimport/beta/beta-openapi.json"
  - "cmd/connectorgen/testdata/sourceimport/beta/beta-operation-source-lock.json"
  - "cmd/connectorgen/testdata/sourceimport/invalid/ambiguous-request.json"
  - "cmd/connectorgen/testdata/sourceimport/invalid/callback-route.json"
  - "cmd/connectorgen/testdata/sourceimport/invalid/cyclic-ref.json"
  - "cmd/connectorgen/testdata/sourceimport/invalid/deep-reference.json"
  - "cmd/connectorgen/testdata/sourceimport/invalid/duplicate-id.json"
  - "cmd/connectorgen/testdata/sourceimport/invalid/external-ref.json"
  - "cmd/connectorgen/testdata/sourceimport/invalid/invalid-relative-path.json"
  - "cmd/connectorgen/testdata/sourceimport/invalid/large-schema.json"
  - "cmd/connectorgen/testdata/sourceimport/invalid/many-operations.json"
  - "cmd/connectorgen/testdata/sourceimport/invalid/many-references.json"
  - "cmd/connectorgen/testdata/sourceimport/invalid/missing-additional-properties.json"
  - "cmd/connectorgen/testdata/sourceimport/invalid/missing-path-parameter.json"
  - "cmd/connectorgen/testdata/sourceimport/invalid/multiple-documents.yaml"
  - "cmd/connectorgen/testdata/sourceimport/invalid/unbounded-request.json"
  - "cmd/connectorgen/testdata/sourceimport/invalid/unresolved-ref.json"
  - "cmd/connectorgen/testdata/sourceimport/invalid/unsupported-encoding.json"
  - "cmd/connectorgen/testdata/sourceimport/invalid/whitespace-path.json"
  - "cmd/connectorgen/testdata/sourceimport/supported/swagger2-body.json"
  - "cmd/connectorgen/validate.go"
  - "data/cli-current-foundations-main-integration-r1/evidence-manifest.json"
  - "data/cli-current-foundations-main-integration-r1/input-manifest.json"
  - "data/cli-current-foundations-main-integration-r1/report.md"
  - "docs/architecture/connector-operation-kernel.md"
  - "docs/cli/connections.md"
  - "docs/cli/connectors.md"
  - "docs/cli/etl.md"
  - "docs/cli/reverse.md"
  - "docs/connectors/github/MANUAL.md"
  - "docs/connectors/github/SKILL.md"
  - "docs/connectors/gmail/MANUAL.md"
  - "docs/connectors/gmail/SKILL.md"
  - "docs/connectors/gorgias/MANUAL.md"
  - "docs/connectors/gorgias/SKILL.md"
  - "docs/direct-read-pages-and-parameters.md"
  - "docs/migration/conventions.md"
  - "docs/skills/pm-100ms/SKILL.md"
  - "docs/skills/pm-7shifts/SKILL.md"
  - "docs/skills/pm-activecampaign/SKILL.md"
  - "docs/skills/pm-acuity-scheduling/SKILL.md"
  - "docs/skills/pm-adjust/SKILL.md"
  - "docs/skills/pm-adobe-commerce-magento/SKILL.md"
  - "docs/skills/pm-agilecrm/SKILL.md"
  - "docs/skills/pm-aha/SKILL.md"
  - "docs/skills/pm-aircall/SKILL.md"
  - "docs/skills/pm-airtable/SKILL.md"
  - "docs/skills/pm-akeneo/SKILL.md"
  - "docs/skills/pm-algolia/SKILL.md"
  - "docs/skills/pm-alpaca-broker-api/SKILL.md"
  - "docs/skills/pm-alpha-vantage/SKILL.md"
  - "docs/skills/pm-amazon-ads/SKILL.md"
  - "docs/skills/pm-amazon-seller-partner/SKILL.md"
  - "docs/skills/pm-amazon-sqs/SKILL.md"
  - "docs/skills/pm-amplitude/SKILL.md"
  - "docs/skills/pm-apify-dataset/SKILL.md"
  - "docs/skills/pm-appcues/SKILL.md"
  - "docs/skills/pm-appfigures/SKILL.md"
  - "docs/skills/pm-appfollow/SKILL.md"
  - "docs/skills/pm-apple-search-ads/SKILL.md"
  - "docs/skills/pm-appsflyer/SKILL.md"
  - "docs/skills/pm-apptivo/SKILL.md"
  - "docs/skills/pm-asana/SKILL.md"
  - "docs/skills/pm-ashby/SKILL.md"
  - "docs/skills/pm-assemblyai/SKILL.md"
  - "docs/skills/pm-auth0/SKILL.md"
  - "docs/skills/pm-aviationstack/SKILL.md"
  - "docs/skills/pm-avni/SKILL.md"
  - "docs/skills/pm-awin-advertiser/SKILL.md"
  - "docs/skills/pm-aws-cloudtrail/SKILL.md"
  - "docs/skills/pm-babelforce/SKILL.md"
  - "docs/skills/pm-bahmni/SKILL.md"
  - "docs/skills/pm-bamboo-hr/SKILL.md"
  - "docs/skills/pm-basecamp/SKILL.md"
  - "docs/skills/pm-beamer/SKILL.md"
  - "docs/skills/pm-bigmailer/SKILL.md"
  - "docs/skills/pm-bing-ads/SKILL.md"
  - "docs/skills/pm-bitbucket/SKILL.md"
  - "docs/skills/pm-bitly/SKILL.md"
  - "docs/skills/pm-blogger/SKILL.md"
  - "docs/skills/pm-bluetally/SKILL.md"
  - "docs/skills/pm-boldsign/SKILL.md"
  - "docs/skills/pm-box-data-extract/SKILL.md"
  - "docs/skills/pm-box/SKILL.md"
  - "docs/skills/pm-braintree/SKILL.md"
  - "docs/skills/pm-braze/SKILL.md"
  - "docs/skills/pm-breezometer/SKILL.md"
  - "docs/skills/pm-breezy-hr/SKILL.md"
  - "docs/skills/pm-brevo/SKILL.md"
  - "docs/skills/pm-brex/SKILL.md"
  - "docs/skills/pm-bugsnag/SKILL.md"
  - "docs/skills/pm-buildkite/SKILL.md"
  - "docs/skills/pm-bunny-inc/SKILL.md"
  - "docs/skills/pm-buzzsprout/SKILL.md"
  - "docs/skills/pm-cal-com/SKILL.md"
  - "docs/skills/pm-calendly/SKILL.md"
  - "docs/skills/pm-callrail/SKILL.md"
  - "docs/skills/pm-campaign-monitor/SKILL.md"
  - "docs/skills/pm-campayn/SKILL.md"
  - "docs/skills/pm-canny/SKILL.md"
  - "docs/skills/pm-capsule-crm/SKILL.md"
  - "docs/skills/pm-captain-data/SKILL.md"
  - "docs/skills/pm-care-quality-commission/SKILL.md"
  - "docs/skills/pm-cart/SKILL.md"
  - "docs/skills/pm-castor-edc/SKILL.md"
  - "docs/skills/pm-chameleon/SKILL.md"
  - "docs/skills/pm-chargebee/SKILL.md"
  - "docs/skills/pm-chargedesk/SKILL.md"
  - "docs/skills/pm-chargify/SKILL.md"
  - "docs/skills/pm-chartmogul/SKILL.md"
  - "docs/skills/pm-chatwoot/SKILL.md"
  - "docs/skills/pm-chift/SKILL.md"
  - "docs/skills/pm-churnkey/SKILL.md"
  - "docs/skills/pm-cimis/SKILL.md"
  - "docs/skills/pm-cin7/SKILL.md"
  - "docs/skills/pm-circa/SKILL.md"
  - "docs/skills/pm-circleci/SKILL.md"
  - "docs/skills/pm-cisco-meraki/SKILL.md"
  - "docs/skills/pm-clarif-ai/SKILL.md"
  - "docs/skills/pm-clazar/SKILL.md"
  - "docs/skills/pm-clickup-api/SKILL.md"
  - "docs/skills/pm-clockify/SKILL.md"
  - "docs/skills/pm-clockodo/SKILL.md"
  - "docs/skills/pm-close-com/SKILL.md"
  - "docs/skills/pm-cloudbeds/SKILL.md"
  - "docs/skills/pm-coassemble/SKILL.md"
  - "docs/skills/pm-coda/SKILL.md"
  - "docs/skills/pm-codefresh/SKILL.md"
  - "docs/skills/pm-coin-api/SKILL.md"
  - "docs/skills/pm-coingecko-coins/SKILL.md"
  - "docs/skills/pm-coinmarketcap/SKILL.md"
  - "docs/skills/pm-commcare/SKILL.md"
  - "docs/skills/pm-commercetools/SKILL.md"
  - "docs/skills/pm-concord/SKILL.md"
  - "docs/skills/pm-configcat/SKILL.md"
  - "docs/skills/pm-confluence/SKILL.md"
  - "docs/skills/pm-convertkit/SKILL.md"
  - "docs/skills/pm-convex/SKILL.md"
  - "docs/skills/pm-copper/SKILL.md"
  - "docs/skills/pm-countercyclical/SKILL.md"
  - "docs/skills/pm-crisp/SKILL.md"
  - "docs/skills/pm-criteo-marketing/SKILL.md"
  - "docs/skills/pm-customer-io/SKILL.md"
  - "docs/skills/pm-customerly/SKILL.md"
  - "docs/skills/pm-datadog/SKILL.md"
  - "docs/skills/pm-datascope/SKILL.md"
  - "docs/skills/pm-dbt/SKILL.md"
  - "docs/skills/pm-defillama/SKILL.md"
  - "docs/skills/pm-delighted/SKILL.md"
  - "docs/skills/pm-deputy/SKILL.md"
  - "docs/skills/pm-devin-ai/SKILL.md"
  - "docs/skills/pm-ding-connect/SKILL.md"
  - "docs/skills/pm-discord/SKILL.md"
  - "docs/skills/pm-dixa/SKILL.md"
  - "docs/skills/pm-dockerhub/SKILL.md"
  - "docs/skills/pm-docuseal/SKILL.md"
  - "docs/skills/pm-dolibarr/SKILL.md"
  - "docs/skills/pm-dremio/SKILL.md"
  - "docs/skills/pm-drift/SKILL.md"
  - "docs/skills/pm-drip/SKILL.md"
  - "docs/skills/pm-dropbox-sign/SKILL.md"
  - "docs/skills/pm-dwolla/SKILL.md"
  - "docs/skills/pm-dynamodb/SKILL.md"
  - "docs/skills/pm-e-conomic/SKILL.md"
  - "docs/skills/pm-easypost/SKILL.md"
  - "docs/skills/pm-easypromos/SKILL.md"
  - "docs/skills/pm-ebay-finance/SKILL.md"
  - "docs/skills/pm-ebay-fulfillment/SKILL.md"
  - "docs/skills/pm-elasticemail/SKILL.md"
  - "docs/skills/pm-elasticsearch/SKILL.md"
  - "docs/skills/pm-emailoctopus/SKILL.md"
  - "docs/skills/pm-employment-hero/SKILL.md"
  - "docs/skills/pm-encharge/SKILL.md"
  - "docs/skills/pm-etl/SKILL.md"
  - "docs/skills/pm-eventbrite/SKILL.md"
  - "docs/skills/pm-eventee/SKILL.md"
  - "docs/skills/pm-eventzilla/SKILL.md"
  - "docs/skills/pm-everhour/SKILL.md"
  - "docs/skills/pm-exchange-rates/SKILL.md"
  - "docs/skills/pm-ezofficeinventory/SKILL.md"
  - "docs/skills/pm-facebook-marketing/SKILL.md"
  - "docs/skills/pm-facebook-pages/SKILL.md"
  - "docs/skills/pm-factorial/SKILL.md"
  - "docs/skills/pm-faker/SKILL.md"
  - "docs/skills/pm-fastbill/SKILL.md"
  - "docs/skills/pm-fastly/SKILL.md"
  - "docs/skills/pm-feishu/SKILL.md"
  - "docs/skills/pm-fillout/SKILL.md"
  - "docs/skills/pm-finage/SKILL.md"
  - "docs/skills/pm-financial-modelling/SKILL.md"
  - "docs/skills/pm-finnhub/SKILL.md"
  - "docs/skills/pm-finnworlds/SKILL.md"
  - "docs/skills/pm-firehydrant/SKILL.md"
  - "docs/skills/pm-fleetio/SKILL.md"
  - "docs/skills/pm-flexmail/SKILL.md"
  - "docs/skills/pm-flexport/SKILL.md"
  - "docs/skills/pm-float/SKILL.md"
  - "docs/skills/pm-flowlu/SKILL.md"
  - "docs/skills/pm-formbricks/SKILL.md"
  - "docs/skills/pm-free-agent-connector/SKILL.md"
  - "docs/skills/pm-freightview/SKILL.md"
  - "docs/skills/pm-freshbooks/SKILL.md"
  - "docs/skills/pm-freshcaller/SKILL.md"
  - "docs/skills/pm-freshchat/SKILL.md"
  - "docs/skills/pm-freshdesk/SKILL.md"
  - "docs/skills/pm-freshsales/SKILL.md"
  - "docs/skills/pm-freshservice/SKILL.md"
  - "docs/skills/pm-front/SKILL.md"
  - "docs/skills/pm-fulcrum/SKILL.md"
  - "docs/skills/pm-fullstory/SKILL.md"
  - "docs/skills/pm-gainsight-px/SKILL.md"
  - "docs/skills/pm-genesys/SKILL.md"
  - "docs/skills/pm-getgist/SKILL.md"
  - "docs/skills/pm-getlago/SKILL.md"
  - "docs/skills/pm-giphy/SKILL.md"
  - "docs/skills/pm-gitbook/SKILL.md"
  - "docs/skills/pm-github/SKILL.md"
  - "docs/skills/pm-gitlab/SKILL.md"
  - "docs/skills/pm-glassfrog/SKILL.md"
  - "docs/skills/pm-gmail/SKILL.md"
  - "docs/skills/pm-gnews/SKILL.md"
  - "docs/skills/pm-gocardless/SKILL.md"
  - "docs/skills/pm-goldcast/SKILL.md"
  - "docs/skills/pm-gologin/SKILL.md"
  - "docs/skills/pm-gong/SKILL.md"
  - "docs/skills/pm-google-ads/SKILL.md"
  - "docs/skills/pm-google-analytics-data-api/SKILL.md"
  - "docs/skills/pm-google-calendar/SKILL.md"
  - "docs/skills/pm-google-classroom/SKILL.md"
  - "docs/skills/pm-google-directory/SKILL.md"
  - "docs/skills/pm-google-forms/SKILL.md"
  - "docs/skills/pm-google-pagespeed-insights/SKILL.md"
  - "docs/skills/pm-google-search-console/SKILL.md"
  - "docs/skills/pm-google-tasks/SKILL.md"
  - "docs/skills/pm-google-webfonts/SKILL.md"
  - "docs/skills/pm-gorgias/SKILL.md"
  - "docs/skills/pm-grafana/SKILL.md"
  - "docs/skills/pm-granola/SKILL.md"
  - "docs/skills/pm-greenhouse/SKILL.md"
  - "docs/skills/pm-gridly/SKILL.md"
  - "docs/skills/pm-guru/SKILL.md"
  - "docs/skills/pm-gutendex/SKILL.md"
  - "docs/skills/pm-harness/SKILL.md"
  - "docs/skills/pm-harvest/SKILL.md"
  - "docs/skills/pm-height/SKILL.md"
  - "docs/skills/pm-hellobaton/SKILL.md"
  - "docs/skills/pm-help-scout/SKILL.md"
  - "docs/skills/pm-hibob/SKILL.md"
  - "docs/skills/pm-high-level/SKILL.md"
  - "docs/skills/pm-hoorayhr/SKILL.md"
  - "docs/skills/pm-hubplanner/SKILL.md"
  - "docs/skills/pm-hubspot/SKILL.md"
  - "docs/skills/pm-hugging-face-datasets/SKILL.md"
  - "docs/skills/pm-humanitix/SKILL.md"
  - "docs/skills/pm-huntr/SKILL.md"
  - "docs/skills/pm-illumina-basespace/SKILL.md"
  - "docs/skills/pm-imagga/SKILL.md"
  - "docs/skills/pm-incident-io/SKILL.md"
  - "docs/skills/pm-inflowinventory/SKILL.md"
  - "docs/skills/pm-insightful/SKILL.md"
  - "docs/skills/pm-insightly/SKILL.md"
  - "docs/skills/pm-instagram/SKILL.md"
  - "docs/skills/pm-instatus/SKILL.md"
  - "docs/skills/pm-intercom/SKILL.md"
  - "docs/skills/pm-interzoid/SKILL.md"
  - "docs/skills/pm-intruder/SKILL.md"
  - "docs/skills/pm-invoiced/SKILL.md"
  - "docs/skills/pm-invoiceninja/SKILL.md"
  - "docs/skills/pm-ip2whois/SKILL.md"
  - "docs/skills/pm-iterable/SKILL.md"
  - "docs/skills/pm-jamf-pro/SKILL.md"
  - "docs/skills/pm-jira/SKILL.md"
  - "docs/skills/pm-jobnimbus/SKILL.md"
  - "docs/skills/pm-jotform/SKILL.md"
  - "docs/skills/pm-judge-me-reviews/SKILL.md"
  - "docs/skills/pm-just-sift/SKILL.md"
  - "docs/skills/pm-justcall/SKILL.md"
  - "docs/skills/pm-k6-cloud/SKILL.md"
  - "docs/skills/pm-katana/SKILL.md"
  - "docs/skills/pm-keka/SKILL.md"
  - "docs/skills/pm-kisi/SKILL.md"
  - "docs/skills/pm-kissmetrics/SKILL.md"
  - "docs/skills/pm-klarna/SKILL.md"
  - "docs/skills/pm-klaus-api/SKILL.md"
  - "docs/skills/pm-klaviyo/SKILL.md"
  - "docs/skills/pm-kyriba/SKILL.md"
  - "docs/skills/pm-kyve/SKILL.md"
  - "docs/skills/pm-launchdarkly/SKILL.md"
  - "docs/skills/pm-leadfeeder/SKILL.md"
  - "docs/skills/pm-lemlist/SKILL.md"
  - "docs/skills/pm-less-annoying-crm/SKILL.md"
  - "docs/skills/pm-lever-hiring/SKILL.md"
  - "docs/skills/pm-lightspeed-retail/SKILL.md"
  - "docs/skills/pm-linear/SKILL.md"
  - "docs/skills/pm-linkedin-ads/SKILL.md"
  - "docs/skills/pm-linkedin-pages/SKILL.md"
  - "docs/skills/pm-linkrunner/SKILL.md"
  - "docs/skills/pm-lob/SKILL.md"
  - "docs/skills/pm-lokalise/SKILL.md"
  - "docs/skills/pm-looker/SKILL.md"
  - "docs/skills/pm-luma/SKILL.md"
  - "docs/skills/pm-mailchimp/SKILL.md"
  - "docs/skills/pm-mailerlite/SKILL.md"
  - "docs/skills/pm-mailersend/SKILL.md"
  - "docs/skills/pm-mailgun/SKILL.md"
  - "docs/skills/pm-mailjet-mail/SKILL.md"
  - "docs/skills/pm-mailjet-sms/SKILL.md"
  - "docs/skills/pm-mailosaur/SKILL.md"
  - "docs/skills/pm-mailtrap/SKILL.md"
  - "docs/skills/pm-mantle/SKILL.md"
  - "docs/skills/pm-marketo/SKILL.md"
  - "docs/skills/pm-marketstack/SKILL.md"
  - "docs/skills/pm-mendeley/SKILL.md"
  - "docs/skills/pm-mention/SKILL.md"
  - "docs/skills/pm-mercado-ads/SKILL.md"
  - "docs/skills/pm-merge/SKILL.md"
  - "docs/skills/pm-metabase/SKILL.md"
  - "docs/skills/pm-metricool/SKILL.md"
  - "docs/skills/pm-microsoft-dataverse/SKILL.md"
  - "docs/skills/pm-microsoft-entra-id/SKILL.md"
  - "docs/skills/pm-microsoft-lists/SKILL.md"
  - "docs/skills/pm-microsoft-teams/SKILL.md"
  - "docs/skills/pm-miro/SKILL.md"
  - "docs/skills/pm-missive/SKILL.md"
  - "docs/skills/pm-mixmax/SKILL.md"
  - "docs/skills/pm-mixpanel/SKILL.md"
  - "docs/skills/pm-mode/SKILL.md"
  - "docs/skills/pm-monday/SKILL.md"
  - "docs/skills/pm-mux/SKILL.md"
  - "docs/skills/pm-my-hours/SKILL.md"
  - "docs/skills/pm-mysql/SKILL.md"
  - "docs/skills/pm-n8n/SKILL.md"
  - "docs/skills/pm-nasa/SKILL.md"
  - "docs/skills/pm-navan/SKILL.md"
  - "docs/skills/pm-nebius-ai/SKILL.md"
  - "docs/skills/pm-netsuite/SKILL.md"
  - "docs/skills/pm-news-api/SKILL.md"
  - "docs/skills/pm-newsdata-io/SKILL.md"
  - "docs/skills/pm-newsdata/SKILL.md"
  - "docs/skills/pm-nexiopay/SKILL.md"
  - "docs/skills/pm-nexus-datasets/SKILL.md"
  - "docs/skills/pm-ninjaone-rmm/SKILL.md"
  - "docs/skills/pm-nocrm/SKILL.md"
  - "docs/skills/pm-northpass-lms/SKILL.md"
  - "docs/skills/pm-notion/SKILL.md"
  - "docs/skills/pm-nutshell/SKILL.md"
  - "docs/skills/pm-nylas/SKILL.md"
  - "docs/skills/pm-nytimes/SKILL.md"
  - "docs/skills/pm-okta/SKILL.md"
  - "docs/skills/pm-omnisend/SKILL.md"
  - "docs/skills/pm-oncehub/SKILL.md"
  - "docs/skills/pm-onepagecrm/SKILL.md"
  - "docs/skills/pm-onesignal/SKILL.md"
  - "docs/skills/pm-onfleet/SKILL.md"
  - "docs/skills/pm-open-data-dc/SKILL.md"
  - "docs/skills/pm-open-exchange-rates/SKILL.md"
  - "docs/skills/pm-openaq/SKILL.md"
  - "docs/skills/pm-openfda/SKILL.md"
  - "docs/skills/pm-openweather/SKILL.md"
  - "docs/skills/pm-opinion-stage/SKILL.md"
  - "docs/skills/pm-opsgenie/SKILL.md"
  - "docs/skills/pm-opuswatch/SKILL.md"
  - "docs/skills/pm-orb/SKILL.md"
  - "docs/skills/pm-oura/SKILL.md"
  - "docs/skills/pm-outbrain-amplify/SKILL.md"
  - "docs/skills/pm-outlook/SKILL.md"
  - "docs/skills/pm-outreach/SKILL.md"
  - "docs/skills/pm-oveit/SKILL.md"
  - "docs/skills/pm-pabbly-subscriptions-billing/SKILL.md"
  - "docs/skills/pm-paddle/SKILL.md"
  - "docs/skills/pm-pagerduty/SKILL.md"
  - "docs/skills/pm-pandadoc/SKILL.md"
  - "docs/skills/pm-paperform/SKILL.md"
  - "docs/skills/pm-papersign/SKILL.md"
  - "docs/skills/pm-pardot/SKILL.md"
  - "docs/skills/pm-partnerize/SKILL.md"
  - "docs/skills/pm-partnerstack/SKILL.md"
  - "docs/skills/pm-payfit/SKILL.md"
  - "docs/skills/pm-paypal-transaction/SKILL.md"
  - "docs/skills/pm-paystack/SKILL.md"
  - "docs/skills/pm-pendo/SKILL.md"
  - "docs/skills/pm-pennylane/SKILL.md"
  - "docs/skills/pm-perigon/SKILL.md"
  - "docs/skills/pm-perk/SKILL.md"
  - "docs/skills/pm-persistiq/SKILL.md"
  - "docs/skills/pm-persona/SKILL.md"
  - "docs/skills/pm-pexels-api/SKILL.md"
  - "docs/skills/pm-phyllo/SKILL.md"
  - "docs/skills/pm-picqer/SKILL.md"
  - "docs/skills/pm-pingdom/SKILL.md"
  - "docs/skills/pm-pinterest/SKILL.md"
  - "docs/skills/pm-pipedrive/SKILL.md"
  - "docs/skills/pm-pipeliner/SKILL.md"
  - "docs/skills/pm-pivotal-tracker/SKILL.md"
  - "docs/skills/pm-piwik/SKILL.md"
  - "docs/skills/pm-plaid/SKILL.md"
  - "docs/skills/pm-planhat/SKILL.md"
  - "docs/skills/pm-plausible/SKILL.md"
  - "docs/skills/pm-pocket/SKILL.md"
  - "docs/skills/pm-pokeapi/SKILL.md"
  - "docs/skills/pm-polygon-stock-api/SKILL.md"
  - "docs/skills/pm-poplar/SKILL.md"
  - "docs/skills/pm-postgres/SKILL.md"
  - "docs/skills/pm-posthog/SKILL.md"
  - "docs/skills/pm-postmarkapp/SKILL.md"
  - "docs/skills/pm-prestashop/SKILL.md"
  - "docs/skills/pm-pretix/SKILL.md"
  - "docs/skills/pm-primetric/SKILL.md"
  - "docs/skills/pm-printify/SKILL.md"
  - "docs/skills/pm-productboard/SKILL.md"
  - "docs/skills/pm-productive/SKILL.md"
  - "docs/skills/pm-public-apis/SKILL.md"
  - "docs/skills/pm-pylon/SKILL.md"
  - "docs/skills/pm-pypi/SKILL.md"
  - "docs/skills/pm-qonto/SKILL.md"
  - "docs/skills/pm-qualaroo/SKILL.md"
  - "docs/skills/pm-quickbooks/SKILL.md"
  - "docs/skills/pm-railz/SKILL.md"
  - "docs/skills/pm-rd-station-marketing/SKILL.md"
  - "docs/skills/pm-recharge/SKILL.md"
  - "docs/skills/pm-recreation/SKILL.md"
  - "docs/skills/pm-recruitee/SKILL.md"
  - "docs/skills/pm-recurly/SKILL.md"
  - "docs/skills/pm-reddit/SKILL.md"
  - "docs/skills/pm-referralhero/SKILL.md"
  - "docs/skills/pm-rentcast/SKILL.md"
  - "docs/skills/pm-repairshopr/SKILL.md"
  - "docs/skills/pm-reply-io/SKILL.md"
  - "docs/skills/pm-retailexpress-by-maropost/SKILL.md"
  - "docs/skills/pm-retently/SKILL.md"
  - "docs/skills/pm-revenuecat/SKILL.md"
  - "docs/skills/pm-revolut-merchant/SKILL.md"
  - "docs/skills/pm-ringcentral/SKILL.md"
  - "docs/skills/pm-rki-covid/SKILL.md"
  - "docs/skills/pm-rocket-chat/SKILL.md"
  - "docs/skills/pm-rocketlane/SKILL.md"
  - "docs/skills/pm-rollbar/SKILL.md"
  - "docs/skills/pm-rootly/SKILL.md"
  - "docs/skills/pm-rss/SKILL.md"
  - "docs/skills/pm-ruddr/SKILL.md"
  - "docs/skills/pm-safetyculture/SKILL.md"
  - "docs/skills/pm-sage-hr/SKILL.md"
  - "docs/skills/pm-salesflare/SKILL.md"
  - "docs/skills/pm-salesforce/SKILL.md"
  - "docs/skills/pm-salesloft/SKILL.md"
  - "docs/skills/pm-sap-fieldglass/SKILL.md"
  - "docs/skills/pm-savvycal/SKILL.md"
  - "docs/skills/pm-scryfall/SKILL.md"
  - "docs/skills/pm-searxng/SKILL.md"
  - "docs/skills/pm-secoda/SKILL.md"
  - "docs/skills/pm-segment/SKILL.md"
  - "docs/skills/pm-sendgrid/SKILL.md"
  - "docs/skills/pm-sendinblue/SKILL.md"
  - "docs/skills/pm-sendowl/SKILL.md"
  - "docs/skills/pm-sendpulse/SKILL.md"
  - "docs/skills/pm-senseforce/SKILL.md"
  - "docs/skills/pm-sentry/SKILL.md"
  - "docs/skills/pm-serpstat/SKILL.md"
  - "docs/skills/pm-service-now/SKILL.md"
  - "docs/skills/pm-sharepoint-lists-enterprise/SKILL.md"
  - "docs/skills/pm-sharetribe/SKILL.md"
  - "docs/skills/pm-shippo/SKILL.md"
  - "docs/skills/pm-shipstation/SKILL.md"
  - "docs/skills/pm-shopwired/SKILL.md"
  - "docs/skills/pm-shortcut/SKILL.md"
  - "docs/skills/pm-shortio/SKILL.md"
  - "docs/skills/pm-shutterstock/SKILL.md"
  - "docs/skills/pm-sigma-computing/SKILL.md"
  - "docs/skills/pm-signnow/SKILL.md"
  - "docs/skills/pm-simfin/SKILL.md"
  - "docs/skills/pm-simplecast/SKILL.md"
  - "docs/skills/pm-simplesat/SKILL.md"
  - "docs/skills/pm-slack/SKILL.md"
  - "docs/skills/pm-smaily/SKILL.md"
  - "docs/skills/pm-smartengage/SKILL.md"
  - "docs/skills/pm-smartreach/SKILL.md"
  - "docs/skills/pm-smartsheets/SKILL.md"
  - "docs/skills/pm-smartwaiver/SKILL.md"
  - "docs/skills/pm-snapchat-marketing/SKILL.md"
  - "docs/skills/pm-solarwinds-service-desk/SKILL.md"
  - "docs/skills/pm-sonar-cloud/SKILL.md"
  - "docs/skills/pm-spacex-api/SKILL.md"
  - "docs/skills/pm-sparkpost/SKILL.md"
  - "docs/skills/pm-split-io/SKILL.md"
  - "docs/skills/pm-spotify-ads/SKILL.md"
  - "docs/skills/pm-spotlercrm/SKILL.md"
  - "docs/skills/pm-square/SKILL.md"
  - "docs/skills/pm-squarespace/SKILL.md"
  - "docs/skills/pm-statsig/SKILL.md"
  - "docs/skills/pm-statuspage/SKILL.md"
  - "docs/skills/pm-stigg/SKILL.md"
  - "docs/skills/pm-stockdata/SKILL.md"
  - "docs/skills/pm-strava/SKILL.md"
  - "docs/skills/pm-stripe/SKILL.md"
  - "docs/skills/pm-survey-sparrow/SKILL.md"
  - "docs/skills/pm-surveycto/SKILL.md"
  - "docs/skills/pm-surveymonkey/SKILL.md"
  - "docs/skills/pm-survicate/SKILL.md"
  - "docs/skills/pm-svix/SKILL.md"
  - "docs/skills/pm-systeme/SKILL.md"
  - "docs/skills/pm-taboola/SKILL.md"
  - "docs/skills/pm-tally-prime/SKILL.md"
  - "docs/skills/pm-tally/SKILL.md"
  - "docs/skills/pm-tavus/SKILL.md"
  - "docs/skills/pm-teamtailor/SKILL.md"
  - "docs/skills/pm-teamwork/SKILL.md"
  - "docs/skills/pm-tempo/SKILL.md"
  - "docs/skills/pm-testrail/SKILL.md"
  - "docs/skills/pm-the-guardian-api/SKILL.md"
  - "docs/skills/pm-thinkific-courses/SKILL.md"
  - "docs/skills/pm-thinkific/SKILL.md"
  - "docs/skills/pm-thrive-learning/SKILL.md"
  - "docs/skills/pm-ticketmaster/SKILL.md"
  - "docs/skills/pm-tickettailor/SKILL.md"
  - "docs/skills/pm-ticktick/SKILL.md"
  - "docs/skills/pm-tiktok-marketing/SKILL.md"
  - "docs/skills/pm-timely/SKILL.md"
  - "docs/skills/pm-tinyemail/SKILL.md"
  - "docs/skills/pm-tmdb/SKILL.md"
  - "docs/skills/pm-todoist/SKILL.md"
  - "docs/skills/pm-toggl/SKILL.md"
  - "docs/skills/pm-track-pms/SKILL.md"
  - "docs/skills/pm-trello/SKILL.md"
  - "docs/skills/pm-tremendous/SKILL.md"
  - "docs/skills/pm-trustpilot/SKILL.md"
  - "docs/skills/pm-tvmaze-schedule/SKILL.md"
  - "docs/skills/pm-twelve-data/SKILL.md"
  - "docs/skills/pm-twilio-taskrouter/SKILL.md"
  - "docs/skills/pm-twilio/SKILL.md"
  - "docs/skills/pm-twitter/SKILL.md"
  - "docs/skills/pm-tyntec-sms/SKILL.md"
  - "docs/skills/pm-typeform/SKILL.md"
  - "docs/skills/pm-ubidots/SKILL.md"
  - "docs/skills/pm-unleash/SKILL.md"
  - "docs/skills/pm-uppromote/SKILL.md"
  - "docs/skills/pm-uptick/SKILL.md"
  - "docs/skills/pm-us-census/SKILL.md"
  - "docs/skills/pm-uservoice/SKILL.md"
  - "docs/skills/pm-vantage/SKILL.md"
  - "docs/skills/pm-veeqo/SKILL.md"
  - "docs/skills/pm-vercel/SKILL.md"
  - "docs/skills/pm-visma-economic/SKILL.md"
  - "docs/skills/pm-vitally/SKILL.md"
  - "docs/skills/pm-vwo/SKILL.md"
  - "docs/skills/pm-waiteraid/SKILL.md"
  - "docs/skills/pm-wasabi-stats-api/SKILL.md"
  - "docs/skills/pm-watchmode/SKILL.md"
  - "docs/skills/pm-weatherstack/SKILL.md"
  - "docs/skills/pm-web-scrapper/SKILL.md"
  - "docs/skills/pm-webflow/SKILL.md"
  - "docs/skills/pm-when-i-work/SKILL.md"
  - "docs/skills/pm-whisky-hunter/SKILL.md"
  - "docs/skills/pm-wikipedia-pageviews/SKILL.md"
  - "docs/skills/pm-woocommerce/SKILL.md"
  - "docs/skills/pm-wordpress/SKILL.md"
  - "docs/skills/pm-workable/SKILL.md"
  - "docs/skills/pm-workday-rest/SKILL.md"
  - "docs/skills/pm-workday/SKILL.md"
  - "docs/skills/pm-workflowmax/SKILL.md"
  - "docs/skills/pm-workramp/SKILL.md"
  - "docs/skills/pm-wrike/SKILL.md"
  - "docs/skills/pm-wufoo/SKILL.md"
  - "docs/skills/pm-xero/SKILL.md"
  - "docs/skills/pm-xkcd/SKILL.md"
  - "docs/skills/pm-xsolla/SKILL.md"
  - "docs/skills/pm-yahoo-finance-price/SKILL.md"
  - "docs/skills/pm-yotpo/SKILL.md"
  - "docs/skills/pm-you-need-a-budget-ynab/SKILL.md"
  - "docs/skills/pm-younium/SKILL.md"
  - "docs/skills/pm-yousign/SKILL.md"
  - "docs/skills/pm-youtube-analytics/SKILL.md"
  - "docs/skills/pm-youtube-data/SKILL.md"
  - "docs/skills/pm-zapier-supported-storage/SKILL.md"
  - "docs/skills/pm-zapsign/SKILL.md"
  - "docs/skills/pm-zendesk-chat/SKILL.md"
  - "docs/skills/pm-zendesk-sunshine/SKILL.md"
  - "docs/skills/pm-zendesk-support/SKILL.md"
  - "docs/skills/pm-zendesk-talk/SKILL.md"
  - "docs/skills/pm-zenefits/SKILL.md"
  - "docs/skills/pm-zoho-analytics-metadata-api/SKILL.md"
  - "docs/skills/pm-zoho-bigin/SKILL.md"
  - "docs/skills/pm-zoho-billing/SKILL.md"
  - "docs/skills/pm-zoho-books/SKILL.md"
  - "docs/skills/pm-zoho-campaign/SKILL.md"
  - "docs/skills/pm-zoho-desk/SKILL.md"
  - "docs/skills/pm-zoho-expense/SKILL.md"
  - "docs/skills/pm-zoho-inventory/SKILL.md"
  - "docs/skills/pm-zoho-invoice/SKILL.md"
  - "docs/skills/pm-zonka-feedback/SKILL.md"
  - "docs/skills/pm-zoom/SKILL.md"
  - "docs/skills/skills.md"
  - "docs/sync-transport-definition.md"
  - "internal/app/app.go"
  - "internal/app/authorization.go"
  - "internal/app/change_capture_dispatch_test.go"
  - "internal/app/declarative_typed_destination_approval.go"
  - "internal/app/durable_coordination.go"
  - "internal/app/etl_mode_dispatch.go"
  - "internal/app/foundations_integration_test.go"
  - "internal/app/issue_label_transport_approval.go"
  - "internal/app/issue_label_warehouse_transport.go"
  - "internal/app/local_warehouse.go"
  - "internal/app/payload_identity_test.go"
  - "internal/app/postgres_transport_approval_test.go"
  - "internal/app/rest_write_command_test.go"
  - "internal/app/reverse_approval_recovery_test.go"
  - "internal/app/reverse_confirmation_test.go"
  - "internal/app/transport_composition_test.go"
  - "internal/app/transport_dispatch.go"
  - "internal/app/transport_dispatch_test.go"
  - "internal/app/types.go"
  - "internal/app/util.go"
  - "internal/cli/agentic_contract_test.go"
  - "internal/cli/cli.go"
  - "internal/cli/cli_test.go"
  - "internal/cli/connector_command_limits_test.go"
  - "internal/cli/connector_command_result_internal_test.go"
  - "internal/cli/docs.go"
  - "internal/cli/errors.go"
  - "internal/cli/etl_transport.go"
  - "internal/cli/etl_transport_test.go"
  - "internal/cli/github_flow_roundtrip_test.go"
  - "internal/cli/golden_transcript_test.go"
  - "internal/cli/reverse_plan_redaction_test.go"
  - "internal/cli/skills.go"
  - "internal/cli/structured_rest_body_help_test.go"
  - "internal/cli/testdata/golden_transcripts.json"
  - "internal/connectors/command_surface.go"
  - "internal/connectors/commandrunner/content_preservation_test.go"
  - "internal/connectors/commandrunner/github_declared_parity_test.go"
  - "internal/connectors/commandrunner/runner.go"
  - "internal/connectors/commandrunner/runner_test.go"
  - "internal/connectors/conformance/dynamic.go"
  - "internal/connectors/conformance/dynamic_test.go"
  - "internal/connectors/conformance/github_exhaustive_proof_internal_test.go"
  - "internal/connectors/conformance/github_exhaustive_proof_test.go"
  - "internal/connectors/conformance/static.go"
  - "internal/connectors/conformance/static_test.go"
  - "internal/connectors/connectors.go"
  - "internal/connectors/connsdk/http.go"
  - "internal/connectors/connsdk/http_test.go"
  - "internal/connectors/connsdk/multipart_bounds_test.go"
  - "internal/connectors/connsdk/stream.go"
  - "internal/connectors/defs/amazon-sqs/operations.json"
  - "internal/connectors/defs/github/certification-matrix.json"
  - "internal/connectors/defs/github/certification-mutation-candidates.json"
  - "internal/connectors/defs/github/certification-sweep.json"
  - "internal/connectors/defs/github/certification.json"
  - "internal/connectors/defs/github/cli_surface.json"
  - "internal/connectors/defs/github/operations.json"
  - "internal/connectors/defs/github/sources/github-operation-descriptor.json"
  - "internal/connectors/defs/github/writes.json"
  - "internal/connectors/defs/gmail/api_surface.json"
  - "internal/connectors/defs/gmail/cli_surface.json"
  - "internal/connectors/defs/gmail/docs.md"
  - "internal/connectors/defs/gmail/operations.json"
  - "internal/connectors/defs/google-ads/operations.json"
  - "internal/connectors/defs/gorgias/api_surface.json"
  - "internal/connectors/defs/gorgias/cli_surface.json"
  - "internal/connectors/defs/gorgias/docs.md"
  - "internal/connectors/defs/gorgias/operations.json"
  - "internal/connectors/defs/help-scout/operations.json"
  - "internal/connectors/defs/jira/operations.json"
  - "internal/connectors/defs/postgres/certification-matrix.json"
  - "internal/connectors/defs/recurly/operations.json"
  - "internal/connectors/defs/workday-rest/operations.json"
  - "internal/connectors/defs/youtube-analytics/operations.json"
  - "internal/connectors/defs/zoom/certification-matrix.json"
  - "internal/connectors/engine/binary_read.go"
  - "internal/connectors/engine/binary_read_test.go"
  - "internal/connectors/engine/bundle.go"
  - "internal/connectors/engine/bundle_test.go"
  - "internal/connectors/engine/connector.go"
  - "internal/connectors/engine/connector_test.go"
  - "internal/connectors/engine/direct_read.go"
  - "internal/connectors/engine/direct_read_paginate.go"
  - "internal/connectors/engine/direct_read_pagination_test.go"
  - "internal/connectors/engine/direct_read_test.go"
  - "internal/connectors/engine/direct_write.go"
  - "internal/connectors/engine/direct_write_multipart_test.go"
  - "internal/connectors/engine/direct_write_test.go"
  - "internal/connectors/engine/github_binary_download_test.go"
  - "internal/connectors/engine/graphql_operation.go"
  - "internal/connectors/engine/graphql_operation_test.go"
  - "internal/connectors/engine/interpolate.go"
  - "internal/connectors/engine/interpolate_test.go"
  - "internal/connectors/engine/operation_direct_read_bindings_test.go"
  - "internal/connectors/engine/operation_headers.go"
  - "internal/connectors/engine/operation_kind.go"
  - "internal/connectors/engine/operation_kind_test.go"
  - "internal/connectors/engine/operation_multipart_test.go"
  - "internal/connectors/engine/operation_parameters.go"
  - "internal/connectors/engine/prepared_write.go"
  - "internal/connectors/engine/rate_limit_coordination_test.go"
  - "internal/connectors/engine/rate_limit_parking.go"
  - "internal/connectors/engine/rate_limit_runtime.go"
  - "internal/connectors/engine/rate_limit_runtime_test.go"
  - "internal/connectors/engine/read.go"
  - "internal/connectors/engine/record_schema_promotion.go"
  - "internal/connectors/engine/record_schema_promotion_test.go"
  - "internal/connectors/engine/response_receipt.go"
  - "internal/connectors/engine/schema.go"
  - "internal/connectors/engine/schema/cli_surface.schema.json"
  - "internal/connectors/engine/schema/operations.schema.json"
  - "internal/connectors/engine/schema/sync_transport.schema.json"
  - "internal/connectors/engine/schema/writes.schema.json"
  - "internal/connectors/engine/schema_test.go"
  - "internal/connectors/engine/sensitive_transform.go"
  - "internal/connectors/engine/status_check.go"
  - "internal/connectors/engine/status_check_test.go"
  - "internal/connectors/engine/structured_rest_body.go"
  - "internal/connectors/engine/structured_rest_body_test.go"
  - "internal/connectors/engine/text_export_test.go"
  - "internal/connectors/engine/write.go"
  - "internal/connectors/engine/write_prepare.go"
  - "internal/connectors/engine/write_query_test.go"
  - "internal/connectors/engine/write_record_hook_test.go"
  - "internal/connectors/engine/write_retry_policy_test.go"
  - "internal/connectors/engine/write_test.go"
  - "internal/connectors/engine/xero_operations_test.go"
  - "internal/connectors/guide.go"
  - "internal/connectors/guide_test.go"
  - "internal/connectors/hooks/google-calendar/hooks_test.go"
  - "internal/connectors/native/ashby/connector_contract_test.go"
  - "internal/connectors/native/ashby/engine_delegate.go"
  - "internal/connectors/native/postgres/transport_source_test.go"
  - "internal/connectors/operation_headers.go"
  - "internal/connectors/sync_transport.go"
  - "internal/connectors/sync_transport_test.go"
  - "internal/connectors/transportpolicy/policy.go"
  - "internal/connectors/write_result_output_test.go"
  - "internal/coordination/durable_store.go"
  - "internal/coordination/durable_store_edges_test.go"
  - "internal/coordination/rate_parking.go"
  - "internal/coordination/rate_parking_test.go"
  - "internal/synccontract/commit.go"
  - "internal/synctransport/arrow_fast_path_controller.go"
  - "internal/synctransport/arrow_fast_path_pipeline.go"
  - "internal/synctransport/orchestrator.go"
  - "internal/synctransport/registry.go"
  - "internal/synctransport/transport_test.go"
  - "internal/synctransport/types.go"
  - "scripts/gen-github-graphql-parity.mjs"
  - "scripts/tests/gen-github-graphql-parity.test.mjs"
  - "website/content/docs/cli-reference.mdx"
  - "website/content/docs/etl.mdx"
  - "website/content/docs/github-cli-surface.mdx"
  - "website/data/connectors.generated.json"
  - "website/lib/connectors.catalog.data.generated.json"
  - "website/lib/docs.generated.ts"
  - "website/scripts/cli-surface.test.mjs"
  - "website/scripts/gen-docs-data.mjs"
  - "website/scripts/gen-docs-data.test.mjs"
  - "website/scripts/gen-github-cli-surface.mjs"
  - "website/scripts/lib/cli-surface.mjs"
findings:
  critical: 12
  warning: 2
  info: 0
  total: 14
status: issues_found
verdict: blockers
---

# Foundation post-fix mapping and generation review

## Narrative Findings (AI reviewer)

### Executive verdict

**Verdict: blockers.** The frozen implementation contains 12 BLOCKER findings and 2 WARNING findings in the mapping/generation lens. The required GitHub generated-artifact gate fails at the reviewed SHA; source-gap operations are deliberately excluded from completeness checks while still advertised as implemented; 14 Google Ads POST reads expose closed empty nested request objects; GraphQL pagination, result selection, secret classification, and Int bounds are incomplete; the v2 GraphQL source projection is not cryptographically bound to the digest it advertises; and both Foundation and connector certification evidence can remain affirmative after the code they describe changes.

The independent finding set was frozen before reading the prior REVIEW.md and REVIEW-FIX.md. Those ledgers were then used only to classify prior closures and identify omissions or regressions.

### Review preflight

- Repository: /Users/karthiksivadas/.treehouse/cli-83d592/57/cli
- Frozen HEAD: 8a8a866ff6d5282c28bda12acceed8a624218f01
- Diff base: e62ae21d428f0d27225f9bff564dc2cd797f6b65
- HEAD was verified before review, before the prior-ledger cross-check, and after the final evidence check.
- Source, test, generated, branch, commit, remote, and PR state were read-only throughout.
- The worktree was clean at preflight. Peer review Markdown that appeared later was treated as permitted review output, not source drift.
- .codegraph/ does not exist, so CodeGraph was unavailable. Review used read-only source tracing, exact JSON cohort audits, and tests.
- AGENTS.md, the repository caveman skill, and the deep gsd-code-reviewer method were read before source review.

### Scope manifest

The frozen range changes 819 files. Forty-five are planning artifacts; 774 non-planning files form the review scope and are enumerated exactly in the frontmatter.

| Cohort | Files | Review treatment |
|---|---:|---|
| Generated connector skills | 554 | Generator, exact-match tests, registry coverage, and representative GitHub output; not superficial prose-by-prose review |
| cmd/connectorgen | 38 | Full changed generator/importer/projector/validator and tests |
| internal/connectors/defs | 25 | Full declarations, source locks, descriptors, operations, writes, CLI/API surfaces, ledgers, and representative output |
| internal/connectors outside defs | 79 | Cross-module engine, command runner, registry, guide, conformance, and tests |
| internal/app | 20 | Reverse-ETL plan/preview/approval/execute and result paths |
| internal/cli | 15 | Generated command/alias/flag reachability and help |
| scripts and script tests | 2 | GitHub REST+GraphQL parity generation and ledger gates |
| website | 11 | Lossless surface mapping, types, generated catalog, and exact tests |
| data | 3 | Foundation evidence manifest and generated data |
| Non-skill documentation | 14 | Generation/certification contracts and representative output |
| Other source/config | 13 | Make gates, registry/build integration, and remaining changed surfaces |
| **Total** | **774** | Deep mapping/generation review |

### Checks run

| Check | Result |
|---|---|
| git rev-parse HEAD | PASS; exact frozen SHA |
| git status --short --untracked-files=all at preflight | PASS; clean |
| go test -timeout 20m ./cmd/connectorgen | PASS, 187.284s |
| go test -timeout 20m ./internal/connectors/engine ./internal/connectors/commandrunner ./internal/connectors/conformance | PASS, 13.265s / 28.176s / 41.027s |
| go test -timeout 20m ./internal/app ./internal/cli | PASS, 351.623s / 1114.517s |
| Focused exact generated-skill test in ./internal/cli | PASS, 211.458s |
| go test ./internal/connectors/connsdk ./internal/connectors and focused guide tests | PASS |
| TestEveryImplementedCommandPassesRuntimePreflight | PASS, 11.456s |
| Node tests for GitHub GraphQL generation, CLI surface, and docs data | PASS, 16 tests |
| Website GitHub generated-surface tests | PASS, 3 tests |
| node --test combined-operation-ledger plus source-drift tests | **FAIL**, 10 pass / 1 fail; github.graphql.query.nodes state differs |
| node scripts/gen-github-graphql-parity.mjs --check | **FAIL**, exit 1 |
| node scripts/github-combined-operation-ledger.mjs --check | **FAIL**, exit 1 |
| go run ./cmd/connectorgen validate internal/connectors/defs | PASS; 552 connectors, zero findings |
| go run ./cmd/connectorgen surface-sync --check | PASS; 552 connectors, zero reported drift |
| certification-matrix --check | PASS; 3 connectors, capability_complete=0, certified=0 |
| certification-candidates --check and certification-sweep --check | PASS; 1,584 sweep rows / 1,580 commands |
| go run ./cmd/agentcontractgen check | PASS |
| Foundation evidence-gate against checked manifest/TDD/review | PASS despite stale code_sha/base_sha |
| git diff --check | PASS |
| GitHub source-gap cohort audit | 351 gap-tagged operations: DELETE 19, GET 15, PATCH 64, POST 151, PUT 102 |
| Implemented zero-input commands joined to gap-tagged GitHub endpoints | 37 commands/endpoints: 36 reverse_etl and 1 direct_read |
| Google Ads closed-empty POST body cohort audit | 31 nodes across 14 implemented POST reads |
| GitHub GraphQL mutation payload audit | 254 of 274 payloads declare a nested result omitted by the fixed document |
| GitHub accepted-evidence audit | 985 records, 13 PM binary digests, 2026-08-17 through 2026-08-19 |

Passing unit, validation, sync, skill, website, and certification checks do not close the findings below: several tests explicitly encode the defective behavior, and two required GitHub artifact checks fail.

### Six-surface reachability and source-driven mapping audit

| Surface | Declared/reachable state at frozen HEAD | Review conclusion |
|---|---|---|
| ETL | GitHub declares 37 streams. Generic catalog/read remains reachable through the registered connector path. Twenty-two source GET routes without a provider-named CLI command are represented by ETL stream reads. | Reachable. No hidden ordinary ETL stream was found, but source-gap completeness is not enforced for the one gap-tagged direct read in PFM-B02. |
| Reverse ETL | GitHub declares 606 write actions; 591 CLI commands have reverse_etl intent. Thirty-two actions without a provider-named command remain selectable through the generic reverse plan action. App.PlanReverseETL resolves the endpoint and WriteValidator generically at internal/app/app.go:1871-1915; preview uses DryRunWriter at 2292-2352; execute uses Writer/OperationDirectWriter at 2870-2933. | Generic and declaration-driven, not GitHub-hard-coded. Reachability exists, but 36 gap-tagged source operations are advertised with no request flags (PFM-B02). |
| Direct read | GitHub has 654 direct_read commands. REST path/query/body/header binding flows from commandrunner.operationDirectOverrides at internal/connectors/commandrunner/runner.go:1554-1644 into engine validation. | Reachable. Defects remain in source-gap projection, Google Ads nested input, GraphQL pagination, secret classification, Int bounds, and fail-open propagation. |
| Direct write | GitHub has 287 direct_write commands, including 274 generated GraphQL mutations plus REST writes. Preview/approval/execute is enforced through OperationDirectWriter at internal/app/app.go:2895-2933. | Reachable. Generated GraphQL documents omit provider resources, expose some secrets, and accept out-of-range Int values. |
| Binary download | GitHub declares 13 binary download operations; commandrunner and engine preflight/execution are covered by the passing runtime cohort. | Reachable. No mapping blocker found. Gmail attachment partiality is correctly explicit because the provider response is a base64url envelope rather than raw bytes. |
| Binary upload | releases assets upload is implemented at internal/connectors/defs/github/cli_surface.json:10664-10709. Its write binds release_id, name, label, and file_path to uploads.github.com, uses body_type binary_upload, root-confines the file, and caps it at 64 MiB at writes.json:5988-6034. | Reachable and declaration-driven. Its ledger row is stale under PFM-B01, but runtime mapping closes prior CF-B05. |

The GitHub CLI surface contains 1,580 commands: 1,555 implemented, 4 unsupported_api, and 21 unsupported_local. Intent counts are auth 3, binary_download 14, config 3, direct_read 654, direct_write 287, etl 15, local_workflow 12, raw_api 1, and reverse_etl 591. The operation registry contains 770 runtime operations: 13 binary downloads, 274 GraphQL mutations, 36 GraphQL queries (31 generated plus 5 supplemental), 1 local-git operation, 231 REST reads, and 215 REST writes.

The GitHub source lock declares REST SHA-256 80850db290cde4eb487e0efb587cf27f305e77b6bef96933ed8a09b5169d5b1d over 12,920,264 bytes and 1,220 operations at internal/connectors/defs/github/sources/github-operation-source-lock.json:5-12. It declares GraphQL SHA-256 c09aba9911b08d2aa8a022578edaf256aa040f38d7fb7196656356ea236c249d over 1,546,421 bytes, 31 query roots, and 274 mutation roots at lines 10995-10999. Counts and copied digest metadata match, but PFM-B09 shows the embedded v2 GraphQL projection is not authenticated by that digest.

Ordinary provider truth remains intact when selected. internal/connectors/engine/graphql_operation_test.go:456-487 asserts the complete receipt retains occurrence ID graphql-occurrence-9007199254740993 and an unclassified provider field remains intact. Required fixes must use explicit source-derived secret classification, not heuristic deletion of token-like fields or occurrence IDs.

## Critical Issues

### PFM-B01: Required GitHub generated-parity artifacts are stale

**Classification:** BLOCKER

**File and symbol evidence:** Makefile:108-115 defines github-parity-artifacts-check and Makefile:158-163 includes it in both verify targets. scripts/gen-github-graphql-parity.mjs:665-676 rebuilds operations, commands, and API transport by filtering old generated entries and appending a new cohort. scripts/tests/github-combined-operation-ledger.test.mjs:259-272 requires the checked ledger to match the real bundle.

**Cross-file call path:** GitHub source lock -> generatedOperation/generatedCommand -> mergeGeneratedBundle at scripts/gen-github-graphql-parity.mjs:665-676 -> operations.json, cli_surface.json, api_surface.json -> github-combined-operation-ledger.mjs -> combined-operation-ledger.json -> Makefile github-parity-artifacts-check -> verify.

**Concrete defect:** At frozen HEAD, node scripts/gen-github-graphql-parity.mjs --check and node scripts/github-combined-operation-ledger.mjs --check exit 1. Semantic CLI content differs only in ordering: the checked file starts GraphQL at command index 1,266 and puts nine new REST commands at 1,571-1,579; regeneration expects those nine before GraphQL and starts GraphQL at 1,275. Operations/API values otherwise byte-match. The combined ledger has 11 stale rows: github.graphql.query.nodes, nine REST command rows, and release upload. Checked implementation progress is 1,454/1,525 rather than regenerated 1,463/1,525. The focused test fails because nodes is actually fixed_projection_only while the checked row says fixed_typename_projection.

**Six-surface impact:** ETL — no runtime difference demonstrated, but the required release gate blocks shipment. Reverse ETL — nine REST classifications and release upload are stale in evidence. Direct read — nodes projection classification is stale. Direct write — shared ledger is not exact. Binary download — runtime unaffected, common gate fails. Binary upload — executable, but ledger row stale.

**Exact fix:** Define one canonical total ordering for authored REST and generated GraphQL commands, apply it regardless of generation order, regenerate cli_surface.json and combined-operation-ledger.json, and keep operations/API artifacts byte-exact.

**Exact behavioral regression test:** Add TestGitHubParityGenerationOrderIsCommutative: create source-projected REST commands before and after GraphQL generation in two temporary trees, assert byte-identical operations/CLI/API artifacts and ledgers, assert the nodes state, and require both --check commands plus make github-parity-artifacts-check to pass.

### PFM-B02: Gap-tagged source operations are skipped by generation and completeness validation while advertised as implemented

**Classification:** BLOCKER

**File and symbol evidence:** cmd/connectorgen/sourceprojection.go:94-109 skips every REST mutation with sourceProjectionHasBlockingGap. Its contract says gaps must not masquerade as coverage at lines 613-616, but validateSourceExecutableCoverage repeats the skip at 688-720, especially 702-704. GitHub repos/update has a rich source body at internal/connectors/defs/github/sources/github-operation-descriptor.json:594164-594238 and a body gap at 598837-598845; writes.json:2535-2547 gives repo2 an empty schema and cli_surface.json:3385-3399 advertises it with no flags. GET /advisories declares query inputs at descriptor lines 45076-45163 and gaps at 46074-46087, while cli_surface.json:11713-11726 advertises it with no flags. commandrunner only materializes declared flags at internal/connectors/commandrunner/runner.go:1554-1644.

**Cross-file call path:** provider OpenAPI -> source descriptor with runtime.gaps -> source projection skips gap -> write/operation remains empty -> surface-sync derives no request fields -> coverage validator skips same identity -> CLI/API declare implemented -> commandrunner sends empty/incomplete request.

**Concrete defect:** The descriptor contains 351 gap-tagged operations. Joining it to implemented zero-flag commands finds 37 source endpoints: 36 reverse_etl and one direct_read. Safety gaps hide required provider inputs without hiding the operation. repo update cannot supply legal update fields; list-global-advisories cannot supply affects, cwes, direction, ecosystem, or the remaining query contract. Validation reports zero drift because it excludes exactly the identities needing review.

**Six-surface impact:** ETL — no ordinary stream mapping proven missing. Reverse ETL — 36 writes are implemented but request-incomplete. Direct read — GET /advisories is implemented with hidden query contract. Direct write — shared projection can count empty typed operations as covered. Binary download/upload — unaffected in the current gap cohort.

**Exact fix:** Remove the gap predicate from completeness accounting. Extend the request-schema foundation to represent bounded union/map variants or retain explicit typed alternatives. A source operation remains visible, but must not be availability=implemented or count as covered until action plus CLI binds every path/query/body/header input; do not omit the operation merely for safety.

**Exact behavioral regression test:** Add TestSourceProjectionGapOperationsCannotMasqueradeAsImplemented using repos/update and /advisories. Assert coverage fails while any field is absent, then succeeds with exact typed flags. Add/change/delete one path, query, body, and header field and assert --check fails before network I/O.

### PFM-B03: Google Ads closes 31 nested provider request objects as empty, making valid POST reads impossible

**Classification:** BLOCKER

**File and symbol evidence:** internal/connectors/defs/google-ads/operations.json:322-403 declares customers.generate.keyword.ideas. Six legal branches — keywordAndUrlSeed, urlSeed, aggregateMetrics, keywordSeed, siteSeed, historicalMetricsOptions — are closed objects with no properties at lines 355-386. Its implemented command at cli_surface.json:317-383 exposes only scalar/array flags. engine.compileOperationDirectReadBodySchema closes objects at internal/connectors/engine/direct_read.go:847-925, and operationReadBody validates at 755-816.

**Cross-file call path:** Google Ads Discovery source -> generated body_schema -> scalar-only CLI flags -> commandrunner.operationDirectOverrides -> engine.operationReadBody -> closed schema validation -> POST.

**Concrete defect:** A cohort audit found 31 closed empty nested objects across 14 implemented Google Ads POST reads. Every non-empty legal provider value fails validation and there is no JSON/nested flag to express it. generateKeywordIdeas cannot supply any useful seed object.

**Six-surface impact:** Direct read — 14 implemented POST reads reject legal input. ETL, reverse ETL, direct write, binary download, and binary upload — unaffected.

**Exact fix:** Resolve Discovery references for all 31 nodes into finite closed child schemas, preserve provider requiredness/unions, and generate JSON or dotted flags for every usable branch. Enforce the exact provider one-of constraint for mutually exclusive seeds.

**Exact behavioral regression test:** Add TestGoogleAdsGeneratedPOSTReadsAcceptDeclaredNestedObjects covering non-empty keywordSeed and an alternate seed with exact provider JSON and one-of enforcement. Add a cohort walk rejecting reachable closed-empty objects without an explicit provider-empty-object annotation.

### PFM-B04: Paginated GraphQL commands allow neither first nor last

**Classification:** BLOCKER

**File and symbol evidence:** scripts/gen-github-graphql-parity.mjs:207-230 marks only source non-null arguments required; pagination only adds properties. commandFlagForArgument at 364-386 does the same. Generated search requires only query/type at internal/connectors/defs/github/operations.json:9558-9602 and leaves first/last optional at cli_surface.json:41032-41069. graphQLOperationVariables rejects mixed directions and bounds supplied sizes at internal/connectors/engine/graphql_operation.go:563-619, but requires neither. Generator test lines 123-131 checks only flag presence; runtime test internal/connectors/engine/graphql_operation_test.go:383-402 omits zero-direction.

**Cross-file call path:** GraphQL source arguments -> rootVariablesSchema/generatedCommand -> optional first/last -> commandrunner body -> graphQLOperationVariables -> fixed query -> GitHub connection resolver.

**Concrete defect:** An implemented paginated query passes local validation with neither first nor last, then GitHub rejects it because connection resolvers require a direction.

**Six-surface impact:** Direct read — every generated GraphQL connection command accepts invalid zero-direction input. Other five surfaces — unaffected.

**Exact fix:** Enforce exactly one of first or last before I/O, using a schema constraint plus runtime validation or a documented deterministic first default. Keep cursors opaque through --page-cursor.

**Exact behavioral regression test:** Add TestGraphQLOperationVariablesRequiresExactlyOnePaginationDirection: neither fails pre-I/O; first-only/last-only pass; both fail; cursor without direction fails. Assert every generated paginated command advertises the same constraint.

### PFM-B05: Backward GraphQL pagination reads forward pageInfo

**Classification:** BLOCKER

**File and symbol evidence:** Generated pagination declares forward/backward variables at scripts/gen-github-graphql-parity.mjs:409-417. graphQLOperationVariables chooses before when last is supplied at internal/connectors/engine/graphql_operation.go:573-596. operationGraphQLDirectRead calls graphQLOperationPage without direction at 688-695. graphQLOperationPage at 938-988 always reads hasNextPage/endCursor.

**Cross-file call path:** CLI --last plus --page-cursor -> commandrunner body -> graphQLOperationVariables inserts before -> provider pageInfo -> graphQLOperationPage ignores backward mode -> DirectReadPage.

**Concrete defect:** A backward response with hasPreviousPage=true/startCursor=prev and hasNextPage=false is reported complete with no continuation; an unrelated forward cursor can also be returned.

**Six-surface impact:** Direct read — backward GraphQL traversal truncates or misnavigates results. Other five surfaces — unaffected.

**Exact fix:** Return/derive the validated direction, pass it to graphQLOperationPage, and select hasPreviousPage/startCursor for backward versus hasNextPage/endCursor for forward.

**Exact behavioral regression test:** Add TestOperationDirectReadBackwardGraphQLPaginationUsesPreviousPageInfo with last set, hasNextPage=false, hasPreviousPage=true, startCursor=previous-cursor, endCursor=forward-cursor; assert HasMore and NextCursor=previous-cursor, plus terminal/malformed cases.

### PFM-B06: GraphQL secret classification misses query arguments and provider-issued response secrets

**Classification:** BLOCKER

**File and symbol evidence:** scripts/gen-github-graphql-parity.mjs:306-335 classifies sensitive input only through mutationPolicy, and generatedCommand at 436-463 gives every query a non-sensitive policy. The source lock declares enterpriseAdministratorInvitationByToken(invitationToken: String!) at internal/connectors/defs/github/sources/github-operation-source-lock.json:11111-11124. Generated operation/CLI accept inline input at operations.json:8994-9018 and cli_surface.json:40544-40564. The repository query selects tempCloneToken at operations.json:9458-9468. regenerateVerifiableDomainToken selects verificationToken at operations.json:19143-19181, but its CLI has no response policy at cli_surface.json:45024-45046; the generator test explicitly expects no sensitive_policy at scripts/tests/gen-github-graphql-parity.test.mjs:151-153. Runtime masks only configured credentials at internal/connectors/connectors.go:1085-1092; response path masking requires SensitivePolicy.ResponseSecretField at internal/connectors/engine/direct_write.go:249-254.

**Cross-file call path:** source argument/result -> query bypasses mutationPolicy or mutation examines only input -> generated CLI lacks env_only/output path -> commandrunner accepts argv or fixed document selects token -> engine masks only configured values -> app publishes result at internal/app/app.go:2924-2929.

**Concrete defect:** Invitation tokens can enter process arguments and shell history. Provider-issued tempCloneToken and verificationToken are not configured credentials, so json_redacted passes them through. This is credential disclosure. Blanket name-based deletion is unacceptable because ordinary token-like provider values and occurrence IDs must remain.

**Six-surface impact:** Reverse ETL — GraphQL mutations can publish provider-issued secrets. Direct read — secret query input is argv-visible and fixed queries can return tokens. Direct write — mutation results can expose verification tokens. ETL and binary surfaces — unaffected.

**Exact fix:** Derive explicit sensitive input/output paths from GraphQL schema plus reviewed provider semantics for Query and Mutation. Mark secret query flags env_only; classify selected secret result paths; mask/store those exact paths at the public boundary; preserve every unclassified field, ID, and occurrence ID.

**Exact behavioral regression test:** Test enterpriseAdministratorInvitationByToken rejects inline input and accepts --from-env. repository.tempCloneToken and regenerateVerifiableDomainToken.verificationToken must be masked/withheld, while issue IDs, clientMutationId, token-count fields, and graphql-occurrence-9007199254740993 remain exact.

### PFM-B07: Generated GraphQL mutations omit created and updated resources

**Classification:** BLOCKER

**File and symbol evidence:** outputSelection at scripts/gen-github-graphql-parity.mjs:258-273 selects only concrete scalar leaves for ordinary object payloads. The source lock declares createIssue returning CreateIssuePayload at internal/connectors/defs/github/sources/github-operation-source-lock.json:13498-13518; its payload contains issue: Issue at 44752-44770. Generated operation internal/connectors/defs/github/operations.json:13432-13444 requests only __typename/clientMutationId. Engine returns only selected data at internal/connectors/engine/direct_write.go:212-234.

**Cross-file call path:** GraphQL payload type system -> outputSelection/scalarLeafSelection -> fixed mutation document -> provider selection -> OperationDirectWriteResult -> reverse run.

**Concrete defect:** 254 of 274 mutation payloads contain nested object/interface/union fields the generator drops. createIssue cannot return created issue ID/number/URL. Writes can report success while withholding authoritative resource identity needed for reconciliation.

**Six-surface impact:** Reverse ETL — GraphQL writes cannot reconcile provider resources. Direct write — most generated mutations discard nested resources. ETL, direct read, binary download/upload — unaffected.

**Exact fix:** Generate bounded schema-derived nested payload selection containing provider identity/status fields, with depth/field ceilings and PFM-B06 secret policy. At minimum select id plus stable identifiers/URLs, while keeping scalar payload status/clientMutationId.

**Exact behavioral regression test:** Add TestGeneratedGraphQLMutationDocumentsPreserveResourceIdentity for createIssue and addComment. Assert document/result contain issue/comment id and URL unchanged, retain clientMutationId/occurrence IDs, mask classified secrets, and stay within limits.

### PFM-B08: GraphQL Int is unrestricted instead of signed 32-bit

**Classification:** BLOCKER

**File and symbol evidence:** scalarSchema maps Int to type integer at scripts/gen-github-graphql-parity.mjs:160-173; commandFlagForArgument maps native integer at 382-386. Source requiredApprovingReviewCount is Int at internal/connectors/defs/github/sources/github-operation-source-lock.json:21100-21105; generated schema has only integer at operations.json:12633-12635. requireClosedBoundedGraphQLVariables checks only collections at internal/connectors/engine/graphql_operation.go:439-468. internal/connectors/engine/schema.go:87-116 has no numeric range fields, and commandrunner parses a platform integer at runner.go:2045-2118.

**Cross-file call path:** source Int -> scalarSchema -> variables_schema/integer flag -> native coercion -> Schema.Validate -> JSON -> GitHub variable coercion.

**Concrete defect:** Values outside [-2147483648, 2147483647] pass locally on 64-bit builds and fail at the provider, including nested JSON Int fields. Aggregate payload byte bounds at graphql_operation.go:732-745 do not enforce scalar domain.

**Six-surface impact:** Reverse ETL/direct write — mutation plans can approve provider-invalid Int input. Direct read — query Int fails only at provider. ETL and binary surfaces — unaffected.

**Exact fix:** Add GraphQL-specific signed-32-bit validation for every Int node, including nested lists/input objects, via exact schema min/max support or a type-tree value walk before I/O.

**Exact behavioral regression test:** Add TestGraphQLIntUsesSigned32BitDomain for reads/writes: bounds pass, adjacent values fail before server call; repeat inside nested object and list.

### PFM-B09: The v2 embedded GraphQL projection is not authenticated by its advertised digest

**Classification:** BLOCKER

**File and symbol evidence:** REST import fetches/recomputes bytes and SHA-256 at cmd/connectorgen/sourceimport.go:650-660. appendLockedGraphQLProjection skips fetch/digest for schema v2 at 716-740, then copies the external SHA/bytes onto embedded schema/descriptors at 748-780. The lock advertises digest/bytes at internal/connectors/defs/github/sources/github-operation-source-lock.json:10995-10998. TestSourceImportVersion2UsesEmbeddedLockedGraphQLProjection at cmd/connectorgen/sourceimport_test.go:125-147 supplies arbitrary embedded fields with a fake all-a digest and succeeds without GraphQL fetch. validateSourceDescriptorAgainstLock merely compares copied metadata at cmd/connectorgen/sourceprojection.go:645-685.

**Cross-file call path:** embedded GraphQL lock -> v2 import bypasses external bytes -> descriptors inherit unrelated digest -> validator compares copied value to origin -> generators trust embedded type system.

**Concrete defect:** A signature/type/count change inside the checked lock can retain the old external digest and pass. The descriptor claims provenance from bytes never hashed; identity/count equality does not establish content identity.

**Six-surface impact:** Reverse ETL/direct write — generated GraphQL mutations lack authenticated source provenance. Direct read — generated GraphQL queries lack it. ETL and REST-based binary surfaces — runtime unaffected; REST lock remains authenticated.

**Exact fix:** Add projection_sha256/projection_bytes over canonical embedded query fields, mutation fields, and type system and verify every import, or fetch an immutable artifact and verify the current digest. Never label embedded bytes with an unrelated digest.

**Exact behavioral regression test:** Replace the permissive test with TestSourceImportVersion2RejectsEmbeddedGraphQLProjectionDigestDrift. Mutate one signature, type field, and root count while retaining digests; each fails. Reordering must normalize identically or reject deterministically.

### PFM-B10: Foundation evidence gate accepts stale implementation evidence

**Classification:** BLOCKER

**File and symbol evidence:** foundationEvidenceManifest parses CodeSHA, ReviewedSHA, BaseSHA, ComponentInputs at cmd/connectorgen/evidencegate.go:15-29. validateFoundationEvidence checks SHA formatting/reviewed SHA only at 76-98; lines 99-149 never compare CodeSHA, BaseSHA, or component inputs with current checkout. The manifest records code_sha 808896a..., reviewed_sha 9e5329f..., base_sha 114a677... at data/cli-current-foundations-main-integration-r1/evidence-manifest.json:1-7, none equal the frozen implementation/base pair. evidencegate_test.go:11-82 lacks stale code/base/component cases. The real gate still succeeds.

**Cross-file call path:** evidence manifest -> runEvidenceGate -> validateFoundationEvidence -> compare only REVIEW.md source_sha -> successful Foundation evidence claim.

**Concrete defect:** Focused checks from an older snapshot remain accepted for current code. The success message claims evidence matches even though code_sha/base_sha are stale and component SHAs are unbound.

**Six-surface impact:** All six surfaces can be certified by checks from a different implementation snapshot.

**Exact fix:** Receive/derive exact implementation HEAD and diff base, require manifest CodeSHA/BaseSHA equality, validate every component SHA/preserving merge against the reviewed graph, and require checks to execute on that clean tree. Bind ReviewedSHA explicitly as review identity only.

**Exact behavioral regression test:** Mutate code_sha, base_sha, each component SHA, and preserving_merge; every stale value fails. Frozen 8a8a866.../e62ae21... passes only after evidence regeneration.

### PFM-B11: Certification matrices treat historical accepted evidence as fresh forever

**Classification:** BLOCKER

**File and symbol evidence:** acceptedEvidence stores identity/timestamp/proof at cmd/connectorgen/certificationmatrix.go:159-183. embeddedEvidenceProof includes PMBinarySHA256/PMCommandFingerprint at certificationproof.go:45-59. validateAcceptedEvidence at certificationmatrix.go:1508-1575 checks shape but never compares implementation/source/config fingerprint. matchingCapabilityEvidence at 1394-1417 matches only identity fields; cellComplete at 1420-1435 treats any match as live. Workflow/sync derivation uses the same evidence path.

**Cross-file call path:** certifications/evidence JSON -> validate -> identity-only match -> LiveTested/LiveEvidence -> capability/workflow/sync/flow cell -> generated certification shard/status.

**Concrete defect:** After source locks, operations, mappings, or pm binary change, old evidence still proves the current cell. The cohort contains 985 GitHub records from 13 PM binary digests. Historical records are useful but cannot establish present behavior.

**Six-surface impact:** Every surface can inherit stale affirmative live evidence after its contract changes.

**Exact fix:** Add a deterministic subject fingerprint covering pm binary/build, connector declarations, source/projection digest, command mapping, relevant config, and proof protocol. Require exact equality for LiveTested; retain nonmatches as historical with stale reason.

**Exact behavioral regression test:** Add TestCertificationEvidenceBecomesStaleWhenSubjectChanges. Change an operation field, CLI maps_to, source digest, and pm binary digest; each clears LiveTested/LiveEvidence and blocks completion until new evidence.

### PFM-B12: params-import fails open when a provider operation or parameter reference disappears

**Classification:** BLOCKER

**File and symbol evidence:** runParamsImport returns success whenever changed is zero at cmd/connectorgen/paramsimport.go:64-82. importConnectorParameters increments total but silently continues when provider path/method is absent at 207-217. resolveOpenAPIParameter returns false for unresolved refs at 465-478; importedParameters silently continues at 481-488.

**Cross-file call path:** provider OpenAPI -> params-import route/ref lookup -> missing entries dropped -> zero/partial parameter set -> surface-sync -> CLI/API -> --check success.

**Concrete defect:** Provider deletion/rename or broken reference can remove all source inputs while --check exits zero. Downstream artifacts remain stale; add/change/delete propagation is not fail-closed.

**Six-surface impact:** ETL/reverse ETL — declarations can retain stale provider assumptions. Direct read/direct write — path/query/header inputs can disappear silently. Binary download/upload — no current binary-specific failure demonstrated.

**Exact fix:** Error for every executable operation whose provider path/method is absent and every unresolved ref. Allow omission only via source-identity-bound reviewed allowlist with reason/expiry; report exact identity/ref and write nothing in --check.

**Exact behavioral regression test:** Add TestParamsImportRejectsDeletedProviderRoute and TestParamsImportRejectsUnresolvedParameterRef. Delete method/path and break component/legacy refs; --check exits nonzero, identifies source, writes nothing.

## Warnings

### PFM-W01: surface-sync preserves stale or incorrect authored parameter contracts

**Classification:** WARNING

**File and symbol evidence:** deriveCommandParameterFlags documents/implements add-only behavior at cmd/connectorgen/surfacesync.go:732-800; existing flags skip at 749-775. TestDeriveCommandParameterFlagsOnlyAdds enshrines it at cmd/connectorgen/paramsimport_test.go:371-406. For operation reads, command-declared query fields absent from rest.parameters are admitted at internal/connectors/engine/direct_read.go:656-689 and assigned a generic bound at 718-727.

**Cross-file call path:** params-import updates operation parameters -> surface-sync sees existing flag -> preserves type/values/required/repeatability/maps_to -> command declaration can authorize stale field -> runtime request.

**Concrete defect:** Add propagation works, but change/delete is not guaranteed. Provider enum, requiredness, type, location, repeatability, or byte bound can change while authored CLI stays stale. Current frozen GitHub audit found no concrete mismatch, so this is a robustness warning.

**Six-surface impact:** Reverse ETL/direct read/direct write can drift; no current ETL or binary mismatch.

**Exact fix:** Synchronize provider-owned maps_to/type/enum/required/repeatability/max_bytes exactly. Preserve only prose unless a source-bound reviewed narrowing annotation tests the difference. Delete flags when provider inputs disappear.

**Exact behavioral regression test:** Change enum/type/location/requiredness/repeatability/max bytes and assert exact update; delete provider parameter and assert flag removal or explicit reviewed exception; assert prose survives.

### PFM-W02: Website and generated skills drop safety-critical flag metadata

**Classification:** WARNING

**File and symbol evidence:** website mapFlags omits env_only, max_bytes, min_items, max_items at website/scripts/lib/cli-surface.mjs:100-118. ConnectorCliFlag lacks them and repeatable at website/lib/connectors.types.ts:37-47. GitHub label delete has max_bytes 8192 at internal/connectors/defs/github/cli_surface.json:2364-2386, but website/data/connectors.generated.json:71645-71661 omits it. create-migration-source has env_only at cli_surface.json:43022-43035, but website output omits it at 123341-123358. Skill rendering prints only required/repeatable at internal/connectors/guide.go:304-317; docs/skills/pm-github/SKILL.md:2689 and 3968 omit max_bytes/env-only. Actual CLI help renders env-only at internal/cli/cli.go:1152-1175.

**Cross-file call path:** cli_surface flag -> website mapFlags/type -> generated website JSON/catalog; separately -> guide renderer -> 554 generated skill files.

**Concrete defect:** Documentation presents secret input as ordinary inline JSON and hides byte/item acceptance limits. GitHub has 297 affected flags: 294 max_bytes and 3 env_only. Output is deterministic but lossy.

**Six-surface impact:** ETL/direct read/reverse ETL/direct write guidance can omit constraints; binary download uses separate injected flags; binary upload name/path limits are not faithfully represented everywhere.

**Exact fix:** Make website/skill models lossless for env_only, max_bytes, numeric/item bounds, repeatable, allow_empty, format, required, values, maps_to. Render env-only as usage constraint.

**Exact behavioral regression test:** Walk every source CLI flag and compare every semantic property in website JSON and skill metadata. Pin label delete max_bytes=8192 and create-migration-source env_only=true.

## Prior-finding closure table for this lens

| Prior ID | Post-fix state | Evidence and relation |
|---|---|---|
| CF-B01 | **PARTIAL / REOPENED** | REST digest/count and GraphQL type projection exist, but embedded v2 GraphQL projection is not digest-bound (PFM-B09). |
| CF-B02 | **PARTIAL / REOPENED** | Canonical descriptor/checks exist, but gap operations skip, params-import fails open, and parity artifacts are stale (PFM-B01/B02/B12/W01). |
| CF-B03 | **PARTIAL / REOPENED** | Many inputs are derived, but 37 implemented gap endpoints have zero flags (PFM-B02). |
| CF-B04 | **PARTIAL / REOPENED** | GraphQL docs/variables are source-derived, but pagination, nested results, secret policy, and Int domain remain wrong (PFM-B04-B08). |
| CF-B05 | **CLOSED** | releases assets upload binds required file/name/release ID, root confinement, uploads host, and 64 MiB cap. |
| CF-B06 | **CLOSED in mapping lens** | Reverse ETL uses generic endpoint resolution/validator/preview/writer; no GitHub hard-coding. |
| CF-B07 | **CLOSED in mapping lens** | Destination auth/config flows through generic resolved runtime; no mapping regression. |
| CF-B08 | **PARTIAL / REOPENED** | Configured credential masking and ordinary field preservation fixed; generated GraphQL secrets unclassified (PFM-B06). |
| CF-B09 | **CLOSED** | No current mapping defect beyond explicit GraphQL result/secret findings. |
| CF-B10 | **CLOSED** | Attempt/run identity remains through reverse execution. |
| CF-B11 | **CLOSED for ordinary receipts; new GraphQL omission** | Receipts/occurrence IDs survive when selected; mutation nested identity omitted (PFM-B07). |
| CF-B19 | **CLOSED for exact enum; new GraphQL defect** | Exact numeric enum handling fixed; signed-32-bit Int is distinct (PFM-B08). |
| CF-B20 | **CLOSED** | Duplicate occurrence validation present and covered. |
| CF-B21 | **CLOSED** | Path/query byte bounds enforced. |
| CF-B22 | **PARTIAL / REOPENED** | REST bodies close, but source gaps advertise empty inputs and command query fields can escape source (PFM-B02/W01). |
| CF-B23 | **CLOSED for ordinary IDs** | Heuristic deletion removed; explicit GraphQL secret classification separately incomplete (PFM-B06). |
| CF-W01 | **CLOSED for determinism; new warning** | Registry-driven exact skill generation passes; renderer drops metadata (PFM-W02). |
| CF-W07 | **PARTIAL / REOPENED** | Website generation deterministic, but semantic flag constraints drop (PFM-W02). |
| CF-W08 | **PARTIAL / REOPENED** | reviewed_sha binds, but current code/base/components and certification freshness do not (PFM-B10/B11). |

### Final verdict

**blockers**

The mapping/generation lens is not shippable at 8a8a866ff6d5282c28bda12acceed8a624218f01. The required parity gate is red, and remaining blockers affect API reachability, provider request validity, backward pagination, secret handling, resource-result preservation, source provenance, and certification freshness. All 12 blockers require the exact regressions above before this lens can be clean.

---

_Reviewed: 2026-08-21T02:59:22Z_  
_Reviewer: gsd-code-reviewer (deep mapping/generation lens)_  
_Frozen diff: e62ae21d428f0d27225f9bff564dc2cd797f6b65..8a8a866ff6d5282c28bda12acceed8a624218f01_
