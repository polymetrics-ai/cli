# pm connectors inspect bamboo-hr

```text
NAME
  pm connectors inspect bamboo-hr - BambooHR connector manual

SYNOPSIS
  pm connectors inspect bamboo-hr
  pm connectors inspect bamboo-hr --json
  pm credentials add <name> --connector bamboo-hr [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes BambooHR employee, metadata, reporting, time off, applicant tracking, benefits, goals, training, time tracking, scheduling, compensation, and webhook-management resources from the current official OpenAPI, with unsupported binary/file/inbound-webhook operations blocked in the typed ledger.

ICON
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
  check_total_rewards_profile_employee_id
  company_benefit_id
  company_report_id
  country_id
  dataset_name
  employee_break_availabilities_id
  employee_break_policies_id
  employee_dependent_id
  employee_id
  employee_table_data_id
  get_an_alert_configuration_id
  get_change_communication_letter_details_id
  get_clock_entry_id
  get_compensation_benchmark_details_job_id
  get_compensation_planning_approval_flows_id
  get_compensation_planning_cycle_budgets_id
  get_compensation_planning_cycle_details_id
  get_compensation_planning_cycle_summary_id
  get_compensation_planning_cycle_worksheet_id
  get_country_by_id_id
  get_employee_id
  get_employee_onboarding_experience_by_id_employee_id
  get_employee_onboarding_experience_by_id_onboarding_experience_id
  get_hour_entry_id
  get_location_id
  get_new_hire_packet_by_id_id
  get_timesheet_id
  get_timesheet_summary_id
  get_timezone_by_id_id
  get_timezone_by_zip_code_zip
  get_total_rewards_printable_statement_employee_id
  get_total_rewards_statement_employee_id
  goal_id
  goal_share_options_search
  list_compensation_planning_cycle_admins_id
  list_employee_files_id
  list_employee_onboarding_experiences_employee_id
  list_employee_verifications_by_employee_employee_id
  list_employees_in_compensation_planning_cycle_id
  member_benefits_calendar_year
  project_id
  report_id
  scheduling_get_schedule_id
  scheduling_get_shift_id
  shift_differential_id
  subdomain
  table
  task_id
  time_off_requests_end
  time_off_requests_start
  time_tracking_record_id
  timesheet_entries_end
  timesheet_entries_start
  webhook_id
  access_token (secret)
  api_key (secret)

ETL STREAMS
  list_alert_configurations:
    fields: additionalRecipientEmails(), bambooAlertId(), customMessage(), customSubject(), dueInterval(), dueWithin(), editorUserId(), employeeIds(), filterListValueIds(), groupBy(), id(), includeLocation(), includePosition(), lastEdited(), limitTrainingToRequired(), listValueIds(), runAtTime(), runAtTimeZone(), schedule(), sendToAdmin(), sendToEmployee(), sendToManager(), userIds()
  get_an_alert_configuration:
    fields: additionalRecipientEmails(), bambooAlertId(), customMessage(), customSubject(), dueInterval(), dueWithin(), editorUserId(), employeeIds(), filterListValueIds(), groupBy(), id(), includeLocation(), includePosition(), lastEdited(), limitTrainingToRequired(), listValueIds(), runAtTime(), runAtTimeZone(), schedule(), sendToAdmin(), sendToEmployee(), sendToManager(), userIds()
  list_alert_templates:
    fields: groupName(), id(), name()
  get_company_equity_settings:
    fields: calculationType(), companyValuation(), currencyCode(), disclaimers(), outstandingShares(), pricePerShare(), sliderMax(), sliderMin(), vestingConditions()
  list_compensation_planning_cycles:
  get_compensation_planning_cycle_details:
  get_compensation_planning_cycle_summary:
  list_employees_in_compensation_planning_cycle:
  get_compensation_planning_cycle_budgets:
  get_compensation_planning_approval_flows:
  list_compensation_planning_cycle_admins:
    fields: addedAt(), addedByEmployeeId(), adminType(), displayName(), employeeId(), firstName(), isRemovable(), jobTitle(), lastName(), photoUrl()
  get_change_communication_letter_details:
  get_compensation_planning_cycle_worksheet:
  list_available_compensation_tools:
    fields: success(), tools(), upsell()
  list_employee_verifications_by_employee:
    fields: archived(), billingProcessed(), createdByUserId(), createdYmdt(), eVerifyStatus(), employeeId(), id(), integrationType(), lastModifiedYmdt(), remoteAccessUrl(), startDateYmdt(), verificationStatus(), verificationStatusNotes(), verificationType()
  get_employee_verification_integration:
    fields: integration()
  list_new_hire_packets:
    fields: arriveByTime(), cancelled(), completedDatetime(), contactEmployeeId(), countNhpGtkySent(), createdByUserId(), createdDatetime(), employeeId(), getToKnowYouEmailSent(), id(), includePhotoOption(), location(), nhpConfigurationId(), nhpTemplateId(), requiresPersonalInformation(), requiresPersonalQuestions(), requiresPhoto(), sendGetToKnowYouEmail(), sentDatetime(), showPayrollDirectDeposit(), showPayrollFederal(), showPayrollState(), status(), viewedDatetime()
  get_new_hire_packet_by_id:
    fields: arriveByTime(), cancelled(), completedDatetime(), contactEmployeeId(), countNhpGtkySent(), createdByUserId(), createdDatetime(), employeeId(), getToKnowYouEmailSent(), id(), includePhotoOption(), location(), nhpConfigurationId(), nhpTemplateId(), requiresPersonalInformation(), requiresPersonalQuestions(), requiresPhoto(), sendGetToKnowYouEmail(), sentDatetime(), showPayrollDirectDeposit(), showPayrollFederal(), showPayrollState(), status(), viewedDatetime()
  get_welcome_new_hires_widget:
    fields: canSeeEmployee(), department(), getToKnowYou(), hireDate(), id(), lastName(), location(), preferredFirstName(), profilePictureUrl()
  list_employee_onboarding_experiences:
    fields: employeeId(), id(), newHirePacketId(), nhpConfigurationId(), nhpTemplateId(), status()
  get_employee_onboarding_experience_by_id:
    fields: employeeId(), id(), newHirePacketId(), nhpConfigurationId(), nhpTemplateId(), status()
  scheduling_list_schedules:
    primary key: id
    fields: createdAt(), deletedAt(), earlyClockInThreshold(), employeeIds(), id(), locationId(), managerUserIds(), name(), startOfWeek(), timezone(), updatedAt()
  scheduling_get_schedule:
    primary key: id
    fields: createdAt(), deletedAt(), earlyClockInThreshold(), employeeIds(), id(), locationId(), managerUserIds(), name(), startOfWeek(), timezone(), updatedAt()
  scheduling_list_timezones:
    primary key: id
    fields: id(), name(), offset()
  scheduling_get_shift:
    primary key: id
    fields: capacity(), color(), createdAt(), deletedAt(), employeeIds(), end(), id(), name(), recurrenceDtend(), recurrenceDtstart(), recurrenceId(), recurrenceRule(), recurrenceUntil(), scheduleId(), start(), status(), timezone(), unpublishedChanges(), updatedAt()
  scheduling_list_shifts:
    primary key: id
    fields: capacity(), color(), createdAt(), deletedAt(), employeeIds(), end(), id(), name(), recurrenceDtend(), recurrenceDtstart(), recurrenceId(), recurrenceRule(), recurrenceUntil(), scheduleId(), start(), status(), timezone(), unpublishedChanges(), updatedAt()
  scheduling_list_shift_assessments:
    primary key: id
    fields: createdAt(), date(), employeeId(), id(), result(), shiftId(), updatedAt(), violations()
  break_assessments:
    primary key: id
    fields: _links(), data(), id(), meta()
  break:
    primary key: id
    fields: availabilityEndTime(), availabilityMaxHoursWorked(), availabilityMinHoursWorked(), availabilityStartTime(), availabilityType(), createdAt(), deletedAt(), duration(), id(), name(), paid(), policyId(), updatedAt()
  break_policy_breaks:
    primary key: id
    fields: _links(), data(), id(), meta()
  break_policy:
    primary key: id
    fields: allEmployeesAssigned(), createdAt(), deletedAt(), description(), id(), name(), updatedAt()
  break_policies:
    primary key: id
    fields: _links(), data(), id(), meta()
  employee_break_availabilities:
    primary key: id
    fields: availabilityType(), available(), availableAfterMinutesWorked(), availableAt(), availableIn(), calculatedAt(), duration(), effectiveAt(), id(), name(), paid(), policyId(), recordedDuration(), timezone(), unavailableAt(), unavailableIn()
  employee_break_policies:
    primary key: id
    fields: _links(), data(), id(), meta()
  break_policy_employees:
    primary key: id
    fields: _links(), data(), id(), meta()
  projects:
    primary key: id
    fields: _links(), data(), id(), meta()
  project:
    primary key: id
    fields: allEmployeesAssigned(), archived(), billable(), createdAt(), deletedAt(), employeeIds(), hasTasks(), id(), includeInPayroll(), name(), updatedAt()
  project_tasks:
    primary key: id
    fields: billable(), createdAt(), deletedAt(), id(), name(), projectId(), updatedAt()
  task:
    primary key: id
    fields: billable(), createdAt(), deletedAt(), id(), name(), projectId(), updatedAt()
  shift_differentials:
    primary key: id
    fields: _links(), data(), id(), meta()
  shift_differential:
    primary key: id
    fields: end(), endDay(), id(), start(), startDay()
  timesheet_entries:
    primary key: id
    fields: approved(), approvedAt(), createdAt(), date(), employeeId(), end(), hours(), id(), note(), projectInfo(), start(), timezone(), type(), updatedAt()
  list_clock_entries:
    fields: clockInLocation(), clockOutLocation(), clockedInBy(), clockedOutBy(), createdAt(), date(), employeeId(), end(), endSource(), hours(), id(), note(), projectId(), schedulingShiftId(), start(), startSource(), taskId(), timesheetId(), timezone(), updatedAt()
  get_clock_entry:
    fields: clockInLocation(), clockOutLocation(), clockedInBy(), clockedOutBy(), createdAt(), date(), employeeId(), end(), endSource(), hours(), id(), note(), projectId(), schedulingShiftId(), start(), startSource(), taskId(), timesheetId(), timezone(), updatedAt()
  list_hour_entries:
    fields: createdAt(), date(), employeeId(), hours(), id(), note(), projectId(), taskId(), timesheetId(), updatedAt()
  get_hour_entry:
    fields: createdAt(), date(), employeeId(), hours(), id(), note(), projectId(), taskId(), timesheetId(), updatedAt()
  list_timesheets:
    fields: approvedAt(), approvedBy(), createdAt(), employeeId(), endDate(), hoursLastChangedAt(), id(), overtimeHours(), startDate(), status(), totalHours(), type(), updatedAt()
  get_timesheet:
    fields: approvedAt(), approvedBy(), createdAt(), employeeId(), endDate(), hoursLastChangedAt(), id(), overtimeHours(), startDate(), status(), totalHours(), type(), updatedAt()
  get_timesheet_summary:
    fields: date(), doubleTimeHours(), overtimeHours(), regularHours(), totalHours()
  webhooks:
    primary key: id
    fields: created(), id(), lastSent(), name(), url()
  webhook:
    primary key: id
    fields: duplicatePostString(), error(), id(), monitorFields(), postFields(), unknownFields()
  monitor_fields:
    primary key: id
    fields: alias(), id(), name()
  post_fields:
    primary key: id
    fields: alias(), id(), name(), pageId(), tableId(), type()
  report_by_id:
    primary key: id
    fields: id()
  datasets_v1:
    primary key: id
    fields: id(), label(), name()
  fields_from_dataset_v1:
    primary key: id
    fields: entityName(), id(), label(), name(), parentName(), parentType()
  reports:
    primary key: id
    fields: id(), name()
  get_legacy_report_id_map:
    fields: legacyReportId(), newReportId(), status()
  datasets_v1_2:
    primary key: id
    fields: id(), label(), name()
  fields_from_dataset_v1_2:
    primary key: id
    fields: entityName(), id(), label(), name(), parentName(), parentType()
  applications:
    primary key: id
    fields: applicant(), appliedDate(), id(), job(), rating(), status()
  statuses:
    primary key: id
    fields: code(), description(), enabled(), id(), manageable(), name(), translatedName()
  company_locations:
    primary key: id
    fields: addressLine1(), addressLine2(), city(), country(), description(), id(), name(), phone(), state(), zipcode()
  hiring_leads:
    primary key: employeeId
    fields: employeeId(), preferredFullName()
  company_benefits:
    primary key: id
    fields: allowsCatchUp(), allowsSuperCatchUp(), benefitVendorId(), companyDeductionId(), deductionTypeId(), endDate(), id(), name(), startDate(), type()
  employee_benefits:
    primary key: employeeId
    fields: employeeBenefit(), employeeId(), payFrequency()
  member_benefit_events:
    primary key: memberId
    fields: coverages(), memberId()
  member_benefits:
    primary key: memberId
    fields: memberId(), plans(), subscriberId()
  benefit_deduction_types:
    primary key: id
    fields: additionalDescription(), allowableBenefitTypes(), canBeCollectedByTrax(), deductionNote(), deductionNoteLink(), deductionNoteLinkText(), deductionTypeName(), defaultDeductionCode(), hideAnnualMax(), id(), managedDeductionType(), nonBenefitDeductionType(), subTypeText(), subTypes()
  company_information:
    primary key: id
    fields: address(), displayName(), id(), legalName(), phone()
  company_profile_integrations:
    primary key: id
    fields: id(), integrations()
  list_compensation_benchmarks:
    fields: benchmarks(), employees(), internalJobPayBand(), isRemote(), jobDetails(), locationDetails()
  get_compensation_benchmark_details:
    fields: annualizationFailed(), compaRatio(), compaRatioStatus(), country(), currencyConversionFailed(), id(), isRemote(), jobTitle(), location(), name(), paidPer(), photoUrl(), rangePenetration(), salary(), varianceFromPayBand(), yearsAtCompany()
  list_compensation_benchmark_sources:
    fields: colorCode(), count(), id(), name(), sort()
  employee_roster:
    primary key: employeeId
    fields: _restrictedFields(), addressLine1(), addressLine2(), age(), allergies(), bestEmail(), birthDate(), birthplace(), citizenship(), citizenshipId(), city(), compensationChangeReason(), compensationChangeReasonId(), compensationComment(), compensationEffectiveDate(), compensationEndDate(), contractEndDate(), country(), countryId(), departmentId(), departmentName(), dietaryRestrictions(), displayName(), divisionId(), divisionName(), eeoJobCategory(), eeoJobCategoryId(), ein(), eligibleForRehire(), eligibleForRehireId(), employeeId(), employeeName(), employeeNumber(), employmentStatusComment(), employmentStatusEffectiveDate(), employmentStatusId(), employmentStatusName(), employmentType(), employmentTypeId(), ethnicity(), ethnicityId(), facebookUrl(), finalDoseAdministrationDate(), finalPayDate(), firstName(), firstNameLastName(), firstNameMiddleInitial(), flsaCode(), flsaCodeId(), gender(), genderIdentity(), genderIdentityId(), hireDate(), homeEmail(), homePhone(), hoursPerPayCycle(), instagramUrl(), isManager(), jacketSize(), jacketSizeId(), jobInformationEffectiveDate(), jobTitleId(), jobTitleName(), lastName(), linkedinUrl(), locationId(), locationName(), maritalStatus(), middleInitial(), middleName(), mobilePhone(), nationalId(), nationalInsuranceCategory(), nationalInsuranceCategoryId(), nationality(), nationalityId(), nickName(), nin(), noticePeriod(), noticePeriodId(), originalHireDate(), overtime(), overtimeRate(), paidPer(), payRate(), paySchedule(), payScheduleId(), payType(), photoUrl(), pinterestUrl(), preferredName(), preferredNameLastName(), probationEndDate(), pronouns(), pronounsId(), proofOfVaccination(), reportsToId(), reportsToName(), secondaryLanguage(), shirtSize(), shirtSizeId(), sin(), skypeUsername(), ssn(), state(), stateId(), status(), tShirtSize(), tShirtSizeId(), taxTypeId(), teams(), tenure(), terminationDate(), terminationReason(), terminationReasonId(), terminationRegrettable(), terminationRegrettableId(), terminationType(), terminationTypeId(), twitterUrl(), userId(), vaccinationStatus(), vaccinationStatusId(), vaccineReceived(), vaccineReceivedId(), veteranStatus(), veteranStatusId(), workEmail(), workPhone(), workPhoneExtension(), zipcode()
  get_employee:
    fields: address1(), bestEmail(), birthDate(), canUploadPhoto(), city(), country(), department(), departmentId(), division(), divisionId(), employeeNumber(), employmentStatus(), employmentStatusId(), exempt(), facebookUrl(), firstName(), gender(), hireDate(), homeEmail(), homePhone(), id(), instagramUrl(), jobTitleId(), jobTitleName(), lastName(), linkedinUrl(), location(), locationId(), marital(), middleName(), mobilePhone(), originalHireDate(), payPeriod(), payRate(), payType(), photoUrl(), pinterestUrl(), preferredName(), reportsToId(), reportsToName(), skypeUsername(), state(), status(), terminationDate(), twitterUrl(), workEmail(), workPhone(), workPhoneExtension()
  list_job_titles_with_employees:
    fields: employees(), id(), title()
  get_compensation_level_group_status_counts:
    fields: draft(), historic(), published()
  get_levels_and_bands_status:
    fields: jobTitles(), levels(), payBands(), review()
  list_compensation_level_groups_and_levels:
    fields: errors(), groupId(), groupName(), levels(), warnings()
  get_pay_bands:
    fields: errors(), groupId(), groupName(), levels(), warnings()
  get_job_title_level_assignments:
    fields: errors(), groupId(), groupName(), levels(), warnings()
  get_levels_and_bands_review:
    fields: errors(), groupId(), groupName(), levels(), warnings()
  get_published_levels_and_bands:
    fields: groupId(), groupName(), levels()
  goals_filters_v1:
    primary key: id
    fields: count(), id(), name()
  goals:
    primary key: id
    fields: actions(), alignsWithOptionId(), completionDate(), description(), dueDate(), id(), lastChangedDateTime(), milestones(), percentComplete(), sharedWithEmployeeIds(), status(), title()
  goals_aggregate_v1:
    primary key: id
    fields: actions(), alignsWithOptionId(), completionDate(), description(), dueDate(), id(), lastChangedDateTime(), milestones(), percentComplete(), sharedWithEmployeeIds(), status(), title()
  goal_creation_permission:
    primary key: id
    fields: canCreateGoals(), id()
  goal_share_options:
    primary key: employeeId
    fields: displayFirstName(), employeeId(), lastName(), photoUrl(), userId()
  goal_comments:
    primary key: id
    fields: authorUserId(), canDelete(), canEdit(), createdAt(), id(), text()
  goal_aggregate:
    primary key: id
    fields: authorUserId(), canDelete(), canEdit(), createdAt(), id(), text()
  alignable_goal_options:
    primary key: id
    fields: id(), title()
  goals_filters_v1_1:
    primary key: id
    fields: actions(), count(), id(), name()
  goals_aggregate_v1_1:
    primary key: id
    fields: actions(), alignsWithOptionId(), completionDate(), description(), dueDate(), id(), lastChangedDateTime(), milestones(), percentComplete(), sharedWithEmployeeIds(), status(), title()
  goals_filters_v1_2:
    primary key: id
    fields: actions(), count(), id(), name()
  goals_aggregate_v1_2:
    primary key: id
    fields: actions(), alignsWithOptionId(), completionDate(), description(), dueDate(), id(), lastChangedDateTime(), milestones(), percentComplete(), sharedWithEmployeeIds(), status(), title()
  time_tracking_record:
    primary key: employeeId
    fields: adjustedHours(), dateAdjusted(), dateHoursWorked(), departmentId(), divisionId(), employeeId(), holidayId(), hoursWorked(), jobCode(), jobData(), jobTitleId(), payCode(), payRate(), project(), projectId(), rateType(), shiftDifferential(), shiftDifferentialId(), taskId(), timeTrackingId(), type()
  get_total_rewards_statement:
    fields: benefitSection(), bonusSection(), calendarSection(), commissionSection(), companyLogoUrl(), companyName(), customDisclaimerInfo(), employeeId(), employeeName(), equitySection(), extraPaySection(), hasMixedCurrencyTypes(), jobTitle(), minStatementYear(), overviewSection(), paySection(), statementYear()
  get_all_provinces:
    fields: countryId(), id(), iso(), label(), name()
  states_by_country_id:
    primary key: id
    fields: id(), iso(), label(), name()
  all_currency_types:
    primary key: id
    fields: code(), id(), name(), symbol(), symbolPosition()
  get_meta_company:
    fields: baseApiUrl(), domain(), id(), name()
  list_industries:
    fields: id(), name()
  get_country_by_id:
    fields: id(), isoCode(), name()
  list_timezones:
    fields: gmtOffset(), id(), name(), utcName()
  get_timezone_by_id:
    fields: gmtOffset(), id(), name(), utcName()
  get_timezone_by_zip_code:
    fields: gmtOffset(), id(), name(), utcName()
  list_bank_holidays:
    fields: date(), id(), name()
  get_currency_conversions:
    fields: baseCurrency(), lastUpdated(), nextUpdate(), rates()
  get_locations:
    fields: address(), archived(), archivedAt(), createdAt(), id(), label(), manageable()
  get_location:
    fields: address(), archived(), archivedAt(), createdAt(), id(), label(), manageable()
  check_total_rewards_profile:
    fields: code(), detail(), fields(), instance(), status(), title(), type()
  get_total_rewards_printable_statement:
    fields: code(), detail(), fields(), instance(), status(), title(), type()
  meta_fields:
    primary key: id
    fields: alias(), deprecated(), id(), name(), type()
  users:
    primary key: id
    fields: id()
  changed_employee_table_data:
    primary key: id
    fields: employees(), id(), table()
  employee_table_data:
    primary key: id
    fields: employeeId(), id()
  changed_employee_ids:
    primary key: id
    fields: employees(), id(), latest()
  employee_time_off_policies:
    primary key: timeOffPolicyId
    fields: accrualStartDate(), timeOffPolicyId(), timeOffTypeId()
  employee_time_off_policies_v1_1:
    primary key: timeOffPolicyId
    fields: accrualStartDate(), timeOffPolicyId(), timeOffTypeId()
  application_details:
    primary key: id
    fields: answer(), archivedDate(), editedDate(), editedEndDate(), hasRevisions(), id(), isArchived(), question()
  job_summaries:
    primary key: id
    fields: activeApplicantsCount(), department(), hiringLead(), id(), location(), newApplicantsCount(), postedDate(), postingUrl(), status(), title(), totalApplicantsCount()
  benefit_coverages:
    primary key: id
    fields: benefitPlanId(), description(), id(), shortName(), sortOrder()
  employee_dependent:
    primary key: id
    fields: addressLine1(), addressLine2(), city(), country(), dateOfBirth(), employeeId(), firstName(), gender(), homePhone(), id(), isStudent(), isUsCitizen(), lastName(), maskedSIN(), maskedSSN(), middleName(), relationship(), state(), zipCode()
  employee_dependents:
    primary key: id
    fields: addressLine1(), addressLine2(), city(), country(), dateOfBirth(), employeeId(), firstName(), gender(), homePhone(), id(), isStudent(), isUsCitizen(), lastName(), maskedSIN(), maskedSSN(), middleName(), relationship(), state(), zipCode()
  employees:
    primary key: id
    fields: department(), display_name(), division(), first_name(), id(), job_title(), last_name(), location(), mobile_phone(), photo_url(), preferred_name(), supervisor(), work_email(), work_phone()
  list_company_files:
    fields: canUploadFiles(), files(), id(), name()
  list_employee_files:
    fields: canDeleteCategory(), canRenameCategory(), canUploadFiles(), displayIfEmpty(), files(), id(), name()
  meta_lists:
    primary key: field_id
    fields: alias(), field_id(), manageable(), multiple(), name(), options()
  tabular_fields:
    primary key: id
    fields: alias(), fields(), id()
  company_report:
    primary key: id
    fields: id()
  time_off_balance:
    primary key: id
    fields: balance(), end(), id(), name(), policyType(), timeOffType(), units(), usedYearToDate()
  time_off_policies:
    primary key: id
    fields: effectiveDate(), id(), name(), timeOffTypeId(), type()
  time_off_requests:
    primary key: id
    fields: actions(), amount(), created(), dates(), employeeId(), end(), id(), name(), notes(), start(), status(), type()
  time_off_types:
    primary key: id
    fields: color(), icon(), id(), name(), units()
  training_types:
    primary key: id
    fields: id()
  training_categories:
    primary key: id
    fields: id()
  whos_out:
    primary key: id
    fields: employeeId(), end(), id(), name(), start(), type()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_an_alert_configuration:
    endpoint: POST /api/v1/alert-configurations
    risk: Create an alert configuration; executes a BambooHR mutation against the configured account.
  update_an_alert_configuration:
    endpoint: PUT /api/v1/alert-configurations/{{ record.id }}
    required fields: id
    risk: Update an alert configuration; executes a BambooHR mutation against the configured account.
  update_company_equity_settings:
    endpoint: PUT /api/v1/compensation/equity/settings
    risk: Update company equity settings; executes a BambooHR mutation against the configured account.
  create_compensation_planning_cycle:
    endpoint: POST /api/v1/compensation/planning_cycles
    risk: Create compensation planning cycle; executes a BambooHR mutation against the configured account.
  update_compensation_planning_cycle:
    endpoint: PUT /api/v1/compensation/planning_cycles/{{ record.id }}
    required fields: id
    risk: Update compensation planning cycle; executes a BambooHR mutation against the configured account.
  delete_compensation_planning_cycle:
    endpoint: DELETE /api/v1/compensation/planning_cycles/{{ record.id }}
    required fields: id
    risk: Delete compensation planning cycle; executes a BambooHR mutation against the configured account.
  add_employees_to_cycle:
    endpoint: POST /api/v1/compensation/planning_cycles/{{ record.id }}/employees
    required fields: id, employeeIds
    risk: Add employees to cycle; executes a BambooHR mutation against the configured account.
  remove_employees_from_cycle:
    endpoint: DELETE /api/v1/compensation/planning_cycles/{{ record.id }}/employees
    required fields: id, employeeIds
    risk: Remove employees from cycle; executes a BambooHR mutation against the configured account.
  add_cycle_admins:
    endpoint: POST /api/v1/compensation/planning_cycles/{{ record.id }}/admins
    required fields: id, employeeIds
    risk: Add cycle admins; executes a BambooHR mutation against the configured account.
  launch_compensation_planning_cycle:
    endpoint: PUT /api/v1/compensation/planning_cycles/{{ record.id }}/launch
    required fields: id
    risk: Launch compensation planning cycle; executes a BambooHR mutation against the configured account.
  complete_compensation_planning_cycle:
    endpoint: PUT /api/v1/compensation/planning_cycles/{{ record.id }}/complete
    required fields: id
    risk: Complete compensation planning cycle; executes a BambooHR mutation against the configured account.
  save_budget_guidelines:
    endpoint: PUT /api/v1/compensation/planning_cycles/{{ record.id }}/budgets/guidelines
    required fields: id
    risk: Save budget guidelines; executes a BambooHR mutation against the configured account.
  save_budget_breakdown:
    endpoint: PUT /api/v1/compensation/planning_cycles/{{ record.id }}/budgets/breakdown
    required fields: id
    risk: Save budget breakdown; executes a BambooHR mutation against the configured account.
  import_budget_breakdown:
    endpoint: POST /api/v1/compensation/planning_cycles/{{ record.id }}/budgets/import
    required fields: id, budgetBreakdown
    risk: Import budget breakdown; executes a BambooHR mutation against the configured account.
  update_approval_flow:
    endpoint: PUT /api/v1/compensation/planning_cycles/{{ record.id }}/approvals/{{ record.template_id }}
    required fields: id, template_id, assignees, flowData
    risk: Update approval flow; executes a BambooHR mutation against the configured account.
  set_final_approver:
    endpoint: POST /api/v1/compensation/planning_cycles/{{ record.id }}/approvals/final_approver/{{ record.employee_id }}
    required fields: id, employee_id
    risk: Set final approver; executes a BambooHR mutation against the configured account.
  remove_from_approval_flow:
    endpoint: DELETE /api/v1/compensation/planning_cycles/{{ record.id }}/approvals/employee/{{ record.employee_id }}
    required fields: id, employee_id
    risk: Remove from approval flow; executes a BambooHR mutation against the configured account.
  save_recommendations:
    endpoint: POST /api/v1/compensation/planning_cycles/{{ record.id }}/recommendations
    required fields: id, employeeId, assigneeEmployeeId, approvalStage
    risk: Save recommendations; executes a BambooHR mutation against the configured account.
  send_recommendations_to_next_stage:
    endpoint: POST /api/v1/compensation/planning_cycles/{{ record.id }}/recommendations/send
    required fields: id, templateId, assigneeEmployeeId, approvalStage
    risk: Send recommendations to next stage; executes a BambooHR mutation against the configured account.
  remove_cycle_admin:
    endpoint: DELETE /api/v1/compensation/planning_cycles/{{ record.id }}/admins/{{ record.employee_id }}
    required fields: id, employee_id
    risk: Remove cycle admin; executes a BambooHR mutation against the configured account.
  save_change_comm_template:
    endpoint: PUT /api/v1/compensation/planning_cycles/{{ record.id }}/change_comm/template
    required fields: id, messageText, subjectText
    risk: Save change comm template; executes a BambooHR mutation against the configured account.
  send_employee_verification_lifecycle_email_by_user:
    endpoint: POST /api/v1/employee-verifications/users/{{ record.user_id }}/send-email
    required fields: user_id, emailType
    risk: Send employee verification lifecycle email by user and email type; executes a BambooHR mutation against the configured account.
  update_employee_verification:
    endpoint: PUT /api/v1/employee-verifications/employees/{{ record.employee_id }}/{{ record.verification_id }}
    required fields: employee_id, verification_id
    risk: Update an employee verification record; executes a BambooHR mutation against the configured account.
  update_employee_verification_integration:
    endpoint: PUT /api/v1/employee-verifications/integration
    required fields: enabled
    risk: Enable or disable the employee verification integration; executes a BambooHR mutation against the configured account.
  create_new_hire_packet:
    endpoint: POST /api/v1/new-hire-packets
    required fields: employeeId
    risk: Create new hire packet; executes a BambooHR mutation against the configured account.
  update_new_hire_packet:
    endpoint: PUT /api/v1/new-hire-packets/{{ record.id }}
    required fields: id
    risk: Update new hire packet; executes a BambooHR mutation against the configured account.
  delete_new_hire_packet:
    endpoint: DELETE /api/v1/new-hire-packets/{{ record.id }}
    required fields: id
    risk: Delete new hire packet; executes a BambooHR mutation against the configured account.
  update_new_hire_packet_gtky_answer_visibility:
    endpoint: PUT /api/v1/new-hire-packets/{{ record.id }}/question-visibility
    required fields: id, hidden
    risk: Update GTKY answer visibility for a new hire packet; executes a BambooHR mutation against the configured account.
  send_new_hire_packet:
    endpoint: POST /api/v1/new-hire-packets/{{ record.id }}/send
    required fields: id
    risk: Send new hire packet; executes a BambooHR mutation against the configured account.
  cancel_new_hire_packet:
    endpoint: POST /api/v1/new-hire-packets/{{ record.id }}/cancel
    required fields: id
    risk: Deletes or removes BambooHR data: Cancel new hire packet; executes a BambooHR mutation against the configured account.
  create_employee_onboarding_experience:
    endpoint: POST /api/v1/employees/{{ record.employee_id }}/onboarding-experiences
    required fields: employee_id
    risk: Create employee onboarding experience; executes a BambooHR mutation against the configured account.
  create_time_tracking_project:
    endpoint: POST /api/v1/time_tracking/projects
    required fields: name
    risk: Create Time Tracking Project through the BambooHR API.
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
  delete_scheduling_delete_shift:
    endpoint: DELETE /api/v1/scheduling/shifts/{{ record.id }}
    required fields: id
    risk: Deletes BambooHR data: Delete Shift.
  update_scheduling_update_shift:
    endpoint: PATCH /api/v1/scheduling/shifts/{{ record.id }}
    required fields: id
    risk: Update Shift through the BambooHR API.
  create_scheduling_create_shift:
    endpoint: POST /api/v1/scheduling/shifts
    required fields: scheduleId, status, color, timezone, start, end
    risk: Create Shift through the BambooHR API.
  create_scheduling_publish_shifts:
    endpoint: POST /api/v1/scheduling/shifts/publish
    required fields: shiftIds
    risk: Publish Shifts through the BambooHR API.
  delete_break:
    endpoint: DELETE /api/v1/time-tracking/breaks/{{ record.id }}
    required fields: id
    risk: Deletes BambooHR data: Delete Break.
  update_break:
    endpoint: PATCH /api/v1/time-tracking/breaks/{{ record.id }}
    required fields: id
    risk: Update Break through the BambooHR API.
  replace_breaks_for_break_policy:
    endpoint: PUT /api/v1/time-tracking/break-policies/{{ record.id }}/breaks
    required fields: id
    risk: Replace Breaks for Break Policy through the BambooHR API.
  create_break:
    endpoint: POST /api/v1/time-tracking/break-policies/{{ record.id }}/breaks
    required fields: id
    risk: Create Break through the BambooHR API.
  set_break_policy_employees:
    endpoint: PUT /api/v1/time-tracking/break-policies/{{ record.id }}/assign
    required fields: id, employeeIds
    risk: Set Employees for Break Policy through the BambooHR API.
  assign_employees_to_break_policy:
    endpoint: POST /api/v1/time-tracking/break-policies/{{ record.id }}/assign
    required fields: id, employeeIds
    risk: Assign Employees to Break Policy through the BambooHR API.
  create_unassign_employees_from_break_policy:
    endpoint: POST /api/v1/time-tracking/break-policies/{{ record.id }}/unassign
    required fields: id, employeeIds
    risk: Deletes or removes BambooHR data: Unassign Employees from Break Policy through the BambooHR API.
  delete_break_policy:
    endpoint: DELETE /api/v1/time-tracking/break-policies/{{ record.id }}
    required fields: id
    risk: Deletes BambooHR data: Delete Break Policy.
  update_break_policy:
    endpoint: PATCH /api/v1/time-tracking/break-policies/{{ record.id }}
    required fields: id
    risk: Update Break Policy through the BambooHR API.
  create_break_policy:
    endpoint: POST /api/v1/time-tracking/break-policies
    required fields: name
    risk: Create Break Policy through the BambooHR API.
  sync_break_policy:
    endpoint: PUT /api/v1/time-tracking/break-policies/{{ record.id }}/sync
    required fields: id
    risk: Sync Break Policy through the BambooHR API.
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
  delete_task:
    endpoint: DELETE /api/v1/time-tracking/tasks/{{ record.id }}
    required fields: id
    risk: Deletes BambooHR data: Delete Time Tracking Task.
  update_task:
    endpoint: PATCH /api/v1/time-tracking/tasks/{{ record.id }}
    required fields: id
    risk: Update Time Tracking Task through the BambooHR API.
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
  delete_timesheet_clock_entries_via_post:
    endpoint: POST /api/v1/time_tracking/clock_entries/delete
    required fields: clockEntryIds
    risk: Deletes or removes BambooHR data: Delete Timesheet Clock Entries through the BambooHR API.
  delete_timesheet_hour_entries_via_post:
    endpoint: POST /api/v1/time_tracking/hour_entries/delete
    required fields: hourEntryIds
    risk: Deletes or removes BambooHR data: Delete Timesheet Hour Entries through the BambooHR API.
  create_timesheet_clock_in_entry:
    endpoint: POST /api/v1/time_tracking/employees/{{ record.employee_id }}/clock_in
    required fields: employee_id
    risk: Create Timesheet Clock-In Entry through the BambooHR API.
  create_timesheet_clock_out_entry:
    endpoint: POST /api/v1/time_tracking/employees/{{ record.employee_id }}/clock_out
    required fields: employee_id
    risk: Create Timesheet Clock-Out Entry through the BambooHR API.
  create_or_update_timesheet_clock_entries:
    endpoint: POST /api/v1/time_tracking/clock_entries/store
    required fields: entries
    risk: Create or Update Timesheet Clock Entries through the BambooHR API.
  create_or_update_timesheet_hour_entries:
    endpoint: POST /api/v1/time_tracking/hour_entries/store
    required fields: hours
    risk: Create or Update Timesheet Hour Entries through the BambooHR API.
  create_clock_entry:
    endpoint: POST /api/v1/time-tracking/clock-entries
    required fields: employeeId, start, end, timezone
    risk: Create Clock Entry; executes a BambooHR mutation against the configured account.
  delete_clock_entry:
    endpoint: DELETE /api/v1/time-tracking/clock-entries/{{ record.id }}
    required fields: id
    risk: Delete Clock Entry; executes a BambooHR mutation against the configured account.
  update_clock_entry:
    endpoint: PATCH /api/v1/time-tracking/clock-entries/{{ record.id }}
    required fields: id
    risk: Update Clock Entry; executes a BambooHR mutation against the configured account.
  clock_in_2:
    endpoint: POST /api/v1/time-tracking/clock-ins
    required fields: employeeId, timezone
    risk: Clock In; executes a BambooHR mutation against the configured account.
  clock_out_2:
    endpoint: POST /api/v1/time-tracking/clock-outs
    required fields: employeeId
    risk: Clock Out; executes a BambooHR mutation against the configured account.
  create_hour_entry:
    endpoint: POST /api/v1/time-tracking/hour-entries
    required fields: employeeId, date, hours
    risk: Create Hour Entry; executes a BambooHR mutation against the configured account.
  delete_hour_entry:
    endpoint: DELETE /api/v1/time-tracking/hour-entries/{{ record.id }}
    required fields: id
    risk: Delete Hour Entry; executes a BambooHR mutation against the configured account.
  update_hour_entry:
    endpoint: PATCH /api/v1/time-tracking/hour-entries/{{ record.id }}
    required fields: id
    risk: Update Hour Entry; executes a BambooHR mutation against the configured account.
  approve_timesheet:
    endpoint: POST /api/v1/time-tracking/timesheet-approvals
    required fields: timesheetId, lastChangedAt
    risk: Approve Timesheet; executes a BambooHR mutation against the configured account.
  create_webhook:
    endpoint: POST /api/v1/webhooks
    required fields: name, url, format
    risk: Create Webhook through the BambooHR API.
  update_webhook:
    endpoint: PUT /api/v1/webhooks/{{ record.id }}
    required fields: id, name, url, format
    risk: Update Webhook through the BambooHR API.
  delete_webhook:
    endpoint: DELETE /api/v1/webhooks/{{ record.id }}
    required fields: id
    risk: Deletes BambooHR data: Delete Webhook.
  put_company_industry_codes:
    endpoint: PUT /api/v1/company-profile-data/industry-codes
    required fields: industryIds
    risk: Update Company Industry Codes; executes a BambooHR mutation against the configured account.
  patch_company_profile_company_information:
    endpoint: PATCH /api/v1/company-profile-data/company-information
    risk: Update company information (phone, address, legal name); executes a BambooHR mutation against the configured account.
  put_company_profile_display_name:
    endpoint: PUT /api/v1/company-profile-data/display-name
    required fields: companyName
    risk: Update company display name; executes a BambooHR mutation against the configured account.
  update_compensation_benchmark:
    endpoint: PUT /api/v1/compensation/benchmarks
    required fields: id, currencyCode, benchmarkValue, benchmarkMin, benchmarkMax
    risk: Update Compensation Benchmark; executes a BambooHR mutation against the configured account.
  create_compensation_benchmark:
    endpoint: POST /api/v1/compensation/benchmarks
    required fields: jobTitleId, currencyCode, benchmarkValue, benchmarkMin, benchmarkMax
    risk: Create Compensation Benchmark; executes a BambooHR mutation against the configured account.
  delete_compensation_benchmark:
    endpoint: DELETE /api/v1/compensation/benchmarks/{{ record.id }}
    required fields: id
    risk: Delete Compensation Benchmark; executes a BambooHR mutation against the configured account.
  update_compensation_benchmark_sources:
    endpoint: PUT /api/v1/compensation/benchmarks/sources
    required fields: benchmarkSources
    risk: Update Compensation Benchmark Sources; executes a BambooHR mutation against the configured account.
  create_compensation_benchmark_source:
    endpoint: POST /api/v1/compensation/benchmarks/sources
    required fields: name
    risk: Create Compensation Benchmark Source; executes a BambooHR mutation against the configured account.
  delete_compensation_benchmark_source:
    endpoint: DELETE /api/v1/compensation/benchmarks/sources
    required fields: id
    risk: Delete Compensation Benchmark Source; executes a BambooHR mutation against the configured account.
  create_employee:
    endpoint: POST /api/v1/employees
    required fields: firstName, lastName
    risk: Create Employee; executes a BambooHR mutation against the configured account.
  update_employee:
    endpoint: POST /api/v1/employees/{{ record.id }}
    required fields: id
    risk: Update Employee; executes a BambooHR mutation against the configured account.
  delete_employee:
    endpoint: DELETE /api/v1/employees/{{ record.id }}
    required fields: id
    risk: Delete employee; executes a BambooHR mutation against the configured account.
  update_table_row:
    endpoint: POST /api/v1/employees/{{ record.id }}/tables/{{ record.table }}/{{ record.row_id }}
    required fields: id, table, row_id
    risk: Update Table Row through the BambooHR API.
  delete_employee_table_row:
    endpoint: DELETE /api/v1/employees/{{ record.id }}/tables/{{ record.table }}/{{ record.row_id }}
    required fields: id, table, row_id
    risk: Deletes BambooHR data: Delete Employee Table Row.
  update_company_file:
    endpoint: POST /api/v1/files/{{ record.file_id }}
    required fields: file_id
    risk: Update Company File; executes a BambooHR mutation against the configured account.
  delete_company_file:
    endpoint: DELETE /api/v1/files/{{ record.file_id }}
    required fields: file_id
    risk: Delete Company File; executes a BambooHR mutation against the configured account.
  update_employee_file:
    endpoint: POST /api/v1/employees/{{ record.id }}/files/{{ record.file_id }}
    required fields: id, file_id
    risk: Update Employee File; executes a BambooHR mutation against the configured account.
  delete_employee_file:
    endpoint: DELETE /api/v1/employees/{{ record.id }}/files/{{ record.file_id }}
    required fields: id, file_id
    risk: Delete Employee File; executes a BambooHR mutation against the configured account.
  update_compensation_level_groups_and_levels:
    endpoint: PUT /api/v1/pay-grades-and-bands/levels
    required fields: groups
    risk: Update Compensation Level Groups and Levels; executes a BambooHR mutation against the configured account.
  update_pay_bands:
    endpoint: PUT /api/v1/pay-grades-and-bands/pay-bands
    required fields: payBands
    risk: Update Pay Bands; executes a BambooHR mutation against the configured account.
  replace_job_title_level_assignments:
    endpoint: PUT /api/v1/pay-grades-and-bands/job-titles
    required fields: jobTitles
    risk: Replace Job Title Level Assignments; executes a BambooHR mutation against the configured account.
  publish_draft_compensation_level_groups:
    endpoint: POST /api/v1/pay-grades-and-bands/publish
    risk: Publish Draft Compensation Level Groups; executes a BambooHR mutation against the configured account.
  delete_compensation_level_groups_or_level:
    endpoint: DELETE /api/v1/pay-grades-and-bands/levels/{{ record.segment }}
    required fields: segment
    risk: Delete Compensation Level Groups or Level; executes a BambooHR mutation against the configured account.
  create_goal:
    endpoint: POST /api/v1/performance/employees/{{ record.employee_id }}/goals
    required fields: employee_id, title, dueDate, sharedWithEmployeeIds
    risk: Create Goal through the BambooHR API.
  update_goal_v1:
    endpoint: PUT /api/v1/performance/employees/{{ record.employee_id }}/goals/{{ record.goal_id }}
    required fields: employee_id, goal_id, dueDate, sharedWithEmployeeIds, title
    risk: Update Goal (v1) through the BambooHR API.
  delete_goal:
    endpoint: DELETE /api/v1/performance/employees/{{ record.employee_id }}/goals/{{ record.goal_id }}
    required fields: employee_id, goal_id
    risk: Deletes BambooHR data: Delete Goal.
  update_goal_progress:
    endpoint: PUT /api/v1/performance/employees/{{ record.employee_id }}/goals/{{ record.goal_id }}/progress
    required fields: employee_id, goal_id, percentComplete
    risk: Update Goal Progress through the BambooHR API.
  update_goal_milestone_progress:
    endpoint: PUT /api/v1/performance/employees/{{ record.employee_id }}/goals/{{ record.goal_id }}/milestones/{{ record.milestone_id }}/progress
    required fields: employee_id, goal_id, milestone_id, complete
    risk: Update Milestone Progress through the BambooHR API.
  update_goal_sharing:
    endpoint: PUT /api/v1/performance/employees/{{ record.employee_id }}/goals/{{ record.goal_id }}/sharedWith
    required fields: employee_id, goal_id
    risk: Update Goal Sharing through the BambooHR API.
  create_goal_comment:
    endpoint: POST /api/v1/performance/employees/{{ record.employee_id }}/goals/{{ record.goal_id }}/comments
    required fields: employee_id, goal_id, text
    risk: Create Goal Comment through the BambooHR API.
  update_goal_comment:
    endpoint: PUT /api/v1/performance/employees/{{ record.employee_id }}/goals/{{ record.goal_id }}/comments/{{ record.comment_id }}
    required fields: employee_id, goal_id, comment_id, text
    risk: Update Goal Comment through the BambooHR API.
  delete_goal_comment:
    endpoint: DELETE /api/v1/performance/employees/{{ record.employee_id }}/goals/{{ record.goal_id }}/comments/{{ record.comment_id }}
    required fields: employee_id, goal_id, comment_id
    risk: Deletes BambooHR data: Delete Goal Comment.
  close_goal:
    endpoint: POST /api/v1/performance/employees/{{ record.employee_id }}/goals/{{ record.goal_id }}/close
    required fields: employee_id, goal_id
    risk: Close Goal through the BambooHR API.
  reopen_goal:
    endpoint: POST /api/v1/performance/employees/{{ record.employee_id }}/goals/{{ record.goal_id }}/reopen
    required fields: employee_id, goal_id
    risk: Reopen Goal through the BambooHR API.
  update_goal_v1_1:
    endpoint: PUT /api/v1_1/performance/employees/{{ record.employee_id }}/goals/{{ record.goal_id }}
    required fields: employee_id, goal_id, title, dueDate, sharedWithEmployeeIds
    risk: Update Goal (v1.1) through the BambooHR API.
  create_time_tracking_hour_record:
    endpoint: POST /api/v1/timetracking/add
    required fields: dateHoursWorked, employeeId, hoursWorked, rateType, timeTrackingId
    risk: Create Hour Record through the BambooHR API.
  create_or_update_time_tracking_hour_records:
    endpoint: POST /api/v1/timetracking/record
    risk: Create or Update Hour Records through the BambooHR API.
  update_time_tracking_record:
    endpoint: PUT /api/v1/timetracking/adjust
    required fields: timeTrackingId, hoursWorked
    risk: Update Hour Record through the BambooHR API.
  delete_time_tracking_hour_record:
    endpoint: DELETE /api/v1/timetracking/delete/{{ record.id }}
    required fields: id
    risk: Deletes BambooHR data: Delete Hour Record.
  add_total_rewards_employees:
    endpoint: POST /api/v1/compensation/total_rewards/employees
    required fields: employeeIds
    risk: Add Employees to Total Rewards; executes a BambooHR mutation against the configured account.
  remove_total_rewards_employees:
    endpoint: DELETE /api/v1/compensation/total_rewards/employees
    required fields: employeeIds
    risk: Remove Employees from Total Rewards; executes a BambooHR mutation against the configured account.
  set_total_rewards_onboarding_step:
    endpoint: PUT /api/v1/compensation/total_rewards/onboarding/{{ record.step_name }}
    required fields: step_name, completed
    risk: Set Total Rewards Onboarding Step Status; executes a BambooHR mutation against the configured account.
  set_total_rewards_custom_disclaimer:
    endpoint: PUT /api/v1/compensation/total_rewards/custom_disclaimer
    required fields: customDisclaimer
    risk: Set Total Rewards Custom Disclaimer; executes a BambooHR mutation against the configured account.
  remove_total_rewards_custom_disclaimer:
    endpoint: DELETE /api/v1/compensation/total_rewards/custom_disclaimer
    risk: Remove Total Rewards Custom Disclaimer; executes a BambooHR mutation against the configured account.
  create_location:
    endpoint: POST /api/v1/hris/org/locations
    required fields: label, address
    risk: Create a job location; executes a BambooHR mutation against the configured account.
  update_location:
    endpoint: PUT /api/v1/hris/org/locations/{{ record.id }}
    required fields: id, label, address
    risk: Update a job location; executes a BambooHR mutation against the configured account.
  delete_location:
    endpoint: DELETE /api/v1/hris/org/locations/{{ record.id }}
    required fields: id
    risk: Delete a job location; executes a BambooHR mutation against the configured account.
  create_table_row:
    endpoint: POST /api/v1/employees/{{ record.id }}/tables/{{ record.table }}
    required fields: id, table
    risk: Create Table Row through the BambooHR API.
  update_table_row_v1_1:
    endpoint: POST /api/v1_1/employees/{{ record.id }}/tables/{{ record.table }}/{{ record.row_id }}
    required fields: id, table, row_id
    risk: Update Table Row v1.1 through the BambooHR API.
  create_table_row_v1_1:
    endpoint: POST /api/v1_1/employees/{{ record.id }}/tables/{{ record.table }}
    required fields: id, table
    risk: Create Table Row v1.1 through the BambooHR API.
  create_employee_file_category:
    endpoint: POST /api/v1/employees/files/categories
    risk: Create Employee File Category; executes a BambooHR mutation against the configured account.
  create_company_file_category:
    endpoint: POST /api/v1/files/categories
    risk: Create Company File Category; executes a BambooHR mutation against the configured account.
  assign_time_off_policies:
    endpoint: PUT /api/v1/employees/{{ record.employee_id }}/time_off/policies
    required fields: employee_id
    risk: Assign Time Off Policies through the BambooHR API.
  assign_time_off_policies_v1_1:
    endpoint: PUT /api/v1_1/employees/{{ record.employee_id }}/time_off/policies
    required fields: employee_id
    risk: Assign Time Off Policies v1.1 through the BambooHR API.
  create_application_comment:
    endpoint: POST /api/v1/applicant_tracking/applications/{{ record.application_id }}/comments
    required fields: application_id, comment
    risk: Create Job Application Comment through the BambooHR API.
  update_applicant_status:
    endpoint: POST /api/v1/applicant_tracking/applications/{{ record.application_id }}/status
    required fields: application_id, status
    risk: Update Applicant Status through the BambooHR API.
  update_employee_dependent:
    endpoint: PUT /api/v1/employeedependents/{{ record.id }}
    required fields: id, employeeId
    risk: Update Employee Dependent through the BambooHR API.
  create_employee_dependent:
    endpoint: POST /api/v1/employeedependents
    required fields: employeeId
    risk: Create Employee Dependent through the BambooHR API.
  update_list_field_values:
    endpoint: PUT /api/v1/meta/lists/{{ record.list_field_id }}
    required fields: list_field_id
    risk: Update List Field Values through the BambooHR API.
  create_time_off_history:
    endpoint: PUT /api/v1/employees/{{ record.employee_id }}/time_off/history
    required fields: employee_id, date
    risk: Create Time Off History Item through the BambooHR API.
  adjust_time_off_balance:
    endpoint: PUT /api/v1/employees/{{ record.employee_id }}/time_off/balance_adjustment
    required fields: employee_id, amount, date, timeOffTypeId
    risk: Adjust Time Off Balance through the BambooHR API.
  create_time_off_request:
    endpoint: PUT /api/v1/employees/{{ record.employee_id }}/time_off/request
    required fields: employee_id, status, start, end, timeOffTypeId
    risk: Create Time Off Request through the BambooHR API.
  update_time_off_request_status:
    endpoint: PUT /api/v1/time_off/requests/{{ record.request_id }}/status
    required fields: request_id, status
    risk: Update Time Off Request Status through the BambooHR API.
  create_training_type:
    endpoint: POST /api/v1/training/type
    required fields: name
    risk: Create Training Type through the BambooHR API.
  update_training_type:
    endpoint: PUT /api/v1/training/type/{{ record.training_type_id }}
    required fields: training_type_id
    risk: Update Training Type through the BambooHR API.
  delete_training_type:
    endpoint: DELETE /api/v1/training/type/{{ record.training_type_id }}
    required fields: training_type_id
    risk: Deletes BambooHR data: Delete Training Type.
  create_training_category:
    endpoint: POST /api/v1/training/category
    required fields: name
    risk: Create Training Category through the BambooHR API.
  update_training_category:
    endpoint: PUT /api/v1/training/category/{{ record.training_category_id }}
    required fields: training_category_id, name
    risk: Update Training Category through the BambooHR API.
  delete_training_category:
    endpoint: DELETE /api/v1/training/category/{{ record.training_category_id }}
    required fields: training_category_id
    risk: Deletes BambooHR data: Delete Training Category.
  create_employee_training_record:
    endpoint: POST /api/v1/training/record/employee/{{ record.employee_id }}
    required fields: employee_id, completed, type
    risk: Create Employee Training Record through the BambooHR API.
  update_employee_training_record:
    endpoint: PUT /api/v1/training/record/{{ record.employee_training_record_id }}
    required fields: employee_training_record_id, completed
    risk: Update Employee Training Record through the BambooHR API.
  delete_employee_training_record:
    endpoint: DELETE /api/v1/training/record/{{ record.employee_training_record_id }}
    required fields: employee_training_record_id
    risk: Deletes BambooHR data: Delete Employee Training Record.

SECURITY
  read risk: external BambooHR API reads across HR, applicant tracking, benefits, payroll-adjacent, compensation, time off, training, goals, metadata, and webhook-management resources
  write risk: typed reverse-ETL mutations create, update, assign, approve, import, or delete BambooHR records according to the selected named action
  approval: reverse ETL writes require plan preview, approval token, and destructive confirmation when declared
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Read BambooHR streams, run bounded JSON direct reads, and safely plan typed BambooHR write actions without generic API passthrough.
  Usage: pm bamboo-hr <command> [flags]
  Source CLI: BambooHR API (OpenAPI 3.1 public-openapi.yaml)
  Global flags:
    --credential (string): Stored BambooHR credential name.
    --connection (string): Alias for --credential.
    --config (string_array): Additional key=value connector config overrides.
    --limit (integer): Maximum ETL records for stream commands.
    --max-bytes (integer): Maximum bytes for bounded direct reads.
  Break
    break policy suggestions get - Get Break Policy Suggestions [intent=direct_read availability=implemented]; notes: Bounded fixed-target JSON POST direct read; schema-gated body flags only.; flags: --prompt
  Countries
    countries options get - Get Countries [intent=direct_read availability=implemented]; notes: Bounded fixed-target JSON direct read; no raw method, path, query, or body passthrough.
  Custom
    custom report request - Request Custom Report [intent=direct_read availability=implemented]; notes: Bounded fixed-target JSON POST direct read; schema-gated body flags only. Object-valued request fields that cannot be represented by typed scalar CLI flags are intentionally omitted; use supported primitive/list flags only, with no generic JSON body passthrough.; flags: --title, --fields, --filter-duplicates
  Data
    data from dataset get-v1 - Get Data from Dataset (v1) [intent=direct_read availability=implemented]; notes: Bounded fixed-target JSON POST direct read; schema-gated body flags only. Object-valued request fields that cannot be represented by typed scalar CLI flags are intentionally omitted; use supported primitive/list flags only, with no generic JSON body passthrough.; flags: --dataset-name, --fields, --group-by, --show-history
    data from dataset get-v2 - Get Data from Dataset (v2) [intent=direct_read availability=implemented]; notes: Bounded fixed-target JSON POST direct read; schema-gated body flags only.; flags: --dataset-name, --fields, --filter, --order-by, --page, --page-size
  Employee
    employee trainings list - List Employee Training Records [intent=direct_read availability=implemented]; notes: Bounded fixed-target JSON direct read; no raw method, path, query, or body passthrough.; flags: --employee-id
  Field
    field options get-v1 - Get Field Options (v1) [intent=direct_read availability=implemented]; notes: Bounded fixed-target JSON POST direct read; schema-gated body flags only.; flags: --dataset-name, --fields, --dependent-fields, --filters
    field options v1 2 get - Get Field Options (v1.2) [intent=direct_read availability=implemented]; notes: Bounded fixed-target JSON POST direct read; schema-gated body flags only.; flags: --dataset-name, --fields, --dependent-fields, --filters
  Webhook
    webhook logs list - List Webhook Logs [intent=direct_read availability=implemented]; notes: Bounded fixed-target JSON direct read; no raw method, path, query, or body passthrough.; flags: --id
  Help topics:
    auth - Use a stored BambooHR API key for Basic auth; OAuth-scoped operations may also require an access token secret.
    safety - No command exposes raw method, path, body, query, shell, file, or passthrough access.

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
