import { randomUUID } from 'node:crypto';
import { ensureMigrated, getPool } from '@/lib/db';
import type { Anchor } from '@/lib/annotations/anchor';

export type BookmarkRecord = {
  id: string;
  postSlug: string;
  userId: string;
  anchor: Anchor;
  createdAt: string;
};

type BookmarkRow = {
  id: string;
  post_slug: string;
  user_id: string;
  section_index: number;
  block_type: 'body' | 'point';
  block_index: number;
  exact: string;
  prefix: string;
  suffix: string;
  start_offset: number;
  created_at: Date;
};

function toRecord(row: BookmarkRow): BookmarkRecord {
  return {
    id: row.id,
    postSlug: row.post_slug,
    userId: row.user_id,
    anchor: {
      sectionIndex: row.section_index,
      blockType: row.block_type,
      blockIndex: row.block_index,
      exact: row.exact,
      prefix: row.prefix,
      suffix: row.suffix,
      startOffset: row.start_offset,
    },
    createdAt: row.created_at.toISOString(),
  };
}

export async function listBookmarks(userId: string, postSlug?: string): Promise<BookmarkRecord[]> {
  await ensureMigrated();
  const { rows } = postSlug
    ? await getPool().query<BookmarkRow>(
        'SELECT * FROM bookmark WHERE user_id = $1 AND post_slug = $2 ORDER BY created_at DESC',
        [userId, postSlug],
      )
    : await getPool().query<BookmarkRow>(
        'SELECT * FROM bookmark WHERE user_id = $1 ORDER BY created_at DESC',
        [userId],
      );
  return rows.map(toRecord);
}

/**
 * Idempotent: re-bookmarking the same selection returns the existing row
 * (the popover toggle can fire twice without duplicating).
 */
export async function insertBookmark(input: {
  postSlug: string;
  userId: string;
  anchor: Anchor;
}): Promise<{ bookmark: BookmarkRecord; created: boolean }> {
  await ensureMigrated();
  const id = randomUUID();
  const inserted = await getPool().query<BookmarkRow>(
    `INSERT INTO bookmark
       (id, post_slug, user_id, section_index, block_type, block_index,
        exact, prefix, suffix, start_offset)
     VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
     ON CONFLICT ON CONSTRAINT bookmark_unique_anchor DO NOTHING
     RETURNING *`,
    [
      id,
      input.postSlug,
      input.userId,
      input.anchor.sectionIndex,
      input.anchor.blockType,
      input.anchor.blockIndex,
      input.anchor.exact,
      input.anchor.prefix,
      input.anchor.suffix,
      input.anchor.startOffset,
    ],
  );
  if (inserted.rows[0]) return { bookmark: toRecord(inserted.rows[0]), created: true };

  const { rows } = await getPool().query<BookmarkRow>(
    `SELECT * FROM bookmark
     WHERE user_id = $1 AND post_slug = $2 AND section_index = $3
       AND block_index = $4 AND start_offset = $5 AND exact = $6`,
    [
      input.userId,
      input.postSlug,
      input.anchor.sectionIndex,
      input.anchor.blockIndex,
      input.anchor.startOffset,
      input.anchor.exact,
    ],
  );
  return { bookmark: toRecord(rows[0]), created: false };
}

export async function getBookmarkOwner(id: string): Promise<string | undefined> {
  await ensureMigrated();
  const { rows } = await getPool().query<{ user_id: string }>(
    'SELECT user_id FROM bookmark WHERE id = $1',
    [id],
  );
  return rows[0]?.user_id;
}

export async function deleteBookmark(id: string): Promise<void> {
  await ensureMigrated();
  await getPool().query('DELETE FROM bookmark WHERE id = $1', [id]);
}
