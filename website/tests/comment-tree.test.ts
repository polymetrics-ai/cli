import { describe, expect, it } from 'vitest';

import { buildCommentTree } from '@/lib/comments/comment-tree';

type TestComment = {
  id: string;
  parentId: string | null;
  createdAt: string;
  body: string;
};

function comment(id: string, parentId: string | null, order: number): TestComment {
  return {
    id,
    parentId,
    createdAt: new Date(Date.UTC(2026, 0, 1, 0, order)).toISOString(),
    body: id,
  };
}

describe('buildCommentTree', () => {
  it('keeps every reply beneath its direct parent at arbitrary depth', () => {
    const comments = [
      comment('nested', 'reply', 3),
      comment('root', null, 0),
      comment('sibling', 'root', 2),
      comment('reply', 'root', 1),
    ];

    const tree = buildCommentTree(comments, 'root');

    expect(tree.map((node) => node.comment.id)).toEqual(['reply', 'sibling']);
    expect(tree[0].children.map((node) => node.comment.id)).toEqual(['nested']);
    expect(tree[0].children[0].children).toEqual([]);
  });

  it('ignores replies whose parent is outside the selected root thread', () => {
    const comments = [
      comment('root', null, 0),
      comment('reply', 'root', 1),
      comment('other-root', null, 2),
      comment('other-reply', 'other-root', 3),
    ];

    expect(buildCommentTree(comments, 'root').map((node) => node.comment.id)).toEqual(['reply']);
  });
});
