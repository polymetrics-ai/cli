# pm connectors inspect trello

```text
NAME
  pm connectors inspect trello - Trello connector manual

SYNOPSIS
  pm connectors inspect trello
  pm connectors inspect trello --json
  pm credentials add <name> --connector trello [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes the supportable Trello REST API surface for boards, cards, lists, members, organizations, labels, checklists, actions, webhooks, search, attachments, plugins, notifications, emoji, and saved searches with fixture-backed declarative operations.

ICON
  asset: icons/trello.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developer.atlassian.com/cloud/trello/rest/

CAPABILITIES
  check=true catalog=true read=true write=true query=true
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  board_ids
  id
  key (secret)
  token (secret)

ETL STREAMS
  checklists:
    fields: fixture(), id(), name()
  lists:
    fields: fixture(), id(), name()
  boards:
    fields: fixture(), id(), name()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

REVERSE ETL ACTIONS
  put_actions_id:
    endpoint: PUT /actions/{{ record.id }}?token={{ secrets.token }}
    required fields: id, text
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  delete_actions_id:
    endpoint: DELETE /actions/{{ record.id }}?token={{ secrets.token }}
    required fields: id
    risk: Deletes or removes Trello data; requires destructive approval.
  put_actions_id_text:
    endpoint: PUT /actions/{{ record.id }}/text?token={{ secrets.token }}
    required fields: id, value
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  post_actions_idaction_reactions:
    endpoint: POST /actions/{{ record.idAction }}/reactions?token={{ secrets.token }}
    required fields: idAction
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  delete_actions_idaction_reactions_id:
    endpoint: DELETE /actions/{{ record.idAction }}/reactions/{{ record.id }}?token={{ secrets.token }}
    required fields: idAction, id
    risk: Deletes or removes Trello data; requires destructive approval.
  put_boards_id:
    endpoint: PUT /boards/{{ record.id }}?token={{ secrets.token }}
    required fields: id
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  delete_boards_id:
    endpoint: DELETE /boards/{{ record.id }}?token={{ secrets.token }}
    required fields: id
    risk: Deletes or removes Trello data; requires destructive approval.
  post_boards_id_labels:
    endpoint: POST /boards/{{ record.id }}/labels?token={{ secrets.token }}
    required fields: id, name, color
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  post_boards_id_lists:
    endpoint: POST /boards/{{ record.id }}/lists?token={{ secrets.token }}
    required fields: id, name
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  put_boards_id_members:
    endpoint: PUT /boards/{{ record.id }}/members?token={{ secrets.token }}
    required fields: id, email
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  put_boards_id_members_idmember:
    endpoint: PUT /boards/{{ record.id }}/members/{{ record.idMember }}?token={{ secrets.token }}
    required fields: id, idMember, type
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  boardsidmembersidmember:
    endpoint: DELETE /boards/{{ record.id }}/members/{{ record.idMember }}?token={{ secrets.token }}
    required fields: id, idMember
    risk: Deletes or removes Trello data; requires destructive approval.
  put_boards_id_memberships_idmembership:
    endpoint: PUT /boards/{{ record.id }}/memberships/{{ record.idMembership }}?token={{ secrets.token }}
    required fields: id, idMembership, type
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  put_boards_id_myprefs_emailposition:
    endpoint: PUT /boards/{{ record.id }}/myPrefs/emailPosition?token={{ secrets.token }}
    required fields: id, value
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  put_boards_id_myprefs_idemaillist:
    endpoint: PUT /boards/{{ record.id }}/myPrefs/idEmailList?token={{ secrets.token }}
    required fields: id, value
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  put_boards_id_my_prefs_showsidebar:
    endpoint: PUT /boards/{{ record.id }}/myPrefs/showSidebar?token={{ secrets.token }}
    required fields: id, value
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  put_boards_id_my_prefs_showsidebaractivity:
    endpoint: PUT /boards/{{ record.id }}/myPrefs/showSidebarActivity?token={{ secrets.token }}
    required fields: id, value
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  put_boards_id_my_prefs_showsidebarboardactions:
    endpoint: PUT /boards/{{ record.id }}/myPrefs/showSidebarBoardActions?token={{ secrets.token }}
    required fields: id, value
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  put_boards_id_my_prefs_showsidebarmembers:
    endpoint: PUT /boards/{{ record.id }}/myPrefs/showSidebarMembers?token={{ secrets.token }}
    required fields: id, value
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  post_boards:
    endpoint: POST /boards/?token={{ secrets.token }}
    required fields: name
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  post_boards_id_calendarkey_generate:
    endpoint: POST /boards/{{ record.id }}/calendarKey/generate?token={{ secrets.token }}
    required fields: id
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  post_boards_id_emailkey_generate:
    endpoint: POST /boards/{{ record.id }}/emailKey/generate?token={{ secrets.token }}
    required fields: id
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  post_boards_id_idtags:
    endpoint: POST /boards/{{ record.id }}/idTags?token={{ secrets.token }}
    required fields: id, value
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  post_boards_id_markedasviewed:
    endpoint: POST /boards/{{ record.id }}/markedAsViewed?token={{ secrets.token }}
    required fields: id
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  post_boards_id_boardplugins:
    endpoint: POST /boards/{{ record.id }}/boardPlugins?token={{ secrets.token }}
    required fields: id
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  delete_boards_id_boardplugins:
    endpoint: DELETE /boards/{{ record.id }}/boardPlugins/{{ record.idPlugin }}?token={{ secrets.token }}
    required fields: id, idPlugin
    risk: Deletes or removes Trello data; requires destructive approval.
  post_boards_id_exports:
    endpoint: POST /boards/{{ record.id }}/exports?token={{ secrets.token }}
    required fields: id
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  delete_boards_id_exports_idexport:
    endpoint: DELETE /boards/{{ record.id }}/exports/{{ record.idExport }}?token={{ secrets.token }}
    required fields: id, idExport
    risk: Deletes or removes Trello data; requires destructive approval.
  post_cards:
    endpoint: POST /cards?token={{ secrets.token }}
    required fields: idList
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  put_cards_id:
    endpoint: PUT /cards/{{ record.id }}?token={{ secrets.token }}
    required fields: id
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  delete_cards_id:
    endpoint: DELETE /cards/{{ record.id }}?token={{ secrets.token }}
    required fields: id
    risk: Deletes or removes Trello data; requires destructive approval.
  post_cards_id_attachments:
    endpoint: POST /cards/{{ record.id }}/attachments?token={{ secrets.token }}
    required fields: id
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  deleted_cards_id_attachments_idattachment:
    endpoint: DELETE /cards/{{ record.id }}/attachments/{{ record.idAttachment }}?token={{ secrets.token }}
    required fields: id, idAttachment
    risk: Deletes or removes Trello data; requires destructive approval.
  post_cards_id_checklists:
    endpoint: POST /cards/{{ record.id }}/checklists?token={{ secrets.token }}
    required fields: id
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  put_cards_id_checkitem_idcheckitem:
    endpoint: PUT /cards/{{ record.id }}/checkItem/{{ record.idCheckItem }}?token={{ secrets.token }}
    required fields: id, idCheckItem
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  delete_cards_id_checkitem_idcheckitem:
    endpoint: DELETE /cards/{{ record.id }}/checkItem/{{ record.idCheckItem }}?token={{ secrets.token }}
    required fields: id, idCheckItem
    risk: Deletes or removes Trello data; requires destructive approval.
  cardsidmembersvoted_1:
    endpoint: POST /cards/{{ record.id }}/membersVoted?token={{ secrets.token }}
    required fields: id, value
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  post_cards_id_stickers:
    endpoint: POST /cards/{{ record.id }}/stickers?token={{ secrets.token }}
    required fields: id, image, top, left, zIndex
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  put_cards_id_stickers_idsticker:
    endpoint: PUT /cards/{{ record.id }}/stickers/{{ record.idSticker }}?token={{ secrets.token }}
    required fields: id, idSticker, top, left, zIndex
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  delete_cards_id_stickers_idsticker:
    endpoint: DELETE /cards/{{ record.id }}/stickers/{{ record.idSticker }}?token={{ secrets.token }}
    required fields: id, idSticker
    risk: Deletes or removes Trello data; requires destructive approval.
  put_cards_id_actions_idaction_comments:
    endpoint: PUT /cards/{{ record.id }}/actions/{{ record.idAction }}/comments?token={{ secrets.token }}
    required fields: id, idAction, text
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  delete_cards_id_actions_id_comments:
    endpoint: DELETE /cards/{{ record.id }}/actions/{{ record.idAction }}/comments?token={{ secrets.token }}
    required fields: id, idAction
    risk: Deletes or removes Trello data; requires destructive approval.
  put_cards_idcard_customfield_idcustomfield_item:
    endpoint: PUT /cards/{{ record.idCard }}/customField/{{ record.idCustomField }}/item?token={{ secrets.token }}
    required fields: idCard, idCustomField
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  put_cards_idcard_customfields:
    endpoint: PUT /cards/{{ record.idCard }}/customFields?token={{ secrets.token }}
    required fields: idCard
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  post_cards_id_actions_comments:
    endpoint: POST /cards/{{ record.id }}/actions/comments?token={{ secrets.token }}
    required fields: id, text
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  post_cards_id_idlabels:
    endpoint: POST /cards/{{ record.id }}/idLabels?token={{ secrets.token }}
    required fields: id
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  post_cards_id_idmembers:
    endpoint: POST /cards/{{ record.id }}/idMembers?token={{ secrets.token }}
    required fields: id
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  post_cards_id_labels:
    endpoint: POST /cards/{{ record.id }}/labels?token={{ secrets.token }}
    required fields: id, color
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  post_cards_id_markassociatednotificationsread:
    endpoint: POST /cards/{{ record.id }}/markAssociatedNotificationsRead?token={{ secrets.token }}
    required fields: id
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  delete_cards_id_idlabels_idlabel:
    endpoint: DELETE /cards/{{ record.id }}/idLabels/{{ record.idLabel }}?token={{ secrets.token }}
    required fields: id, idLabel
    risk: Deletes or removes Trello data; requires destructive approval.
  delete_id_idmembers_idmember:
    endpoint: DELETE /cards/{{ record.id }}/idMembers/{{ record.idMember }}?token={{ secrets.token }}
    required fields: id, idMember
    risk: Deletes or removes Trello data; requires destructive approval.
  delete_cards_id_membersvoted_idmember:
    endpoint: DELETE /cards/{{ record.id }}/membersVoted/{{ record.idMember }}?token={{ secrets.token }}
    required fields: id, idMember
    risk: Deletes or removes Trello data; requires destructive approval.
  put_cards_idcard_checklist_idchecklist_checkitem_idcheckitem:
    endpoint: PUT /cards/{{ record.idCard }}/checklist/{{ record.idChecklist }}/checkItem/{{ record.idCheckItem }}?token={{ secrets.token }}
    required fields: idCard, idChecklist, idCheckItem
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  delete_cards_id_checklists_idchecklist:
    endpoint: DELETE /cards/{{ record.id }}/checklists/{{ record.idChecklist }}?token={{ secrets.token }}
    required fields: id, idChecklist
    risk: Deletes or removes Trello data; requires destructive approval.
  post_checklists:
    endpoint: POST /checklists?token={{ secrets.token }}
    required fields: idCard
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  put_checlists_id:
    endpoint: PUT /checklists/{{ record.id }}?token={{ secrets.token }}
    required fields: id
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  delete_checklists_id:
    endpoint: DELETE /checklists/{{ record.id }}?token={{ secrets.token }}
    required fields: id
    risk: Deletes or removes Trello data; requires destructive approval.
  put_checklists_id_field:
    endpoint: PUT /checklists/{{ record.id }}/{{ record.field }}?token={{ secrets.token }}
    required fields: id, field, value
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  post_checklists_id_checkitems:
    endpoint: POST /checklists/{{ record.id }}/checkItems?token={{ secrets.token }}
    required fields: id, name
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  delete_checklists_id_checkitems_idcheckitem:
    endpoint: DELETE /checklists/{{ record.id }}/checkItems/{{ record.idCheckItem }}?token={{ secrets.token }}
    required fields: id, idCheckItem
    risk: Deletes or removes Trello data; requires destructive approval.
  post_customfields:
    endpoint: POST /customFields?token={{ secrets.token }}
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  put_customfields_id:
    endpoint: PUT /customFields/{{ record.id }}?token={{ secrets.token }}
    required fields: id
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  delete_customfields_id:
    endpoint: DELETE /customFields/{{ record.id }}?token={{ secrets.token }}
    required fields: id
    risk: Deletes or removes Trello data; requires destructive approval.
  get_customfields_id_options:
    endpoint: POST /customFields/{{ record.id }}/options?token={{ secrets.token }}
    required fields: id
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  delete_customfields_options_idcustomfieldoption:
    endpoint: DELETE /customFields/{{ record.id }}/options/{{ record.idCustomFieldOption }}?token={{ secrets.token }}
    required fields: id, idCustomFieldOption
    risk: Deletes or removes Trello data; requires destructive approval.
  put_labels_id:
    endpoint: PUT /labels/{{ record.id }}?token={{ secrets.token }}
    required fields: id
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  delete_labels_id:
    endpoint: DELETE /labels/{{ record.id }}?token={{ secrets.token }}
    required fields: id
    risk: Deletes or removes Trello data; requires destructive approval.
  put_labels_id_field:
    endpoint: PUT /labels/{{ record.id }}/{{ record.field }}?token={{ secrets.token }}
    required fields: id, field, value
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  post_labels:
    endpoint: POST /labels?token={{ secrets.token }}
    required fields: name, color, idBoard
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  put_lists_id:
    endpoint: PUT /lists/{{ record.id }}?token={{ secrets.token }}
    required fields: id
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  post_lists:
    endpoint: POST /lists?token={{ secrets.token }}
    required fields: name, idBoard
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  post_lists_id_archiveallcards:
    endpoint: POST /lists/{{ record.id }}/archiveAllCards?token={{ secrets.token }}
    required fields: id
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  post_lists_id_moveallcards:
    endpoint: POST /lists/{{ record.id }}/moveAllCards?token={{ secrets.token }}
    required fields: id, idBoard, idList
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  put_lists_id_closed:
    endpoint: PUT /lists/{{ record.id }}/closed?token={{ secrets.token }}
    required fields: id
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  put_id_idboard:
    endpoint: PUT /lists/{{ record.id }}/idBoard?token={{ secrets.token }}
    required fields: id, value
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  put_lists_id_field:
    endpoint: PUT /lists/{{ record.id }}/{{ record.field }}?token={{ secrets.token }}
    required fields: id, field
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  put_members_id:
    endpoint: PUT /members/{{ record.id }}?token={{ secrets.token }}
    required fields: id
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  post_members_id_boardbackgrounds_1:
    endpoint: POST /members/{{ record.id }}/boardBackgrounds?token={{ secrets.token }}
    required fields: id, file
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  put_members_id_boardbackgrounds_idbackground:
    endpoint: PUT /members/{{ record.id }}/boardBackgrounds/{{ record.idBackground }}?token={{ secrets.token }}
    required fields: id, idBackground
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  delete_members_id_boardbackgrounds_idbackground:
    endpoint: DELETE /members/{{ record.id }}/boardBackgrounds/{{ record.idBackground }}?token={{ secrets.token }}
    required fields: id, idBackground
    risk: Deletes or removes Trello data; requires destructive approval.
  post_members_id_boardstars:
    endpoint: POST /members/{{ record.id }}/boardStars?token={{ secrets.token }}
    required fields: id, idBoard, pos
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  put_members_id_boardstars_idstar:
    endpoint: PUT /members/{{ record.id }}/boardStars/{{ record.idStar }}?token={{ secrets.token }}
    required fields: id, idStar
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  delete_members_id_boardstars_idstar:
    endpoint: DELETE /members/{{ record.id }}/boardStars/{{ record.idStar }}?token={{ secrets.token }}
    required fields: id, idStar
    risk: Deletes or removes Trello data; requires destructive approval.
  membersidcustomboardbackgrounds_1:
    endpoint: POST /members/{{ record.id }}/customBoardBackgrounds?token={{ secrets.token }}
    required fields: id, file
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  put_members_id_customboardbackgrounds_idbackground:
    endpoint: PUT /members/{{ record.id }}/customBoardBackgrounds/{{ record.idBackground }}?token={{ secrets.token }}
    required fields: id, idBackground
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  delete_members_id_customboardbackgrounds_idbackground:
    endpoint: DELETE /members/{{ record.id }}/customBoardBackgrounds/{{ record.idBackground }}?token={{ secrets.token }}
    required fields: id, idBackground
    risk: Deletes or removes Trello data; requires destructive approval.
  post_members_id_customemoji:
    endpoint: POST /members/{{ record.id }}/customEmoji?token={{ secrets.token }}
    required fields: id, file, name
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  post_members_id_customstickers:
    endpoint: POST /members/{{ record.id }}/customStickers?token={{ secrets.token }}
    required fields: id, file
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  delete_members_id_customstickers_idsticker:
    endpoint: DELETE /members/{{ record.id }}/customStickers/{{ record.idSticker }}?token={{ secrets.token }}
    required fields: id, idSticker
    risk: Deletes or removes Trello data; requires destructive approval.
  post_members_id_savedsearches:
    endpoint: POST /members/{{ record.id }}/savedSearches?token={{ secrets.token }}
    required fields: id, name, query, pos
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  put_members_id_savedsearches_idsearch:
    endpoint: PUT /members/{{ record.id }}/savedSearches/{{ record.idSearch }}?token={{ secrets.token }}
    required fields: id, idSearch
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  delete_members_id_savedsearches_idsearch:
    endpoint: DELETE /members/{{ record.id }}/savedSearches/{{ record.idSearch }}?token={{ secrets.token }}
    required fields: id, idSearch
    risk: Deletes or removes Trello data; requires destructive approval.
  membersidavatar:
    endpoint: POST /members/{{ record.id }}/avatar?token={{ secrets.token }}
    required fields: id, file
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  post_members_id_onetimemessagesdismissed:
    endpoint: POST /members/{{ record.id }}/oneTimeMessagesDismissed?token={{ secrets.token }}
    required fields: id, value
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  put_members_id_notification_channel_settings_channel_blocked_keys:
    endpoint: PUT /members/{{ record.id }}/notificationsChannelSettings?token={{ secrets.token }}
    required fields: id
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  put_members_id_notification_channel_settings_channel_blocked_keys_2:
    endpoint: PUT /members/{{ record.id }}/notificationsChannelSettings/{{ record.channel }}?token={{ secrets.token }}
    required fields: id, channel
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  put_members_id_notification_channel_settings_channel_blocked_keys_3:
    endpoint: PUT /members/{{ record.id }}/notificationsChannelSettings/{{ record.channel }}/{{ record.blockedKeys }}?token={{ secrets.token }}
    required fields: id, channel, blockedKeys
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  put_notifications_id:
    endpoint: PUT /notifications/{{ record.id }}?token={{ secrets.token }}
    required fields: id
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  post_notifications_all_read:
    endpoint: POST /notifications/all/read?token={{ secrets.token }}
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  put_notifications_id_unread:
    endpoint: PUT /notifications/{{ record.id }}/unread?token={{ secrets.token }}
    required fields: id
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  post_organizations:
    endpoint: POST /organizations?token={{ secrets.token }}
    required fields: displayName
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  put_organizations_id:
    endpoint: PUT /organizations/{{ record.id }}?token={{ secrets.token }}
    required fields: id
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  delete_organizations_id:
    endpoint: DELETE /organizations/{{ record.id }}?token={{ secrets.token }}
    required fields: id
    risk: Deletes or removes Trello data; requires destructive approval.
  post_organizations_id_exports:
    endpoint: POST /organizations/{{ record.id }}/exports?token={{ secrets.token }}
    required fields: id
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  put_organizations_id_members:
    endpoint: PUT /organizations/{{ record.id }}/members?token={{ secrets.token }}
    required fields: id, email, fullName
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  post_organizations_id_tags:
    endpoint: POST /organizations/{{ record.id }}/tags?token={{ secrets.token }}
    required fields: id
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  put_organizations_id_members_idmember:
    endpoint: PUT /organizations/{{ record.id }}/members/{{ record.idMember }}?token={{ secrets.token }}
    required fields: id, idMember, type
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  delete_organizations_id_members:
    endpoint: DELETE /organizations/{{ record.id }}/members/{{ record.idMember }}?token={{ secrets.token }}
    required fields: id, idMember
    risk: Deletes or removes Trello data; requires destructive approval.
  put_organizations_id_members_idmember_deactivated:
    endpoint: PUT /organizations/{{ record.id }}/members/{{ record.idMember }}/deactivated?token={{ secrets.token }}
    required fields: id, idMember, value
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  post_organizations_id_logo:
    endpoint: POST /organizations/{{ record.id }}/logo?token={{ secrets.token }}
    required fields: id
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  delete_organizations_id_logo:
    endpoint: DELETE /organizations/{{ record.id }}/logo?token={{ secrets.token }}
    required fields: id
    risk: Deletes or removes Trello data; requires destructive approval.
  organizations_id_members_idmember_all:
    endpoint: DELETE /organizations/{{ record.id }}/members/{{ record.idMember }}/all?token={{ secrets.token }}
    required fields: id, idMember
    risk: Deletes or removes Trello data; requires destructive approval.
  delete_organizations_id_prefs_associateddomain:
    endpoint: DELETE /organizations/{{ record.id }}/prefs/associatedDomain?token={{ secrets.token }}
    required fields: id
    risk: Deletes or removes Trello data; requires destructive approval.
  delete_organizations_id_prefs_orginviterestrict:
    endpoint: DELETE /organizations/{{ record.id }}/prefs/orgInviteRestrict?token={{ secrets.token }}
    required fields: id
    risk: Deletes or removes Trello data; requires destructive approval.
  delete_organizations_id_tags_idtag:
    endpoint: DELETE /organizations/{{ record.id }}/tags/{{ record.idTag }}?token={{ secrets.token }}
    required fields: id, idTag
    risk: Deletes or removes Trello data; requires destructive approval.
  put_plugins_id:
    endpoint: PUT /plugins/{{ record.id }}/?token={{ secrets.token }}
    required fields: id
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  post_plugins_idplugin_listing:
    endpoint: POST /plugins/{{ record.idPlugin }}/listing?token={{ secrets.token }}
    required fields: idPlugin
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  put_plugins_idplugin_listings_idlisting:
    endpoint: PUT /plugins/{{ record.idPlugin }}/listings/{{ record.idListing }}?token={{ secrets.token }}
    required fields: idPlugin, idListing
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  post_webhooks:
    endpoint: POST /webhooks/?token={{ secrets.token }}
    required fields: callbackURL, idModel
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  put_webhooks_id:
    endpoint: PUT /webhooks/{{ record.id }}?token={{ secrets.token }}
    required fields: id
    risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.
  delete_webhooks_id:
    endpoint: DELETE /webhooks/{{ record.id }}?token={{ secrets.token }}
    required fields: id
    risk: Deletes or removes Trello data; requires destructive approval.

SECURITY
  read risk: External Trello API reads of boards, cards, lists, members, organizations, labels, checklists, actions, notifications, plugins, webhooks, search results, and attachments metadata.
  write risk: Typed Trello REST mutations can create, update, move, archive, delete, or otherwise change collaboration data and webhook delivery.
  approval: Reverse ETL writes require plan, preview, explicit approval, and execute; destructive deletes require destructive confirmation.
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Work with Trello REST resources from pm while the full operation ledger remains in connector metadata.
  Usage: pm trello <read|write> <operation> [flags]
  Global flags:
    --json (boolean): Write machine-readable JSON output.
    --credential (string): Use a saved Trello connector credential.
    --config (string_array): Set connector config as key=value (for example id=<board-id>).
    --limit (integer): Maximum records to emit for ETL read commands.
    --max-bytes (integer): Maximum bytes for direct-read JSON responses.
  Trello read and direct-read commands
    read get-actions-id - Get an Action [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-actions-id-board - Get the Board for an Action [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-actions-id-card - Get the Card for an Action [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-actions-id-list - Get the List for an Action [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-actions-id-member - Get the Member of an Action [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-actions-id-membercreator - Get the Member Creator of an Action [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-actions-id-organization - Get the Organization of an Action [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-actions-idaction-reactions - Get Action's Reactions [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id-action
    read get-actions-idaction-reactions-id - Get Action's Reaction [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id-action, --id
    read get-actions-idaction-reactionsummary - List Action's summary of Reactions [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id-action
    read get-boards-id-memberships - Get Memberships of a Board [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-boards-id - Get a Board [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-boards-id-actions - Get Actions of a Board [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --board-id
    read get-boards-id-boardstars - Get boardStars on a Board [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --board-id
    read checklists - Get Checklists on a Board [intent=etl availability=implemented stream=checklists]
    read get-boards-id-cards - Get Cards on a Board [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-boards-id-customfields - Get Custom Fields for Board [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-boards-id-labels - Get Labels on a Board [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read lists - Get Lists on a Board [intent=etl availability=implemented stream=lists]
    read get-boards-id-members - Get the Members of a Board [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-boards-id-boardplugins - Get Enabled Power-Ups on Board [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-board-id-plugins - Get Power-Ups on a Board [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-boards-id-exports-idexport - Get an Export for a Board [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id, --id-export
    read get-boards-id-exports-idexport-download - Download an Export for a Board [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id, --id-export
    read get-boards-id-exports-mostrecent - Get a Board's Most Recent Export [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-cards-id - Get a Card [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-cards-id-actions - Get Actions on a Card [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-cards-id-attachments - Get Attachments on a Card [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-cards-id-attachments-idattachment - Get an Attachment on a Card [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id, --id-attachment
    read get-cards-id-board - Get the Board the Card is on [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-cards-id-checkitemstates - Get checkItems on a Card [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-cards-id-checklists - Get Checklists on a Card [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-cards-id-checkitem-idcheckitem - Get checkItem on a Card [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id, --id-check-item
    read get-cards-id-list - Get the List of a Card [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-cards-id-members - Get the Members of a Card [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-cards-id-membersvoted - Get Members who have voted on a Card [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-cards-id-plugindata - Get pluginData on a Card [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-cards-id-stickers - Get Stickers on a Card [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-cards-id-stickers-idsticker - Get a Sticker on a Card [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id, --id-sticker
    read get-cards-id-customfielditems - Get Custom Field Items for a Card [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-checklists-id - Get a Checklist [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-checklists-id-board - Get the Board the Checklist is on [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-checklists-id-cards - Get the Card a Checklist is on [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-checklists-id-checkitems - Get Checkitems on a Checklist [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-checklists-id-checkitems-idcheckitem - Get a Checkitem on a Checklist [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id, --id-check-item
    read get-customfields-id - Get a Custom Field [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read post-customfields-id-options - Get Options of Custom Field drop down [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-customfields-options-idcustomfieldoption - Get Option of Custom Field dropdown [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id, --id-custom-field-option
    read emoji - List available Emoji [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.
    read get-labels-id - Get a Label [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-lists-id - Get a List [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-lists-id-actions - Get Actions for a List [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-lists-id-board - Get the Board a List is on [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-lists-id-cards - Get Cards in a List [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-members-id - Get a Member [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-members-id-actions - Get a Member's Actions [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-members-id-boardbackgrounds - Get Member's custom Board backgrounds [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-members-id-boardbackgrounds-idbackground - Get a boardBackground of a Member [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id, --id-background
    read get-members-id-boardstars - Get a Member's boardStars [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-members-id-boardstars-idstar - Get a boardStar of Member [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id, --id-star
    read boards - Get Boards that Member belongs to [intent=etl availability=implemented stream=boards]
    read get-members-id-boardsinvited - Get Boards the Member has been invited to [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-members-id-cards - Get Cards the Member is on [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-members-id-customboardbackgrounds - Get a Member's custom Board Backgrounds [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-members-id-customboardbackgrounds-idbackground - Get custom Board Background of Member [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id, --id-background
    read get-members-id-customemoji - Get a Member's customEmojis [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read membersidcustomemojiidemoji - Get a Member's custom Emoji [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id, --id-emoji
    read get-members-id-customstickers - Get Member's custom Stickers [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-members-id-customstickers-idsticker - Get a Member's custom Sticker [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id, --id-sticker
    read get-members-id-notifications - Get Member's Notifications [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-members-id-organizations - Get Member's Organizations [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-members-id-organizationsinvited - Get Organizations a Member has been invited to [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-members-id-savedsearches - Get Member's saved searched [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-members-id-savedsearches-idsearch - Get a saved search [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id, --id-search
    read get-members-id-notification-channel-settings - Get a Member's notification channel settings [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-members-id-notification-channel-settings-channel - Get blocked notification keys of Member on this channel [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id, --channel
    read get-notifications-id - Get a Notification [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-notifications-id-board - Get the Board a Notification is on [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-notifications-id-card - Get the Card a Notification is on [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-notifications-id-list - Get the List a Notification is on [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read notificationsidmember - Get the Member a Notification is about (not the creator) [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-notifications-id-membercreator - Get the Member who created the Notification [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-notifications-id-organization - Get a Notification's associated Organization [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-organizations-id - Get an Organization [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-organizations-id-actions - Get Actions for Organization [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-organizations-id-boards - Get Boards in an Organization [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-organizations-id-exports - Retrieve Organization's Exports [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-organizations-id-members - Get the Members of an Organization [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-organizations-id-memberships - Get Memberships of an Organization [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-organizations-id-memberships-idmembership - Get a Membership of an Organization [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id, --id-membership
    read get-organizations-id-plugindata - Get the pluginData Scoped to Organization [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-organizations-id-tags - Get Tags of an Organization [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-organizations-id-newbillableguests-idboard - Get Organizations new billable guests [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id, --id-board
    read get-plugins-id - Get a Plugin [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-plugins-id-compliance-memberprivacy - Get Plugin's Member privacy compliance [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
    read get-search - Search Trello [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --query
    read get-search-members - Search for Members [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --query
    read get-webhooks-id - Get a Webhook [intent=direct_read availability=implemented]; notes: Fixed Trello JSON direct read; response fields with secret/token/download/content-like names are redacted.; flags: --id
  Representative Trello reverse ETL write commands
    write create-board - Create a Trello board. [intent=reverse_etl availability=implemented write=post_boards]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.; flags: --name, --desc
    write create-list - Create a Trello list on a board. [intent=reverse_etl availability=implemented write=post_lists]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.; flags: --name, --id-board, --pos
    write create-card - Create a Trello card. [intent=reverse_etl availability=implemented write=post_cards]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.; flags: --id-list, --name, --desc, --due, --start, --due-complete
    write update-card - Update a Trello card. [intent=reverse_etl availability=implemented write=put_cards_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.; flags: --id, --name, --desc, --closed, --id-list
    write comment-card - Add a comment to a Trello card. [intent=reverse_etl availability=implemented write=post_cards_id_actions_comments]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.; flags: --id, --text
    write delete-card - Delete a Trello card. [intent=reverse_etl availability=implemented write=delete_cards_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Deletes or removes Trello data; requires destructive approval.; flags: --id
    write create-label - Create a Trello label. [intent=reverse_etl availability=implemented write=post_labels]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.; flags: --name, --color, --id-board
    write create-webhook - Create a Trello webhook. [intent=reverse_etl availability=implemented write=post_webhooks]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Mutates Trello data through a typed closed-schema REST action; execution requires plan, preview, approval, and execute.; flags: --callback-url, --id-model, --description
    write delete-webhook - Delete a Trello webhook. [intent=reverse_etl availability=implemented write=delete_webhooks_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Deletes or removes Trello data; requires destructive approval.; flags: --id
  Help topics:
    full-surface - All 261 official REST operations are tracked in api_surface.json; inspect the connector manifest for 3 streams, 95 direct reads, and 121 writes.
    safety - Writes are reverse ETL only: plan, preview, approve, execute; no raw HTTP escape hatch is exposed.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect trello

  # Inspect as structured JSON
  pm connectors inspect trello --json

AGENT WORKFLOW
  - Run pm connectors inspect trello before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
