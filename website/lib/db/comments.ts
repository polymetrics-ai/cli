import { randomUUID } from 'node:crypto';
import { ensureMigrated, getPool } from '@/lib/db';
import type { Anchor } from '@/lib/annotations/anchor';

export type CommentRecord = {
  id: string;
  postSlug: string;
  userId: string;
  body: string;
  /** Null for replies — they inherit the root note's thread anchor. */
  anchor: Anchor | null;
  parentId: string | null;
  createdAt: string;
  author: { name: string; image: string | null };
};

type CommentRow = {
  id: string;
  post_slug: string;
  user_id: string;
  body: string;
  parent_id: string | null;
  section_index: number | null;
  block_type: 'body' | 'point' | null;
  block_index: number | null;
  exact: string | null;
  prefix: string | null;
  suffix: string | null;
  start_offset: number | null;
  created_at: Date;
  author_name: string;
  author_image: string | null;
};

function toRecord(row: CommentRow): CommentRecord {
  const anchor: Anchor | null =
    row.exact !== null &&
    row.section_index !== null &&
    row.block_type !== null &&
    row.block_index !== null &&
    row.start_offset !== null
      ? {
          sectionIndex: row.section_index,
          blockType: row.block_type,
          blockIndex: row.block_index,
          exact: row.exact,
          prefix: row.prefix ?? '',
          suffix: row.suffix ?? '',
          startOffset: row.start_offset,
        }
      : null;

  return {
    id: row.id,
    postSlug: row.post_slug,
    userId: row.user_id,
    body: row.body,
    anchor,
    parentId: row.parent_id,
    createdAt: row.created_at.toISOString(),
    author: { name: row.author_name, image: row.author_image },
  };
}

const SELECT = `
  SELECT c.id, c.post_slug, c.user_id, c.body, c.parent_id,
         c.section_index, c.block_type, c.block_index,
         c.exact, c.prefix, c.suffix, c.start_offset, c.created_at,
         u."name" AS author_name, u."image" AS author_image
  FROM comment c
  JOIN "user" u ON u."id" = c.user_id
`;

export async function listCommentsBySlug(postSlug: string): Promise<CommentRecord[]> {
  await ensureMigrated();
  const { rows } = await getPool().query<CommentRow>(
    `${SELECT} WHERE c.post_slug = $1 ORDER BY c.created_at ASC`,
    [postSlug],
  );
  return rows.map(toRecord);
}

export async function insertComment(input: {
  postSlug: string;
  userId: string;
  body: string;
  anchor?: Anchor;
  parentId?: string;
}): Promise<CommentRecord> {
  await ensureMigrated();
  const id = randomUUID();
  await getPool().query(
    `INSERT INTO comment
       (id, post_slug, user_id, body, parent_id, section_index, block_type, block_index,
        exact, prefix, suffix, start_offset)
     VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
    [
      id,
      input.postSlug,
      input.userId,
      input.body,
      input.parentId ?? null,
      input.anchor?.sectionIndex ?? null,
      input.anchor?.blockType ?? null,
      input.anchor?.blockIndex ?? null,
      input.anchor?.exact ?? null,
      input.anchor?.prefix ?? null,
      input.anchor?.suffix ?? null,
      input.anchor?.startOffset ?? null,
    ],
  );
  const { rows } = await getPool().query<CommentRow>(`${SELECT} WHERE c.id = $1`, [id]);
  return toRecord(rows[0]);
}

export async function getCommentMeta(
  id: string,
): Promise<{ userId: string; postSlug: string } | undefined> {
  await ensureMigrated();
  const { rows } = await getPool().query<{ user_id: string; post_slug: string }>(
    'SELECT user_id, post_slug FROM comment WHERE id = $1',
    [id],
  );
  return rows[0] ? { userId: rows[0].user_id, postSlug: rows[0].post_slug } : undefined;
}

export async function getCommentOwner(id: string): Promise<string | undefined> {
  return (await getCommentMeta(id))?.userId;
}

/** Replies are removed by the parent_id ON DELETE CASCADE. */
export async function deleteComment(id: string): Promise<void> {
  await ensureMigrated();
  await getPool().query('DELETE FROM comment WHERE id = $1', [id]);
}
