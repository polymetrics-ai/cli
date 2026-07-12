'use client';

import { useId, useMemo, useState } from 'react';
import { ChevronDown, LocateFixed, MessageSquare, Reply, Trash2 } from 'lucide-react';
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
import { buildCommentTree } from '@/lib/comments/comment-tree';
import type { CommentTreeNode } from '@/lib/comments/comment-tree';

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
  const fieldId = useId();
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
    <div className="mt-3 border-l-2 border-surface-cta-primary bg-surface-1 p-2.5">
      <label
        htmlFor={fieldId}
        className="mb-1.5 block font-mono text-[9px] uppercase tracking-widest text-text-disabled"
      >
        {placeholder.replace(/…$/, '')}
      </label>
      <textarea
        id={fieldId}
        name="reply"
        autoComplete="off"
        value={body}
        onChange={(event) => setBody(event.target.value.slice(0, REPLY_MAX))}
        rows={2}
        placeholder="Write a thoughtful reply…"
        onKeyDown={(event) => {
          if (event.key === 'Escape') onDone();
          if (event.key === 'Enter' && (event.metaKey || event.ctrlKey)) {
            event.preventDefault();
            void submit();
          }
        }}
        className="w-full resize-none border border-line-structure bg-surface-bg px-2.5 py-2 text-[13px] leading-relaxed text-text-primary outline-none placeholder:text-text-disabled focus:border-line-cta focus-visible:ring-1 focus-visible:ring-surface-cta-primary"
      />
      <div className="mt-1.5 flex items-center gap-2">
        <span className="mr-auto font-mono text-[9px] tabular-nums text-text-disabled">
          {body.length}/{REPLY_MAX}
        </span>
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

function ThreadReply({
  node,
  depth,
  rootId,
  replyingTo,
  viewerAdmin,
  onReply,
  onStopReply,
}: {
  node: CommentTreeNode<CommentDto>;
  depth: number;
  rootId: string;
  replyingTo: string | null;
  viewerAdmin: boolean;
  onReply: (id: string, rootId: string) => void;
  onStopReply: () => void;
}) {
  const { comment, children } = node;

  return (
    <article
      data-reply-id={comment.id}
      data-thread-depth={depth}
      className={comment.pending ? 'annotation-pending' : ''}
      aria-label={`Reply by ${comment.author.name}`}
    >
      <AuthorRow comment={comment} />
      <p className="mt-1.5 break-words text-[13px] leading-[1.65] text-text-secondary">
        {comment.body}
      </p>
      <div className="mt-2 flex min-h-6 items-center gap-3">
        <button
          type="button"
          aria-label={`Reply to ${comment.author.name}`}
          onClick={() => onReply(comment.id, rootId)}
          className="flex items-center gap-1 font-mono text-[9px] uppercase tracking-widest text-text-disabled transition-colors hover:text-line-cta focus-visible:outline focus-visible:outline-1 focus-visible:outline-offset-2 focus-visible:outline-line-cta"
        >
          <Reply className="h-2.5 w-2.5" aria-hidden="true" />
          Reply
        </button>
        {comment.mine || viewerAdmin ? <DeleteButton comment={comment} /> : null}
      </div>

      {replyingTo === comment.id ? (
        <ReplyComposer
          parentId={comment.id}
          placeholder={`Reply to ${comment.author.name}…`}
          onDone={onStopReply}
        />
      ) : null}

      {children.length > 0 ? (
        <div
          className={`comment-thread-children mt-3 flex flex-col gap-3 border-l border-line-structure ${
            depth >= 4 ? 'pl-2' : 'ml-2 pl-3'
          }`}
        >
          {children.map((child) => (
            <ThreadReply
              key={child.comment.id}
              node={child}
              depth={depth + 1}
              rootId={rootId}
              replyingTo={replyingTo}
              viewerAdmin={viewerAdmin}
              onReply={onReply}
              onStopReply={onStopReply}
            />
          ))}
        </div>
      ) : null}
    </article>
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
    replyCounts,
    viewerAdmin,
    signedIn,
    requestSignIn,
  } = useAnnotations();
  const [replyingTo, setReplyingTo] = useState<string | null>(null);
  const [expanded, setExpanded] = useState<Set<string>>(new Set());

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
        <SheetContent side="right" className="blog-comments-sheet overflow-y-auto border-l border-line-structure bg-surface-bg p-0">
          <SheetHeader className="border-b border-line-structure bg-surface-1 px-5 py-4">
            <p className="font-mono text-[10px] uppercase tracking-widest text-text-disabled">
              Reader conversation
            </p>
            <div className="flex items-center gap-2 pr-8">
              <SheetTitle className="font-square text-[18px] font-semibold text-text-primary">
                Notes on this article
              </SheetTitle>
              <span className="border border-line-structure bg-surface-bg px-1.5 py-0.5 font-mono text-[10px] tabular-nums text-line-cta">
                {count}
              </span>
            </div>
            <SheetDescription className="text-[12px] text-text-tertiary">
              Select a passage to add a note, or open a thread below to read every reply.
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
                    const replyTree = buildCommentTree(comments, comment.id);
                    const visibleReplyCount = replyCounts.get(comment.id) ?? 0;
                    const isExpanded = expanded.has(comment.id);
                    return (
                      <div key={comment.id} className="corner-box-corners border border-line-structure bg-surface-bg p-3">
                        {comment.anchor ? (
                          <blockquote className="truncate border-l-2 border-surface-cta-primary bg-surface-1 px-2.5 py-1 text-[12px] italic text-text-tertiary">
                            {comment.anchor.exact}
                          </blockquote>
                        ) : null}
                        <div className="mt-2.5">
                          <AuthorRow comment={comment}>
                            {comment.mine || viewerAdmin ? <DeleteButton comment={comment} /> : null}
                          </AuthorRow>
                        </div>
                        <p className="mt-2 break-words text-[14px] leading-[1.65] text-text-primary">
                          {comment.body}
                        </p>

                        <div className="mt-2.5 flex min-h-7 flex-wrap items-center gap-3 border-t border-line-structure pt-2">
                          {sectionIndex !== -1 ? (
                            <button
                              type="button"
                              onClick={() => jumpTo(comment)}
                              className="flex items-center gap-1 font-mono text-[10px] uppercase tracking-widest text-line-cta transition-colors hover:text-text-primary"
                            >
                              <LocateFixed className="h-3 w-3" aria-hidden="true" />
                              View passage
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
                          {visibleReplyCount > 0 ? (
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
                              {visibleReplyCount} {visibleReplyCount === 1 ? 'reply' : 'replies'}
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

                        {isExpanded && replyTree.length > 0 ? (
                          <div className="mt-3 flex flex-col gap-3 border-l-2 border-surface-cta-primary bg-surface-1/50 py-1 pl-3">
                            {replyTree.map((node) => (
                              <ThreadReply
                                key={node.comment.id}
                                node={node}
                                depth={1}
                                rootId={comment.id}
                                replyingTo={replyingTo}
                                viewerAdmin={viewerAdmin}
                                onReply={startReply}
                                onStopReply={() => setReplyingTo(null)}
                              />
                            ))}
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
