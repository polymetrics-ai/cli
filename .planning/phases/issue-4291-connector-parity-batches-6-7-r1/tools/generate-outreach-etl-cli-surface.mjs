#!/usr/bin/env node

import { readFileSync, writeFileSync } from 'node:fs';

const connector = 'outreach';
const root = `internal/connectors/defs/${connector}`;
const streams = JSON.parse(readFileSync(`${root}/streams.json`, 'utf8'));
const lock = JSON.parse(readFileSync(`${root}/sources/${connector}-operation-source-lock.json`, 'utf8'));
const apiSurface = JSON.parse(readFileSync(`${root}/api_surface.json`, 'utf8'));

if (lock.schema_version !== 2 || lock.connector !== connector) {
  throw new Error('expected the pinned Outreach schema-version 2 source lock');
}

function normalizedPath(path) {
  return path
    .replace(/\{\{\s*config\.[^}]+\}\}/g, '{}')
    .replace(/\{[^}]+\}/g, '{}');
}

function words(name) {
  return name.split('_').map((word) => word[0].toUpperCase() + word.slice(1)).join(' ');
}

const sourcesByPath = new Map();
for (const operation of lock.rest.operations) {
  if (operation.protocol !== 'rest' || operation.method !== 'GET') {
    continue;
  }
  const key = `${operation.method} ${normalizedPath(operation.path)}`;
  const prior = sourcesByPath.get(key) ?? [];
  prior.push(operation);
  sourcesByPath.set(key, prior);
}

const commands = streams.streams.map((stream) => {
  const key = `GET ${normalizedPath(`/api/v2${stream.path}`)}`;
  const matches = sourcesByPath.get(key) ?? [];
  if (matches.length !== 1) {
    throw new Error(`${stream.name}: expected exactly one pinned GET source for ${key}, found ${matches.length}`);
  }
  const source = matches[0];
  const endpoints = apiSurface.endpoints.filter((endpoint) => (
    endpoint.method === 'GET' && endpoint.covered_by?.stream === stream.name
  ));
  if (endpoints.length !== 1 || endpoints[0].path !== source.path) {
    throw new Error(`${stream.name}: expected one API-surface endpoint matching ${source.path}`);
  }
  return {
    path: `${stream.name.replaceAll('_', '-')} list`,
    summary: `Read Outreach ${words(stream.name)} as ETL records.`,
    intent: 'etl',
    availability: 'implemented',
    stream: stream.name,
    source_url: source.source_url,
    api_surface: [{ method: endpoints[0].method, path: endpoints[0].path }],
    examples: [`pm outreach ${stream.name.replaceAll('_', '-')} list --json`],
  };
});

const surface = {
  tagline: 'Read Outreach REST API v2 records and safely plan typed Outreach mutations.',
  usage: 'pm outreach <command> [flags]',
  source_cli: {
    name: 'Outreach REST API v2',
    docs: 'https://api.outreach.io/api/v2/schema/openapi.json',
    reference: 'Pinned Outreach OpenAPI 3.0.3 and the public Custom Objects documentation audit.',
    source: 'provider_api',
  },
  groups: [{
    id: 'etl',
    title: 'ETL streams',
    commands: commands.map((command) => command.path.split(' ')[0]),
  }],
  global_flags: [
    { name: 'credential', type: 'string', summary: 'Credential name to use for the Outreach request.' },
    { name: 'connection', type: 'string', summary: 'Alias for --credential.' },
    { name: 'config', type: 'string_array', summary: 'Connector config override as key=value; never pass secret values here.' },
    { name: 'json', type: 'boolean', summary: 'Emit machine-readable JSON output.' },
    { name: 'limit', type: 'integer', summary: 'Maximum ETL records to emit.' },
  ],
  commands,
};

writeFileSync(`${root}/cli_surface.json`, `${JSON.stringify(surface, null, 2)}\n`);
