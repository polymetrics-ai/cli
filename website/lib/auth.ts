import { betterAuth } from 'better-auth';
import { getPool } from '@/lib/db';

const baseURL = process.env.BETTER_AUTH_URL ?? 'http://localhost:3000';

function provider(idVar: string, secretVar: string) {
  const clientId = process.env[idVar];
  const clientSecret = process.env[secretVar];
  if (!clientId || !clientSecret) return undefined;
  return { clientId, clientSecret };
}

// Providers are included only when their credentials are present so the
// site builds and boots (docs, blog, existing APIs) without any OAuth
// secrets configured — sign-in simply offers what is wired up.
const github = provider('GITHUB_CLIENT_ID', 'GITHUB_CLIENT_SECRET');
const google = provider('GOOGLE_CLIENT_ID', 'GOOGLE_CLIENT_SECRET');
const linkedin = provider('LINKEDIN_CLIENT_ID', 'LINKEDIN_CLIENT_SECRET');

export const auth = betterAuth({
  database: getPool(),
  baseURL,
  trustedOrigins: [baseURL],
  socialProviders: {
    ...(github ? { github } : {}),
    ...(google ? { google } : {}),
    ...(linkedin ? { linkedin } : {}),
  },
  session: {
    cookieCache: {
      enabled: true,
      maxAge: 300,
    },
  },
  rateLimit: {
    enabled: true,
  },
});
