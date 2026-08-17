---
name: pm-keka
description: Keka connector knowledge and safe action guide.
---

# pm-keka

## Purpose

Reads and writes the documented Keka HRMS REST API surface for Core HR, documents, leave, attendance, payroll, PSA, PMS, hire, expense, assets, requisitions, skills, and BGV resources.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- attachment_id
- base_url
- bgv_id
- candidate_id
- client_id
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
- client_secret (secret)

## ETL Streams

- employees:
  - primary key: id
  - fields: department(), displayName(), email(), employeeNumber(), employmentStatus(), firstName(), id(), jobTitle(), lastName()
- attendance:
  - primary key: id
  - fields: attendanceDate(), employeeId(), id(), shiftEndTime(), shiftStartTime(), status(), totalGrossHours()
- leave_types:
  - primary key: id
  - fields: id(), identifier(), isActive(), leaveTypeUnit(), name()
- leave_requests:
  - primary key: id
  - fields: dayCount(), employeeId(), fromDate(), id(), leaveTypeId(), status(), toDate()
- clients:
  - primary key: id
  - fields: code(), id(), isActive(), name()
- projects:
  - primary key: id
  - fields: billingType(), clientId(), code(), id(), name(), status()
- employee:
  - primary key: id
  - fields: accountStatus(), attendanceNumber(), bandInfo(), bloodGroup(), captureSchemeInfo(), city(), contingentType(), countryCode(), currentAddress(), customFields(), dateOfBirth(), displayName(), dottedLineManager(), educationDetails(), email(), employeeNumber(), employmentStatus(), exitDate(), exitReason(), exitStatus(), exitType(), expensePolicyInfo(), experienceDetails(), firstName(), gender(), groups(), homePhone(), id(), image(), invitationStatus(), isPrivate(), isProfileComplete(), jobTitle(), joiningDate(), l2Manager(), lastName(), leavePlanInfo(), maritalStatus(), marriageDate(), middleName(), mobilePhone(), overtimePolicyInfo(), payGradeInfo(), permanentAddress(), personalEmail(), probationEndDate(), professionalSummary(), relations(), reportsTo(), resignationSubmittedDate(), secondaryJobTitle(), shiftPolicyInfo(), timeType(), trackingPolicyInfo(), weeklyOffPolicyInfo(), workPhone(), workerType()
- employee_update_fields:
  - fields: jobFields(), profileFields()
- groups:
  - primary key: id
  - fields: code(), description(), groupTypeId(), id(), name()
- group_types:
  - primary key: id
  - fields: id(), isSystemDefined(), name(), systemGroupType()
- departments:
  - primary key: id
  - fields: departmentHeads(), description(), id(), isArchived(), name(), parentId()
- locations:
  - primary key: id
  - fields: address(), description(), id(), name()
- job_titles:
  - primary key: id
  - fields: description(), id(), name()
- currencies:
  - primary key: id
  - fields: code(), id(), name()
- notice_periods:
  - primary key: id
  - fields: id(), name()
- exit_reasons:
  - fields: exitReason(), terminationReason()
- document_types:
  - primary key: id
  - fields: documentFields(), id(), name()
- employee_documents:
  - primary key: id
  - fields: attachments(), attributes(), id(), name()
- employee_document_attachment_download_urls:
  - fields: fileURL()
- leave_balances:
  - fields: employeeIdentifier(), employeeName(), employeeNumber(), leaveBalance()
- leave_plans:
  - primary key: id
  - fields: id(), name()
- capture_schemes:
  - primary key: id
  - fields: id(), name()
- shift_policies:
  - primary key: id
  - fields: id(), name()
- holiday_calendars:
  - primary key: id
  - fields: id(), name()
- tracking_policies:
  - primary key: id
  - fields: id(), name()
- weekly_off_policies:
  - primary key: id
  - fields: id(), name()
- salary_components:
  - primary key: id
  - fields: accountingCode(), id(), identifier(), title()
- pay_groups:
  - primary key: identifier
  - fields: description(), identifier(), legalEntityId(), legalEntityName(), name()
- pay_cycles:
- pay_register:
- pay_batches:
- batch_payments:
- pay_grades:
  - primary key: id
  - fields: id(), name()
- pay_bands:
- employee_salaries:
  - primary key: id
  - fields: ctc(), deductions(), earnings(), effectiveFrom(), employee(), gross(), id(), remunerationType()
