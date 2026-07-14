import { NextRequest } from 'next/server';
import { describe, expect, it } from 'vitest';

import { GET } from '../../app/api/raw/[...slug]/route';

async function rawMarkdown(slug: string[]) {
  const request = new NextRequest(`http://localhost/api/raw/${slug.join('/')}`);
  const response = await GET(request, { params: Promise.resolve({ slug }) });
  return {
    response,
    text: await response.text(),
  };
}

describe('raw Markdown API', () => {
  it('returns the docs index for the copy-markdown index alias', async () => {
    const { response, text } = await rawMarkdown(['index']);

    expect(response.status).toBe(200);
    expect(response.headers.get('content-type')).toContain('text/markdown');
    expect(text).toContain('The extract-query-act loop');
    expect(text).not.toMatch(/^---/);
  });

  it('returns generated docs Markdown without filesystem access', async () => {
    const { response, text } = await rawMarkdown(['quickstart']);

    expect(response.status).toBe(200);
    expect(text).toContain('Install pm');
    expect(text).toContain('pm etl run --connection demo');
  });

  it('returns synthetic connector Markdown', async () => {
    const { response, text } = await rawMarkdown(['connectors', '100ms']);

    expect(response.status).toBe(200);
    expect(text).toContain('# 100ms connector');
    expect(text).toContain('management_token');
  });

  it('returns synthetic Asana connector Markdown from generated catalog data', async () => {
    const { response, text } = await rawMarkdown(['connectors', 'asana']);

    expect(response.status).toBe(200);
    expect(response.headers.get('content-type')).toContain('text/markdown');
    expect(text).toContain('# Asana connector');
    expect(text).toContain('Reads Asana workspaces, projects, tasks');
    expect(text).toContain('- **Write:** Yes');
    expect(text).toContain('| `workspace_memberships` | gid | - | No |');
    expect(text).toContain('| `create_task` | POST | create |');
  });

  it('rejects traversal-like docs slugs', async () => {
    const { response, text } = await rawMarkdown(['..', 'quickstart']);

    expect(response.status).toBe(404);
    expect(text).toBe('Not found');
  });
});
