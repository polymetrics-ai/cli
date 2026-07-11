import type { BlogPost } from '@/lib/blog';

/**
 * W3C TextQuoteSelector-style anchor addressed to one block (a paragraph
 * or bullet point) of one section of a blog post. Blocks are plain
 * strings in `BLOG_POSTS`, so resolution is pure string search — no DOM.
 */
export type BlockType = 'body' | 'point';

export type Anchor = {
  sectionIndex: number;
  blockType: BlockType;
  blockIndex: number;
  /** The selected text itself. */
  exact: string;
  /** Up to CONTEXT_LENGTH chars of block text before the selection. */
  prefix: string;
  /** Up to CONTEXT_LENGTH chars of block text after the selection. */
  suffix: string;
  /** Char offset of `exact` within the block at capture time. */
  startOffset: number;
};

export type ResolvedRange = {
  sectionIndex: number;
  blockType: BlockType;
  blockIndex: number;
  start: number;
  end: number;
  orphaned: false;
};

export type Orphaned = { orphaned: true };

export type Resolution = ResolvedRange | Orphaned;

export const CONTEXT_LENGTH = 64;
export const EXACT_MAX_LENGTH = 1000;

export function blockText(post: BlogPost, blockType: BlockType, sectionIndex: number, blockIndex: number): string | undefined {
  const section = post.sections[sectionIndex];
  if (!section) return undefined;
  const blocks = blockType === 'body' ? section.body : section.points;
  return blocks?.[blockIndex];
}

/** Every (blockType, blockIndex) pair a section offers, in document order. */
function sectionBlocks(post: BlogPost, sectionIndex: number): Array<{ blockType: BlockType; blockIndex: number; text: string }> {
  const section = post.sections[sectionIndex];
  if (!section) return [];
  const blocks: Array<{ blockType: BlockType; blockIndex: number; text: string }> = [];
  section.body.forEach((text, blockIndex) => blocks.push({ blockType: 'body', blockIndex, text }));
  section.points?.forEach((text, blockIndex) => blocks.push({ blockType: 'point', blockIndex, text }));
  return blocks;
}

function occurrences(haystack: string, needle: string): number[] {
  const found: number[] = [];
  if (!needle) return found;
  let index = haystack.indexOf(needle);
  while (index !== -1) {
    found.push(index);
    index = haystack.indexOf(needle, index + 1);
  }
  return found;
}

function overlapScore(text: string, anchor: Anchor, start: number): number {
  const before = text.slice(Math.max(0, start - anchor.prefix.length), start);
  const after = text.slice(start + anchor.exact.length, start + anchor.exact.length + anchor.suffix.length);

  let score = 0;
  // Compare prefix right-to-left and suffix left-to-right: the chars
  // touching the selection are the strongest signal.
  for (let i = 1; i <= Math.min(before.length, anchor.prefix.length); i++) {
    if (before[before.length - i] === anchor.prefix[anchor.prefix.length - i]) score++;
    else break;
  }
  for (let i = 0; i < Math.min(after.length, anchor.suffix.length); i++) {
    if (after[i] === anchor.suffix[i]) score++;
    else break;
  }
  return score;
}

function bestInBlock(text: string, anchor: Anchor, location: { blockType: BlockType; blockIndex: number; sectionIndex: number }): ResolvedRange | undefined {
  const starts = occurrences(text, anchor.exact);
  if (starts.length === 0) return undefined;

  let bestStart = starts[0];
  if (starts.length > 1) {
    let bestScore = -1;
    for (const start of starts) {
      // Exact original offset wins outright among equals.
      const score = overlapScore(text, anchor, start) + (start === anchor.startOffset ? 0.5 : 0);
      if (score > bestScore) {
        bestScore = score;
        bestStart = start;
      }
    }
  }

  return {
    sectionIndex: location.sectionIndex,
    blockType: location.blockType,
    blockIndex: location.blockIndex,
    start: bestStart,
    end: bestStart + anchor.exact.length,
    orphaned: false,
  };
}

