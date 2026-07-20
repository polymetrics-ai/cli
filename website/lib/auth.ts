import { betterAuth } from 'better-auth';
import {
  configuredTrustedProviders,
  shouldRequireLocalEmailVerified,
  type SocialProviderId,
} from '@/lib/auth-config';
import { getPool } from '@/lib/db';

const baseURL = process.env.BETTER_AUTH_URL ?? 'http://localhost:3000';

type OAuthProviderConfig = {
  clientId: string;
  clientSecret: string;
};

function provider(idVar: string, secretVar: string) {
  const clientId = process.env[idVar];
  const clientSecret = process.env[secretVar];
  if (!clientId || !clientSecret) return undefined;
  return { clientId, clientSecret };
}

// GitHub is included only when both credentials are present so the public
// site still builds and boots without OAuth secrets in build environments.
const github = provider('GITHUB_CLIENT_ID', 'GITHUB_CLIENT_SECRET');

const socialProviders = {
  ...(github ? { github } : {}),
} satisfies Partial<Record<SocialProviderId, OAuthProviderConfig>>;

export const auth = betterAuth({
  database: getPool(),
  baseURL,
  trustedOrigins: [baseURL],
  socialProviders,
  account: {
    accountLinking: {
      enabled: true,
      trustedProviders: configuredTrustedProviders(socialProviders),
      requireLocalEmailVerified: shouldRequireLocalEmailVerified(),
      updateUserInfoOnLink: true,
    },
  },
  session: {
    cookieCache: {
      enabled: true,
      maxAge: 300,
    },
  },
  // Playwright-only escape hatch: e2e specs sign in with credentials
  // instead of real OAuth. Double-gated so it can never ship enabled.
  emailAndPassword: {
    enabled: process.env.E2E_TEST_AUTH === '1' && process.env.NODE_ENV !== 'production',
  },
  rateLimit: {
    enabled: true,
  },
});
