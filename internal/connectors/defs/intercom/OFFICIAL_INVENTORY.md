# Intercom official API inventory

Official sources inventoried on 2026-08-01:

- Intercom REST page data: `https://developers.intercom.com/page-data/docs/references/rest-api/api.intercom.io/data.json`
- Intercom OpenAPI shared JSON: `https://developers.intercom.com/page-data/shared/oas-docs/references/%402.16/rest-api/api.intercom.io.yaml.json`
- OpenAPI title/version: `Intercom API` / `2.16`

## Reconciled lane counts

| etl_read | reverse_etl_write | direct_read_query_search | binary_file | cdc_changefeed | excluded_not_applicable | total |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 55 | 114 | 42 | 7 | 12 | 1 | 231 |

## Connector-local implementation disposition

| implemented ETL streams | typed write actions declared | blocked/planned operation-ledger rows | live certified |
| ---: | ---: | ---: | ---: |
| 5 | 114 | 112 | 0 |

Destructive/delete operations are included in the official operation inventory. They are not blanket-excluded as unsafe; implemented destructive write actions declare `confirm: "destructive"` and remain behind reverse ETL plan -> preview -> explicit approval -> execute. Direct, binary, and CDC-like operation rows remain blocked by default until their shared execution foundations and fixture evidence are present.

## Operation inventory

