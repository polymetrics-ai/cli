'use client';

import { useMemo, useState } from 'react';
import { ChevronDown, MessageSquare, Reply, Trash2 } from 'lucide-react';
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet';
import { Skeleton } from '@/components/shadcn/ui/skeleton';
import { useAnnotations } from '@/components/blog/annotations-provider';
import { AuthorChip } from '@/components/blog/author-chip';
import { relativeTime } from '@/components/blog/highlight-text';
import type { CommentDto } from '@/components/blog/annotations-provider';
import type { BlogSection } from '@/lib/blog';

const REPLY_MAX = 2000;

function DeleteButton({ comment }: { comment: CommentDto }) {
  const { deleteComment } = useAnnotations();
  const [confirming, setConfirming] = useState(false);

  if (!confirming) {
    return (
      <button
        type="button"
        aria-label="Delete note"
        onClick={() => setConfirming(true)}
        className="ml-auto flex items-center gap-1 font-mono text-[9px] uppercase tracking-widest text-text-disabled transition-colors hover:text-destructive"
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
        className="border border-destructive bg-destructive px-1.5 py-0.5 text-white"
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

function AuthorRow({ comment, children }: { comment: CommentDto; children?: React.ReactNode }) {
  return (
    <div className="flex items-center gap-1.5">
      <AuthorChip author={comment.author} />
      <span className="shrink-0 font-mono text-[9px] text-text-disabled">
        {relativeTime(comment.createdAt)}
      </span>
      {children}
    </div>
  );
}

function ReplyComposer({
  parentId,
  placeholder,
  onDone,
}: {
  parentId: string;
  placeholder: string;
  onDone: () => void;
}) {
  const { submitReply } = useAnnotations();
  const [body, setBody] = useState('');
  const [submitting, setSubmitting] = useState(false);

  async function submit() {
    const trimmed = body.trim();
    if (!trimmed || submitting) return;
    setSubmitting(true);
    const ok = await submitReply(parentId, trimmed);
    setSubmitting(false);
    if (ok) {
      setBody('');
      onDone();
    }
  }

  return (
    <div className="mt-2">
      <textarea
        autoFocus
        value={body}
        onChange={(event) => setBody(event.target.value.slice(0, REPLY_MAX))}
        rows={2}
        placeholder={placeholder}
        onKeyDown={(event) => {
          if (event.key === 'Escape') onDone();
          if (event.key === 'Enter' && (event.metaKey || event.ctrlKey)) {
            event.preventDefault();
            void submit();
          }
        }}
        className="w-full resize-none border border-line-structure bg-surface-bg px-2.5 py-1.5 text-[13px] leading-relaxed text-text-primary outline-none placeholder:text-text-disabled focus:border-line-cta"
      />
      <div className="mt-1.5 flex items-center justify-end gap-2">
        <button
          type="button"
          onClick={onDone}
          className="px-2 py-1 font-square text-[10px] font-semibold uppercase tracking-wider text-text-tertiary transition-colors hover:text-text-primary"
        >
          Cancel
        </button>
        <button
          type="button"
          disabled={!body.trim() || submitting}
          onClick={() => void submit()}
          className="border border-line-cta bg-line-cta px-2.5 py-1 font-square text-[10px] font-semibold uppercase tracking-wider text-surface-bg transition-opacity hover:opacity-90 disabled:pointer-events-none disabled:opacity-50"
        >
          {submitting ? 'Posting…' : 'Post reply'}
        </button>
      </div>
    </div>
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
  const [replyingTo, setReplyingTo] = useState<string | null>(null);
  const [expanded, setExpanded] = useState<Set<string>>(new Set());

  const byId = useMemo(() => new Map(comments.map((c) => [c.id, c])), [comments]);

  /** Root ancestor for any comment (replies may nest arbitrarily deep). */
  const rootOf = useMemo(() => {
    const map = new Map<string, string>();
    for (const comment of comments) {
      let cursor = comment;
      let guard = 0;
      while (cursor.parentId && guard < 100) {
        const parent = byId.get(cursor.parentId);
        if (!parent) break;
        cursor = parent;
        guard += 1;
      }
      map.set(comment.id, cursor.id);
    }
    return map;
  }, [comments, byId]);

  const repliesByRoot = useMemo(() => {
    const map = new Map<string, CommentDto[]>();
    for (const comment of comments) {
      if (!comment.parentId) continue;
      const root = rootOf.get(comment.id) ?? comment.parentId;
      const list = map.get(root) ?? [];
      list.push(comment);
      map.set(root, list);
    }
    for (const list of map.values()) {
      list.sort((a, b) => a.createdAt.localeCompare(b.createdAt));
    }
    return map;
  }, [comments, rootOf]);

  const grouped = useMemo(() => {
    const groups = new Map<number, CommentDto[]>();
    for (const comment of comments) {
      if (comment.pending || comment.parentId) continue;
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

  function startReply(id: string, rootId: string) {
    if (!signedIn) {
      setSheetOpen(false);
      requestSignIn();
      return;
    }
    setReplyingTo(id);
    // Keep the thread open so the new reply is visible once it posts.
    setExpanded((current) => new Set(current).add(rootId));
  }

  const count = comments.filter((c) => !c.pending).length;

  return (
    <>
      {/* Floating trigger — shown wherever the margin rail is hidden (<80rem). */}
      <button
        type="button"
        onClick={() => setSheetOpen(true)}
        aria-label={`Open notes (${count})`}
        className="blog-notes-fab fixed bottom-5 right-5 z-40 items-center gap-2 border border-line-cta bg-surface-bg px-3 py-2.5 shadow-[0_18px_60px_rgba(12,31,23,0.16)] transition-colors hover:bg-surface-1"
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
                  {list.map((comment) => {
                    const replies = repliesByRoot.get(comment.id) ?? [];
                    const visibleReplies = replies.filter((r) => !r.pending);
                    const isExpanded = expanded.has(comment.id) || replies.some((r) => r.pending);
                    return (
                      <div key={comment.id} className="corner-box-corners border border-line-structure bg-surface-bg p-3">
                        {comment.anchor ? (
                          <blockquote className="truncate border-l-2 border-surface-cta-primary bg-surface-1 px-2.5 py-1 text-[12px] italic text-text-tertiary">
                            {comment.anchor.exact}
                          </blockquote>
                        ) : null}
                        <p className="mt-2 text-[13px] leading-relaxed text-text-secondary">
                          {comment.body}
                        </p>
                        <div className="mt-2.5">
                          <AuthorRow comment={comment}>
                            {comment.mine || viewerAdmin ? <DeleteButton comment={comment} /> : null}
                          </AuthorRow>
                        </div>

                        <div className="mt-2 flex items-center gap-3">
                          {sectionIndex !== -1 ? (
                            <button
                              type="button"
                              onClick={() => jumpTo(comment)}
                              className="font-mono text-[10px] uppercase tracking-widest text-line-cta transition-colors hover:text-text-primary"
                            >
                              Jump to text →
                            </button>
                          ) : null}
                          <button
                            type="button"
                            onClick={() => startReply(comment.id, comment.id)}
                            className="flex items-center gap-1 font-mono text-[10px] uppercase tracking-widest text-text-tertiary transition-colors hover:text-text-primary"
                          >
                            <Reply className="h-3 w-3" aria-hidden="true" />
                            Reply
                          </button>
                          {visibleReplies.length > 0 ? (
                            <button
                              type="button"
                              aria-expanded={isExpanded}
                              onClick={() =>
                                setExpanded((current) => {
                                  const next = new Set(current);
                                  if (next.has(comment.id)) next.delete(comment.id);
                                  else next.add(comment.id);
                                  return next;
                                })
                              }
                              className="ml-auto flex items-center gap-1 border border-line-structure bg-surface-1 px-1.5 py-0.5 font-mono text-[10px] uppercase tracking-widest text-line-cta transition-colors hover:bg-surface-2"
                            >
                              <ChevronDown
                                className={`h-3 w-3 transition-transform motion-reduce:transition-none ${isExpanded ? 'rotate-180' : ''}`}
                                aria-hidden="true"
                              />
                              {visibleReplies.length} {visibleReplies.length === 1 ? 'reply' : 'replies'}
                            </button>
                          ) : null}
                        </div>

                        {replyingTo === comment.id ? (
                          <ReplyComposer
                            parentId={comment.id}
                            placeholder={`Reply to ${comment.author.name}…`}
                            onDone={() => setReplyingTo(null)}
                          />
                        ) : null}

                        {isExpanded && replies.length > 0 ? (
                          <div className="mt-3 flex flex-col gap-2.5 border-l border-line-structure pl-3">
                            {replies.map((reply) => {
                              const directParent =
                                reply.parentId && reply.parentId !== comment.id
                                  ? byId.get(reply.parentId)
                                  : null;
                              return (
                                <div
                                  key={reply.id}
                                  data-reply-id={reply.id}
                                  className={reply.pending ? 'annotation-pending' : ''}
                                >
                                  <p className="text-[13px] leading-relaxed text-text-secondary">
                                    {directParent ? (
                                      <span className="mr-1 font-mono text-[11px] text-line-cta">
                                        @{directParent.author.name}
                                      </span>
                                    ) : null}
                                    {reply.body}
                                  </p>
                                  <div className="mt-1.5">
                                    <AuthorRow comment={reply}>
                                      <button
                                        type="button"
                                        aria-label={`Reply to ${reply.author.name}`}
                                        onClick={() => startReply(reply.id, comment.id)}
                                        className="ml-2 flex items-center gap-1 font-mono text-[9px] uppercase tracking-widest text-text-disabled transition-colors hover:text-text-primary"
                                      >
                                        <Reply className="h-2.5 w-2.5" aria-hidden="true" />
                                        Reply
                                      </button>
                                      {reply.mine || viewerAdmin ? <DeleteButton comment={reply} /> : null}
                                    </AuthorRow>
                                  </div>
                                  {replyingTo === reply.id ? (
                                    <ReplyComposer
                                      parentId={reply.id}
                                      placeholder={`Reply to ${reply.author.name}…`}
                                      onDone={() => setReplyingTo(null)}
                                    />
                                  ) : null}
                                </div>
                              );
                            })}
                          </div>
                        ) : null}
                      </div>
                    );
                  })}
                </div>
              </div>
            ))}
          </div>
        </SheetContent>
      </Sheet>
    </>
  );
}
