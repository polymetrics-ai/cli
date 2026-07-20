import { auth } from '@/lib/auth';
import { ensureMigrated } from '@/lib/db';

export type SessionUser = {
  id: string;
  name: string;
  email: string;
  image: string | null;
  isAdmin: boolean;
};

function adminEmails(): Set<string> {
  return new Set(
    (process.env.ADMIN_EMAILS ?? '')
      .split(',')
      .map((email) => email.trim().toLowerCase())
      .filter(Boolean),
  );
}

/**
 * Single seam between route handlers and Better Auth: handlers depend on
 * this module (and tests mock it) instead of auth internals.
 */
export async function getSessionUser(requestHeaders: Headers): Promise<SessionUser | null> {
  await ensureMigrated();
  const session = await auth.api.getSession({ headers: requestHeaders });
  if (!session) return null;

  const email = session.user.email.toLowerCase();
  return {
    id: session.user.id,
    name: session.user.name,
    email,
    image: session.user.image ?? null,
    isAdmin: adminEmails().has(email),
  };
}
