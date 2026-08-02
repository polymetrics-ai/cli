#!/usr/bin/env python3
import json
import re
import shutil
from pathlib import Path

# Generator for the scope-corrected 60-action CloudTrail ledger. The current
# connector-local surface exposes fixed read streams through connector-local
# discovery/fan-out; direct reads and writes remain blocked.
ROOT = Path(__file__).resolve().parents[3]
# The script is stored under .planning/phases/aws-cloudtrail-parity-wave04/traces.
if not (ROOT / 'internal/connectors').exists():
    ROOT = Path.cwd()
LEDGER = ROOT / '.planning/phases/aws-cloudtrail-parity-wave04/traces/aws-cloudtrail-api-field-ledger-typed.json'
DEF = ROOT / 'internal/connectors/defs/aws-cloudtrail'

ops = {row['action']: row for row in json.loads(LEDGER.read_text())}

def snake(name: str) -> str:
    out = []
    for i, ch in enumerate(name):
        if ch.isupper() and i > 0 and (not name[i-1].isupper() or (i+1 < len(name) and name[i+1].islower())):
            out.append('_')
        out.append(ch.lower())
    return ''.join(out)

DIRECT = [
    'CancelQuery', 'DescribeQuery', 'GenerateQuery', 'GetQueryResults', 'ListInsightsData',
    'ListInsightsMetricData', 'ListQueries', 'LookupEvents', 'SearchSampleQueries', 'StartQuery',
]
WRITES = [
    'AddTags', 'CreateChannel', 'CreateDashboard', 'CreateEventDataStore', 'CreateTrail',
    'DeleteChannel', 'DeleteDashboard', 'DeleteEventDataStore', 'DeleteResourcePolicy', 'DeleteTrail',
    'DeregisterOrganizationDelegatedAdmin', 'DisableFederation', 'EnableFederation',
    'PutEventConfiguration', 'PutEventSelectors', 'PutInsightSelectors', 'PutResourcePolicy',
    'RegisterOrganizationDelegatedAdmin', 'RemoveTags', 'RestoreEventDataStore',
    'StartDashboardRefresh', 'StartEventDataStoreIngestion', 'StartImport', 'StartLogging',
    'StopEventDataStoreIngestion', 'StopImport', 'StopLogging', 'UpdateChannel', 'UpdateDashboard',
    'UpdateEventDataStore', 'UpdateTrail',
]
STREAM_READ_QUERY_OVERRIDES = {
    'GetInsightSelectors': {'TrailName': 'trail-fixture'},
}
STREAMS = [a for a in ops if a not in DIRECT and a not in WRITES]
assert len(DIRECT) == 10, len(DIRECT)
assert len(WRITES) == 31, len(WRITES)
assert len(STREAMS) == 19, (len(STREAMS), STREAMS)

EVENT_FIELDS = [
    'eventTime','eventVersion','userIdentity','eventSource','eventName','awsRegion','sourceIPAddress','userAgent',
    'errorCode','errorMessage','requestParameters','responseElements','additionalEventData','requestID','eventID',
    'eventType','apiVersion','managementEvent','readOnly','resources','recipientAccountId','serviceEventDetails',
    'sharedEventID','vpcEndpointId','vpcEndpointAccountId','eventCategory','addendum','sessionCredentialFromConsole',
    'eventContext','edgeDeviceDetails','tlsDetails'
]

def typed_type(base: str, nullable: bool):
    return [base, 'null'] if nullable else base

