'use client';

import { useEffect, useRef, useState } from 'react';
import { createPortal } from 'react-dom';
import { useAnnotations } from '@/components/blog/annotations-provider';

const BODY_MAX = 2000;
const WIDTH = 340;

export function CommentComposer() {
  const { draft, composing, closeComposer, submitComment } = useAnnotations();
  const [body, setBody] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    if (composing) textareaRef.current?.focus();
    else setBody('');
  }, [composing]);

  if (!composing || !draft) return null;

  const anchoredTop = draft.rect.height > 0 ? draft.rect.top + draft.rect.height + 12 : 120;
  const top = Math.min(anchoredTop, window.innerHeight - 260);
  const left =
    draft.rect.width > 0
      ? Math.min(Math.max(draft.rect.left, 16), window.innerWidth - WIDTH - 16)
      : Math.max((window.innerWidth - WIDTH) / 2, 16);

  const remaining = BODY_MAX - body.length;

  async function submit() {
    const trimmed = body.trim();
    if (!trimmed || submitting) return;
    setSubmitting(true);
    await submitComment(trimmed);
    setSubmitting(false);
  }

  return createPortal(
    <div
      className="corner-box-corners annotation-popover-enter fixed z-50 border border-line-cta bg-surface-bg shadow-[0_18px_60px_rgba(12,31,23,0.16)]"
      style={{ top, left, width: WIDTH }}
      role="dialog"
      aria-label="New note"
      onKeyDown={(event) => {
        if (event.key === 'Escape') {
          event.stopPropagation();
          closeComposer();
        }
        if (event.key === 'Enter' && (event.metaKey || event.ctrlKey)) {
          event.preventDefault();
          void submit();
        }
      }}
    >
      <div className="border-b border-line-structure px-4 py-3">
        <p className="font-mono text-[10px] uppercase tracking-widest text-text-disabled">New note</p>
        <blockquote className="mt-2 line-clamp-2 border-l-2 border-surface-cta-primary bg-surface-1 px-3 py-1.5 text-[13px] italic leading-relaxed text-text-tertiary">
          {draft.anchor.exact}
        </blockquote>
      </div>

      <div className="px-4 py-3">
        <textarea
          ref={textareaRef}
          value={body}
          onChange={(event) => setBody(event.target.value.slice(0, BODY_MAX))}
          rows={4}
          placeholder="Add your note…"
          className="w-full resize-none border border-line-structure bg-surface-bg px-3 py-2 text-[14px] leading-relaxed text-text-primary outline-none placeholder:text-text-disabled focus:border-line-cta"
        />
        <div className="mt-2 flex items-center justify-between">
          <span
            className={`font-mono text-[10px] tracking-wider ${
              remaining < 100 ? 'text-destructive' : 'text-text-disabled'
            }`}
          >
            {body.length}/{BODY_MAX}
          </span>
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={closeComposer}
              className="px-3 py-1.5 font-square text-[11px] font-semibold uppercase tracking-wider text-text-tertiary transition-colors hover:text-text-primary"
            >
              Cancel
            </button>
            <button
              type="button"
              disabled={!body.trim() || submitting}
              onClick={() => void submit()}
              className="border border-line-cta bg-line-cta px-3.5 py-1.5 font-square text-[11px] font-semibold uppercase tracking-wider text-surface-bg transition-opacity hover:opacity-90 disabled:pointer-events-none disabled:opacity-50"
            >
              {submitting ? 'Posting…' : 'Post note'}
            </button>
          </div>
        </div>
      </div>
    </div>,
    document.body,
  );
}
