'use client';

import { useCallback, useEffect, useRef, useState } from 'react';
import { createPortal } from 'react-dom';
import { Bookmark, MessageSquare } from 'lucide-react';
import { CONTEXT_LENGTH, EXACT_MAX_LENGTH } from '@/lib/annotations/anchor';
import type { Anchor, BlockType } from '@/lib/annotations/anchor';
import { stashDraft, useAnnotations } from '@/components/blog/annotations-provider';

/**
 * Reads the live selection and, when it sits inside exactly one
 * annotatable block, converts it to a text-quote anchor.
 */
function anchorFromSelection(): { anchor: Anchor; rect: DOMRect } | null {
  const selection = window.getSelection();
  if (!selection || selection.isCollapsed || selection.rangeCount === 0) return null;

  const range = selection.getRangeAt(0);
  const startBlock = closestBlock(range.startContainer);
  const endBlock = closestBlock(range.endContainer);
  if (!startBlock || startBlock !== endBlock) return null;

  const exact = range.toString();
  if (!exact.trim() || exact.length > EXACT_MAX_LENGTH) return null;

  // Offset of the selection start within the block's full text — measured
  // through the DOM so existing <mark>/<span> splits don't skew it.
  const probe = document.createRange();
  probe.selectNodeContents(startBlock);
  probe.setEnd(range.startContainer, range.startOffset);
  const startOffset = probe.toString().length;

  const blockText = startBlock.textContent ?? '';
  const anchor: Anchor = {
    sectionIndex: Number(startBlock.getAttribute('data-section-index')),
    blockType: (startBlock.getAttribute('data-block-type') ?? 'body') as BlockType,
    blockIndex: Number(startBlock.getAttribute('data-block-index')),
    exact,
    prefix: blockText.slice(Math.max(0, startOffset - CONTEXT_LENGTH), startOffset),
    suffix: blockText.slice(startOffset + exact.length, startOffset + exact.length + CONTEXT_LENGTH),
    startOffset,
  };
  if (Number.isNaN(anchor.sectionIndex) || Number.isNaN(anchor.blockIndex)) return null;

  return { anchor, rect: range.getBoundingClientRect() };
}

function closestBlock(node: Node): Element | null {
  const element = node instanceof Element ? node : node.parentElement;
  return element?.closest('[data-annotation-block]') ?? null;
}

const POPOVER_HEIGHT = 38;

export function SelectionPopover() {
  const {
    slug,
    signedIn,
    draft,
    setDraft,
    composing,
    openComposer,
    toggleBookmark,
    isDraftBookmarked,
    requestSignIn,
  } = useAnnotations();
  const [mounted, setMounted] = useState(false);
  const skipNextClear = useRef(false);

  useEffect(() => setMounted(true), []);

  const capture = useCallback(() => {
    const result = anchorFromSelection();
    if (result) {
      skipNextClear.current = true;
      setDraft(result);
    }
  }, [setDraft]);

  useEffect(() => {
    function onMouseUp() {
      // Wait a tick so the selection has settled.
      window.setTimeout(capture, 0);
    }
    function onKeyUp(event: KeyboardEvent) {
      if (event.key === 'Shift' || event.key.startsWith('Arrow')) window.setTimeout(capture, 0);
      if (event.key === 'Escape') setDraft(null);
    }
    function onSelectionChange() {
      if (skipNextClear.current) {
        skipNextClear.current = false;
        return;
      }
      const selection = window.getSelection();
      if (!selection || selection.isCollapsed) setDraft(null);
    }
    function onScroll() {
      // Position is viewport-fixed; hide rather than track while scrolling.
      setDraft(null);
    }

    document.addEventListener('mouseup', onMouseUp);
    document.addEventListener('keyup', onKeyUp);
    document.addEventListener('selectionchange', onSelectionChange);
    window.addEventListener('scroll', onScroll, { passive: true });
    return () => {
      document.removeEventListener('mouseup', onMouseUp);
      document.removeEventListener('keyup', onKeyUp);
      document.removeEventListener('selectionchange', onSelectionChange);
      window.removeEventListener('scroll', onScroll);
    };
  }, [capture, setDraft]);

  if (!mounted || !draft || composing || draft.rect.width === 0) return null;

  const flip = draft.rect.top < POPOVER_HEIGHT + 16;
  const top = flip
    ? draft.rect.top + draft.rect.height + 10
    : draft.rect.top - POPOVER_HEIGHT - 10;
  const centered = draft.rect.left + draft.rect.width / 2;
  const left = Math.min(Math.max(centered - 92, 12), window.innerWidth - 196);

  function handleBookmark() {
    if (!signedIn) {
      if (draft) stashDraft({ slug, anchor: draft.anchor });
      requestSignIn();
      return;
    }
    void toggleBookmark();
  }

  function handleComment() {
    if (!signedIn) {
      if (draft) stashDraft({ slug, anchor: draft.anchor });
      requestSignIn();
      return;
    }
    openComposer();
  }

  const buttonCls =
    'flex items-center gap-1.5 px-3 py-2 font-square text-[11px] font-semibold uppercase tracking-wider text-text-secondary transition-colors hover:bg-surface-1 hover:text-text-primary';

  return createPortal(
    <div
      role="toolbar"
      aria-label="Annotate selection"
      className="corner-box-corners annotation-popover-enter fixed z-50 flex border border-line-cta bg-surface-bg shadow-[4px_4px_0_0_#c0d6c8]"
      style={{ top, left }}
    >
      <button type="button" className={buttonCls} onClick={handleBookmark}>
        <Bookmark
          className={`h-3.5 w-3.5 ${isDraftBookmarked ? 'fill-[#34d399] text-line-cta' : ''}`}
          aria-hidden="true"
        />
        {isDraftBookmarked ? 'Saved' : 'Bookmark'}
      </button>
      <span aria-hidden className="w-px self-stretch bg-line-structure" />
      <button type="button" className={buttonCls} onClick={handleComment}>
        <MessageSquare className="h-3.5 w-3.5" aria-hidden="true" />
        Comment
      </button>
      <span
        aria-hidden
        className={`absolute left-1/2 h-2 w-2 -translate-x-1/2 rotate-45 border-line-cta bg-surface-bg ${
          flip ? '-top-1 border-l border-t' : '-bottom-1 border-b border-r'
        }`}
      />
    </div>,
    document.body,
  );
}
