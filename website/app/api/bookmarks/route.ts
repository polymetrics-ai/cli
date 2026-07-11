import { NextResponse } from 'next/server';
import { getSessionUser } from '@/lib/auth-session';
import { isValidAnchorShape, blockText } from '@/lib/annotations/anchor';
import { checkRateLimit } from '@/lib/annotations/rate-limit';
import { insertBookmark, listBookmarks } from '@/lib/db/bookmarks';
import { getBlogPost } from '@/lib/blog';

export const dynamic = 'force-dynamic';

function error(message: string, status: number, headers?: HeadersInit) {
  return NextResponse.json({ error: message }, { status, headers });
}

export async function GET(request: Request) {
  const user = await getSessionUser(request.headers);
  if (!user) return error('sign in required', 401);

  const slug = new URL(request.url).searchParams.get('slug') ?? undefined;
  if (slug && !getBlogPost(slug)) return error('unknown post slug', 400);

  const bookmarks = await listBookmarks(user.id, slug);
  return NextResponse.json({
    bookmarks: bookmarks.map((bookmark) => ({
      id: bookmark.id,
      postSlug: bookmark.postSlug,
      anchor: bookmark.anchor,
      createdAt: bookmark.createdAt,
    })),
  });
}

export async function POST(request: Request) {
  const user = await getSessionUser(request.headers);
  if (!user) return error('sign in to bookmark', 401);

  let payload: { slug?: unknown; anchor?: unknown };
  try {
    payload = await request.json();
  } catch {
    return error('invalid JSON body', 400);
  }

  if (typeof payload.slug !== 'string') return error('unknown post slug', 400);
  const post = getBlogPost(payload.slug);
  if (!post) return error('unknown post slug', 400);
  if (!isValidAnchorShape(payload.anchor)) return error('invalid anchor', 400);
  if (
    payload.anchor.sectionIndex >= post.sections.length ||
    blockText(post, payload.anchor.blockType, payload.anchor.sectionIndex, payload.anchor.blockIndex) === undefined
  ) {
    return error('anchor out of range', 400);
  }

  const rate = checkRateLimit('bookmark:create', user.id);
  if (!rate.allowed) {
    return error('too many bookmarks, slow down', 429, {
      'Retry-After': String(rate.retryAfterSeconds),
    });
  }

  const { bookmark, created } = await insertBookmark({
    postSlug: payload.slug,
    userId: user.id,
    anchor: payload.anchor,
  });

  return NextResponse.json(
    {
      bookmark: {
        id: bookmark.id,
        postSlug: bookmark.postSlug,
        anchor: bookmark.anchor,
        createdAt: bookmark.createdAt,
      },
    },
    { status: created ? 201 : 200 },
  );
}
