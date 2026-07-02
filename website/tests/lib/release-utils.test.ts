import { describe, expect, it } from 'vitest';

import {
  formatBytes,
  releaseExcerpt,
  selectPreferredAsset,
  type ReleaseAsset,
} from '../../lib/release-utils';

const assets: ReleaseAsset[] = [
  {
    id: 1,
    name: 'pm-linux-amd64.tar.gz',
    url: 'https://example.com/linux',
    downloadCount: 10,
    size: 18_000_000,
    contentType: 'application/gzip',
  },
  {
    id: 2,
    name: 'pm-darwin-arm64.tar.gz',
    url: 'https://example.com/darwin-arm64',
    downloadCount: 7,
    size: 17_500_000,
    contentType: 'application/gzip',
  },
  {
    id: 3,
    name: 'pm-windows-amd64.zip',
    url: 'https://example.com/windows',
    downloadCount: 3,
    size: 19_000_000,
    contentType: 'application/zip',
  },
];

describe('release utilities', () => {
  it('selects the closest asset for the current platform', () => {
    expect(selectPreferredAsset(assets, 'MacIntel arm64')?.name).toBe('pm-darwin-arm64.tar.gz');
    expect(selectPreferredAsset(assets, 'Windows x86_64')?.name).toBe('pm-windows-amd64.zip');
    expect(selectPreferredAsset(assets, 'Linux x86_64')?.name).toBe('pm-linux-amd64.tar.gz');
  });

  it('returns null when no binary assets exist', () => {
    expect(selectPreferredAsset([], 'MacIntel')).toBeNull();
  });

  it('formats bytes and release markdown excerpts for compact UI rows', () => {
    expect(formatBytes(18_000_000)).toBe('17 MB');
    const excerpt = releaseExcerpt('## Added\n\n- Binary release assets\n- Release notes', 40);
    expect(excerpt).toContain('Binary release assets');
    expect(excerpt.length).toBeLessThanOrEqual(40);
    expect(excerpt.endsWith('...')).toBe(true);
  });
});
