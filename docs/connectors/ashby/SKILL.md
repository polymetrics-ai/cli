---
name: pm-ashby
description: Ashby connector knowledge and safe action guide.
---

# pm-ashby

## Purpose

Reads Ashby applicant-tracking REST resources and exposes reviewed reverse-ETL/direct-read surfaces from the official Ashby OpenAPI. Fixture-only; not live-certified.

## Icon

- asset: icons/ashby.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.ashbyhq.com/

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- max_pages
- mode
- page_size
- start_date
- api_key (secret)

## ETL Streams

- candidates:
  - primary key: id
  - cursor: updatedAt
  - fields: applicationIds(), company(), createdAt(), creditedToUser(), customFields(), emailAddresses(), fileHandles(), fraudStatus(), id(), location(), name(), phoneNumbers(), position(), primaryEmailAddress(), primaryPhoneNumber(), profileUrl(), resumeFileHandle(), school(), socialLinks(), source(), tags(), timezone(), updatedAt()
- jobs:
  - primary key: id
  - cursor: updatedAt
  - fields: author(), brandId(), closedAt(), compensation(), confidential(), createdAt(), customFields(), customRequisitionId(), defaultInterviewPlanId(), departmentId(), employmentType(), hiringTeam(), id(), interviewPlanIds(), jobPostingIds(), location(), locationId(), openedAt(), openings(), status(), title(), updatedAt()
- applications:
  - primary key: id
  - cursor: updatedAt
  - fields: appliedViaJobPostingId(), archiveReason(), archivedAt(), candidate(), createdAt(), creditedToUser(), currentInterviewStage(), customFields(), hiringTeam(), id(), job(), openings(), source(), status(), submitterClientIp(), submitterUserAgent(), updatedAt()
- users:
  - primary key: id
  - cursor: updatedAt
  - fields: customFields(), email(), firstName(), globalRole(), id(), isEnabled(), lastName(), managerId(), updatedAt()
- api_key_info:
  - primary key: title
  - cursor: createdAt
  - fields: createdAt(), scopes(), title()
- audit_log_list:
  - primary key: id
  - cursor: timestamp
  - fields: actor(), category(), changedFields(), description(), id(), parentAction(), request(), target(), timestamp()
- application_info:
  - primary key: id
  - cursor: updatedAt
  - fields: applicationFormSubmissions(), applicationHistory(), appliedViaJobPostingId(), archiveReason(), archivedAt(), candidate(), createdAt(), creditedToUser(), currentInterviewStage(), customFields(), hiringTeam(), id(), job(), openings(), referrals(), resumeFileHandle(), source(), status(), submitterClientIp(), submitterUserAgent(), updatedAt()
- application_hiring_team_role_list:
  - primary key: id
  - fields: id(), title()
- application_list_history:
  - primary key: id
  - cursor: enteredStageAt
  - fields: actorId(), allowedActions(), enteredStageAt(), id(), leftStageAt(), stageId(), stageNumber(), title()
- application_list_criteria_evaluations:
  - primary key: id
  - cursor: evaluatedAt
  - fields: criterion(), evaluatedAt(), id(), outcome(), outcomeNumber(), reasoning(), skipReason(), status()
- application_feedback_list:
  - primary key: id
  - cursor: submittedAt
  - fields: applicationHistoryId(), applicationId(), creditedToUser(), feedbackFormDefinitionId(), formDefinition(), id(), interviewEventId(), interviewId(), submittedAt(), submittedByUser(), submittedValues()
- candidate_info:
  - primary key: id
  - cursor: updatedAt
  - fields: applicationIds(), company(), createdAt(), creditedToUser(), customFields(), emailAddresses(), fileHandles(), fraudStatus(), id(), location(), name(), phoneNumbers(), position(), primaryEmailAddress(), primaryPhoneNumber(), profileUrl(), resumeFileHandle(), school(), socialLinks(), source(), tags(), timezone(), updatedAt()
- candidate_list_client_info:
  - primary key: id
  - cursor: createdAt
  - fields: candidateId(), createdAt(), id(), ipAddress(), relatedEntityId(), relatedEntityType(), userAgent()
- candidate_list_fraud_checks:
  - primary key: id
  - cursor: createdAt
  - fields: applicationId(), candidateId(), createdAt(), fraudSignals(), id()
- candidate_list_notes:
  - primary key: id
  - cursor: createdAt
  - fields: author(), content(), createdAt(), id(), isPrivate()
- candidate_list_projects:
  - primary key: id
  - cursor: createdAt
  - fields: authorId(), confidential(), createdAt(), customFieldEntries(), description(), id(), isArchived(), title()
- candidate_tag_list:
  - primary key: id
  - fields: id(), isArchived(), title()
- communication_template_list:
  - primary key: id
  - cursor: updatedAt
  - fields: createdAt(), id(), intendedTypes(), title(), updatedAt()
- feedback_form_definition_list:
  - primary key: id
  - fields: formDefinition(), id(), interviewId(), isArchived(), isDefaultForm(), organizationId(), title()
- feedback_form_definition_info:
  - primary key: id
  - fields: formDefinition(), id(), interviewId(), isArchived(), isDefaultForm(), organizationId(), title()
- job_posting_list:
  - primary key: id
  - cursor: updatedAt
  - fields: applicationDeadline(), applyLink(), compensationTierSummary(), departmentName(), employmentType(), externalLink(), id(), isListed(), jobId(), locationIds(), locationName(), publishedDate(), shouldDisplayCompensationOnJobBoard(), status(), teamName(), title(), updatedAt(), workplaceType()
- job_posting_info:
  - primary key: id
  - cursor: updatedAt
  - fields: address(), applicationConfirmationEmailTemplateId(), applicationDeadline(), applicationFormDefinition(), applicationLimitCalloutHtml(), applyLink(), compensation(), departmentName(), descriptionHtml(), descriptionParts(), descriptionPlain(), descriptionSocial(), employmentType(), externalLink(), id(), isListed(), isRemote(), job(), jobId(), linkedData(), locationAddress(), locationIds(), locationName(), publishedDate(), status(), suppressDescriptionClosing(), suppressDescriptionOpening(), surveyFormDefinitions(), teamName(), teamNameHierarchy(), title(), updatedAt(), workplaceType()
- job_info:
  - primary key: id
  - cursor: updatedAt
  - fields: author(), brandId(), closedAt(), compensation(), confidential(), createdAt(), customFields(), customRequisitionId(), defaultInterviewPlanId(), departmentId(), employmentType(), hiringTeam(), id(), interviewPlanIds(), jobPostingIds(), location(), locationId(), openedAt(), openings(), status(), title(), updatedAt()
- job_board_list:
  - primary key: id
  - fields: id(), isInternal(), title()
- job_interview_plan_info:
  - primary key: jobId
  - fields: interviewPlanId(), jobId(), stages()
- job_template_list:
  - primary key: id
  - cursor: updatedAt
  - fields: createdAt(), defaultInterviewPlanId(), departmentId(), id(), interviewPlanIds(), location(), locationId(), status(), title(), updatedAt()
- department_info:
  - primary key: id
  - cursor: updatedAt
  - fields: createdAt(), externalName(), extraData(), id(), isArchived(), name(), parentId(), updatedAt()
- department_list:
  - primary key: id
  - cursor: updatedAt
  - fields: createdAt(), externalName(), extraData(), id(), isArchived(), name(), parentId(), updatedAt()
- location_list:
  - primary key: id
  - fields: address(), externalName(), extraData(), id(), isArchived(), isRemote(), name(), parentLocationId(), type(), workplaceType()
- location_info:
  - primary key: id
  - fields: address(), externalName(), extraData(), id(), isArchived(), isRemote(), name(), parentLocationId(), type(), workplaceType()
- interview_plan_list:
  - primary key: id
  - cursor: updatedAt
  - fields: createdAt(), id(), isArchived(), title(), updatedAt()
- interview_stage_list:
  - primary key: id
  - fields: id(), interviewPlanId(), interviewStageGroupId(), orderInInterviewPlan(), title(), type()
- interview_stage_group_list:
  - primary key: id
  - fields: id(), order(), stageType(), title()
- offer_list:
  - primary key: id
  - cursor: decidedAt
  - fields: acceptanceStatus(), applicationId(), decidedAt(), formDefinition(), id(), latestVersion(), offerStatus(), versions()
- offer_info:
  - primary key: id
  - cursor: decidedAt
  - fields: acceptanceStatus(), applicationId(), decidedAt(), formDefinition(), id(), latestVersion(), offerStatus(), versions()
- opening_info:
  - primary key: id
  - cursor: openedAt
  - fields: archivedAt(), closeReasonId(), closedAt(), id(), isArchived(), latestVersion(), openedAt(), openingState()
- opening_list:
  - primary key: id
  - cursor: openedAt
  - fields: archivedAt(), closeReasonId(), closedAt(), id(), isArchived(), latestVersion(), openedAt(), openingState()
- project_info:
  - primary key: id
  - cursor: createdAt
  - fields: authorId(), confidential(), createdAt(), customFieldEntries(), description(), id(), isArchived(), title()
- project_list:
  - primary key: id
  - cursor: createdAt
  - fields: authorId(), confidential(), createdAt(), customFieldEntries(), description(), id(), isArchived(), title()
- source_list:
  - primary key: id
  - fields: id(), isArchived(), sourceType(), title()
- source_tracking_link_list:
  - primary key: id
  - fields: code(), enabled(), id(), link(), sourceId()
- archive_reason_list:
  - primary key: id
  - fields: id(), isArchived(), reasonType(), text()
- brand_list:
  - primary key: id
  - fields: hostedJobsPageSlug(), id(), name()
- custom_field_list:
  - primary key: id
  - fields: description(), fieldType(), id(), isArchived(), isPrivate(), isRequired(), objectType(), selectableValues(), title()
- custom_field_info:
  - primary key: id
  - fields: description(), fieldType(), id(), isArchived(), isPrivate(), isRequired(), objectType(), selectableValues(), title()
- hiring_team_role_list:
  - primary key: value
  - fields: value()
- user_info:
  - primary key: id
  - cursor: updatedAt
  - fields: customFields(), email(), firstName(), globalRole(), id(), isEnabled(), lastName(), managerId(), updatedAt()
- user_list_interviewer_pauses:
  - primary key: id
  - cursor: createdAt
  - fields: comment(), createdAt(), endsAt(), id(), startsAt(), userId()
- referral_form_info:
  - primary key: id
  - fields: formDefinition(), id(), isArchived(), isDefaultForm(), organizationId(), title()
- email_sender_list:
  - primary key: email
  - fields: displayName(), email(), type()
- sequence_info:
  - primary key: id
  - cursor: createdAt
  - fields: applicationId(), candidateId(), createdAt(), id(), sequenceTemplateId(), stages(), status()
- sequence_list:
  - primary key: id
  - cursor: createdAt
  - fields: applicationId(), candidateId(), createdAt(), id(), sequenceTemplateId(), stages(), status()
- sequence_template_info:
  - primary key: id
  - cursor: updatedAt
  - fields: id(), isArchived(), stages(), title(), unsubscribeLinkActive(), updatedAt()
- sequence_template_list:
  - primary key: id
  - cursor: updatedAt
  - fields: id(), isArchived(), stages(), title(), unsubscribeLinkActive(), updatedAt()
- interview_schedule_list:
  - primary key: id
  - cursor: updatedAt
  - fields: applicationId(), createdAt(), id(), interviewEvents(), interviewStageId(), scheduledBy(), status(), updatedAt()
- take_home_assignment_list:
  - primary key: id
  - cursor: updatedAt
  - fields: applicationId(), candidateId(), createdAt(), feedbackFormDefinitionId(), id(), interview(), interviewId(), interviewStageId(), reviewers(), status(), submission(), updatedAt()
- take_home_assignment_info:
  - primary key: id
  - cursor: updatedAt
  - fields: applicationId(), candidateId(), createdAt(), feedbackFormDefinitionId(), id(), interview(), interviewId(), interviewStageId(), reviewers(), status(), submission(), updatedAt()
- interview_event_list:
  - primary key: id
  - cursor: updatedAt
  - fields: createdAt(), endTime(), extraData(), feedbackLink(), hasSubmittedFeedback(), id(), interview(), interviewId(), interviewScheduleId(), interviewerCalendarEventId(), interviewerUserIds(), interviewers(), location(), meetingLink(), notetakerTranscriptId(), startTime(), updatedAt()
- interview_briefing_info:
  - primary key: id
  - fields: application(), applicationId(), candidate(), feedbackFormDefinition(), feedbackFormDefinitionId(), hasSubmittedFeedback(), id(), interview(), interviewId(), interviewStageId(), interviewers(), job()
- interview_info:
  - primary key: id
  - fields: externalTitle(), feedbackFormDefinitionId(), id(), instructionsHtml(), instructionsPlain(), isArchived(), isDebrief(), isFeedbackRequested(), isFeedbackRequired(), jobId(), title(), type()
