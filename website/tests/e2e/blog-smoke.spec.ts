import { expect, test } from '@playwright/test';

import { BLOG_POSTS } from '@/lib/blog';

test.describe('blog UI smoke', () => {
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

  test('links the blog from the desktop navbar', async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    await page.goto('/');

    // Wait for hydration (session slot renders) so the click isn't
    // swallowed by React's pre-hydration event replay.
    await expect(page.getByRole('button', { name: 'Sign in' })).toBeVisible();
    await page.getByRole('navigation').getByRole('link', { name: 'Blog', exact: true }).click();
    await expect(page).toHaveURL(/\/blog$/);
  });

  test('swaps Get Demo for the auth slot on blog routes only', async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });

    await page.goto('/blog/one-cli-to-rule-them-all');
    await expect(
      page
        .locator('header button[aria-label="Account menu"], header button:has-text("Sign in")')
        .first(),
    ).toBeVisible();
    await expect(page.locator('header').getByRole('link', { name: 'Get Demo' })).toBeHidden();

    await page.goto('/');
    await expect(page.locator('header').getByRole('link', { name: 'Get Demo' })).toBeVisible();
  });
});
