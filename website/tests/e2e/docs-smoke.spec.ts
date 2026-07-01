import { expect, test } from '@playwright/test';

test.describe('docs UI smoke', () => {
  test('renders the homepage terminal without breaking the app shell', async ({ page }) => {
    await page.goto('/');

    await expect(page.getByRole('heading', { level: 1 })).toContainText('Extract');
    await expect(page.getByText('pm · quickstart')).toBeVisible();
    await expect(page.getByText('pm etl run --connection my-github')).toBeVisible();
  });

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

    await page.goto('/docs/connectors/source-100ms');
    const connectorToc = page.getByRole('navigation', { name: 'On this page' });
    await expect(connectorToc).toBeVisible();
    await expect(connectorToc.getByRole('link', { name: 'Configuration' })).toBeVisible();
  });

  test('renders connector catalog, connector detail, and generated data.json', async ({ page }) => {
    await page.goto('/docs/connectors');

    await expect(
      page.getByRole('heading', { level: 1, name: 'Search every connector page' }),
    ).toBeVisible();

    await page.getByLabel('Search connectors').fill('100ms');
    await expect(page.getByRole('link', { name: /100ms/ })).toBeVisible();

    await page.goto('/docs/connectors/source-100ms');
    await expect(page.getByRole('heading', { level: 1, name: '100ms' })).toBeVisible();
    await expect(page.getByRole('link', { name: 'data.json' })).toHaveAttribute(
      'href',
      '/docs/connectors/source-100ms/data.json',
    );

    const response = await page.request.get('/docs/connectors/source-100ms/data.json');
    expect(response.ok()).toBe(true);
    expect(response.headers()['content-type']).toMatch(/application\/json/);

    const payload = await response.json();
    expect(payload).toMatchObject({
      slug: 'source-100ms',
      name: '100ms',
      enrichment: expect.any(Object),
    });
    expect(payload).not.toHaveProperty('docUrl');
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
