import { ensureMigrated, getPool } from '@/lib/db';

export type ProfileSettings = {
  profileVisible: boolean;
  profileUrl: string | null;
  providerUsername: string | null;
  providerProfileUrl: string | null;
};

const DEFAULTS: ProfileSettings = {
  profileVisible: false,
  profileUrl: null,
  providerUsername: null,
  providerProfileUrl: null,
};

type SettingsRow = {
  profile_visible: boolean;
  profile_url: string | null;
  provider_username: string | null;
  provider_profile_url: string | null;
};

export async function getProfileSettings(userId: string): Promise<ProfileSettings> {
  await ensureMigrated();
  const { rows } = await getPool().query<SettingsRow>(
    'SELECT profile_visible, profile_url, provider_username, provider_profile_url FROM profile_settings WHERE user_id = $1',
    [userId],
  );
  const row = rows[0];
  if (!row) return DEFAULTS;
  return {
    profileVisible: row.profile_visible,
    profileUrl: row.profile_url,
    providerUsername: row.provider_username,
    providerProfileUrl: row.provider_profile_url,
  };
}

export async function upsertProfileSettings(
  userId: string,
  input: { profileVisible: boolean; profileUrl: string | null },
): Promise<void> {
  await ensureMigrated();
  await getPool().query(
    `INSERT INTO profile_settings (user_id, profile_visible, profile_url, updated_at)
     VALUES ($1, $2, $3, NOW())
     ON CONFLICT (user_id) DO UPDATE
       SET profile_visible = EXCLUDED.profile_visible,
           profile_url = EXCLUDED.profile_url,
           updated_at = NOW()`,
    [userId, input.profileVisible, input.profileUrl],
  );
}

export async function setProviderProfile(
  userId: string,
  username: string,
  profileUrl: string,
): Promise<void> {
  await ensureMigrated();
  await getPool().query(
    `INSERT INTO profile_settings (user_id, provider_username, provider_profile_url, updated_at)
     VALUES ($1, $2, $3, NOW())
     ON CONFLICT (user_id) DO UPDATE
       SET provider_username = EXCLUDED.provider_username,
           provider_profile_url = EXCLUDED.provider_profile_url,
           updated_at = NOW()`,
    [userId, username, profileUrl],
  );
}

/** The GitHub account id for a user, if they signed in with GitHub. */
export async function getGithubAccountId(userId: string): Promise<string | undefined> {
  await ensureMigrated();
  const { rows } = await getPool().query<{ accountId: string }>(
    `SELECT "accountId" FROM "account" WHERE "userId" = $1 AND "providerId" = 'github' LIMIT 1`,
    [userId],
  );
  return rows[0]?.accountId;
}