- interview_list:
  - primary key: id
  - fields: externalTitle(), feedbackFormDefinitionId(), id(), instructionsHtml(), instructionsPlain(), isArchived(), isDebrief(), isFeedbackRequested(), isFeedbackRequired(), jobId(), title(), type()
- interview_stage_info:
  - primary key: id
  - fields: id(), interviewPlanId(), interviewStageGroupId(), orderInInterviewPlan(), title(), type()
- survey_form_definition_info:
  - primary key: id
  - fields: formDefinition(), id(), isArchived(), surveyType(), title()
- survey_form_definition_list:
  - primary key: id
  - fields: formDefinition(), id(), isArchived(), surveyType(), title()
- survey_request_list:
  - primary key: id
  - fields: applicationId(), candidateId(), id(), surveyFormDefinitionId(), surveyUrl()
- survey_submission_list:
  - primary key: id
  - cursor: submittedAt
  - fields: applicationId(), candidateId(), formDefinition(), id(), submittedAt(), submittedValues(), surveyFormDefinitionId(), surveyType()
- webhook_info:
  - primary key: id
  - fields: enabled(), id(), requestUrl(), webhookType()
- interviewer_pool_list:
  - primary key: id
  - fields: id(), isArchived(), title(), trainingPath()
- interviewer_pool_info:
  - primary key: id
  - fields: id(), isArchived(), qualifiedMembers(), title(), trainees(), trainingPath()
- close_reason_list:
  - primary key: id
  - fields: id(), isArchived(), reasonText()
- report_synchronous:
  - primary key: requestId
  - fields: failureReason(), reportData(), requestId(), status()
- approval_list:
  - primary key: id
  - cursor: createdAt
  - fields: approvalDefinitionId(), completedAt(), createdAt(), entityId(), entityType(), id(), steps(), submittedAt()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_application:
  - endpoint: POST /application.create
  - required fields: candidateId, jobId
  - risk: Executes Ashby application.create through the documented POST /application.create endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.
- update_application:
  - endpoint: POST /application.update
  - required fields: applicationId
  - risk: Executes Ashby application.update through the documented POST /application.update endpoint; reverse ETL plan, preview, approval, and execute are required.
- delete_application:
  - endpoint: POST /application.delete
  - required fields: applicationId
  - risk: Executes Ashby application.delete through the documented POST /application.delete endpoint; reverse ETL plan, preview, approval, and execute are required.
- add_application_hiring_team_member:
  - endpoint: POST /application.addHiringTeamMember
  - required fields: applicationId, teamMemberId, roleId
  - risk: Executes Ashby application.addHiringTeamMember through the documented POST /application.addHiringTeamMember endpoint; reverse ETL plan, preview, approval, and execute are required.
- remove_application_hiring_team_member:
  - endpoint: POST /application.removeHiringTeamMember
  - required fields: applicationId, teamMemberId, roleId
  - risk: Executes Ashby application.removeHiringTeamMember through the documented POST /application.removeHiringTeamMember endpoint; reverse ETL plan, preview, approval, and execute are required.
- change_application_stage:
  - endpoint: POST /application.change_stage
  - required fields: applicationId, interviewStageId
  - risk: Executes Ashby application.change_stage through the documented POST /application.change_stage endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.
- change_application_stage_2:
  - endpoint: POST /application.changeStage
  - required fields: applicationId, interviewStageId
  - risk: Executes Ashby application.changeStage through the documented POST /application.changeStage endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.
- change_application_source:
  - endpoint: POST /application.change_source
  - required fields: applicationId, sourceId
  - risk: Executes Ashby application.change_source through the documented POST /application.change_source endpoint; reverse ETL plan, preview, approval, and execute are required.
- change_application_source_2:
  - endpoint: POST /application.changeSource
  - required fields: applicationId, sourceId
  - risk: Executes Ashby application.changeSource through the documented POST /application.changeSource endpoint; reverse ETL plan, preview, approval, and execute are required.
- transfer_application:
  - endpoint: POST /application.transfer
  - required fields: applicationId, jobId, interviewPlanId, interviewStageId
  - risk: Executes Ashby application.transfer through the documented POST /application.transfer endpoint; reverse ETL plan, preview, approval, and execute are required.
- update_application_history:
  - endpoint: POST /application.updateHistory
  - required fields: applicationId, applicationHistory
  - risk: Executes Ashby application.updateHistory through the documented POST /application.updateHistory endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.
- submit_application_feedback:
  - endpoint: POST /applicationFeedback.submit
  - required fields: feedbackForm, formDefinitionId, applicationId
  - risk: Executes Ashby applicationFeedback.submit through the documented POST /applicationFeedback.submit endpoint; reverse ETL plan, preview, approval, and execute are required.
- create_application_feedback_request:
  - endpoint: POST /applicationFeedbackRequest.create
  - required fields: applicationId, interviewId, interviewerUserId
  - risk: Executes Ashby applicationFeedbackRequest.create through the documented POST /applicationFeedbackRequest.create endpoint; reverse ETL plan, preview, approval, and execute are required.
- submit_application_form:
  - endpoint: POST /applicationForm.submit
  - required fields: jobPostingId, applicationForm, allowSubmissionForUnpublishedJobPosting
  - risk: Executes Ashby applicationForm.submit through the documented POST /applicationForm.submit endpoint; reverse ETL plan, preview, approval, and execute are required.
- update_approval_definition:
  - endpoint: POST /approvalDefinition.update
  - required fields: entityType, entityId, approvalStepDefinitions
  - risk: Executes Ashby approvalDefinition.update through the documented POST /approvalDefinition.update endpoint; reverse ETL plan, preview, approval, and execute are required.
- create_candidate:
  - endpoint: POST /candidate.create
  - required fields: name
  - risk: Executes Ashby candidate.create through the documented POST /candidate.create endpoint; reverse ETL plan, preview, approval, and execute are required.
- upload_candidate_resume:
  - endpoint: POST /candidate.uploadResume
  - required fields: candidateId, resumeHandle
  - risk: Executes Ashby candidate.uploadResume through the documented POST /candidate.uploadResume endpoint; reverse ETL plan, preview, approval, and execute are required.
- upload_candidate_file:
  - endpoint: POST /candidate.uploadFile
  - required fields: candidateId, fileHandle
  - risk: Executes Ashby candidate.uploadFile through the documented POST /candidate.uploadFile endpoint; reverse ETL plan, preview, approval, and execute are required.
- update_candidate:
  - endpoint: POST /candidate.update
  - required fields: candidateId
  - risk: Executes Ashby candidate.update through the documented POST /candidate.update endpoint; reverse ETL plan, preview, approval, and execute are required.
- create_candidate_note:
  - endpoint: POST /candidate.createNote
  - required fields: candidateId, note
  - risk: Executes Ashby candidate.createNote through the documented POST /candidate.createNote endpoint; reverse ETL plan, preview, approval, and execute are required.
- set_candidate_fraud_status:
  - endpoint: POST /candidate.setFraudStatus
  - required fields: candidateId, fraudStatus
  - risk: Executes Ashby candidate.setFraudStatus through the documented POST /candidate.setFraudStatus endpoint; reverse ETL plan, preview, approval, and execute are required.
- anonymize_candidate:
  - endpoint: POST /candidate.anonymize
  - required fields: candidateId
  - risk: Executes Ashby candidate.anonymize through the documented POST /candidate.anonymize endpoint; reverse ETL plan, preview, approval, and execute are required.
- remove_candidate_project:
  - endpoint: POST /candidate.removeProject
  - required fields: candidateId, projectId
  - risk: Executes Ashby candidate.removeProject through the documented POST /candidate.removeProject endpoint; reverse ETL plan, preview, approval, and execute are required.
- add_candidate_project:
  - endpoint: POST /candidate.addProject
  - required fields: candidateId, projectId
  - risk: Executes Ashby candidate.addProject through the documented POST /candidate.addProject endpoint; reverse ETL plan, preview, approval, and execute are required.
- add_candidate_email_message:
  - endpoint: POST /candidate.addEmailMessage
  - required fields: candidateId, emailProviderEmailId, subject, from, to, body
  - risk: Executes Ashby candidate.addEmailMessage through the documented POST /candidate.addEmailMessage endpoint; reverse ETL plan, preview, approval, and execute are required.
- add_candidate_tag:
  - endpoint: POST /candidate.addTag
  - required fields: candidateId, tagId
  - risk: Executes Ashby candidate.addTag through the documented POST /candidate.addTag endpoint; reverse ETL plan, preview, approval, and execute are required.
- remove_candidate_tag:
  - endpoint: POST /candidate.removeTag
  - required fields: candidateId, tagId
  - risk: Executes Ashby candidate.removeTag through the documented POST /candidate.removeTag endpoint; reverse ETL plan, preview, approval, and execute are required.
- push_candidate_to_hris:
  - endpoint: POST /candidate.pushToHris
  - required fields: applicationId, externalSystem
  - risk: Executes Ashby candidate.pushToHris through the documented POST /candidate.pushToHris endpoint; reverse ETL plan, preview, approval, and execute are required.
- create_candidate_tag:
  - endpoint: POST /candidateTag.create
  - required fields: title
  - risk: Executes Ashby candidateTag.create through the documented POST /candidateTag.create endpoint; reverse ETL plan, preview, approval, and execute are required.
- archive_candidate_tag:
  - endpoint: POST /candidateTag.archive
  - required fields: tagId
  - risk: Executes Ashby candidateTag.archive through the documented POST /candidateTag.archive endpoint; reverse ETL plan, preview, approval, and execute are required.
- update_job_posting_compensation:
  - endpoint: POST /jobPosting.updateCompensation
  - required fields: jobPostingId, compensationTiers
  - risk: Executes Ashby jobPosting.updateCompensation through the documented POST /jobPosting.updateCompensation endpoint; reverse ETL plan, preview, approval, and execute are required.
- update_job_posting:
  - endpoint: POST /jobPosting.update
  - required fields: jobPostingId
  - risk: Executes Ashby jobPosting.update through the documented POST /jobPosting.update endpoint; reverse ETL plan, preview, approval, and execute are required.
- set_job_posting_status:
  - endpoint: POST /jobPosting.setStatus
  - required fields: jobPostingId, status
  - risk: Executes Ashby jobPosting.setStatus through the documented POST /jobPosting.setStatus endpoint; reverse ETL plan, preview, approval, and execute are required.
- create_job:
  - endpoint: POST /job.create
  - required fields: title, teamId, locationId
  - risk: Executes Ashby job.create through the documented POST /job.create endpoint; reverse ETL plan, preview, approval, and execute are required.
- update_job:
  - endpoint: POST /job.update
  - required fields: jobId
  - risk: Executes Ashby job.update through the documented POST /job.update endpoint; reverse ETL plan, preview, approval, and execute are required.
- update_job_compensation:
  - endpoint: POST /job.updateCompensation
  - required fields: jobId, compensationTiers
  - risk: Executes Ashby job.updateCompensation through the documented POST /job.updateCompensation endpoint; reverse ETL plan, preview, approval, and execute are required.
- set_job_status:
  - endpoint: POST /job.setStatus
  - required fields: jobId, status
  - risk: Executes Ashby job.setStatus through the documented POST /job.setStatus endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.
- archive_department:
  - endpoint: POST /department.archive
  - required fields: departmentId
  - risk: Executes Ashby department.archive through the documented POST /department.archive endpoint; reverse ETL plan, preview, approval, and execute are required.
- restore_department:
  - endpoint: POST /department.restore
  - required fields: departmentId
  - risk: Executes Ashby department.restore through the documented POST /department.restore endpoint; reverse ETL plan, preview, approval, and execute are required.
- create_department:
  - endpoint: POST /department.create
  - required fields: name
  - risk: Executes Ashby department.create through the documented POST /department.create endpoint; reverse ETL plan, preview, approval, and execute are required.
- move_department:
  - endpoint: POST /department.move
  - required fields: departmentId
  - risk: Executes Ashby department.move through the documented POST /department.move endpoint; reverse ETL plan, preview, approval, and execute are required.
- update_department:
  - endpoint: POST /department.update
  - required fields: departmentId, name
  - risk: Executes Ashby department.update through the documented POST /department.update endpoint; reverse ETL plan, preview, approval, and execute are required.
- archive_location:
  - endpoint: POST /location.archive
  - required fields: locationId
  - risk: Executes Ashby location.archive through the documented POST /location.archive endpoint; reverse ETL plan, preview, approval, and execute are required.
- create_location:
  - endpoint: POST /location.create
  - required fields: name, type
  - risk: Executes Ashby location.create through the documented POST /location.create endpoint; reverse ETL plan, preview, approval, and execute are required.
- restore_location:
  - endpoint: POST /location.restore
  - required fields: locationId
  - risk: Executes Ashby location.restore through the documented POST /location.restore endpoint; reverse ETL plan, preview, approval, and execute are required.
- move_location:
  - endpoint: POST /location.move
  - required fields: locationId, parentLocationHierarchyId
  - risk: Executes Ashby location.move through the documented POST /location.move endpoint; reverse ETL plan, preview, approval, and execute are required.