- employee_fnf_details:
  - primary key: id
  - fields: comments(), contributions(), deductions(), earnings(), employeeNumber(), esiNumber(), exitRequestStatus(), id(), lastWorkingDay(), leaveEncashmentDays(), lossOfPayDays(), netAmount(), netRecovery(), noOfPayDays(), okToRehire(), panNumber(), payAction(), pfNumber(), resignationNote(), settlementDate(), terminationNoticeDate(), terminationReason(), terminationType(), uanNumber()
- client:
- billing_roles:
  - primary key: id
  - fields: billingRate(), id(), name()
- project_phases:
- project:
- project_allocations:
  - primary key: id
  - fields: employee(), endDate(), id(), startDate()
- project_time_entries:
- project_tasks:
- project_task_time_entries:
- project_task_assignees:
  - primary key: id
  - fields: assignedTo(), description(), endDate(), estimatedHours(), id(), name(), projectId(), startDate(), taskBillingType(), taskType()
- timesheet_entries:
- pms_timeframes:
- goals:
  - primary key: id
  - fields: childGoals(), currentValue(), departmentId(), description(), employeeId(), employeeNumber(), endDate(), id(), initialValue(), isPrivate(), metricType(), name(), parentGoal(), progress(), startDate(), status(), tags(), targetValue(), timeFrameId(), type()
- badges:
  - primary key: id
  - fields: description(), id(), name(), status()
- praise:
- review_groups:
  - primary key: id
  - fields: description(), id(), name()
- review_cycles:
  - primary key: id
  - fields: fromDate(), id(), isAdhoc(), name(), reviewGroup(), status(), toDate()
- reviews:
  - primary key: id
  - fields: employee(), id(), ratings(), reviewCycle(), reviewGroup(), status(), summary()
- hire_jobs:
  - primary key: id
  - fields: createdBy(), createdOn(), departmentName(), description(), experience(), id(), jobLocations(), jobType(), noOfOpenings(), orgJobId(), publishedBy(), publishedOn(), status(), targetHireDate(), title(), totalHiredPositions()
- job_application_fields:
  - primary key: id
  - fields: fieldName(), fieldOptions(), fieldType(), id(), isSystemGenerated(), required()
- candidates:
  - primary key: id
  - fields: additionalCandidateDetails(), educationDetails(), email(), experienceDetails(), firstName(), gender(), id(), jobApplicationDetails(), lastName(), middleName(), mobilePhone(), skills()
- candidate_interviews:
  - primary key: id
  - fields: candidateId(), endTime(), id(), interviewDate(), interviewType(), jobId(), panelMembers(), scheduledBy(), scheduledDate(), stageId(), startTime(), timeZoneId()
- candidate_scorecards:
- preboarding_candidates:
  - primary key: id
  - fields: countryCode(), department(), email(), expectedDateOfJoining(), firstName(), gender(), id(), jobTitle(), lastName(), middleName(), mobileNumber(), stage(), status(), workLocation()
- expense_categories:
  - primary key: id
  - fields: categoryType(), code(), description(), id(), name()
- expense_claims:
  - primary key: id
  - fields: approvalStatus(), claimNumber(), employee(), expenses(), id(), submittedOn(), title()
- expense_policies:
  - primary key: id
  - fields: id(), name()
- assets:
  - primary key: id
  - fields: assetCategoryId(), assetConditionId(), assetId(), assetName(), assetTypeId(), assignedOn(), assignedTo(), id(), status()
- asset_types:
  - primary key: id
  - fields: id(), name()
- asset_categories:
  - primary key: id
  - fields: id(), name()
- asset_conditions:
  - primary key: id
  - fields: id(), name()
- requisition_requests:
  - primary key: id
  - fields: additionalComments(), additionalFields(), budget(), department(), hired(), id(), isArchived(), isPriority(), jobNumber(), jobType(), locations(), openPositions(), requestedBy(), requestedOn(), requisitionFor(), requisitionTypes(), status(), subDepartment(), targetHiringDate(), toBeReplaced()
- employee_skills:
  - primary key: id
  - fields: id(), rating(), skillName()
- bgv_requests:
  - primary key: id
  - fields: bgvDecision(), candidateId(), checks(), email(), firstName(), gender(), id(), lastName(), middleName(), mobileNumber(), status()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

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
