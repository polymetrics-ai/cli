# Ashby Connector

## Overview

Reads Ashby applicant-tracking REST resources and exposes reviewed reverse-ETL/direct-read surfaces from the official Ashby OpenAPI.

Readable streams: `candidates`, `jobs`, `applications`, `users`, `api_key_info`, `audit_log_list`, `application_info`, `application_hiring_team_role_list`, `application_list_history`, `application_list_criteria_evaluations`, `application_feedback_list`, `candidate_info`, `candidate_list_client_info`, `candidate_list_fraud_checks`, `candidate_list_notes`, `candidate_list_projects`, `candidate_tag_list`, `communication_template_list`, `feedback_form_definition_list`, `feedback_form_definition_info`, `job_posting_list`, `job_posting_info`, `job_info`, `job_board_list`, `job_interview_plan_info`, `job_template_list`, `department_info`, `department_list`, `location_list`, `location_info`, `interview_plan_list`, `interview_stage_list`, `interview_stage_group_list`, `offer_list`, `offer_info`, `opening_info`, `opening_list`, `project_info`, `project_list`, `source_list`, `source_tracking_link_list`, `archive_reason_list`, `brand_list`, `custom_field_list`, `custom_field_info`, `hiring_team_role_list`, `user_info`, `user_list_interviewer_pauses`, `email_sender_list`, `sequence_info`, `sequence_list`, `sequence_template_info`, `sequence_template_list`, `interview_schedule_list`, `take_home_assignment_list`, `take_home_assignment_info`, `interview_event_list`, `interview_briefing_info`, `interview_info`, `interview_list`, `interview_stage_info`, `survey_form_definition_info`, `survey_form_definition_list`, `survey_request_list`, `survey_submission_list`, `webhook_info`, `interviewer_pool_list`, `interviewer_pool_info`, `close_reason_list`, `report_synchronous`, `approval_list`.

Service API documentation: https://developers.ashbyhq.com/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Ashby API key. Provide from an environment variable or stdin; never inline in prompts or docs.

Authentication uses declared mode(s): `basic`.

## Execution contract

Connection check: `POST /candidate.list`
Check JSON body: `limit`=1.

## Streams notes

- `candidates`: `POST /candidate.list`; records `results`
  - JSON body: `createdAfter`={{ query.createdAfter }}, `createdBefore`={{ query.createdBefore }}, `limit`=100.
  - Pagination: `cursor`.
- `jobs`: `POST /job.list`; records `results`
  - JSON body: `closedAfter`={{ query.closedAfter }}, `closedBefore`={{ query.closedBefore }}, `createdAfter`={{ query.createdAfter }}, `includeUnpublishedJobPostingsIds`={{ query.includeUnpublishedJobPostingsIds }}, `limit`=100, `openedAfter`={{ query.openedAfter }}, `openedBefore`={{ query.openedBefore }}.
  - Pagination: `cursor`.
- `applications`: `POST /application.list`; records `results`
  - JSON body: `createdAfter`={{ query.createdAfter }}, `createdBefore`={{ query.createdBefore }}, `jobId`={{ query.jobId }}, `limit`=100, `status`={{ query.status }}.
  - Pagination: `cursor`.
- `users`: `POST /user.list`; records `results`
  - JSON body: `includeDeactivated`={{ query.includeDeactivated }}, `limit`=100.
  - Pagination: `cursor`.
- `api_key_info`: `POST /apiKey.info`; records `results`
- `audit_log_list`: `POST /auditLog.list`; records `results`
  - JSON body: `endDate`={{ query.endDate }}, `limit`=100, `startDate`={{ query.startDate }}.
  - Pagination: `cursor`.
- `application_info`: `POST /application.info`; records `results`
  - JSON body: `applicationId`={{ query.applicationId }}, `submittedFormInstanceId`={{ query.submittedFormInstanceId }}.
- `application_hiring_team_role_list`: `POST /applicationHiringTeamRole.list`; records `results`
  - JSON body: `limit`=100.
  - Pagination: `cursor`.
- `application_list_history`: `POST /application.listHistory`; records `results`
  - JSON body: `applicationId`={{ query.applicationId }}, `limit`=100.
  - Pagination: `cursor`.
- `application_list_criteria_evaluations`: `POST /application.listCriteriaEvaluations`; records `results`
  - JSON body: `applicationId`={{ query.applicationId }}, `limit`=100.
  - Pagination: `cursor`.
