'use client';

import { useMemo, useState } from 'react';
import { MessageSquare, Trash2 } from 'lucide-react';
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet';
import { Skeleton } from '@/components/shadcn/ui/skeleton';
import { useAnnotations } from '@/components/blog/annotations-provider';
import { relativeTime } from '@/components/blog/highlight-text';
import type { CommentDto } from '@/components/blog/annotations-provider';
import type { BlogSection } from '@/lib/blog';

function DeleteButton({ comment }: { comment: CommentDto }) {
  const { deleteComment } = useAnnotations();
  const [confirming, setConfirming] = useState(false);

  if (!confirming) {
    return (
      <button
        type="button"
        aria-label="Delete note"
        onClick={() => setConfirming(true)}
        className="ml-auto flex items-center gap-1 font-mono text-[9px] uppercase tracking-widest text-text-disabled transition-colors hover:text-[--destructive]"
      >
        <Trash2 className="h-3 w-3" aria-hidden="true" />
        Delete
      </button>
    );
  }
  return (
    <span className="ml-auto flex items-center gap-2 font-mono text-[9px] uppercase tracking-widest">
      <button
        type="button"
        onClick={() => void deleteComment(comment.id)}
        className="border border-[#b42318] bg-[#b42318] px-1.5 py-0.5 text-white"
      >
        Confirm
      </button>
      <button
        type="button"
        onClick={() => setConfirming(false)}
        className="text-text-tertiary hover:text-text-primary"
      >
        Keep
      </button>
    </span>
  );
}

export function CommentsSheet({ sections }: { sections: BlogSection[] }) {
  const {
    comments,
    resolutions,
    loading,
    sheetOpen,
    setSheetOpen,
    setActiveId,
    viewerAdmin,
    signedIn,
    requestSignIn,
  } = useAnnotations();

  const grouped = useMemo(() => {
    const groups = new Map<number, CommentDto[]>();
    for (const comment of comments) {
      if (comment.pending) continue;
      const resolution = resolutions.get(comment.id);
      const sectionIndex =
        resolution && !resolution.orphaned ? resolution.sectionIndex : -1; // -1 = context changed
      const list = groups.get(sectionIndex) ?? [];
      list.push(comment);
      groups.set(sectionIndex, list);
    }
    return [...groups.entries()].sort(([a], [b]) => (a === -1 ? 1 : b === -1 ? -1 : a - b));
  }, [comments, resolutions]);

  function jumpTo(comment: CommentDto) {
    setSheetOpen(false);
    setActiveId(comment.id);
    window.setTimeout(() => {
      document.getElementById(`annotation-${comment.id}`)?.scrollIntoView({
        block: 'center',
        behavior: matchMedia('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth',
      });
    }, 250);
  }

  const count = comments.filter((c) => !c.pending).length;

  return (
    <>
      {/* Mobile trigger — the rail is hidden below lg. */}
      <button
        type="button"
        onClick={() => setSheetOpen(true)}
        aria-label={`Open notes (${count})`}
        className="blog-notes-fab fixed bottom-5 right-5 z-40 items-center gap-2 border border-line-cta bg-surface-bg px-3 py-2.5 shadow-[4px_4px_0_0_#c0d6c8] transition-colors hover:bg-surface-1"
      >
        <MessageSquare className="h-4 w-4 text-line-cta" aria-hidden="true" />
        <span className="font-mono text-[11px] font-bold text-text-secondary">{count}</span>
      </button>

      <Sheet open={sheetOpen} onOpenChange={setSheetOpen}>
        <SheetContent side="right" className="w-[380px] max-w-[calc(100vw-2rem)] overflow-y-auto border-l border-line-structure bg-surface-bg p-0 sm:max-w-[380px]">
          <SheetHeader className="border-b border-line-structure px-5 py-4">
            <p className="font-mono text-[10px] uppercase tracking-widest text-text-disabled">
              Marginalia
            </p>
            <SheetTitle className="font-square text-[18px] font-semibold text-text-primary">
              Notes on this article
            </SheetTitle>
            <SheetDescription className="text-[12px] text-text-tertiary">
              {count === 0 ? 'No notes yet.' : `${count} note${count === 1 ? '' : 's'} from readers.`}
            </SheetDescription>
          </SheetHeader>

          <div className="flex flex-col gap-6 px-5 py-5">
            {loading ? (
              <div className="flex flex-col gap-3">
                <Skeleton className="h-20 w-full rounded-none" />
                <Skeleton className="h-20 w-full rounded-none" />
              </div>
            ) : null}

            {!loading && count === 0 ? (
              <div className="with-stripes border border-line-structure p-3">
                <p className="bg-surface-bg p-2 text-[13px] leading-relaxed text-text-tertiary">
                  Select any passage in the article to leave the first note.
                  {!signedIn ? (
                    <button
                      type="button"
                      onClick={() => {
                        setSheetOpen(false);
                        requestSignIn();
                      }}
                      className="mt-1.5 block font-mono text-[10px] uppercase tracking-widest text-line-cta"
                    >
                      Sign in to join →
                    </button>
                  ) : null}
                </p>
              </div>
            ) : null}

            {grouped.map(([sectionIndex, list]) => (
              <div key={sectionIndex}>
                <p className="border-b border-line-structure pb-1.5 font-mono text-[10px] uppercase tracking-widest text-text-disabled">
                  {sectionIndex === -1
                    ? 'Context changed'
                    : `${String(sectionIndex + 1).padStart(2, '0')} — ${sections[sectionIndex]?.heading ?? ''}`}
                </p>
                <div className="mt-3 flex flex-col gap-3">
                  {list.map((comment) => (
                    <div key={comment.id} className="corner-box-corners border border-line-structure bg-surface-bg p-3">
                      <blockquote className="truncate border-l-2 border-[#34d399] bg-surface-1 px-2.5 py-1 text-[12px] italic text-text-tertiary">
                        {comment.anchor.exact}
                      </blockquote>
                      <p className="mt-2 text-[13px] leading-relaxed text-text-secondary">
                        {comment.body}
                      </p>
                      <div className="mt-2.5 flex items-center gap-1.5">
                        {comment.author.image ? (
                          // eslint-disable-next-line @next/next/no-img-element -- OAuth avatar host varies
                          <img src={comment.author.image} alt="" width={14} height={14} className="h-3.5 w-3.5 border border-line-structure object-cover" />
                        ) : (
                          <span className="flex h-3.5 w-3.5 items-center justify-center border border-line-cta bg-surface-cta-primary font-mono text-[8px] font-bold text-line-cta">
                            {(comment.author.name[0] ?? '?').toUpperCase()}
                          </span>
                        )}
                        <span className="truncate text-[11px] font-medium text-text-secondary">
                          {comment.author.name}
                        </span>
                        <span className="shrink-0 font-mono text-[9px] text-text-disabled">
                          {relativeTime(comment.createdAt)}
                        </span>
                        {comment.mine || viewerAdmin ? <DeleteButton comment={comment} /> : null}
                      </div>
                      {sectionIndex !== -1 ? (
                        <button
                          type="button"
                          onClick={() => jumpTo(comment)}
                          className="mt-2 font-mono text-[10px] uppercase tracking-widest text-line-cta transition-colors hover:text-text-primary"
                        >
                          Jump to text →
                        </button>
                      ) : null}
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </SheetContent>
      </Sheet>
    </>
  );
}
