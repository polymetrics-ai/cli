---
name: pm-keka
description: Keka connector knowledge and safe action guide.
---

# pm-keka

## Purpose

Reads and writes the documented Keka HRMS REST API surface for Core HR, documents, leave, attendance, payroll, PSA, PMS, hire, expense, assets, requisitions, skills, and BGV resources.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- attachment_id
- base_url (required)
- bgv_id
- candidate_id
- client_id (required)
- document_id
- document_type_id
- employee_id
- grant_type
- job_id
- mode
- pay_batch_id
- pay_cycle_id
- pay_group_id
- project_id
- scope
- task_id
- token_url
- api_key (secret)
- client_secret (secret) (required)

## ETL Streams

- employees:
  - primary key: id
  - fields: department(string), displayName(string), email(string), employeeNumber(string), employmentStatus(string), firstName(string), id(string), jobTitle(string), lastName(string)
- attendance:
  - primary key: id
  - fields: attendanceDate(string), employeeId(string), id(string), shiftEndTime(string), shiftStartTime(string), status(string), totalGrossHours(number)
- leave_types:
  - primary key: id
  - fields: id(string), identifier(string), isActive(boolean), leaveTypeUnit(string), name(string)
- leave_requests:
  - primary key: id
  - fields: dayCount(number), employeeId(string), fromDate(string), id(string), leaveTypeId(string), status(string), toDate(string)
- clients:
  - primary key: id
  - fields: code(string), id(string), isActive(boolean), name(string)
- projects:
  - primary key: id
  - fields: billingType(string), clientId(string), code(string), id(string), name(string), status(string)
- employee:
  - primary key: id
  - fields: accountStatus(integer), attendanceNumber(string), bandInfo(null), bloodGroup(integer), captureSchemeInfo(object), city(string), contingentType(object), countryCode(string), currentAddress(null), customFields(array), dateOfBirth(string), displayName(string), dottedLineManager(object), educationDetails(array), email(string), employeeNumber(string), employmentStatus(integer), exitDate(null), exitReason(null), exitStatus(integer), exitType(integer), expensePolicyInfo(object), experienceDetails(array), firstName(string), gender(integer), groups(array), homePhone(null), id(string), image(object), invitationStatus(integer), isPrivate(boolean), isProfileComplete(boolean), jobTitle(object), joiningDate(string), l2Manager(object), lastName(string), leavePlanInfo(object), maritalStatus(integer), marriageDate(null), middleName(null), mobilePhone(string), overtimePolicyInfo(null), payGradeInfo(null), permanentAddress(null), personalEmail(null), probationEndDate(string), professionalSummary(null), relations(array), reportsTo(object), resignationSubmittedDate(null), secondaryJobTitle(null), shiftPolicyInfo(object), timeType(integer), trackingPolicyInfo(null), weeklyOffPolicyInfo(object), workPhone(null), workerType(integer)
- employee_update_fields:
  - fields: jobFields(array), profileFields(array)
- groups:
  - primary key: id
  - fields: code(null), description(string), groupTypeId(string), id(string), name(string)
- group_types:
  - primary key: id
  - fields: id(string), isSystemDefined(boolean), name(string), systemGroupType(integer)
- departments:
  - primary key: id
  - fields: departmentHeads(array), description(string), id(string), isArchived(boolean), name(string), parentId(null)
- locations:
  - primary key: id
  - fields: address(object), description(string), id(string), name(string)
- job_titles:
  - primary key: id
  - fields: description(string), id(string), name(string)
- currencies:
  - primary key: id
  - fields: code(string), id(string), name(string)
- notice_periods:
  - primary key: id
  - fields: id(string), name(string)
- exit_reasons:
  - fields: exitReason(array), terminationReason(array)
- document_types:
  - primary key: id
  - fields: documentFields(array), id(string), name(string)
- employee_documents:
  - primary key: id
  - fields: attachments(array), attributes(array), id(string), name(string)
- employee_document_attachment_download_urls:
  - fields: fileURL(string)
- leave_balances:
  - fields: employeeIdentifier(string), employeeName(string), employeeNumber(string), leaveBalance(array)
- leave_plans:
  - primary key: id
  - fields: id(string), name(string)
- capture_schemes:
  - primary key: id
  - fields: id(string), name(string)
- shift_policies:
  - primary key: id
  - fields: id(string), name(string)
- holiday_calendars:
  - primary key: id
  - fields: id(string), name(string)
- tracking_policies:
  - primary key: id
  - fields: id(string), name(string)
- weekly_off_policies:
  - primary key: id
  - fields: id(string), name(string)
- salary_components:
  - primary key: id
  - fields: accountingCode(null), id(string), identifier(string), title(string)
- pay_groups:
  - primary key: identifier
  - fields: description(string), identifier(string), legalEntityId(string), legalEntityName(string), name(string)
- pay_cycles:
- pay_register:
- pay_batches:
- batch_payments:
- pay_grades:
  - primary key: id
  - fields: id(string), name(string)
