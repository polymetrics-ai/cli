import { NextResponse } from 'next/server';
import { getSessionUser } from '@/lib/auth-session';
import { isValidAnchorShape, blockText } from '@/lib/annotations/anchor';
import { checkRateLimit } from '@/lib/annotations/rate-limit';
import { insertComment, listCommentsBySlug } from '@/lib/db/comments';
import { getBlogPost } from '@/lib/blog';
import type { Anchor } from '@/lib/annotations/anchor';

export const dynamic = 'force-dynamic';

export const COMMENT_MAX_LENGTH = 2000;

function error(message: string, status: number, headers?: HeadersInit) {
  return NextResponse.json({ error: message }, { status, headers });
}

/**
 * The server validates anchor shape and addressing (the block must exist)
 * but not the quote text — content edits must never brick old anchors;
 * the client resolves and orphans gracefully.
 */
function validateAnchor(slug: string, anchor: unknown): string | Anchor {
  const post = getBlogPost(slug);
  if (!post) return 'unknown post slug';
  if (!isValidAnchorShape(anchor)) return 'invalid anchor';
  if (anchor.sectionIndex >= post.sections.length) return 'anchor section out of range';
  if (blockText(post, anchor.blockType, anchor.sectionIndex, anchor.blockIndex) === undefined) {
    return 'anchor block out of range';
  }
  return anchor;
}

export async function GET(request: Request) {
  const slug = new URL(request.url).searchParams.get('slug');
  if (!slug || !getBlogPost(slug)) return error('unknown post slug', 400);

  const [user, comments] = await Promise.all([
    getSessionUser(request.headers),
    listCommentsBySlug(slug),
  ]);

  return NextResponse.json({
    comments: comments.map((comment) => ({
      id: comment.id,
      body: comment.body,
      anchor: comment.anchor,
      createdAt: comment.createdAt,
      author: comment.author,
      mine: user !== null && comment.userId === user.id,
    })),
    viewer: { admin: user?.isAdmin ?? false },
  });
}

export async function POST(request: Request) {
  const user = await getSessionUser(request.headers);
  if (!user) return error('sign in to comment', 401);

  let payload: { slug?: unknown; body?: unknown; anchor?: unknown };
  try {
    payload = await request.json();
  } catch {
    return error('invalid JSON body', 400);
  }

  if (typeof payload.slug !== 'string') return error('unknown post slug', 400);
  if (typeof payload.body !== 'string' || payload.body.trim().length === 0) {
    return error('comment body is required', 400);
  }
  if (payload.body.length > COMMENT_MAX_LENGTH) {
    return error(`comment body exceeds ${COMMENT_MAX_LENGTH} characters`, 400);
  }

  const anchor = validateAnchor(payload.slug, payload.anchor);
  if (typeof anchor === 'string') return error(anchor, 400);

  const rate = checkRateLimit('comment:create', user.id);
  if (!rate.allowed) {
    return error('too many comments, slow down', 429, {
      'Retry-After': String(rate.retryAfterSeconds),
    });
  }

  const comment = await insertComment({
    postSlug: payload.slug,
    userId: user.id,
    body: payload.body.trim(),
    anchor,
  });

  return NextResponse.json(
    {
      comment: {
        id: comment.id,
        body: comment.body,
        anchor: comment.anchor,
        createdAt: comment.createdAt,
        author: comment.author,
        mine: true,
      },
    },
    { status: 201 },
  );
}
