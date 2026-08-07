# Monday planned-operation reclassification evidence

Captured from official developer.monday.com reference pages (public docs only; no
live provider calls, no credentials). Recorded so the top-50 sweep does not repay
for these doc reads.

## Shared-foundation blocker (named)

`internal/connectors/engine/graphql.go` → `resolveGraphQLVariableTemplate` supports
only `string | integer | number | boolean`. There is no array/object GraphQL
variable type, so a GraphQL argument typed `[ID!]!` or `SomeInput!` cannot be
populated from a record field. Verified repo-wide: 0 connectors pass a literal
array or object as a GraphQL variable; the only templated variable types in use
across all 551 connectors are `integer` (5) and `boolean` (3).

Consequence: an operation is genuinely blocked **only when a REQUIRED argument is
a list or input object**. Optional list/object arguments can simply be omitted
from a typed bounded contract, so they do NOT block authoring.

## Result: the existing `planned` classification over-blocks

The prior wave marked 96 commands planned. Applying the rule above, the following
are **NOT blocked** and are authorable today with scalar-only contracts. This list
is verified against the cited pages, not inferred.

| Command | Required args (all scalar) | Source page |
|---|---|---|
| `reverse create-board` | `board_name: String!`, `board_kind: BoardKind!` | /reference/boards |
| `reverse create-update` | `body: String!` | /reference/updates |
| `reverse delete-form-tag` | `formToken: String!`, `tagId: String!` | /reference/form |
| `reverse update-folder` | `folder_id: ID!` | /reference/folders |
| `reverse create-view` | `board_id: ID!`, `type: ViewKind!` | /reference/board-views |
| `reverse update-view` | `board_id: ID!`, `type: ViewKind!`, `view_id: ID!` | /reference/board-views |
| `reverse update-view-table` | `board_id: ID!`, `view_id: ID!` | /reference/board-views |
| `reverse attach-dropdown-managed-column` | `board_id: ID!`, `managed_column_id: ID!` | /reference/managed-columns |
| `reverse update-dropdown-managed-column` | `id: String!`, `revision: Int!` | /reference/managed-columns |
| `reverse create-object` | `name: String!`, `object_type_unique_key: String!`, `privacy_kind: PrivacyKind!` | /reference/objects |
| `reverse update-status-column` | `board_id: ID!`, `id: String!`, `revision: String!` | /reference/status |
| `query departments` | none required (`ids: [ID!]` is optional) | /reference/departments |
| `query folders` | none required; `limit`/`page` are scalar bounding args | /reference/folders |

## Confirmed genuinely blocked (required list/object argument)

- **/reference/users — all 10 planned**: `activate_users`, `deactivate_users`,
  `clear_users_department` (`user_ids: [ID!]!`); `add_users_to_board`,
  `add_users_to_workspace`, `delete_subscribers_from_board`,
  `delete_users_from_workspace`, `update_users_role` (`user_ids: [ID!]!`);
  `invite_users` (`emails: [String!]!`); `update_email_domain`
  (`input: UpdateEmailDomainAttributesInput!`); `update_multiple_users`
  (`user_updates: [UserUpdateInput!]!`).
- **/reference/teams — all 9 planned**: every one takes `team_ids: [ID!]!` or
  `user_ids: [ID!]!`; `create_team` takes `input: CreateTeamAttributesInput!`.
- **/reference/form — 6 of 7**: `create_form_question`, `create_form_tag`,
  `set_form_password`, `update_form`, `update_form_question`,
  `update_form_settings` each take a required input object.
- **/reference/departments**: `create_department` (`data: ...Input!`),
  `update_department` (`data: ...Input`), `assign_department_members` and
  `unassign_department_owners` (`user_ids: [ID!]!`).
- **/reference/favorites — all 3**: `create_favorite`, `delete_favorite`,
  `update_favorite_position` each take a single required input object.
- **/reference/dashboards-and-widgets — all 3**: `create_dashboard`
  (`board_ids: [ID!]!`), `create_widget` (`parent: WidgetParentInput!`),
  `update_overview_hierarchy` (`attributes: ...!`).
- **/reference/managed-columns — 2**: `create_dropdown_managed_column` and
  `create_status_managed_column` both take `settings: ...Input!`.
- **/reference/objects — 2**: `add_subscribers_to_object` (`user_ids`),
  `update_object` (`input: UpdateObjectInput!`).
- **/reference/status — 1**: `create_status_column` requires the
  `defaults: CreateStatusColumnSettingsInput` labels array to be meaningful.
- **/reference/assets — 2**: `add_file_to_column`, `add_file_to_update` are
  binary/multipart uploads; blocked separately from the variable-typing issue.

## Not yet audited

Pages still unchecked when this worker was folded into the sweep:
object-schemas (5), articles (2), portfolio (2), app (2), dropdown (2),
marketplace-app-discounts (2), validations (2), items (2),
`boards → update-board-hierarchy` (1), and the single-operation pages:
aggregate, audit-logs, export-markdown-from-doc, items-page-by-column-values,
mute-board-settings, notifications-settings, replies, batch-extend-trial-period,
configure-categorize-ai-column, create-doc, create-doc-blocks,
create-object-relations, create-timeline-item, enroll-items-to-sequence,
ingest-items, run-prompt, update-app-feature, update-app-lifecycle-subscription,
update-directory-resources-attributes, update-workspace.