- update_location_address:
  - endpoint: POST /location.updateAddress
  - required fields: locationId
  - risk: Executes Ashby location.updateAddress through the documented POST /location.updateAddress endpoint; reverse ETL plan, preview, approval, and execute are required.
- update_location_name:
  - endpoint: POST /location.updateName
  - required fields: locationId, name
  - risk: Executes Ashby location.updateName through the documented POST /location.updateName endpoint; reverse ETL plan, preview, approval, and execute are required.
- update_location_workplace_type:
  - endpoint: POST /location.updateWorkplaceType
  - required fields: locationId, workplaceType
  - risk: Executes Ashby location.updateWorkplaceType through the documented POST /location.updateWorkplaceType endpoint; reverse ETL plan, preview, approval, and execute are required.
- update_location_external_name:
  - endpoint: POST /location.updateExternalName
  - required fields: locationId
  - risk: Executes Ashby location.updateExternalName through the documented POST /location.updateExternalName endpoint; reverse ETL plan, preview, approval, and execute are required.
- approve_offer:
  - endpoint: POST /offer.approve
  - required fields: offerVersionId
  - risk: Executes Ashby offer.approve through the documented POST /offer.approve endpoint; reverse ETL plan, preview, approval, and execute are required.
- create_offer:
  - endpoint: POST /offer.create
  - required fields: offerProcessId, offerFormId, offerForm
  - risk: Executes Ashby offer.create through the documented POST /offer.create endpoint; reverse ETL plan, preview, approval, and execute are required.
- start_offer:
  - endpoint: POST /offer.start
  - required fields: offerProcessId
  - risk: Executes Ashby offer.start through the documented POST /offer.start endpoint; reverse ETL plan, preview, approval, and execute are required.
- start_offer_approval_process:
  - endpoint: POST /offer.startApprovalProcess
  - required fields: offerVersionId
  - risk: Executes Ashby offer.startApprovalProcess through the documented POST /offer.startApprovalProcess endpoint; reverse ETL plan, preview, approval, and execute are required.
- update_offer:
  - endpoint: POST /offer.update
  - required fields: offerId, offerForm
  - risk: Executes Ashby offer.update through the documented POST /offer.update endpoint; reverse ETL plan, preview, approval, and execute are required.
- set_offer_status:
  - endpoint: POST /offer.setStatus
  - required fields: offerId, acceptanceStatus
  - risk: Executes Ashby offer.setStatus through the documented POST /offer.setStatus endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.
- set_offer_decided_at:
  - endpoint: POST /offer.setDecidedAt
  - required fields: offerId, decidedAt
  - risk: Executes Ashby offer.setDecidedAt through the documented POST /offer.setDecidedAt endpoint; reverse ETL plan, preview, approval, and execute are required.
- start_offer_process:
  - endpoint: POST /offerProcess.start
  - required fields: applicationId
  - risk: Executes Ashby offerProcess.start through the documented POST /offerProcess.start endpoint; reverse ETL plan, preview, approval, and execute are required.
- create_opening:
  - endpoint: POST /opening.create
  - risk: Executes Ashby opening.create through the documented POST /opening.create endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.
- update_opening:
  - endpoint: POST /opening.update
  - required fields: openingId
  - risk: Executes Ashby opening.update through the documented POST /opening.update endpoint; reverse ETL plan, preview, approval, and execute are required.
- set_opening_archived:
  - endpoint: POST /opening.setArchived
  - required fields: openingId, archive
  - risk: Executes Ashby opening.setArchived through the documented POST /opening.setArchived endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.
- set_opening_opening_state:
  - endpoint: POST /opening.setOpeningState
  - required fields: openingId, openingState
  - risk: Executes Ashby opening.setOpeningState through the documented POST /opening.setOpeningState endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.
- add_opening_job:
  - endpoint: POST /opening.addJob
  - required fields: openingId, jobId
  - risk: Executes Ashby opening.addJob through the documented POST /opening.addJob endpoint; reverse ETL plan, preview, approval, and execute are required.
- remove_opening_job:
  - endpoint: POST /opening.removeJob
  - required fields: openingId, jobId
  - risk: Executes Ashby opening.removeJob through the documented POST /opening.removeJob endpoint; reverse ETL plan, preview, approval, and execute are required.
- add_opening_location:
  - endpoint: POST /opening.addLocation
  - required fields: openingId, locationId
  - risk: Executes Ashby opening.addLocation through the documented POST /opening.addLocation endpoint; reverse ETL plan, preview, approval, and execute are required.
- remove_opening_location:
  - endpoint: POST /opening.removeLocation
  - required fields: openingId, locationId
  - risk: Executes Ashby opening.removeLocation through the documented POST /opening.removeLocation endpoint; reverse ETL plan, preview, approval, and execute are required.
- create_custom_field:
  - endpoint: POST /customField.create
  - required fields: fieldType, objectType, title
  - risk: Executes Ashby customField.create through the documented POST /customField.create endpoint; reverse ETL plan, preview, approval, and execute are required.
- set_custom_field_value:
  - endpoint: POST /customField.setValue
  - required fields: objectId, objectType, fieldId, fieldValue
  - risk: Executes Ashby customField.setValue through the documented POST /customField.setValue endpoint; reverse ETL plan, preview, approval, and execute are required.
- set_custom_field_values:
  - endpoint: POST /customField.setValues
  - required fields: objectId, objectType, values
  - risk: Executes Ashby customField.setValues through the documented POST /customField.setValues endpoint; reverse ETL plan, preview, approval, and execute are required.
- update_custom_field_selectable_values:
  - endpoint: POST /customField.updateSelectableValues
  - required fields: customFieldId, selectableValues
  - risk: Executes Ashby customField.updateSelectableValues through the documented POST /customField.updateSelectableValues endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.
- update_assessment:
  - endpoint: POST /assessment.update
  - required fields: assessment_id, timestamp
  - risk: Executes Ashby assessment.update through the documented POST /assessment.update endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.
- add_assessment_completed_to_candidate:
  - endpoint: POST /assessment.addCompletedToCandidate
  - required fields: candidateId, partnerId, assessment, timestamp
  - risk: Executes Ashby assessment.addCompletedToCandidate through the documented POST /assessment.addCompletedToCandidate endpoint; reverse ETL plan, preview, approval, and execute are required.
- add_hiring_team_member:
  - endpoint: POST /hiringTeam.addMember
  - required fields: teamMemberId, roleId
  - risk: Executes Ashby hiringTeam.addMember through the documented POST /hiringTeam.addMember endpoint; reverse ETL plan, preview, approval, and execute are required.
- remove_hiring_team_member:
  - endpoint: POST /hiringTeam.removeMember
  - required fields: teamMemberId, roleId
  - risk: Executes Ashby hiringTeam.removeMember through the documented POST /hiringTeam.removeMember endpoint; reverse ETL plan, preview, approval, and execute are required.
- interviewer_user_settings:
  - endpoint: POST /user.interviewerSettings
  - required fields: userId
  - risk: Executes Ashby user.interviewerSettings through the documented POST /user.interviewerSettings endpoint; reverse ETL plan, preview, approval, and execute are required.
- update_user_interviewer_settings:
  - endpoint: POST /user.updateInterviewerSettings
  - required fields: userId
  - risk: Executes Ashby user.updateInterviewerSettings through the documented POST /user.updateInterviewerSettings endpoint; reverse ETL plan, preview, approval, and execute are required.
- create_user_interviewer_pause:
  - endpoint: POST /user.createInterviewerPause
  - required fields: userId
  - risk: Executes Ashby user.createInterviewerPause through the documented POST /user.createInterviewerPause endpoint; reverse ETL plan, preview, approval, and execute are required.
- delete_user_interviewer_pause:
  - endpoint: POST /user.deleteInterviewerPause
  - required fields: interviewerPauseId
  - risk: Executes Ashby user.deleteInterviewerPause through the documented POST /user.deleteInterviewerPause endpoint; reverse ETL plan, preview, approval, and execute are required.
- set_user_custom_field_value:
  - endpoint: POST /user.setCustomFieldValue
  - required fields: userId, fieldId, fieldValue
  - risk: Executes Ashby user.setCustomFieldValue through the documented POST /user.setCustomFieldValue endpoint; reverse ETL plan, preview, approval, and execute are required.
- set_user_custom_field_values:
  - endpoint: POST /user.setCustomFieldValues
  - required fields: userId, values
  - risk: Executes Ashby user.setCustomFieldValues through the documented POST /user.setCustomFieldValues endpoint; reverse ETL plan, preview, approval, and execute are required.
- create_referral:
  - endpoint: POST /referral.create
  - required fields: id, creditedToUserId, fieldSubmissions
  - risk: Executes Ashby referral.create through the documented POST /referral.create endpoint; reverse ETL plan, preview, approval, and execute are required.
- add_sequence:
  - endpoint: POST /sequence.add
  - required fields: sequenceTemplateId, candidateId, start
  - risk: Executes Ashby sequence.add through the documented POST /sequence.add endpoint; reverse ETL plan, preview, approval, and execute are required.
- cancel_sequence:
  - endpoint: POST /sequence.cancel
  - required fields: sequenceId
  - risk: Executes Ashby sequence.cancel through the documented POST /sequence.cancel endpoint; reverse ETL plan, preview, approval, and execute are required.
- discard_sequence:
  - endpoint: POST /sequence.discard
  - required fields: sequenceId
  - risk: Executes Ashby sequence.discard through the documented POST /sequence.discard endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required.
- update_sequence_stage:
  - endpoint: POST /sequence.updateStage
  - required fields: sequenceId, stageId
  - risk: Executes Ashby sequence.updateStage through the documented POST /sequence.updateStage endpoint; reverse ETL plan, preview, approval, and execute are required.
- start_sequence:
  - endpoint: POST /sequence.start
  - required fields: sequenceId
  - risk: Executes Ashby sequence.start through the documented POST /sequence.start endpoint; reverse ETL plan, preview, approval, and execute are required.
- create_interview_schedule:
  - endpoint: POST /interviewSchedule.create
  - required fields: applicationId, interviewEvents
  - risk: Executes Ashby interviewSchedule.create through the documented POST /interviewSchedule.create endpoint; reverse ETL plan, preview, approval, and execute are required.
- update_interview_schedule:
  - endpoint: POST /interviewSchedule.update
  - required fields: interviewScheduleId
  - risk: Executes Ashby interviewSchedule.update through the documented POST /interviewSchedule.update endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.
- cancel_interview_schedule:
  - endpoint: POST /interviewSchedule.cancel
  - required fields: id
  - risk: Executes Ashby interviewSchedule.cancel through the documented POST /interviewSchedule.cancel endpoint; reverse ETL plan, preview, approval, and execute are required.
- create_survey_request:
  - endpoint: POST /surveyRequest.create
  - required fields: candidateId, applicationId, surveyFormDefinitionId
  - risk: Executes Ashby surveyRequest.create through the documented POST /surveyRequest.create endpoint; reverse ETL plan, preview, approval, and execute are required.
- create_survey_submission:
  - endpoint: POST /surveySubmission.create
  - required fields: surveyFormDefinitionId, candidateId, applicationId, submittedValues
  - risk: Executes Ashby surveySubmission.create through the documented POST /surveySubmission.create endpoint; reverse ETL plan, preview, approval, and execute are required.
- create_webhook:
  - endpoint: POST /webhook.create
  - required fields: webhookType, requestUrl, secretToken
  - risk: Executes Ashby webhook.create through the documented POST /webhook.create endpoint; reverse ETL plan, preview, approval, and execute are required.
- update_webhook:
  - endpoint: POST /webhook.update
  - required fields: webhookId
  - risk: Executes Ashby webhook.update through the documented POST /webhook.update endpoint; reverse ETL plan, preview, approval, and execute are required.
- delete_webhook:
  - endpoint: POST /webhook.delete
  - required fields: webhookId
  - risk: Executes Ashby webhook.delete through the documented POST /webhook.delete endpoint; reverse ETL plan, preview, approval, and execute are required.
- archive_interviewer_pool:
  - endpoint: POST /interviewerPool.archive
  - required fields: interviewerPoolId
  - risk: Executes Ashby interviewerPool.archive through the documented POST /interviewerPool.archive endpoint; reverse ETL plan, preview, approval, and execute are required.
- restore_interviewer_pool:
  - endpoint: POST /interviewerPool.restore
  - required fields: interviewerPoolId
  - risk: Executes Ashby interviewerPool.restore through the documented POST /interviewerPool.restore endpoint; reverse ETL plan, preview, approval, and execute are required.
- create_interviewer_pool:
  - endpoint: POST /interviewerPool.create
  - required fields: title
  - risk: Executes Ashby interviewerPool.create through the documented POST /interviewerPool.create endpoint; reverse ETL plan, preview, approval, and execute are required.
- update_interviewer_pool:
  - endpoint: POST /interviewerPool.update
  - required fields: interviewerPoolId
  - risk: Executes Ashby interviewerPool.update through the documented POST /interviewerPool.update endpoint; reverse ETL plan, preview, approval, and execute are required.
