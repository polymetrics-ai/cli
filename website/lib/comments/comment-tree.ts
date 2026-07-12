export type ThreadableComment = {
  id: string;
  parentId: string | null;
  createdAt: string;
};

export type CommentTreeNode<T extends ThreadableComment> = {
  comment: T;
  children: CommentTreeNode<T>[];
};

/** Build one root's reply tree while guarding against malformed cycles. */
export function buildCommentTree<T extends ThreadableComment>(
  comments: readonly T[],
  rootId: string,
): CommentTreeNode<T>[] {
  const childrenByParent = new Map<string, T[]>();

  for (const comment of comments) {
    if (!comment.parentId) continue;
    const children = childrenByParent.get(comment.parentId) ?? [];
    children.push(comment);
    childrenByParent.set(comment.parentId, children);
  }

  for (const children of childrenByParent.values()) {
    children.sort((left, right) => left.createdAt.localeCompare(right.createdAt));
  }

  const visited = new Set([rootId]);
  const visit = (parentId: string): CommentTreeNode<T>[] => {
    const nodes: CommentTreeNode<T>[] = [];
    for (const comment of childrenByParent.get(parentId) ?? []) {
      if (visited.has(comment.id)) continue;
      visited.add(comment.id);
      nodes.push({ comment, children: visit(comment.id) });
    }
    return nodes;
  };

  return visit(rootId);
}