| # | Lane | Method | Path | Operation ID | Summary | Disposition |
| ---: | --- | --- | --- | --- | --- | --- |
| 1 | `etl_read` | `GET` | `/admins` | `listAdmins` | List all admins | covered_by stream `admins` |
| 2 | `cdc_changefeed` | `GET` | `/admins/activity_log_event_types` | `listActivityLogEventTypes` | List all activity log event types | operation-ledger blocked/planned |
| 3 | `cdc_changefeed` | `GET` | `/admins/activity_logs` | `listActivityLogs` | List all activity logs | operation-ledger blocked/planned |
| 4 | `cdc_changefeed` | `POST` | `/admins/activity_logs/search` | `searchActivityLogs` | Search activity logs | operation-ledger blocked/planned |
| 5 | `direct_read_query_search` | `GET` | `/admins/{admin_id}` | `retrieveAdmin` | Retrieve an admin | operation-ledger blocked/planned |
| 6 | `reverse_etl_write` | `PUT` | `/admins/{admin_id}/away` | `setAwayAdmin` | Set an admin to away | covered_by write `set_away_admin` |
| 7 | `etl_read` | `GET` | `/ai/content_import_sources` | `listContentImportSources` | List content import sources | operation-ledger blocked/planned |
| 8 | `reverse_etl_write` | `POST` | `/ai/content_import_sources` | `createContentImportSource` | Create a content import source | covered_by write `create_content_import_source` |
| 9 | `reverse_etl_write` | `DELETE` | `/ai/content_import_sources/{source_id}` | `deleteContentImportSource` | Delete a content import source | covered_by write `delete_content_import_source` |
| 10 | `direct_read_query_search` | `GET` | `/ai/content_import_sources/{source_id}` | `getContentImportSource` | Retrieve a content import source | operation-ledger blocked/planned |
| 11 | `reverse_etl_write` | `PUT` | `/ai/content_import_sources/{source_id}` | `updateContentImportSource` | Update a content import source | covered_by write `update_content_import_source` |
| 12 | `etl_read` | `GET` | `/ai/external_pages` | `listExternalPages` | List external pages | operation-ledger blocked/planned |
| 13 | `reverse_etl_write` | `POST` | `/ai/external_pages` | `createExternalPage` | Create an external page (or update an external page by external ID) | covered_by write `create_external_page` |
| 14 | `reverse_etl_write` | `DELETE` | `/ai/external_pages/{page_id}` | `deleteExternalPage` | Delete an external page | covered_by write `delete_external_page` |
| 15 | `direct_read_query_search` | `GET` | `/ai/external_pages/{page_id}` | `getExternalPage` | Retrieve an external page | operation-ledger blocked/planned |
| 16 | `reverse_etl_write` | `PUT` | `/ai/external_pages/{page_id}` | `updateExternalPage` | Update an external page | covered_by write `update_external_page` |
| 17 | `etl_read` | `GET` | `/articles` | `listArticles` | List all articles | operation-ledger blocked/planned |
| 18 | `reverse_etl_write` | `POST` | `/articles` | `createArticle` | Create an article | covered_by write `create_article` |
| 19 | `etl_read` | `GET` | `/articles/search` | `searchArticles` | Search for articles | operation-ledger blocked/planned |
| 20 | `reverse_etl_write` | `DELETE` | `/articles/{article_id}` | `deleteArticle` | Delete an article | covered_by write `delete_article` |
| 21 | `direct_read_query_search` | `GET` | `/articles/{article_id}` | `retrieveArticle` | Retrieve an article | operation-ledger blocked/planned |
| 22 | `reverse_etl_write` | `PUT` | `/articles/{article_id}` | `updateArticle` | Update an article | covered_by write `update_article` |
| 23 | `reverse_etl_write` | `POST` | `/articles/{article_id}/tags` | `attachTagToArticle` | Add a tag to an article | covered_by write `attach_tag_to_article` |
| 24 | `reverse_etl_write` | `DELETE` | `/articles/{article_id}/tags/{id}` | `detachTagFromArticle` | Remove a tag from an article | covered_by write `detach_tag_from_article` |
| 25 | `etl_read` | `GET` | `/articles/{article_id}/versions` | `listArticleVersions` | List article versions | operation-ledger blocked/planned |
| 26 | `direct_read_query_search` | `GET` | `/articles/{article_id}/versions/{id}` | `retrieveArticleVersion` | Retrieve an article version | operation-ledger blocked/planned |
| 27 | `direct_read_query_search` | `GET` | `/articles/{id}/draft` | `retrieveArticleDraft` | Retrieve an article draft | operation-ledger blocked/planned |
| 28 | `reverse_etl_write` | `PUT` | `/articles/{id}/draft` | `stageArticleDraft` | Stage an article draft | covered_by write `stage_article_draft` |
| 29 | `reverse_etl_write` | `POST` | `/articles/{id}/draft/publish` | `publishArticleDraft` | Publish an article draft | covered_by write `publish_article_draft` |
| 30 | `etl_read` | `GET` | `/audiences` | `listAudiences` | List all audiences | operation-ledger blocked/planned |
| 31 | `reverse_etl_write` | `POST` | `/audiences` | `createAudience` | Create an audience | covered_by write `create_audience` |
| 32 | `reverse_etl_write` | `DELETE` | `/audiences/{id}` | `deleteAudience` | Delete an audience | covered_by write `delete_audience` |
| 33 | `direct_read_query_search` | `GET` | `/audiences/{id}` | `retrieveAudience` | Retrieve an audience | operation-ledger blocked/planned |
| 34 | `reverse_etl_write` | `PUT` | `/audiences/{id}` | `updateAudience` | Update an audience | covered_by write `update_audience` |
| 35 | `etl_read` | `GET` | `/away_status_reasons` | `listAwayStatusReasons` | List all away status reasons | operation-ledger blocked/planned |
| 36 | `etl_read` | `GET` | `/brands` | `listBrands` | List all brands | operation-ledger blocked/planned |
| 37 | `direct_read_query_search` | `GET` | `/brands/{id}` | `retrieveBrand` | Retrieve a brand | operation-ledger blocked/planned |
| 38 | `etl_read` | `GET` | `/calls` | `listCalls` | List all calls | operation-ledger blocked/planned |
| 39 | `etl_read` | `POST` | `/calls/search` | `listCallsWithTranscripts` | List calls with transcripts | operation-ledger blocked/planned |
| 40 | `direct_read_query_search` | `GET` | `/calls/{call_id}` | `showCall` | Get a call | operation-ledger blocked/planned |
| 41 | `direct_read_query_search` | `GET` | `/calls/{call_id}/recording` | `showCallRecording` | Get call recording by call id | operation-ledger blocked/planned |
| 42 | `direct_read_query_search` | `GET` | `/calls/{call_id}/transcript` | `showCallTranscript` | Get call transcript by call id | operation-ledger blocked/planned |
| 43 | `etl_read` | `GET` | `/companies` | `retrieveCompany` | Retrieve companies | covered_by stream `companies` |
| 44 | `reverse_etl_write` | `POST` | `/companies` | `createOrUpdateCompany` | Create or Update a company | covered_by write `create_or_update_company` |
| 45 | `etl_read` | `POST` | `/companies/list` | `listAllCompanies` | List all companies | operation-ledger blocked/planned |
| 46 | `etl_read` | `GET` | `/companies/scroll` | `scrollOverAllCompanies` | Scroll over all companies | operation-ledger blocked/planned |
| 47 | `reverse_etl_write` | `DELETE` | `/companies/{company_id}` | `deleteCompany` | Delete a company | covered_by write `delete_company` |
| 48 | `direct_read_query_search` | `GET` | `/companies/{company_id}` | `RetrieveACompanyById` | Retrieve a company by ID | operation-ledger blocked/planned |
| 49 | `reverse_etl_write` | `PUT` | `/companies/{company_id}` | `UpdateCompany` | Update a company | covered_by write `update_company` |
| 50 | `etl_read` | `GET` | `/companies/{company_id}/contacts` | `ListAttachedContacts` | List attached contacts | operation-ledger blocked/planned |
| 51 | `etl_read` | `GET` | `/companies/{company_id}/notes` | `listCompanyNotes` | List all company notes | operation-ledger blocked/planned |
| 52 | `reverse_etl_write` | `POST` | `/companies/{company_id}/notes` | `createCompanyNote` | Create a company note | covered_by write `create_company_note` |
| 53 | `etl_read` | `GET` | `/companies/{company_id}/segments` | `ListAttachedSegmentsForCompanies` | List attached segments for companies | operation-ledger blocked/planned |
| 54 | `etl_read` | `GET` | `/contacts` | `ListContacts` | List all contacts | covered_by stream `contacts` |
| 55 | `reverse_etl_write` | `POST` | `/contacts` | `CreateContact` | Create contact | covered_by write `create_contact` |
| 56 | `direct_read_query_search` | `GET` | `/contacts/find_by_external_id/{external_id}` | `ShowContactByExternalId` | Get a contact by External ID | operation-ledger blocked/planned |
| 57 | `reverse_etl_write` | `POST` | `/contacts/merge` | `MergeContact` | Merge a lead and a user | covered_by write `merge_contact` |
| 58 | `etl_read` | `POST` | `/contacts/search` | `SearchContacts` | Search contacts | operation-ledger blocked/planned |
| 59 | `reverse_etl_write` | `DELETE` | `/contacts/{contact_id}` | `DeleteContact` | Delete a contact | covered_by write `delete_contact` |
| 60 | `direct_read_query_search` | `GET` | `/contacts/{contact_id}` | `ShowContact` | Get a contact | operation-ledger blocked/planned |
| 61 | `reverse_etl_write` | `PUT` | `/contacts/{contact_id}` | `UpdateContact` | Update a contact | covered_by write `update_contact` |
| 62 | `reverse_etl_write` | `POST` | `/contacts/{contact_id}/archive` | `ArchiveContact` | Archive contact | covered_by write `archive_contact` |
| 63 | `reverse_etl_write` | `POST` | `/contacts/{contact_id}/block` | `BlockContact` | Block contact | covered_by write `block_contact` |
| 64 | `etl_read` | `GET` | `/contacts/{contact_id}/companies` | `listCompaniesForAContact` | List attached companies for contact | operation-ledger blocked/planned |
| 65 | `reverse_etl_write` | `POST` | `/contacts/{contact_id}/companies` | `attachContactToACompany` | Attach a Contact to a Company | covered_by write `attach_contact_to_acompany` |
| 66 | `reverse_etl_write` | `DELETE` | `/contacts/{contact_id}/companies/{company_id}` | `detachContactFromACompany` | Detach a contact from a company | covered_by write `detach_contact_from_acompany` |
| 67 | `etl_read` | `GET` | `/contacts/{contact_id}/notes` | `listNotes` | List all notes | operation-ledger blocked/planned |
| 68 | `reverse_etl_write` | `POST` | `/contacts/{contact_id}/notes` | `createNote` | Create a note | covered_by write `create_note` |
| 69 | `etl_read` | `GET` | `/contacts/{contact_id}/segments` | `listSegmentsForAContact` | List attached segments for contact | operation-ledger blocked/planned |
| 70 | `etl_read` | `GET` | `/contacts/{contact_id}/subscriptions` | `listSubscriptionsForAContact` | List subscriptions for a contact | operation-ledger blocked/planned |
| 71 | `reverse_etl_write` | `POST` | `/contacts/{contact_id}/subscriptions` | `attachSubscriptionTypeToContact` | Add subscription to a contact | covered_by write `attach_subscription_type_to_contact` |
| 72 | `reverse_etl_write` | `DELETE` | `/contacts/{contact_id}/subscriptions/{subscription_id}` | `detachSubscriptionTypeToContact` | Remove subscription from a contact | covered_by write `detach_subscription_type_to_contact` |
| 73 | `etl_read` | `GET` | `/contacts/{contact_id}/tags` | `listTagsForAContact` | List tags attached to a contact | operation-ledger blocked/planned |
| 74 | `reverse_etl_write` | `POST` | `/contacts/{contact_id}/tags` | `attachTagToContact` | Add tag to a contact | covered_by write `attach_tag_to_contact` |
| 75 | `reverse_etl_write` | `DELETE` | `/contacts/{contact_id}/tags/{tag_id}` | `detachTagFromContact` | Remove tag from a contact | covered_by write `detach_tag_from_contact` |
| 76 | `reverse_etl_write` | `POST` | `/contacts/{contact_id}/unarchive` | `UnarchiveContact` | Unarchive contact | covered_by write `unarchive_contact` |
| 77 | `etl_read` | `GET` | `/contacts/{id}/banners` | `listContactBanners` | List banners for a contact | operation-ledger blocked/planned |
| 78 | `reverse_etl_write` | `POST` | `/contacts/{id}/banners/{view_id}/dismiss` | `dismissContactBanner` | Dismiss a banner for a contact | covered_by write `dismiss_contact_banner` |
| 79 | `etl_read` | `GET` | `/contacts/{id}/merge_history` | `ListContactMergeHistory` | Get contact merge history | operation-ledger blocked/planned |
| 80 | `reverse_etl_write` | `POST` | `/content/bulk_actions` | `bulkContentActions` | Run a bulk action on Knowledge Hub content | covered_by write `bulk_content_actions` |
| 81 | `etl_read` | `GET` | `/content/search` | `searchContent` | Search knowledge base contents | operation-ledger blocked/planned |
| 82 | `etl_read` | `GET` | `/content_snippets` | `listContentSnippets` | List all content snippets | operation-ledger blocked/planned |
| 83 | `reverse_etl_write` | `POST` | `/content_snippets` | `createContentSnippet` | Create a content snippet | covered_by write `create_content_snippet` |
| 84 | `reverse_etl_write` | `POST` | `/content_snippets/{content_snippet_id}/tags` | `attachTagToContentSnippet` | Add a tag to a content snippet | covered_by write `attach_tag_to_content_snippet` |
| 85 | `reverse_etl_write` | `DELETE` | `/content_snippets/{content_snippet_id}/tags/{id}` | `detachTagFromContentSnippet` | Remove a tag from a content snippet | covered_by write `detach_tag_from_content_snippet` |
| 86 | `reverse_etl_write` | `DELETE` | `/content_snippets/{id}` | `deleteContentSnippet` | Delete a content snippet | covered_by write `delete_content_snippet` |
| 87 | `direct_read_query_search` | `GET` | `/content_snippets/{id}` | `getContentSnippet` | Retrieve a content snippet | operation-ledger blocked/planned |
| 88 | `reverse_etl_write` | `PUT` | `/content_snippets/{id}` | `updateContentSnippet` | Update a content snippet | covered_by write `update_content_snippet` |
| 89 | `etl_read` | `GET` | `/conversations` | `listConversations` | List all conversations | covered_by stream `conversations` |
| 90 | `reverse_etl_write` | `POST` | `/conversations` | `createConversation` | Creates a conversation | covered_by write `create_conversation` |
| 91 | `etl_read` | `GET` | `/conversations/attributes` | `listConversationAttributes` | List all conversation attributes | operation-ledger blocked/planned |
| 92 | `reverse_etl_write` | `POST` | `/conversations/attributes` | `createConversationAttribute` | Create a conversation attribute | covered_by write `create_conversation_attribute` |
| 93 | `reverse_etl_write` | `DELETE` | `/conversations/attributes/{id}` | `deleteConversationAttribute` | Delete (archive) a conversation attribute | covered_by write `delete_conversation_attribute` |
| 94 | `direct_read_query_search` | `GET` | `/conversations/attributes/{id}` | `getConversationAttribute` | Get a conversation attribute | operation-ledger blocked/planned |
| 95 | `reverse_etl_write` | `PUT` | `/conversations/attributes/{id}` | `updateConversationAttribute` | Update a conversation attribute | covered_by write `update_conversation_attribute` |
| 96 | `reverse_etl_write` | `POST` | `/conversations/attributes/{id}/options` | `createConversationAttributeOption` | Add an option to a list conversation attribute | covered_by write `create_conversation_attribute_option` |
| 97 | `reverse_etl_write` | `DELETE` | `/conversations/attributes/{id}/options/{option_id}` | `deleteConversationAttributeOption` | Archive an option on a list conversation attribute | covered_by write `delete_conversation_attribute_option` |
| 98 | `reverse_etl_write` | `PUT` | `/conversations/attributes/{id}/options/{option_id}` | `updateConversationAttributeOption` | Update an option on a list conversation attribute | covered_by write `update_conversation_attribute_option` |
| 99 | `cdc_changefeed` | `GET` | `/conversations/deleted` | `listDeletedConversationIds` | List all deleted conversation IDs | operation-ledger blocked/planned |
| 100 | `reverse_etl_write` | `POST` | `/conversations/redact` | `redactConversation` | Redact a conversation part | covered_by write `redact_conversation` |
| 101 | `etl_read` | `POST` | `/conversations/search` | `searchConversations` | Search conversations | operation-ledger blocked/planned |
| 102 | `reverse_etl_write` | `DELETE` | `/conversations/{conversation_id}` | `deleteConversation` | Delete a conversation | covered_by write `delete_conversation` |
| 103 | `direct_read_query_search` | `GET` | `/conversations/{conversation_id}` | `retrieveConversation` | Retrieve a conversation | operation-ledger blocked/planned |
| 104 | `reverse_etl_write` | `PUT` | `/conversations/{conversation_id}` | `updateConversation` | Update a conversation | covered_by write `update_conversation` |
| 105 | `reverse_etl_write` | `POST` | `/conversations/{conversation_id}/convert` | `convertConversationToTicket` | Convert a conversation to a ticket | covered_by write `convert_conversation_to_ticket` |
| 106 | `reverse_etl_write` | `POST` | `/conversations/{conversation_id}/customers` | `attachContactToConversation` | Attach a contact to a conversation | covered_by write `attach_contact_to_conversation` |
| 107 | `reverse_etl_write` | `DELETE` | `/conversations/{conversation_id}/customers/{contact_id}` | `detachContactFromConversation` | Detach a contact from a group conversation | covered_by write `detach_contact_from_conversation` |
| 108 | `reverse_etl_write` | `POST` | `/conversations/{conversation_id}/parts` | `manageConversation` | Manage a conversation | covered_by write `manage_conversation` |
| 109 | `reverse_etl_write` | `POST` | `/conversations/{conversation_id}/reply` | `replyConversation` | Reply to a conversation | covered_by write `reply_conversation` |
| 110 | `reverse_etl_write` | `POST` | `/conversations/{conversation_id}/tags` | `attachTagToConversation` | Add tag to a conversation | covered_by write `attach_tag_to_conversation` |
| 111 | `reverse_etl_write` | `DELETE` | `/conversations/{conversation_id}/tags/{tag_id}` | `detachTagFromConversation` | Remove tag from a conversation | covered_by write `detach_tag_from_conversation` |
| 112 | `etl_read` | `GET` | `/conversations/{id}/handling_events` | `listHandlingEvents` | List handling events | operation-ledger blocked/planned |
| 113 | `reverse_etl_write` | `POST` | `/conversations/{id}/merge` | `mergeConversation` | Merge a conversation | covered_by write `merge_conversation` |
| 114 | `etl_read` | `GET` | `/conversations/{id}/side_conversations` | `listSideConversations` | List side conversations | operation-ledger blocked/planned |
| 115 | `reverse_etl_write` | `DELETE` | `/custom_object_instances/{custom_object_type_identifier}` | `deleteCustomObjectInstancesById` | Delete a Custom Object Instance by External ID | covered_by write `delete_custom_object_instances_by_id` |
| 116 | `etl_read` | `GET` | `/custom_object_instances/{custom_object_type_identifier}` | `listCustomObjectInstances` | List Custom Object Instances | operation-ledger blocked/planned |
| 117 | `reverse_etl_write` | `POST` | `/custom_object_instances/{custom_object_type_identifier}` | `createCustomObjectInstances` | Create or Update a Custom Object Instance | covered_by write `create_custom_object_instances` |
| 118 | `reverse_etl_write` | `DELETE` | `/custom_object_instances/{custom_object_type_identifier}/{custom_object_instance_id}` | `deleteCustomObjectInstancesByExternalId` | Delete a Custom Object Instance by ID | covered_by write `delete_custom_object_instances_by_external_id` |
| 119 | `direct_read_query_search` | `GET` | `/custom_object_instances/{custom_object_type_identifier}/{custom_object_instance_id}` | `getCustomObjectInstancesById` | Get Custom Object Instance by ID | operation-ledger blocked/planned |
| 120 | `etl_read` | `GET` | `/data_attributes` | `lisDataAttributes` | List all data attributes | operation-ledger blocked/planned |
| 121 | `reverse_etl_write` | `POST` | `/data_attributes` | `createDataAttribute` | Create a data attribute | covered_by write `create_data_attribute` |
| 122 | `reverse_etl_write` | `PUT` | `/data_attributes/{data_attribute_id}` | `updateDataAttribute` | Update a data attribute | covered_by write `update_data_attribute` |
| 123 | `cdc_changefeed` | `GET` | `/data_connectors` | `listDataConnectors` | List all data connectors | operation-ledger blocked/planned |
| 124 | `reverse_etl_write` | `POST` | `/data_connectors` | `createDataConnector` | Create a data connector | covered_by write `create_data_connector` |
| 125 | `cdc_changefeed` | `GET` | `/data_connectors/{data_connector_id}/execution_results` | `listDataConnectorExecutionResults` | List execution results for a data connector | operation-ledger blocked/planned |
| 126 | `cdc_changefeed` | `GET` | `/data_connectors/{data_connector_id}/execution_results/{id}` | `showDataConnectorExecutionResult` | Retrieve an execution result | operation-ledger blocked/planned |
| 127 | `reverse_etl_write` | `DELETE` | `/data_connectors/{id}` | `deleteDataConnector` | Delete a data connector | covered_by write `delete_data_connector` |
| 128 | `cdc_changefeed` | `GET` | `/data_connectors/{id}` | `RetrieveDataConnector` | Retrieve a data connector | operation-ledger blocked/planned |
| 129 | `reverse_etl_write` | `PATCH` | `/data_connectors/{id}` | `updateDataConnector` | Update a data connector | covered_by write `update_data_connector` |
| 130 | `binary_file` | `GET` | `/download/content/data/{job_identifier}` | `downloadDataExport` | Download content data export | operation-ledger blocked/planned |
| 131 | `binary_file` | `GET` | `/download/reporting_data/{job_identifier}` | `not-recorded` | Download completed export job data | operation-ledger blocked/planned |
| 132 | `etl_read` | `GET` | `/emails` | `listEmails` | List all email settings | operation-ledger blocked/planned |
| 133 | `direct_read_query_search` | `GET` | `/emails/{id}` | `retrieveEmail` | Retrieve an email setting | operation-ledger blocked/planned |
| 134 | `cdc_changefeed` | `GET` | `/events` | `lisDataEvents` | List all data events | operation-ledger blocked/planned |
| 135 | `reverse_etl_write` | `POST` | `/events` | `createDataEvent` | Submit a data event | covered_by write `create_data_event` |
| 136 | `cdc_changefeed` | `POST` | `/events/summaries` | `dataEventSummaries` | Create event summaries | operation-ledger blocked/planned |
| 137 | `reverse_etl_write` | `POST` | `/export/cancel/{job_identifier}` | `cancelDataExport` | Cancel content data export | covered_by write `cancel_data_export` |
| 138 | `binary_file` | `POST` | `/export/content/data` | `createDataExport` | Create content data export | operation-ledger blocked/planned |
| 139 | `binary_file` | `GET` | `/export/content/data/{job_identifier}` | `getDataExport` | Show content data export | operation-ledger blocked/planned |
| 140 | `binary_file` | `POST` | `/export/reporting_data/enqueue` | `not-recorded` | Enqueue a new reporting data export job | operation-ledger blocked/planned |
| 141 | `etl_read` | `GET` | `/export/reporting_data/get_datasets` | `not-recorded` | List available datasets and attributes | operation-ledger blocked/planned |
| 142 | `binary_file` | `GET` | `/export/reporting_data/{job_identifier}` | `not-recorded` | Get export job status | operation-ledger blocked/planned |
| 143 | `binary_file` | `GET` | `/export/workflows/{id}` | `exportWorkflow` | Export a workflow | operation-ledger blocked/planned |
| 144 | `reverse_etl_write` | `POST` | `/fin/csat` | `submitFinCsat` | Submit a CSAT rating | covered_by write `submit_fin_csat` |
| 145 | `reverse_etl_write` | `POST` | `/fin/reply` | `replyToFin` | Reply to Fin | covered_by write `reply_to_fin` |
| 146 | `reverse_etl_write` | `POST` | `/fin/start` | `startFinConversation` | Start a conversation with Fin | covered_by write `start_fin_conversation` |
| 147 | `direct_read_query_search` | `GET` | `/fin_voice/collect/{id}` | `collectFinVoiceCallById` | Collect Fin Voice call by ID | operation-ledger blocked/planned |
| 148 | `direct_read_query_search` | `GET` | `/fin_voice/conversation/{conversation_id}` | `collectFinVoiceCallsByConversationId` | Collect Fin Voice calls by conversation ID | operation-ledger blocked/planned |
| 149 | `direct_read_query_search` | `GET` | `/fin_voice/external_id/{external_id}` | `collectFinVoiceCallByExternalId` | Collect Fin Voice call by external ID | operation-ledger blocked/planned |
| 150 | `direct_read_query_search` | `GET` | `/fin_voice/phone_number/{phone_number}` | `collectFinVoiceCallByPhoneNumber` | Collect Fin Voice call by phone number | operation-ledger blocked/planned |
| 151 | `reverse_etl_write` | `POST` | `/fin_voice/register` | `registerFinVoiceCall` | Register a Fin Voice call | covered_by write `register_fin_voice_call` |
| 152 | `etl_read` | `GET` | `/help_center/collections` | `listAllCollections` | List all collections | operation-ledger blocked/planned |
| 153 | `reverse_etl_write` | `POST` | `/help_center/collections` | `createCollection` | Create a collection | covered_by write `create_collection` |
| 154 | `reverse_etl_write` | `DELETE` | `/help_center/collections/{collection_id}` | `deleteCollection` | Delete a collection | covered_by write `delete_collection` |
| 155 | `direct_read_query_search` | `GET` | `/help_center/collections/{collection_id}` | `retrieveCollection` | Retrieve a collection | operation-ledger blocked/planned |
| 156 | `reverse_etl_write` | `PUT` | `/help_center/collections/{collection_id}` | `updateCollection` | Update a collection | covered_by write `update_collection` |
| 157 | `etl_read` | `GET` | `/help_center/help_centers` | `listHelpCenters` | List all Help Centers | operation-ledger blocked/planned |
| 158 | `direct_read_query_search` | `GET` | `/help_center/help_centers/{help_center_id}` | `retrieveHelpCenter` | Retrieve a Help Center | operation-ledger blocked/planned |
| 159 | `etl_read` | `GET` | `/help_center/help_centers/{help_center_id}/redirects` | `listHelpCenterRedirects` | List all redirects for a help center | operation-ledger blocked/planned |
| 160 | `reverse_etl_write` | `POST` | `/help_center/help_centers/{help_center_id}/redirects` | `createHelpCenterRedirect` | Create a redirect | covered_by write `create_help_center_redirect` |
| 161 | `reverse_etl_write` | `DELETE` | `/help_center/help_centers/{help_center_id}/redirects/{id}` | `deleteHelpCenterRedirect` | Delete a redirect | covered_by write `delete_help_center_redirect` |
| 162 | `direct_read_query_search` | `GET` | `/help_center/help_centers/{help_center_id}/redirects/{id}` | `retrieveHelpCenterRedirect` | Retrieve a redirect | operation-ledger blocked/planned |
| 163 | `etl_read` | `GET` | `/internal_articles` | `listInternalArticles` | List all articles | operation-ledger blocked/planned |
| 164 | `reverse_etl_write` | `POST` | `/internal_articles` | `createInternalArticle` | Create an internal article | covered_by write `create_internal_article` |
| 165 | `etl_read` | `GET` | `/internal_articles/search` | `searchInternalArticles` | Search for internal articles | operation-ledger blocked/planned |
| 166 | `reverse_etl_write` | `DELETE` | `/internal_articles/{internal_article_id}` | `deleteInternalArticle` | Delete an internal article | covered_by write `delete_internal_article` |
| 167 | `direct_read_query_search` | `GET` | `/internal_articles/{internal_article_id}` | `retrieveInternalArticle` | Retrieve an internal article | operation-ledger blocked/planned |
| 168 | `reverse_etl_write` | `PUT` | `/internal_articles/{internal_article_id}` | `updateInternalArticle` | Update an internal article | covered_by write `update_internal_article` |
| 169 | `reverse_etl_write` | `POST` | `/internal_articles/{internal_article_id}/tags` | `attachTagToInternalArticle` | Add a tag to an internal article | covered_by write `attach_tag_to_internal_article` |
| 170 | `reverse_etl_write` | `DELETE` | `/internal_articles/{internal_article_id}/tags/{id}` | `detachTagFromInternalArticle` | Remove a tag from an internal article | covered_by write `detach_tag_from_internal_article` |
| 171 | `direct_read_query_search` | `GET` | `/ip_allowlist` | `getIpAllowlist` | Get IP allowlist settings | operation-ledger blocked/planned |
| 172 | `reverse_etl_write` | `PUT` | `/ip_allowlist` | `updateIpAllowlist` | Update IP allowlist settings | covered_by write `update_ip_allowlist` |
| 173 | `cdc_changefeed` | `GET` | `/jobs/status/{job_id}` | `jobsStatus` | Retrieve job status | operation-ledger blocked/planned |
| 174 | `etl_read` | `GET` | `/macros` | `listMacros` | List all macros | operation-ledger blocked/planned |
| 175 | `direct_read_query_search` | `GET` | `/macros/{id}` | `getMacro` | Retrieve a macro | operation-ledger blocked/planned |
| 176 | `excluded_not_applicable` | `GET` | `/me` | `identifyAdmin` | Identify an admin | operation-ledger duplicate of `GET /admins` |
| 177 | `reverse_etl_write` | `POST` | `/messages` | `createMessage` | Create a message | covered_by write `create_message` |
| 178 | `cdc_changefeed` | `GET` | `/messages/status` | `getWhatsAppMessageStatus` | Get statuses of all messages sent based on the specified ruleset_id | operation-ledger blocked/planned |
| 179 | `direct_read_query_search` | `GET` | `/messages/whatsapp/status` | `RetrieveWhatsAppMessageStatus` | Retrieve WhatsApp message delivery status | operation-ledger blocked/planned |
| 180 | `etl_read` | `GET` | `/news/news_items` | `listNewsItems` | List all news items | operation-ledger blocked/planned |
| 181 | `reverse_etl_write` | `POST` | `/news/news_items` | `createNewsItem` | Create a news item | covered_by write `create_news_item` |
| 182 | `reverse_etl_write` | `DELETE` | `/news/news_items/{news_item_id}` | `deleteNewsItem` | Delete a news item | covered_by write `delete_news_item` |
| 183 | `direct_read_query_search` | `GET` | `/news/news_items/{news_item_id}` | `retrieveNewsItem` | Retrieve a news item | operation-ledger blocked/planned |
| 184 | `reverse_etl_write` | `PUT` | `/news/news_items/{news_item_id}` | `updateNewsItem` | Update a news item | covered_by write `update_news_item` |
| 185 | `etl_read` | `GET` | `/news/newsfeeds` | `listNewsfeeds` | List all newsfeeds | operation-ledger blocked/planned |
| 186 | `direct_read_query_search` | `GET` | `/news/newsfeeds/{newsfeed_id}` | `retrieveNewsfeed` | Retrieve a newsfeed | operation-ledger blocked/planned |
| 187 | `etl_read` | `GET` | `/news/newsfeeds/{newsfeed_id}/items` | `listLiveNewsfeedItems` | List all live newsfeed items | operation-ledger blocked/planned |
| 188 | `direct_read_query_search` | `GET` | `/notes/{note_id}` | `retrieveNote` | Retrieve a note | operation-ledger blocked/planned |
| 189 | `etl_read` | `GET` | `/office_hours_schedules` | `listOfficeHoursSchedules` | List all office hours schedules | operation-ledger blocked/planned |
| 190 | `reverse_etl_write` | `POST` | `/office_hours_schedules` | `createOfficeHoursSchedule` | Create an office hours schedule | covered_by write `create_office_hours_schedule` |
| 191 | `reverse_etl_write` | `DELETE` | `/office_hours_schedules/{id}` | `deleteOfficeHoursSchedule` | Delete an office hours schedule | covered_by write `delete_office_hours_schedule` |
| 192 | `direct_read_query_search` | `GET` | `/office_hours_schedules/{id}` | `getOfficeHoursSchedule` | Retrieve an office hours schedule | operation-ledger blocked/planned |
| 193 | `reverse_etl_write` | `PUT` | `/office_hours_schedules/{id}` | `updateOfficeHoursSchedule` | Update an office hours schedule | covered_by write `update_office_hours_schedule` |
| 194 | `etl_read` | `GET` | `/office_hours_schedules/{office_hours_schedule_id}/office_hours_exceptions` | `listOfficeHoursExceptions` | List all office hours exceptions | operation-ledger blocked/planned |
| 195 | `reverse_etl_write` | `POST` | `/office_hours_schedules/{office_hours_schedule_id}/office_hours_exceptions` | `createOfficeHoursException` | Create an office hours exception | covered_by write `create_office_hours_exception` |
| 196 | `reverse_etl_write` | `DELETE` | `/office_hours_schedules/{office_hours_schedule_id}/office_hours_exceptions/{id}` | `deleteOfficeHoursException` | Delete an office hours exception | covered_by write `delete_office_hours_exception` |
| 197 | `direct_read_query_search` | `GET` | `/office_hours_schedules/{office_hours_schedule_id}/office_hours_exceptions/{id}` | `getOfficeHoursException` | Retrieve an office hours exception | operation-ledger blocked/planned |
| 198 | `reverse_etl_write` | `PUT` | `/office_hours_schedules/{office_hours_schedule_id}/office_hours_exceptions/{id}` | `updateOfficeHoursException` | Update an office hours exception | covered_by write `update_office_hours_exception` |
| 199 | `reverse_etl_write` | `POST` | `/phone_call_redirects` | `createPhoneSwitch` | Create a phone Switch | covered_by write `create_phone_switch` |
| 200 | `etl_read` | `GET` | `/segments` | `listSegments` | List all segments | operation-ledger blocked/planned |
| 201 | `direct_read_query_search` | `GET` | `/segments/{segment_id}` | `retrieveSegment` | Retrieve a segment | operation-ledger blocked/planned |
| 202 | `etl_read` | `GET` | `/subscription_types` | `listSubscriptionTypes` | List subscription types | operation-ledger blocked/planned |
| 203 | `etl_read` | `GET` | `/tags` | `listTags` | List all tags | covered_by stream `tags` |
| 204 | `reverse_etl_write` | `POST` | `/tags` | `createTag` | Create or update a tag, Tag or untag companies, Tag contacts | covered_by write `create_tag` |
| 205 | `reverse_etl_write` | `DELETE` | `/tags/{tag_id}` | `deleteTag` | Delete tag | covered_by write `delete_tag` |
| 206 | `direct_read_query_search` | `GET` | `/tags/{tag_id}` | `findTag` | Find a specific tag | operation-ledger blocked/planned |
| 207 | `etl_read` | `GET` | `/teams` | `listTeams` | List all teams | operation-ledger blocked/planned |
| 208 | `direct_read_query_search` | `GET` | `/teams/{team_id}` | `retrieveTeam` | Retrieve a team | operation-ledger blocked/planned |
| 209 | `direct_read_query_search` | `GET` | `/teams/{team_id}/metrics` | `getTeamMetrics` | Retrieve team metrics | operation-ledger blocked/planned |
| 210 | `etl_read` | `GET` | `/ticket_states` | `listTicketStates` | List all ticket states | operation-ledger blocked/planned |
| 211 | `etl_read` | `GET` | `/ticket_types` | `listTicketTypes` | List all ticket types | operation-ledger blocked/planned |
| 212 | `reverse_etl_write` | `POST` | `/ticket_types` | `createTicketType` | Create a ticket type | covered_by write `create_ticket_type` |
| 213 | `direct_read_query_search` | `GET` | `/ticket_types/{ticket_type_id}` | `getTicketType` | Retrieve a ticket type | operation-ledger blocked/planned |
| 214 | `reverse_etl_write` | `PUT` | `/ticket_types/{ticket_type_id}` | `updateTicketType` | Update a ticket type | covered_by write `update_ticket_type` |
| 215 | `reverse_etl_write` | `POST` | `/ticket_types/{ticket_type_id}/attributes` | `createTicketTypeAttribute` | Create a new attribute for a ticket type | covered_by write `create_ticket_type_attribute` |
| 216 | `reverse_etl_write` | `PUT` | `/ticket_types/{ticket_type_id}/attributes/{attribute_id}` | `updateTicketTypeAttribute` | Update an existing attribute for a ticket type | covered_by write `update_ticket_type_attribute` |
| 217 | `reverse_etl_write` | `POST` | `/tickets` | `createTicket` | Create a ticket | covered_by write `create_ticket` |
| 218 | `reverse_etl_write` | `POST` | `/tickets/enqueue` | `enqueueCreateTicket` | Enqueue create ticket | covered_by write `enqueue_create_ticket` |
| 219 | `etl_read` | `POST` | `/tickets/search` | `searchTickets` | Search tickets | operation-ledger blocked/planned |
| 220 | `reverse_etl_write` | `DELETE` | `/tickets/{ticket_id}` | `deleteTicket` | Delete a ticket | covered_by write `delete_ticket` |
| 221 | `direct_read_query_search` | `GET` | `/tickets/{ticket_id}` | `getTicket` | Retrieve a ticket | operation-ledger blocked/planned |
| 222 | `reverse_etl_write` | `PUT` | `/tickets/{ticket_id}` | `updateTicket` | Update a ticket | covered_by write `update_ticket` |
| 223 | `reverse_etl_write` | `POST` | `/tickets/{ticket_id}/change_type` | `changeTicketType` | Change ticket type | covered_by write `change_ticket_type` |
| 224 | `reverse_etl_write` | `POST` | `/tickets/{ticket_id}/linked_conversations` | `linkConversationToTicket` | Link a conversation to a ticket | covered_by write `link_conversation_to_ticket` |
| 225 | `reverse_etl_write` | `DELETE` | `/tickets/{ticket_id}/linked_conversations/{id}` | `unlinkConversationFromTicket` | Unlink a conversation from a ticket | covered_by write `unlink_conversation_from_ticket` |
| 226 | `reverse_etl_write` | `POST` | `/tickets/{ticket_id}/reply` | `replyTicket` | Reply to a ticket | covered_by write `reply_ticket` |
| 227 | `reverse_etl_write` | `POST` | `/tickets/{ticket_id}/tags` | `attachTagToTicket` | Add tag to a ticket | covered_by write `attach_tag_to_ticket` |
| 228 | `reverse_etl_write` | `DELETE` | `/tickets/{ticket_id}/tags/{tag_id}` | `detachTagFromTicket` | Remove tag from a ticket | covered_by write `detach_tag_from_ticket` |
| 229 | `direct_read_query_search` | `GET` | `/visitors` | `retrieveVisitorWithUserId` | Retrieve a visitor with User ID | operation-ledger blocked/planned |
| 230 | `reverse_etl_write` | `PUT` | `/visitors` | `updateVisitor` | Update a visitor | covered_by write `update_visitor` |
| 231 | `reverse_etl_write` | `POST` | `/visitors/convert` | `convertVisitor` | Convert a visitor | covered_by write `convert_visitor` |
