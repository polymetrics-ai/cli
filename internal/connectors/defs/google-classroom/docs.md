# Google Classroom Connector

## Overview

Reads Classroom courses and course-scoped resources through fixed REST routes and OAuth2 refresh-token authentication.

Readable streams: `courses`, `teachers`, `students`, `course_work`, `announcements`.

Service API documentation: https://developers.google.com/workspace/classroom/reference/rest.

## Auth setup

Connection fields:

- `client_id` (required, secret, string);
- `client_refresh_token` (required, secret, string);
- `client_secret` (required, secret, string);

Authentication uses declared mode(s): `oauth2_refresh_token`.

## Execution contract

Default stream pagination: `cursor`.

Connection check: `GET /v1/courses`
Check query: `pageSize`=`1`.

## Streams notes

- `courses`: `GET /v1/courses`; records `courses`
- `teachers`: `GET /v1/courses/{{ fanout.id }}/teachers`; records `teachers`
- `students`: `GET /v1/courses/{{ fanout.id }}/students`; records `students`
- `course_work`: `GET /v1/courses/{{ fanout.id }}/courseWork`; records `courseWork`
- `announcements`: `GET /v1/courses/{{ fanout.id }}/announcements`; records `announcements`

## Write actions & risks

This connector's write surface is declared separately in the rendered execution bundle.