- `application_feedback_list`: `POST /applicationFeedback.list`; records `results`
  - JSON body: `applicationId`={{ query.applicationId }}, `createdAfter`={{ query.createdAfter }}, `limit`=100.
  - Pagination: `cursor`.
- `candidate_info`: `POST /candidate.info`; records `results`
  - JSON body: `externalMappingId`={{ query.externalMappingId }}, `id`={{ query.id }}.
- `candidate_list_client_info`: `POST /candidate.listClientInfo`; records `results`
  - JSON body: `candidateId`={{ query.candidateId }}, `limit`=100.
  - Pagination: `cursor`.
- `candidate_list_fraud_checks`: `POST /candidate.listFraudChecks`; records `results`
  - JSON body: `candidateId`={{ query.candidateId }}, `limit`=100.
  - Pagination: `cursor`.
- `candidate_list_notes`: `POST /candidate.listNotes`; records `results`
  - JSON body: `candidateId`={{ query.candidateId }}, `limit`=100.
  - Pagination: `cursor`.
- `candidate_list_projects`: `POST /candidate.listProjects`; records `results`
  - JSON body: `candidateId`={{ query.candidateId }}, `limit`=100.
  - Pagination: `cursor`.
- `candidate_tag_list`: `POST /candidateTag.list`; records `results`
  - JSON body: `includeArchived`={{ query.includeArchived }}, `limit`=100.
  - Pagination: `cursor`.
- `communication_template_list`: `POST /communicationTemplate.list`; records `results`
  - JSON body: `limit`=100.
  - Pagination: `cursor`.
- `feedback_form_definition_list`: `POST /feedbackFormDefinition.list`; records `results`
  - JSON body: `includeArchived`={{ query.includeArchived }}, `limit`=100.
  - Pagination: `cursor`.
- `feedback_form_definition_info`: `POST /feedbackFormDefinition.info`; records `results`
  - JSON body: `feedbackFormDefinitionId`={{ query.feedbackFormDefinitionId }}.
- `job_posting_list`: `POST /jobPosting.list`; records `results`
  - JSON body: `department`={{ query.department }}, `includeUnpublishedJobPostings`={{ query.includeUnpublishedJobPostings }}, `jobBoardId`={{ query.jobBoardId }}, `limit`=100, `listedOnly`={{ query.listedOnly }}, `location`={{ query.location }}.
  - Pagination: `cursor`.
- `job_posting_info`: `POST /jobPosting.info`; records `results`
  - JSON body: `includeUnpublishedJobPostings`={{ query.includeUnpublishedJobPostings }}, `jobBoardId`={{ query.jobBoardId }}, `jobPostingId`={{ query.jobPostingId }}.
- `job_info`: `POST /job.info`; records `results`
  - JSON body: `id`={{ query.id }}, `includeUnpublishedJobPostingsIds`={{ query.includeUnpublishedJobPostingsIds }}.
- `job_board_list`: `POST /jobBoard.list`; records `results`
  - JSON body: `limit`=100.
  - Pagination: `cursor`.
- `job_interview_plan_info`: `POST /jobInterviewPlan.info`; records `results`
  - JSON body: `jobId`={{ query.jobId }}.
- `job_template_list`: `POST /jobTemplate.list`; records `results`
  - JSON body: `limit`=100.
  - Pagination: `cursor`.
- `department_info`: `POST /department.info`; records `results`
  - JSON body: `departmentId`={{ query.departmentId }}.
- `department_list`: `POST /department.list`; records `results`
  - JSON body: `includeArchived`={{ query.includeArchived }}, `limit`=100.
  - Pagination: `cursor`.
- `location_list`: `POST /location.list`; records `results`
  - JSON body: `includeArchived`={{ query.includeArchived }}, `includeLocationHierarchy`={{ query.includeLocationHierarchy }}, `limit`=100.
  - Pagination: `cursor`.
- `location_info`: `POST /location.info`; records `results`
  - JSON body: `locationId`={{ query.locationId }}.
- `interview_plan_list`: `POST /interviewPlan.list`; records `results`
  - JSON body: `includeArchived`={{ query.includeArchived }}, `limit`=100.
  - Pagination: `cursor`.
- `interview_stage_list`: `POST /interviewStage.list`; records `results`
  - JSON body: `interviewPlanId`={{ query.interviewPlanId }}, `limit`=100.
  - Pagination: `cursor`.
