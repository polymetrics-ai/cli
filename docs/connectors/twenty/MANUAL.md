# pm connectors inspect twenty

```text
NAME
  pm connectors inspect twenty - Twenty CRM connector manual

SYNOPSIS
  pm connectors inspect twenty
  pm connectors inspect twenty --json
  pm credentials add <name> --connector twenty [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Twenty CRM companies, people, opportunities, notes, tasks, messages, calendar events, workflows, workspace members, and the rest of the 28-object workspace surface, and writes create/update/batch/delete mutations through the Twenty REST API.

ICON
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  replication_start_date
  api_key (secret) (required)

ETL STREAMS
  attachments:
    primary key: id
    cursor: updatedAt
    fields: createdAt(string), createdBy(object), deletedAt(string), file(array), fileCategory(string), fullPath(string), id(string), name(string), position(number), searchVector(string), targetCompany(object), targetDashboard(object), targetNote(object), targetOpportunity(object), targetPerson(object), targetTask(object), targetWorkflow(object), updatedAt(string), updatedBy(object)
  blocklists:
    primary key: id
    cursor: updatedAt
    fields: createdAt(string), createdBy(object), deletedAt(string), handle(string), id(string), position(number), searchVector(string), updatedAt(string), updatedBy(object), workspaceMember(object)
  calendar_channel_event_associations:
    primary key: id
    cursor: updatedAt
    fields: calendarChannelId(string), calendarEvent(object), createdAt(string), createdBy(object), deletedAt(string), eventExternalId(string), id(string), position(number), recurringEventExternalId(string), searchVector(string), updatedAt(string), updatedBy(object)
  calendar_event_participants:
    primary key: id
    cursor: updatedAt
    fields: calendarEvent(object), createdAt(string), createdBy(object), deletedAt(string), displayName(string), handle(string), id(string), isOrganizer(boolean), person(object), position(number), responseStatus(string), searchVector(string), updatedAt(string), updatedBy(object), workspaceMember(object)
  calendar_events:
    primary key: id
    cursor: updatedAt
    fields: calendarChannelEventAssociations(object), calendarEventParticipants(object), callRecorderPreference(string), callRecordings(object), conferenceLink(object), conferenceSolution(string), createdAt(string), createdBy(object), deletedAt(string), description(string), endsAt(string), externalCreatedAt(string), externalUpdatedAt(string), iCalUid(string), id(string), isCanceled(boolean), isFullDay(boolean), lastContactForCompanies(object), lastContactForOpportunities(object), lastContactForPeople(object), lastMeetingForPeople(object), location(string), position(number), searchVector(string), startsAt(string), title(string), updatedAt(string), updatedBy(object)
  call_recordings:
    primary key: id
    cursor: updatedAt
    fields: applicationId(string), audio(array), calendarEvent(object), callRecorderFailureReason(string), createdAt(string), createdBy(object), deletedAt(string), endedAt(string), externalBotId(string), externalRecordingId(string), id(string), position(number), recordingRequestStatus(string), searchVector(string), startedAt(string), status(string), summary(string), title(string), transcript(), updatedAt(string), updatedBy(object), video(array)
  companies:
    primary key: id
    cursor: updatedAt
    fields: accountOwner(object), address(object), annualRevenue(object), attachments(object), createdAt(string), createdBy(object), deletedAt(string), domainName(object), id(string), lastContactAt(string), lastContactItemCalendarEvent(object), lastContactItemMessage(object), linkedinLink(object), name(string), noteTargets(object), opportunities(object), pdlAffiliatedProfiles(array), pdlAlternativeDomains(array), pdlAlternativeNames(array), pdlCompanyType(string), pdlEmployeeCount(number), pdlEmployeeCountByCountry(), pdlEnrichmentStatus(string), pdlFacebookUrl(object), pdlFoundedYear(number), pdlFundingStages(array), pdlHeadline(string), pdlId(string), pdlIndustry(string), pdlIndustryDetail(string), pdlLastEnrichedAt(string), pdlLastFundingDate(string), pdlLatestFundingStage(string), pdlLegalName(string), pdlLinkedinId(string), pdlLocationContinent(string), pdlLocationMetro(string), pdlMicExchange(string), pdlNaics(), pdlNumberFundingRounds(number), pdlRawPayload(), pdlSic(), pdlSizeRange(string), pdlSummary(string), pdlTags(array), pdlTicker(string), pdlTotalFunding(object), pdlTwitterUrl(object), people(object), position(number), searchVector(string), taskTargets(object), timelineActivities(object), updatedAt(string), updatedBy(object)
  dashboards:
    primary key: id
    cursor: updatedAt
    fields: attachments(object), createdAt(string), createdBy(object), deletedAt(string), id(string), pageLayoutId(string), position(number), searchVector(string), timelineActivities(object), title(string), updatedAt(string), updatedBy(object)
  message_campaigns:
    primary key: id
    cursor: updatedAt
    fields: bodyTemplate(string), bouncedCount(number), complainedCount(number), createdAt(string), createdBy(object), deletedAt(string), failedCount(number), fromAddress(object), id(string), list(object), messages(object), position(number), recipients(object), searchVector(string), sentAt(string), sentCount(number), status(string), subject(string), timelineActivities(object), unsubscribeTopicId(string), updatedAt(string), updatedBy(object)
  message_channel_message_association_message_folders:
    primary key: id
    cursor: updatedAt
    fields: createdAt(string), createdBy(object), deletedAt(string), id(string), messageChannelMessageAssociation(object), messageFolderId(string), position(number), searchVector(string), updatedAt(string), updatedBy(object)
  message_channel_message_associations:
    primary key: id
    cursor: updatedAt
    fields: createdAt(string), createdBy(object), deletedAt(string), direction(string), id(string), message(object), messageChannelId(string), messageExternalId(string), messageFolders(object), messageThread(object), messageThreadExternalId(string), position(number), searchVector(string), updatedAt(string), updatedBy(object)
  message_list_members:
    primary key: id
    cursor: updatedAt
    fields: createdAt(string), createdBy(object), deletedAt(string), id(string), list(object), person(object), position(number), searchVector(string), updatedAt(string), updatedBy(object)
  message_lists:
    primary key: id
    cursor: updatedAt
    fields: campaigns(object), createdAt(string), createdBy(object), deletedAt(string), id(string), members(object), name(string), position(number), searchVector(string), timelineActivities(object), updatedAt(string), updatedBy(object)
  message_participants:
    primary key: id
    cursor: updatedAt
    fields: createdAt(string), createdBy(object), deletedAt(string), displayName(string), handle(string), id(string), message(object), messageCampaign(object), person(object), position(number), role(string), searchVector(string), updatedAt(string), updatedBy(object), workspaceMember(object)
  message_threads:
    primary key: id
    cursor: updatedAt
    fields: createdAt(string), createdBy(object), deletedAt(string), id(string), messageChannelMessageAssociations(object), messages(object), position(number), searchVector(string), subject(string), updatedAt(string), updatedBy(object)
  messages:
    primary key: id
    cursor: updatedAt
    fields: createdAt(string), createdBy(object), deletedAt(string), deliveryStatus(string), headerMessageId(string), id(string), isDraft(boolean), lastContactForCompanies(object), lastContactForOpportunities(object), lastContactForPeople(object), lastEmailForPeople(object), messageCampaign(object), messageChannelMessageAssociations(object), messageParticipants(object), messageThread(object), position(number), receivedAt(string), searchVector(string), subject(string), text(string), updatedAt(string), updatedBy(object)
  note_targets:
    primary key: id
    cursor: updatedAt
    fields: createdAt(string), createdBy(object), deletedAt(string), id(string), note(object), position(number), searchVector(string), targetCompany(object), targetOpportunity(object), targetPerson(object), updatedAt(string), updatedBy(object)
  notes:
    primary key: id
    cursor: updatedAt
    fields: attachments(object), bodyV2(string), createdAt(string), createdBy(object), deletedAt(string), id(string), noteTargets(object), position(number), searchVector(string), timelineActivities(object), title(string), updatedAt(string), updatedBy(object)
  opportunities:
    primary key: id
    cursor: updatedAt
    fields: amount(object), attachments(object), closeDate(string), company(object), createdAt(string), createdBy(object), deletedAt(string), id(string), lastContactAt(string), lastContactItemCalendarEvent(object), lastContactItemMessage(object), name(string), noteTargets(object), owner(object), pointOfContact(object), position(number), searchVector(string), stage(string), taskTargets(object), timelineActivities(object), updatedAt(string), updatedBy(object)
  people:
    primary key: id
    cursor: updatedAt
    fields: attachments(object), avatarFile(array), avatarUrl(string), calendarEventParticipants(object), company(object), createdAt(string), createdBy(object), deletedAt(string), emails(object), id(string), jobTitle(string), lastContactAt(string), lastContactBy(object), lastContactItemCalendarEvent(object), lastContactItemMessage(object), lastEmail(object), lastInboundAt(string), lastMeeting(object), lastOutboundAt(string), linkedinLink(object), listMemberships(object), messageParticipants(object), name(object), noteTargets(object), pdlBirthDate(string), pdlBirthYear(number), pdlCertifications(), pdlEducation(), pdlEnrichmentStatus(string), pdlExperience(), pdlFacebookUrl(object), pdlGithubUrl(object), pdlHeadline(string), pdlId(string), pdlIndustry(string), pdlInferredSalary(string), pdlInterests(array), pdlJobOnetCode(string), pdlJobRole(string), pdlJobStartDate(string), pdlJobSummary(string), pdlJobTitleClass(string), pdlJobTitleSubRole(string), pdlLanguages(), pdlLastEnrichedAt(string), pdlLikelihood(number), pdlLinkedinConnections(number), pdlLinkedinUsername(string), pdlLocation(object), pdlLocationMetro(string), pdlNameAliases(array), pdlProfiles(), pdlRawPayload(), pdlSeniority(array), pdlSex(string), pdlSkills(array), pdlSummary(string), pdlTwitterUrl(object), pdlYearsExperience(number), phones(object), pointOfContactForOpportunities(object), position(number), searchVector(string), taskTargets(object), timelineActivities(object), updatedAt(string), updatedBy(object)
  task_targets:
    primary key: id
    cursor: updatedAt
    fields: createdAt(string), createdBy(object), deletedAt(string), id(string), position(number), searchVector(string), targetCompany(object), targetOpportunity(object), targetPerson(object), task(object), updatedAt(string), updatedBy(object)
  tasks:
    primary key: id
    cursor: updatedAt
    fields: assignee(object), attachments(object), bodyV2(string), createdAt(string), createdBy(object), deletedAt(string), dueAt(string), id(string), position(number), searchVector(string), status(string), taskTargets(object), timelineActivities(object), title(string), updatedAt(string), updatedBy(object)
  timeline_activities:
    primary key: id
    cursor: updatedAt
    fields: createdAt(string), createdBy(object), deletedAt(string), happensAt(string), id(string), linkedObjectMetadataId(string), linkedRecordCachedName(string), linkedRecordId(string), name(string), position(number), properties(), searchVector(string), targetCompany(object), targetDashboard(object), targetMessageCampaign(object), targetMessageList(object), targetNote(object), targetOpportunity(object), targetPerson(object), targetTask(object), targetWorkflow(object), targetWorkflowRun(object), targetWorkflowVersion(object), updatedAt(string), updatedBy(object), workspaceMember(object)
  workflow_automated_triggers:
    primary key: id
    cursor: updatedAt
    fields: createdAt(string), createdBy(object), deletedAt(string), id(string), position(number), searchVector(string), settings(), type(string), updatedAt(string), updatedBy(object), workflow(object)
  workflow_runs:
    primary key: id
    cursor: updatedAt
    fields: createdAt(string), createdBy(object), deletedAt(string), endedAt(string), enqueuedAt(string), id(string), name(string), position(number), searchVector(string), startedAt(string), state(), status(string), stepLogs(), timelineActivities(object), updatedAt(string), updatedBy(object), workflow(object), workflowVersion(object)
  workflow_versions:
    primary key: id
    cursor: updatedAt
    fields: createdAt(string), createdBy(object), deletedAt(string), id(string), name(string), position(number), runs(object), searchVector(string), status(string), steps(), timelineActivities(object), trigger(), updatedAt(string), updatedBy(object), workflow(object)
  workflows:
    primary key: id
    cursor: updatedAt
    fields: attachments(object), automatedTriggers(object), createdAt(string), createdBy(object), deletedAt(string), id(string), lastPublishedVersionId(string), name(string), position(number), runs(object), searchVector(string), statuses(array), timelineActivities(object), updatedAt(string), updatedBy(object), versions(object)
  workspace_members:
    primary key: id
    cursor: updatedAt
    fields: accountOwnerForCompanies(object), assignedTasks(object), avatarUrl(string), blocklist(object), calendarEventParticipants(object), calendarStartDay(number), colorScheme(string), createdAt(string), createdBy(object), dateFormat(string), deletedAt(string), id(string), jobTitle(string), lastContactForPeople(object), locale(string), messageParticipants(object), name(object), numberFormat(string), ownedOpportunities(object), position(number), searchVector(string), timeFormat(string), timeZone(string), timelineActivities(object), updatedAt(string), updatedBy(object), userEmail(string), userId(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_attachments:
    endpoint: POST /rest/attachments
    risk: creates a Twenty attachments record; external mutation; reverse ETL plan, preview, approval, execute required
  update_attachments:
    endpoint: PATCH /rest/attachments/{{ record.id }}
    required fields: id
    risk: updates an existing Twenty attachments record; external mutation; reverse ETL plan, preview, approval, execute required
  batch_attachments:
    endpoint: POST /rest/batch/attachments
    required fields: records
    risk: bulk creates up to 60 Twenty attachments records; external mutation; reverse ETL plan, preview, approval, execute required; Twenty enforces batch size
  create_blocklists:
    endpoint: POST /rest/blocklists
    risk: creates a Twenty blocklists record; external mutation; reverse ETL plan, preview, approval, execute required
  update_blocklists:
    endpoint: PATCH /rest/blocklists/{{ record.id }}
    required fields: id
    risk: updates an existing Twenty blocklists record; external mutation; reverse ETL plan, preview, approval, execute required
  batch_blocklists:
    endpoint: POST /rest/batch/blocklists
    required fields: records
    risk: bulk creates up to 60 Twenty blocklists records; external mutation; reverse ETL plan, preview, approval, execute required; Twenty enforces batch size
  create_calendar_channel_event_associations:
    endpoint: POST /rest/calendarChannelEventAssociations
    risk: creates a Twenty calendar_channel_event_associations record; external mutation; reverse ETL plan, preview, approval, execute required
  update_calendar_channel_event_associations:
    endpoint: PATCH /rest/calendarChannelEventAssociations/{{ record.id }}
    required fields: id
    risk: updates an existing Twenty calendar_channel_event_associations record; external mutation; reverse ETL plan, preview, approval, execute required
  batch_calendar_channel_event_associations:
    endpoint: POST /rest/batch/calendarChannelEventAssociations
    required fields: records
    risk: bulk creates up to 60 Twenty calendar_channel_event_associations records; external mutation; reverse ETL plan, preview, approval, execute required; Twenty enforces batch size
  create_calendar_event_participants:
    endpoint: POST /rest/calendarEventParticipants
    risk: creates a Twenty calendar_event_participants record; external mutation; reverse ETL plan, preview, approval, execute required
  update_calendar_event_participants:
    endpoint: PATCH /rest/calendarEventParticipants/{{ record.id }}
    required fields: id
    risk: updates an existing Twenty calendar_event_participants record; external mutation; reverse ETL plan, preview, approval, execute required
  batch_calendar_event_participants:
    endpoint: POST /rest/batch/calendarEventParticipants
    required fields: records
    risk: bulk creates up to 60 Twenty calendar_event_participants records; external mutation; reverse ETL plan, preview, approval, execute required; Twenty enforces batch size
  create_calendar_events:
    endpoint: POST /rest/calendarEvents
    risk: creates a Twenty calendar_events record; external mutation; reverse ETL plan, preview, approval, execute required
  update_calendar_events:
    endpoint: PATCH /rest/calendarEvents/{{ record.id }}
    required fields: id
    risk: updates an existing Twenty calendar_events record; external mutation; reverse ETL plan, preview, approval, execute required
  batch_calendar_events:
    endpoint: POST /rest/batch/calendarEvents
    required fields: records
    risk: bulk creates up to 60 Twenty calendar_events records; external mutation; reverse ETL plan, preview, approval, execute required; Twenty enforces batch size
  create_call_recordings:
    endpoint: POST /rest/callRecordings
    risk: creates a Twenty call_recordings record; external mutation; reverse ETL plan, preview, approval, execute required
  update_call_recordings:
    endpoint: PATCH /rest/callRecordings/{{ record.id }}
    required fields: id
    risk: updates an existing Twenty call_recordings record; external mutation; reverse ETL plan, preview, approval, execute required
  batch_call_recordings:
    endpoint: POST /rest/batch/callRecordings
    required fields: records
    risk: bulk creates up to 60 Twenty call_recordings records; external mutation; reverse ETL plan, preview, approval, execute required; Twenty enforces batch size
  create_companies:
    endpoint: POST /rest/companies
    risk: creates a Twenty companies record; external mutation; reverse ETL plan, preview, approval, execute required
  update_companies:
    endpoint: PATCH /rest/companies/{{ record.id }}
    required fields: id
    risk: updates an existing Twenty companies record; external mutation; reverse ETL plan, preview, approval, execute required
  batch_companies:
    endpoint: POST /rest/batch/companies
    required fields: records
    risk: bulk creates up to 60 Twenty companies records; external mutation; reverse ETL plan, preview, approval, execute required; Twenty enforces batch size
  create_dashboards:
    endpoint: POST /rest/dashboards
    risk: creates a Twenty dashboards record; external mutation; reverse ETL plan, preview, approval, execute required
  update_dashboards:
    endpoint: PATCH /rest/dashboards/{{ record.id }}
    required fields: id
    risk: updates an existing Twenty dashboards record; external mutation; reverse ETL plan, preview, approval, execute required
  batch_dashboards:
    endpoint: POST /rest/batch/dashboards
    required fields: records
    risk: bulk creates up to 60 Twenty dashboards records; external mutation; reverse ETL plan, preview, approval, execute required; Twenty enforces batch size
  create_message_campaigns:
    endpoint: POST /rest/messageCampaigns
    risk: creates a Twenty message_campaigns record; external mutation; reverse ETL plan, preview, approval, execute required
  update_message_campaigns:
    endpoint: PATCH /rest/messageCampaigns/{{ record.id }}
    required fields: id
    risk: updates an existing Twenty message_campaigns record; external mutation; reverse ETL plan, preview, approval, execute required
  batch_message_campaigns:
    endpoint: POST /rest/batch/messageCampaigns
    required fields: records
    risk: bulk creates up to 60 Twenty message_campaigns records; external mutation; reverse ETL plan, preview, approval, execute required; Twenty enforces batch size
  create_message_channel_message_association_message_folders:
    endpoint: POST /rest/messageChannelMessageAssociationMessageFolders
    risk: creates a Twenty message_channel_message_association_message_folders record; external mutation; reverse ETL plan, preview, approval, execute required
  update_message_channel_message_association_message_folders:
    endpoint: PATCH /rest/messageChannelMessageAssociationMessageFolders/{{ record.id }}
    required fields: id
    risk: updates an existing Twenty message_channel_message_association_message_folders record; external mutation; reverse ETL plan, preview, approval, execute required
  batch_message_channel_message_association_message_folders:
    endpoint: POST /rest/batch/messageChannelMessageAssociationMessageFolders
    required fields: records
    risk: bulk creates up to 60 Twenty message_channel_message_association_message_folders records; external mutation; reverse ETL plan, preview, approval, execute required; Twenty enforces batch size
  create_message_channel_message_associations:
    endpoint: POST /rest/messageChannelMessageAssociations
    risk: creates a Twenty message_channel_message_associations record; external mutation; reverse ETL plan, preview, approval, execute required
  update_message_channel_message_associations:
    endpoint: PATCH /rest/messageChannelMessageAssociations/{{ record.id }}
    required fields: id
    risk: updates an existing Twenty message_channel_message_associations record; external mutation; reverse ETL plan, preview, approval, execute required
  batch_message_channel_message_associations:
    endpoint: POST /rest/batch/messageChannelMessageAssociations
    required fields: records
    risk: bulk creates up to 60 Twenty message_channel_message_associations records; external mutation; reverse ETL plan, preview, approval, execute required; Twenty enforces batch size
  create_message_list_members:
    endpoint: POST /rest/messageListMembers
    risk: creates a Twenty message_list_members record; external mutation; reverse ETL plan, preview, approval, execute required
  update_message_list_members:
    endpoint: PATCH /rest/messageListMembers/{{ record.id }}
    required fields: id
    risk: updates an existing Twenty message_list_members record; external mutation; reverse ETL plan, preview, approval, execute required
  batch_message_list_members:
    endpoint: POST /rest/batch/messageListMembers
    required fields: records
    risk: bulk creates up to 60 Twenty message_list_members records; external mutation; reverse ETL plan, preview, approval, execute required; Twenty enforces batch size
  create_message_lists:
    endpoint: POST /rest/messageLists
    risk: creates a Twenty message_lists record; external mutation; reverse ETL plan, preview, approval, execute required
  update_message_lists:
    endpoint: PATCH /rest/messageLists/{{ record.id }}
    required fields: id
    risk: updates an existing Twenty message_lists record; external mutation; reverse ETL plan, preview, approval, execute required
  batch_message_lists:
    endpoint: POST /rest/batch/messageLists
    required fields: records
    risk: bulk creates up to 60 Twenty message_lists records; external mutation; reverse ETL plan, preview, approval, execute required; Twenty enforces batch size
  create_message_participants:
    endpoint: POST /rest/messageParticipants
    risk: creates a Twenty message_participants record; external mutation; reverse ETL plan, preview, approval, execute required
  update_message_participants:
    endpoint: PATCH /rest/messageParticipants/{{ record.id }}
    required fields: id
    risk: updates an existing Twenty message_participants record; external mutation; reverse ETL plan, preview, approval, execute required
  batch_message_participants:
    endpoint: POST /rest/batch/messageParticipants
    required fields: records
    risk: bulk creates up to 60 Twenty message_participants records; external mutation; reverse ETL plan, preview, approval, execute required; Twenty enforces batch size
  create_message_threads:
    endpoint: POST /rest/messageThreads
    risk: creates a Twenty message_threads record; external mutation; reverse ETL plan, preview, approval, execute required
  update_message_threads:
    endpoint: PATCH /rest/messageThreads/{{ record.id }}
    required fields: id
    risk: updates an existing Twenty message_threads record; external mutation; reverse ETL plan, preview, approval, execute required
  batch_message_threads:
    endpoint: POST /rest/batch/messageThreads
    required fields: records
    risk: bulk creates up to 60 Twenty message_threads records; external mutation; reverse ETL plan, preview, approval, execute required; Twenty enforces batch size
  create_messages:
    endpoint: POST /rest/messages
    risk: creates a Twenty messages record; external mutation; reverse ETL plan, preview, approval, execute required
  update_messages:
    endpoint: PATCH /rest/messages/{{ record.id }}
    required fields: id
    risk: updates an existing Twenty messages record; external mutation; reverse ETL plan, preview, approval, execute required
  batch_messages:
    endpoint: POST /rest/batch/messages
    required fields: records
    risk: bulk creates up to 60 Twenty messages records; external mutation; reverse ETL plan, preview, approval, execute required; Twenty enforces batch size
  create_note_targets:
    endpoint: POST /rest/noteTargets
    risk: creates a Twenty note_targets record; external mutation; reverse ETL plan, preview, approval, execute required
  update_note_targets:
    endpoint: PATCH /rest/noteTargets/{{ record.id }}
    required fields: id
    risk: updates an existing Twenty note_targets record; external mutation; reverse ETL plan, preview, approval, execute required
  batch_note_targets:
    endpoint: POST /rest/batch/noteTargets
    required fields: records
    risk: bulk creates up to 60 Twenty note_targets records; external mutation; reverse ETL plan, preview, approval, execute required; Twenty enforces batch size
  create_notes:
    endpoint: POST /rest/notes
    risk: creates a Twenty notes record; external mutation; reverse ETL plan, preview, approval, execute required
  update_notes:
    endpoint: PATCH /rest/notes/{{ record.id }}
    required fields: id
    risk: updates an existing Twenty notes record; external mutation; reverse ETL plan, preview, approval, execute required
  batch_notes:
    endpoint: POST /rest/batch/notes
    required fields: records
    risk: bulk creates up to 60 Twenty notes records; external mutation; reverse ETL plan, preview, approval, execute required; Twenty enforces batch size
  create_opportunities:
    endpoint: POST /rest/opportunities
    risk: creates a Twenty opportunities record; external mutation; reverse ETL plan, preview, approval, execute required
  update_opportunities:
    endpoint: PATCH /rest/opportunities/{{ record.id }}
    required fields: id
    risk: updates an existing Twenty opportunities record; external mutation; reverse ETL plan, preview, approval, execute required
  batch_opportunities:
    endpoint: POST /rest/batch/opportunities
    required fields: records
    risk: bulk creates up to 60 Twenty opportunities records; external mutation; reverse ETL plan, preview, approval, execute required; Twenty enforces batch size
  create_people:
    endpoint: POST /rest/people
    risk: creates a Twenty people record; external mutation; reverse ETL plan, preview, approval, execute required
  update_people:
    endpoint: PATCH /rest/people/{{ record.id }}
    required fields: id
    risk: updates an existing Twenty people record; external mutation; reverse ETL plan, preview, approval, execute required
  batch_people:
    endpoint: POST /rest/batch/people
    required fields: records
    risk: bulk creates up to 60 Twenty people records; external mutation; reverse ETL plan, preview, approval, execute required; Twenty enforces batch size
  create_task_targets:
    endpoint: POST /rest/taskTargets
    risk: creates a Twenty task_targets record; external mutation; reverse ETL plan, preview, approval, execute required
  update_task_targets:
    endpoint: PATCH /rest/taskTargets/{{ record.id }}
    required fields: id
    risk: updates an existing Twenty task_targets record; external mutation; reverse ETL plan, preview, approval, execute required
  batch_task_targets:
    endpoint: POST /rest/batch/taskTargets
    required fields: records
    risk: bulk creates up to 60 Twenty task_targets records; external mutation; reverse ETL plan, preview, approval, execute required; Twenty enforces batch size
  create_tasks:
    endpoint: POST /rest/tasks
    risk: creates a Twenty tasks record; external mutation; reverse ETL plan, preview, approval, execute required
  update_tasks:
    endpoint: PATCH /rest/tasks/{{ record.id }}
    required fields: id
    risk: updates an existing Twenty tasks record; external mutation; reverse ETL plan, preview, approval, execute required
  batch_tasks:
    endpoint: POST /rest/batch/tasks
    required fields: records
    risk: bulk creates up to 60 Twenty tasks records; external mutation; reverse ETL plan, preview, approval, execute required; Twenty enforces batch size
  create_timeline_activities:
    endpoint: POST /rest/timelineActivities
    risk: creates a Twenty timeline_activities record; external mutation; reverse ETL plan, preview, approval, execute required
  update_timeline_activities:
    endpoint: PATCH /rest/timelineActivities/{{ record.id }}
    required fields: id
    risk: updates an existing Twenty timeline_activities record; external mutation; reverse ETL plan, preview, approval, execute required
  batch_timeline_activities:
    endpoint: POST /rest/batch/timelineActivities
    required fields: records
    risk: bulk creates up to 60 Twenty timeline_activities records; external mutation; reverse ETL plan, preview, approval, execute required; Twenty enforces batch size
  create_workflow_automated_triggers:
    endpoint: POST /rest/workflowAutomatedTriggers
    risk: creates a Twenty workflow_automated_triggers record; external mutation; reverse ETL plan, preview, approval, execute required
  update_workflow_automated_triggers:
    endpoint: PATCH /rest/workflowAutomatedTriggers/{{ record.id }}
    required fields: id
    risk: updates an existing Twenty workflow_automated_triggers record; external mutation; reverse ETL plan, preview, approval, execute required
  batch_workflow_automated_triggers:
    endpoint: POST /rest/batch/workflowAutomatedTriggers
    required fields: records
    risk: bulk creates up to 60 Twenty workflow_automated_triggers records; external mutation; reverse ETL plan, preview, approval, execute required; Twenty enforces batch size
  create_workflow_runs:
    endpoint: POST /rest/workflowRuns
    risk: creates a Twenty workflow_runs record; external mutation; reverse ETL plan, preview, approval, execute required
  update_workflow_runs:
    endpoint: PATCH /rest/workflowRuns/{{ record.id }}
    required fields: id
    risk: updates an existing Twenty workflow_runs record; external mutation; reverse ETL plan, preview, approval, execute required
  batch_workflow_runs:
    endpoint: POST /rest/batch/workflowRuns
    required fields: records
    risk: bulk creates up to 60 Twenty workflow_runs records; external mutation; reverse ETL plan, preview, approval, execute required; Twenty enforces batch size
  create_workflow_versions:
    endpoint: POST /rest/workflowVersions
    risk: creates a Twenty workflow_versions record; external mutation; reverse ETL plan, preview, approval, execute required
  update_workflow_versions:
    endpoint: PATCH /rest/workflowVersions/{{ record.id }}
    required fields: id
    risk: updates an existing Twenty workflow_versions record; external mutation; reverse ETL plan, preview, approval, execute required
  batch_workflow_versions:
    endpoint: POST /rest/batch/workflowVersions
    required fields: records
    risk: bulk creates up to 60 Twenty workflow_versions records; external mutation; reverse ETL plan, preview, approval, execute required; Twenty enforces batch size
  create_workflows:
    endpoint: POST /rest/workflows
    risk: creates a Twenty workflows record; external mutation; reverse ETL plan, preview, approval, execute required
  update_workflows:
    endpoint: PATCH /rest/workflows/{{ record.id }}
    required fields: id
    risk: updates an existing Twenty workflows record; external mutation; reverse ETL plan, preview, approval, execute required
  batch_workflows:
    endpoint: POST /rest/batch/workflows
    required fields: records
    risk: bulk creates up to 60 Twenty workflows records; external mutation; reverse ETL plan, preview, approval, execute required; Twenty enforces batch size
  create_workspace_members:
    endpoint: POST /rest/workspaceMembers
    risk: creates a Twenty workspace_members record; external mutation; reverse ETL plan, preview, approval, execute required
  update_workspace_members:
    endpoint: PATCH /rest/workspaceMembers/{{ record.id }}
    required fields: id
    risk: updates an existing Twenty workspace_members record; external mutation; reverse ETL plan, preview, approval, execute required
  batch_workspace_members:
    endpoint: POST /rest/batch/workspaceMembers
    required fields: records
    risk: bulk creates up to 60 Twenty workspace_members records; external mutation; reverse ETL plan, preview, approval, execute required; Twenty enforces batch size
  delete_attachments:
    endpoint: DELETE /rest/attachments/{{ record.id }}
    required fields: id
    risk: destructive external deletion; typed confirmation required
  delete_blocklists:
    endpoint: DELETE /rest/blocklists/{{ record.id }}
    required fields: id
    risk: destructive external deletion; typed confirmation required
  delete_calendar_channel_event_associations:
    endpoint: DELETE /rest/calendarChannelEventAssociations/{{ record.id }}
    required fields: id
    risk: destructive external deletion; typed confirmation required
  delete_calendar_event_participants:
    endpoint: DELETE /rest/calendarEventParticipants/{{ record.id }}
    required fields: id
    risk: destructive external deletion; typed confirmation required
  delete_calendar_events:
    endpoint: DELETE /rest/calendarEvents/{{ record.id }}
    required fields: id
    risk: destructive external deletion; typed confirmation required
  delete_call_recordings:
    endpoint: DELETE /rest/callRecordings/{{ record.id }}
    required fields: id
    risk: destructive external deletion; typed confirmation required
  delete_companies:
    endpoint: DELETE /rest/companies/{{ record.id }}
    required fields: id
    risk: destructive external deletion; typed confirmation required
  delete_dashboards:
    endpoint: DELETE /rest/dashboards/{{ record.id }}
    required fields: id
    risk: destructive external deletion; typed confirmation required
  delete_message_campaigns:
    endpoint: DELETE /rest/messageCampaigns/{{ record.id }}
    required fields: id
    risk: destructive external deletion; typed confirmation required
  delete_message_channel_message_association_message_folders:
    endpoint: DELETE /rest/messageChannelMessageAssociationMessageFolders/{{ record.id }}
    required fields: id
    risk: destructive external deletion; typed confirmation required
  delete_message_channel_message_associations:
    endpoint: DELETE /rest/messageChannelMessageAssociations/{{ record.id }}
    required fields: id
    risk: destructive external deletion; typed confirmation required
  delete_message_list_members:
    endpoint: DELETE /rest/messageListMembers/{{ record.id }}
    required fields: id
    risk: destructive external deletion; typed confirmation required
  delete_message_lists:
    endpoint: DELETE /rest/messageLists/{{ record.id }}
    required fields: id
    risk: destructive external deletion; typed confirmation required
  delete_message_participants:
    endpoint: DELETE /rest/messageParticipants/{{ record.id }}
    required fields: id
    risk: destructive external deletion; typed confirmation required
  delete_message_threads:
    endpoint: DELETE /rest/messageThreads/{{ record.id }}
    required fields: id
    risk: destructive external deletion; typed confirmation required
  delete_messages:
    endpoint: DELETE /rest/messages/{{ record.id }}
    required fields: id
    risk: destructive external deletion; typed confirmation required
  delete_note_targets:
    endpoint: DELETE /rest/noteTargets/{{ record.id }}
    required fields: id
    risk: destructive external deletion; typed confirmation required
  delete_notes:
    endpoint: DELETE /rest/notes/{{ record.id }}
    required fields: id
    risk: destructive external deletion; typed confirmation required
  delete_opportunities:
    endpoint: DELETE /rest/opportunities/{{ record.id }}
    required fields: id
    risk: destructive external deletion; typed confirmation required
  delete_people:
    endpoint: DELETE /rest/people/{{ record.id }}
    required fields: id
    risk: destructive external deletion; typed confirmation required
  delete_task_targets:
    endpoint: DELETE /rest/taskTargets/{{ record.id }}
    required fields: id
    risk: destructive external deletion; typed confirmation required
  delete_tasks:
    endpoint: DELETE /rest/tasks/{{ record.id }}
    required fields: id
    risk: destructive external deletion; typed confirmation required
  delete_timeline_activities:
    endpoint: DELETE /rest/timelineActivities/{{ record.id }}
    required fields: id
    risk: destructive external deletion; typed confirmation required
  delete_workflow_automated_triggers:
    endpoint: DELETE /rest/workflowAutomatedTriggers/{{ record.id }}
    required fields: id
    risk: destructive external deletion; typed confirmation required
  delete_workflow_runs:
    endpoint: DELETE /rest/workflowRuns/{{ record.id }}
    required fields: id
    risk: destructive external deletion; typed confirmation required
  delete_workflow_versions:
    endpoint: DELETE /rest/workflowVersions/{{ record.id }}
    required fields: id
    risk: destructive external deletion; typed confirmation required
  delete_workflows:
    endpoint: DELETE /rest/workflows/{{ record.id }}
    required fields: id
    risk: destructive external deletion; typed confirmation required
  delete_workspace_members:
    endpoint: DELETE /rest/workspaceMembers/{{ record.id }}
    required fields: id
    risk: destructive external deletion; typed confirmation required

SECURITY
  read risk: external Twenty CRM API read of CRM, messaging, calendar, workflow, and workspace-member data
  write risk: creates, updates, batch-writes, and deletes records across all 28 Twenty CRM REST objects
  approval: required for every create_<object>, update_<object>, batch_<object>, and delete_<object> action across the Twenty REST object surface; delete_<object> actions additionally require typed --confirm destructive
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Work with Twenty CRM objects from the command line.
  Usage: pm twenty <object> <verb> [flags]
  Global flags:
    --json (boolean): Write machine-readable JSON output.
    --credential (string): Use a saved Twenty credential by name.: maps_to=credential
    --connection (string): Alias for selecting a saved Twenty connector credential.: maps_to=connection
    --limit (integer): Limit records returned by implemented list commands.: maps_to=limit
    --preview (boolean): Build a reverse ETL preview for implemented write commands before approval/execution.: maps_to=preview
    --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
  Twenty Objects
    attachments list - List Twenty attachments records [intent=etl availability=implemented stream=attachments]; flags: --limit
    attachments get - Get one Twenty attachments record by ID [intent=direct_read availability=implemented operation=twenty.attachments.get]; flags: --id (required), --page, --page-cursor
    attachments create - Create a Twenty attachments record [intent=reverse_etl availability=implemented write=create_attachments]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a Twenty attachments record through the REST API.; flags: --name, --position
    attachments update - Update a Twenty attachments record [intent=reverse_etl availability=implemented write=update_attachments]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Updates an existing Twenty attachments record through the REST API.; flags: --id, --name, --position
    attachments batch - Batch-create Twenty attachments records [intent=reverse_etl availability=implemented write=batch_attachments]; approval: reverse ETL batch writes require plan, preview, approval, execute.; risk: Bulk creates Twenty attachments records through the REST API.; flags: --records (required)
    attachments delete - Delete a Twenty attachments record [intent=reverse_etl availability=implemented write=delete_attachments]; approval: reverse ETL deletes require plan, preview, approval, execute, and typed --confirm destructive.; risk: Destructively deletes a Twenty attachments record through the REST API.; flags: --id
    blocklists list - List Twenty blocklists records [intent=etl availability=implemented stream=blocklists]; flags: --limit
    blocklists get - Get one Twenty blocklists record by ID [intent=direct_read availability=implemented operation=twenty.blocklists.get]; flags: --id (required), --page, --page-cursor
    blocklists create - Create a Twenty blocklists record [intent=reverse_etl availability=implemented write=create_blocklists]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a Twenty blocklists record through the REST API.; flags: --handle, --position
    blocklists update - Update a Twenty blocklists record [intent=reverse_etl availability=implemented write=update_blocklists]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Updates an existing Twenty blocklists record through the REST API.; flags: --id, --handle, --position
    blocklists batch - Batch-create Twenty blocklists records [intent=reverse_etl availability=implemented write=batch_blocklists]; approval: reverse ETL batch writes require plan, preview, approval, execute.; risk: Bulk creates Twenty blocklists records through the REST API.; flags: --records (required)
    blocklists delete - Delete a Twenty blocklists record [intent=reverse_etl availability=implemented write=delete_blocklists]; approval: reverse ETL deletes require plan, preview, approval, execute, and typed --confirm destructive.; risk: Destructively deletes a Twenty blocklists record through the REST API.; flags: --id
    calendar-channel-event-associations list - List Twenty calendar channel event associations records [intent=etl availability=implemented stream=calendar_channel_event_associations]; flags: --limit
    calendar-channel-event-associations get - Get one Twenty calendar channel event associations record by ID [intent=direct_read availability=implemented operation=twenty.calendar-channel-event-associations.get]; flags: --id (required), --page, --page-cursor
    calendar-channel-event-associations create - Create a Twenty calendar channel event associations record [intent=reverse_etl availability=implemented write=create_calendar_channel_event_associations]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a Twenty calendar channel event associations record through the REST API.; flags: --calendar-channel-id, --event-external-id, --position, --recurring-event-external-id
    calendar-channel-event-associations update - Update a Twenty calendar channel event associations record [intent=reverse_etl availability=implemented write=update_calendar_channel_event_associations]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Updates an existing Twenty calendar channel event associations record through the REST API.; flags: --id, --calendar-channel-id, --event-external-id, --position, --recurring-event-external-id
    calendar-channel-event-associations batch - Batch-create Twenty calendar channel event associations records [intent=reverse_etl availability=implemented write=batch_calendar_channel_event_associations]; approval: reverse ETL batch writes require plan, preview, approval, execute.; risk: Bulk creates Twenty calendar channel event associations records through the REST API.; flags: --records (required)
    calendar-channel-event-associations delete - Delete a Twenty calendar channel event associations record [intent=reverse_etl availability=implemented write=delete_calendar_channel_event_associations]; approval: reverse ETL deletes require plan, preview, approval, execute, and typed --confirm destructive.; risk: Destructively deletes a Twenty calendar channel event associations record through the REST API.; flags: --id
    calendar-event-participants list - List Twenty calendar event participants records [intent=etl availability=implemented stream=calendar_event_participants]; flags: --limit
    calendar-event-participants get - Get one Twenty calendar event participants record by ID [intent=direct_read availability=implemented operation=twenty.calendar-event-participants.get]; flags: --id (required), --page, --page-cursor
    calendar-event-participants create - Create a Twenty calendar event participants record [intent=reverse_etl availability=implemented write=create_calendar_event_participants]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a Twenty calendar event participants record through the REST API.; flags: --display-name, --handle, --is-organizer, --position, --response-status
    calendar-event-participants update - Update a Twenty calendar event participants record [intent=reverse_etl availability=implemented write=update_calendar_event_participants]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Updates an existing Twenty calendar event participants record through the REST API.; flags: --id, --display-name, --handle, --is-organizer, --position, --response-status
    calendar-event-participants batch - Batch-create Twenty calendar event participants records [intent=reverse_etl availability=implemented write=batch_calendar_event_participants]; approval: reverse ETL batch writes require plan, preview, approval, execute.; risk: Bulk creates Twenty calendar event participants records through the REST API.; flags: --records (required)
    calendar-event-participants delete - Delete a Twenty calendar event participants record [intent=reverse_etl availability=implemented write=delete_calendar_event_participants]; approval: reverse ETL deletes require plan, preview, approval, execute, and typed --confirm destructive.; risk: Destructively deletes a Twenty calendar event participants record through the REST API.; flags: --id
    calendar-events list - List Twenty calendar events records [intent=etl availability=implemented stream=calendar_events]; flags: --limit
    calendar-events get - Get one Twenty calendar events record by ID [intent=direct_read availability=implemented operation=twenty.calendar-events.get]; flags: --id (required), --page, --page-cursor
    calendar-events create - Create a Twenty calendar events record [intent=reverse_etl availability=implemented write=create_calendar_events]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a Twenty calendar events record through the REST API.; flags: --call-recorder-preference, --conference-solution, --description, --ends-at, --external-created-at, --external-updated-at, --i-cal-uid, --is-canceled, --is-full-day, --location, --position, --starts-at, --title
    calendar-events update - Update a Twenty calendar events record [intent=reverse_etl availability=implemented write=update_calendar_events]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Updates an existing Twenty calendar events record through the REST API.; flags: --id, --call-recorder-preference, --conference-solution, --description, --ends-at, --external-created-at, --external-updated-at, --i-cal-uid, --is-canceled, --is-full-day, --location, --position, --starts-at, --title
    calendar-events batch - Batch-create Twenty calendar events records [intent=reverse_etl availability=implemented write=batch_calendar_events]; approval: reverse ETL batch writes require plan, preview, approval, execute.; risk: Bulk creates Twenty calendar events records through the REST API.; flags: --records (required)
    calendar-events delete - Delete a Twenty calendar events record [intent=reverse_etl availability=implemented write=delete_calendar_events]; approval: reverse ETL deletes require plan, preview, approval, execute, and typed --confirm destructive.; risk: Destructively deletes a Twenty calendar events record through the REST API.; flags: --id
    call-recordings list - List Twenty call recordings records [intent=etl availability=implemented stream=call_recordings]; flags: --limit
    call-recordings get - Get one Twenty call recordings record by ID [intent=direct_read availability=implemented operation=twenty.call-recordings.get]; flags: --id (required), --page, --page-cursor
    call-recordings create - Create a Twenty call recordings record [intent=reverse_etl availability=implemented write=create_call_recordings]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a Twenty call recordings record through the REST API.; flags: --application-id, --call-recorder-failure-reason, --ended-at, --external-bot-id, --external-recording-id, --position, --recording-request-status, --started-at, --status, --summary, --title
    call-recordings update - Update a Twenty call recordings record [intent=reverse_etl availability=implemented write=update_call_recordings]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Updates an existing Twenty call recordings record through the REST API.; flags: --id, --application-id, --call-recorder-failure-reason, --ended-at, --external-bot-id, --external-recording-id, --position, --recording-request-status, --started-at, --status, --summary, --title
    call-recordings batch - Batch-create Twenty call recordings records [intent=reverse_etl availability=implemented write=batch_call_recordings]; approval: reverse ETL batch writes require plan, preview, approval, execute.; risk: Bulk creates Twenty call recordings records through the REST API.; flags: --records (required)
    call-recordings delete - Delete a Twenty call recordings record [intent=reverse_etl availability=implemented write=delete_call_recordings]; approval: reverse ETL deletes require plan, preview, approval, execute, and typed --confirm destructive.; risk: Destructively deletes a Twenty call recordings record through the REST API.; flags: --id
    companies list - List Twenty companies records [intent=etl availability=implemented stream=companies]; flags: --limit
    companies get - Get one Twenty companies record by ID [intent=direct_read availability=implemented operation=twenty.companies.get]; flags: --id (required), --page, --page-cursor
    companies create - Create a Twenty companies record [intent=reverse_etl availability=implemented write=create_companies]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a Twenty companies record through the REST API.; flags: --last-contact-at, --name, --pdl-company-type, --pdl-employee-count, --pdl-enrichment-status, --pdl-founded-year, --pdl-funding-stages, --pdl-headline, --pdl-id, --pdl-industry, --pdl-industry-detail, --pdl-last-enriched-at, --pdl-last-funding-date, --pdl-latest-funding-stage, --pdl-legal-name, --pdl-linkedin-id, --pdl-location-continent, --pdl-location-metro, --pdl-mic-exchange, --pdl-number-funding-rounds, --pdl-size-range, --pdl-summary, --pdl-ticker, --position
    companies update - Update a Twenty companies record [intent=reverse_etl availability=implemented write=update_companies]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Updates an existing Twenty companies record through the REST API.; flags: --id, --last-contact-at, --name, --pdl-company-type, --pdl-employee-count, --pdl-enrichment-status, --pdl-founded-year, --pdl-funding-stages, --pdl-headline, --pdl-id, --pdl-industry, --pdl-industry-detail, --pdl-last-enriched-at, --pdl-last-funding-date, --pdl-latest-funding-stage, --pdl-legal-name, --pdl-linkedin-id, --pdl-location-continent, --pdl-location-metro, --pdl-mic-exchange, --pdl-number-funding-rounds, --pdl-size-range, --pdl-summary, --pdl-ticker, --position
    companies batch - Batch-create Twenty companies records [intent=reverse_etl availability=implemented write=batch_companies]; approval: reverse ETL batch writes require plan, preview, approval, execute.; risk: Bulk creates Twenty companies records through the REST API.; flags: --records (required)
    companies delete - Delete a Twenty companies record [intent=reverse_etl availability=implemented write=delete_companies]; approval: reverse ETL deletes require plan, preview, approval, execute, and typed --confirm destructive.; risk: Destructively deletes a Twenty companies record through the REST API.; flags: --id
    dashboards list - List Twenty dashboards records [intent=etl availability=implemented stream=dashboards]; flags: --limit
    dashboards get - Get one Twenty dashboards record by ID [intent=direct_read availability=implemented operation=twenty.dashboards.get]; flags: --id (required), --page, --page-cursor
    dashboards create - Create a Twenty dashboards record [intent=reverse_etl availability=implemented write=create_dashboards]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a Twenty dashboards record through the REST API.; flags: --page-layout-id, --position, --title
    dashboards update - Update a Twenty dashboards record [intent=reverse_etl availability=implemented write=update_dashboards]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Updates an existing Twenty dashboards record through the REST API.; flags: --id, --page-layout-id, --position, --title
    dashboards batch - Batch-create Twenty dashboards records [intent=reverse_etl availability=implemented write=batch_dashboards]; approval: reverse ETL batch writes require plan, preview, approval, execute.; risk: Bulk creates Twenty dashboards records through the REST API.; flags: --records (required)
    dashboards delete - Delete a Twenty dashboards record [intent=reverse_etl availability=implemented write=delete_dashboards]; approval: reverse ETL deletes require plan, preview, approval, execute, and typed --confirm destructive.; risk: Destructively deletes a Twenty dashboards record through the REST API.; flags: --id
    message-campaigns list - List Twenty message campaigns records [intent=etl availability=implemented stream=message_campaigns]; flags: --limit
    message-campaigns get - Get one Twenty message campaigns record by ID [intent=direct_read availability=implemented operation=twenty.message-campaigns.get]; flags: --id (required), --page, --page-cursor
    message-campaigns create - Create a Twenty message campaigns record [intent=reverse_etl availability=implemented write=create_message_campaigns]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a Twenty message campaigns record through the REST API.; flags: --body-template, --bounced-count, --complained-count, --failed-count, --position, --sent-at, --sent-count, --status, --subject, --unsubscribe-topic-id
    message-campaigns update - Update a Twenty message campaigns record [intent=reverse_etl availability=implemented write=update_message_campaigns]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Updates an existing Twenty message campaigns record through the REST API.; flags: --id, --body-template, --bounced-count, --complained-count, --failed-count, --position, --sent-at, --sent-count, --status, --subject, --unsubscribe-topic-id
    message-campaigns batch - Batch-create Twenty message campaigns records [intent=reverse_etl availability=implemented write=batch_message_campaigns]; approval: reverse ETL batch writes require plan, preview, approval, execute.; risk: Bulk creates Twenty message campaigns records through the REST API.; flags: --records (required)
    message-campaigns delete - Delete a Twenty message campaigns record [intent=reverse_etl availability=implemented write=delete_message_campaigns]; approval: reverse ETL deletes require plan, preview, approval, execute, and typed --confirm destructive.; risk: Destructively deletes a Twenty message campaigns record through the REST API.; flags: --id
    message-channel-message-association-message-folders list - List Twenty message channel message association message folders records [intent=etl availability=implemented stream=message_channel_message_association_message_folders]; flags: --limit
    message-channel-message-association-message-folders get - Get one Twenty message channel message association message folders record by ID [intent=direct_read availability=implemented operation=twenty.message-channel-message-association-message-folders.get]; flags: --id (required), --page, --page-cursor
    message-channel-message-association-message-folders create - Create a Twenty message channel message association message folders record [intent=reverse_etl availability=implemented write=create_message_channel_message_association_message_folders]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a Twenty message channel message association message folders record through the REST API.; flags: --message-folder-id, --position
    message-channel-message-association-message-folders update - Update a Twenty message channel message association message folders record [intent=reverse_etl availability=implemented write=update_message_channel_message_association_message_folders]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Updates an existing Twenty message channel message association message folders record through the REST API.; flags: --id, --message-folder-id, --position
    message-channel-message-association-message-folders batch - Batch-create Twenty message channel message association message folders records [intent=reverse_etl availability=implemented write=batch_message_channel_message_association_message_folders]; approval: reverse ETL batch writes require plan, preview, approval, execute.; risk: Bulk creates Twenty message channel message association message folders records through the REST API.; flags: --records (required)
    message-channel-message-association-message-folders delete - Delete a Twenty message channel message association message folders record [intent=reverse_etl availability=implemented write=delete_message_channel_message_association_message_folders]; approval: reverse ETL deletes require plan, preview, approval, execute, and typed --confirm destructive.; risk: Destructively deletes a Twenty message channel message association message folders record through the REST API.; flags: --id
    message-channel-message-associations list - List Twenty message channel message associations records [intent=etl availability=implemented stream=message_channel_message_associations]; flags: --limit
    message-channel-message-associations get - Get one Twenty message channel message associations record by ID [intent=direct_read availability=implemented operation=twenty.message-channel-message-associations.get]; flags: --id (required), --page, --page-cursor
    message-channel-message-associations create - Create a Twenty message channel message associations record [intent=reverse_etl availability=implemented write=create_message_channel_message_associations]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a Twenty message channel message associations record through the REST API.; flags: --direction, --message-channel-id, --message-external-id, --message-thread-external-id, --position
    message-channel-message-associations update - Update a Twenty message channel message associations record [intent=reverse_etl availability=implemented write=update_message_channel_message_associations]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Updates an existing Twenty message channel message associations record through the REST API.; flags: --id, --direction, --message-channel-id, --message-external-id, --message-thread-external-id, --position
    message-channel-message-associations batch - Batch-create Twenty message channel message associations records [intent=reverse_etl availability=implemented write=batch_message_channel_message_associations]; approval: reverse ETL batch writes require plan, preview, approval, execute.; risk: Bulk creates Twenty message channel message associations records through the REST API.; flags: --records (required)
    message-channel-message-associations delete - Delete a Twenty message channel message associations record [intent=reverse_etl availability=implemented write=delete_message_channel_message_associations]; approval: reverse ETL deletes require plan, preview, approval, execute, and typed --confirm destructive.; risk: Destructively deletes a Twenty message channel message associations record through the REST API.; flags: --id
    message-list-members list - List Twenty message list members records [intent=etl availability=implemented stream=message_list_members]; flags: --limit
    message-list-members get - Get one Twenty message list members record by ID [intent=direct_read availability=implemented operation=twenty.message-list-members.get]; flags: --id (required), --page, --page-cursor
    message-list-members create - Create a Twenty message list members record [intent=reverse_etl availability=implemented write=create_message_list_members]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a Twenty message list members record through the REST API.; flags: --position
    message-list-members update - Update a Twenty message list members record [intent=reverse_etl availability=implemented write=update_message_list_members]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Updates an existing Twenty message list members record through the REST API.; flags: --id, --position
    message-list-members batch - Batch-create Twenty message list members records [intent=reverse_etl availability=implemented write=batch_message_list_members]; approval: reverse ETL batch writes require plan, preview, approval, execute.; risk: Bulk creates Twenty message list members records through the REST API.; flags: --records (required)
    message-list-members delete - Delete a Twenty message list members record [intent=reverse_etl availability=implemented write=delete_message_list_members]; approval: reverse ETL deletes require plan, preview, approval, execute, and typed --confirm destructive.; risk: Destructively deletes a Twenty message list members record through the REST API.; flags: --id
    message-lists list - List Twenty message lists records [intent=etl availability=implemented stream=message_lists]; flags: --limit
    message-lists get - Get one Twenty message lists record by ID [intent=direct_read availability=implemented operation=twenty.message-lists.get]; flags: --id (required), --page, --page-cursor
    message-lists create - Create a Twenty message lists record [intent=reverse_etl availability=implemented write=create_message_lists]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a Twenty message lists record through the REST API.; flags: --name, --position
    message-lists update - Update a Twenty message lists record [intent=reverse_etl availability=implemented write=update_message_lists]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Updates an existing Twenty message lists record through the REST API.; flags: --id, --name, --position
    message-lists batch - Batch-create Twenty message lists records [intent=reverse_etl availability=implemented write=batch_message_lists]; approval: reverse ETL batch writes require plan, preview, approval, execute.; risk: Bulk creates Twenty message lists records through the REST API.; flags: --records (required)
    message-lists delete - Delete a Twenty message lists record [intent=reverse_etl availability=implemented write=delete_message_lists]; approval: reverse ETL deletes require plan, preview, approval, execute, and typed --confirm destructive.; risk: Destructively deletes a Twenty message lists record through the REST API.; flags: --id
    message-participants list - List Twenty message participants records [intent=etl availability=implemented stream=message_participants]; flags: --limit
    message-participants get - Get one Twenty message participants record by ID [intent=direct_read availability=implemented operation=twenty.message-participants.get]; flags: --id (required), --page, --page-cursor
    message-participants create - Create a Twenty message participants record [intent=reverse_etl availability=implemented write=create_message_participants]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a Twenty message participants record through the REST API.; flags: --display-name, --handle, --position, --role
    message-participants update - Update a Twenty message participants record [intent=reverse_etl availability=implemented write=update_message_participants]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Updates an existing Twenty message participants record through the REST API.; flags: --id, --display-name, --handle, --position, --role
    message-participants batch - Batch-create Twenty message participants records [intent=reverse_etl availability=implemented write=batch_message_participants]; approval: reverse ETL batch writes require plan, preview, approval, execute.; risk: Bulk creates Twenty message participants records through the REST API.; flags: --records (required)
    message-participants delete - Delete a Twenty message participants record [intent=reverse_etl availability=implemented write=delete_message_participants]; approval: reverse ETL deletes require plan, preview, approval, execute, and typed --confirm destructive.; risk: Destructively deletes a Twenty message participants record through the REST API.; flags: --id
    message-threads list - List Twenty message threads records [intent=etl availability=implemented stream=message_threads]; flags: --limit
    message-threads get - Get one Twenty message threads record by ID [intent=direct_read availability=implemented operation=twenty.message-threads.get]; flags: --id (required), --page, --page-cursor
    message-threads create - Create a Twenty message threads record [intent=reverse_etl availability=implemented write=create_message_threads]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a Twenty message threads record through the REST API.; flags: --position, --subject
    message-threads update - Update a Twenty message threads record [intent=reverse_etl availability=implemented write=update_message_threads]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Updates an existing Twenty message threads record through the REST API.; flags: --id, --position, --subject
    message-threads batch - Batch-create Twenty message threads records [intent=reverse_etl availability=implemented write=batch_message_threads]; approval: reverse ETL batch writes require plan, preview, approval, execute.; risk: Bulk creates Twenty message threads records through the REST API.; flags: --records (required)
    message-threads delete - Delete a Twenty message threads record [intent=reverse_etl availability=implemented write=delete_message_threads]; approval: reverse ETL deletes require plan, preview, approval, execute, and typed --confirm destructive.; risk: Destructively deletes a Twenty message threads record through the REST API.; flags: --id
    messages list - List Twenty messages records [intent=etl availability=implemented stream=messages]; flags: --limit
    messages get - Get one Twenty messages record by ID [intent=direct_read availability=implemented operation=twenty.messages.get]; flags: --id (required), --page, --page-cursor
    messages create - Create a Twenty messages record [intent=reverse_etl availability=implemented write=create_messages]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a Twenty messages record through the REST API.; flags: --delivery-status, --header-message-id, --is-draft, --position, --received-at, --subject, --text
    messages update - Update a Twenty messages record [intent=reverse_etl availability=implemented write=update_messages]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Updates an existing Twenty messages record through the REST API.; flags: --id, --delivery-status, --header-message-id, --is-draft, --position, --received-at, --subject, --text
    messages batch - Batch-create Twenty messages records [intent=reverse_etl availability=implemented write=batch_messages]; approval: reverse ETL batch writes require plan, preview, approval, execute.; risk: Bulk creates Twenty messages records through the REST API.; flags: --records (required)
    messages delete - Delete a Twenty messages record [intent=reverse_etl availability=implemented write=delete_messages]; approval: reverse ETL deletes require plan, preview, approval, execute, and typed --confirm destructive.; risk: Destructively deletes a Twenty messages record through the REST API.; flags: --id
    note-targets list - List Twenty note targets records [intent=etl availability=implemented stream=note_targets]; flags: --limit
    note-targets get - Get one Twenty note targets record by ID [intent=direct_read availability=implemented operation=twenty.note-targets.get]; flags: --id (required), --page, --page-cursor
    note-targets create - Create a Twenty note targets record [intent=reverse_etl availability=implemented write=create_note_targets]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a Twenty note targets record through the REST API.; flags: --position
    note-targets update - Update a Twenty note targets record [intent=reverse_etl availability=implemented write=update_note_targets]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Updates an existing Twenty note targets record through the REST API.; flags: --id, --position
    note-targets batch - Batch-create Twenty note targets records [intent=reverse_etl availability=implemented write=batch_note_targets]; approval: reverse ETL batch writes require plan, preview, approval, execute.; risk: Bulk creates Twenty note targets records through the REST API.; flags: --records (required)
    note-targets delete - Delete a Twenty note targets record [intent=reverse_etl availability=implemented write=delete_note_targets]; approval: reverse ETL deletes require plan, preview, approval, execute, and typed --confirm destructive.; risk: Destructively deletes a Twenty note targets record through the REST API.; flags: --id
    notes list - List Twenty notes records [intent=etl availability=implemented stream=notes]; flags: --limit
    notes get - Get one Twenty notes record by ID [intent=direct_read availability=implemented operation=twenty.notes.get]; flags: --id (required), --page, --page-cursor
    notes create - Create a Twenty notes record [intent=reverse_etl availability=implemented write=create_notes]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a Twenty notes record through the REST API.; flags: --body-v2, --position, --title
    notes update - Update a Twenty notes record [intent=reverse_etl availability=implemented write=update_notes]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Updates an existing Twenty notes record through the REST API.; flags: --id, --body-v2, --position, --title
    notes batch - Batch-create Twenty notes records [intent=reverse_etl availability=implemented write=batch_notes]; approval: reverse ETL batch writes require plan, preview, approval, execute.; risk: Bulk creates Twenty notes records through the REST API.; flags: --records (required)
    notes delete - Delete a Twenty notes record [intent=reverse_etl availability=implemented write=delete_notes]; approval: reverse ETL deletes require plan, preview, approval, execute, and typed --confirm destructive.; risk: Destructively deletes a Twenty notes record through the REST API.; flags: --id
    opportunities list - List Twenty opportunities records [intent=etl availability=implemented stream=opportunities]; flags: --limit
    opportunities get - Get one Twenty opportunities record by ID [intent=direct_read availability=implemented operation=twenty.opportunities.get]; flags: --id (required), --page, --page-cursor
    opportunities create - Create a Twenty opportunities record [intent=reverse_etl availability=implemented write=create_opportunities]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a Twenty opportunities record through the REST API.; flags: --close-date, --last-contact-at, --name, --position, --stage
    opportunities update - Update a Twenty opportunities record [intent=reverse_etl availability=implemented write=update_opportunities]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Updates an existing Twenty opportunities record through the REST API.; flags: --id, --close-date, --last-contact-at, --name, --position, --stage
    opportunities batch - Batch-create Twenty opportunities records [intent=reverse_etl availability=implemented write=batch_opportunities]; approval: reverse ETL batch writes require plan, preview, approval, execute.; risk: Bulk creates Twenty opportunities records through the REST API.; flags: --records (required)
    opportunities delete - Delete a Twenty opportunities record [intent=reverse_etl availability=implemented write=delete_opportunities]; approval: reverse ETL deletes require plan, preview, approval, execute, and typed --confirm destructive.; risk: Destructively deletes a Twenty opportunities record through the REST API.; flags: --id
    people list - List Twenty people records [intent=etl availability=implemented stream=people]; flags: --limit
    people get - Get one Twenty people record by ID [intent=direct_read availability=implemented operation=twenty.people.get]; flags: --id (required), --page, --page-cursor
    people create - Create a Twenty people record [intent=reverse_etl availability=implemented write=create_people]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a Twenty people record through the REST API.; flags: --avatar-url, --job-title, --last-contact-at, --last-inbound-at, --last-outbound-at, --pdl-birth-date, --pdl-birth-year, --pdl-enrichment-status, --pdl-headline, --pdl-id, --pdl-industry, --pdl-inferred-salary, --pdl-job-onet-code, --pdl-job-role, --pdl-job-start-date, --pdl-job-summary, --pdl-job-title-class, --pdl-job-title-sub-role, --pdl-last-enriched-at, --pdl-likelihood, --pdl-linkedin-connections, --pdl-linkedin-username, --pdl-location-metro, --pdl-seniority, --pdl-sex, --pdl-summary, --pdl-years-experience, --position
    people update - Update a Twenty people record [intent=reverse_etl availability=implemented write=update_people]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Updates an existing Twenty people record through the REST API.; flags: --id, --avatar-url, --job-title, --last-contact-at, --last-inbound-at, --last-outbound-at, --pdl-birth-date, --pdl-birth-year, --pdl-enrichment-status, --pdl-headline, --pdl-id, --pdl-industry, --pdl-inferred-salary, --pdl-job-onet-code, --pdl-job-role, --pdl-job-start-date, --pdl-job-summary, --pdl-job-title-class, --pdl-job-title-sub-role, --pdl-last-enriched-at, --pdl-likelihood, --pdl-linkedin-connections, --pdl-linkedin-username, --pdl-location-metro, --pdl-seniority, --pdl-sex, --pdl-summary, --pdl-years-experience, --position
    people batch - Batch-create Twenty people records [intent=reverse_etl availability=implemented write=batch_people]; approval: reverse ETL batch writes require plan, preview, approval, execute.; risk: Bulk creates Twenty people records through the REST API.; flags: --records (required)
    people delete - Delete a Twenty people record [intent=reverse_etl availability=implemented write=delete_people]; approval: reverse ETL deletes require plan, preview, approval, execute, and typed --confirm destructive.; risk: Destructively deletes a Twenty people record through the REST API.; flags: --id
    task-targets list - List Twenty task targets records [intent=etl availability=implemented stream=task_targets]; flags: --limit
    task-targets get - Get one Twenty task targets record by ID [intent=direct_read availability=implemented operation=twenty.task-targets.get]; flags: --id (required), --page, --page-cursor
    task-targets create - Create a Twenty task targets record [intent=reverse_etl availability=implemented write=create_task_targets]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a Twenty task targets record through the REST API.; flags: --position
    task-targets update - Update a Twenty task targets record [intent=reverse_etl availability=implemented write=update_task_targets]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Updates an existing Twenty task targets record through the REST API.; flags: --id, --position
    task-targets batch - Batch-create Twenty task targets records [intent=reverse_etl availability=implemented write=batch_task_targets]; approval: reverse ETL batch writes require plan, preview, approval, execute.; risk: Bulk creates Twenty task targets records through the REST API.; flags: --records (required)
    task-targets delete - Delete a Twenty task targets record [intent=reverse_etl availability=implemented write=delete_task_targets]; approval: reverse ETL deletes require plan, preview, approval, execute, and typed --confirm destructive.; risk: Destructively deletes a Twenty task targets record through the REST API.; flags: --id
    tasks list - List Twenty tasks records [intent=etl availability=implemented stream=tasks]; flags: --limit
    tasks get - Get one Twenty tasks record by ID [intent=direct_read availability=implemented operation=twenty.tasks.get]; flags: --id (required), --page, --page-cursor
    tasks create - Create a Twenty tasks record [intent=reverse_etl availability=implemented write=create_tasks]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a Twenty tasks record through the REST API.; flags: --body-v2, --due-at, --position, --status, --title
    tasks update - Update a Twenty tasks record [intent=reverse_etl availability=implemented write=update_tasks]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Updates an existing Twenty tasks record through the REST API.; flags: --id, --body-v2, --due-at, --position, --status, --title
    tasks batch - Batch-create Twenty tasks records [intent=reverse_etl availability=implemented write=batch_tasks]; approval: reverse ETL batch writes require plan, preview, approval, execute.; risk: Bulk creates Twenty tasks records through the REST API.; flags: --records (required)
    tasks delete - Delete a Twenty tasks record [intent=reverse_etl availability=implemented write=delete_tasks]; approval: reverse ETL deletes require plan, preview, approval, execute, and typed --confirm destructive.; risk: Destructively deletes a Twenty tasks record through the REST API.; flags: --id
    timeline-activities list - List Twenty timeline activities records [intent=etl availability=implemented stream=timeline_activities]; flags: --limit
    timeline-activities get - Get one Twenty timeline activities record by ID [intent=direct_read availability=implemented operation=twenty.timeline-activities.get]; flags: --id (required), --page, --page-cursor
    timeline-activities create - Create a Twenty timeline activities record [intent=reverse_etl availability=implemented write=create_timeline_activities]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a Twenty timeline activities record through the REST API.; flags: --happens-at, --name, --position
    timeline-activities update - Update a Twenty timeline activities record [intent=reverse_etl availability=implemented write=update_timeline_activities]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Updates an existing Twenty timeline activities record through the REST API.; flags: --id, --happens-at, --name, --position
    timeline-activities batch - Batch-create Twenty timeline activities records [intent=reverse_etl availability=implemented write=batch_timeline_activities]; approval: reverse ETL batch writes require plan, preview, approval, execute.; risk: Bulk creates Twenty timeline activities records through the REST API.; flags: --records (required)
    timeline-activities delete - Delete a Twenty timeline activities record [intent=reverse_etl availability=implemented write=delete_timeline_activities]; approval: reverse ETL deletes require plan, preview, approval, execute, and typed --confirm destructive.; risk: Destructively deletes a Twenty timeline activities record through the REST API.; flags: --id
    workflow-automated-triggers list - List Twenty workflow automated triggers records [intent=etl availability=implemented stream=workflow_automated_triggers]; flags: --limit
    workflow-automated-triggers get - Get one Twenty workflow automated triggers record by ID [intent=direct_read availability=implemented operation=twenty.workflow-automated-triggers.get]; flags: --id (required), --page, --page-cursor
    workflow-automated-triggers create - Create a Twenty workflow automated triggers record [intent=reverse_etl availability=implemented write=create_workflow_automated_triggers]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a Twenty workflow automated triggers record through the REST API.; flags: --position, --type
    workflow-automated-triggers update - Update a Twenty workflow automated triggers record [intent=reverse_etl availability=implemented write=update_workflow_automated_triggers]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Updates an existing Twenty workflow automated triggers record through the REST API.; flags: --id, --position, --type
    workflow-automated-triggers batch - Batch-create Twenty workflow automated triggers records [intent=reverse_etl availability=implemented write=batch_workflow_automated_triggers]; approval: reverse ETL batch writes require plan, preview, approval, execute.; risk: Bulk creates Twenty workflow automated triggers records through the REST API.; flags: --records (required)
    workflow-automated-triggers delete - Delete a Twenty workflow automated triggers record [intent=reverse_etl availability=implemented write=delete_workflow_automated_triggers]; approval: reverse ETL deletes require plan, preview, approval, execute, and typed --confirm destructive.; risk: Destructively deletes a Twenty workflow automated triggers record through the REST API.; flags: --id
    workflow-runs list - List Twenty workflow runs records [intent=etl availability=implemented stream=workflow_runs]; flags: --limit
    workflow-runs get - Get one Twenty workflow runs record by ID [intent=direct_read availability=implemented operation=twenty.workflow-runs.get]; flags: --id (required), --page, --page-cursor
    workflow-runs create - Create a Twenty workflow runs record [intent=reverse_etl availability=implemented write=create_workflow_runs]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a Twenty workflow runs record through the REST API.; flags: --ended-at, --enqueued-at, --name, --started-at, --status
    workflow-runs update - Update a Twenty workflow runs record [intent=reverse_etl availability=implemented write=update_workflow_runs]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Updates an existing Twenty workflow runs record through the REST API.; flags: --id, --ended-at, --enqueued-at, --name, --started-at, --status
    workflow-runs batch - Batch-create Twenty workflow runs records [intent=reverse_etl availability=implemented write=batch_workflow_runs]; approval: reverse ETL batch writes require plan, preview, approval, execute.; risk: Bulk creates Twenty workflow runs records through the REST API.; flags: --records (required)
    workflow-runs delete - Delete a Twenty workflow runs record [intent=reverse_etl availability=implemented write=delete_workflow_runs]; approval: reverse ETL deletes require plan, preview, approval, execute, and typed --confirm destructive.; risk: Destructively deletes a Twenty workflow runs record through the REST API.; flags: --id
    workflow-versions list - List Twenty workflow versions records [intent=etl availability=implemented stream=workflow_versions]; flags: --limit
    workflow-versions get - Get one Twenty workflow versions record by ID [intent=direct_read availability=implemented operation=twenty.workflow-versions.get]; flags: --id (required), --page, --page-cursor
    workflow-versions create - Create a Twenty workflow versions record [intent=reverse_etl availability=implemented write=create_workflow_versions]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a Twenty workflow versions record through the REST API.; flags: --name, --status
    workflow-versions update - Update a Twenty workflow versions record [intent=reverse_etl availability=implemented write=update_workflow_versions]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Updates an existing Twenty workflow versions record through the REST API.; flags: --id, --name, --status
    workflow-versions batch - Batch-create Twenty workflow versions records [intent=reverse_etl availability=implemented write=batch_workflow_versions]; approval: reverse ETL batch writes require plan, preview, approval, execute.; risk: Bulk creates Twenty workflow versions records through the REST API.; flags: --records (required)
    workflow-versions delete - Delete a Twenty workflow versions record [intent=reverse_etl availability=implemented write=delete_workflow_versions]; approval: reverse ETL deletes require plan, preview, approval, execute, and typed --confirm destructive.; risk: Destructively deletes a Twenty workflow versions record through the REST API.; flags: --id
    workflows list - List Twenty workflows records [intent=etl availability=implemented stream=workflows]; flags: --limit
    workflows get - Get one Twenty workflows record by ID [intent=direct_read availability=implemented operation=twenty.workflows.get]; flags: --id (required), --page, --page-cursor
    workflows create - Create a Twenty workflows record [intent=reverse_etl availability=implemented write=create_workflows]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a Twenty workflows record through the REST API.; flags: --last-published-version-id, --name, --position, --statuses
    workflows update - Update a Twenty workflows record [intent=reverse_etl availability=implemented write=update_workflows]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Updates an existing Twenty workflows record through the REST API.; flags: --id, --last-published-version-id, --name, --position, --statuses
    workflows batch - Batch-create Twenty workflows records [intent=reverse_etl availability=implemented write=batch_workflows]; approval: reverse ETL batch writes require plan, preview, approval, execute.; risk: Bulk creates Twenty workflows records through the REST API.; flags: --records (required)
    workflows delete - Delete a Twenty workflows record [intent=reverse_etl availability=implemented write=delete_workflows]; approval: reverse ETL deletes require plan, preview, approval, execute, and typed --confirm destructive.; risk: Destructively deletes a Twenty workflows record through the REST API.; flags: --id
    workspace-members list - List Twenty workspace members records [intent=etl availability=implemented stream=workspace_members]; flags: --limit
    workspace-members get - Get one Twenty workspace members record by ID [intent=direct_read availability=implemented operation=twenty.workspace-members.get]; flags: --id (required), --page, --page-cursor
    workspace-members create - Create a Twenty workspace members record [intent=reverse_etl availability=implemented write=create_workspace_members]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a Twenty workspace members record through the REST API.; notes: Write schema exposes nested object/array fields only; CLI scalar flag coercion is not exposed for this command, so use reverse ETL records rather than direct scalar flags.
    workspace-members update - Update a Twenty workspace members record [intent=reverse_etl availability=implemented write=update_workspace_members]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Updates an existing Twenty workspace members record through the REST API.; notes: Write schema exposes nested object/array fields only; CLI scalar flag coercion is not exposed for this command, so use reverse ETL records rather than direct scalar flags.; flags: --id
    workspace-members batch - Batch-create Twenty workspace members records [intent=reverse_etl availability=implemented write=batch_workspace_members]; approval: reverse ETL batch writes require plan, preview, approval, execute.; risk: Bulk creates Twenty workspace members records through the REST API.; flags: --records (required)
    workspace-members delete - Delete a Twenty workspace members record [intent=reverse_etl availability=implemented write=delete_workspace_members]; approval: reverse ETL deletes require plan, preview, approval, execute, and typed --confirm destructive.; risk: Destructively deletes a Twenty workspace members record through the REST API.; flags: --id
  Help topics:
    writes - Twenty write commands use reverse ETL plan, preview, approval, execute.
    direct-read - Get-by-id commands are bounded, operation-backed direct reads with redacted JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect twenty

  # Inspect as structured JSON
  pm connectors inspect twenty --json

AGENT WORKFLOW
  - Run pm connectors inspect twenty before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
