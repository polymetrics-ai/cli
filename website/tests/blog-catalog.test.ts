import { describe, expect, it } from 'vitest';

import { getBlogPost } from '@/lib/blog';

describe('blog catalog', () => {
  it('publishes the human harnesses essay with verified repository evidence', () => {
    const post = getBlogPost('human-harnesses');

    expect(post).toBeDefined();
    expect(post).toMatchObject({
      title: 'Humans Need Harnesses Too',
      publishedAt: '2026-08-04',
      updatedAt: '2026-08-04',
      category: 'Build in public',
    });
    expect(post?.tags).toEqual(
      expect.arrayContaining(['human harnesses', 'GitHub Actions', 'approval gates']),
    );

    const headings = post?.sections.map((section) => section.heading);
    expect(headings).toEqual(
      expect.arrayContaining([
        'The repository became a harness',
        'What the harness still does not do',
      ]),
    );

    const articleText = post?.sections
      .flatMap((section) => [
        ...section.body,
        ...(section.points ?? []),
        section.code ?? '',
      ])
      .join(' ');

    expect(articleText).toContain('29,129');
    expect(articleText).toContain('14,783');
    expect(articleText).toContain('14,346');
    expect(articleText).toContain('2,903');
    expect(articleText).toContain('7,088');
  });
});