- `interview_stage_group_list`: `POST /interviewStageGroup.list`; records `results`
  - JSON body: `limit`=100.
  - Pagination: `cursor`.
- `offer_list`: `POST /offer.list`; records `results`
  - JSON body: `applicationId`={{ query.applicationId }}, `createdAfter`={{ query.createdAfter }}, `limit`=100.
  - Pagination: `cursor`.
- `offer_info`: `POST /offer.info`; records `results`
  - JSON body: `excludeFormDefinition`={{ query.excludeFormDefinition }}, `offerId`={{ query.offerId }}.
- `opening_info`: `POST /opening.info`; records `results`
  - JSON body: `openingId`={{ query.openingId }}.
- `opening_list`: `POST /opening.list`; records `results`
  - JSON body: `createdAfter`={{ query.createdAfter }}, `limit`=100.
  - Pagination: `cursor`.
- `project_info`: `POST /project.info`; records `results`
  - JSON body: `projectId`={{ query.projectId }}.
- `project_list`: `POST /project.list`; records `results`
  - JSON body: `createdAfter`={{ query.createdAfter }}, `limit`=100.
  - Pagination: `cursor`.
- `source_list`: `POST /source.list`; records `results`
  - JSON body: `includeArchived`={{ query.includeArchived }}, `limit`=100.
  - Pagination: `cursor`.
- `source_tracking_link_list`: `POST /sourceTrackingLink.list`; records `results`
  - JSON body: `includeDisabled`={{ query.includeDisabled }}, `limit`=100, `sourceId`={{ query.sourceId }}.
  - Pagination: `cursor`.
- `archive_reason_list`: `POST /archiveReason.list`; records `results`
  - JSON body: `includeArchived`={{ query.includeArchived }}, `limit`=100.
  - Pagination: `cursor`.
- `brand_list`: `POST /brand.list`; records `results`
  - JSON body: `limit`=100.
  - Pagination: `cursor`.
- `custom_field_list`: `POST /customField.list`; records `results`
  - JSON body: `includeArchived`={{ query.includeArchived }}, `limit`=100.
  - Pagination: `cursor`.
- `custom_field_info`: `POST /customField.info`; records `results`
  - JSON body: `customFieldId`={{ query.customFieldId }}.
- `hiring_team_role_list`: `POST /hiringTeamRole.list`; records `results`
  - JSON body: `limit`=100.
  - Pagination: `cursor`.
- `user_info`: `POST /user.info`; records `results`
  - JSON body: `userId`={{ query.userId }}.
- `user_list_interviewer_pauses`: `POST /user.listInterviewerPauses`; records `results`
  - JSON body: `limit`=100, `userId`={{ query.userId }}.
  - Pagination: `cursor`.
- `email_sender_list`: `POST /emailSender.list`; records `results`
  - JSON body: `limit`=100.
  - Pagination: `cursor`.
- `sequence_info`: `POST /sequence.info`; records `results`
  - JSON body: `sequenceId`={{ query.sequenceId }}.
- `sequence_list`: `POST /sequence.list`; records `results`
  - JSON body: `candidateId`={{ query.candidateId }}, `limit`=100.
  - Pagination: `cursor`.
- `sequence_template_info`: `POST /sequenceTemplate.info`; records `results`
  - JSON body: `sequenceTemplateId`={{ query.sequenceTemplateId }}.
- `sequence_template_list`: `POST /sequenceTemplate.list`; records `results`
  - JSON body: `includeArchived`={{ query.includeArchived }}, `limit`=100.
  - Pagination: `cursor`.
- `interview_schedule_list`: `POST /interviewSchedule.list`; records `results`
  - JSON body: `applicationId`={{ query.applicationId }}, `createdAfter`={{ query.createdAfter }}, `interviewStageId`={{ query.interviewStageId }}, `limit`=100.
  - Pagination: `cursor`.
- `take_home_assignment_list`: `POST /takeHomeAssignment.list`; records `results`
  - JSON body: `applicationId`={{ query.applicationId }}, `candidateId`={{ query.candidateId }}, `limit`=100.
  - Pagination: `cursor`.
- `take_home_assignment_info`: `POST /takeHomeAssignment.info`; records `results`
  - JSON body: `takeHomeAssignmentId`={{ query.takeHomeAssignmentId }}.
