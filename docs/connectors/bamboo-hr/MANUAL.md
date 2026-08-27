# pm connectors inspect bamboo-hr

```text
NAME
  pm connectors inspect bamboo-hr - BambooHR connector manual

SYNOPSIS
  pm connectors inspect bamboo-hr
  pm connectors inspect bamboo-hr --json
  pm credentials add <name> --connector bamboo-hr [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes BambooHR employee, metadata, reporting, time off, applicant tracking, benefits, goals, training, time tracking, scheduling, and webhook resources that are available through the documented Basic-auth API surface.

ICON
  id: bamboohr
  asset: icons/bamboohr.svg
  source: upstream_registry
  review_status: upstream_seeded

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  application_id
  break_id
  break_policy_breaks_id
  break_policy_employees_id
  break_policy_id
  changed_employee_ids_since
  changed_employee_table_data_since
  company_benefit_id
  company_report_id
  country_id
  dataset_name
  employee_break_availabilities_id
  employee_break_policies_id
  employee_dependent_id
  employee_id
  employee_table_data_id
  goal_id
  goal_share_options_search
  member_benefits_calendar_year
  project_id
  report_id
  scheduling_get_schedule_id
  scheduling_get_shift_id
  shift_differential_id
  subdomain (required)
  table
  task_id
  time_off_requests_end
  time_off_requests_start
  time_tracking_record_id
  timesheet_entries_end
  timesheet_entries_start
  webhook_id
  api_key (secret) (required)

ETL STREAMS
  employees:
    primary key: id
    fields: department(string), display_name(string), division(string), first_name(string), id(string), job_title(string), last_name(string), location(string), mobile_phone(string), photo_url(string), preferred_name(string), supervisor(string), work_email(string), work_phone(string)
  meta_fields:
    primary key: id
    fields: alias(string), deprecated(boolean), id(string), name(string), type(string)
  meta_lists:
    primary key: field_id
    fields: alias(string), field_id(string), manageable(boolean), multiple(boolean), name(string), options(array)
  time_off_types:
    primary key: id
    fields: color(string), icon(string), id(string), name(string), units(string)
  applications:
    primary key: id
    fields: applicant(object), appliedDate(string), id(integer), job(object), rating(integer), status(object)
  application_details:
    primary key: id
    fields: answer(object), archivedDate(string), editedDate(string), editedEndDate(string), hasRevisions(boolean), id(string), isArchived(boolean), question(object)
  hiring_leads:
    primary key: employeeId
    fields: employeeId(integer), preferredFullName(string)
  job_summaries:
    primary key: id
    fields: activeApplicantsCount(integer), department(object), hiringLead(object), id(integer), location(object), newApplicantsCount(integer), postedDate(string), postingUrl(string), status(object), title(object), totalApplicantsCount(integer)
  company_locations:
    primary key: id
    fields: addressLine1(string), addressLine2(string), city(string), country(object), description(string), id(integer), name(string), phone(string), state(object), zipcode(string)
  statuses:
    primary key: id
    fields: code(string), description(string), enabled(boolean), id(string), manageable(boolean), name(string), translatedName(string)
  company_benefits:
    primary key: id
    fields: allowsCatchUp(boolean), allowsSuperCatchUp(boolean), benefitVendorId(string), companyDeductionId(string), deductionTypeId(string), endDate(string), id(string), name(string), startDate(string), type(string)
  company_benefit_types:
    primary key: id
    fields: canBeAcaPlan(boolean), canCoExistEnrollment(boolean), id(string), isReimbursementPlan(boolean), name(string), slug(string)
  company_benefit:
    primary key: benefitVendorId
    fields: benefitType(string), benefitVendorId(string), deductionTypeId(integer), description(string), endDate(string), meetAcaMin(string), minEssentialCoverage(string), name(string), planUrl(string), reimbursementAmount(number), reimbursementFrequency(string), safeHarbor(string), ssoLoginUrl(string), ssoLoginUrlLinkText(string), startDate(string)
  employee_benefits:
    primary key: employeeId
    fields: employeeBenefit(array), employeeId(integer), payFrequency(string)
  member_benefit_events:
    primary key: memberId
    fields: coverages(array), memberId(string)
  benefit_coverages:
    primary key: id
    fields: benefitPlanId(string), description(string), id(string), shortName(string), sortOrder(string)
  member_benefits:
    primary key: memberId
    fields: memberId(string), plans(array), subscriberId(string)
  benefit_deduction_types:
    primary key: id
    fields: additionalDescription(string), allowableBenefitTypes(array), canBeCollectedByTrax(boolean), deductionNote(string), deductionNoteLink(string), deductionNoteLinkText(string), deductionTypeName(string), defaultDeductionCode(string), hideAnnualMax(boolean), id(integer), managedDeductionType(string), nonBenefitDeductionType(boolean), subTypeText(string), subTypes(array)
  company_profile_integrations:
    primary key: id
    fields: id(string), integrations(array)
  company_eins:
    primary key: id
    fields: eins(array), id(string)
  company_information:
    primary key: id
    fields: address(object), displayName(string), id(string), legalName(string), phone(string)
  reports:
    primary key: id
    fields: id(integer), name(string)
  report_by_id:
    primary key: id
    fields: id(string)
  datasets_v1:
    primary key: id
    fields: id(string), label(string), name(string)
  fields_from_dataset_v1:
    primary key: id
    fields: entityName(string), id(string), label(string), name(string), parentName(string), parentType(string)
  employee_dependents:
    primary key: id
    fields: addressLine1(string), addressLine2(string), city(string), country(string), dateOfBirth(string), employeeId(string), firstName(string), gender(string), homePhone(string), id(string), isStudent(string), isUsCitizen(string), lastName(string), maskedSIN(string), maskedSSN(string), middleName(string), relationship(string), state(string), zipCode(string)
  employee_dependent:
    primary key: id
    fields: addressLine1(string), addressLine2(string), city(string), country(string), dateOfBirth(string), employeeId(string), firstName(string), gender(string), homePhone(string), id(string), isStudent(string), isUsCitizen(string), lastName(string), maskedSIN(string), maskedSSN(string), middleName(string), relationship(string), state(string), zipCode(string)
  employee_roster:
    primary key: employeeId
    fields: _restrictedFields(array), addressLine1(string), addressLine2(string), age(string), allergies(string), bestEmail(string), birthDate(string), birthplace(string), citizenship(string), citizenshipId(string), city(string), compensationChangeReason(string), compensationChangeReasonId(string), compensationComment(string), compensationEffectiveDate(string), compensationEndDate(string), contractEndDate(string), country(string), countryId(string), departmentId(string), departmentName(string), dietaryRestrictions(string), displayName(string), divisionId(string), divisionName(string), eeoJobCategory(string), eeoJobCategoryId(string), ein(string), eligibleForRehire(string), eligibleForRehireId(string), employeeId(string), employeeName(string), employeeNumber(string), employmentStatusComment(string), employmentStatusEffectiveDate(string), employmentStatusId(string), employmentStatusName(string), employmentType(string), employmentTypeId(string), ethnicity(string), ethnicityId(string), facebookUrl(string), finalDoseAdministrationDate(string), finalPayDate(string), firstName(string), firstNameLastName(string), firstNameMiddleInitial(string), flsaCode(string), flsaCodeId(string), gender(string), genderIdentity(string), genderIdentityId(array), hireDate(string), homeEmail(string), homePhone(string), hoursPerPayCycle(string), instagramUrl(string), isManager(boolean), jacketSize(string), jacketSizeId(string), jobInformationEffectiveDate(string), jobTitleId(string), jobTitleName(string), lastName(string), linkedinUrl(string), locationId(string), locationName(string), maritalStatus(string), middleInitial(string), middleName(string), mobilePhone(string), nationalId(string), nationalInsuranceCategory(string), nationalInsuranceCategoryId(string), nationality(string), nationalityId(string), nickName(string), nin(string), noticePeriod(string), noticePeriodId(string), originalHireDate(string), overtime(string), overtimeRate(object), paidPer(string), payRate(object), paySchedule(string), payScheduleId(string), payType(string), photoUrl(string), pinterestUrl(string), preferredName(string), preferredNameLastName(string), probationEndDate(string), pronouns(string), pronounsId(string), proofOfVaccination(boolean), reportsToId(string), reportsToName(string), secondaryLanguage(string), shirtSize(string), shirtSizeId(string), sin(string), skypeUsername(string), ssn(string), state(string), stateId(string), status(string), tShirtSize(string), tShirtSizeId(string), taxTypeId(string), teams(array), tenure(string), terminationDate(string), terminationReason(string), terminationReasonId(string), terminationRegrettable(string), terminationRegrettableId(string), terminationType(string), terminationTypeId(string), twitterUrl(string), userId(string), vaccinationStatus(string), vaccinationStatusId(string), vaccineReceived(string), vaccineReceivedId(string), veteranStatus(string), veteranStatusId(array), workEmail(string), workPhone(string), workPhoneExtension(string), zipcode(string)
  changed_employee_ids:
    primary key: id
    fields: employees(object), id(string), latest(string)
  changed_employee_table_data:
    primary key: id
    fields: employees(object), id(string), table(string)
  time_off_balance:
    primary key: id
    fields: balance(string), end(string), id(string), name(string), policyType(string), timeOffType(string), units(string), usedYearToDate(string)
  employee_time_off_policies:
    primary key: timeOffPolicyId
    fields: accrualStartDate(string), timeOffPolicyId(string), timeOffTypeId(integer)
  employee_table_data:
    primary key: id
    fields: employeeId(string), id(string)
  all_currency_types:
    primary key: id
    fields: code(string), id(integer), name(string), symbol(string), symbolPosition(integer)
  states_by_country_id:
    primary key: id
    fields: id(integer), iso(string), label(string), name(string)
  tabular_fields:
    primary key: id
    fields: alias(string), fields(array), id(string)
  time_off_policies:
    primary key: id
    fields: effectiveDate(string), id(integer), name(string), timeOffTypeId(integer), type(string)
  users:
    primary key: id
    fields: id(string)
  goals:
    primary key: id
    fields: actions(object), alignsWithOptionId(string), completionDate(string), description(string), dueDate(string), id(string), lastChangedDateTime(string), milestones(array), percentComplete(integer), sharedWithEmployeeIds(array), status(string), title(string)
  goals_aggregate_v1:
    primary key: id
    fields: actions(object), alignsWithOptionId(string), completionDate(string), description(string), dueDate(string), id(string), lastChangedDateTime(string), milestones(array), percentComplete(integer), sharedWithEmployeeIds(array), status(string), title(string)
  alignable_goal_options:
    primary key: id
    fields: id(string), title(string)
  goal_creation_permission:
    primary key: id
    fields: canCreateGoals(boolean), id(string)
  goals_filters_v1:
    primary key: id
    fields: count(integer), id(string), name(string)
  goal_share_options:
    primary key: employeeId
    fields: displayFirstName(string), employeeId(integer), lastName(string), photoUrl(string), userId(integer)
  goal_aggregate:
    primary key: id
    fields: authorUserId(integer), canDelete(boolean), canEdit(boolean), createdAt(string), id(string), text(string)
  goal_comments:
    primary key: id
    fields: authorUserId(integer), canDelete(boolean), canEdit(boolean), createdAt(string), id(string), text(string)
  company_report:
    primary key: id
    fields: id(string)
  scheduling_list_schedules:
    primary key: id
    fields: createdAt(string), deletedAt(string), earlyClockInThreshold(integer), employeeIds(array), id(string), locationId(integer), managerUserIds(array), name(string), startOfWeek(string), timezone(string), updatedAt(string)
  scheduling_get_schedule:
    primary key: id
    fields: createdAt(string), deletedAt(string), earlyClockInThreshold(integer), employeeIds(array), id(string), locationId(integer), managerUserIds(array), name(string), startOfWeek(string), timezone(string), updatedAt(string)
  scheduling_list_shift_assessments:
    primary key: id
    fields: createdAt(string), date(string), employeeId(integer), id(string), result(), shiftId(string), updatedAt(string), violations(array)
  scheduling_list_shifts:
    primary key: id
    fields: capacity(integer), color(string), createdAt(string), deletedAt(string), employeeIds(array), end(string), id(string), name(string), recurrenceDtend(string), recurrenceDtstart(string), recurrenceId(string), recurrenceRule(string), recurrenceUntil(string), scheduleId(string), start(string), status(), timezone(string), unpublishedChanges(object), updatedAt(string)
  scheduling_get_shift:
    primary key: id
    fields: capacity(integer), color(string), createdAt(string), deletedAt(string), employeeIds(array), end(string), id(string), name(string), recurrenceDtend(string), recurrenceDtstart(string), recurrenceId(string), recurrenceRule(string), recurrenceUntil(string), scheduleId(string), start(string), status(), timezone(string), unpublishedChanges(object), updatedAt(string)
  scheduling_list_timezones:
    primary key: id
    fields: id(string), name(string), offset(string)
  break_assessments:
    primary key: id
    fields: _links(object), data(array), id(string), meta(object)
  break_policies:
    primary key: id
    fields: _links(object), data(array), id(string), meta(object)
  break_policy:
    primary key: id
    fields: allEmployeesAssigned(boolean), createdAt(string), deletedAt(string), description(string), id(string), name(string), updatedAt(string)
  break_policy_breaks:
    primary key: id
    fields: _links(object), data(array), id(string), meta(object)
  break_policy_employees:
    primary key: id
    fields: _links(object), data(array), id(string), meta(object)
  break:
    primary key: id
    fields: availabilityEndTime(string), availabilityMaxHoursWorked(number), availabilityMinHoursWorked(number), availabilityStartTime(string), availabilityType(), createdAt(string), deletedAt(string), duration(integer), id(string), name(string), paid(boolean), policyId(string), updatedAt(string)
  employee_break_availabilities:
    primary key: id
    fields: availabilityType(), available(boolean), availableAfterMinutesWorked(integer), availableAt(string), availableIn(integer), calculatedAt(string), duration(integer), effectiveAt(string), id(string), name(string), paid(boolean), policyId(string), recordedDuration(integer), timezone(string), unavailableAt(string), unavailableIn(integer)
  employee_break_policies:
    primary key: id
    fields: _links(object), data(array), id(string), meta(object)
  projects:
    primary key: id
    fields: _links(object), data(array), id(string), meta(object)
  project:
    primary key: id
    fields: allEmployeesAssigned(boolean), archived(boolean), billable(boolean), createdAt(string), deletedAt(string), employeeIds(array), hasTasks(boolean), id(integer), includeInPayroll(boolean), name(string), updatedAt(string)
  project_tasks:
    primary key: id
    fields: billable(boolean), createdAt(string), deletedAt(string), id(integer), name(string), projectId(integer), updatedAt(string)
  shift_differentials:
    primary key: id
    fields: _links(object), data(array), id(string), meta(object)
  shift_differential:
    primary key: id
    fields: end(string), endDay(string), id(integer), start(string), startDay(string)
  task:
    primary key: id
    fields: billable(boolean), createdAt(string), deletedAt(string), id(integer), name(string), projectId(integer), updatedAt(string)
  time_off_requests:
    primary key: id
    fields: actions(object), amount(object), created(string), dates(object), employeeId(string), end(string), id(string), name(string), notes(object), start(string), status(object), type(object)
  whos_out:
    primary key: id
    fields: employeeId(integer), end(string), id(integer), name(string), start(string), type(string)
  timesheet_entries:
    primary key: id
    fields: approved(boolean), approvedAt(string), createdAt(string), date(string), employeeId(integer), end(string), hours(number), id(integer), note(string), projectInfo(object), start(string), timezone(string), type(string), updatedAt(string)
  time_tracking_record:
    primary key: employeeId
    fields: adjustedHours(string), dateAdjusted(string), dateHoursWorked(string), departmentId(string), divisionId(string), employeeId(string), holidayId(string), hoursWorked(string), jobCode(string), jobData(string), jobTitleId(string), payCode(string), payRate(string), project(object), projectId(string), rateType(string), shiftDifferential(object), shiftDifferentialId(string), taskId(string), timeTrackingId(string), type(string)
  training_categories:
    primary key: id
    fields: id(string)
  training_types:
    primary key: id
    fields: id(string)
  webhooks:
    primary key: id
    fields: created(string), id(string), lastSent(string), name(string), url(string)
  monitor_fields:
    primary key: id
    fields: alias(string), id(string), name(string)
  post_fields:
    primary key: id
    fields: alias(string), id(integer), name(string), pageId(integer), tableId(integer), type(string)
  webhook:
    primary key: id
    fields: duplicatePostString(array), error(string), id(string), monitorFields(array), postFields(array), unknownFields(array)
  employee_time_off_policies_v1_1:
    primary key: timeOffPolicyId
    fields: accrualStartDate(string), timeOffPolicyId(string), timeOffTypeId(integer)
  goals_aggregate_v1_1:
    primary key: id
    fields: actions(object), alignsWithOptionId(string), completionDate(string), description(string), dueDate(string), id(string), lastChangedDateTime(string), milestones(array), percentComplete(integer), sharedWithEmployeeIds(array), status(string), title(string)
  goals_filters_v1_1:
    primary key: id
    fields: actions(object), count(integer), id(string), name(string)
  datasets_v1_2:
    primary key: id
    fields: id(string), label(string), name(string)
  fields_from_dataset_v1_2:
    primary key: id
    fields: entityName(string), id(string), label(string), name(string), parentName(string), parentType(string)
  goals_aggregate_v1_2:
    primary key: id
    fields: actions(object), alignsWithOptionId(string), completionDate(string), description(string), dueDate(string), id(string), lastChangedDateTime(string), milestones(array), percentComplete(integer), sharedWithEmployeeIds(array), status(string), title(string)
  goals_filters_v1_2:
    primary key: id
    fields: actions(object), count(integer), id(string), name(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

REVERSE ETL ACTIONS
  create_application_comment:
    endpoint: POST /api/v1/applicant_tracking/applications/{{ record.application_id }}/comments
    required fields: application_id, comment
    risk: Create Job Application Comment through the BambooHR API.
  update_applicant_status:
    endpoint: POST /api/v1/applicant_tracking/applications/{{ record.application_id }}/status
    required fields: application_id, status
    risk: Update Applicant Status through the BambooHR API.
  add_new_company_benefit:
    endpoint: POST /api/v1/benefit/company_benefit
    risk: Add a new company benefit through the BambooHR API.
  delete_company_benefit:
    endpoint: DELETE /api/v1/benefit/company_benefit/{{ record.id }}
    required fields: id
    risk: Deletes BambooHR data: Delete a company benefit.
  update_company_benefit:
    endpoint: PUT /api/v1/benefit/company_benefit/{{ record.id }}
    required fields: id
    risk: Update a company benefit through the BambooHR API.
  create_employee_benefit:
    endpoint: POST /api/v1/benefit/employee_benefit
    risk: Add an employee benefit through the BambooHR API.
  add_benefit_group_employee:
    endpoint: POST /api/v1/benefitgroupemployees
    risk: Add a benefit group employee through the BambooHR API.
  clear_employee_deposit:
    endpoint: DELETE /api/v1/employee_direct_deposit_accounts/{{ record.id }}
    required fields: id
    risk: Deletes BambooHR data: Clear an employee's direct deposit information.
  add_employee_deposit:
    endpoint: POST /api/v1/employee_direct_deposit_accounts/{{ record.id }}
    required fields: id
    risk: Add an employee's direct deposit information through the BambooHR API.
  add_employee_paystub:
    endpoint: POST /api/v1/employee_pay_stub
    risk: Add an employee's paystub through the BambooHR API.
  clear_employee_paystub:
    endpoint: DELETE /api/v1/employee_pay_stub/{{ record.id }}
    required fields: id
    risk: Deletes BambooHR data: Delete an employee's paystub.
  add_employee_unpaid_paystubs:
    endpoint: POST /api/v1/employee_unpaid_pay_stubs
    risk: Add an employee's unpaid paystubs through the BambooHR API.
  clear_employee_unpaid_paystubs:
    endpoint: DELETE /api/v1/employee_unpaid_pay_stubs/{{ record.id }}
    required fields: id
    risk: Deletes BambooHR data: Clear an employee's unpaid paystubs.
  clear_employee_withholding:
    endpoint: DELETE /api/v1/employee_withholding/{{ record.id }}
    required fields: id
    risk: Deletes BambooHR data: Clear an employee's default withholdings.
  add_employee_withholding:
    endpoint: POST /api/v1/employee_withholding/{{ record.id }}
    required fields: id
    risk: Add an employee's default withholdings through the BambooHR API.
  create_employee_dependent:
    endpoint: POST /api/v1/employeedependents
    required fields: employeeId
    risk: Create Employee Dependent through the BambooHR API.
  update_employee_dependent:
    endpoint: PUT /api/v1/employeedependents/{{ record.id }}
    required fields: id, employeeId
    risk: Update Employee Dependent through the BambooHR API.
  adjust_time_off_balance:
    endpoint: PUT /api/v1/employees/{{ record.employee_id }}/time_off/balance_adjustment
    required fields: employee_id, amount, date, timeOffTypeId
    risk: Adjust Time Off Balance through the BambooHR API.
  create_time_off_history:
    endpoint: PUT /api/v1/employees/{{ record.employee_id }}/time_off/history
    required fields: employee_id, date
    risk: Create Time Off History Item through the BambooHR API.
  assign_time_off_policies:
    endpoint: PUT /api/v1/employees/{{ record.employee_id }}/time_off/policies
    required fields: employee_id
    risk: Assign Time Off Policies through the BambooHR API.
  create_time_off_request:
    endpoint: PUT /api/v1/employees/{{ record.employee_id }}/time_off/request
    required fields: employee_id, status, start, end, timeOffTypeId
    risk: Create Time Off Request through the BambooHR API.
  create_table_row:
    endpoint: POST /api/v1/employees/{{ record.id }}/tables/{{ record.table }}
    required fields: id, table
    risk: Create Table Row through the BambooHR API.
  delete_employee_table_row:
    endpoint: DELETE /api/v1/employees/{{ record.id }}/tables/{{ record.table }}/{{ record.row_id }}
    required fields: id, table, row_id
    risk: Deletes BambooHR data: Delete Employee Table Row.
  update_table_row:
    endpoint: POST /api/v1/employees/{{ record.id }}/tables/{{ record.table }}/{{ record.row_id }}
    required fields: id, table, row_id
    risk: Update Table Row through the BambooHR API.
  update_list_field_values:
    endpoint: PUT /api/v1/meta/lists/{{ record.list_field_id }}
    required fields: list_field_id
    risk: Update List Field Values through the BambooHR API.
  create_goal:
    endpoint: POST /api/v1/performance/employees/{{ record.employee_id }}/goals
    required fields: employee_id, title, dueDate, sharedWithEmployeeIds
    risk: Create Goal through the BambooHR API.
  delete_goal:
    endpoint: DELETE /api/v1/performance/employees/{{ record.employee_id }}/goals/{{ record.goal_id }}
    required fields: employee_id, goal_id
    risk: Deletes BambooHR data: Delete Goal.
  update_goal_v1:
    endpoint: PUT /api/v1/performance/employees/{{ record.employee_id }}/goals/{{ record.goal_id }}
    required fields: employee_id, goal_id, dueDate, sharedWithEmployeeIds, title
    risk: Update Goal (v1) through the BambooHR API.
  close_goal:
    endpoint: POST /api/v1/performance/employees/{{ record.employee_id }}/goals/{{ record.goal_id }}/close
    required fields: employee_id, goal_id
    risk: Close Goal through the BambooHR API.
  create_goal_comment:
    endpoint: POST /api/v1/performance/employees/{{ record.employee_id }}/goals/{{ record.goal_id }}/comments
    required fields: employee_id, goal_id, text
    risk: Create Goal Comment through the BambooHR API.
  delete_goal_comment:
    endpoint: DELETE /api/v1/performance/employees/{{ record.employee_id }}/goals/{{ record.goal_id }}/comments/{{ record.comment_id }}
    required fields: employee_id, goal_id, comment_id
    risk: Deletes BambooHR data: Delete Goal Comment.
  update_goal_comment:
    endpoint: PUT /api/v1/performance/employees/{{ record.employee_id }}/goals/{{ record.goal_id }}/comments/{{ record.comment_id }}
    required fields: employee_id, goal_id, comment_id, text
    risk: Update Goal Comment through the BambooHR API.
  update_goal_milestone_progress:
    endpoint: PUT /api/v1/performance/employees/{{ record.employee_id }}/goals/{{ record.goal_id }}/milestones/{{ record.milestone_id }}/progress
    required fields: employee_id, goal_id, milestone_id, complete
    risk: Update Milestone Progress through the BambooHR API.
  update_goal_progress:
    endpoint: PUT /api/v1/performance/employees/{{ record.employee_id }}/goals/{{ record.goal_id }}/progress
    required fields: employee_id, goal_id, percentComplete
    risk: Update Goal Progress through the BambooHR API.
  reopen_goal:
    endpoint: POST /api/v1/performance/employees/{{ record.employee_id }}/goals/{{ record.goal_id }}/reopen
    required fields: employee_id, goal_id
    risk: Reopen Goal through the BambooHR API.
  update_goal_sharing:
    endpoint: PUT /api/v1/performance/employees/{{ record.employee_id }}/goals/{{ record.goal_id }}/sharedWith
    required fields: employee_id, goal_id
    risk: Update Goal Sharing through the BambooHR API.
  create_scheduling_create_schedule:
    endpoint: POST /api/v1/scheduling/schedules
    required fields: name, locationId, startOfWeek
    risk: Create Schedule through the BambooHR API.
  delete_scheduling_delete_schedule:
    endpoint: DELETE /api/v1/scheduling/schedules/{{ record.id }}
    required fields: id
    risk: Deletes BambooHR data: Delete Schedule.
  update_scheduling_update_schedule:
    endpoint: PATCH /api/v1/scheduling/schedules/{{ record.id }}
    required fields: id
    risk: Update Schedule through the BambooHR API.
  create_scheduling_create_shift:
    endpoint: POST /api/v1/scheduling/shifts
    required fields: scheduleId, status, color, timezone, start, end
    risk: Create Shift through the BambooHR API.
  create_scheduling_publish_shifts:
    endpoint: POST /api/v1/scheduling/shifts/publish
    required fields: shiftIds
    risk: Publish Shifts through the BambooHR API.
  delete_scheduling_delete_shift:
    endpoint: DELETE /api/v1/scheduling/shifts/{{ record.id }}
    required fields: id
    risk: Deletes BambooHR data: Delete Shift.
  update_scheduling_update_shift:
    endpoint: PATCH /api/v1/scheduling/shifts/{{ record.id }}
    required fields: id
    risk: Update Shift through the BambooHR API.
  create_break_policy:
    endpoint: POST /api/v1/time-tracking/break-policies
    required fields: name
    risk: Create Break Policy through the BambooHR API.
  delete_break_policy:
    endpoint: DELETE /api/v1/time-tracking/break-policies/{{ record.id }}
    required fields: id
    risk: Deletes BambooHR data: Delete Break Policy.
  update_break_policy:
    endpoint: PATCH /api/v1/time-tracking/break-policies/{{ record.id }}
    required fields: id
    risk: Update Break Policy through the BambooHR API.
  assign_employees_to_break_policy:
    endpoint: POST /api/v1/time-tracking/break-policies/{{ record.id }}/assign
    required fields: id, employeeIds
    risk: Assign Employees to Break Policy through the BambooHR API.
  set_break_policy_employees:
    endpoint: PUT /api/v1/time-tracking/break-policies/{{ record.id }}/assign
    required fields: id, employeeIds
    risk: Set Employees for Break Policy through the BambooHR API.
  create_break:
    endpoint: POST /api/v1/time-tracking/break-policies/{{ record.id }}/breaks
    required fields: id
    risk: Create Break through the BambooHR API.
  replace_breaks_for_break_policy:
    endpoint: PUT /api/v1/time-tracking/break-policies/{{ record.id }}/breaks
    required fields: id
    risk: Replace Breaks for Break Policy through the BambooHR API.
  sync_break_policy:
    endpoint: PUT /api/v1/time-tracking/break-policies/{{ record.id }}/sync
    required fields: id
    risk: Sync Break Policy through the BambooHR API.
  create_unassign_employees_from_break_policy:
    endpoint: POST /api/v1/time-tracking/break-policies/{{ record.id }}/unassign
    required fields: id, employeeIds
    risk: Unassign Employees from Break Policy through the BambooHR API.
  delete_break:
    endpoint: DELETE /api/v1/time-tracking/breaks/{{ record.id }}
    required fields: id
    risk: Deletes BambooHR data: Delete Break.
  update_break:
    endpoint: PATCH /api/v1/time-tracking/breaks/{{ record.id }}
    required fields: id
    risk: Update Break through the BambooHR API.
  create_project:
    endpoint: POST /api/v1/time-tracking/projects
    required fields: name
    risk: Create Time Tracking Project through the BambooHR API.
  delete_project:
    endpoint: DELETE /api/v1/time-tracking/projects/{{ record.id }}
    required fields: id
    risk: Deletes BambooHR data: Delete Time Tracking Project.
  update_project:
    endpoint: PATCH /api/v1/time-tracking/projects/{{ record.id }}
    required fields: id
    risk: Update Time Tracking Project through the BambooHR API.
  create_project_task:
    endpoint: POST /api/v1/time-tracking/projects/{{ record.project_id }}/tasks
    required fields: project_id, name
    risk: Create Time Tracking Project Task through the BambooHR API.
  create_shift_differential:
    endpoint: POST /api/v1/time-tracking/shift-differentials
    required fields: name, rate, rateType, times
    risk: Create Time Tracking Shift Differential through the BambooHR API.
  delete_shift_differential:
    endpoint: DELETE /api/v1/time-tracking/shift-differentials/{{ record.id }}
    required fields: id
    risk: Deletes BambooHR data: Delete Time Tracking Shift Differential.
  update_shift_differential:
    endpoint: PATCH /api/v1/time-tracking/shift-differentials/{{ record.id }}
    required fields: id
    risk: Update Time Tracking Shift Differential through the BambooHR API.
  delete_task:
    endpoint: DELETE /api/v1/time-tracking/tasks/{{ record.id }}
    required fields: id
    risk: Deletes BambooHR data: Delete Time Tracking Task.
  update_task:
    endpoint: PATCH /api/v1/time-tracking/tasks/{{ record.id }}
    required fields: id
    risk: Update Time Tracking Task through the BambooHR API.
  update_time_off_request_status:
    endpoint: PUT /api/v1/time_off/requests/{{ record.request_id }}/status
    required fields: request_id, status
    risk: Update Time Off Request Status through the BambooHR API.
  delete_clock_entries:
    endpoint: DELETE /api/v1/time_tracking/clock_entries
    risk: Deletes BambooHR data: Delete clock entries.
  store_clock_entries:
    endpoint: POST /api/v1/time_tracking/clock_entries
    risk: Store clock entries through the BambooHR API.
  delete_timesheet_clock_entries_via_post:
    endpoint: POST /api/v1/time_tracking/clock_entries/delete
    required fields: clockEntryIds
    risk: Delete Timesheet Clock Entries through the BambooHR API.
  create_or_update_timesheet_clock_entries:
    endpoint: POST /api/v1/time_tracking/clock_entries/store
    required fields: entries
    risk: Create or Update Timesheet Clock Entries through the BambooHR API.
  clock_in:
    endpoint: POST /api/v1/time_tracking/clock_in/{{ record.employee_id }}
    required fields: employee_id
    risk: Clock in (employee id optional) through the BambooHR API.
  clock_out:
    endpoint: POST /api/v1/time_tracking/clock_out/{{ record.employee_id }}
    required fields: employee_id
    risk: Clock out (employee id optional) through the BambooHR API.
  store_daily_entries:
    endpoint: POST /api/v1/time_tracking/daily_entries
    risk: Store daily entries through the BambooHR API.
  clock_in_data:
    endpoint: POST /api/v1/time_tracking/employee/{{ record.employee_id }}/clock_in/data
    required fields: employee_id
    risk: Edit information on the currently clocked in entry through the BambooHR API.
  clock_out_employee_at_specific_time:
    endpoint: POST /api/v1/time_tracking/employee/{{ record.employee_id }}/clock_out/datetime
    required fields: employee_id
    risk: Clock out an employee at a specific time through the BambooHR API.
  create_timesheet_clock_in_entry:
    endpoint: POST /api/v1/time_tracking/employees/{{ record.employee_id }}/clock_in
    required fields: employee_id
    risk: Create Timesheet Clock-In Entry through the BambooHR API.
  create_timesheet_clock_out_entry:
    endpoint: POST /api/v1/time_tracking/employees/{{ record.employee_id }}/clock_out
    required fields: employee_id
    risk: Create Timesheet Clock-Out Entry through the BambooHR API.
  delete_timesheet_hour_entries_via_post:
    endpoint: POST /api/v1/time_tracking/hour_entries/delete
    required fields: hourEntryIds
    risk: Delete Timesheet Hour Entries through the BambooHR API.
  create_or_update_timesheet_hour_entries:
    endpoint: POST /api/v1/time_tracking/hour_entries/store
    required fields: hours
    risk: Create or Update Timesheet Hour Entries through the BambooHR API.
  create_time_tracking_project:
    endpoint: POST /api/v1/time_tracking/projects
    required fields: name
    risk: Create Time Tracking Project through the BambooHR API.
  approve_employee_timesheets:
    endpoint: POST /api/v1/time_tracking/timesheets/approve
    required fields: lastChanged, timesheets
    risk: Approve employee timesheets through the BambooHR API.
  clock_out_and_approve_employee_timesheets:
    endpoint: POST /api/v1/time_tracking/timesheets/clock_out_and_approve
    risk: Approve timesheets for employees that are currently clocked in through the BambooHR API.
  create_time_tracking_hour_record:
    endpoint: POST /api/v1/timetracking/add
    required fields: dateHoursWorked, employeeId, hoursWorked, rateType, timeTrackingId
    risk: Create Hour Record through the BambooHR API.
  update_time_tracking_record:
    endpoint: PUT /api/v1/timetracking/adjust
    required fields: timeTrackingId, hoursWorked
    risk: Update Hour Record through the BambooHR API.
  delete_time_tracking_hour_record:
    endpoint: DELETE /api/v1/timetracking/delete/{{ record.id }}
    required fields: id
    risk: Deletes BambooHR data: Delete Hour Record.
  create_or_update_time_tracking_hour_records:
    endpoint: POST /api/v1/timetracking/record
    risk: Create or Update Hour Records through the BambooHR API.
  create_training_category:
    endpoint: POST /api/v1/training/category
    required fields: name
    risk: Create Training Category through the BambooHR API.
  delete_training_category:
    endpoint: DELETE /api/v1/training/category/{{ record.training_category_id }}
    required fields: training_category_id
    risk: Deletes BambooHR data: Delete Training Category.
  update_training_category:
    endpoint: PUT /api/v1/training/category/{{ record.training_category_id }}
    required fields: training_category_id, name
    risk: Update Training Category through the BambooHR API.
  create_employee_training_record:
    endpoint: POST /api/v1/training/record/employee/{{ record.employee_id }}
    required fields: employee_id, completed, type
    risk: Create Employee Training Record through the BambooHR API.
  delete_employee_training_record:
    endpoint: DELETE /api/v1/training/record/{{ record.employee_training_record_id }}
    required fields: employee_training_record_id
    risk: Deletes BambooHR data: Delete Employee Training Record.
  update_employee_training_record:
    endpoint: PUT /api/v1/training/record/{{ record.employee_training_record_id }}
    required fields: employee_training_record_id, completed
    risk: Update Employee Training Record through the BambooHR API.
  create_training_type:
    endpoint: POST /api/v1/training/type
    required fields: name
    risk: Create Training Type through the BambooHR API.
  delete_training_type:
    endpoint: DELETE /api/v1/training/type/{{ record.training_type_id }}
    required fields: training_type_id
    risk: Deletes BambooHR data: Delete Training Type.
  update_training_type:
    endpoint: PUT /api/v1/training/type/{{ record.training_type_id }}
    required fields: training_type_id
    risk: Update Training Type through the BambooHR API.
  create_webhook:
    endpoint: POST /api/v1/webhooks
    required fields: name, url, format
    risk: Create Webhook through the BambooHR API.
  delete_webhook:
    endpoint: DELETE /api/v1/webhooks/{{ record.id }}
    required fields: id
    risk: Deletes BambooHR data: Delete Webhook.
  update_webhook:
    endpoint: PUT /api/v1/webhooks/{{ record.id }}
    required fields: id, name, url, format
    risk: Update Webhook through the BambooHR API.
  assign_time_off_policies_v1_1:
    endpoint: PUT /api/v1_1/employees/{{ record.employee_id }}/time_off/policies
    required fields: employee_id
    risk: Assign Time Off Policies v1.1 through the BambooHR API.
  create_table_row_v1_1:
    endpoint: POST /api/v1_1/employees/{{ record.id }}/tables/{{ record.table }}
    required fields: id, table
    risk: Create Table Row v1.1 through the BambooHR API.
  update_table_row_v1_1:
    endpoint: POST /api/v1_1/employees/{{ record.id }}/tables/{{ record.table }}/{{ record.row_id }}
    required fields: id, table, row_id
    risk: Update Table Row v1.1 through the BambooHR API.
  update_goal_v1_1:
    endpoint: PUT /api/v1_1/performance/employees/{{ record.employee_id }}/goals/{{ record.goal_id }}
    required fields: employee_id, goal_id, title, dueDate, sharedWithEmployeeIds
    risk: Update Goal (v1.1) through the BambooHR API.
  update_company_benefit_properties:
    endpoint: POST /api/v1_2/benefit/company_benefit/{{ record.id }}
    required fields: id
    risk: Update a company benefit through the BambooHR API.

SECURITY
  read risk: external BambooHR API reads across HR, applicant tracking, benefits, payroll-adjacent, time off, training, goals, and metadata resources
  write risk: creates, updates, assigns, approves, or deletes BambooHR HR records according to the selected reverse-ETL action
  approval: reverse ETL writes require plan preview and approval token
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Declared bamboo-hr API commands.
  Usage: pm bamboo-hr <command> [flags]
  Global flags:
    --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
  Other Commands
    operations get-api-v1-1-employees-employee-id-time-off-policies - Declared etl: GET /api/v1_1/employees/{employeeId}/time_off/policies. [intent=etl availability=implemented stream=employee_time_off_policies_v1_1]; notes: Provider GET /api/v1_1/employees/{employeeId}/time_off/policies is bound to the existing employee_time_off_policies_v1_1 stream with its connector-owned schema and pagination contract.
    operations put-api-v1-1-employees-employee-id-time-off-policies - Declared direct write: PUT /api/v1_1/employees/{employeeId}/time_off/policies. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: PUT /api/v1_1/employees/{employeeId}/time_off/policies.; notes: Blocked: locked source operation bamboo-hr.provider.assign-time-off-policies-v1-1-2 has no declaration-owned executable direct_write route.
    operations post-api-v1-1-employees-id-tables-table - Declared direct write: POST /api/v1_1/employees/{id}/tables/{table}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1_1/employees/{id}/tables/{table}.; notes: Blocked: locked source operation bamboo-hr.provider.create-table-row-v1-1-3 has no declaration-owned executable direct_write route.
    operations post-api-v1-1-employees-id-tables-table-row-id - Declared direct write: POST /api/v1_1/employees/{id}/tables/{table}/{rowId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1_1/employees/{id}/tables/{table}/{rowId}.; notes: Blocked: locked source operation bamboo-hr.provider.update-table-row-v1-1-4 has no declaration-owned executable direct_write route.
    operations put-api-v1-1-performance-employees-employee-id-goals-goal-id - Declared direct write: PUT /api/v1_1/performance/employees/{employeeId}/goals/{goalId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: PUT /api/v1_1/performance/employees/{employeeId}/goals/{goalId}.; notes: Blocked: locked source operation bamboo-hr.provider.update-goal-v1-1-5 has no declaration-owned executable direct_write route.
    operations get-api-v1-1-performance-employees-employee-id-goals-aggregate - Declared etl: GET /api/v1_1/performance/employees/{employeeId}/goals/aggregate. [intent=etl availability=implemented stream=goals_aggregate_v1_1]; notes: Provider GET /api/v1_1/performance/employees/{employeeId}/goals/aggregate is bound to the existing goals_aggregate_v1_1 stream with its connector-owned schema and pagination contract.
    operations get-api-v1-1-performance-employees-employee-id-goals-filters - Declared etl: GET /api/v1_1/performance/employees/{employeeId}/goals/filters. [intent=etl availability=implemented stream=goals_filters_v1_1]; notes: Provider GET /api/v1_1/performance/employees/{employeeId}/goals/filters is bound to the existing goals_filters_v1_1 stream with its connector-owned schema and pagination contract.
    operations get-api-v1-2-datasets - Declared etl: GET /api/v1_2/datasets. [intent=etl availability=implemented stream=datasets_v1_2]; notes: Provider GET /api/v1_2/datasets is bound to the existing datasets_v1_2 stream with its connector-owned schema and pagination contract.
    operations post-api-v1-2-datasets-dataset-name-field-options - Declared direct write: POST /api/v1_2/datasets/{datasetName}/field-options. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Read-like POST endpoint requires request-body query execution; the current declarative read path does not send stream bodies and this must not be exposed as a write action.; notes: Blocked: locked source operation bamboo-hr.provider.get-field-options-v1-2-9 has no declaration-owned executable direct_write route.
    operations get-api-v1-2-datasets-dataset-name-fields - Declared etl: GET /api/v1_2/datasets/{datasetName}/fields. [intent=etl availability=implemented stream=fields_from_dataset_v1_2]; notes: Provider GET /api/v1_2/datasets/{datasetName}/fields is bound to the existing fields_from_dataset_v1_2 stream with its connector-owned schema and pagination contract.
    operations get-api-v1-2-performance-employees-employee-id-goals-aggregate - Declared etl: GET /api/v1_2/performance/employees/{employeeId}/goals/aggregate. [intent=etl availability=implemented stream=goals_aggregate_v1_2]; notes: Provider GET /api/v1_2/performance/employees/{employeeId}/goals/aggregate is bound to the existing goals_aggregate_v1_2 stream with its connector-owned schema and pagination contract.
    operations get-api-v1-2-performance-employees-employee-id-goals-filters - Declared etl: GET /api/v1_2/performance/employees/{employeeId}/goals/filters. [intent=etl availability=implemented stream=goals_filters_v1_2]; notes: Provider GET /api/v1_2/performance/employees/{employeeId}/goals/filters is bound to the existing goals_filters_v1_2 stream with its connector-owned schema and pagination contract.
    operations get-api-v1-alert-configurations - Declared direct read: GET /api/v1/alert-configurations. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.list-alert-configurations-13 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations post-api-v1-alert-configurations - Declared direct write: POST /api/v1/alert-configurations. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.create-alert-configuration-14 has no declaration-owned executable direct_write route.
    operations get-api-v1-alert-configurations-id - Declared direct read: GET /api/v1/alert-configurations/{id}. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.get-alert-configuration-15 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations put-api-v1-alert-configurations-id - Declared direct write: PUT /api/v1/alert-configurations/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.replace-alert-configuration-16 has no declaration-owned executable direct_write route.
    operations get-api-v1-alerts - Declared direct read: GET /api/v1/alerts. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.list-alert-templates-17 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations post-api-v1-applicant-tracking-application - Declared direct write: POST /api/v1/applicant_tracking/application. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: File, photo, CSV/PDF, or multipart payload is outside the JSON record/write dialect.; notes: Blocked: locked source operation bamboo-hr.provider.create-candidate-18 has no declaration-owned executable direct_write route.
    operations get-api-v1-applicant-tracking-applications - Declared etl: GET /api/v1/applicant_tracking/applications. [intent=etl availability=implemented stream=applications]; notes: Provider GET /api/v1/applicant_tracking/applications is bound to the existing applications stream with its connector-owned schema and pagination contract.
    operations get-api-v1-applicant-tracking-applications-application-id - Declared etl: GET /api/v1/applicant_tracking/applications/{applicationId}. [intent=etl availability=implemented stream=application_details]; notes: Provider GET /api/v1/applicant_tracking/applications/{applicationId} is bound to the existing application_details stream with its connector-owned schema and pagination contract.
    operations post-api-v1-applicant-tracking-applications-application-id-comments - Declared direct write: POST /api/v1/applicant_tracking/applications/{applicationId}/comments. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/applicant_tracking/applications/{applicationId}/comments.; notes: Blocked: locked source operation bamboo-hr.provider.create-application-comment-21 has no declaration-owned executable direct_write route.
    operations post-api-v1-applicant-tracking-applications-application-id-status - Declared direct write: POST /api/v1/applicant_tracking/applications/{applicationId}/status. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/applicant_tracking/applications/{applicationId}/status.; notes: Blocked: locked source operation bamboo-hr.provider.update-applicant-status-22 has no declaration-owned executable direct_write route.
    operations get-api-v1-applicant-tracking-hiring-leads - Declared etl: GET /api/v1/applicant_tracking/hiring_leads. [intent=etl availability=implemented stream=hiring_leads]; notes: Provider GET /api/v1/applicant_tracking/hiring_leads is bound to the existing hiring_leads stream with its connector-owned schema and pagination contract.
    operations post-api-v1-applicant-tracking-job-opening - Declared direct write: POST /api/v1/applicant_tracking/job_opening. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: File, photo, CSV/PDF, or multipart payload is outside the JSON record/write dialect.; notes: Blocked: locked source operation bamboo-hr.provider.create-job-opening-24 has no declaration-owned executable direct_write route.
    operations get-api-v1-applicant-tracking-jobs - Declared etl: GET /api/v1/applicant_tracking/jobs. [intent=etl availability=implemented stream=job_summaries]; notes: Provider GET /api/v1/applicant_tracking/jobs is bound to the existing job_summaries stream with its connector-owned schema and pagination contract.
    operations get-api-v1-applicant-tracking-locations - Declared etl: GET /api/v1/applicant_tracking/locations. [intent=etl availability=implemented stream=company_locations]; notes: Provider GET /api/v1/applicant_tracking/locations is bound to the existing company_locations stream with its connector-owned schema and pagination contract.
    operations get-api-v1-applicant-tracking-statuses - Declared etl: GET /api/v1/applicant_tracking/statuses. [intent=etl availability=implemented stream=statuses]; notes: Provider GET /api/v1/applicant_tracking/statuses is bound to the existing statuses stream with its connector-owned schema and pagination contract.
    operations get-api-v1-benefit-company-benefit - Declared etl: GET /api/v1/benefit/company_benefit. [intent=etl availability=implemented stream=company_benefits]; notes: Provider GET /api/v1/benefit/company_benefit is bound to the existing company_benefits stream with its connector-owned schema and pagination contract.
    operations get-api-v1-benefit-employee-benefit - Declared etl: GET /api/v1/benefit/employee_benefit. [intent=etl availability=implemented stream=employee_benefits]; notes: Provider GET /api/v1/benefit/employee_benefit is bound to the existing employee_benefits stream with its connector-owned schema and pagination contract.
    operations get-api-v1-benefit-member-benefit - Declared etl: GET /api/v1/benefit/member_benefit. [intent=etl availability=implemented stream=member_benefit_events]; notes: Provider GET /api/v1/benefit/member_benefit is bound to the existing member_benefit_events stream with its connector-owned schema and pagination contract.
    operations get-api-v1-benefitcoverages - Declared etl: GET /api/v1/benefitcoverages. [intent=etl availability=implemented stream=benefit_coverages]; notes: Provider GET /api/v1/benefitcoverages is bound to the existing benefit_coverages stream with its connector-owned schema and pagination contract.
    operations get-api-v1-benefits-member-benefits - Declared etl: GET /api/v1/benefits/member-benefits. [intent=etl availability=implemented stream=member_benefits]; notes: Provider GET /api/v1/benefits/member-benefits is bound to the existing member_benefits stream with its connector-owned schema and pagination contract.
    operations get-api-v1-benefits-settings-deduction-types-all - Declared etl: GET /api/v1/benefits/settings/deduction_types/all. [intent=etl availability=implemented stream=benefit_deduction_types]; notes: Provider GET /api/v1/benefits/settings/deduction_types/all is bound to the existing benefit_deduction_types stream with its connector-owned schema and pagination contract.
    operations get-api-v1-company-information - Declared etl: GET /api/v1/company_information. [intent=etl availability=implemented stream=company_information]; notes: Provider GET /api/v1/company_information is bound to the existing company_information stream with its connector-owned schema and pagination contract.
    operations patch-api-v1-company-profile-data-company-information - Declared direct write: PATCH /api/v1/company-profile-data/company-information. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.patch-company-profile-company-information-35 has no declaration-owned executable direct_write route.
    operations put-api-v1-company-profile-data-display-name - Declared direct write: PUT /api/v1/company-profile-data/display-name. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.put-company-profile-display-name-36 has no declaration-owned executable direct_write route.
    operations put-api-v1-company-profile-data-industry-codes - Declared direct write: PUT /api/v1/company-profile-data/industry-codes. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.put-company-industry-codes-37 has no declaration-owned executable direct_write route.
    operations get-api-v1-company-profile-integrations - Declared etl: GET /api/v1/company-profile-integrations. [intent=etl availability=implemented stream=company_profile_integrations]; notes: Provider GET /api/v1/company-profile-integrations is bound to the existing company_profile_integrations stream with its connector-owned schema and pagination contract.
    operations get-api-v1-compensation-benchmarks - Declared direct read: GET /api/v1/compensation/benchmarks. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.list-compensation-benchmarks-39 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations post-api-v1-compensation-benchmarks - Declared direct write: POST /api/v1/compensation/benchmarks. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.create-compensation-benchmark-40 has no declaration-owned executable direct_write route.
    operations put-api-v1-compensation-benchmarks - Declared direct write: PUT /api/v1/compensation/benchmarks. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.update-compensation-benchmark-41 has no declaration-owned executable direct_write route.
    operations delete-api-v1-compensation-benchmarks-id - Declared direct write: DELETE /api/v1/compensation/benchmarks/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.delete-compensation-benchmark-42 has no declaration-owned executable direct_write route.
    operations get-api-v1-compensation-benchmarks-details - Declared direct read: GET /api/v1/compensation/benchmarks/details. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.get-compensation-benchmark-details-43 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations get-api-v1-compensation-benchmarks-details-export - Declared direct read: GET /api/v1/compensation/benchmarks/details/export. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.export-compensation-benchmark-details-44 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations post-api-v1-compensation-benchmarks-import - Declared direct write: POST /api/v1/compensation/benchmarks/import. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.import-compensation-benchmarks-45 has no declaration-owned executable direct_write route.
    operations delete-api-v1-compensation-benchmarks-sources - Declared direct write: DELETE /api/v1/compensation/benchmarks/sources. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.delete-compensation-benchmark-source-46 has no declaration-owned executable direct_write route.
    operations get-api-v1-compensation-benchmarks-sources - Declared direct read: GET /api/v1/compensation/benchmarks/sources. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.list-compensation-benchmark-sources-47 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations post-api-v1-compensation-benchmarks-sources - Declared direct write: POST /api/v1/compensation/benchmarks/sources. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.create-compensation-benchmark-source-48 has no declaration-owned executable direct_write route.
    operations put-api-v1-compensation-benchmarks-sources - Declared direct write: PUT /api/v1/compensation/benchmarks/sources. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.update-compensation-benchmark-sources-49 has no declaration-owned executable direct_write route.
    operations get-api-v1-compensation-equity-settings - Declared direct read: GET /api/v1/compensation/equity/settings. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.db49fb29f9f04d59afad7c01ce860418-50 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations put-api-v1-compensation-equity-settings - Declared direct write: PUT /api/v1/compensation/equity/settings. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.c5880b509783cd9d7fce9ddf5d6af1be-51 has no declaration-owned executable direct_write route.
    operations get-api-v1-compensation-planning-cycles - Declared direct read: GET /api/v1/compensation/planning_cycles. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.b65f246186b41a9783a9397c11c703b4-52 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations post-api-v1-compensation-planning-cycles - Declared direct write: POST /api/v1/compensation/planning_cycles. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.e2ac4e1535f296cb8901f209e04caa83-53 has no declaration-owned executable direct_write route.
    operations delete-api-v1-compensation-planning-cycles-id - Declared direct write: DELETE /api/v1/compensation/planning_cycles/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.22ad75be25455279e2987c80851af5fc-54 has no declaration-owned executable direct_write route.
    operations get-api-v1-compensation-planning-cycles-id - Declared direct read: GET /api/v1/compensation/planning_cycles/{id}. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.5c2b55158b0950b1e9211655666645b6-55 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations put-api-v1-compensation-planning-cycles-id - Declared direct write: PUT /api/v1/compensation/planning_cycles/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.100b0cf8c5207b35697ff10370fd5fe1-56 has no declaration-owned executable direct_write route.
    operations get-api-v1-compensation-planning-cycles-id-admins - Declared direct read: GET /api/v1/compensation/planning_cycles/{id}/admins. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.b3c51254de6918637a971fe4af382a53-57 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations post-api-v1-compensation-planning-cycles-id-admins - Declared direct write: POST /api/v1/compensation/planning_cycles/{id}/admins. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.c7c32ed5278ac67e2e518bf7484a75dc-58 has no declaration-owned executable direct_write route.
    operations delete-api-v1-compensation-planning-cycles-id-admins-employee-id - Declared direct write: DELETE /api/v1/compensation/planning_cycles/{id}/admins/{employeeId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.ef7619b0ee4c8dc079aaea870cfbe81b-59 has no declaration-owned executable direct_write route.
    operations get-api-v1-compensation-planning-cycles-id-approvals - Declared direct read: GET /api/v1/compensation/planning_cycles/{id}/approvals. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.4e886b18264480611f380805301c49c4-60 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations put-api-v1-compensation-planning-cycles-id-approvals-template-id - Declared direct write: PUT /api/v1/compensation/planning_cycles/{id}/approvals/{templateId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.cf87b8e09a001b6fb81dfce6c20ab9e3-61 has no declaration-owned executable direct_write route.
    operations delete-api-v1-compensation-planning-cycles-id-approvals-employee-employee-id - Declared direct write: DELETE /api/v1/compensation/planning_cycles/{id}/approvals/employee/{employeeId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.b1e467e0eef72350eec61fcfeaf4e19d-62 has no declaration-owned executable direct_write route.
    operations post-api-v1-compensation-planning-cycles-id-approvals-final-approver-employee-id - Declared direct write: POST /api/v1/compensation/planning_cycles/{id}/approvals/final_approver/{employeeId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.5c4aab35a34f5760ec044104b5232bf5-63 has no declaration-owned executable direct_write route.
    operations get-api-v1-compensation-planning-cycles-id-budgets - Declared direct read: GET /api/v1/compensation/planning_cycles/{id}/budgets. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.7efceaee2c010f88244dd01ee81e6e7b-64 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations put-api-v1-compensation-planning-cycles-id-budgets-breakdown - Declared direct write: PUT /api/v1/compensation/planning_cycles/{id}/budgets/breakdown. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.1d64402ee192568adbd5e3179a91e6e2-65 has no declaration-owned executable direct_write route.
    operations put-api-v1-compensation-planning-cycles-id-budgets-guidelines - Declared direct write: PUT /api/v1/compensation/planning_cycles/{id}/budgets/guidelines. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.dacd313af2106213fc4696175941ce65-66 has no declaration-owned executable direct_write route.
    operations post-api-v1-compensation-planning-cycles-id-budgets-import - Declared direct write: POST /api/v1/compensation/planning_cycles/{id}/budgets/import. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.1d1fc0f164cb51973a0206b8e2fb2d2d-67 has no declaration-owned executable direct_write route.
    operations get-api-v1-compensation-planning-cycles-id-change-comm - Declared direct read: GET /api/v1/compensation/planning_cycles/{id}/change_comm. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.d6987e300672a00c7cfe59afebb64156-68 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations put-api-v1-compensation-planning-cycles-id-change-comm-template - Declared direct write: PUT /api/v1/compensation/planning_cycles/{id}/change_comm/template. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.c79f9c5950f983e59d2626faa30c00a1-69 has no declaration-owned executable direct_write route.
    operations put-api-v1-compensation-planning-cycles-id-complete - Declared direct write: PUT /api/v1/compensation/planning_cycles/{id}/complete. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.f4b431363af6573af46750f32632e88b-70 has no declaration-owned executable direct_write route.
    operations delete-api-v1-compensation-planning-cycles-id-employees - Declared direct write: DELETE /api/v1/compensation/planning_cycles/{id}/employees. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.89a5068111ec499135c7d6e9a53d5a30-71 has no declaration-owned executable direct_write route.
    operations get-api-v1-compensation-planning-cycles-id-employees - Declared direct read: GET /api/v1/compensation/planning_cycles/{id}/employees. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.a6b8da1348a3151fe95adc03aaf64447-72 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations post-api-v1-compensation-planning-cycles-id-employees - Declared direct write: POST /api/v1/compensation/planning_cycles/{id}/employees. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.3958585c861325ea7a2cd30a8c74f042-73 has no declaration-owned executable direct_write route.
    operations put-api-v1-compensation-planning-cycles-id-launch - Declared direct write: PUT /api/v1/compensation/planning_cycles/{id}/launch. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.3a19f07aa737dc826ba43b9a1c1cd257-74 has no declaration-owned executable direct_write route.
    operations post-api-v1-compensation-planning-cycles-id-recommendations - Declared direct write: POST /api/v1/compensation/planning_cycles/{id}/recommendations. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.f3883a522dadbe9e11b34f8b656e3adb-75 has no declaration-owned executable direct_write route.
    operations post-api-v1-compensation-planning-cycles-id-recommendations-send - Declared direct write: POST /api/v1/compensation/planning_cycles/{id}/recommendations/send. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.a05b6d5f564f805d688ff2c1e37c3990-76 has no declaration-owned executable direct_write route.
    operations get-api-v1-compensation-planning-cycles-id-summary - Declared direct read: GET /api/v1/compensation/planning_cycles/{id}/summary. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.9bc279d788f6e86b4cd8b2e0d3de91b1-77 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations get-api-v1-compensation-planning-cycles-id-worksheet - Declared direct read: GET /api/v1/compensation/planning_cycles/{id}/worksheet. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.329acecaa6df729733d0752aa9f6b204-78 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations get-api-v1-compensation-planning-cycles-id-worksheet-export - Declared direct read: GET /api/v1/compensation/planning_cycles/{id}/worksheet/export. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.593d5bff120edf2a218a92022a682728-79 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations get-api-v1-compensation-tools - Declared direct read: GET /api/v1/compensation/tools. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.9f398e2652ea47a6dc5121ce5184222a-80 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations get-api-v1-compensation-total-rewards-employee-id - Declared direct read: GET /api/v1/compensation/total_rewards/{employeeId}. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.check-total-rewards-profile-81 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations get-api-v1-compensation-total-rewards-employee-id-printable - Declared direct read: GET /api/v1/compensation/total_rewards/{employeeId}/printable. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.get-total-rewards-printable-statement-82 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations get-api-v1-compensation-total-rewards-employee-id-statement - Declared direct read: GET /api/v1/compensation/total_rewards/{employeeId}/statement. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.get-total-rewards-statement-83 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations delete-api-v1-compensation-total-rewards-custom-disclaimer - Declared direct write: DELETE /api/v1/compensation/total_rewards/custom_disclaimer. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.remove-total-rewards-custom-disclaimer-84 has no declaration-owned executable direct_write route.
    operations put-api-v1-compensation-total-rewards-custom-disclaimer - Declared direct write: PUT /api/v1/compensation/total_rewards/custom_disclaimer. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.set-total-rewards-custom-disclaimer-85 has no declaration-owned executable direct_write route.
    operations delete-api-v1-compensation-total-rewards-employees - Declared direct write: DELETE /api/v1/compensation/total_rewards/employees. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.remove-total-rewards-employees-86 has no declaration-owned executable direct_write route.
    operations post-api-v1-compensation-total-rewards-employees - Declared direct write: POST /api/v1/compensation/total_rewards/employees. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.add-total-rewards-employees-87 has no declaration-owned executable direct_write route.
    operations put-api-v1-compensation-total-rewards-onboarding-step-name - Declared direct write: PUT /api/v1/compensation/total_rewards/onboarding/{stepName}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.set-total-rewards-onboarding-step-88 has no declaration-owned executable direct_write route.
    operations get-api-v1-custom-reports - Declared etl: GET /api/v1/custom-reports. [intent=etl availability=implemented stream=reports]; notes: Provider GET /api/v1/custom-reports is bound to the existing reports stream with its connector-owned schema and pagination contract.
    operations get-api-v1-custom-reports-report-id - Declared etl: GET /api/v1/custom-reports/{reportId}. [intent=etl availability=implemented stream=report_by_id]; notes: Provider GET /api/v1/custom-reports/{reportId} is bound to the existing report_by_id stream with its connector-owned schema and pagination contract.
    operations get-api-v1-custom-reports-legacy-field-map - Declared direct read: GET /api/v1/custom-reports/legacy-field-map. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.get-legacy-report-field-map-91 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations get-api-v1-custom-reports-legacy-id-map - Declared direct read: GET /api/v1/custom-reports/legacy-id-map. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.get-legacy-report-id-map-92 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations get-api-v1-datasets - Declared etl: GET /api/v1/datasets. [intent=etl availability=implemented stream=datasets_v1]; notes: Provider GET /api/v1/datasets is bound to the existing datasets_v1 stream with its connector-owned schema and pagination contract.
    operations post-api-v1-datasets-dataset-name - Declared direct write: POST /api/v1/datasets/{datasetName}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Deprecated ad-hoc report endpoint; BambooHR directs callers to dataset APIs instead.; notes: Blocked: locked source operation bamboo-hr.provider.get-data-from-dataset-v1-94 has no declaration-owned executable direct_write route.
    operations post-api-v1-datasets-dataset-name-field-options - Declared direct write: POST /api/v1/datasets/{datasetName}/field-options. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Deprecated ad-hoc report endpoint; BambooHR directs callers to dataset APIs instead.; notes: Blocked: locked source operation bamboo-hr.provider.get-field-options-v1-95 has no declaration-owned executable direct_write route.
    operations get-api-v1-datasets-dataset-name-fields - Declared etl: GET /api/v1/datasets/{datasetName}/fields. [intent=etl availability=implemented stream=fields_from_dataset_v1]; notes: Provider GET /api/v1/datasets/{datasetName}/fields is bound to the existing fields_from_dataset_v1 stream with its connector-owned schema and pagination contract.
    operations get-api-v1-employee-verifications-employees-employee-id - Declared direct read: GET /api/v1/employee-verifications/employees/{employeeId}. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.list-employee-verifications-by-employee-97 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations put-api-v1-employee-verifications-employees-employee-id-verification-id - Declared direct write: PUT /api/v1/employee-verifications/employees/{employeeId}/{verificationId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.update-employee-verification-98 has no declaration-owned executable direct_write route.
    operations get-api-v1-employee-verifications-integration - Declared direct read: GET /api/v1/employee-verifications/integration. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.get-employee-verification-integration-99 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations put-api-v1-employee-verifications-integration - Declared direct write: PUT /api/v1/employee-verifications/integration. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.update-employee-verification-integration-100 has no declaration-owned executable direct_write route.
    operations post-api-v1-employee-verifications-users-user-id-send-email - Declared direct write: POST /api/v1/employee-verifications/users/{userId}/send-email. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.send-employee-verification-lifecycle-email-by-user-101 has no declaration-owned executable direct_write route.
    operations get-api-v1-employeedependents - Declared etl: GET /api/v1/employeedependents. [intent=etl availability=implemented stream=employee_dependents]; notes: Provider GET /api/v1/employeedependents is bound to the existing employee_dependents stream with its connector-owned schema and pagination contract.
    operations post-api-v1-employeedependents - Declared direct write: POST /api/v1/employeedependents. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/employeedependents.; notes: Blocked: locked source operation bamboo-hr.provider.create-employee-dependent-103 has no declaration-owned executable direct_write route.
    operations get-api-v1-employeedependents-id - Declared etl: GET /api/v1/employeedependents/{id}. [intent=etl availability=implemented stream=employee_dependent]; notes: Provider GET /api/v1/employeedependents/{id} is bound to the existing employee_dependent stream with its connector-owned schema and pagination contract.
    operations put-api-v1-employeedependents-id - Declared direct write: PUT /api/v1/employeedependents/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: PUT /api/v1/employeedependents/{id}.; notes: Blocked: locked source operation bamboo-hr.provider.update-employee-dependent-105 has no declaration-owned executable direct_write route.
    operations get-api-v1-employees - Declared etl: GET /api/v1/employees. [intent=etl availability=implemented stream=employee_roster]; notes: Provider GET /api/v1/employees is bound to the existing employee_roster stream with its connector-owned schema and pagination contract.
    operations post-api-v1-employees - Declared direct write: POST /api/v1/employees. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: File, photo, CSV/PDF, or multipart payload is outside the JSON record/write dialect.; notes: Blocked: locked source operation bamboo-hr.provider.create-employee-107 has no declaration-owned executable direct_write route.
    operations get-api-v1-employees-employee-id-onboarding-experiences - Declared direct read: GET /api/v1/employees/{employeeId}/onboarding-experiences. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.0158de7cde2a4c4cf577f0b25070d809-108 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations post-api-v1-employees-employee-id-onboarding-experiences - Declared direct write: POST /api/v1/employees/{employeeId}/onboarding-experiences. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.288aa996aba16d7a495c62321ea999a9-109 has no declaration-owned executable direct_write route.
    operations get-api-v1-employees-employee-id-onboarding-experiences-onboarding-experience-id - Declared direct read: GET /api/v1/employees/{employeeId}/onboarding-experiences/{onboardingExperienceId}. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.847dd061d1d1859e7ce8cb3adfc9faf2-110 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations post-api-v1-employees-employee-id-photo - Declared direct write: POST /api/v1/employees/{employeeId}/photo. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: File, photo, CSV/PDF, or multipart payload is outside the JSON record/write dialect.; notes: Blocked: locked source operation bamboo-hr.provider.upload-employee-photo-111 has no declaration-owned executable direct_write route.
    operations get-api-v1-employees-employee-id-photo-size - Declared direct read: GET /api/v1/employees/{employeeId}/photo/{size}. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.get-employee-photo-112 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations put-api-v1-employees-employee-id-time-off-balance-adjustment - Declared direct write: PUT /api/v1/employees/{employeeId}/time_off/balance_adjustment. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: PUT /api/v1/employees/{employeeId}/time_off/balance_adjustment.; notes: Blocked: locked source operation bamboo-hr.provider.adjust-time-off-balance-113 has no declaration-owned executable direct_write route.
    operations get-api-v1-employees-employee-id-time-off-calculator - Declared etl: GET /api/v1/employees/{employeeId}/time_off/calculator. [intent=etl availability=implemented stream=time_off_balance]; notes: Provider GET /api/v1/employees/{employeeId}/time_off/calculator is bound to the existing time_off_balance stream with its connector-owned schema and pagination contract.
    operations put-api-v1-employees-employee-id-time-off-history - Declared direct write: PUT /api/v1/employees/{employeeId}/time_off/history. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: PUT /api/v1/employees/{employeeId}/time_off/history.; notes: Blocked: locked source operation bamboo-hr.provider.create-time-off-history-115 has no declaration-owned executable direct_write route.
    operations get-api-v1-employees-employee-id-time-off-policies - Declared etl: GET /api/v1/employees/{employeeId}/time_off/policies. [intent=etl availability=implemented stream=employee_time_off_policies]; notes: Provider GET /api/v1/employees/{employeeId}/time_off/policies is bound to the existing employee_time_off_policies stream with its connector-owned schema and pagination contract.
    operations put-api-v1-employees-employee-id-time-off-policies - Declared direct write: PUT /api/v1/employees/{employeeId}/time_off/policies. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: PUT /api/v1/employees/{employeeId}/time_off/policies.; notes: Blocked: locked source operation bamboo-hr.provider.assign-time-off-policies-v1-117 has no declaration-owned executable direct_write route.
    operations put-api-v1-employees-employee-id-time-off-request - Declared direct write: PUT /api/v1/employees/{employeeId}/time_off/request. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: PUT /api/v1/employees/{employeeId}/time_off/request.; notes: Blocked: locked source operation bamboo-hr.provider.create-time-off-request-118 has no declaration-owned executable direct_write route.
    operations delete-api-v1-employees-id - Declared direct write: DELETE /api/v1/employees/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.delete-employee-119 has no declaration-owned executable direct_write route.
    operations get-api-v1-employees-id - Declared direct read: GET /api/v1/employees/{id}. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.get-employee-120 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations post-api-v1-employees-id - Declared direct write: POST /api/v1/employees/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: File, photo, CSV/PDF, or multipart payload is outside the JSON record/write dialect.; notes: Blocked: locked source operation bamboo-hr.provider.update-employee-121 has no declaration-owned executable direct_write route.
    operations post-api-v1-employees-id-files - Declared direct write: POST /api/v1/employees/{id}/files. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: File, photo, CSV/PDF, or multipart payload is outside the JSON record/write dialect.; notes: Blocked: locked source operation bamboo-hr.provider.upload-employee-file-122 has no declaration-owned executable direct_write route.
    operations delete-api-v1-employees-id-files-file-id - Declared direct write: DELETE /api/v1/employees/{id}/files/{fileId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: File, photo, CSV/PDF, or multipart payload is outside the JSON record/write dialect.; notes: Blocked: locked source operation bamboo-hr.provider.delete-employee-file-123 has no declaration-owned executable direct_write route.
    operations get-api-v1-employees-id-files-file-id - Declared direct read: GET /api/v1/employees/{id}/files/{fileId}. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.get-employee-file-124 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations post-api-v1-employees-id-files-file-id - Declared direct write: POST /api/v1/employees/{id}/files/{fileId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: File, photo, CSV/PDF, or multipart payload is outside the JSON record/write dialect.; notes: Blocked: locked source operation bamboo-hr.provider.update-employee-file-125 has no declaration-owned executable direct_write route.
    operations get-api-v1-employees-id-files-view - Declared direct read: GET /api/v1/employees/{id}/files/view. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.list-employee-files-126 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations get-api-v1-employees-id-tables-table - Declared etl: GET /api/v1/employees/{id}/tables/{table}. [intent=etl availability=implemented stream=employee_table_data]; notes: Provider GET /api/v1/employees/{id}/tables/{table} is bound to the existing employee_table_data stream with its connector-owned schema and pagination contract.
    operations post-api-v1-employees-id-tables-table - Declared direct write: POST /api/v1/employees/{id}/tables/{table}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/employees/{id}/tables/{table}.; notes: Blocked: locked source operation bamboo-hr.provider.create-table-row-v1-128 has no declaration-owned executable direct_write route.
    operations delete-api-v1-employees-id-tables-table-row-id - Declared direct write: DELETE /api/v1/employees/{id}/tables/{table}/{rowId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: Declared provider mutation: DELETE /api/v1/employees/{id}/tables/{table}/{rowId}.; notes: Blocked: locked source operation bamboo-hr.provider.delete-employee-table-row-129 has no declaration-owned executable direct_write route.
    operations post-api-v1-employees-id-tables-table-row-id - Declared direct write: POST /api/v1/employees/{id}/tables/{table}/{rowId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/employees/{id}/tables/{table}/{rowId}.; notes: Blocked: locked source operation bamboo-hr.provider.update-table-row-v1-130 has no declaration-owned executable direct_write route.
    operations get-api-v1-employees-changed - Declared etl: GET /api/v1/employees/changed. [intent=etl availability=implemented stream=changed_employee_ids]; notes: Provider GET /api/v1/employees/changed is bound to the existing changed_employee_ids stream with its connector-owned schema and pagination contract.
    operations get-api-v1-employees-changed-tables-table - Declared etl: GET /api/v1/employees/changed/tables/{table}. [intent=etl availability=implemented stream=changed_employee_table_data]; notes: Provider GET /api/v1/employees/changed/tables/{table} is bound to the existing changed_employee_table_data stream with its connector-owned schema and pagination contract.
    operations get-api-v1-employees-directory - Declared etl: GET /api/v1/employees/directory. [intent=etl availability=implemented stream=employees]; notes: Provider GET /api/v1/employees/directory is bound to the existing employees stream with its connector-owned schema and pagination contract.
    operations post-api-v1-employees-files-categories - Declared direct write: POST /api/v1/employees/files/categories. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: File, photo, CSV/PDF, or multipart payload is outside the JSON record/write dialect.; notes: Blocked: locked source operation bamboo-hr.provider.create-employee-file-category-134 has no declaration-owned executable direct_write route.
    operations post-api-v1-files - Declared direct write: POST /api/v1/files. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: File, photo, CSV/PDF, or multipart payload is outside the JSON record/write dialect.; notes: Blocked: locked source operation bamboo-hr.provider.upload-company-file-135 has no declaration-owned executable direct_write route.
    operations delete-api-v1-files-file-id - Declared direct write: DELETE /api/v1/files/{fileId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: File, photo, CSV/PDF, or multipart payload is outside the JSON record/write dialect.; notes: Blocked: locked source operation bamboo-hr.provider.delete-company-file-136 has no declaration-owned executable direct_write route.
    operations get-api-v1-files-file-id - Declared direct read: GET /api/v1/files/{fileId}. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.get-company-file-137 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations post-api-v1-files-file-id - Declared direct write: POST /api/v1/files/{fileId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: File, photo, CSV/PDF, or multipart payload is outside the JSON record/write dialect.; notes: Blocked: locked source operation bamboo-hr.provider.update-company-file-138 has no declaration-owned executable direct_write route.
    operations post-api-v1-files-categories - Declared direct write: POST /api/v1/files/categories. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: File, photo, CSV/PDF, or multipart payload is outside the JSON record/write dialect.; notes: Blocked: locked source operation bamboo-hr.provider.create-company-file-category-139 has no declaration-owned executable direct_write route.
    operations get-api-v1-files-view - Declared direct read: GET /api/v1/files/view. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.list-company-files-140 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations get-api-v1-holidays - Declared direct read: GET /api/v1/holidays. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.list-company-holidays-141 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations post-api-v1-holidays - Declared direct write: POST /api/v1/holidays. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider operation has no connector-owned operation contract, command, or CLI surface. It remains declaration-pending rather than being treated as an engine gap.; notes: Blocked: locked source operation bamboo-hr.provider.create-company-holiday-142 has no declaration-owned executable direct_write route.
    operations delete-api-v1-holidays-id - Declared direct write: DELETE /api/v1/holidays/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: Documented provider operation has no connector-owned operation contract, command, or CLI surface. It remains declaration-pending rather than being treated as an engine gap.; notes: Blocked: locked source operation bamboo-hr.provider.delete-company-holiday-143 has no declaration-owned executable direct_write route.
    operations get-api-v1-holidays-id - Declared direct read: GET /api/v1/holidays/{id}. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.get-company-holiday-144 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations patch-api-v1-holidays-id - Declared direct write: PATCH /api/v1/holidays/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider operation has no connector-owned operation contract, command, or CLI surface. It remains declaration-pending rather than being treated as an engine gap.; notes: Blocked: locked source operation bamboo-hr.provider.update-company-holiday-145 has no declaration-owned executable direct_write route.
    operations post-api-v1-holidays-bulk-insert - Declared direct write: POST /api/v1/holidays/bulk-insert. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider operation has no connector-owned operation contract, command, or CLI surface. It remains declaration-pending rather than being treated as an engine gap.; notes: Blocked: locked source operation bamboo-hr.provider.bulk-insert-company-holidays-146 has no declaration-owned executable direct_write route.
    operations get-api-v1-holidays-catalog - Declared direct read: GET /api/v1/holidays/catalog. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.list-catalog-holidays-147 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations get-api-v1-holidays-catalog-uuid - Declared direct read: GET /api/v1/holidays/catalog/{uuid}. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.get-catalog-holiday-148 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations get-api-v1-hris-org-locations - Declared direct read: GET /api/v1/hris/org/locations. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.get-locations-149 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations post-api-v1-hris-org-locations - Declared direct write: POST /api/v1/hris/org/locations. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.create-location-150 has no declaration-owned executable direct_write route.
    operations delete-api-v1-hris-org-locations-id - Declared direct write: DELETE /api/v1/hris/org/locations/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.delete-location-151 has no declaration-owned executable direct_write route.
    operations get-api-v1-hris-org-locations-id - Declared direct read: GET /api/v1/hris/org/locations/{id}. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.get-location-152 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations put-api-v1-hris-org-locations-id - Declared direct write: PUT /api/v1/hris/org/locations/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.update-location-153 has no declaration-owned executable direct_write route.
    operations post-api-v1-login - Declared direct write: POST /api/v1/login. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.login-154 has no declaration-owned executable direct_write route.
    operations get-api-v1-meta-bank-holidays - Declared direct read: GET /api/v1/meta/bank-holidays. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.list-bank-holidays-155 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations get-api-v1-meta-company - Declared direct read: GET /api/v1/meta/company. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.get-meta-company-156 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations get-api-v1-meta-countries-id - Declared direct read: GET /api/v1/meta/countries/{id}. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.get-country-by-id-157 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations get-api-v1-meta-countries-options - Declared direct read: GET /api/v1/meta/countries/options. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.get-countries-options-158 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations get-api-v1-meta-currency-conversions - Declared direct read: GET /api/v1/meta/currency-conversions. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.get-currency-conversions-159 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations get-api-v1-meta-currency-types - Declared etl: GET /api/v1/meta/currency/types. [intent=etl availability=implemented stream=all_currency_types]; notes: Provider GET /api/v1/meta/currency/types is bound to the existing all_currency_types stream with its connector-owned schema and pagination contract.
    operations get-api-v1-meta-fields - Declared etl: GET /api/v1/meta/fields. [intent=etl availability=implemented stream=meta_fields]; notes: Provider GET /api/v1/meta/fields is bound to the existing meta_fields stream with its connector-owned schema and pagination contract.
    operations get-api-v1-meta-industries - Declared direct read: GET /api/v1/meta/industries. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.list-industries-162 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations get-api-v1-meta-lists - Declared etl: GET /api/v1/meta/lists. [intent=etl availability=implemented stream=meta_lists]; notes: Provider GET /api/v1/meta/lists is bound to the existing meta_lists stream with its connector-owned schema and pagination contract.
    operations put-api-v1-meta-lists-list-field-id - Declared direct write: PUT /api/v1/meta/lists/{listFieldId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: PUT /api/v1/meta/lists/{listFieldId}.; notes: Blocked: locked source operation bamboo-hr.provider.update-list-field-values-164 has no declaration-owned executable direct_write route.
    operations get-api-v1-meta-provinces - Declared direct read: GET /api/v1/meta/provinces. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.get-all-provinces-165 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations get-api-v1-meta-provinces-country-id - Declared etl: GET /api/v1/meta/provinces/{countryId}. [intent=etl availability=implemented stream=states_by_country_id]; notes: Provider GET /api/v1/meta/provinces/{countryId} is bound to the existing states_by_country_id stream with its connector-owned schema and pagination contract.
    operations get-api-v1-meta-tables - Declared etl: GET /api/v1/meta/tables. [intent=etl availability=implemented stream=tabular_fields]; notes: Provider GET /api/v1/meta/tables is bound to the existing tabular_fields stream with its connector-owned schema and pagination contract.
    operations get-api-v1-meta-time-off-policies - Declared etl: GET /api/v1/meta/time_off/policies. [intent=etl availability=implemented stream=time_off_policies]; notes: Provider GET /api/v1/meta/time_off/policies is bound to the existing time_off_policies stream with its connector-owned schema and pagination contract.
    operations get-api-v1-meta-time-off-types - Declared etl: GET /api/v1/meta/time_off/types. [intent=etl availability=implemented stream=time_off_types]; notes: Provider GET /api/v1/meta/time_off/types is bound to the existing time_off_types stream with its connector-owned schema and pagination contract.
    operations get-api-v1-meta-timezones - Declared direct read: GET /api/v1/meta/timezones. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.5c5fb0f1211ae1c9451753f92f1053b6-170 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations get-api-v1-meta-timezones-id - Declared direct read: GET /api/v1/meta/timezones/{id}. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.baa7162824294d030115568d1d8e6ca7-171 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations get-api-v1-meta-timezones-by-zip-zip - Declared direct read: GET /api/v1/meta/timezones/by-zip/{zip}. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.10d66d8561dd7dac50ff9c21ef63d83b-172 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations get-api-v1-meta-users - Declared etl: GET /api/v1/meta/users. [intent=etl availability=implemented stream=users]; notes: Provider GET /api/v1/meta/users is bound to the existing users stream with its connector-owned schema and pagination contract.
    operations get-api-v1-new-hire-packets - Declared direct read: GET /api/v1/new-hire-packets. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.f44b802c30cdea2b9076b3f82f99c74d-174 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations post-api-v1-new-hire-packets - Declared direct write: POST /api/v1/new-hire-packets. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.ec1ba8e76f33960b018d0d7518fe97b5-175 has no declaration-owned executable direct_write route.
    operations delete-api-v1-new-hire-packets-id - Declared direct write: DELETE /api/v1/new-hire-packets/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.caa7fc488bcfaef14125398f2ebb987d-176 has no declaration-owned executable direct_write route.
    operations get-api-v1-new-hire-packets-id - Declared direct read: GET /api/v1/new-hire-packets/{id}. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.696f0a229cdde60b733568e3c4d043d9-177 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations put-api-v1-new-hire-packets-id - Declared direct write: PUT /api/v1/new-hire-packets/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.1ab0279d46023eb951a434f24df885f1-178 has no declaration-owned executable direct_write route.
    operations post-api-v1-new-hire-packets-id-cancel - Declared direct write: POST /api/v1/new-hire-packets/{id}/cancel. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.19c7e26a1347ae7eb22919e9b0595c19-179 has no declaration-owned executable direct_write route.
    operations put-api-v1-new-hire-packets-id-question-visibility - Declared direct write: PUT /api/v1/new-hire-packets/{id}/question-visibility. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.update-new-hire-packet-gtky-answer-visibility-180 has no declaration-owned executable direct_write route.
    operations post-api-v1-new-hire-packets-id-send - Declared direct write: POST /api/v1/new-hire-packets/{id}/send. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.f49b0f1f2fb1ef2c408ba12916ee9baa-181 has no declaration-owned executable direct_write route.
    operations get-api-v1-onboarding-new-hire-widget - Declared direct read: GET /api/v1/onboarding/new-hire-widget. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.044949386f2d655c6a627ef53f9434b7-182 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations get-api-v1-pay-grades-and-bands - Declared direct read: GET /api/v1/pay-grades-and-bands. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.get-published-levels-and-bands-183 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations post-api-v1-pay-grades-and-bands-import - Declared direct write: POST /api/v1/pay-grades-and-bands/import. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.upload-levels-and-bands-csv-184 has no declaration-owned executable direct_write route.
    operations get-api-v1-pay-grades-and-bands-job-titles - Declared direct read: GET /api/v1/pay-grades-and-bands/job-titles. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.get-job-title-level-assignments-185 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations put-api-v1-pay-grades-and-bands-job-titles - Declared direct write: PUT /api/v1/pay-grades-and-bands/job-titles. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.replace-job-title-level-assignments-186 has no declaration-owned executable direct_write route.
    operations get-api-v1-pay-grades-and-bands-job-titles-with-employees - Declared direct read: GET /api/v1/pay-grades-and-bands/job-titles-with-employees. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.list-job-titles-with-employees-187 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations get-api-v1-pay-grades-and-bands-levels - Declared direct read: GET /api/v1/pay-grades-and-bands/levels. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.list-compensation-level-groups-and-levels-188 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations put-api-v1-pay-grades-and-bands-levels - Declared direct write: PUT /api/v1/pay-grades-and-bands/levels. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.update-compensation-level-groups-and-levels-189 has no declaration-owned executable direct_write route.
    operations delete-api-v1-pay-grades-and-bands-levels-segment - Declared direct write: DELETE /api/v1/pay-grades-and-bands/levels/{segment}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.delete-compensation-level-groups-or-level-190 has no declaration-owned executable direct_write route.
    operations get-api-v1-pay-grades-and-bands-pay-bands - Declared direct read: GET /api/v1/pay-grades-and-bands/pay-bands. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.get-pay-bands-191 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations put-api-v1-pay-grades-and-bands-pay-bands - Declared direct write: PUT /api/v1/pay-grades-and-bands/pay-bands. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.update-pay-bands-192 has no declaration-owned executable direct_write route.
    operations post-api-v1-pay-grades-and-bands-publish - Declared direct write: POST /api/v1/pay-grades-and-bands/publish. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OAuth-scoped public API operation; this bundle intentionally preserves the legacy BambooHR API-key Basic-auth credential model.; notes: Blocked: locked source operation bamboo-hr.provider.publish-draft-compensation-level-groups-193 has no declaration-owned executable direct_write route.
    operations get-api-v1-pay-grades-and-bands-review - Declared direct read: GET /api/v1/pay-grades-and-bands/review. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.get-levels-and-bands-review-194 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations get-api-v1-pay-grades-and-bands-status - Declared direct read: GET /api/v1/pay-grades-and-bands/status. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.get-levels-and-bands-status-195 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations get-api-v1-pay-grades-and-bands-status-counts - Declared direct read: GET /api/v1/pay-grades-and-bands/status-counts. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.get-compensation-level-group-status-counts-196 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations get-api-v1-performance-employees-employee-id-goals - Declared etl: GET /api/v1/performance/employees/{employeeId}/goals. [intent=etl availability=implemented stream=goals]; notes: Provider GET /api/v1/performance/employees/{employeeId}/goals is bound to the existing goals stream with its connector-owned schema and pagination contract.
    operations post-api-v1-performance-employees-employee-id-goals - Declared direct write: POST /api/v1/performance/employees/{employeeId}/goals. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/performance/employees/{employeeId}/goals.; notes: Blocked: locked source operation bamboo-hr.provider.create-goal-198 has no declaration-owned executable direct_write route.
    operations delete-api-v1-performance-employees-employee-id-goals-goal-id - Declared direct write: DELETE /api/v1/performance/employees/{employeeId}/goals/{goalId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: Declared provider mutation: DELETE /api/v1/performance/employees/{employeeId}/goals/{goalId}.; notes: Blocked: locked source operation bamboo-hr.provider.delete-goal-199 has no declaration-owned executable direct_write route.
    operations put-api-v1-performance-employees-employee-id-goals-goal-id - Declared direct write: PUT /api/v1/performance/employees/{employeeId}/goals/{goalId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: PUT /api/v1/performance/employees/{employeeId}/goals/{goalId}.; notes: Blocked: locked source operation bamboo-hr.provider.update-goal-v1-200 has no declaration-owned executable direct_write route.
    operations get-api-v1-performance-employees-employee-id-goals-goal-id-aggregate - Declared etl: GET /api/v1/performance/employees/{employeeId}/goals/{goalId}/aggregate. [intent=etl availability=implemented stream=goal_aggregate]; notes: Provider GET /api/v1/performance/employees/{employeeId}/goals/{goalId}/aggregate is bound to the existing goal_aggregate stream with its connector-owned schema and pagination contract.
    operations post-api-v1-performance-employees-employee-id-goals-goal-id-close - Declared direct write: POST /api/v1/performance/employees/{employeeId}/goals/{goalId}/close. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/performance/employees/{employeeId}/goals/{goalId}/close.; notes: Blocked: locked source operation bamboo-hr.provider.close-goal-202 has no declaration-owned executable direct_write route.
    operations get-api-v1-performance-employees-employee-id-goals-goal-id-comments - Declared etl: GET /api/v1/performance/employees/{employeeId}/goals/{goalId}/comments. [intent=etl availability=implemented stream=goal_comments]; notes: Provider GET /api/v1/performance/employees/{employeeId}/goals/{goalId}/comments is bound to the existing goal_comments stream with its connector-owned schema and pagination contract.
    operations post-api-v1-performance-employees-employee-id-goals-goal-id-comments - Declared direct write: POST /api/v1/performance/employees/{employeeId}/goals/{goalId}/comments. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/performance/employees/{employeeId}/goals/{goalId}/comments.; notes: Blocked: locked source operation bamboo-hr.provider.create-goal-comment-204 has no declaration-owned executable direct_write route.
    operations delete-api-v1-performance-employees-employee-id-goals-goal-id-comments-comment-id - Declared direct write: DELETE /api/v1/performance/employees/{employeeId}/goals/{goalId}/comments/{commentId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: Declared provider mutation: DELETE /api/v1/performance/employees/{employeeId}/goals/{goalId}/comments/{commentId}.; notes: Blocked: locked source operation bamboo-hr.provider.delete-goal-comment-205 has no declaration-owned executable direct_write route.
    operations put-api-v1-performance-employees-employee-id-goals-goal-id-comments-comment-id - Declared direct write: PUT /api/v1/performance/employees/{employeeId}/goals/{goalId}/comments/{commentId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: PUT /api/v1/performance/employees/{employeeId}/goals/{goalId}/comments/{commentId}.; notes: Blocked: locked source operation bamboo-hr.provider.update-goal-comment-206 has no declaration-owned executable direct_write route.
    operations put-api-v1-performance-employees-employee-id-goals-goal-id-milestones-milestone-id-progress - Declared direct write: PUT /api/v1/performance/employees/{employeeId}/goals/{goalId}/milestones/{milestoneId}/progress. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: PUT /api/v1/performance/employees/{employeeId}/goals/{goalId}/milestones/{milestoneId}/progress.; notes: Blocked: locked source operation bamboo-hr.provider.update-goal-milestone-progress-207 has no declaration-owned executable direct_write route.
    operations put-api-v1-performance-employees-employee-id-goals-goal-id-progress - Declared direct write: PUT /api/v1/performance/employees/{employeeId}/goals/{goalId}/progress. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: PUT /api/v1/performance/employees/{employeeId}/goals/{goalId}/progress.; notes: Blocked: locked source operation bamboo-hr.provider.update-goal-progress-208 has no declaration-owned executable direct_write route.
    operations post-api-v1-performance-employees-employee-id-goals-goal-id-reopen - Declared direct write: POST /api/v1/performance/employees/{employeeId}/goals/{goalId}/reopen. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/performance/employees/{employeeId}/goals/{goalId}/reopen.; notes: Blocked: locked source operation bamboo-hr.provider.reopen-goal-209 has no declaration-owned executable direct_write route.
    operations put-api-v1-performance-employees-employee-id-goals-goal-id-shared-with - Declared direct write: PUT /api/v1/performance/employees/{employeeId}/goals/{goalId}/sharedWith. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: PUT /api/v1/performance/employees/{employeeId}/goals/{goalId}/sharedWith.; notes: Blocked: locked source operation bamboo-hr.provider.update-goal-sharing-210 has no declaration-owned executable direct_write route.
    operations get-api-v1-performance-employees-employee-id-goals-aggregate - Declared etl: GET /api/v1/performance/employees/{employeeId}/goals/aggregate. [intent=etl availability=implemented stream=goals_aggregate_v1]; notes: Provider GET /api/v1/performance/employees/{employeeId}/goals/aggregate is bound to the existing goals_aggregate_v1 stream with its connector-owned schema and pagination contract.
    operations get-api-v1-performance-employees-employee-id-goals-alignment-options - Declared etl: GET /api/v1/performance/employees/{employeeId}/goals/alignmentOptions. [intent=etl availability=implemented stream=alignable_goal_options]; notes: Provider GET /api/v1/performance/employees/{employeeId}/goals/alignmentOptions is bound to the existing alignable_goal_options stream with its connector-owned schema and pagination contract.
    operations get-api-v1-performance-employees-employee-id-goals-can-create-goals - Declared etl: GET /api/v1/performance/employees/{employeeId}/goals/canCreateGoals. [intent=etl availability=implemented stream=goal_creation_permission]; notes: Provider GET /api/v1/performance/employees/{employeeId}/goals/canCreateGoals is bound to the existing goal_creation_permission stream with its connector-owned schema and pagination contract.
    operations get-api-v1-performance-employees-employee-id-goals-filters - Declared etl: GET /api/v1/performance/employees/{employeeId}/goals/filters. [intent=etl availability=implemented stream=goals_filters_v1]; notes: Provider GET /api/v1/performance/employees/{employeeId}/goals/filters is bound to the existing goals_filters_v1 stream with its connector-owned schema and pagination contract.
    operations get-api-v1-performance-employees-employee-id-goals-share-options - Declared etl: GET /api/v1/performance/employees/{employeeId}/goals/shareOptions. [intent=etl availability=implemented stream=goal_share_options]; notes: Provider GET /api/v1/performance/employees/{employeeId}/goals/shareOptions is bound to the existing goal_share_options stream with its connector-owned schema and pagination contract.
    operations get-api-v1-reports-id - Declared etl: GET /api/v1/reports/{id}. [intent=etl availability=implemented stream=company_report]; notes: Provider GET /api/v1/reports/{id} is bound to the existing company_report stream with its connector-owned schema and pagination contract.
    operations post-api-v1-reports-custom - Declared direct write: POST /api/v1/reports/custom. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Deprecated ad-hoc report endpoint; BambooHR directs callers to dataset APIs instead.; notes: Blocked: locked source operation bamboo-hr.provider.request-custom-report-217 has no declaration-owned executable direct_write route.
    operations get-api-v1-scheduling-schedules - Declared etl: GET /api/v1/scheduling/schedules. [intent=etl availability=implemented stream=scheduling_list_schedules]; notes: Provider GET /api/v1/scheduling/schedules is bound to the existing scheduling_list_schedules stream with its connector-owned schema and pagination contract.
    operations post-api-v1-scheduling-schedules - Declared direct write: POST /api/v1/scheduling/schedules. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/scheduling/schedules.; notes: Blocked: locked source operation bamboo-hr.provider.scheduling-create-schedule-219 has no declaration-owned executable direct_write route.
    operations delete-api-v1-scheduling-schedules-id - Declared direct write: DELETE /api/v1/scheduling/schedules/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: Declared provider mutation: DELETE /api/v1/scheduling/schedules/{id}.; notes: Blocked: locked source operation bamboo-hr.provider.scheduling-delete-schedule-220 has no declaration-owned executable direct_write route.
    operations get-api-v1-scheduling-schedules-id - Declared etl: GET /api/v1/scheduling/schedules/{id}. [intent=etl availability=implemented stream=scheduling_get_schedule]; notes: Provider GET /api/v1/scheduling/schedules/{id} is bound to the existing scheduling_get_schedule stream with its connector-owned schema and pagination contract.
    operations patch-api-v1-scheduling-schedules-id - Declared direct write: PATCH /api/v1/scheduling/schedules/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: PATCH /api/v1/scheduling/schedules/{id}.; notes: Blocked: locked source operation bamboo-hr.provider.scheduling-update-schedule-222 has no declaration-owned executable direct_write route.
    operations get-api-v1-scheduling-schedules-id-pdf - Declared direct read: GET /api/v1/scheduling/schedules/{id}/pdf. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.scheduling-get-schedule-pdf-223 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations get-api-v1-scheduling-shift-assessments - Declared etl: GET /api/v1/scheduling/shift-assessments. [intent=etl availability=implemented stream=scheduling_list_shift_assessments]; notes: Provider GET /api/v1/scheduling/shift-assessments is bound to the existing scheduling_list_shift_assessments stream with its connector-owned schema and pagination contract.
    operations get-api-v1-scheduling-shifts - Declared etl: GET /api/v1/scheduling/shifts. [intent=etl availability=implemented stream=scheduling_list_shifts]; notes: Provider GET /api/v1/scheduling/shifts is bound to the existing scheduling_list_shifts stream with its connector-owned schema and pagination contract.
    operations post-api-v1-scheduling-shifts - Declared direct write: POST /api/v1/scheduling/shifts. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/scheduling/shifts.; notes: Blocked: locked source operation bamboo-hr.provider.scheduling-create-shift-226 has no declaration-owned executable direct_write route.
    operations delete-api-v1-scheduling-shifts-id - Declared direct write: DELETE /api/v1/scheduling/shifts/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: Declared provider mutation: DELETE /api/v1/scheduling/shifts/{id}.; notes: Blocked: locked source operation bamboo-hr.provider.scheduling-delete-shift-227 has no declaration-owned executable direct_write route.
    operations get-api-v1-scheduling-shifts-id - Declared etl: GET /api/v1/scheduling/shifts/{id}. [intent=etl availability=implemented stream=scheduling_get_shift]; notes: Provider GET /api/v1/scheduling/shifts/{id} is bound to the existing scheduling_get_shift stream with its connector-owned schema and pagination contract.
    operations patch-api-v1-scheduling-shifts-id - Declared direct write: PATCH /api/v1/scheduling/shifts/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: PATCH /api/v1/scheduling/shifts/{id}.; notes: Blocked: locked source operation bamboo-hr.provider.scheduling-update-shift-229 has no declaration-owned executable direct_write route.
    operations post-api-v1-scheduling-shifts-publish - Declared direct write: POST /api/v1/scheduling/shifts/publish. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/scheduling/shifts/publish.; notes: Blocked: locked source operation bamboo-hr.provider.scheduling-publish-shifts-230 has no declaration-owned executable direct_write route.
    operations get-api-v1-scheduling-timezones - Declared etl: GET /api/v1/scheduling/timezones. [intent=etl availability=implemented stream=scheduling_list_timezones]; notes: Provider GET /api/v1/scheduling/timezones is bound to the existing scheduling_list_timezones stream with its connector-owned schema and pagination contract.
    operations get-api-v1-time-off-requests - Declared etl: GET /api/v1/time_off/requests. [intent=etl availability=implemented stream=time_off_requests]; notes: Provider GET /api/v1/time_off/requests is bound to the existing time_off_requests stream with its connector-owned schema and pagination contract.
    operations put-api-v1-time-off-requests-request-id-status - Declared direct write: PUT /api/v1/time_off/requests/{requestId}/status. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: PUT /api/v1/time_off/requests/{requestId}/status.; notes: Blocked: locked source operation bamboo-hr.provider.update-time-off-request-status-233 has no declaration-owned executable direct_write route.
    operations get-api-v1-time-off-whos-out - Declared etl: GET /api/v1/time_off/whos_out. [intent=etl availability=implemented stream=whos_out]; notes: Provider GET /api/v1/time_off/whos_out is bound to the existing whos_out stream with its connector-owned schema and pagination contract.
    operations post-api-v1-time-tracking-clock-entries-delete - Declared direct write: POST /api/v1/time_tracking/clock_entries/delete. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/time_tracking/clock_entries/delete.; notes: Blocked: locked source operation bamboo-hr.provider.delete-timesheet-clock-entries-via-post-235 has no declaration-owned executable direct_write route.
    operations post-api-v1-time-tracking-clock-entries-store - Declared direct write: POST /api/v1/time_tracking/clock_entries/store. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/time_tracking/clock_entries/store.; notes: Blocked: locked source operation bamboo-hr.provider.create-or-update-timesheet-clock-entries-236 has no declaration-owned executable direct_write route.
    operations post-api-v1-time-tracking-employees-employee-id-clock-in - Declared direct write: POST /api/v1/time_tracking/employees/{employeeId}/clock_in. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/time_tracking/employees/{employeeId}/clock_in.; notes: Blocked: locked source operation bamboo-hr.provider.create-timesheet-clock-in-entry-237 has no declaration-owned executable direct_write route.
    operations post-api-v1-time-tracking-employees-employee-id-clock-out - Declared direct write: POST /api/v1/time_tracking/employees/{employeeId}/clock_out. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/time_tracking/employees/{employeeId}/clock_out.; notes: Blocked: locked source operation bamboo-hr.provider.create-timesheet-clock-out-entry-238 has no declaration-owned executable direct_write route.
    operations post-api-v1-time-tracking-hour-entries-delete - Declared direct write: POST /api/v1/time_tracking/hour_entries/delete. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/time_tracking/hour_entries/delete.; notes: Blocked: locked source operation bamboo-hr.provider.delete-timesheet-hour-entries-via-post-239 has no declaration-owned executable direct_write route.
    operations post-api-v1-time-tracking-hour-entries-store - Declared direct write: POST /api/v1/time_tracking/hour_entries/store. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/time_tracking/hour_entries/store.; notes: Blocked: locked source operation bamboo-hr.provider.create-or-update-timesheet-hour-entries-240 has no declaration-owned executable direct_write route.
    operations post-api-v1-time-tracking-projects - Declared direct write: POST /api/v1/time_tracking/projects. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/time_tracking/projects.; notes: Blocked: locked source operation bamboo-hr.provider.create-time-tracking-project-legacy-241 has no declaration-owned executable direct_write route.
    operations get-api-v1-time-tracking-timesheet-entries - Declared etl: GET /api/v1/time_tracking/timesheet_entries. [intent=etl availability=implemented stream=timesheet_entries]; notes: Provider GET /api/v1/time_tracking/timesheet_entries is bound to the existing timesheet_entries stream with its connector-owned schema and pagination contract.
    operations get-api-v1-time-tracking-break-assessments - Declared etl: GET /api/v1/time-tracking/break-assessments. [intent=etl availability=implemented stream=break_assessments]; notes: Provider GET /api/v1/time-tracking/break-assessments is bound to the existing break_assessments stream with its connector-owned schema and pagination contract.
    operations get-api-v1-time-tracking-break-policies - Declared etl: GET /api/v1/time-tracking/break-policies. [intent=etl availability=implemented stream=break_policies]; notes: Provider GET /api/v1/time-tracking/break-policies is bound to the existing break_policies stream with its connector-owned schema and pagination contract.
    operations post-api-v1-time-tracking-break-policies - Declared direct write: POST /api/v1/time-tracking/break-policies. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/time-tracking/break-policies.; notes: Blocked: locked source operation bamboo-hr.provider.create-break-policy-245 has no declaration-owned executable direct_write route.
    operations delete-api-v1-time-tracking-break-policies-id - Declared direct write: DELETE /api/v1/time-tracking/break-policies/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: Declared provider mutation: DELETE /api/v1/time-tracking/break-policies/{id}.; notes: Blocked: locked source operation bamboo-hr.provider.delete-break-policy-246 has no declaration-owned executable direct_write route.
    operations get-api-v1-time-tracking-break-policies-id - Declared etl: GET /api/v1/time-tracking/break-policies/{id}. [intent=etl availability=implemented stream=break_policy]; notes: Provider GET /api/v1/time-tracking/break-policies/{id} is bound to the existing break_policy stream with its connector-owned schema and pagination contract.
    operations patch-api-v1-time-tracking-break-policies-id - Declared direct write: PATCH /api/v1/time-tracking/break-policies/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: PATCH /api/v1/time-tracking/break-policies/{id}.; notes: Blocked: locked source operation bamboo-hr.provider.update-break-policy-248 has no declaration-owned executable direct_write route.
    operations post-api-v1-time-tracking-break-policies-id-assign - Declared direct write: POST /api/v1/time-tracking/break-policies/{id}/assign. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/time-tracking/break-policies/{id}/assign.; notes: Blocked: locked source operation bamboo-hr.provider.assign-employees-to-break-policy-249 has no declaration-owned executable direct_write route.
    operations put-api-v1-time-tracking-break-policies-id-assign - Declared direct write: PUT /api/v1/time-tracking/break-policies/{id}/assign. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: PUT /api/v1/time-tracking/break-policies/{id}/assign.; notes: Blocked: locked source operation bamboo-hr.provider.set-break-policy-employees-250 has no declaration-owned executable direct_write route.
    operations get-api-v1-time-tracking-break-policies-id-breaks - Declared etl: GET /api/v1/time-tracking/break-policies/{id}/breaks. [intent=etl availability=implemented stream=break_policy_breaks]; notes: Provider GET /api/v1/time-tracking/break-policies/{id}/breaks is bound to the existing break_policy_breaks stream with its connector-owned schema and pagination contract.
    operations post-api-v1-time-tracking-break-policies-id-breaks - Declared direct write: POST /api/v1/time-tracking/break-policies/{id}/breaks. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/time-tracking/break-policies/{id}/breaks.; notes: Blocked: locked source operation bamboo-hr.provider.create-break-252 has no declaration-owned executable direct_write route.
    operations put-api-v1-time-tracking-break-policies-id-breaks - Declared direct write: PUT /api/v1/time-tracking/break-policies/{id}/breaks. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: PUT /api/v1/time-tracking/break-policies/{id}/breaks.; notes: Blocked: locked source operation bamboo-hr.provider.replace-breaks-for-break-policy-253 has no declaration-owned executable direct_write route.
    operations get-api-v1-time-tracking-break-policies-id-employees - Declared etl: GET /api/v1/time-tracking/break-policies/{id}/employees. [intent=etl availability=implemented stream=break_policy_employees]; notes: Provider GET /api/v1/time-tracking/break-policies/{id}/employees is bound to the existing break_policy_employees stream with its connector-owned schema and pagination contract.
    operations put-api-v1-time-tracking-break-policies-id-sync - Declared direct write: PUT /api/v1/time-tracking/break-policies/{id}/sync. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: PUT /api/v1/time-tracking/break-policies/{id}/sync.; notes: Blocked: locked source operation bamboo-hr.provider.sync-break-policy-255 has no declaration-owned executable direct_write route.
    operations post-api-v1-time-tracking-break-policies-id-unassign - Declared direct write: POST /api/v1/time-tracking/break-policies/{id}/unassign. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/time-tracking/break-policies/{id}/unassign.; notes: Blocked: locked source operation bamboo-hr.provider.unassign-employees-from-break-policy-256 has no declaration-owned executable direct_write route.
    operations post-api-v1-time-tracking-break-policies-suggestions - Declared direct write: POST /api/v1/time-tracking/break-policies/suggestions. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Read-like POST endpoint requires request-body query execution; the current declarative read path does not send stream bodies and this must not be exposed as a write action.; notes: Blocked: locked source operation bamboo-hr.provider.get-break-policy-suggestions-257 has no declaration-owned executable direct_write route.
    operations delete-api-v1-time-tracking-breaks-id - Declared direct write: DELETE /api/v1/time-tracking/breaks/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: Declared provider mutation: DELETE /api/v1/time-tracking/breaks/{id}.; notes: Blocked: locked source operation bamboo-hr.provider.delete-break-258 has no declaration-owned executable direct_write route.
    operations get-api-v1-time-tracking-breaks-id - Declared etl: GET /api/v1/time-tracking/breaks/{id}. [intent=etl availability=implemented stream=break]; notes: Provider GET /api/v1/time-tracking/breaks/{id} is bound to the existing break stream with its connector-owned schema and pagination contract.
    operations patch-api-v1-time-tracking-breaks-id - Declared direct write: PATCH /api/v1/time-tracking/breaks/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: PATCH /api/v1/time-tracking/breaks/{id}.; notes: Blocked: locked source operation bamboo-hr.provider.update-break-260 has no declaration-owned executable direct_write route.
    operations get-api-v1-time-tracking-clock-entries - Declared direct read: GET /api/v1/time-tracking/clock-entries. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.list-clock-entries-261 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations post-api-v1-time-tracking-clock-entries - Declared direct write: POST /api/v1/time-tracking/clock-entries. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider operation has no connector-owned operation contract, command, or CLI surface. It remains declaration-pending rather than being treated as an engine gap.; notes: Blocked: locked source operation bamboo-hr.provider.create-clock-entry-262 has no declaration-owned executable direct_write route.
    operations delete-api-v1-time-tracking-clock-entries-id - Declared direct write: DELETE /api/v1/time-tracking/clock-entries/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: Documented provider operation has no connector-owned operation contract, command, or CLI surface. It remains declaration-pending rather than being treated as an engine gap.; notes: Blocked: locked source operation bamboo-hr.provider.delete-clock-entry-263 has no declaration-owned executable direct_write route.
    operations get-api-v1-time-tracking-clock-entries-id - Declared direct read: GET /api/v1/time-tracking/clock-entries/{id}. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.get-clock-entry-264 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations patch-api-v1-time-tracking-clock-entries-id - Declared direct write: PATCH /api/v1/time-tracking/clock-entries/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider operation has no connector-owned operation contract, command, or CLI surface. It remains declaration-pending rather than being treated as an engine gap.; notes: Blocked: locked source operation bamboo-hr.provider.update-clock-entry-265 has no declaration-owned executable direct_write route.
    operations post-api-v1-time-tracking-clock-ins - Declared direct write: POST /api/v1/time-tracking/clock-ins. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider operation has no connector-owned operation contract, command, or CLI surface. It remains declaration-pending rather than being treated as an engine gap.; notes: Blocked: locked source operation bamboo-hr.provider.clock-in-266 has no declaration-owned executable direct_write route.
    operations post-api-v1-time-tracking-clock-outs - Declared direct write: POST /api/v1/time-tracking/clock-outs. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider operation has no connector-owned operation contract, command, or CLI surface. It remains declaration-pending rather than being treated as an engine gap.; notes: Blocked: locked source operation bamboo-hr.provider.clock-out-267 has no declaration-owned executable direct_write route.
    operations get-api-v1-time-tracking-employees-id-break-availabilities - Declared etl: GET /api/v1/time-tracking/employees/{id}/break-availabilities. [intent=etl availability=implemented stream=employee_break_availabilities]; notes: Provider GET /api/v1/time-tracking/employees/{id}/break-availabilities is bound to the existing employee_break_availabilities stream with its connector-owned schema and pagination contract.
    operations get-api-v1-time-tracking-employees-id-break-policies - Declared etl: GET /api/v1/time-tracking/employees/{id}/break-policies. [intent=etl availability=implemented stream=employee_break_policies]; notes: Provider GET /api/v1/time-tracking/employees/{id}/break-policies is bound to the existing employee_break_policies stream with its connector-owned schema and pagination contract.
    operations get-api-v1-time-tracking-hour-entries - Declared direct read: GET /api/v1/time-tracking/hour-entries. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.list-hour-entries-270 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations post-api-v1-time-tracking-hour-entries - Declared direct write: POST /api/v1/time-tracking/hour-entries. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider operation has no connector-owned operation contract, command, or CLI surface. It remains declaration-pending rather than being treated as an engine gap.; notes: Blocked: locked source operation bamboo-hr.provider.create-hour-entry-271 has no declaration-owned executable direct_write route.
    operations delete-api-v1-time-tracking-hour-entries-id - Declared direct write: DELETE /api/v1/time-tracking/hour-entries/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: Documented provider operation has no connector-owned operation contract, command, or CLI surface. It remains declaration-pending rather than being treated as an engine gap.; notes: Blocked: locked source operation bamboo-hr.provider.delete-hour-entry-272 has no declaration-owned executable direct_write route.
    operations get-api-v1-time-tracking-hour-entries-id - Declared direct read: GET /api/v1/time-tracking/hour-entries/{id}. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.get-hour-entry-273 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations patch-api-v1-time-tracking-hour-entries-id - Declared direct write: PATCH /api/v1/time-tracking/hour-entries/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider operation has no connector-owned operation contract, command, or CLI surface. It remains declaration-pending rather than being treated as an engine gap.; notes: Blocked: locked source operation bamboo-hr.provider.update-hour-entry-274 has no declaration-owned executable direct_write route.
    operations get-api-v1-time-tracking-projects - Declared etl: GET /api/v1/time-tracking/projects. [intent=etl availability=implemented stream=projects]; notes: Provider GET /api/v1/time-tracking/projects is bound to the existing projects stream with its connector-owned schema and pagination contract.
    operations post-api-v1-time-tracking-projects-2 - Declared direct write: POST /api/v1/time-tracking/projects. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/time-tracking/projects.; notes: Blocked: locked source operation bamboo-hr.provider.create-time-tracking-project-276 has no declaration-owned executable direct_write route.
    operations delete-api-v1-time-tracking-projects-id - Declared direct write: DELETE /api/v1/time-tracking/projects/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: Declared provider mutation: DELETE /api/v1/time-tracking/projects/{id}.; notes: Blocked: locked source operation bamboo-hr.provider.delete-project-277 has no declaration-owned executable direct_write route.
    operations patch-api-v1-time-tracking-projects-id - Declared direct write: PATCH /api/v1/time-tracking/projects/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: PATCH /api/v1/time-tracking/projects/{id}.; notes: Blocked: locked source operation bamboo-hr.provider.update-project-278 has no declaration-owned executable direct_write route.
    operations get-api-v1-time-tracking-projects-project-id - Declared etl: GET /api/v1/time-tracking/projects/{projectId}. [intent=etl availability=implemented stream=project]; notes: Provider GET /api/v1/time-tracking/projects/{projectId} is bound to the existing project stream with its connector-owned schema and pagination contract.
    operations get-api-v1-time-tracking-projects-project-id-tasks - Declared etl: GET /api/v1/time-tracking/projects/{projectId}/tasks. [intent=etl availability=implemented stream=project_tasks]; notes: Provider GET /api/v1/time-tracking/projects/{projectId}/tasks is bound to the existing project_tasks stream with its connector-owned schema and pagination contract.
    operations post-api-v1-time-tracking-projects-project-id-tasks - Declared direct write: POST /api/v1/time-tracking/projects/{projectId}/tasks. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/time-tracking/projects/{projectId}/tasks.; notes: Blocked: locked source operation bamboo-hr.provider.create-project-task-281 has no declaration-owned executable direct_write route.
    operations get-api-v1-time-tracking-shift-differentials - Declared etl: GET /api/v1/time-tracking/shift-differentials. [intent=etl availability=implemented stream=shift_differentials]; notes: Provider GET /api/v1/time-tracking/shift-differentials is bound to the existing shift_differentials stream with its connector-owned schema and pagination contract.
    operations post-api-v1-time-tracking-shift-differentials - Declared direct write: POST /api/v1/time-tracking/shift-differentials. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/time-tracking/shift-differentials.; notes: Blocked: locked source operation bamboo-hr.provider.create-shift-differential-283 has no declaration-owned executable direct_write route.
    operations delete-api-v1-time-tracking-shift-differentials-id - Declared direct write: DELETE /api/v1/time-tracking/shift-differentials/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: Declared provider mutation: DELETE /api/v1/time-tracking/shift-differentials/{id}.; notes: Blocked: locked source operation bamboo-hr.provider.delete-shift-differential-284 has no declaration-owned executable direct_write route.
    operations get-api-v1-time-tracking-shift-differentials-id - Declared etl: GET /api/v1/time-tracking/shift-differentials/{id}. [intent=etl availability=implemented stream=shift_differential]; notes: Provider GET /api/v1/time-tracking/shift-differentials/{id} is bound to the existing shift_differential stream with its connector-owned schema and pagination contract.
    operations patch-api-v1-time-tracking-shift-differentials-id - Declared direct write: PATCH /api/v1/time-tracking/shift-differentials/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: PATCH /api/v1/time-tracking/shift-differentials/{id}.; notes: Blocked: locked source operation bamboo-hr.provider.update-shift-differential-286 has no declaration-owned executable direct_write route.
    operations delete-api-v1-time-tracking-tasks-id - Declared direct write: DELETE /api/v1/time-tracking/tasks/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: Declared provider mutation: DELETE /api/v1/time-tracking/tasks/{id}.; notes: Blocked: locked source operation bamboo-hr.provider.delete-task-287 has no declaration-owned executable direct_write route.
    operations get-api-v1-time-tracking-tasks-id - Declared etl: GET /api/v1/time-tracking/tasks/{id}. [intent=etl availability=implemented stream=task]; notes: Provider GET /api/v1/time-tracking/tasks/{id} is bound to the existing task stream with its connector-owned schema and pagination contract.
    operations patch-api-v1-time-tracking-tasks-id - Declared direct write: PATCH /api/v1/time-tracking/tasks/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: PATCH /api/v1/time-tracking/tasks/{id}.; notes: Blocked: locked source operation bamboo-hr.provider.update-task-289 has no declaration-owned executable direct_write route.
    operations post-api-v1-time-tracking-timesheet-approvals - Declared direct write: POST /api/v1/time-tracking/timesheet-approvals. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider operation has no connector-owned operation contract, command, or CLI surface. It remains declaration-pending rather than being treated as an engine gap.; notes: Blocked: locked source operation bamboo-hr.provider.approve-timesheet-290 has no declaration-owned executable direct_write route.
    operations get-api-v1-time-tracking-timesheets - Declared direct read: GET /api/v1/time-tracking/timesheets. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.list-timesheets-291 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations get-api-v1-time-tracking-timesheets-id - Declared direct read: GET /api/v1/time-tracking/timesheets/{id}. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.get-timesheet-292 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations get-api-v1-time-tracking-timesheets-id-summary - Declared direct read: GET /api/v1/time-tracking/timesheets/{id}/summary. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.get-timesheet-summary-293 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations post-api-v1-timetracking-add - Declared direct write: POST /api/v1/timetracking/add. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/timetracking/add.; notes: Blocked: locked source operation bamboo-hr.provider.create-time-tracking-hour-record-294 has no declaration-owned executable direct_write route.
    operations put-api-v1-timetracking-adjust - Declared direct write: PUT /api/v1/timetracking/adjust. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: PUT /api/v1/timetracking/adjust.; notes: Blocked: locked source operation bamboo-hr.provider.update-time-tracking-record-295 has no declaration-owned executable direct_write route.
    operations delete-api-v1-timetracking-delete-id - Declared direct write: DELETE /api/v1/timetracking/delete/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: Declared provider mutation: DELETE /api/v1/timetracking/delete/{id}.; notes: Blocked: locked source operation bamboo-hr.provider.delete-time-tracking-hour-record-296 has no declaration-owned executable direct_write route.
    operations post-api-v1-timetracking-record - Declared direct write: POST /api/v1/timetracking/record. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/timetracking/record.; notes: Blocked: locked source operation bamboo-hr.provider.create-or-update-time-tracking-hour-records-297 has no declaration-owned executable direct_write route.
    operations get-api-v1-timetracking-record-id - Declared etl: GET /api/v1/timetracking/record/{id}. [intent=etl availability=implemented stream=time_tracking_record]; notes: Provider GET /api/v1/timetracking/record/{id} is bound to the existing time_tracking_record stream with its connector-owned schema and pagination contract.
    operations get-api-v1-training-category - Declared etl: GET /api/v1/training/category. [intent=etl availability=implemented stream=training_categories]; notes: Provider GET /api/v1/training/category is bound to the existing training_categories stream with its connector-owned schema and pagination contract.
    operations post-api-v1-training-category - Declared direct write: POST /api/v1/training/category. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/training/category.; notes: Blocked: locked source operation bamboo-hr.provider.create-training-category-300 has no declaration-owned executable direct_write route.
    operations delete-api-v1-training-category-training-category-id - Declared direct write: DELETE /api/v1/training/category/{trainingCategoryId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: Declared provider mutation: DELETE /api/v1/training/category/{trainingCategoryId}.; notes: Blocked: locked source operation bamboo-hr.provider.delete-training-category-301 has no declaration-owned executable direct_write route.
    operations put-api-v1-training-category-training-category-id - Declared direct write: PUT /api/v1/training/category/{trainingCategoryId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: PUT /api/v1/training/category/{trainingCategoryId}.; notes: Blocked: locked source operation bamboo-hr.provider.update-training-category-302 has no declaration-owned executable direct_write route.
    operations delete-api-v1-training-record-employee-training-record-id - Declared direct write: DELETE /api/v1/training/record/{employeeTrainingRecordId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: Declared provider mutation: DELETE /api/v1/training/record/{employeeTrainingRecordId}.; notes: Blocked: locked source operation bamboo-hr.provider.delete-employee-training-record-303 has no declaration-owned executable direct_write route.
    operations put-api-v1-training-record-employee-training-record-id - Declared direct write: PUT /api/v1/training/record/{employeeTrainingRecordId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: PUT /api/v1/training/record/{employeeTrainingRecordId}.; notes: Blocked: locked source operation bamboo-hr.provider.update-employee-training-record-304 has no declaration-owned executable direct_write route.
    operations get-api-v1-training-record-employee-employee-id - Declared direct read: GET /api/v1/training/record/employee/{employeeId}. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.list-employee-trainings-305 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations post-api-v1-training-record-employee-employee-id - Declared direct write: POST /api/v1/training/record/employee/{employeeId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/training/record/employee/{employeeId}.; notes: Blocked: locked source operation bamboo-hr.provider.create-employee-training-record-306 has no declaration-owned executable direct_write route.
    operations get-api-v1-training-type - Declared etl: GET /api/v1/training/type. [intent=etl availability=implemented stream=training_types]; notes: Provider GET /api/v1/training/type is bound to the existing training_types stream with its connector-owned schema and pagination contract.
    operations post-api-v1-training-type - Declared direct write: POST /api/v1/training/type. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/training/type.; notes: Blocked: locked source operation bamboo-hr.provider.create-training-type-308 has no declaration-owned executable direct_write route.
    operations delete-api-v1-training-type-training-type-id - Declared direct write: DELETE /api/v1/training/type/{trainingTypeId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: Declared provider mutation: DELETE /api/v1/training/type/{trainingTypeId}.; notes: Blocked: locked source operation bamboo-hr.provider.delete-training-type-309 has no declaration-owned executable direct_write route.
    operations put-api-v1-training-type-training-type-id - Declared direct write: PUT /api/v1/training/type/{trainingTypeId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: PUT /api/v1/training/type/{trainingTypeId}.; notes: Blocked: locked source operation bamboo-hr.provider.update-training-type-310 has no declaration-owned executable direct_write route.
    operations get-api-v1-webhooks - Declared etl: GET /api/v1/webhooks. [intent=etl availability=implemented stream=webhooks]; notes: Provider GET /api/v1/webhooks is bound to the existing webhooks stream with its connector-owned schema and pagination contract.
    operations post-api-v1-webhooks - Declared direct write: POST /api/v1/webhooks. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/webhooks.; notes: Blocked: locked source operation bamboo-hr.provider.create-webhook-312 has no declaration-owned executable direct_write route.
    operations delete-api-v1-webhooks-id - Declared direct write: DELETE /api/v1/webhooks/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: Declared provider mutation: DELETE /api/v1/webhooks/{id}.; notes: Blocked: locked source operation bamboo-hr.provider.delete-webhook-313 has no declaration-owned executable direct_write route.
    operations get-api-v1-webhooks-id - Declared etl: GET /api/v1/webhooks/{id}. [intent=etl availability=implemented stream=webhook]; notes: Provider GET /api/v1/webhooks/{id} is bound to the existing webhook stream with its connector-owned schema and pagination contract.
    operations put-api-v1-webhooks-id - Declared direct write: PUT /api/v1/webhooks/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: PUT /api/v1/webhooks/{id}.; notes: Blocked: locked source operation bamboo-hr.provider.update-webhook-315 has no declaration-owned executable direct_write route.
    operations get-api-v1-webhooks-id-log - Declared direct read: GET /api/v1/webhooks/{id}/log. [intent=direct_read availability=partial]; notes: Blocked: locked source operation bamboo-hr.provider.list-webhook-logs-316 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations get-api-v1-webhooks-monitor-fields - Declared etl: GET /api/v1/webhooks/monitor_fields. [intent=etl availability=implemented stream=monitor_fields]; notes: Provider GET /api/v1/webhooks/monitor_fields is bound to the existing monitor_fields stream with its connector-owned schema and pagination contract.
    operations get-api-v1-webhooks-post-fields - Declared etl: GET /api/v1/webhooks/post-fields. [intent=etl availability=implemented stream=post_fields]; notes: Provider GET /api/v1/webhooks/post-fields is bound to the existing post_fields stream with its connector-owned schema and pagination contract.
    operations post-api-v2-datasets-dataset-name-data - Declared direct write: POST /api/v2/datasets/{datasetName}/data. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Read-like POST endpoint requires request-body query execution; the current declarative read path does not send stream bodies and this must not be exposed as a write action.; notes: Blocked: locked source operation bamboo-hr.provider.get-data-from-dataset-v2-319 has no declaration-owned executable direct_write route.
    operations post-api-v1-benefit-company-benefit - Declared direct write: POST /api/v1/benefit/company_benefit. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/benefit/company_benefit.; notes: Blocked: locked source operation bamboo-hr.local-api-surface.post-api-v1-benefit-company-benefit-320 has no declaration-owned executable direct_write route.
    operations get-api-v1-benefit-company-benefit-type - Declared etl: GET /api/v1/benefit/company_benefit/type. [intent=etl availability=implemented stream=company_benefit_types]; notes: Provider GET /api/v1/benefit/company_benefit/type is bound to the existing company_benefit_types stream with its connector-owned schema and pagination contract.
    operations delete-api-v1-benefit-company-benefit-id - Declared direct write: DELETE /api/v1/benefit/company_benefit/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: Declared provider mutation: DELETE /api/v1/benefit/company_benefit/{id}.; notes: Blocked: locked source operation bamboo-hr.local-api-surface.delete-api-v1-benefit-company-benefit-id-322 has no declaration-owned executable direct_write route.
    operations get-api-v1-benefit-company-benefit-id - Declared etl: GET /api/v1/benefit/company_benefit/{id}. [intent=etl availability=implemented stream=company_benefit]; notes: Provider GET /api/v1/benefit/company_benefit/{id} is bound to the existing company_benefit stream with its connector-owned schema and pagination contract.
    operations put-api-v1-benefit-company-benefit-id - Declared direct write: PUT /api/v1/benefit/company_benefit/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: PUT /api/v1/benefit/company_benefit/{id}.; notes: Blocked: locked source operation bamboo-hr.local-api-surface.put-api-v1-benefit-company-benefit-id-324 has no declaration-owned executable direct_write route.
    operations post-api-v1-benefit-employee-benefit - Declared direct write: POST /api/v1/benefit/employee_benefit. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/benefit/employee_benefit.; notes: Blocked: locked source operation bamboo-hr.local-api-surface.post-api-v1-benefit-employee-benefit-325 has no declaration-owned executable direct_write route.
    operations post-api-v1-benefitgroupemployees - Declared direct write: POST /api/v1/benefitgroupemployees. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/benefitgroupemployees.; notes: Blocked: locked source operation bamboo-hr.local-api-surface.post-api-v1-benefitgroupemployees-326 has no declaration-owned executable direct_write route.
    operations get-api-v1-company-eins - Declared etl: GET /api/v1/company_eins. [intent=etl availability=implemented stream=company_eins]; notes: Provider GET /api/v1/company_eins is bound to the existing company_eins stream with its connector-owned schema and pagination contract.
    operations delete-api-v1-employee-direct-deposit-accounts-id - Declared direct write: DELETE /api/v1/employee_direct_deposit_accounts/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: Declared provider mutation: DELETE /api/v1/employee_direct_deposit_accounts/{id}.; notes: Blocked: locked source operation bamboo-hr.local-api-surface.delete-api-v1-employee-direct-deposit-accounts-id-328 has no declaration-owned executable direct_write route.
    operations post-api-v1-employee-direct-deposit-accounts-id - Declared direct write: POST /api/v1/employee_direct_deposit_accounts/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/employee_direct_deposit_accounts/{id}.; notes: Blocked: locked source operation bamboo-hr.local-api-surface.post-api-v1-employee-direct-deposit-accounts-id-329 has no declaration-owned executable direct_write route.
    operations post-api-v1-employee-pay-stub - Declared direct write: POST /api/v1/employee_pay_stub. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/employee_pay_stub.; notes: Blocked: locked source operation bamboo-hr.local-api-surface.post-api-v1-employee-pay-stub-330 has no declaration-owned executable direct_write route.
    operations delete-api-v1-employee-pay-stub-id - Declared direct write: DELETE /api/v1/employee_pay_stub/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: Declared provider mutation: DELETE /api/v1/employee_pay_stub/{id}.; notes: Blocked: locked source operation bamboo-hr.local-api-surface.delete-api-v1-employee-pay-stub-id-331 has no declaration-owned executable direct_write route.
    operations post-api-v1-employee-unpaid-pay-stubs - Declared direct write: POST /api/v1/employee_unpaid_pay_stubs. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/employee_unpaid_pay_stubs.; notes: Blocked: locked source operation bamboo-hr.local-api-surface.post-api-v1-employee-unpaid-pay-stubs-332 has no declaration-owned executable direct_write route.
    operations delete-api-v1-employee-unpaid-pay-stubs-id - Declared direct write: DELETE /api/v1/employee_unpaid_pay_stubs/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: Declared provider mutation: DELETE /api/v1/employee_unpaid_pay_stubs/{id}.; notes: Blocked: locked source operation bamboo-hr.local-api-surface.delete-api-v1-employee-unpaid-pay-stubs-id-333 has no declaration-owned executable direct_write route.
    operations delete-api-v1-employee-withholding-id - Declared direct write: DELETE /api/v1/employee_withholding/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: Declared provider mutation: DELETE /api/v1/employee_withholding/{id}.; notes: Blocked: locked source operation bamboo-hr.local-api-surface.delete-api-v1-employee-withholding-id-334 has no declaration-owned executable direct_write route.
    operations post-api-v1-employee-withholding-id - Declared direct write: POST /api/v1/employee_withholding/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/employee_withholding/{id}.; notes: Blocked: locked source operation bamboo-hr.local-api-surface.post-api-v1-employee-withholding-id-335 has no declaration-owned executable direct_write route.
    operations delete-api-v1-time-tracking-clock-entries - Declared direct write: DELETE /api/v1/time_tracking/clock_entries. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: Declared provider mutation: DELETE /api/v1/time_tracking/clock_entries.; notes: Blocked: locked source operation bamboo-hr.local-api-surface.delete-api-v1-time-tracking-clock-entries-336 has no declaration-owned executable direct_write route.
    operations post-api-v1-time-tracking-clock-entries-2 - Declared direct write: POST /api/v1/time_tracking/clock_entries. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/time_tracking/clock_entries.; notes: Blocked: locked source operation bamboo-hr.local-api-surface.post-api-v1-time-tracking-clock-entries-337 has no declaration-owned executable direct_write route.
    operations post-api-v1-time-tracking-clock-in-employee-id - Declared direct write: POST /api/v1/time_tracking/clock_in/{employeeId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/time_tracking/clock_in/{employeeId}.; notes: Blocked: locked source operation bamboo-hr.local-api-surface.post-api-v1-time-tracking-clock-in-employeeid-338 has no declaration-owned executable direct_write route.
    operations post-api-v1-time-tracking-clock-out-employee-id - Declared direct write: POST /api/v1/time_tracking/clock_out/{employeeId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/time_tracking/clock_out/{employeeId}.; notes: Blocked: locked source operation bamboo-hr.local-api-surface.post-api-v1-time-tracking-clock-out-employeeid-339 has no declaration-owned executable direct_write route.
    operations post-api-v1-time-tracking-daily-entries - Declared direct write: POST /api/v1/time_tracking/daily_entries. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/time_tracking/daily_entries.; notes: Blocked: locked source operation bamboo-hr.local-api-surface.post-api-v1-time-tracking-daily-entries-340 has no declaration-owned executable direct_write route.
    operations post-api-v1-time-tracking-employee-employee-id-clock-in-data - Declared direct write: POST /api/v1/time_tracking/employee/{employeeId}/clock_in/data. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/time_tracking/employee/{employeeId}/clock_in/data.; notes: Blocked: locked source operation bamboo-hr.local-api-surface.post-api-v1-time-tracking-employee-employeeid-clock-in-data-341 has no declaration-owned executable direct_write route.
    operations post-api-v1-time-tracking-employee-employee-id-clock-out-datetime - Declared direct write: POST /api/v1/time_tracking/employee/{employeeId}/clock_out/datetime. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/time_tracking/employee/{employeeId}/clock_out/datetime.; notes: Blocked: locked source operation bamboo-hr.local-api-surface.post-api-v1-time-tracking-employee-employeeid-clock-out-datetime-342 has no declaration-owned executable direct_write route.
    operations post-api-v1-time-tracking-timesheets-approve - Declared direct write: POST /api/v1/time_tracking/timesheets/approve. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/time_tracking/timesheets/approve.; notes: Blocked: locked source operation bamboo-hr.local-api-surface.post-api-v1-time-tracking-timesheets-approve-343 has no declaration-owned executable direct_write route.
    operations post-api-v1-time-tracking-timesheets-clock-out-and-approve - Declared direct write: POST /api/v1/time_tracking/timesheets/clock_out_and_approve. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1/time_tracking/timesheets/clock_out_and_approve.; notes: Blocked: locked source operation bamboo-hr.local-api-surface.post-api-v1-time-tracking-timesheets-clock-out-and-approve-344 has no declaration-owned executable direct_write route.
    operations post-api-v1-2-benefit-company-benefit-id - Declared direct write: POST /api/v1_2/benefit/company_benefit/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/v1_2/benefit/company_benefit/{id}.; notes: Blocked: locked source operation bamboo-hr.local-api-surface.post-api-v1-2-benefit-company-benefit-id-345 has no declaration-owned executable direct_write route.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect bamboo-hr

  # Inspect as structured JSON
  pm connectors inspect bamboo-hr --json

AGENT WORKFLOW
  - Run pm connectors inspect bamboo-hr before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
