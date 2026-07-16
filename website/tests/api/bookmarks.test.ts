import { afterAll, beforeEach, describe, expect, it, vi } from 'vitest';
import { randomUUID } from 'node:crypto';
import { BLOG_POSTS } from '@/lib/blog';
import { resetRateLimits } from '@/lib/annotations/rate-limit';
import type { SessionUser } from '@/lib/auth-session';

vi.mock('@/lib/auth-session', () => ({
  getSessionUser: vi.fn(),
}));

import { getSessionUser } from '@/lib/auth-session';
import { GET, POST } from '@/app/api/bookmarks/route';
import { DELETE } from '@/app/api/bookmarks/[id]/route';
import { ensureMigrated, getPool } from '@/lib/db';

const mockSession = vi.mocked(getSessionUser);

const post = BLOG_POSTS[0];
const slug = post.slug;

function validAnchor() {
  const text = post.sections[0].body[0];
  return {
    sectionIndex: 0,
    blockType: 'body' as const,
    blockIndex: 0,
    exact: text.slice(0, 16),
    prefix: '',
    suffix: text.slice(16, 32),
    startOffset: 0,
  };
}

function user(id: string): SessionUser {
  return { id, name: 'Bookmarker', email: `${id.slice(0, 8)}@test.dev`, image: null, isAdmin: false };
}

const alice = randomUUID();
const bob = randomUUID();

async function createDbUser(id: string) {
  await getPool().query(
    `INSERT INTO "user" ("id", "name", "email") VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
    [id, 'Bookmarker', `${id}@test.dev`],
  );
}

function postRequest(body: unknown) {
  return new Request('http://localhost/api/bookmarks', {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify(body),
  });
}

beforeEach(async () => {
  resetRateLimits();
  mockSession.mockReset();
  await ensureMigrated();
  await createDbUser(alice);
  await createDbUser(bob);
  await getPool().query('DELETE FROM bookmark WHERE user_id IN ($1, $2)', [alice, bob]);
});

afterAll(async () => {
  await getPool().query('DELETE FROM bookmark WHERE user_id IN ($1, $2)', [alice, bob]);
  await getPool().query('DELETE FROM "user" WHERE "id" IN ($1, $2)', [alice, bob]);
  await getPool().end();
});

describe('/api/bookmarks', () => {
  it('requires auth for GET and POST', async () => {
    mockSession.mockResolvedValue(null);
    expect((await GET(new Request('http://localhost/api/bookmarks'))).status).toBe(401);
    expect((await POST(postRequest({ slug, anchor: validAnchor() }))).status).toBe(401);
  });

  it('creates then toggles idempotently (201 then 200, one row)', async () => {
    mockSession.mockResolvedValue(user(alice));
    const first = await POST(postRequest({ slug, anchor: validAnchor() }));
    expect(first.status).toBe(201);
    const second = await POST(postRequest({ slug, anchor: validAnchor() }));
    expect(second.status).toBe(200);

    const { bookmark: a } = await first.json();
    const { bookmark: b } = await second.json();
    expect(b.id).toBe(a.id);

    const listed = await GET(new Request(`http://localhost/api/bookmarks?slug=${slug}`));
    const { bookmarks } = await listed.json();
    expect(bookmarks).toHaveLength(1);
  });

  it('keeps body and point anchors distinct when their quote coordinates match', async () => {
    mockSession.mockResolvedValue(user(alice));
    const sectionIndex = post.sections.findIndex(
      (section) => section.body.length > 0 && (section.points?.length ?? 0) > 0,
    );
    expect(sectionIndex).toBeGreaterThanOrEqual(0);

    const shared = {
      ...validAnchor(),
      sectionIndex,
      blockIndex: 0,
      exact: 'same quote',
      prefix: '',
      suffix: '',
      startOffset: 0,
    };
    const body = await POST(postRequest({ slug, anchor: { ...shared, blockType: 'body' } }));
    const point = await POST(postRequest({ slug, anchor: { ...shared, blockType: 'point' } }));

    expect(body.status).toBe(201);
    expect(point.status).toBe(201);
    const bodyBookmark = (await body.json()).bookmark;
    const pointBookmark = (await point.json()).bookmark;
    expect(pointBookmark.id).not.toBe(bodyBookmark.id);

    const listed = await GET(new Request(`http://localhost/api/bookmarks?slug=${slug}`));
    expect((await listed.json()).bookmarks).toHaveLength(2);
  });

  it('keeps bookmarks private per user', async () => {
    mockSession.mockResolvedValue(user(alice));
    await POST(postRequest({ slug, anchor: validAnchor() }));

    mockSession.mockResolvedValue(user(bob));
    const listed = await GET(new Request(`http://localhost/api/bookmarks?slug=${slug}`));
    const { bookmarks } = await listed.json();
    expect(bookmarks).toHaveLength(0);
  });

  it('rejects invalid anchors with 400', async () => {
    mockSession.mockResolvedValue(user(alice));
    const response = await POST(postRequest({ slug, anchor: { nope: true } }));
    expect(response.status).toBe(400);
  });

  it('only the owner can delete a bookmark', async () => {
    mockSession.mockResolvedValue(user(alice));
    const created = await POST(postRequest({ slug, anchor: validAnchor() }));
    const { bookmark } = await created.json();
    const args = (id: string) =>
      [
        new Request(`http://localhost/api/bookmarks/${id}`, { method: 'DELETE' }),
        { params: Promise.resolve({ id }) },
      ] as const;

    mockSession.mockResolvedValue(user(bob));
    expect((await DELETE(...args(bookmark.id))).status).toBe(403);

    mockSession.mockResolvedValue(user(alice));
    expect((await DELETE(...args(bookmark.id))).status).toBe(204);
    expect((await DELETE(...args(bookmark.id))).status).toBe(404);
  });
});
