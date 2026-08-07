#!/usr/bin/env node
// Generates Mailchimp connector definition parity files from the current official
// Mailchimp Marketing Swagger root and provider-owned path refs. Public docs only;
// no live provider data, credentials, or writes.

import fs from 'node:fs';
import path from 'node:path';

const ROOT_URL = 'https://api.mailchimp.com/schema/3.0/Swagger.json';
const OUT = 'internal/connectors/defs/mailchimp';
const TRACE = '.planning/phases/mailchimp-parity-wave03-r1/traces';
const METHODS = new Set(['get', 'post', 'put', 'patch', 'delete']);
const TODAY = '2026-07-31';

const cache = new Map();
async function fetchJSON(url) {
  if (cache.has(url)) return cache.get(url);
  const res = await fetch(url, { headers: { 'user-agent': 'polymetrics-mailchimp-parity-generator' } });
  if (!res.ok) throw new Error(`${res.status} ${res.statusText} ${url}`);
  const json = await res.json();
  cache.set(url, json);
  return json;
}

function snake(input) {
  return String(input || '')
    .replace(/([a-z0-9])([A-Z])/g, '$1_$2')
    .replace(/[^A-Za-z0-9]+/g, '_')
    .replace(/_+/g, '_')
    .replace(/^_+|_+$/g, '')
    .toLowerCase();
}

function kebab(input) {
  return snake(input).replace(/_/g, '-');
}

function uniqueName(base, used) {
  let name = base || 'operation';
  let i = 2;
  while (used.has(name)) name = `${base}_${i++}`;
  used.add(name);
  return name;
}

function pathVars(p) {
  return [...String(p).matchAll(/\{([^}]+)\}/g)].map(m => m[1]);
}

function streamPath(p) {
  return p.replace(/\{([^}]+)\}/g, (_, v) => `{{ config.${v} }}`);
}

function resourceSegments(p) {
  return p.split('/').filter(Boolean).filter(seg => !seg.startsWith('{')).map(snake);
}

function streamNameFor(p, recordPath, used) {
  let segs = resourceSegments(p);
  if (recordPath && recordPath !== '.' && !segs.includes(snake(recordPath))) segs.push(snake(recordPath));
  return uniqueName(segs.join('_') || 'root', used);
}

function actionNameFor(method, p, operationId, used) {
  const m = method.toLowerCase();
  const segs = resourceSegments(p);
  const last = segs[segs.length - 1] || 'resource';
  let base;
  if (segs.includes('actions')) {
    const idx = segs.indexOf('actions');
    const act = segs[idx + 1] || last;
    const parent = segs[idx - 1] || segs[0] || 'resource';
    base = `${act}_${parent}`;
  } else if (m === 'post') {
    base = `post_${segs.join('_')}`;
  } else if (m === 'patch') {
    base = `patch_${segs.join('_')}`;
  } else if (m === 'put') {
    base = `put_${segs.join('_')}`;
  } else if (m === 'delete') {
    base = `delete_${segs.join('_')}`;
  } else {
    base = snake(operationId || `${m}_${segs.join('_')}`);
  }
  return uniqueName(base, used);
}

function commandPathFor(p, suffix = 'get') {
  const segs = resourceSegments(p).map(kebab);
  const trimmed = segs.slice(0, 4); // keep help paths bounded but meaningful
  if (trimmed[trimmed.length - 1] !== suffix) trimmed.push(suffix);
  return trimmed.join(' ');
}

function simplifySchemaType(schema) {
  if (!schema) return { type: ['string', 'integer', 'number', 'boolean', 'object', 'array', 'null'] };
  if (schema.enum) return { type: ['string', 'null'] };
  if (schema.type) {
    if (schema.type === 'integer') return { type: ['integer', 'null'] };
    if (schema.type === 'number') return { type: ['number', 'integer', 'null'] };
    if (schema.type === 'boolean') return { type: ['boolean', 'null'] };
    if (schema.type === 'array') return { type: ['array', 'null'], items: { type: ['string', 'integer', 'number', 'boolean', 'object', 'array', 'null'] } };
    if (schema.type === 'object') return { type: ['object', 'null'] };
    return { type: [schema.type, 'null'] };
  }
  if (schema.$ref || schema.properties) return { type: ['object', 'null'] };
  return { type: ['string', 'integer', 'number', 'boolean', 'object', 'array', 'null'] };
}

