import { Pool } from 'pg';

/**
 * Annotation e2e scenarios assert on margin-rail geometry (cards vs
 * cluster badges), which depends on how many notes already exist. Start
 * every run from a clean slate. CI databases are fresh anyway; locally
 * this resets the throwaway dev container's annotations only — users,
 * sessions, and bookmarks-page data are recreated by the specs.
 */
export default async function globalSetup(): Promise<void> {
  if (!process.env.DATABASE_URL) return;
  const pool = new Pool({ connectionString: process.env.DATABASE_URL });
  try {
    await pool.query('DELETE FROM comment');
    await pool.query('DELETE FROM bookmark');
  } catch {
    // Tables may not exist yet on a brand-new database — the app
    // migrates on first request, and the specs tolerate empty state.
  } finally {
    await pool.end();
  }
}
