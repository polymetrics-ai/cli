'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';

export type SidebarShortcut = {
  key: string;
  href: string;
};

function shouldIgnoreShortcutTarget(target: EventTarget | null) {
  if (!(target instanceof HTMLElement)) return false;
  const tag = target.tagName;
  return tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT' || target.isContentEditable;
}

export function SidebarShortcuts({ shortcuts }: { shortcuts: SidebarShortcut[] }) {
  const router = useRouter();

  useEffect(() => {
    const map = new Map(shortcuts.map((shortcut) => [
      shortcut.key.toLowerCase(),
      shortcut.href,
    ]));

    function onKeyDown(event: KeyboardEvent) {
      if (event.repeat || event.metaKey || event.ctrlKey || event.altKey) return;
      if (shouldIgnoreShortcutTarget(event.target)) return;

      const href = map.get(event.key.toLowerCase());
      if (!href) return;

      event.preventDefault();
      if (href.startsWith('http')) {
        window.open(href, '_blank', 'noopener,noreferrer');
      } else {
        router.push(href);
      }
    }

    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
  }, [router, shortcuts]);

  return null;
}
