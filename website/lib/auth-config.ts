export const SOCIAL_PROVIDER_IDS = ['github'] as const;

export type SocialProviderId = (typeof SOCIAL_PROVIDER_IDS)[number];

export function configuredTrustedProviders(
  providers: Partial<Record<SocialProviderId, unknown>>,
): SocialProviderId[] {
  return SOCIAL_PROVIDER_IDS.filter((provider) => Boolean(providers[provider]));
}

export function shouldRequireLocalEmailVerified(nodeEnv = process.env.NODE_ENV): boolean {
  return nodeEnv === 'production';
}