async function deref(schema) {
  if (!schema) return schema;
  if (schema.$ref) return fetchJSON(schema.$ref);
  return schema;
}

async function collectionInfo(op) {
  const resp = op.responses && (op.responses['200'] || op.responses[200]);
  if (!resp || !resp.schema) return null;
  const responseSchema = await deref(resp.schema);
  const props = responseSchema?.properties || {};
  // Mailchimp collection envelopes consistently expose total_items alongside
  // the record array. Single-object GET responses can also contain array
  // fields (tags, variants, questions, lines, etc.); those must stay typed
  // direct reads so the connector does not silently drop the parent resource.
  if (!props.total_items) return null;
  for (const [name, prop] of Object.entries(props)) {
    if (name === '_links') continue;
    const p = await deref(prop);
    if (p?.type === 'array' || p?.items) {
      const itemSchema = await deref(p.items || {});
      return { recordsPath: name, itemSchema: itemSchema || {}, responseSchema: responseSchema || {} };
    }
  }
  return null;
}

async function bodySchemaFor(op) {
  for (const rawParam of op.parameters || []) {
    const param = rawParam.$ref ? await fetchJSON(rawParam.$ref) : rawParam;
    if (param.in === 'body' && param.schema) return deref(param.schema);
  }
  return null;
}

async function queryParamsFor(op) {
  const out = [];
  for (const rawParam of op.parameters || []) {
    const param = rawParam.$ref ? await fetchJSON(rawParam.$ref) : rawParam;
    if (param.in === 'query' && param.name) out.push(param);
  }
  return out;
}

function pickPK(props, vars) {
  for (const key of ['id', 'web_id', 'email_id', 'campaign_id', 'list_id', 'store_id', 'subscriber_hash', 'email_address', 'name', 'domain', 'month', 'url', 'ip']) {
    if (props[key]) return [key];
  }
  if (vars.length > 0) return [vars[0]];
  return ['id'];
}

function cursorFor(props) {
  for (const [field, param] of [
    ['date_created', 'since_date_created'],
    ['create_time', 'since_create_time'],
    ['send_time', 'since_send_time'],
    ['last_changed', 'since_last_changed'],
    ['updated_at', 'since_updated_at']
  ]) {
    if (props[field]) return { field, param };
  }
  return null;
}

function schemaForStream(name, itemSchema, vars) {
  const rawProps = itemSchema.properties || {};
  const props = {};
  for (const [key, value] of Object.entries(rawProps)) {
    props[key] = simplifySchemaType(value);
  }
  for (const v of vars) props[v] = { type: ['string', 'null'] };
  const pk = pickPK(props, vars);
  for (const p of pk) if (!props[p]) props[p] = { type: ['string', 'integer', 'null'] };
  const schema = {
    '$schema': 'http://json-schema.org/draft-07/schema#',
    title: name,
    type: 'object',
    'x-primary-key': pk,
    properties: props
  };
  const cursor = cursorFor(props);
  if (cursor) schema['x-cursor-field'] = cursor.field;
  return { schema, cursor };
}

function sampleValue(field, schema) {
  const f = field.toLowerCase();
  const types = Array.isArray(schema?.type) ? schema.type : [schema?.type || 'string'];
  // Honor the JSON type before field-name heuristics so provider fields such
  // as email_channel (object), test_emails (array), and update_existing
  // (boolean) produce schema-valid sanitized fixtures.
  if (types.includes('array')) return [];
  if (types.includes('object')) return {};
  if (types.includes('boolean')) return true;
  if (types.includes('integer')) return 1;
  if (types.includes('number')) return 1.5;
  if (f.includes('email')) return 'fixture@example.invalid';
  if (f.includes('url')) return 'https://example.invalid/mailchimp-fixture';
  if (f.includes('date') || f.endsWith('_time') || f.includes('timestamp')) return '2026-01-01T00:00:00Z';
  return `fixture_${field}`;
}

