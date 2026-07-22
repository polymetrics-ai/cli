'use client';

import { useState } from 'react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogTitle,
} from '@/components/ui/dialog';
import { signIn } from '@/lib/auth-client';

const GITHUB_ICON_PATH =
  'M12 .5C5.37.5 0 5.78 0 12.29c0 5.2 3.44 9.6 8.21 11.16.6.11.82-.25.82-.56 0-.28-.01-1.02-.02-2-3.34.71-4.04-1.58-4.04-1.58-.55-1.37-1.34-1.74-1.34-1.74-1.09-.73.08-.72.08-.72 1.2.08 1.84 1.21 1.84 1.21 1.07 1.79 2.81 1.27 3.5.97.11-.76.42-1.27.76-1.56-2.67-.3-5.47-1.31-5.47-5.83 0-1.29.47-2.34 1.24-3.17-.12-.3-.54-1.52.12-3.16 0 0 1.01-.32 3.3 1.21a11.6 11.6 0 0 1 3-.4c1.02 0 2.05.13 3 .4 2.29-1.53 3.3-1.21 3.3-1.21.66 1.64.24 2.86.12 3.16.77.83 1.24 1.88 1.24 3.17 0 4.53-2.81 5.53-5.49 5.82.43.37.81 1.1.81 2.22 0 1.6-.01 2.9-.01 3.29 0 .31.22.68.82.56A12.01 12.01 0 0 0 24 12.29C24 5.78 18.63.5 12 .5Z';

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
  const [pending, setPending] = useState(false);

  async function start() {
    setPending(true);
    try {
      await signIn.social({
        provider: 'github',
        callbackURL: callbackURL ?? window.location.pathname,
      });
    } catch {
      setPending(false);
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
            Use GitHub to comment on passages and keep private bookmarks across the blog.
          </DialogDescription>
        </div>

        <div className="px-6 py-5">
          <button
            type="button"
            disabled={pending}
            onClick={start}
            className="flex w-full items-center gap-3 border border-line-structure bg-surface-1 px-4 py-2.5 text-left transition-colors hover:border-line-cta hover:bg-surface-2 disabled:pointer-events-none disabled:opacity-60"
          >
            <svg
              aria-hidden
              width="16"
              height="16"
              viewBox="0 0 24 24"
              fill="currentColor"
              className="shrink-0 text-text-secondary"
            >
              <path d={GITHUB_ICON_PATH} />
            </svg>
            <span className="font-square text-[11px] font-semibold uppercase tracking-wider text-text-secondary">
              {pending ? 'Redirecting…' : 'Continue with GitHub'}
            </span>
          </button>
        </div>

        <p className="border-t border-line-structure px-6 py-4 font-mono text-[10px] leading-relaxed text-text-disabled">
          We only read your GitHub name, avatar, and email. No posts on your behalf.
        </p>
      </DialogContent>
    </Dialog>
  );
}
