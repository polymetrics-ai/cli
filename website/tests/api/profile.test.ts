import { afterAll, beforeEach, describe, expect, it, vi } from 'vitest';
import { randomUUID } from 'node:crypto';
import type { SessionUser } from '@/lib/auth-session';

vi.mock('@/lib/auth-session', () => ({
  getSessionUser: vi.fn(),
}));

import { getSessionUser } from '@/lib/auth-session';
import { GET, PUT } from '@/app/api/profile/route';
import { ensureMigrated, getPool } from '@/lib/db';

const mockSession = vi.mocked(getSessionUser);

const alice = randomUUID();

function user(id: string): SessionUser {
  return { id, name: 'Profiled', email: `${id.slice(0, 8)}@test.dev`, image: null, isAdmin: false };
}

function putRequest(body: unknown) {
  return new Request('http://localhost/api/profile', {
    method: 'PUT',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify(body),
  });
}

beforeEach(async () => {
  mockSession.mockReset();
  await ensureMigrated();
  await getPool().query(
    `INSERT INTO "user" ("id", "name", "email") VALUES ($1, 'Profiled', $2) ON CONFLICT DO NOTHING`,
    [alice, `${alice}@test.dev`],
  );
  await getPool().query('DELETE FROM profile_settings WHERE user_id = $1', [alice]);
});

afterAll(async () => {
  await getPool().query('DELETE FROM "user" WHERE "id" = $1', [alice]);
  await getPool().end();
});

describe('/api/profile', () => {
  it('requires auth for GET and PUT', async () => {
    mockSession.mockResolvedValue(null);
    expect((await GET(new Request('http://localhost/api/profile'))).status).toBe(401);
    expect((await PUT(putRequest({ profileVisible: true }))).status).toBe(401);
  });

  it('defaults to a private profile', async () => {
    mockSession.mockResolvedValue(user(alice));
    const response = await GET(new Request('http://localhost/api/profile'));
    const { settings } = await response.json();
    expect(settings.profileVisible).toBe(false);
    expect(settings.profileUrl).toBeNull();
  });

  it('saves visibility and a valid https url', async () => {
    mockSession.mockResolvedValue(user(alice));
    const saved = await PUT(putRequest({ profileVisible: true, profileUrl: 'https://example.dev' }));
    expect(saved.status).toBe(200);
    const { settings } = await saved.json();
    expect(settings.profileVisible).toBe(true);
    expect(settings.profileUrl).toBe('https://example.dev/');
  });

  it('rejects non-https and oversized urls', async () => {
    mockSession.mockResolvedValue(user(alice));
    expect((await PUT(putRequest({ profileVisible: true, profileUrl: 'http://x.dev' }))).status).toBe(400);
    expect(
      (await PUT(putRequest({ profileVisible: true, profileUrl: `https://x.dev/${'a'.repeat(200)}` }))).status,
    ).toBe(400);
    expect((await PUT(putRequest({ profileVisible: 'yes' }))).status).toBe(400);
  });

  it('never returns an email field', async () => {
    mockSession.mockResolvedValue(user(alice));
    const response = await GET(new Request('http://localhost/api/profile'));
    expect(JSON.stringify(await response.json())).not.toContain('@test.dev');
  });
});