function streamFixtureBody(recordPath, schema, streamName, vars, page = 1) {
  const record = {};
  for (const pk of schema['x-primary-key'] || ['id']) record[pk] = sampleValue(pk, schema.properties?.[pk] || { type: 'string' });
  for (const v of vars) record[v] = `fixture_${v}`;
  for (const cursor of [schema['x-cursor-field']].filter(Boolean)) record[cursor] = `2026-01-0${page}T00:00:00Z`;
  for (const [k, v] of Object.entries(schema.properties || {})) {
    if (record[k] === undefined && Object.keys(record).length < 8) record[k] = sampleValue(k, v);
  }
  const body = { total_items: page === 1 ? 2 : 0 };
  body[recordPath] = page === 1 ? [record] : [];
  return body;
}

function recordSchemaForWrite(method, p, bodySchema, destructive) {
  const vars = pathVars(p);
  const properties = {};
  for (const v of vars) properties[v] = { type: 'string' };
  const bodyFields = [];
  const bodyProps = bodySchema?.properties || {};
  for (const [k, raw] of Object.entries(bodyProps)) {
    properties[k] = simplifySchemaType(raw);
    bodyFields.push(k);
  }
  if (bodyFields.length === 0 && !['delete'].includes(method.toLowerCase()) && !/\/actions\//.test(p)) {
    properties.payload = { type: ['object', 'string', 'null'], description: 'Typed Mailchimp request payload field for provider-defined optional content.' };
    bodyFields.push('payload');
  }
  const schema = {
    type: 'object',
    additionalProperties: false,
    required: vars,
    properties
  };
  if (vars.length === 0 && bodyFields.length > 0) schema.minProperties = 1;
  if (destructive) schema.description = 'Closed schema for approval-gated Mailchimp mutation.';
  return { schema, pathFields: vars, bodyFields };
}

function writeRecordAndBody(action, schemaInfo) {
  const record = {};
  for (const f of schemaInfo.pathFields) record[f] = `fixture_${f}`;
  for (const f of schemaInfo.bodyFields.slice(0, 3)) {
    if (record[f] === undefined) record[f] = sampleValue(f, schemaInfo.schema.properties[f]);
  }
  if (Object.keys(record).length === 0) record.payload = 'fixture_payload';
  const body = {};
  for (const [k, v] of Object.entries(record)) if (!schemaInfo.pathFields.includes(k)) body[k] = v;
  return { record, body };
}

function concreteWritePath(p, record) {
  return p.replace(/\{([^}]+)\}/g, (_, v) => encodeURIComponent(record[v] || `fixture_${v}`));
}

function isDestructive(method, p, summary) {
  if (method.toUpperCase() === 'DELETE') return true;
  return /delete|archive|forget|send|publish|unpublish|cancel|pause|start|trigger|verify/i.test(`${p} ${summary}`);
}

function operationRow(model, risk, reason, sourceURL, notes) {
  const row = { model, status: 'blocked', risk, blocked_by_default: true, reason };
  if (sourceURL) row.source_url = sourceURL;
  if (notes) row.notes = notes;
  return row;
}

function cleanDir(dir) {
  fs.rmSync(dir, { recursive: true, force: true });
  fs.mkdirSync(dir, { recursive: true });
}

function writeJSON(file, value) {
  fs.mkdirSync(path.dirname(file), { recursive: true });
  fs.writeFileSync(file, JSON.stringify(value, null, 2) + '\n');
}

function writeText(file, value) {
  fs.mkdirSync(path.dirname(file), { recursive: true });
  fs.writeFileSync(file, value);
}

