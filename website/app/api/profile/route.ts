import { NextResponse } from 'next/server';
import { getSessionUser } from '@/lib/auth-session';
import {
  getGithubAccountId,
  getProfileSettings,
  setProviderProfile,
  upsertProfileSettings,
} from '@/lib/db/profile';

export const dynamic = 'force-dynamic';

const PROFILE_URL_MAX = 200;

function error(message: string, status: number) {
  return NextResponse.json({ error: message }, { status });
}

export async function GET(request: Request) {
  const user = await getSessionUser(request.headers);
  if (!user) return error('sign in required', 401);

  const settings = await getProfileSettings(user.id);
  return NextResponse.json({ settings });
}

/**
 * Best-effort GitHub profile derivation: the account row stores the
 * numeric GitHub user id; the public users API maps it to login +
 * profile URL. Failures (offline, rate limit) never block the save.
 */
async function deriveGithubProfile(userId: string): Promise<void> {
  const accountId = await getGithubAccountId(userId);
  if (!accountId) return;
  const existing = await getProfileSettings(userId);
  if (existing.providerUsername) return;

  try {
    const response = await fetch(`https://api.github.com/user/${accountId}`, {
      headers: { accept: 'application/vnd.github+json' },
      signal: AbortSignal.timeout(5000),
    });
    if (!response.ok) return;
    const data = (await response.json()) as { login?: string; html_url?: string };
    if (data.login && data.html_url) {
      await setProviderProfile(userId, data.login, data.html_url);
    }
  } catch {
    /* derived data is optional */
  }
}

export async function PUT(request: Request) {
  const user = await getSessionUser(request.headers);
  if (!user) return error('sign in required', 401);

  let payload: { profileVisible?: unknown; profileUrl?: unknown };
  try {
    payload = await request.json();
  } catch {
    return error('invalid JSON body', 400);
  }

  if (typeof payload.profileVisible !== 'boolean') {
    return error('profileVisible must be a boolean', 400);
  }

  let profileUrl: string | null = null;
  if (payload.profileUrl !== undefined && payload.profileUrl !== null && payload.profileUrl !== '') {
    if (typeof payload.profileUrl !== 'string' || payload.profileUrl.length > PROFILE_URL_MAX) {
      return error(`profileUrl must be a string of at most ${PROFILE_URL_MAX} characters`, 400);
    }
    let parsed: URL;
    try {
      parsed = new URL(payload.profileUrl);
    } catch {
      return error('profileUrl must be a valid URL', 400);
    }
    if (parsed.protocol !== 'https:') return error('profileUrl must use https', 400);
    profileUrl = parsed.toString();
  }

  await upsertProfileSettings(user.id, { profileVisible: payload.profileVisible, profileUrl });
  await deriveGithubProfile(user.id);

  const settings = await getProfileSettings(user.id);
  return NextResponse.json({ settings });
}
