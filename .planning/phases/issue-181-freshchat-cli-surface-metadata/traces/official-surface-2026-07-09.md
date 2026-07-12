# Freshchat official API surface — sanitized planning extraction

Source: https://developers.freshchat.com/api/
Fetched: 2026-07-09
Purpose: issue #181 planning baseline for parent #180.

Raw HTML was not committed because the official docs page contains secret-shaped Authorization examples. This artifact contains only sanitized method/path rows.

## Count

- Current bundle baseline: 34 API surface rows.
- Sanitized official extraction: 34 operation rows.
- Extraction note: a docs example typo emitted `GET /metric`; it is not counted as an official operation. `GET /users/{user_id}/conversations` is retained from the official navigation/body.

## Sanitized operation rows

| # | Method | Path | Current classification |
| ---: | --- | --- | --- |
| 1 | GET | `/accounts/configuration` | stream `account_configuration` |
| 2 | GET | `/agents` | stream `agents` |
| 3 | POST | `/agents` | write `create_agent` |
| 4 | GET | `/agents/{agent_id}` | stream `agent_details` |
| 5 | PUT | `/agents/{agent_id}` | write `update_agent` |
| 6 | GET | `/agents/status` | stream `agent_statuses` |
| 7 | PATCH | `/agents/{agent_id}` | write `update_agent_status` |
| 8 | DELETE | `/agents/{agent_id}` | write `delete_agent` |
| 9 | GET | `/users` | stream `users` |
| 10 | POST | `/users` | write `create_user` |
| 11 | POST | `/users/fetch` | excluded/planned direct-read gap |
| 12 | GET | `/users/{user_id}` | stream `user_details` |
| 13 | PUT | `/users/{user_id}` | write `update_user` |
| 14 | GET | `/users/{user_id}/conversations` | stream `user_conversations` |
| 15 | DELETE | `/users/{user_id}` | write `delete_user` |
| 16 | POST | `/conversations` | write `create_conversation` |
| 17 | GET | `/conversations/{conversation_id}` | stream `conversation_detail` |
| 18 | PUT | `/conversations/{conversation_id}` | write `update_conversation` |
| 19 | GET | `/conversations/{conversation_id}/messages` | stream `conversation_messages` |
| 20 | POST | `/conversations/{conversation_id}/messages` | write `send_conversation_message` |
| 21 | GET | `/conversations/fields` | stream `conversation_fields` |
| 22 | GET | `/groups` | stream `groups` |
| 23 | GET | `/channels` | stream `channels` |
| 24 | GET | `/roles` | stream `roles` |
| 25 | POST | `/csat/{conversation_id}` | write `create_csat_rating` |
| 26 | POST | `/files/upload` | binary/multipart gap |
| 27 | POST | `/images/upload` | binary/multipart gap |
| 28 | POST | `/outbound-messages/whatsapp` | write `send_outbound_whatsapp_message` |
| 29 | GET | `/outbound-messages` | stream `outbound_messages` |
| 30 | POST | `/reports/raw` | write `extract_report` |
| 31 | GET | `/reports/raw/{id}` | stream `report_status` |
| 32 | GET | `/metrics/historical` | stream `historical_metrics` |
| 33 | GET | `/metrics/instant` | stream `instant_metrics` |
| 34 | GET | `/business-hours/within-bh` | stream `business_hours_status` |

## Follow-up classification notes

- #184 should convert legacy `excluded` rows into operation-ledger rows where appropriate and avoid blanket exclusions for sensitive/admin/destructive operations.
- #185/#186 should decide whether `/users/fetch`, file/image uploads, and report artifact policies become bounded direct-read/binary operations or remain explicitly blocked with non-product-scope reasons.
