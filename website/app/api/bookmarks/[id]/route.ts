import { NextResponse } from 'next/server';
import { getSessionUser } from '@/lib/auth-session';
import { deleteBookmark, getBookmarkOwner } from '@/lib/db/bookmarks';

export const dynamic = 'force-dynamic';

export async function DELETE(request: Request, { params }: { params: Promise<{ id: string }> }) {
  const user = await getSessionUser(request.headers);
  if (!user) return NextResponse.json({ error: 'sign in required' }, { status: 401 });

  const { id } = await params;
  const owner = await getBookmarkOwner(id);
  if (!owner) return NextResponse.json({ error: 'bookmark not found' }, { status: 404 });

  // Bookmarks are private — only the owner may remove one, admin or not.
  if (owner !== user.id) {
    return NextResponse.json({ error: 'not your bookmark' }, { status: 403 });
  }

  await deleteBookmark(id);
  return new NextResponse(null, { status: 204 });
}