- add_interviewer_pool_user:
  - endpoint: POST /interviewerPool.addUser
  - required fields: interviewerPoolId, userId
  - risk: Executes Ashby interviewerPool.addUser through the documented POST /interviewerPool.addUser endpoint; reverse ETL plan, preview, approval, and execute are required.
- remove_interviewer_pool_user:
  - endpoint: POST /interviewerPool.removeUser
  - required fields: interviewerPoolId, userId
  - risk: Executes Ashby interviewerPool.removeUser through the documented POST /interviewerPool.removeUser endpoint; reverse ETL plan, preview, approval, and execute are required.
- generate_report:
  - endpoint: POST /report.generate
  - required fields: reportId
  - risk: Executes Ashby report.generate through the documented POST /report.generate endpoint; reverse ETL plan, preview, approval, and execute are required.

## Security

- read risk: bounded Ashby POST reads using documented endpoints, Basic API-key auth, page-size and max-pages bounds, and sanitized replay fixtures
- write risk: named reverse-ETL actions only; no generic HTTP method/path/body; destructive actions require typed confirmation
- approval: reverse ETL writes require plan -> preview -> explicit approval -> execute
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Ashby applicant-tracking connector with typed REST streams, bounded direct reads, and gated reverse-ETL writes.
- Usage: pm connectors command ashby <command> [flags]
- Source CLI: Ashby Public API (https://developers.ashbyhq.com/reference/candidateaddtag)
- ETL streams
- Bounded direct reads
- Reverse ETL writes
- Other Commands
  - candidate list - Lists all candidates in the organization with pagination and incremental sync support.

Use the `syncToken` parameter to retrieve only candidates updated since  [intent=etl availability=implemented stream=candidates]; notes: Fixed Ashby stream for candidate.list; flags map only to documented request body fields.; flags: --created-after, --created-before
  - job list - Lists all jobs.

See the [Pagination and Incremental Synchronization](/docs/pagination-and-incremental-sync) guide for detailed usage examples.

**Requires the  [intent=etl availability=implemented stream=jobs]; notes: Fixed Ashby stream for job.list; flags map only to documented request body fields.; flags: --status, --created-after, --opened-after, --opened-before, --closed-after, --closed-before, --include-unpublished-job-postings-ids, --expand
  - application list - Gets all applications in the organization.

See the [Pagination and Incremental Synchronization](/docs/pagination-and-incremental-sync) guide for detailed usage [intent=etl availability=implemented stream=applications]; notes: Fixed Ashby stream for application.list; flags map only to documented request body fields.; flags: --created-after, --created-before, --status, --job-id, --expand
  - user list - Lists all users in the organization with pagination support.

By default, only active (enabled) users are returned. Use `includeDeactivated: true` to include de [intent=etl availability=implemented stream=users]; notes: Fixed Ashby stream for user.list; flags map only to documented request body fields.; flags: --include-deactivated
  - api-key info - Returns details for the API key used to make the request.

**Requires the [`apiKeysRead`](authentication#permissions-apikeyinfo) permission.** [intent=etl availability=implemented stream=api_key_info]; notes: Fixed Ashby stream for apiKey.info; flags map only to documented request body fields.
  - audit-log list - > **Beta**
>
> **This API is in active development and only available in a closed beta with early design partners.**

Lists an organization's audit log entries, [intent=etl availability=implemented stream=audit_log_list]; notes: Fixed Ashby stream for auditLog.list; flags map only to documented request body fields.; flags: --start-date, --end-date, --actor-ids, --target-ids, --target-types, --categories
  - application create - Consider a candidate for a job (e.g. when sourcing a candidate for a job posting).

If you're submitting an application as a job board, use the [`applicationFor [intent=reverse_etl availability=implemented write=create_application]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby application.create through the documented POST /application.create endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --candidate-id, --job-id, --interview-plan-id, --interview-stage-id, --source-id, --credited-to-user-id, --created-at, --application-history
  - application update - Update an application.

To set values for custom fields on Applications, use the [`customField.setValue`](ref:customfieldsetvalue) endpoint.

**Requires the [`c [intent=reverse_etl availability=implemented write=update_application]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby application.update through the documented POST /application.update endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --application-id, --source-id, --credited-to-user-id, --created-at, --send-notifications
  - application delete - Deletes an application by id.

**Requires the [`candidatesDelete`](authentication#permissions-applicationdelete) permission.** [intent=reverse_etl availability=implemented write=delete_application]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby application.delete through the documented POST /application.delete endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --application-id
  - application info - Fetch application details by application id or by submitted form instance id (which is returned by the `applicationForm.submit` endpoint). If both `applicationI [intent=etl availability=implemented stream=application_info]; notes: Fixed Ashby stream for application.info; flags map only to documented request body fields. Requires at least one documented selector: applicationId, submittedFormInstanceId.; flags: --application-id, --submitted-form-instance-id, --expand
  - application add-hiring-team-member - Adds an Ashby user to the hiring team at the application level.

**Requires the [`candidatesWrite`](authentication#permissions-applicationaddhiringteammember) p [intent=reverse_etl availability=implemented write=add_application_hiring_team_member]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby application.addHiringTeamMember through the documented POST /application.addHiringTeamMember endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --application-id, --team-member-id, --role-id
  - application remove-hiring-team-member - Unassigns a hiring team role from an Ashby user at the application level.

**Requires the [`candidatesWrite`](authentication#permissions-applicationremovehiring [intent=reverse_etl availability=implemented write=remove_application_hiring_team_member]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby application.removeHiringTeamMember through the documented POST /application.removeHiringTeamMember endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --application-id, --team-member-id, --role-id
  - application-hiring-team-role list - Gets all available hiring team roles for applications in the organization.

**Requires the [`candidatesRead`](authentication#permissions-applicationhiringteamro [intent=etl availability=implemented stream=application_hiring_team_role_list]; notes: Fixed Ashby stream for applicationHiringTeamRole.list; flags map only to documented request body fields.
  - application change-stage - **Deprecated.** Use [`application.changeStage`](ref:applicationchangestage) instead.

Change the stage of an application.

**Requires the [`candidatesWrite`](au [intent=reverse_etl availability=implemented write=change_application_stage]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby application.change_stage through the documented POST /application.change_stage endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --application-id, --interview-stage-id, --archive-reason-id
  - application change-stage-2 - Change the stage of an application.

**Requires the [`candidatesWrite`](authentication#permissions-applicationchangestage) permission.** [intent=reverse_etl availability=implemented write=change_application_stage_2]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby application.changeStage through the documented POST /application.changeStage endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --application-id, --interview-stage-id, --archive-reason-id
  - application change-source - **Deprecated.** Use [`application.changeSource`](ref:applicationchangesource) instead.

Change the source of an application.

**Requires the [`candidatesWrite`] [intent=reverse_etl availability=implemented write=change_application_source]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby application.change_source through the documented POST /application.change_source endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --application-id, --source-id
  - application change-source-2 - Change the source of an application.

**Requires the [`candidatesWrite`](authentication#permissions-applicationchangesource) permission.** [intent=reverse_etl availability=implemented write=change_application_source_2]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby application.changeSource through the documented POST /application.changeSource endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --application-id, --source-id
  - application transfer - Transfer an application to a different job.

**Requires the [`candidatesWrite`](authentication#permissions-applicationtransfer) permission.** [intent=reverse_etl availability=implemented write=transfer_application]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby application.transfer through the documented POST /application.transfer endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --application-id, --job-id, --interview-plan-id, --interview-stage-id, --start-automatic-activities
  - application update-history - Update the history of an application. Used to update stage timestamps and to delete history events.

**Also requires the `Allow updating application history?` s [intent=reverse_etl availability=implemented write=update_application_history]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby application.updateHistory through the documented POST /application.updateHistory endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --application-id, --application-history-0-stage-id, --application-history-0-stage-number, --application-history-0-entered-stage-at
  - application list-history - Fetch a paginated list of application history items for an application.

This endpoint supports pagination only (not incremental sync). See the [Pagination and  [intent=etl availability=implemented stream=application_list_history]; notes: Fixed Ashby stream for application.listHistory; flags map only to documented request body fields.; flags: --application-id
  - application list-criteria-evaluations - Fetch a paginated list of AI criteria evaluations for an application.

This endpoint returns the AI-generated criteria evaluations that assess how well a candid [intent=etl availability=implemented stream=application_list_criteria_evaluations]; notes: Fixed Ashby stream for application.listCriteriaEvaluations; flags map only to documented request body fields.; flags: --application-id
  - application-feedback list - List all interview scorecards and feedback submissions associated with an application.

Each feedback submission contains:
- **formDefinition**: The structure o [intent=etl availability=implemented stream=application_feedback_list]; notes: Fixed Ashby stream for applicationFeedback.list; flags map only to documented request body fields.; flags: --application-id, --created-after
  - application-feedback submit - Application feedback forms support a variety of field types.

The values accepted for each field depend on the type of field that's being filled out:
- `Boolean [intent=reverse_etl availability=partial write=submit_application_feedback]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby applicationFeedback.submit through the documented POST /applicationFeedback.submit endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed. This command has nested object/array requirements that are implemented by the reverse-ETL action schema but are not fully expressible as scalar CLI flags; use file/warehouse reverse-ETL inputs for execution.; flags: --feedback-form-field-submissions-0-path, --form-definition-id, --application-id, --user-id, --interview-event-id
  - application-feedback-request create - Request feedback on an application without scheduling an interview.
The `interviewEventId` returned in the response can be provided to `applicationFeedback.subm [intent=reverse_etl availability=implemented write=create_application_feedback_request]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby applicationFeedbackRequest.create through the documented POST /applicationFeedbackRequest.create endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --application-id, --interview-id, --interviewer-user-id
  - application-form submit - Submits a completed application form for a job posting.

**Requires the [`candidatesWrite`](authentication#permissions-applicationformsubmit) permission.** [intent=reverse_etl availability=partial write=submit_application_form]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby applicationForm.submit through the documented POST /applicationForm.submit endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed. This command has nested object/array requirements that are implemented by the reverse-ETL action schema but are not fully expressible as scalar CLI flags; use file/warehouse reverse-ETL inputs for execution.; flags: --job-posting-id, --application-form-field-submissions-0-path, --allow-submission-for-unpublished-job-posting, --tag-ids
  - approval-definition update - Create or update an approval definition for a specific entity that requires approval. The entity requiring approval must be within scope of an approval in Ashby [intent=reverse_etl availability=implemented write=update_approval_definition]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby approvalDefinition.update through the documented POST /approvalDefinition.update endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --entity-type, --entity-id, --approval-step-definitions-0-approvals-required, --approval-step-definitions-0-approvers-0-user-id, --approval-step-definitions-0-approvers-0-type, --submit-approval-request
  - candidate search - Searches for candidates by email and/or name.

**Requires the [`candidatesRead`](authentication#permissions-candidatesearch) permission.** [intent=direct_read availability=implemented operation=ashby.direct.candidate.search]; approval: none; risk: bounded JSON direct read; credential-marked response fields are redacted, and non-credential identity fields remain complete in trusted live local output; notes: Fixed Ashby POST direct read; no raw method/path/body override is exposed.; flags: --email, --name
  - candidate info - Fetches details about a single candidate by id or external mapping id.

**Requires the [`candidatesRead`](authentication#permissions-candidateinfo) permission.* [intent=etl availability=implemented stream=candidate_info]; notes: Fixed Ashby stream for candidate.info; flags map only to documented request body fields. Requires at least one documented selector: id, externalMappingId.; flags: --id, --external-mapping-id
  - candidate create - Creates a new candidate.

**Requires the [`candidatesWrite`](authentication#permissions-candidatecreate) permission.** [intent=reverse_etl availability=implemented write=create_candidate]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby candidate.create through the documented POST /candidate.create endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --name, --email, --phone-number, --linked-in-url, --github-url, --website, --alternate-email-addresses, --source-id, --credited-to-user-id, --created-at
  - candidate upload-resume - Uploads a resume for a candidate. Accepts either a multipart/form-data
request with a `resume` file part, or a JSON body with a `resumeHandle`
previously create [intent=reverse_etl availability=implemented write=upload_candidate_resume]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby candidate.uploadResume through the documented POST /candidate.uploadResume endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --candidate-id, --resume-handle
  - candidate upload-file - Uploads a file for a candidate. Accepts either a multipart/form-data
request with a `file` file part, or a JSON body with a `fileHandle`
previously created via  [intent=reverse_etl availability=implemented write=upload_candidate_file]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby candidate.uploadFile through the documented POST /candidate.uploadFile endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --candidate-id, --file-handle
  - candidate update - Updates an existing candidate.

**Requires the [`candidatesWrite`](authentication#permissions-candidateupdate) permission.** [intent=reverse_etl availability=implemented write=update_candidate]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby candidate.update through the documented POST /candidate.update endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --candidate-id, --name, --email, --phone-number, --linked-in-url, --github-url, --website-url, --alternate-email, --social-links, --source-id, --credited-to-user-id, --created-at, --send-notifications
  - candidate create-note - Creates a note on a candidate.

**Requires the [`candidatesWrite`](authentication#permissions-candidatecreatenote) permission.** [intent=reverse_etl availability=implemented write=create_candidate_note]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby candidate.createNote through the documented POST /candidate.createNote endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --candidate-id, --note-type, --note-value, --send-notifications, --is-private, --created-at
  - candidate list-client-info - Lists client information records (IP, user agent) collected for a candidate.

**Requires the [`candidatesRead`](authentication#permissions-candidatelistclientin [intent=etl availability=implemented stream=candidate_list_client_info]; notes: Fixed Ashby stream for candidate.listClientInfo; flags map only to documented request body fields.; flags: --candidate-id
  - candidate list-fraud-checks - Lists the fraud checks performed on a candidate.

**Requires the [`candidatesRead`](authentication#permissions-candidatelistfraudchecks) permission.** [intent=etl availability=implemented stream=candidate_list_fraud_checks]; notes: Fixed Ashby stream for candidate.listFraudChecks; flags map only to documented request body fields.; flags: --candidate-id
  - candidate set-fraud-status - Updates the manual fraud-review status of a candidate.

**Requires the [`candidatesWrite`](authentication#permissions-candidatesetfraudstatus) permission.** [intent=reverse_etl availability=implemented write=set_candidate_fraud_status]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby candidate.setFraudStatus through the documented POST /candidate.setFraudStatus endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --candidate-id, --fraud-status
  - candidate list-notes - Lists the notes that have been added to a candidate.

**Requires the [`candidatesRead`](authentication#permissions-candidatelistnotes) permission.** [intent=etl availability=implemented stream=candidate_list_notes]; notes: Fixed Ashby stream for candidate.listNotes; flags map only to documented request body fields.; flags: --candidate-id
  - candidate anonymize - Anonymizes a candidate's personally identifiable information.

**Requires the [`candidatesWrite`](authentication#permissions-candidateanonymize) permission.** [intent=reverse_etl availability=implemented write=anonymize_candidate]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby candidate.anonymize through the documented POST /candidate.anonymize endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --candidate-id
  - candidate remove-project - Removes the candidate from a project.

**Requires the [`candidatesWrite`](authentication#permissions-candidateremoveproject) permission.** [intent=reverse_etl availability=implemented write=remove_candidate_project]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby candidate.removeProject through the documented POST /candidate.removeProject endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --candidate-id, --project-id
  - candidate add-project - Adds a candidate to a project.

**Requires the [`candidatesWrite`](authentication#permissions-candidateaddproject) permission.** [intent=reverse_etl availability=implemented write=add_candidate_project]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby candidate.addProject through the documented POST /candidate.addProject endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --candidate-id, --project-id
  - candidate add-email-message - Attaches an existing email message (e.g. fetched from a partner provider) to a candidate.

**Requires the [`candidatesWrite`](authentication#permissions-candida [intent=reverse_etl availability=implemented write=add_candidate_email_message]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby candidate.addEmailMessage through the documented POST /candidate.addEmailMessage endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --candidate-id, --email-provider-email-id, --subject, --from, --to, --message-body, --user-id, --sent-at, --cc, --message-url, --message-id-header, --thread-id, --is-private
  - candidate add-tag - Adds a tag to a candidate.

**Requires the [`candidatesWrite`](authentication#permissions-candidateaddtag) permission.** [intent=reverse_etl availability=implemented write=add_candidate_tag]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby candidate.addTag through the documented POST /candidate.addTag endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --candidate-id, --tag-id
  - candidate remove-tag - Removes a tag from a candidate.

**Requires the [`candidatesWrite`](authentication#permissions-candidateremovetag) permission.** [intent=reverse_etl availability=implemented write=remove_candidate_tag]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby candidate.removeTag through the documented POST /candidate.removeTag endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --candidate-id, --tag-id
  - candidate list-projects - Lists the projects a candidate has been added to.

**Requires the [`candidatesRead`](authentication#permissions-candidatelistprojects) permission.** [intent=etl availability=implemented stream=candidate_list_projects]; notes: Fixed Ashby stream for candidate.listProjects; flags map only to documented request body fields.; flags: --candidate-id
  - candidate push-to-hris - > Beta
>
> This feature is in beta and may not be available for all organizations.

Pushes a candidate's data to an HRIS system (e.g. Workday, BambooHR, ADP).

 [intent=reverse_etl availability=implemented write=push_candidate_to_hris]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby candidate.pushToHris through the documented POST /candidate.pushToHris endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --application-id, --external-system, --integration-partner-id
  - candidate-tag create - Creates a candidate tag.

If a tag already exists with the given title, the existing tag will be returned.

**Requires the [`hiringProcessMetadataWrite`](authen [intent=reverse_etl availability=implemented write=create_candidate_tag]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby candidateTag.create through the documented POST /candidateTag.create endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --title
  - candidate-tag archive - Archives a candidate tag.

**Requires the [`hiringProcessMetadataWrite`](authentication#permissions-candidatetagarchive) permission.** [intent=reverse_etl availability=implemented write=archive_candidate_tag]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby candidateTag.archive through the documented POST /candidateTag.archive endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --tag-id
  - candidate-tag list - Lists all candidate tags.

See the [Pagination and Incremental Synchronization](/docs/pagination-and-incremental-sync) guide for detailed usage examples.

**Req [intent=etl availability=implemented stream=candidate_tag_list]; notes: Fixed Ashby stream for candidateTag.list; flags map only to documented request body fields.; flags: --include-archived
  - communication-template list - List all enabled communication templates.

**Requires the [`hiringProcessMetadataRead`](authentication#permissions-communicationtemplatelist) permission.** [intent=etl availability=implemented stream=communication_template_list]; notes: Fixed Ashby stream for communicationTemplate.list; flags map only to documented request body fields.
  - feedback-form-definition list - Lists all feedback form definitions.

See the [Pagination and Incremental Synchronization](/docs/pagination-and-incremental-sync) guide for detailed usage examp [intent=etl availability=implemented stream=feedback_form_definition_list]; notes: Fixed Ashby stream for feedbackFormDefinition.list; flags map only to documented request body fields.; flags: --include-archived
  - feedback-form-definition info - Returns a single feedback form by id

**Requires the [`hiringProcessMetadataRead`](authentication#permissions-feedbackformdefinitioninfo) permission.** [intent=etl availability=implemented stream=feedback_form_definition_info]; notes: Fixed Ashby stream for feedbackFormDefinition.info; flags map only to documented request body fields.; flags: --feedback-form-definition-id
  - job-posting list - Lists published job postings. By default, only published job postings are returned.

Set `includeUnpublishedJobPostings` to `true` to also include unpublished ( [intent=etl availability=implemented stream=job_posting_list]; notes: Fixed Ashby stream for jobPosting.list; flags map only to documented request body fields.; flags: --location, --department, --listed-only, --include-unpublished-job-postings, --job-board-id
  - job-posting info - Retrieve an individual job posting.

Set `includeUnpublishedJobPostings` to `true` when fetching an unpublished (draft) job posting. This flag is required for d [intent=etl availability=implemented stream=job_posting_info]; notes: Fixed Ashby stream for jobPosting.info; flags map only to documented request body fields.; flags: --job-posting-id, --job-board-id, --include-unpublished-job-postings, --expand
  - job-posting update-compensation - Updates compensation for an existing job posting.

Set `includeUnpublishedJobPostings` to `true` when updating an unpublished (draft) job posting. This flag is  [intent=reverse_etl availability=implemented write=update_job_posting_compensation]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby jobPosting.updateCompensation through the documented POST /jobPosting.updateCompensation endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --job-posting-id, --compensation-tiers-0-components-0-compensation-type, --compensation-tiers-0-components-0-interval, --include-unpublished-job-postings
  - job-posting update - Updates an existing job posting.

Set `includeUnpublishedJobPostings` to `true` when updating an unpublished (draft) job posting. This flag is required for draf [intent=reverse_etl availability=implemented write=update_job_posting]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby jobPosting.update through the documented POST /jobPosting.update endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --job-posting-id, --title, --workplace-type, --suppress-description-opening, --suppress-description-closing, --application-confirmation-email-template-id, --include-unpublished-job-postings
  - job-posting set-status - Sets the status of a job posting. Use this to publish or unpublish a job posting.

Set `status` to `Published` to publish a draft job posting. The posting must  [intent=reverse_etl availability=implemented write=set_job_posting_status]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby jobPosting.setStatus through the documented POST /jobPosting.setStatus endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --job-posting-id, --status
  - job info - Fetches details of a single job by id.

**Requires the [`jobsRead`](authentication#permissions-jobinfo) permission.** [intent=etl availability=implemented stream=job_info]; notes: Fixed Ashby stream for job.info; flags map only to documented request body fields.; flags: --id, --include-unpublished-job-postings-ids, --expand
  - job create - Creates a new job.

**Requires the [`jobsWrite`](authentication#permissions-jobcreate) permission.** [intent=reverse_etl availability=implemented write=create_job]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby job.create through the documented POST /job.create endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --title, --team-id, --location-id, --default-interview-plan-id, --job-template-id, --employment-type, --brand-id
  - job update - Updates an existing job. At least one field other than `jobId` must be supplied.

**Requires the [`jobsWrite`](authentication#permissions-jobupdate) permission. [intent=reverse_etl availability=implemented write=update_job]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby job.update through the documented POST /job.update endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --job-id, --title, --team-id, --location-id, --default-interview-plan-id, --employment-type, --custom-requisition-id
  - job update-compensation - Replaces the compensation tiers on a job. Pass an empty array to clear existing compensation.

**Requires the [`jobsWrite`](authentication#permissions-jobupdate [intent=reverse_etl availability=implemented write=update_job_compensation]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby job.updateCompensation through the documented POST /job.updateCompensation endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --job-id, --compensation-tiers-0-components-0-compensation-type, --compensation-tiers-0-components-0-interval
  - job search - Searches jobs by title or custom requisition id. At least one of `title` or `requisitionId` must be provided.

**Requires the [`jobsRead`](authentication#permis [intent=direct_read availability=implemented operation=ashby.direct.job.search]; approval: none; risk: bounded JSON direct read; credential-marked response fields are redacted, and non-credential identity fields remain complete in trusted live local output; notes: Fixed Ashby POST direct read; no raw method/path/body override is exposed.; flags: --title, --requisition-id
  - job set-status - Sets the status of a job.

**Requires the [`jobsWrite`](authentication#permissions-jobsetstatus) permission.** [intent=reverse_etl availability=implemented write=set_job_status]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby job.setStatus through the documented POST /job.setStatus endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --job-id, --status
  - job-board list - List all enabled job boards.

**Requires the [`jobsRead`](authentication#permissions-jobboardlist) permission.** [intent=etl availability=implemented stream=job_board_list]; notes: Fixed Ashby stream for jobBoard.list; flags map only to documented request body fields.
  - job-interview-plan info - Returns a job's interview plan, including activities and interviews that need to be scheduled at each stage.

**Requires the [`jobsRead`](authentication#permiss [intent=etl availability=implemented stream=job_interview_plan_info]; notes: Fixed Ashby stream for jobInterviewPlan.info; flags map only to documented request body fields.; flags: --job-id
  - job-template list - List all active and inactive job templates.

See the [Pagination and Incremental Synchronization](/docs/pagination-and-incremental-sync) guide for detailed usag [intent=etl availability=implemented stream=job_template_list]; notes: Fixed Ashby stream for jobTemplate.list; flags map only to documented request body fields.; flags: --expand
  - department archive - Archives a department.

**Requires the [`organizationWrite`](authentication#permissions-departmentarchive) permission.** [intent=reverse_etl availability=implemented write=archive_department]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby department.archive through the documented POST /department.archive endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --department-id
  - department restore - Restores a department.

**Requires the [`organizationWrite`](authentication#permissions-departmentrestore) permission.** [intent=reverse_etl availability=implemented write=restore_department]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby department.restore through the documented POST /department.restore endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --department-id
  - department create - Creates a department.

**Requires the [`organizationWrite`](authentication#permissions-departmentcreate) permission.** [intent=reverse_etl availability=implemented write=create_department]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby department.create through the documented POST /department.create endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --name, --external-name, --parent-id
  - department info - Fetch department details by id.

**Requires the [`organizationRead`](authentication#permissions-departmentinfo) permission.** [intent=etl availability=implemented stream=department_info]; notes: Fixed Ashby stream for department.info; flags map only to documented request body fields.; flags: --department-id
  - department list - Lists all departments.

See the [Pagination and Incremental Synchronization](/docs/pagination-and-incremental-sync) guide for detailed usage examples.

**Requir [intent=etl availability=implemented stream=department_list]; notes: Fixed Ashby stream for department.list; flags map only to documented request body fields.; flags: --include-archived
  - department move - Moves a department to another parent.

**Requires the [`organizationWrite`](authentication#permissions-departmentmove) permission.** [intent=reverse_etl availability=implemented write=move_department]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby department.move through the documented POST /department.move endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --department-id, --parent-id
  - department update - Updates a department.

**Requires the [`organizationWrite`](authentication#permissions-departmentupdate) permission.** [intent=reverse_etl availability=implemented write=update_department]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby department.update through the documented POST /department.update endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --department-id, --name, --external-name
  - location archive - Archives a location or location hierarchy.

**Requires the [`organizationWrite`](authentication#permissions-locationarchive) permission.** [intent=reverse_etl availability=implemented write=archive_location]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby location.archive through the documented POST /location.archive endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --location-id
  - location create - Creates a location or location hierarchy.

**Requires the [`organizationWrite`](authentication#permissions-locationcreate) permission.** [intent=reverse_etl availability=implemented write=create_location]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby location.create through the documented POST /location.create endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --name, --type, --parent-location-id, --is-remote, --workplace-type, --external-name
  - location restore - Restores an archived location or location hierarchy.

**Requires the [`organizationWrite`](authentication#permissions-locationrestore) permission.** [intent=reverse_etl availability=implemented write=restore_location]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby location.restore through the documented POST /location.restore endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --location-id
  - location list - List all locations. Regions are not returned.

See the [Pagination and Incremental Synchronization](/docs/pagination-and-incremental-sync) guide for detailed us [intent=etl availability=implemented stream=location_list]; notes: Fixed Ashby stream for location.list; flags map only to documented request body fields.; flags: --include-archived, --include-location-hierarchy
  - location info - Gets details for a single location by id.

**Requires the [`organizationRead`](authentication#permissions-locationinfo) permission.** [intent=etl availability=implemented stream=location_info]; notes: Fixed Ashby stream for location.info; flags map only to documented request body fields.; flags: --location-id
  - location move - Moves a location in location hierarchy.

**Requires the [`organizationWrite`](authentication#permissions-locationmove) permission.** [intent=reverse_etl availability=implemented write=move_location]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby location.move through the documented POST /location.move endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --location-id, --parent-location-hierarchy-id
  - location update-address - Update an address of a location or location hierarchy.

**Requires the [`organizationWrite`](authentication#permissions-locationupdateaddress) permission.** [intent=reverse_etl availability=implemented write=update_location_address]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby location.updateAddress through the documented POST /location.updateAddress endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --location-id
  - location update-name - Update location's name.

**Requires the [`organizationWrite`](authentication#permissions-locationupdatename) permission.** [intent=reverse_etl availability=implemented write=update_location_name]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby location.updateName through the documented POST /location.updateName endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --location-id, --name
  - location update-workplace-type - Update location's workplace type.

**Requires the [`organizationWrite`](authentication#permissions-locationupdateworkplacetype) permission.** [intent=reverse_etl availability=implemented write=update_location_workplace_type]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby location.updateWorkplaceType through the documented POST /location.updateWorkplaceType endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --location-id, --workplace-type
  - location update-external-name - Update a location's external (candidate-facing) name.

**Requires the [`organizationWrite`](authentication#permissions-locationupdateexternalname) permission.** [intent=reverse_etl availability=implemented write=update_location_external_name]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby location.updateExternalName through the documented POST /location.updateExternalName endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --location-id, --external-name
  - interview-plan list - List published interview plans. Draft interview plans are not returned. If `includeArchived` is true, archived interview plans are also included. Job-specific i [intent=etl availability=implemented stream=interview_plan_list]; notes: Fixed Ashby stream for interviewPlan.list; flags map only to documented request body fields.; flags: --include-archived
  - interview-stage list - List all interview stages for an interview plan in order.

**Requires the [`interviewsRead`](authentication#permissions-interviewstagelist) permission.** [intent=etl availability=implemented stream=interview_stage_list]; notes: Fixed Ashby stream for interviewStage.list; flags map only to documented request body fields.; flags: --interview-plan-id
  - interview-stage-group list - List all interview stage groups in the organization in order.

**Requires the [`interviewsRead`](authentication#permissions-interviewstagegrouplist) permission. [intent=etl availability=implemented stream=interview_stage_group_list]; notes: Fixed Ashby stream for interviewStageGroup.list; flags map only to documented request body fields.
  - notetaker-transcript info - Fetches metadata and a pre-signed download URL for an AI
Notetaker transcript recording.

**Prerequisites:**
- Your organization must have the **AI Notetaker ad [intent=direct_read availability=implemented operation=ashby.direct.notetaker.transcript.info]; approval: none; risk: bounded JSON direct read; credential-marked response fields are redacted, and Ashby signed URL fields are preserved (results.url/results.transcriptUrl) in trusted live local output; notes: Fixed Ashby POST direct read; no raw method/path/body override is exposed.; flags: --notetaker-transcript-id (required)
  - offer approve - Approves an offer or a specific approval step within an offer's approval process.

This endpoint mimics the behavior of the "Force Approve" function in the Ashb [intent=reverse_etl availability=implemented write=approve_offer]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby offer.approve through the documented POST /offer.approve endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --offer-version-id, --approval-step-id, --user-id, --exclude-form-definition
  - offer create - Creates a new Offer

Offer forms support a variety of field types. The values accepted for each field depend on the type of field that's being filled out:
- `Bo [intent=reverse_etl availability=partial write=create_offer]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby offer.create through the documented POST /offer.create endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed. This command has nested object/array requirements that are implemented by the reverse-ETL action schema but are not fully expressible as scalar CLI flags; use file/warehouse reverse-ETL inputs for execution.; flags: --offer-process-id, --offer-form-id, --offer-form-field-submissions-0-path, --exclude-form-definition
  - offer list - Get a list of all offers with their latest version.

See the [Pagination and Incremental Synchronization](/docs/pagination-and-incremental-sync) guide for detai [intent=etl availability=implemented stream=offer_list]; notes: Fixed Ashby stream for offer.list; flags map only to documented request body fields.; flags: --created-after, --offer-status, --acceptance-status, --application-id, --approval-status
  - offer info - Returns details about a single offer by id

**Requires the [`offersRead`](authentication#permissions-offerinfo) permission.** [intent=etl availability=implemented stream=offer_info]; notes: Fixed Ashby stream for offer.info; flags map only to documented request body fields.; flags: --offer-id, --exclude-form-definition
  - offer start - The offer.start endpoint creates and returns an offer version instance that can be filled out and submitted
using the `offer.create` endpoint.

In order to crea [intent=reverse_etl availability=implemented write=start_offer]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby offer.start through the documented POST /offer.start endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --offer-process-id
  - offer start-approval-process - Starts the approval process for an offer in a "WaitingOnApprovalStart" state.
Once started, the approval is sent to the configured approvers.

The offer version [intent=reverse_etl availability=implemented write=start_offer_approval_process]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby offer.startApprovalProcess through the documented POST /offer.startApprovalProcess endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --offer-version-id, --note, --exclude-form-definition
  - offer update - Updates an existing Offer

Offer forms support a variety of field types. The values accepted for each field depend on the type of field that's being filled out: [intent=reverse_etl availability=partial write=update_offer]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby offer.update through the documented POST /offer.update endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed. This command has nested object/array requirements that are implemented by the reverse-ETL action schema but are not fully expressible as scalar CLI flags; use file/warehouse reverse-ETL inputs for execution.; flags: --offer-id, --offer-form-field-submissions-0-path, --exclude-form-definition
  - offer set-status - Updates an offer's acceptance status.

Ashby derives the offer status from the provided acceptance status; `offerStatus` can't be set independently.

**Requires [intent=reverse_etl availability=implemented write=set_offer_status]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby offer.setStatus through the documented POST /offer.setStatus endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --offer-id, --acceptance-status, --exclude-form-definition
  - offer set-decided-at - Updates an offer's decidedAt timestamp.

**Requires the [`offersWrite`](authentication#permissions-offersetdecidedat) permission.** [intent=reverse_etl availability=implemented write=set_offer_decided_at]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby offer.setDecidedAt through the documented POST /offer.setDecidedAt endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --offer-id, --decided-at, --exclude-form-definition
  - offer-process start - Starts an offer process for a candidate.

**Requires the [`offersWrite`](authentication#permissions-offerprocessstart) permission.** [intent=reverse_etl availability=implemented write=start_offer_process]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby offerProcess.start through the documented POST /offerProcess.start endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --application-id
  - opening info - Retrieves an opening by its UUID.

**Requires the [`jobsRead`](authentication#permissions-openinginfo) permission.** [intent=etl availability=implemented stream=opening_info]; notes: Fixed Ashby stream for opening.info; flags map only to documented request body fields.; flags: --opening-id
  - opening create - Creates an opening.

To set values of custom fields on Openings, use the [`customField.setValue`](ref:customfieldsetvalue) endpoint.

**Requires the [`jobsWrite [intent=reverse_etl availability=implemented write=create_opening]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby opening.create through the documented POST /opening.create endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --identifier, --description, --team-id, --location-ids, --job-ids, --target-hire-date, --target-start-date, --is-backfill, --employment-type, --opening-state
  - opening update - Updates an opening.

To set values for custom fields on Openings, use the [`customField.setValue`](ref:customfieldsetvalue) endpoint.

**Requires the [`jobsWrit [intent=reverse_etl availability=implemented write=update_opening]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby opening.update through the documented POST /opening.update endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --opening-id, --identifier, --description, --team-id, --target-hire-date, --target-start-date, --is-backfill, --employment-type
  - opening set-archived - Sets the archived state of an opening.

**Requires the [`jobsWrite`](authentication#permissions-openingsetarchived) permission.** [intent=reverse_etl availability=implemented write=set_opening_archived]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby opening.setArchived through the documented POST /opening.setArchived endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --opening-id, --archive
  - opening set-opening-state - Sets the state of an opening.

**Requires the [`jobsWrite`](authentication#permissions-openingsetopeningstate) permission.** [intent=reverse_etl availability=implemented write=set_opening_opening_state]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby opening.setOpeningState through the documented POST /opening.setOpeningState endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --opening-id, --opening-state, --close-reason-id
  - opening add-job - Adds a job to an opening.

**Requires the [`jobsWrite`](authentication#permissions-openingaddjob) permission.** [intent=reverse_etl availability=implemented write=add_opening_job]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby opening.addJob through the documented POST /opening.addJob endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --opening-id, --job-id
  - opening remove-job - Removes a job from an opening.

**Requires the [`jobsWrite`](authentication#permissions-openingremovejob) permission.** [intent=reverse_etl availability=implemented write=remove_opening_job]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby opening.removeJob through the documented POST /opening.removeJob endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --opening-id, --job-id
  - opening add-location - Adds a location to an opening.

**Requires the [`jobsWrite`](authentication#permissions-openingaddlocation) permission.** [intent=reverse_etl availability=implemented write=add_opening_location]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby opening.addLocation through the documented POST /opening.addLocation endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --opening-id, --location-id
  - opening remove-location - Removes a location from an opening.

**Requires the [`jobsWrite`](authentication#permissions-openingremovelocation) permission.** [intent=reverse_etl availability=implemented write=remove_opening_location]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby opening.removeLocation through the documented POST /opening.removeLocation endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --opening-id, --location-id
  - opening list - Lists openings.

See the [Pagination and Incremental Synchronization](/docs/pagination-and-incremental-sync) guide for detailed usage examples.

**Requires the  [intent=etl availability=implemented stream=opening_list]; notes: Fixed Ashby stream for opening.list; flags map only to documented request body fields.; flags: --created-after
  - opening search - Searches for openings by identifier.

**Requires the [`jobsRead`](authentication#permissions-openingsearch) permission.** [intent=direct_read availability=implemented operation=ashby.direct.opening.search]; approval: none; risk: bounded JSON direct read; credential-marked response fields are redacted, and non-credential identity fields remain complete in trusted live local output; notes: Fixed Ashby POST direct read; no raw method/path/body override is exposed.; flags: --identifier (required)
  - project info - Retrieves a project by its UUID.

**Requires the [`candidatesRead`](authentication#permissions-projectinfo) permission.** [intent=etl availability=implemented stream=project_info]; notes: Fixed Ashby stream for project.info; flags map only to documented request body fields.; flags: --project-id
  - project list - Lists projects.

See the [Pagination and Incremental Synchronization](/docs/pagination-and-incremental-sync) guide for detailed usage examples.

**Requires the  [intent=etl availability=implemented stream=project_list]; notes: Fixed Ashby stream for project.list; flags map only to documented request body fields.; flags: --created-after
  - project search - Search for projects by title.

Responses are limited to 100 results. Consider refining your search or using /project.list to paginate through all projects, if y [intent=direct_read availability=implemented operation=ashby.direct.project.search]; approval: none; risk: bounded JSON direct read; credential-marked response fields are redacted, and non-credential identity fields remain complete in trusted live local output; notes: Fixed Ashby POST direct read; no raw method/path/body override is exposed.; flags: --title (required)
  - source list - List all sources

**Requires the [`hiringProcessMetadataRead`](authentication#permissions-sourcelist) permission.** [intent=etl availability=implemented stream=source_list]; notes: Fixed Ashby stream for source.list; flags map only to documented request body fields.; flags: --include-archived
  - source-tracking-link list - List all source custom tracking links

**Requires the [`hiringProcessMetadataRead`](authentication#permissions-sourcetrackinglinklist) permission.** [intent=etl availability=implemented stream=source_tracking_link_list]; notes: Fixed Ashby stream for sourceTrackingLink.list; flags map only to documented request body fields.; flags: --include-disabled, --source-id
  - archive-reason list - Lists archive reasons.

**Requires the [`hiringProcessMetadataRead`](authentication#permissions-archivereasonlist) permission.** [intent=etl availability=implemented stream=archive_reason_list]; notes: Fixed Ashby stream for archiveReason.list; flags map only to documented request body fields.; flags: --include-archived
  - brand list - Lists all brands for the organization.

**Requires the [`organizationRead`](authentication#permissions-brandlist) permission.** [intent=etl availability=implemented stream=brand_list]; notes: Fixed Ashby stream for brand.list; flags map only to documented request body fields.
  - custom-field list - Lists all custom fields.

See the [Pagination and Incremental Synchronization](/docs/pagination-and-incremental-sync) guide for detailed usage examples.

**Requ [intent=etl availability=implemented stream=custom_field_list]; notes: Fixed Ashby stream for customField.list; flags map only to documented request body fields.; flags: --include-archived
  - custom-field create - Create a new custom field.

**Requires the [`hiringProcessMetadataWrite`](authentication#permissions-customfieldcreate) permission.** [intent=reverse_etl availability=implemented write=create_custom_field]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby customField.create through the documented POST /customField.create endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --field-type, --object-type, --title, --description, --selectable-values, --is-date-only-field, --is-exposable-to-candidate, --is-private
  - custom-field set-value - Set the value of a custom field for a given object.

**Note:** When updating multiple custom fields on the same object, use
[`customField.setValues`](#operation [intent=reverse_etl availability=partial write=set_custom_field_value]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby customField.setValue through the documented POST /customField.setValue endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed. This command has a documented fieldValue union/nested input that is implemented by the reverse-ETL action schema but is not safely expressible as scalar CLI flags; use file/warehouse reverse-ETL inputs for execution.; flags: --object-id, --object-type, --field-id
  - custom-field info - Get information about a custom field.

**Requires the [`hiringProcessMetadataRead`](authentication#permissions-customfieldinfo) permission.** [intent=etl availability=implemented stream=custom_field_info]; notes: Fixed Ashby stream for customField.info; flags map only to documented request body fields.; flags: --custom-field-id
  - custom-field set-values - Set the values of multiple custom fields for a given object in a single call.
This is the recommended approach when updating multiple fields on the same object
 [intent=reverse_etl availability=partial write=set_custom_field_values]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby customField.setValues through the documented POST /customField.setValues endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed. This command has a documented fieldValue union/nested input that is implemented by the reverse-ETL action schema but is not safely expressible as scalar CLI flags; use file/warehouse reverse-ETL inputs for execution.; flags: --object-id, --object-type, --values-0-field-id
  - custom-field update-selectable-values - Update the selectable values for a custom field.

This endpoint merges the provided selectable values with the existing values for a custom field.

**Merge beha [intent=reverse_etl availability=implemented write=update_custom_field_selectable_values]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby customField.updateSelectableValues through the documented POST /customField.updateSelectableValues endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --custom-field-id, --selectable-values-0-label, --selectable-values-0-value
  - assessment update - Update Ashby about the status of a started assessment.

`assessment_status` is required unless `cancelled_reason` is provided.

**Requires the [`candidatesWrite [intent=reverse_etl availability=implemented write=update_assessment]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby assessment.update through the documented POST /assessment.update endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --assessment-id, --timestamp, --metadata
  - assessment add-completed-to-candidate - Add a completed assessment to a candidate.

**Requires the [`candidatesWrite`](authentication#permissions-assessmentaddcompletedtocandidate) permission.** [intent=reverse_etl availability=implemented write=add_assessment_completed_to_candidate]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby assessment.addCompletedToCandidate through the documented POST /assessment.addCompletedToCandidate endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --candidate-id, --partner-id, --assessment-assessment-type-id, --assessment-assessment-id, --assessment-assessment-name, --assessment-result-identifier, --assessment-result-label, --assessment-result-type, --assessment-result-value, --assessment-metadata-0-identifier, --assessment-metadata-0-label, --assessment-metadata-0-type, --assessment-metadata-0-value, --timestamp
  - hiring-team add-member - Adds an Ashby user to the hiring team at the application, job, or opening level.

**Requires the [`organizationWrite`](authentication#permissions-hiringteamaddm [intent=reverse_etl availability=implemented write=add_hiring_team_member]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby hiringTeam.addMember through the documented POST /hiringTeam.addMember endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Provide an applicationId, jobId, or openingId target. Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --application-id, --job-id, --opening-id, --team-member-id, --role-id
  - hiring-team remove-member - Removes an Ashby user from the hiring team at the application, job, or opening level.

**Requires the [`organizationWrite`](authentication#permissions-hiringtea [intent=reverse_etl availability=implemented write=remove_hiring_team_member]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby hiringTeam.removeMember through the documented POST /hiringTeam.removeMember endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Provide an applicationId, jobId, or openingId target. Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --application-id, --job-id, --opening-id, --team-member-id, --role-id
  - hiring-team-role list - Lists the possible hiring team roles in an organization

**Requires the [`organizationRead`](authentication#permissions-hiringteamrolelist) permission.** [intent=etl availability=implemented stream=hiring_team_role_list]; notes: Fixed Ashby stream for hiringTeamRole.list; defaults to namesOnly=true role-title results. namesOnly=false object results are blocked pending variant-schema foundation ashby_hiring_team_role_list_names_only_false.
  - user info - Retrieves detailed information about a specific user by their ID.

**Requires the [`organizationRead`](authentication#permissions-userinfo) permission.** [intent=etl availability=implemented stream=user_info]; notes: Fixed Ashby stream for user.info; flags map only to documented request body fields.; flags: --user-id
  - user search - Searches for users by email address.

Returns an array containing the user if found, or an empty array if no user with the given email exists.

**Requires the [ [intent=direct_read availability=implemented operation=ashby.direct.user.search]; approval: none; risk: bounded JSON direct read; credential-marked response fields are redacted, and non-credential identity fields remain complete in trusted live local output; notes: Fixed Ashby POST direct read; no raw method/path/body override is exposed.; flags: --email (required)
  - user interviewer-settings - Get interviewer settings for a user.

**Requires the [`organizationRead`](authentication#permissions-userinterviewersettings) permission.** [intent=reverse_etl availability=implemented write=interviewer_user_settings]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby user.interviewerSettings through the documented POST /user.interviewerSettings endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --user-id
  - user update-interviewer-settings - Update interviewer settings for a user.

Either limit can be provided, or both can be provided. If only one is provided, the other will remain unchanged. If a l [intent=reverse_etl availability=implemented write=update_user_interviewer_settings]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby user.updateInterviewerSettings through the documented POST /user.updateInterviewerSettings endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --user-id, --daily-limit, --weekly-limit
  - user create-interviewer-pause - Creates an interviewer pause for a user. While paused, the user will not be scheduled for interviews.

A user can only have one interviewer pause at a time (whe [intent=reverse_etl availability=implemented write=create_user_interviewer_pause]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby user.createInterviewerPause through the documented POST /user.createInterviewerPause endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --user-id, --starts-at, --ends-at, --comment
  - user list-interviewer-pauses - Lists all active or scheduled interviewer pauses for a user.

**Requires the [`organizationRead`](authentication#permissions-userlistinterviewerpauses) permissi [intent=etl availability=implemented stream=user_list_interviewer_pauses]; notes: Fixed Ashby stream for user.listInterviewerPauses; flags map only to documented request body fields.; flags: --user-id
  - user delete-interviewer-pause - Deletes an interviewer pause.

**Requires the [`organizationWrite`](authentication#permissions-userdeleteinterviewerpause) permission.** [intent=reverse_etl availability=implemented write=delete_user_interviewer_pause]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby user.deleteInterviewerPause through the documented POST /user.deleteInterviewerPause endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --interviewer-pause-id
  - user set-custom-field-value - Set the value of a custom field on an employee.

The values accepted in the `fieldValue` param depend on the type of field being updated. See the [customField.s [intent=reverse_etl availability=partial write=set_user_custom_field_value]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby user.setCustomFieldValue through the documented POST /user.setCustomFieldValue endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed. This command has a documented fieldValue union/nested input that is implemented by the reverse-ETL action schema but is not safely expressible as scalar CLI flags; use file/warehouse reverse-ETL inputs for execution.; flags: --user-id, --field-id
  - user set-custom-field-values - Set the values of multiple custom fields on an employee in a single call.
This is the recommended approach when updating multiple fields on the same employee
to [intent=reverse_etl availability=partial write=set_user_custom_field_values]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby user.setCustomFieldValues through the documented POST /user.setCustomFieldValues endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed. This command has a documented fieldValue union/nested input that is implemented by the reverse-ETL action schema but is not safely expressible as scalar CLI flags; use file/warehouse reverse-ETL inputs for execution.; flags: --user-id, --values-0-field-id
  - referral create - Creates a referral

**Requires the [`candidatesWrite`](authentication#permissions-referralcreate) permission.** [intent=reverse_etl availability=partial write=create_referral]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby referral.create through the documented POST /referral.create endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed. This command has nested object/array requirements that are implemented by the reverse-ETL action schema but are not fully expressible as scalar CLI flags; use file/warehouse reverse-ETL inputs for execution.; flags: --id, --credited-to-user-id, --field-submissions-0-path, --created-at
  - referral-form info - Fetches the default referral form or creates a default referral form if none exists.

**Requires the [`hiringProcessMetadataRead`](authentication#permissions-re [intent=etl availability=implemented stream=referral_form_info]; notes: Fixed Ashby stream for referralForm.info; flags map only to documented request body fields.
  - email-sender list - > Beta
>
> This endpoint is in beta and may not be available for all organizations.

Lists the email addresses available to the acting identity for sending emai [intent=etl availability=implemented stream=email_sender_list]; notes: Fixed Ashby stream for emailSender.list; flags map only to documented request body fields.
  - sequence add - > Beta
>
> This endpoint is in beta and may not be available for all organizations.

Enrolls a candidate in a reusable sourcing sequence. Set start to false to  [intent=reverse_etl availability=implemented write=add_sequence]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby sequence.add through the documented POST /sequence.add endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --sequence-template-id, --candidate-id, --start, --application-id
  - sequence cancel - Cancels a running sourcing sequence (campaign) for a candidate.

**Requires the [`sourcingWrite`](authentication#permissions-sequencecancel) permission.** [intent=reverse_etl availability=implemented write=cancel_sequence]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby sequence.cancel through the documented POST /sequence.cancel endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --sequence-id, --reason
  - sequence discard - > Beta
>
> This endpoint is in beta and may not be available for all organizations.

Permanently discards a NotStarted sourcing sequence draft owned by the acti [intent=reverse_etl availability=implemented write=discard_sequence]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby sequence.discard through the documented POST /sequence.discard endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --sequence-id
  - sequence info - > Beta
>
> This endpoint is in beta and may not be available for all organizations.

Retrieves a candidate's sourcing sequence enrollment when it is visible to  [intent=etl availability=implemented stream=sequence_info]; notes: Fixed Ashby stream for sequence.info; flags map only to documented request body fields.; flags: --sequence-id
  - sequence list - > Beta
>
> This endpoint is in beta and may not be available for all organizations.

Lists visible started and ended sourcing sequence enrollments across the or [intent=etl availability=implemented stream=sequence_list]; notes: Fixed Ashby stream for sequence.list; flags map only to documented request body fields.; flags: --candidate-id
  - sequence update-stage - > Beta
>
> This endpoint is in beta and may not be available for all organizations.

Replaces the supplied subject or HTML body of one email stage in a not-star [intent=reverse_etl availability=implemented write=update_sequence_stage]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby sequence.updateStage through the documented POST /sequence.updateStage endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --sequence-id, --stage-id, --subject, --body-html
  - sequence start - > Beta
>
> This endpoint is in beta and may not be available for all organizations.

Validates and starts an existing NotStarted sourcing sequence draft owned b [intent=reverse_etl availability=implemented write=start_sequence]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby sequence.start through the documented POST /sequence.start endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --sequence-id
  - sequence-template info - Retrieves metadata for a reusable sourcing sequence template visible to the caller. Archived templates may be returned. Message subjects and bodies are not expo [intent=etl availability=implemented stream=sequence_template_info]; notes: Fixed Ashby stream for sequenceTemplate.info; flags map only to documented request body fields.; flags: --sequence-template-id
  - sequence-template list - Lists reusable sourcing sequence templates visible to the caller. Returns template and cadence metadata only; message subjects and bodies are not exposed.

Arch [intent=etl availability=implemented stream=sequence_template_list]; notes: Fixed Ashby stream for sequenceTemplate.list; flags map only to documented request body fields.; flags: --include-archived
  - interview-schedule list - Gets all interview schedules in the organization.

See the [Pagination and Incremental Synchronization](/docs/pagination-and-incremental-sync) guide for detaile [intent=etl availability=implemented stream=interview_schedule_list]; notes: Fixed Ashby stream for interviewSchedule.list; flags map only to documented request body fields.; flags: --application-id, --interview-stage-id, --created-after
  - interview-schedule create - Create a scheduled interview in Ashby.

**Requires the [`interviewsWrite`](authentication#permissions-interviewschedulecreate) permission.** [intent=reverse_etl availability=implemented write=create_interview_schedule]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby interviewSchedule.create through the documented POST /interviewSchedule.create endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --application-id, --interview-events-0-start-time, --interview-events-0-end-time, --interview-events-0-interviewers-0-email
  - interview-schedule update - Update an interview schedule. This endpoint allows you to add, cancel, or update interview events associated with an interview schedule.

In order to update an  [intent=reverse_etl availability=implemented write=update_interview_schedule]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby interviewSchedule.update through the documented POST /interviewSchedule.update endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required for archive/close/cancel-capable request values.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --interview-schedule-id, --interview-event-id-to-cancel, --allow-feedback-deletion
  - interview-schedule cancel - Cancel an interview schedule by id.

**Requires the [`interviewsWrite`](authentication#permissions-interviewschedulecancel) permission.** [intent=reverse_etl availability=implemented write=cancel_interview_schedule]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby interviewSchedule.cancel through the documented POST /interviewSchedule.cancel endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --id, --allow-reschedule
  - take-home-assignment list - > Beta
>
> This endpoint is in beta and may not be available for all organizations.

Lists take-home assignments visible to the caller, including candidate subm [intent=etl availability=implemented stream=take_home_assignment_list]; notes: Fixed Ashby stream for takeHomeAssignment.list; flags map only to documented request body fields.; flags: --application-id, --candidate-id, --expand
  - take-home-assignment info - > Beta
>
> This endpoint is in beta and may not be available for all organizations.

Retrieves a single take-home assignment by id and links it to its interview [intent=etl availability=implemented stream=take_home_assignment_info]; notes: Fixed Ashby stream for takeHomeAssignment.info; flags map only to documented request body fields.; flags: --take-home-assignment-id, --expand
  - interview-event list - Lists interview events associated with an interview schedule.

See the [Pagination and Incremental Synchronization](/docs/pagination-and-incremental-sync) guide [intent=etl availability=implemented stream=interview_event_list]; notes: Fixed Ashby stream for interviewEvent.list; flags map only to documented request body fields.; flags: --interview-schedule-id, --expand, --created-after
  - interview-briefing info - Fetch the briefing data for an interview event. Returns the application,
interview, per-interviewer status, and the feedback form definition id
needed to render [intent=etl availability=implemented stream=interview_briefing_info]; notes: Fixed Ashby stream for interviewBriefing.info; flags map only to documented request body fields.; flags: --interview-event-id, --expand
  - interview info - Fetch interview details by id.

**Requires the [`interviewsRead`](authentication#permissions-interviewinfo) permission.** [intent=etl availability=implemented stream=interview_info]; notes: Fixed Ashby stream for interview.info; flags map only to documented request body fields.; flags: --id
  - interview list - List all interviews.

See the [Pagination and Incremental Synchronization](/docs/pagination-and-incremental-sync) guide for detailed usage examples.

**Requires [intent=etl availability=implemented stream=interview_list]; notes: Fixed Ashby stream for interview.list; flags map only to documented request body fields.; flags: --include-archived, --include-non-shared-interviews, --exclude-archived-schedule-template-interviews
  - interview-stage info - Retrieves detailed information about a specific interview stage by its ID.

**Requires the [`interviewsRead`](authentication#permissions-interviewstageinfo) per [intent=etl availability=implemented stream=interview_stage_info]; notes: Fixed Ashby stream for interviewStage.info; flags map only to documented request body fields.; flags: --interview-stage-id
  - file info - Retrieve the URL for a file referenced by a public API file handle (candidate files, resumes, offer letters, and signature-request files).

**Please note** that [intent=direct_read availability=implemented operation=ashby.direct.file.info]; approval: none; risk: bounded JSON direct read; credential-marked response fields are redacted, and Ashby signed URL fields are preserved (results.url/results.transcriptUrl) in trusted live local output; notes: Fixed Ashby POST direct read; no raw method/path/body override is exposed.; flags: --file-handle (required)
  - survey-form-definition info - Returns details about a single survey form definition by id.

**Requires the [`hiringProcessMetadataRead`](authentication#permissions-surveyformdefinitioninfo)  [intent=etl availability=implemented stream=survey_form_definition_info]; notes: Fixed Ashby stream for surveyFormDefinition.info; flags map only to documented request body fields.; flags: --survey-form-definition-id
  - survey-form-definition list - Lists all survey form definitions.

See the [Pagination and Incremental Synchronization](/docs/pagination-and-incremental-sync) guide for detailed usage example [intent=etl availability=implemented stream=survey_form_definition_list]; notes: Fixed Ashby stream for surveyFormDefinition.list; flags map only to documented request body fields.
  - survey-request create - This endpoint generates a survey request and returns a survey URL. You can send this URL to a candidate to allow them to complete a survey.

**Note that calling [intent=reverse_etl availability=implemented write=create_survey_request]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby surveyRequest.create through the documented POST /surveyRequest.create endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --candidate-id, --application-id, --survey-form-definition-id
  - survey-request list - Lists all survey requests.

See the [Pagination and Incremental Synchronization](/docs/pagination-and-incremental-sync) guide for detailed usage examples.

**Re [intent=etl availability=implemented stream=survey_request_list]; notes: Fixed Ashby stream for surveyRequest.list; flags map only to documented request body fields.; flags: --survey-type, --application-id, --candidate-id, --created-after
  - survey-submission create - Creates a survey submission for a candidate and application.

**Requires the [`candidatesWrite`](authentication#permissions-surveysubmissioncreate) permission.* [intent=reverse_etl availability=partial write=create_survey_submission]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby surveySubmission.create through the documented POST /surveySubmission.create endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed. This command has nested object/array requirements that are implemented by the reverse-ETL action schema but are not fully expressible as scalar CLI flags; use file/warehouse reverse-ETL inputs for execution.; flags: --survey-form-definition-id, --candidate-id, --application-id
  - survey-submission list - Lists all survey submissions of a given `surveyType`.

See the [Pagination and Incremental Synchronization](/docs/pagination-and-incremental-sync) guide for det [intent=etl availability=implemented stream=survey_submission_list]; notes: Fixed Ashby stream for surveySubmission.list; flags map only to documented request body fields.; flags: --survey-type, --created-after
  - webhook create - Creates a webhook setting.

**Requires the [`apiKeysWrite`](authentication#permissions-webhookcreate) permission.** [intent=reverse_etl availability=implemented write=create_webhook]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby webhook.create through the documented POST /webhook.create endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --webhook-type, --request-url, --secret-token
  - webhook info - Retrieves information about a specific webhook setting by its ID.

**Requires the [`apiKeysRead`](authentication#permissions-webhookinfo) permission.** [intent=etl availability=implemented stream=webhook_info]; notes: Fixed Ashby stream for webhook.info; flags map only to documented request body fields.; flags: --webhook-id
  - webhook update - Updates a webhook setting. One of `enabled`, `requestUrl`, or `secretToken` must be provided.

**Requires the [`apiKeysWrite`](authentication#permissions-webhoo [intent=reverse_etl availability=implemented write=update_webhook]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby webhook.update through the documented POST /webhook.update endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --webhook-id, --enabled, --request-url, --secret-token
  - webhook delete - Deletes a webhook setting.

**Requires the [`apiKeysWrite`](authentication#permissions-webhookdelete) permission.** [intent=reverse_etl availability=implemented write=delete_webhook]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby webhook.delete through the documented POST /webhook.delete endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --webhook-id
  - interviewer-pool list - List all interviewer pools.

See the [Pagination and Incremental Synchronization](/docs/pagination-and-incremental-sync) guide for detailed usage examples.

**R [intent=etl availability=implemented stream=interviewer_pool_list]; notes: Fixed Ashby stream for interviewerPool.list; flags map only to documented request body fields.; flags: --include-archived-pools, --include-archived-training-stages
  - interviewer-pool info - Get information about an interviewer pool.

**Requires the [`hiringProcessMetadataRead`](authentication#permissions-interviewerpoolinfo) permission.** [intent=etl availability=implemented stream=interviewer_pool_info]; notes: Fixed Ashby stream for interviewerPool.info; flags map only to documented request body fields.; flags: --interviewer-pool-id
  - interviewer-pool archive - Archives an interviewer pool.

**Requires the [`hiringProcessMetadataWrite`](authentication#permissions-interviewerpoolarchive) permission.** [intent=reverse_etl availability=implemented write=archive_interviewer_pool]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby interviewerPool.archive through the documented POST /interviewerPool.archive endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --interviewer-pool-id
  - interviewer-pool restore - Restores an archived interviewer pool.

**Requires the [`hiringProcessMetadataWrite`](authentication#permissions-interviewerpoolrestore) permission.** [intent=reverse_etl availability=implemented write=restore_interviewer_pool]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby interviewerPool.restore through the documented POST /interviewerPool.restore endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --interviewer-pool-id
  - interviewer-pool create - Create an interviewer pool.

**Requires the [`hiringProcessMetadataWrite`](authentication#permissions-interviewerpoolcreate) permission.** [intent=reverse_etl availability=implemented write=create_interviewer_pool]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby interviewerPool.create through the documented POST /interviewerPool.create endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --title, --requires-training
  - interviewer-pool update - Update an interviewer pool.

**Requires the [`hiringProcessMetadataWrite`](authentication#permissions-interviewerpoolupdate) permission.** [intent=reverse_etl availability=implemented write=update_interviewer_pool]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby interviewerPool.update through the documented POST /interviewerPool.update endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --interviewer-pool-id, --title, --requires-training
  - interviewer-pool add-user - Add a user to an interviewer pool.

**Requires the [`hiringProcessMetadataWrite`](authentication#permissions-interviewerpooladduser) permission.** [intent=reverse_etl availability=implemented write=add_interviewer_pool_user]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby interviewerPool.addUser through the documented POST /interviewerPool.addUser endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --interviewer-pool-id, --user-id, --interviewer-pool-training-path-stage-id
  - interviewer-pool remove-user - Remove a user from an interviewer pool.

**Requires the [`hiringProcessMetadataWrite`](authentication#permissions-interviewerpoolremoveuser) permission.** [intent=reverse_etl availability=implemented write=remove_interviewer_pool_user]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby interviewerPool.removeUser through the documented POST /interviewerPool.removeUser endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --interviewer-pool-id, --user-id
  - close-reason list - Lists all close reasons for jobs or openings.

**Requires the [`hiringProcessMetadataRead`](authentication#permissions-closereasonlist) permission.** [intent=etl availability=implemented stream=close_reason_list]; notes: Fixed Ashby stream for closeReason.list; flags map only to documented request body fields.; flags: --include-archived
  - report generate - > Beta
>
> This endpoint is currently in beta and may change without notice.

Generates a new report or polls the status of an existing report generation.

**Tw [intent=reverse_etl availability=implemented write=generate_report]; approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; risk: Executes Ashby report.generate through the documented POST /report.generate endpoint; reverse ETL plan, preview, approval, and execute are required.; notes: Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed.; flags: --report-id, --include-headers-in-data, --result-style, --request-id
  - report synchronous - > Beta
>
> This endpoint is currently in beta and may change without notice.

Retrieves report data synchronously.

**Timeout:** 30 seconds. If a report is timi [intent=etl availability=implemented stream=report_synchronous]; notes: Fixed Ashby stream for report.synchronous; flags map only to documented request body fields.; flags: --report-id, --include-headers-in-data, --result-style
  - approval list - Gets all approvals in the organization. You can optionally filter by entity type and entity ID.

See the [Pagination and Incremental Synchronization](/docs/pagi [intent=etl availability=implemented stream=approval_list]; notes: Fixed Ashby stream for approval.list; flags map only to documented request body fields.; flags: --entity-type, --entity-id
- Help topics:
  - ashby safety - Ashby writes are named, schema-validated actions only; reverse ETL must use plan, preview, explicit approval, and execute.
  - ashby parity - Public Ashby OpenAPI coverage ledger is recorded in api_surface.json with blocked webhook/partner/binary workflow reasons.

## Commands

### Inspect as a manual

```bash
pm connectors inspect ashby
```

### Inspect as structured JSON

```bash
pm connectors inspect ashby --json
```

## Agent Rules

- Run pm connectors inspect ashby before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
