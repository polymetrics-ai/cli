'use client';

import { useEffect, useState } from 'react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogTitle,
} from '@/components/ui/dialog';

type Settings = {
  profileVisible: boolean;
  profileUrl: string | null;
  providerUsername: string | null;
  providerProfileUrl: string | null;
};

export function ProfileSettingsDialog({
  open,
  onOpenChange,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const [settings, setSettings] = useState<Settings | null>(null);
  const [visible, setVisible] = useState(false);
  const [url, setUrl] = useState('');
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<string | null>(null);

  useEffect(() => {
    if (!open) return;
    setMessage(null);
    fetch('/api/profile')
      .then((response) => (response.ok ? response.json() : null))
      .then((data: { settings: Settings } | null) => {
        if (!data) return;
        setSettings(data.settings);
        setVisible(data.settings.profileVisible);
        setUrl(data.settings.profileUrl ?? '');
      })
      .catch(() => {});
  }, [open]);

  async function save() {
    if (saving) return;
    setSaving(true);
    setMessage(null);
    try {
      const response = await fetch('/api/profile', {
        method: 'PUT',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ profileVisible: visible, profileUrl: url.trim() || null }),
      });
      const data = await response.json();
      if (!response.ok) {
        setMessage(data.error ?? 'saving failed');
        return;
      }
      setSettings(data.settings);
      setMessage('Saved');
      window.setTimeout(() => onOpenChange(false), 600);
    } catch {
      setMessage('saving failed');
    } finally {
      setSaving(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="w-[380px] max-w-[calc(100vw-2rem)] border border-line-structure bg-surface-bg p-0 shadow-[0_18px_60px_rgba(12,31,23,0.16)]">
        <div className="border-b border-line-structure px-6 py-5">
          <p className="font-mono text-[10px] uppercase tracking-widest text-text-disabled">
            Profile
          </p>
          <DialogTitle className="mt-2 font-square text-[20px] font-semibold leading-tight text-text-primary">
            How readers see you
          </DialogTitle>
          <DialogDescription className="mt-2 text-[13px] leading-relaxed text-text-tertiary">
            Your name and avatar always appear on your notes. Details below show
            only if you make your profile visible.
          </DialogDescription>
        </div>

        <div className="flex flex-col gap-4 px-6 py-5">
          <label className="flex cursor-pointer items-start gap-3">
            <input
              type="checkbox"
              checked={visible}
              onChange={(event) => setVisible(event.target.checked)}
              className="mt-0.5 h-4 w-4 shrink-0 cursor-pointer appearance-none border border-line-structure bg-surface-bg outline-none checked:border-line-cta checked:bg-surface-cta-primary focus-visible:outline-1 focus-visible:outline-surface-cta-primary"
            />
            <span>
              <span className="block text-[13px] font-medium text-text-primary">
                Make my profile visible
              </span>
              <span className="mt-0.5 block text-[12px] leading-snug text-text-tertiary">
                Readers hovering your name see member-since, note count, and your links.
              </span>
            </span>
          </label>

          {settings?.providerUsername ? (
            <div className="border border-line-structure bg-surface-1 px-3 py-2">
              <p className="font-mono text-[9px] uppercase tracking-widest text-text-disabled">
                Linked GitHub profile
              </p>
              <p className="mt-1 truncate font-mono text-[12px] text-text-secondary">
                {settings.providerProfileUrl ?? `@${settings.providerUsername}`}
              </p>
            </div>
          ) : null}

          <label className="flex flex-col gap-1.5">
            <span className="font-mono text-[10px] uppercase tracking-widest text-text-disabled">
              Website (optional, https)
            </span>
            <input
              type="url"
              value={url}
              onChange={(event) => setUrl(event.target.value)}
              placeholder="https://your-site.dev"
              className="w-full border border-line-structure bg-surface-bg px-3 py-2 text-[13px] text-text-primary outline-none placeholder:text-text-disabled focus:border-line-cta"
            />
          </label>
        </div>

        <div className="flex items-center justify-between border-t border-line-structure px-6 py-4">
          <span className="font-mono text-[10px] uppercase tracking-widest text-text-disabled">
            {message ?? ''}
          </span>
          <button
            type="button"
            disabled={saving || settings === null}
            onClick={() => void save()}
            className="border border-line-cta bg-line-cta px-4 py-2 font-square text-[11px] font-semibold uppercase tracking-wider text-surface-bg transition-opacity hover:opacity-90 disabled:pointer-events-none disabled:opacity-50"
          >
            {saving ? 'Saving…' : 'Save'}
          </button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
