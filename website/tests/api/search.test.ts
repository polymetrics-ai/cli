import { describe, expect, it } from 'vitest';

import { GET } from '../../app/api/search/route';

type SearchResult = {
  title: string;
  description: string;
  url: string;
  kind: 'doc' | 'connector';
  section: string;
  snippet: string;
  score: number;
};

type SearchPayload = {
  query: string;
  total: number;
  results: SearchResult[];
};

async function search(params: Record<string, string>): Promise<SearchPayload> {
  const url = new URL('http://localhost/api/search');
  for (const [key, value] of Object.entries(params)) {
    url.searchParams.set(key, value);
  }

  const response = await GET(new Request(url));
  expect(response.status).toBe(200);
  return response.json() as Promise<SearchPayload>;
}

describe('search API index', () => {
  it('indexes docs MDX pages', async () => {
    const payload = await search({ q: 'quickstart', limit: '8' });

    expect(payload.query).toBe('quickstart');
    expect(payload.results).toContainEqual(
      expect.objectContaining({
        kind: 'doc',
        title: 'Quickstart',
        url: '/docs/quickstart',
      }),
    );
  });

  it('indexes generated connector metadata and setup fields', async () => {
    const payload = await search({ q: 'management_token', limit: '12' });

    expect(payload.results).toContainEqual(
      expect.objectContaining({
        kind: 'connector',
        title: '100ms connector',
        url: '/docs/connectors/100ms',
      }),
    );
  });

  it('includes the generated connector catalog landing page', async () => {
    const payload = await search({ q: 'connector catalog', limit: '10' });

    expect(payload.results).toContainEqual(
      expect.objectContaining({
        kind: 'doc',
        title: 'Connector catalog',
        url: '/docs/connectors',
      }),
    );
  });

  it('normalizes alternate query params and caps large limits', async () => {
    const payload = await search({ search: 'agent', limit: '500' });

    expect(payload.query).toBe('agent');
    expect(payload.results.length).toBeGreaterThan(0);
    expect(payload.results.length).toBeLessThanOrEqual(50);
    expect(payload.total).toBe(payload.results.length);
  });
});