/**
 * Resolve an anchor against the current post content.
 *
 * Strategy, cheapest first:
 *  1. Exact offset hit in the addressed block.
 *  2. Occurrence scan in the addressed block, scored by prefix/suffix.
 *  3. Any block in the addressed section (content shifted locally).
 *  4. Any block in the post (content moved across sections).
 *  5. Orphaned — the quote no longer exists; callers keep the comment
 *     visible in list surfaces but render no highlight.
 */
export function resolveAnchor(post: BlogPost, anchor: Anchor): Resolution {
  if (!anchor.exact || anchor.exact.length > EXACT_MAX_LENGTH) return { orphaned: true };

  const addressed = blockText(post, anchor.blockType, anchor.sectionIndex, anchor.blockIndex);
  if (addressed !== undefined) {
    if (addressed.slice(anchor.startOffset, anchor.startOffset + anchor.exact.length) === anchor.exact) {
      return {
        sectionIndex: anchor.sectionIndex,
        blockType: anchor.blockType,
        blockIndex: anchor.blockIndex,
        start: anchor.startOffset,
        end: anchor.startOffset + anchor.exact.length,
        orphaned: false,
      };
    }
    const inBlock = bestInBlock(addressed, anchor, anchor);
    if (inBlock) return inBlock;
  }

  const inSection = sectionBlocks(post, anchor.sectionIndex)
    .map((block) => bestInBlock(block.text, anchor, { ...block, sectionIndex: anchor.sectionIndex }))
    .find(Boolean);
  if (inSection) return inSection;

  for (let sectionIndex = 0; sectionIndex < post.sections.length; sectionIndex++) {
    if (sectionIndex === anchor.sectionIndex) continue;
    const hit = sectionBlocks(post, sectionIndex)
      .map((block) => bestInBlock(block.text, anchor, { ...block, sectionIndex }))
      .find(Boolean);
    if (hit) return hit;
  }

  return { orphaned: true };
}

export type Segment = {
  text: string;
  /** Ids of the annotations covering this segment; empty = plain text. */
  ids: string[];
};

/**
 * Split a block's text into alternating plain/highlighted segments from a
 * set of resolved ranges. Overlaps are preserved (a segment can carry
 * several ids), which the highlight renderer uses to deepen the tint.
 */
export function segmentBlock(text: string, ranges: Array<{ id: string; start: number; end: number }>): Segment[] {
  const bounded = ranges
    .map((range) => ({
      id: range.id,
      start: Math.max(0, Math.min(range.start, text.length)),
      end: Math.max(0, Math.min(range.end, text.length)),
    }))
    .filter((range) => range.end > range.start);

  if (bounded.length === 0) return text ? [{ text, ids: [] }] : [];

  const cuts = new Set<number>([0, text.length]);
  for (const range of bounded) {
    cuts.add(range.start);
    cuts.add(range.end);
  }
  const points = [...cuts].sort((a, b) => a - b);

  const segments: Segment[] = [];
  for (let i = 0; i < points.length - 1; i++) {
    const start = points[i];
    const end = points[i + 1];
    if (end <= start) continue;
    const ids = bounded.filter((range) => range.start <= start && range.end >= end).map((range) => range.id);
    segments.push({ text: text.slice(start, end), ids });
  }
  return segments;
}

export function isValidAnchorShape(value: unknown): value is Anchor {
  if (typeof value !== 'object' || value === null) return false;
  const anchor = value as Record<string, unknown>;
  return (
    Number.isInteger(anchor.sectionIndex) &&
    (anchor.sectionIndex as number) >= 0 &&
    (anchor.blockType === 'body' || anchor.blockType === 'point') &&
    Number.isInteger(anchor.blockIndex) &&
    (anchor.blockIndex as number) >= 0 &&
    typeof anchor.exact === 'string' &&
    anchor.exact.length > 0 &&
    anchor.exact.length <= EXACT_MAX_LENGTH &&
    typeof anchor.prefix === 'string' &&
    anchor.prefix.length <= CONTEXT_LENGTH &&
    typeof anchor.suffix === 'string' &&
    anchor.suffix.length <= CONTEXT_LENGTH &&
    Number.isInteger(anchor.startOffset) &&
    (anchor.startOffset as number) >= 0
  );
}
