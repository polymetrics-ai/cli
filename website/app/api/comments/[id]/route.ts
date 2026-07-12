import { NextResponse } from 'next/server';
import { getSessionUser } from '@/lib/auth-session';
import { deleteComment, getCommentOwner } from '@/lib/db/comments';

export const dynamic = 'force-dynamic';

export async function DELETE(request: Request, { params }: { params: Promise<{ id: string }> }) {
  const user = await getSessionUser(request.headers);
  if (!user) return NextResponse.json({ error: 'sign in required' }, { status: 401 });

  const { id } = await params;
  const owner = await getCommentOwner(id);
  if (!owner) return NextResponse.json({ error: 'comment not found' }, { status: 404 });

  if (owner !== user.id && !user.isAdmin) {
    return NextResponse.json({ error: 'not your comment' }, { status: 403 });
  }

  await deleteComment(id);
  return new NextResponse(null, { status: 204 });
}
