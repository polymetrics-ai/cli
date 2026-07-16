import { readFileSync } from 'node:fs';
import path from 'node:path';
import { createElement } from 'react';
import { renderToStaticMarkup } from 'react-dom/server';
import { describe, expect, it } from 'vitest';

import { PmLogoMark } from '@/components/brand/pm-logo-mark';

const websiteRoot = path.resolve(import.meta.dirname, '..');
const repositoryRoot = path.resolve(websiteRoot, '..');

function readRepositoryFile(relativePath: string) {
  return readFileSync(path.join(repositoryRoot, relativePath), 'utf8');
}

describe('PM brand mark', () => {
  it('keeps P static, blinks M, and renders no cursor glyph', () => {
    const markup = renderToStaticMarkup(
      createElement(PmLogoMark, { title: 'Polymetrics CLI' }),
    );

    expect(markup).toContain('aria-label="Polymetrics CLI"');
    expect(markup).toContain('>P</text>');
    expect(markup).toContain('pm-logo-mark__m');
    expect(markup).toContain('>M</text>');
    expect(markup).not.toContain('>_</text>');
    expect(markup).toContain('@keyframes pm-logo-mark-m');
    expect(markup).toContain('@media (prefers-reduced-motion: reduce)');
  });

  it('uses the shared mark in every persistent website brand surface', () => {
    const consumers = [
      'website/components/home/navbar.tsx',
      'website/components/home/home-sidebar.tsx',
      'website/components/home/site-footer.tsx',
    ];

    for (const consumer of consumers) {
      const source = readRepositoryFile(consumer);
      expect(source, consumer).toContain(
        "import { PmLogoMark } from '@/components/brand/pm-logo-mark';",
      );
      expect(source, consumer).toContain('<PmLogoMark');
      expect(source, consumer).not.toContain('cursor-blink');
    }
  });
});

describe('repository license boundaries', () => {
  it('uses AGPL-3.0-only by default and MIT for connector definitions', () => {
    const rootLicense = readRepositoryFile('LICENSE');
    const definitionsLicense = readRepositoryFile('internal/connectors/defs/LICENSE');
    const licensingMap = readRepositoryFile('LICENSING.md');

    expect(rootLicense.trimStart()).toMatch(
      /^GNU AFFERO GENERAL PUBLIC LICENSE\n\s+Version 3/,
    );
    expect(rootLicense).toContain('13. Remote Network Interaction; Use with the GNU General Public License.');
    expect(definitionsLicense).toMatch(/^MIT License\n\nCopyright \(c\) 2026 Polymetrics AI/);
    expect(licensingMap).toContain('default license is `AGPL-3.0-only`');
    expect(licensingMap).toContain('`internal/connectors/defs/**` | `MIT`');
  });

  it('keeps maintained repository and website copy aligned with the license map', () => {
    const maintainedCopy = [
      'README.md',
      'NOTICE',
      'CONTRIBUTING.md',
      '.github/claude-review-rubric.md',
      'website/app/(home)/page.tsx',
      'website/components/home/faq-accordion.tsx',
      'website/components/home/home-sidebar.tsx',
      'website/components/home/site-footer.tsx',
    ].map(readRepositoryFile).join('\n');

    expect(maintainedCopy).not.toMatch(/Elastic License|Elastic-2\.0|public source/i);
    expect(maintainedCopy).toContain('AGPL-3.0-only');
    expect(maintainedCopy).toContain('MIT');
  });
});
