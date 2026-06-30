'use client';

import { useState } from 'react';
import { AlertTriangle, Check, Copy } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';

interface CopyMarkdownProps {
  slug: string | string[] | undefined;
  label?: string;
  className?: string;
}

export function CopyMarkdown({ slug, label = 'Copy as Markdown', className }: CopyMarkdownProps) {
  const [state, setState] = useState<'idle' | 'copied' | 'error'>('idle');

  const rawPath = slug
    ? `/api/raw/${Array.isArray(slug) ? slug.join('/') : slug}`
    : '/api/raw/index';

  async function handleClick() {
    try {
      const res = await fetch(rawPath);
      if (!res.ok) throw new Error('Failed to fetch');
      const text = await res.text();
      await navigator.clipboard.writeText(text);
      setState('copied');
      setTimeout(() => setState('idle'), 2000);
    } catch {
      setState('error');
      setTimeout(() => setState('idle'), 2000);
    }
  }

  const Icon = state === 'copied' ? Check : state === 'error' ? AlertTriangle : Copy;
  const text = state === 'idle' ? label : state === 'copied' ? 'Copied' : 'Error';

  return (
    <Button
      onClick={handleClick}
      aria-label="Copy page as Markdown"
      variant="quiet"
      className={cn(
        'shrink-0 border-line-structure bg-surface-bg text-text-tertiary hover:border-line-cta hover:bg-surface-2 hover:text-text-primary',
        state === 'copied' && 'border-line-cta text-line-cta',
        state === 'error' && 'border-red-300 text-red-700',
        className,
      )}
    >
      <Icon className="h-3.5 w-3.5" aria-hidden="true" />
      <span aria-live="polite">{text}</span>
    </Button>
  );
}
