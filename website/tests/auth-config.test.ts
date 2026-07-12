import { describe, expect, it } from 'vitest';

import {
  configuredTrustedProviders,
  shouldRequireLocalEmailVerified,
} from '@/lib/auth-config';

describe('auth account linking config', () => {
  it('trusts only configured social providers', () => {
    expect(
      configuredTrustedProviders({
        github: { configured: true },
        google: undefined,
        linkedin: { configured: true },
      }),
    ).toEqual(['github', 'linkedin']);
  });

  it('keeps local email verification strict in production only', () => {
    expect(shouldRequireLocalEmailVerified('production')).toBe(true);
    expect(shouldRequireLocalEmailVerified('development')).toBe(false);
    expect(shouldRequireLocalEmailVerified('test')).toBe(false);
  });
});
