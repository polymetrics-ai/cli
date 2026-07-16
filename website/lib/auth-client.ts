'use client';

import { useEffect, useState } from 'react';
import { createAuthClient } from 'better-auth/react';

export const authClient = createAuthClient();

export const { signIn, signOut, useSession } = authClient;

/**
 * Better Auth can resolve its client store before React hydrates. Hold the
 * server and first browser render in the same pending state, then expose the
 * live session after mount so auth-dependent markup never mismatches.
 */
export function useHydratedSession() {
  const session = useSession();
  const [hydrated, setHydrated] = useState(false);

  useEffect(() => setHydrated(true), []);

  return {
    ...session,
    data: hydrated ? session.data : undefined,
    isPending: !hydrated || session.isPending,
  };
}