- `interview_event_list`: `POST /interviewEvent.list`; records `results`
  - JSON body: `createdAfter`={{ query.createdAfter }}, `interviewScheduleId`={{ query.interviewScheduleId }}, `limit`=100.
  - Pagination: `cursor`.
- `interview_briefing_info`: `POST /interviewBriefing.info`; records `results`
  - JSON body: `interviewEventId`={{ query.interviewEventId }}.
- `interview_info`: `POST /interview.info`; records `results`
  - JSON body: `id`={{ query.id }}.
- `interview_list`: `POST /interview.list`; records `results`
  - JSON body: `excludeArchivedScheduleTemplateInterviews`={{ query.excludeArchivedScheduleTemplateInterviews }}, `includeArchived`={{ query.includeArchived }}, `includeNonSharedInterviews`={{ query.includeNonSharedInterviews }}, `limit`=100.
  - Pagination: `cursor`.
- `interview_stage_info`: `POST /interviewStage.info`; records `results`
  - JSON body: `interviewStageId`={{ query.interviewStageId }}.
- `survey_form_definition_info`: `POST /surveyFormDefinition.info`; records `results`
  - JSON body: `surveyFormDefinitionId`={{ query.surveyFormDefinitionId }}.
- `survey_form_definition_list`: `POST /surveyFormDefinition.list`; records `results`
  - JSON body: `limit`=100.
  - Pagination: `cursor`.
- `survey_request_list`: `POST /surveyRequest.list`; records `results`
  - JSON body: `applicationId`={{ query.applicationId }}, `candidateId`={{ query.candidateId }}, `createdAfter`={{ query.createdAfter }}, `limit`=100, `surveyType`={{ query.surveyType }}.
  - Pagination: `cursor`.
- `survey_submission_list`: `POST /surveySubmission.list`; records `results`
  - JSON body: `createdAfter`={{ query.createdAfter }}, `limit`=100, `surveyType`={{ query.surveyType }}.
  - Pagination: `cursor`.
- `webhook_info`: `POST /webhook.info`; records `results`
  - JSON body: `webhookId`={{ query.webhookId }}.
- `interviewer_pool_list`: `POST /interviewerPool.list`; records `results`
  - JSON body: `includeArchivedPools`={{ query.includeArchivedPools }}, `includeArchivedTrainingStages`={{ query.includeArchivedTrainingStages }}, `limit`=100.
  - Pagination: `cursor`.
- `interviewer_pool_info`: `POST /interviewerPool.info`; records `results`
  - JSON body: `interviewerPoolId`={{ query.interviewerPoolId }}.
- `close_reason_list`: `POST /closeReason.list`; records `results`
  - JSON body: `includeArchived`={{ query.includeArchived }}, `limit`=100.
  - Pagination: `cursor`.
- `report_synchronous`: `POST /report.synchronous`; records `results`
  - JSON body: `includeHeadersInData`={{ query.includeHeadersInData }}, `reportId`={{ query.reportId }}, `resultStyle`={{ query.resultStyle }}.
- `approval_list`: `POST /approval.list`; records `results`
  - JSON body: `entityId`={{ query.entityId }}, `entityType`={{ query.entityType }}, `limit`=100.
  - Pagination: `cursor`.

## Write actions & risks

This connector's write surface is declared separately in the rendered execution bundle.


## Declared response errors