def schema_type(aws_type: str, nullable: bool = True):
    t = aws_type.lower()
    if 'boolean' in t:
        return {'type': typed_type('boolean', nullable)}
    if 'integer' in t:
        return {'type': typed_type('integer', nullable)}
    if 'timestamp' in t:
        return {'type': ['string', 'integer', 'null'] if nullable else ['string', 'integer'], 'description': 'AWS timestamp; CLI flags use RFC3339, native requests send Unix seconds.'}
    if 'array of strings' in t:
        return {'type': typed_type('array', nullable), 'items': {'type': 'string'}}
    if 'array of' in t:
        return {'type': typed_type('array', nullable), 'items': {'type': 'object', 'additionalProperties': True}}
    if 'map' in t or 'object' in t:
        return {'type': typed_type('object', nullable), 'additionalProperties': True}
    return {'type': typed_type('string', nullable)}

def write_schema(action: str):
    row = ops[action]
    props = {f['name']: schema_type(f['aws_type'], nullable=False) for f in row['request_fields']}
    req = [f['name'] for f in row['request_fields'] if f['required']]
    return {'type': 'object', 'required': req, 'additionalProperties': False, 'properties': props}

def sample_value(field: dict):
    name = field['name']
    t = field['aws_type'].lower()
    if 'boolean' in t:
        return True
    if 'integer' in t:
        if name in ('MaxResults', 'MaxQueryResults'):
            return 1
        if name == 'Period':
            return 60
        if name == 'RetentionPeriod':
            return 90
        return 1
    if 'timestamp' in t:
        return '2026-01-01T00:00:00Z'
    if 'array of strings' in t:
        return [sample_string(name)]
    if 'array of tag objects' in t.lower():
        return [{'Key': 'fixture-key', 'Value': 'fixture-value'}]
    if 'array of destination objects' in t.lower():
        return [{'Type': 'EVENT_DATA_STORE', 'Location': 'arn:aws:cloudtrail:us-east-1:123456789012:eventdatastore/fixture'}]
    if 'array of' in t:
        return [{'Name': 'fixture'}]
    if 'map' in t:
        return {'fixtureKey': 'fixtureValue'}
    if 'object' in t:
        return {'Name': 'fixture'}
    return sample_string(name)

def sample_string(name: str) -> str:
    samples = {
        'Name': 'trail-fixture',
        'TrailName': 'trail-fixture',
        'Channel': 'arn:aws:cloudtrail:us-east-1:123456789012:channel/fixture-channel',
        'ChannelArn': 'arn:aws:cloudtrail:us-east-1:123456789012:channel/fixture-channel',
        'DashboardId': 'arn:aws:cloudtrail:us-east-1:123456789012:dashboard/fixture-dashboard',
        'EventDataStore': 'arn:aws:cloudtrail:us-east-1:123456789012:eventdatastore/fixture-store',
        'EventDataStores': 'arn:aws:cloudtrail:us-east-1:123456789012:eventdatastore/fixture-store',
        'ResourceArn': 'arn:aws:cloudtrail:us-east-1:123456789012:eventdatastore/fixture-store',
        'ResourceId': 'arn:aws:cloudtrail:us-east-1:123456789012:trail/fixture-trail',
        'ResourceIdList': 'arn:aws:cloudtrail:us-east-1:123456789012:trail/fixture-trail',
        'QueryId': '11111111-1111-1111-1111-111111111111',
        'RefreshId': '1234567890',
        'ImportId': '22222222-2222-2222-2222-222222222222',
        'DelegatedAdminAccountId': '123456789012',
        'MemberAccountId': '123456789012',
        'EventName': 'ConsoleLogin',
        'EventSource': 'signin.amazonaws.com',
        'InsightType': 'ApiCallRateInsight',
        'DataType': 'InsightsEvents',
        'InsightSource': 'arn:aws:cloudtrail:us-east-1:123456789012:trail/fixture-trail',
        'Prompt': 'show console logins',
        'SearchPhrase': 'sample',
        'QueryStatement': 'SELECT eventTime, eventName FROM fixture_events LIMIT 10',
        'S3BucketName': 'fixture-cloudtrail-logs',
        'ResourcePolicy': '{"Version":"2012-10-17","Statement":[]}',
        'FederationRoleArn': 'arn:aws:iam::123456789012:role/fixture-cloudtrail-federation',
        'Source': 'aws.partner/fixture',
        'DeliveryS3Uri': 's3://fixture-query-results/prefix',
    }
    if name in samples:
        return samples[name]
    if name.endswith('Arn') or name.endswith('ARN'):
        return 'arn:aws:cloudtrail:us-east-1:123456789012:trail/fixture-trail'
    return 'fixture-' + re.sub(r'[^a-z0-9]+', '-', snake(name)).strip('-')

