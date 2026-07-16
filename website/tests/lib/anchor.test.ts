import { describe, expect, it } from 'vitest';
import { resolveAnchor, segmentBlock, isValidAnchorShape } from '@/lib/annotations/anchor';
import type { Anchor } from '@/lib/annotations/anchor';
import type { BlogPost } from '@/lib/blog';

const post = {
  slug: 'test-post',
  title: 'Test',
  description: '',
  publishedAt: '2026-01-01',
  updatedAt: '2026-01-01',
  readingTime: '1 min read',
  category: 'Test',
  tags: [],
  summary: '',
  sections: [
    {
      heading: 'First section',
      body: [
        'The quick brown fox jumps over the lazy dog.',
        'A second paragraph mentions the quick brown fox again for ambiguity.',
      ],
      points: ['Keep the loop small.', 'Ship one binary.'],
    },
    {
      heading: 'Second section',
      body: ['Entirely different content lives here about connectors.'],
    },
  ],
} satisfies BlogPost;

function anchor(overrides: Partial<Anchor>): Anchor {
  return {
    sectionIndex: 0,
    blockType: 'body',
    blockIndex: 0,
    exact: 'quick brown fox',
    prefix: 'The ',
    suffix: ' jumps',
    startOffset: 4,
    ...overrides,
  };
}

describe('resolveAnchor', () => {
  it('hits the exact stored offset (fast path)', () => {
    const result = resolveAnchor(post, anchor({}));
    expect(result).toEqual({
      sectionIndex: 0,
      blockType: 'body',
      blockIndex: 0,
      start: 4,
      end: 19,
      orphaned: false,
    });
  });

  it('re-finds text whose offset shifted after an edit', () => {
    const result = resolveAnchor(post, anchor({ startOffset: 30 }));
    expect(result).toMatchObject({ start: 4, end: 19, orphaned: false });
  });

  it('disambiguates duplicate quotes by prefix/suffix context', () => {
    // "quick brown fox" appears in body[0] and body[1]; context from
    // body[1] ("the … again") must select the second occurrence there.
    const result = resolveAnchor(
      post,
      anchor({
        blockIndex: 1,
        prefix: 'mentions the ',
        suffix: ' again for',
        startOffset: 32,
      }),
    );
    expect(result).toMatchObject({ blockIndex: 1, orphaned: false });
    const text = post.sections[0].body[1];
    if ('start' in result) {
      expect(text.slice(result.start, result.end)).toBe('quick brown fox');
    }
  });

  it('falls back to sibling blocks in the addressed section', () => {
    // Anchor addresses body[0] but the quote only exists in points[1].
    const result = resolveAnchor(
      post,
      anchor({ exact: 'one binary', prefix: 'Ship ', suffix: '.', startOffset: 5 }),
    );
    expect(result).toMatchObject({ blockType: 'point', blockIndex: 1, orphaned: false });
  });

  it('falls back to other sections when content moved', () => {
    const result = resolveAnchor(
      post,
      anchor({ exact: 'about connectors', prefix: 'lives here ', suffix: '.', startOffset: 0 }),
    );
    expect(result).toMatchObject({ sectionIndex: 1, blockType: 'body', orphaned: false });
  });

  it('orphans when the quote no longer exists anywhere', () => {
    expect(resolveAnchor(post, anchor({ exact: 'vanished text' }))).toEqual({ orphaned: true });
  });

  it('orphans on out-of-range section addressing with missing quote', () => {
    expect(resolveAnchor(post, anchor({ sectionIndex: 9, exact: 'vanished text' }))).toEqual({
      orphaned: true,
    });
  });
});

describe('segmentBlock', () => {
  const text = 'abcdefghij';

  it('returns one plain segment when nothing is highlighted', () => {
    expect(segmentBlock(text, [])).toEqual([{ text, ids: [] }]);
  });

  it('splits around a single range', () => {
    expect(segmentBlock(text, [{ id: 'a', start: 2, end: 5 }])).toEqual([
      { text: 'ab', ids: [] },
      { text: 'cde', ids: ['a'] },
      { text: 'fghij', ids: [] },
    ]);
  });

  it('merges overlapping ranges into multi-id segments', () => {
    const segments = segmentBlock(text, [
      { id: 'a', start: 1, end: 6 },
      { id: 'b', start: 4, end: 9 },
    ]);
    expect(segments).toEqual([
      { text: 'a', ids: [] },
      { text: 'bcd', ids: ['a'] },
      { text: 'ef', ids: ['a', 'b'] },
      { text: 'ghi', ids: ['b'] },
      { text: 'j', ids: [] },
    ]);
  });

  it('clamps out-of-bounds ranges instead of throwing', () => {
    expect(segmentBlock(text, [{ id: 'a', start: 8, end: 40 }])).toEqual([
      { text: 'abcdefgh', ids: [] },
      { text: 'ij', ids: ['a'] },
    ]);
  });
});

describe('isValidAnchorShape', () => {
  it('accepts a well-formed anchor', () => {
    expect(isValidAnchorShape(anchor({}))).toBe(true);
  });

  it.each([
    ['missing exact', anchor({ exact: '' })],
    ['oversized exact', anchor({ exact: 'x'.repeat(1001) })],
    ['oversized prefix', anchor({ prefix: 'x'.repeat(65) })],
    ['negative offset', anchor({ startOffset: -1 })],
    ['bad block type', { ...anchor({}), blockType: 'code' }],
    ['non-object', 'nope'],
  ])('rejects %s', (_label, value) => {
    expect(isValidAnchorShape(value)).toBe(false);
  });
});
