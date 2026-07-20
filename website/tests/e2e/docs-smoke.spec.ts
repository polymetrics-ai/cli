import { expect, test } from '@playwright/test';
import type { Page } from '@playwright/test';

test.describe('docs UI smoke', () => {
  async function awaitNavbarHydrated(page: Page): Promise<void> {
    await expect(page.locator('header[data-navbar-hydrated="true"]')).toBeVisible();
  }

  test('renders an MDX docs page with docs actions', async ({ page }) => {
    await page.goto('/docs/quickstart');

    await expect(page.getByRole('heading', { level: 1, name: 'Quickstart' })).toBeVisible();
    await expect(page.getByRole('heading', { name: 'Install pm' })).toBeVisible();
    await expect(page.getByRole('link', { name: 'llms.txt' })).toHaveAttribute(
      'href',
      '/llms.txt',
    );
    await expect(page.getByText('pm etl run --connection demo --stream customers --json')).toBeVisible();
  });

  test('renders the desktop on-page TOC and follows heading anchors', async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    await page.goto('/docs/quickstart');

    const toc = page.getByRole('navigation', { name: 'On this page' });
    await expect(toc).toBeVisible();
    await expect(page.locator('[data-site-toc]')).toHaveCount(1);

    await expect(toc.getByRole('link', { name: 'Install pm' })).toHaveAttribute(
      'aria-current',
      'location',
    );

    await toc.getByRole('link', { name: 'Extract data' }).click();
    await expect(page).toHaveURL(/\/docs\/quickstart#extract-data$/);
    await expect(toc.getByRole('link', { name: 'Extract data' })).toHaveAttribute(
      'aria-current',
      'location',
    );

    await page.goto('/docs/connectors/100ms');
    const connectorToc = page.getByRole('navigation', { name: 'On this page' });
    await expect(connectorToc).toBeVisible();
    await expect(page.locator('[data-site-toc]')).toHaveCount(1);
    await expect(connectorToc.getByRole('link', { name: 'Status' })).toBeVisible();
  });

  test('uses one shared right TOC on the homepage and stays light only', async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    await page.addInitScript(() => {
      localStorage.setItem('theme', 'dark');
      document.documentElement.classList.add('dark');
    });
    await page.goto('/');

    const toc = page.getByRole('navigation', { name: 'On this page' });
    await expect(toc).toBeVisible();
    await expect(page.locator('[data-site-toc]')).toHaveCount(1);
    await expect(toc.getByRole('link', { name: 'Overview' })).toHaveAttribute(
      'aria-current',
      'location',
    );

    // Wait for a client-only navbar marker; clicking earlier can race
    // React's pre-hydration event replay.
    await awaitNavbarHydrated(page);
    await toc.getByRole('link', { name: 'The loop' }).click();
    await expect(page).toHaveURL(/\/#loop$/);
    await expect(toc.getByRole('link', { name: 'The loop' })).toHaveAttribute(
      'aria-current',
      'location',
    );

    await expect(page.locator('.site-toc-svg')).toBeVisible();
    await expect(page.locator('.site-toc-path-active')).toHaveCount(1);
    await expect(page.locator('.site-toc-path-active')).toHaveAttribute('d', /M /);
    await expect(page.locator('.site-toc-node rect')).toBeVisible();
    const tocBox = await page.locator('[data-site-toc]').boundingBox();
    expect(Math.round(tocBox?.width ?? 0)).toBe(256);
    const bodyBackground = await page.evaluate(() => getComputedStyle(document.body).backgroundColor);
    const colorScheme = await page.evaluate(
      () => getComputedStyle(document.documentElement).colorScheme,
    );
    expect(bodyBackground).not.toBe('rgb(18, 18, 18)');
    expect(colorScheme).toContain('light');
  });

  test('opens the changelog route from the main navigation', async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    await page.goto('/');

    await awaitNavbarHydrated(page);
    await page.locator('header').getByRole('link', { name: 'Changelog' }).click();
    await expect(page).toHaveURL(/\/changelog$/);
    await expect(
      page.getByRole('heading', { level: 1, name: 'Changes that ship the loop.' }),
    ).toBeVisible();
  });

  test('renders connector catalog, connector detail, and generated data.json', async ({ page }) => {
    await page.goto('/docs/connectors');

    await expect(
      page.getByRole('heading', { level: 1, name: 'Search every connector page' }),
    ).toBeVisible();

    await page.getByLabel('Search connectors').fill('100ms');
    await expect(page.getByRole('link', { name: /100ms/ })).toBeVisible();

    await page.goto('/docs/connectors/100ms');
    await expect(page.getByRole('heading', { level: 1, name: '100ms' })).toBeVisible();
    await expect(page.getByRole('link', { name: 'data.json', exact: true })).toHaveAttribute(
      'href',
      '/docs/connectors/100ms/data.json',
    );

    const response = await page.request.get('/docs/connectors/100ms/data.json');
    expect(response.ok()).toBe(true);
    expect(response.headers()['content-type']).toMatch(/application\/json/);

    const payload = await response.json();
    expect(payload).toMatchObject({
      slug: '100ms',
      name: '100ms',
      category: 'api',
      capabilities: expect.objectContaining({
        read: true,
        write: true,
      }),
    });
    expect(payload.docsMd).toContain('management_token');
  });

  test('opens docs search and returns docs plus connector results', async ({ page }) => {
    await page.goto('/docs');

    await page.getByRole('button', { name: 'Search documentation' }).click();
    const input = page.getByPlaceholder(/Search .* connectors/);
    await expect(input).toBeVisible();

    await input.fill('quickstart');
    await expect(page.getByText('Quickstart').first()).toBeVisible();

    await input.fill('management_token');
    await expect(page.getByText('100ms connector')).toBeVisible();
  });
});