def sample_record(action: str, include_optional=False):
    row = ops[action]
    fields = row['request_fields'] if include_optional else [f for f in row['request_fields'] if f['required']]
    if not fields:
        fields = row['request_fields'][:1]
    return {f['name']: sample_value(f) for f in fields}

def flag_type(field: dict):
    t = field['aws_type'].lower()
    if 'boolean' in t:
        return 'boolean'
    if 'integer' in t:
        return 'integer'
    if 'array of strings' in t:
        return 'string_array'
    return 'string'

def flag_for(field: dict, namespace: str):
    name = snake(field['name']).replace('_','-')
    flag = {'name': name, 'type': flag_type(field), 'summary': field['name'] + '.', 'maps_to': namespace + '.' + field['name']}
    if flag['type'] == 'string':
        flag['allow_empty'] = False
        if 'timestamp' in field['aws_type'].lower():
            flag['format'] = 'date-time'
    return flag

def op_id(action: str):
    return 'aws-cloudtrail.' + snake(action).replace('_','-').replace('-', '_')

def direct_command_path(action: str):
    mapping = {
        'CancelQuery': 'query cancel',
        'DescribeQuery': 'query describe',
        'GenerateQuery': 'query generate',
        'GetQueryResults': 'query results',
        'ListInsightsData': 'insights data',
        'ListInsightsMetricData': 'insights metric-data',
        'ListQueries': 'query list',
        'LookupEvents': 'events lookup',
        'SearchSampleQueries': 'sample-queries search',
        'StartQuery': 'query start',
    }
    return mapping[action]

def stream_command_path(action: str):
    return 'read ' + snake(action).replace('_','-')

def source_url(action: str):
    return f'https://docs.aws.amazon.com/awscloudtrail/latest/APIReference/API_{action}.html'

def summary_for(action: str):
    return f'Run the documented AWS CloudTrail {action} operation through a fixed signed JSON-RPC request.'

def json_dump(path: Path, data):
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(data, indent=2) + '\n')

# Clean generated connector-local mutable files.
for child in ['schemas', 'fixtures']:
    shutil.rmtree(DEF / child, ignore_errors=True)
for name in ['metadata.json','spec.json','streams.json','writes.json','operations.json','api_surface.json','cli_surface.json','certification.json','docs.md']:
    try:
        (DEF / name).unlink()
    except FileNotFoundError:
        pass

metadata = {
    'name': 'aws-cloudtrail',
    'display_name': 'AWS CloudTrail',
    'description': 'Reads AWS CloudTrail configuration and resource metadata through fixed AWS JSON-RPC streams. Provider query/direct-read and write/admin actions remain planned until shared promoted-native forwarding exposes them safely at runtime.',
    'integration_type': 'api',
    'release_stage': 'alpha',
    'capabilities': {'check': True, 'read': True, 'write': False, 'query': False, 'cdc': False, 'dynamic_schema': False},
    'batch': {'read_page_size': 50, 'write_batch_size': 1},
    'risk': {
        'read': 'bounded AWS CloudTrail JSON-RPC reads using fixed action names, SigV4 authentication, and connector-local resource discovery for parameterized streams',
        'write': 'blocked/planned: CloudTrail write/admin actions require shared promoted-native manifest, command-surface, validation, and dry-run forwarding before they can be safely exposed.',
        'approval': 'No CloudTrail writes are exposed in this scope-corrected connector surface; future writes must preserve plan -> preview -> approval -> execute.'
    },
    'docs_url': 'https://docs.aws.amazon.com/awscloudtrail/latest/APIReference/API_Operations.html'
}
json_dump(DEF/'metadata.json', metadata)

