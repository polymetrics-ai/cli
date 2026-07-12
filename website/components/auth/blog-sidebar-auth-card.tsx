'use client';

import { useState } from 'react';
import Link from 'next/link';
import { Bookmark, IdCard, LogOut, MessageSquare } from 'lucide-react';
import { CornerBox } from '@/components/ui/corner-box';
import { ProfileSettingsDialog } from '@/components/auth/profile-settings-dialog';
import { SignInDialog } from '@/components/auth/sign-in-dialog';
import { signOut, useSession } from '@/lib/auth-client';

function AccountAvatar({ name, image }: { name: string; image: string | null | undefined }) {
  if (image) {
    return (
      // eslint-disable-next-line @next/next/no-img-element -- external OAuth avatar hosts are unknown ahead of time
      <img
        src={image}
        alt=""
        width={32}
        height={32}
        className="h-8 w-8 shrink-0 border border-line-structure object-cover"
      />
    );
  }

  const initial = (name.trim()[0] ?? '?').toUpperCase();
  return (
    <span className="flex h-8 w-8 shrink-0 items-center justify-center border border-line-cta bg-surface-cta-primary font-mono text-[12px] font-bold text-line-cta">
      {initial}
    </span>
  );
}

function ActionButton({
  children,
  onClick,
}: {
  children: React.ReactNode;
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="flex w-full items-center gap-2 border border-line-structure bg-surface-1 px-2.5 py-2 text-left text-[12px] text-text-secondary transition-colors hover:border-line-cta hover:bg-surface-2 hover:text-text-primary focus-visible:outline-1 focus-visible:outline-surface-cta-primary"
    >
      {children}
    </button>
  );
}

export function BlogSidebarAuthCard() {
  const { data: session, isPending } = useSession();
  const [signInOpen, setSignInOpen] = useState(false);
  const [profileOpen, setProfileOpen] = useState(false);

  return (
    <div
      data-blog-auth-card
      data-session-ready={isPending ? 'false' : 'true'}
      className="w-full min-w-0 border-t border-line-structure p-2"
    >
      <CornerBox className="min-h-[166px] min-w-0 overflow-hidden p-3">
        {isPending ? (
          <div className="flex h-full min-h-[140px] flex-col justify-between" aria-label="Loading account">
            <div>
              <span className="block h-3 w-28 bg-surface-1" />
              <span className="mt-4 block h-4 w-32 bg-surface-1" />
              <span className="mt-2 block h-3 w-full bg-surface-1" />
              <span className="mt-1.5 block h-3 w-3/4 bg-surface-1" />
            </div>
            <span className="block h-8 w-full border border-line-structure bg-surface-1" />
          </div>
        ) : !session ? (
          <>
            <div className="flex items-center justify-between gap-2">
              <p className="font-mono text-[10px] uppercase tracking-widest text-text-disabled">
                Join blog discussion
              </p>
              <MessageSquare className="h-3.5 w-3.5 text-line-cta" aria-hidden="true" />
            </div>
            <p className="mt-3 text-[12px] leading-relaxed text-text-tertiary">
              Sign in to comment on passages and keep private bookmarks.
            </p>
            <button
              type="button"
              onClick={() => setSignInOpen(true)}
              className="mt-4 flex w-full items-center justify-center border border-line-cta bg-line-cta px-3 py-2 font-square text-[11px] font-semibold uppercase tracking-wider text-surface-bg transition-opacity hover:opacity-90 focus-visible:outline-1 focus-visible:outline-surface-cta-primary"
            >
              Sign in
            </button>
            <SignInDialog open={signInOpen} onOpenChange={setSignInOpen} />
          </>
        ) : (
          <>
            <div className="flex items-center justify-between gap-2">
              <p className="font-mono text-[10px] uppercase tracking-widest text-text-disabled">
                Reading as
              </p>
              <MessageSquare className="h-3.5 w-3.5 text-line-cta" aria-hidden="true" />
            </div>
            <div className="mt-3 flex min-w-0 items-center gap-2.5">
              <AccountAvatar name={session.user.name} image={session.user.image} />
              <div className="min-w-0">
                <p className="truncate text-[13px] font-medium text-text-primary">
                  {session.user.name}
                </p>
                <p className="mt-0.5 truncate font-mono text-[10px] text-text-disabled">
                  {session.user.email}
                </p>
              </div>
            </div>
            <div className="mt-4 grid gap-1.5">
              <Link
                href="/bookmarks"
                className="flex items-center gap-2 border border-line-structure bg-surface-1 px-2.5 py-2 text-[12px] text-text-secondary transition-colors hover:border-line-cta hover:bg-surface-2 hover:text-text-primary focus-visible:outline-1 focus-visible:outline-surface-cta-primary"
              >
                <Bookmark className="h-3.5 w-3.5" aria-hidden="true" />
                Bookmarks
              </Link>
              <ActionButton onClick={() => setProfileOpen(true)}>
                <IdCard className="h-3.5 w-3.5" aria-hidden="true" />
                Profile
              </ActionButton>
              <ActionButton onClick={() => void signOut()}>
                <LogOut className="h-3.5 w-3.5" aria-hidden="true" />
                Sign out
              </ActionButton>
            </div>
            <ProfileSettingsDialog open={profileOpen} onOpenChange={setProfileOpen} />
          </>
        )}
      </CornerBox>
    </div>
  );
}
