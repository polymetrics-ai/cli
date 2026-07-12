'use client';

import { useState } from 'react';
import Link from 'next/link';
import { Bookmark, IdCard, LogOut } from 'lucide-react';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { SignInDialog } from '@/components/auth/sign-in-dialog';
import { ProfileSettingsDialog } from '@/components/auth/profile-settings-dialog';
import { signOut, useSession } from '@/lib/auth-client';

function Avatar({ name, image }: { name: string; image: string | null | undefined }) {
  if (image) {
    return (
      // eslint-disable-next-line @next/next/no-img-element -- external OAuth avatar hosts are unknown ahead of time
      <img
        src={image}
        alt=""
        width={24}
        height={24}
        className="h-6 w-6 border border-line-structure object-cover"
      />
    );
  }
  const initial = (name.trim()[0] ?? '?').toUpperCase();
  return (
    <span className="flex h-6 w-6 items-center justify-center border border-line-cta bg-surface-cta-primary font-mono text-[11px] font-bold text-line-cta">
      {initial}
    </span>
  );
}

export function UserMenu() {
  const { data: session, isPending } = useSession();
  const [signInOpen, setSignInOpen] = useState(false);
  const [profileOpen, setProfileOpen] = useState(false);

  // Reserve the slot while the session loads so the navbar never shifts.
  if (isPending) {
    return <span aria-hidden className="mx-1 inline-block h-6 w-6" />;
  }

  if (!session) {
    return (
      <>
        <button
          type="button"
          onClick={() => setSignInOpen(true)}
          className="mx-1 whitespace-nowrap border border-line-structure bg-surface-bg px-2.5 py-1 font-square text-[11px] font-semibold uppercase tracking-wider text-text-secondary transition-colors hover:border-line-cta hover:text-text-primary"
        >
          Sign in
        </button>
        <SignInDialog open={signInOpen} onOpenChange={setSignInOpen} />
      </>
    );
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <button
          type="button"
          aria-label="Account menu"
          className="mx-1 flex items-center outline-none transition-opacity hover:opacity-80 focus-visible:outline-2 focus-visible:outline-[--surface-cta-primary]"
        >
          <Avatar name={session.user.name} image={session.user.image} />
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent
        align="end"
        sideOffset={10}
        className="w-56 border border-line-structure bg-surface-1 p-1.5 text-text-primary shadow-[0_18px_60px_rgba(12,31,23,0.16)]"
      >
        <DropdownMenuLabel className="px-2.5 py-2">
          <span className="block truncate text-[13px] font-medium text-text-primary">
            {session.user.name}
          </span>
          <span className="mt-0.5 block truncate font-mono text-[10px] text-text-disabled">
            {session.user.email}
          </span>
        </DropdownMenuLabel>
        <DropdownMenuSeparator className="bg-line-structure" />
        <DropdownMenuItem asChild className="cursor-pointer px-2.5 py-2 text-[13px] text-text-secondary focus:bg-surface-bg focus:text-text-primary">
          <Link href="/bookmarks">
            <Bookmark className="h-3.5 w-3.5" aria-hidden="true" />
            Bookmarks
          </Link>
        </DropdownMenuItem>
        <DropdownMenuItem
          className="cursor-pointer px-2.5 py-2 text-[13px] text-text-secondary focus:bg-surface-bg focus:text-text-primary"
          onSelect={() => setProfileOpen(true)}
        >
          <IdCard className="h-3.5 w-3.5" aria-hidden="true" />
          Profile
        </DropdownMenuItem>
        <DropdownMenuItem
          className="cursor-pointer px-2.5 py-2 text-[13px] text-text-secondary focus:bg-surface-bg focus:text-text-primary"
          onSelect={() => void signOut()}
        >
          <LogOut className="h-3.5 w-3.5" aria-hidden="true" />
          Sign out
        </DropdownMenuItem>
      </DropdownMenuContent>
      <ProfileSettingsDialog open={profileOpen} onOpenChange={setProfileOpen} />
    </DropdownMenu>
  );
}