spec = {
    '$schema': 'http://json-schema.org/draft-07/schema#',
    'title': 'AWS CloudTrail Connection Specification',
    'type': 'object',
    'required': ['aws_key_id','aws_region_name','aws_secret_key'],
    'properties': {
        'aws_key_id': {'type': 'string', 'x-secret': True, 'description': 'AWS access key ID. Provide with --from-env or stdin; never paste the value into chat or logs.'},
        'aws_region_name': {'type': 'string', 'description': 'AWS region for CloudTrail, for example us-east-1.'},
        'aws_secret_key': {'type': 'string', 'x-secret': True, 'description': 'AWS secret access key. Provide with --from-env or stdin; never paste the value into chat or logs.'},
        'base_url': {'type': 'string', 'description': 'Optional CloudTrail endpoint override for fixture/local testing. Production defaults to https://cloudtrail.<region>.amazonaws.com.'},
        'page_size': {'type': 'string', 'description': 'Optional bounded MaxResults value for paginated read/query actions.'},
        'max_pages': {'type': 'string', 'description': 'Optional maximum page count for paginated read/query actions; use a positive integer.'},
        'mode': {'type': 'string', 'description': 'Set to fixture for credential-free native fixture tests.'}
    }
}
json_dump(DEF/'spec.json', spec)

streams_json = {
    'base': {
        'url': 'https://cloudtrail.{{ config.aws_region_name }}.amazonaws.com',
        'user_agent': 'polymetrics-go-cli',
        'headers': {'Content-Type': 'application/x-amz-json-1.1', 'Accept': 'application/x-amz-json-1.1'},
        'auth': [{'mode': 'none'}],
        'pagination': {'type': 'none'},
        'check': {'method': 'POST', 'path': '/'},
        'error_map': [
            {'status': 400, 'match_body': 'InvalidClientTokenId', 'class': 'auth_failed', 'hint': 'AWS rejected the access key; rotate aws_key_id/aws_secret_key.'},
            {'status': 400, 'match_body': 'AccessDenied', 'class': 'permission_denied', 'hint': 'credential lacks permission for the fixed CloudTrail action.'},
            {'status': 403, 'class': 'permission_denied', 'hint': 'credential lacks permission for CloudTrail.'}
        ]
    },
    'streams': []
}
for action in STREAMS:
    stream = snake(action)
    streams_json['streams'].append({
        'name': stream,
        'method': 'POST',
        'path': '/',
        'records': {'path': 'data'},
        'schema': f'schemas/{stream}.json',
        'projection': 'passthrough'
    })
json_dump(DEF/'streams.json', streams_json)

for action in STREAMS:
    stream = snake(action)
    props = {
        'pm_record_id': {'type': 'string'},
        'operation': {'type': 'string'},
    }
    for f in ops[action]['response_fields']:
        props[f['name']] = schema_type(f.get('aws_type',''))
    schema = {
        '$schema': 'http://json-schema.org/draft-07/schema#',
        'title': stream,
        'type': 'object',
        'x-primary-key': ['pm_record_id'],
        'properties': props,
        'additionalProperties': True
    }
    json_dump(DEF/f'schemas/{stream}.json', schema)

writes = {'actions': []}
json_dump(DEF/'writes.json', writes)

operations = {'operations': []}
json_dump(DEF/'operations.json', operations)

