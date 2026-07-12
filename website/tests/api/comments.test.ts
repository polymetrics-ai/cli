import { afterAll, beforeEach, describe, expect, it, vi } from 'vitest';
import { randomUUID } from 'node:crypto';
import { BLOG_POSTS } from '@/lib/blog';
import { resetRateLimits } from '@/lib/annotations/rate-limit';
import type { SessionUser } from '@/lib/auth-session';

vi.mock('@/lib/auth-session', () => ({
  getSessionUser: vi.fn(),
}));

import { getSessionUser } from '@/lib/auth-session';
import { GET, POST } from '@/app/api/comments/route';
import { DELETE } from '@/app/api/comments/[id]/route';
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
    exact: text.slice(4, 24),
    prefix: text.slice(0, 4),
    suffix: text.slice(24, 44),
    startOffset: 4,
  };
}

function user(id: string, isAdmin = false): SessionUser {
  return { id, name: `Test ${id.slice(0, 6)}`, email: `${id.slice(0, 8)}@test.dev`, image: null, isAdmin };
}

async function createDbUser(id: string) {
  await getPool().query(
    `INSERT INTO "user" ("id", "name", "email") VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
    [id, `Test ${id.slice(0, 6)}`, `${id}@test.dev`],
  );
}

function getRequest() {
  return new Request(`http://localhost/api/comments?slug=${slug}`);
}

function postRequest(body: unknown) {
  return new Request('http://localhost/api/comments', {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify(body),
  });
}

function deleteRequest(id: string) {
  return [
    new Request(`http://localhost/api/comments/${id}`, { method: 'DELETE' }),
    { params: Promise.resolve({ id }) },
  ] as const;
}

const owner = randomUUID();
const stranger = randomUUID();

beforeEach(async () => {
  resetRateLimits();
  mockSession.mockReset();
  await ensureMigrated();
  await createDbUser(owner);
  await createDbUser(stranger);
  await getPool().query('DELETE FROM comment WHERE user_id IN ($1, $2)', [owner, stranger]);
});

afterAll(async () => {
  await getPool().query('DELETE FROM comment WHERE user_id IN ($1, $2)', [owner, stranger]);
  await getPool().query('DELETE FROM "user" WHERE "id" IN ($1, $2)', [owner, stranger]);
  await getPool().end();
});

describe('POST /api/comments', () => {
  it('rejects unauthenticated posts with 401', async () => {
    mockSession.mockResolvedValue(null);
    const response = await POST(postRequest({ slug, body: 'hi', anchor: validAnchor() }));
    expect(response.status).toBe(401);
  });

  it('rejects an unknown slug with 400', async () => {
    mockSession.mockResolvedValue(user(owner));
    const response = await POST(postRequest({ slug: 'no-such-post', body: 'hi', anchor: validAnchor() }));
    expect(response.status).toBe(400);
  });

  it('rejects an oversized body with 400', async () => {
    mockSession.mockResolvedValue(user(owner));
    const response = await POST(postRequest({ slug, body: 'x'.repeat(2001), anchor: validAnchor() }));
    expect(response.status).toBe(400);
  });

  it('rejects out-of-range anchors with 400', async () => {
    mockSession.mockResolvedValue(user(owner));
    const response = await POST(
      postRequest({ slug, body: 'hi', anchor: { ...validAnchor(), sectionIndex: 99 } }),
    );
    expect(response.status).toBe(400);
  });

  it('creates a comment (201) and lists it with author but never email', async () => {
    mockSession.mockResolvedValue(user(owner));
    const created = await POST(postRequest({ slug, body: 'A thoughtful note', anchor: validAnchor() }));
    expect(created.status).toBe(201);
    const { comment } = await created.json();
    expect(comment.mine).toBe(true);
    expect(comment.author.name).toBeTruthy();

    const listed = await GET(getRequest());
    expect(listed.status).toBe(200);
    const { comments } = await listed.json();
    const found = comments.find((c: { id: string }) => c.id === comment.id);
    expect(found.body).toBe('A thoughtful note');
    expect(JSON.stringify(found)).not.toContain('@test.dev');
  });

  it('marks mine=false for other readers', async () => {
    mockSession.mockResolvedValue(user(owner));
    const created = await POST(postRequest({ slug, body: 'mine flag test', anchor: validAnchor() }));
    const { comment } = await created.json();

    mockSession.mockResolvedValue(user(stranger));
    const listed = await GET(getRequest());
    const { comments } = await listed.json();
    expect(comments.find((c: { id: string }) => c.id === comment.id).mine).toBe(false);
  });

  it('rate limits a burst with 429 + Retry-After', async () => {
    mockSession.mockResolvedValue(user(owner));
    for (let i = 0; i < 5; i++) {
      const ok = await POST(postRequest({ slug, body: `burst ${i}`, anchor: validAnchor() }));
      expect(ok.status).toBe(201);
    }
    const blocked = await POST(postRequest({ slug, body: 'one too many', anchor: validAnchor() }));
    expect(blocked.status).toBe(429);
    expect(Number(blocked.headers.get('Retry-After'))).toBeGreaterThan(0);
  });
});

describe('DELETE /api/comments/[id]', () => {
  async function createComment(): Promise<string> {
    mockSession.mockResolvedValue(user(owner));
    const created = await POST(postRequest({ slug, body: 'deletable', anchor: validAnchor() }));
    const { comment } = await created.json();
    return comment.id;
  }

  it('lets the author delete their own comment', async () => {
    const id = await createComment();
    const response = await DELETE(...deleteRequest(id));
    expect(response.status).toBe(204);
  });

  it('blocks strangers with 403', async () => {
    const id = await createComment();
    mockSession.mockResolvedValue(user(stranger));
    const response = await DELETE(...deleteRequest(id));
    expect(response.status).toBe(403);
  });

  it('lets an admin delete any comment', async () => {
    const id = await createComment();
    mockSession.mockResolvedValue(user(stranger, true));
    const response = await DELETE(...deleteRequest(id));
    expect(response.status).toBe(204);
  });

  it('404s on a missing comment', async () => {
    mockSession.mockResolvedValue(user(owner));
    const response = await DELETE(...deleteRequest(randomUUID()));
    expect(response.status).toBe(404);
  });
});