- pay_bands:
- employee_salaries:
  - primary key: id
  - fields: ctc(integer), deductions(array), earnings(array), effectiveFrom(string), employee(object), gross(integer), id(string), remunerationType(integer)
- employee_fnf_details:
  - primary key: id
  - fields: comments(string), contributions(array), deductions(array), earnings(array), employeeNumber(string), esiNumber(null), exitRequestStatus(integer), id(string), lastWorkingDay(string), leaveEncashmentDays(integer), lossOfPayDays(integer), netAmount(integer), netRecovery(null), noOfPayDays(integer), okToRehire(boolean), panNumber(null), payAction(integer), pfNumber(null), resignationNote(null), settlementDate(string), terminationNoticeDate(string), terminationReason(string), terminationType(integer), uanNumber(null)
- client:
- billing_roles:
  - primary key: id
  - fields: billingRate(object), id(string), name(string)
- project_phases:
- project:
- project_allocations:
  - primary key: id
  - fields: employee(object), endDate(string), id(string), startDate(string)
- project_time_entries:
- project_tasks:
- project_task_time_entries:
- project_task_assignees:
  - primary key: id
  - fields: assignedTo(array), description(string), endDate(string), estimatedHours(number), id(string), name(string), projectId(string), startDate(string), taskBillingType(integer), taskType(integer)
- timesheet_entries:
- pms_timeframes:
- goals:
  - primary key: id
  - fields: childGoals(array), currentValue(integer), departmentId(null), description(string), employeeId(string), employeeNumber(string), endDate(string), id(string), initialValue(integer), isPrivate(boolean), metricType(integer), name(string), parentGoal(null), progress(integer), startDate(string), status(integer), tags(array), targetValue(integer), timeFrameId(string), type(integer)
- badges:
  - primary key: id
  - fields: description(string), id(string), name(string), status(integer)
- praise:
- review_groups:
  - primary key: id
  - fields: description(null), id(string), name(string)
- review_cycles:
  - primary key: id
  - fields: fromDate(string), id(string), isAdhoc(boolean), name(string), reviewGroup(object), status(integer), toDate(string)
- reviews:
  - primary key: id
  - fields: employee(object), id(string), ratings(array), reviewCycle(object), reviewGroup(object), status(integer), summary(null)
- hire_jobs:
  - primary key: id
  - fields: createdBy(string), createdOn(string), departmentName(string), description(string), experience(null), id(string), jobLocations(array), jobType(string), noOfOpenings(string), orgJobId(null), publishedBy(string), publishedOn(string), status(integer), targetHireDate(string), title(string), totalHiredPositions(string)
- job_application_fields:
  - primary key: id
  - fields: fieldName(string), fieldOptions(array), fieldType(integer), id(string), isSystemGenerated(boolean), required(boolean)
- candidates:
  - primary key: id
  - fields: additionalCandidateDetails(object), educationDetails(array), email(string), experienceDetails(array), firstName(string), gender(integer), id(string), jobApplicationDetails(object), lastName(string), middleName(null), mobilePhone(object), skills(array)
- candidate_interviews:
  - primary key: id
  - fields: candidateId(string), endTime(object), id(string), interviewDate(string), interviewType(string), jobId(string), panelMembers(string), scheduledBy(string), scheduledDate(string), stageId(string), startTime(object), timeZoneId(string)
- candidate_scorecards:
- preboarding_candidates:
  - primary key: id
  - fields: countryCode(string), department(null), email(string), expectedDateOfJoining(string), firstName(string), gender(integer), id(string), jobTitle(null), lastName(string), middleName(null), mobileNumber(string), stage(integer), status(integer), workLocation(string)
- expense_categories:
  - primary key: id
  - fields: categoryType(integer), code(string), description(string), id(string), name(string)
- expense_claims:
  - primary key: id
  - fields: approvalStatus(integer), claimNumber(string), employee(object), expenses(array), id(string), submittedOn(string), title(string)
- expense_policies:
  - primary key: id
  - fields: id(string), name(string)
- assets:
  - primary key: id
  - fields: assetCategoryId(string), assetConditionId(string), assetId(string), assetName(string), assetTypeId(string), assignedOn(string), assignedTo(object), id(string), status(integer)
- asset_types:
  - primary key: id
  - fields: id(string), name(string)
- asset_categories:
  - primary key: id
  - fields: id(string), name(string)
- asset_conditions:
  - primary key: id
  - fields: id(string), name(string)
- requisition_requests:
  - primary key: id
  - fields: additionalComments(string), additionalFields(null), budget(string), department(string), hired(integer), id(string), isArchived(boolean), isPriority(boolean), jobNumber(string), jobType(integer), locations(array), openPositions(integer), requestedBy(string), requestedOn(string), requisitionFor(string), requisitionTypes(array), status(integer), subDepartment(null), targetHiringDate(string), toBeReplaced(array)
- employee_skills:
  - primary key: id
  - fields: id(string), rating(integer), skillName(string)
