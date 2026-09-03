# pm connectors inspect ashby

```text
NAME
  pm connectors inspect ashby - Ashby connector manual

SYNOPSIS
  pm connectors inspect ashby
  pm connectors inspect ashby --json
  pm credentials add <name> --connector ashby [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Ashby applicant-tracking REST resources and exposes reviewed reverse-ETL/direct-read surfaces from the official Ashby OpenAPI. Fixture-only; not live-certified.

ICON
  id: ashby
  asset: icons/ashby.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.ashbyhq.com/

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  api_key (secret) (required)

ETL STREAMS
  candidates:
    primary key: id
    fields: applicationIds(array), company(string), createdAt(string), creditedToUser(object), customFields(array), emailAddresses(array), fileHandles(array), fraudStatus(string), id(string), location(object), name(string), phoneNumbers(array), position(string), primaryEmailAddress(object), primaryPhoneNumber(object), profileUrl(string), resumeFileHandle(object), school(string), socialLinks(array), source(object), tags(array), timezone(string), updatedAt(string)
  jobs:
    primary key: id
    fields: author(object), brandId(string), closedAt(string), compensation(object), confidential(boolean), createdAt(string), customFields(array), customRequisitionId(string), defaultInterviewPlanId(string), departmentId(string), employmentType(string), hiringTeam(array), id(string), interviewPlanIds(array), jobPostingIds(array), location(object), locationId(string), openedAt(string), openings(array), status(string), title(string), updatedAt(string)
  applications:
    primary key: id
    fields: appliedViaJobPostingId(string), archiveReason(object), archivedAt(string), candidate(object), createdAt(string), creditedToUser(object), currentInterviewStage(object), customFields(array), hiringTeam(array), id(string), job(object), openings(array), source(object), status(string), submitterClientIp(string), submitterUserAgent(string), updatedAt(string)
  users:
    primary key: id
    fields: customFields(array), email(string), firstName(string), globalRole(string), id(string), isEnabled(boolean), lastName(string), managerId(string), updatedAt(string)
  api_key_info:
    primary key: title
    fields: createdAt(string), scopes(array), title(string)
  audit_log_list:
    primary key: id
    fields: actor(object), category(string), changedFields(object), description(string), id(string), parentAction(object), request(object), target(object), timestamp(string)
  application_info:
    primary key: id
    fields: applicationFormSubmissions(array), applicationHistory(array), appliedViaJobPostingId(string), archiveReason(object), archivedAt(string), candidate(object), createdAt(string), creditedToUser(object), currentInterviewStage(object), customFields(array), hiringTeam(array), id(string), job(object), openings(array), referrals(array), resumeFileHandle(object), source(object), status(string), submitterClientIp(string), submitterUserAgent(string), updatedAt(string)
  application_hiring_team_role_list:
    primary key: id
    fields: id(string), title(string)
  application_list_history:
    primary key: id
    fields: actorId(string), allowedActions(array), enteredStageAt(string), id(string), leftStageAt(string), stageId(string), stageNumber(integer), title(string)
  application_list_criteria_evaluations:
    primary key: id
    fields: criterion(object), evaluatedAt(string), id(string), outcome(string), outcomeNumber(number), reasoning(string), skipReason(string), status(string)
  application_feedback_list:
    primary key: id
    fields: applicationHistoryId(string), applicationId(string), creditedToUser(object), feedbackFormDefinitionId(string), formDefinition(object), id(string), interviewEventId(string), interviewId(string), submittedAt(string), submittedByUser(object), submittedValues(object)
  candidate_info:
    primary key: id
    fields: applicationIds(array), company(string), createdAt(string), creditedToUser(object), customFields(array), emailAddresses(array), fileHandles(array), fraudStatus(string), id(string), location(object), name(string), phoneNumbers(array), position(string), primaryEmailAddress(object), primaryPhoneNumber(object), profileUrl(string), resumeFileHandle(object), school(string), socialLinks(array), source(object), tags(array), timezone(string), updatedAt(string)
  candidate_list_client_info:
    primary key: id
    fields: candidateId(string), createdAt(string), id(string), ipAddress(string), relatedEntityId(string), relatedEntityType(string), userAgent(string)
  candidate_list_fraud_checks:
    primary key: id
    fields: applicationId(string), candidateId(string), createdAt(string), fraudSignals(array), id(string)
  candidate_list_notes:
    primary key: id
    fields: author(object), content(string), createdAt(string), id(string), isPrivate(boolean)
  candidate_list_projects:
    primary key: id
    fields: authorId(string), confidential(boolean), createdAt(string), customFieldEntries(array), description(string), id(string), isArchived(boolean), title(string)
  candidate_tag_list:
    primary key: id
    fields: id(string), isArchived(boolean), title(string)
  communication_template_list:
    primary key: id
    fields: createdAt(string), id(string), intendedTypes(array), title(string), updatedAt(string)
  feedback_form_definition_list:
    primary key: id
    fields: formDefinition(object), id(string), interviewId(string), isArchived(boolean), isDefaultForm(boolean), organizationId(string), title(string)
  feedback_form_definition_info:
    primary key: id
    fields: formDefinition(object), id(string), interviewId(string), isArchived(boolean), isDefaultForm(boolean), organizationId(string), title(string)
  job_posting_list:
    primary key: id
    fields: applicationDeadline(string), applyLink(string), compensationTierSummary(string), departmentName(string), employmentType(string), externalLink(string), id(string), isListed(boolean), jobId(string), locationIds(object), locationName(string), publishedDate(string), shouldDisplayCompensationOnJobBoard(boolean), status(string), teamName(string), title(string), updatedAt(string), workplaceType(string)
  job_posting_info:
    primary key: id
    fields: address(object), applicationConfirmationEmailTemplateId(string), applicationDeadline(string), applicationFormDefinition(object), applicationLimitCalloutHtml(string), applyLink(string), compensation(object), departmentName(string), descriptionHtml(string), descriptionParts(object), descriptionPlain(string), descriptionSocial(string), employmentType(string), externalLink(string), id(string), isListed(boolean), isRemote(boolean), job(object), jobId(string), linkedData(object), locationAddress(string), locationIds(object), locationName(string), publishedDate(string), status(string), suppressDescriptionClosing(boolean), suppressDescriptionOpening(boolean), surveyFormDefinitions(array), teamName(string), teamNameHierarchy(array), title(string), updatedAt(string), workplaceType(string)
  job_info:
    primary key: id
    fields: author(object), brandId(string), closedAt(string), compensation(object), confidential(boolean), createdAt(string), customFields(array), customRequisitionId(string), defaultInterviewPlanId(string), departmentId(string), employmentType(string), hiringTeam(array), id(string), interviewPlanIds(array), jobPostingIds(array), location(object), locationId(string), openedAt(string), openings(array), status(string), title(string), updatedAt(string)
  job_board_list:
    primary key: id
    fields: id(string), isInternal(boolean), title(string)
  job_interview_plan_info:
    primary key: jobId
    fields: interviewPlanId(string), jobId(string), stages(array)
  job_template_list:
    primary key: id
    fields: createdAt(string), defaultInterviewPlanId(string), departmentId(string), id(string), interviewPlanIds(array), location(object), locationId(string), status(string), title(string), updatedAt(string)
  department_info:
    primary key: id
    fields: createdAt(string), externalName(string), extraData(object), id(string), isArchived(boolean), name(string), parentId(string), updatedAt(string)
  department_list:
    primary key: id
    fields: createdAt(string), externalName(string), extraData(object), id(string), isArchived(boolean), name(string), parentId(string), updatedAt(string)
  location_list:
    primary key: id
    fields: address(object), externalName(string), extraData(object), id(string), isArchived(boolean), isRemote(boolean), name(string), parentLocationId(string), type(string), workplaceType(string)
  location_info:
    primary key: id
    fields: address(object), externalName(string), extraData(object), id(string), isArchived(boolean), isRemote(boolean), name(string), parentLocationId(string), type(string), workplaceType(string)
  interview_plan_list:
    primary key: id
    fields: createdAt(string), id(string), isArchived(boolean), title(string), updatedAt(string)
  interview_stage_list:
    primary key: id
    fields: id(string), interviewPlanId(string), interviewStageGroupId(string), orderInInterviewPlan(integer), title(string), type(string)
  interview_stage_group_list:
    primary key: id
    fields: id(string), order(integer), stageType(string), title(string)
  offer_list:
    primary key: id
    fields: acceptanceStatus(string), applicationId(string), decidedAt(string), formDefinition(object), id(string), latestVersion(object), offerStatus(string), versions(array)
  offer_info:
    primary key: id
    fields: acceptanceStatus(string), applicationId(string), decidedAt(string), formDefinition(object), id(string), latestVersion(object), offerStatus(string), versions(array)
  opening_info:
    primary key: id
    fields: archivedAt(string), closeReasonId(string), closedAt(string), id(string), isArchived(boolean), latestVersion(object), openedAt(string), openingState(string)
  opening_list:
    primary key: id
    fields: archivedAt(string), closeReasonId(string), closedAt(string), id(string), isArchived(boolean), latestVersion(object), openedAt(string), openingState(string)
  project_info:
    primary key: id
    fields: authorId(string), confidential(boolean), createdAt(string), customFieldEntries(array), description(string), id(string), isArchived(boolean), title(string)
  project_list:
    primary key: id
    fields: authorId(string), confidential(boolean), createdAt(string), customFieldEntries(array), description(string), id(string), isArchived(boolean), title(string)
  source_list:
    primary key: id
    fields: id(string), isArchived(boolean), sourceType(object), title(string)
  source_tracking_link_list:
    primary key: id
    fields: code(string), enabled(boolean), id(string), link(string), sourceId(string)
  archive_reason_list:
    primary key: id
    fields: id(string), isArchived(boolean), reasonType(string), text(string)
  brand_list:
    primary key: id
    fields: hostedJobsPageSlug(string), id(string), name(string)
  custom_field_list:
    primary key: id
    fields: description(string), fieldType(string), id(string), isArchived(boolean), isPrivate(boolean), isRequired(boolean), objectType(string), selectableValues(array), title(string)
  custom_field_info:
    primary key: id
    fields: description(string), fieldType(string), id(string), isArchived(boolean), isPrivate(boolean), isRequired(boolean), objectType(string), selectableValues(array), title(string)
  hiring_team_role_list:
    primary key: value
    fields: value(string)
  user_info:
    primary key: id
    fields: customFields(array), email(string), firstName(string), globalRole(string), id(string), isEnabled(boolean), lastName(string), managerId(string), updatedAt(string)
  user_list_interviewer_pauses:
    primary key: id
    fields: comment(string), createdAt(string), endsAt(string), id(string), startsAt(string), userId(string)
  email_sender_list:
    primary key: email
    fields: displayName(string), email(string), type(string)
  sequence_info:
    primary key: id
    fields: applicationId(string), candidateId(string), createdAt(string), id(string), sequenceTemplateId(string), stages(array), status(string)
  sequence_list:
    primary key: id
    fields: applicationId(string), candidateId(string), createdAt(string), id(string), sequenceTemplateId(string), stages(array), status(string)
  sequence_template_info:
    primary key: id
    fields: id(string), isArchived(boolean), stages(array), title(string), unsubscribeLinkActive(boolean), updatedAt(string)
  sequence_template_list:
    primary key: id
    fields: id(string), isArchived(boolean), stages(array), title(string), unsubscribeLinkActive(boolean), updatedAt(string)
  interview_schedule_list:
    primary key: id
    fields: applicationId(string), createdAt(string), id(string), interviewEvents(array), interviewStageId(string), scheduledBy(object), status(string), updatedAt(string)
  take_home_assignment_list:
    primary key: id
    fields: applicationId(string), candidateId(string), createdAt(string), feedbackFormDefinitionId(string), id(string), interview(object), interviewId(string), interviewStageId(string), reviewers(array), status(string), submission(object), updatedAt(string)
  take_home_assignment_info:
    primary key: id
    fields: applicationId(string), candidateId(string), createdAt(string), feedbackFormDefinitionId(string), id(string), interview(object), interviewId(string), interviewStageId(string), reviewers(array), status(string), submission(object), updatedAt(string)
  interview_event_list:
    primary key: id
    fields: createdAt(string), endTime(string), extraData(object), feedbackLink(string), hasSubmittedFeedback(boolean), id(string), interview(object), interviewId(string), interviewScheduleId(string), interviewerCalendarEventId(string), interviewerUserIds(array), interviewers(array), location(string), meetingLink(string), notetakerTranscriptId(string), startTime(string), updatedAt(string)
  interview_briefing_info:
    primary key: id
    cursor: candidate
    fields: application(object), applicationId(string), candidate(object), feedbackFormDefinition(object), feedbackFormDefinitionId(string), hasSubmittedFeedback(boolean), id(string), interview(object), interviewId(string), interviewStageId(string), interviewers(array), job(object)
  interview_info:
    primary key: id
    fields: externalTitle(string), feedbackFormDefinitionId(string), id(string), instructionsHtml(string), instructionsPlain(string), isArchived(boolean), isDebrief(boolean), isFeedbackRequested(boolean), isFeedbackRequired(boolean), jobId(string), title(string), type(string)
  interview_list:
    primary key: id
    fields: externalTitle(string), feedbackFormDefinitionId(string), id(string), instructionsHtml(string), instructionsPlain(string), isArchived(boolean), isDebrief(boolean), isFeedbackRequested(boolean), isFeedbackRequired(boolean), jobId(string), title(string), type(string)
  interview_stage_info:
    primary key: id
    fields: id(string), interviewPlanId(string), interviewStageGroupId(string), orderInInterviewPlan(integer), title(string), type(string)
  survey_form_definition_info:
    primary key: id
    fields: formDefinition(object), id(string), isArchived(boolean), surveyType(string), title(string)
  survey_form_definition_list:
    primary key: id
    fields: formDefinition(object), id(string), isArchived(boolean), surveyType(string), title(string)
  survey_request_list:
    primary key: id
    cursor: candidateId
    fields: applicationId(string), candidateId(string), id(string), surveyFormDefinitionId(string), surveyUrl(string)
  survey_submission_list:
    primary key: id
    fields: applicationId(string), candidateId(string), formDefinition(object), id(string), submittedAt(string), submittedValues(object), surveyFormDefinitionId(string), surveyType(string)
  webhook_info:
    primary key: id
    fields: enabled(boolean), id(string), requestUrl(string), webhookType(string)
  interviewer_pool_list:
    primary key: id
    fields: id(string), isArchived(boolean), title(string), trainingPath(object)
  interviewer_pool_info:
    primary key: id
    fields: id(string), isArchived(boolean), qualifiedMembers(array), title(string), trainees(array), trainingPath(object)
  close_reason_list:
    primary key: id
    fields: id(string), isArchived(boolean), reasonText(string)
  report_synchronous:
    primary key: requestId
    fields: failureReason(string), reportData(object), requestId(string), status(string)
  approval_list:
    primary key: id
    fields: approvalDefinitionId(string), completedAt(string), createdAt(string), entityId(string), entityType(string), id(string), steps(array), submittedAt(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_application:
    endpoint: POST /application.create
    required fields: candidateId, jobId
    risk: Executes Ashby application.create through the documented POST /application.create endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.
  update_application:
    endpoint: POST /application.update
    required fields: applicationId
    risk: Executes Ashby application.update through the documented POST /application.update endpoint; reverse ETL plan, preview, approval, and execute are required.
  delete_application:
    endpoint: POST /application.delete
    required fields: applicationId
    risk: Executes Ashby application.delete through the documented POST /application.delete endpoint; reverse ETL plan, preview, approval, and execute are required.
  add_application_hiring_team_member:
    endpoint: POST /application.addHiringTeamMember
    required fields: applicationId, teamMemberId, roleId
    risk: Executes Ashby application.addHiringTeamMember through the documented POST /application.addHiringTeamMember endpoint; reverse ETL plan, preview, approval, and execute are required.
  remove_application_hiring_team_member:
    endpoint: POST /application.removeHiringTeamMember
    required fields: applicationId, teamMemberId, roleId
    risk: Executes Ashby application.removeHiringTeamMember through the documented POST /application.removeHiringTeamMember endpoint; reverse ETL plan, preview, approval, and execute are required.
  change_application_stage:
    endpoint: POST /application.change_stage
    required fields: applicationId, interviewStageId
    risk: Executes Ashby application.change_stage through the documented POST /application.change_stage endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.
  change_application_stage_2:
    endpoint: POST /application.changeStage
    required fields: applicationId, interviewStageId
    risk: Executes Ashby application.changeStage through the documented POST /application.changeStage endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.
  change_application_source:
    endpoint: POST /application.change_source
    required fields: applicationId, sourceId
    risk: Executes Ashby application.change_source through the documented POST /application.change_source endpoint; reverse ETL plan, preview, approval, and execute are required.
  change_application_source_2:
    endpoint: POST /application.changeSource
    required fields: applicationId, sourceId
    risk: Executes Ashby application.changeSource through the documented POST /application.changeSource endpoint; reverse ETL plan, preview, approval, and execute are required.
  transfer_application:
    endpoint: POST /application.transfer
    required fields: applicationId, jobId, interviewPlanId, interviewStageId
    risk: Executes Ashby application.transfer through the documented POST /application.transfer endpoint; reverse ETL plan, preview, approval, and execute are required.
  update_application_history:
    endpoint: POST /application.updateHistory
    required fields: applicationId, applicationHistory
    risk: Executes Ashby application.updateHistory through the documented POST /application.updateHistory endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.
  submit_application_feedback:
    endpoint: POST /applicationFeedback.submit
    required fields: feedbackForm, formDefinitionId, applicationId
    risk: Executes Ashby applicationFeedback.submit through the documented POST /applicationFeedback.submit endpoint; reverse ETL plan, preview, approval, and execute are required.
  create_application_feedback_request:
    endpoint: POST /applicationFeedbackRequest.create
    required fields: applicationId, interviewId, interviewerUserId
    risk: Executes Ashby applicationFeedbackRequest.create through the documented POST /applicationFeedbackRequest.create endpoint; reverse ETL plan, preview, approval, and execute are required.
  update_approval_definition:
    endpoint: POST /approvalDefinition.update
    required fields: entityType, entityId, approvalStepDefinitions
    risk: Executes Ashby approvalDefinition.update through the documented POST /approvalDefinition.update endpoint; reverse ETL plan, preview, approval, and execute are required.
  create_candidate:
    endpoint: POST /candidate.create
    required fields: name
    risk: Executes Ashby candidate.create through the documented POST /candidate.create endpoint; reverse ETL plan, preview, approval, and execute are required.
  upload_candidate_resume:
    endpoint: POST /candidate.uploadResume
    required fields: candidateId, resumeHandle
    risk: Executes Ashby candidate.uploadResume through the documented POST /candidate.uploadResume endpoint; reverse ETL plan, preview, approval, and execute are required.
  upload_candidate_file:
    endpoint: POST /candidate.uploadFile
    required fields: candidateId, fileHandle
    risk: Executes Ashby candidate.uploadFile through the documented POST /candidate.uploadFile endpoint; reverse ETL plan, preview, approval, and execute are required.
  update_candidate:
    endpoint: POST /candidate.update
    required fields: candidateId
    risk: Executes Ashby candidate.update through the documented POST /candidate.update endpoint; reverse ETL plan, preview, approval, and execute are required.
  create_candidate_note:
    endpoint: POST /candidate.createNote
    required fields: candidateId, note
    risk: Executes Ashby candidate.createNote through the documented POST /candidate.createNote endpoint; reverse ETL plan, preview, approval, and execute are required.
  set_candidate_fraud_status:
    endpoint: POST /candidate.setFraudStatus
    required fields: candidateId, fraudStatus
    risk: Executes Ashby candidate.setFraudStatus through the documented POST /candidate.setFraudStatus endpoint; reverse ETL plan, preview, approval, and execute are required.
  anonymize_candidate:
    endpoint: POST /candidate.anonymize
    required fields: candidateId
    risk: Executes Ashby candidate.anonymize through the documented POST /candidate.anonymize endpoint; reverse ETL plan, preview, approval, and execute are required.
  remove_candidate_project:
    endpoint: POST /candidate.removeProject
    required fields: candidateId, projectId
    risk: Executes Ashby candidate.removeProject through the documented POST /candidate.removeProject endpoint; reverse ETL plan, preview, approval, and execute are required.
  add_candidate_project:
    endpoint: POST /candidate.addProject
    required fields: candidateId, projectId
    risk: Executes Ashby candidate.addProject through the documented POST /candidate.addProject endpoint; reverse ETL plan, preview, approval, and execute are required.
  add_candidate_email_message:
    endpoint: POST /candidate.addEmailMessage
    required fields: candidateId, emailProviderEmailId, subject, from, to, body
    risk: Executes Ashby candidate.addEmailMessage through the documented POST /candidate.addEmailMessage endpoint; reverse ETL plan, preview, approval, and execute are required.
  add_candidate_tag:
    endpoint: POST /candidate.addTag
    required fields: candidateId, tagId
    risk: Executes Ashby candidate.addTag through the documented POST /candidate.addTag endpoint; reverse ETL plan, preview, approval, and execute are required.
  remove_candidate_tag:
    endpoint: POST /candidate.removeTag
    required fields: candidateId, tagId
    risk: Executes Ashby candidate.removeTag through the documented POST /candidate.removeTag endpoint; reverse ETL plan, preview, approval, and execute are required.
  push_candidate_to_hris:
    endpoint: POST /candidate.pushToHris
    required fields: applicationId, externalSystem
    risk: Executes Ashby candidate.pushToHris through the documented POST /candidate.pushToHris endpoint; reverse ETL plan, preview, approval, and execute are required.
  create_candidate_tag:
    endpoint: POST /candidateTag.create
    required fields: title
    risk: Executes Ashby candidateTag.create through the documented POST /candidateTag.create endpoint; reverse ETL plan, preview, approval, and execute are required.
  archive_candidate_tag:
    endpoint: POST /candidateTag.archive
    required fields: tagId
    risk: Executes Ashby candidateTag.archive through the documented POST /candidateTag.archive endpoint; reverse ETL plan, preview, approval, and execute are required.
  update_job_posting_compensation:
    endpoint: POST /jobPosting.updateCompensation
    required fields: jobPostingId, compensationTiers
    risk: Executes Ashby jobPosting.updateCompensation through the documented POST /jobPosting.updateCompensation endpoint; reverse ETL plan, preview, approval, and execute are required.
  update_job_posting:
    endpoint: POST /jobPosting.update
    required fields: jobPostingId
    risk: Executes Ashby jobPosting.update through the documented POST /jobPosting.update endpoint; reverse ETL plan, preview, approval, and execute are required.
  set_job_posting_status:
    endpoint: POST /jobPosting.setStatus
    required fields: jobPostingId, status
    risk: Executes Ashby jobPosting.setStatus through the documented POST /jobPosting.setStatus endpoint; reverse ETL plan, preview, approval, and execute are required.
  create_job:
    endpoint: POST /job.create
    required fields: title, teamId, locationId
    risk: Executes Ashby job.create through the documented POST /job.create endpoint; reverse ETL plan, preview, approval, and execute are required.
  update_job:
    endpoint: POST /job.update
    required fields: jobId
    risk: Executes Ashby job.update through the documented POST /job.update endpoint; reverse ETL plan, preview, approval, and execute are required.
  update_job_compensation:
    endpoint: POST /job.updateCompensation
    required fields: jobId, compensationTiers
    risk: Executes Ashby job.updateCompensation through the documented POST /job.updateCompensation endpoint; reverse ETL plan, preview, approval, and execute are required.
  set_job_status:
    endpoint: POST /job.setStatus
    required fields: jobId, status
    risk: Executes Ashby job.setStatus through the documented POST /job.setStatus endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.
  archive_department:
    endpoint: POST /department.archive
    required fields: departmentId
    risk: Executes Ashby department.archive through the documented POST /department.archive endpoint; reverse ETL plan, preview, approval, and execute are required.
  restore_department:
    endpoint: POST /department.restore
    required fields: departmentId
    risk: Executes Ashby department.restore through the documented POST /department.restore endpoint; reverse ETL plan, preview, approval, and execute are required.
  create_department:
    endpoint: POST /department.create
    required fields: name
    risk: Executes Ashby department.create through the documented POST /department.create endpoint; reverse ETL plan, preview, approval, and execute are required.
  move_department:
    endpoint: POST /department.move
    required fields: departmentId
    risk: Executes Ashby department.move through the documented POST /department.move endpoint; reverse ETL plan, preview, approval, and execute are required.
  update_department:
    endpoint: POST /department.update
    required fields: departmentId, name
    risk: Executes Ashby department.update through the documented POST /department.update endpoint; reverse ETL plan, preview, approval, and execute are required.
  archive_location:
    endpoint: POST /location.archive
    required fields: locationId
    risk: Executes Ashby location.archive through the documented POST /location.archive endpoint; reverse ETL plan, preview, approval, and execute are required.
  create_location:
    endpoint: POST /location.create
    required fields: name, type
    risk: Executes Ashby location.create through the documented POST /location.create endpoint; reverse ETL plan, preview, approval, and execute are required.
  restore_location:
    endpoint: POST /location.restore
    required fields: locationId
    risk: Executes Ashby location.restore through the documented POST /location.restore endpoint; reverse ETL plan, preview, approval, and execute are required.
  move_location:
    endpoint: POST /location.move
    required fields: locationId, parentLocationHierarchyId
    risk: Executes Ashby location.move through the documented POST /location.move endpoint; reverse ETL plan, preview, approval, and execute are required.
  update_location_address:
    endpoint: POST /location.updateAddress
    required fields: locationId
    risk: Executes Ashby location.updateAddress through the documented POST /location.updateAddress endpoint; reverse ETL plan, preview, approval, and execute are required.
  update_location_name:
    endpoint: POST /location.updateName
    required fields: locationId, name
    risk: Executes Ashby location.updateName through the documented POST /location.updateName endpoint; reverse ETL plan, preview, approval, and execute are required.
  update_location_workplace_type:
    endpoint: POST /location.updateWorkplaceType
    required fields: locationId, workplaceType
    risk: Executes Ashby location.updateWorkplaceType through the documented POST /location.updateWorkplaceType endpoint; reverse ETL plan, preview, approval, and execute are required.
  update_location_external_name:
    endpoint: POST /location.updateExternalName
    required fields: locationId
    risk: Executes Ashby location.updateExternalName through the documented POST /location.updateExternalName endpoint; reverse ETL plan, preview, approval, and execute are required.
  approve_offer:
    endpoint: POST /offer.approve
    required fields: offerVersionId
    risk: Executes Ashby offer.approve through the documented POST /offer.approve endpoint; reverse ETL plan, preview, approval, and execute are required.
  create_offer:
    endpoint: POST /offer.create
    required fields: offerProcessId, offerFormId, offerForm
    risk: Executes Ashby offer.create through the documented POST /offer.create endpoint; reverse ETL plan, preview, approval, and execute are required.
  start_offer:
    endpoint: POST /offer.start
    required fields: offerProcessId
    risk: Executes Ashby offer.start through the documented POST /offer.start endpoint; reverse ETL plan, preview, approval, and execute are required.
  start_offer_approval_process:
    endpoint: POST /offer.startApprovalProcess
    required fields: offerVersionId
    risk: Executes Ashby offer.startApprovalProcess through the documented POST /offer.startApprovalProcess endpoint; reverse ETL plan, preview, approval, and execute are required.
  update_offer:
    endpoint: POST /offer.update
    required fields: offerId, offerForm
    risk: Executes Ashby offer.update through the documented POST /offer.update endpoint; reverse ETL plan, preview, approval, and execute are required.
  set_offer_status:
    endpoint: POST /offer.setStatus
    required fields: offerId, acceptanceStatus
    risk: Executes Ashby offer.setStatus through the documented POST /offer.setStatus endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.
  set_offer_decided_at:
    endpoint: POST /offer.setDecidedAt
    required fields: offerId, decidedAt
    risk: Executes Ashby offer.setDecidedAt through the documented POST /offer.setDecidedAt endpoint; reverse ETL plan, preview, approval, and execute are required.
  start_offer_process:
    endpoint: POST /offerProcess.start
    required fields: applicationId
    risk: Executes Ashby offerProcess.start through the documented POST /offerProcess.start endpoint; reverse ETL plan, preview, approval, and execute are required.
  create_opening:
    endpoint: POST /opening.create
    risk: Executes Ashby opening.create through the documented POST /opening.create endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.
  update_opening:
    endpoint: POST /opening.update
    required fields: openingId
    risk: Executes Ashby opening.update through the documented POST /opening.update endpoint; reverse ETL plan, preview, approval, and execute are required.
  set_opening_archived:
    endpoint: POST /opening.setArchived
    required fields: openingId, archive
    risk: Executes Ashby opening.setArchived through the documented POST /opening.setArchived endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.
  set_opening_opening_state:
    endpoint: POST /opening.setOpeningState
    required fields: openingId, openingState
    risk: Executes Ashby opening.setOpeningState through the documented POST /opening.setOpeningState endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.
  add_opening_job:
    endpoint: POST /opening.addJob
    required fields: openingId, jobId
    risk: Executes Ashby opening.addJob through the documented POST /opening.addJob endpoint; reverse ETL plan, preview, approval, and execute are required.
  remove_opening_job:
    endpoint: POST /opening.removeJob
    required fields: openingId, jobId
    risk: Executes Ashby opening.removeJob through the documented POST /opening.removeJob endpoint; reverse ETL plan, preview, approval, and execute are required.
  add_opening_location:
    endpoint: POST /opening.addLocation
    required fields: openingId, locationId
    risk: Executes Ashby opening.addLocation through the documented POST /opening.addLocation endpoint; reverse ETL plan, preview, approval, and execute are required.
  remove_opening_location:
    endpoint: POST /opening.removeLocation
    required fields: openingId, locationId
    risk: Executes Ashby opening.removeLocation through the documented POST /opening.removeLocation endpoint; reverse ETL plan, preview, approval, and execute are required.
  create_custom_field:
    endpoint: POST /customField.create
    required fields: fieldType, objectType, title
    risk: Executes Ashby customField.create through the documented POST /customField.create endpoint; reverse ETL plan, preview, approval, and execute are required.
  set_custom_field_value:
    endpoint: POST /customField.setValue
    required fields: objectId, objectType, fieldId, fieldValue
    risk: Executes Ashby customField.setValue through the documented POST /customField.setValue endpoint; reverse ETL plan, preview, approval, and execute are required.
  set_custom_field_values:
    endpoint: POST /customField.setValues
    required fields: objectId, objectType, values
    risk: Executes Ashby customField.setValues through the documented POST /customField.setValues endpoint; reverse ETL plan, preview, approval, and execute are required.
  update_custom_field_selectable_values:
    endpoint: POST /customField.updateSelectableValues
    required fields: customFieldId, selectableValues
    risk: Executes Ashby customField.updateSelectableValues through the documented POST /customField.updateSelectableValues endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.
  update_assessment:
    endpoint: POST /assessment.update
    required fields: assessment_id, timestamp
    risk: Executes Ashby assessment.update through the documented POST /assessment.update endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.
  add_assessment_completed_to_candidate:
    endpoint: POST /assessment.addCompletedToCandidate
    required fields: candidateId, partnerId, assessment, timestamp
    risk: Executes Ashby assessment.addCompletedToCandidate through the documented POST /assessment.addCompletedToCandidate endpoint; reverse ETL plan, preview, approval, and execute are required.
  add_hiring_team_member:
    endpoint: POST /hiringTeam.addMember
    required fields: teamMemberId, roleId
    risk: Executes Ashby hiringTeam.addMember through the documented POST /hiringTeam.addMember endpoint; reverse ETL plan, preview, approval, and execute are required.
  remove_hiring_team_member:
    endpoint: POST /hiringTeam.removeMember
    required fields: teamMemberId, roleId
    risk: Executes Ashby hiringTeam.removeMember through the documented POST /hiringTeam.removeMember endpoint; reverse ETL plan, preview, approval, and execute are required.
  update_user_interviewer_settings:
    endpoint: POST /user.updateInterviewerSettings
    required fields: userId
    risk: Executes Ashby user.updateInterviewerSettings through the documented POST /user.updateInterviewerSettings endpoint; reverse ETL plan, preview, approval, and execute are required.
  create_user_interviewer_pause:
    endpoint: POST /user.createInterviewerPause
    required fields: userId
    risk: Executes Ashby user.createInterviewerPause through the documented POST /user.createInterviewerPause endpoint; reverse ETL plan, preview, approval, and execute are required.
  delete_user_interviewer_pause:
    endpoint: POST /user.deleteInterviewerPause
    required fields: interviewerPauseId
    risk: Executes Ashby user.deleteInterviewerPause through the documented POST /user.deleteInterviewerPause endpoint; reverse ETL plan, preview, approval, and execute are required.
  set_user_custom_field_value:
    endpoint: POST /user.setCustomFieldValue
    required fields: userId, fieldId, fieldValue
    risk: Executes Ashby user.setCustomFieldValue through the documented POST /user.setCustomFieldValue endpoint; reverse ETL plan, preview, approval, and execute are required.
  set_user_custom_field_values:
    endpoint: POST /user.setCustomFieldValues
    required fields: userId, values
    risk: Executes Ashby user.setCustomFieldValues through the documented POST /user.setCustomFieldValues endpoint; reverse ETL plan, preview, approval, and execute are required.
  create_referral:
    endpoint: POST /referral.create
    required fields: id, creditedToUserId, fieldSubmissions
    risk: Executes Ashby referral.create through the documented POST /referral.create endpoint; reverse ETL plan, preview, approval, and execute are required.
  add_sequence:
    endpoint: POST /sequence.add
    required fields: sequenceTemplateId, candidateId, start
    risk: Executes Ashby sequence.add through the documented POST /sequence.add endpoint; reverse ETL plan, preview, approval, and execute are required.
  cancel_sequence:
    endpoint: POST /sequence.cancel
    required fields: sequenceId
    risk: Executes Ashby sequence.cancel through the documented POST /sequence.cancel endpoint; reverse ETL plan, preview, approval, and execute are required.
  discard_sequence:
    endpoint: POST /sequence.discard
    required fields: sequenceId
    risk: Executes Ashby sequence.discard through the documented POST /sequence.discard endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required.
  update_sequence_stage:
    endpoint: POST /sequence.updateStage
    required fields: sequenceId, stageId
    risk: Executes Ashby sequence.updateStage through the documented POST /sequence.updateStage endpoint; reverse ETL plan, preview, approval, and execute are required.
  start_sequence:
    endpoint: POST /sequence.start
    required fields: sequenceId
    risk: Executes Ashby sequence.start through the documented POST /sequence.start endpoint; reverse ETL plan, preview, approval, and execute are required.
  create_interview_schedule:
    endpoint: POST /interviewSchedule.create
    required fields: applicationId, interviewEvents
    risk: Executes Ashby interviewSchedule.create through the documented POST /interviewSchedule.create endpoint; reverse ETL plan, preview, approval, and execute are required.
  update_interview_schedule:
    endpoint: POST /interviewSchedule.update
    required fields: interviewScheduleId
    risk: Executes Ashby interviewSchedule.update through the documented POST /interviewSchedule.update endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.
  cancel_interview_schedule:
    endpoint: POST /interviewSchedule.cancel
    required fields: id
    risk: Executes Ashby interviewSchedule.cancel through the documented POST /interviewSchedule.cancel endpoint; reverse ETL plan, preview, approval, and execute are required.
  create_survey_request:
    endpoint: POST /surveyRequest.create
    required fields: candidateId, applicationId, surveyFormDefinitionId
    risk: Executes Ashby surveyRequest.create through the documented POST /surveyRequest.create endpoint; reverse ETL plan, preview, approval, and execute are required.
  create_survey_submission:
    endpoint: POST /surveySubmission.create
    required fields: surveyFormDefinitionId, candidateId, applicationId, submittedValues
    risk: Executes Ashby surveySubmission.create through the documented POST /surveySubmission.create endpoint; reverse ETL plan, preview, approval, and execute are required.
  create_webhook:
    endpoint: POST /webhook.create
    required fields: webhookType, requestUrl, secretToken
    risk: Executes Ashby webhook.create through the documented POST /webhook.create endpoint; reverse ETL plan, preview, approval, and execute are required.
  update_webhook:
    endpoint: POST /webhook.update
    required fields: webhookId
    risk: Executes Ashby webhook.update through the documented POST /webhook.update endpoint; reverse ETL plan, preview, approval, and execute are required.
  delete_webhook:
    endpoint: POST /webhook.delete
    required fields: webhookId
    risk: Executes Ashby webhook.delete through the documented POST /webhook.delete endpoint; reverse ETL plan, preview, approval, and execute are required.
  archive_interviewer_pool:
    endpoint: POST /interviewerPool.archive
    required fields: interviewerPoolId
    risk: Executes Ashby interviewerPool.archive through the documented POST /interviewerPool.archive endpoint; reverse ETL plan, preview, approval, and execute are required.
  restore_interviewer_pool:
    endpoint: POST /interviewerPool.restore
    required fields: interviewerPoolId
    risk: Executes Ashby interviewerPool.restore through the documented POST /interviewerPool.restore endpoint; reverse ETL plan, preview, approval, and execute are required.
  create_interviewer_pool:
    endpoint: POST /interviewerPool.create
    required fields: title
    risk: Executes Ashby interviewerPool.create through the documented POST /interviewerPool.create endpoint; reverse ETL plan, preview, approval, and execute are required.
  update_interviewer_pool:
    endpoint: POST /interviewerPool.update
    required fields: interviewerPoolId
    risk: Executes Ashby interviewerPool.update through the documented POST /interviewerPool.update endpoint; reverse ETL plan, preview, approval, and execute are required.
  add_interviewer_pool_user:
    endpoint: POST /interviewerPool.addUser
    required fields: interviewerPoolId, userId
    risk: Executes Ashby interviewerPool.addUser through the documented POST /interviewerPool.addUser endpoint; reverse ETL plan, preview, approval, and execute are required.
  remove_interviewer_pool_user:
    endpoint: POST /interviewerPool.removeUser
    required fields: interviewerPoolId, userId
    risk: Executes Ashby interviewerPool.removeUser through the documented POST /interviewerPool.removeUser endpoint; reverse ETL plan, preview, approval, and execute are required.

SECURITY
  read risk: external Ashby API reads through fixed declaration-bound routes
  write risk: named reverse-ETL actions only; no generic HTTP method/path/body; destructive actions require typed confirmation
  approval: reverse ETL writes require plan -> preview -> explicit approval -> execute
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Ashby applicant-tracking connector with typed REST streams, bounded direct reads, and gated reverse-ETL writes.
  Usage: pm connectors command ashby <command> [flags]
  Global flags:
    --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
  ETL streams
  Bounded direct reads
  Reverse ETL writes
  Other Commands
    candidate list - Full-refresh-only Ashby candidate list read. Opaque syncToken checkpointing is unavailable pending ashby-sync-token-checkpoint-foundation. [intent=etl availability=implemented stream=candidates]; notes: Fixed Ashby stream for candidate.list; flags map only to documented request body fields. Opaque syncToken checkpointing is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --created-after, --created-before
    job list - Full-refresh-only Ashby job list read. Opaque syncToken checkpointing is unavailable pending ashby-sync-token-checkpoint-foundation. [intent=etl availability=implemented stream=jobs]; notes: Fixed Ashby stream for job.list; flags map only to documented request body fields. Repeatable array request variants (--status, --expand) are blocked pending connector-stream-repeatable-array-foundation. Opaque syncToken checkpointing is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --created-after, --opened-after, --opened-before, --closed-after, --closed-before, --include-unpublished-job-postings-ids
    application list - Full-refresh-only Ashby application list read. Opaque syncToken checkpointing is unavailable pending ashby-sync-token-checkpoint-foundation. [intent=etl availability=implemented stream=applications]; notes: Fixed Ashby stream for application.list; flags map only to documented request body fields. Repeatable array request variants (--expand) are blocked pending connector-stream-repeatable-array-foundation. Opaque syncToken checkpointing is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --created-after, --created-before, --status (non-empty), --job-id (non-empty)
    user list - Full-refresh-only Ashby user list read. Opaque syncToken checkpointing is unavailable pending ashby-sync-token-checkpoint-foundation. [intent=etl availability=implemented stream=users]; notes: Fixed Ashby stream for user.list; flags map only to documented request body fields. Opaque syncToken checkpointing is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --include-deactivated
    api-key info - Returns details for the API key used to make the request. Requires the apiKeysRead permission. [intent=etl availability=implemented stream=api_key_info]; notes: Fixed Ashby stream for apiKey.info; flags map only to documented request body fields. Incremental execution is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.
    audit-log list - Beta This API is in active development and only available in a closed beta with early design partners. [intent=etl availability=implemented stream=audit_log_list]; notes: Fixed Ashby stream for auditLog.list; flags map only to documented request body fields. Repeatable array request variants (--actor-ids, --target-ids, --target-types, --categories) are blocked pending connector-stream-repeatable-array-foundation. Incremental execution is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --start-date (non-empty, format=date-time), --end-date (non-empty, format=date-time)
    application create - Consider a candidate for a job (e.g. when sourcing a candidate for a job posting). [intent=reverse_etl availability=implemented write=create_application]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby application.create through the documented POST /application.create endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --candidate-id (non-empty), --job-id (non-empty), --interview-plan-id (non-empty), --interview-stage-id (non-empty), --source-id (non-empty), --credited-to-user-id (non-empty), --created-at (non-empty, format=date-time), --application-history
    application update - Update an application. To set values for custom fields on Applications, use the customField.setValue endpoint. [intent=reverse_etl availability=implemented write=update_application]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby application.update through the documented POST /application.update endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --application-id (non-empty), --source-id (non-empty), --credited-to-user-id (non-empty), --created-at (non-empty, format=date-time), --send-notifications
    application delete - Deletes an application by id. Requires the candidatesDelete permission. [intent=reverse_etl availability=implemented write=delete_application]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby application.delete through the documented POST /application.delete endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --application-id (non-empty)
    application info - Fetch application details by application id or by submitted form instance id (which is returned by the applicationForm.submit endpoint). [intent=etl availability=implemented stream=application_info]; notes: Fixed Ashby stream for application.info; flags map only to documented request body fields. Requires at least one documented selector: applicationId, submittedFormInstanceId. Repeatable array request variants (--expand) are blocked pending connector-stream-repeatable-array-foundation. Incremental execution is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --application-id (non-empty), --submitted-form-instance-id (non-empty)
    application add-hiring-team-member - Adds an Ashby user to the hiring team at the application level. [intent=reverse_etl availability=implemented write=add_application_hiring_team_member]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby application.addHiringTeamMember through the documented POST /application.addHiringTeamMember endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --application-id (non-empty), --team-member-id (non-empty), --role-id (non-empty)
    application remove-hiring-team-member - Unassigns a hiring team role from an Ashby user at the application level. [intent=reverse_etl availability=implemented write=remove_application_hiring_team_member]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby application.removeHiringTeamMember through the documented POST /application.removeHiringTeamMember endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --application-id (non-empty), --team-member-id (non-empty), --role-id (non-empty)
    application-hiring-team-role list - Gets all available hiring team roles for applications in the organization. [intent=etl availability=implemented stream=application_hiring_team_role_list]; notes: Fixed Ashby stream for applicationHiringTeamRole.list; flags map only to documented request body fields.
    application change-stage - Deprecated. Use application.changeStage instead. Change the stage of an application. [intent=reverse_etl availability=implemented write=change_application_stage]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby application.change_stage through the documented POST /application.change_stage endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --application-id (non-empty), --interview-stage-id (non-empty), --archive-reason-id (non-empty)
    application change-stage-2 - Change the stage of an application. Requires the candidatesWrite permission. [intent=reverse_etl availability=implemented write=change_application_stage_2]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby application.changeStage through the documented POST /application.changeStage endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --application-id (non-empty), --interview-stage-id (non-empty), --archive-reason-id (non-empty)
    application change-source - Deprecated. Use application.changeSource instead. Change the source of an application. [intent=reverse_etl availability=implemented write=change_application_source]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby application.change_source through the documented POST /application.change_source endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --application-id (non-empty), --source-id (non-empty)
    application change-source-2 - Change the source of an application. Requires the candidatesWrite permission. [intent=reverse_etl availability=implemented write=change_application_source_2]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby application.changeSource through the documented POST /application.changeSource endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --application-id (non-empty), --source-id (non-empty)
    application transfer - Transfer an application to a different job. Requires the candidatesWrite permission. [intent=reverse_etl availability=implemented write=transfer_application]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby application.transfer through the documented POST /application.transfer endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --application-id (non-empty), --job-id (non-empty), --interview-plan-id (non-empty), --interview-stage-id (non-empty), --start-automatic-activities
    application update-history - Update the history of an application. Used to update stage timestamps and to delete history events. Also requires the Allow updating application history? [intent=reverse_etl availability=implemented write=update_application_history]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby application.updateHistory through the documented POST /application.updateHistory endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --application-id (non-empty), --application-history-0-stage-id (non-empty), --application-history-0-stage-number, --application-history-0-entered-stage-at (non-empty, format=date-time)
    application list-history - Fetch a paginated list of application history items for an application. This endpoint supports pagination only (not incremental sync). [intent=etl availability=implemented stream=application_list_history]; notes: Fixed Ashby stream for application.listHistory; flags map only to documented request body fields.; flags: --application-id (required, non-empty)
    application list-criteria-evaluations - Fetch a paginated list of AI criteria evaluations for an application. [intent=etl availability=implemented stream=application_list_criteria_evaluations]; notes: Fixed Ashby stream for application.listCriteriaEvaluations; flags map only to documented request body fields. Incremental execution is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --application-id (required, non-empty)
    application-feedback list - Full-refresh-only Ashby application-feedback list read. Opaque syncToken checkpointing is unavailable pending ashby-sync-token-checkpoint-foundation. [intent=etl availability=implemented stream=application_feedback_list]; notes: Fixed Ashby stream for applicationFeedback.list; flags map only to documented request body fields. Opaque syncToken checkpointing is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --application-id (non-empty), --created-after
    application-feedback submit - Application feedback forms support a variety of field types. [intent=reverse_etl availability=partial write=submit_application_feedback]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby applicationFeedback.submit through the documented POST /applicationFeedback.submit endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed. This command has nested object/array requirements that are implemented by the reverse-ETL action schema but are not fully expressible as scalar CLI flags; use file/warehouse reverse-ETL inputs for execution.; flags: --feedback-form-field-submissions-0-path (non-empty), --form-definition-id (non-empty), --application-id (non-empty), --user-id (non-empty), --interview-event-id (non-empty)
    application-feedback-request create - Request feedback on an application without scheduling an interview. [intent=reverse_etl availability=implemented write=create_application_feedback_request]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby applicationFeedbackRequest.create through the documented POST /applicationFeedbackRequest.create endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --application-id (non-empty), --interview-id (non-empty), --interviewer-user-id (non-empty)
    approval-definition update - Create or update an approval definition for a specific entity that requires approval. [intent=reverse_etl availability=implemented write=update_approval_definition]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby approvalDefinition.update through the documented POST /approvalDefinition.update endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --entity-type (non-empty), --entity-id (non-empty), --approval-step-definitions-0-approvals-required, --approval-step-definitions-0-approvers-0-user-id (non-empty), --approval-step-definitions-0-approvers-0-type (non-empty), --submit-approval-request
    candidate search - Searches for candidates by email and/or name. Requires the candidatesRead permission. [intent=direct_read availability=implemented operation=ashby.direct.candidate.search]; approval: none; risk: bounded JSON direct read; credential-marked response fields are redacted, and non-credential identity fields remain complete in trusted live local output; notes: Fixed Ashby POST direct read; no raw method/path/body override is exposed.; flags: --email (non-empty), --name (non-empty), --page, --page-cursor
    candidate info - Fetches details about a single candidate by id or external mapping id. Requires the candidatesRead permission. [intent=etl availability=implemented stream=candidate_info]; notes: Fixed Ashby stream for candidate.info; flags map only to documented request body fields. Requires at least one documented selector: id, externalMappingId. Incremental execution is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --id (non-empty), --external-mapping-id (non-empty)
    candidate create - Creates a new candidate. Requires the candidatesWrite permission. [intent=reverse_etl availability=implemented write=create_candidate]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby candidate.create through the documented POST /candidate.create endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --name (non-empty), --email (non-empty), --phone-number (non-empty), --linked-in-url (non-empty), --github-url (non-empty), --website (non-empty), --alternate-email-addresses, --source-id (non-empty), --credited-to-user-id (non-empty), --created-at (non-empty, format=date-time)
    candidate upload-resume - Uploads a resume for a candidate. Accepts either a multipart/form-data request with a resume file part, or a JSON body with a resumeHandle previously... [intent=reverse_etl availability=implemented write=upload_candidate_resume]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby candidate.uploadResume through the documented POST /candidate.uploadResume endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --candidate-id (non-empty), --resume-handle (non-empty)
    candidate upload-file - Uploads a file for a candidate. Accepts either a multipart/form-data request with a file file part, or a JSON body with a fileHandle previously created... [intent=reverse_etl availability=implemented write=upload_candidate_file]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby candidate.uploadFile through the documented POST /candidate.uploadFile endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --candidate-id (non-empty), --file-handle (non-empty)
    candidate update - Updates an existing candidate. Requires the candidatesWrite permission. [intent=reverse_etl availability=implemented write=update_candidate]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby candidate.update through the documented POST /candidate.update endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --candidate-id (non-empty), --name (non-empty), --email (non-empty), --phone-number (non-empty), --linked-in-url (non-empty), --github-url (non-empty), --website-url (non-empty), --alternate-email (non-empty), --social-links, --source-id (non-empty), --credited-to-user-id (non-empty), --created-at (non-empty, format=date-time), --send-notifications
    candidate create-note - Creates a note on a candidate. Requires the candidatesWrite permission. [intent=reverse_etl availability=implemented write=create_candidate_note]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby candidate.createNote through the documented POST /candidate.createNote endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --candidate-id (non-empty), --note-type, --note-value (non-empty), --send-notifications, --is-private, --created-at (non-empty, format=date-time)
    candidate list-client-info - Full-refresh-only Ashby candidate list-client-info read. Opaque syncToken checkpointing is unavailable pending ashby-sync-token-checkpoint-foundation. [intent=etl availability=implemented stream=candidate_list_client_info]; notes: Fixed Ashby stream for candidate.listClientInfo; flags map only to documented request body fields. Opaque syncToken checkpointing is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --candidate-id (required, non-empty)
    candidate list-fraud-checks - Lists the fraud checks performed on a candidate. Requires the candidatesRead permission. [intent=etl availability=implemented stream=candidate_list_fraud_checks]; notes: Fixed Ashby stream for candidate.listFraudChecks; flags map only to documented request body fields. Incremental execution is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --candidate-id (required, non-empty)
    candidate set-fraud-status - Updates the manual fraud-review status of a candidate. Requires the candidatesWrite permission. [intent=reverse_etl availability=implemented write=set_candidate_fraud_status]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby candidate.setFraudStatus through the documented POST /candidate.setFraudStatus endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --candidate-id (non-empty), --fraud-status
    candidate list-notes - Full-refresh-only Ashby candidate list-notes read. Opaque syncToken checkpointing is unavailable pending ashby-sync-token-checkpoint-foundation. [intent=etl availability=implemented stream=candidate_list_notes]; notes: Fixed Ashby stream for candidate.listNotes; flags map only to documented request body fields. Opaque syncToken checkpointing is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --candidate-id (required, non-empty)
    candidate anonymize - Anonymizes a candidate's personally identifiable information. Requires the candidatesWrite permission. [intent=reverse_etl availability=implemented write=anonymize_candidate]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby candidate.anonymize through the documented POST /candidate.anonymize endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --candidate-id (non-empty)
    candidate remove-project - Removes the candidate from a project. Requires the candidatesWrite permission. [intent=reverse_etl availability=implemented write=remove_candidate_project]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby candidate.removeProject through the documented POST /candidate.removeProject endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --candidate-id (non-empty), --project-id (non-empty)
    candidate add-project - Adds a candidate to a project. Requires the candidatesWrite permission. [intent=reverse_etl availability=implemented write=add_candidate_project]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby candidate.addProject through the documented POST /candidate.addProject endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --candidate-id (non-empty), --project-id (non-empty)
    candidate add-email-message - Attaches an existing email message (e.g. fetched from a partner provider) to a candidate. [intent=reverse_etl availability=implemented write=add_candidate_email_message]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby candidate.addEmailMessage through the documented POST /candidate.addEmailMessage endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --candidate-id (non-empty), --email-provider-email-id (non-empty), --subject (non-empty), --from (non-empty), --to (non-empty), --message-body (non-empty), --user-id (non-empty), --sent-at (non-empty, format=date-time), --cc (non-empty), --message-url (non-empty), --message-id-header (non-empty), --thread-id (non-empty), --is-private
    candidate add-tag - Adds a tag to a candidate. Requires the candidatesWrite permission. [intent=reverse_etl availability=implemented write=add_candidate_tag]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby candidate.addTag through the documented POST /candidate.addTag endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --candidate-id (non-empty), --tag-id (non-empty)
    candidate remove-tag - Removes a tag from a candidate. Requires the candidatesWrite permission. [intent=reverse_etl availability=implemented write=remove_candidate_tag]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby candidate.removeTag through the documented POST /candidate.removeTag endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --candidate-id (non-empty), --tag-id (non-empty)
    candidate list-projects - Lists the projects a candidate has been added to. Requires the candidatesRead permission. [intent=etl availability=implemented stream=candidate_list_projects]; notes: Fixed Ashby stream for candidate.listProjects; flags map only to documented request body fields. Incremental execution is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --candidate-id (required, non-empty)
    candidate push-to-hris - Beta This feature is in beta and may not be available for all organizations. Pushes a candidate's data to an HRIS system (e.g. Workday, BambooHR, ADP). [intent=reverse_etl availability=implemented write=push_candidate_to_hris]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby candidate.pushToHris through the documented POST /candidate.pushToHris endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --application-id (non-empty), --external-system, --integration-partner-id (non-empty)
    candidate-tag create - Creates a candidate tag. If a tag already exists with the given title, the existing tag will be returned. [intent=reverse_etl availability=implemented write=create_candidate_tag]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby candidateTag.create through the documented POST /candidateTag.create endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --title (non-empty)
    candidate-tag archive - Archives a candidate tag. Requires the hiringProcessMetadataWrite permission. [intent=reverse_etl availability=implemented write=archive_candidate_tag]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby candidateTag.archive through the documented POST /candidateTag.archive endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --tag-id (non-empty)
    candidate-tag list - Full-refresh-only Ashby candidate-tag list read. Opaque syncToken checkpointing is unavailable pending ashby-sync-token-checkpoint-foundation. [intent=etl availability=implemented stream=candidate_tag_list]; notes: Fixed Ashby stream for candidateTag.list; flags map only to documented request body fields. Opaque syncToken checkpointing is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --include-archived
    communication-template list - List all enabled communication templates. Requires the hiringProcessMetadataRead permission. [intent=etl availability=implemented stream=communication_template_list]; notes: Fixed Ashby stream for communicationTemplate.list; flags map only to documented request body fields. Incremental execution is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.
    feedback-form-definition list - Full-refresh-only Ashby feedback-form-definition list read. Opaque syncToken checkpointing is unavailable pending ashby-sync-token-checkpoint-foundation. [intent=etl availability=implemented stream=feedback_form_definition_list]; notes: Fixed Ashby stream for feedbackFormDefinition.list; flags map only to documented request body fields. Opaque syncToken checkpointing is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --include-archived
    feedback-form-definition info - Returns a single feedback form by id Requires the hiringProcessMetadataRead permission. [intent=etl availability=implemented stream=feedback_form_definition_info]; notes: Fixed Ashby stream for feedbackFormDefinition.info; flags map only to documented request body fields.; flags: --feedback-form-definition-id (required, non-empty)
    job-posting list - Lists published job postings. By default, only published job postings are returned. [intent=etl availability=implemented stream=job_posting_list]; notes: Fixed Ashby stream for jobPosting.list; flags map only to documented request body fields. Incremental execution is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --location (non-empty), --department (non-empty), --listed-only, --include-unpublished-job-postings, --job-board-id (non-empty)
    job-posting info - Retrieve an individual job posting. Set includeUnpublishedJobPostings to true when fetching an unpublished (draft) job posting. [intent=etl availability=implemented stream=job_posting_info]; notes: Fixed Ashby stream for jobPosting.info; flags map only to documented request body fields. Repeatable array request variants (--expand) are blocked pending connector-stream-repeatable-array-foundation. Incremental execution is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --job-posting-id (required, non-empty), --job-board-id (non-empty), --include-unpublished-job-postings
    job-posting update-compensation - Updates compensation for an existing job posting. Set includeUnpublishedJobPostings to true when updating an unpublished (draft) job posting. [intent=reverse_etl availability=implemented write=update_job_posting_compensation]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby jobPosting.updateCompensation through the documented POST /jobPosting.updateCompensation endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --job-posting-id (non-empty), --compensation-tiers-0-components-0-compensation-type, --compensation-tiers-0-components-0-interval, --include-unpublished-job-postings
    job-posting update - Updates an existing job posting. Set includeUnpublishedJobPostings to true when updating an unpublished (draft) job posting. [intent=reverse_etl availability=implemented write=update_job_posting]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby jobPosting.update through the documented POST /jobPosting.update endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --job-posting-id (non-empty), --title (non-empty), --workplace-type, --suppress-description-opening, --suppress-description-closing, --application-confirmation-email-template-id (non-empty), --include-unpublished-job-postings
    job-posting set-status - Sets the status of a job posting. Use this to publish or unpublish a job posting. Set status to Published to publish a draft job posting. [intent=reverse_etl availability=implemented write=set_job_posting_status]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby jobPosting.setStatus through the documented POST /jobPosting.setStatus endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --job-posting-id (non-empty), --status
    job info - Fetches details of a single job by id. Requires the jobsRead permission. [intent=etl availability=implemented stream=job_info]; notes: Fixed Ashby stream for job.info; flags map only to documented request body fields. Repeatable array request variants (--expand) are blocked pending connector-stream-repeatable-array-foundation. Incremental execution is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --id (required, non-empty), --include-unpublished-job-postings-ids
    job create - Creates a new job. Requires the jobsWrite permission. [intent=reverse_etl availability=implemented write=create_job]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby job.create through the documented POST /job.create endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --title (non-empty), --team-id (non-empty), --location-id (non-empty), --default-interview-plan-id (non-empty), --job-template-id (non-empty), --employment-type, --brand-id (non-empty)
    job update - Updates an existing job. At least one field other than jobId must be supplied. Requires the jobsWrite permission. [intent=reverse_etl availability=implemented write=update_job]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby job.update through the documented POST /job.update endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --job-id (non-empty), --title (non-empty), --team-id (non-empty), --location-id (non-empty), --default-interview-plan-id (non-empty), --employment-type, --custom-requisition-id (non-empty)
    job update-compensation - Replaces the compensation tiers on a job. Pass an empty array to clear existing compensation. [intent=reverse_etl availability=implemented write=update_job_compensation]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby job.updateCompensation through the documented POST /job.updateCompensation endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --job-id (non-empty), --compensation-tiers-0-components-0-compensation-type, --compensation-tiers-0-components-0-interval
    job search - Searches jobs by title or custom requisition id. At least one of title or requisitionId must be provided. [intent=direct_read availability=implemented operation=ashby.direct.job.search]; approval: none; risk: bounded JSON direct read; credential-marked response fields are redacted, and non-credential identity fields remain complete in trusted live local output; notes: Fixed Ashby POST direct read; no raw method/path/body override is exposed.; flags: --title (non-empty), --requisition-id (non-empty), --page, --page-cursor
    job set-status - Sets the status of a job. Requires the jobsWrite permission. [intent=reverse_etl availability=implemented write=set_job_status]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby job.setStatus through the documented POST /job.setStatus endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --job-id (non-empty), --status
    job-board list - List all enabled job boards. Requires the jobsRead permission. [intent=etl availability=implemented stream=job_board_list]; notes: Fixed Ashby stream for jobBoard.list; flags map only to documented request body fields.
    job-interview-plan info - Returns a job's interview plan, including activities and interviews that need to be scheduled at each stage. [intent=etl availability=implemented stream=job_interview_plan_info]; notes: Fixed Ashby stream for jobInterviewPlan.info; flags map only to documented request body fields.; flags: --job-id (required, non-empty)
    job-template list - Full-refresh-only Ashby job-template list read. Opaque syncToken checkpointing is unavailable pending ashby-sync-token-checkpoint-foundation. [intent=etl availability=implemented stream=job_template_list]; notes: Fixed Ashby stream for jobTemplate.list; flags map only to documented request body fields. Repeatable array request variants (--expand) are blocked pending connector-stream-repeatable-array-foundation. Opaque syncToken checkpointing is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.
    department archive - Archives a department. Requires the organizationWrite permission. [intent=reverse_etl availability=implemented write=archive_department]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby department.archive through the documented POST /department.archive endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --department-id (non-empty)
    department restore - Restores a department. Requires the organizationWrite permission. [intent=reverse_etl availability=implemented write=restore_department]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby department.restore through the documented POST /department.restore endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --department-id (non-empty)
    department create - Creates a department. Requires the organizationWrite permission. [intent=reverse_etl availability=implemented write=create_department]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby department.create through the documented POST /department.create endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --name (non-empty), --external-name (non-empty), --parent-id (non-empty)
    department info - Fetch department details by id. Requires the organizationRead permission. [intent=etl availability=implemented stream=department_info]; notes: Fixed Ashby stream for department.info; flags map only to documented request body fields. Incremental execution is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --department-id (required, non-empty)
    department list - Full-refresh-only Ashby department list read. Opaque syncToken checkpointing is unavailable pending ashby-sync-token-checkpoint-foundation. [intent=etl availability=implemented stream=department_list]; notes: Fixed Ashby stream for department.list; flags map only to documented request body fields. Opaque syncToken checkpointing is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --include-archived
    department move - Moves a department to another parent. Requires the organizationWrite permission. [intent=reverse_etl availability=implemented write=move_department]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby department.move through the documented POST /department.move endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --department-id (non-empty), --parent-id (non-empty)
    department update - Updates a department. Requires the organizationWrite permission. [intent=reverse_etl availability=implemented write=update_department]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby department.update through the documented POST /department.update endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --department-id (non-empty), --name (non-empty), --external-name (non-empty)
    location archive - Archives a location or location hierarchy. Requires the organizationWrite permission. [intent=reverse_etl availability=implemented write=archive_location]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby location.archive through the documented POST /location.archive endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --location-id (non-empty)
    location create - Creates a location or location hierarchy. Requires the organizationWrite permission. [intent=reverse_etl availability=implemented write=create_location]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby location.create through the documented POST /location.create endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --name (non-empty), --type, --parent-location-id (non-empty), --is-remote, --workplace-type, --external-name (non-empty)
    location restore - Restores an archived location or location hierarchy. Requires the organizationWrite permission. [intent=reverse_etl availability=implemented write=restore_location]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby location.restore through the documented POST /location.restore endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --location-id (non-empty)
    location list - Full-refresh-only Ashby location list read. Opaque syncToken checkpointing is unavailable pending ashby-sync-token-checkpoint-foundation. [intent=etl availability=implemented stream=location_list]; notes: Fixed Ashby stream for location.list; flags map only to documented request body fields. Opaque syncToken checkpointing is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --include-archived, --include-location-hierarchy
    location info - Gets details for a single location by id. Requires the organizationRead permission. [intent=etl availability=implemented stream=location_info]; notes: Fixed Ashby stream for location.info; flags map only to documented request body fields.; flags: --location-id (required, non-empty)
    location move - Moves a location in location hierarchy. Requires the organizationWrite permission. [intent=reverse_etl availability=implemented write=move_location]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby location.move through the documented POST /location.move endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --location-id (non-empty), --parent-location-hierarchy-id (non-empty)
    location update-address - Update an address of a location or location hierarchy. Requires the organizationWrite permission. [intent=reverse_etl availability=implemented write=update_location_address]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby location.updateAddress through the documented POST /location.updateAddress endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --location-id (non-empty)
    location update-name - Update location's name. Requires the organizationWrite permission. [intent=reverse_etl availability=implemented write=update_location_name]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby location.updateName through the documented POST /location.updateName endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --location-id (non-empty), --name (non-empty)
    location update-workplace-type - Update location's workplace type. Requires the organizationWrite permission. [intent=reverse_etl availability=implemented write=update_location_workplace_type]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby location.updateWorkplaceType through the documented POST /location.updateWorkplaceType endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --location-id (non-empty), --workplace-type
    location update-external-name - Update a location's external (candidate-facing) name. Requires the organizationWrite permission. [intent=reverse_etl availability=implemented write=update_location_external_name]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby location.updateExternalName through the documented POST /location.updateExternalName endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --location-id (non-empty), --external-name (non-empty)
    interview-plan list - Full-refresh-only Ashby interview-plan list read. Opaque syncToken checkpointing is unavailable pending ashby-sync-token-checkpoint-foundation. [intent=etl availability=implemented stream=interview_plan_list]; notes: Fixed Ashby stream for interviewPlan.list; flags map only to documented request body fields. Opaque syncToken checkpointing is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --include-archived
    interview-stage list - List all interview stages for an interview plan in order. Requires the interviewsRead permission. [intent=etl availability=implemented stream=interview_stage_list]; notes: Fixed Ashby stream for interviewStage.list; flags map only to documented request body fields.; flags: --interview-plan-id (required, non-empty)
    interview-stage-group list - List all interview stage groups in the organization in order. Requires the interviewsRead permission. [intent=etl availability=implemented stream=interview_stage_group_list]; notes: Fixed Ashby stream for interviewStageGroup.list; flags map only to documented request body fields.
    notetaker-transcript info - Fetches metadata and a pre-signed download URL for an AI Notetaker transcript recording. [intent=direct_read availability=implemented operation=ashby.direct.notetaker.transcript.info]; approval: none; risk: bounded JSON direct read; credential-marked response fields are redacted, and Ashby signed URL fields are preserved (results.url/results.transcriptUrl) in trusted live local output; notes: Fixed Ashby POST direct read; no raw method/path/body override is exposed.; flags: --notetaker-transcript-id (required, non-empty), --page, --page-cursor
    offer approve - Approves an offer or a specific approval step within an offer's approval process. [intent=reverse_etl availability=implemented write=approve_offer]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby offer.approve through the documented POST /offer.approve endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --offer-version-id (non-empty), --approval-step-id (non-empty), --user-id (non-empty), --exclude-form-definition
    offer create - Creates a new Offer Offer forms support a variety of field types. [intent=reverse_etl availability=partial write=create_offer]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby offer.create through the documented POST /offer.create endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed. This command has nested object/array requirements that are implemented by the reverse-ETL action schema but are not fully expressible as scalar CLI flags; use file/warehouse reverse-ETL inputs for execution.; flags: --offer-process-id (non-empty), --offer-form-id (non-empty), --offer-form-field-submissions-0-path (non-empty), --exclude-form-definition
    offer list - Full-refresh-only Ashby offer list read. Opaque syncToken checkpointing is unavailable pending ashby-sync-token-checkpoint-foundation. [intent=etl availability=implemented stream=offer_list]; notes: Fixed Ashby stream for offer.list; flags map only to documented request body fields. Repeatable array request variants (--offer-status, --acceptance-status, --approval-status) are blocked pending connector-stream-repeatable-array-foundation. Opaque syncToken checkpointing is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --created-after, --application-id (non-empty)
    offer info - Returns details about a single offer by id Requires the offersRead permission. [intent=etl availability=implemented stream=offer_info]; notes: Fixed Ashby stream for offer.info; flags map only to documented request body fields. Incremental execution is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --offer-id (required, non-empty), --exclude-form-definition
    offer start - The offer.start endpoint creates and returns an offer version instance that can be filled out and submitted using the offer.create endpoint. [intent=reverse_etl availability=implemented write=start_offer]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby offer.start through the documented POST /offer.start endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --offer-process-id (non-empty)
    offer start-approval-process - Starts the approval process for an offer in a "WaitingOnApprovalStart" state. Once started, the approval is sent to the configured approvers. [intent=reverse_etl availability=implemented write=start_offer_approval_process]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby offer.startApprovalProcess through the documented POST /offer.startApprovalProcess endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --offer-version-id (non-empty), --note (non-empty), --exclude-form-definition
    offer update - Updates an existing Offer Offer forms support a variety of field types. [intent=reverse_etl availability=partial write=update_offer]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby offer.update through the documented POST /offer.update endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed. This command has nested object/array requirements that are implemented by the reverse-ETL action schema but are not fully expressible as scalar CLI flags; use file/warehouse reverse-ETL inputs for execution.; flags: --offer-id (non-empty), --offer-form-field-submissions-0-path (non-empty), --exclude-form-definition
    offer set-status - Updates an offer's acceptance status. Ashby derives the offer status from the provided acceptance status; offerStatus can't be set independently. [intent=reverse_etl availability=implemented write=set_offer_status]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby offer.setStatus through the documented POST /offer.setStatus endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --offer-id (non-empty), --acceptance-status, --exclude-form-definition
    offer set-decided-at - Updates an offer's decidedAt timestamp. Requires the offersWrite permission. [intent=reverse_etl availability=implemented write=set_offer_decided_at]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby offer.setDecidedAt through the documented POST /offer.setDecidedAt endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --offer-id (non-empty), --decided-at (non-empty, format=date-time), --exclude-form-definition
    offer-process start - Starts an offer process for a candidate. Requires the offersWrite permission. [intent=reverse_etl availability=implemented write=start_offer_process]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby offerProcess.start through the documented POST /offerProcess.start endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --application-id (non-empty)
    opening info - Retrieves an opening by its UUID. Requires the jobsRead permission. [intent=etl availability=implemented stream=opening_info]; notes: Fixed Ashby stream for opening.info; flags map only to documented request body fields. Incremental execution is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --opening-id (required, non-empty)
    opening create - Creates an opening. To set values of custom fields on Openings, use the customField.setValue endpoint. [intent=reverse_etl availability=implemented write=create_opening]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby opening.create through the documented POST /opening.create endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --identifier (non-empty), --description (non-empty), --team-id (non-empty), --location-ids, --job-ids, --target-hire-date (non-empty, format=date-time), --target-start-date (non-empty, format=date-time), --is-backfill, --employment-type, --opening-state
    opening update - Updates an opening. To set values for custom fields on Openings, use the customField.setValue endpoint. [intent=reverse_etl availability=implemented write=update_opening]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby opening.update through the documented POST /opening.update endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --opening-id (non-empty), --identifier (non-empty), --description (non-empty), --team-id (non-empty), --target-hire-date (non-empty, format=date-time), --target-start-date (non-empty, format=date-time), --is-backfill, --employment-type
    opening set-archived - Sets the archived state of an opening. Requires the jobsWrite permission. [intent=reverse_etl availability=implemented write=set_opening_archived]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby opening.setArchived through the documented POST /opening.setArchived endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --opening-id (non-empty), --archive
    opening set-opening-state - Sets the state of an opening. Requires the jobsWrite permission. [intent=reverse_etl availability=implemented write=set_opening_opening_state]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby opening.setOpeningState through the documented POST /opening.setOpeningState endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --opening-id (non-empty), --opening-state, --close-reason-id (non-empty)
    opening add-job - Adds a job to an opening. Requires the jobsWrite permission. [intent=reverse_etl availability=implemented write=add_opening_job]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby opening.addJob through the documented POST /opening.addJob endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --opening-id (non-empty), --job-id (non-empty)
    opening remove-job - Removes a job from an opening. Requires the jobsWrite permission. [intent=reverse_etl availability=implemented write=remove_opening_job]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby opening.removeJob through the documented POST /opening.removeJob endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --opening-id (non-empty), --job-id (non-empty)
    opening add-location - Adds a location to an opening. Requires the jobsWrite permission. [intent=reverse_etl availability=implemented write=add_opening_location]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby opening.addLocation through the documented POST /opening.addLocation endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --opening-id (non-empty), --location-id (non-empty)
    opening remove-location - Removes a location from an opening. Requires the jobsWrite permission. [intent=reverse_etl availability=implemented write=remove_opening_location]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby opening.removeLocation through the documented POST /opening.removeLocation endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --opening-id (non-empty), --location-id (non-empty)
    opening list - Full-refresh-only Ashby opening list read. Opaque syncToken checkpointing is unavailable pending ashby-sync-token-checkpoint-foundation. [intent=etl availability=implemented stream=opening_list]; notes: Fixed Ashby stream for opening.list; flags map only to documented request body fields. Opaque syncToken checkpointing is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --created-after
    opening search - Searches for openings by identifier. Requires the jobsRead permission. [intent=direct_read availability=implemented operation=ashby.direct.opening.search]; approval: none; risk: bounded JSON direct read; credential-marked response fields are redacted, and non-credential identity fields remain complete in trusted live local output; notes: Fixed Ashby POST direct read; no raw method/path/body override is exposed.; flags: --identifier (required, non-empty), --page, --page-cursor
    project info - Retrieves a project by its UUID. Requires the candidatesRead permission. [intent=etl availability=implemented stream=project_info]; notes: Fixed Ashby stream for project.info; flags map only to documented request body fields. Incremental execution is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --project-id (required, non-empty)
    project list - Full-refresh-only Ashby project list read. Opaque syncToken checkpointing is unavailable pending ashby-sync-token-checkpoint-foundation. [intent=etl availability=implemented stream=project_list]; notes: Fixed Ashby stream for project.list; flags map only to documented request body fields. Opaque syncToken checkpointing is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --created-after
    project search - Search for projects by title. Responses are limited to 100 results. [intent=direct_read availability=implemented operation=ashby.direct.project.search]; approval: none; risk: bounded JSON direct read; credential-marked response fields are redacted, and non-credential identity fields remain complete in trusted live local output; notes: Fixed Ashby POST direct read; no raw method/path/body override is exposed.; flags: --title (required, non-empty), --page, --page-cursor
    source list - List all sources Requires the hiringProcessMetadataRead permission. [intent=etl availability=implemented stream=source_list]; notes: Fixed Ashby stream for source.list; flags map only to documented request body fields.; flags: --include-archived
    source-tracking-link list - Full-refresh-only Ashby source-tracking-link list read. Opaque syncToken checkpointing is unavailable pending ashby-sync-token-checkpoint-foundation. [intent=etl availability=implemented stream=source_tracking_link_list]; notes: Fixed Ashby stream for sourceTrackingLink.list; flags map only to documented request body fields. Opaque syncToken checkpointing is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --include-disabled, --source-id (non-empty)
    archive-reason list - Lists archive reasons. Requires the hiringProcessMetadataRead permission. [intent=etl availability=implemented stream=archive_reason_list]; notes: Fixed Ashby stream for archiveReason.list; flags map only to documented request body fields.; flags: --include-archived
    brand list - Lists all brands for the organization. Requires the organizationRead permission. [intent=etl availability=implemented stream=brand_list]; notes: Fixed Ashby stream for brand.list; flags map only to documented request body fields.
    custom-field list - Full-refresh-only Ashby custom-field list read. Opaque syncToken checkpointing is unavailable pending ashby-sync-token-checkpoint-foundation. [intent=etl availability=implemented stream=custom_field_list]; notes: Fixed Ashby stream for customField.list; flags map only to documented request body fields. Opaque syncToken checkpointing is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --include-archived
    custom-field create - Create a new custom field. Requires the hiringProcessMetadataWrite permission. [intent=reverse_etl availability=implemented write=create_custom_field]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby customField.create through the documented POST /customField.create endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --field-type, --object-type, --title (non-empty), --description (non-empty), --selectable-values, --is-date-only-field, --is-exposable-to-candidate, --is-private
    custom-field set-value - Set the value of a custom field for a given object. [intent=reverse_etl availability=partial write=set_custom_field_value]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby customField.setValue through the documented POST /customField.setValue endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed. This command has a documented fieldValue union/nested input that is implemented by the reverse-ETL action schema but is not safely expressible as scalar CLI flags; use file/warehouse reverse-ETL inputs for execution.; flags: --object-id (non-empty), --object-type, --field-id (non-empty)
    custom-field info - Get information about a custom field. Requires the hiringProcessMetadataRead permission. [intent=etl availability=implemented stream=custom_field_info]; notes: Fixed Ashby stream for customField.info; flags map only to documented request body fields.; flags: --custom-field-id (required, non-empty)
    custom-field set-values - Set the values of multiple custom fields for a given object in a single call. [intent=reverse_etl availability=partial write=set_custom_field_values]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby customField.setValues through the documented POST /customField.setValues endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed. This command has a documented fieldValue union/nested input that is implemented by the reverse-ETL action schema but is not safely expressible as scalar CLI flags; use file/warehouse reverse-ETL inputs for execution.; flags: --object-id (non-empty), --object-type, --values-0-field-id (non-empty)
    custom-field update-selectable-values - Update the selectable values for a custom field. This endpoint merges the provided selectable values with the existing values for a custom field. [intent=reverse_etl availability=implemented write=update_custom_field_selectable_values]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby customField.updateSelectableValues through the documented POST /customField.updateSelectableValues endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --custom-field-id (non-empty), --selectable-values-0-label (non-empty), --selectable-values-0-value (non-empty)
    assessment update - Update Ashby about the status of a started assessment. assessment_status is required unless cancelled_reason is provided. [intent=reverse_etl availability=implemented write=update_assessment]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby assessment.update through the documented POST /assessment.update endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --assessment-id (non-empty), --timestamp, --metadata
    assessment add-completed-to-candidate - Add a completed assessment to a candidate. Requires the candidatesWrite permission. [intent=reverse_etl availability=implemented write=add_assessment_completed_to_candidate]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby assessment.addCompletedToCandidate through the documented POST /assessment.addCompletedToCandidate endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --candidate-id (non-empty), --partner-id (non-empty), --assessment-assessment-type-id (non-empty), --assessment-assessment-id (non-empty), --assessment-assessment-name (non-empty), --assessment-result-identifier (non-empty), --assessment-result-label (non-empty), --assessment-result-type, --assessment-result-value, --assessment-metadata-0-identifier (non-empty), --assessment-metadata-0-label (non-empty), --assessment-metadata-0-type, --assessment-metadata-0-value, --timestamp
    hiring-team add-member - Adds an Ashby user to the hiring team at the application, job, or opening level. [intent=reverse_etl availability=implemented write=add_hiring_team_member]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby hiringTeam.addMember through the documented POST /hiringTeam.addMember endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Provide an applicationId, jobId, or openingId target. Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --application-id (non-empty), --job-id (non-empty), --opening-id (non-empty), --team-member-id (non-empty), --role-id (non-empty)
    hiring-team remove-member - Removes an Ashby user from the hiring team at the application, job, or opening level. [intent=reverse_etl availability=implemented write=remove_hiring_team_member]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby hiringTeam.removeMember through the documented POST /hiringTeam.removeMember endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Provide an applicationId, jobId, or openingId target. Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --application-id (non-empty), --job-id (non-empty), --opening-id (non-empty), --team-member-id (non-empty), --role-id (non-empty)
    hiring-team-role list - Lists the possible hiring team roles in an organization Requires the organizationRead permission. [intent=etl availability=implemented stream=hiring_team_role_list]; notes: Fixed Ashby stream for hiringTeamRole.list; defaults to namesOnly=true role-title results. namesOnly=false object results are blocked pending variant-schema foundation ashby_hiring_team_role_list_names_only_false.
    user info - Retrieves detailed information about a specific user by their ID. Requires the organizationRead permission. [intent=etl availability=implemented stream=user_info]; notes: Fixed Ashby stream for user.info; flags map only to documented request body fields. Incremental execution is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --user-id (required, non-empty)
    user search - Searches for users by email address. Returns an array containing the user if found, or an empty array if no user with the given email exists. [intent=direct_read availability=implemented operation=ashby.direct.user.search]; approval: none; risk: bounded JSON direct read; credential-marked response fields are redacted, and non-credential identity fields remain complete in trusted live local output; notes: Fixed Ashby POST direct read; no raw method/path/body override is exposed.; flags: --email (required, non-empty), --page, --page-cursor
    user interviewer-settings - Get interviewer settings for a user. Requires the organizationRead permission. [intent=direct_read availability=implemented operation=ashby.direct.user.interviewer.settings]; approval: none; risk: bounded JSON direct read; credential-marked response fields are redacted, and non-credential interviewer settings remain complete in trusted live local output; notes: Fixed Ashby POST direct read; no raw method/path/body override is exposed.; flags: --user-id (required, non-empty), --page, --page-cursor
    user update-interviewer-settings - Update interviewer settings for a user. Either limit can be provided, or both can be provided. If only one is provided, the other will remain unchanged. [intent=reverse_etl availability=implemented write=update_user_interviewer_settings]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby user.updateInterviewerSettings through the documented POST /user.updateInterviewerSettings endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --user-id (non-empty), --daily-limit, --weekly-limit
    user create-interviewer-pause - Creates an interviewer pause for a user. While paused, the user will not be scheduled for interviews. [intent=reverse_etl availability=implemented write=create_user_interviewer_pause]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby user.createInterviewerPause through the documented POST /user.createInterviewerPause endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --user-id (non-empty), --starts-at (non-empty, format=date-time), --ends-at (non-empty, format=date-time), --comment (non-empty)
    user list-interviewer-pauses - Lists all active or scheduled interviewer pauses for a user. [intent=etl availability=implemented stream=user_list_interviewer_pauses]; notes: Fixed Ashby stream for user.listInterviewerPauses; flags map only to documented request body fields. Incremental execution is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --user-id (required, non-empty)
    user delete-interviewer-pause - Deletes an interviewer pause. Requires the organizationWrite permission. [intent=reverse_etl availability=implemented write=delete_user_interviewer_pause]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby user.deleteInterviewerPause through the documented POST /user.deleteInterviewerPause endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --interviewer-pause-id (non-empty)
    user set-custom-field-value - Set the value of a custom field on an employee. The values accepted in the fieldValue param depend on the type of field being updated. [intent=reverse_etl availability=partial write=set_user_custom_field_value]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby user.setCustomFieldValue through the documented POST /user.setCustomFieldValue endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed. This command has a documented fieldValue union/nested input that is implemented by the reverse-ETL action schema but is not safely expressible as scalar CLI flags; use file/warehouse reverse-ETL inputs for execution.; flags: --user-id (non-empty), --field-id (non-empty)
    user set-custom-field-values - Set the values of multiple custom fields on an employee in a single call. [intent=reverse_etl availability=partial write=set_user_custom_field_values]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby user.setCustomFieldValues through the documented POST /user.setCustomFieldValues endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed. This command has a documented fieldValue union/nested input that is implemented by the reverse-ETL action schema but is not safely expressible as scalar CLI flags; use file/warehouse reverse-ETL inputs for execution.; flags: --user-id (non-empty), --values-0-field-id (non-empty)
    referral create - Creates a referral Requires the candidatesWrite permission. [intent=reverse_etl availability=partial write=create_referral]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby referral.create through the documented POST /referral.create endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed. This command has nested object/array requirements that are implemented by the reverse-ETL action schema but are not fully expressible as scalar CLI flags; use file/warehouse reverse-ETL inputs for execution.; flags: --id (non-empty), --credited-to-user-id (non-empty), --field-submissions-0-path (non-empty), --created-at (non-empty, format=date-time)
    email-sender list - Beta This endpoint is in beta and may not be available for all organizations. [intent=etl availability=implemented stream=email_sender_list]; notes: Fixed Ashby stream for emailSender.list; flags map only to documented request body fields.
    sequence add - Beta This endpoint is in beta and may not be available for all organizations. Enrolls a candidate in a reusable sourcing sequence. [intent=reverse_etl availability=implemented write=add_sequence]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby sequence.add through the documented POST /sequence.add endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --sequence-template-id (non-empty), --candidate-id (non-empty), --start, --application-id (non-empty)
    sequence cancel - Cancels a running sourcing sequence (campaign) for a candidate. Requires the sourcingWrite permission. [intent=reverse_etl availability=implemented write=cancel_sequence]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby sequence.cancel through the documented POST /sequence.cancel endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --sequence-id (non-empty), --reason
    sequence discard - Beta This endpoint is in beta and may not be available for all organizations. [intent=reverse_etl availability=implemented write=discard_sequence]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby sequence.discard through the documented POST /sequence.discard endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --sequence-id (non-empty)
    sequence info - Beta This endpoint is in beta and may not be available for all organizations. [intent=etl availability=implemented stream=sequence_info]; notes: Fixed Ashby stream for sequence.info; flags map only to documented request body fields. Incremental execution is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --sequence-id (required, non-empty)
    sequence list - Full-refresh-only Ashby sequence list read. Opaque syncToken checkpointing is unavailable pending ashby-sync-token-checkpoint-foundation. [intent=etl availability=implemented stream=sequence_list]; notes: Fixed Ashby stream for sequence.list; flags map only to documented request body fields. Opaque syncToken checkpointing is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --candidate-id (non-empty)
    sequence update-stage - Beta This endpoint is in beta and may not be available for all organizations. [intent=reverse_etl availability=implemented write=update_sequence_stage]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby sequence.updateStage through the documented POST /sequence.updateStage endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --sequence-id (non-empty), --stage-id (non-empty), --subject (non-empty), --body-html (non-empty)
    sequence start - Beta This endpoint is in beta and may not be available for all organizations. [intent=reverse_etl availability=implemented write=start_sequence]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby sequence.start through the documented POST /sequence.start endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --sequence-id (non-empty)
    sequence-template info - Retrieves metadata for a reusable sourcing sequence template visible to the caller. Archived templates may be returned. [intent=etl availability=implemented stream=sequence_template_info]; notes: Fixed Ashby stream for sequenceTemplate.info; flags map only to documented request body fields. Incremental execution is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --sequence-template-id (required, non-empty)
    sequence-template list - Full-refresh-only Ashby sequence-template list read. Opaque syncToken checkpointing is unavailable pending ashby-sync-token-checkpoint-foundation. [intent=etl availability=implemented stream=sequence_template_list]; notes: Fixed Ashby stream for sequenceTemplate.list; flags map only to documented request body fields. Opaque syncToken checkpointing is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --include-archived
    interview-schedule list - Full-refresh-only Ashby interview-schedule list read. Opaque syncToken checkpointing is unavailable pending ashby-sync-token-checkpoint-foundation. [intent=etl availability=implemented stream=interview_schedule_list]; notes: Fixed Ashby stream for interviewSchedule.list; flags map only to documented request body fields. Opaque syncToken checkpointing is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --application-id (non-empty), --interview-stage-id (non-empty), --created-after
    interview-schedule create - Create a scheduled interview in Ashby. Requires the interviewsWrite permission. [intent=reverse_etl availability=implemented write=create_interview_schedule]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby interviewSchedule.create through the documented POST /interviewSchedule.create endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --application-id (non-empty), --interview-events-0-start-time (non-empty), --interview-events-0-end-time (non-empty), --interview-events-0-interviewers-0-email (non-empty)
    interview-schedule update - Update an interview schedule. This endpoint allows you to add, cancel, or update interview events associated with an interview schedule. [intent=reverse_etl availability=implemented write=update_interview_schedule]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby interviewSchedule.update through the documented POST /interviewSchedule.update endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --interview-schedule-id (non-empty), --interview-event-id-to-cancel (non-empty), --allow-feedback-deletion
    interview-schedule cancel - Cancel an interview schedule by id. Requires the interviewsWrite permission. [intent=reverse_etl availability=implemented write=cancel_interview_schedule]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby interviewSchedule.cancel through the documented POST /interviewSchedule.cancel endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --id (non-empty), --allow-reschedule
    take-home-assignment list - Full-refresh-only Ashby take-home-assignment list read. Opaque syncToken checkpointing is unavailable pending ashby-sync-token-checkpoint-foundation. [intent=etl availability=implemented stream=take_home_assignment_list]; notes: Fixed Ashby stream for takeHomeAssignment.list; flags map only to documented request body fields. Repeatable array request variants (--expand) are blocked pending connector-stream-repeatable-array-foundation. Opaque syncToken checkpointing is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --application-id (non-empty), --candidate-id (non-empty)
    take-home-assignment info - Beta This endpoint is in beta and may not be available for all organizations. [intent=etl availability=implemented stream=take_home_assignment_info]; notes: Fixed Ashby stream for takeHomeAssignment.info; flags map only to documented request body fields. Repeatable array request variants (--expand) are blocked pending connector-stream-repeatable-array-foundation. Incremental execution is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --take-home-assignment-id (required, non-empty)
    interview-event list - Full-refresh-only Ashby interview-event list read. Opaque syncToken checkpointing is unavailable pending ashby-sync-token-checkpoint-foundation. [intent=etl availability=implemented stream=interview_event_list]; notes: Fixed Ashby stream for interviewEvent.list; flags map only to documented request body fields. Repeatable array request variants (--expand) are blocked pending connector-stream-repeatable-array-foundation. Opaque syncToken checkpointing is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --interview-schedule-id (required, non-empty), --created-after
    interview-briefing info - Fetch the briefing data for an interview event. [intent=etl availability=implemented stream=interview_briefing_info]; notes: Fixed Ashby stream for interviewBriefing.info; flags map only to documented request body fields. Repeatable array request variants (--expand) are blocked pending connector-stream-repeatable-array-foundation.; flags: --interview-event-id (required, non-empty)
    interview info - Fetch interview details by id. Requires the interviewsRead permission. [intent=etl availability=implemented stream=interview_info]; notes: Fixed Ashby stream for interview.info; flags map only to documented request body fields.; flags: --id (required, non-empty)
    interview list - Full-refresh-only Ashby interview list read. Opaque syncToken checkpointing is unavailable pending ashby-sync-token-checkpoint-foundation. [intent=etl availability=implemented stream=interview_list]; notes: Fixed Ashby stream for interview.list; flags map only to documented request body fields. Opaque syncToken checkpointing is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --include-archived, --include-non-shared-interviews, --exclude-archived-schedule-template-interviews
    interview-stage info - Retrieves detailed information about a specific interview stage by its ID. [intent=etl availability=implemented stream=interview_stage_info]; notes: Fixed Ashby stream for interviewStage.info; flags map only to documented request body fields.; flags: --interview-stage-id (required, non-empty)
    file info - Retrieve the URL for a file referenced by a public API file handle (candidate files, resumes, offer letters, and signature-request files). [intent=direct_read availability=implemented operation=ashby.direct.file.info]; approval: none; risk: bounded JSON direct read; credential-marked response fields are redacted, and Ashby signed URL fields are preserved (results.url/results.transcriptUrl) in trusted live local output; notes: Fixed Ashby POST direct read; no raw method/path/body override is exposed.; flags: --file-handle (required, non-empty), --page, --page-cursor
    survey-form-definition info - Returns details about a single survey form definition by id. [intent=etl availability=implemented stream=survey_form_definition_info]; notes: Fixed Ashby stream for surveyFormDefinition.info; flags map only to documented request body fields.; flags: --survey-form-definition-id (required, non-empty)
    survey-form-definition list - Full-refresh-only Ashby survey-form-definition list read. Opaque syncToken checkpointing is unavailable pending ashby-sync-token-checkpoint-foundation. [intent=etl availability=implemented stream=survey_form_definition_list]; notes: Fixed Ashby stream for surveyFormDefinition.list; flags map only to documented request body fields. Opaque syncToken checkpointing is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.
    survey-request create - This endpoint generates a survey request and returns a survey URL. You can send this URL to a candidate to allow them to complete a survey. [intent=reverse_etl availability=implemented write=create_survey_request]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby surveyRequest.create through the documented POST /surveyRequest.create endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --candidate-id (non-empty), --application-id (non-empty), --survey-form-definition-id (non-empty)
    survey-request list - Full-refresh-only Ashby survey-request list read. Opaque syncToken checkpointing is unavailable pending ashby-sync-token-checkpoint-foundation. [intent=etl availability=implemented stream=survey_request_list]; notes: Fixed Ashby stream for surveyRequest.list; flags map only to documented request body fields. Opaque syncToken checkpointing is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --survey-type (required), --application-id (non-empty), --candidate-id (non-empty), --created-after
    survey-submission create - Creates a survey submission for a candidate and application. Requires the candidatesWrite permission. [intent=reverse_etl availability=partial write=create_survey_submission]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby surveySubmission.create through the documented POST /surveySubmission.create endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed. This command has nested object/array requirements that are implemented by the reverse-ETL action schema but are not fully expressible as scalar CLI flags; use file/warehouse reverse-ETL inputs for execution.; flags: --survey-form-definition-id (non-empty), --candidate-id (non-empty), --application-id (non-empty)
    survey-submission list - Full-refresh-only Ashby survey-submission list read. Opaque syncToken checkpointing is unavailable pending ashby-sync-token-checkpoint-foundation. [intent=etl availability=implemented stream=survey_submission_list]; notes: Fixed Ashby stream for surveySubmission.list; flags map only to documented request body fields. Opaque syncToken checkpointing is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --survey-type (required), --created-after
    webhook create - Creates a webhook setting. Requires the apiKeysWrite permission. [intent=reverse_etl availability=implemented write=create_webhook]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby webhook.create through the documented POST /webhook.create endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --webhook-type, --request-url (non-empty), --secret-token (non-empty)
    webhook info - Retrieves information about a specific webhook setting by its ID. Requires the apiKeysRead permission. [intent=etl availability=implemented stream=webhook_info]; notes: Fixed Ashby stream for webhook.info; flags map only to documented request body fields.; flags: --webhook-id (required, non-empty)
    webhook update - Updates a webhook setting. One of enabled, requestUrl, or secretToken must be provided. [intent=reverse_etl availability=implemented write=update_webhook]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby webhook.update through the documented POST /webhook.update endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --webhook-id (non-empty), --enabled, --request-url (non-empty), --secret-token (non-empty)
    webhook delete - Deletes a webhook setting. Requires the apiKeysWrite permission. [intent=reverse_etl availability=implemented write=delete_webhook]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby webhook.delete through the documented POST /webhook.delete endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --webhook-id (non-empty)
    interviewer-pool list - Full-refresh-only Ashby interviewer-pool list read. Opaque syncToken checkpointing is unavailable pending ashby-sync-token-checkpoint-foundation. [intent=etl availability=implemented stream=interviewer_pool_list]; notes: Fixed Ashby stream for interviewerPool.list; flags map only to documented request body fields. Opaque syncToken checkpointing is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --include-archived-pools, --include-archived-training-stages
    interviewer-pool info - Get information about an interviewer pool. Requires the hiringProcessMetadataRead permission. [intent=etl availability=implemented stream=interviewer_pool_info]; notes: Fixed Ashby stream for interviewerPool.info; flags map only to documented request body fields.; flags: --interviewer-pool-id (required, non-empty)
    interviewer-pool archive - Archives an interviewer pool. Requires the hiringProcessMetadataWrite permission. [intent=reverse_etl availability=implemented write=archive_interviewer_pool]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby interviewerPool.archive through the documented POST /interviewerPool.archive endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --interviewer-pool-id (non-empty)
    interviewer-pool restore - Restores an archived interviewer pool. Requires the hiringProcessMetadataWrite permission. [intent=reverse_etl availability=implemented write=restore_interviewer_pool]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby interviewerPool.restore through the documented POST /interviewerPool.restore endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --interviewer-pool-id (non-empty)
    interviewer-pool create - Create an interviewer pool. Requires the hiringProcessMetadataWrite permission. [intent=reverse_etl availability=implemented write=create_interviewer_pool]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby interviewerPool.create through the documented POST /interviewerPool.create endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --title (non-empty), --requires-training
    interviewer-pool update - Update an interviewer pool. Requires the hiringProcessMetadataWrite permission. [intent=reverse_etl availability=implemented write=update_interviewer_pool]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby interviewerPool.update through the documented POST /interviewerPool.update endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --interviewer-pool-id (non-empty), --title (non-empty), --requires-training
    interviewer-pool add-user - Add a user to an interviewer pool. Requires the hiringProcessMetadataWrite permission. [intent=reverse_etl availability=implemented write=add_interviewer_pool_user]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby interviewerPool.addUser through the documented POST /interviewerPool.addUser endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --interviewer-pool-id (non-empty), --user-id (non-empty), --interviewer-pool-training-path-stage-id (non-empty)
    interviewer-pool remove-user - Remove a user from an interviewer pool. Requires the hiringProcessMetadataWrite permission. [intent=reverse_etl availability=implemented write=remove_interviewer_pool_user]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby interviewerPool.removeUser through the documented POST /interviewerPool.removeUser endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --interviewer-pool-id (non-empty), --user-id (non-empty)
    close-reason list - Lists all close reasons for jobs or openings. Requires the hiringProcessMetadataRead permission. [intent=etl availability=implemented stream=close_reason_list]; notes: Fixed Ashby stream for closeReason.list; flags map only to documented request body fields.; flags: --include-archived
    report generate - Start an Ashby report generation or check an existing request. [intent=direct_read availability=implemented operation=ashby.direct.report.generate]; approval: none; risk: bounded JSON direct read that starts or polls a documented Ashby report generation and returns at most 1 MiB of redacted JSON; the connector does not fetch returned report URLs or poll automatically; notes: Fixed Ashby POST direct read; no raw method/path/body override is exposed.; flags: --report-id (required, non-empty), --include-headers-in-data, --result-style, --request-id (non-empty), --page, --page-cursor
    report synchronous - Beta This endpoint is currently in beta and may change without notice. Retrieves report data synchronously. Timeout: 30 seconds. [intent=etl availability=implemented stream=report_synchronous]; notes: Fixed Ashby stream for report.synchronous; flags map only to documented request body fields.; flags: --report-id (required, non-empty), --include-headers-in-data, --result-style
    approval list - Full-refresh-only Ashby approval list read. Opaque syncToken checkpointing is unavailable pending ashby-sync-token-checkpoint-foundation. [intent=etl availability=implemented stream=approval_list]; notes: Fixed Ashby stream for approval.list; flags map only to documented request body fields. Opaque syncToken checkpointing is blocked pending ashby-sync-token-checkpoint-foundation; this stream is full-refresh only.; flags: --entity-type, --entity-id (non-empty)
  Help topics:
    ashby safety - Ashby writes are named, schema-validated actions only; reverse ETL must use plan, preview, explicit approval, and execute.
    ashby parity - Public Ashby OpenAPI coverage ledger is recorded in execution bundle with blocked webhook/partner/binary workflow reasons.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect ashby

  # Inspect as structured JSON
  pm connectors inspect ashby --json

AGENT WORKFLOW
  - Run pm connectors inspect ashby before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
