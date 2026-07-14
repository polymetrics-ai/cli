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

  it('returns Asana generated connector catalog data for website docs parity', async () => {
    const { response, json } = await connectorData('asana');

    expect(response.status).toBe(200);
    expect(json).toMatchObject({
      slug: 'asana',
      name: 'Asana',
      category: 'api',
      releaseStage: 'ga',
      capabilities: {
        check: true,
        read: true,
        write: true,
        query: false,
        cdc: false,
        dynamicSchema: false,
      },
    });
    expect((json.streams as Array<{ name: string }>).map((stream) => stream.name)).toEqual(
      expect.arrayContaining([
        'workspaces',
        'projects',
        'tasks',
        'custom_fields',
        'workspace_memberships',
      ]),
    );
    expect(
      (json.cliSurface as { commands: Array<Record<string, unknown>> }).commands,
    ).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          path: 'workspace-memberships list',
          intent: 'etl',
          availability: 'implemented',
          stream: 'workspace_memberships',
          examples: expect.arrayContaining([
            'pm asana workspace-memberships list --workspace <workspace-gid> --json',
          ]),
        }),
      ]),
    );
  });

  it('returns a 404 JSON payload for unknown connectors', async () => {
    const { response, json } = await connectorData('missing-connector');

    expect(response.status).toBe(404);
    expect(json).toEqual({ error: 'Connector not found' });
  });
});
