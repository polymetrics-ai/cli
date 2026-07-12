'use client';

import { useMemo } from 'react';
import { segmentBlock } from '@/lib/annotations/anchor';
import type { BlockType } from '@/lib/annotations/anchor';
import { useAnnotations } from '@/components/blog/annotations-provider';

const BOOKMARK_PREFIX = 'bookmark:';

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
}: {
  text: string;
  sectionIndex: number;
  blockType: BlockType;
  blockIndex: number;
}) {
  const { comments, bookmarks, resolutions, activeId, setActiveId, setHovered } = useAnnotations();

  const segments = useMemo(() => {
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
    return segmentBlock(text, ranges);
  }, [text, sectionIndex, blockType, blockIndex, comments, bookmarks, resolutions]);

  const commentsById = useMemo(() => new Map(comments.map((c) => [c.id, c])), [comments]);

  return (
    <>
      {segments.map((segment, index) => {
        const commentIds = segment.ids.filter((id) => !id.startsWith(BOOKMARK_PREFIX));
        const hasBookmark = segment.ids.some((id) => id.startsWith(BOOKMARK_PREFIX));

        if (commentIds.length === 0 && !hasBookmark) {
          return <span key={index}>{segment.text}</span>;
        }

        if (commentIds.length === 0) {
          // Own private bookmark: quiet dotted underline, no background.
          const bookmarkIds = segment.ids
            .filter((id) => id.startsWith(BOOKMARK_PREFIX))
            .map((id) => id.slice(BOOKMARK_PREFIX.length))
            .join(' ');
          return (
            <span
              key={index}
              className="border-b border-dotted border-[#34d399] decoration-clone"
              data-annotation-bookmark={bookmarkIds}
            >
              {segment.text}
            </span>
          );
        }

        const primary = commentIds[0];
        const comment = commentsById.get(primary);
        const isActive = commentIds.includes(activeId ?? '');
        const depth = commentIds.length > 1;

        return (
          <mark
            key={index}
            id={`annotation-${primary}`}
            data-annotation-mark={commentIds.join(' ')}
            tabIndex={0}
            role="button"
            aria-label={`View note by ${comment?.author.name ?? 'a reader'}`}
            className={[
              'cursor-pointer border-b text-inherit transition-colors duration-150',
              hasBookmark ? 'border-dotted' : '',
              isActive
                ? 'bg-[#34d399]/40 border-[#0f3d2e]'
                : depth
                  ? 'bg-[#34d399]/35 border-[#34d399] hover:bg-[#34d399]/45'
                  : 'bg-[#34d399]/20 border-[#34d399] hover:bg-[#34d399]/35',
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
 * Floating preview card shown while hovering/focusing a highlight —
 * author, time, and the first lines of the note.
 */
export function HoverPreview() {
  const { hovered, comments, setActiveId } = useAnnotations();
  if (!hovered) return null;
  const comment = comments.find((c) => c.id === hovered.id);
  if (!comment) return null;

  const top = hovered.rect.top - 8;
  const left = Math.min(Math.max(hovered.rect.left, 16), window.innerWidth - 296);

  return (
    <div
      className="corner-box-corners pointer-events-none fixed z-50 w-[280px] -translate-y-full border border-line-structure bg-surface-bg p-3 shadow-[4px_4px_0_0_#c0d6c8]"
      style={{ top, left }}
      role="tooltip"
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
      </div>
      <p className="mt-2 line-clamp-2 text-[13px] leading-relaxed text-text-tertiary">{comment.body}</p>
      <button
        type="button"
        tabIndex={-1}
        onClick={() => setActiveId(comment.id)}
        className="mt-2 font-mono text-[10px] uppercase tracking-widest text-line-cta"
      >
        View note →
      </button>
    </div>
  );
}
