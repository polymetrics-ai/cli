# Keka

## Overview

Keka reads the documented HRMS API surface published at https://apidocs.keka.com/. Streams cover Core HR, documents, leave, attendance, payroll, PSA, PMS, hiring, expense, assets, requisitions, skills, and BGV resources. The original legacy streams remain first in the catalog and keep their legacy projection schemas: employees, attendance, leave_types, leave_requests, clients, and projects.

## Auth setup

Set `base_url` to the company-specific API root, for example `https://<company>.keka.com/api/v1`. The bundle uses the existing `keka` custom AuthHook to exchange `client_id`, `client_secret`, optional `api_key`, `grant_type`, and `scope` at `token_url`; secret-shaped values stay in credential storage and are marked `x-secret` where appropriate.

## Streams notes

Keka list endpoints use `pageNumber` and `pageSize` with records under `data`. Detail and singleton endpoints override pagination with `type: none` and read the object at `data` or, for exit reasons, the response root. Streams whose paths require an existing employee, document type, project, task, job, candidate, payroll batch, or BGV identifier read those IDs from optional `config.*` fields declared in `spec.json`. New Pass B streams use passthrough projection with sample-derived schemas so Keka module-specific fields are not silently dropped.

## Write actions & risks

The bundle exposes 24 object-bodied Keka mutations as write actions, including employee create/update, exit requests, leave requests, payroll payment status, PSA clients/projects/tasks/allocations, goal progress, praise, hiring candidates and notes, preboarding candidates, asset assignment, and BGV report updates. Writes execute one HTTP request per record and require the platform reverse-ETL plan, preview, and approval flow before live mutation. Path identifiers such as `employee_id`, `project_id`, `task_id`, `job_id`, and `candidate_id` are removed from the JSON body because they are already carried in the URL.

## Known limits

- The OAuth token endpoint is a non-data endpoint implemented by the Keka AuthHook, not exposed as a write.
- The official collection contains three named GET operations without request URLs: payroll bonus types, hire job boards, and skills. They are recorded in `api_surface.json` as excluded rather than inventing paths.
- Attendance punch pushes and employee skill additions require a top-level JSON array request body. The current declarative write dialect sends JSON/form objects only, so those two operations are excluded with explicit ENGINE_GAP reasons.
- Path-scoped streams require caller-provided IDs in config; this bundle does not chain multi-level fan-out for every parent/child resource combination.
- Dynamic conformance remains skipped because Keka authentication is owned by a custom token-exchange AuthHook whose token URL is not replayable by the generic fixture harness; static conformance still validates the expanded bundle.