async function main() {
  const root = await fetchJSON(ROOT_URL);
  const pathSpecs = [];
  for (const [p, entry] of Object.entries(root.paths || {})) {
    const spec = entry.$ref ? await fetchJSON(entry.$ref) : entry;
    pathSpecs.push({ path: p, ref: entry.$ref || ROOT_URL, spec });
  }

  const operations = [];
  for (const ps of pathSpecs) {
    for (const [method, op] of Object.entries(ps.spec)) {
      if (!METHODS.has(method)) continue;
      operations.push({ method: method.toUpperCase(), lower: method, path: ps.path, ref: ps.ref, op });
    }
  }
  operations.sort((a, b) => a.path === b.path ? a.method.localeCompare(b.method) : a.path.localeCompare(b.path));

  cleanDir(path.join(OUT, 'schemas'));
  cleanDir(path.join(OUT, 'fixtures'));

  const specProps = {
    access_token: { type: 'string', 'x-secret': true, description: 'Mailchimp OAuth access token. Takes precedence over api_key when both are set. Sent as Bearer auth; never logged.' },
    api_key: { type: 'string', 'x-secret': true, description: 'Mailchimp API key. Sent as HTTP Basic auth (username "anystring") when access_token is unset; never logged.' },
    data_center: { type: 'string', description: 'Mailchimp datacenter token (for example, us6) used to build https://<data_center>.api.mailchimp.com/3.0.' },
    start_date: { type: 'string', format: 'date-time', description: 'Optional RFC3339 lower bound used by stream definitions that expose Mailchimp since_* filters.' },
    mode: { type: 'string', description: 'Runtime mode: live (default) or fixture for credential-free conformance.' },
    search_query: { type: 'string', description: 'Optional default search term for typed search direct-read commands.' }
  };
  for (const op of operations) for (const v of pathVars(op.path)) specProps[v] ||= { type: 'string', description: `Optional Mailchimp path identifier for {${v}} operations and nested streams.` };

  const streams = [];
  const streamSchemas = new Map();
  const writes = [];
  const directOps = [];
  const cliCommands = [];
  const apiEndpoints = [];
  const usedStreams = new Set();
  const usedWrites = new Set();
  const usedCommands = new Set();

  for (const entry of operations) {
    const { method, lower, path: endpointPath, op, ref } = entry;
    const summary = (op.summary || '').replace(/\s+/g, ' ').trim();
    if (method === 'GET' && (endpointPath === '/' || endpointPath === '/ping')) {
      apiEndpoints.push({ method, path: endpointPath, operation: operationRow('local_workflow', 'low', 'Provider metadata/health endpoint; `pm connectors inspect` and `pm etl check` cover local discovery/health without exposing this as a data operation.', ref) });
      continue;
    }

    if (method === 'POST' && endpointPath === '/batches') {
      apiEndpoints.push({ method, path: endpointPath, operation: operationRow('disallowed', 'high', 'Mailchimp batch creation accepts arbitrary method/path/body operations; exposing it would be a generic HTTP write passthrough, which AGENTS.md forbids.', ref, 'Use named stream, direct-read, or reverse-ETL actions instead of raw batch operation submission.') });
      continue;
    }

    if (method === 'GET') {
      const info = await collectionInfo(op);
      const isSearch = endpointPath === '/search-members' || endpointPath === '/search-campaigns' || /tag-search/.test(endpointPath);
      if (info && !isSearch) {
        const vars = pathVars(endpointPath);
        const name = streamNameFor(endpointPath, info.recordsPath, usedStreams);
        const { schema, cursor } = schemaForStream(name, info.itemSchema, vars);
        streamSchemas.set(name, schema);
        const stream = {
          name,
          path: streamPath(endpointPath),
          records: { path: info.recordsPath },
          query: {},
          computed_fields: {},
          schema: `schemas/${name}.json`
        };
        if (Object.keys(stream.query).length === 0) delete stream.query;
        for (const v of vars) stream.computed_fields[v] = `{{ config.${v} }}`;
        if (Object.keys(stream.computed_fields).length === 0) delete stream.computed_fields;
        if (cursor && ['lists', 'campaigns', 'reports', 'automations'].includes(name)) {
          stream.incremental = { cursor_field: cursor.field, request_param: cursor.param, param_format: 'rfc3339', start_config_key: 'start_date' };
        }
        streams.push(stream);
        apiEndpoints.push({ method, path: endpointPath, covered_by: { stream: name } });
        const fixtureDir = path.join(OUT, 'fixtures', 'streams', name);
        const q = { count: '100' };
        if (stream.incremental) q[stream.incremental.request_param] = '2020-01-01T00:00:00Z';
        writeJSON(path.join(fixtureDir, 'page_1.json'), { request: { method: 'GET', path: concreteWritePath(endpointPath, Object.fromEntries(vars.map(v => [v, 'synthetic-conformance-value']))), query: q }, response: { status: 200, body: streamFixtureBody(info.recordsPath, schema, name, vars, 1) } });
        if (name === 'lists') {
          writeJSON(path.join(fixtureDir, 'page_2.json'), { request: { method: 'GET', path: '/lists', query: { count: '100', offset: '100' } }, response: { status: 200, body: streamFixtureBody(info.recordsPath, schema, name, vars, 2) } });
        }
        cliCommands.push({ path: uniqueName(`${kebab(name)} list`, usedCommands).replace(/_/g, '-'), summary: `${summary || `Read ${name}`} as ETL records.`, intent: 'etl', availability: 'implemented', stream: name, api_surface: [{ method, path: endpointPath }], examples: [`pm mailchimp ${kebab(name)} list --json --limit 25`] });
        continue;
      }

      const opID = `mailchimp.${snake(op.operationId || `${method}_${endpointPath}`)}`;
      const cmdPath = uniqueName(commandPathFor(endpointPath, 'get'), usedCommands).replace(/_/g, '-');
      const flags = [];
      for (const v of pathVars(endpointPath)) flags.push({ name: kebab(v), type: 'string', summary: `Mailchimp ${v} path identifier.`, maps_to: `path.${v}`, allow_empty: false });
      if (isSearch) flags.push({ name: 'query', type: 'string', summary: 'Search query text.', maps_to: 'query.query', allow_empty: false });
      directOps.push({ id: opID, kind: 'rest_read', summary: summary || `${method} ${endpointPath}`, description: `Bounded Mailchimp direct read for ${method} ${endpointPath}.`, source_url: ref, risk: isSearch ? 'medium' : 'low', approval: 'none: read-only provider operation with connector-scoped endpoint, typed flags, bounded response bytes, and JSON redaction', output_policy: 'json_redacted', rest: { method, path: endpointPath, max_bytes: 1048576 } });
      cliCommands.push({ path: cmdPath, summary: summary || `Read ${endpointPath}.`, intent: 'direct_read', availability: 'implemented', operation: opID, api_surface: [{ method, path: endpointPath }], flags, output_policy: 'json_redacted', examples: [`pm mailchimp ${cmdPath} --json`] });
      apiEndpoints.push({ method, path: endpointPath, covered_by: { direct_read: cmdPath } });
      continue;
    }

    const destructive = isDestructive(method, endpointPath, summary);
    const bodySchema = await bodySchemaFor(op);
    const actionName = actionNameFor(lower, endpointPath, op.operationId, usedWrites);
    const schemaInfo = recordSchemaForWrite(lower, endpointPath, bodySchema, destructive);
    const action = {
      name: actionName,
      kind: lower === 'delete' ? 'delete' : (lower === 'patch' ? 'update' : (lower === 'put' ? 'upsert' : 'custom')),
      method,
      path: streamPath(endpointPath).replace(/config\./g, 'record.'),
      path_fields: schemaInfo.pathFields,
      body_fields: schemaInfo.bodyFields,
      record_schema: schemaInfo.schema,
      risk: `${destructive ? 'Destructive or externally visible' : 'Externally visible'} Mailchimp mutation: ${summary || `${method} ${endpointPath}`}. Reverse ETL must plan, preview, receive explicit approval, and then execute.`
    };
    if (schemaInfo.pathFields.length === 0) delete action.path_fields;
    if (schemaInfo.bodyFields.length === 0 || lower === 'delete' || /\/actions\//.test(endpointPath)) delete action.body_fields;
    if (lower === 'delete' || /\/actions\//.test(endpointPath)) action.body_type = 'none';
    if (lower === 'delete') action.delete = { idempotent: true, missing_ok_status: [404] };
    if (destructive) action.confirm = 'destructive';
    const redact = [];
    for (const k of Object.keys(schemaInfo.schema.properties || {})) if (/email|subscriber|contact|file_data|url|domain|phone/i.test(k)) redact.push(k);
    if (redact.length > 0) action.redact_fields = [...new Set(redact)];
    writes.push(action);
    apiEndpoints.push({ method, path: endpointPath, covered_by: { write: actionName } });
    const { record, body } = writeRecordAndBody(action, schemaInfo);
    const expect = { method, path: concreteWritePath(endpointPath, record) };
    if (action.body_type !== 'none' && Object.keys(body).length > 0) expect.body = body;
    writeJSON(path.join(OUT, 'fixtures', 'writes', `${actionName}.json`), { record, expect, response: { status: method === 'POST' ? 201 : 200, body: { id: `fixture_${actionName}` } } });

    const flags = [];
    for (const reqField of [...(schemaInfo.schema.required || [])]) flags.push({ name: kebab(reqField), type: 'string', summary: `Required ${reqField} record field.`, maps_to: `record.${reqField}`, allow_empty: false });
    cliCommands.push({ path: uniqueName(`actions ${kebab(actionName)}`, usedCommands).replace(/_/g, '-'), summary: summary || `${method} ${endpointPath}`, intent: 'reverse_etl', availability: 'implemented', write: actionName, flags, risk: action.risk, approval: destructive ? 'reverse ETL plan -> preview -> explicit approval -> execute with destructive confirmation' : 'reverse ETL plan -> preview -> explicit approval -> execute', api_surface: [{ method, path: endpointPath }], examples: [`pm mailchimp actions ${kebab(actionName)} --preview --json`] });
  }

  for (const [name, schema] of streamSchemas) writeJSON(path.join(OUT, 'schemas', `${name}.json`), schema);

  writeJSON(path.join(OUT, 'metadata.json'), {
    name: 'mailchimp',
    display_name: 'Mailchimp',
    description: 'Reads Mailchimp Marketing API audiences, members, campaigns, reports, automations, templates, files, batches, webhooks, ecommerce, reporting, and related resources; exposes typed approval-gated Mailchimp mutations where the declarative engine can model the documented operation safely.',
    integration_type: 'api',
    docs_url: ROOT_URL,
    release_stage: 'ga',
    capabilities: { check: true, read: true, write: true, query: false, cdc: false, dynamic_schema: false },
    batch: { read_page_size: 100, write_batch_size: 1 },
    rate_limit: { requests_per_minute: 10 },
    risk: {
      read: 'external Mailchimp Marketing API reads using configured datacenter credentials; nested streams may require operation-specific identifier config',
      write: 'approval-gated Mailchimp mutations against audiences, campaigns, automations, templates, files, webhooks, ecommerce, and related resources',
      approval: 'reverse ETL writes require plan, preview, explicit approval, and destructive confirmation where declared'
    }
  });

  writeJSON(path.join(OUT, 'spec.json'), {
    '$schema': 'http://json-schema.org/draft-07/schema#',
    title: 'Mailchimp Connection Specification',
    type: 'object',
    required: ['data_center'],
    properties: Object.fromEntries(Object.entries(specProps).sort(([a], [b]) => a.localeCompare(b)))
  });

  writeJSON(path.join(OUT, 'streams.json'), {
    base: {
      url: 'https://{{ config.data_center }}.api.mailchimp.com/3.0',
      user_agent: 'polymetrics-go-cli',
      auth: [
        { mode: 'bearer', token: '{{ secrets.access_token }}', when: '{{ secrets.access_token }}' },
        { mode: 'basic', username: 'anystring', password: '{{ secrets.api_key }}', when: '{{ secrets.api_key }}' }
      ],
      pagination: { type: 'offset_limit', limit_param: 'count', offset_param: 'offset', page_size: 100, max_pages: 5 },
      check: { method: 'GET', path: '/lists', query: { count: '1' } },
      error_map: [
        { status: 401, hint: 'mailchimp credentials are missing or invalid; re-run pm credentials add mailchimp' },
        { status: 429, class: 'rate_limited', hint: 'mailchimp rate limit exceeded; retry later or lower the request rate' }
      ]
    },
    streams
  });

  writeJSON(path.join(OUT, 'writes.json'), { actions: writes });
  writeJSON(path.join(OUT, 'operations.json'), { operations: directOps });
  writeJSON(path.join(OUT, 'api_surface.json'), {
    api: 'Mailchimp Marketing API v3.0',
    docs: ROOT_URL,
    reviewed_at: TODAY,
    operation_ledger_version: 1,
    scope: 'Current official Mailchimp Marketing Swagger 2.0 root plus all 181 provider-owned path refs. Count unit is HTTP method plus normalized path. Executable rows are definition-owned streams, typed direct reads, or named reverse-ETL actions. Blocked rows are exact safety/runtime blockers, not advertised executable work.',
    endpoints: apiEndpoints
  });

  writeJSON(path.join(OUT, 'cli_surface.json'), {
    tagline: 'Inspect, read, search, and safely plan typed Mailchimp Marketing API operations.',
    usage: 'pm mailchimp <command> [flags]',
    source_cli: { name: 'Mailchimp Marketing API', docs: ROOT_URL, reference: 'Official Swagger 2.0 schema 3.0.91', source: 'provider_api' },
    groups: [
      { id: 'streams', title: 'ETL streams', commands: ['lists', 'campaigns', 'reports', 'automations', 'ecommerce', 'templates', 'file-manager'] },
      { id: 'direct-reads', title: 'Typed direct reads and search', commands: ['search-members', 'search-campaigns', 'reports', 'campaigns', 'lists'] },
      { id: 'actions', title: 'Approval-gated reverse ETL actions', commands: ['actions'] }
    ],
    global_flags: [
      { name: 'credential', type: 'string', summary: 'Credential name to use for the Mailchimp request.' },
      { name: 'connection', type: 'string', summary: 'Alias for --credential.' },
      { name: 'config', type: 'string_array', summary: 'Connector config override as key=value.' },
      { name: 'json', type: 'boolean', summary: 'Emit machine-readable JSON output.' },
      { name: 'limit', type: 'integer', summary: 'Maximum ETL records to emit.' },
      { name: 'max-bytes', type: 'integer', summary: 'Maximum direct-read response bytes; operations are capped by their metadata.' },
      { name: 'plan', type: 'string', summary: 'Execute an approved reverse-ETL plan by id.' },
      { name: 'preview', type: 'boolean', summary: 'Preview a reverse-ETL write command without making a network mutation.' },
      { name: 'approve', type: 'string', summary: 'Approval token required to execute a reverse-ETL plan.' },
      { name: 'confirm', type: 'string', summary: 'Typed confirmation challenge for destructive reverse-ETL writes.' }
    ],
    commands: cliCommands,
    help_topics: [
      { name: 'auth', summary: 'Use access_token or api_key from environment/stdin; never paste secrets into prompts or shell history.' },
      { name: 'safety', summary: 'Writes are reverse-ETL only: plan, preview, explicit approval, execute; destructive actions require confirmation.' }
    ]
  });

  writeJSON(path.join(OUT, 'fixtures', 'check.json'), { request: { method: 'GET', path: '/lists', query: { count: '1' } }, response: { status: 200, body: { lists: [], total_items: 0 } } });

  const counts = { streams: streams.length, writes: writes.length, direct_reads: directOps.length, blocked: apiEndpoints.filter(e => e.operation).length, endpoints: apiEndpoints.length };
  writeJSON(path.join(TRACE, 'mailchimp-generation-summary.json'), counts);

  const docs = `# Overview\n\nMailchimp uses the official Marketing API Swagger root and all provider-owned path refs from ${ROOT_URL}. This bundle models ${counts.endpoints} current official operations: ${counts.streams} ETL streams, ${counts.direct_reads} typed direct reads/search commands, ${counts.writes} named reverse-ETL write actions, and ${counts.blocked} blocked operation-ledger rows.\n\nThe connector covers audiences/lists and members, campaigns, automations and customer journeys, reports and reporting resources, templates, file-manager metadata, batches, batch webhooks, ecommerce resources, landing pages, connected sites, verified domains, conversations, account exports, authorized apps, Facebook ads, and typed provider search.\n\n## Auth setup\n\nConnection fields:\n\n- \`data_center\` (required): Mailchimp datacenter token such as \`us6\`; builds \`https://<data_center>.api.mailchimp.com/3.0\`.\n- \`access_token\` (optional secret): OAuth bearer token; preferred when present.\n- \`api_key\` (optional secret): API key used as HTTP Basic password when no bearer token is present.\n- \`start_date\` (optional): RFC3339 lower bound used by supported top-level incremental streams.\n- Optional identifier config such as \`list_id\`, \`campaign_id\`, \`workflow_id\`, \`subscriber_hash\`, \`template_id\`, \`store_id\`, and related path variables is used by nested streams and direct-read commands when a command flag does not supply the value.\n\nSecrets must be provided from environment variables or stdin through \`pm credentials add\`; never paste secret values into chat, docs, logs, or shell history.\n\n## Streams notes\n\nThe base reader uses bounded offset/count pagination with \`count=100\` and \`max_pages=5\` for fixture-safe and operator-bounded reads. Top-level \`lists\`, \`campaigns\`, \`reports\`, and \`automations\` retain incremental lower-bound parameters when a cursor or \`start_date\` is available. Nested streams are explicit ETL streams with schema projection and sanitized fixtures; they require the relevant identifier config when the stream path contains provider IDs.\n\nExecutable stream count: ${counts.streams}. Every executable stream has a sanitized fixture page; \`lists\` has a two-page fixture to prove offset pagination terminates.\n\n## Write actions & risks\n\nThis bundle declares ${counts.writes} named reverse-ETL actions. Each action has a closed \`record_schema\`, path fields for provider identifiers, risk text, redaction for sensitive identifier/content fields, and fixture-backed request-shape coverage. Destructive or externally irreversible actions such as deletes, archives, sends, publishes, pauses/starts, and triggers declare \`confirm: "destructive"\`. DELETE actions declare provider-idempotent 404 handling where supported by the HTTP delete semantics.\n\nReverse ETL remains: plan -> preview -> explicit approval -> execute. No action exposes a raw method/path/body escape hatch.\n\n## Known limits\n\n- \`POST /batches\` remains blocked because the official body is an arbitrary batch of method/path/body operations; exposing it would be a generic HTTP write passthrough forbidden by repo policy.\n- \`GET /\` and \`GET /ping\` remain blocked as local metadata/health workflows; \`pm connectors inspect\` and \`pm etl check\` are the typed local surfaces.\n- Fixture-only validation does not certify live provider behavior. Certified/live status remains \`0\` until a separately approved live executor runs with redacted artifacts.\n- Nested direct reads rely on typed flags or matching config values for path variables; no generic raw API command is exposed.\n`;
  writeText(path.join(OUT, 'docs.md'), docs);

  console.log(JSON.stringify(counts, null, 2));
}

main().catch(err => { console.error(err); process.exit(1); });
