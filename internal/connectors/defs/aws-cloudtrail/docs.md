# AWS CloudTrail Connector

## Overview

Reads CloudTrail management events through fixed, declaration-bound LookupEvents requests signed with AWS Signature Version 4.

Readable streams: `management_events`, `read_only_events`, `write_only_events`, `console_logins`.

Service API documentation: https://docs.aws.amazon.com/awscloudtrail/latest/APIReference/API_LookupEvents.html.

## Auth setup

Connection fields:

- `aws_key_id` (required, secret, string); AWS access key ID.
- `aws_secret_key` (required, secret, string); AWS secret access key.
- `aws_session_token` (optional, secret, string); Optional AWS session token.
- `start_date` (optional, string); Optional RFC3339 StartTime lower bound sent in every CloudTrail LookupEvents request.

Authentication uses declared mode(s): `aws_sigv4`, `aws_sigv4`.

## Execution contract

Connection check: `POST /`
Check JSON body: `MaxResults`=1.

## Streams notes

- `management_events`: `POST /`; records `Events`
  - JSON body: `MaxResults`=50, `StartTime`={{ incremental.lower_bound }}.
  - Incremental cursor: `EventTime`.
- `read_only_events`: `POST /`; records `Events`
  - JSON body: `LookupAttributes`=[{'AttributeKey': 'ReadOnly', 'AttributeValue': 'true'}], `MaxResults`=50, `StartTime`={{ incremental.lower_bound }}.
  - Incremental cursor: `EventTime`.
- `write_only_events`: `POST /`; records `Events`
  - JSON body: `LookupAttributes`=[{'AttributeKey': 'ReadOnly', 'AttributeValue': 'false'}], `MaxResults`=50, `StartTime`={{ incremental.lower_bound }}.
  - Incremental cursor: `EventTime`.
- `console_logins`: `POST /`; records `Events`
  - JSON body: `LookupAttributes`=[{'AttributeKey': 'EventName', 'AttributeValue': 'ConsoleLogin'}], `MaxResults`=50, `StartTime`={{ incremental.lower_bound }}.
  - Incremental cursor: `EventTime`.

## Write actions & risks

This connector's write surface is declared separately in the rendered execution bundle.
