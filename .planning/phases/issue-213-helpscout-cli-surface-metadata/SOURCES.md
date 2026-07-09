# Sources: Help Scout CLI Surface Metadata

- Official docs: https://developer.helpscout.com/mailbox-api/
- Crawled endpoint nav pages: 146
- Unique normalized method/path rows: 145
- Method split: DELETE 19, GET 79, PATCH 6, POST 21, PUT 20
- Duplicate docs page: `GET /v2/conversations/{conversation_id}/threads/{thread_id}/original-source` appears as JSON and RFC 822 variants.

| Method | Path | Docs title | Source |
|---|---|---|---|
| GET | `/v2/conversations` | List Conversations | https://developer.helpscout.com/mailbox-api/endpoints/conversations/list/ |
| DELETE | `/v2/conversations/{conversation_id}/attachments/{attachment_id}` | Delete Attachment | https://developer.helpscout.com/mailbox-api/endpoints/conversations/attachments/delete/ |
| GET | `/v2/conversations/{conversation_id}/attachments/{attachment_id}/file` | Download Attachment File | https://developer.helpscout.com/mailbox-api/endpoints/conversations/attachments/get-attachment-file/ |
| GET | `/v2/conversations/{conversation_id}/attachments/{attachment_id}/data` | Get Attachment Data | https://developer.helpscout.com/mailbox-api/endpoints/conversations/attachments/get-data/ |
| POST | `/v2/conversations/{conversation_id}/threads/{thread_id}/attachments` | Upload Attachment | https://developer.helpscout.com/mailbox-api/endpoints/conversations/attachments/create/ |
| PUT | `/v2/conversations/{conversation_id}/fields` | Update Custom Fields | https://developer.helpscout.com/mailbox-api/endpoints/conversations/custom_fields/update/ |
| DELETE | `/v2/conversations/{conversation_id}/snooze` | Delete Snooze | https://developer.helpscout.com/mailbox-api/endpoints/conversations/snooze/delete/ |
| PUT | `/v2/conversations/{conversation_id}/snooze` | Update Snooze | https://developer.helpscout.com/mailbox-api/endpoints/conversations/snooze/update/ |
| PUT | `/v2/conversations/{conversation_id}/tags` | Update Tags | https://developer.helpscout.com/mailbox-api/endpoints/conversations/tags/update/ |
| DELETE | `/v2/conversations/{conversation_id}/threads/{thread_id}/schedule` | Delete Thread Schedule | https://developer.helpscout.com/mailbox-api/endpoints/conversations/threads/schedule/delete/ |
| PATCH | `/v2/conversations/{conversation_id}/threads/{thread_id}/schedule` | Publish Thread Schedule | https://developer.helpscout.com/mailbox-api/endpoints/conversations/threads/schedule/publish/ |
| PUT | `/v2/conversations/{conversation_id}/threads/{thread_id}/schedule` | Update Thread Schedule | https://developer.helpscout.com/mailbox-api/endpoints/conversations/threads/schedule/update/ |
| POST | `/v2/conversations/{conversation_id}/chats` | Create Chat Thread | https://developer.helpscout.com/mailbox-api/endpoints/conversations/threads/chat/ |
| POST | `/v2/conversations/{conversation_id}/customer` | Create Customer Thread | https://developer.helpscout.com/mailbox-api/endpoints/conversations/threads/customer/ |
| POST | `/v2/conversations/{conversation_id}/notes` | Create Note | https://developer.helpscout.com/mailbox-api/endpoints/conversations/threads/note/ |
| POST | `/v2/conversations/{conversation_id}/phones` | Create Phone Thread | https://developer.helpscout.com/mailbox-api/endpoints/conversations/threads/phone/ |
| POST | `/v2/conversations/{conversation_id}/reply` | Create Reply Thread | https://developer.helpscout.com/mailbox-api/endpoints/conversations/threads/reply/ |
| GET | `/v2/conversations/{conversation_id}/threads/{thread_id}/original-source` | Get Thread Original Source / Get Thread Original Source (RFC 822) | https://developer.helpscout.com/mailbox-api/endpoints/conversations/threads/thread-source-json/ |
| GET | `/v2/conversations/{conversation_id}/threads` | List Threads | https://developer.helpscout.com/mailbox-api/endpoints/conversations/threads/list/ |
| GET | `/v3/conversations/{conversation_id}/threads` | List Threads (v3) | https://developer.helpscout.com/mailbox-api/endpoints/conversations/threads/list-v3/ |
| PATCH | `/v2/conversations/{conversation_id}/threads/{thread_id}` | Update Thread | https://developer.helpscout.com/mailbox-api/endpoints/conversations/threads/update/ |
| POST | `/v2/conversations` | Create Conversation | https://developer.helpscout.com/mailbox-api/endpoints/conversations/create/ |
| DELETE | `/v2/conversations/{conversation_id}` | Delete Conversation | https://developer.helpscout.com/mailbox-api/endpoints/conversations/delete/ |
| GET | `/v2/conversations/{conversation_id}` | Get Conversation | https://developer.helpscout.com/mailbox-api/endpoints/conversations/get/ |
| GET | `/v3/conversations/{conversation_id}` | Get Conversation (v3) | https://developer.helpscout.com/mailbox-api/endpoints/conversations/get-v3/ |
| PATCH | `/v2/conversations/{conversation_id}` | Update Conversation | https://developer.helpscout.com/mailbox-api/endpoints/conversations/update/ |
| GET | `/v2/customers` | List Customers | https://developer.helpscout.com/mailbox-api/endpoints/customers/list/ |
| POST | `/v2/customers/{customer_id}/address` | Create Address | https://developer.helpscout.com/mailbox-api/endpoints/customers/address/create/ |
| DELETE | `/v2/customers/{customer_id}/address` | Delete Address | https://developer.helpscout.com/mailbox-api/endpoints/customers/address/delete/ |
| GET | `/v2/customers/{customer_id}/address` | Get Address | https://developer.helpscout.com/mailbox-api/endpoints/customers/address/get/ |
| PUT | `/v2/customers/{customer_id}/address` | Update Address | https://developer.helpscout.com/mailbox-api/endpoints/customers/address/update/ |
| POST | `/v2/customers/{customer_id}/chats` | Create Chat Handle | https://developer.helpscout.com/mailbox-api/endpoints/customers/chat_handles/create/ |
| DELETE | `/v2/customers/{customer_id}/chats/{chat_id}` | Delete Chat Handle | https://developer.helpscout.com/mailbox-api/endpoints/customers/chat_handles/delete/ |
| GET | `/v2/customers/{customer_id}/chats` | List Chats Handles | https://developer.helpscout.com/mailbox-api/endpoints/customers/chat_handles/list/ |
| PUT | `/v2/customers/{customer_id}/chats/{chat_id}` | Update Chat Handles | https://developer.helpscout.com/mailbox-api/endpoints/customers/chat_handles/update/ |
| POST | `/v2/customers/{customer_id}/emails` | Create Email | https://developer.helpscout.com/mailbox-api/endpoints/customers/emails/create/ |
| DELETE | `/v2/customers/{customer_id}/emails/{email_id}` | Delete Email | https://developer.helpscout.com/mailbox-api/endpoints/customers/emails/delete/ |
| GET | `/v2/customers/{customer_id}/emails` | List Emails | https://developer.helpscout.com/mailbox-api/endpoints/customers/emails/list/ |
| PUT | `/v2/customers/{customer_id}/emails/{email_id}` | Update Email | https://developer.helpscout.com/mailbox-api/endpoints/customers/emails/update/ |
| POST | `/v2/customers/{customer_id}/phones` | Create Phone | https://developer.helpscout.com/mailbox-api/endpoints/customers/phones/create/ |
| DELETE | `/v2/customers/{customer_id}/phones/{phone_id}` | Delete Phone | https://developer.helpscout.com/mailbox-api/endpoints/customers/phones/delete/ |
| GET | `/v2/customers/{customer_id}/phones` | List Phones | https://developer.helpscout.com/mailbox-api/endpoints/customers/phones/list/ |
| PUT | `/v2/customers/{customer_id}/phones/{phone_id}` | Update Phone | https://developer.helpscout.com/mailbox-api/endpoints/customers/phones/update/ |
| POST | `/v2/customers/{customer_id}/social-profiles` | Create Social Profile | https://developer.helpscout.com/mailbox-api/endpoints/customers/social_profiles/create/ |
| DELETE | `/v2/customers/{customer_id}/social-profiles/{social_profile_id}` | Delete Social Profile | https://developer.helpscout.com/mailbox-api/endpoints/customers/social_profiles/delete/ |
| GET | `/v2/customers/{customer_id}/social-profiles` | List Social Profiles | https://developer.helpscout.com/mailbox-api/endpoints/customers/social_profiles/list/ |
| PUT | `/v2/customers/{customer_id}/social-profiles/{social_profile_id}` | Update Social Profile | https://developer.helpscout.com/mailbox-api/endpoints/customers/social_profiles/update/ |
| POST | `/v2/customers/{customer_id}/websites` | Create Website | https://developer.helpscout.com/mailbox-api/endpoints/customers/websites/create/ |
| DELETE | `/v2/customers/{customer_id}/websites/{website_id}` | Delete Website | https://developer.helpscout.com/mailbox-api/endpoints/customers/websites/delete/ |
| GET | `/v2/customers/{customer_id}/websites` | List Websites | https://developer.helpscout.com/mailbox-api/endpoints/customers/websites/list/ |
| PUT | `/v2/customers/{customer_id}/websites/{website_id}` | Update Website | https://developer.helpscout.com/mailbox-api/endpoints/customers/websites/update/ |
| POST | `/v2/customers` | Create Customer | https://developer.helpscout.com/mailbox-api/endpoints/customers/create/ |
| DELETE | `/v2/customers/{customer_id}` | Delete Customer | https://developer.helpscout.com/mailbox-api/endpoints/customers/delete/ |
| DELETE | `/v2/customers/{customer_id}?async=true` | Delete Customer Asynchronously | https://developer.helpscout.com/mailbox-api/endpoints/customers/delete-async/ |
| GET | `/v2/customers/{customer_id}` | Get Customer | https://developer.helpscout.com/mailbox-api/endpoints/customers/get/ |
| GET | `/v3/customers` | List Customers (v3) | https://developer.helpscout.com/mailbox-api/endpoints/customers/list-v3/ |
| PUT | `/v2/customers/{customer_id}` | Overwrite Customer | https://developer.helpscout.com/mailbox-api/endpoints/customers/overwrite/ |
| PATCH | `/v2/customers/{customer_id}` | Update Customer | https://developer.helpscout.com/mailbox-api/endpoints/customers/update/ |
| GET | `/v2/mailboxes/{mailbox_id}/routing` | Get Routing configuration | https://developer.helpscout.com/mailbox-api/endpoints/inboxes/routing/routing-configuration-get/ |
| PUT | `/v2/mailboxes/{mailbox_id}/routing` | Update Routing configuration | https://developer.helpscout.com/mailbox-api/endpoints/inboxes/routing/routing-configuration-update/ |
| POST | `/v2/mailboxes/{mailbox_id}/saved-replies` | Create Saved Reply | https://developer.helpscout.com/mailbox-api/endpoints/inboxes/saved_Replies/saved-replies-create/ |
| DELETE | `/v2/mailboxes/{mailbox_id}/saved-replies/{saved_reply_id}` | Delete Saved Reply | https://developer.helpscout.com/mailbox-api/endpoints/inboxes/saved_Replies/saved-replies-delete/ |
| GET | `/v2/mailboxes/{mailbox_id}/saved-replies/{saved_reply_id}` | Get Saved Reply | https://developer.helpscout.com/mailbox-api/endpoints/inboxes/saved_Replies/saved-replies-get/ |
| GET | `/v2/mailboxes/{mailbox_id}/saved-replies` | List Saved Replies | https://developer.helpscout.com/mailbox-api/endpoints/inboxes/saved_Replies/saved-replies-list/ |
| PUT | `/v2/mailboxes/{mailbox_id}/saved-replies/{saved_reply_id}` | Update Saved Reply | https://developer.helpscout.com/mailbox-api/endpoints/inboxes/saved_Replies/saved-replies-update/ |
| GET | `/v2/mailboxes/{mailbox_id}` | Get Inbox | https://developer.helpscout.com/mailbox-api/endpoints/inboxes/get/ |
| GET | `/v2/mailboxes/{mailbox_id}/fields` | List Inbox Custom Fields | https://developer.helpscout.com/mailbox-api/endpoints/inboxes/inbox-fields/ |
| GET | `/v2/mailboxes/{mailbox_id}/folders` | List Inbox Folders | https://developer.helpscout.com/mailbox-api/endpoints/inboxes/inbox-folders/ |
| GET | `/v2/mailboxes` | List Inboxes | https://developer.helpscout.com/mailbox-api/endpoints/inboxes/list/ |
| GET | `/v2/organizations/{organization_id}` | Get Organization by ID | https://developer.helpscout.com/mailbox-api/endpoints/organizations/get/ |
| POST | `/v2/organizations/properties` | Create Organization Property Definition | https://developer.helpscout.com/mailbox-api/endpoints/organizations/properties/create/ |
| DELETE | `/v2/organizations/properties/{property_slug}` | Delete Organization Property Definition | https://developer.helpscout.com/mailbox-api/endpoints/organizations/properties/delete/ |
| GET | `/v2/organizations/properties/{property_slug}` | Get Organization Property Definition | https://developer.helpscout.com/mailbox-api/endpoints/organizations/properties/get/ |
| GET | `/v2/organizations/properties` | List Organization Property Definitions | https://developer.helpscout.com/mailbox-api/endpoints/organizations/properties/list/ |
| DELETE | `/v2/organizations/{organization_id}/properties/{property_slug}` | Remove Organization Property Value | https://developer.helpscout.com/mailbox-api/endpoints/organizations/properties/remove-value/ |
| PUT | `/v2/organizations/{organization_id}/properties/{property_slug}` | Set Organization Property Value | https://developer.helpscout.com/mailbox-api/endpoints/organizations/properties/set-value/ |
| PUT | `/v2/organizations/properties/{property_slug}` | Update Organization Property Definition | https://developer.helpscout.com/mailbox-api/endpoints/organizations/properties/update/ |
| POST | `/v2/organizations` | Create Organization | https://developer.helpscout.com/mailbox-api/endpoints/organizations/create/ |
| DELETE | `/v2/organizations/{organization_id}` | Delete Organization by ID | https://developer.helpscout.com/mailbox-api/endpoints/organizations/delete/ |
| GET | `/v2/organizations/{organization_id}/conversations` | Get Organization Conversations | https://developer.helpscout.com/mailbox-api/endpoints/organizations/get-conversations/ |
| GET | `/v2/organizations/{organization_id}/customers` | Get Organization Customers | https://developer.helpscout.com/mailbox-api/endpoints/organizations/get-customers/ |
| GET | `/v2/organizations` | List Organizations | https://developer.helpscout.com/mailbox-api/endpoints/organizations/list/ |
| PUT | `/v2/organizations/{organization_id}` | Update Organization | https://developer.helpscout.com/mailbox-api/endpoints/organizations/update/ |
| POST | `/v2/customer-properties` | Create Customer Property Definition | https://developer.helpscout.com/mailbox-api/endpoints/properties/create/ |
| DELETE | `/v2/customer-properties/{property_slug}` | Delete Customer Property Definition | https://developer.helpscout.com/mailbox-api/endpoints/properties/delete/ |
| GET | `/v2/customer-properties` | List Customer Property Definitions | https://developer.helpscout.com/mailbox-api/endpoints/properties/list/ |
| PATCH | `/v2/customers/{customer_id}/properties` | Update Customer Properties | https://developer.helpscout.com/mailbox-api/endpoints/properties/update/ |
| GET | `/v2/ratings/{rating_id}` | Get Satisfaction Rating | https://developer.helpscout.com/mailbox-api/endpoints/ratings/get/ |
| GET | `/v2/reports/company` | Company Overall Report | https://developer.helpscout.com/mailbox-api/endpoints/reports/company/reports-company-overall/ |
| GET | `/v2/reports/company/customers-helped` | Company Customers Helped | https://developer.helpscout.com/mailbox-api/endpoints/reports/company/reports-company-customers-helped/ |
| GET | `/v2/reports/company/drilldown` | Company Drilldown | https://developer.helpscout.com/mailbox-api/endpoints/reports/company/reports-company-drilldown/ |
| GET | `/v2/reports/conversations` | Conversations - Overall Report | https://developer.helpscout.com/mailbox-api/endpoints/reports/conversations/reports-conversations-overall/ |
| GET | `/v2/reports/conversations/volume-by-channel` | All Channels - Volumes by Channel | https://developer.helpscout.com/mailbox-api/endpoints/reports/conversations/reports-conversations-volume-by-channel/ |
| GET | `/v2/reports/conversations/busy-times` | Conversations - Busiest Time of Day | https://developer.helpscout.com/mailbox-api/endpoints/reports/conversations/reports-conversations-busy-times/ |
| GET | `/v2/reports/conversations/drilldown` | Conversations - Drilldown | https://developer.helpscout.com/mailbox-api/endpoints/reports/conversations/reports-conversations-drilldown/ |
| GET | `/v2/reports/conversations/fields-drilldown` | Conversations - Drilldown by Field | https://developer.helpscout.com/mailbox-api/endpoints/reports/conversations/reports-conversations-field-drilldown/ |
| GET | `/v2/reports/conversations/new` | Conversations - New Conversations | https://developer.helpscout.com/mailbox-api/endpoints/reports/conversations/reports-conversations-new/ |
| GET | `/v2/reports/conversations/new-drilldown` | Conversations - New Conversations Drilldown | https://developer.helpscout.com/mailbox-api/endpoints/reports/conversations/reports-conversations-new-drilldown/ |
| GET | `/v2/reports/conversations/received-messages` | Conversations - Received Messages Statistics | https://developer.helpscout.com/mailbox-api/endpoints/reports/conversations/reports-conversations-received-messages/ |
| GET | `/v2/reports/docs` | Docs Overall Report | https://developer.helpscout.com/mailbox-api/endpoints/reports/docs/reports-docs-overall/ |
| GET | `/v2/reports/happiness` | Happiness Overall Report | https://developer.helpscout.com/mailbox-api/endpoints/reports/happiness/reports-happiness-overall/ |
| GET | `/v2/reports/happiness/ratings` | Happiness Ratings Report | https://developer.helpscout.com/mailbox-api/endpoints/reports/happiness/reports-happiness-ratings/ |
| GET | `/v2/reports/productivity` | Productivity Overall Report | https://developer.helpscout.com/mailbox-api/endpoints/reports/productivity/reports-productivity-overall/ |
| GET | `/v2/reports/productivity/first-response-time` | Productivity - First Response Time | https://developer.helpscout.com/mailbox-api/endpoints/reports/productivity/reports-productivity-first-response-time/ |
| GET | `/v2/reports/productivity/replies-sent` | Productivity - Replies Sent | https://developer.helpscout.com/mailbox-api/endpoints/reports/productivity/reports-productivity-replies-sent/ |
| GET | `/v2/reports/productivity/resolution-time` | Productivity - Resolution Time | https://developer.helpscout.com/mailbox-api/endpoints/reports/productivity/reports-productivity-resolution-time/ |
| GET | `/v2/reports/productivity/resolved` | Productivity - Resolved | https://developer.helpscout.com/mailbox-api/endpoints/reports/productivity/reports-productivity-resolved/ |
| GET | `/v2/reports/productivity/response-time` | Productivity - Response Time | https://developer.helpscout.com/mailbox-api/endpoints/reports/productivity/reports-productivity-respose-time/ |
| GET | `/v2/reports/user` | User/Team Overall Report | https://developer.helpscout.com/mailbox-api/endpoints/reports/user/reports-user/ |
| GET | `/v2/reports/user/conversation-history` | User Conversation History | https://developer.helpscout.com/mailbox-api/endpoints/reports/user/reports-user-conversation-history/ |
| GET | `/v2/reports/user/customers-helped` | User Customers Helped | https://developer.helpscout.com/mailbox-api/endpoints/reports/user/reports-user-customer-helped/ |
| GET | `/v2/reports/user/drilldown` | User Drill-down | https://developer.helpscout.com/mailbox-api/endpoints/reports/user/reports-user-drilldown/ |
| GET | `/v2/reports/user/happiness` | User Happiness | https://developer.helpscout.com/mailbox-api/endpoints/reports/user/reports-user-happiness/ |
| GET | `/v2/reports/user/ratings` | User Happiness drilldown | https://developer.helpscout.com/mailbox-api/endpoints/reports/user/reports-user-happiness-drilldown/ |
| GET | `/v2/reports/user/replies` | User Replies | https://developer.helpscout.com/mailbox-api/endpoints/reports/user/reports-user-replies/ |
| GET | `/v2/reports/user/resolutions` | User Resolutions | https://developer.helpscout.com/mailbox-api/endpoints/reports/user/reports-user-resolutions/ |
| GET | `/v2/reports/user/chat` | User/Team Chat Report | https://developer.helpscout.com/mailbox-api/endpoints/reports/user/reports-user-chat/ |
| GET | `/v2/reports/chat` | Chat Report | https://developer.helpscout.com/mailbox-api/endpoints/reports/chat/ |
| GET | `/v2/reports/email` | Email Report | https://developer.helpscout.com/mailbox-api/endpoints/reports/email/ |
| GET | `/v2/reports/phone` | Phone Report | https://developer.helpscout.com/mailbox-api/endpoints/reports/phone/ |
| GET | `/v3/system-users/{system_user_id}` | Get System User | https://developer.helpscout.com/mailbox-api/endpoints/system-users/get-system-user/ |
| GET | `/v3/system-users` | List System Users | https://developer.helpscout.com/mailbox-api/endpoints/system-users/list-system-users/ |
| GET | `/v2/tags/{tag_id}` | Get Tag by ID | https://developer.helpscout.com/mailbox-api/endpoints/tags/get/ |
| GET | `/v2/tags` | List Tags | https://developer.helpscout.com/mailbox-api/endpoints/tags/list/ |
| GET | `/v2/teams/{team_id}/members` | List Team Members | https://developer.helpscout.com/mailbox-api/endpoints/teams/list-team-members/ |
| GET | `/v2/teams` | List Teams | https://developer.helpscout.com/mailbox-api/endpoints/teams/list-teams/ |
| PUT | `/v2/teams/{team_id}/members` | Update Team Members | https://developer.helpscout.com/mailbox-api/endpoints/teams/update-team-members/ |
| GET | `/v2/users` | List Users | https://developer.helpscout.com/mailbox-api/endpoints/users/list/ |
| GET | `/v2/users/{user_id}/conversation-reassignment` | Get Conversation Reassignment configuration | https://developer.helpscout.com/mailbox-api/endpoints/users/conversation%20reassignment/get/ |
| PUT | `/v2/users/{user_id}/conversation-reassignment` | Update Conversation Reassignment configuration | https://developer.helpscout.com/mailbox-api/endpoints/users/conversation%20reassignment/update/ |
| GET | `/v2/users/{user_id}/status` | Get user status | https://developer.helpscout.com/mailbox-api/endpoints/users/status/get/ |
| GET | `/v2/users/status` | List users statuses | https://developer.helpscout.com/mailbox-api/endpoints/users/status/list/ |
| PUT | `/v2/users/{user_id}/status` | Set user status | https://developer.helpscout.com/mailbox-api/endpoints/users/status/set/ |
| POST | `/v2/users` | Create User | https://developer.helpscout.com/mailbox-api/endpoints/users/create/ |
| DELETE | `/v2/users/{user_id}` | Delete User | https://developer.helpscout.com/mailbox-api/endpoints/users/delete/ |
| GET | `/v2/users/me` | Get Resource Owner | https://developer.helpscout.com/mailbox-api/endpoints/users/me/ |
| GET | `/v2/users/{user_id}` | Get User | https://developer.helpscout.com/mailbox-api/endpoints/users/get/ |
| POST | `/v2/webhooks` | Create Webhook | https://developer.helpscout.com/mailbox-api/endpoints/webhooks/create/ |
| DELETE | `/v2/webhooks/{webhook_id}` | Delete Webhook | https://developer.helpscout.com/mailbox-api/endpoints/webhooks/delete/ |
| GET | `/v2/webhooks/{webhook_id}` | Get Webhook | https://developer.helpscout.com/mailbox-api/endpoints/webhooks/get/ |
| GET | `/v2/webhooks` | List Webhooks | https://developer.helpscout.com/mailbox-api/endpoints/webhooks/list/ |
| PUT | `/v2/webhooks/{webhook_id}` | Update Webhook | https://developer.helpscout.com/mailbox-api/endpoints/webhooks/update/ |
| GET | `/v2/workflows` | List Workflows | https://developer.helpscout.com/mailbox-api/endpoints/workflows/list/ |
| POST | `/v2/workflows/{workflow_id}/run` | Run Manual Workflows | https://developer.helpscout.com/mailbox-api/endpoints/workflows/run/ |
| PATCH | `/v2/workflows/{workflow_id}` | Update workflow status | https://developer.helpscout.com/mailbox-api/endpoints/workflows/update/ |
