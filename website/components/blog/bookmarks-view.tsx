'use client';

import { useEffect, useMemo, useState } from 'react';
import Link from 'next/link';
import { Bookmark, Trash2 } from 'lucide-react';
import { CornerBox } from '@/components/ui/corner-box';
import { Skeleton } from '@/components/shadcn/ui/skeleton';
import { SiteFooter } from '@/components/home/site-footer';
import { SignInDialog } from '@/components/auth/sign-in-dialog';
import { useSession } from '@/lib/auth-client';
import { getBlogPost } from '@/lib/blog';
import type { Anchor } from '@/lib/annotations/anchor';

type BookmarkDto = {
  id: string;
  postSlug: string;
  anchor: Anchor;
  createdAt: string;
};

function Shell({ children }: { children: React.ReactNode }) {
  return (
    <main className="mx-auto w-full px-4 pt-10 sm:px-8 md:max-w-[680px] md:px-0 xl:max-w-[840px]">
      <header className="mb-10 border-b border-line-structure pb-8">
        <span className="inline-flex items-center gap-2 font-mono text-[12px] uppercase tracking-widest text-text-disabled">
          <Bookmark className="h-3.5 w-3.5 text-line-cta" aria-hidden="true" />
          Private library
        </span>
        <h1 className="mt-4 font-square text-[44px] font-semibold leading-[1] text-text-primary">
          Bookmarks
        </h1>
        <p className="mt-4 max-w-[52ch] text-[14px] leading-relaxed text-text-tertiary">
          Passages you saved across the blog. Only you can see these.
        </p>
      </header>
      {children}
      <SiteFooter />
    </main>
  );
}

export function BookmarksView() {
  const { data: session, isPending } = useSession();
  const [bookmarks, setBookmarks] = useState<BookmarkDto[] | null>(null);
  const [signInOpen, setSignInOpen] = useState(false);

  useEffect(() => {
    if (!session) return;
    const controller = new AbortController();
    fetch('/api/bookmarks', { signal: controller.signal })
      .then((response) => (response.ok ? response.json() : { bookmarks: [] }))
      .then((data: { bookmarks: BookmarkDto[] }) => setBookmarks(data.bookmarks))
      .catch(() => setBookmarks([]));
    return () => controller.abort();
  }, [session]);

  const grouped = useMemo(() => {
    const groups = new Map<string, BookmarkDto[]>();
    for (const bookmark of bookmarks ?? []) {
      const list = groups.get(bookmark.postSlug) ?? [];
      list.push(bookmark);
      groups.set(bookmark.postSlug, list);
    }
    return [...groups.entries()];
  }, [bookmarks]);

  async function remove(id: string) {
    setBookmarks((current) => (current ?? []).filter((b) => b.id !== id));
    await fetch(`/api/bookmarks/${id}`, { method: 'DELETE' });
  }

  if (isPending) {
    return (
      <Shell>
        <div className="flex flex-col gap-3">
          <Skeleton className="h-16 w-full rounded-none" />
          <Skeleton className="h-16 w-full rounded-none" />
        </div>
      </Shell>
    );
  }

  if (!session) {
    return (
      <Shell>
        <CornerBox withStripes className="p-6">
          <div className="bg-surface-bg p-4">
            <p className="font-mono text-[10px] uppercase tracking-widest text-text-disabled">
              Sign in required
            </p>
            <p className="mt-3 text-[14px] leading-relaxed text-text-tertiary">
              Bookmarks are tied to your account. Sign in to see the passages you saved.
            </p>
            <button
              type="button"
              onClick={() => setSignInOpen(true)}
              className="mt-4 border border-line-cta bg-line-cta px-4 py-2 font-square text-[11px] font-semibold uppercase tracking-wider text-surface-bg transition-opacity hover:opacity-90"
            >
              Sign in
            </button>
          </div>
        </CornerBox>
        <SignInDialog open={signInOpen} onOpenChange={setSignInOpen} callbackURL="/bookmarks" />
      </Shell>
    );
  }

  return (
    <Shell>
      {bookmarks === null ? (
        <div className="flex flex-col gap-3">
          <Skeleton className="h-16 w-full rounded-none" />
          <Skeleton className="h-16 w-full rounded-none" />
        </div>
      ) : bookmarks.length === 0 ? (
        <CornerBox withStripes className="p-6">
          <p className="bg-surface-bg p-4 text-[14px] leading-relaxed text-text-tertiary">
            Nothing saved yet. Select any passage in a{' '}
            <Link href="/blog" className="text-line-cta underline underline-offset-2">
              blog article
            </Link>{' '}
            and choose Bookmark.
          </p>
        </CornerBox>
      ) : (
        <div className="flex flex-col gap-10">
          {grouped.map(([slug, list]) => {
            const post = getBlogPost(slug);
            return (
              <section key={slug}>
                <div className="flex items-baseline justify-between border-b border-line-structure pb-2">
                  <h2 className="font-square text-[16px] font-semibold text-text-primary">
                    {post?.title ?? slug}
                  </h2>
                  <span className="font-mono text-[10px] uppercase tracking-widest text-text-disabled">
                    {list.length} saved
                  </span>
                </div>
                <div className="mt-3 flex flex-col gap-2.5">
                  {list.map((bookmark) => (
                    <CornerBox key={bookmark.id} hoverStripes className="group/bm relative">
                      <div className="flex items-start gap-3 p-3.5">
                        <Link href={`/blog/${slug}?b=${bookmark.id}`} className="min-w-0 flex-1">
                          <blockquote className="line-clamp-2 border-l-2 border-[#34d399] bg-surface-1 px-3 py-1.5 text-[13px] italic leading-relaxed text-text-secondary">
                            {bookmark.anchor.exact}
                          </blockquote>
                          <span className="mt-2 block font-mono text-[10px] uppercase tracking-widest text-text-disabled">
                            Saved {bookmark.createdAt.slice(0, 10)} · Jump to passage →
                          </span>
                        </Link>
                        <button
                          type="button"
                          aria-label="Remove bookmark"
                          onClick={() => void remove(bookmark.id)}
                          className="shrink-0 border border-transparent p-1.5 text-text-disabled transition-colors hover:border-line-structure hover:text-[#b42318]"
                        >
                          <Trash2 className="h-3.5 w-3.5" aria-hidden="true" />
                        </button>
                      </div>
                    </CornerBox>
                  ))}
                </div>
              </section>
            );
          })}
        </div>
      )}
    </Shell>
  );
}
