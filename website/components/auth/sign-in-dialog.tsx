'use client';

import { useState } from 'react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogTitle,
} from '@/components/ui/dialog';
import { signIn } from '@/lib/auth-client';

type Provider = 'github' | 'google' | 'linkedin';

const PROVIDERS: { id: Provider; label: string; path: string }[] = [
  {
    id: 'github',
    label: 'Continue with GitHub',
    path: 'M12 .5C5.37.5 0 5.78 0 12.29c0 5.2 3.44 9.6 8.21 11.16.6.11.82-.25.82-.56 0-.28-.01-1.02-.02-2-3.34.71-4.04-1.58-4.04-1.58-.55-1.37-1.34-1.74-1.34-1.74-1.09-.73.08-.72.08-.72 1.2.08 1.84 1.21 1.84 1.21 1.07 1.79 2.81 1.27 3.5.97.11-.76.42-1.27.76-1.56-2.67-.3-5.47-1.31-5.47-5.83 0-1.29.47-2.34 1.24-3.17-.12-.3-.54-1.52.12-3.16 0 0 1.01-.32 3.3 1.21a11.6 11.6 0 0 1 3-.4c1.02 0 2.05.13 3 .4 2.29-1.53 3.3-1.21 3.3-1.21.66 1.64.24 2.86.12 3.16.77.83 1.24 1.88 1.24 3.17 0 4.53-2.81 5.53-5.49 5.82.43.37.81 1.1.81 2.22 0 1.6-.01 2.9-.01 3.29 0 .31.22.68.82.56A12.01 12.01 0 0 0 24 12.29C24 5.78 18.63.5 12 .5Z',
  },
  {
    id: 'google',
    label: 'Continue with Google',
    path: 'M23.49 12.27c0-.79-.07-1.54-.2-2.27H12v4.51h6.47a5.53 5.53 0 0 1-2.4 3.58v3h3.86c2.26-2.09 3.56-5.17 3.56-8.82ZM12 24c3.24 0 5.95-1.08 7.93-2.91l-3.86-3c-1.08.72-2.45 1.16-4.07 1.16-3.13 0-5.78-2.11-6.73-4.96H1.29v3.09A11.99 11.99 0 0 0 12 24ZM5.27 14.29a7.19 7.19 0 0 1 0-4.58V6.62H1.29a12.04 12.04 0 0 0 0 10.76l3.98-3.09ZM12 4.75c1.77 0 3.35.61 4.6 1.8l3.42-3.42C17.95 1.19 15.24 0 12 0 7.31 0 3.26 2.69 1.29 6.62l3.98 3.09C6.22 6.86 8.87 4.75 12 4.75Z',
  },
  {
    id: 'linkedin',
    label: 'Continue with LinkedIn',
    path: 'M20.45 20.45h-3.56v-5.57c0-1.33-.03-3.04-1.85-3.04-1.85 0-2.13 1.45-2.13 2.94v5.67H9.35V9h3.42v1.56h.05c.48-.9 1.64-1.85 3.37-1.85 3.6 0 4.27 2.37 4.27 5.46v6.28ZM5.34 7.43a2.06 2.06 0 1 1 0-4.13 2.06 2.06 0 0 1 0 4.13ZM7.12 20.45H3.55V9h3.57v11.45ZM22.22 0H1.77C.79 0 0 .77 0 1.73v20.54C0 23.23.79 24 1.77 24h20.45c.98 0 1.78-.77 1.78-1.73V1.73C24 .77 23.2 0 22.22 0Z',
  },
];

export function SignInDialog({
  open,
  onOpenChange,
  callbackURL,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  /** Where OAuth returns the reader; defaults to the current page. */
  callbackURL?: string;
}) {
  const [pending, setPending] = useState<Provider | null>(null);

  async function start(provider: Provider) {
    setPending(provider);
    try {
      await signIn.social({
        provider,
        callbackURL: callbackURL ?? window.location.pathname,
      });
    } catch {
      setPending(null);
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="w-[360px] max-w-[calc(100vw-2rem)] border border-line-structure bg-surface-bg p-0 shadow-[0_18px_60px_rgba(12,31,23,0.16)]">
        <div className="border-b border-line-structure px-6 py-5">
          <p className="font-mono text-[10px] uppercase tracking-widest text-text-disabled">
            Sign in
          </p>
          <DialogTitle className="mt-2 font-square text-[20px] font-semibold leading-tight text-text-primary">
            Join the discussion
          </DialogTitle>
          <DialogDescription className="mt-2 text-[13px] leading-relaxed text-text-tertiary">
            Comment on passages and keep private bookmarks across the blog.
          </DialogDescription>
        </div>

        <div className="flex flex-col gap-2 px-6 py-5">
          {PROVIDERS.map(({ id, label, path }) => (
            <button
              key={id}
              type="button"
              disabled={pending !== null}
              onClick={() => start(id)}
              className="flex w-full items-center gap-3 border border-line-structure bg-surface-1 px-4 py-2.5 text-left transition-colors hover:border-line-cta hover:bg-surface-2 disabled:pointer-events-none disabled:opacity-60"
            >
              <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor" aria-hidden className="shrink-0 text-text-secondary">
                <path d={path} />
              </svg>
              <span className="font-square text-[11px] font-semibold uppercase tracking-wider text-text-secondary">
                {pending === id ? 'Redirecting…' : label}
              </span>
            </button>
          ))}
        </div>

        <p className="border-t border-line-structure px-6 py-4 font-mono text-[10px] leading-relaxed text-text-disabled">
          We only read your name, avatar, and email. No posts on your behalf.
        </p>
      </DialogContent>
    </Dialog>
  );
}
