#!/usr/bin/env python3
import json
import re
import shutil
from pathlib import Path

# Historical generator for the initial 60/19/10/31 implementation. The final
# scope-corrected commit intentionally post-processes its output to expose only
# 19 implemented streams and to mark the 10 direct-read plus 31 write/admin
# actions blocked/planned until shared promoted-native forwarding lands.
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
    'description': 'Reads AWS CloudTrail configuration and event metadata, runs bounded CloudTrail query/lookups, and executes typed approval-gated CloudTrail administration actions through fixed AWS JSON-RPC operations.',
    'integration_type': 'api',
    'release_stage': 'alpha',
    'capabilities': {'check': True, 'read': True, 'write': True, 'query': False, 'cdc': False, 'dynamic_schema': False},
    'batch': {'read_page_size': 50, 'write_batch_size': 1},
    'risk': {
        'read': 'bounded AWS CloudTrail JSON-RPC reads using fixed action names and SigV4 authentication',
        'write': 'typed CloudTrail trail, channel, dashboard, event-data-store, import, logging, federation, selector, tagging, and resource-policy actions only',
        'approval': 'reverse ETL writes require plan, preview, explicit approval, and destructive confirmation where declared'
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
        'start_date': {'type': 'string', 'description': 'Optional lower bound for LookupEvents-style reads; YYYY-MM-DD or RFC3339.'},
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
for action in WRITES:
    name = snake(action)
    kind = 'delete' if action.startswith('Delete') else 'update'
    if action.startswith('Create') or action in ('AddTags','RegisterOrganizationDelegatedAdmin','StartImport'):
        kind = 'create'
    confirm = 'destructive' if action.startswith(('Delete','Deregister','Disable','Remove','Stop')) or action in ('PutResourcePolicy','RegisterOrganizationDelegatedAdmin','StartLogging','StartDashboardRefresh') else ''
    entry = {
        'name': name,
        'kind': kind,
        'method': 'POST',
        'path': '/',
        'body_type': 'json',
        'record_schema': write_schema(action),
        'risk': f'Executes AWS CloudTrail {action} through a fixed typed JSON-RPC action; no raw action or path is accepted.'
    }
    if confirm:
        entry['confirm'] = confirm
    # Redact policy and selector documents or identifiers that may appear in operator-visible errors.
    redacts = [f['name'] for f in ops[action]['request_fields'] if f['name'] in ('ResourcePolicy','QueryStatement','Prompt')]
    if redacts:
        entry['redact_fields'] = redacts
    if kind == 'delete':
        entry['delete'] = {'idempotent': True, 'missing_ok_status': [404]}
    writes['actions'].append(entry)
json_dump(DEF/'writes.json', writes)

operations = {'operations': []}
for action in DIRECT:
    body_schema = write_schema(action)
    operations['operations'].append({
        'id': op_id(action),
        'kind': 'rest_read',
        'summary': summary_for(action),
        'description': f'Bounded AWS CloudTrail {action} operation using a fixed JSON-RPC target and closed request schema.',
        'source_url': source_url(action),
        'risk': 'high' if action in ('CancelQuery','StartQuery','GenerateQuery') else 'medium',
        'approval': 'none: bounded CloudTrail provider query/lookup operation with a closed request schema and redacted JSON output',
        'output_policy': 'json_redacted',
        'rest': {'method': 'POST', 'path': '/', 'content_type': 'application/json', 'max_bytes': 1048576, 'body_schema': body_schema}
    })
json_dump(DEF/'operations.json', operations)

# API surface rows in official action order.
endpoints = []
for action in ops:
    row = {'method': 'POST', 'path': f'cloudtrail_api_action:{action}'}
    if action in STREAMS:
        row['covered_by'] = {'stream': snake(action)}
    elif action in DIRECT:
        row['covered_by'] = {'direct_read': direct_command_path(action)}
    else:
        row['covered_by'] = {'write': snake(action)}
    endpoints.append(row)
api_surface = {
    'api': 'AWS CloudTrail API Reference actions',
    'docs': 'https://docs.aws.amazon.com/awscloudtrail/latest/APIReference/API_Operations.html',
    'reviewed_at': '2026-07-31',
    'operation_ledger_version': 1,
    'scope': 'Complete AWS CloudTrail API action inventory from the official Actions page. The official CloudTrail event-record contents page documents 31 top-level event record fields for LookupEvents payloads; those are schema fields, not separate API operations or CDC feeds.',
    'endpoints': endpoints
}
json_dump(DEF/'api_surface.json', api_surface)

# CLI surface: implemented ETL and direct read commands; reverse ETL writes are available through generic pm reverse plan/preview/run using writes.json actions.
commands = []
for action in STREAMS:
    row = ops[action]
    commands.append({
        'path': stream_command_path(action),
        'summary': summary_for(action),
        'intent': 'etl',
        'availability': 'implemented',
        'stream': snake(action),
        'source_url': source_url(action),
        'flags': [flag_for(f, 'query') for f in row['request_fields']],
        'examples': [f'pm aws-cloudtrail {stream_command_path(action)} --credential cloudtrail --json'],
        'notes': 'Runs a fixed CloudTrail JSON-RPC read action; request fields are allow-listed command flags.'
    })
for action in DIRECT:
    row = ops[action]
    commands.append({
        'path': direct_command_path(action),
        'summary': summary_for(action),
        'intent': 'direct_read',
        'availability': 'implemented',
        'operation': op_id(action),
        'source_url': source_url(action),
        'flags': [flag_for(f, 'body') for f in row['request_fields']],
        'examples': [f'pm aws-cloudtrail {direct_command_path(action)} --credential cloudtrail --json'],
        'output_policy': 'json_redacted',
        'redact_fields': ['CloudTrailEvent','AccessKeyId','userIdentity','requestParameters','responseElements','additionalEventData'],
        'risk': 'high' if action in ('CancelQuery','StartQuery','GenerateQuery') else 'medium',
        'approval': 'none: typed bounded CloudTrail provider query/lookup operation',
        'notes': 'No raw query text escape hatch is exposed beyond the documented CloudTrail QueryStatement field on StartQuery.'
    })
cli_surface = {
    'tagline': 'AWS CloudTrail fixed-action reads, provider queries, and approval-gated administration actions.',
    'usage': 'pm aws-cloudtrail <command> [flags] --credential <name> [--json]',
    'source_cli': {'name': 'AWS CloudTrail API', 'docs': 'https://docs.aws.amazon.com/awscloudtrail/latest/APIReference/API_Operations.html', 'reference': 'https://docs.aws.amazon.com/awscloudtrail/latest/APIReference/API_Operations.html', 'source': 'official AWS documentation'},
    'groups': [
        {'id': 'read', 'title': 'ETL read streams', 'commands': [stream_command_path(a) for a in STREAMS]},
        {'id': 'query', 'title': 'Bounded query and lookup operations', 'commands': [direct_command_path(a) for a in DIRECT]},
        {'id': 'write', 'title': 'Reverse ETL writes', 'commands': []}
    ],
    'commands': commands,
    'help_topics': [
        {'name': 'aws-cloudtrail', 'summary': 'AWS CloudTrail connector command surface and safety contract.'},
        {'name': 'aws-cloudtrail writes', 'summary': 'Use pm reverse plan/preview/run with aws-cloudtrail write actions; direct provider shortcuts are intentionally not exposed.'}
    ]
}
json_dump(DEF/'cli_surface.json', cli_surface)

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
    read_query = {f['name']: str(sample_value(f)) if not isinstance(sample_value(f), (dict, list)) else json.dumps(sample_value(f)) for f in ops[action]['request_fields'] if f['required']}
    page = {'request': {'method': 'POST', 'path': '/', 'query': {}}, 'response': {'status': 200, 'body': response_body_for_stream(action)}}
    if read_query:
        page['read_query'] = read_query
    json_dump(DEF/f'fixtures/streams/{stream}/page_1.json', page)

for action in WRITES:
    name = snake(action)
    rec = sample_record(action)
    fx = {'record': rec, 'expect': {'method': 'POST', 'path': '/', 'body': rec}, 'response': {'status': 200, 'body': {}}}
    json_dump(DEF/f'fixtures/writes/{name}.json', fx)

for action in DIRECT:
    rec = sample_record(action)
    fx = {'operation': op_id(action), 'request': {'method': 'POST', 'path': '/', 'body': rec}, 'response': {'status': 200, 'body': {'operation': action, 'pm_record_id': snake(action)+'_fixture_1'}}}
    json_dump(DEF/f'fixtures/direct_reads/{snake(action)}.json', fx)

# docs.md
write_list = ', '.join(f'`{snake(a)}`' for a in WRITES)
stream_list = ', '.join(f'`{snake(a)}`' for a in STREAMS)
direct_list = ', '.join(f'`{direct_command_path(a)}`' for a in DIRECT)
docs = f"""# Overview

AWS CloudTrail connector parity is implemented from the official AWS CloudTrail API Reference Actions page fetched on 2026-07-31. The bundle enumerates 60 official CloudTrail API actions exactly once: 19 ETL/read streams, 10 bounded provider query/lookup direct reads, and 31 typed reverse-ETL write actions. The official CloudTrail event-record contents page documents event record version 1.11 and 31 top-level event fields; those fields are schema payload fields for LookupEvents, not CDC/changefeed operations.

Readable streams: {stream_list}.

Bounded direct/provider query commands: {direct_list}.

Typed write actions: {write_list}.

## Auth setup

Use `pm credentials add <name> --connector aws-cloudtrail` and provide secrets only from environment variables or stdin:

- `aws_key_id` (required secret)
- `aws_secret_key` (required secret)
- `aws_region_name` (required config, for example `us-east-1`)

Optional config fields are `base_url` for local fixture endpoints, `page_size`, `max_pages`, `start_date`, and `mode=fixture` for credential-free tests. Do not place AWS secret values in chat, command history, docs, fixtures, or issue comments.

## Streams notes

Every stream uses a fixed AWS CloudTrail JSON-RPC action with SigV4 authentication and no raw action/path/header/body escape hatch. Paginated actions pass bounded `MaxResults` and follow `NextToken` until it is absent or `max_pages` is reached. Streams whose official request schema has required fields expose those fields through definition-owned command flags and fixture `read_query` values. The connector keeps `query=false` at the metadata layer; CloudTrail provider query operations are modeled as typed direct reads, not as warehouse SQL.

## Write actions & risks

Reverse ETL writes are declared in `writes.json` as closed top-level AWS request schemas. Runtime execution remains the standard `pm reverse` flow: plan -> preview -> explicit approval -> execute. Destructive/admin operations such as delete, stop, disable, resource-policy, delegated-admin, and logging controls declare destructive confirmation metadata. The native executor maps each write action to one fixed CloudTrail `X-Amz-Target`; operators cannot supply arbitrary AWS action names or raw request bodies. Delete actions are treated idempotently for provider missing-resource HTTP 404 responses in fixture and local replay.

## Known limits

- This work is fixture-only and local-test verified; it does not certify live AWS provider behavior.
- CloudTrail event record fields are parsed as payload/schema fields where CloudTrail returns them, but they are not counted as CDC because AWS does not document a CloudTrail changefeed/subscription API in the audited sources.
- Provider query direct reads are bounded JSON responses with redaction. They do not expose a generic CloudTrail Lake SQL runner; only the documented `StartQuery` `QueryStatement` field exists as a typed, fixed-target operation.
- Nested AWS structures are typed as closed top-level fields with object/array payloads because the official action pages link many nested reusable shapes; this bundle does not add raw generic sub-body passthrough beyond those documented fields.
"""
(DEF/'docs.md').write_text(docs)

print('generated aws-cloudtrail bundle:', len(STREAMS), 'streams', len(DIRECT), 'direct ops', len(WRITES), 'writes')
