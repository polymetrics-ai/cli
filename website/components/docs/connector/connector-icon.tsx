import * as React from 'react';
import { cn } from '@/lib/utils';

type ConnectorIconLike = {
  publicPath: string;
  id?: string;
  path?: string;
  source?: string;
  reviewStatus?: string;
} | null | undefined;

interface ConnectorIconProps {
  icon: ConnectorIconLike;
  name: string;
  className?: string;
  imageClassName?: string;
}

function initialsForName(name: string): string {
  const words = name
    .replace(/[^a-z0-9]+/gi, ' ')
    .trim()
    .split(/\s+/)
    .filter(Boolean);

  if (words.length === 0) return 'PM';
  if (words.length === 1) return words[0].slice(0, 2).toUpperCase();
  return `${words[0][0]}${words[1][0]}`.toUpperCase();
}

function isGenericIcon(icon: ConnectorIconLike): boolean {
  if (!icon) return true;
  return (
    icon.id === 'pm-sample' ||
    icon.path === 'icons/pm-sample.svg' ||
    icon.publicPath.endsWith('/pm-sample.svg')
  );
}

function fallbackTone(name: string): string {
  const tones = [
    '[background:#e4efe8] text-line-cta',
    '[background:#e9ecef] text-[#24415f]',
    '[background:#eee9df] text-[#5a3716]',
    '[background:#e7e4ef] text-[#3f3264]',
    '[background:#edf0df] text-[#425116]',
  ];
  let hash = 0;
  for (const ch of name) hash = (hash * 31 + ch.charCodeAt(0)) >>> 0;
  return tones[hash % tones.length];
}

export function ConnectorIcon({
  icon,
  name,
  className,
  imageClassName,
}: ConnectorIconProps) {
  const showImage = !!icon?.publicPath && !isGenericIcon(icon);

  return (
    <span
      aria-hidden="true"
      className={cn(
        'flex h-8 w-8 shrink-0 items-center justify-center overflow-hidden border border-line-structure text-[11px] font-semibold leading-none',
        showImage ? 'bg-surface-1 text-line-cta' : fallbackTone(name),
        className,
      )}
    >
      {showImage ? (
        <img
          src={icon.publicPath}
          alt=""
          width={20}
          height={20}
          loading="lazy"
          decoding="async"
          className={cn('h-5 w-5 object-contain', imageClassName)}
        />
      ) : (
        <span className="font-mono">{initialsForName(name)}</span>
      )}
    </span>
  );
}
