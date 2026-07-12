'use client';

import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { MessageSquare } from 'lucide-react';
import { useAnnotations } from '@/components/blog/annotations-provider';
import { relativeTime } from '@/components/blog/highlight-text';
import type { CommentDto } from '@/components/blog/annotations-provider';

const CARD_HEIGHT = 104;
const GAP = 8;
const CLUSTER_DRIFT = 140;

type Placed =
  | { kind: 'note'; comment: CommentDto; top: number }
  | { kind: 'cluster'; comments: CommentDto[]; top: number };

/**
 * Medium-style margin notes: each comment card sits beside its highlight,
 * pushed down to avoid overlaps; runs that drift too far from their
 * targets collapse into a "+N notes" badge that opens the sheet.
 * Positions are measured from the live DOM (highlight offsetTop relative
 * to the grid container) and re-measured on resize and font load.
 */
export function MarginNotesRail({ containerRef }: { containerRef: React.RefObject<HTMLDivElement | null> }) {
  const {
    comments,
    resolutions,
    activeId,
    setActiveId,
    setSheetOpen,
    signedIn,
    requestSignIn,
    loading,
  } = useAnnotations();
  const [targets, setTargets] = useState<Map<string, number>>(new Map());
  const [minTop, setMinTop] = useState(0);
  const railRef = useRef<HTMLDivElement>(null);
  const wrapperRef = useRef<HTMLDivElement>(null);

  const measure = useCallback(() => {
    const container = containerRef.current;
    if (!container) return;
    const containerTop = container.getBoundingClientRect().top;
    const next = new Map<string, number>();
    for (const comment of comments) {
      if (comment.pending) continue;
      const resolution = resolutions.get(comment.id);
      if (!resolution || resolution.orphaned) continue;
      const mark = document.getElementById(`annotation-${comment.id}`);
      if (!mark) continue;
      next.set(comment.id, mark.getBoundingClientRect().top - containerTop);
    }
    setTargets(next);
    const wrapper = wrapperRef.current;
    if (wrapper) setMinTop(Math.max(0, wrapper.getBoundingClientRect().top - containerTop));
  }, [comments, resolutions, containerRef]);

  useEffect(() => {
    measure();
    const container = containerRef.current;
    if (!container) return;
    const observer = new ResizeObserver(measure);
    observer.observe(container);
    // Chakra Petch / Geist load after hydration and shift line boxes.
    document.fonts?.ready.then(measure).catch(() => {});
    return () => observer.disconnect();
  }, [measure, containerRef]);

  const placed = useMemo<Placed[]>(() => {
    // Work in wrapper coordinates: a note's ideal top is its highlight's
    // offset minus the wrapper's own offset, clamped so notes never rise
    // above the rail (the Summary card owns the top of the column).
    const anchored = comments
      .filter((comment) => targets.has(comment.id))
      .sort((a, b) => (targets.get(a.id) ?? 0) - (targets.get(b.id) ?? 0));

    const result: Placed[] = [];
    let cursor = 0;
    let index = 0;
    while (index < anchored.length) {
      const comment = anchored[index];
      const target = Math.max(0, (targets.get(comment.id) ?? 0) - minTop);
      const top = Math.max(target, cursor);

      // If this card would sit far below its highlight, gather the whole
      // crowded run into one badge at the run's first target instead.
      if (top - target > CLUSTER_DRIFT) {
        const cluster: CommentDto[] = [comment];
        while (
          index + 1 < anchored.length &&
          Math.max(0, (targets.get(anchored[index + 1].id) ?? 0) - minTop) < cursor
        ) {
          index += 1;
          cluster.push(anchored[index]);
        }
        result.push({ kind: 'cluster', comments: cluster, top });
        cursor = top + 40 + GAP;
        index += 1;
        continue;
      }

      result.push({ kind: 'note', comment, top });
      cursor = top + CARD_HEIGHT + GAP;
      index += 1;
    }
    return result;
  }, [comments, targets, minTop]);

  const orphaned = comments.filter((comment) => {
    const resolution = resolutions.get(comment.id);
    return resolution?.orphaned && !comment.pending;
  });

  const anchoredCount = placed.reduce(
    (count, item) => count + (item.kind === 'note' ? 1 : item.comments.length),
    0,
  );

  function activate(comment: CommentDto) {
    setActiveId(comment.id);
    document.getElementById(`annotation-${comment.id}`)?.scrollIntoView({
      block: 'center',
      behavior: matchMedia('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth',
    });
  }

  return (
    <div ref={railRef} className="blog-margin-rail mt-6" data-margin-rail>
      <div className="flex items-center justify-between border-b border-line-structure pb-2">
        <p className="font-mono text-[10px] uppercase tracking-widest text-text-disabled">
          Marginalia
        </p>
        <button
          type="button"
          onClick={() => setSheetOpen(true)}
          aria-label="Open all notes"
          className="flex items-center gap-1.5 border border-line-structure bg-surface-1 px-1.5 py-0.5 font-mono text-[10px] text-text-tertiary transition-colors hover:border-line-cta hover:text-text-primary"
        >
          <MessageSquare className="h-3 w-3" aria-hidden="true" />
          {comments.filter((c) => !c.pending).length}
        </button>
      </div>

      {!loading && comments.length === 0 ? (
        <div className="with-stripes mt-3 border border-line-structure p-3">
          <p className="bg-surface-bg p-2 text-[12px] leading-relaxed text-text-tertiary">
            Select any passage to leave the first note.
            {!signedIn ? (
              <button
                type="button"
                onClick={requestSignIn}
                className="mt-1 block font-mono text-[10px] uppercase tracking-widest text-line-cta"
              >
                Sign in to join →
              </button>
            ) : null}
          </p>
        </div>
      ) : null}

      {/* Anchored notes float alongside their highlights; the wrapper is
          position:absolute within the article grid, offset by the rail's
          own position so coordinates line up. */}
      <div ref={wrapperRef} className="relative" style={{ height: 0 }}>
        {placed.map((item) =>
          item.kind === 'note' ? (
            <button
              key={item.comment.id}
              type="button"
              onClick={() => activate(item.comment)}
              data-margin-note={item.comment.id}
              className={`corner-box-corners absolute w-full border bg-surface-bg p-2.5 text-left transition-[top,border-color] duration-150 ease-out motion-reduce:transition-none ${
                activeId === item.comment.id
                  ? 'z-10 border-line-cta shadow-[3px_3px_0_0_#c0d6c8]'
                  : 'border-line-structure hover:border-line-cta'
              }`}
              style={{ top: item.top }}
            >
              <span className="flex items-center gap-1.5">
                {item.comment.author.image ? (
                  // eslint-disable-next-line @next/next/no-img-element -- OAuth avatar host varies
                  <img src={item.comment.author.image} alt="" width={14} height={14} className="h-3.5 w-3.5 border border-line-structure object-cover" />
                ) : (
                  <span className="flex h-3.5 w-3.5 items-center justify-center border border-line-cta bg-surface-cta-primary font-mono text-[8px] font-bold text-line-cta">
                    {(item.comment.author.name[0] ?? '?').toUpperCase()}
                  </span>
                )}
                <span className="truncate text-[11px] font-medium text-text-secondary">
                  {item.comment.author.name}
                </span>
                <span className="ml-auto shrink-0 font-mono text-[9px] text-text-disabled">
                  {relativeTime(item.comment.createdAt)}
                </span>
              </span>
              <span className="mt-1.5 line-clamp-3 block text-[12px] leading-snug text-text-tertiary">
                {item.comment.body}
              </span>
            </button>
          ) : (
            <button
              key={`cluster-${item.comments[0].id}`}
              type="button"
              onClick={() => setSheetOpen(true)}
              className="absolute border border-[#34d399] bg-surface-1 px-2 py-1 font-mono text-[10px] uppercase tracking-widest text-line-cta transition-colors hover:bg-surface-2"
              style={{ top: item.top }}
            >
              +{item.comments.length} notes
            </button>
          ),
        )}
        {/* Spacer so the rail contributes its own height to the column. */}
        <div style={{ height: placedHeight(placed) }} aria-hidden />
      </div>

      {orphaned.length > 0 ? (
        <div className="mt-4 border-t border-line-structure pt-3">
          <p className="font-mono text-[9px] uppercase tracking-widest text-text-disabled">
            Context changed
          </p>
          {orphaned.map((comment) => (
            <div key={comment.id} className="with-stripes mt-2 border border-line-structure p-2">
              <blockquote className="bg-surface-bg px-2 py-1 text-[11px] italic leading-snug text-text-disabled">
                “{comment.anchor.exact.slice(0, 80)}
                {comment.anchor.exact.length > 80 ? '…' : ''}”
              </blockquote>
              <p className="mt-1 line-clamp-2 px-2 text-[12px] leading-snug text-text-tertiary">
                {comment.body}
              </p>
            </div>
          ))}
        </div>
      ) : null}

      <span className="sr-only">{anchoredCount} anchored notes</span>
    </div>
  );
}

function placedHeight(placed: Placed[]): number {
  let bottom = 0;
  for (const item of placed) {
    const height = item.kind === 'note' ? CARD_HEIGHT : 32;
    bottom = Math.max(bottom, item.top + height);
  }
  return bottom;
}