# API surface rows in official action order.
def blocked_operation(action: str):
    if action in DIRECT:
        return {
            'model': 'direct_read',
            'status': 'blocked',
            'risk': 'high',
            'blocked_by_default': True,
            'reason': 'Planned/blocked after scope correction: this CloudTrail provider query command requires shared promoted-native CommandSurface/OperationDirectReader forwarding that was reverted; it is not executable through the default pm aws-cloudtrail command surface.',
            'source_url': source_url(action),
            'notes': 'Re-enable only with a shared-runtime slice that forwards operation direct reads from promoted native connectors.'
        }
    kind = 'destructive_action' if action.startswith(('Delete','Disable','Stop')) else 'admin_reverse_etl'
    return {
        'model': kind,
        'status': 'blocked',
        'risk': 'critical' if kind == 'destructive_action' else 'high',
        'blocked_by_default': True,
        'reason': 'Planned/blocked after scope correction: this CloudTrail write/admin command requires shared promoted-native Manifest/CommandSurface/WriteValidator/DryRunWriter forwarding for typed plan-preview-approval execution; that shared runtime support was reverted, so the action is not exposed as an executable write.',
        'source_url': source_url(action),
        'notes': 'No raw AWS action/path/body escape hatch is provided; this remains planned until typed write metadata is runtime-visible.'
    }

endpoints = []
for action in ops:
    row = {'method': 'POST', 'path': f'cloudtrail_api_action:{action}'}
    if action in STREAMS:
        row['method'] = 'READ_POST'
        row['covered_by'] = {'stream': snake(action)}
    else:
        row['operation'] = blocked_operation(action)
    endpoints.append(row)
api_surface = {
    'api': 'AWS CloudTrail API Reference actions',
    'docs': 'https://docs.aws.amazon.com/awscloudtrail/latest/APIReference/API_Operations.html',
    'reviewed_at': '2026-07-31',
    'operation_ledger_version': 1,
    'scope': 'Complete AWS CloudTrail API action inventory from the official Actions page. Scope-corrected surface implements 19 ETL/read streams through the connector-local runtime using fixed action mappings and resource discovery/fan-out where an AWS read action requires identifiers. The 10 provider query/direct-read actions and 31 write/admin actions are blocked/planned because their safe exposure requires shared promoted-native forwarding outside this connector-local slice. Event record fields remain payload schema evidence, not CDC operations.',
    'endpoints': endpoints
}
json_dump(DEF/'api_surface.json', api_surface)

cert = {'schema_version': 1, 'source': {'default_stream': 'describe_trails', 'live_unavailable': [{'kind': 'no_credentials', 'contains': ['aws-cloudtrail connector requires secrets aws_key_id and aws_secret_key']}]} }
json_dump(DEF/'certification.json', cert)

# Fixtures.
json_dump(DEF/'fixtures/check.json', {'request': {'method': 'POST', 'path': '/', 'query': {}}, 'response': {'status': 200, 'body': {'trailList': []}}})

def response_body_for_stream(action: str):
    stream = snake(action)
    row = ops[action]
    base_obj = {'pm_record_id': f'{stream}_fixture_1', 'operation': action}
    # Include common documented response fields with synthetic values.
    for f in row['response_fields']:
        base_obj[f['name']] = sample_value({'name': f['name'], 'aws_type': f.get('aws_type','String')})
    if action == 'DescribeTrails':
        return {'trailList': [base_obj]}
    if action.startswith('List'):
        # Pick the first array response field if one exists.
        arr = None
        for f in row['response_fields']:
            if 'array' in f.get('aws_type','').lower():
                arr = f['name']; break
        if arr:
            item = dict(base_obj)
            item.pop(arr, None)
            return {arr: [item]}
    return base_obj

for action in STREAMS:
    stream = snake(action)
    read_query = dict(STREAM_READ_QUERY_OVERRIDES.get(action, {}))
    read_query.update({f['name']: str(sample_value(f)) if not isinstance(sample_value(f), (dict, list)) else json.dumps(sample_value(f)) for f in ops[action]['request_fields'] if f['required']})
    page = {'request': {'method': 'POST', 'path': '/', 'query': {}}, 'response': {'status': 200, 'body': response_body_for_stream(action)}}
    if read_query:
        page['read_query'] = read_query
    json_dump(DEF/f'fixtures/streams/{stream}/page_1.json', page)

