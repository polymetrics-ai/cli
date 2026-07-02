import { NextRequest } from 'next/server';
import { describe, expect, it } from 'vitest';

import { proxy } from '../proxy';

function request(url: string, headers: Record<string, string>) {
  return new NextRequest(url, { headers });
}

describe('website HTTPS middleware', () => {
  it('redirects forwarded HTTP traffic to HTTPS', () => {
    const response = proxy(
      request('http://cli.polymetrics.ai/docs?tab=install', {
        host: 'cli.polymetrics.ai:3000',
        'x-forwarded-proto': 'http',
      }),
    );

    expect(response.status).toBe(308);
    expect(response.headers.get('location')).toBe('https://cli.polymetrics.ai/docs?tab=install');
  });

  it('adds HSTS on forwarded HTTPS traffic', () => {
    const response = proxy(
      request('https://cli.polymetrics.ai/', {
        host: 'cli.polymetrics.ai',
        'x-forwarded-proto': 'https',
      }),
    );

    expect(response.headers.get('strict-transport-security')).toBe(
      'max-age=63072000; includeSubDomains; preload',
    );
  });

  it('does not force HTTPS for local development', () => {
    const response = proxy(
      request('http://localhost:3000/docs', {
        host: 'localhost:3000',
        'x-forwarded-proto': 'http',
      }),
    );

    expect(response.status).not.toBe(308);
    expect(response.headers.get('location')).toBeNull();
  });
});
