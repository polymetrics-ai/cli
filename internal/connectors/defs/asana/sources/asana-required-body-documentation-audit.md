# Asana required-body documentation audit

Captured 2026-08-24 for the 22 required `data` bodies that remain after source
projection. This audit treats Asana's published endpoint pages as a first-class
provider source. Asana's rendered page directs clients to append `.md` for its
Markdown representation; the SHA-256 and byte count below pin that
provider-supported representation. Every capture returned HTTP 200.

The OpenAPI baseline is the existing
[`asana-operation-source-lock.json`](asana-operation-source-lock.json):
`https://raw.githubusercontent.com/Asana/openapi/56796a67a3c093eedf55fd9682357957a2ebfd85/defs/asana_oas.yaml`,
SHA-256 `cb3b90f4e0af56035eab0c648974f625b942a28a7144aa6c2326e38ca0bb3d56`,
3,066,750 bytes. For each row, the page's **OpenAPI definition** code fence
contains the listed request-body pointer and schema pointer. The data schema
or every `allOf` branch deliberately omits `additionalProperties: false`; under
OpenAPI 3.0 that remains open. No endpoint page supplies prose that says the
listed fields are exhaustive. Consequently, none is a provider-bounded body
that can safely be turned into a closed connector record schema.

| Operation | Endpoint | Pinned provider documentation | Bytes / SHA-256 | Provider documentation location and result |
|---|---|---|---|---|
| `createCustomField` | `POST /custom_fields` | [page](https://developers.asana.com/reference/createcustomfield) · [pinned Markdown](https://developers.asana.com/reference/createcustomfield.md) | 64,850 · `c71c45616f09e10df55491a193299893dd9b9cc822ca76c3d07661a17fd09c0d` | `paths["/custom_fields"].post.requestBody…data → CustomFieldCreateRequest`; open `allOf`; no closure. |
| `createGoal` | `POST /goals` | [page](https://developers.asana.com/reference/creategoal) · [pinned Markdown](https://developers.asana.com/reference/creategoal.md) | 90,096 · `7cfcf3b7371763239c4dbc0f4b331a14c280eb932937a79429565b1c8aa846e9` | `paths["/goals"].post.requestBody…data → GoalRequest`; open `allOf`; no closure. |
| `createOooEntry` | `POST /ooo_entries` | [page](https://developers.asana.com/reference/createoooentry) · [pinned Markdown](https://developers.asana.com/reference/createoooentry.md) | 22,154 · `4fcfb7edd3744cb940d5ddc00b63250da5c10589c3d5928c1ec0fc853e20e41f` | `paths["/ooo_entries"].post.requestBody…data → OooEntryCreateRequest`; both data `allOf` branches open. |
| `createPortfolio` | `POST /portfolios` | [page](https://developers.asana.com/reference/createportfolio) · [pinned Markdown](https://developers.asana.com/reference/createportfolio.md) | 84,283 · `f420c066ead9bea293d7685159f69a354d05c071d260547fe020aaf64b38559c` | `paths["/portfolios"].post.requestBody…data → PortfolioRequest`; open `allOf`; no closure. |
| `createProject` | `POST /projects` | [page](https://developers.asana.com/reference/createproject) · [pinned Markdown](https://developers.asana.com/reference/createproject.md) | 101,544 · `07b173d166ad5f6ef2cf94918af7bbacef3ecd229378c8d03865b88c1ed85114` | `…data → ProjectRequest`; open root and explicit dynamic `custom_fields` key map. |
| `createProjectBrief` | `POST /projects/{project_gid}/project_briefs` | [page](https://developers.asana.com/reference/createprojectbrief) · [pinned Markdown](https://developers.asana.com/reference/createprojectbrief.md) | 26,694 · `49cfabbdd2d83a37ed726699f1d4599510b820ac116932f4858515b26161e6ec` | `…data → ProjectBriefRequest`; open `allOf`; no closure. |
| `createProjectStatusForProject` | `POST /projects/{project_gid}/project_statuses` | [page](https://developers.asana.com/reference/createprojectstatusforproject) · [pinned Markdown](https://developers.asana.com/reference/createprojectstatusforproject.md) | 26,761 · `719146ea7de3a75dd06a6765914cd141feb7deaf2fb474167f9a650befb38d3b` | `…data → ProjectStatusBase`; both `allOf` branches open. |
| `createStatusForObject` | `POST /status_updates` | [page](https://developers.asana.com/reference/createstatusforobject) · [pinned Markdown](https://developers.asana.com/reference/createstatusforobject.md) | 36,321 · `519199b8fc9de36203628fe2a5735aa09cf1b7b68d6f2355465b454ce273c603` | `…data → StatusUpdateRequest`; open `allOf`; no closure. |
| `createTag` | `POST /tags` | [page](https://developers.asana.com/reference/createtag) · [pinned Markdown](https://developers.asana.com/reference/createtag.md) | 27,466 · `a09f1fa26cd76f7114eeda0e3def90144268715e46f893cf15ddc69be58f244d` | `…data → TagCreateRequest`; open `allOf`; no closure. |
| `createTask` | `POST /tasks` | [page](https://developers.asana.com/reference/createtask) · [pinned Markdown](https://developers.asana.com/reference/createtask.md) | 94,411 · `e46fc82def7fd7c867d61de2cff98f5b41300fa4cbff3b25c6d88d3fa37d589b` | `…data → TaskCreateRequest`; open root and explicit dynamic `custom_fields` key map. |
| `createSubtaskForTask` | `POST /tasks/{task_gid}/subtasks` | [page](https://developers.asana.com/reference/createsubtaskfortask) · [pinned Markdown](https://developers.asana.com/reference/createsubtaskfortask.md) | 94,545 · `9d43a90b2fb2ee5b5f94ef93847a5cb962261e1660caa526c4f7f6a7482f76ac` | `…data → TaskCreateRequest`; open root and explicit dynamic `custom_fields` key map. |
| `createTeam` | `POST /teams` | [page](https://developers.asana.com/reference/createteam) · [pinned Markdown](https://developers.asana.com/reference/createteam.md) | 73,991 · `824358c902eaf6d0575efa359805cb13b5c2fcc388285ecb89c3777ff601d26f` | `…data → TeamRequest`; both `allOf` branches open. |
| `createProjectForTeam` | `POST /teams/{team_gid}/projects` | [page](https://developers.asana.com/reference/createprojectforteam) · [pinned Markdown](https://developers.asana.com/reference/createprojectforteam.md) | 100,880 · `f01c7023f429673634d4a33ac6de0f12312a886cb5fda785fac0904f44b4488c` | `…data → ProjectRequest`; open root and explicit dynamic `custom_fields` key map. |
| `createProjectForWorkspace` | `POST /workspaces/{workspace_gid}/projects` | [page](https://developers.asana.com/reference/createprojectforworkspace) · [pinned Markdown](https://developers.asana.com/reference/createprojectforworkspace.md) | 101,302 · `bda37a70edbd37d4203cb584f5bb513626f8a3462edd9092313a10731b874200` | `…data → ProjectRequest`; open root and explicit dynamic `custom_fields` key map. |
| `createTagForWorkspace` | `POST /workspaces/{workspace_gid}/tags` | [page](https://developers.asana.com/reference/createtagforworkspace) · [pinned Markdown](https://developers.asana.com/reference/createtagforworkspace.md) | 28,215 · `54a8502bf3142450eb22db2c7e780d9fe85e301c910a6e5e85b334aec198d56d` | `…data → TagCreateTagForWorkspaceRequest`; open `allOf`; no closure. |
| `updateGoalRelationship` | `PUT /goal_relationships/{goal_relationship_gid}` | [page](https://developers.asana.com/reference/updategoalrelationship) · [pinned Markdown](https://developers.asana.com/reference/updategoalrelationship.md) | 29,066 · `6310667dac6705146fe2165243cfaa43555a0734e21c2c1407e4cd606f3edc79` | `…data → GoalRelationshipRequest`; both `allOf` branches open. |
| `updateGoal` | `PUT /goals/{goal_gid}` | [page](https://developers.asana.com/reference/updategoal) · [pinned Markdown](https://developers.asana.com/reference/updategoal.md) | 91,681 · `1ed17b7b28e309b645a3b3b09b420fa6321329522a700368bac66a1acfbb4139` | `…data → GoalUpdateRequest`; open root and explicit dynamic `custom_fields` key map. |
| `updatePortfolio` | `PUT /portfolios/{portfolio_gid}` | [page](https://developers.asana.com/reference/updateportfolio) · [pinned Markdown](https://developers.asana.com/reference/updateportfolio.md) | 85,156 · `33f64e0a091747a285498c317d97aaa6df1a14e77bfd59f8a1d2cc309f77e902` | `…data → PortfolioUpdateRequest`; open root and explicit dynamic `custom_fields` key map. |
| `updateProjectBrief` | `PUT /project_briefs/{project_brief_gid}` | [page](https://developers.asana.com/reference/updateprojectbrief) · [pinned Markdown](https://developers.asana.com/reference/updateprojectbrief.md) | 26,718 · `28739946a1c53cb3a2cc02555083aec2ff9e2de9d18e48b195799d03fa4c25a1` | `…data → ProjectBriefRequest`; open `allOf`; no closure. |
| `updateTeam` | `PUT /teams/{team_gid}` | [page](https://developers.asana.com/reference/updateteam) · [pinned Markdown](https://developers.asana.com/reference/updateteam.md) | 74,567 · `3e70be6241e7face63fab1cf97f4ee54d71d30203e2dc296cca439c3093065f0` | `…data → TeamRequest`; both `allOf` branches open. |
| `updateUser` | `PUT /users/{user_gid}` | [page](https://developers.asana.com/reference/updateuser) · [pinned Markdown](https://developers.asana.com/reference/updateuser.md) | 40,280 · `6d495107d03fdbf24312cfd81e348517193a72d3748fe2827d8fc17c9ea84b97` | `…data → UserUpdateRequest`; open root and explicit dynamic `custom_fields` key map. |
| `updateUserForWorkspace` | `PUT /workspaces/{workspace_gid}/users/{user_gid}` | [page](https://developers.asana.com/reference/updateuserforworkspace) · [pinned Markdown](https://developers.asana.com/reference/updateuserforworkspace.md) | 40,609 · `37ce4adec74a162c4dca04380f5ff45f43fca0287049bc1c6dfd95057d657f63` | `…data → UserUpdateRequest`; open root and explicit dynamic `custom_fields` key map. |

## Result

Documentation-resolvable required bodies: **0 of 22**.

Provider-unbounded required bodies with both source classes checked: **22 of
22**. The nine explicit dynamic maps are `createProject`,
`createProjectForTeam`, `createProjectForWorkspace`, `createSubtaskForTask`,
`createTask`, `updateGoal`, `updatePortfolio`, `updateUser`, and
`updateUserForWorkspace`; the remaining 13 data roots are open because neither
the pinned OpenAPI artifact nor the pinned endpoint page closes them. A
connector-local closed schema would reject provider-permitted inputs, so no
contract or availability claim is changed by this audit.