- `candidates`: `path`=`errors`, `success_path`=`success`.
- `jobs`: `path`=`errors`, `success_path`=`success`.
- `applications`: `path`=`errors`, `success_path`=`success`.
- `users`: `path`=`errors`, `success_path`=`success`.
- `api_key_info`: `path`=`errors`, `success_path`=`success`.
- `audit_log_list`: `path`=`errors`, `success_path`=`success`.
- `application_info`: `path`=`errors`, `success_path`=`success`.
- `application_hiring_team_role_list`: `path`=`errors`, `success_path`=`success`.
- `application_list_history`: `path`=`errors`, `success_path`=`success`.
- `application_list_criteria_evaluations`: `path`=`errors`, `success_path`=`success`.
- `application_feedback_list`: `path`=`errors`, `success_path`=`success`.
- `candidate_info`: `path`=`errors`, `success_path`=`success`.
- `candidate_list_client_info`: `path`=`errors`, `success_path`=`success`.
- `candidate_list_fraud_checks`: `path`=`errors`, `success_path`=`success`.
- `candidate_list_notes`: `path`=`errors`, `success_path`=`success`.
- `candidate_list_projects`: `path`=`errors`, `success_path`=`success`.
- `candidate_tag_list`: `path`=`errors`, `success_path`=`success`.
- `communication_template_list`: `path`=`errors`, `success_path`=`success`.
- `feedback_form_definition_list`: `path`=`errors`, `success_path`=`success`.
- `feedback_form_definition_info`: `path`=`errors`, `success_path`=`success`.
- `job_posting_list`: `path`=`errors`, `success_path`=`success`.
- `job_posting_info`: `path`=`errors`, `success_path`=`success`.
- `job_info`: `path`=`errors`, `success_path`=`success`.
- `job_board_list`: `path`=`errors`, `success_path`=`success`.
- `job_interview_plan_info`: `path`=`errors`, `success_path`=`success`.
- `job_template_list`: `path`=`errors`, `success_path`=`success`.
- `department_info`: `path`=`errors`, `success_path`=`success`.
- `department_list`: `path`=`errors`, `success_path`=`success`.
- `location_list`: `path`=`errors`, `success_path`=`success`.
- `location_info`: `path`=`errors`, `success_path`=`success`.
- `interview_plan_list`: `path`=`errors`, `success_path`=`success`.
- `interview_stage_list`: `path`=`errors`, `success_path`=`success`.
- `interview_stage_group_list`: `path`=`errors`, `success_path`=`success`.
- `offer_list`: `path`=`errors`, `success_path`=`success`.
- `offer_info`: `path`=`errors`, `success_path`=`success`.
- `opening_info`: `path`=`errors`, `success_path`=`success`.
- `opening_list`: `path`=`errors`, `success_path`=`success`.
- `project_info`: `path`=`errors`, `success_path`=`success`.
- `project_list`: `path`=`errors`, `success_path`=`success`.
- `source_list`: `path`=`errors`, `success_path`=`success`.
- `source_tracking_link_list`: `path`=`errors`, `success_path`=`success`.
- `archive_reason_list`: `path`=`errors`, `success_path`=`success`.
- `brand_list`: `path`=`errors`, `success_path`=`success`.
- `custom_field_list`: `path`=`errors`, `success_path`=`success`.
- `custom_field_info`: `path`=`errors`, `success_path`=`success`.
- `hiring_team_role_list`: `path`=`errors`, `success_path`=`success`.
- `user_info`: `path`=`errors`, `success_path`=`success`.
- `user_list_interviewer_pauses`: `path`=`errors`, `success_path`=`success`.
- `email_sender_list`: `path`=`errors`, `success_path`=`success`.
- `sequence_info`: `path`=`errors`, `success_path`=`success`.
- `sequence_list`: `path`=`errors`, `success_path`=`success`.
- `sequence_template_info`: `path`=`errors`, `success_path`=`success`.
- `sequence_template_list`: `path`=`errors`, `success_path`=`success`.
- `interview_schedule_list`: `path`=`errors`, `success_path`=`success`.
- `take_home_assignment_list`: `path`=`errors`, `success_path`=`success`.
- `take_home_assignment_info`: `path`=`errors`, `success_path`=`success`.
- `interview_event_list`: `path`=`errors`, `success_path`=`success`.
- `interview_briefing_info`: `path`=`errors`, `success_path`=`success`.
- `interview_info`: `path`=`errors`, `success_path`=`success`.
- `interview_list`: `path`=`errors`, `success_path`=`success`.
- `interview_stage_info`: `path`=`errors`, `success_path`=`success`.
- `survey_form_definition_info`: `path`=`errors`, `success_path`=`success`.
- `survey_form_definition_list`: `path`=`errors`, `success_path`=`success`.
- `survey_request_list`: `path`=`errors`, `success_path`=`success`.
- `survey_submission_list`: `path`=`errors`, `success_path`=`success`.
- `webhook_info`: `path`=`errors`, `success_path`=`success`.
- `interviewer_pool_list`: `path`=`errors`, `success_path`=`success`.
- `interviewer_pool_info`: `path`=`errors`, `success_path`=`success`.
- `close_reason_list`: `path`=`errors`, `success_path`=`success`.
- `report_synchronous`: `path`=`errors`, `success_path`=`success`.
- `approval_list`: `path`=`errors`, `success_path`=`success`.
