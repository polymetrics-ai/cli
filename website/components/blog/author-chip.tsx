'use client';

import { useState } from 'react';
import { ExternalLink } from 'lucide-react';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import type { AuthorDto } from '@/components/blog/annotations-provider';

function Avatar({ author, size }: { author: AuthorDto; size: number }) {
  if (author.image) {
    return (
      // eslint-disable-next-line @next/next/no-img-element -- OAuth avatar host varies
      <img
        src={author.image}
        alt=""
        width={size}
        height={size}
        style={{ width: size, height: size }}
        className="border border-line-structure object-cover"
      />
    );
  }
  return (
    <span
      style={{ width: size, height: size }}
      className="flex items-center justify-center border border-line-cta bg-surface-cta-primary font-mono text-[8px] font-bold text-line-cta"
    >
      {(author.name[0] ?? '?').toUpperCase()}
    </span>
  );
}

const GITHUB_PATH =
  'M12 .5C5.37.5 0 5.78 0 12.29c0 5.2 3.44 9.6 8.21 11.16.6.11.82-.25.82-.56 0-.28-.01-1.02-.02-2-3.34.71-4.04-1.58-4.04-1.58-.55-1.37-1.34-1.74-1.34-1.74-1.09-.73.08-.72.08-.72 1.2.08 1.84 1.21 1.84 1.21 1.07 1.79 2.81 1.27 3.5.97.11-.76.42-1.27.76-1.56-2.67-.3-5.47-1.31-5.47-5.83 0-1.29.47-2.34 1.24-3.17-.12-.3-.54-1.52.12-3.16 0 0 1.01-.32 3.3 1.21a11.6 11.6 0 0 1 3-.4c1.02 0 2.05.13 3 .4 2.29-1.53 3.3-1.21 3.3-1.21.66 1.64.24 2.86.12 3.16.77.83 1.24 1.88 1.24 3.17 0 4.53-2.81 5.53-5.49 5.82.43.37.81 1.1.81 2.22 0 1.6-.01 2.9-.01 3.29 0 .31.22.68.82.56A12.01 12.01 0 0 0 24 12.29C24 5.78 18.63.5 12 .5Z';

/**
 * A commenter's identity, expandable into a profile preview: real
 * details for authors who opted in, a dynamic placeholder otherwise.
 * Opens on hover/focus, pins on click (Escape/blur closes — radix).
 */
export function AuthorChip({ author }: { author: AuthorDto }) {
  const [open, setOpen] = useState(false);
  const profile = author.profile ?? null;

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button
          type="button"
          aria-label={`View ${author.name}'s profile`}
          onMouseEnter={() => setOpen(true)}
          onMouseLeave={() => setOpen(false)}
          className="flex min-w-0 items-center gap-1.5 outline-none focus-visible:outline-1 focus-visible:outline-surface-cta-primary"
        >
          <Avatar author={author} size={14} />
          <span className="truncate text-[11px] font-medium text-text-secondary underline decoration-line-structure decoration-dotted underline-offset-2">
            {author.name}
          </span>
        </button>
      </PopoverTrigger>
      <PopoverContent
        side="top"
        align="start"
        sideOffset={6}
        onOpenAutoFocus={(event) => event.preventDefault()}
        onMouseEnter={() => setOpen(true)}
        onMouseLeave={() => setOpen(false)}
        className="corner-box-corners w-[240px] border border-line-structure bg-surface-bg p-0"
        data-author-preview={profile ? 'visible' : 'private'}
      >
        {profile ? (
          <div className="p-3">
            <div className="flex items-center gap-2">
              <Avatar author={author} size={24} />
              <span className="truncate text-[13px] font-medium text-text-primary">
                {author.name}
              </span>
            </div>
            <div className="mt-2.5 flex flex-wrap gap-x-3 gap-y-1 font-mono text-[9px] uppercase tracking-widest text-text-disabled">
              <span>Member since {profile.memberSince.slice(0, 4)}</span>
              <span>
                {profile.noteCount} note{profile.noteCount === 1 ? '' : 's'}
              </span>
            </div>
            {profile.providerProfileUrl || profile.profileUrl ? (
              <div className="mt-2.5 flex flex-col gap-1.5 border-t border-line-structure pt-2.5">
                {profile.providerProfileUrl ? (
                  <a
                    href={profile.providerProfileUrl}
                    target="_blank"
                    rel="noreferrer"
                    className="flex items-center gap-1.5 font-mono text-[10px] text-text-secondary transition-colors hover:text-text-primary"
                  >
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor" aria-hidden>
                      <path d={GITHUB_PATH} />
                    </svg>
                    {profile.providerUsername ?? 'GitHub'}
                  </a>
                ) : null}
                {profile.profileUrl ? (
                  <a
                    href={profile.profileUrl}
                    target="_blank"
                    rel="noreferrer"
                    className="flex items-center gap-1.5 truncate font-mono text-[10px] text-text-secondary transition-colors hover:text-text-primary"
                  >
                    <ExternalLink className="h-3 w-3 shrink-0" aria-hidden="true" />
                    <span className="truncate">{profile.profileUrl.replace(/^https:\/\//, '')}</span>
                  </a>
                ) : null}
              </div>
            ) : null}
          </div>
        ) : (
          <div className="with-stripes p-2">
            <div className="bg-surface-bg p-2.5">
              <div className="flex items-center gap-2">
                <Avatar author={author} size={24} />
                <span className="truncate text-[13px] font-medium text-text-primary">
                  {author.name}
                </span>
              </div>
              <p className="mt-2 font-mono text-[9px] uppercase tracking-widest text-text-disabled">
                Profile private
              </p>
              <p className="mt-1 text-[11px] leading-snug text-text-tertiary">
                This reader keeps their profile private.
              </p>
            </div>
          </div>
        )}
      </PopoverContent>
    </Popover>
  );
}
