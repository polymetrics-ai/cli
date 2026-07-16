import { expect, test } from '@playwright/test';
import type { Page } from '@playwright/test';

import { BLOG_POSTS } from '@/lib/blog';

function expectedGitHubDiscussionUrl(title: string): string {
  return `https://github.com/polymetrics-ai/cli/discussions?discussions_q=${encodeURIComponent(`"${title}"`)}`;
}

test.describe('blog UI smoke', () => {
  async function awaitNavbarHydrated(page: Page): Promise<void> {
    await expect(page.locator('header[data-navbar-hydrated="true"]')).toBeVisible();
  }

  async function expectDesktopRails(
    page: Page,
    path: string,
    leftSelector: string,
    rightSelector: string,
  ): Promise<void> {
    await page.goto(path);
    await expect(page.locator(leftSelector)).toBeVisible();
    await expect(page.locator(`${rightSelector}[data-site-toc]`)).toBeVisible();
  }

  test('lists every blog post on the index page', async ({ page }) => {
    await page.goto('/blog');

    await expect(
      page.getByRole('heading', { level: 1, name: 'Notes for the data loop.' }),
    ).toBeVisible();
    for (const post of BLOG_POSTS) {
      await expect(page.getByRole('heading', { name: post.title })).toBeVisible();
    }
  });

  test('renders a blog post with its sections', async ({ page }) => {
    const post = BLOG_POSTS[0];
    await page.goto(`/blog/${post.slug}`);

    await expect(page.getByRole('heading', { level: 1, name: post.title })).toBeVisible();
    for (const section of post.sections) {
      await expect(page.getByRole('heading', { name: section.heading })).toBeVisible();
    }
  });

  test('opens inline GitHub references without leaving the article', async ({ page }) => {
    await page.route('https://api.github.com/**', (route) => route.abort());
    await page.goto('/blog/human-harnesses');
    const articleUrl = page.url();

    await expect(page.locator('[data-github-evidence-trail]')).toHaveCount(0);
    const claimLink = page.getByRole('link', { name: 'Preview PR #27 evidence' }).first();
    const citationLink = page.getByRole('link', { name: 'Open citation 1: PR #27' }).first();
    await expect(claimLink).toHaveText('PR #27');
    await expect(citationLink).toHaveText('[1]');

    await claimLink.click();
    const preview = page.getByRole('dialog', { name: 'GitHub evidence: PR #27' });
    await expect(preview).toBeVisible();
    await expect(preview.getByText('Verified snapshot')).toBeVisible();
    await expect(preview.getByText('Closed, not merged')).toBeVisible();
    await expect(preview.getByText('1,961,878', { exact: true })).toBeVisible();

    const sourceLink = preview.getByRole('link', { name: 'Open PR #27 on GitHub' });
    await expect(sourceLink).toHaveAttribute(
      'href',
      'https://github.com/polymetrics-ai/cli/pull/27',
    );
    await expect(sourceLink).toHaveAttribute('target', '_blank');
    expect(page.url()).toBe(articleUrl);

    await page.keyboard.press('Escape');
    await expect(preview).toBeHidden();
    await expect(claimLink).toBeFocused();

    await citationLink.click();
    await expect(preview).toBeVisible();
    await page.keyboard.press('Escape');
    await expect(citationLink).toBeFocused();

    const starLink = page.getByRole('link', { name: 'Star Polymetrics on GitHub' });
    await expect(starLink).toHaveAttribute('href', 'https://github.com/polymetrics-ai/cli');
    await expect(starLink).toHaveAttribute('target', '_blank');

    await page.setViewportSize({ width: 390, height: 844 });
    await citationLink.click();
    await expect(preview).toBeVisible();
    const previewBox = await preview.boundingBox();
    expect(previewBox).not.toBeNull();
    expect(previewBox!.x).toBeGreaterThanOrEqual(0);
    expect(previewBox!.x + previewBox!.width).toBeLessThanOrEqual(390);
    expect(Math.abs(previewBox!.x + previewBox!.width / 2 - 195)).toBeLessThanOrEqual(2);
    expect(
      await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth),
    ).toBe(true);
  });

  test('replaces the local review loop with its responsive editorial image', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 });
    await page.goto('/blog/human-harnesses');

    const figure = page.locator('[data-blog-section-image]');
    const image = figure.getByRole('img', {
      name: 'Separate review and repair stations passing the same verified artifact around a loop.',
    });
    await image.scrollIntoViewIfNeeded();
    await expect(image).toBeVisible();
    await expect.poll(() => image.evaluate((element) => (element as HTMLImageElement).naturalWidth))
      .toBeGreaterThan(0);
    await expect(image).toHaveAttribute('src', /04-review-repair-loop\.webp/);
    await expect(
      page.getByText('implement -> verify -> review exact head -> disposition', { exact: false }),
    ).toHaveCount(0);
    await expect(figure).toContainText(
      'Review and repair stay separate while verification sends the same artifact back around the loop.',
    );
    expect(
      await figure.evaluate((element) => ({
        previousBlock: (element.previousElementSibling as HTMLElement | null)?.dataset.blockIndex,
        nextBlock: (element.nextElementSibling as HTMLElement | null)?.dataset.blockIndex,
      })),
    ).toEqual({ previousBlock: '2', nextBlock: '3' });

    const box = await figure.boundingBox();
    expect(box).not.toBeNull();
    expect(box!.x).toBeGreaterThanOrEqual(0);
    expect(box!.x + box!.width).toBeLessThanOrEqual(390);
    expect(
      await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth),
    ).toBe(true);
  });

  test('links the blog from the desktop navbar', async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    await page.goto('/');

    // Wait for hydration so the click isn't swallowed by React's
    // pre-hydration event replay.
    await awaitNavbarHydrated(page);
    await page.getByRole('navigation').getByRole('link', { name: 'Blog', exact: true }).click();
    await expect(page).toHaveURL(/\/blog$/);
  });

  test('keeps desktop navigation and search controls clear of each other', async ({ page }) => {
    await page.setViewportSize({ width: 1152, height: 800 });
    await page.goto('/');
    await awaitNavbarHydrated(page);

    const navLinks = page.locator('[data-navbar-links]');
    const search = page.getByRole('button', { name: 'Search documentation' });
    const ctas = page.locator('.navbar-desktop-cta');
    await expect(navLinks).toBeVisible();
    await expect(search).toBeVisible();
    await expect(ctas).toBeVisible();

    const [linksBox, searchBox] = await Promise.all([
      navLinks.boundingBox(),
      search.boundingBox(),
    ]);
    expect(linksBox).not.toBeNull();
    expect(searchBox).not.toBeNull();
    expect(linksBox!.x + linksBox!.width + 12).toBeLessThanOrEqual(searchBox!.x);
  });

  test('keeps navbar CTAs globally and moves blog sign-in above GitHub discussion', async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const post = BLOG_POSTS[0];
    const discussionUrl = expectedGitHubDiscussionUrl(post.title);

    await page.goto(`/blog/${post.slug}`);
    await awaitNavbarHydrated(page);
    await expect(page.locator('header').getByRole('link', { name: 'Get Started' })).toBeVisible();
    await expect(page.locator('header').getByRole('link', { name: 'Get Demo' })).toBeVisible();
    await expect(page.locator('header').getByRole('button', { name: 'Sign in' })).toHaveCount(0);
    await expect(page.locator('.home-sidebar-panel [data-blog-auth-card]')).toHaveCount(0);

    const blogDiscussionCard = page.locator('.page-aside-panel [data-blog-auth-card]');
    await expect(blogDiscussionCard).toBeVisible();
    await expect(blogDiscussionCard.getByText('Join blog discussion')).toBeVisible();
    await expect(blogDiscussionCard.getByRole('button', { name: 'Sign in' })).toBeVisible();
    await expect(blogDiscussionCard.getByRole('link', { name: 'Open GitHub Discussion' })).toHaveCount(0);

    const githubDiscussionLink = page.locator('.page-aside-panel [data-github-discussion-link]');
    await expect(githubDiscussionLink).toHaveAttribute('href', discussionUrl);
    await expect(githubDiscussionLink).toContainText('GitHub discussion');

    await page.goto('/');
    await awaitNavbarHydrated(page);
    await expect(page.locator('header').getByRole('link', { name: 'Get Demo' })).toBeVisible();
    await expect(page.locator('[data-blog-auth-card]')).toHaveCount(0);

    await page.goto('/docs');
    await awaitNavbarHydrated(page);
    await expect(page.locator('header').getByRole('link', { name: 'Get Demo' })).toBeVisible();
    await expect(page.locator('[data-blog-auth-card]')).toHaveCount(0);

    await page.goto('/bookmarks');
    await awaitNavbarHydrated(page);
    await expect(page.locator('.page-aside-panel [data-blog-auth-card]')).toBeVisible();
    await expect(page.locator('.page-aside-panel [data-github-discussion-link]')).toBeVisible();
  });

  test('mounts both sidebars on every desktop page shell', async ({ page }) => {
    await page.setViewportSize({ width: 1560, height: 1000 });

    await expectDesktopRails(page, '/', '.home-sidebar-panel', '.home-aside-panel');
    await expectDesktopRails(page, '/blog', '.home-sidebar-panel', '.page-aside-panel');
    await expectDesktopRails(
      page,
      `/blog/${BLOG_POSTS[0].slug}`,
      '.home-sidebar-panel',
      '.page-aside-panel',
    );
    await expectDesktopRails(page, '/bookmarks', '.home-sidebar-panel', '.page-aside-panel');
    await expectDesktopRails(page, '/changelog', '.home-sidebar-panel', '.page-aside-panel');
    await expectDesktopRails(page, '/patterns', '.home-sidebar-panel', '.page-aside-panel');
    await expectDesktopRails(page, '/docs', '.docs-sidebar-panel', '.docs-toc-panel');
  });
});