# docs.md
write_list = ', '.join(f'`{a}`' for a in WRITES)
stream_list = ', '.join(f'`{snake(a)}`' for a in STREAMS)
direct_list = ', '.join(f'`{a}`' for a in DIRECT)
docs = f"""# Overview

AWS CloudTrail connector parity was audited from the official AWS CloudTrail API Reference Actions page. The scope-corrected bundle still enumerates all 60 official CloudTrail API actions exactly once, and 19 ETL/read stream actions are implemented and runtime-reachable in this connector-local slice. Streams whose AWS read action requires an identifier use connector-local discovery/fan-out from the fixed list/describe streams rather than caller-supplied raw action/path/header/body input. The 10 provider query/direct-read actions and 31 write/admin actions are recorded as blocked/planned in `api_surface.json` because they require shared promoted-native command-surface, manifest, validation, dry-run, and operation-direct-read forwarding that was reverted from this branch. The official CloudTrail event-record contents page documents event record version 1.11 and 31 top-level event fields; those fields are schema payload fields for LookupEvents, not CDC/changefeed operations.

Implemented readable streams: {stream_list}.

Blocked/planned provider query operations: {direct_list}.

Blocked/planned write/admin operations: {write_list}.

## Auth setup

Use `pm credentials add <name> --connector aws-cloudtrail` and provide secrets only from environment variables or stdin:

- `aws_key_id` (required secret)
- `aws_secret_key` (required secret)
- `aws_region_name` (required config, for example `us-east-1`)

Optional config fields are `base_url` for local fixture endpoints, `page_size`, `max_pages`, and `mode=fixture` for credential-free tests. Do not place AWS secret values in chat, command history, docs, fixtures, or issue comments.

## Streams notes

Every implemented stream uses a fixed AWS CloudTrail JSON-RPC action with SigV4 authentication and no raw action/path/header/body escape hatch. Paginated actions pass bounded `MaxResults` and follow `NextToken` until it is absent or `max_pages` is reached. Streams for resource-detail actions derive required identifiers through connector-local discovery/fan-out from `DescribeTrails`, `ListChannels`, `ListDashboards`, `ListEventDataStores`, and `ListImports`. The connector keeps `query=false` at the metadata layer; CloudTrail provider query operations are blocked/planned until shared runtime forwarding makes operation-direct reads visible safely.

## Write actions & risks

No CloudTrail reverse-ETL write actions are exposed by the scope-corrected runtime surface. The audited write/admin API actions remain blocked/planned in `api_surface.json`; they must not be documented as executable until shared promoted-native forwarding exposes bundle manifests, command surfaces, write validation, and dry-run previews. Any future write slice must preserve the standard plan -> preview -> explicit approval -> execute flow, destructive confirmation metadata, fixed CloudTrail `X-Amz-Target` mapping, and no raw AWS action or request-body escape hatch.

## Known limits

- This work is fixture-only and local-test verified; it does not certify live AWS provider behavior.
- CloudTrail event record fields are parsed as payload/schema fields where CloudTrail returns them, but they are not counted as CDC because AWS does not document a CloudTrail changefeed/subscription API in the audited sources.
- Provider query direct reads are blocked/planned. They do not expose a generic CloudTrail Lake SQL runner.
- Resource-detail streams depend on the corresponding list/describe action returning identifiers; when no resources are discovered they emit no records.
- Write/admin actions are blocked/planned. No CloudTrail write action is listed as executable in generated catalog/docs/help for this corrective head.
"""
(DEF/'docs.md').write_text(docs)

print('generated aws-cloudtrail bundle:', len(STREAMS), 'streams', len(DIRECT) + len(WRITES), 'blocked operations')
