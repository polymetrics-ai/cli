import { NextRequest } from 'next/server';
import { describe, expect, it } from 'vitest';

import {
  GET,
  generateStaticParams,
} from '../../app/docs/connectors/[slug]/data.json/route';
import { CONNECTOR_CATALOG } from '../../lib/connectors.catalog.generated';

async function connectorData(slug: string) {
  const request = new NextRequest(`http://localhost/docs/connectors/${slug}/data.json`);
  const response = await GET(request, { params: Promise.resolve({ slug }) });
  return {
    response,
    json: (await response.json()) as Record<string, unknown>,
  };
}

describe('connector data route', () => {
  it('generates one static JSON route per catalog connector', () => {
    const params = generateStaticParams();

    expect(params).toHaveLength(CONNECTOR_CATALOG.length);
    expect(new Set(params.map((param) => param.slug)).size).toBe(params.length);
    expect(params).toContainEqual({ slug: '100ms' });
  });

  it('returns public generated connector catalog data', async () => {
    const { response, json } = await connectorData('100ms');

    expect(response.status).toBe(200);
    expect(json).toMatchObject({
      slug: '100ms',
      name: '100ms',
      category: 'api',
      capabilities: expect.objectContaining({
        check: true,
        read: true,
        write: true,
      }),
    });
    expect(json.streams).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ name: 'rooms' }),
      ]),
    );
    expect(json.writeActions).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ name: 'create_room' }),
      ]),
    );
    expect(json.docsMd).toContain('management_token');
  });

  it('returns GitHub CLI surface metadata for docs rendering', async () => {
    const { response, json } = await connectorData('github');

    expect(response.status).toBe(200);
    expect(json.cliSurface).toMatchObject({
      usage: 'pm github <command> <subcommand> [flags]',
      globalFlags: expect.arrayContaining([
        expect.objectContaining({ name: 'json', type: 'boolean' }),
      ]),
      groups: expect.arrayContaining([
        expect.objectContaining({
          title: 'Core Commands',
          commands: expect.arrayContaining(['issue', 'pr', 'repo', 'release']),
        }),
      ]),
      commands: expect.arrayContaining([
        expect.objectContaining({
          path: 'issue list',
          intent: 'etl',
          availability: 'implemented',
          stream: 'issues',
        }),
        expect.objectContaining({
          path: 'issue create',
          intent: 'reverse_etl',
          availability: 'implemented',
          write: 'create_issue',
        }),
      ]),
    });
  });

  it('returns Chatwoot CLI surface metadata for docs rendering', async () => {
    const { response, json } = await connectorData('chatwoot');

    expect(response.status).toBe(200);
    expect(json.cliSurface).toMatchObject({
      usage: 'pm chatwoot <command> <subcommand> [flags]',
      globalFlags: expect.arrayContaining([
        expect.objectContaining({ name: 'json', type: 'boolean' }),
      ]),
      groups: expect.arrayContaining([
        expect.objectContaining({
          title: 'Support Desk Commands',
          commands: expect.arrayContaining(['conversation', 'contact', 'message']),
        }),
        expect.objectContaining({
          title: 'Blocked Admin And Configuration Commands',
          commands: expect.arrayContaining(['account', 'automation', 'webhook']),
        }),
      ]),
      commands: expect.arrayContaining([
        expect.objectContaining({
          path: 'conversation list',
          intent: 'etl',
          availability: 'implemented',
          stream: 'conversations',
        }),
        expect.objectContaining({
          path: 'message send',
          intent: 'reverse_etl',
          availability: 'implemented',
          write: 'send_message',
        }),
        expect.objectContaining({
          path: 'platform account create',
          intent: 'reverse_etl',
          availability: 'unsafe_or_disallowed',
        }),
      ]),
    });
  });

  it('returns a 404 JSON payload for unknown connectors', async () => {
    const { response, json } = await connectorData('missing-connector');

    expect(response.status).toBe(404);
    expect(json).toEqual({ error: 'Connector not found' });
  });
});
