import { Pool } from 'pg';
import { MIGRATIONS } from '@/lib/db/migrations';

/**
 * Arbitrary but stable advisory-lock key so concurrent server starts
 * (or CI workers sharing one database) never race the migration runner.
 */
const MIGRATION_LOCK_KEY = 813_002_931;

let pool: Pool | undefined;
let migrated: Promise<void> | undefined;

export function getPool(): Pool {
  pool ??= new Pool({
    connectionString: process.env.DATABASE_URL,
    max: 10,
  });
  return pool;
}

async function runMigrations(): Promise<void> {
  const client = await getPool().connect();
  try {
    await client.query('SELECT pg_advisory_lock($1)', [MIGRATION_LOCK_KEY]);
    await client.query(
      `CREATE TABLE IF NOT EXISTS schema_migrations (
         id INTEGER PRIMARY KEY,
         name TEXT NOT NULL,
         applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
       )`,
    );
    const { rows } = await client.query<{ id: number }>('SELECT id FROM schema_migrations');
    const applied = new Set(rows.map((row) => row.id));

    for (const migration of MIGRATIONS) {
      if (applied.has(migration.id)) continue;
      await client.query('BEGIN');
      try {
        await client.query(migration.sql);
        await client.query('INSERT INTO schema_migrations (id, name) VALUES ($1, $2)', [
          migration.id,
          migration.name,
        ]);
        await client.query('COMMIT');
      } catch (error) {
        await client.query('ROLLBACK');
        throw error;
      }
    }
  } finally {
    await client.query('SELECT pg_advisory_unlock($1)', [MIGRATION_LOCK_KEY]).catch(() => {});
    client.release();
  }
}

/**
 * Lazily applies pending migrations exactly once per process. Every
 * route handler that touches the database awaits this first, so a fresh
 * container becomes consistent on its first request without a separate
 * migration step in the deploy pipeline.
 */
export function ensureMigrated(): Promise<void> {
  migrated ??= runMigrations().catch((error) => {
    migrated = undefined;
    throw error;
  });
  return migrated;
}
