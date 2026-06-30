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
    expect(params).toContainEqual({ slug: 'source-100ms' });
  });

  it('returns public generated connector data with enrichment', async () => {
    const { response, json } = await connectorData('source-100ms');

    expect(response.status).toBe(200);
    expect(json).toMatchObject({
      slug: 'source-100ms',
      name: '100ms',
      type: 'source',
      enrichment: expect.objectContaining({
        prerequisites: expect.any(Array),
        authMethods: expect.arrayContaining([
          expect.objectContaining({ name: 'Management token' }),
        ]),
      }),
    });
    expect(json).not.toHaveProperty('docUrl');
    expect(json.config).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          name: 'management_token',
          required: true,
          secret: true,
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
