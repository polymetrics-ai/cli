'use client';

import { Fragment, useEffect, useMemo, useState } from 'react';
import { createPortal } from 'react-dom';
import { MessageSquareText, X } from 'lucide-react';
import { segmentBlock } from '@/lib/annotations/anchor';
import type { BlockType } from '@/lib/annotations/anchor';
import type { BlogEvidence } from '@/lib/blog';
import { useAnnotations } from '@/components/blog/annotations-provider';

const BOOKMARK_PREFIX = 'bookmark:';
const EVIDENCE_PREFIX = 'evidence:';

type InlineEvidenceReference = {
  evidence: BlogEvidence;
  number: number;
  text: string;
};

/**
 * Renders one block's text with `<mark>` highlights for every comment
 * (and dotted underlines for the reader's own private bookmarks) whose
 * anchor resolved into this block. Pure string slicing — the block text
 * is the source of truth, so server HTML and client HTML always agree.
 */
export function HighlightedBlock({
  text,
  sectionIndex,
  blockType,
  blockIndex,
  evidenceReferences = [],
  onEvidenceOpen,
}: {
  text: string;
  sectionIndex: number;
  blockType: BlockType;
  blockIndex: number;
  evidenceReferences?: InlineEvidenceReference[];
  onEvidenceOpen?: (evidence: BlogEvidence, trigger: HTMLElement) => void;
}) {
  const { comments, bookmarks, resolutions, activeId, setActiveId, setHovered } = useAnnotations();

  const { segments, evidenceRanges } = useMemo(() => {
    const ranges: Array<{ id: string; start: number; end: number }> = [];
    for (const comment of comments) {
      const resolution = resolutions.get(comment.id);
      if (
        resolution &&
        !resolution.orphaned &&
        resolution.sectionIndex === sectionIndex &&
        resolution.blockType === blockType &&
        resolution.blockIndex === blockIndex
      ) {
        ranges.push({ id: comment.id, start: resolution.start, end: resolution.end });
      }
    }
    for (const bookmark of bookmarks) {
      const resolution = resolutions.get(bookmark.id);
      if (
        resolution &&
        !resolution.orphaned &&
        resolution.sectionIndex === sectionIndex &&
        resolution.blockType === blockType &&
        resolution.blockIndex === blockIndex
      ) {
        ranges.push({ id: BOOKMARK_PREFIX + bookmark.id, start: resolution.start, end: resolution.end });
      }
    }
    const nextEvidenceRanges = evidenceReferences.flatMap((reference) => {
      const start = text.indexOf(reference.text);
      if (start < 0) return [];
      const id = EVIDENCE_PREFIX + reference.evidence.id;
      const range = { id, start, end: start + reference.text.length, reference };
      ranges.push(range);
      return [range];
    });
    return { segments: segmentBlock(text, ranges), evidenceRanges: nextEvidenceRanges };
  }, [
    text,
    sectionIndex,
    blockType,
    blockIndex,
    comments,
    bookmarks,
    resolutions,
    evidenceReferences,
  ]);

  const commentsById = useMemo(() => new Map(comments.map((c) => [c.id, c])), [comments]);

  let offset = 0;

  const evidenceLink = (
    reference: InlineEvidenceReference,
    label: string,
    triggerLabel: string,
    className: string,
  ) => (
    <a
      href={reference.evidence.url}
      target="_blank"
      rel="noreferrer"
      aria-label={triggerLabel}
      className={className}
      onClick={(event) => {
        if (event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return;
        event.preventDefault();
        onEvidenceOpen?.(reference.evidence, event.currentTarget);
      }}
    >
      {label}
    </a>
  );

  return (
    <>
      {segments.map((segment, index) => {
        const start = offset;
        const end = start + segment.text.length;
        offset = end;
        const commentIds = segment.ids.filter(
          (id) => !id.startsWith(BOOKMARK_PREFIX) && !id.startsWith(EVIDENCE_PREFIX),
        );
        const hasBookmark = segment.ids.some((id) => id.startsWith(BOOKMARK_PREFIX));
        const evidenceRange = evidenceRanges.find(
          (range) => segment.ids.includes(range.id) && start >= range.start && end <= range.end,
        );
        const endingEvidence = evidenceRanges.filter((range) => range.end === end);

        let content;

        if (commentIds.length === 0 && !hasBookmark) {
          content = evidenceRange
            ? evidenceLink(
                evidenceRange.reference,
                segment.text,
                `Preview ${evidenceRange.reference.evidence.label} evidence`,
                'font-medium text-line-cta underline decoration-line-structure underline-offset-[3px] transition-colors hover:decoration-line-cta focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-line-cta',
              )
            : <span>{segment.text}</span>;
        } else if (commentIds.length === 0) {
          // Own private bookmark: quiet dotted underline, no background.
          const bookmarkIds = segment.ids
            .filter((id) => id.startsWith(BOOKMARK_PREFIX))
            .map((id) => id.slice(BOOKMARK_PREFIX.length))
            .join(' ');
          content = (
            <span
              className="border-b border-dotted border-surface-cta-primary decoration-clone"
              data-annotation-bookmark={bookmarkIds}
            >
              {segment.text}
            </span>
          );
        } else {
          const primary = commentIds[0];
          const comment = commentsById.get(primary);
          const isActive = commentIds.includes(activeId ?? '');
          const depth = commentIds.length > 1;

          content = (
            <mark
              id={`annotation-${primary}`}
              data-annotation-mark={commentIds.join(' ')}
              tabIndex={0}
              role="button"
              aria-label={`View note by ${comment?.author.name ?? 'a reader'}`}
              aria-pressed={isActive}
              className={[
                'cursor-pointer border-b text-inherit transition-colors duration-150 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-line-cta',
                hasBookmark ? 'border-dotted' : '',
                isActive
                  ? 'bg-surface-cta-primary/40 border-line-cta'
                  : depth
                    ? 'bg-surface-cta-primary/35 border-surface-cta-primary hover:bg-surface-cta-primary/45'
                    : 'bg-surface-cta-primary/20 border-surface-cta-primary hover:bg-surface-cta-primary/35',
                comment?.pending ? 'annotation-pending' : '',
              ].join(' ')}
              onMouseEnter={(event) => setHovered({ id: primary, rect: event.currentTarget.getBoundingClientRect() })}
              onMouseLeave={() => setHovered(null)}
              onFocus={(event) => setHovered({ id: primary, rect: event.currentTarget.getBoundingClientRect() })}
              onBlur={() => setHovered(null)}
              onClick={() => setActiveId(isActive ? null : primary)}
              onKeyDown={(event) => {
                if (event.key === 'Enter' || event.key === ' ') {
                  event.preventDefault();
                  setActiveId(isActive ? null : primary);
                }
              }}
            >
              {segment.text}
            </mark>
          );
        }

        return (
          <Fragment key={index}>
            {content}
            {endingEvidence.map(({ reference }) => (
              <sup key={reference.evidence.id} className="ml-0.5 align-super leading-none">
                {evidenceLink(
                  reference,
                  `[${reference.number}]`,
                  `Open citation ${reference.number}: ${reference.evidence.label}`,
                  'font-mono text-[0.68em] font-semibold text-line-cta no-underline transition-colors hover:bg-surface-cta-primary/25 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-line-cta',
                )}
              </sup>
            ))}
          </Fragment>
        );
      })}
    </>
  );
}

/** Relative "3d ago" stamp, deterministic enough for hover surfaces. */
export function relativeTime(iso: string): string {
  const then = new Date(iso).getTime();
  const minutes = Math.round((Date.now() - then) / 60_000);
  if (minutes < 1) return 'just now';
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.round(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.round(hours / 24);
  if (days < 30) return `${days}d ago`;
  return new Date(iso).toISOString().slice(0, 10);
}

/**
 * Hovering gives a lightweight preview. Activating a highlight pins the
 * complete note so readers can move into the card and open its thread.
 */
export function HoverPreview() {
  const {
    hovered,
    comments,
    activeId,
    setActiveId,
    setSheetOpen,
    replyCounts,
  } = useAnnotations();
  const [activeRect, setActiveRect] = useState<DOMRect | null>(null);

  useEffect(() => {
    if (!activeId) {
      setActiveRect(null);
      return;
    }

    const update = () => {
      const mark = document.querySelector<HTMLElement>(
        `[data-annotation-mark~="${CSS.escape(activeId)}"]`,
      );
      setActiveRect(mark?.getBoundingClientRect() ?? null);
    };
    update();
    window.addEventListener('resize', update);
    window.addEventListener('scroll', update, { passive: true, capture: true });
    return () => {
      window.removeEventListener('resize', update);
      window.removeEventListener('scroll', update, true);
    };
  }, [activeId]);

  useEffect(() => {
    if (!activeId) return;
    const dismiss = (event: KeyboardEvent) => {
      if (event.key === 'Escape') setActiveId(null);
    };
    window.addEventListener('keydown', dismiss);
    return () => window.removeEventListener('keydown', dismiss);
  }, [activeId, setActiveId]);

  const pinned = Boolean(activeId && activeRect);
  const previewId = pinned ? activeId : hovered?.id;
  const rect = pinned ? activeRect : hovered?.rect;
  const comment = comments.find((candidate) => candidate.id === previewId);
  if (!comment || !rect) return null;
  const replies = replyCounts.get(comment.id) ?? 0;

  const width = pinned ? Math.min(340, window.innerWidth - 32) : Math.min(280, window.innerWidth - 32);
  const placeBeside = pinned && rect.right + width + 12 <= window.innerWidth - 16;
  const left = placeBeside
    ? rect.right + 12
    : Math.min(Math.max(rect.left, 16), window.innerWidth - width - 16);
  const placeAbove = !placeBeside && rect.top > (pinned ? 260 : 180);
  const top = placeBeside
    ? Math.min(Math.max(16, rect.top - 12), Math.max(16, window.innerHeight - 240))
    : placeAbove
      ? rect.top - 10
      : rect.bottom + 10;

  return createPortal(
    <div
      className={`corner-box-corners fixed z-[70] border bg-surface-bg shadow-[0_18px_60px_rgba(12,31,23,0.18)] ${
        pinned ? '' : 'annotation-popover-enter'
      } ${
        pinned
          ? 'max-h-[min(420px,calc(100vh-2rem))] overflow-y-auto overscroll-contain border-line-cta p-4'
          : 'pointer-events-none border-line-structure p-3'
      } ${placeAbove ? '-translate-y-full' : ''}`}
      style={{ top, left, width }}
      role={pinned ? 'dialog' : 'tooltip'}
      aria-label={pinned ? `Note by ${comment.author.name}` : undefined}
      data-note-preview={pinned ? 'pinned' : 'hover'}
    >
      <div className="flex items-center gap-2">
        {comment.author.image ? (
          // eslint-disable-next-line @next/next/no-img-element -- OAuth avatar host varies
          <img src={comment.author.image} alt="" width={16} height={16} className="h-4 w-4 border border-line-structure object-cover" />
        ) : (
          <span className="flex h-4 w-4 items-center justify-center border border-line-cta bg-surface-cta-primary font-mono text-[9px] font-bold text-line-cta">
            {(comment.author.name[0] ?? '?').toUpperCase()}
          </span>
        )}
        <span className="truncate text-[12px] font-medium text-text-primary">{comment.author.name}</span>
        <span className="ml-auto shrink-0 font-mono text-[10px] text-text-disabled">
          {relativeTime(comment.createdAt)}
        </span>
        {pinned ? (
          <button
            type="button"
            onClick={() => setActiveId(null)}
            aria-label="Close note"
            className="ml-1 flex size-7 shrink-0 items-center justify-center border border-line-structure text-text-tertiary transition-colors hover:border-line-cta hover:bg-surface-1 hover:text-text-primary focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-line-cta"
          >
            <X className="size-3.5" aria-hidden="true" />
          </button>
        ) : null}
      </div>
      <p
        className={`mt-2 break-words text-[13px] leading-relaxed ${
          pinned ? 'text-text-secondary' : 'line-clamp-2 text-text-tertiary'
        }`}
      >
        {comment.body}
      </p>
      {pinned ? (
        <div className="mt-3 flex items-center gap-3 border-t border-line-structure pt-3">
          <button
            type="button"
            onClick={() => {
              setSheetOpen(true);
              setActiveId(null);
            }}
            className="flex items-center gap-1.5 border border-line-cta bg-line-cta px-2.5 py-1.5 font-mono text-[10px] uppercase tracking-widest text-surface-bg transition-opacity hover:opacity-90 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-line-cta"
          >
            <MessageSquareText className="size-3" aria-hidden="true" />
            Open thread
          </button>
          <span className="font-mono text-[9px] uppercase tracking-widest text-text-disabled">
            {replies} {replies === 1 ? 'reply' : 'replies'}
          </span>
        </div>
      ) : replies > 0 ? (
        <p className="mt-2 font-mono text-[9px] uppercase tracking-widest text-text-disabled">
          Click to read · {replies} {replies === 1 ? 'reply' : 'replies'}
        </p>
      ) : (
        <p className="mt-2 font-mono text-[9px] uppercase tracking-widest text-text-disabled">
          Click to keep open
        </p>
      )}
    </div>,
    document.body,
  );
}