- bgv_requests:
  - primary key: id
  - fields: bgvDecision(integer), candidateId(string), checks(array), email(string), firstName(string), gender(integer), id(string), lastName(string), middleName(null), mobileNumber(null), status(integer)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Reverse ETL Actions

- create_employee:
  - endpoint: POST /hris/employees
  - risk: Create Employee through the Keka API.
- update_employee_personal_details:
  - endpoint: PUT /hris/employees/{{ record.employee_id }}/personaldetails
  - required fields: employee_id
  - risk: Update Employee Personal Details through the Keka API.
- update_employee_job_details:
  - endpoint: PUT /hris/employees/{{ record.employee_id }}/jobdetails
  - required fields: employee_id
  - risk: Update Employee Job Details through the Keka API.
- create_exit_request:
  - endpoint: POST /hris/employees/{{ record.employee_id }}/exitrequest
  - required fields: employee_id
  - risk: Create Exit Request through the Keka API.
- update_exit_request:
  - endpoint: PUT /hris/employees/{{ record.employee_id }}/exitrequest
  - required fields: employee_id
  - risk: Update Exit Request through the Keka API.
- create_leave_request:
  - endpoint: POST /time/leaverequests
  - risk: Create Leave Request through the Keka API.
- update_payment_status:
  - endpoint: PUT /payroll/paygroups/{{ record.pay_group_id }}/paycycles/{{ record.pay_cycle_id }}/paybatches/{{ record.pay_batch_id }}/payments
  - required fields: pay_group_id, pay_cycle_id, pay_batch_id
  - risk: Update Payment Status through the Keka API.
- create_client:
  - endpoint: POST /psa/clients
  - risk: Create Client through the Keka API.
- update_client:
  - endpoint: PUT /psa/clients/{{ record.client_id }}
  - required fields: client_id
  - risk: Update Client through the Keka API.
- create_project_phase:
  - endpoint: POST /psa/projects/{{ record.project_id }}/phases
  - required fields: project_id
  - risk: Create Project Phase through the Keka API.
- create_project:
  - endpoint: POST /psa/projects
  - risk: Create Project through the Keka API.
- update_project_details:
  - endpoint: PUT /psa/projects/{{ record.project_id }}
  - required fields: project_id
  - risk: Update Project Details through the Keka API.
- add_project_allocation:
  - endpoint: POST /psa/projects/{{ record.project_id }}/allocations
  - required fields: project_id
  - risk: Add Project Allocation through the Keka API.
- create_project_task:
  - endpoint: POST /psa/projects/{{ record.project_id }}/tasks
  - required fields: project_id
  - risk: Create Project Task through the Keka API.
- update_project_task:
  - endpoint: PUT /psa/projects/{{ record.project_id }}/tasks/{{ record.task_id }}
  - required fields: project_id, task_id
  - risk: Update Project Task through the Keka API.
- update_goal_progress:
  - endpoint: PUT /pms/goals/{{ record.goal_id }}/progress
  - required fields: goal_id
  - risk: Update Goal Progress through the Keka API.
- create_praise:
  - endpoint: POST /pms/praise
  - risk: Create Praise through the Keka API.
- update_candidate:
  - endpoint: PUT /hire/jobs/{{ record.job_id }}/candidate/{{ record.candidate_id }}
  - required fields: job_id, candidate_id
  - risk: Update Candidate through the Keka API.
- add_candidate_notes:
  - endpoint: POST /hire/jobs/{{ record.job_id }}/candidate/{{ record.candidate_id }}/notes
  - required fields: job_id, candidate_id
  - risk: Add Candidate Notes through the Keka API.
- create_candidate:
  - endpoint: POST /v1/hire/jobs/{{ record.job_id }}/candidate
  - required fields: job_id
  - risk: Create Candidate through the Keka API.
- create_preboarding_candidate:
  - endpoint: POST /hire/preboarding/candidates
  - risk: Create Preboarding Candidate through the Keka API.
- update_preboarding_candidate:
  - endpoint: PUT /hire/preboarding/candidates/{{ record.preboarding_candidate_id }}
  - required fields: preboarding_candidate_id
  - risk: Update Preboarding Candidate through the Keka API.
- update_asset_assignment:
  - endpoint: PUT /assets/{{ record.asset_id }}/allocation
  - required fields: asset_id
  - risk: Update Asset Assignment through the Keka API.
- add_bgv_request_report:
  - endpoint: PUT /hris/bgv/{{ record.bgv_id }}/requests/{{ record.request_id }}
  - required fields: bgv_id, request_id
  - risk: Add Bgv Request Report through the Keka API.

## Security

- read risk: external Keka HRMS API read of employee, attendance, leave, payroll, PSA, hiring, expense, asset, requisition, skill, and BGV data
- write risk: live Keka API mutations can create or update employee, leave, payroll payment, client, project, performance, hiring, asset, skill, and BGV records
- approval: reverse ETL writes require plan, preview, and approval token before live Keka mutations execute
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect keka
```

### Inspect as structured JSON

```bash
pm connectors inspect keka --json
```

## Agent Rules

- Run pm connectors inspect keka before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
