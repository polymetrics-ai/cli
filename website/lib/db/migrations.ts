export type Migration = {
  id: number;
  name: string;
  sql: string;
};

/**
 * Migrations are embedded as strings (not .sql files) so the standalone
 * server bundle never depends on filesystem layout at runtime.
 *
 * The auth tables mirror Better Auth's core schema for the built-in
 * Kysely adapter (camelCase, quoted identifiers). Do not rename columns
 * without regenerating via `npx @better-auth/cli generate`.
 */
export const MIGRATIONS: Migration[] = [
  {
    id: 1,
    name: 'init-auth-and-annotations',
    sql: `
CREATE TABLE IF NOT EXISTS "user" (
  "id" TEXT PRIMARY KEY,
  "name" TEXT NOT NULL,
  "email" TEXT NOT NULL UNIQUE,
  "emailVerified" BOOLEAN NOT NULL DEFAULT FALSE,
  "image" TEXT,
  "createdAt" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  "updatedAt" TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS "session" (
  "id" TEXT PRIMARY KEY,
  "expiresAt" TIMESTAMPTZ NOT NULL,
  "token" TEXT NOT NULL UNIQUE,
  "createdAt" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  "updatedAt" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  "ipAddress" TEXT,
  "userAgent" TEXT,
  "userId" TEXT NOT NULL REFERENCES "user" ("id") ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS "session_userId_idx" ON "session" ("userId");

CREATE TABLE IF NOT EXISTS "account" (
  "id" TEXT PRIMARY KEY,
  "accountId" TEXT NOT NULL,
  "providerId" TEXT NOT NULL,
  "userId" TEXT NOT NULL REFERENCES "user" ("id") ON DELETE CASCADE,
  "accessToken" TEXT,
  "refreshToken" TEXT,
  "idToken" TEXT,
  "accessTokenExpiresAt" TIMESTAMPTZ,
  "refreshTokenExpiresAt" TIMESTAMPTZ,
  "scope" TEXT,
  "password" TEXT,
  "createdAt" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  "updatedAt" TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS "account_userId_idx" ON "account" ("userId");

CREATE TABLE IF NOT EXISTS "verification" (
  "id" TEXT PRIMARY KEY,
  "identifier" TEXT NOT NULL,
  "value" TEXT NOT NULL,
  "expiresAt" TIMESTAMPTZ NOT NULL,
  "createdAt" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  "updatedAt" TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS "verification_identifier_idx" ON "verification" ("identifier");

-- Annotation anchor columns follow the W3C TextQuoteSelector model:
-- the selected text plus surrounding context, addressed to one block
-- (paragraph or bullet) of one section of a post. Anchors are resolved
-- client-side; rows are never invalidated by content edits.
CREATE TABLE IF NOT EXISTS comment (
  id TEXT PRIMARY KEY,
  post_slug TEXT NOT NULL,
  user_id TEXT NOT NULL REFERENCES "user" ("id") ON DELETE CASCADE,
  body TEXT NOT NULL,
  parent_id TEXT,
  section_index INTEGER NOT NULL,
  block_type TEXT NOT NULL,
  block_index INTEGER NOT NULL,
  exact TEXT NOT NULL,
  prefix TEXT NOT NULL DEFAULT '',
  suffix TEXT NOT NULL DEFAULT '',
  start_offset INTEGER NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS comment_slug_idx ON comment (post_slug, created_at);

CREATE TABLE IF NOT EXISTS bookmark (
  id TEXT PRIMARY KEY,
  post_slug TEXT NOT NULL,
  user_id TEXT NOT NULL REFERENCES "user" ("id") ON DELETE CASCADE,
  section_index INTEGER NOT NULL,
  block_type TEXT NOT NULL,
  block_index INTEGER NOT NULL,
  exact TEXT NOT NULL,
  prefix TEXT NOT NULL DEFAULT '',
  suffix TEXT NOT NULL DEFAULT '',
  start_offset INTEGER NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT bookmark_unique_anchor
    UNIQUE (user_id, post_slug, section_index, block_index, start_offset, exact)
);
CREATE INDEX IF NOT EXISTS bookmark_user_idx ON bookmark (user_id, created_at);
`,
  },
  {
    id: 2,
    name: 'threaded-replies',
    sql: `
-- Replies carry no anchor of their own — they inherit the root note's
-- thread. A row is either a root (anchored) or a reply (parented).
ALTER TABLE comment ALTER COLUMN section_index DROP NOT NULL;
ALTER TABLE comment ALTER COLUMN block_type DROP NOT NULL;
ALTER TABLE comment ALTER COLUMN block_index DROP NOT NULL;
ALTER TABLE comment ALTER COLUMN exact DROP NOT NULL;
ALTER TABLE comment ALTER COLUMN prefix DROP NOT NULL;
ALTER TABLE comment ALTER COLUMN suffix DROP NOT NULL;
ALTER TABLE comment ALTER COLUMN start_offset DROP NOT NULL;

ALTER TABLE comment
  ADD CONSTRAINT comment_parent_fk
  FOREIGN KEY (parent_id) REFERENCES comment (id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS comment_parent_idx ON comment (parent_id);

ALTER TABLE comment
  ADD CONSTRAINT comment_shape CHECK (
    parent_id IS NOT NULL OR (
      section_index IS NOT NULL AND block_type IS NOT NULL AND block_index IS NOT NULL
      AND exact IS NOT NULL AND prefix IS NOT NULL AND suffix IS NOT NULL
      AND start_offset IS NOT NULL
    )
  );
`,
  },
];
